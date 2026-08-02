package main

// Le gabarit de la fenetre de reglages : un diorama vivant en haut (le singe y
// marche a sa taille reelle, ses humeurs defilent, on peut le taper), cinq
// onglets en dessous, et un pied qui compte ce qui a change.
//
// La page ne connait aucun reglage en dur : elle recoit l'etat, les libelles et
// les sprites depuis le Go (voir reglagesui.go et reglagesmots.go) et se
// construit a partir de la.

import "html/template"

var gabaritReglages = template.Must(template.New("reglages").Parse(pageHTML))

const pageHTML = `<!doctype html>
<html lang="fr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{ .Titre }}</title>
<style>
@font-face{font-family:'Pixel';src:url("{{ .PolicePixel }}") format('woff2');font-display:block}
@font-face{font-family:'PixelUI';src:url("{{ .PoliceUI }}") format('woff2');font-display:block}

:root{
  --bg:#171a23; --pnl:#1e222d; --fld:#12151d; --bar:#1a1e28;
  --brd:#323a4c; --brd2:#2b3244; --txt:#e8eaf1; --mut:#98a0b4;
  --a:#c98a4b; --a2:#e8b586; --ombre:#05070a;
}
*{box-sizing:border-box;margin:0;padding:0}
html,body{height:100%}
body{
  background:var(--bg); color:var(--txt);
  font-family:'PixelUI',ui-monospace,monospace; font-size:15px; line-height:1.45;
  overflow:hidden; user-select:none; -webkit-user-select:none;
  -webkit-font-smoothing:none;
}
.px{font-family:'Pixel',monospace;letter-spacing:.5px;text-transform:uppercase}
img{image-rendering:pixelated}
button{font:inherit;color:inherit;background:none;border:0;cursor:pointer}
:focus-visible{outline:2px solid var(--a2);outline-offset:2px}

.app{display:flex;flex-direction:column;height:100%}

/* ── le diorama ───────────────────────────────────────────────────── */
.diorama{
  position:relative;height:164px;flex:0 0 164px;overflow:hidden;
  background:#0d1016;border-bottom:3px solid var(--brd);
}
.ciel{position:absolute;inset:0;background:linear-gradient(180deg,#1b2231 0%,rgba(27,34,49,0) 100%)}
.couche{position:absolute;left:0;right:0;bottom:32px;height:70px;background-repeat:repeat-x}
.c3{background-image:
  radial-gradient(circle at 34px 30px,#1a212e 0 30px,rgba(0,0,0,0) 31px),
  radial-gradient(circle at 106px 46px,#1a212e 0 21px,rgba(0,0,0,0) 22px),
  linear-gradient(90deg,rgba(0,0,0,0) 0 32px,#1a212e 32px 36px,rgba(0,0,0,0) 37px 103px,#1a212e 103px 106px,rgba(0,0,0,0) 107px);
  background-size:140px 100%}
.c2{background-image:
  radial-gradient(circle at 48px 28px,#1e2634 0 33px,rgba(0,0,0,0) 34px),
  linear-gradient(90deg,rgba(0,0,0,0) 0 43px,#1e2634 43px 50px,rgba(0,0,0,0) 51px);
  background-size:191px 100%;bottom:32px;height:62px}
.c1{background-image:
  radial-gradient(circle at 24px 30px,#232c3c 0 22px,rgba(0,0,0,0) 23px),
  radial-gradient(circle at 72px 34px,#232c3c 0 18px,rgba(0,0,0,0) 19px);
  background-size:97px 100%;bottom:32px;height:52px}
.sol{position:absolute;left:0;right:0;bottom:0;height:32px;background:#262d3d;
  border-top:3px solid #3a4459;
  background-image:linear-gradient(90deg,rgba(0,0,0,.18) 0 4px,rgba(0,0,0,0) 4px 8px);
  background-size:8px 100%}

.scene{position:absolute;bottom:29px;left:50%;transform:translateX(-50%);
  display:flex;flex-direction:column;align-items:center;cursor:pointer}
.sprite{background-repeat:no-repeat;image-rendering:pixelated}
.coeursTete{position:absolute;bottom:calc(100% + 3px);left:50%;transform:translateX(-50%);
  display:flex;gap:3px}
.coeursTete img{width:11px;height:auto}
.bulle{position:absolute;bottom:calc(100% + 20px);left:50%;transform:translateX(-50%);
  white-space:nowrap;background:var(--txt);color:#12151d;padding:3px 7px;font-size:12px;
  border:2px solid #05070a;opacity:0;transition:opacity .2s}
.bulle.on{opacity:1}
.bulle:after{content:"";position:absolute;top:100%;left:50%;margin-left:-3px;
  border:4px solid rgba(0,0,0,0);border-top-color:var(--txt)}
.envol{position:absolute;pointer-events:none;animation:envol 1s ease-out forwards}
.envol img{width:12px;height:auto}
@keyframes envol{to{transform:translate(var(--dx),-34px);opacity:0}}

.humeurs{position:absolute;top:10px;left:10px;display:flex;flex-direction:column;gap:5px;
  background:rgba(13,16,22,.72);border:2px solid var(--brd2);padding:7px 9px}
.humeur{display:flex;align-items:center;gap:7px}
.humeur .nom{font-size:7px;width:64px;color:var(--mut)}
.jauge{display:flex;gap:2px}
.jauge i{width:5px;height:8px;background:#2b3244;display:block}
.jauge i.on{background:var(--a)}
.jauge i.on.haut{background:var(--a2)}

.astuce{position:absolute;bottom:7px;right:10px;font-size:11px;color:#6f7890}
.direct{position:absolute;top:11px;right:10px;font-size:10px;color:var(--a);
  display:flex;align-items:center;gap:4px}
.direct b{width:5px;height:5px;background:var(--a);display:block;animation:clign 1.4s steps(2) infinite}
@keyframes clign{50%{opacity:.25}}

/* ── onglets ──────────────────────────────────────────────────────── */
.onglets{display:flex;background:var(--bar);border-bottom:3px solid var(--brd);flex:0 0 auto}
.onglet{flex:1;padding:11px 4px 9px;font-size:8px;color:var(--mut);
  border-right:2px solid var(--brd2);display:flex;align-items:center;justify-content:center;gap:5px}
.onglet:last-child{border-right:0}
.onglet:hover{color:var(--txt);background:#20252f}
.onglet.actif{color:#12151d;background:var(--a2)}
.onglet .ic{font-size:10px;line-height:1}

/* ── panneaux ─────────────────────────────────────────────────────── */
.panneaux{flex:1;overflow-y:auto;overflow-x:hidden;padding:14px 16px 18px}
.panneaux::-webkit-scrollbar{width:12px}
.panneaux::-webkit-scrollbar-track{background:var(--fld)}
.panneaux::-webkit-scrollbar-thumb{background:var(--brd);border:2px solid var(--fld)}
.panneau{display:none;flex-direction:column;gap:14px}
.panneau.actif{display:flex}

.section{border:2px solid var(--brd2);background:var(--pnl)}
.section > h2{font-family:'Pixel',monospace;font-size:8px;letter-spacing:.5px;text-transform:uppercase;
  color:var(--a2);background:var(--bar);padding:8px 11px;border-bottom:2px solid var(--brd2)}
.section > .corps{display:flex;flex-direction:column}

.reglage{padding:11px 12px 12px;border-top:2px solid var(--brd2)}
.corps > .reglage:first-child{border-top:0}
.reglage.imbrique{margin-left:14px;border-left:2px solid var(--brd2);background:rgba(0,0,0,.14)}
.reglage.eteint{opacity:.42}
.reglage.eteint .ctrl{pointer-events:none}
.tete{display:flex;align-items:baseline;gap:10px;margin-bottom:8px}
.tete .nom{flex:1;font-size:15px}
.tete .val{font-family:'Pixel',monospace;font-size:8px;color:var(--a2);text-align:right;
  font-variant-numeric:tabular-nums}
.aide{font-size:12.5px;color:var(--mut);margin-top:8px;line-height:1.4;max-width:62ch}

/* curseur 8-bit */
.curseur{position:relative;height:18px;display:flex;align-items:center;cursor:ew-resize;touch-action:none}
.rail{position:absolute;left:0;right:0;height:10px;background:var(--fld);border:2px solid var(--brd2)}
.remplit{position:absolute;left:2px;top:2px;bottom:2px;background:var(--a);
  background-image:linear-gradient(90deg,rgba(255,255,255,.16) 0 2px,rgba(0,0,0,0) 2px 4px);
  background-size:4px 100%}
.pouce{position:absolute;width:14px;height:18px;background:var(--a2);
  border:2px solid var(--ombre);box-shadow:inset 0 -3px 0 rgba(0,0,0,.28);margin-left:-7px}

/* segments */
.segment{display:flex;border:2px solid var(--brd2);background:var(--fld);width:fit-content}
.segment button{padding:6px 13px;font-size:13px;color:var(--mut);border-right:2px solid var(--brd2)}
.segment button:last-child{border-right:0}
.segment button:hover{color:var(--txt)}
.segment button.actif{background:var(--a);color:#12151d}

/* interrupteur */
.bascule{width:44px;height:22px;background:var(--fld);border:2px solid var(--brd2);
  position:relative;flex:0 0 auto}
.bascule i{position:absolute;top:0;left:0;width:18px;height:18px;background:#5a6478;
  border:2px solid var(--ombre);transition:left .12s steps(3)}
.bascule.on{background:#3a2a18;border-color:var(--a)}
.bascule.on i{left:22px;background:var(--a2)}

/* coeurs */
.rangeeCoeurs{display:flex;gap:5px}
.rangeeCoeurs button{padding:2px}
.rangeeCoeurs img{width:18px;height:auto;display:block}
.rangeeCoeurs button:hover img{transform:translateY(-2px)}

/* cartes de personnage */
.cartes{display:grid;grid-template-columns:1fr 1fr;gap:10px}
.carte{border:2px solid var(--brd2);background:var(--fld);padding:10px;text-align:left;
  display:flex;align-items:center;gap:11px}
.carte:hover{border-color:#4a5570}
.carte.actif{border-color:var(--a);background:#241d14}
.carte .vitrine{width:52px;height:52px;flex:0 0 52px;display:flex;align-items:center;justify-content:center;
  background:#0d1016;border:2px solid var(--brd2);overflow:hidden}
.carte .nom{font-size:14px}
.carte .sous{font-size:11px;color:var(--mut)}

/* intervalle a deux poignees */
.doubleCurseur{position:relative;height:18px;touch-action:none}

/* chemin de config */
.chemin{display:flex;gap:8px;align-items:center}
.chemin code{flex:1;background:var(--fld);border:2px solid var(--brd2);padding:7px 9px;
  font-size:11px;color:var(--mut);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;
  font-family:ui-monospace,monospace;user-select:text;-webkit-user-select:text}

/* ── pied ─────────────────────────────────────────────────────────── */
.pied{flex:0 0 auto;display:flex;align-items:center;gap:12px;padding:10px 14px;
  background:var(--bar);border-top:3px solid var(--brd)}
.etat{flex:1;font-size:8px;color:var(--mut)}
.etat.sale{color:var(--a2)}
.lien{font-size:12px;color:var(--mut);text-decoration:underline;text-underline-offset:3px}
.lien:hover{color:var(--txt)}
.btn{font-family:'Pixel',monospace;font-size:8px;letter-spacing:.5px;padding:10px 14px;
  background:var(--pnl);border:2px solid var(--brd);box-shadow:inset 0 -3px 0 rgba(0,0,0,.35)}
.btn:hover{background:#252a37}
.btn:active{box-shadow:inset 0 2px 0 rgba(0,0,0,.35);transform:translateY(1px)}
.btn.primaire{background:var(--a);color:#12151d;border-color:var(--a2)}
.btn.primaire:hover{background:var(--a2)}
.btn:disabled{opacity:.4;cursor:default;transform:none}
</style>
</head>
<body>
<div class="app">
  <header class="diorama" id="diorama">
    <div class="ciel"></div>
    <div class="couche c3" id="c3"></div>
    <div class="couche c2" id="c2"></div>
    <div class="couche c1" id="c1"></div>
    <div class="sol" id="c0"></div>
    <div class="humeurs" id="humeurs"></div>
    <div class="direct" id="direct" hidden><b></b><span></span></div>
    <div class="scene" id="scene">
      <div class="bulle" id="bulle"></div>
      <div class="coeursTete" id="coeursTete"></div>
      <div class="sprite" id="singe"></div>
    </div>
    <div class="astuce" id="astuce"></div>
  </header>

  <nav class="onglets" id="onglets"></nav>
  <main class="panneaux" id="panneaux"></main>

  <footer class="pied">
    <div class="etat px" id="etat"></div>
    <button class="lien" id="defaut"></button>
    <button class="btn" id="annuler"></button>
    <button class="btn primaire" id="enregistrer"></button>
  </footer>
</div>

<script>
const D = {{ .Donnees }};
const M = D.mots, CH = M.champs, DEF = D.defauts;
const cfg = JSON.parse(JSON.stringify(D.cfg));
const base = JSON.parse(JSON.stringify(D.cfg));
const $ = (s) => document.querySelector(s);
const el = (t, c, txt) => { const e = document.createElement(t); if (c) e.className = c;
  if (txt !== undefined) e.textContent = txt; return e; };

/* ═══ mise en forme des valeurs ═══════════════════════════════════ */
const U = M.unites;
function mot(echelle, v){ for (const p of M.echelles[echelle]) if (v <= p.max) return p.n;
  return M.echelles[echelle][M.echelles[echelle].length-1].n; }
function duree(s){
  if (s < 60) return Math.round(s) + ' ' + U.s;
  const m = s/60;
  return (m < 10 ? m.toFixed(1).replace('.0','') : Math.round(m)) + ' ' + U.min;
}
const FORMATS = {
  taille: (v) => '×' + v.toFixed(1),
  chance: (v) => Math.round(v*100) + ' % · ' + mot('chance', v),
  vitesse: (v) => v.toFixed(1) + ' · ' + mot('vitesse', v),
  px: (v) => Math.round(v) + ' ' + U.px,
  duree: (v) => duree(v),
  reveil: (v) => v <= 0 ? U.secoue : duree(v),
  cache: (v) => v <= 0 ? U.toujours : duree(v),
  sieste: (v) => v <= 0 ? U.epuise : duree(v),
  intervalle: (v) => U.a + ' ' + duree(v[0]) + ' ' + U.b + ' ' + duree(v[1]),
};

/* ═══ ce que contient chaque onglet ═══════════════════════════════ */
const PLAN = {
  apparence: [
    { s:'taille',     c:[ {k:'taille', t:'curseur', min:.5, max:2, pas:.1, f:'taille'} ] },
    { s:'personnage', c:[ {k:'planche', t:'cartes'} ] },
    { s:'langue',     c:[ {k:'langue', t:'segment', o:M.options.langue} ] },
    { s:'bulle',      c:[ {k:'echelle_bulle', t:'segment', o:M.options.echelle_bulle} ] },
  ],
  vie: [
    { s:'sante', c:[
      {k:'coeurs', t:'coeurs'},
      {k:'secondes_avant_resurrection', t:'curseur', min:0, max:300, pas:5, f:'reveil'},
    ]},
    { s:'absences', c:[
      {k:'cache_apres_clic', t:'curseur', min:0, max:120, pas:5, f:'cache'},
      {k:'secondes_avant_sieste', t:'curseur', min:0, max:900, pas:15, f:'sieste'},
    ]},
  ],
  caractere: [
    { s:'social', c:[
      {k:'chance_ami', t:'curseur', min:0, max:1, pas:.05, f:'chance'},
      {k:'distance_arret', t:'curseur', min:16, max:240, pas:4, f:'px', dep:'chance_ami'},
      {k:'chance_chasse', t:'curseur', min:0, max:1, pas:.05, f:'chance'},
      {k:'chance_vol', t:'curseur', min:0, max:1, pas:.05, f:'chance'},
      {k:'secondes_avant_vie_seule', t:'curseur', min:0, max:120, pas:5, f:'duree'},
    ]},
    { s:'aventure', c:[
      {k:'vitesse', t:'curseur', min:.5, max:8, pas:.1, f:'vitesse'},
      {k:'chance_grimpe', t:'curseur', min:0, max:1, pas:.05, f:'chance'},
      {k:'chance_jeu', t:'curseur', min:0, max:1, pas:.05, f:'chance'},
    ]},
    { s:'corps', c:[
      {k:'chance_repas', t:'curseur', min:0, max:1, pas:.05, f:'chance'},
      {k:'chance_crotte', t:'curseur', min:0, max:1, pas:.05, f:'chance', dep:'chance_repas'},
      {k:'chance_jet_crotte', t:'curseur', min:0, max:1, pas:.05, f:'chance', dep:'chance_crotte'},
      {k:'jet_mode', t:'segment', o:M.options.jet_mode, dep:'chance_jet_crotte'},
    ]},
  ],
  paroles: [
    { s:'bavardage', c:[
      {k:'parle', t:'bascule'},
      {k:'secondes_entre_paroles', t:'double', min:10, max:600, pas:10, f:'intervalle', dep:'parle'},
      {k:'duree_bulle', t:'curseur', min:1, max:20, pas:.5, f:'duree', dep:'parle'},
    ]},
  ],
  appli: [
    { s:'entretien', c:[ {k:'auto_update', t:'bascule'}, {k:'demarrage', t:'bascule'} ] },
    { s:'fichier',   c:[ {k:'', t:'chemin'} ] },
  ],
};

/* ═══ construction des onglets ════════════════════════════════════ */
const controles = [];
let ongletActif = M.onglets[0].cle;

function construire(){
  const nav = $('#onglets'), zone = $('#panneaux');
  for (const o of M.onglets){
    const b = el('button', 'onglet px');
    b.appendChild(el('span', 'ic', o.icone));
    b.appendChild(el('span', null, o.nom));
    b.onclick = () => { ongletActif = o.cle; majOnglets(); };
    b.dataset.cle = o.cle;
    nav.appendChild(b);

    const p = el('div', 'panneau');
    p.dataset.cle = o.cle;
    for (const sec of PLAN[o.cle]){
      const s = el('section', 'section');
      s.appendChild(el('h2', null, M.sections[sec.s]));
      const corps = el('div', 'corps');
      for (const champ of sec.c) corps.appendChild(ligne(champ));
      s.appendChild(corps);
      p.appendChild(s);
    }
    zone.appendChild(p);
  }
  majOnglets();
}

function majOnglets(){
  for (const b of document.querySelectorAll('.onglet'))
    b.classList.toggle('actif', b.dataset.cle === ongletActif);
  for (const p of document.querySelectorAll('.panneau'))
    p.classList.toggle('actif', p.dataset.cle === ongletActif);
  $('#panneaux').scrollTop = 0;
}

// ligne fabrique une ligne de reglage : titre, valeur, controle, explication.
function ligne(champ){
  const r = el('div', 'reglage');
  if (champ.dep) r.classList.add('imbrique');
  const info = CH[champ.k];

  if (champ.t === 'chemin'){ r.appendChild(blocChemin()); return r; }

  const tete = el('div', 'tete');
  tete.appendChild(el('span', 'nom', info.nom));
  const val = el('span', 'val px');
  tete.appendChild(val);
  r.appendChild(tete);

  let ctrl;
  if (champ.t === 'curseur')      ctrl = curseur(champ, val);
  else if (champ.t === 'double')  ctrl = doubleCurseur(champ, val);
  else if (champ.t === 'segment') ctrl = segment(champ);
  else if (champ.t === 'bascule') ctrl = bascule(champ);
  else if (champ.t === 'coeurs')  ctrl = rangeeCoeurs(champ, val);
  else if (champ.t === 'cartes')  ctrl = cartes(champ);

  if (champ.t === 'bascule'){
    tete.appendChild(ctrl.el);            // l'interrupteur vit dans le titre
  } else {
    const boite = el('div', 'ctrl');
    boite.appendChild(ctrl.el);
    r.appendChild(boite);
  }
  r.appendChild(el('p', 'aide', info.aide));

  controles.push({ champ, ligne:r, maj:ctrl.maj });
  return r;
}

/* ═══ les controles ═══════════════════════════════════════════════ */
function curseur(champ, val){
  const c = el('div', 'curseur');
  c.tabIndex = 0;
  const rail = el('div', 'rail'), rempli = el('div', 'remplit'), pouce = el('div', 'pouce');
  rail.appendChild(rempli);
  c.appendChild(rail); c.appendChild(pouce);

  const part = () => (cfg[champ.k] - champ.min) / (champ.max - champ.min);
  const maj = () => {
    const p = Math.max(0, Math.min(1, part()));
    rempli.style.width = 'calc(' + (p*100) + '% - 4px)';
    pouce.style.left = (p*100) + '%';
    if (val) val.textContent = FORMATS[champ.f](cfg[champ.k]);
    c.setAttribute('aria-valuenow', cfg[champ.k]);
  };
  const poser = (x) => {
    const b = c.getBoundingClientRect();
    const p = Math.max(0, Math.min(1, (x - b.left) / b.width));
    const brut = champ.min + p * (champ.max - champ.min);
    const v = Math.round(brut / champ.pas) * champ.pas;
    definir(champ.k, arrondi(Math.max(champ.min, Math.min(champ.max, v))));
  };
  c.addEventListener('pointerdown', (e) => { c.setPointerCapture(e.pointerId); poser(e.clientX); });
  c.addEventListener('pointermove', (e) => { if (e.buttons) poser(e.clientX); });
  c.addEventListener('keydown', (e) => {
    const d = e.key === 'ArrowLeft' ? -1 : e.key === 'ArrowRight' ? 1 : 0;
    if (!d) return;
    e.preventDefault();
    definir(champ.k, arrondi(Math.max(champ.min, Math.min(champ.max, cfg[champ.k] + d*champ.pas))));
  });
  c.setAttribute('role','slider');
  c.setAttribute('aria-label', CH[champ.k].nom);
  return { el:c, maj };
}

// arrondi rattrape les miettes du calcul flottant (0.30000000000000004)
function arrondi(v){ return Math.round(v * 1000) / 1000; }

function doubleCurseur(champ, val){
  const c = el('div', 'doubleCurseur');
  const rail = el('div', 'rail'), rempli = el('div', 'remplit');
  const pa = el('div', 'pouce'), pb = el('div', 'pouce');
  pa.tabIndex = 0; pb.tabIndex = 0;
  rail.appendChild(rempli);
  c.appendChild(rail); c.appendChild(pa); c.appendChild(pb);

  const part = (v) => (v - champ.min) / (champ.max - champ.min);
  const maj = () => {
    const a = part(cfg[champ.k][0]), b = part(cfg[champ.k][1]);
    pa.style.left = (a*100) + '%'; pb.style.left = (b*100) + '%';
    rempli.style.left = 'calc(' + (a*100) + '% + 2px)';
    rempli.style.width = ((b-a)*100) + '%';
    val.textContent = FORMATS[champ.f](cfg[champ.k]);
  };
  const poser = (x) => {
    const bb = c.getBoundingClientRect();
    const p = Math.max(0, Math.min(1, (x - bb.left) / bb.width));
    let v = champ.min + p * (champ.max - champ.min);
    v = Math.round(v / champ.pas) * champ.pas;
    const paire = cfg[champ.k].slice();
    // on deplace la poignee la plus proche, et on garde a <= b
    const i = Math.abs(v - paire[0]) <= Math.abs(v - paire[1]) ? 0 : 1;
    paire[i] = Math.max(champ.min, Math.min(champ.max, v));
    if (paire[0] > paire[1]) paire[1] = paire[0];
    definir(champ.k, paire);
  };
  c.addEventListener('pointerdown', (e) => { c.setPointerCapture(e.pointerId); poser(e.clientX); });
  c.addEventListener('pointermove', (e) => { if (e.buttons) poser(e.clientX); });
  return { el:c, maj };
}

function segment(champ){
  const s = el('div', 'segment');
  const boutons = [];
  for (const o of champ.o){
    const b = el('button', null, o.n);
    b.onclick = () => definir(champ.k, o.v);
    boutons.push([b, o.v]);
    s.appendChild(b);
  }
  const maj = () => { for (const [b, v] of boutons) b.classList.toggle('actif', cfg[champ.k] === v); };
  return { el:s, maj };
}

function bascule(champ){
  const b = el('button', 'bascule');
  b.appendChild(el('i'));
  b.setAttribute('role', 'switch');
  b.setAttribute('aria-label', CH[champ.k].nom);
  b.onclick = () => definir(champ.k, !cfg[champ.k]);
  const maj = () => { b.classList.toggle('on', !!cfg[champ.k]);
    b.setAttribute('aria-checked', cfg[champ.k] ? 'true' : 'false'); };
  return { el:b, maj };
}

function rangeeCoeurs(champ, val){
  const r = el('div', 'rangeeCoeurs');
  const imgs = [];
  for (let i = 1; i <= 9; i++){
    const b = el('button');
    const img = el('img');
    img.alt = '';
    b.appendChild(img);
    b.onclick = () => definir(champ.k, i);
    imgs.push(img);
    r.appendChild(b);
  }
  const maj = () => {
    imgs.forEach((img, i) => { img.src = i < cfg[champ.k] ? D.coeurs.plein : D.coeurs.vide; });
    val.textContent = cfg[champ.k];
  };
  return { el:r, maj };
}

function cartes(champ){
  const g = el('div', 'cartes');
  const items = [];
  for (const p of D.planches){
    const b = el('button', 'carte');
    const vitrine = el('div', 'vitrine');
    const sp = el('div', 'sprite');
    dessinerSprite(sp, p.marche, p.celL, p.celH, 46 / Math.max(p.celL, p.celH));
    animer(sp, p.marche);
    vitrine.appendChild(sp);
    const txt = el('div');
    txt.appendChild(el('div', 'nom', p.nom));
    txt.appendChild(el('div', 'sous', p.sous));
    b.appendChild(vitrine); b.appendChild(txt);
    b.onclick = () => definir(champ.k, p.cle);
    items.push([b, p.cle]);
    g.appendChild(b);
  }
  const maj = () => { for (const [b, v] of items) b.classList.toggle('actif', cfg[champ.k] === v); };
  return { el:g, maj };
}

function blocChemin(){
  const d = el('div');
  const c = el('div', 'chemin');
  const code = el('code', null, D.chemin);
  const b = el('button', 'btn', M.boutons.dossier);
  b.onclick = () => fetch('/dossier', {method:'POST'});
  c.appendChild(code); c.appendChild(b);
  d.appendChild(c);
  const p = el('p', 'aide', M.divers.version + ' ' + D.version);
  d.appendChild(p);
  return d;
}

/* ═══ sprites ═════════════════════════════════════════════════════ */
function dessinerSprite(node, bande, celL, celH, echelle){
  if (!bande || !bande.cadres){ node.style.display = 'none'; return; }
  const w = Math.max(1, Math.round(celL * echelle)), h = Math.max(1, Math.round(celH * echelle));
  node.style.display = '';
  node.style.width = w + 'px';
  node.style.height = h + 'px';
  node.style.backgroundImage = 'url("' + bande.png + '")';
  node.style.backgroundSize = (w * bande.cadres) + 'px ' + h + 'px';
  node.dataset.pas = w;
}
function animer(node, bande){
  if (!bande || !bande.cadres) return;
  let i = 0;
  setInterval(() => {
    i = (i + 1) % bande.cadres;
    node.style.backgroundPositionX = (-i * Number(node.dataset.pas)) + 'px';
  }, Math.max(60, bande.ms));
}

/* ═══ le diorama ══════════════════════════════════════════════════ */
const singe = $('#singe'), scene = $('#scene'), bulle = $('#bulle');
let etatSinge = 'marche', cadre = 0, horloge = 0, decor = 0, coeursRestants = 0, sursis = 0;

function plancheActive(){
  return D.planches.find(p => p.cle === cfg.planche) || D.planches[0];
}
// echelleEcran : le sprite est dessine en pixels physiques par l'application ;
// la page, elle, compte en pixels CSS. Diviser par le ratio de l'ecran remet
// l'apercu a la taille reelle vue par l'utilisateur.
function echelleEcran(){ return cfg.taille / (window.devicePixelRatio || 1); }

function poserSinge(){
  const p = plancheActive();
  const b = p[etatSinge] && p[etatSinge].cadres ? p[etatSinge] : p.repos;
  dessinerSprite(singe, b, p.celL, p.celH, echelleEcran());
  // ses pieds se posent sur le sol, quelle que soit la planche
  scene.style.bottom = (29 - p.pied * echelleEcran()) + 'px';
  singe.style.backgroundPositionX = '0px';
  cadre = 0;
  majCoeursTete();
}

function majCoeursTete(){
  const z = $('#coeursTete');
  z.innerHTML = '';
  if (etatSinge === 'meurt') return;
  if (coeursRestants > cfg.coeurs) coeursRestants = cfg.coeurs;
  for (let i = 0; i < cfg.coeurs; i++){
    const img = el('img');
    img.src = i < coeursRestants ? D.coeurs.plein : D.coeurs.vide;
    img.alt = '';
    z.appendChild(img);
  }
}

function battement(dt){
  const p = plancheActive();
  const b = p[etatSinge] && p[etatSinge].cadres ? p[etatSinge] : p.repos;
  if (!b || !b.cadres) return;

  horloge += dt;
  const ms = Math.max(60, b.ms);
  if (horloge >= ms){
    horloge = 0;
    const fige = etatSinge === 'meurt';
    if (!fige || cadre < b.cadres - 1) cadre = (cadre + 1) % b.cadres;
    singe.style.backgroundPositionX = (-cadre * Number(singe.dataset.pas || 0)) + 'px';
  }

  if (etatSinge === 'marche'){
    decor += dt * 0.045 * cfg.vitesse;
    $('#c3').style.backgroundPositionX = (-decor * 0.25) + 'px';
    $('#c2').style.backgroundPositionX = (-decor * 0.5) + 'px';
    $('#c1').style.backgroundPositionX = (-decor) + 'px';
    $('#c0').style.backgroundPositionX = (-decor * 1.6) + 'px';
  }

  if (sursis > 0){
    sursis -= dt;
    if (sursis <= 0 && etatSinge !== 'meurt') changerEtat('marche');
  }
}

function changerEtat(e){
  if (etatSinge === e) return;
  etatSinge = e;
  poserSinge();
}

// taper : chaque clic lui coute un coeur ; a terre, un clic le releve.
scene.onclick = () => {
  if (etatSinge === 'meurt'){
    coeursRestants = cfg.coeurs;
    changerEtat('marche');
    for (let i = 0; i < 3; i++) envolerCoeur(true);
    return;
  }
  coeursRestants = Math.max(0, coeursRestants - 1);
  envolerCoeur(false);
  if (coeursRestants === 0){ changerEtat('meurt'); majCoeursTete(); return; }
  changerEtat('touche');
  sursis = 500;
  majCoeursTete();
};

function envolerCoeur(plein){
  const e = el('div', 'envol');
  const img = el('img');
  img.src = plein ? D.coeurs.plein : D.coeurs.vide;
  img.alt = '';
  e.appendChild(img);
  e.style.left = (10 + Math.floor(Math.random()*24)) + 'px';
  e.style.bottom = '60%';
  e.style.setProperty('--dx', (Math.floor(Math.random()*24) - 12) + 'px');
  scene.appendChild(e);
  setTimeout(() => e.remove(), 1000);
}

/* les humeurs, lues sur le vrai singe */
const ORDRE = ['energie','ennui','bonheur','peur'];
const barres = {};
function construireHumeurs(){
  const z = $('#humeurs');
  for (const k of ORDRE){
    const l = el('div', 'humeur');
    l.appendChild(el('span', 'nom px', M.humeurs[k]));
    const j = el('div', 'jauge');
    const cases = [];
    for (let i = 0; i < 8; i++){ const c = el('i'); cases.push(c); j.appendChild(c); }
    barres[k] = cases;
    l.appendChild(j);
    z.appendChild(l);
  }
  $('#direct').querySelector('span').textContent = M.divers.vivant;
}
function peindreHumeurs(v){
  for (const k of ORDRE){
    const n = Math.round((v[k] || 0) * 8);
    barres[k].forEach((c, i) => {
      c.classList.toggle('on', i < n);
      c.classList.toggle('haut', i < n && i >= 6);
    });
  }
}
async function lireHumeurs(){
  try {
    const r = await fetch('/humeurs');
    const v = await r.json();
    if (v && Object.keys(v).length){ peindreHumeurs(v); $('#direct').hidden = false; }
  } catch (e) { /* le singe n'est pas la : on garde la derniere image */ }
}

/* ═══ etat, enregistrement ════════════════════════════════════════ */
function definir(cle, valeur){ cfg[cle] = valeur; rafraichir(); }

function rafraichir(){
  for (const c of controles){
    c.maj();
    if (c.champ.dep){
      const parent = cfg[c.champ.dep];
      c.ligne.classList.toggle('eteint', !parent);
    }
  }
  poserSinge();
  majPied();
}

function modifiees(){
  let n = 0;
  for (const k of Object.keys(base))
    if (JSON.stringify(base[k]) !== JSON.stringify(cfg[k])) n++;
  return n;
}

function majPied(){
  const n = modifiees();
  const e = $('#etat');
  e.textContent = n === 0 ? M.divers.propre
    : n === 1 ? M.divers.saleUn
    : M.divers.saleN.replace('%d', n);
  e.classList.toggle('sale', n > 0);
  $('#enregistrer').disabled = n === 0;
}

async function enregistrer(){
  $('#enregistrer').disabled = true;
  $('#annuler').disabled = true;
  const e = $('#etat');
  e.textContent = M.divers.redemarre;
  e.classList.add('sale');
  await fetch('/enregistrer', {
    method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(cfg),
  });
}

/* ═══ demarrage ═══════════════════════════════════════════════════ */
construire();
construireHumeurs();
$('#astuce').textContent = M.divers.apercuAide;
$('#defaut').textContent = M.boutons.defaut;
$('#annuler').textContent = M.boutons.annuler;
$('#enregistrer').textContent = M.boutons.enregistrer;
$('#defaut').onclick = () => { Object.assign(cfg, JSON.parse(JSON.stringify(DEF))); rafraichir(); };
$('#annuler').onclick = () => fetch('/annuler', {method:'POST'});
$('#enregistrer').onclick = enregistrer;

coeursRestants = cfg.coeurs;
rafraichir();
peindreHumeurs({energie:.8, ennui:.2, bonheur:.6, peur:0});
lireHumeurs();
setInterval(lireHumeurs, 900);

let dernier = performance.now();
requestAnimationFrame(function tour(t){
  battement(t - dernier);
  dernier = t;
  requestAnimationFrame(tour);
});
</script>
</body>
</html>`
