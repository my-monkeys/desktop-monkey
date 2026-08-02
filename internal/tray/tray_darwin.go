//go:build darwin

// Package tray ajoute une icone du singe dans la barre des menus de macOS, avec
// un petit menu : lancer au demarrage, ouvrir les reglages, quitter. Le detail
// Cocoa vit dans tray_darwin_bridge.m ; ce fichier l'habille de l'interface
// commune (la meme que sous Windows).
package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "tray_darwin_bridge.h"
*/
import "C"

import (
	"image"
	"strings"
	"unsafe"
)

// Nouvelle installe l'icone et son menu. exe est le chemin de l'executable
// (pour le lancement au demarrage), config le fichier de reglages a ouvrir, et
// icone une image RGBA (l'icone du singe) — nil pour un emoji de repli.
func Nouvelle(nomApp, exe, config string, icone *image.RGBA) error {
	cNom := C.CString(nomApp)
	cExe := C.CString(exe)
	cCfg := C.CString(config)
	defer C.free(unsafe.Pointer(cNom))
	defer C.free(unsafe.Pointer(cExe))
	defer C.free(unsafe.Pointer(cCfg))

	var pix *C.uchar
	var w, h C.int
	if icone != nil {
		b := icone.Bounds()
		w, h = C.int(b.Dx()), C.int(b.Dy())
		pix = (*C.uchar)(unsafe.Pointer(&icone.Pix[0]))
	}
	C.tray_init(cNom, cExe, cCfg, pix, w, h)
	return nil
}

// QuitDemande indique si l'utilisateur a choisi Quitter dans le menu. La boucle
// principale l'interroge a chaque image pour s'arreter proprement.
func QuitDemande() bool { return C.tray_quit_requested() != 0 }

// MajJauges remplace les lignes d'humeur montrees en tete du menu (elles
// apparaissent a sa prochaine ouverture).
func MajJauges(lignes []string) {
	c := C.CString(strings.Join(lignes, "\n"))
	defer C.free(unsafe.Pointer(c))
	C.tray_maj_jauges(c)
}

// ReglagesDemande indique que "Open settings" vient d'etre choisi, et consomme
// la demande : la boucle principale ouvre alors le dialogue natif.
func ReglagesDemande() bool { return C.tray_reglages_requested() != 0 }

// Fermer retire l'icone de la barre des menus.
func Fermer() { C.tray_fermer() }

// AuDemarrage indique si le singe est lance a l'ouverture de session.
func AuDemarrage() bool { return C.tray_au_demarrage() != 0 }

// DefinirDemarrage active ou desactive le lancement a l'ouverture de session.
func DefinirDemarrage(actif bool) {
	v := C.int(0)
	if actif {
		v = 1
	}
	C.tray_definir_demarrage(v)
}
