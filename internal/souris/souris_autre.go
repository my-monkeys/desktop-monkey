//go:build !windows && !darwin

package souris

// Repli pour les plateformes ou la lecture globale de la souris n'est pas
// implementee (Linux, etc.). Windows et macOS ont leur propre fichier ;
// l'affichage bureau n'y est de toute facon pas disponible (voir calque).

// Position renvoie les coordonnees du curseur en pixels ecran.
func Position() (int, int) { return 0, 0 }

// BoutonGauche indique si le bouton gauche est physiquement enfonce.
func BoutonGauche() bool { return false }

// Placer deplace le curseur systeme.
func Placer(x, y int) {}

// TailleEcran renvoie les dimensions de l'ecran principal.
func TailleEcran() (int, int) { return 1920, 1080 }

// BasTravail renvoie le bas de la zone de travail (sommet de la barre des
// taches). La valeur simule une barre de 40 pixels.
func BasTravail() int { return 1040 }

// Disponible indique si la lecture globale de la souris fonctionne.
func Disponible() bool { return false }
