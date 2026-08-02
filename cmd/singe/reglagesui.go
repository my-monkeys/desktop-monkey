package main

// Le menu "Open settings" ouvre une petite page de reglages plutot qu'un JSON
// brut : l'application embarque un serveur HTTP minuscule, accessible
// uniquement depuis la machine (127.0.0.1, port ephemere). La page presente
// des curseurs pour l'essentiel ; enregistrer ecrit config.json (en preservant
// les cles non exposees) puis relance le singe pour tout appliquer.

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// demarrerReglagesUI lance le serveur et renvoie l'URL de la page, ou "" si le
// port ne peut pas s'ouvrir (le menu retombera alors sur le fichier brut).
func demarrerReglagesUI() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Printf("reglages web indisponibles : %v", err)
		return ""
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", pageReglages)
	mux.HandleFunc("/enregistrer", enregistrerReglages)
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String()
	log.Printf("reglages : %s", url)
	return url
}

// champ decrit un curseur de la page.
type champ struct {
	Cle, Nom, Aide string
	Min, Max, Pas  float64
	Valeur         float64
}

type donneesPage struct {
	FR      bool
	Titre   string
	Chemin  string
	Taille  float64
	Vitesse float64
	Langue  string
	Parle   bool
	Coeurs  int
	Sieste  float64
	Chances []champ
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
		Chemin:  filepath.Join(dossierConfig(), "config.json"),
		Taille:  valeurOu(r.Taille, 1),
		Vitesse: r.Vitesse,
		Langue:  r.Langue,
		Parle:   r.Parle,
		Coeurs:  r.Coeurs,
		Sieste:  r.AvantSieste,
		Chances: []champ{
			{"chance_ami", nom("Ami du curseur", "Cursor friend"), nom("à quel point il colle ton curseur et le suit partout", "how much he clings to your cursor and follows it around"), 0, 1, 0.05, r.ChanceAmi},
			{"chance_chasse", nom("Chasse", "Hunting"), nom("le suivi tourne parfois à la poursuite, coups de banane inclus", "following sometimes turns into a chase, banana whacks included"), 0, 1, 0.05, r.ChanceChasse},
			{"chance_vol", nom("Vol du curseur", "Cursor theft"), nom("il peut s'enfuir avec ta flèche — secoue la souris pour te libérer", "he may run off with your arrow — shake the mouse to break free"), 0, 1, 0.05, r.ChanceVol},
			{"chance_grimpe", nom("Escalade", "Climbing"), nom("il grimpe aux bords de l'écran puis se laisse tomber (aïe)", "he climbs the screen edges then lets go (ouch)"), 0, 1, 0.05, r.ChanceGrimpe},
			{"chance_jeu", nom("Jeu", "Play"), nom("des petits bonds partout, juste pour le plaisir", "little bounces around, just for fun"), 0, 1, 0.05, r.ChanceJeu},
			{"chance_crotte", nom("Crottes", "Poops"), nom("après un repas, il faut bien que ça sorte…", "after a meal, it has to come out…"), 0, 1, 0.05, r.ChanceCrotte},
			{"chance_jet_crotte", nom("Jet de crottes", "Poop throwing"), nom("il ramasse ses vieilles crottes et les jette — parfois sur toi", "he picks up old poops and throws them — sometimes at you"), 0, 2, 0.1, r.ChanceJetCrotte},
		},
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

	chemin := filepath.Join(dossierConfig(), "config.json")
	brut, err := json.MarshalIndent(r, "", "  ")
	if err == nil {
		err = os.WriteFile(chemin, brut, 0o644)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fr := enFrancais(r.Langue)
	msg := "Saved! The monkey is restarting… you can close this tab."
	if fr {
		msg = "Enregistré ! Le singe redémarre… tu peux fermer cet onglet."
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<meta charset="utf-8"><body style="background:#1b1f2a;color:#e7e9ef;font-family:system-ui;display:grid;place-items:center;height:95vh"><p style="font-size:18px">🐒 %s</p></body>`, msg)

	// le redemarrage applique tout (taille comprise) ; on laisse la reponse
	// partir avant de mourir
	go func() {
		time.Sleep(400 * time.Millisecond)
		redemarrer()
	}()
}

// redemarrer relance le meme executable et s'efface.
func redemarrer() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("redemarrage impossible : %v", err)
		return
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout, cmd.Stderr = nil, nil
	if err := cmd.Start(); err != nil {
		log.Printf("redemarrage : %v", err)
		return
	}
	os.Exit(0)
}

func valeurOu(v, defaut float64) float64 {
	if v <= 0 {
		return defaut
	}
	return v
}

var gabaritReglages = template.Must(template.New("reglages").Parse(`<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Titre}}</title>
<style>
  :root{ --bg:#1b1f2a; --panel:#232836; --line:#333a4c; --ink:#e7e9ef; --dim:#9aa1b4; --accent:#b5793f; }
  *{box-sizing:border-box}
  body{margin:0;background:var(--bg);color:var(--ink);font:14px/1.5 system-ui,-apple-system,"Segoe UI",sans-serif;
       display:flex;justify-content:center;padding:28px 16px}
  form{width:min(480px,100%)}
  h1{font-size:18px;margin:0 0 4px}
  .sub{color:var(--dim);font-size:12px;margin:0 0 18px;word-break:break-all}
  fieldset{border:1px solid var(--line);border-radius:10px;background:var(--panel);
           padding:14px 16px;margin:0 0 14px}
  legend{padding:0 6px;color:var(--dim);font-size:12px;text-transform:uppercase;letter-spacing:.06em}
  .ligne{display:grid;grid-template-columns:1fr auto;gap:2px 10px;align-items:center;margin:10px 0}
  .ligne label{font-weight:500}
  .aide{grid-column:1/-1;color:var(--dim);font-size:12px;margin-top:-2px}
  output{font-variant-numeric:tabular-nums;color:var(--accent);font-weight:600;min-width:38px;text-align:right}
  input[type=range]{grid-column:1/-1;width:100%;accent-color:var(--accent)}
  select,input[type=number]{background:var(--bg);color:var(--ink);border:1px solid var(--line);
       border-radius:7px;padding:6px 8px;font:inherit}
  .duo{display:flex;gap:10px;align-items:center;justify-content:space-between;margin:10px 0}
  button{width:100%;padding:12px;border:0;border-radius:9px;background:var(--accent);color:#fff;
         font:600 15px inherit;cursor:pointer;margin-top:4px}
  button:hover{filter:brightness(1.08)}
</style>
<form method="post" action="/enregistrer">
  <h1>🐒 {{.Titre}}</h1>
  <p class="sub">{{.Chemin}}</p>

  <fieldset>
    <legend>{{if .FR}}Apparence{{else}}Appearance{{end}}</legend>
    <div class="ligne">
      <label for="taille">{{if .FR}}Taille{{else}}Size{{end}}</label>
      <output id="vtaille">{{printf "%.2f" .Taille}}</output>
      <input type="range" id="taille" name="taille" min="0.5" max="2" step="0.05" value="{{.Taille}}"
             oninput="vtaille.value=(+this.value).toFixed(2)">
      <span class="aide">{{if .FR}}sa taille à l'écran — 1 = normale, 0.5 = mini, 2 = géant{{else}}his on-screen size — 1 = normal, 0.5 = tiny, 2 = giant{{end}}</span>
    </div>
    <div class="ligne">
      <label for="vitesse">{{if .FR}}Vitesse{{else}}Speed{{end}}</label>
      <output id="vvitesse">{{printf "%.1f" .Vitesse}}</output>
      <input type="range" id="vitesse" name="vitesse" min="0.5" max="6" step="0.1" value="{{.Vitesse}}"
             oninput="vvitesse.value=(+this.value).toFixed(1)">
      <span class="aide">{{if .FR}}à quelle allure il se déplace sur le bureau{{else}}how fast he moves around the desktop{{end}}</span>
    </div>
    <div class="duo">
      <label for="langue">{{if .FR}}Langue{{else}}Language{{end}}</label>
      <select id="langue" name="langue">
        <option value="auto" {{if eq .Langue "auto"}}selected{{end}}>{{if .FR}}auto (système){{else}}auto (system){{end}}</option>
        <option value="fr" {{if eq .Langue "fr"}}selected{{end}}>français</option>
        <option value="en" {{if eq .Langue "en"}}selected{{end}}>English</option>
      </select>
    </div>
    <div class="duo">
      <label for="parle">{{if .FR}}Bulles de dialogue{{else}}Speech bubbles{{end}}</label>
      <input type="checkbox" id="parle" name="parle" {{if .Parle}}checked{{end}}>
    </div>
  </fieldset>

  <fieldset>
    <legend>{{if .FR}}Vie{{else}}Life{{end}}</legend>
    <div class="duo">
      <label for="coeurs">{{if .FR}}Cœurs (clics avant K.O.){{else}}Hearts (clicks before K.O.){{end}}</label>
      <input type="number" id="coeurs" name="coeurs" min="1" max="9" value="{{.Coeurs}}">
    </div>
    <div class="ligne">
      <label for="sieste">{{if .FR}}Sieste après (secondes){{else}}Nap after (seconds){{end}}</label>
      <output id="vsieste">{{printf "%.0f" .Sieste}}</output>
      <input type="range" id="sieste" name="secondes_avant_sieste" min="10" max="600" step="5" value="{{.Sieste}}"
             oninput="vsieste.value=this.value">
      <span class="aide">{{if .FR}}secondes sans bouger la souris avant qu'il s'endorme{{else}}seconds of mouse stillness before he falls asleep{{end}}</span>
    </div>
  </fieldset>

  <fieldset>
    <legend>{{if .FR}}Caractère{{else}}Character{{end}}</legend>
    {{range .Chances}}
    <div class="ligne">
      <label for="{{.Cle}}">{{.Nom}}</label>
      <output id="v{{.Cle}}">{{printf "%.2f" .Valeur}}</output>
      <input type="range" id="{{.Cle}}" name="{{.Cle}}" min="{{.Min}}" max="{{.Max}}" step="{{.Pas}}" value="{{.Valeur}}"
             oninput="v{{.Cle}}.value=(+this.value).toFixed(2)">
      <span class="aide">{{.Aide}}</span>
    </div>
    {{end}}
  </fieldset>

  <button>{{if .FR}}Enregistrer — le singe redémarre{{else}}Save — the monkey restarts{{end}}</button>
</form>
`))
