//go:build !windows

package calque

import (
	"errors"
	"image"
)

// L'affichage sur le bureau est pour l'instant implemente uniquement sous
// Windows. Sur macOS il passera par une NSWindow a fond transparent ; le reste
// du programme (comportement, sprites, bulles) est deja independant de la
// plateforme et n'aura pas a changer.

type Calque struct{}

// Ouvrir renvoie une erreur explicite sur les plateformes non supportees.
func Ouvrir(nom string, larg, haut int) (*Calque, error) {
	return nil, errors.New("affichage bureau non implemente sur cette plateforme (Windows uniquement pour l'instant)")
}

func (c *Calque) Afficher(img *image.RGBA, x, y int) error { return nil }
func (c *Calque) TraiterMessages() bool                    { return false }
func (c *Calque) Fermer()                                  {}
