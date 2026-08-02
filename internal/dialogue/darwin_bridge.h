// Pont Objective-C pour la fenetre native de reglages (NSTabView + curseurs).
// Declare en C pur pour cgo.
#ifndef SINGE_DIALOGUE_DARWIN_H
#define SINGE_DIALOGUE_DARWIN_H

// Ouvre la fenetre de reglages decrite par un JSON :
//
//	{
//	  "titre": "...", "enregistrer": "...", "annuler": "...",
//	  "onglets": [{
//	    "titre": "...",
//	    "champs": [{
//	      "type": "curseur" | "entier" | "choix" | "case",
//	      "cle": "...", "nom": "...", "aide": "...",
//	      "min": 0, "max": 1, "pas": 0.05, "valeur": 0.5,     curseur/entier
//	      "options": ["auto"], "libelles": ["auto"], "texte": "auto",  choix
//	      "coche": true                                        case
//	    }]
//	  }]
//	}
//
// Sans effet si la fenetre est deja ouverte (elle revient au premier plan).
void dialogue_ouvrir(const char *desc);

// Renvoie le JSON des valeurs enregistrees (cle -> nombre/booleen/texte), ou
// NULL si rien n'a ete enregistre depuis le dernier appel. L'appelant libere
// la chaine avec free().
char *dialogue_resultat(void);

#endif
