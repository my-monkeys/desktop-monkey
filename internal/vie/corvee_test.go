package vie

import "testing"

// Le bord de l'ecran est hors de portee : la marge de bord empeche son centre
// d'y arriver. Il doit quand meme se declarer arrive, sinon il pietine la avec
// sa crotte a la main sans jamais la lancer.
func TestCorveeArriveMemeSiLeBordEstHorsDePortee(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	_, cy := v.Centre()
	v.EnvoyerVers(0, cy) // le bord gauche, que son centre ne touchera jamais

	for i := 0; i < 60*12 && !v.Arrive(); i++ {
		v.Avancer(dt, 900, 900, false)
	}
	if !v.Arrive() {
		cx, _ := v.Centre()
		t.Fatalf("toujours pas arrive : centre en %.0f, cible 0", cx)
	}
	// il est bien alle jusqu'au mur, pas juste renonce sur place
	if v.X > v.r.MargeBord+1 {
		t.Errorf("il s'est arrete en %.0f, la marge de bord est a %.0f", v.X, v.r.MargeBord)
	}
}

// Les quatre coins de l'ecran sont hors de portee eux aussi : la marge tient le
// singe a distance sur les deux axes a la fois. C'est le cas signale — il
// s'arretait dans le coin, sa crotte a la main, definitivement.
func TestCorveeArriveDansLesCoins(t *testing.T) {
	modele := nouvelleVie(t, ReglagesParDefaut())
	larg, haut := modele.ecranL, modele.ecranH

	for _, coin := range [][2]float64{{0, 0}, {larg, 0}, {0, haut}, {larg, haut}} {
		v := nouvelleVie(t, ReglagesParDefaut())
		v.EnvoyerVers(coin[0], coin[1])
		for i := 0; i < 60*15 && !v.Arrive(); i++ {
			v.Avancer(dt, 900, 900, false)
		}
		if !v.Arrive() {
			t.Errorf("coin %.0f,%.0f : jamais arrive (position %.0f,%.0f)",
				coin[0], coin[1], v.X, v.Y)
			continue
		}
		// il est bien alle se coller au coin, pas renonce en chemin
		m := v.r.MargeBord
		colleX := v.X <= m+1 || v.X >= v.ecranL-v.largeur-m-1
		colleY := v.Y <= m+1 || v.Y >= v.ecranH-v.hauteur-m-1
		if !colleX || !colleY {
			t.Errorf("coin %.0f,%.0f : arrete en %.0f,%.0f, pas dans le coin",
				coin[0], coin[1], v.X, v.Y)
		}
	}
}

// Une cible atteignable reste traitee normalement : il y va, et il y arrive.
func TestCorveeArriveNormalement(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	cible := [2]float64{700, 600}
	v.EnvoyerVers(cible[0], cible[1])
	for i := 0; i < 60*20 && !v.Arrive(); i++ {
		v.Avancer(dt, 900, 900, false)
	}
	if !v.Arrive() {
		t.Fatal("il n'a pas rejoint un point pourtant accessible")
	}
	cx, cy := v.Centre()
	if d := (cx-cible[0])*(cx-cible[0]) + (cy-cible[1])*(cy-cible[1]); d > 36 {
		t.Errorf("arrive a %.0f,%.0f au lieu de %.0f,%.0f", cx, cy, cible[0], cible[1])
	}
}
