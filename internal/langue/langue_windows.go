//go:build windows

// Package langue detecte la langue de l'utilisateur, pour choisir le recueil
// de phrases embarque.
package langue

import "syscall"

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	procUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
)

// Francais indique si l'interface de l'utilisateur est en francais.
func Francais() bool {
	r, _, _ := procUILanguage.Call()
	return r&0x3ff == 0x0c // LANG_FRENCH, quel que soit le pays
}
