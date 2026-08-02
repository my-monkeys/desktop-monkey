package main

import (
	"image"
	"log"
	"math"

	"github.com/my-monkeys/desktop-monkey/internal/calque"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
)

// Le singe laisse parfois une crotte sur le bureau. Chaque crotte est une
// petite fenetre en couche autonome, posee la ou il l'a pondue : la fenetre du
// singe est petite et le suit, elle ne peut pas dessiner ailleurs a l'ecran.
//
// Une crotte joue son apparition, fume en attendant, et explose au clic.

const (
	crotteMax  = 6                // nombre maximum de crottes a l'ecran
	crotteClic = 4                // tolerance de clic autour du sprite, en pixels
	nomCrotte  = "CrotteDeBureau" // classe de fenetre partagee par les crottes
)

// phases de la vie d'une crotte
const (
	crApparait = iota
	crRepos
	crExplose
	crVole // le singe l'a lancee : elle file (vers la souris, ou hors ecran)
	crFinie
)

// modes de lancer
const (
	versBord   = iota // jetee hors de l'ecran, coupee au bord
	versSouris        // lancee sur le curseur, elle explose a l'arrivee
)

type crotte struct {
	solX, solY int     // point de sol sur le bureau (base de la crotte)
	fenX, fenY int     // coin haut gauche de sa fenetre (rendu)
	fx, fy     float64 // meme coin, en flottant (mouvement fluide en vol/porte)
	phase      int
	ms         float64 // avancement dans l'animation de la phase
	age        float64 // secondes depuis qu'elle est posee
	cal        *calque.Calque

	porte  bool    // le singe la tient
	vx, vy float64 // vitesse en vol (px/s)
	mode   int     // versBord ou versSouris
}

// chargerCrottes prepare la planche des crottes ; son absence desactive
// simplement la mecanique.
func (s *scene) chargerCrottes() {
	p, err := planche.Charger(ressources.Fichiers, "assets/crotte.json", echelleAff)
	if err != nil {
		log.Printf("crottes indisponibles : %v", err)
		return
	}
	s.pc = p
	// base du tas dans la cellule : hauteur moins les rangees vides du dessous
	a := p.Obtenir("repos", "")
	s.crotteSol = p.Hauteur - a.VideEnBas
}

// pondre depose une crotte au point de sol donne, si la planche existe, si on
// peut ouvrir des fenetres, et si le plafond n'est pas atteint.
func (s *scene) pondre(x, y float64) {
	if s.pc == nil || !s.fenetresOK || len(s.crottes) >= crotteMax {
		return
	}
	x = s.ecarterCrotte(x)
	cal, err := calque.Ouvrir(nomCrotte, s.pc.Largeur, s.pc.Hauteur)
	if err != nil {
		log.Printf("fenetre de crotte : %v", err)
		return
	}
	c := &crotte{
		solX:  int(x),
		solY:  int(y),
		fenX:  int(x) - s.pc.Largeur/2,
		fenY:  int(y) - s.crotteSol,
		phase: crApparait,
		cal:   cal,
	}
	c.fx, c.fy = float64(c.fenX), float64(c.fenY)
	s.crottes = append(s.crottes, c)
}

// ecarterCrotte decale l'abscisse d'une nouvelle crotte pour qu'elle ne se
// pose pas sur une crotte deja la : le singe defeque souvent au meme endroit
// s'il ne bouge pas, et les tas se superposaient. On s'ecarte de part et
// d'autre, par pas d'une largeur de tas, jusqu'a trouver un creneau libre.
func (s *scene) ecarterCrotte(x float64) float64 {
	ecart := float64(s.pc.Largeur)
	libre := func(cx float64) bool {
		for _, c := range s.crottes {
			if math.Abs(float64(c.solX)-cx) < ecart {
				return false
			}
		}
		return true
	}
	if libre(x) {
		return x
	}
	for d := ecart; d <= float64(s.ecranL); d += ecart {
		if cand := x + d; cand <= float64(s.ecranL) && libre(cand) {
			return cand
		}
		if cand := x - d; cand >= 0 && libre(cand) {
			return cand
		}
	}
	return x
}

// gererCrottes fait vivre les crottes : avancement des animations, clic qui
// declenche l'explosion, et retrait de celles qui ont fini d'exploser.
func (s *scene) gererCrottes(dt float64, curseurX, curseurY int, bouton bool) {
	frontDescendant := bouton && !s.boutonCrotteAvant
	s.boutonCrotteAvant = bouton

	// le clic la plus recente (au-dessus) explose en priorite ; une crotte
	// portee ou lancee n'est pas cliquable
	if frontDescendant {
		for i := len(s.crottes) - 1; i >= 0; i-- {
			c := s.crottes[i]
			if (c.phase == crApparait || c.phase == crRepos) && !c.porte &&
				s.crotteSousCurseur(c, curseurX, curseurY) {
				c.phase, c.ms = crExplose, 0
				break
			}
		}
	}

	reste := s.crottes[:0]
	for _, c := range s.crottes {
		c.ms += dt * 1000
		if c.phase == crRepos && !c.porte {
			c.age += dt
		}
		switch c.phase {
		case crApparait:
			if a := s.pc.Obtenir("apparait", ""); a.Finie(int(c.ms)) {
				c.phase, c.ms = crRepos, 0
			}
		case crVole:
			s.volerCrotte(c, dt, curseurX, curseurY)
		case crExplose:
			if a := s.pc.Obtenir("explose", ""); a.Finie(int(c.ms)) {
				c.phase = crFinie
			}
		}
		if c.phase == crFinie {
			c.cal.Fermer()
			continue
		}
		reste = append(reste, c)
	}
	s.crottes = reste
}

// volerCrotte fait avancer une crotte lancee : vers la souris (elle explose a
// l'arrivee), ou hors de l'ecran par un bord (Windows rogne la fenetre au bord,
// elle semble se faire couper en sortant).
func (s *scene) volerCrotte(c *crotte, dt float64, curseurX, curseurY int) {
	c.fx += c.vx * dt
	c.fy += c.vy * dt
	c.fenX, c.fenY = int(c.fx), int(c.fy)

	if c.mode == versSouris {
		cx := c.fx + float64(s.pc.Largeur)/2
		cy := c.fy + float64(s.crotteSol)
		if math.Hypot(cx-float64(curseurX), cy-float64(curseurY)) < 20 {
			c.phase, c.ms = crExplose, 0
		}
		return
	}
	// versBord : disparait une fois entierement sortie d'un cote
	if c.fx <= -float64(s.pc.Largeur) || c.fx >= float64(s.ecranL) {
		c.phase = crFinie
	}
}

// crotteSousCurseur indique si le curseur touche un pixel opaque de la crotte.
func (s *scene) crotteSousCurseur(c *crotte, curseurX, curseurY int) bool {
	lx, ly := curseurX-c.fenX, curseurY-c.fenY
	if lx < -crotteClic || ly < -crotteClic ||
		lx >= s.pc.Largeur+crotteClic || ly >= s.pc.Hauteur+crotteClic {
		return false
	}
	img := s.crotteImage(c)
	if img == nil {
		return false
	}
	// on cherche un pixel dessine dans un petit rayon autour du curseur, ce qui
	// donne la tolerance de clic sans exiger de viser au pixel pres
	for dy := -crotteClic; dy <= crotteClic; dy++ {
		for dx := -crotteClic; dx <= crotteClic; dx++ {
			x, y := lx+dx, ly+dy
			if x < 0 || y < 0 || x >= s.pc.Largeur || y >= s.pc.Hauteur {
				continue
			}
			if img.Pix[img.PixOffset(x, y)+3] > 40 {
				return true
			}
		}
	}
	return false
}

// crotteImage renvoie l'image a afficher pour la phase courante.
func (s *scene) crotteImage(c *crotte) *image.RGBA {
	switch c.phase {
	case crApparait:
		return s.pc.Obtenir("apparait", "").Image(int(c.ms))
	case crExplose:
		return s.pc.Obtenir("explose", "").Image(int(c.ms))
	default:
		return s.pc.Obtenir("repos", "").Image(int(c.ms))
	}
}

// afficherCrottes peint chaque crotte dans sa fenetre. A appeler dans la boucle
// principale, apres l'affichage du singe. Une crotte portee est dessinee dans
// la scene du singe (voir composer) : sa fenetre affiche du transparent, pour
// que la bulle et les coeurs restent au-dessus d'elle.
func (s *scene) afficherCrottes() {
	for _, c := range s.crottes {
		img := s.crotteImage(c)
		if img == nil {
			continue
		}
		if c.porte {
			if s.videCrotte == nil {
				s.videCrotte = image.NewRGBA(image.Rect(0, 0, s.pc.Largeur, s.pc.Hauteur))
			}
			img = s.videCrotte
		}
		if err := c.cal.Afficher(img, c.fenX, c.fenY); err != nil {
			log.Printf("affichage crotte : %v", err)
		}
	}
}

// fermerCrottes libere toutes les fenetres de crottes.
func (s *scene) fermerCrottes() {
	for _, c := range s.crottes {
		c.cal.Fermer()
	}
	s.crottes = nil
}
