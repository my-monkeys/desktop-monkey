package vie

import "math"

// Les jauges d'humeur donnent au singe une vie interieure : quatre valeurs
// entre 0 et 1 qui derivent avec le temps, reagissent a ce qu'on lui fait, et
// ponderent ses choix d'activite. Un singe ignore s'ennuie et fait des
// betises ; un singe caline est joueur ; un singe frappe prend peur.
//
// Les facteurs de ponderation valent ~1 aux valeurs de depart : le
// comportement par defaut reste celui d'avant les jauges.

// Jauge est une humeur exposee a l'affichage (le menu du tray).
type Jauge struct {
	Nom    string // identifiant stable : "energie", "ennui", "bonheur", "peur"
	Valeur float64
}

// valeurs de depart (l'etat "neutre" des facteurs)
const (
	energieDepart = 0.8
	ennuiDepart   = 0.2
	bonheurDepart = 0.6
	peurDepart    = 0.0
)

// Jauges renvoie l'etat des humeurs, dans un ordre stable.
func (v *Vie) Jauges() []Jauge {
	return []Jauge{
		{"energie", v.energie},
		{"ennui", v.ennui},
		{"bonheur", v.bonheur},
		{"peur", v.peur},
	}
}

// majJauges fait deriver les humeurs d'un pas de temps. bouge indique un vrai
// mouvement de souris de l'utilisateur (pas un coup de patte du singe).
func (v *Vie) majJauges(dt float64, bouge bool) {
	// l'ennui vient d'un curseur mort ; l'activite le dissipe
	if bouge || v.etat == Joue || v.etat == Chasse || v.etat == Vol || v.etat == Grimpe {
		v.ennui = clampJauge(v.ennui - 0.08*dt)
	} else {
		v.ennui = clampJauge(v.ennui + 0.012*dt)
	}

	// l'energie s'use, la sieste la repare (le repas aussi, en evenement)
	switch v.etat {
	case Sieste:
		v.energie = clampJauge(v.energie + 0.06*dt)
	case Chasse, Vol, Grimpe, Joue, Fuite:
		v.energie = clampJauge(v.energie - 0.010*dt)
	default:
		v.energie = clampJauge(v.energie - 0.004*dt)
	}

	// le bonheur revient doucement vers le neutre, la peur s'estompe
	v.bonheur += (bonheurDepart - v.bonheur) * 0.02 * dt
	v.peur = clampJauge(v.peur - 0.03*dt)
}

// --- facteurs de ponderation (≈1 aux valeurs de depart) -----------------------

// facteurAmi module l'envie d'aller voir le curseur : un singe heureux y court,
// un singe apeure s'abstient.
func (v *Vie) facteurAmi() float64 {
	return (0.4 + v.bonheur) * (1 - 0.6*v.peur)
}

// facteurEnnui module les betises (chasse, vol, jet de crotte) : elles montent
// avec l'ennui.
func (v *Vie) facteurEnnui() float64 {
	return 0.8 + v.ennui
}

// facteurActif module les depenses physiques (grimpe, jeu) : il faut s'ennuyer
// un peu et avoir de l'energie.
func (v *Vie) facteurActif() float64 {
	return v.facteurEnnui() * (0.2 + v.energie)
}

// FacteurEnnui est expose pour la scene (le jet de crotte est decide la-bas).
func (v *Vie) FacteurEnnui() float64 { return v.facteurEnnui() }

// pondere applique un facteur d'humeur a une chance de base. Une base >= 1
// signifie "toujours" (reglage force, tests, demos) : elle ignore les humeurs.
func pondere(base, facteur float64) float64 {
	if base >= 1 {
		return base
	}
	return base * facteur
}

// --- evenements ---------------------------------------------------------------

func (v *Vie) jaugesCoup() {
	v.bonheur = clampJauge(v.bonheur - 0.15)
	v.peur = clampJauge(v.peur + 0.3)
}

func (v *Vie) jaugesGuili() {
	v.bonheur = clampJauge(v.bonheur + 0.25)
	v.ennui = clampJauge(v.ennui - 0.5)
	v.peur = clampJauge(v.peur - 0.2)
}

func (v *Vie) jaugesRepas() {
	v.energie = clampJauge(v.energie + 0.3)
}

// --- guili : secouer la souris pres de lui le chatouille ----------------------

const (
	guiliRayon    = 150 // distance du centre en dessous de laquelle ca chatouille
	guiliAmpMin   = 10  // amplitude minimale d'un aller-retour, en pixels
	guiliFenetre  = 1.1 // secondes pour accumuler les demi-tours
	guiliCooldown = 3.0 // secondes entre deux fous rires
)

// detecterGuili compte les demi-tours du curseur pres du singe, dans n'importe
// quelle direction (on secoue comme on veut : de gauche a droite, de haut en
// bas, en diagonale). dxg/dyg est le deplacement du curseur depuis l'image
// precedente ; bouge vaut false si ce mouvement vient du singe lui-meme.
func (v *Vie) detecterGuili(dt, dxg, dyg, curseurX, curseurY float64, bouge bool) {
	if v.guiliCool > 0 {
		v.guiliCool -= dt
	}
	switch v.etat {
	case Mort, Cadavre, Cache, Vol, Touche, Fuite:
		v.guiliInv, v.guiliAmp, v.guiliTemps = 0, 0, 0
		return
	}

	cx, cy := v.Centre()
	pres := math.Hypot(curseurX-cx, curseurY-cy) < guiliRayon
	if !pres || !bouge {
		if !pres {
			v.guiliInv, v.guiliAmp, v.guiliTemps = 0, 0, 0
			v.guiliPvx, v.guiliPvy = 0, 0
		}
		return
	}

	v.guiliTemps += dt
	if v.guiliTemps > guiliFenetre {
		v.guiliInv, v.guiliAmp, v.guiliTemps = 0, 0, 0
	}
	d := math.Hypot(dxg, dyg)
	if d < 2 {
		return
	}
	// un demi-tour : le deplacement repart a l'oppose du precedent
	if (v.guiliPvx != 0 || v.guiliPvy != 0) &&
		dxg*v.guiliPvx+dyg*v.guiliPvy < 0 && v.guiliAmp >= guiliAmpMin {
		v.guiliInv++
		v.guiliAmp = 0
	}
	v.guiliPvx, v.guiliPvy = dxg, dyg
	v.guiliAmp += d

	if v.guiliInv >= 3 && v.guiliCool <= 0 {
		v.guiliInv, v.guiliAmp, v.guiliTemps = 0, 0, 0
		v.guiliCool = guiliCooldown
		v.jaugesGuili()
		v.faceAu(curseurX, curseurY)
		v.Evenement = "guili"
	}
}

func clampJauge(x float64) float64 {
	return math.Max(0, math.Min(1, x))
}
