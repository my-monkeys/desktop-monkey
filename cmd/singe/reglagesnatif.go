package main

// Outils partages par la page de reglages : l'apercu du singe, l'ecriture de
// la configuration et le redemarrage qui applique tout.

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/my-monkeys/desktop-monkey/internal/maj"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
)

type apercu struct {
	Png          string // PNG du singe au repos, en base64
	BaseL, BaseH int    // pixels physiques par unite de taille
}

// apercuSinge prepare l'image de l'apercu : le singe au repos, en pixels
// physiques par unite de taille (facteur de plateforme compris) — la page la
// remet a l'echelle de l'ecran via devicePixelRatio.
func apercuSinge(chemin string) (apercu, bool) {
	p, err := planche.Charger(ressources.Fichiers, chemin, 1)
	if err != nil {
		return apercu{}, false
	}
	img := p.Obtenir("repos", "droite").Image(0)
	if img == nil {
		return apercu{}, false
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return apercu{}, false
	}
	return apercu{
		Png:   base64.StdEncoding.EncodeToString(buf.Bytes()),
		BaseL: img.Bounds().Dx() * facteurAffichage,
		BaseH: img.Bounds().Dy() * facteurAffichage,
	}, true
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
