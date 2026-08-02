package main

// Construction de la fenetre native de reglages (internal/dialogue) : la
// description des onglets part d'ici, et l'enregistrement reecrit config.json
// puis relance le singe. La ou le dialogue natif n'existe pas (Windows pour
// l'instant), le menu ouvre la page web a la place.

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/my-monkeys/desktop-monkey/internal/dialogue"
)

type champNatif struct {
	Type     string   `json:"type"`
	Cle      string   `json:"cle"`
	Nom      string   `json:"nom"`
	Aide     string   `json:"aide,omitempty"`
	Min      float64  `json:"min,omitempty"`
	Max      float64  `json:"max,omitempty"`
	Pas      float64  `json:"pas,omitempty"`
	Valeur   float64  `json:"valeur,omitempty"`
	Options  []string `json:"options,omitempty"`
	Libelles []string `json:"libelles,omitempty"`
	Texte    string   `json:"texte,omitempty"`
	Coche    bool     `json:"coche,omitempty"`
}

type ongletNatif struct {
	Titre  string       `json:"titre"`
	Champs []champNatif `json:"champs"`
}

type descNative struct {
	Titre       string        `json:"titre"`
	Enregistrer string        `json:"enregistrer"`
	Annuler     string        `json:"annuler"`
	Onglets     []ongletNatif `json:"onglets"`
}

// ouvrirReglagesNatifs affiche la fenetre, pre-remplie de la config actuelle.
func ouvrirReglagesNatifs() {
	r := chargerReglages()
	fr := enFrancais(r.Langue)
	nom := func(frTxt, enTxt string) string {
		if fr {
			return frTxt
		}
		return enTxt
	}

	curseur := func(cle, frTxt, enTxt, aideFr, aideEn string, min, max, pas, val float64) champNatif {
		return champNatif{Type: "curseur", Cle: cle, Nom: nom(frTxt, enTxt),
			Aide: nom(aideFr, aideEn), Min: min, Max: max, Pas: pas, Valeur: val}
	}

	d := descNative{
		Titre:       nom("Réglages du singe", "Monkey settings"),
		Enregistrer: nom("Enregistrer et redémarrer", "Save and restart"),
		Annuler:     nom("Annuler", "Cancel"),
		Onglets: []ongletNatif{
			{
				Titre: nom("Apparence", "Appearance"),
				Champs: []champNatif{
					curseur("taille", "Taille", "Size", "", "", 0.5, 2, 0.05, valeurOu(r.Taille, 1)),
					curseur("vitesse", "Vitesse", "Speed", "", "", 0.5, 6, 0.1, r.Vitesse),
					{Type: "choix", Cle: "langue", Nom: nom("Langue", "Language"),
						Options:  []string{"auto", "fr", "en"},
						Libelles: []string{nom("auto (système)", "auto (system)"), "français", "English"},
						Texte:    valeurTexteOu(r.Langue, "auto")},
					{Type: "case", Cle: "parle", Nom: nom("Bulles de dialogue", "Speech bubbles"), Coche: r.Parle},
				},
			},
			{
				Titre: nom("Vie", "Life"),
				Champs: []champNatif{
					{Type: "entier", Cle: "coeurs", Nom: nom("Cœurs (clics avant K.O.)", "Hearts (clicks before K.O.)"),
						Min: 1, Max: 9, Pas: 1, Valeur: float64(r.Coeurs)},
					curseur("secondes_avant_sieste", "Sieste après (secondes)", "Nap after (seconds)",
						"", "", 10, 600, 5, r.AvantSieste),
				},
			},
			{
				Titre: nom("Caractère", "Character"),
				Champs: []champNatif{
					curseur("chance_ami", "Ami du curseur", "Cursor friend",
						"envie d'aller le voir", "urge to go see it", 0, 1, 0.05, r.ChanceAmi),
					curseur("chance_chasse", "Chasse", "Hunting",
						"poursuites qui tournent à l'attaque", "pursuits that turn into attacks", 0, 1, 0.05, r.ChanceChasse),
					curseur("chance_vol", "Vol du curseur", "Cursor theft",
						"attaques qui finissent en rapt", "attacks that end in a snatch", 0, 1, 0.05, r.ChanceVol),
					curseur("chance_grimpe", "Escalade", "Climbing",
						"envie de grimper aux bords", "urge to climb the edges", 0, 1, 0.05, r.ChanceGrimpe),
					curseur("chance_jeu", "Jeu", "Play",
						"envie de sautiller", "urge to bounce around", 0, 1, 0.05, r.ChanceJeu),
					curseur("chance_crotte", "Crottes", "Poops",
						"besoin pressant (après un repas)", "pressing need (after a meal)", 0, 1, 0.05, r.ChanceCrotte),
					curseur("chance_jet_crotte", "Jet de crottes", "Poop throwing",
						"corvées de nettoyage", "cleanup chores", 0, 2, 0.1, r.ChanceJetCrotte),
				},
			},
		},
	}

	brut, err := json.Marshal(d)
	if err != nil {
		log.Printf("description des reglages : %v", err)
		return
	}
	dialogue.Ouvrir(string(brut))
}

// recolterReglagesNatifs applique un eventuel enregistrement du dialogue :
// reecrit config.json (cles non exposees preservees) puis relance le singe.
func recolterReglagesNatifs() {
	brut, ok := dialogue.Resultat()
	if !ok {
		return
	}
	var vals map[string]any
	if err := json.Unmarshal([]byte(brut), &vals); err != nil {
		log.Printf("reglages enregistres illisibles : %v", err)
		return
	}

	r := chargerReglages()
	lireF := func(cle string, dst *float64) {
		if v, ok := vals[cle].(float64); ok {
			*dst = v
		}
	}
	lireF("taille", &r.Taille)
	lireF("vitesse", &r.Vitesse)
	lireF("secondes_avant_sieste", &r.AvantSieste)
	lireF("chance_ami", &r.ChanceAmi)
	lireF("chance_chasse", &r.ChanceChasse)
	lireF("chance_vol", &r.ChanceVol)
	lireF("chance_grimpe", &r.ChanceGrimpe)
	lireF("chance_jeu", &r.ChanceJeu)
	lireF("chance_crotte", &r.ChanceCrotte)
	lireF("chance_jet_crotte", &r.ChanceJetCrotte)
	if v, ok := vals["coeurs"].(float64); ok {
		r.Coeurs = int(v)
	}
	if l, ok := vals["langue"].(string); ok && (l == "auto" || l == "fr" || l == "en") {
		r.Langue = l
	}
	if p, ok := vals["parle"].(bool); ok {
		r.Parle = p
	}

	chemin := filepath.Join(dossierConfig(), "config.json")
	sortie, err := json.MarshalIndent(r, "", "  ")
	if err == nil {
		err = os.WriteFile(chemin, sortie, 0o644)
	}
	if err != nil {
		log.Printf("ecriture des reglages : %v", err)
		return
	}
	log.Printf("reglages enregistres, redemarrage")
	redemarrer()
}

func valeurTexteOu(v, defaut string) string {
	if v == "" {
		return defaut
	}
	return v
}
