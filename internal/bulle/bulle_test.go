package bulle

import (
	"image"
	"image/png"
	"os"
	"testing"
)

// TestRendu ecrit un apercu de la bulle sur fond uni, pour verifier a l'oeil
// les couleurs, les bords en escalier et le centrage du texte.
func TestRendu(t *testing.T) {
	b, err := Nouvelle(2)
	if err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 300, 120))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 0x20, 0x30, 0x60, 0xff
	}
	b.Dessiner(img, "Me revoilà !", 20, 10, 1)

	larg, haut := b.Taille("Me revoilà !")
	if larg <= 0 || haut <= 0 {
		t.Fatalf("taille invalide : %dx%d", larg, haut)
	}
	// le milieu du corps doit avoir pris la couleur de fond de la bulle
	if c := img.RGBAAt(20+larg/2, 10+haut/3); c != couleurFond {
		t.Errorf("interieur = %v, attendu %v", c, couleurFond)
	}

	f, err := os.Create("/tmp/bulle-test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// TestEmojiIgnores verifie que les caracteres absents de la police sont retires
// plutot que rendus en carre vide.
func TestEmojiIgnores(t *testing.T) {
	b, err := Nouvelle(2)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b.nettoyer("Coucou toi 🐒"), "Coucou toi"; got != want {
		t.Errorf("nettoyer = %q, attendu %q", got, want)
	}
}

// TestPoliceAccents verifie que la police bitmap couvre le francais.
func TestPoliceAccents(t *testing.T) {
	b, err := Nouvelle(2)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range "àâäéèêëîïôöùûüçÀÉÊÇ'…!?" {
		if _, ok := b.face.GlyphAdvance(r); !ok {
			t.Errorf("glyphe manquant : %q", r)
		}
	}
}
