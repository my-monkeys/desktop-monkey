package main

// Le serveur de la fenetre de reglages.
//
// La page est servie par un serveur HTTP minuscule, joignable uniquement depuis
// la machine (127.0.0.1, port ephemere), et affichee dans une petite fenetre
// native (voir internal/dialogue). Tout est embarque — polices, sprites,
// coeurs — donc la fenetre reste juste sans connexion.
//
// Le Go envoie l'etat et les libelles en un seul bloc JSON ; la page se
// construit a partir de la. Enregistrer renvoie le tout en JSON, reecrit
// config.json et relance le singe pour appliquer la taille.

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/my-monkeys/desktop-monkey/internal/ressources"
	"github.com/my-monkeys/desktop-monkey/internal/tray"
)

// demandeFermeture est posee par le bouton Annuler ; la boucle principale la
// consomme et ferme la fenetre native.
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
	mux.HandleFunc("/dossier", montrerDossier)
	mux.HandleFunc("/humeurs", servirHumeurs)
	mux.HandleFunc("/favicon.ico", servirIcone)
	go func() { _ = http.Serve(ln, mux) }()
	url := "http://" + ln.Addr().String()
	log.Printf("reglages : %s", url)
	return url
}

// envoi est la configuration telle que la page la manipule : les cles sont
// celles du fichier de configuration, pour que l'aller-retour reste lisible.
type envoi struct {
	Taille          float64    `json:"taille"`
	Planche         string     `json:"planche"`
	Langue          string     `json:"langue"`
	EchelleBulle    int        `json:"echelle_bulle"`
	Coeurs          int        `json:"coeurs"`
	Resurrection    float64    `json:"secondes_avant_resurrection"`
	CacheApresClic  float64    `json:"cache_apres_clic"`
	AvantSieste     float64    `json:"secondes_avant_sieste"`
	ChanceAmi       float64    `json:"chance_ami"`
	DistArret       float64    `json:"distance_arret"`
	ChanceChasse    float64    `json:"chance_chasse"`
	ChanceVol       float64    `json:"chance_vol"`
	Vitesse         float64    `json:"vitesse"`
	ChanceGrimpe    float64    `json:"chance_grimpe"`
	ChanceJeu       float64    `json:"chance_jeu"`
	SeuilVieSeule   float64    `json:"secondes_avant_vie_seule"`
	ChanceRepas     float64    `json:"chance_repas"`
	ChanceCrotte    float64    `json:"chance_crotte"`
	ChanceJetCrotte float64    `json:"chance_jet_crotte"`
	JetMode         int        `json:"jet_mode"`
	Parle           bool       `json:"parle"`
	EntreParoles    [2]float64 `json:"secondes_entre_paroles"`
	DureeBulle      float64    `json:"duree_bulle"`
	AutoMaj         bool       `json:"auto_update"`
	Demarrage       bool       `json:"demarrage"`
}

func envoiDepuis(r Reglages, demarrage bool) envoi {
	return envoi{
		Taille: r.Taille, Planche: r.Planche, Langue: r.Langue,
		EchelleBulle: r.EchelleBulle, Coeurs: r.Coeurs,
		Resurrection: r.Resurrection, CacheApresClic: r.CacheApresClic,
		AvantSieste: r.AvantSieste, ChanceAmi: r.ChanceAmi, DistArret: r.DistArret,
		ChanceChasse: r.ChanceChasse, ChanceVol: r.ChanceVol, Vitesse: r.Vitesse,
		ChanceGrimpe: r.ChanceGrimpe, ChanceJeu: r.ChanceJeu,
		SeuilVieSeule: r.SeuilVieSeule, ChanceRepas: r.ChanceRepas,
		ChanceCrotte: r.ChanceCrotte, ChanceJetCrotte: r.ChanceJetCrotte,
		JetMode: r.JetMode, Parle: r.Parle, EntreParoles: r.EntreParoles,
		DureeBulle: r.DureeBulle, AutoMaj: r.AutoMaj, Demarrage: demarrage,
	}
}

// appliquer repose la configuration recue sur celle du disque : les cles que la
// page n'expose pas sont preservees.
func (e envoi) appliquer(r Reglages) Reglages {
	r.Taille, r.Planche, r.Langue = e.Taille, e.Planche, e.Langue
	r.EchelleBulle, r.Coeurs = e.EchelleBulle, e.Coeurs
	r.Resurrection, r.CacheApresClic = e.Resurrection, e.CacheApresClic
	r.AvantSieste, r.ChanceAmi, r.DistArret = e.AvantSieste, e.ChanceAmi, e.DistArret
	r.ChanceChasse, r.ChanceVol = e.ChanceChasse, e.ChanceVol
	r.Vitesse, r.ChanceGrimpe, r.ChanceJeu = e.Vitesse, e.ChanceGrimpe, e.ChanceJeu
	r.SeuilVieSeule, r.ChanceRepas = e.SeuilVieSeule, e.ChanceRepas
	r.ChanceCrotte, r.ChanceJetCrotte = e.ChanceCrotte, e.ChanceJetCrotte
	r.JetMode, r.Parle = e.JetMode, e.Parle
	r.EntreParoles, r.DureeBulle, r.AutoMaj = e.EntreParoles, e.DureeBulle, e.AutoMaj
	return r
}

func pageReglages(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	r := chargerReglages()
	fr := enFrancais(r.Langue)
	demarrage := tray.AuDemarrage()

	donnees, err := json.Marshal(map[string]any{
		"fr":       fr,
		"cfg":      envoiDepuis(r, demarrage),
		"defauts":  envoiDepuis(reglagesParDefaut(), demarrage),
		"planches": planchesWeb(fr),
		"coeurs": map[string]string{
			"plein": "data:image/png;base64," + coeurBase64(true),
			"vide":  "data:image/png;base64," + coeurBase64(false),
		},
		"chemin":  filepath.Join(dossierConfig(), "config.json"),
		"version": version,
		"mots":    motsReglages(fr),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := struct {
		Titre       string
		Donnees     template.JS
		PolicePixel template.URL
		PoliceUI    template.URL
	}{Titre: titreFenetre(fr), Donnees: template.JS(donnees)}
	if p := policePixel(); p != "" {
		page.PolicePixel = template.URL("data:font/woff2;base64," + p)
	}
	if p, err := ressourceBase64("assets/pixelify.woff2"); err == nil {
		page.PoliceUI = template.URL("data:font/woff2;base64," + p)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := gabaritReglages.Execute(w, page); err != nil {
		log.Printf("page reglages : %v", err)
	}
}

func ressourceBase64(chemin string) (string, error) {
	brut, err := ressources.Fichiers.ReadFile(chemin)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(brut), nil
}

func enregistrerReglages(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}
	var e envoi
	if err := json.NewDecoder(req.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// le lancement au demarrage ne vit pas dans le fichier de configuration :
	// c'est le systeme qui le retient (base de registre, ou LaunchAgent)
	tray.DefinirDemarrage(e.Demarrage)

	if err := ecrireReglages(e.appliquer(chargerReglages())); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

	// on laisse la reponse partir, puis le singe renait avec ses nouveaux
	// reglages (la taille ne peut pas changer a chaud)
	go func() {
		time.Sleep(400 * time.Millisecond)
		redemarrer()
	}()
}

func annulerReglages(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	demandeFermeture.Store(true)
}

func montrerDossier(w http.ResponseWriter, req *http.Request) {
	ouvrirDossierConfig()
	w.WriteHeader(http.StatusNoContent)
}

// servirHumeurs renvoie les jauges reelles du singe : l'apercu les montre en
// direct, ce ne sont pas des chiffres decoratifs.
func servirHumeurs(w http.ResponseWriter, req *http.Request) {
	out := map[string]float64{}
	for _, j := range humeursActuelles() {
		out[j.Nom] = j.Valeur
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func servirIcone(w http.ResponseWriter, req *http.Request) {
	brut, err := base64.StdEncoding.DecodeString(coeurBase64(true))
	if err != nil {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(brut)
}
