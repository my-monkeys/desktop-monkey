//go:build !windows && !darwin

package tray

import "image"

// Sur les plateformes autres que Windows, le tray n'est pas encore implemente
// (Linux, etc.). Ces fonctions ne font rien, pour que l'application compile.

// Nouvelle installe l'icone du tray (sans effet ici).
func Nouvelle(nomApp, exe, config string, icone *image.RGBA) error { return nil }

// Fermer retire l'icone (sans effet ici).
func Fermer() {}

// QuitDemande indique si l'utilisateur a demande a quitter via le menu.
func QuitDemande() bool { return false }
