package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Les cases des planches ont de larges marges transparentes. L'apercu doit les
// retirer, sinon le singe apparait minuscule au milieu de rien.
func TestBandesRecadreesSurLeCorps(t *testing.T) {
	planches := planchesWeb(true)
	if len(planches) != 2 {
		t.Fatalf("2 planches attendues, %d obtenues", len(planches))
	}
	for _, p := range planches {
		if p.CelL <= 0 || p.CelH <= 0 {
			t.Errorf("%s : case vide (%dx%d)", p.Cle, p.CelL, p.CelH)
		}
		// toutes les poses sont jouables : Obtenir se rabat sur le repos
		for nom, b := range map[string]bande{
			"marche": p.Marche, "repos": p.Repos, "touche": p.Touche, "meurt": p.Meurt,
		} {
			if b.Cadres == 0 {
				t.Errorf("%s : la pose %q n'a aucune image", p.Cle, nom)
			}
			if !strings.HasPrefix(b.Png, "data:image/png;base64,") {
				t.Errorf("%s : la pose %q n'est pas une image embarquee", p.Cle, nom)
			}
		}
		if p.Pied < 0 {
			t.Errorf("%s : ecart au sol negatif (%d)", p.Cle, p.Pied)
		}
	}
}

// La page se rend, et embarque la configuration que la fenetre va manipuler.
func TestPageReglagesSeRend(t *testing.T) {
	rec := httptest.NewRecorder()
	pageReglages(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("code %d", rec.Code)
	}
	corps := rec.Body.String()
	for _, attendu := range []string{"chance_jet_crotte", "secondes_entre_paroles", "data:font/woff2"} {
		if !strings.Contains(corps, attendu) {
			t.Errorf("la page ne contient pas %q", attendu)
		}
	}
}

// L'aller-retour ne doit rien perdre : ce que la page renvoie se repose sur la
// configuration du disque sans ecraser ce qu'elle n'expose pas.
func TestEnvoiConserveLesValeurs(t *testing.T) {
	r := reglagesParDefaut()
	r.Taille = 1.6
	r.EntreParoles = [2]float64{30, 90}
	r.JetMode = 1

	brut, err := json.Marshal(envoiDepuis(r, true))
	if err != nil {
		t.Fatal(err)
	}
	var e envoi
	if err := json.Unmarshal(brut, &e); err != nil {
		t.Fatal(err)
	}
	if got := e.appliquer(reglagesParDefaut()); got != r {
		t.Errorf("aller-retour perdu :\n reçu   %+v\n attendu %+v", got, r)
	}
	if !e.Demarrage {
		t.Error("le lancement au demarrage n'a pas suivi")
	}
}
