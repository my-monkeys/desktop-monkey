#!/usr/bin/env python3
"""Fabrique l'icone macOS (AppIcon.iconset) a partir du sprite du singe.

    python3 outils/icone.py <sortie.iconset>

Le singe (pixel art) est agrandi au plus proche voisin pour rester net, puis
pose au centre d'un fond bleu nuit a coins arrondis facon macOS. iconutil
transforme ensuite le dossier en .icns.
"""
import json
import os
import sys

from PIL import Image, ImageDraw

RACINE = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ASSETS = os.path.join(RACINE, "internal", "ressources", "assets")


def sprite():
    d = json.load(open(os.path.join(ASSETS, "singe2.json"), encoding="utf-8"))
    feuille = Image.open(os.path.join(ASSETS, d["image"])).convert("RGBA")
    cl, ch = d["cellule"]
    c, l = d["animations"]["repos_droite"]["cellules"][1]
    img = feuille.crop((c * cl, l * ch, c * cl + cl, l * ch + ch))
    return img.crop(img.getbbox())  # recadre sur le singe


def master(taille=1024):
    singe = sprite()
    board = Image.new("RGBA", (taille, taille), (0, 0, 0, 0))

    # fond a coins arrondis, degrade bleu nuit vertical
    fond = Image.new("RGBA", (taille, taille), (0, 0, 0, 0))
    dr = ImageDraw.Draw(fond)
    haut, bas = (0x22, 0x2c, 0x44), (0x3a, 0x4b, 0x70)
    for y in range(taille):
        t = y / (taille - 1)
        c = tuple(int(haut[i] + t * (bas[i] - haut[i])) for i in range(3))
        dr.line([(0, y), (taille, y)], fill=c + (255,))
    masque = Image.new("L", (taille, taille), 0)
    ImageDraw.Draw(masque).rounded_rectangle(
        [0, 0, taille - 1, taille - 1], radius=int(taille * 0.22), fill=255)
    board.paste(fond, (0, 0), masque)

    # singe agrandi au plus proche voisin, ~62 % de l'icone, centre
    cible = int(taille * 0.62)
    ratio = min(cible / singe.width, cible / singe.height)
    nl, nh = int(singe.width * ratio), int(singe.height * ratio)
    grand = singe.resize((nl, nh), Image.NEAREST)
    board.paste(grand, ((taille - nl) // 2, (taille - nh) // 2), grand)
    return board


def main():
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    dossier = sys.argv[1]
    os.makedirs(dossier, exist_ok=True)
    m = master(1024)
    # tailles requises par iconutil (avec les @2x)
    for taille in (16, 32, 128, 256, 512):
        m.resize((taille, taille), Image.LANCZOS).save(
            os.path.join(dossier, f"icon_{taille}x{taille}.png"))
        d2 = taille * 2
        m.resize((d2, d2), Image.LANCZOS).save(
            os.path.join(dossier, f"icon_{taille}x{taille}@2x.png"))
    print(dossier)


if __name__ == "__main__":
    main()
