// Package planche charge les planches de sprites et leurs animations.
//
// Une planche est decrite par un fichier JSON, ce qui permet de changer
// entierement de jeu de sprites sans toucher au code : il suffit de fournir une
// nouvelle image et son descripteur.
//
//	{
//	  "image": "singe.png",
//	  "cellule": [48, 48],        largeur, hauteur d'une case
//	  "echelle": 1.333,           agrandissement a l'affichage
//	  "vue": "dessus",            "dessus" (4 directions) ou "profil"
//	  "animations": {
//	    "marche_bas": {"cellules": [[0,0],[1,0],[2,0],[1,0]], "ms": 120},
//	    "repos_bas":  {"cellules": [[1,0]], "ms": 400},
//	    "meurt":      {"cellules": [[0,4],[1,4]], "ms": 90, "boucle": false}
//	  }
//	}
//
// Les noms d'animation suivent la convention "action" ou "action_direction",
// les directions etant bas, gauche, droite et haut. Une planche vue de profil
// n'a que gauche et droite ; le code appelant demande une direction et recoit
// l'animation la plus proche disponible.
//
// Les images sont mises a l'echelle une fois pour toutes au chargement, par
// echantillonnage au plus proche voisin : c'est ce qui garde les contours nets
// du pixel art, la ou une interpolation les rendrait flous.
package planche

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"io/fs"
	"path"
	"strings"
)

// Animation est une suite d'images affichees en boucle ou une seule fois.
type Animation struct {
	Images []*image.RGBA
	MS     int // duree d'affichage d'une image
	Boucle bool

	// VideEnHaut est le nombre de rangees transparentes au-dessus du
	// personnage, la plus faible parmi les images de cette animation. Les
	// sprites laissent souvent de la marge dans leur cellule ; sans cette
	// mesure, ce qu'on place au-dessus du personnage semblerait flotter.
	//
	// La valeur est propre a chaque animation : une attaque qui projette un
	// objet en l'air occupe toute la cellule, la ou un simple repos en laisse
	// un bon tiers vide.
	VideEnHaut int

	// VideEnBas est son pendant sous le personnage. Il sert a poser le corps
	// exactement sur la barre des taches, et non le bas de sa cellule.
	VideEnBas int
}

// Duree renvoie la duree totale d'un cycle, en millisecondes.
func (a *Animation) Duree() int { return a.MS * len(a.Images) }

// Image renvoie l'image a afficher apres ms millisecondes d'animation.
func (a *Animation) Image(ms int) *image.RGBA {
	if len(a.Images) == 0 {
		return nil
	}
	i := ms / max(a.MS, 1)
	if i >= len(a.Images) {
		if a.Boucle {
			i %= len(a.Images)
		} else {
			i = len(a.Images) - 1
		}
	}
	return a.Images[i]
}

// Finie indique si une animation non bouclee est arrivee a son terme.
func (a *Animation) Finie(ms int) bool { return !a.Boucle && ms >= a.Duree() }

type descripteur struct {
	Nom        string  `json:"nom"`
	Image      string  `json:"image"`
	Cellule    [2]int  `json:"cellule"`
	Echelle    float64 `json:"echelle"`
	Vue        string  `json:"vue"`
	Animations map[string]struct {
		Cellules [][2]int `json:"cellules"`
		MS       int      `json:"ms"`
		Boucle   *bool    `json:"boucle"`
	} `json:"animations"`
}

// Planche regroupe toutes les animations d'un personnage.
type Planche struct {
	Nom              string
	Vue              string
	Largeur, Hauteur int // dimensions d'une image, apres mise a l'echelle

	animations map[string]*Animation
}

// Charger lit un descripteur et son image depuis un systeme de fichiers
// (typiquement embarque dans le binaire via go:embed).
func Charger(fsys fs.FS, chemin string) (*Planche, error) {
	brut, err := fs.ReadFile(fsys, chemin)
	if err != nil {
		return nil, fmt.Errorf("descripteur %s : %w", chemin, err)
	}

	var d descripteur
	if err := json.Unmarshal(brut, &d); err != nil {
		return nil, fmt.Errorf("descripteur %s : %w", chemin, err)
	}
	if d.Echelle == 0 {
		d.Echelle = 1
	}

	octets, err := fs.ReadFile(fsys, path.Join(path.Dir(chemin), d.Image))
	if err != nil {
		return nil, fmt.Errorf("image %s : %w", d.Image, err)
	}
	decodee, _, err := image.Decode(bytes.NewReader(octets))
	if err != nil {
		return nil, fmt.Errorf("image %s : %w", d.Image, err)
	}

	feuille := versRGBA(decodee)
	celL, celH := d.Cellule[0], d.Cellule[1]
	finL := int(float64(celL) * d.Echelle)
	finH := int(float64(celH) * d.Echelle)

	p := &Planche{
		Nom:        d.Nom,
		Vue:        strings.ToLower(d.Vue),
		Largeur:    finL,
		Hauteur:    finH,
		animations: make(map[string]*Animation, len(d.Animations)),
	}
	if p.Vue == "" {
		p.Vue = "dessus"
	}

	cache := map[[2]int]*image.RGBA{}
	for nom, spec := range d.Animations {
		anim := &Animation{MS: spec.MS, Boucle: true}
		if anim.MS <= 0 {
			anim.MS = 150
		}
		if spec.Boucle != nil {
			anim.Boucle = *spec.Boucle
		}
		for _, c := range spec.Cellules {
			if _, vu := cache[c]; !vu {
				zone := image.Rect(c[0]*celL, c[1]*celH, c[0]*celL+celL, c[1]*celH+celH)
				cache[c] = agrandir(feuille, zone, finL, finH)
			}
			anim.Images = append(anim.Images, cache[c])
		}
		if len(anim.Images) > 0 {
			anim.VideEnHaut = finH
			anim.VideEnBas = finH
			for _, img := range anim.Images {
				if v := videEnHaut(img); v < anim.VideEnHaut {
					anim.VideEnHaut = v
				}
				if v := videEnBas(img); v < anim.VideEnBas {
					anim.VideEnBas = v
				}
			}
			p.animations[nom] = anim
		}
	}
	if len(p.animations) == 0 {
		return nil, fmt.Errorf("planche %s : aucune animation", chemin)
	}

	return p, nil
}

// videEnHaut compte les rangees entierement transparentes en haut d'une image.
func videEnHaut(img *image.RGBA) int {
	b := img.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if img.Pix[img.PixOffset(x, y)+3] != 0 {
				return y
			}
		}
	}
	return b.Dy()
}

// videEnBas compte les rangees entierement transparentes en bas d'une image.
func videEnBas(img *image.RGBA) int {
	b := img.Bounds()
	for y := b.Dy() - 1; y >= 0; y-- {
		for x := 0; x < b.Dx(); x++ {
			if img.Pix[img.PixOffset(x, y)+3] != 0 {
				return b.Dy() - 1 - y
			}
		}
	}
	return b.Dy()
}

func versRGBA(src image.Image) *image.RGBA {
	if r, ok := src.(*image.RGBA); ok {
		return r
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

// agrandir extrait une cellule et la redimensionne au plus proche voisin.
func agrandir(src *image.RGBA, zone image.Rectangle, larg, haut int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, larg, haut))
	sl, sh := zone.Dx(), zone.Dy()
	for y := 0; y < haut; y++ {
		sy := zone.Min.Y + y*sh/haut
		for x := 0; x < larg; x++ {
			sx := zone.Min.X + x*sl/larg
			i := dst.PixOffset(x, y)
			j := src.PixOffset(sx, sy)
			copy(dst.Pix[i:i+4], src.Pix[j:j+4])
		}
	}
	return dst
}

// Action indique si une action existe, quelle que soit la direction.
func (p *Planche) Action(action string) bool {
	for nom := range p.animations {
		if nom == action || strings.HasPrefix(nom, action+"_") {
			return true
		}
	}
	return false
}

// Obtenir renvoie l'animation la plus specifique disponible pour une action et
// une direction, avec repli progressif : "marche_gauche", puis "marche", puis
// une animation de repos, puis n'importe laquelle. Ne renvoie jamais nil, ce
// qui evite au code appelant d'avoir a verifier ce que contient la planche.
func (p *Planche) Obtenir(action, direction string) *Animation {
	if direction != "" {
		if a, ok := p.animations[action+"_"+direction]; ok {
			return a
		}
	}
	if a, ok := p.animations[action]; ok {
		return a
	}
	for _, repli := range []string{"repos_bas", "repos", "repos_droite", "marche_bas"} {
		if a, ok := p.animations[repli]; ok {
			return a
		}
	}
	for _, a := range p.animations {
		return a
	}
	return nil
}
