// Pont Objective-C vers Cocoa pour la fenetre en couche du singe sous macOS.
// Les fonctions sont declarees en C pur pour etre appelables depuis cgo.
#ifndef SINGE_DARWIN_BRIDGE_H
#define SINGE_DARWIN_BRIDGE_H

// A appeler une fois, sur le fil principal, avant toute fenetre.
void calque_init(void);

// Cree une fenetre sans bordure, transparente, flottante et traversante.
// Renvoie un handle opaque (NSWindow retenu).
void* calque_ouvrir(int w, int h);

// Peint l'image RGBA (alpha premultiplie) dans la fenetre et la place a l'ecran.
// x, y sont en coordonnees ecran haut-gauche ; screenH sert a retourner l'axe Y
// vers la convention bas-gauche de Cocoa.
void calque_afficher(void* win, unsigned char* pix, int w, int h,
                     int x, int y, int screenH);

// Regle si la fenetre laisse passer les clics (traversante) ou les capture.
// On la rend capturante seulement quand le curseur est sur un pixel dessine,
// pour absorber le clic sur le sprite sans bloquer le bureau ailleurs.
void calque_traversant(void* win, int oui);

// Vide la file d'evenements de l'application. Renvoie toujours 1.
int calque_pump(void);

// Detruit une fenetre.
void calque_fermer(void* win);

// Dimensions de l'ecran principal en points.
void calque_ecran(int* w, int* h);

#endif
