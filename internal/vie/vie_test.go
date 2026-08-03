package vie

import (
	"math"
	"testing"

	"github.com/my-monkeys/desktop-monkey/internal/planche"
	"github.com/my-monkeys/desktop-monkey/internal/ressources"
)

const dt = 1.0 / 60

func nouvelleVie(t *testing.T, r Reglages) *Vie {
	t.Helper()
	p, err := planche.Charger(ressources.Fichiers, "assets/singe2.json")
	if err != nil {
		t.Fatalf("planche : %v", err)
	}
	return Nouvelle(p, r, 1920, 1080, 1040)
}

func TestTroisCoupsPuisMort(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())

	if !v.Coup() {
		t.Fatal("le premier coup n'a pas porte")
	}
	if v.Etat() != Touche {
		t.Fatalf("apres un coup : etat %v, attendu touche", v.Etat())
	}
	if restants, total := v.Coeurs(); restants != 2 || total != 3 {
		t.Fatalf("coeurs %d/%d, attendu 2/3", restants, total)
	}
	if v.PrendreEvenement() != "touche" {
		t.Fatal("evenement touche attendu")
	}

	v.Coup()
	v.Coup()
	if v.Etat() != Mort {
		t.Fatalf("apres trois coups : etat %v, attendu mort", v.Etat())
	}
	if v.PrendreEvenement() != "mort" {
		t.Fatal("evenement mort attendu")
	}
}

func TestCoupIgnoreSurLeCadavre(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	v.Coup()
	v.Coup()
	v.Coup()
	v.Retomber()
	if v.Coup() {
		t.Fatal("un coup sur le cadavre ne doit pas porter")
	}
}

func TestCadavreTombeSurLaBarre(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	v.Coup()
	v.Coup()
	v.Coup()
	for i := 0; i < 600 && !v.AnimationMorteFinie(); i++ {
		v.Avancer(dt, 0, 0, false)
	}
	v.Retomber()
	if v.Etat() != Cadavre {
		t.Fatalf("etat %v, attendu cadavre", v.Etat())
	}

	for i := 0; i < 600; i++ {
		v.Avancer(dt, 0, 0, false)
	}

	a, _ := v.Animation()
	basDuCorps := v.Y + v.hauteur - float64(a.VideEnBas)
	if math.Abs(basDuCorps-1040) > 1 {
		t.Fatalf("le corps repose a y=%.1f, attendu 1040", basDuCorps)
	}
}

func TestCadavreChuteAvecAnimationDeChute(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	v.Coup()
	v.Coup()
	v.Coup()
	for i := 0; i < 600 && !v.AnimationMorteFinie(); i++ {
		v.Avancer(dt, 0, 0, false)
	}
	v.Retomber()

	// en l'air : l'animation de chute, vivante
	for i := 0; i < 5; i++ {
		v.Avancer(dt, 0, 0, false)
	}
	if v.animCourante != "tombe" || v.animFigee {
		t.Fatalf("en chute : anim %q (figee %v), attendu tombe animee", v.animCourante, v.animFigee)
	}

	// pose : retour a la pose morte figee
	for i := 0; i < 600; i++ {
		v.Avancer(dt, 0, 0, false)
	}
	if v.animCourante != "meurt" || !v.animFigee {
		t.Fatalf("au sol : anim %q (figee %v), attendu meurt figee", v.animCourante, v.animFigee)
	}
}

func TestCadavreRelacheRetombe(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	v.Coup()
	v.Coup()
	v.Coup()
	for i := 0; i < 600 && !v.AnimationMorteFinie(); i++ {
		v.Avancer(dt, 0, 0, false)
	}
	v.Retomber()
	for i := 0; i < 600; i++ {
		v.Avancer(dt, 0, 0, false)
	}

	// on l'attrape au centre, on le remonte au milieu de l'ecran, on lache
	cx, cy := v.Centre()
	v.Avancer(dt, cx, cy, true)
	for i := 0; i < 60; i++ {
		v.Avancer(dt, cx, cy-float64(i*8), true)
	}
	if v.Y > 900 {
		t.Fatalf("le cadavre n'a pas suivi la souris (y=%.0f)", v.Y)
	}
	v.Avancer(dt, cx, cy-472, false) // relache
	for i := 0; i < 600; i++ {
		v.Avancer(dt, cx, cy-472, false)
	}

	a, _ := v.Animation()
	basDuCorps := v.Y + v.hauteur - float64(a.VideEnBas)
	if math.Abs(basDuCorps-1040) > 1 {
		t.Fatalf("relache, le corps repose a y=%.1f, attendu 1040", basDuCorps)
	}
}

func TestChasseEtAbandon(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi = 1
	r.ChanceChasse = 1
	r.ChanceVol = 0     // sans quoi il volerait parfois le curseur au lieu d'abandonner
	r.SeuilVieSeule = 3 // l'ami immobile est vite delaisse : pas de re-chasse apres l'abandon
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	// un curseur vivant a cote de lui pendant la fin du repos initial
	cx, cy := v.Centre()
	evts := map[string]bool{}
	for i := 0; i < 30; i++ {
		v.Avancer(dt, cx+40+float64(i%2)*10, cy, false)
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}
	if v.Etat() != Chasse {
		t.Fatalf("etat %v, attendu chasse", v.Etat())
	}
	if !evts["chasse"] {
		t.Fatal("evenement chasse attendu")
	}

	// curseur fige : il finit par se lasser
	for i := 0; i < 60*8; i++ {
		v.Avancer(dt, cx+45, cy, false)
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}
	if v.Etat() == Chasse {
		t.Fatal("la chasse aurait du etre abandonnee")
	}
	if !evts["abandon"] {
		t.Fatal("evenement abandon attendu")
	}
}

func TestGrimpeEtRedescend(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceGrimpe = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu = 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	aGrimpe := false
	for i := 0; i < 60*40; i++ {
		v.Avancer(dt, 5, 5, false)
		if v.Etat() == Grimpe {
			aGrimpe = true
		}
		if aGrimpe && v.Etat() != Grimpe {
			a, _ := v.Animation()
			bas := v.Y + v.hauteur - float64(a.VideEnBas)
			if math.Abs(bas-1040) > 8 {
				t.Fatalf("apres l'escalade, le corps est a y=%.1f, attendu ~1040", bas)
			}
			return
		}
	}
	t.Fatalf("l'escalade n'a pas eu lieu ou ne s'est pas terminee (etat %v)", v.Etat())
}

func TestVolDuCurseurPuisLache(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi, r.ChanceChasse, r.ChanceVol = 1, 1, 1
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	evts := map[string]bool{}
	note := func() {
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}

	// un curseur vivant a cote de lui : la chasse demarre
	cx, cy := v.Centre()
	for i := 0; i < 300 && v.Etat() != Chasse; i++ {
		v.Avancer(dt, cx+40+float64(i%2)*8, cy, false)
		note()
	}
	if v.Etat() != Chasse {
		t.Fatalf("etat %v, attendu chasse", v.Etat())
	}

	// le curseur reste au contact : il finit par le voler
	for i := 0; i < 240 && v.Etat() == Chasse; i++ {
		ncx, ncy := v.Centre()
		v.Avancer(dt, ncx+40, ncy, false)
		note()
	}
	if v.Etat() != Vol {
		t.Fatalf("etat %v, attendu vol", v.Etat())
	}
	if !evts["vole"] {
		t.Fatal("evenement vole attendu")
	}
	if _, _, ok := v.TientCurseur(); !ok {
		t.Fatal("TientCurseur devrait repondre pendant le vol")
	}

	// personne ne se debat : il s'enfuit puis lache la fleche de lui-meme
	for i := 0; i < 600 && v.Etat() == Vol; i++ {
		x, y, _ := v.TientCurseur()
		v.Avancer(dt, x, y, false)
		note()
	}
	if v.Etat() == Vol {
		t.Fatal("le vol aurait du se terminer")
	}
	if !evts["lache"] {
		t.Fatal("evenement lache attendu")
	}
	if _, _, ok := v.TientCurseur(); ok {
		t.Fatal("le curseur devrait etre rendu")
	}
}

func TestSecouerRanime(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	v.Coup()
	v.Coup()
	v.Coup()
	for i := 0; i < 600 && !v.AnimationMorteFinie(); i++ {
		v.Avancer(dt, 0, 0, false)
	}
	v.Retomber()
	for i := 0; i < 600; i++ {
		v.Avancer(dt, 0, 0, false)
	}

	// on l'attrape puis on le secoue frenetiquement
	cx, cy := v.Centre()
	v.Avancer(dt, cx, cy, true)
	evts := map[string]bool{}
	for i := 0; i < 120 && v.Etat() == Cadavre; i++ {
		dx := 40.0
		if i%2 == 1 {
			dx = -40
		}
		v.Avancer(dt, cx+dx, cy, true)
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}
	if v.Etat() == Cadavre {
		t.Fatal("les secousses auraient du le ranimer")
	}
	if !evts["ranime"] {
		t.Fatal("evenement ranime attendu")
	}
	if restants, total := v.Coeurs(); restants != total {
		t.Fatalf("ranime avec %d/%d coeurs, attendu le plein", restants, total)
	}
}

func TestJoueBonditEtSeDeplace(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceJeu = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceGrimpe = 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	for i := 0; i < 300 && v.Etat() != Joue; i++ {
		v.Avancer(dt, 5, 5, false)
	}
	if v.Etat() != Joue {
		t.Fatalf("etat %v, attendu joue", v.Etat())
	}

	x0, y0 := v.X, v.Y
	decolle := false
	for i := 0; i < 90; i++ {
		v.Avancer(dt, 5, 5, false)
		if v.Y < y0-5 {
			decolle = true
		}
	}
	if !decolle {
		t.Fatal("les sauts devraient le faire decoller de sa ligne")
	}
	if math.Abs(v.X-x0) < 30 {
		t.Fatalf("les sauts devraient le deplacer (dx=%.0f)", v.X-x0)
	}

	// il finit toujours par se reposer sur sa ligne de depart
	v.r.ChanceJeu = 0 // sans quoi il enchainerait les sessions sans fin
	for i := 0; i < 600 && v.Etat() == Joue; i++ {
		v.Avancer(dt, 5, 5, false)
	}
	if v.Etat() == Joue {
		t.Fatal("la session de jeu aurait du se terminer")
	}
	if math.Abs(v.Y-y0) > 1 {
		t.Fatalf("il devrait finir sur sa ligne de depart (y=%.1f, depart %.1f)", v.Y, y0)
	}
}

func TestAttaqueRepousseLeCurseur(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi, r.ChanceChasse = 1, 1
	r.ChanceVol = 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	cx, cy := v.Centre()
	for i := 0; i < 300 && v.Etat() != Chasse; i++ {
		v.Avancer(dt, cx+40+float64(i%2)*8, cy, false)
	}
	if v.Etat() != Chasse {
		t.Fatalf("etat %v, attendu chasse", v.Etat())
	}

	// la souris ne bouge que sous ses bananes : chaque jet qui fait mouche la
	// repousse, et on applique la poussee comme le ferait la boucle principale.
	// C'est la scene qui voit l'impact, on le simule ici.
	sx, sy := cx+40.0, cy
	poussees := 0
	evts := map[string]bool{}
	for i := 0; i < 60*8; i++ {
		v.Avancer(dt, sx, sy, false)
		if _, ok := v.PrendreLancer(); ok {
			v.BananeTouche(sx, sy)
		}
		if px, py, ok := v.PrendrePoussee(); ok {
			ncx, ncy := v.Centre()
			if math.Hypot(px-ncx, py-ncy) <= math.Hypot(sx-ncx, sy-ncy) {
				t.Fatal("l'impact devrait eloigner le curseur du singe")
			}
			sx, sy = px, py
			poussees++
		}
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}
	if poussees == 0 {
		t.Fatal("aucune banane n'a repousse le curseur")
	}
	// ces mouvements ne viennent pas de l'utilisateur : il finit par se lasser
	if !evts["abandon"] {
		t.Fatal("la chasse aurait du etre abandonnee malgre les poussees")
	}
}

func TestFuiteApresResurrection(t *testing.T) {
	v := nouvelleVie(t, ReglagesParDefaut())
	v.Coup()
	v.Coup()
	v.Coup()
	for i := 0; i < 600 && !v.AnimationMorteFinie(); i++ {
		v.Avancer(dt, 0, 0, false)
	}
	v.Retomber()
	for i := 0; i < 600; i++ {
		v.Avancer(dt, 0, 0, false)
	}
	cx, cy := v.Centre()
	v.Avancer(dt, cx, cy, true)
	for i := 0; i < 120 && v.Etat() == Cadavre; i++ {
		dx := 40.0
		if i%2 == 1 {
			dx = -40
		}
		v.Avancer(dt, cx+dx, cy, true)
	}
	if v.Etat() == Cadavre {
		t.Fatal("les secousses auraient du le ranimer")
	}

	// la souris reste plantee a cote de lui : panique et prise de distance
	sx, sy := v.Centre()
	sx -= 120
	fui := false
	evts := map[string]bool{}
	for i := 0; i < 60*6; i++ {
		v.Avancer(dt, sx, sy, false)
		if v.Etat() == Fuite {
			fui = true
		}
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}
	if !fui {
		t.Fatal("frais ressuscite, il aurait du fuir la souris")
	}
	if !evts["fuite"] {
		t.Fatal("evenement fuite attendu")
	}
	ncx, ncy := v.Centre()
	if d := math.Hypot(ncx-sx, ncy-sy); d < 150 {
		t.Fatalf("il aurait du garder ses distances (d=%.0f)", d)
	}
}

func TestSansCrainteApresLaPeur(t *testing.T) {
	r := ReglagesParDefaut()
	r.DureeCrainte = 0.5
	v := nouvelleVie(t, r)
	v.Ranimer()
	for i := 0; i < 60; i++ { // une seconde : la peur est passee
		v.Avancer(dt, 5, 5, false)
	}
	cx, cy := v.Centre()
	v.Avancer(dt, cx+100, cy, false)
	if v.Etat() == Fuite {
		t.Fatal("la peur passee, il ne devrait plus fuir")
	}
}

func TestGrimpePauseImmobile(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceGrimpe = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu = 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	enPause := func() bool { return v.Etat() == Grimpe && v.phaseGrimpe == grimpePause }
	for i := 0; i < 60*30 && !enPause(); i++ {
		v.Avancer(dt, 5, 700, false)
	}
	if !enPause() {
		t.Fatalf("la pause d'escalade n'a jamais ete atteinte (etat %v)", v.Etat())
	}
	if !v.animFigee {
		t.Fatal("accroche en haut du mur, l'animation devrait etre figee")
	}

	enChute := func() bool { return v.Etat() == Grimpe && v.phaseGrimpe == grimpeChute }
	for i := 0; i < 60*5 && !enChute(); i++ {
		v.Avancer(dt, 5, 700, false)
	}
	if !enChute() {
		t.Fatal("la chute n'a pas suivi la pause")
	}
	if v.animFigee {
		t.Fatal("en chute, l'animation devrait reprendre vie")
	}
}

func TestGrimpeSansTeleportation(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceGrimpe = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu = 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)
	v.Y = 40 // deja tout en haut : la cible d'escalade tiree sera sous lui

	for i := 0; i < 60*30 && v.Etat() != Grimpe; i++ {
		v.Avancer(dt, 5, 700, false)
	}
	if v.Etat() != Grimpe {
		t.Fatalf("etat %v, attendu grimpe", v.Etat())
	}
	// pendant la montee, l'ordonnee ne doit jamais augmenter d'un coup
	for i := 0; i < 60*20 && v.Etat() == Grimpe && v.phaseGrimpe != grimpeChute; i++ {
		avant := v.Y
		v.Avancer(dt, 5, 700, false)
		if v.phaseGrimpe == grimpeMonte && v.Y > avant+0.01 {
			t.Fatalf("teleportation vers le bas pendant la montee (%.0f -> %.0f)", avant, v.Y)
		}
	}
}

func TestRegeneration(t *testing.T) {
	r := ReglagesParDefaut()
	r.RegenCoeur = 1
	v := nouvelleVie(t, r)
	v.Coup()
	if restants, _ := v.Coeurs(); restants != 2 {
		t.Fatalf("coeurs %d, attendu 2", restants)
	}
	for i := 0; i < 60*3; i++ {
		v.Avancer(dt, 0, 0, false)
	}
	if restants, _ := v.Coeurs(); restants != 3 {
		t.Fatalf("apres l'accalmie : coeurs %d, attendu 3", restants)
	}
}

func TestAmiRejointMemeDeLoin(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi = 1
	r.ChanceChasse = 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	// le curseur vit a l'autre bout de l'ecran : il y va quand meme
	for i := 0; i < 60; i++ {
		v.Avancer(dt, 1850+float64(i%2)*8, 1000, false)
	}
	if v.Etat() != Suivi {
		t.Fatalf("etat %v, attendu suivi d'un ami lointain", v.Etat())
	}
}

func TestVieSeuleQuandLAmiDort(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi = 1
	r.ChanceChasse = 0 // une chasse sur curseur fige durerait 5 s, au-dela de la tolerance du test
	r.SeuilVieSeule = 1
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	// le premier pas de temps voit le curseur "apparaitre" en (500,500), ce
	// qui compte comme un mouvement : on ne juge qu'une fois le seuil passe
	for i := 0; i < 60*20; i++ {
		v.Avancer(dt, 500, 500, false)
		if i > 60*3 {
			if e := v.Etat(); e == Suivi || e == Chasse {
				t.Fatalf("curseur immobile : il ne devrait pas le suivre (etat %v)", e)
			}
		}
	}
}

func TestRetrouvailles(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi = 1
	r.ChanceChasse = 0
	r.ChanceRepas, r.ChanceJeu, r.ChanceGrimpe = 0, 0, 0
	r.SeuilVieSeule = 2
	v := nouvelleVie(t, r)

	// l'ami s'endort : le singe vaque
	for i := 0; i < 60*5; i++ {
		v.Avancer(dt, 500, 500, false)
	}
	if e := v.Etat(); e == Suivi || e == Chasse {
		t.Fatalf("apres une longue immobilite : etat %v", e)
	}

	// l'ami revient a la vie : il accourt sur-le-champ
	evts := map[string]bool{}
	for i := 0; i < 30 && v.Etat() != Suivi; i++ {
		v.Avancer(dt, 500+float64(i)*6, 500, false)
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
	}
	if v.Etat() != Suivi {
		t.Fatalf("l'ami revenu, il devrait accourir (etat %v)", v.Etat())
	}
	if !evts["retrouvailles"] {
		t.Fatal("evenement retrouvailles attendu")
	}
}

func TestPromenadeAnimeLaMarche(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu, r.ChanceGrimpe = 0, 0, 0, 0
	v := nouvelleVie(t, r)
	v.passerA(Promenade)
	v.cibleX, v.cibleY = v.X+400, v.Y+400 // diagonale parfaite

	reposEnMouvement, frames := 0, 0
	for i := 0; i < 60*4 && v.Etat() == Promenade; i++ {
		x0, y0 := v.X, v.Y
		v.Avancer(dt, 5, 5, false)
		if math.Hypot(v.X-x0, v.Y-y0) > 0.5 {
			frames++
			if v.animCourante != "marche" {
				reposEnMouvement++
			}
		}
	}
	if frames < 60 {
		t.Fatalf("la promenade n'a presque pas bouge (%d images)", frames)
	}
	// seules les toutes premieres images, le temps d'accelerer, ont le droit
	// de montrer le repos
	if reposEnMouvement > 4 {
		t.Fatalf("%d/%d images en mouvement sans animation de marche", reposEnMouvement, frames)
	}
}

func TestChuteDEscaladeCouteUnCoeur(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceGrimpe = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu = 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)

	evts := map[string]bool{}
	aGrimpe := false
	for i := 0; i < 60*40; i++ {
		v.Avancer(dt, 5, 700, false)
		if v.Etat() == Grimpe {
			aGrimpe = true
		}
		if e := v.PrendreEvenement(); e != "" {
			evts[e] = true
		}
		if aGrimpe && v.Etat() != Grimpe {
			break
		}
	}
	if !aGrimpe {
		t.Fatal("l'escalade n'a pas eu lieu")
	}
	if restants, _ := v.Coeurs(); restants != 2 {
		t.Fatalf("apres la chute : %d coeurs, attendu 2", restants)
	}
	if !evts["aie_chute"] {
		t.Fatal("evenement aie_chute attendu")
	}
}

func TestNeGrimpePasAvecUnSeulCoeur(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceGrimpe = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu = 0, 0, 0
	r.RegenCoeur = 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)
	v.Coup()
	v.Coup() // il ne reste qu'un coeur

	for i := 0; i < 60*30; i++ {
		v.Avancer(dt, 5, 700, false)
		if v.Etat() == Grimpe {
			t.Fatal("blesse a un coeur, il ne devrait pas risquer l'escalade")
		}
	}
}

func TestDefequePondUneCrotte(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceCrotte = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu, r.ChanceGrimpe = 0, 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	v := nouvelleVie(t, r)
	v.aMange = true // precondition : il faut avoir mange pour pondre

	// curseur fige loin : il finit son repos et va se soulager
	for i := 0; i < 60*5 && v.Etat() != Defeque; i++ {
		v.Avancer(dt, 5, 5, false)
	}
	if v.Etat() != Defeque {
		t.Fatalf("etat %v, attendu defeque", v.Etat())
	}
	// la crotte tombe derriere lui (cote oppose a son regard), pas sous ses pieds
	fx, fy := v.X+v.largeur/2, v.Y+v.hauteur
	if v.direction == "gauche" {
		fx += v.largeur * 0.3
	} else {
		fx -= v.largeur * 0.3
	}

	// il pond a la fin de l'accroupissement
	pondu := false
	for i := 0; i < 60*3; i++ {
		v.Avancer(dt, 5, 5, false)
		if x, y, ok := v.PrendreCrotte(); ok {
			if math.Abs(x-fx) > 2 || math.Abs(y-fy) > 2 {
				t.Fatalf("crotte pondue en (%.0f,%.0f), attendu ~(%.0f,%.0f)", x, y, fx, fy)
			}
			pondu = true
			break
		}
	}
	if !pondu {
		t.Fatal("aucune crotte pondue apres l'accroupissement")
	}
	// il pond a mi-accroupissement, marque une pause, puis s'ecarte
	sorti := false
	for i := 0; i < 60*3; i++ {
		v.Avancer(dt, 5, 5, false)
		if v.Etat() != Defeque {
			sorti = true
			break
		}
	}
	if !sorti {
		t.Fatal("il devrait s'ecarter apres avoir pondu")
	}
}

// Sans avoir mange, il ne pond pas, meme avec ChanceCrotte au maximum.
func TestPasDeCrotteSansRepas(t *testing.T) {
	r := ReglagesParDefaut()
	r.ChanceCrotte = 1
	r.ChanceAmi, r.ChanceRepas, r.ChanceJeu, r.ChanceGrimpe = 0, 0, 0, 0
	r.DureeRepos = [2]float64{0.1, 0.1}
	// promenades courtes : une longue marche retarderait la prochaine decision
	r.DureePromenade = [2]float64{0.2, 0.3}
	v := nouvelleVie(t, r)

	for i := 0; i < 60*4; i++ {
		v.Avancer(dt, 5, 5, false)
		if v.Etat() == Defeque {
			t.Fatal("il pond alors qu'il n'a jamais mange")
		}
	}

	// une fois nourri, il finit par se soulager
	v.aMange = true
	pondu := false
	for i := 0; i < 60*8 && !pondu; i++ {
		v.Avancer(dt, 5, 5, false)
		if v.Etat() == Defeque {
			pondu = true
		}
	}
	if !pondu {
		t.Fatal("apres avoir mange, il devrait pouvoir pondre")
	}
}
