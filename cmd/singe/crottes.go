package main

import (
	"image"
	"log"

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
	crFinie
)

type crotte struct {
	solX, solY int // point de sol sur le bureau (base de la crotte)
	fenX, fenY int // coin haut gauche de sa fenetre
	phase      int
	ms         float64 // avancement dans l'animation de la phase
	cal        *calque.Calque
}

// chargerCrottes prepare la planche des crottes ; son absence desactive
// simplement la mecanique.
func (s *scene) chargerCrottes() {
	p, err := planche.Charger(ressources.Fichiers, "assets/crotte.json")
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
	s.crottes = append(s.crottes, c)
}

// gererCrottes fait vivre les crottes : avancement des animations, clic qui
// declenche l'explosion, et retrait de celles qui ont fini d'exploser.
func (s *scene) gererCrottes(dt float64, curseurX, curseurY int, bouton bool) {
	frontDescendant := bouton && !s.boutonCrotteAvant
	s.boutonCrotteAvant = bouton

	// le clic la plus recente (au-dessus) explose en priorite
	if frontDescendant {
		for i := len(s.crottes) - 1; i >= 0; i-- {
			c := s.crottes[i]
			if (c.phase == crApparait || c.phase == crRepos) && s.crotteSousCurseur(c, curseurX, curseurY) {
				c.phase, c.ms = crExplose, 0
				break
			}
		}
	}

	reste := s.crottes[:0]
	for _, c := range s.crottes {
		c.ms += dt * 1000
		switch c.phase {
		case crApparait:
			if a := s.pc.Obtenir("apparait", ""); a.Finie(int(c.ms)) {
				c.phase, c.ms = crRepos, 0
			}
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
// principale, apres l'affichage du singe.
func (s *scene) afficherCrottes() {
	for _, c := range s.crottes {
		if img := s.crotteImage(c); img != nil {
			if err := c.cal.Afficher(img, c.fenX, c.fenY); err != nil {
				log.Printf("affichage crotte : %v", err)
			}
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
