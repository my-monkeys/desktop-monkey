//go:build darwin

// Package calque affiche une image a canal alpha directement sur le bureau.
//
// Sous macOS, on utilise une NSWindow sans bordure a fond transparent, posee
// au-dessus des autres fenetres et traversante. C'est l'equivalent de la
// fenetre en couche Windows. Le detail Cocoa vit dans darwin_bridge.m ; ce
// fichier ne fait que l'habiller de l'interface commune Calque.
package calque

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "darwin_bridge.h"
*/
import "C"

import (
	"fmt"
	"image"
	"sync"
	"unsafe"
)

// L'application Cocoa ne s'initialise qu'une fois, meme si plusieurs fenetres
// sont ouvertes (le singe et chaque crotte).
var initCocoa sync.Once

// Calque est une fenetre sans bordure, toujours visible et traversante, dont le
// contenu est une image a canal alpha.
type Calque struct {
	win        unsafe.Pointer
	larg, haut int
	ecranH     int
}

// Ouvrir cree la fenetre. A appeler depuis le fil principal (Cocoa l'exige) :
// c'est garanti par le runtime.LockOSThread pose des l'init du programme.
func Ouvrir(nom string, larg, haut int) (*Calque, error) {
	initCocoa.Do(func() { C.calque_init() })

	win := C.calque_ouvrir(C.int(larg), C.int(haut))
	if win == nil {
		return nil, fmt.Errorf("creation de la fenetre macOS impossible")
	}
	var w, h C.int
	C.calque_ecran(&w, &h)
	return &Calque{win: unsafe.Pointer(win), larg: larg, haut: haut, ecranH: int(h)}, nil
}

// Afficher place la fenetre en (x, y) et y peint l'image.
func (c *Calque) Afficher(img *image.RGBA, x, y int) error {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > c.larg || h > c.haut {
		return fmt.Errorf("image %dx%d plus grande que le calque %dx%d", w, h, c.larg, c.haut)
	}
	C.calque_afficher(c.win, (*C.uchar)(unsafe.Pointer(&img.Pix[0])),
		C.int(w), C.int(h), C.int(x), C.int(y), C.int(c.ecranH))
	return nil
}

// PasserDevant ramene la fenetre du singe devant les crottes (creees apres).
func (c *Calque) PasserDevant() { C.calque_devant(c.win) }

// Traversant regle si la fenetre laisse passer les clics. On la rend
// capturante uniquement quand le curseur est sur un pixel dessine, pour
// absorber le clic sur le sprite sans bloquer le bureau ailleurs. Windows le
// fait au pixel pres tout seul ; macOS a besoin de ce coup de main.
func (c *Calque) Traversant(oui bool) {
	if c.win != nil {
		C.calque_traversant(c.win, boolToInt(oui))
	}
}

func boolToInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

// TraiterMessages vide la file d'evenements de l'application. A appeler a
// chaque tour de boucle. Renvoie toujours true (la fermeture passe par la fin
// de la boucle principale, pas par un evenement).
func (c *Calque) TraiterMessages() bool { return C.calque_pump() != 0 }

// Fermer detruit la fenetre.
func (c *Calque) Fermer() {
	if c.win != nil {
		C.calque_fermer(c.win)
		c.win = nil
	}
}
