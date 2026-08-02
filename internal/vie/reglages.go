package vie

// Reglages pilote le comportement. Les durees sont en secondes.
type Reglages struct {
	Vitesse    float64
	MargeBord  float64
	DistArret  float64 // distance a laquelle il s'arrete du curseur
	DistRepart float64 // et a partir de laquelle il repart

	DureeRepos     [2]float64
	DureePromenade [2]float64
	DureeRepas     [2]float64

	// L'amitie : le curseur est son meilleur ami. Tant qu'il vit (a bouge
	// recemment), le rejoindre est de loin son activite preferee, ou qu'il
	// soit a l'ecran. Quand l'ami ne bouge plus, le singe vit sa vie.
	ChanceAmi     float64 // probabilite d'aller voir un curseur encore vivant
	SeuilVieSeule float64 // secondes d'immobilite avant qu'il vaque a ses occupations

	AvantSieste  float64 // inactivite de la souris avant de s'endormir
	ChanceRepas  float64 // probabilite de manger plutot que de flaner
	ChanceJeu    float64 // probabilite de sautiller plutot que de flaner
	ChanceGrimpe float64 // probabilite d'aller escalader un bord de l'ecran
	ChanceCrotte float64 // probabilite de s'accroupir et pondre une crotte

	// La chasse : au lieu de suivre gentiment, il poursuit le curseur et
	// l'attaque quand il le rattrape.
	ChanceChasse  float64 // proportion des poursuites qui tournent a la chasse
	VitesseChasse float64 // multiplicateur de vitesse pendant la chasse
	PorteeAttaque float64 // distance en deca de laquelle son jet bouscule aussi le curseur
	EntreLancers  float64 // secondes entre deux bananes, pendant une chasse
	AbandonChasse float64 // secondes de curseur immobile avant d'abandonner
	DureeChasse   float64 // duree maximale d'une chasse
	ChanceVol     float64 // probabilite de voler le curseur au bout d'une chasse

	// La sante.
	Coeurs       int     // nombre de coups avant la mort
	RegenCoeur   float64 // secondes sans coup pour regagner un coeur (0 = jamais)
	Resurrection float64 // secondes avant que le cadavre se releve (0 = jamais)

	// La peur, apres une resurrection.
	DureeCrainte float64 // secondes de crainte apres etre revenu a la vie
	DistFuite    float64 // distance du curseur qui le fait detaler tant qu'il a peur

	// Duree d'absence apres la mort. A zero, le singe ne disparait pas : son
	// cadavre tombe sur la barre des taches et peut etre deplace a la souris.
	CacheApresClic float64
}

// ReglagesParDefaut renvoie un comportement equilibre.
func ReglagesParDefaut() Reglages {
	return Reglages{
		Vitesse:        2.6,
		MargeBord:      10,
		DistArret:      64,
		DistRepart:     95,
		DureeRepos:     [2]float64{3, 9},
		DureePromenade: [2]float64{4, 11},
		DureeRepas:     [2]float64{4, 7},
		ChanceAmi:      0.85,
		SeuilVieSeule:  10,
		AvantSieste:    180,
		ChanceRepas:    0.18,
		ChanceJeu:      0.10,
		ChanceGrimpe:   0.12,
		ChanceCrotte:   0.07,
		ChanceChasse:   0.35,
		VitesseChasse:  1.7,
		PorteeAttaque:  46,
		EntreLancers:   1.1,
		AbandonChasse:  5,
		DureeChasse:    25,
		ChanceVol:      0.35,
		Coeurs:         3,
		RegenCoeur:     45,
		Resurrection:   0,
		DureeCrainte:   12,
		DistFuite:      170,
		CacheApresClic: 0,
	}
}
