package main

import (
	"image"
	"math"
	"testing"

	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
	"github.com/my-monkeys/desktop-monkey/internal/vie"
)

// La banane doit se decouper toute seule du sprite de lancer : c'est elle qui
// devient le projectile.
func TestBananeSeDecoupeDuSprite(t *testing.T) {
	p, err := planche.Charger(ressources.Fichiers, "assets/singe2.json", 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"lance", "lance_saut"} {
		for _, sens := range []string{"droite", "gauche"} {
			a := p.Obtenir(action, sens)
			if a == nil {
				t.Fatalf("%s_%s absent de la planche", action, sens)
			}
			boite, ok := amasJaune(a.Images[len(a.Images)-1])
			if !ok {
				t.Fatalf("%s_%s : pas de banane dans la derniere image", action, sens)
			}
			// une banane, pas un pixel perdu ni la moitie du singe
			if l, h := boite.Dx(), boite.Dy(); l < 3 || l > 12 || h < 4 || h > 14 {
				t.Errorf("%s_%s : banane de %dx%d, taille inattendue", action, sens, l, h)
			}
			// elle a quitte la main : loin du bord ou il la tient
			if sens == "droite" && boite.Min.X < a.Images[0].Bounds().Dx()/2 {
				t.Errorf("%s_droite : la banane est encore sur lui (x=%d)", action, boite.Min.X)
			}
		}
	}
}

// Le singe RPG n'a pas de banane : la mecanique doit s'effacer sans bruit.
func TestPlancheSansBananeNeCassePas(t *testing.T) {
	p, err := planche.Charger(ressources.Fichiers, "assets/singe.json", 1)
	if err != nil {
		t.Fatal(err)
	}
	a := p.Obtenir("lance", "droite") // il se rabat sur le repos
	if a == nil || len(a.Images) == 0 {
		t.Fatal("aucune image de repli")
	}
	if _, ok := amasJaune(a.Images[len(a.Images)-1]); ok {
		t.Error("une banane a ete trouvee dans une planche qui n'en a pas")
	}
}

// sceneBanane monte le strict necessaire pour eprouver le vol : la planche du
// singe, la banane decoupee, et les dimensions de l'ecran.
func sceneBanane(t *testing.T) *scene {
	t.Helper()
	echelleAff = 1 // taille de reference, pour que les mesures soient stables
	p, err := planche.Charger(ressources.Fichiers, "assets/singe2.json", echelleAff)
	if err != nil {
		t.Fatal(err)
	}
	s := &scene{pSinge: p, ecranL: 1680, hautEcran: 1050}
	s.chargerBananes()
	if s.bananeImg == nil {
		t.Fatal("la banane n'a pas ete decoupee du sprite")
	}
	return s
}

// La banane part de la main, passe par le curseur, puis retombe hors de l'ecran.
func TestBananeSuitSaCourbeJusquHorsEcran(t *testing.T) {
	s := sceneBanane(t)
	const cibleX, cibleY = 1200, 400
	b, ok := s.nouvelleBanane(vie.Lancer{
		CibleX: cibleX, CibleY: cibleY, Action: "lance", Direction: "droite",
	}, 200, 500)
	if !ok {
		t.Fatal("le jet a ete refuse")
	}
	if b.fx < 200 || b.fx > 200+float64(s.pSinge.Largeur) {
		t.Errorf("elle part de %.0f, hors de la cellule du singe", b.fx)
	}

	// on la suit pas a pas : elle doit froler la cible, puis sortir par le bas
	const dt = 1.0 / 60
	plusPres, sortie, monte := 1e9, false, false
	y0 := b.fy
	for i := 0; i < 600; i++ {
		b.vy += graviteBanane * dt
		b.fx += b.vx * dt
		b.fy += b.vy * dt
		if b.fy < y0-4 {
			monte = true // la cloche : elle s'eleve avant de retomber
		}
		if d := math.Hypot(cibleX-b.fx, cibleY-b.fy); d < plusPres {
			plusPres = d
		}
		if b.fy > float64(s.hautEcran) {
			sortie = true
			break
		}
	}
	if !monte {
		t.Error("elle n'est jamais montee : ce n'est pas une cloche mais une ligne")
	}
	if plusPres > 40 {
		t.Errorf("elle est passee a %.0f px de la cible, trop loin", plusPres)
	}
	if !sortie {
		t.Error("elle n'est jamais sortie par le bas de l'ecran")
	}
}

// Un jet vers la gauche part bien a gauche.
func TestBananeVersLaGauche(t *testing.T) {
	s := sceneBanane(t)
	b, ok := s.nouvelleBanane(vie.Lancer{
		CibleX: 100, CibleY: 500, Action: "lance_saut", Direction: "gauche",
	}, 900, 500)
	if !ok {
		t.Fatal("le jet a ete refuse")
	}
	if b.vx >= 0 {
		t.Errorf("vitesse horizontale %.0f : elle ne part pas vers la gauche", b.vx)
	}
}

// suivre fait voler une banane image par image et dit si elle a touche le
// curseur en chemin, sans passer par les fenetres du systeme.
func suivre(b banane, img image.Rectangle, curseurX, curseurY float64, pas int) bool {
	demiL, demiH := float64(img.Dx())/2, float64(img.Dy())/2
	rayon := math.Hypot(demiL, demiH) + margeTouche*echelleAff
	const dt = 1.0 / 60
	for i := 0; i < pas; i++ {
		avantX, avantY := b.fx+demiL, b.fy+demiH
		b.vy += graviteBanane * dt
		b.fx += b.vx * dt
		b.fy += b.vy * dt
		if distanceAuSegment(curseurX, curseurY, avantX, avantY,
			b.fx+demiL, b.fy+demiH) <= rayon {
			return true
		}
	}
	return false
}

// Une banane lancee sur le curseur doit le toucher, pas le traverser.
func TestBananeToucheLeCurseur(t *testing.T) {
	s := sceneBanane(t)
	for _, cible := range [][2]float64{{1200, 400}, {1600, 900}, {260, 520}, {900, 120}} {
		b, ok := s.nouvelleBanane(vie.Lancer{
			CibleX: cible[0], CibleY: cible[1], Action: "lance", Direction: "droite",
		}, 200, 500)
		if !ok {
			t.Fatal("le jet a ete refuse")
		}
		if !suivre(b, s.bananeImg.Bounds(), cible[0], cible[1], 400) {
			t.Errorf("la banane a traverse le curseur en %.0f,%.0f sans le toucher",
				cible[0], cible[1])
		}
	}
}

// Meme tres rapide, elle ne doit pas sauter par-dessus le curseur entre deux
// images : c'est tout le trajet qui compte, pas la seule position d'arrivee.
func TestBananeRapideNeTraversePas(t *testing.T) {
	s := sceneBanane(t)
	img := s.bananeImg.Bounds()
	// 4200 px/s, soit 70 px par image : bien plus que sa largeur. Le curseur est
	// pose entre deux positions echantillonnees (x = 380 puis 450), la ou un
	// simple test d'arrivee ne verrait jamais rien.
	depart := banane{fx: 100, fy: 500, vx: 4200, vy: 0}
	if !suivre(depart, img, 400, 505, 60) {
		t.Error("la banane a franchi le curseur d'un bond")
	}
	// et pour prouver que c'est bien le trajet qui sauve : sans lui, elle rate
	if toucheSansTrajet(depart, img, 400, 505, 60) {
		t.Error("le test image par image suffisait : le cas n'est pas probant")
	}
}

// toucheSansTrajet refait le vol en ne regardant que les positions d'arrivee,
// comme le ferait un test naif : il sert de temoin.
func toucheSansTrajet(b banane, img image.Rectangle, curseurX, curseurY float64, pas int) bool {
	demiL, demiH := float64(img.Dx())/2, float64(img.Dy())/2
	rayon := math.Hypot(demiL, demiH) + margeTouche*echelleAff
	const dt = 1.0 / 60
	for i := 0; i < pas; i++ {
		b.vy += graviteBanane * dt
		b.fx += b.vx * dt
		b.fy += b.vy * dt
		if math.Hypot(curseurX-(b.fx+demiL), curseurY-(b.fy+demiH)) <= rayon {
			return true
		}
	}
	return false
}

// Et elle ne doit pas toucher ce qu'elle ne croise pas.
func TestBananeManqueLeCurseurEloigne(t *testing.T) {
	s := sceneBanane(t)
	b, ok := s.nouvelleBanane(vie.Lancer{
		CibleX: 1200, CibleY: 400, Action: "lance", Direction: "droite",
	}, 200, 500)
	if !ok {
		t.Fatal("le jet a ete refuse")
	}
	if suivre(b, s.bananeImg.Bounds(), 300, 90, 400) {
		t.Error("elle a touche un curseur qui n'etait pas sur son chemin")
	}
}
