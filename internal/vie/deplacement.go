package vie

import "math"

// Ce fichier regroupe le deplacement du singe : la marche lissee vers un
// point, l'orientation du sprite et les petits utilitaires de geometrie.

// avancerVers deplace le singe vers un point, avec une vitesse lissee : il
// accelere en partant et freine a l'approche, au lieu de partir et de
// s'arreter d'un coup.
func (v *Vie) avancerVers(x, y, dt, facteur float64) {
	cx, cy := v.Centre()
	dx, dy := x-cx, y-cy
	d := math.Hypot(dx, dy)
	if d < 1 {
		return
	}

	// demarche un peu sinueuse en promenade, plutot qu'une ligne parfaite
	if v.etat == Promenade {
		dev := 0.35 * math.Sin(v.depuis*2.2) * math.Min(1, d/90)
		a := math.Atan2(dy, dx) + dev
		dx, dy = math.Cos(a)*d, math.Sin(a)*d
	}

	cible := v.r.Vitesse * facteur
	const rayonFrein = 52.0
	if d < rayonFrein {
		cible *= math.Max(0.25, d/rayonFrein)
	}
	v.vitesse += (cible - v.vitesse) * math.Min(1, 5*dt)
	v.aAvance = true

	pas := v.vitesse * dt * 60 // vitesse exprimee par image a 60 Hz
	if pas > d {
		pas = d
	}
	v.X += pas * dx / d
	v.Y += pas * dy / d

	m := v.r.MargeBord
	v.X = clamp(v.X, m, v.ecranL-v.largeur-m)
	v.Y = clamp(v.Y, m, v.ecranH-v.hauteur-m)

	if v.p.Vue == "profil" {
		if math.Abs(dx) > 3 {
			v.direction = siPositif(dx, "droite", "gauche")
		}
		return
	}
	if math.Abs(dx) > math.Abs(dy) {
		if math.Abs(dx) > 3 {
			v.direction = siPositif(dx, "droite", "gauche")
		}
	} else {
		v.direction = siPositif(dy, "bas", "haut")
	}
}

// suivre fait marcher le singe derriere le curseur, en s'arretant a distance
// respectueuse.
func (v *Vie) suivre(dt, curseurX, curseurY float64) {
	cx, cy := v.Centre()
	d := math.Hypot(curseurX-cx, curseurY-cy)
	if v.enMarche {
		if d <= v.r.DistArret {
			v.enMarche = false
		}
	} else if d >= v.r.DistRepart {
		v.enMarche = true
	}
	if v.enMarche {
		v.avancerVers(curseurX, curseurY, dt, 1)
		v.animCourante = "marche"
	} else {
		v.animCourante = "repos"
		v.faceAu(curseurX, curseurY)
	}
	// l'ami ne bouge plus depuis un bon moment : il retourne vivre sa vie
	if v.dernierMvt > v.r.SeuilVieSeule {
		v.passerA(Repos)
		v.Evenement = "delaisse"
	}
}

// fuir eloigne le singe du curseur a toutes jambes, jusqu'a ce qu'il ait
// repris ses distances ou que la peur soit passee.
func (v *Vie) fuir(dt, curseurX, curseurY float64) {
	cx, cy := v.Centre()
	dx, dy := cx-curseurX, cy-curseurY
	d := math.Hypot(dx, dy)
	if d < 1 {
		dx, dy, d = 1, 0, 1
	}
	v.animCourante = "marche"
	v.avancerVers(cx+400*dx/d, cy+400*dy/d, dt, v.r.VitesseChasse)
	if d > v.r.DistFuite*2 || v.crainte <= 0 {
		v.passerA(Repos)
	}
}

// faceAu oriente le singe vers un point sans le deplacer.
func (v *Vie) faceAu(x, y float64) {
	cx, cy := v.Centre()
	dx, dy := x-cx, y-cy
	if v.p.Vue == "profil" || math.Abs(dx) > math.Abs(dy) {
		if math.Abs(dx) > 3 {
			v.direction = siPositif(dx, "droite", "gauche")
		}
		return
	}
	v.direction = siPositif(dy, "bas", "haut")
}

// dansLeCadre indique si un point est sur le corps du singe, avec une
// tolerance. On se limite aux pixels dessines (les marges vides de la cellule
// sont exclues) : cliquer a cote de lui ne l'atteint pas, ce qui laisse la
// place a une crotte collee a son dos sans qu'un clic dessus le blesse.
func (v *Vie) dansLeCadre(x, y float64) bool {
	const marge = 6
	gauche, droite, haut, bas := 0.0, 0.0, 0.0, 0.0
	if a, _ := v.Animation(); a != nil {
		gauche, droite = float64(a.VideAGauche), float64(a.VideADroite)
		haut, bas = float64(a.VideEnHaut), float64(a.VideEnBas)
	}
	return x >= v.X+gauche-marge && x <= v.X+v.largeur-droite+marge &&
		y >= v.Y+haut-marge && y <= v.Y+v.hauteur-bas+marge
}

func clamp(x, lo, hi float64) float64 {
	if hi < lo {
		return lo
	}
	return math.Min(math.Max(x, lo), hi)
}

func siPositif(x float64, oui, non string) string {
	if x > 0 {
		return oui
	}
	return non
}
