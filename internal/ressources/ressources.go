// Package ressources embarque les images, descripteurs et phrases dans le
// binaire, pour que l'application soit livrable en un seul fichier.
package ressources

import "embed"

//go:embed assets
var Fichiers embed.FS
