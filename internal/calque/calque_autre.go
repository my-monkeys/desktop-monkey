//go:build !windows && !darwin

package calque

import (
	"errors"
	"image"
)

// L'affichage sur le bureau est implemente pour Windows et macOS (fichiers
// dedies). Sur les autres plateformes (Linux, etc.), il n'y a pas encore de
// support ; le reste du programme (comportement, sprites, bulles) est
// independant de la plateforme.

type Calque struct{}

// Ouvrir renvoie une erreur explicite sur les plateformes non supportees.
func Ouvrir(nom string, larg, haut int) (*Calque, error) {
	return nil, errors.New("affichage bureau non implemente sur cette plateforme (Windows et macOS uniquement)")
}

func (c *Calque) Afficher(img *image.RGBA, x, y int) error { return nil }
func (c *Calque) Traversant(oui bool)                      {}
func (c *Calque) TraiterMessages() bool                    { return false }
func (c *Calque) Fermer()                                  {}
