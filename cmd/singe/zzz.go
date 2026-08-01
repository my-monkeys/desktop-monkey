package main

import (
	"image"
	"image/color"
	"math"
)

// Pendant la sieste, des petits « Z » s'envolent au-dessus du singe : dessines
// en code comme les coeurs, pas d'asset. Trois Z decales d'un tiers de cycle
// montent en zigzaguant et s'estompent.

// glyphe d'un Z, 5x5 blocs
var glypheZ = [5]string{
	"#####",
	"   # ",
	"  #  ",
	" #   ",
	"#####",
}

const (
	zzzCycle  = 2.5  // duree de l'ascension d'un Z, en secondes
	zzzMontee = 26.0 // hauteur de l'ascension, en pixels de base
)

// dessinerZzz peint les Z au-dessus de (x, y) — le sommet de la tete du
// dormeur — a l'instant t (secondes depuis le debut de la sieste).
func dessinerZzz(dst *image.RGBA, x, y int, t float64, ech int) {
	teinte := color.RGBA{0xdd, 0xe6, 0xf5, 0xff} // blanc bleute, facon songe
	for k := 0; k < 3; k++ {
		p := math.Mod(t+float64(k)*zzzCycle/3, zzzCycle) / zzzCycle
		// premier cycle : les Z n'existent pas encore tous, pour eviter
		// qu'ils apparaissent d'un bloc au beau milieu de leur montee
		if t < float64(k)*zzzCycle/3 {
			continue
		}
		alpha := 1 - p
		zx := x + int(math.Sin(p*4*math.Pi+float64(k)*2)*5*float64(ech)) + (k-1)*3*ech
		zy := y - int(p*zzzMontee*float64(ech)) - 2*ech
		poserGlyphe(dst, glypheZ, zx, zy, ech, teinte, alpha)
	}
}

// poserGlyphe dessine un motif en blocs de taille ech, fondu par alpha.
func poserGlyphe(dst *image.RGBA, g [5]string, x, y, ech int, c color.RGBA, alpha float64) {
	if alpha <= 0 {
		return
	}
	b := dst.Bounds()
	for gy, ligne := range g {
		for gx, ch := range ligne {
			if ch != '#' {
				continue
			}
			for dy := 0; dy < ech; dy++ {
				for dx := 0; dx < ech; dx++ {
					px, py := x+gx*ech+dx, y+gy*ech+dy
					if px < b.Min.X || py < b.Min.Y || px >= b.Max.X || py >= b.Max.Y {
						continue
					}
					o := dst.PixOffset(px, py)
					// melange manuel : le fond de scene est transparent, on
					// pose la teinte avec son alpha
					a := alpha
					dst.Pix[o+0] = melange(dst.Pix[o+0], c.R, a)
					dst.Pix[o+1] = melange(dst.Pix[o+1], c.G, a)
					dst.Pix[o+2] = melange(dst.Pix[o+2], c.B, a)
					if na := uint8(a * 255); dst.Pix[o+3] < na {
						dst.Pix[o+3] = na
					}
				}
			}
		}
	}
}

func melange(fond, dessus uint8, a float64) uint8 {
	return uint8(float64(fond)*(1-a) + float64(dessus)*a)
}
