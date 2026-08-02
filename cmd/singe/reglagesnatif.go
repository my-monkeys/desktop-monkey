package main

// Ce que la page de reglages a besoin pour se dessiner : l'apercu anime du
// singe, ses coeurs, la police pixel — plus l'ecriture de la configuration et
// le redemarrage qui applique le tout.

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/my-monkeys/desktop-monkey/internal/coeurs"
	"github.com/my-monkeys/desktop-monkey/internal/maj"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
)

// apercu decrit la bande d'images de marche envoyee a la page : les cadres y
// sont poses cote a cote, la page les fait defiler.
type apercu struct {
	Png    string // la bande entiere, PNG en base64
	Cadres int    // nombre d'images dans la bande
	CelL   int    // largeur d'une image, en pixels physiques par unite de taille
	CelH   int    // sa hauteur
	MS     int    // duree d'affichage d'une image
}

// apercuMarche compose la bande de l'animation de marche. Les dimensions sont
// donnees en pixels physiques pour une taille de 1 : la page les divise par la
// densite de son ecran, donc l'apercu fait la taille reelle.
func apercuMarche(chemin string) (apercu, bool) {
	p, err := planche.Charger(ressources.Fichiers, chemin, 1)
	if err != nil {
		return apercu{}, false
	}
	a := p.Obtenir("marche", "droite")
	if a == nil || len(a.Images) == 0 {
		return apercu{}, false
	}

	l, h := p.Largeur, p.Hauteur
	bande := image.NewRGBA(image.Rect(0, 0, l*len(a.Images), h))
	for i, img := range a.Images {
		draw.Draw(bande, image.Rect(i*l, 0, i*l+l, h), img, image.Point{}, draw.Src)
	}
	return apercu{
		Png:    versPngBase64(bande),
		Cadres: len(a.Images),
		CelL:   l * facteurAffichage,
		CelH:   h * facteurAffichage,
		MS:     a.MS,
	}, true
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
