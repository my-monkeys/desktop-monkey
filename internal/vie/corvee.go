package vie

import "math"

// L'etat Corvee est le seul que le singe ne decide pas lui-meme : la scene l'y
// met pour lui faire faire une besogne autour des crottes (aller en chercher
// une, la porter au bord, la lancer). La scene fournit la cible et lit son
// arrivee ; le singe se contente de marcher jusque-la, puis d'attendre.

// EnvoyerVers place le singe en Corvee et lui donne un point a rejoindre (centre
// du sprite). A rappeler pour changer de cible en cours de besogne. La cible
// peut etre hors de portee — le bord de l'ecran, par exemple : il ira alors
// aussi loin qu'il peut, et se declarera arrive (voir corvee).
func (v *Vie) EnvoyerVers(x, y float64) {
	if v.etat != Corvee {
		v.passerA(Corvee)
	}
	v.dirX, v.dirY = x, y
	v.arrive = false
	v.pietine = 0
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
	avant := math.Hypot(v.dirX-cx, v.dirY-cy)
	if avant < 6 {
		v.arrive = true
		v.animCourante = "repos"
		v.pietine = 0
		return
	}

	v.animCourante = "marche"
	v.avancerVers(v.dirX, v.dirY, dt, 1)

	// La scene vise parfois un point que la marge de bord lui interdit
	// d'atteindre : le bord de l'ecran, ou un curseur dans un coin. Sans ce
	// garde-fou il pietinait la, sa crotte a la main, sans jamais la lancer.
	//
	// On regarde le terrain gagne vers la cible, et non le deplacement : plaque
	// contre un mur, il glisse encore le long sans jamais s'en rapprocher.
	ncx, ncy := v.Centre()
	if avant-math.Hypot(v.dirX-ncx, v.dirY-ncy) < gainMinimal {
		if v.pietine += dt; v.pietine >= dureePietine {
			v.arrive = true
			v.animCourante = "repos"
		}
		return
	}
	v.pietine = 0
}

const (
	// gainMinimal est le terrain a gagner par image pour qu'il avance vraiment.
	// Sa marche en couvre une cinquantaine de fois plus.
	gainMinimal = 0.05
	// dureePietine est le temps sans progres au bout duquel il tient sa cible
	// pour atteinte.
	dureePietine = 0.4
)
