package main

import (
	"math"

	"github.com/my-monkeys/desktop-monkey/internal/vie"
)

// Quand une crotte traine depuis un moment, le singe va parfois s'en occuper :
// il la rejoint, la ramasse, et soit l'emporte au bord de l'ecran pour la jeter
// dehors, soit la lance sur le curseur. Toute cette choregraphie vit dans la
// scene ; le singe n'est qu'un pantin en etat Corvee (voir internal/vie).

// phases de la corvee
const (
	corVaChercher = iota // il marche vers la crotte
	corPorteBord         // il la porte jusqu'au bord
	corVise              // il fait le geste de lancer
)

const (
	ageJetCrotte = 4.0 // secondes avant qu'une crotte devienne bonne a jeter
	cooldownJet  = 9.0 // secondes de repos entre deux corvees
)

// gererCorvee fait avancer la besogne en cours, ou en declenche une.
func (s *scene) gererCorvee(dt float64, cx, cy int) {
	if s.pc == nil || !s.fenetresOK {
		return
	}
	if s.corCooldown > 0 {
		s.corCooldown -= dt
	}
	if s.corCrotte == nil {
		s.peutEtreLancerCorvee(dt, cx, cy)
		return
	}

	c := s.corCrotte
	// la crotte a ete cliquee, ou est deja partie : on laisse tomber
	if !s.contientCrotte(c) || c.phase == crExplose || c.phase == crVole {
		s.finirCorvee()
		return
	}

	switch s.corPhase {
	case corVaChercher:
		if s.v.Arrive() {
			c.porte = true
			c.phase, c.ms = crRepos, 0
			if s.corMode == versBord {
				_, cyy := s.v.Centre()
				s.v.EnvoyerVers(s.bordVise(c), cyy)
				s.corPhase = corPorteBord
			} else {
				s.v.Viser(float64(cx))
				s.corTimer, s.corPhase = 0.35, corVise
			}
		}
	case corPorteBord:
		if s.v.Arrive() {
			s.v.Viser(s.bordVise(c))
			s.corTimer, s.corPhase = 0.35, corVise
		}
	case corVise:
		s.corTimer -= dt
		if s.corTimer <= 0 {
			s.lancerCrotte(c, cx, cy)
			s.v.Liberer()
			s.finirCorvee()
			s.corCooldown = cooldownJet
		}
	}

	// pendant tout le transport, la crotte suit les mains du singe
	if c.porte {
		hx, hy := s.v.PositionMain()
		c.fx = hx - float64(s.pc.Largeur)/2
		c.fy = hy - float64(s.crotteSol)
		c.fenX, c.fenY = int(c.fx), int(c.fy)
	}
}

// peutEtreLancerCorvee demarre une besogne s'il est desoeuvre, qu'une crotte
// traine, et que le sort le decide.
func (s *scene) peutEtreLancerCorvee(dt float64, cx, cy int) {
	if s.corCooldown > 0 {
		return
	}
	if et := s.v.Etat(); et != vie.Repos && et != vie.Promenade {
		return
	}
	c := s.crotteAJeter()
	if c == nil || s.alea.Float64() > s.r.ChanceJetCrotte*dt {
		return
	}
	s.corCrotte, s.corPhase = c, corVaChercher
	// il la jette dehors, ou la lance sur la souris
	s.corMode = versBord
	if s.r.JetMode == versSouris || (s.r.JetMode < 0 && s.alea.Float64() < 0.5) {
		s.corMode = versSouris
	}
	s.v.EnvoyerVers(float64(c.solX), float64(c.solY)-float64(s.pc.Hauteur)*0.25)
}

// lancerCrotte donne son elan a la crotte : droit sur le curseur, ou hors de
// l'ecran par le bord le plus proche (avec un petit arc).
func (s *scene) lancerCrotte(c *crotte, cx, cy int) {
	c.phase, c.ms, c.mode = crVole, 0, s.corMode
	hx, hy := s.v.PositionMain()
	c.fx = hx - float64(s.pc.Largeur)/2
	c.fy = hy - float64(s.crotteSol)

	if s.corMode == versSouris {
		px := c.fx + float64(s.pc.Largeur)/2
		py := c.fy + float64(s.crotteSol)
		dx, dy := float64(cx)-px, float64(cy)-py
		d := math.Hypot(dx, dy)
		if d < 1 {
			d = 1
		}
		const vitesse = 950.0
		c.vx, c.vy = vitesse*dx/d, vitesse*dy/d
		return
	}
	const vitesse = 750.0
	if float64(c.solX) < float64(s.ecranL)/2 {
		c.vx = -vitesse
	} else {
		c.vx = vitesse
	}
	c.vy = -140 // un rien vers le haut, pour l'allure
}

// bordVise renvoie l'abscisse du bord vers lequel jeter la crotte.
func (s *scene) bordVise(c *crotte) float64 {
	if float64(c.solX) < float64(s.ecranL)/2 {
		return 0
	}
	return float64(s.ecranL)
}

// crotteAJeter renvoie une crotte posee depuis assez longtemps, libre.
func (s *scene) crotteAJeter() *crotte {
	for _, c := range s.crottes {
		if c.phase == crRepos && !c.porte && c.age > ageJetCrotte {
			return c
		}
	}
	return nil
}

func (s *scene) contientCrotte(c *crotte) bool {
	for _, x := range s.crottes {
		if x == c {
			return true
		}
	}
	return false
}

func (s *scene) finirCorvee() {
	if s.corCrotte != nil {
		s.corCrotte.porte = false
	}
	s.corCrotte, s.corPhase = nil, corVaChercher
}
