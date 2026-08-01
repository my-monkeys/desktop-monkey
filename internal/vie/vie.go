// Package vie contient le comportement du singe : ce qu'il fait, quand, et
// pourquoi il passe d'une activite a une autre.
//
// Le principe est une machine a etats ou chaque activite a une duree tiree au
// hasard dans un intervalle. Quand elle expire, le singe choisit la suivante en
// fonction du contexte : depuis combien de temps la souris n'a pas bouge, ou
// elle se trouve, et un peu de hasard pour qu'il ne soit jamais previsible.
//
// La sante (coeurs, coups recus, mort et cadavre) est dans sante.go.
package vie

import (
	"math"
	"math/rand"
	"time"

	"github.com/my-monkeys/desktop-monkey/internal/planche"
)

// Etat est l'activite en cours.
type Etat int

const (
	Repos     Etat = iota // il ne fait rien de particulier
	Promenade             // il marche vers un point choisi au hasard
	Suivi                 // il suit le curseur
	Chasse                // il poursuit le curseur pour l'attaquer
	Vol                   // il s'enfuit avec le curseur qu'il a vole
	Fuite                 // frais ressuscite, il fuit une souris trop proche
	Grimpe                // il escalade un bord de l'ecran
	Joue                  // il sautille sur place
	Corvee                // il va chercher une crotte pour s'en debarrasser (pilote par la scene)
	Defeque               // il s'accroupit et pond une crotte
	Repas                 // il mange
	Sieste                // il dort, apres une longue inactivite
	Touche                // il encaisse un coup
	Mort                  // il n'a plus de coeur : agonie puis explosion
	Cadavre               // il git sur la barre des taches, et se laisse deplacer
	Cache                 // absent de l'ecran pour un moment
)

func (e Etat) String() string {
	switch e {
	case Repos:
		return "repos"
	case Promenade:
		return "promenade"
	case Suivi:
		return "suivi"
	case Chasse:
		return "chasse"
	case Vol:
		return "vol"
	case Fuite:
		return "fuite"
	case Grimpe:
		return "grimpe"
	case Joue:
		return "joue"
	case Defeque:
		return "defeque"
	case Repas:
		return "repas"
	case Sieste:
		return "sieste"
	case Touche:
		return "touche"
	case Mort:
		return "mort"
	case Cadavre:
		return "cadavre"
	case Cache:
		return "cache"
	}
	return "?"
}

// Vie tient l'etat du singe et le fait evoluer.
type Vie struct {
	r    Reglages
	p    *planche.Planche
	alea *rand.Rand

	X, Y             float64 // coin haut gauche du sprite
	largeur, hauteur float64
	ecranL, ecranH   float64
	basEcran         float64 // sommet de la barre des taches : le sol du cadavre

	etat      Etat
	depuis    float64 // secondes passees dans l'etat courant
	duree     float64 // duree prevue de l'etat courant
	direction string

	cibleX, cibleY float64

	// etat Corvee : cible imposee par la scene (aller chercher/porter une
	// crotte), et minuterie du geste de lancer
	dirX, dirY float64
	arrive     bool
	viseMS     float64

	animMS       float64
	animCourante string

	dernierMvt  float64 // secondes depuis le dernier mouvement de souris
	curseurX    float64
	curseurY    float64
	boutonAvant bool

	vitesse  float64 // vitesse lissee, pour des departs et arrets progressifs
	aAvance  bool    // un deplacement a eu lieu pendant ce pas de temps
	enMarche bool    // en suivi : hysteresis entre s'arreter et repartir

	attaqueDite bool // une seule phrase d'attaque par chasse

	enPortee       float64 // compte a rebours entre deux coups de patte
	volTente       bool    // une seule tentative de vol du curseur par chasse
	volX, volY     float64 // position imposee au curseur vole
	fuiteX, fuiteY float64 // destination de la fuite avec le curseur

	pousseX, pousseY   float64 // ou envoyer le curseur repousse par un coup
	pousseEnAttente    bool    // une poussee attend d'etre appliquee
	attenduX, attenduY float64 // position du curseur si la poussee est appliquee
	attenduValide      bool    // le prochain mouvement peut venir de la poussee

	phaseGrimpe int     // avancement de l'escalade (voir acrobaties.go)
	pauseGrimpe float64 // duree de la pause en haut du mur

	jeuSolY float64 // ligne sur laquelle rebondissent ses sauts
	jeuDir  float64 // sens des bonds, -1 ou 1

	dragXPrec, dragYPrec   float64 // suivi des secousses du cadavre traine
	dragVxPrec, dragVyPrec float64
	secousses              float64

	crainte float64 // temps de peur restant apres une resurrection

	// jauges d'humeur (voir jauges.go) et detection du guili
	energie, ennui float64
	bonheur, peur  float64
	guiliInv       int     // inversions de direction accumulees
	guiliAmp       float64 // amplitude du va-et-vient en cours
	guiliTemps     float64 // age de la fenetre d'accumulation
	guiliCool      float64 // repos entre deux fous rires
	guiliDroite    bool    // direction du dernier mouvement

	aMange       bool    // il a mange depuis sa derniere crotte : il peut pondre
	aPondu       bool    // une crotte vient d'etre pondue, en attente de recolte
	crotteLachee bool    // la crotte de l'accroupissement en cours est deja tombee
	pondX, pondY float64 // ou elle est tombee (point de sol)

	coeursRestants int
	depuisCoup     float64 // depuis le dernier coup recu (affichage des coeurs)
	sansCoup       float64 // idem, mais remis a zero par la regeneration
	flash          float64 // temps restant du flash de degats
	reculX, reculY float64 // recul imprime par le dernier coup
	coteGauche     bool    // alternance de l'animation de coup

	vy        float64 // vitesse verticale du cadavre en chute
	animFigee bool    // le cadavre garde la derniere image de son agonie
	deplace   bool    // le cadavre est en cours de deplacement a la souris
	prise     [2]float64

	// Signale a l'appelant qu'un evenement notable vient de se produire, pour
	// declencher une phrase adaptee.
	Evenement string
}

// Nouvelle cree un singe au centre de l'ecran. basEcran est l'ordonnee du
// sommet de la barre des taches, la ou le cadavre se pose.
func Nouvelle(p *planche.Planche, r Reglages, ecranL, ecranH, basEcran int) *Vie {
	if r.Coeurs < 1 {
		r.Coeurs = 1
	}
	l, h := p.Largeur, p.Hauteur
	v := &Vie{
		r:              r,
		p:              p,
		alea:           rand.New(rand.NewSource(time.Now().UnixNano())),
		largeur:        float64(l),
		hauteur:        float64(h),
		ecranL:         float64(ecranL),
		ecranH:         float64(ecranH),
		basEcran:       float64(basEcran),
		direction:      "bas",
		coeursRestants: r.Coeurs,
		depuisCoup:     1e9,
		energie:        energieDepart,
		ennui:          ennuiDepart,
		bonheur:        bonheurDepart,
		peur:           peurDepart,
	}
	v.X = v.ecranL/2 - v.largeur/2
	v.Y = v.ecranH/2 - v.hauteur/2
	v.passerA(Repos)
	return v
}

// Etat renvoie l'activite en cours.
func (v *Vie) Etat() Etat { return v.etat }

// Centre renvoie le centre du sprite, en pixels ecran.
func (v *Vie) Centre() (float64, float64) {
	return v.X + v.largeur/2, v.Y + v.hauteur/2
}

// Taille renvoie les dimensions affichees du sprite.
func (v *Vie) Taille() (float64, float64) { return v.largeur, v.hauteur }

// HautDuCorps renvoie l'ordonnee du sommet reel du personnage, en ignorant les
// rangees transparentes que l'animation en cours laisse au-dessus de lui.
func (v *Vie) HautDuCorps() float64 {
	a, _ := v.Animation()
	return v.Y + float64(a.VideEnHaut)
}

// Animation renvoie l'animation a afficher et l'avancement dans celle-ci.
func (v *Vie) Animation() (*planche.Animation, int) {
	a := v.p.Obtenir(v.animCourante, v.direction)
	if v.animFigee {
		// derniere image, immobile — Duree() tout court reboucleraient les
		// animations en boucle sur leur premiere image
		return a, a.Duree() - a.MS
	}
	return a, int(v.animMS)
}

func (v *Vie) entre(bornes [2]float64) float64 {
	return bornes[0] + v.alea.Float64()*(bornes[1]-bornes[0])
}

func (v *Vie) passerA(e Etat) {
	v.etat = e
	v.depuis = 0
	v.animMS = 0

	switch e {
	case Repos:
		v.duree = v.entre(v.r.DureeRepos)
		v.animCourante = "repos"
	case Promenade:
		v.duree = v.entre(v.r.DureePromenade)
		v.animCourante = "marche"
		// tirees dans l'espace du centre du sprite, celui de avancerVers —
		// une cible en coordonnees de coin serait inatteignable pres des bords
		m := v.r.MargeBord
		v.cibleX = m + v.largeur/2 + v.alea.Float64()*(v.ecranL-v.largeur-2*m)
		v.cibleY = m + v.hauteur/2 + v.alea.Float64()*(v.ecranH-v.hauteur-2*m)
	case Suivi:
		v.duree = 0 // dure tant que le curseur reste interessant
		v.animCourante = "marche"
		v.enMarche = true
	case Chasse:
		v.duree = v.r.DureeChasse
		v.animCourante = "marche"
		v.attaqueDite = false
		v.volTente = false
		v.enPortee = 0
	case Fuite:
		v.duree = 0 // il fuit tant que la souris reste sur ses talons
		v.animCourante = "marche"
	case Vol:
		v.duree = 2.8 + v.alea.Float64()*1.4
		v.animCourante = "marche"
		// il detale vers l'autre bout de l'ecran, a peu pres
		m := v.r.MargeBord
		v.fuiteX = clamp(v.ecranL-v.X+(v.alea.Float64()-0.5)*300, m, v.ecranL-v.largeur-m)
		v.fuiteY = clamp(v.ecranH-v.Y+(v.alea.Float64()-0.5)*300, m, v.ecranH-v.hauteur-m)
		v.volX, v.volY = v.Centre()
	case Grimpe:
		v.duree = 25 // garde-fou : les phases s'enchainent bien avant
		v.animCourante = "marche"
		v.phaseGrimpe = grimpeApproche
		m := v.r.MargeBord
		if v.X+v.largeur/2 < v.ecranL/2 {
			v.cibleX = m
		} else {
			v.cibleX = v.ecranL - v.largeur - m
		}
		v.cibleY = m + v.alea.Float64()*v.ecranH*0.35
		if v.cibleY > v.Y {
			// deja plus haut que la cible : grimper l'y teleporterait vers le bas
			v.cibleY = v.Y
		}
		v.pauseGrimpe = 1.2 + v.alea.Float64()*1.5
	case Joue:
		v.duree = 1.5 + v.alea.Float64()*1.5
		v.animCourante = choisir(v.p, "saute", "repos")
		v.vy = 0
		v.jeuSolY = v.Y
		v.jeuDir = 1
		if v.alea.Float64() < 0.5 {
			v.jeuDir = -1
		}
	case Defeque:
		// il s'accroupit dos a l'ecran ; faute d'animation dediee, il se met en
		// repos vers le haut. Il pond a mi-parcours, reste un instant, puis
		// s'ecarte (voir le cas Defeque dans Avancer).
		v.duree = 1.5
		v.crotteLachee = false
		v.animCourante = choisir(v.p, "defeque", "repos")
		v.direction = "haut"
	case Repas:
		v.duree = v.entre(v.r.DureeRepas)
		v.animCourante = choisir(v.p, "mange", "repos")
	case Sieste:
		v.duree = 0
		v.animCourante = choisir(v.p, "dort", "repos")
	case Touche:
		v.duree = 0.55
		v.animCourante = choisir(v.p, "touche", "repos")
	case Mort:
		v.duree = 0
		v.animCourante = choisir(v.p, "meurt", "repos")
	case Cadavre:
		v.duree = 0
		v.animCourante = choisir(v.p, "meurt", "repos")
		v.animFigee = true
		v.vy = 0
	case Cache:
		v.duree = v.r.CacheApresClic
		v.animCourante = "repos"
	}
	if e != Cadavre {
		v.animFigee = false
	}
	// sans animation de sommeil dediee, la sieste emprunte la pose allongee de
	// l'agonie (derniere image, figee) : il dort couche, pas assis
	if e == Sieste && !v.p.Action("dort") && v.p.Action("meurt") {
		v.animCourante = "meurt"
		if v.direction != "gauche" && v.direction != "droite" {
			v.direction = "droite" // la pose n'existe qu'en profil
		}
		v.animFigee = true
	}
}

// choisir renvoie la premiere action que la planche connait.
func choisir(p *planche.Planche, essais ...string) string {
	for _, a := range essais {
		if p.Action(a) {
			return a
		}
	}
	return essais[len(essais)-1]
}

// Visible indique s'il faut dessiner le singe.
func (v *Vie) Visible() bool { return v.etat != Cache }

// Avancer fait progresser le comportement de dt secondes.
func (v *Vie) Avancer(dt float64, curseurX, curseurY float64, boutonEnfonce bool) {
	v.depuis += dt
	v.animMS += dt * 1000
	v.depuisCoup += dt
	v.sansCoup += dt
	if v.flash > 0 {
		v.flash -= dt
	}
	if v.crainte > 0 {
		v.crainte -= dt
	}
	v.aAvance = false
	x0, y0 := v.X, v.Y

	// --- activite de la souris ------------------------------------------------
	bougeSeul := math.Hypot(curseurX-v.curseurX, curseurY-v.curseurY) > 2
	if bougeSeul && v.attenduValide &&
		math.Hypot(curseurX-v.attenduX, curseurY-v.attenduY) < 4 {
		bougeSeul = false // ce mouvement vient d'un coup de patte, pas de l'utilisateur
	}
	v.attenduValide = false
	if bougeSeul {
		absence := v.dernierMvt
		v.dernierMvt = 0
		if v.etat == Sieste {
			// il se reveille d'un bond quand la planche le permet
			if v.p.Action("saute") {
				v.passerA(Joue)
			} else {
				v.passerA(Repos)
			}
			v.Evenement = "reveil"
		} else if absence > v.r.SeuilVieSeule && v.crainte <= 0 {
			// l'ami revient apres une longue absence : il accourt
			switch v.etat {
			case Repos, Promenade, Repas:
				if v.alea.Float64() < v.r.ChanceAmi {
					if v.alea.Float64() < v.r.ChanceChasse {
						v.passerA(Chasse)
					} else {
						v.passerA(Suivi)
					}
					v.Evenement = "retrouvailles"
				}
			}
		}
	} else {
		v.dernierMvt += dt
	}

	// humeurs : derive lente, et chatouilles si on secoue la souris pres de lui
	v.majJauges(dt, bougeSeul)
	v.detecterGuili(dt, curseurX-v.curseurX, curseurX, curseurY, bougeSeul)

	v.curseurX, v.curseurY = curseurX, curseurY

	v.regenerer()

	// --- clic sur le singe ----------------------------------------------------
	frontDescendant := boutonEnfonce && !v.boutonAvant
	v.boutonAvant = boutonEnfonce

	if v.etat == Cadavre {
		v.avancerCadavre(dt, curseurX, curseurY, boutonEnfonce, frontDescendant)
		return
	}
	if frontDescendant && v.Visible() && v.etat != Mort && v.dansLeCadre(curseurX, curseurY) {
		v.Coup()
		return
	}

	// frais ressuscite, une souris trop proche le fait detaler
	if v.crainte > 0 {
		switch v.etat {
		case Repos, Promenade, Repas, Joue, Suivi:
			cx, cy := v.Centre()
			if math.Hypot(curseurX-cx, curseurY-cy) < v.r.DistFuite {
				v.passerA(Fuite)
				v.Evenement = "fuite"
			}
		}
	}

	switch v.etat {
	case Cache:
		if v.depuis >= v.duree {
			v.reapparaitre()
		}

	case Mort:
		// l'appelant gere la suite via AnimationMorteFinie / Retomber

	case Sieste:
		// on n'en sort que si la souris bouge (traite plus haut)

	case Touche:
		v.encaisser(dt)

	case Suivi:
		v.suivre(dt, curseurX, curseurY)

	case Chasse:
		v.chasser(dt, curseurX, curseurY)

	case Vol:
		v.enfuir(dt, curseurX, curseurY)

	case Fuite:
		v.fuir(dt, curseurX, curseurY)

	case Grimpe:
		v.grimper(dt)

	case Promenade:
		cx, cy := v.Centre()
		if math.Hypot(v.cibleX-cx, v.cibleY-cy) < 6 || v.depuis >= v.duree {
			v.choisirSuite()
		} else {
			// reaffirmee chaque image : la retrogradation en repos des pas
			// quasi nuls (ci-dessous) ne doit pas lui coller pour de bon
			v.animCourante = "marche"
			v.avancerVers(v.cibleX, v.cibleY, dt, 1)
		}

	case Joue:
		v.sautiller(dt)

	case Corvee:
		v.corvee(dt)

	case Defeque:
		// il pond a mi-accroupissement, la crotte tombe derriere lui (sous son
		// corps ; le singe passe devant, cf calque.PasserDevant)...
		if !v.crotteLachee && v.depuis >= 0.5 {
			recul := v.largeur * 0.3
			v.pondX = v.X + v.largeur/2
			if v.direction == "gauche" {
				v.pondX += recul // il regarde a gauche : la crotte sort a droite
			} else {
				v.pondX -= recul
			}
			v.pondY = v.Y + v.hauteur
			v.aPondu = true
			v.crotteLachee = true
			v.Evenement = "crotte"
		}
		// ...il reste un instant, puis s'ecarte d'un pas pour la reveler
		if v.depuis >= v.duree {
			pas := v.largeur
			cible := v.X + v.largeur/2
			if v.direction == "gauche" {
				cible -= pas
			} else {
				cible += pas
			}
			m := v.r.MargeBord
			v.passerA(Promenade)
			v.cibleX = clamp(cible, m+v.largeur/2, v.ecranL-v.largeur/2-m)
			v.cibleY = v.Y + v.hauteur/2
		}

	case Repos, Repas:
		if v.depuis >= v.duree {
			if v.etat == Repas {
				v.aMange = true // il a mange : il a maintenant de quoi pondre
				v.jaugesRepas() // et ca requinque
			}
			v.choisirSuite()
		}
	}

	// sans deplacement ce pas-ci, l'elan retombe doucement
	if !v.aAvance {
		v.vitesse -= v.vitesse * math.Min(1, 8*dt)
	}

	// une animation de course sur place trahirait l'illusion : quand le pas
	// reel est quasi nul (demarrage, mur, coin), il se tient juste debout
	if v.animCourante == "marche" && math.Hypot(v.X-x0, v.Y-y0) < 25*dt {
		v.animCourante = "repos"
	}
}

// choisirSuite decide de la prochaine activite. Les chances de base (reglages)
// sont ponderees par les jauges d'humeur (voir jauges.go) : un singe qui
// s'ennuie fait des betises, un singe epuise se repose.
func (v *Vie) choisirSuite() {
	// endormissement apres une longue absence de l'utilisateur, ou d'epuisement
	if (v.r.AvantSieste > 0 && v.dernierMvt >= v.r.AvantSieste) || v.energie < 0.15 {
		v.passerA(Sieste)
		v.Evenement = "sieste"
		return
	}

	// le curseur est son meilleur ami : tant qu'il vit, le rejoindre passe
	// avant tout le reste, ou qu'il soit — sauf s'il a encore peur de lui
	if v.crainte <= 0 && v.dernierMvt < v.r.SeuilVieSeule &&
		v.alea.Float64() < pondere(v.r.ChanceAmi, v.facteurAmi()) {
		if v.alea.Float64() < pondere(v.r.ChanceChasse, v.facteurEnnui()) {
			v.passerA(Chasse)
			v.Evenement = "chasse"
		} else {
			v.passerA(Suivi)
		}
		return
	}

	// un singe qui manque d'energie pense davantage a manger
	if v.p.Action("mange") && v.alea.Float64() < pondere(v.r.ChanceRepas, 1.8-v.energie) {
		v.passerA(Repas)
		v.Evenement = "repas"
		return
	}

	if v.p.Action("saute") && v.alea.Float64() < pondere(v.r.ChanceJeu, v.facteurActif()) {
		v.passerA(Joue)
		v.Evenement = "joue"
		return
	}

	// l'escalade se termine par une chute qui coute un coeur : avec un seul,
	// l'instinct de survie l'emporte
	if v.p.Action("grimpe") && v.coeursRestants > 1 &&
		v.alea.Float64() < pondere(v.r.ChanceGrimpe, v.facteurActif()) {
		v.passerA(Grimpe)
		v.Evenement = "grimpe"
		return
	}

	// besoin pressant : il s'accroupit et pond. Mais il faut d'abord avoir
	// mange — pas de crotte sortie de nulle part. L'appelant recolte la crotte.
	if v.aMange && v.r.ChanceCrotte > 0 && v.alea.Float64() < v.r.ChanceCrotte {
		v.aMange = false // ce qu'il avait dans le ventre est evacue
		v.passerA(Defeque)
		return
	}

	if v.etat == Repos {
		v.passerA(Promenade)
	} else {
		v.passerA(Repos)
	}
}

func (v *Vie) reapparaitre() {
	m := v.r.MargeBord
	v.X = m + v.alea.Float64()*(v.ecranL-v.largeur-2*m)
	v.Y = m + v.alea.Float64()*(v.ecranH-v.hauteur-2*m)
	v.boutonAvant = true // ignore un bouton deja enfonce a son retour
	v.coeursRestants = v.r.Coeurs
	v.crainte = v.r.DureeCrainte
	v.passerA(Repos)
	v.Evenement = "retour"
}

// PrendreEvenement renvoie l'evenement en attente et le consomme.
func (v *Vie) PrendreEvenement() string {
	e := v.Evenement
	v.Evenement = ""
	return e
}

// PrendreCrotte signale qu'une crotte vient d'etre pondue et ou (point de sol).
// Renvoie false s'il n'y a rien a recolter. A appeler a chaque pas.
func (v *Vie) PrendreCrotte() (x, y float64, ok bool) {
	if !v.aPondu {
		return 0, 0, false
	}
	v.aPondu = false
	return v.pondX, v.pondY, true
}
