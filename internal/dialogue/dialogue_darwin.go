//go:build darwin

// Package dialogue affiche la fenetre de reglages : une petite fenetre native
// qui rend la page HTML servie par l'application (la meme interface sur toutes
// les plateformes). Sous macOS, c'est une NSWindow portant une WKWebView.
package dialogue

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit
#include <stdlib.h>
#include "darwin_bridge.h"
*/
import "C"

import "unsafe"

// Ouvrir montre la fenetre sur l'URL donnee ; si elle est deja ouverte, elle
// revient simplement au premier plan.
func Ouvrir(url, titre string, larg, haut int) {
	cu := C.CString(url)
	ct := C.CString(titre)
	defer C.free(unsafe.Pointer(cu))
	defer C.free(unsafe.Pointer(ct))
	C.dialogue_ouvrir(cu, ct, C.int(larg), C.int(haut))
}

// Fermer ferme la fenetre si elle est ouverte.
func Fermer() { C.dialogue_fermer() }

// Disponible indique que la plateforme sait afficher la fenetre.
func Disponible() bool { return true }
