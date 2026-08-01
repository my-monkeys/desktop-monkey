//go:build !windows

package tray

import "image"

// Sur les plateformes autres que Windows, le tray n'est pas encore implemente
// (macOS aura une icone dans la barre des menus). Ces fonctions ne font rien,
// pour que l'application compile et se lance sans zone de notification.

// Nouvelle installe l'icone du tray (sans effet ici).
func Nouvelle(nomApp, exe, config string, icone *image.RGBA) error { return nil }

// Fermer retire l'icone (sans effet ici).
func Fermer() {}
