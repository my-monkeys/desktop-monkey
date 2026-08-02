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

// On le repose ou on veut : il ne retombe pas, il reste ou on l'a laisse.
func TestLacheEnLAirIlResteLa(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	x, y := surLui(v)

	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, x, y, true)
	}
	v.Avancer(dt, x, 60, true)  // on le monte tout en haut
	v.Avancer(dt, x, 60, false) // et on lache
	pose := v.Y

	if v.Etat() != Repos {
		t.Fatalf("etat %v, attendu repos", v.Etat())
	}
	if pose >= v.solY()-1 {
		t.Fatalf("il est deja au sol (%.0f) : le test ne prouve rien", pose)
	}
	// quelques instants plus tard il n'a toujours pas glisse vers le bas
	for i := 0; i < 30; i++ {
		v.Avancer(dt, x, 60, false)
	}
	if v.Y > pose+1 {
		t.Fatalf("il est descendu de %.0f px apres avoir ete repose", v.Y-pose)
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

// Attrape, il vient se caler sous le curseur : celui-ci doit tomber sur son
// corps, quel que soit l'endroit ou on a appuye.
func TestAttrapeIlSeCaleSousLeCurseur(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	// on appuie sur son bord, pas en son milieu
	x, y := v.X+8, v.Y+10
	if !v.dansLeCadre(x, y) {
		x, y = surLui(v)
	}
	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, x, y, true)
	}
	if v.Etat() != Porte {
		t.Fatalf("etat %v, attendu porte", v.Etat())
	}
	for _, p := range [][2]float64{{600, 300}, {900, 700}, {300, 500}} {
		v.Avancer(dt, p[0], p[1], true)
		if !v.dansLeCadre(p[0], p[1]) {
			t.Fatalf("curseur en %.0f,%.0f hors de son corps (singe en %.0f,%.0f)",
				p[0], p[1], v.X, v.Y)
		}
	}
}

// Porte, il regarde du cote ou on l'emmene.
func TestPorteIlRegardeVersOuOnLEmmene(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	x, y := surLui(v)
	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, x, y, true)
	}

	// on le tire vers la gauche, puis vers la droite
	for _, cas := range []struct {
		vers    float64
		attendu string
	}{{-300, "gauche"}, {300, "droite"}, {-300, "gauche"}} {
		but := x + cas.vers
		for i := 0; i < 20; i++ {
			v.Avancer(dt, but, y, true)
		}
		if v.direction != cas.attendu {
			t.Fatalf("emmene vers %s : direction %q, attendu %q",
				cas.attendu, v.direction, cas.attendu)
		}
		x = but
	}
}

// Un tremblement de main ne doit pas le faire loucher.
func TestPorteLeTremblementNeLeFaitPasLoucher(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	x, y := surLui(v)
	for i := 0; i < pasPourSeuil; i++ {
		v.Avancer(dt, x, y, true)
	}
	for i := 0; i < 40; i++ {
		v.Avancer(dt, x-260, y, true) // il regarde a gauche, franchement
	}
	depart := v.direction

	for i := 0; i < 60; i++ { // puis la main tremble de deux pixels
		v.Avancer(dt, x-260+float64(i%2)*2, y, true)
	}
	if v.direction != depart {
		t.Fatalf("le tremblement l'a fait passer de %q a %q", depart, v.direction)
	}
}
