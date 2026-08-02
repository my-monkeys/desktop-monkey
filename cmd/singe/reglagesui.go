package main

// La page de reglages : une seule interface HTML, la meme sur toutes les
// plateformes, affichee dans une petite fenetre native (webview — voir
// internal/dialogue). L'application embarque un serveur HTTP minuscule,
// accessible uniquement depuis la machine (127.0.0.1, port ephemere).
// Enregistrer reecrit config.json (en preservant les cles non exposees) puis
// relance le singe ; Annuler ferme simplement la fenetre.

import (
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
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String()
	log.Printf("reglages : %s", url)
	return url
}

// champPage decrit un curseur de caractere.
type champPage struct {
	Cle, Nom, Aide string
	Min, Max, Pas  float64
	Valeur         float64
}

type donneesPage struct {
	FR       bool
	Titre    string
	Version  string
	Chemin   string
	Taille   float64
	Vitesse  float64
	Langue   string
	Parle    bool
	AutoMaj  bool
	Coeurs   int
	Sieste   float64
	Chances  []champPage
	ApercuB6 template.URL // data-URI du sprite (template.URL : les data: passent)
	ApercuL  int          // pixels physiques par unite de taille
	ApercuH  int
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
	}
	if ap, ok := apercuSinge(r.Planche); ok {
		d.ApercuB6 = template.URL("data:image/png;base64," + ap.Png)
		d.ApercuL, d.ApercuH = ap.BaseL, ap.BaseH
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
	fmt.Fprintf(w, `<meta charset="utf-8"><body style="background:#171a23;color:#e7e9ef;font-family:system-ui;display:grid;place-items:center;height:95vh;margin:0"><p style="font-size:17px">🐒 %s</p></body>`, msg)

	go func() {
		// on laisse la reponse partir, puis on renait
		time.Sleep(400 * time.Millisecond)
		redemarrer()
	}()
}

func annulerReglages(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	demandeFermeture.Store(true)
}

var gabaritReglages = template.Must(template.New("reglages").Parse(`<!doctype html>
<html><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Titre}}</title>
<style>
  :root{
    --bg:#171a23; --panel:#1f2430; --panel2:#242a38; --line:#323a4c;
    --ink:#e8eaf1; --dim:#98a0b4; --accent:#c98a4b; --accent2:#e8b586;
  }
  *{box-sizing:border-box;-webkit-user-select:none;user-select:none}
  html,body{height:100%}
  body{margin:0;background:var(--bg);color:var(--ink);
       font:13.5px/1.5 -apple-system,"Segoe UI",system-ui,sans-serif;
       display:flex;flex-direction:column}

  header{display:flex;align-items:baseline;gap:10px;padding:16px 20px 10px}
  header h1{font-size:16px;font-weight:650;margin:0}
  header .v{color:var(--dim);font-size:11px}

  nav{display:flex;gap:4px;padding:0 20px;border-bottom:1px solid var(--line)}
  nav button{appearance:none;border:0;background:none;color:var(--dim);
    font:600 13px inherit;padding:8px 14px 10px;cursor:pointer;
    border-bottom:2px solid transparent;margin-bottom:-1px}
  nav button.actif{color:var(--ink);border-bottom-color:var(--accent)}
  nav button:hover{color:var(--ink)}

  main{flex:1;overflow-y:auto;padding:14px 20px}
  section{display:none}
  section.actif{display:block}

  .champ{padding:10px 0;border-bottom:1px solid rgba(255,255,255,.04)}
  .champ:last-child{border-bottom:0}
  .ligne{display:flex;justify-content:space-between;align-items:center;gap:12px}
  .ligne label{font-weight:550}
  output{font-variant-numeric:tabular-nums;color:var(--accent2);font-weight:650;font-size:13px}
  .aide{color:var(--dim);font-size:11.5px;margin-top:1px}
  input[type=range]{width:100%;margin:8px 0 0;accent-color:var(--accent)}
  select,input[type=number]{background:var(--panel2);color:var(--ink);
    border:1px solid var(--line);border-radius:7px;padding:5px 8px;font:inherit}
  input[type=checkbox]{width:17px;height:17px;accent-color:var(--accent)}

  .apercu{margin-top:14px;background:var(--panel);border:1px solid var(--line);
    border-radius:10px;padding:10px 12px 0}
  .apercu .t{color:var(--dim);font-size:11px;letter-spacing:.06em;text-transform:uppercase}
  .scene{position:relative;height:calc(18px + var(--hmax));}
  .scene img{position:absolute;left:50%;bottom:8px;transform:translateX(-50%);
    image-rendering:pixelated}
  .sol{position:absolute;left:8px;right:8px;bottom:8px;height:1px;background:var(--line)}

  footer{display:flex;justify-content:flex-end;gap:10px;padding:12px 20px;
    border-top:1px solid var(--line);background:var(--panel)}
  button.btn{font:600 13px inherit;border-radius:8px;padding:9px 16px;cursor:pointer;border:1px solid var(--line)}
  .btn.sec{background:var(--panel2);color:var(--ink)}
  .btn.sec:hover{border-color:var(--dim)}
  .btn.pri{background:var(--accent);border-color:var(--accent);color:#14161c}
  .btn.pri:hover{filter:brightness(1.1)}
</style></head>
<body>
<header><h1>🐒 {{.Titre}}</h1><span class="v">v{{.Version}}</span></header>
<nav>
  <button type="button" class="actif" data-t="apparence">{{if .FR}}Apparence{{else}}Appearance{{end}}</button>
  <button type="button" data-t="vie">{{if .FR}}Vie{{else}}Life{{end}}</button>
  <button type="button" data-t="caractere">{{if .FR}}Caractère{{else}}Character{{end}}</button>
</nav>
<form id="f" method="post" action="/enregistrer" style="display:contents">
<main>
  <section id="apparence" class="actif">
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
      <div class="ligne"><label for="parle">{{if .FR}}Bulles de dialogue{{else}}Speech bubbles{{end}}</label>
        <input type="checkbox" id="parle" name="parle" {{if .Parle}}checked{{end}}></div>
      <div class="aide">{{if .FR}}il commente sa vie de temps en temps{{else}}he comments on his life now and then{{end}}</div>
    </div>
    <div class="champ">
      <div class="ligne"><label for="auto_update">{{if .FR}}Mises à jour automatiques{{else}}Automatic updates{{end}}</label>
        <input type="checkbox" id="auto_update" name="auto_update" {{if .AutoMaj}}checked{{end}}></div>
      <div class="aide">{{if .FR}}il s'améliore tout seul quand une nouvelle version sort{{else}}he upgrades himself when a new version is out{{end}}</div>
    </div>
    {{if .ApercuB6}}
    <div class="apercu">
      <div class="t">{{if .FR}}Aperçu (taille réelle){{else}}Preview (actual size){{end}}</div>
      <div class="scene"><div class="sol"></div><img id="apercu" src="{{.ApercuB6}}" alt=""></div>
    </div>
    {{end}}
  </section>

  <section id="vie">
    <div class="champ">
      <div class="ligne"><label for="coeurs">{{if .FR}}Cœurs{{else}}Hearts{{end}}</label>
        <input type="number" id="coeurs" name="coeurs" min="1" max="9" value="{{.Coeurs}}"></div>
      <div class="aide">{{if .FR}}clics encaissés avant de tomber raide — secoue son corps pour le ranimer{{else}}clicks he takes before dropping dead — shake his body to revive him{{end}}</div>
    </div>
    <div class="champ">
      <div class="ligne"><label for="sieste">{{if .FR}}Sieste après{{else}}Nap after{{end}}</label>
        <output id="vsieste">{{printf "%.0f" .Sieste}} s</output></div>
      <div class="aide">{{if .FR}}secondes sans bouger la souris avant qu'il s'endorme{{else}}seconds of mouse stillness before he falls asleep{{end}}</div>
      <input type="range" id="sieste" name="secondes_avant_sieste" min="10" max="600" step="5" value="{{.Sieste}}">
    </div>
  </section>

  <section id="caractere">
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
  <button type="button" class="btn sec" id="annuler">{{if .FR}}Annuler{{else}}Cancel{{end}}</button>
  <button type="submit" class="btn pri">{{if .FR}}Enregistrer et redémarrer{{else}}Save and restart{{end}}</button>
</footer>
</form>
<script>
  // onglets
  document.querySelectorAll('nav button').forEach(b => b.onclick = () => {
    document.querySelectorAll('nav button').forEach(x => x.classList.remove('actif'));
    document.querySelectorAll('section').forEach(x => x.classList.remove('actif'));
    b.classList.add('actif');
    document.getElementById(b.dataset.t).classList.add('actif');
  });
  // valeurs en direct
  const maj = (id, out, fix, suf) => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('input', () =>
      document.getElementById(out).value = (+el.value).toFixed(fix) + (suf||''));
  };
  maj('taille','vtaille',2); maj('vitesse','vvitesse',1); maj('sieste','vsieste',0,' s');
  {{range .Chances}}maj('{{.Cle}}','v{{.Cle}}',2);
  {{end}}
  // apercu : pixels physiques -> pixels CSS de cet ecran, ancre au sol
  const ap = document.getElementById('apercu');
  if (ap) {
    const bl = {{.ApercuL}}, bh = {{.ApercuH}}, sc = document.querySelector('.scene');
    const hmax = Math.ceil(bh * 2 / devicePixelRatio);
    sc.style.setProperty('--hmax', hmax + 'px');
    const red = () => {
      const t = +document.getElementById('taille').value;
      ap.style.width  = (bl * t / devicePixelRatio) + 'px';
      ap.style.height = (bh * t / devicePixelRatio) + 'px';
    };
    document.getElementById('taille').addEventListener('input', red);
    red();
  }
  // annuler : la fenetre native se ferme
  document.getElementById('annuler').onclick = () => fetch('/annuler').catch(()=>{});
</script>
</body></html>
`))
