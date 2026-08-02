//go:build !darwin && !windows

// Sur les autres plateformes, pas de fenetre : ces fonctions ne font rien.
package dialogue

// Ouvrir n'a pas d'effet ici.
func Ouvrir(url, titre string, larg, haut int) {}

// Fermer n'a pas d'effet ici.
func Fermer() {}

// Disponible indique que la plateforme ne sait pas afficher la fenetre.
func Disponible() bool { return false }
