//go:build darwin

package main

// facteurAffichage agrandit le singe et son interface sur macOS. Les fenetres y
// sont mesurees en points (l'ecran Retina rend deux pixels par point), ce qui,
// sans correction, donne un singe deux fois plus petit qu'a resolution egale
// sous Windows. On double donc tout ce qui est dessine.
const facteurAffichage = 2
