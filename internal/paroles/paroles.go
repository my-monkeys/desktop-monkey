// Package paroles choisit ce que le singe raconte.
//
// Les phrases sont rangees par contexte dans un fichier JSON embarque, que
// l'utilisateur peut remplacer par le sien sans recompiler l'application.
package paroles

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"time"
)

// Recueil regroupe les phrases par contexte.
type Recueil struct {
	phrases map[string][]string
	alea    *rand.Rand
	dernier map[string]string // evite de repeter deux fois la meme phrase
}

// Charger lit le recueil embarque, puis le remplace par le fichier utilisateur
// s'il existe. Un fichier utilisateur invalide est ignore plutot que de rendre
// le singe muet.
func Charger(fsys fs.FS, chemin, chemESurcharge string) (*Recueil, error) {
	brut, err := fs.ReadFile(fsys, chemin)
	if err != nil {
		return nil, err
	}
	var m map[string][]string
	if err := json.Unmarshal(sansBOM(brut), &m); err != nil {
		return nil, err
	}

	if chemESurcharge != "" {
		if perso, err := os.ReadFile(chemESurcharge); err == nil {
			var mp map[string][]string
			if err := json.Unmarshal(sansBOM(perso), &mp); err == nil {
				m = mp
			} else {
				log.Printf("phrases.json perso ignore (%v)", err)
			}
		}
	}

	return &Recueil{
		phrases: m,
		alea:    rand.New(rand.NewSource(time.Now().UnixNano())),
		dernier: map[string]string{},
	}, nil
}

// sansBOM retire le BOM UTF-8 que le Bloc-notes de Windows ajoute volontiers.
func sansBOM(b []byte) []byte {
	return bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))
}

// Pour renvoie une phrase du contexte demande, en retombant sur "hasard" si le
// contexte n'existe pas. Renvoie false s'il n'y a rien a dire.
func (r *Recueil) Pour(contexte string) (string, bool) {
	liste := r.phrases[contexte]
	if len(liste) == 0 {
		liste = r.phrases["hasard"]
	}
	if len(liste) == 0 {
		return "", false
	}
	if len(liste) == 1 {
		return liste[0], true
	}
	// tire jusqu'a obtenir autre chose que la phrase precedente
	for i := 0; i < 6; i++ {
		p := liste[r.alea.Intn(len(liste))]
		if p != r.dernier[contexte] {
			r.dernier[contexte] = p
			return p, true
		}
	}
	return liste[r.alea.Intn(len(liste))], true
}

// SelonHeure renvoie le contexte adapte au moment de la journee.
func SelonHeure(t time.Time) string {
	switch h := t.Hour(); {
	case h < 5:
		return "nuit"
	case h < 11:
		return "matin"
	case h < 18:
		return "hasard"
	case h < 23:
		return "soir"
	default:
		return "nuit"
	}
}
