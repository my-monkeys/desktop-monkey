#!/usr/bin/env python3
"""Genere un apercu des animations d'une planche : une planche de contact avec
toutes les images, et un GIF ou chaque animation tourne a sa vitesse reelle.

    python3 outils/apercu_animations.py <descripteur.json> <sortie_sans_extension>
"""
import json
import os
import sys

from PIL import Image, ImageDraw

ZOOM = 3
MARGE = 10
LIBELLE = 150
FOND = (26, 24, 30)
TEXTE = (232, 228, 220)
DISCRET = (140, 134, 128)


def charger(chemin):
    with open(chemin, encoding="utf-8") as f:
        d = json.load(f)
    feuille = Image.open(os.path.join(os.path.dirname(chemin), d["image"])).convert("RGBA")
    cl, ch = d["cellule"]

    animations = {}
    for nom, spec in d["animations"].items():
        images = [
            feuille.crop((c * cl, l * ch, c * cl + cl, l * ch + ch))
            for c, l in spec["cellules"]
        ]
        animations[nom] = {
            "images": images,
            "ms": spec.get("ms", 150),
            "boucle": spec.get("boucle", True),
        }
    return animations, cl, ch


def planche_de_contact(animations, cl, ch, sortie):
    """Une ligne par animation, toutes ses images cote a cote."""
    noms = sorted(a for a in animations if not a.endswith("_gauche"))
    colonnes = max(len(animations[n]["images"]) for n in noms)

    cw, chh = cl * ZOOM, ch * ZOOM
    larg = LIBELLE + colonnes * (cw + MARGE) + MARGE
    haut = MARGE + len(noms) * (chh + MARGE + 16)

    board = Image.new("RGB", (larg, haut), FOND)
    dr = ImageDraw.Draw(board)

    for i, nom in enumerate(noms):
        a = animations[nom]
        y = MARGE + i * (chh + MARGE + 16)
        etiquette = nom.replace("_droite", "")
        dr.text((10, y + chh // 2 - 12), etiquette, fill=TEXTE)
        detail = f"{len(a['images'])} img · {a['ms']} ms"
        if not a["boucle"]:
            detail += " · 1 fois"
        dr.text((10, y + chh // 2 + 2), detail, fill=DISCRET)

        for j, img in enumerate(a["images"]):
            vignette = img.resize((cw, chh), Image.NEAREST)
            x = LIBELLE + j * (cw + MARGE)
            # damier leger, pour distinguer le sprite du fond
            for py in range(0, chh, 8):
                for px in range(0, cw, 8):
                    if (px // 8 + py // 8) % 2 == 0:
                        dr.rectangle([x + px, y + py, x + px + 7, y + py + 7],
                                     fill=(38, 36, 42))
            board.paste(vignette, (x, y), vignette)

    board.save(sortie)
    return sortie, len(noms)


def gif_anime(animations, cl, ch, sortie):
    """Toutes les animations jouees en parallele, a leur vitesse reelle."""
    noms = sorted(a for a in animations if not a.endswith("_gauche"))
    cw, chh = cl * ZOOM, ch * ZOOM
    par_ligne = 5
    lignes = (len(noms) + par_ligne - 1) // par_ligne
    larg = MARGE + par_ligne * (cw + MARGE)
    haut = MARGE + lignes * (chh + MARGE + 14)

    PAS = 40  # ms par image du GIF
    duree = max(len(animations[n]["images"]) * animations[n]["ms"] for n in noms)
    total = max(1, duree // PAS)

    images = []
    for t in range(total):
        board = Image.new("RGB", (larg, haut), FOND)
        dr = ImageDraw.Draw(board)
        for i, nom in enumerate(noms):
            a = animations[nom]
            col, lig = i % par_ligne, i // par_ligne
            x = MARGE + col * (cw + MARGE)
            y = MARGE + lig * (chh + MARGE + 14)

            k = (t * PAS) // a["ms"]
            if not a["boucle"]:
                k = min(k, len(a["images"]) - 1)
            else:
                k %= len(a["images"])

            board.paste(a["images"][k].resize((cw, chh), Image.NEAREST), (x, y),
                        a["images"][k].resize((cw, chh), Image.NEAREST))
            dr.text((x, y + chh + 2), nom.replace("_droite", ""), fill=DISCRET)
        images.append(board)

    images[0].save(sortie, save_all=True, append_images=images[1:],
                   duration=PAS, loop=0, optimize=True)
    return sortie


def main():
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    descripteur, base = sys.argv[1], sys.argv[2]

    animations, cl, ch = charger(descripteur)
    png, n = planche_de_contact(animations, cl, ch, base + ".png")
    gif = gif_anime(animations, cl, ch, base + ".gif")
    print(f"{n} animations · {png} · {gif}")


if __name__ == "__main__":
    main()
