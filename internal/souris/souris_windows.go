//go:build windows

// Package souris lit l'etat physique de la souris a l'echelle de tout l'ecran.
//
// La fenetre de l'application laisse passer les clics (elle est "traversante"),
// donc elle ne recoit aucun evenement souris : on interroge directement le
// systeme. Sous Windows cela passe par user32.dll, en appels syscall purs — pas
// de cgo, ce qui permet de compiler l'executable Windows depuis un Mac.
package souris

import (
	"syscall"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetCursorPos  = user32.NewProc("GetCursorPos")
	procGetAsyncKey   = user32.NewProc("GetAsyncKeyState")
	procSystemMetrics = user32.NewProc("GetSystemMetrics")
	procSystemParams  = user32.NewProc("SystemParametersInfoW")
	procSetCursorPos  = user32.NewProc("SetCursorPos")
)

const (
	vkBoutonGauche = 0x01
	smEcranLargeur = 0
	smEcranHauteur = 1
	spiZoneTravail = 0x0030
)

type point struct{ X, Y int32 }

type rectangle struct{ Gauche, Haut, Droite, Bas int32 }

// Position renvoie les coordonnees du curseur en pixels ecran.
func Position() (int, int) {
	var p point
	r, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	if r == 0 {
		return 0, 0
	}
	return int(p.X), int(p.Y)
}

// BoutonGauche indique si le bouton gauche est physiquement enfonce, meme si le
// clic est destine a une autre fenetre.
func BoutonGauche() bool {
	r, _, _ := procGetAsyncKey.Call(vkBoutonGauche)
	return r&0x8000 != 0
}

// Placer deplace le curseur systeme. C'est ce qui permet au singe de voler la
// fleche et de s'enfuir avec.
func Placer(x, y int) {
	procSetCursorPos.Call(uintptr(x), uintptr(y))
}

// TailleEcran renvoie les dimensions de l'ecran principal.
func TailleEcran() (int, int) {
	l, _, _ := procSystemMetrics.Call(smEcranLargeur)
	h, _, _ := procSystemMetrics.Call(smEcranHauteur)
	return int(l), int(h)
}

// BasTravail renvoie l'ordonnee du bas de la zone de travail, c'est-a-dire le
// sommet de la barre des taches quand elle est en bas de l'ecran.
func BasTravail() int {
	var r rectangle
	ret, _, _ := procSystemParams.Call(spiZoneTravail, 0, uintptr(unsafe.Pointer(&r)), 0)
	if ret == 0 {
		_, h := TailleEcran()
		return h
	}
	return int(r.Bas)
}

// Disponible indique si la lecture globale de la souris fonctionne.
func Disponible() bool { return true }
