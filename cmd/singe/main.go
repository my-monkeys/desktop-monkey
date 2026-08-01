// Commande singe : une petite mascotte qui vit sur le bureau.
//
// L'application affiche une fenetre sans bordure, sans icone dans la barre des
// taches, toujours au premier plan et traversante : clics et frappes la
// traversent comme si elle n'existait pas. Elle ne recoit donc aucun evenement
// souris, et lit l'etat du curseur directement aupres du systeme.
//
// La fenetre a la taille d'une petite scene (le singe et sa bulle) et se
// deplace avec lui, plutot que de couvrir tout l'ecran : c'est plus leger, et
// un defaut d'affichage reste sans consequence pour l'utilisateur.
package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/my-monkeys/desktop-monkey/internal/bulle"
	"github.com/my-monkeys/desktop-monkey/internal/calque"
	"github.com/my-monkeys/desktop-monkey/internal/coeurs"
	"github.com/my-monkeys/desktop-monkey/internal/langue"
	"github.com/my-monkeys/desktop-monkey/internal/paroles"
	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
	"github.com/my-monkeys/desktop-monkey/internal/souris"
	"github.com/my-monkeys/desktop-monkey/internal/vie"
)

const (
	nomApp = "SingeDeBureau"

	// Dimensions de la scene : assez large pour la plus grande bulle, assez
	// haute pour la bulle posee au-dessus du singe.
	sceneLarg = 420
	sceneHaut = 300

	imagesParSeconde = 60

	// affichage des points de vie apres un coup
	echelleCoeur       = 2
	dureeAfficheCoeurs = 2.8
)

func main() {
	log.SetFlags(log.Ltime)
	ouvrirJournal()

	var planCLI, capture string
	var sansConfig bool
	flag.StringVar(&planCLI, "planche", "", "descripteur de planche a utiliser")
	flag.BoolVar(&sansConfig, "sans-config", false, "ignorer le config.json utilisateur")
	flag.StringVar(&capture, "capture", "", "rendre une image dans ce fichier PNG et quitter (mise au point)")
	flag.Parse()

	r := reglagesParDefaut()
	if !sansConfig {
		r = chargerReglages()
		ecrireConfigExemple()
	}
	if planCLI != "" {
		r.Planche = planCLI
	}

	if capture != "" {
		log.SetOutput(os.Stderr)
		if err := rendreDansFichier(r, capture); err != nil {
			log.Fatalf("capture : %v", err)
		}
		return
	}

	if err := lancer(r); err != nil {
		log.Fatalf("arret : %v", err)
	}
}

// rendreDansFichier compose une image de la scene et l'ecrit sur disque, sans
// ouvrir de fenetre. Permet de verifier le rendu sur n'importe quelle
// plateforme, y compris la ou l'affichage bureau n'est pas implemente.
func rendreDansFichier(r Reglages, chemin string) error {
	s, err := nouvelleScene(r, 1920, 1080, 1040)
	if err != nil {
		return err
	}
	// on laisse passer une seconde de simulation : le fondu d'apparition de la
	// bulle est termine, on juge donc le rendu etabli
	for i := 0; i < imagesParSeconde; i++ {
		s.avancer(1.0 / imagesParSeconde)
	}

	img := image.NewRGBA(image.Rect(0, 0, sceneLarg, sceneHaut))
	s.composer(img)

	// damier gris, pour distinguer transparence et blanc a l'oeil
	fond := image.NewRGBA(img.Bounds())
	for y := 0; y < sceneHaut; y++ {
		for x := 0; x < sceneLarg; x++ {
			c := color.RGBA{0x50, 0x54, 0x5c, 0xff}
			if (x/12+y/12)%2 == 0 {
				c = color.RGBA{0x3c, 0x40, 0x46, 0xff}
			}
			fond.SetRGBA(x, y, c)
		}
	}
	draw.Draw(fond, fond.Bounds(), img, image.Point{}, draw.Over)

	f, err := os.Create(chemin)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Printf("texte=%q resteBulle=%.1f visible=%v", s.texte, s.resteBulle, s.v.Visible())
	return png.Encode(f, fond)
}

func lancer(r Reglages) error {
	// La fenetre appartient au fil qui l'a creee : c'est aussi lui qui doit
	// traiter ses messages, donc toute la boucle reste sur ce fil.
	runtime.LockOSThread()

	larg, haut := souris.TailleEcran()
	bas := souris.BasTravail()
	log.Printf("demarrage | ecran %dx%d | barre des taches a %d | planche %s",
		larg, haut, bas, r.Planche)

	s, err := nouvelleScene(r, larg, haut, bas)
	if err != nil {
		return err
	}

	c, err := calque.Ouvrir(nomApp, sceneLarg, sceneHaut)
	if err != nil {
		return err
	}
	defer c.Fermer()
	log.Printf("fenetre ouverte (%dx%d)", sceneLarg, sceneHaut)

	// les crottes vivent dans leurs propres fenetres : on ne les ouvre qu'ici
	s.fenetresOK = true
	defer s.fermerCrottes()

	image := image.NewRGBA(image.Rect(0, 0, sceneLarg, sceneHaut))
	intervalle := time.Second / imagesParSeconde
	dt := intervalle.Seconds()
	horloge := time.NewTicker(intervalle)
	defer horloge.Stop()

	for range horloge.C {
		if !c.TraiterMessages() {
			log.Printf("fermeture demandee")
			return nil
		}
		s.avancer(dt)
		x, y := s.composer(image)
		if err := c.Afficher(image, x, y); err != nil {
			return err
		}
		s.afficherCrottes()
		s.majTraversance(c, image, x, y)
	}
	return nil
}

// majTraversance rend chaque fenetre capturante uniquement quand le curseur est
// sur un de ses pixels dessines : le clic sur le sprite est alors absorbe, le
// reste du bureau reste cliquable. Sans effet sous Windows, qui teste deja les
// clics au pixel pres.
func (s *scene) majTraversance(c *calque.Calque, img *image.RGBA, fenX, fenY int) {
	cx, cy := souris.Position()
	c.Traversant(!opaqueAutour(img, cx-fenX, cy-fenY, 2))
	for _, cr := range s.crottes {
		cr.cal.Traversant(!s.crotteSousCurseur(cr, cx, cy))
	}
}

// opaqueAutour indique si un pixel dessine se trouve dans un petit rayon autour
// de (x, y), ce qui donne une tolerance sans exiger de viser au pixel pres.
func opaqueAutour(img *image.RGBA, x, y, r int) bool {
	b := img.Bounds()
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			px, py := x+dx, y+dy
			if px < 0 || py < 0 || px >= b.Dx() || py >= b.Dy() {
				continue
			}
			if img.Pix[img.PixOffset(px, py)+3] > 40 {
				return true
			}
		}
	}
	return false
}

// scene tient tout ce qui est affiche et sait se dessiner.
type scene struct {
	r Reglages
	v *vie.Vie
	b *bulle.Bulle
	p *paroles.Recueil

	coeur      *planche.Animation
	coeurActif bool
	coeurX     float64
	coeurY     float64
	coeurMS    float64

	texte      string
	resteBulle float64
	prochaine  float64

	mortTraitee bool
	teinte      *image.RGBA // sprite rougi pendant le flash de degats
	ecranL      int
	alea        *rand.Rand

	pc                *planche.Planche // planche des crottes (optionnelle)
	crotteSol         int              // ordonnee de la base du tas dans sa cellule
	crottes           []*crotte
	boutonCrotteAvant bool
	fenetresOK        bool // true quand on peut ouvrir des fenetres (pas en capture)
}

func nouvelleScene(r Reglages, larg, haut, bas int) (*scene, error) {
	p, err := planche.Charger(ressources.Fichiers, r.Planche)
	if err != nil {
		return nil, err
	}

	reg := vie.ReglagesParDefaut()
	reg.Vitesse = r.Vitesse
	reg.DistArret = r.DistArret
	reg.DistRepart = r.DistArret * 1.45
	reg.AvantSieste = r.AvantSieste
	reg.ChanceRepas = r.ChanceRepas
	reg.ChanceAmi = r.ChanceAmi
	reg.SeuilVieSeule = r.SeuilVieSeule
	reg.ChanceChasse = r.ChanceChasse
	reg.ChanceVol = r.ChanceVol
	reg.ChanceGrimpe = r.ChanceGrimpe
	reg.ChanceJeu = r.ChanceJeu
	reg.ChanceCrotte = r.ChanceCrotte
	reg.Coeurs = r.Coeurs
	reg.Resurrection = r.Resurrection
	reg.CacheApresClic = r.CacheApresClic

	s := &scene{
		r:      r,
		v:      vie.Nouvelle(p, reg, larg, haut, bas),
		ecranL: larg,
		alea:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if s.b, err = bulle.Nouvelle(r.EchelleBulle); err != nil {
		return nil, err
	}

	// L'explosion de coeur et les phrases sont optionnelles : leur absence ne
	// doit pas empecher l'application de demarrer.
	if pc, err := planche.Charger(ressources.Fichiers, "assets/coeur.json"); err == nil {
		s.coeur = pc.Obtenir("explose", "")
	} else {
		log.Printf("explosion de coeur indisponible : %v", err)
	}
	perso := filepath.Join(dossierConfig(), "phrases.json")
	if rec, err := paroles.Charger(ressources.Fichiers, recueilPhrases(r.Langue), perso); err == nil {
		s.p = rec
	} else {
		log.Printf("phrases indisponibles : %v", err)
	}

	s.chargerCrottes()

	s.programmerParole()
	s.parler("bonjour")
	return s, nil
}

func (s *scene) parler(contexte string) {
	if !s.r.Parle || s.p == nil {
		return
	}
	if t, ok := s.p.Pour(contexte); ok {
		s.texte = t
		s.resteBulle = s.r.DureeBulle
	}
}

func (s *scene) programmerParole() {
	a, b := s.r.EntreParoles[0], s.r.EntreParoles[1]
	if b <= a {
		b = a + 1
	}
	s.prochaine = a + s.alea.Float64()*(b-a)
}

func (s *scene) avancer(dt float64) {
	cx, cy := souris.Position()
	bouton := souris.BoutonGauche()
	s.v.Avancer(dt, float64(cx), float64(cy), bouton)

	// le singe tient le curseur : c'est lui qui decide ou il va ; et ses
	// coups de patte repoussent la fleche
	if x, y, ok := s.v.TientCurseur(); ok {
		souris.Placer(int(x), int(y))
	} else if x, y, ok := s.v.PrendrePoussee(); ok {
		souris.Placer(int(x), int(y))
	}

	// crottes : depose quand le singe vient de pondre, puis vie et clics
	if x, y, ok := s.v.PrendreCrotte(); ok {
		s.pondre(x, y)
	}
	s.gererCrottes(dt, cx, cy, bouton)

	// enchainement de la mort : agonie, coeur, disparition
	if s.v.Etat() == vie.Mort && !s.mortTraitee && s.v.AnimationMorteFinie() {
		s.coeurX, s.coeurY = s.v.Centre()
		s.coeurActif, s.coeurMS = true, 0
		s.v.Retomber()
		s.mortTraitee = true
	}
	if e := s.v.Etat(); e != vie.Mort && e != vie.Cache && e != vie.Cadavre {
		s.mortTraitee = false
	}
	if s.coeurActif {
		s.coeurMS += dt * 1000
		if s.coeur == nil || s.coeurMS >= float64(s.coeur.Duree()) {
			s.coeurActif = false
		}
	}

	// paroles : evenements du singe, sinon minuterie
	if e := s.v.PrendreEvenement(); e != "" {
		if e == "ranime" {
			// le retour a la vie a droit a sa propre explosion de coeur
			s.coeurX, s.coeurY = s.v.Centre()
			s.coeurActif, s.coeurMS = true, 0
		}
		s.parler(e)
		s.programmerParole()
	} else if s.v.Visible() && s.v.Etat() != vie.Sieste && s.v.Etat() != vie.Cadavre {
		s.prochaine -= dt
		if s.prochaine <= 0 {
			s.parler(paroles.SelonHeure(time.Now()))
			s.programmerParole()
		}
	}
	if s.resteBulle > 0 {
		s.resteBulle -= dt
	}
}

// composer dessine la scene et renvoie ou placer la fenetre a l'ecran.
func (s *scene) composer(dst *image.RGBA) (int, int) {
	for i := range dst.Pix {
		dst.Pix[i] = 0
	}

	spriteL, spriteH := s.v.Taille()

	// Le singe est pose en bas au centre de la scene ; la fenetre est ensuite
	// decalee pour qu'il tombe a sa position reelle a l'ecran.
	sx := sceneLarg/2 - int(spriteL)/2
	sy := sceneHaut - int(spriteH)
	fenX := int(s.v.X) - sx
	fenY := int(s.v.Y) - sy

	if s.v.Visible() {
		anim, ms := s.v.Animation()
		if img := anim.Image(ms); img != nil {
			if f := s.v.Flash(); f > 0 && int(f*18)%2 == 0 {
				img = s.teinter(img)
			}
			poser(dst, img, sx, sy)
		}
	}

	if s.coeurActif && s.coeur != nil {
		if img := s.coeur.Image(int(s.coeurMS)); img != nil {
			b := img.Bounds()
			poser(dst, img,
				int(s.coeurX)-fenX-b.Dx()/2,
				int(s.coeurY)-fenY-b.Dy()/2)
		}
	}

	// points de vie au-dessus de la tete, le temps de quelques secondes apres
	// un coup ; la bulle eventuelle est poussee d'autant vers le haut
	hautCoeurs := 0
	if d := s.v.DepuisCoup(); d < dureeAfficheCoeurs && s.v.Visible() && s.v.Etat() != vie.Cadavre {
		alpha := 1.0
		if reste := dureeAfficheCoeurs - d; reste < 0.4 {
			alpha = reste / 0.4
		}
		restants, total := s.v.Coeurs()
		cx := s.dansEcran(sceneLarg/2-coeurs.Largeur(total, echelleCoeur)/2,
			coeurs.Largeur(total, echelleCoeur), fenX) + coeurs.Largeur(total, echelleCoeur)/2
		cy := sy + int(s.v.HautDuCorps()-s.v.Y) - coeurs.Hauteur(echelleCoeur) - 3
		if cy < 0 {
			cy = 0
		}
		coeurs.Dessiner(dst, cx, cy, restants, total, echelleCoeur, d, alpha)
		hautCoeurs = coeurs.Hauteur(echelleCoeur) + 5
	}

	if s.resteBulle > 0 && s.texte != "" && s.v.Visible() && s.v.Etat() != vie.Cadavre {
		alpha := 1.0
		if e := s.r.DureeBulle - s.resteBulle; e < 0.25 {
			alpha = e / 0.25
		} else if s.resteBulle < 0.4 {
			alpha = s.resteBulle / 0.4
		}
		larg, haut := s.b.Taille(s.texte)
		bx := s.dansEcran(sceneLarg/2-larg/2, larg, fenX)
		// posee juste au-dessus de la tete, pas du cadre du sprite
		by := sy + int(s.v.HautDuCorps()-s.v.Y) - haut - 1 - hautCoeurs
		if by < 0 {
			by = 0
		}
		s.b.Dessiner(dst, s.texte, bx, by, alpha)
	}

	return fenX, fenY
}

// dansEcran decale une abscisse de scene pour qu'un element de largeur larg
// reste visible quand le singe est colle a un bord de l'ecran.
func (s *scene) dansEcran(x, larg, fenX int) int {
	if fenX+x < 0 {
		x = -fenX
	}
	if fenX+x+larg > s.ecranL {
		x = s.ecranL - fenX - larg
	}
	return x
}

// teinter rougit le sprite, l'espace d'un battement, pour marquer le coup.
// Les pixels sont a alpha premultiplie : la teinte est ponderee par l'alpha.
func (s *scene) teinter(src *image.RGBA) *image.RGBA {
	if s.teinte == nil || s.teinte.Bounds() != src.Bounds() {
		s.teinte = image.NewRGBA(src.Bounds())
	}
	const force = 0.55
	for i := 0; i < len(src.Pix); i += 4 {
		a := float64(src.Pix[i+3])
		s.teinte.Pix[i+0] = uint8(float64(src.Pix[i+0])*(1-force) + 0xff*force*a/255)
		s.teinte.Pix[i+1] = uint8(float64(src.Pix[i+1])*(1-force) + 0x38*force*a/255)
		s.teinte.Pix[i+2] = uint8(float64(src.Pix[i+2])*(1-force) + 0x38*force*a/255)
		s.teinte.Pix[i+3] = src.Pix[i+3]
	}
	return s.teinte
}

// recueilPhrases choisit le fichier de phrases embarque : anglais par
// defaut, francais si le systeme de l'utilisateur l'est (ou si la config le
// force avec "fr" / "en").
func recueilPhrases(reglage string) string {
	if reglage == "fr" || (reglage != "en" && langue.Francais()) {
		return "assets/phrases.fr.json"
	}
	return "assets/phrases.en.json"
}

func poser(dst, src *image.RGBA, x, y int) {
	b := src.Bounds()
	draw.Draw(dst, image.Rect(x, y, x+b.Dx(), y+b.Dy()), src, b.Min, draw.Over)
}

// ouvrirJournal redirige les messages vers un fichier a cote de la config :
// l'application est lancee sans console, ses erreurs seraient invisibles.
func ouvrirJournal() {
	d := dossierConfig()
	if err := os.MkdirAll(d, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(d, "singe.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	log.SetOutput(f)
}
