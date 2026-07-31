// Package bulle dessine la bulle de dialogue au-dessus du singe.
//
// Le style suit celui des sprites : tout est trace pixel par pixel, sans aucun
// lissage, puis agrandi au plus proche voisin. Les coins et la pointe sont donc
// en escalier, et le texte utilise une police bitmap plutot qu'une police
// vectorielle — sans quoi la bulle jurerait avec le pixel art.
package bulle

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	margeX, margeY = 4, 3 // espace entre le texte et le bord, en pixels d'origine
	largeurMax     = 150  // au-dela, le texte passe a la ligne
	pointeBase     = 5    // largeur de la pointe a sa base
	pointeHaut     = 3    // hauteur de la pointe
)

// coinsRetraits donne le retrait horizontal des premieres lignes du cadre, ce
// qui dessine un coin arrondi en escalier. Applique aussi en bas, en miroir.
var coinsRetraits = []int{2, 1}

// Palette de boite de dialogue de jeu : fond blanc, contour et texte noirs.
var (
	couleurFond  = color.RGBA{0xff, 0xff, 0xff, 0xff}
	couleurBord  = color.RGBA{0x00, 0x00, 0x00, 0xff}
	couleurTexte = color.RGBA{0x00, 0x00, 0x00, 0xff}
)

// Bulle rend un texte dans un phylactere pixelise.
type Bulle struct {
	face    font.Face
	echelle int

	// Le rendu ne change qu'avec le texte : on garde le dernier sous la main
	// plutot que de le refaire soixante fois par seconde.
	dernierTexte   string
	dernierRendu   *image.RGBA
	dernierePointe int
}

// Nouvelle prepare la bulle. echelle est le facteur d'agrandissement, a aligner
// sur celui des sprites pour que les pixels aient la meme taille.
func Nouvelle(echelle int) (*Bulle, error) {
	if echelle < 1 {
		echelle = 1
	}
	return &Bulle{face: bitmapfont.Face, echelle: echelle}, nil
}

// nettoyer retire les caracteres absents de la police, qui s'afficheraient
// sinon en carre vide. Les phrases peuvent ainsi comporter des emoji : ils sont
// simplement ignores.
func (b *Bulle) nettoyer(s string) string {
	var sortie []rune
	for _, r := range s {
		if _, ok := b.face.GlyphAdvance(r); ok {
			sortie = append(sortie, r)
		}
	}
	// espaces de bord
	for len(sortie) > 0 && sortie[0] == ' ' {
		sortie = sortie[1:]
	}
	for len(sortie) > 0 && sortie[len(sortie)-1] == ' ' {
		sortie = sortie[:len(sortie)-1]
	}
	return string(sortie)
}

// couper repartit le texte en lignes tenant dans largeurMax.
func (b *Bulle) couper(s string) []string {
	mots := champs(b.nettoyer(s))
	if len(mots) == 0 {
		return nil
	}
	var lignes []string
	ligne := mots[0]
	for _, m := range mots[1:] {
		essai := ligne + " " + m
		if font.MeasureString(b.face, essai).Ceil() > largeurMax {
			lignes = append(lignes, ligne)
			ligne = m
		} else {
			ligne = essai
		}
	}
	return append(lignes, ligne)
}

func champs(s string) []string {
	var out []string
	courant := ""
	for _, r := range s {
		if r == ' ' {
			if courant != "" {
				out = append(out, courant)
				courant = ""
			}
			continue
		}
		courant += string(r)
	}
	if courant != "" {
		out = append(out, courant)
	}
	return out
}

// rendre construit l'image de la bulle : trace a l'echelle 1, puis agrandie.
func (b *Bulle) rendre(texte string) *image.RGBA {
	lignes := b.couper(texte)
	if len(lignes) == 0 {
		return nil
	}

	interligne := b.face.Metrics().Height.Ceil()
	txtL := 0
	for _, l := range lignes {
		if w := font.MeasureString(b.face, l).Ceil(); w > txtL {
			txtL = w
		}
	}
	larg := txtL + 2*margeX
	corpsH := len(lignes)*interligne + 2*margeY
	haut := corpsH + pointeHaut
	pointeX := larg / 2

	brut := image.NewRGBA(image.Rect(0, 0, larg, haut))

	// La silhouette est d'abord remplie, puis on repasse en couleur de bord
	// tout pixel qui touche le vide : le contour epouse ainsi coins et pointe
	// sans avoir a les tracer separement.
	dans := func(x, y int) bool { return dansSilhouette(x, y, larg, corpsH, haut, pointeX) }
	for y := 0; y < haut; y++ {
		for x := 0; x < larg; x++ {
			if dans(x, y) {
				brut.SetRGBA(x, y, couleurFond)
			}
		}
	}
	for y := 0; y < haut; y++ {
		for x := 0; x < larg; x++ {
			if !dans(x, y) {
				continue
			}
			if !dans(x-1, y) || !dans(x+1, y) || !dans(x, y-1) || !dans(x, y+1) {
				brut.SetRGBA(x, y, couleurBord)
			}
		}
	}

	dessinateur := &font.Drawer{
		Dst:  brut,
		Src:  image.NewUniform(couleurTexte),
		Face: b.face,
	}
	base := margeY + b.face.Metrics().Ascent.Ceil()
	for i, l := range lignes {
		w := font.MeasureString(b.face, l).Ceil()
		dessinateur.Dot = fixed.P((larg-w)/2, base+i*interligne)
		dessinateur.DrawString(l)
	}

	b.dernierePointe = pointeX * b.echelle
	return agrandir(brut, b.echelle)
}

// dansSilhouette decrit le contour : un cadre aux coins retires en escalier,
// prolonge vers le bas par une pointe triangulaire.
func dansSilhouette(x, y, larg, corpsH, haut, pointeX int) bool {
	if x < 0 || y < 0 || x >= larg || y >= haut {
		return false
	}
	if y < corpsH {
		retrait := 0
		if y < len(coinsRetraits) {
			retrait = coinsRetraits[y]
		} else if bas := corpsH - 1 - y; bas < len(coinsRetraits) {
			retrait = coinsRetraits[bas]
		}
		return x >= retrait && x < larg-retrait
	}
	demi := pointeBase/2 - (y - corpsH)
	return demi >= 0 && x >= pointeX-demi && x <= pointeX+demi
}

func agrandir(src *image.RGBA, z int) *image.RGBA {
	if z == 1 {
		return src
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx()*z, b.Dy()*z))
	for y := 0; y < b.Dy()*z; y++ {
		for x := 0; x < b.Dx()*z; x++ {
			i := dst.PixOffset(x, y)
			j := src.PixOffset(x/z, y/z)
			copy(dst.Pix[i:i+4], src.Pix[j:j+4])
		}
	}
	return dst
}

func (b *Bulle) image(texte string) *image.RGBA {
	if texte != b.dernierTexte || b.dernierRendu == nil {
		b.dernierRendu = b.rendre(texte)
		b.dernierTexte = texte
	}
	return b.dernierRendu
}

// Taille renvoie les dimensions qu'occupera la bulle pour ce texte.
func (b *Bulle) Taille(texte string) (int, int) {
	img := b.image(texte)
	if img == nil {
		return 0, 0
	}
	return img.Bounds().Dx(), img.Bounds().Dy()
}

// Dessiner peint la bulle dans dst, coin haut gauche en (x, y).
// alpha, entre 0 et 1, permet les fondus d'apparition et de disparition.
func (b *Bulle) Dessiner(dst *image.RGBA, texte string, x, y int, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	img := b.image(texte)
	if img == nil {
		return
	}
	composer(dst, img, x, y, alpha)
}

// composer superpose src sur dst en attenuant l'ensemble.
//
// Les couleurs de Go sont a alpha premultiplie : les quatre composantes doivent
// etre attenuees ensemble. N'attenuer que l'alpha produirait des couleurs
// invalides, rendues en teintes aberrantes.
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
