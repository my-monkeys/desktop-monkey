package vie

import (
	"math"
	"testing"
)

// pasPourSeuil est le nombre de pas de temps qui depasse le seuil d'appui long.
var pasPourSeuil = int(math.Ceil(seuilAppuiLong/dt)) + 1

// surLui renvoie un point du corps du singe, celui qu'un clic devrait toucher.
func surLui(v *Vie) (float64, float64) { return v.Centre() }

func TestAppuiCourtEstUnCoup(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	x, y := surLui(v)

	v.Avancer(dt, x, y, true)  // on appuie sur lui
	v.Avancer(dt, x, y, false) // et on relache aussitot

	if v.Etat() != Touche {
		t.Fatalf("etat %v, attendu touche", v.Etat())
	}
	if restants, _ := v.Coeurs(); restants != 2 {
		t.Fatalf("%d coeurs restants, attendu 2", restants)
	}
}

func TestAppuiMaintenuLeSouleve(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	x, y := surLui(v)

	// on garde le bouton enfonce au-dela du seuil
	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, x, y, true)
	}
	if v.Etat() != Porte {
		t.Fatalf("etat %v, attendu porte", v.Etat())
	}
	if restants, _ := v.Coeurs(); restants != 3 {
		t.Fatalf("attraper lui a coute un coeur (%d restants)", restants)
	}

	// il suit le curseur tant qu'on ne lache pas
	avantX, avantY := v.X, v.Y
	v.Avancer(dt, x+180, y-90, true)
	if v.X-avantX < 170 || avantY-v.Y < 80 {
		t.Fatalf("il n'a pas suivi le curseur : %.0f,%.0f -> %.0f,%.0f",
			avantX, avantY, v.X, v.Y)
	}
}

func TestLacheEnLAirIlRetombe(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	x, y := surLui(v)

	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, x, y, true)
	}
	// on le monte tout en haut, puis on lache
	v.Avancer(dt, x, 60, true)
	v.Avancer(dt, x, 60, false)
	if v.Etat() != Chute {
		t.Fatalf("etat %v, attendu chute", v.Etat())
	}

	sol := v.solY()
	for i := 0; i < 600 && v.Etat() == Chute; i++ {
		v.Avancer(dt, x, 60, false)
	}
	if v.Etat() != Repos {
		t.Fatalf("apres la chute : etat %v, attendu repos", v.Etat())
	}
	if v.Y != sol {
		t.Fatalf("il a atterri a %.1f, le sol est a %.1f", v.Y, sol)
	}
}

func TestAppuiAcoteNeFaitRien(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	depart := v.Etat()

	// un appui long loin de lui ne doit ni le frapper ni le soulever
	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, 5, 5, true)
	}
	if v.Etat() == Porte || v.Etat() == Touche {
		t.Fatalf("etat %v apres un appui a cote (depart %v)", v.Etat(), depart)
	}
	if restants, _ := v.Coeurs(); restants != 3 {
		t.Fatalf("%d coeurs restants, attendu 3", restants)
	}
}
