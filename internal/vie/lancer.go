package vie

// Le lancer de banane.
//
// La planche porte deux gestes de lancer : un simple, et un ou il saute en
// arriere avant de jeter. Il tire l'un ou l'autre a chaque fois, ce qui suffit
// a rendre la chasse moins mecanique.
//
// La banane elle-meme n'est pas dessinee ici : le geste fini, la vie signale un
// lancer, et la scene fait apparaitre une vraie banane a la place de celle du
// sprite, qui part alors en cloche (voir cmd/singe/bananes.go). C'est ce qui
// lui permet d'attaquer de loin : avant, l'attaque n'existait qu'a bout de
// bras.

import "math"

// Lancer decrit un jet de banane qui vient de partir : ou il visait, et dans
// quelle pose, pour que la scene sache d'ou sort le fruit.
type Lancer struct {
	CibleX, CibleY float64
	Action         string // "lance" ou "lance_saut"
	Direction      string
}

// gestesLancer sont les deux facons de jeter une banane, par ordre de
// preference si la planche ne les connait pas toutes.
var gestesLancer = []string{"lance", "lance_saut"}

// commencerLancer arme le geste : il se tourne vers sa cible et joue l'une des
// deux animations jusqu'au bout.
func (v *Vie) commencerLancer(curseurX, curseurY float64) {
	v.faceAu(curseurX, curseurY)

	geste := gestesLancer[v.alea.Intn(len(gestesLancer))]
	v.animCourante = choisir(v.p, geste, "attaque", "marche")
	v.animMS = 0
	v.lanceAction = v.animCourante

	a, _ := v.Animation()
	if a == nil || len(a.Images) == 0 {
		v.lanceMS = 0.3
		return
	}
	v.lanceMS = float64(len(a.Images)*a.MS) / 1000
	v.depuisLancer = 0
}

// avancerLancer fait s'ecouler le geste. La banane part a la derniere image,
// la ou le sprite la montre deja loin de sa main : le relais est invisible.
func (v *Vie) avancerLancer(dt, curseurX, curseurY float64) {
	v.lanceMS -= dt
	if v.lanceMS > 0 {
		return
	}
	v.lanceMS = 0
	v.lanceEnAttente = true
	v.lancer = Lancer{
		CibleX: curseurX, CibleY: curseurY,
		Action: v.lanceAction, Direction: v.direction,
	}
	if !v.attaqueDite {
		v.attaqueDite = true
		v.Evenement = "attaque"
	}
	// de pres, le jet bouscule le curseur comme le faisait le coup de patte —
	// et peut toujours tourner au vol de la fleche
	cx, cy := v.Centre()
	if math.Hypot(curseurX-cx, curseurY-cy) <= v.r.PorteeAttaque {
		v.porterCoup(curseurX, curseurY)
	}
}

// PrendreLancer renvoie le jet qui vient de partir, et le consomme. La scene
// s'en sert pour faire voler une vraie banane.
func (v *Vie) PrendreLancer() (Lancer, bool) {
	if !v.lanceEnAttente {
		return Lancer{}, false
	}
	v.lanceEnAttente = false
	return v.lancer, true
}
