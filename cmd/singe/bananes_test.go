package main

import (
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
