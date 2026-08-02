// Package coeurs dessine les points de vie du singe : une rangee de petits
// coeurs en pixel art, affichee au-dessus de sa tete apres un coup.
//
// Comme la bulle, tout est trace pixel par pixel puis agrandi au plus proche
// voisin, pour rester dans le style des sprites.
package coeurs

import (
	"image"
	"image/color"
)

const (
	largCoeur = 7
	hautCoeur = 6
	espace    = 2 // entre deux coeurs, en pixels d'origine

	// le coeur qui vient d'etre perdu clignote ce temps-la avant de s'eteindre
	dureeClignote = 0.9
)

var motif = [hautCoeur]string{
	".XX.XX.",
	"XXXXXXX",
	"XXXXXXX",
	".XXXXX.",
	"..XXX..",
	"...X...",
}

var (
	rouge = color.RGBA{0xe8, 0x3a, 0x45, 0xff}
	eclat = color.RGBA{0xff, 0xb5, 0xbc, 0xff} // reflet en haut a gauche
	ombre = color.RGBA{0x30, 0x2b, 0x30, 0xaa} // coeur vide, en creux
)

// deux images par echelle : coeur plein et coeur vide
var cache = map[[2]int]*image.RGBA{}

func rendu(plein bool, echelle int) *image.RGBA {
	etat := 0
	if plein {
		etat = 1
	}
	if img, ok := cache[[2]int{etat, echelle}]; ok {
		return img
	}

	img := image.NewRGBA(image.Rect(0, 0, largCoeur*echelle, hautCoeur*echelle))
	for y, ligne := range motif {
		for x, c := range ligne {
			if c != 'X' {
				continue
			}
			couleur := ombre
			if plein {
				couleur = rouge
				if x == 1 && y == 1 {
					couleur = eclat
				}
			}
			for dy := 0; dy < echelle; dy++ {
				for dx := 0; dx < echelle; dx++ {
					img.SetRGBA(x*echelle+dx, y*echelle+dy, couleur)
				}
			}
		}
	}
	cache[[2]int{etat, echelle}] = img
	return img
}

// Icone renvoie l'image d'un seul coeur, plein ou vide. La fenetre de reglages
// s'en sert pour montrer les points de vie tels qu'ils apparaissent vraiment
// au-dessus du singe.
func Icone(plein bool, echelle int) *image.RGBA { return rendu(plein, echelle) }

// Hauteur renvoie la hauteur de la rangee, en pixels.
func Hauteur(echelle int) int { return hautCoeur * echelle }

// Largeur renvoie la largeur totale de la rangee, en pixels.
func Largeur(total, echelle int) int {
	if total <= 0 {
		return 0
	}
	return (total*largCoeur + (total-1)*espace) * echelle
}

// Dessiner peint la rangee de coeurs centree sur cx, bord haut en y. Le coeur
// perdu au dernier coup clignote un court instant avant de s'eteindre. alpha,
// entre 0 et 1, permet le fondu de disparition.
func Dessiner(dst *image.RGBA, cx, y, restants, total, echelle int, depuisCoup, alpha float64) {
	if alpha <= 0 || total <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	largTotal := (total*largCoeur + (total-1)*espace) * echelle
	x := cx - largTotal/2
	for i := 0; i < total; i++ {
		plein := i < restants
		if i == restants && depuisCoup < dureeClignote && int(depuisCoup*16)%2 == 0 {
			plein = true // le coeur perdu bat encore un peu
		}
		composer(dst, rendu(plein, echelle), x, y, alpha)
		x += (largCoeur + espace) * echelle
	}
}

// composer superpose src sur dst en attenuant l'ensemble. Les couleurs de Go
// sont a alpha premultiplie : les quatre composantes s'attenuent ensemble.
func composer(dst, src *image.RGBA, x, y int, alpha float64) {
	b := src.Bounds()
	for sy := 0; sy < b.Dy(); sy++ {
		dy := y + sy
		if dy < 0 || dy >= dst.Bounds().Dy() {
			continue
		}
		for sx := 0; sx < b.Dx(); sx++ {
			dx := x + sx
			if dx < 0 || dx >= dst.Bounds().Dx() {
				continue
			}
			i := src.PixOffset(sx, sy)
			sa := float64(src.Pix[i+3]) * alpha
			if sa == 0 {
				continue
			}
			j := dst.PixOffset(dx, dy)
			inv := 1 - sa/255
			for k := 0; k < 4; k++ {
				s := float64(src.Pix[i+k]) * alpha
				dst.Pix[j+k] = uint8(s + float64(dst.Pix[j+k])*inv)
			}
		}
	}
}
