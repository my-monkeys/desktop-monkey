// Commande apercu_gifs fabrique les GIF animes du README en pilotant le vrai
// moteur de comportement : on cree un singe (internal/vie), on lui joue une
// souris scriptee (il s'approche, clique, s'eloigne...) et on filme sa
// reaction reelle — il marche, suit, encaisse, grimpe, pond. Rien n'est simule
// a la main : ce sont les memes etats et animations que sur le bureau.
//
//	go run ./cmd/apercu_gifs
//
// Ecrit dans docs/. Simple utilitaire d'auteur, non embarque dans l'appli.
package main

import (
	"image"
	"image/color"
	"image/gif"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/my-monkeys/desktop-monkey/internal/coeurs"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
	"github.com/my-monkeys/desktop-monkey/internal/vie"
)

const (
	out  = 2       // agrandissement final : pixel art net une fois affiche
	fps  = 20      // images par seconde du rendu et du GIF
	sol  = 12      // hauteur de la "barre des taches" en bas de la scene
	echC = 3       // echelle des coeurs
)

var (
	singe, crotte *planche.Planche
	crotteSol     int // base du tas dans sa cellule
)

func main() {
	var err error
	if singe, err = planche.Charger(ressources.Fichiers, "assets/singe2.json"); err != nil {
		log.Fatal(err)
	}
	if crotte, err = planche.Charger(ressources.Fichiers, "assets/crotte.json"); err != nil {
		log.Fatal(err)
	}
	crotteSol = crotte.Hauteur - crotte.Obtenir("repos", "").VideEnBas

	if err := os.MkdirAll("docs", 0o755); err != nil {
		log.Fatal(err)
	}

	for _, s := range scenes() {
		ecrire("docs/"+s.nom+".gif", rendre(s))
	}
	log.Printf("GIF ecrits dans docs/")
}

// scene decrit un petit scenario : une taille, une duree, des probabilites de
// comportement forcees, et la trajectoire de la souris image par image.
type scene struct {
	nom        string
	larg, haut int
	images     int
	regle      func(*vie.Reglages)
	souris     func(f int, v *vie.Vie, larg, haut int) (x, y float64, clic bool)
}

func scenes() []scene {
	// souris posee immobile hors du singe (coin haut), pour les scenes ou elle
	// ne joue pas de role actif
	coin := func(f int, v *vie.Vie, larg, haut int) (float64, float64, bool) {
		return float64(larg) - 22, 16, false
	}
	return []scene{
		{
			// il marche sur le bureau en suivant la souris de va-et-vient
			nom: "marche", larg: 360, haut: 168, images: 96,
			regle: func(r *vie.Reglages) {
				r.ChanceAmi = 1
				r.ChanceChasse = 0
				r.SeuilVieSeule = 0.3
				r.Vitesse = 3.2
			},
			souris: func(f int, v *vie.Vie, larg, haut int) (float64, float64, bool) {
				t := float64(f) / fps
				x := 44 + (float64(larg)-88)*(0.5+0.5*math.Sin(t*0.9))
				return x, float64(haut-sol) - 46, false
			},
		},
		{
			// il chasse la souris et l'attaque a coups de banane
			nom: "chasse", larg: 340, haut: 168, images: 110,
			regle: func(r *vie.Reglages) {
				r.ChanceAmi = 1
				r.ChanceChasse = 1
				r.ChanceVol = 0
				r.SeuilVieSeule = 0.3
			},
			souris: func(f int, v *vie.Vie, larg, haut int) (float64, float64, bool) {
				t := float64(f) / fps
				x := 70 + (float64(larg)-140)*(0.5+0.5*math.Sin(t*0.55))
				return x, float64(haut-sol) - 40, false
			},
		},
		{
			// on lui clique dessus : la souris le suit, coeurs, recul a chaque
			// coup, puis il meurt et tombe sur la barre des taches
			nom: "degats", larg: 340, haut: 184, images: 130,
			regle: func(r *vie.Reglages) {
				r.ChanceAmi = 0
				r.DureeRepos = [2]float64{40, 40} // il reste sur place entre les coups
				r.Coeurs = 3
			},
			souris: func(f int, v *vie.Vie, larg, haut int) (float64, float64, bool) {
				cx, cy := v.Centre() // la souris colle au singe, meme quand il recule
				if f < 14 {          // petite approche depuis la gauche
					k := float64(f) / 14
					cx = 30 + (cx-30)*k
					cy = float64(haut)/2 + (cy-float64(haut)/2)*k
				}
				// trois coups : un front montant isole a 0,9 s / 1,7 s / 2,5 s
				clic := f == 18 || f == 34 || f == 50
				return cx, cy, clic
			},
		},
		{
			// il escalade un bord de l'ecran puis se laisse tomber
			nom: "grimpe", larg: 240, haut: 232, images: 130,
			regle: func(r *vie.Reglages) {
				r.ChanceAmi = 0
				r.ChanceGrimpe = 1
				r.DureeRepos = [2]float64{0.2, 0.35}
			},
			souris: coin,
		},
		{
			// il pond, puis s'ecarte pour reveler le tas fumant : souris immobile
			// => il vaque et pond ; souris qui bouge => il va la rejoindre
			nom: "pond", larg: 300, haut: 178, images: 128,
			regle: func(r *vie.Reglages) {
				r.ChanceAmi = 1
				r.SeuilVieSeule = 0.15 // des que la souris se fige, il vit sa vie
				r.ChanceCrotte = 1
				r.DureeRepos = [2]float64{0.3, 0.45}
				r.Vitesse = 3.4
				r.DistArret = 10 // il va coller a la souris, donc bien loin du tas
				r.DistRepart = 18
			},
			souris: func(f int, v *vie.Vie, larg, haut int) (float64, float64, bool) {
				y := float64(haut-sol) - 40
				if f < 50 { // immobile au centre : il s'installe et pond
					return float64(larg) / 2, y, false
				}
				// puis la souris file tout a gauche : il la suit et laisse le tas
				k := math.Min(1, float64(f-50)/26)
				return float64(larg)/2 - (float64(larg)/2-24)*k, y, false
			},
		},
	}
}

type tas struct {
	x, y int
	ms   float64
}

// rendre joue le scenario et filme le singe : chaque image est le rendu reel de
// l'etat courant (animation, coeurs, crottes) plus la souris scriptee.
func rendre(s scene) *gif.GIF {
	r := vie.ReglagesParDefaut()
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu = 0, 0, 0
	r.ChanceGrimpe, r.ChanceCrotte, r.ChanceChasse = 0, 0, 0
	r.AvantSieste = 1e9
	r.DureeRepos = [2]float64{0.3, 0.5}
	if s.regle != nil {
		s.regle(&r)
	}
	v := vie.Nouvelle(singe, r, s.larg, s.haut, s.haut-sol)

	var tases []tas
	g := &gif.GIF{LoopCount: 0}
	const dt = 1.0 / fps
	for f := 0; f < s.images; f++ {
		cx, cy, clic := s.souris(f, v, s.larg, s.haut)
		v.Avancer(dt, cx, cy, clic)
		if px, py, ok := v.PrendreCrotte(); ok {
			tases = append(tases, tas{int(px), int(py), 0})
		}
		for i := range tases {
			tases[i].ms += dt * 1000
		}
		ajouter(g, dessiner(s, v, tases, cx, cy), 100/fps)
	}
	return g
}

func dessiner(s scene, v *vie.Vie, tases []tas, cx, cy float64) *image.RGBA {
	dst := fond(s.larg, s.haut)

	for _, t := range tases { // les crottes, derriere le singe
		img := crotteImage(t.ms)
		poser(dst, img, t.x-crotte.Largeur/2, t.y-crotteSol)
	}

	if v.Visible() {
		anim, ms := v.Animation()
		if img := anim.Image(ms); img != nil {
			if fl := v.Flash(); fl > 0 && int(fl*18)%2 == 0 {
				img = teinter(img)
			}
			poser(dst, img, int(v.X), int(v.Y))
		}
	}

	if d := v.DepuisCoup(); d < 2.8 && v.Visible() && v.Etat() != vie.Cadavre {
		restants, total := v.Coeurs()
		larg, _ := v.Taille()
		cyh := int(v.HautDuCorps()) - coeurs.Hauteur(echC) - 4
		if cyh < 0 {
			cyh = 0
		}
		coeurs.Dessiner(dst, int(v.X)+int(larg)/2, cyh, restants, total, echC, d, 1)
	}

	dessinerCurseur(dst, int(cx), int(cy), s.larg, s.haut)
	return agrandir(dst, out)
}

// crotteImage : l'apparition, puis la fumee en boucle.
func crotteImage(ms float64) *image.RGBA {
	a := crotte.Obtenir("apparait", "")
	if !a.Finie(int(ms)) {
		return a.Image(int(ms))
	}
	repos := crotte.Obtenir("repos", "")
	return repos.Image(int(ms) - a.Duree())
}

// fond : degrade bleu nuit en bandes (peu de couleurs pour le GIF) et une
// barre des taches sombre en bas, sur laquelle le singe se tient.
func fond(larg, haut int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, larg, haut))
	h := color.RGBA{0x1b, 0x24, 0x38, 0xff}
	b := color.RGBA{0x2c, 0x39, 0x56, 0xff}
	const bandes = 20
	for y := 0; y < haut; y++ {
		t := float64(y*bandes/haut) / float64(bandes-1)
		c := color.RGBA{
			uint8(float64(h.R) + t*float64(b.R-h.R)),
			uint8(float64(h.G) + t*float64(b.G-h.G)),
			uint8(float64(h.B) + t*float64(b.B-h.B)), 0xff,
		}
		for x := 0; x < larg; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	barre := color.RGBA{0x14, 0x18, 0x24, 0xff}
	for y := haut - sol; y < haut; y++ {
		for x := 0; x < larg; x++ {
			img.SetRGBA(x, y, barre)
		}
	}
	for x := 0; x < larg; x++ { // fin liseret clair au sommet de la barre
		img.SetRGBA(x, haut-sol, color.RGBA{0x3a, 0x46, 0x64, 0xff})
	}
	return img
}

func poser(dst, src *image.RGBA, x, y int) {
	b := src.Bounds()
	for sy := 0; sy < b.Dy(); sy++ {
		for sx := 0; sx < b.Dx(); sx++ {
			o := src.PixOffset(b.Min.X+sx, b.Min.Y+sy)
			if src.Pix[o+3] < 128 {
				continue
			}
			px, py := x+sx, y+sy
			if px < 0 || py < 0 || px >= dst.Bounds().Dx() || py >= dst.Bounds().Dy() {
				continue
			}
			dst.SetRGBA(px, py, color.RGBA{src.Pix[o], src.Pix[o+1], src.Pix[o+2], 0xff})
		}
	}
}

// teinter rougit le singe le temps d'un flash de degats.
func teinter(src *image.RGBA) *image.RGBA {
	d := image.NewRGBA(src.Bounds())
	copy(d.Pix, src.Pix)
	for i := 0; i < len(d.Pix); i += 4 {
		if d.Pix[i+3] == 0 {
			continue
		}
		d.Pix[i] = uint8(min(255, int(d.Pix[i])+90))
		d.Pix[i+1] /= 2
		d.Pix[i+2] /= 2
	}
	return d
}

// dessinerCurseur stampe une fleche blanche facon pointeur.
func dessinerCurseur(dst *image.RGBA, x, y, larg, haut int) {
	m := []string{
		"1", "12", "122", "1222", "12222", "122222", "1222221",
		"12222110", "12211010", "1201", "0010",
	}
	for gy, ligne := range m {
		for gx, c := range ligne {
			var col color.RGBA
			switch c {
			case '1':
				col = color.RGBA{0x10, 0x10, 0x10, 0xff}
			case '2':
				col = color.RGBA{0xff, 0xff, 0xff, 0xff}
			default:
				continue
			}
			px, py := x+gx, y+gy
			if px >= 0 && px < larg && py >= 0 && py < haut {
				dst.SetRGBA(px, py, col)
			}
		}
	}
}

func agrandir(src *image.RGBA, z int) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*z, b.Dy()*z))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			o := src.PixOffset(x, y)
			c := color.RGBA{src.Pix[o], src.Pix[o+1], src.Pix[o+2], src.Pix[o+3]}
			for dy := 0; dy < z; dy++ {
				for dx := 0; dx < z; dx++ {
					dst.SetRGBA(x*z+dx, y*z+dy, c)
				}
			}
		}
	}
	return dst
}

func ajouter(g *gif.GIF, img *image.RGBA, delaiCs int) {
	if delaiCs < 2 {
		delaiCs = 2
	}
	g.Image = append(g.Image, versPaletted(img))
	g.Delay = append(g.Delay, delaiCs)
}

// versPaletted construit une palette exacte (peu de couleurs : pixel art + fond
// en bandes), en reduisant la precision si besoin pour tenir dans 256 couleurs.
func versPaletted(img *image.RGBA) *image.Paletted {
	for _, bits := range []uint{8, 7, 6, 5, 4} {
		masque := uint8(0xff << (8 - bits))
		vues := map[color.RGBA]int{}
		var pal color.Palette
		ok := true
		for i := 0; i < len(img.Pix); i += 4 {
			c := color.RGBA{img.Pix[i] & masque, img.Pix[i+1] & masque, img.Pix[i+2] & masque, 0xff}
			if _, vu := vues[c]; !vu {
				if len(pal) >= 256 {
					ok = false
					break
				}
				vues[c] = len(pal)
				pal = append(pal, c)
			}
		}
		if !ok {
			continue
		}
		out := image.NewPaletted(img.Bounds(), pal)
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				o := img.PixOffset(x, y)
				c := color.RGBA{img.Pix[o] & masque, img.Pix[o+1] & masque, img.Pix[o+2] & masque, 0xff}
				out.SetColorIndex(x, y, uint8(vues[c]))
			}
		}
		return out
	}
	log.Fatal("trop de couleurs pour un GIF")
	return nil
}

func ecrire(chemin string, g *gif.GIF) {
	f, err := os.Create(chemin)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, g); err != nil {
		log.Fatal(err)
	}
	log.Printf("%s : %d images", filepath.Base(chemin), len(g.Image))
}
