/* Desktop Monkey — la vie du site.
 *
 * Trois choses seulement : le singe qui se promene en bas de page (il suit le
 * curseur, s'ennuie, parle, et meurt si on le tape), les jauges d'humeur qui
 * derivent dans le faux menu du tray, et la fenetre de reglages qui reagit
 * pour de vrai. Aucune dependance, aucun reglage n'est envoye nulle part.
 */
(() => {
  'use strict';

  // La planche de sprites : 16 colonnes de 48x32. Memes lignes que dans
  // l'application (voir internal/ressources/assets/singe2.json).
  const CEL_L = 48, CEL_H = 32, PLANCHE_L = 768, PLANCHE_H = 576;
  const LIGNE = { reposD: 0, marcheD: 1, meurtD: 3, reposG: 9, marcheG: 10, meurtG: 12 };

  const PAROLES = {
    touche: ["Ouch!", "Ow ow ow!", "Hey, that hurts!", "Why are you hitting me?",
             "I thought we were friends…", "Not the face!", "Careful, I'm fragile!"],
    hasard: ["Remember to drink water 💧", "I'm proud of you",
             "Nice cursor. Would be a shame if someone… took it.",
             "Did you know I can climb curtains?", "I'm supervising your work. Good work.",
             "I saw a banana fly by, brb", "Your screen is very comfy",
             "Stretch a little, it feels great"],
    mort: ["You got me 💔", "Tell the banana I loved her…", "Farewell, cruel world…"],
    ranime: ["The power of love 💕", "I'm alive!", "I saw a giant glowing banana…", "Where am I?"],
    retour: ["I'm back 🐒", "Did you miss me?", "Clean slate, let's go"]
  };

  const auHasard = (l) => l[Math.floor(Math.random() * l.length)];
  const borner = (v, min, max) => Math.max(min, Math.min(max, v));
  const doux = matchMedia('(prefers-reduced-motion: reduce)').matches;

  // poser une image de la planche sur un element, a l'echelle voulue
  function image(el, colonne, ligne, echelle) {
    el.style.width = CEL_L * echelle + 'px';
    el.style.height = CEL_H * echelle + 'px';
    el.style.backgroundSize = PLANCHE_L * echelle + 'px ' + PLANCHE_H * echelle + 'px';
    el.style.backgroundPosition = -(colonne * CEL_L * echelle) + 'px ' + -(ligne * CEL_H * echelle) + 'px';
  }

  /* ————————————————————————— copier la commande ————————————————————————— */

  const COMMANDE = 'brew install --cask my-monkeys/tap/desktop-monkey';

  function copierDansLePressePapier(texte) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(texte).catch(() => copieDeSecours(texte));
    }
    copieDeSecours(texte);
    return Promise.resolve();
  }

  // vieux navigateurs, ou page servie sans contexte securise
  function copieDeSecours(texte) {
    const z = document.createElement('textarea');
    z.value = texte;
    z.style.position = 'fixed';
    z.style.opacity = '0';
    document.body.appendChild(z);
    z.select();
    try { document.execCommand('copy'); } catch (_) { /* tant pis */ }
    z.remove();
  }

  const etatsCopie = document.querySelectorAll('[data-copie-etat]');
  let minuterieCopie;
  document.querySelectorAll('[data-copier]').forEach((bouton) => {
    bouton.addEventListener('click', () => {
      copierDansLePressePapier(COMMANDE).then(() => {
        etatsCopie.forEach((e) => { e.textContent = '✓ Copied!'; });
        clearTimeout(minuterieCopie);
        minuterieCopie = setTimeout(() => {
          etatsCopie.forEach((e) => { e.textContent = 'Click the line to copy'; });
        }, 2000);
      });
    });
  });

  /* ————————————————————————— les humeurs ————————————————————————— */

  const humeurs = { energy: 74, boredom: 36, happiness: 88, fear: 14 };

  // une jauge de dix crans : les pleins prennent l'accent, les autres restent gris
  function jauge(valeur) {
    const n = borner(Math.round(valeur / 10), 0, 10);
    return { pleins: '▰'.repeat(n), vides: '▰'.repeat(10 - n) };
  }

  function motDHumeur() {
    if (humeurs.fear > 60) return 'keeping his distance today';
    if (humeurs.boredom > 65) return 'plotting something';
    if (humeurs.energy < 30) return 'looking for a nap spot';
    if (humeurs.happiness > 75) return 'very pleased with you';
    return 'just hanging around';
  }

  const quip = document.querySelector('[data-quip]');
  function afficherHumeurs() {
    for (const nom of Object.keys(humeurs)) {
      const j = jauge(humeurs[nom]);
      const pleins = document.querySelector('[data-jauge="' + nom + '-on"]');
      const vides = document.querySelector('[data-jauge="' + nom + '-off"]');
      if (pleins) pleins.textContent = j.pleins;
      if (vides) vides.textContent = j.vides;
    }
    if (quip) quip.textContent = motDHumeur();
  }
  afficherHumeurs();

  if (!doux) {
    setInterval(() => {
      for (const nom of Object.keys(humeurs)) {
        humeurs[nom] = borner(humeurs[nom] + (Math.random() * 16 - 8), 6, 97);
      }
      afficherHumeurs();
    }, 1500);
  }

  /* ————————————————————————— la fenetre de reglages ————————————————————————— */

  const reglages = { taille: 1.4, vitesse: 1, coeurs: 3 };

  // onglets
  const boutonsOnglet = document.querySelectorAll('[data-onglet]');
  boutonsOnglet.forEach((bouton) => {
    bouton.addEventListener('click', () => {
      boutonsOnglet.forEach((b) => {
        const actif = b === bouton;
        b.setAttribute('aria-selected', String(actif));
        document.getElementById('tab-' + b.dataset.onglet).hidden = !actif;
      });
    });
  });

  // curseurs simples : la valeur s'affiche a cote de son etiquette
  function brancherCurseur(id, sortie, format, surChangement) {
    const entree = document.getElementById(id);
    const cible = document.querySelector('[data-out="' + sortie + '"]');
    if (!entree || !cible) return;
    const majAffichage = () => {
      const v = parseFloat(entree.value);
      cible.textContent = format(v);
      if (surChangement) surChangement(v);
    };
    entree.addEventListener('input', majAffichage);
    majAffichage();
  }

  brancherCurseur('dm-size', 'size', (v) => v.toFixed(1) + '×', (v) => { reglages.taille = v; });
  brancherCurseur('dm-speed', 'speed', (v) => v.toFixed(1) + '×', (v) => { reglages.vitesse = v; });

  // coeurs : cliquer le n-ieme regle la resistance du singe
  const coeurs = document.querySelectorAll('[data-coeur]');
  const noteCoeurs = document.querySelector('[data-coeurs-note]');
  function afficherCoeurs() {
    coeurs.forEach((c) => {
      const plein = Number(c.dataset.coeur) <= reglages.coeurs;
      c.toggleAttribute('data-plein', plein);
      c.setAttribute('aria-pressed', String(plein));
    });
    if (noteCoeurs) {
      noteCoeurs.textContent = reglages.coeurs === 1
        ? "one click and he's gone"
        : reglages.coeurs + ' clicks to knock him out';
    }
  }
  coeurs.forEach((c) => c.addEventListener('click', () => {
    reglages.coeurs = Number(c.dataset.coeur);
    afficherCoeurs();
  }));
  afficherCoeurs();

  // interrupteurs
  document.querySelectorAll('[data-bascule]').forEach((b) => {
    b.addEventListener('click', () => {
      b.setAttribute('aria-checked', b.getAttribute('aria-checked') === 'true' ? 'false' : 'true');
    });
  });

  // caractere : quatre curseurs qui composent une petite phrase
  const traits = { mischief: 72, affection: 80, shyness: 28, appetite: 66 };
  const persona = document.querySelector('[data-persona]');
  function afficherPersona() {
    if (!persona) return;
    persona.textContent =
      (traits.mischief > 60 ? 'a bold little troublemaker' : 'a well-behaved houseguest') +
      (traits.affection > 60 ? ' who really likes you' : ' who tolerates you') +
      (traits.appetite > 70 ? ', and he is always hungry.' : '.');
  }
  [['dm-mis', 'mischief'], ['dm-aff', 'affection'], ['dm-shy', 'shyness'], ['dm-app', 'appetite']]
    .forEach(([id, trait]) => brancherCurseur(id, trait, (v) => String(Math.round(v)), (v) => {
      traits[trait] = v;
      afficherPersona();
    }));
  afficherPersona();

  // enregistrer : le singe fait mine de redemarrer
  const noteSauve = document.querySelector('[data-sauve-note]');
  const boutonSauver = document.querySelector('[data-sauver]');
  let minuterieSauve;
  if (boutonSauver && noteSauve) {
    boutonSauver.addEventListener('click', () => {
      noteSauve.textContent = 'Saved. He restarted — same monkey, new settings.';
      dire(auHasard(PAROLES.retour), 2600);
      clearTimeout(minuterieSauve);
      minuterieSauve = setTimeout(() => {
        noteSauve.textContent = 'He restarts when you save. Takes about a second.';
      }, 2600);
    });
  }

  /* ————————————————————————— le singe du bas de page ————————————————————————— */

  const promeneur = document.querySelector('[data-promeneur]');
  const bulle = document.querySelector('[data-bulle]');
  const apercu = document.querySelector('[data-apercu]');

  let bulleJusqua = 0;
  function dire(texte, duree) {
    if (!bulle) return;
    bulle.textContent = texte;
    bulle.style.opacity = '1';
    bulleJusqua = performance.now() + (duree || 3200);
  }

  const souris = { x: innerWidth * 0.5, vu: -9999 };
  addEventListener('pointermove', (e) => {
    souris.x = e.clientX;
    souris.vu = performance.now();
  }, { passive: true });

  const singe = {
    x: innerWidth * 0.18, sens: 1, mode: 'walk',
    t: 0, cible: null, prochaineErrance: 0, mortA: 0
  };
  let tapes = 0, derniereTape = 0, prochaineParole = performance.now() + 9000;

  if (promeneur) {
    promeneur.addEventListener('click', () => {
      const maintenant = performance.now();
      if (singe.mode === 'mort') return;
      if (maintenant - derniereTape > 9000) tapes = 0;
      derniereTape = maintenant;
      tapes++;
      if (tapes >= 3) {
        singe.mode = 'mort';
        singe.t = 0;
        singe.mortA = maintenant;
        dire(auHasard(PAROLES.mort), 2600);
      } else {
        dire(auHasard(PAROLES.touche), 1800);
      }
    });
  }

  let dernierInstant = performance.now();
  let tApercu = 0;

  function battement(maintenant) {
    const dt = Math.min(64, maintenant - dernierInstant);
    dernierInstant = maintenant;

    if (promeneur) {
      const maxX = Math.max(0, innerWidth - CEL_L * 2);

      if (singe.mode === 'mort') {
        singe.t += dt;
        const i = Math.min(5, Math.floor(singe.t / 130));
        image(promeneur, i, singe.sens > 0 ? LIGNE.meurtD : LIGNE.meurtG, 2);
        if (maintenant - singe.mortA > 3400) {
          singe.mode = 'idle';
          singe.t = 0;
          tapes = 0;
          dire(auHasard(PAROLES.ranime), 2600);
        }
      } else {
        // le curseur reste son meilleur ami tant qu'il a bouge recemment
        const suit = maintenant - souris.vu < 5000;
        if (suit) {
          singe.cible = borner(souris.x - CEL_L, 0, maxX);
        } else if (maintenant > singe.prochaineErrance) {
          singe.cible = Math.random() * maxX;
          singe.prochaineErrance = maintenant + 3200 + Math.random() * 4200;
        }

        const ecart = (singe.cible == null ? singe.x : singe.cible) - singe.x;
        if (Math.abs(ecart) > 24 && !doux) {
          singe.sens = ecart > 0 ? 1 : -1;
          singe.x = borner(singe.x + singe.sens * 0.105 * dt, 0, maxX);
          singe.mode = 'walk';
        } else if (singe.mode !== 'idle') {
          singe.mode = 'idle';
          singe.t = 0;
          if (suit && maintenant > prochaineParole) {
            dire(auHasard(PAROLES.hasard), 3400);
            prochaineParole = maintenant + 14000 + Math.random() * 12000;
          }
        }

        singe.t += dt;
        if (singe.mode === 'walk') {
          image(promeneur, Math.floor(singe.t / 84) % 8, singe.sens > 0 ? LIGNE.marcheD : LIGNE.marcheG, 2);
        } else {
          image(promeneur, Math.floor(singe.t / 220) % 4, singe.sens > 0 ? LIGNE.reposD : LIGNE.reposG, 2);
        }
      }

      promeneur.style.transform = 'translateX(' + Math.round(singe.x) + 'px)';

      if (bulle) {
        const l = bulle.offsetWidth || 0;
        bulle.style.transform = 'translateX(' +
          Math.round(borner(singe.x + 44 - l / 2, 6, innerWidth - l - 6)) + 'px)';
        if (bulleJusqua && maintenant > bulleJusqua) {
          bulle.style.opacity = '0';
          bulleJusqua = 0;
        }
      }
    }

    // l'apercu de la fenetre de reglages marche a la taille et a l'allure choisies
    if (apercu) {
      const ms = 84 / Math.max(0.4, reglages.vitesse);
      tApercu += doux ? 0 : dt;
      image(apercu, Math.floor(tApercu / ms) % 8, LIGNE.marcheD, reglages.taille * 1.5);
    }

    requestAnimationFrame(battement);
  }

  requestAnimationFrame(battement);
})();
