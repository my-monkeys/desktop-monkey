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

// Boite englobante de tous les ecrans : origine (ox,oy) en haut a gauche de
// l'union, dans le repere global de CoreGraphics. Le singe vit dans cet espace
// unifie, donc il peut passer d'un ecran a l'autre et suivre le curseur ou
// qu'il soit — au lieu d'etre coince sur l'ecran principal.
static void souris_union(int* ox, int* oy, int* w, int* h) {
    uint32_t n = 0;
    CGGetActiveDisplayList(0, NULL, &n);
    if (n == 0) { *ox = *oy = 0; *w = *h = 0; return; }
    if (n > 16) n = 16;
    CGDirectDisplayID ids[16];
    CGGetActiveDisplayList(n, ids, &n);
    CGRect u = CGDisplayBounds(ids[0]);
    for (uint32_t i = 1; i < n; i++) u = CGRectUnion(u, CGDisplayBounds(ids[i]));
    *ox = (int)u.origin.x;
    *oy = (int)u.origin.y;
    *w  = (int)u.size.width;
    *h  = (int)u.size.height;
}
static void souris_pos(int* x, int* y) {
    int ox, oy, w, h;
    souris_union(&ox, &oy, &w, &h);
    CGEventRef e = CGEventCreate(NULL);
    CGPoint p = CGEventGetLocation(e);
    CFRelease(e);
    *x = (int)p.x - ox;
    *y = (int)p.y - oy;
}
static int souris_bouton(void) {
    return CGEventSourceButtonState(kCGEventSourceStateCombinedSessionState,
                                    kCGMouseButtonLeft) ? 1 : 0;
}
static void souris_taille(int* w, int* h) {
    int ox, oy;
    souris_union(&ox, &oy, w, h);
}
static void souris_placer(int x, int y) {
    int ox, oy, w, h;
    souris_union(&ox, &oy, &w, &h);
    CGWarpMouseCursorPosition(CGPointMake(x + ox, y + oy));
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
