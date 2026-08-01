//go:build !darwin

package main

// facteurAffichage vaut 1 hors macOS : la taille des sprites est prise telle
// quelle (voir la version darwin pour le pourquoi du doublement).
const facteurAffichage = 1
