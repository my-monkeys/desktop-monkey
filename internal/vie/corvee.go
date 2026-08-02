package vie

import "math"

// L'etat Corvee est le seul que le singe ne decide pas lui-meme : la scene l'y
// met pour lui faire faire une besogne autour des crottes (aller en chercher
// une, la porter au bord, la lancer). La scene fournit la cible et lit son
// arrivee ; le singe se contente de marcher jusque-la, puis d'attendre.

// EnvoyerVers place le singe en Corvee et lui donne un point a rejoindre (centre
// du sprite). A rappeler pour changer de cible en cours de besogne.
func (v *Vie) EnvoyerVers(x, y float64) {
	if v.etat != Corvee {
		v.passerA(Corvee)
	}
	v.dirX, v.dirY = x, y
	v.arrive = false
}

// EnCorvee indique si le singe est aux ordres de la scene.
func (v *Vie) EnCorvee() bool { return v.etat == Corvee }

// Arrive indique s'il a rejoint le dernier point demande.
func (v *Vie) Arrive() bool { return v.etat == Corvee && v.arrive }

// Viser le tourne vers un point et declenche le geste de lancer (l'animation
// d'attaque, le temps d'un swing).
func (v *Vie) Viser(x float64) {
	v.faceAu(x, v.Y)
	v.viseMS = 0.35
	v.animMS = 0
}

// Liberer le rend a sa vie autonome.
func (v *Vie) Liberer() {
	v.viseMS = 0
	v.passerA(Repos)
}

// PositionMain renvoie ou une crotte portee doit etre dessinee (les mains du
// singe, un peu au-dessus de son centre).
func (v *Vie) PositionMain() (x, y float64) {
	return v.X + v.largeur/2, v.Y + v.hauteur*0.35
}

func (v *Vie) corvee(dt float64) {
	if v.viseMS > 0 {
		// le temps du geste de lancer : il reste sur place, animation d'attaque
		v.viseMS -= dt
		v.animCourante = choisir(v.p, "lance", "repos")
		return
	}
	cx, cy := v.Centre()
	if math.Hypot(v.dirX-cx, v.dirY-cy) < 6 {
		v.arrive = true
		v.animCourante = "repos"
		return
	}
	v.animCourante = "marche"
	v.avancerVers(v.dirX, v.dirY, dt, 1)
}
