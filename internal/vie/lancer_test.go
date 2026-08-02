package vie

import "testing"

// chasseur met le singe en chasse et lui garde une proie vivante : sans
// mouvement du curseur il se lasse et abandonne.
func chasseur(t *testing.T, cible [2]float64, pas int, f func(l Lancer)) {
	t.Helper()
	v := nouvelleVie(t, ReglagesParDefaut())
	v.passerA(Chasse)
	for i := 0; i < pas; i++ {
		x := cible[0] + float64(i%7) // la proie fremit, il reste interesse
		v.Avancer(dt, x, cible[1], false)
		if l, ok := v.PrendreLancer(); ok {
			f(l)
		}
		if v.Etat() != Chasse {
			v.passerA(Chasse) // vol de curseur ou lassitude : on le remet en chasse
		}
	}
}

// Les deux gestes de lancer doivent tous les deux sortir.
func TestLancerAlterneLesDeuxGestes(t *testing.T) {
	vus := map[string]int{}
	chasseur(t, [2]float64{1500, 300}, 6000, func(l Lancer) { vus[l.Action]++ })

	if len(vus) < 2 {
		t.Fatalf("un seul geste sur %d lancers : %v", total(vus), vus)
	}
	for _, geste := range []string{"lance", "lance_saut"} {
		if vus[geste] == 0 {
			t.Errorf("le geste %q n'est jamais sorti (%v)", geste, vus)
		}
	}
}

// Il doit pouvoir jeter sa banane de l'autre bout de l'ecran : avant, l'attaque
// n'existait qu'a bout de bras.
func TestIlLanceDeLoin(t *testing.T) {
	n := 0
	var vu Lancer
	chasseur(t, [2]float64{1850, 60}, 600, func(l Lancer) { n++; vu = l })

	if n == 0 {
		t.Fatal("aucune banane lancee sur une proie lointaine")
	}
	if vu.CibleX < 1800 {
		t.Errorf("il visait %.0f, pas la proie", vu.CibleX)
	}
	if vu.Direction != "droite" {
		t.Errorf("il visait a droite mais regarde vers %q", vu.Direction)
	}
}

func total(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}
