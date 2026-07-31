package vie

import "math"

// Ce fichier regroupe la sante du singe : les coeurs, les coups recus, la
// mort, puis la chute et la manipulation du cadavre.

// perdreCoeur retire un coeur : flash de degats, affichage des coeurs, et
// mort si c'etait le dernier. Renvoie true si le singe meurt.
func (v *Vie) perdreCoeur() bool {
	v.coeursRestants--
	v.depuisCoup = 0
	v.sansCoup = 0
	v.flash = 0.45
	if v.coeursRestants <= 0 {
		v.passerA(Mort)
		v.Evenement = "mort"
		return true
	}
	return false
}

// Coup encaisse un clic : le singe perd un coeur, accuse le choc, et meurt
// quand il n'en reste plus. Renvoie true si le coup a porte.
func (v *Vie) Coup() bool {
	if v.etat == Mort || v.etat == Cadavre || v.etat == Cache {
		return false
	}
	if v.perdreCoeur() {
		return true
	}

	// l'animation de coup alterne d'un cote a l'autre a chaque clic
	v.coteGauche = !v.coteGauche
	v.passerA(Touche)
	if v.coteGauche {
		v.direction = "gauche"
	} else {
		v.direction = "droite"
	}

	// petit recul a l'oppose du curseur
	cx, cy := v.Centre()
	dx, dy := cx-v.curseurX, cy-v.curseurY
	d := math.Hypot(dx, dy)
	if d < 1 {
		dx, dy, d = 0, -1, 1
	}
	v.reculX, v.reculY = 190*dx/d, 190*dy/d

	v.Evenement = "touche"
	return true
}

// encaisser fait vivre l'etat Touche : le recul s'amortit, puis le singe
// reprend sa vie une fois l'animation jouee.
func (v *Vie) encaisser(dt float64) {
	m := v.r.MargeBord
	v.X = clamp(v.X+v.reculX*dt, m, v.ecranL-v.largeur-m)
	v.Y = clamp(v.Y+v.reculY*dt, m, v.ecranH-v.hauteur-m)
	amorti := math.Pow(0.002, dt)
	v.reculX *= amorti
	v.reculY *= amorti

	a, ms := v.Animation()
	if a.Finie(ms) || v.depuis >= v.duree {
		v.passerA(Repos)
	}
}

// regenerer redonne un coeur apres une longue accalmie, pour que de vieux
// clics oublies ne laissent pas le singe au bord de la mort.
func (v *Vie) regenerer() {
	if v.r.RegenCoeur <= 0 || v.coeursRestants <= 0 || v.coeursRestants >= v.r.Coeurs {
		return
	}
	if v.sansCoup >= v.r.RegenCoeur {
		v.coeursRestants++
		v.sansCoup = 0
	}
}

// Coeurs renvoie le nombre de coeurs restants et le total.
func (v *Vie) Coeurs() (restants, total int) {
	return v.coeursRestants, v.r.Coeurs
}

// DepuisCoup renvoie le temps ecoule depuis le dernier coup recu, pour
// l'affichage des coeurs et de leur clignotement.
func (v *Vie) DepuisCoup() float64 { return v.depuisCoup }

// Flash renvoie le temps restant du flash de degats, zero hors des coups.
func (v *Vie) Flash() float64 { return math.Max(0, v.flash) }

// AnimationMorteFinie indique si l'agonie est terminee, moment ou l'appelant
// declenche l'explosion de coeur.
func (v *Vie) AnimationMorteFinie() bool {
	if v.etat != Mort {
		return false
	}
	a, ms := v.Animation()
	if a.Boucle {
		return v.depuis >= 0.6 // pas d'animation de mort : court delai
	}
	return a.Finie(ms)
}

// Retomber decide de ce qui suit l'explosion de coeur : soit le singe
// s'eclipse pour un moment, soit son corps tombe sur la barre des taches.
func (v *Vie) Retomber() {
	if v.r.CacheApresClic > 0 {
		v.passerA(Cache)
		return
	}
	v.passerA(Cadavre)
}

// Ranimer remet le singe sur ses pattes, coeurs au complet — mais l'episode
// le laisse craintif un moment.
func (v *Vie) Ranimer() {
	v.deplace = false
	v.coeursRestants = v.r.Coeurs
	v.crainte = v.r.DureeCrainte
	v.passerA(Repos)
	v.Evenement = "retour"
}

// avancerCadavre gere le corps une fois mort : deplacable a la souris, il
// tombe des qu'on le lache, jusqu'a se poser sur la barre des taches. Bien
// secoue pendant qu'on le tient, il revient a la vie.
func (v *Vie) avancerCadavre(dt, cx, cy float64, boutonEnfonce, frontDescendant bool) {
	saisi := frontDescendant && v.dansLeCadre(cx, cy)
	v.deplacerCadavre(cx, cy, boutonEnfonce, frontDescendant)
	if saisi {
		v.dragXPrec, v.dragYPrec = cx, cy
		v.dragVxPrec, v.dragVyPrec = 0, 0
		v.secousses = 0
	}
	if v.deplace {
		v.vy = 0
		v.depuis = 0
		v.pendre()
		v.secouer(dt, cx, cy)
		return
	}
	v.tomber(dt)
	v.poseCadavre()
	if v.r.Resurrection > 0 && v.depuis >= v.r.Resurrection && v.Y >= v.solY() {
		v.Ranimer()
	}
}

// pendre fige le corps dans sa pose morte pendant qu'on le tient.
func (v *Vie) pendre() {
	v.animCourante = choisir(v.p, "meurt", "repos")
	v.animFigee = true
}

// poseCadavre choisit l'apparence du corps : l'animation de chute tant qu'il
// est en l'air, la derniere image de l'agonie une fois pose.
func (v *Vie) poseCadavre() {
	if (v.vy != 0 || v.Y < v.solY()-0.5) && v.p.Action("tombe") {
		if v.animFigee { // il vient de partir en chute
			v.animFigee = false
			v.animMS = 0
		}
		v.animCourante = "tombe"
		return
	}
	if !v.animFigee { // il vient de se poser
		v.pendre()
		// les deux animations ne laissent pas le meme vide sous elles : on
		// recale le corps sur le sol de la pose morte
		v.Y = v.solY()
	}
}

// secouer mesure les demi-tours brusques du cadavre traine a la souris ; bien
// secoue, le singe revient a la vie.
func (v *Vie) secouer(dt, cx, cy float64) {
	vx := (cx - v.dragXPrec) / dt
	vy := (cy - v.dragYPrec) / dt
	v.dragXPrec, v.dragYPrec = cx, cy
	v.secousses = math.Max(0, v.secousses-1.2*dt)
	if math.Hypot(vx, vy) > 500 {
		if vx*v.dragVxPrec+vy*v.dragVyPrec < 0 { // demi-tour franc
			v.secousses++
		}
		v.dragVxPrec, v.dragVyPrec = vx, vy
	}
	if v.secousses < 3 {
		return
	}
	v.secousses = 0
	v.Ranimer()
	v.Evenement = "ranime"
}

// deplacerCadavre laisse attraper le cadavre a la souris et le suivre.
func (v *Vie) deplacerCadavre(cx, cy float64, boutonEnfonce, frontDescendant bool) {
	if frontDescendant && v.dansLeCadre(cx, cy) {
		v.deplace = true
		v.prise = [2]float64{cx - v.X, cy - v.Y}
	}
	if !boutonEnfonce {
		v.deplace = false
		return
	}
	if !v.deplace {
		return
	}
	m := v.r.MargeBord
	v.X = clamp(cx-v.prise[0], m, v.ecranL-v.largeur-m)
	v.Y = clamp(cy-v.prise[1], m, v.solY())
}

// tomber applique la gravite au cadavre jusqu'au sommet de la barre des
// taches, avec un petit rebond a l'atterrissage.
func (v *Vie) tomber(dt float64) {
	const gravite = 2800 // pixels/s2
	sol := v.solY()
	if v.Y >= sol && math.Abs(v.vy) < 1 {
		v.Y = sol
		return
	}
	v.vy += gravite * dt
	v.Y += v.vy * dt
	if v.Y >= sol {
		v.Y = sol
		if v.vy > 300 {
			v.vy = -0.28 * v.vy
		} else {
			v.vy = 0
		}
	}
}

// solY est l'ordonnee du sprite quand le corps repose sur la barre des taches,
// en tenant compte du vide que l'animation laisse sous lui.
func (v *Vie) solY() float64 {
	a, _ := v.Animation()
	return v.basEcran - v.hauteur + float64(a.VideEnBas)
}
