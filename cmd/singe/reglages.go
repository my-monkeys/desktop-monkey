package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Reglages regroupe ce que l'utilisateur peut modifier sans recompiler.
type Reglages struct {
	Planche         string     `json:"planche"`
	Langue          string     `json:"langue"`
	Taille          float64    `json:"taille"`
	EchelleBulle    int        `json:"echelle_bulle"`
	Vitesse         float64    `json:"vitesse"`
	DistArret       float64    `json:"distance_arret"`
	AvantSieste     float64    `json:"secondes_avant_sieste"`
	ChanceRepas     float64    `json:"chance_repas"`
	ChanceAmi       float64    `json:"chance_ami"`
	SeuilVieSeule   float64    `json:"secondes_avant_vie_seule"`
	ChanceChasse    float64    `json:"chance_chasse"`
	ChanceVol       float64    `json:"chance_vol"`
	ChanceGrimpe    float64    `json:"chance_grimpe"`
	ChanceJeu       float64    `json:"chance_jeu"`
	ChanceCrotte    float64    `json:"chance_crotte"`
	ChanceJetCrotte float64    `json:"chance_jet_crotte"`
	JetMode         int        `json:"jet_mode"` // -1 aleatoire, 0 hors ecran, 1 sur la souris
	Coeurs          int        `json:"coeurs"`
	Resurrection    float64    `json:"secondes_avant_resurrection"`
	CacheApresClic  float64    `json:"cache_apres_clic"`
	Parle           bool       `json:"parle"`
	EntreParoles    [2]float64 `json:"secondes_entre_paroles"`
	DureeBulle      float64    `json:"duree_bulle"`
}

func reglagesParDefaut() Reglages {
	return Reglages{
		Planche:         "assets/singe2.json",
		Langue:          "auto",
		Taille:          1,
		EchelleBulle:    1,
		Vitesse:         2.6,
		DistArret:       64,
		AvantSieste:     180,
		ChanceRepas:     0.18,
		ChanceAmi:       0.85,
		SeuilVieSeule:   10,
		ChanceChasse:    0.35,
		ChanceVol:       0.35,
		ChanceGrimpe:    0.12,
		ChanceJeu:       0.10,
		ChanceCrotte:    0.07,
		ChanceJetCrotte: 0.35,
		JetMode:         -1,
		Coeurs:          3,
		Resurrection:    0,
		CacheApresClic:  0,
		Parle:           true,
		EntreParoles:    [2]float64{90, 240},
		DureeBulle:      6,
	}
}

func dossierConfig() string {
	switch runtime.GOOS {
	case "windows":
		if d := os.Getenv("APPDATA"); d != "" {
			return filepath.Join(d, nomApp)
		}
	case "darwin":
		if d, err := os.UserHomeDir(); err == nil {
			return filepath.Join(d, "Library", "Application Support", nomApp)
		}
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, nomApp)
	}
	return "."
}

func chargerReglages() Reglages {
	r := reglagesParDefaut()
	chemin := filepath.Join(dossierConfig(), "config.json")
	if brut, err := os.ReadFile(chemin); err == nil {
		// le Bloc-notes de Windows ajoute volontiers un BOM que le JSON refuse
		brut = bytes.TrimPrefix(brut, []byte("\xef\xbb\xbf"))
		if err := json.Unmarshal(brut, &r); err != nil {
			log.Printf("config.json ignore (%v)", err)
		}
	}
	return r
}

// ecrireConfigExemple depose un config.json au premier lancement, pour que les
// reglages soient decouvrables sans lire la documentation.
func ecrireConfigExemple() {
	d := dossierConfig()
	if err := os.MkdirAll(d, 0o755); err != nil {
		return
	}
	chemin := filepath.Join(d, "config.json")
	if _, err := os.Stat(chemin); err == nil {
		return
	}
	if brut, err := json.MarshalIndent(reglagesParDefaut(), "", "  "); err == nil {
		_ = os.WriteFile(chemin, brut, 0o644)
	}
}
