//go:build darwin

// Package souris lit l'etat physique de la souris a l'echelle de tout l'ecran.
//
// Sous macOS, tout passe par CoreGraphics (Quartz) : position du curseur, etat
// du bouton, dimensions de l'ecran, et deplacement force du curseur (pour que
// le singe puisse voler la fleche). Les coordonnees sont en points, origine en
// haut a gauche de l'ecran principal, y vers le bas — la meme convention que
// sous Windows, donc le reste du programme n'a rien a changer.
package souris

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

static void souris_pos(int* x, int* y) {
    CGEventRef e = CGEventCreate(NULL);
    CGPoint p = CGEventGetLocation(e);
    CFRelease(e);
    *x = (int)p.x;
    *y = (int)p.y;
}
static int souris_bouton(void) {
    return CGEventSourceButtonState(kCGEventSourceStateCombinedSessionState,
                                    kCGMouseButtonLeft) ? 1 : 0;
}
static void souris_taille(int* w, int* h) {
    CGRect b = CGDisplayBounds(CGMainDisplayID());
    *w = (int)b.size.width;
    *h = (int)b.size.height;
}
static void souris_placer(int x, int y) {
    CGWarpMouseCursorPosition(CGPointMake(x, y));
    // sans cela, un court delai ignore les mouvements physiques apres un warp
    CGAssociateMouseAndMouseCursorPosition(true);
}
*/
import "C"

// Position renvoie les coordonnees du curseur en points ecran.
func Position() (int, int) {
	var x, y C.int
	C.souris_pos(&x, &y)
	return int(x), int(y)
}

// BoutonGauche indique si le bouton gauche est physiquement enfonce, meme si le
// clic est destine a une autre fenetre.
func BoutonGauche() bool { return C.souris_bouton() != 0 }

// TailleEcran renvoie les dimensions de l'ecran principal, en points.
func TailleEcran() (int, int) {
	var w, h C.int
	C.souris_taille(&w, &h)
	return int(w), int(h)
}

// BasTravail renvoie le sol du cadavre. macOS n'a pas de barre des taches en
// bas de l'ecran : le sol est le bas de l'ecran.
func BasTravail() int {
	_, h := TailleEcran()
	return h
}

// Placer deplace le curseur systeme. C'est ce qui permet au singe de voler la
// fleche et de s'enfuir avec.
func Placer(x, y int) { C.souris_placer(C.int(x), C.int(y)) }

// Disponible indique si la lecture globale de la souris fonctionne.
func Disponible() bool { return true }
