package main

// Les textes de la fenetre de reglages, en francais et en anglais.
//
// Chaque reglage porte un nom court et une phrase qui dit ce qu'il change —
// pas la valeur, l'effet : « il s'approche du curseur et le suit un moment »
// vaut mieux que « probabilite de l'etat Ami ».

func titreFenetre(fr bool) string {
	if fr {
		return "Réglages · Desktop Monkey"
	}
	return "Settings · Desktop Monkey"
}

// motsReglages renvoie tout le vocabulaire de la page en une seule structure,
// serialisee vers la page en JSON.
func motsReglages(fr bool) map[string]any {
	t := func(f, e string) string {
		if fr {
			return f
		}
		return e
	}

	// champ decrit un reglage : son nom, la phrase qui l'explique, et parfois
	// une note sur le cas particulier de la valeur zero.
	champ := func(nom, aide string) map[string]string {
		return map[string]string{"nom": nom, "aide": aide}
	}

	return map[string]any{
		"onglets": []map[string]string{
			{"cle": "apparence", "nom": t("APPARENCE", "LOOKS"), "icone": "◆"},
			{"cle": "vie", "nom": t("VIE", "LIFE"), "icone": "♥"},
			{"cle": "caractere", "nom": t("CARACTÈRE", "CHARACTER"), "icone": "★"},
			{"cle": "paroles", "nom": t("PAROLES", "SPEECH"), "icone": "▣"},
			{"cle": "appli", "nom": t("APPLI", "APP"), "icone": "⚙"},
		},

		"sections": map[string]string{
			"taille":     t("Taille", "Size"),
			"personnage": t("Personnage", "Character sheet"),
			"langue":     t("Langue", "Language"),
			"bulle":      t("Bulles", "Speech bubbles"),
			"sante":      t("Points de vie", "Health"),
			"absences":   t("Absences", "Time away"),
			"social":     t("Avec toi", "With you"),
			"aventure":   t("Aventures", "Adventures"),
			"corps":      t("Besoins", "Body"),
			"bavardage":  t("Bavardage", "Chatter"),
			"entretien":  t("Entretien", "Maintenance"),
			"fichier":    t("Configuration", "Configuration"),
		},

		"champs": map[string]map[string]string{
			"taille": champ(t("Taille du singe", "Monkey size"),
				t("À quel point il est grand sur ton bureau. L'aperçu au-dessus est à l'échelle réelle.",
					"How big he is on your desktop. The preview above is at actual scale.")),
			"planche": champ(t("Personnage", "Character"),
				t("Le jeu de sprites qu'il utilise. Les deux savent marcher, grimper et se faire câliner.",
					"Which sprite sheet he uses. Both know how to walk, climb and be petted.")),
			"langue": champ(t("Langue", "Language"),
				t("Celle de ses répliques et de cette fenêtre. « Auto » suit la langue du système.",
					"For his lines and for this window. \"Auto\" follows your system language.")),
			"echelle_bulle": champ(t("Taille des bulles", "Bubble size"),
				t("La taille du texte quand il parle.", "How big his speech bubbles are.")),

			"coeurs": champ(t("Points de vie", "Hit points"),
				t("Combien de clics il encaisse avant de tomber. Chaque coup lui prend un cœur.",
					"How many clicks he takes before he falls. Each hit costs him one heart.")),
			"secondes_avant_resurrection": champ(t("Résurrection", "Revival"),
				t("Le temps qu'il reste par terre avant de se relever tout seul.",
					"How long he stays down before getting back up on his own.")),
			"cache_apres_clic": champ(t("Il se cache", "He hides"),
				t("Après un coup, il disparaît un moment pour bouder.",
					"After a hit, he vanishes for a while to sulk.")),
			"secondes_avant_sieste": champ(t("Avant la sieste", "Before napping"),
				t("Sans souris qui bouge pendant tout ce temps, il s'allonge et s'endort.",
					"With no mouse movement for this long, he lies down and falls asleep.")),

			"chance_ami": champ(t("Il vient te voir", "He comes to you"),
				t("Il s'approche du curseur et reste un moment à côté.",
					"He walks over to your cursor and hangs around for a while.")),
			"distance_arret": champ(t("Distance", "Stopping distance"),
				t("À quelle distance du curseur il s'arrête quand il vient te voir.",
					"How far from the cursor he stops when he comes over.")),
			"chance_chasse": champ(t("Il court après la souris", "He chases the mouse"),
				t("Il poursuit le curseur au lieu de simplement l'accompagner.",
					"He runs after the cursor instead of just keeping it company.")),
			"chance_vol": champ(t("Il pousse le curseur", "He pushes the cursor"),
				t("Il attrape ta souris et la déplace de quelques pixels. Facétieux.",
					"He grabs your mouse and nudges it a few pixels. Cheeky.")),
			"vitesse": champ(t("Vitesse", "Speed"),
				t("À quelle allure il traverse l'écran.", "How fast he crosses the screen.")),
			"chance_grimpe": champ(t("Il grimpe", "He climbs"),
				t("Il escalade le bord de l'écran et s'y accroche.",
					"He climbs up the edge of the screen and hangs there.")),
			"chance_jeu": champ(t("Il joue", "He plays"),
				t("Cabrioles et pitreries sur place.", "Tumbles and antics on the spot.")),
			"secondes_avant_vie_seule": champ(t("Avant de s'occuper", "Before wandering off"),
				t("Le temps sans souris qui bouge après lequel il commence à vivre sa vie.",
					"How long without mouse movement before he starts doing his own thing.")),
			"chance_repas": champ(t("Il mange", "He eats"),
				t("Il grignote une banane. Il faut qu'il mange avant de pouvoir faire caca.",
					"He snacks on a banana. He has to eat before he can poop.")),
			"chance_crotte": champ(t("Il fait caca", "He poops"),
				t("Un petit tas apparaît derrière lui, puis il se décale. Un clic dessus le nettoie.",
					"A little pile appears behind him, then he steps aside. Click it to clean up.")),
			"chance_jet_crotte": champ(t("Il jette son caca", "He throws his poop"),
				t("Quand un tas traîne depuis un moment, il va le chercher et s'en débarrasse.",
					"When a pile has been sitting around, he picks it up and gets rid of it.")),
			"jet_mode": champ(t("Il vise", "Target"),
				t("Où il l'envoie : par-dessus le bord de l'écran, ou droit sur ton curseur.",
					"Where it goes: over the edge of the screen, or straight at your cursor.")),

			"parle": champ(t("Il parle", "He talks"),
				t("Il commente ce qu'il fait dans une petite bulle.",
					"He comments on what he's doing in a little bubble.")),
			"secondes_entre_paroles": champ(t("Entre deux phrases", "Between lines"),
				t("Le délai avant sa prochaine remarque, tiré au hasard dans cet intervalle.",
					"How long before his next remark, picked at random in this range.")),
			"duree_bulle": champ(t("Durée d'une bulle", "Bubble duration"),
				t("Combien de temps la bulle reste à l'écran.",
					"How long a speech bubble stays on screen.")),

			"auto_update": champ(t("Mises à jour automatiques", "Automatic updates"),
				t("Il va chercher les nouvelles versions tout seul et redémarre pour les installer.",
					"He fetches new versions on his own and restarts to install them.")),
			"demarrage": champ(t("Lancer au démarrage", "Launch at startup"),
				t("Il est là dès l'ouverture de ta session.",
					"He shows up as soon as you log in.")),
		},

		"options": map[string]any{
			"langue": []map[string]string{
				{"v": "auto", "n": t("Auto", "Auto")},
				{"v": "fr", "n": "Français"},
				{"v": "en", "n": "English"},
			},
			"echelle_bulle": []map[string]any{
				{"v": 1, "n": t("Petite", "Small")},
				{"v": 2, "n": t("Moyenne", "Medium")},
				{"v": 3, "n": t("Grande", "Large")},
			},
			"jet_mode": []map[string]any{
				{"v": -1, "n": t("Au hasard", "Random")},
				{"v": 0, "n": t("Hors de l'écran", "Off screen")},
				{"v": 1, "n": t("Sur le curseur", "At the cursor")},
			},
		},

		"humeurs": map[string]string{
			"energie": t("ÉNERGIE", "ENERGY"),
			"ennui":   t("ENNUI", "BOREDOM"),
			"bonheur": t("BONHEUR", "HAPPINESS"),
			"peur":    t("PEUR", "FEAR"),
		},

		// vocabulaire des valeurs : la page fabrique « 35 % · parfois »
		"echelles": map[string]any{
			"chance": []map[string]any{
				{"max": 0.001, "n": t("jamais", "never")},
				{"max": 0.15, "n": t("rarement", "rarely")},
				{"max": 0.4, "n": t("parfois", "sometimes")},
				{"max": 0.7, "n": t("souvent", "often")},
				{"max": 1.01, "n": t("tout le temps", "constantly")},
			},
			"vitesse": []map[string]any{
				{"max": 1.5, "n": t("flânerie", "strolling")},
				{"max": 3.5, "n": t("tranquille", "easygoing")},
				{"max": 5.5, "n": t("pressé", "brisk")},
				{"max": 99, "n": t("survolté", "hyperactive")},
			},
		},

		"unites": map[string]string{
			"s":        t("s", "s"),
			"min":      t("min", "min"),
			"px":       "px",
			"jamais":   t("jamais", "never"),
			"secoue":   t("secoue-le pour le réveiller", "shake him awake"),
			"toujours": t("il reste visible", "he stays visible"),
			"epuise":   t("seulement épuisé", "only when exhausted"),
			"aucune":   t("il ne se lève plus", "he never gets up"),
			"a":        t("de", "from"),
			"b":        t("à", "to"),
		},

		"boutons": map[string]string{
			"enregistrer": t("ENREGISTRER", "SAVE"),
			"annuler":     t("ANNULER", "CANCEL"),
			"defaut":      t("Tout remettre à zéro", "Reset everything"),
			"dossier":     t("Ouvrir le dossier", "Reveal folder"),
		},

		"divers": map[string]string{
			"propre":     t("Aucune modification", "No changes"),
			"saleUn":     t("1 modification", "1 change"),
			"saleN":      t("%d modifications", "%d changes"),
			"redemarre":  t("Il redémarre pour appliquer…", "Restarting to apply…"),
			"version":    t("version", "version"),
			"apercuAide": t("Clique dessus pour l'embêter", "Click him to bother him"),
			"vivant":     t("EN DIRECT", "LIVE"),
		},
	}
}
