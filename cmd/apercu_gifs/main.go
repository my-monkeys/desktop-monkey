// Commande apercu_gifs fabrique les petits GIF animes du README : une
// animation par fonctionnalite (attaque, saut, grimpe, cacas...), rendue avec
// le vrai moteur de sprites du projet et posee sur un fond de bureau discret.
//
//	go run ./cmd/apercu_gifs
//
// Les fichiers sont ecrits dans docs/. Ce n'est pas embarque dans l'appli : un
// simple utilitaire d'auteur, comme cmd/embete.
package main

import (
	"image"
	"image/color"
	"image/gif"
	"log"
	"os"
	"path/filepath"

	"github.com/my-monkeys/desktop-monkey/internal/coeurs"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
)

const (
	canL = 480 // largeur de la scene
	canH = 230 // hauteur de la scene
	marB = 22  // marge sous le sprite (le "sol")
)

func main() {
	// sprites rendus a 2x (par-dessus l'echelle du descripteur) : le GIF est
	// ensuite affiche un peu reduit dans le README, donc bien net.
	singe, err := planche.Charger(ressources.Fichiers, "assets/singe2.json", 2)
	if err != nil {
		log.Fatal(err)
	}
	crot, err := planche.Charger(ressources.Fichiers, "assets/crotte.json", 3)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll("docs", 0o755); err != nil {
		log.Fatal(err)
	}

	// Chaque scene : un nom de fichier, l'animation jouee, et des options.
	ecrire("docs/attaque.gif", scene(singe.Obtenir("attaque", "droite"), 3, opts{}))
	ecrire("docs/saut.gif", scene(singe.Obtenir("saute", "droite"), 3, opts{}))
	ecrire("docs/grimpe.gif", scene(singe.Obtenir("grimpe", "droite"), 3, opts{colle: gauche}))
	ecrire("docs/chute-coeur.gif", scene(singe.Obtenir("tombe", "droite"), 3, opts{}))
	ecrire("docs/hero.gif", scene(singe.Obtenir("marche", "droite"), 4, opts{}))
	ecrire("docs/vol-curseur.gif", scene(singe.Obtenir("marche", "droite"), 4, opts{curseur: true}))
	ecrire("docs/coeurs.gif", scene(singe.Obtenir("touche", "droite"), 3, opts{coeurs: true}))
	ecrire("docs/cadavre-taskbar.gif", mort(singe.Obtenir("meurt", "droite")))
	ecrire("docs/poop.gif", poop(crot))

	log.Printf("GIF ecrits dans docs/")
}

// ancrage horizontal du sprite
type ancrage int

const (
	centre ancrage = iota
	gauche
)

type opts struct {
	colle   ancrage // ou coller le sprite horizontalement
	curseur bool    // dessine un curseur porte par le singe
	coeurs  bool    // trois coeurs au-dessus de la tete, qui s'egrenent
}

// scene joue une animation en boucle, repetee assez de fois pour durer, et
// renvoie les images du GIF avec leurs delais (en centiemes de seconde).
func scene(a *planche.Animation, tours int, o opts) *gif.GIF {
	if a == nil || len(a.Images) == 0 {
		log.Fatal("animation manquante")
	}
	g := &gif.GIF{LoopCount: 0}
	total := len(a.Images) * tours
	for i := 0; i < total; i++ {
		src := a.Images[i%len(a.Images)]
		dst := fond()
		sx := (canL - src.Bounds().Dx()) / 2
		if o.colle == gauche {
			sx = canL/4 - src.Bounds().Dx()/2
		}
		sy := canH - src.Bounds().Dy() - marB
		poser(dst, src, sx, sy)

		if o.curseur {
			dessinerCurseur(dst, sx+src.Bounds().Dx()-16, sy+src.Bounds().Dy()/2)
		}
		if o.coeurs {
			// 3 coeurs le premier tiers, 2 le deuxieme, 1 le dernier
			restants := 3 - i*3/total
			if restants < 1 {
				restants = 1
			}
			ech := 6
			cx := canL / 2
			cy := sy - coeurs.Hauteur(ech) - 12
			coeurs.Dessiner(dst, cx, cy, restants, 3, ech, 0.2, 1)
		}
		ajouter(g, dst, a.MS/10)
	}
	return g
}

// mort joue l'animation de mort puis laisse le corps au sol un instant.
func mort(a *planche.Animation) *gif.GIF {
	g := &gif.GIF{LoopCount: 0}
	for _, src := range a.Images {
		dst := fond()
		sx := (canL - src.Bounds().Dx()) / 2
		sy := canH - src.Bounds().Dy() - marB
		poser(dst, src, sx, sy)
		ajouter(g, dst, a.MS/10)
	}
	// le corps reste allonge ~1,4 s avant que la boucle ne reprenne
	last := a.Images[len(a.Images)-1]
	dst := fond()
	poser(dst, last, (canL-last.Bounds().Dx())/2, canH-last.Bounds().Dy()-marB)
	ajouter(g, dst, 140)
	return g
}

// poop joue l'apparition de la crotte puis sa fumee, en boucle.
func poop(p *planche.Planche) *gif.GIF {
	g := &gif.GIF{LoopCount: 0}
	apparait := p.Obtenir("apparait", "")
	repos := p.Obtenir("repos", "")
	pose := func(a *planche.Animation, tours int) {
		for i := 0; i < len(a.Images)*tours; i++ {
			src := a.Images[i%len(a.Images)]
			dst := fond()
			sx := (canL - src.Bounds().Dx()) / 2
			sy := canH - src.Bounds().Dy() - marB
			poser(dst, src, sx, sy)
			ajouter(g, dst, a.MS/10)
		}
	}
	pose(apparait, 1)
	pose(repos, 3)
	return g
}

// fond fabrique un fond de bureau : un degrade vertical bleu nuit, en bandes
// pour garder peu de couleurs (le GIF n'en accepte que 256).
func fond() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, canL, canH))
	haut := color.RGBA{0x1b, 0x24, 0x38, 0xff}
	bas := color.RGBA{0x2f, 0x3d, 0x5c, 0xff}
	const bandes = 22
	for y := 0; y < canH; y++ {
		t := float64(y*bandes/canH) / float64(bandes-1)
		c := color.RGBA{
			uint8(float64(haut.R) + t*float64(bas.R-haut.R)),
			uint8(float64(haut.G) + t*float64(bas.G-haut.G)),
			uint8(float64(haut.B) + t*float64(bas.B-haut.B)),
			0xff,
		}
		for x := 0; x < canL; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

// poser compose un sprite sur le fond (alpha simple : opaque ou transparent).
func poser(dst, src *image.RGBA, x, y int) {
	b := src.Bounds()
	for sy := 0; sy < b.Dy(); sy++ {
		for sx := 0; sx < b.Dx(); sx++ {
			o := src.PixOffset(b.Min.X+sx, b.Min.Y+sy)
			if src.Pix[o+3] < 128 {
				continue
			}
			dst.SetRGBA(x+sx, y+sy, color.RGBA{
				src.Pix[o], src.Pix[o+1], src.Pix[o+2], 0xff,
			})
		}
	}
}

// dessinerCurseur stampe une petite fleche blanche (curseur vole).
func dessinerCurseur(dst *image.RGBA, x, y int) {
	// silhouette d'un pointeur classique, 0 = rien, 1 = noir, 2 = blanc
	m := []string{
		"1",
		"12",
		"122",
		"1222",
		"12222",
		"122222",
		"1222221",
		"12222110",
		"12211010",
		"1201",
		"0010",
	}
	const z = 2 // curseur dessine a la meme echelle 2x que les sprites
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
			for dy := 0; dy < z; dy++ {
				for dx := 0; dx < z; dx++ {
					px, py := x+gx*z+dx, y+gy*z+dy
					if px >= 0 && px < canL && py >= 0 && py < canH {
						dst.SetRGBA(px, py, col)
					}
				}
			}
		}
	}
}

// ajouter quantifie une image et l'ajoute au GIF avec son delai.
func ajouter(g *gif.GIF, img *image.RGBA, delaiCs int) {
	if delaiCs < 2 {
		delaiCs = 2
	}
	g.Image = append(g.Image, versPaletted(img))
	g.Delay = append(g.Delay, delaiCs)
}

// versPaletted construit une palette exacte de l'image (peu de couleurs :
// pixel art + fond en bandes) et l'indexe. Reduit la precision si besoin pour
// tenir dans les 256 couleurs du format GIF.
func versPaletted(img *image.RGBA) *image.Paletted {
	for _, bits := range []uint{8, 7, 6, 5, 4} {
		masque := uint8(0xff << (8 - bits))
		vues := map[color.RGBA]int{}
		var pal color.Palette
		ok := true
		for i := 0; i < len(img.Pix); i += 4 {
			c := color.RGBA{
				img.Pix[i] & masque, img.Pix[i+1] & masque,
				img.Pix[i+2] & masque, 0xff,
			}
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
				c := color.RGBA{
					img.Pix[o] & masque, img.Pix[o+1] & masque,
					img.Pix[o+2] & masque, 0xff,
				}
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
