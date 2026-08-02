//go:build !darwin

// Sur les autres plateformes, le dialogue natif n'est pas (encore) implemente :
// l'appelant retombe sur la page de reglages web.
package dialogue

// Ouvrir n'a pas d'effet ici.
func Ouvrir(descJSON string) {}

// Resultat ne renvoie jamais rien ici.
func Resultat() (string, bool) { return "", false }

// Disponible indique que la plateforme n'offre pas le dialogue natif.
func Disponible() bool { return false }
