// Pont Objective-C vers Cocoa pour l'icone du singe dans la barre des menus de
// macOS (NSStatusItem) et son petit menu. Declare en C pur pour cgo.
#ifndef SINGE_TRAY_DARWIN_H
#define SINGE_TRAY_DARWIN_H

// Installe l'icone et son menu. nomApp sert d'etiquette au LaunchAgent, exe est
// le chemin de l'executable (lancement au demarrage), config le fichier de
// reglages a ouvrir, et icon une image RGBA iw x ih (l'icone du singe).
void tray_init(const char *nomApp, const char *exe, const char *config,
               unsigned char *icon, int iw, int ih);

// Renvoie 1 si l'utilisateur a choisi Quitter dans le menu.
int tray_quit_requested(void);

// Retire l'icone de la barre des menus.
void tray_fermer(void);

#endif
