//go:build darwin

package main

import "runtime"

// Cocoa exige que la fenetre et sa boucle d'evenements vivent sur le vrai fil
// principal du processus. On y verrouille donc la goroutine principale des le
// demarrage, avant tout autre code — sans quoi le runtime Go pourrait la
// deplacer sur un autre fil et l'interface refuserait de s'afficher.
func init() {
	runtime.LockOSThread()
}
