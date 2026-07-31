#!/usr/bin/env python3
"""Assemble les bandes d'animation d'un pack de sprites en une planche unique.

Le moteur attend une seule image decoupee en grille reguliere, alors que les
packs du commerce livrent souvent une bande par animation, avec des largeurs
d'image differentes. Cet outil normalise tout : chaque image est centree dans
une cellule de taille fixe, une ligne par animation, et les versions tournees
vers la gauche sont generees par symetrie.

    python3 outils/composer_planche.py <dossier_des_bandes> <sortie.png> <sortie.json>

Le pack attendu ici est « Pixel Art Monkey » de tiki-ted : des fichiers nommes
monkey-right-<action>.png, tous hauts de 32 pixels.
"""
import json
import os
import sys

from PIL import Image

# nom de fichier -> (action, nombre d'images, ms par image, boucle)
ANIMATIONS = [
    ("monkey-right-idle.png",   "repos",   4,  200, True),
    ("monkey-right-run.png",    "marche",  8,   80, True),
    ("monkey-right-eat.png",    "mange",  16,   90, True),
    ("monkey-right-die.png",    "meurt",   6,  120, False),
    ("monkey-right-hurt.png",   "touche",  3,  110, False),
    ("monkey-right-jump.png",   "saute",   4,  110, True),
    ("monkey-right-fall.png",   "tombe",   3,  110, True),
    ("monkey-right-attack.png", "attaque", 11,  80, True),
    ("monkey-climb-up.png",     "grimpe",  4,  120, True),
]

CELL_L, CELL_H = 48, 32     # la plus grande image du pack fait 48x32
ECHELLE = 2                 # 32 px de haut -> 64 px a l'ecran


def decouper(chemin, nombre):
    """Renvoie les images d'une bande, deduites de sa largeur totale."""
    bande = Image.open(chemin).convert("RGBA")
    larg, haut = bande.size
    if larg % nombre:
        raise SystemExit(f"{os.path.basename(chemin)} : {larg} px "
                         f"non divisible en {nombre} images")
    pas = larg // nombre
    return [bande.crop((i * pas, 0, (i + 1) * pas, haut)) for i in range(nombre)]


def main():
    if len(sys.argv) != 4:
        raise SystemExit(__doc__)
    source, sortie_png, sortie_json = sys.argv[1:]

    bandes = []
    for fichier, action, nombre, ms, boucle in ANIMATIONS:
        chemin = os.path.join(source, fichier)
        if not os.path.exists(chemin):
            print(f"  ignore : {fichier} absent")
            continue
        bandes.append((action, decouper(chemin, nombre), ms, boucle))

    if not bandes:
        raise SystemExit("aucune bande trouvee")

    colonnes = max(len(images) for _, images, _, _ in bandes)
    lignes = len(bandes) * 2          # une ligne vers la droite, une vers la gauche
    planche = Image.new("RGBA", (colonnes * CELL_L, lignes * CELL_H), (0, 0, 0, 0))

    animations = {}
    for i, (action, images, ms, boucle) in enumerate(bandes):
        for sens, decalage in (("droite", 0), ("gauche", len(bandes))):
            ligne = i + decalage
            cellules = []
            for j, img in enumerate(images):
                vignette = img.transpose(Image.FLIP_LEFT_RIGHT) if sens == "gauche" else img
                # centre l'image dans sa cellule : le personnage garde sa place
                # quelle que soit la largeur d'origine de l'animation
                x = j * CELL_L + (CELL_L - vignette.width) // 2
                y = ligne * CELL_H + (CELL_H - vignette.height)
                planche.paste(vignette, (x, y), vignette)
                cellules.append([j, ligne])
            spec = {"cellules": cellules, "ms": ms}
            if not boucle:
                spec["boucle"] = False
            animations[f"{action}_{sens}"] = spec

    planche.save(sortie_png)

    with open(sortie_json, "w", encoding="utf-8") as f:
        json.dump({
            "nom": "Pixel Art Monkey (tiki-ted)",
            "image": os.path.basename(sortie_png),
            "cellule": [CELL_L, CELL_H],
            "echelle": ECHELLE,
            "vue": "profil",
            "animations": animations,
        }, f, indent=2, ensure_ascii=False)

    print(f"planche  : {sortie_png}  ({planche.width}x{planche.height})")
    print(f"descripteur : {sortie_json}")
    print(f"animations : {', '.join(sorted(a for a, _, _, _ in bandes))}")


if __name__ == "__main__":
    main()
