package vie

// Attraper le singe a la souris.
//
// Un appui sur lui veut dire deux choses selon sa duree : relache tout de
// suite, c'est un coup ; maintenu, on l'attrape et il suit le curseur tant
// qu'on ne lache pas. Le verdict ne peut donc pas tomber a l'enfoncement : on
// retient l'appui le temps de savoir, ce qui suspend brievement son
// comportement — ce qui se voit comme une hesitation, pas comme un blocage.
//
// On le repose ou on veut : lache, il reprend sa vie sur place. Il se promene
// deja partout sur l'ecran, un singe pose en plein ciel n'a donc rien
// d'anormal — c'est un deplacement, pas une punition.

// seuilAppuiLong est le temps au-dela duquel un appui n'est plus un coup mais
// une prise. Assez court pour que taper reste vif, assez long pour qu'un clic
// maladroit ne le souleve pas.
const seuilAppuiLong = 0.22

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

// attraper le souleve : il vient se caler sous le curseur, pris par la nuque,
// et non pas suspendu la ou le doigt s'est pose — un singe tenu a cote du
// curseur ne ressemble a rien.
func (v *Vie) attraper(curseurX, curseurY float64) {
	v.prise = priseNuque(v)
	v.passerA(Porte)
	v.Evenement = "attrape"
	v.jaugesAttrape()
}

// priseNuque renvoie l'ecart entre le curseur et le coin de la cellule quand on
// le tient : le curseur tombe au milieu de son corps dessine, un peu au-dessus
// des epaules. Les marges vides de la cellule sont retirees, sinon la prise
// serait decalee d'une planche a l'autre.
func priseNuque(v *Vie) [2]float64 {
	gauche, droite, haut, bas := 0.0, 0.0, 0.0, 0.0
	if a, _ := v.Animation(); a != nil {
		gauche, droite = float64(a.VideAGauche), float64(a.VideADroite)
		haut, bas = float64(a.VideEnHaut), float64(a.VideEnBas)
	}
	corpsH := v.hauteur - haut - bas
	return [2]float64{
		(gauche + v.largeur - droite) / 2, // milieu du corps, horizontalement
		haut + corpsH*0.2,                 // la nuque : il pend sous le curseur
	}
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

// lacher le repose la ou il est : il reprend sa vie sur place.
func (v *Vie) lacher() {
	v.vy = 0
	v.passerA(Repos)
	v.Evenement = "repose"
}
