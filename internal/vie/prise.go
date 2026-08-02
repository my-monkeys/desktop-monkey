package vie

import "math"

// Attraper le singe a la souris.
//
// Un appui sur lui veut dire deux choses selon sa duree : relache tout de
// suite, c'est un coup ; maintenu, on l'attrape et il suit le curseur tant
// qu'on ne lache pas. Le verdict ne peut donc pas tomber a l'enfoncement : on
// retient l'appui le temps de savoir, ce qui suspend brievement son
// comportement — ce qui se voit comme une hesitation, pas comme un blocage.
//
// Lache en l'air, il retombe (etat Chute, meme gravite que le cadavre).

// seuilAppuiLong est le temps au-dela duquel un appui n'est plus un coup mais
// une prise. Assez court pour que taper reste vif, assez long pour qu'un clic
// maladroit ne le souleve pas.
const seuilAppuiLong = 0.22

// hauteurLachePeur est la chute a partir de laquelle il a vraiment peur en
// arrivant.
const hauteurLachePeur = 220

// appuyer arbitre l'appui en cours. Il renvoie vrai quand il a pris la main sur
// ce pas de temps (coup porte, prise commencee, ou verdict encore en attente).
func (v *Vie) appuyer(dt float64, curseurX, curseurY float64, boutonEnfonce, frontDescendant bool) bool {
	if frontDescendant && v.Visible() && v.etat != Mort && v.dansLeCadre(curseurX, curseurY) {
		v.appuiSur = true
		v.appuiDepuis = 0
	}
	if !v.appuiSur {
		return false
	}

	v.appuiDepuis += dt
	if !boutonEnfonce {
		// relache avant le seuil : c'etait un coup
		v.appuiSur = false
		v.Coup()
		return true
	}
	if v.appuiDepuis >= seuilAppuiLong {
		v.appuiSur = false
		v.attraper(curseurX, curseurY)
	}
	return true
}

// attraper le souleve : il quitte ce qu'il faisait et suit le curseur.
func (v *Vie) attraper(curseurX, curseurY float64) {
	v.prise = [2]float64{curseurX - v.X, curseurY - v.Y}
	v.passerA(Porte)
	v.Evenement = "attrape"
	v.jaugesAttrape()
}

// porter le fait suivre le curseur, et le lache des que le bouton remonte.
func (v *Vie) porter(dt float64, curseurX, curseurY float64, boutonEnfonce bool) {
	if !boutonEnfonce {
		v.lacher()
		return
	}
	m := v.r.MargeBord
	v.X = clamp(curseurX-v.prise[0], m, v.ecranL-v.largeur-m)
	v.Y = clamp(curseurY-v.prise[1], m, v.solY())
	v.jaugesPorte(dt)
}

// lacher le laisse retomber : au sol il reprend sa vie, de haut il a peur.
func (v *Vie) lacher() {
	if v.Y >= v.solY() {
		v.Y = v.solY()
		v.passerA(Repos)
		v.Evenement = "repose"
		return
	}
	v.hauteurLache = v.solY() - v.Y
	v.vy = 0
	v.passerA(Chute)
}

// chuter applique la gravite jusqu'au sol, puis le remet sur pattes.
func (v *Vie) chuter(dt float64) {
	v.tomber(dt)
	if v.Y >= v.solY() && math.Abs(v.vy) < 1 {
		if v.hauteurLache >= hauteurLachePeur {
			v.jaugesChute()
			v.Evenement = "aie_chute" // la mauvaise reception a deja ses repliques
		} else {
			v.Evenement = "repose"
		}
		v.hauteurLache = 0
		v.passerA(Repos)
	}
}
