package main

// La page de reglages : une seule interface HTML, la meme sur toutes les
// plateformes, affichee dans une petite fenetre native (webview — voir
// internal/dialogue). Elle reprend l'habillage du site : ardoise sombre, un
// seul accent ambre, titres en police pixel embarquee dans le binaire.
//
// L'application embarque un serveur HTTP minuscule, accessible uniquement
// depuis la machine (127.0.0.1, port ephemere). Enregistrer reecrit
// config.json (en preservant les cles non exposees) puis relance le singe ;
// Annuler ferme simplement la fenetre.

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"time"
)

// demandeFermeture est posee par le bouton Annuler de la page ; la boucle
// principale la consomme et ferme la fenetre native.
var demandeFermeture atomic.Bool

// demarrerReglagesUI lance le serveur et renvoie l'URL de la page, ou "" si le
// port ne peut pas s'ouvrir.
func demarrerReglagesUI() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Printf("reglages indisponibles : %v", err)
		return ""
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", pageReglages)
	mux.HandleFunc("/enregistrer", enregistrerReglages)
	mux.HandleFunc("/annuler", annulerReglages)
	mux.HandleFunc("/favicon.ico", icone)
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String()
	log.Printf("reglages : %s", url)
	return url
}

// champPage decrit un curseur de l'onglet Caractere.
type champPage struct {
	Cle, Nom, Aide string
	Min, Max, Pas  float64
	Valeur         float64
}

type donneesPage struct {
	FR      bool
	Titre   string
	Version string
	Chemin  string
	Taille  float64
	Vitesse float64
	Langue  string
	Parle   bool
	AutoMaj bool
	Coeurs  int
	Sieste  float64
	Chances []champPage

	// l'apercu anime, les coeurs et la police, tous embarques en data-URI
	Ap          apercu
	ApSrc       template.URL
	CoeurPlein  template.URL
	CoeurVide   template.URL
	PolicePixel template.URL
}

func pageReglages(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	r := chargerReglages()
	fr := enFrancais(r.Langue)
	nom := func(frTxt, enTxt string) string {
		if fr {
			return frTxt
		}
		return enTxt
	}
	d := donneesPage{
		FR:      fr,
		Titre:   nom("Réglages du singe", "Monkey settings"),
		Version: version,
		Chemin:  filepath.Join(dossierConfig(), "config.json"),
		Taille:  valeurOu(r.Taille, 1),
		Vitesse: r.Vitesse,
		Langue:  valeurTexteOu(r.Langue, "auto"),
		Parle:   r.Parle,
		AutoMaj: r.AutoMaj,
		Coeurs:  r.Coeurs,
		Sieste:  r.AvantSieste,
		Chances: []champPage{
			{"chance_ami", nom("Ami du curseur", "Cursor friend"),
				nom("à quel point il colle ton curseur et le suit partout", "how much he clings to your cursor and follows it around"), 0, 1, 0.05, r.ChanceAmi},
			{"chance_chasse", nom("Chasse", "Hunting"),
				nom("le suivi tourne parfois à la poursuite, coups de banane inclus", "following sometimes turns into a chase, banana whacks included"), 0, 1, 0.05, r.ChanceChasse},
			{"chance_vol", nom("Vol du curseur", "Cursor theft"),
				nom("il peut s'enfuir avec ta flèche — secoue la souris pour te libérer", "he may run off with your arrow — shake the mouse to break free"), 0, 1, 0.05, r.ChanceVol},
			{"chance_grimpe", nom("Escalade", "Climbing"),
				nom("il grimpe aux bords de l'écran puis se laisse tomber (aïe)", "he climbs the screen edges then lets go (ouch)"), 0, 1, 0.05, r.ChanceGrimpe},
			{"chance_jeu", nom("Jeu", "Play"),
				nom("des petits bonds partout, juste pour le plaisir", "little bounces around, just for fun"), 0, 1, 0.05, r.ChanceJeu},
			{"chance_crotte", nom("Crottes", "Poops"),
				nom("après un repas, il faut bien que ça sorte…", "after a meal, it has to come out…"), 0, 1, 0.05, r.ChanceCrotte},
			{"chance_jet_crotte", nom("Jet de crottes", "Poop throwing"),
				nom("il ramasse ses vieilles crottes et les jette — parfois sur toi", "he picks up old poops and throws them — sometimes at you"), 0, 2, 0.1, r.ChanceJetCrotte},
		},
		CoeurPlein: template.URL("data:image/png;base64," + coeurBase64(true)),
		CoeurVide:  template.URL("data:image/png;base64," + coeurBase64(false)),
	}
	if p := policePixel(); p != "" {
		d.PolicePixel = template.URL("data:font/woff2;base64," + p)
	}
	if ap, ok := apercuMarche(r.Planche); ok {
		d.Ap = ap
		d.ApSrc = template.URL("data:image/png;base64," + ap.Png)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := gabaritReglages.Execute(w, d); err != nil {
		log.Printf("page reglages : %v", err)
	}
}

func enregistrerReglages(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// on part de la config actuelle : les cles non exposees sont preservees
	r := chargerReglages()
	lireF := func(cle string, dst *float64) {
		if v := req.FormValue(cle); v != "" {
			fmt.Sscanf(v, "%g", dst)
		}
	}
	lireF("taille", &r.Taille)
	lireF("vitesse", &r.Vitesse)
	lireF("secondes_avant_sieste", &r.AvantSieste)
	lireF("chance_ami", &r.ChanceAmi)
	lireF("chance_chasse", &r.ChanceChasse)
	lireF("chance_vol", &r.ChanceVol)
	lireF("chance_grimpe", &r.ChanceGrimpe)
	lireF("chance_jeu", &r.ChanceJeu)
	lireF("chance_crotte", &r.ChanceCrotte)
	lireF("chance_jet_crotte", &r.ChanceJetCrotte)
	if v := req.FormValue("coeurs"); v != "" {
		fmt.Sscanf(v, "%d", &r.Coeurs)
	}
	if l := req.FormValue("langue"); l == "auto" || l == "fr" || l == "en" {
		r.Langue = l
	}
	r.Parle = req.FormValue("parle") != ""
	r.AutoMaj = req.FormValue("auto_update") != ""

	if err := ecrireReglages(r); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fr := enFrancais(r.Langue)
	msg := "Saved! The monkey is restarting…"
	if fr {
		msg = "Enregistré ! Le singe redémarre…"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<meta charset="utf-8"><body style="background:#171a23;color:#e8eaf1;`+
		`font:15px -apple-system,'Segoe UI',system-ui,sans-serif;display:grid;`+
		`place-items:center;height:95vh;margin:0">🐒 %s</body>`, msg)

	go func() {
		// on laisse la reponse partir, puis on renait
		time.Sleep(400 * time.Millisecond)
		redemarrer()
	}()
}

// icone sert le coeur du singe comme icone d'onglet.
func icone(w http.ResponseWriter, req *http.Request) {
	brut, err := base64.StdEncoding.DecodeString(coeurBase64(true))
	if err != nil {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(brut)
}

func annulerReglages(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	demandeFermeture.Store(true)
}

var gabaritReglages = template.Must(template.New("reglages").Parse(`<!doctype html>
<html lang="{{if .FR}}fr{{else}}en{{end}}"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Titre}}</title>
<link rel="icon" href="/favicon.ico">
<style>
{{if .PolicePixel}}
  @font-face{font-family:"Pixel";src:url({{.PolicePixel}}) format("woff2");font-display:swap}
{{end}}
  :root{
    --a:#c98a4b; --a2:#e8b586;
    --bg:#171a23; --pnl:#1e222d; --pnl2:#1a1e28; --deep:#12151d;
    --brd:#323a4c; --ink:#e8eaf1; --mut:#98a0b4;
    --pixel:"Pixel",ui-monospace,monospace;
    --sans:-apple-system,BlinkMacSystemFont,"Segoe UI",system-ui,sans-serif;
    --mono:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;
  }
  *{box-sizing:border-box}
  html,body{height:100%}
  body{margin:0;background:var(--bg);color:var(--ink);
       font:13.5px/1.55 var(--sans);display:flex;flex-direction:column;
       -webkit-font-smoothing:antialiased;
       -webkit-user-select:none;user-select:none;cursor:default}
  :focus-visible{outline:2px solid var(--a2);outline-offset:2px;border-radius:4px}

  header{display:flex;align-items:baseline;gap:9px;padding:15px 20px 10px}
  header h1{margin:0;font:400 11px/1.5 var(--pixel);letter-spacing:.5px}
  header .v{color:var(--mut);font:11px var(--mono)}

  nav{display:flex;gap:2px;padding:0 20px;background:var(--pnl2);
      border-top:1px solid var(--brd);border-bottom:1px solid var(--brd)}
  nav button{appearance:none;background:none;border:0;border-bottom:2px solid transparent;
    color:var(--mut);font:400 8px/1 var(--pixel);letter-spacing:.5px;
    padding:13px 12px 11px;cursor:pointer;margin-bottom:-1px}
  nav button:hover{color:var(--ink)}
  nav button[aria-selected=true]{color:var(--ink);border-bottom-color:var(--a)}

  main{flex:1;overflow-y:auto;padding:4px 20px 16px}
  section[hidden]{display:none}

  .colonnes{display:grid;gap:8px 26px;align-items:start;
    grid-template-columns:repeat(auto-fit,minmax(250px,1fr))}
  .pile>.champ:last-child{border-bottom:0}

  .champ{padding:12px 0;border-bottom:1px solid rgba(255,255,255,.045)}
  .champ:last-child{border-bottom:0}
  .ligne{display:flex;justify-content:space-between;align-items:center;gap:12px}
  .ligne label,.ligne .nom{font-weight:550}
  output{font:600 12.5px var(--mono);color:var(--a2);font-variant-numeric:tabular-nums}
  .aide{color:var(--mut);font-size:11.5px;margin-top:2px}
  input[type=range]{width:100%;margin:9px 0 0;accent-color:var(--a)}
  select{background:var(--deep);color:var(--ink);border:1px solid var(--brd);
    border-radius:7px;padding:6px 9px;font:inherit}

  /* interrupteur : une vraie case a cocher, simplement habillee */
  .bascule{position:relative;width:44px;height:24px;flex:0 0 auto}
  .bascule input{position:absolute;inset:0;width:100%;height:100%;margin:0;opacity:0;cursor:pointer}
  .bascule i{position:absolute;inset:0;border-radius:12px;background:var(--deep);
    border:1px solid var(--brd);transition:background .16s,border-color .16s;pointer-events:none}
  .bascule i::after{content:"";position:absolute;top:2px;left:2px;width:18px;height:18px;
    border-radius:50%;background:var(--ink);transition:transform .16s}
  .bascule input:checked+i{background:var(--a);border-color:var(--a)}
  .bascule input:checked+i::after{transform:translateX(20px)}
  .bascule input:focus-visible+i{outline:2px solid var(--a2);outline-offset:2px}

  /* les coeurs, exactement ceux qu'il porte au-dessus de la tete */
  .coeurs{display:flex;gap:7px;align-items:center;margin-top:10px;flex-wrap:wrap}
  .coeurs button{width:28px;height:24px;padding:0;border:0;background:none;cursor:pointer;
    background-image:url({{.CoeurVide}});background-size:28px 24px;background-repeat:no-repeat;
    image-rendering:pixelated;transition:transform .12s}
  .coeurs button[data-plein]{background-image:url({{.CoeurPlein}})}
  .coeurs button:hover{transform:translateY(-2px)}

  /* l'apercu : il marche a la taille choisie, pose sur son sol */
  .apercu{margin-top:12px;background:var(--pnl);border:1px solid var(--brd);
    border-radius:10px;padding:10px 12px 0;overflow:hidden}
  .colonne-apercu{padding-top:12px}
  .colonne-apercu .apercu{margin-top:0}
  .apercu .t{color:var(--mut);font-size:10.5px;letter-spacing:.06em;text-transform:uppercase}
  .scene{position:relative;display:flex;align-items:flex-end;justify-content:center;
    height:min(calc(var(--hmax,120px) + 22px), 190px)}
  .singe{position:relative;bottom:10px;background-repeat:no-repeat;image-rendering:pixelated}
  .sol{position:absolute;left:-12px;right:-12px;bottom:0;height:8px;
    background:var(--pnl2);border-top:1px solid var(--brd)}

  footer{display:flex;align-items:center;gap:10px;padding:11px 20px;
    border-top:1px solid var(--brd);background:var(--pnl2)}
  footer .note{flex:1;min-width:0;color:var(--mut);font-size:11.5px}
  footer .chemin{display:block;font:10px var(--mono);color:#4b5468;
    white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
  .btn{font:600 13px var(--sans);border-radius:8px;padding:9px 15px;cursor:pointer;
    border:1px solid var(--brd);background:var(--pnl);color:var(--ink)}
  .btn:hover{border-color:var(--mut)}
  .btn.pri{background:var(--a);border-color:var(--a);color:#14161c}
  .btn.pri:hover{background:var(--a2);border-color:var(--a2)}

  @media (prefers-reduced-motion: reduce){*{transition-duration:.001ms !important}}
</style></head>
<body>
<header><h1>🐒 {{.Titre}}</h1><span class="v">v{{.Version}}</span></header>

<nav role="tablist">
  <button type="button" role="tab" aria-selected="true"  data-onglet="apparence">{{if .FR}}APPARENCE{{else}}APPEARANCE{{end}}</button>
  <button type="button" role="tab" aria-selected="false" data-onglet="vie">{{if .FR}}VIE{{else}}LIFE{{end}}</button>
  <button type="button" role="tab" aria-selected="false" data-onglet="caractere">{{if .FR}}CARACTÈRE{{else}}CHARACTER{{end}}</button>
</nav>

<form id="f" method="post" action="/enregistrer" style="display:contents">
<main>

  <section id="apparence" role="tabpanel">
    <div class="colonnes">
    <div class="pile">
    <div class="champ">
      <div class="ligne"><label for="taille">{{if .FR}}Taille{{else}}Size{{end}}</label>
        <output id="vtaille">{{printf "%.2f" .Taille}}</output></div>
      <div class="aide">{{if .FR}}sa taille à l'écran — 1 = normale, 0.5 = mini, 2 = géant{{else}}his on-screen size — 1 = normal, 0.5 = tiny, 2 = giant{{end}}</div>
      <input type="range" id="taille" name="taille" min="0.5" max="2" step="0.05" value="{{.Taille}}">
    </div>

    <div class="champ">
      <div class="ligne"><label for="vitesse">{{if .FR}}Vitesse{{else}}Speed{{end}}</label>
        <output id="vvitesse">{{printf "%.1f" .Vitesse}}</output></div>
      <div class="aide">{{if .FR}}à quelle allure il se déplace sur le bureau{{else}}how fast he moves around the desktop{{end}}</div>
      <input type="range" id="vitesse" name="vitesse" min="0.5" max="6" step="0.1" value="{{.Vitesse}}">
    </div>

    <div class="champ">
      <div class="ligne"><label for="langue">{{if .FR}}Langue{{else}}Language{{end}}</label>
        <select id="langue" name="langue">
          <option value="auto" {{if eq .Langue "auto"}}selected{{end}}>{{if .FR}}auto (système){{else}}auto (system){{end}}</option>
          <option value="fr" {{if eq .Langue "fr"}}selected{{end}}>français</option>
          <option value="en" {{if eq .Langue "en"}}selected{{end}}>English</option>
        </select></div>
      <div class="aide">{{if .FR}}la langue de ses bulles et de ses menus{{else}}the language of his bubbles and menus{{end}}</div>
    </div>

    <div class="champ">
      <div class="ligne"><span class="nom">{{if .FR}}Bulles de dialogue{{else}}Speech bubbles{{end}}</span>
        <span class="bascule"><input type="checkbox" id="parle" name="parle" {{if .Parle}}checked{{end}}
          aria-label="{{if .FR}}Bulles de dialogue{{else}}Speech bubbles{{end}}"><i></i></span></div>
      <div class="aide">{{if .FR}}il commente sa vie de temps en temps{{else}}he comments on his life now and then{{end}}</div>
    </div>

    <div class="champ">
      <div class="ligne"><span class="nom">{{if .FR}}Mises à jour automatiques{{else}}Automatic updates{{end}}</span>
        <span class="bascule"><input type="checkbox" id="auto_update" name="auto_update" {{if .AutoMaj}}checked{{end}}
          aria-label="{{if .FR}}Mises à jour automatiques{{else}}Automatic updates{{end}}"><i></i></span></div>
      <div class="aide">{{if .FR}}il s'améliore tout seul quand une nouvelle version sort{{else}}he upgrades himself when a new version is out{{end}}</div>
    </div>
    </div>

    {{if .ApSrc}}
    <div class="colonne-apercu">
      <div class="apercu">
        <div class="t">{{if .FR}}Aperçu — taille réelle{{else}}Preview — actual size{{end}}</div>
        <div class="scene"><span class="sol"></span><span class="singe" id="apercu" role="img"
          aria-label="{{if .FR}}Le singe à la taille choisie{{else}}The monkey at the chosen size{{end}}"></span></div>
      </div>
      <div class="aide">{{if .FR}}il marche ici comme il marchera sur ton bureau{{else}}he walks here exactly as he will on your desktop{{end}}</div>
    </div>
    {{end}}
  </section>

  <section id="vie" role="tabpanel" hidden>
    <div class="champ">
      <div class="ligne"><span class="nom" id="lbl-coeurs">{{if .FR}}Cœurs{{else}}Hearts{{end}}</span>
        <output id="vcoeurs">{{.Coeurs}}</output></div>
      <div class="aide">{{if .FR}}clics encaissés avant de tomber raide — secoue son corps pour le ranimer{{else}}clicks he takes before dropping dead — shake his body to revive him{{end}}</div>
      <div class="coeurs" role="group" aria-labelledby="lbl-coeurs" id="coeurs"></div>
      <input type="hidden" name="coeurs" id="coeurs-val" value="{{.Coeurs}}">
    </div>

    <div class="champ">
      <div class="ligne"><label for="sieste">{{if .FR}}Sieste après{{else}}Nap after{{end}}</label>
        <output id="vsieste">{{printf "%.0f" .Sieste}} s</output></div>
      <div class="aide">{{if .FR}}secondes sans bouger la souris avant qu'il s'endorme{{else}}seconds of mouse stillness before he falls asleep{{end}}</div>
      <input type="range" id="sieste" name="secondes_avant_sieste" min="10" max="600" step="5" value="{{.Sieste}}">
    </div>
  </section>

  <section id="caractere" role="tabpanel" hidden>
    {{range .Chances}}
    <div class="champ">
      <div class="ligne"><label for="{{.Cle}}">{{.Nom}}</label>
        <output id="v{{.Cle}}">{{printf "%.2f" .Valeur}}</output></div>
      <div class="aide">{{.Aide}}</div>
      <input type="range" id="{{.Cle}}" name="{{.Cle}}" min="{{.Min}}" max="{{.Max}}" step="{{.Pas}}" value="{{.Valeur}}">
    </div>
    {{end}}
  </section>

</main>

<footer>
  <span class="note">{{if .FR}}Il redémarre quand tu enregistres.{{else}}He restarts when you save.{{end}}
    <span class="chemin" title="{{.Chemin}}">{{.Chemin}}</span></span>
  <button type="button" class="btn" id="annuler">{{if .FR}}Annuler{{else}}Cancel{{end}}</button>
  <button type="submit" class="btn pri">{{if .FR}}Enregistrer{{else}}Save{{end}}</button>
</footer>
</form>

<script>
  // onglets
  var onglets = document.querySelectorAll('[data-onglet]');
  onglets.forEach(function (b) {
    b.addEventListener('click', function () {
      onglets.forEach(function (x) {
        var actif = x === b;
        x.setAttribute('aria-selected', String(actif));
        document.getElementById(x.dataset.onglet).hidden = !actif;
      });
    });
  });

  // la valeur s'affiche a cote de chaque curseur
  function suivre(id, sortie, dec, suffixe) {
    var e = document.getElementById(id), s = document.getElementById(sortie);
    if (!e || !s) return;
    var maj = function () { s.value = (+e.value).toFixed(dec) + (suffixe || ''); };
    e.addEventListener('input', maj);
    maj();
  }
  suivre('taille', 'vtaille', 2);
  suivre('vitesse', 'vvitesse', 1);
  suivre('sieste', 'vsieste', 0, ' s');
  {{range .Chances}}suivre('{{.Cle}}', 'v{{.Cle}}', 2);
  {{end}}

  // les coeurs : cliquer le n-ieme regle sa resistance
  (function () {
    var boite = document.getElementById('coeurs'),
        champ = document.getElementById('coeurs-val'),
        sortie = document.getElementById('vcoeurs'), max = 9;
    if (!boite) return;
    function peindre(n) {
      champ.value = n;
      sortie.value = n;
      for (var i = 1; i <= max; i++) {
        var b = boite.children[i - 1];
        if (i <= n) b.setAttribute('data-plein', ''); else b.removeAttribute('data-plein');
        b.setAttribute('aria-pressed', String(i <= n));
      }
    }
    for (var i = 1; i <= max; i++) {
      var b = document.createElement('button');
      b.type = 'button';
      b.setAttribute('aria-label', String(i));
      b.addEventListener('click', (function (n) { return function () { peindre(n); }; })(i));
      boite.appendChild(b);
    }
    peindre(+champ.value || 3);
  })();

  {{if .ApSrc}}
  // l'apercu marche a la taille choisie : la bande d'images defile, et les
  // pixels physiques sont ramenes a la densite de cet ecran-ci
  (function () {
    var el = document.getElementById('apercu'), taille = document.getElementById('taille');
    var CADRES = {{.Ap.Cadres}}, CEL_L = {{.Ap.CelL}}, CEL_H = {{.Ap.CelH}}, MS = {{.Ap.MS}};
    var densite = window.devicePixelRatio || 1, i = 0;
    el.style.backgroundImage = 'url({{.ApSrc}})';
    document.querySelector('.scene').style.setProperty('--hmax', (CEL_H * 2 / densite) + 'px');
    function dessiner() {
      var t = +taille.value, l = CEL_L * t / densite, h = CEL_H * t / densite;
      el.style.width = l + 'px';
      el.style.height = h + 'px';
      el.style.backgroundSize = (l * CADRES) + 'px ' + h + 'px';
      el.style.backgroundPosition = (-i * l) + 'px 0';
    }
    taille.addEventListener('input', dessiner);
    setInterval(function () { i = (i + 1) % CADRES; dessiner(); }, MS);
    dessiner();
  })();
  {{end}}

  document.getElementById('annuler').addEventListener('click', function () {
    fetch('/annuler').catch(function () {});
  });
</script>
</body></html>
`))
