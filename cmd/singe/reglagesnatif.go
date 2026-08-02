package main

// Ce que la page de reglages a besoin pour se dessiner : les planches de
// sprites decoupees en bandes d'animation, les coeurs, la police pixel, les
// humeurs reelles du singe — plus l'ecriture de la configuration et le
// redemarrage qui applique le tout.

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/my-monkeys/desktop-monkey/internal/coeurs"
	"github.com/my-monkeys/desktop-monkey/internal/maj"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
	"github.com/my-monkeys/desktop-monkey/internal/vie"
)

// bande est une animation mise a plat : ses images cote a cote dans un seul
// PNG, que la page fait defiler.
type bande struct {
	Png    string `json:"png"`    // data-URI
	Cadres int    `json:"cadres"` // nombre d'images
	MS     int    `json:"ms"`     // duree d'une image
}

// plancheWeb est tout ce qu'il faut a la page pour animer un personnage.
type plancheWeb struct {
	Cle    string `json:"cle"`  // le chemin du descripteur, tel qu'il ira dans la config
	Nom    string `json:"nom"`  // "Singe pixel"
	Sous   string `json:"sous"` // "de profil · tiki-ted"
	CelL   int    `json:"celL"` // taille d'une image, en pixels physiques
	CelH   int    `json:"celH"` // pour une taille de 1
	Pied   int    `json:"pied"` // vide sous ses pieds quand il est debout
	Marche bande  `json:"marche"`
	Repos  bande  `json:"repos"`
	Touche bande  `json:"touche"`
	Meurt  bande  `json:"meurt"`
}

// planchesWeb prepare les deux personnages livres avec l'application.
func planchesWeb(fr bool) []plancheWeb {
	nom := func(frTxt, enTxt string) string {
		if fr {
			return frTxt
		}
		return enTxt
	}
	descripteurs := []struct{ cle, nom, sous string }{
		{"assets/singe2.json", nom("Singe pixel", "Pixel monkey"), nom("de profil · tiki-ted", "side view · tiki-ted")},
		{"assets/singe.json", nom("Singe RPG", "RPG monkey"), nom("vu de dessus · WhtDragon", "top-down · WhtDragon")},
	}

	var out []plancheWeb
	for _, d := range descripteurs {
		p, err := planche.Charger(ressources.Fichiers, d.cle, 1)
		if err != nil {
			log.Printf("planche %s indisponible : %v", d.cle, err)
			continue
		}
		// une seule boite pour toutes les poses : le singe ne saute pas d'un
		// pixel quand il passe de la marche au repos
		boite := boiteAnimee(p, actionsApercu)
		// la pose couchee descend plus bas que la marche : on retient l'ecart
		// pour poser ses pieds sur le sol de l'apercu
		debout := boiteAnimee(p, []string{"marche", "repos"})
		out = append(out, plancheWeb{
			Cle: d.cle, Nom: d.nom, Sous: d.sous,
			CelL:   boite.Dx() * facteurAffichage,
			CelH:   boite.Dy() * facteurAffichage,
			Pied:   (boite.Max.Y - debout.Max.Y) * facteurAffichage,
			Marche: bandeDe(p, "marche", boite),
			Repos:  bandeDe(p, "repos", boite),
			Touche: bandeDe(p, "touche", boite),
			Meurt:  bandeDe(p, "meurt", boite),
		})
	}
	return out
}

// actionsApercu sont les poses que la fenetre de reglages sait jouer.
var actionsApercu = []string{"marche", "repos", "touche", "meurt"}

// boiteAnimee cherche le rectangle qui contient le singe dans toutes ses poses,
// en coordonnees locales a une case. Les cases ont de larges marges
// transparentes : sans ce recadrage, l'apercu montrerait surtout du vide.
func boiteAnimee(p *planche.Planche, actions []string) image.Rectangle {
	var b image.Rectangle
	for _, nom := range actions {
		a := p.Obtenir(nom, "droite")
		if a == nil {
			continue
		}
		for _, img := range a.Images {
			b = b.Union(boiteOpaque(img))
		}
	}
	if b.Empty() {
		return image.Rect(0, 0, p.Largeur, p.Hauteur)
	}
	return b
}

// boiteOpaque renvoie les limites des pixels visibles d'une image, ramenees a
// son coin superieur gauche.
func boiteOpaque(img image.Image) image.Rectangle {
	d := img.Bounds()
	minX, minY, maxX, maxY := d.Dx(), d.Dy(), 0, 0
	for y := d.Min.Y; y < d.Max.Y; y++ {
		for x := d.Min.X; x < d.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a > 2048 {
				lx, ly := x-d.Min.X, y-d.Min.Y
				if lx < minX {
					minX = lx
				}
				if lx >= maxX {
					maxX = lx + 1
				}
				if ly < minY {
					minY = ly
				}
				if ly >= maxY {
					maxY = ly + 1
				}
			}
		}
	}
	if minX >= maxX || minY >= maxY {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX, maxY)
}

// bandeDe met une animation a plat, recadree sur boite. Obtenir se rabat sur le
// repos quand la planche ne connait pas l'action, donc il y a toujours quelque
// chose a voir.
func bandeDe(p *planche.Planche, action string, boite image.Rectangle) bande {
	a := p.Obtenir(action, "droite")
	if a == nil || len(a.Images) == 0 {
		return bande{}
	}
	l, h := boite.Dx(), boite.Dy()
	bnd := image.NewRGBA(image.Rect(0, 0, l*len(a.Images), h))
	for i, img := range a.Images {
		draw.Draw(bnd, image.Rect(i*l, 0, i*l+l, h), img,
			img.Bounds().Min.Add(boite.Min), draw.Src)
	}
	return bande{
		Png:    "data:image/png;base64," + versPngBase64(bnd),
		Cadres: len(a.Images),
		MS:     a.MS,
	}
}

// coeurBase64 renvoie l'image d'un coeur (plein ou vide) telle que le singe la
// porte au-dessus de la tete.
func coeurBase64(plein bool) string { return versPngBase64(coeurs.Icone(plein, 4)) }

// policePixel renvoie la police des titres, embarquee dans le binaire : la
// fenetre de reglages doit rester belle sans connexion.
func policePixel() string {
	brut, err := ressources.Fichiers.ReadFile("assets/pixel.woff2")
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(brut)
}

func versPngBase64(img image.Image) string {
	var tampon bytes.Buffer
	if err := png.Encode(&tampon, img); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(tampon.Bytes())
}

// humeurs publie l'etat des jauges pour la page de reglages : elle les lit a
// intervalle regulier et les affiche en direct, ce ne sont pas des chiffres
// inventes.
var humeurs atomic.Value // []vie.Jauge

func publierHumeurs(j []vie.Jauge) { humeurs.Store(j) }

func humeursActuelles() []vie.Jauge {
	if v, ok := humeurs.Load().([]vie.Jauge); ok {
		return v
	}
	return nil
}

// ouvrirDossierConfig montre le fichier de configuration dans le gestionnaire
// de fichiers du systeme.
func ouvrirDossierConfig() {
	d := dossierConfig()
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", d).Start()
	case "windows":
		_ = exec.Command("explorer", d).Start()
	default:
		_ = exec.Command("xdg-open", d).Start()
	}
}

// redemarrer relance le meme executable et s'efface : c'est ce qui applique
// les nouveaux reglages, taille comprise.
func redemarrer() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("redemarrage impossible : %v", err)
		return
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	if err := cmd.Start(); err != nil {
		log.Printf("redemarrage : %v", err)
		return
	}
	os.Exit(0)
}

func valeurOu(v, defaut float64) float64 {
	if v <= 0 {
		return defaut
	}
	return v
}

func valeurTexteOu(v, defaut string) string {
	if v == "" {
		return defaut
	}
	return v
}

// majPrete est posee quand une nouvelle version a ete telechargee et mise en
// place : la boucle principale redemarre alors le singe.
var majPrete atomic.Bool

// surveillerMisesAJour verifie la derniere release a intervalles reguliers et
// prepare la mise a jour quand une version plus recente existe.
func surveillerMisesAJour() {
	time.Sleep(30 * time.Second) // laisser le demarrage respirer
	for {
		if tag, url, ok := maj.Verifier(version); ok {
			log.Printf("mise a jour %s disponible, telechargement", tag)
			if err := maj.Appliquer(url); err != nil {
				log.Printf("mise a jour : %v", err)
			} else {
				majPrete.Store(true)
				return // la boucle principale redemarre
			}
		}
		time.Sleep(24 * time.Hour)
	}
}
