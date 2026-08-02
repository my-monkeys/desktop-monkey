//go:build darwin

// Package dialogue affiche la fenetre native de reglages : une vraie boite de
// dialogue Cocoa a onglets (NSTabView), decrite en JSON par l'appelant. Le
// detail AppKit vit dans darwin_bridge.m.
package dialogue

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "darwin_bridge.h"
*/
import "C"

import "unsafe"

// Ouvrir affiche la fenetre construite depuis la description JSON (voir le
// bridge pour le format). Sans effet si elle est deja ouverte.
func Ouvrir(descJSON string) {
	c := C.CString(descJSON)
	defer C.free(unsafe.Pointer(c))
	C.dialogue_ouvrir(c)
}

// Resultat renvoie les valeurs enregistrees (JSON cle -> valeur), une seule
// fois : les appels suivants renvoient false jusqu'au prochain enregistrement.
func Resultat() (string, bool) {
	p := C.dialogue_resultat()
	if p == nil {
		return "", false
	}
	s := C.GoString(p)
	C.free(unsafe.Pointer(p))
	return s, true
}

// Disponible indique que la plateforme offre le dialogue natif.
func Disponible() bool { return true }
