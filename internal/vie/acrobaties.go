package vie

import "math"

// Ce fichier contient les numeros du singe : la chasse au curseur, le vol de
// la fleche, et l'escalade des bords de l'ecran.

// chasser fait courir le singe apres le curseur en lui jetant des bananes. De
// pres, le jet bouscule la fleche, et peut tourner au vol pur et simple.
func (v *Vie) chasser(dt, curseurX, curseurY float64) {
	cx, cy := v.Centre()
	d := math.Hypot(curseurX-cx, curseurY-cy)

	switch {
	case v.lanceMS > 0:
		// le geste est parti : il reste plante le temps de le finir
		v.avancerLancer(dt, curseurX, curseurY)
		if v.etat != Chasse {
			return // le jet a tourne au vol de curseur
		}
	case d > v.r.PorteeAttaque:
		v.animCourante = "marche"
		v.avancerVers(curseurX, curseurY, dt, v.r.VitesseChasse)
	default:
		// a bout de bras il ne recule pas, il attend le prochain jet
		v.animCourante = choisir(v.p, "repos", "marche")
		v.faceAu(curseurX, curseurY)
	}

	// la banane part a intervalles reguliers, quelle que soit la distance :
	// c'est ce qui lui permet d'attaquer d'un bout a l'autre de l'ecran
	if v.lanceMS <= 0 {
		v.depuisLancer += dt
		if v.depuisLancer >= v.r.EntreLancers {
			v.commencerLancer(curseurX, curseurY)
		}
	}
	// la proie ne bouge plus, ou la chasse s'eternise : il s'en desinteresse
	if v.dernierMvt > v.r.AbandonChasse || v.depuis >= v.duree {
		v.passerA(Repos)
		v.Evenement = "abandon"
	}
}

// porterCoup conclut un swing d'attaque. Le premier coup d'une chasse peut
// tourner au vol de la fleche ; les autres la repoussent, comme un vrai coup.
func (v *Vie) porterCoup(curseurX, curseurY float64) {
	if !v.volTente {
		v.volTente = true
		if v.alea.Float64() < v.r.ChanceVol {
			v.passerA(Vol)
			v.Evenement = "vole"
			return
		}
	}
	cx, cy := v.Centre()
	dx, dy := curseurX-cx, curseurY-cy
	d := math.Hypot(dx, dy)
	if d < 1 {
		dx, dy, d = 1, 0, 1
	}
	force := 45 + v.alea.Float64()*25
	v.pousseX = clamp(curseurX+force*dx/d, 0, v.ecranL-1)
	v.pousseY = clamp(curseurY+force*dy/d, 0, v.ecranH-1)
	v.pousseEnAttente = true
	v.attenduX, v.attenduY = v.pousseX, v.pousseY
	v.attenduValide = true
}

// PrendrePoussee renvoie ou envoyer le curseur repousse par un coup de patte,
// et consomme la poussee. C'est l'appelant qui deplace le curseur systeme.
func (v *Vie) PrendrePoussee() (x, y float64, ok bool) {
	if !v.pousseEnAttente {
		return 0, 0, false
	}
	v.pousseEnAttente = false
	return v.pousseX, v.pousseY, true
}

// enfuir traine le curseur vole vers l'autre bout de l'ecran, jusqu'a ce que
// la proie se debatte ou qu'il arrive a destination.
func (v *Vie) enfuir(dt, curseurX, curseurY float64) {
	// un grand ecart entre la position imposee et la position lue signifie
	// que l'utilisateur se debat : il lache tout de suite
	if v.depuis > 0.5 && math.Hypot(curseurX-v.volX, curseurY-v.volY) > 30 {
		v.lacherCurseur()
		return
	}
	v.animCourante = "marche"
	v.avancerVers(v.fuiteX+v.largeur/2, v.fuiteY+v.hauteur/2, dt, v.r.VitesseChasse)
	v.volX, v.volY = v.Centre()
	if v.depuis >= v.duree ||
		(math.Abs(v.X-v.fuiteX) < 4 && math.Abs(v.Y-v.fuiteY) < 4) {
		v.lacherCurseur()
	}
}

// lacherCurseur rend le curseur et retourne a la vie normale.
func (v *Vie) lacherCurseur() {
	// sans ceci, le curseur pose a cote de lui passerait pour une proie
	// fraiche et il repartirait aussitot en chasse
	v.dernierMvt = 3
	v.passerA(Repos)
	v.Evenement = "lache"
}

// TientCurseur indique si le singe tient le curseur, et ou le poser a l'ecran.
// C'est l'appelant qui deplace reellement le curseur systeme.
func (v *Vie) TientCurseur() (x, y float64, ok bool) {
	if v.etat != Vol {
		return 0, 0, false
	}
	return v.volX, v.volY, true
}

// Les phases de l'escalade.
const (
	grimpeApproche = iota // rejoindre le bord vertical le plus proche
	grimpeMonte           // escalader le long du bord
	grimpePause           // souffler la-haut
	grimpeChute           // se laisser tomber
)

// grimper enchaine les phases de l'escalade : rejoindre le bord, monter,
// souffler en haut, puis se laisser tomber jusqu'en bas de l'ecran.
func (v *Vie) grimper(dt float64) {
	switch v.phaseGrimpe {
	case grimpeApproche:
		v.animCourante = "marche"
		v.avancerVers(v.cibleX+v.largeur/2, v.Y+v.hauteur/2, dt, 1)
		if math.Abs(v.X-v.cibleX) < 2 {
			v.phaseGrimpe = grimpeMonte
			v.animMS = 0
		}

	case grimpeMonte:
		v.animCourante = choisir(v.p, "grimpe", "marche")
		v.faceAuMur()
		v.Y = math.Max(v.Y-v.r.Vitesse*0.55*dt*60, v.cibleY)
		if v.Y <= v.cibleY+0.5 {
			v.phaseGrimpe = grimpePause
			// arrive, il s'accroche sans bouger : l'animation d'escalade ne
			// doit pas continuer a mouliner sur place
			v.animFigee = true
		}

	case grimpePause:
		v.animCourante = choisir(v.p, "grimpe", "repos")
		v.pauseGrimpe -= dt
		if v.pauseGrimpe <= 0 {
			v.animFigee = false
			v.phaseGrimpe = grimpeChute
			v.vy = 0
			v.animMS = 0
			v.Evenement = "tombe"
		}

	case grimpeChute:
		v.animCourante = choisir(v.p, "tombe", "marche")
		v.tomber(dt)
		if v.Y >= v.solY()-0.5 && v.vy == 0 {
			// la reception est brutale : elle coute un coeur
			if !v.perdreCoeur() {
				v.reculX, v.reculY = 0, 0
				v.passerA(Touche)
				v.Evenement = "aie_chute"
			}
		}
	}
	if v.etat == Grimpe && v.depuis >= v.duree {
		v.passerA(Repos)
	}
}

// faceAuMur tourne le singe vers le bord qu'il escalade.
func (v *Vie) faceAuMur() {
	if v.cibleX < v.ecranL/2 {
		v.direction = "gauche"
	} else {
		v.direction = "droite"
	}
}

// sautiller fait faire au singe de vrais bonds : il avance en decrivant de
// petites paraboles sur sa ligne de depart, et rebondit sur les bords.
func (v *Vie) sautiller(dt float64) {
	const gravite = 2200.0
	v.animCourante = choisir(v.p, "saute", "repos")
	v.direction = siPositif(v.jeuDir, "droite", "gauche")

	m := v.r.MargeBord
	v.X += v.jeuDir * v.r.Vitesse * 0.8 * 60 * dt
	if v.X <= m || v.X >= v.ecranL-v.largeur-m {
		v.jeuDir = -v.jeuDir
	}
	v.X = clamp(v.X, m, v.ecranL-v.largeur-m)

	v.Y += v.vy * dt
	v.vy += gravite * dt
	if v.Y >= v.jeuSolY {
		v.Y = v.jeuSolY
		// il ne s'arrete qu'entre deux bonds, jamais en l'air
		if v.depuis >= v.duree {
			v.choisirSuite()
			return
		}
		v.vy = -(260 + v.alea.Float64()*90)
	}
}
