package vie

import (
	"math"
	"testing"
)

// Aux valeurs de depart, les facteurs de ponderation valent ~1 : les jauges ne
// changent pas l'equilibre par defaut du singe.
func TestFacteursNeutresAuDepart(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	for nom, f := range map[string]float64{
		"ami":   v.facteurAmi(),
		"ennui": v.facteurEnnui(),
		"actif": v.facteurActif(),
	} {
		if math.Abs(f-1) > 0.05 {
			t.Errorf("facteur %s = %.2f au depart, attendu ~1", nom, f)
		}
	}
}

// Un curseur immobile ennuie le singe, peu a peu.
func TestEnnuiMonteQuandLaSourisDort(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceChasse, r.ChanceGrimpe, r.ChanceJeu = 0, 0, 0
	r.AvantSieste = 1e9
	v := nouvelleVie(t, r)

	avant := v.ennui
	for i := 0; i < 60*30; i++ { // 30 s sans bouger
		v.Avancer(dt, 5, 5, false)
	}
	if v.ennui <= avant {
		t.Errorf("ennui %.2f apres 30 s d'immobilite, attendu > %.2f", v.ennui, avant)
	}
}

// Un coup le rend malheureux et craintif.
func TestCoupEffraie(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	bonheurAvant := v.bonheur
	v.Coup()
	if v.peur <= 0 {
		t.Error("la peur devrait monter apres un coup")
	}
	if v.bonheur >= bonheurAvant {
		t.Error("le bonheur devrait baisser apres un coup")
	}
}

// Secouer la souris pres de lui declenche le guili et le rend heureux ; le meme
// va-et-vient loin de lui ne fait rien.
func TestGuiliRendHeureux(t *testing.T) {
	secouer := func(v *Vie, cx, cy float64) bool {
		// oscillation horizontale de ±20 px, plusieurs allers-retours
		for i := 0; i < 60; i++ {
			x := cx + 20*math.Sin(float64(i)*0.8)
			v.Avancer(dt, x, cy, false)
			if v.Evenement == "guili" {
				return true
			}
		}
		return false
	}

	r := ReglagesParDefaut()
	r.ChanceAmi, r.ChanceChasse = 0, 0
	v := nouvelleVie(t, r)
	bonheurAvant := v.bonheur
	cx, cy := v.Centre()
	if !secouer(v, cx+30, cy) {
		t.Fatal("secouer la souris pres de lui devrait declencher un guili")
	}
	if v.bonheur <= bonheurAvant {
		t.Error("le guili devrait le rendre plus heureux")
	}

	v2 := nouvelleVie(t, r)
	cx2, cy2 := v2.Centre()
	if secouer(v2, cx2+500, cy2) {
		t.Error("secouer loin de lui ne devrait pas le chatouiller")
	}

	// le secouage vertical chatouille aussi
	v3 := nouvelleVie(t, r)
	cx3, cy3 := v3.Centre()
	vertical := false
	for i := 0; i < 60; i++ {
		y := cy3 + 20*math.Sin(float64(i)*0.8)
		v3.Avancer(dt, cx3+30, y, false)
		if v3.Evenement == "guili" {
			vertical = true
			break
		}
	}
	if !vertical {
		t.Error("secouer verticalement pres de lui devrait le chatouiller")
	}
}
