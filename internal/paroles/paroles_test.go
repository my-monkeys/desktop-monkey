package paroles

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/my-monkeys/singe-de-bureau/internal/ressources"
)

func TestSurchargeAvecBOM(t *testing.T) {
	chemin := filepath.Join(t.TempDir(), "phrases.json")
	contenu := "\xef\xbb\xbf" + `{"bonjour": ["Salut BOM"]}`
	if err := os.WriteFile(chemin, []byte(contenu), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := Charger(ressources.Fichiers, "assets/phrases.json", chemin)
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	p, ok := r.Pour("bonjour")
	if !ok || p != "Salut BOM" {
		t.Fatalf("phrase %q, attendu la surcharge malgre le BOM", p)
	}
}

func TestSurchargeInvalideIgnoree(t *testing.T) {
	chemin := filepath.Join(t.TempDir(), "phrases.json")
	if err := os.WriteFile(chemin, []byte("pas du json"), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := Charger(ressources.Fichiers, "assets/phrases.json", chemin)
	if err != nil {
		t.Fatalf("un fichier perso invalide ne doit pas faire echouer le chargement : %v", err)
	}
	if _, ok := r.Pour("bonjour"); !ok {
		t.Fatal("les phrases embarquees devraient rester disponibles")
	}
}
