#!/usr/bin/env python3
"""Assemble un GIF a partir des captures plein ecran de la VM, recadrees sur le
singe (positions dans rects.csv).

    python3 outils/assembler.py <scene> [largeur_crop] [hauteur_crop] [echelle]

Lit captures/<scene>/f###.png + rects.csv, ecrit docs/<scene>.gif.
"""
import os
import sys

from PIL import Image, ImageDraw

# silhouette d'un pointeur classique (pointe en 0,0), a peindre a la position
# reelle du curseur (CopyFromScreen ne capture pas le curseur materiel)
FLECHE = [(0, 0), (0, 22), (5, 17), (9, 25), (12, 23), (8, 16), (15, 16)]


def dessiner_curseur(im, x, y):
    d = ImageDraw.Draw(im)
    pts = [(x + px, y + py) for px, py in FLECHE]
    d.polygon(pts, fill=(255, 255, 255), outline=(20, 20, 20))


def main():
    scene = sys.argv[1]
    cw = int(sys.argv[2]) if len(sys.argv) > 2 else 520
    ch = int(sys.argv[3]) if len(sys.argv) > 3 else 360
    ech = float(sys.argv[4]) if len(sys.argv) > 4 else 0.8

    dossier = f"captures/{scene}"
    rects = {}
    with open(f"{dossier}/rects.csv") as f:
        for ligne in f:
            ligne = ligne.strip()
            if not ligne:
                continue
            vals = [int(x) for x in ligne.split(",")]
            n, l, t, r, b = vals[:5]
            cur = (vals[5], vals[6]) if len(vals) >= 7 else None
            rects[n] = (l, t, r, b, cur)

    frames = []
    for n in sorted(rects):
        p = f"{dossier}/f{n:03d}.png"
        if not os.path.exists(p):
            continue
        im = Image.open(p).convert("RGB")
        l, t, r, b, cur = rects[n]
        cx = (l + r) // 2      # centre horizontal de la fenetre du singe
        cy = b - 42            # le singe est en bas de sa fenetre
        # le filigrane "Activate Windows" est en bas a droite de l'ecran : on
        # saute les images ou le singe y traine, pour ne pas le montrer
        if cx > 900 and cy > 520:
            continue
        x0 = max(0, min(im.width - cw, cx - cw // 2))
        y0 = max(0, min(im.height - ch, cy - ch // 2))
        crop = im.crop((x0, y0, x0 + cw, y0 + ch))
        # curseur : CopyFromScreen ne le capture pas, on le peint a sa place
        if cur is not None:
            px, py = cur[0] - x0, cur[1] - y0
            if -20 <= px <= cw and -20 <= py <= ch:
                dessiner_curseur(crop, px, py)
        if ech != 1.0:
            crop = crop.resize((int(cw * ech), int(ch * ech)), Image.LANCZOS)
        frames.append(crop)

    if not frames:
        raise SystemExit("aucune image")

    # palette adaptative commune, sans dither. Piege : le fond bleu occupe
    # presque toute l'image, donc une palette naive n'a plus de couleur pour le
    # petit singe brun (il virait au gris-bleu). On sur-pondere donc le singe
    # (au centre du cadrage) en l'agrandissant a cote d'une image de reference.
    mid = frames[len(frames) // 2]
    w, h = mid.size
    mono = mid.crop((w // 2 - 60, h // 2 - 50, w // 2 + 60, h // 2 + 55)).resize((w, h), Image.NEAREST)
    base = Image.new("RGB", (w, h * 2))
    base.paste(mid, (0, 0))
    base.paste(mono, (0, h))
    couleurs = int(os.environ.get("COULEURS", "256"))
    pal = base.quantize(colors=couleurs, method=Image.MEDIANCUT)
    frames = [fr.quantize(palette=pal, dither=Image.NONE) for fr in frames]

    os.makedirs("docs", exist_ok=True)
    sortie = f"docs/{scene}.gif"
    frames[0].save(sortie, save_all=True, append_images=frames[1:],
                   duration=100, loop=0, optimize=True, disposal=2)
    ko = os.path.getsize(sortie) // 1024
    print(f"{sortie} : {len(frames)} images, {ko} Ko")


if __name__ == "__main__":
    main()
