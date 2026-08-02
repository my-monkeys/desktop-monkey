// Pont Objective-C pour la fenetre de reglages (NSWindow + WKWebView).
// Declare en C pur pour cgo.
#ifndef SINGE_DIALOGUE_DARWIN_H
#define SINGE_DIALOGUE_DARWIN_H

// Ouvre (ou ramene au premier plan) la fenetre, chargee sur l'URL donnee.
void dialogue_ouvrir(const char *url, const char *titre, int larg, int haut);

// Ferme la fenetre si elle est ouverte.
void dialogue_fermer(void);

#endif
