package main

// Les bananes qu'il jette.
//
// Le sprite du lancer montre deja une banane s'eloigner de sa main, mais elle
// est prisonniere de la cellule : elle ne peut pas depasser les quelques
// pixels du dessin. On la decoupe donc du sprite au demarrage — le plus gros
// amas jaune de la derniere image du geste — pour en faire une vraie banane,
// qui prend le relais exactement la ou celle du dessin s'arrete, part en
// cloche vers le curseur et finit par tomber hors de l'ecran.

import (
	"image"
	"image/color"
	"log"
	"math"

	"github.com/my-monkeys/desktop-monkey/internal/calque"
	"github.com/my-monkeys/desktop-monkey/internal/vie"
)

const (
	nomBanane     = "BananeDeBureau" // classe de fenetre partagee par les bananes
	bananeMax     = 8                // plafond de bananes en vol
	graviteBanane = 2000             // pixels/s2, comme la chute du singe
	volMin        = 0.35             // duree de vol minimale jusqu'a la cible
	volMax        = 0.9              // et maximale : plus c'est loin, plus ca plane
	volDistance   = 1400             // distance au-dela de laquelle le vol dure volMax
)

// banane est un fruit en vol : une petite fenetre qui suit une parabole.
type banane struct {
	fx, fy float64 // coin haut gauche de sa fenetre
	vx, vy float64 // vitesse, en pixels par seconde
	cal    *calque.Calque
}

// departBanane retient, pour un geste et un sens donnes, ou se trouve la
// banane dans la derniere image : c'est de la que part la vraie.
type departBanane struct{ dx, dy float64 }

// chargerBananes decoupe la banane du sprite et note son point de depart pour
// chaque geste. Une planche sans banane (le singe RPG) desactive simplement la
// mecanique.
func (s *scene) chargerBananes() {
	marge := int(math.Round(echelleAff))
	if marge < 1 {
		marge = 1
	}
	s.bananeDepart = map[string]departBanane{}
	for _, action := range []string{"lance", "lance_saut"} {
		for _, sens := range []string{"droite", "gauche"} {
			a := s.pSinge.Obtenir(action, sens)
			if a == nil || len(a.Images) == 0 {
				continue
			}
			derniere := a.Images[len(a.Images)-1]
			boite, ok := amasJaune(derniere)
			if !ok {
				continue
			}
			boite = boite.Inset(-marge).Intersect(derniere.Bounds())
			if s.bananeImg == nil {
				s.bananeImg = decouper(derniere, boite)
			}
			// le depart est donne dans le repere de la cellule
			s.bananeDepart[action+"_"+sens] = departBanane{
				dx: float64(boite.Min.X - derniere.Bounds().Min.X),
				dy: float64(boite.Min.Y - derniere.Bounds().Min.Y),
			}
		}
	}
	if s.bananeImg == nil {
		log.Print("pas de banane dans cette planche : le singe se battra a mains nues")
	}
}

// nouvelleBanane calcule le depart d'un jet : d'ou part le fruit, et a quelle
// vitesse pour qu'il passe par le curseur. Au-dela, la gravite l'emporte et il
// finit par sortir de l'ecran par le bas.
func (s *scene) nouvelleBanane(l vie.Lancer, celluleX, celluleY float64) (banane, bool) {
	depart, ok := s.bananeDepart[l.Action+"_"+l.Direction]
	if !ok || s.bananeImg == nil {
		return banane{}, false
	}
	b := s.bananeImg.Bounds()
	x := celluleX + depart.dx
	y := celluleY + depart.dy

	// duree de vol : plus la cible est loin, plus la cloche est ample
	cx, cy := x+float64(b.Dx())/2, y+float64(b.Dy())/2
	d := math.Hypot(l.CibleX-cx, l.CibleY-cy)
	t := volMin + (volMax-volMin)*math.Min(1, d/volDistance)

	return banane{
		fx: x, fy: y,
		vx: (l.CibleX - cx) / t,
		vy: (l.CibleY-cy)/t - 0.5*graviteBanane*t,
	}, true
}

// lancerBanane met en vol une banane partie de la main du singe.
func (s *scene) lancerBanane(l vie.Lancer, celluleX, celluleY float64) {
	if !s.fenetresOK || len(s.bananes) >= bananeMax {
		return
	}
	b, ok := s.nouvelleBanane(l, celluleX, celluleY)
	if !ok {
		return
	}
	img := s.bananeImg.Bounds()
	cal, err := calque.Ouvrir(nomBanane, img.Dx(), img.Dy())
	if err != nil {
		log.Printf("fenetre de banane : %v", err)
		return
	}
	cal.Traversant(true) // une banane en vol ne doit pas gener les clics
	b.cal = cal
	s.bananes = append(s.bananes, &b)
}

// gererBananes fait suivre leur parabole aux bananes en vol et ramasse celles
// qui ont quitte l'ecran.
func (s *scene) gererBananes(dt float64) {
	vivantes := s.bananes[:0]
	for _, b := range s.bananes {
		b.vy += graviteBanane * dt
		b.fx += b.vx * dt
		b.fy += b.vy * dt

		hors := b.fy > float64(s.hautEcran) ||
			b.fx < -float64(s.bananeImg.Bounds().Dx()) ||
			b.fx > float64(s.ecranL)
		if hors {
			b.cal.Fermer()
			continue
		}
		b.cal.Afficher(s.bananeImg, int(b.fx), int(b.fy))
		vivantes = append(vivantes, b)
	}
	s.bananes = vivantes
}

func (s *scene) fermerBananes() {
	for _, b := range s.bananes {
		b.cal.Fermer()
	}
	s.bananes = nil
}

// amasJaune cherche la banane dans une image : le plus gros groupe de pixels
// jaunes tenants. Le sprite en compte parfois un isole sur le corps du singe,
// qu'il ne faut pas confondre avec le fruit.
func amasJaune(img image.Image) (image.Rectangle, bool) {
	b := img.Bounds()
	l, h := b.Dx(), b.Dy()
	vu := make([]bool, l*h)
	jaune := make([]bool, l*h)
	for y := 0; y < h; y++ {
		for x := 0; x < l; x++ {
			jaune[y*l+x] = estJaune(img.At(b.Min.X+x, b.Min.Y+y))
		}
	}

	var meilleure image.Rectangle
	meilleurN := 0
	pile := make([]int, 0, 64)
	for depart := range jaune {
		if !jaune[depart] || vu[depart] {
			continue
		}
		pile, vu[depart] = append(pile[:0], depart), true
		n := 0
		// on part du rectangle vide : Union l'ignore, alors qu'un rectangle
		// sentinelle serait remis a l'endroit par image.Rect
		var boite image.Rectangle
		for len(pile) > 0 {
			i := pile[len(pile)-1]
			pile = pile[:len(pile)-1]
			x, y := i%l, i/l
			n++
			boite = boite.Union(image.Rect(x, y, x+1, y+1))
			for _, v := range [4]int{i - 1, i + 1, i - l, i + l} {
				if v < 0 || v >= len(jaune) || vu[v] || !jaune[v] {
					continue
				}
				if (v == i-1 && x == 0) || (v == i+1 && x == l-1) {
					continue // pas de passage d'un bord a l'autre
				}
				vu[v] = true
				pile = append(pile, v)
			}
		}
		if n > meilleurN {
			meilleurN, meilleure = n, boite
		}
	}
	if meilleurN < 4 {
		return image.Rectangle{}, false
	}
	return meilleure.Add(b.Min), true
}

// estJaune reconnait le jaune de la banane. Le vert doit y etre presque aussi
// fort que le rouge : c'est ce qui la separe du poil brun du singe, ou le vert
// est nettement en retrait (#b36739 contre #fbf236).
func estJaune(c color.Color) bool {
	r, g, bl, a := c.RGBA()
	if a>>8 < 128 {
		return false
	}
	r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
	return r8 > 140 && g8 > 100 && b8 < 140 &&
		r8-b8 > 60 && g8-b8 > 40 && g8*10 > r8*9
}

func decouper(img image.Image, boite image.Rectangle) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, boite.Dx(), boite.Dy()))
	for y := 0; y < boite.Dy(); y++ {
		for x := 0; x < boite.Dx(); x++ {
			out.Set(x, y, img.At(boite.Min.X+x, boite.Min.Y+y))
		}
	}
	return out
}
