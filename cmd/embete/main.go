//go:build windows

// Une fenetre qui pose une question sans bonne reponse : quel que soit le
// bouton choisi, une fenetre de plus s'ouvre, et celle qu'on vient de cliquer
// reste. Elles s'accumulent donc a l'ecran.
//
// La derniere n'offre plus de choix mais un bouton « Fermer tout », qui se
// derobe des que la souris l'approche. Il finit par se laisser attraper : sans
// quoi il n'y aurait plus d'autre issue que le gestionnaire des taches.
//
//	embete.exe                    dix fenetres, la dixieme etant la derniere
//	embete.exe -max 3             pour essayer sans cliquer neuf fois
//	embete.exe -esquives 0        le bouton ne fuit pas
//	embete.exe -max 0             sans fin, aucune fenetre finale
//
// Les libelles des boutons sont ecrits en dur plutot que repris d'une
// MessageBox : celle-ci les traduit selon la langue de Windows, et afficherait
// « Yes / No » sur un systeme anglais.
//
// Pour tout arreter a distance : Stop-Process -Name embete
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	registerClassExW    = user32.NewProc("RegisterClassExW")
	createWindowExW     = user32.NewProc("CreateWindowExW")
	defWindowProcW      = user32.NewProc("DefWindowProcW")
	showWindow          = user32.NewProc("ShowWindow")
	updateWindow        = user32.NewProc("UpdateWindow")
	getMessageW         = user32.NewProc("GetMessageW")
	translateMessage    = user32.NewProc("TranslateMessage")
	dispatchMessageW    = user32.NewProc("DispatchMessageW")
	postMessageW        = user32.NewProc("PostMessageW")
	postQuitMessage     = user32.NewProc("PostQuitMessage")
	isDialogMessageW    = user32.NewProc("IsDialogMessageW")
	getSystemMetrics    = user32.NewProc("GetSystemMetrics")
	getClientRect       = user32.NewProc("GetClientRect")
	sendMessageW        = user32.NewProc("SendMessageW")
	loadCursorW         = user32.NewProc("LoadCursorW")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
	setTimer            = user32.NewProc("SetTimer")
	killTimer           = user32.NewProc("KillTimer")
	getCursorPos        = user32.NewProc("GetCursorPos")
	screenToClient      = user32.NewProc("ScreenToClient")
	moveWindow          = user32.NewProc("MoveWindow")

	setBkMode      = gdi32.NewProc("SetBkMode")
	getStockObject = gdi32.NewProc("GetStockObject")
	createFontW    = gdi32.NewProc("CreateFontW")

	getModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmClose          = 0x0010
	wmSetFont        = 0x0030
	wmTimer          = 0x0113
	wmCommand        = 0x0111
	wmCtlColorStatic = 0x0138
	wmOuvrir         = 0x8000 // WM_APP : notre propre demande d'ouverture

	wsChild         = 0x40000000
	wsVisible       = 0x10000000
	wsTabStop       = 0x00010000
	wsCaption       = 0x00C00000
	bsDefPushButton = 0x0001
	ssCenter        = 0x00000001

	wsExTopMost       = 0x00000008
	wsExDlgModalFrame = 0x00000001

	swShow          = 5
	idcArrow        = 32512
	fondTransparent = 1
	brosseBlanche   = 0
	largeurEcran    = 0
	hauteurEcran    = 1

	idOui    = 1
	idNon    = 2
	idFermer = 3

	minuteur = 1 // identifiant du minuteur de fuite

	largeur, hauteur   = 400, 200 // fenetres a question
	largeurF, hauteurF = 560, 300 // fenetre finale : de la place pour fuir
	largeurB, hauteurB = 150, 36  // bouton « Fermer tout »
	marge              = 70       // distance a laquelle le bouton s'enfuit
	ecartMin           = 150      // distance minimale visee apres un bond
)

type classeFenetre struct {
	taille       uint32
	style        uint32
	procedure    uintptr
	extraClasse  int32
	extraFenetre int32
	instance     syscall.Handle
	icone        syscall.Handle
	curseur      syscall.Handle
	fond         syscall.Handle
	menu         *uint16
	nom          *uint16
	petiteIcone  syscall.Handle
}

type message struct {
	fenetre syscall.Handle
	code    uint32
	wparam  uintptr
	lparam  uintptr
	temps   uint32
	point   point
}

type point struct{ x, y int32 }

type rectangle struct{ gauche, haut, droite, bas int32 }

var (
	nomClasse = must(syscall.UTF16PtrFromString("EmbeteFenetre"))
	instance  syscall.Handle
	police    uintptr

	maximum        int // 0 : sans fin, aucune fenetre finale
	limiteEsquives int
	ouvertes       int

	finale   uintptr // fenetre au bouton fuyant, 0 tant qu'elle n'existe pas
	bouton   uintptr // le bouton « Fermer tout »
	bx, by   int32   // sa position dans la fenetre finale
	lcF, hcF int32   // zone utile de la fenetre finale
	esquives int
)

func must(p *uint16, err error) *uint16 {
	if err != nil {
		panic(err)
	}
	return p
}

func texte(s string) uintptr {
	return uintptr(unsafe.Pointer(must(syscall.UTF16PtrFromString(s))))
}

// procedure traite les messages de toutes les fenetres.
func procedure(fenetre, code, wparam, lparam uintptr) uintptr {
	switch code {
	case wmCommand:
		switch wparam & 0xFFFF {
		case idOui, idNon:
			// La fenetre est ouverte plus tard, depuis la boucle de messages,
			// et non ici : le bouton detient encore la capture de la souris, et
			// lui passer devant maintenant lui fait renvoyer sa notification
			// en boucle — un seul clic ouvrait ainsi une poignee de fenetres.
			postMessageW.Call(0, wmOuvrir, 0, 0)
		case idFermer:
			postQuitMessage.Call(0) // quitter fait disparaitre toutes les fenetres
		}
		return 0

	case wmTimer:
		esquiver(fenetre)
		return 0

	case wmCtlColorStatic:
		// Sans cela le libelle garderait le gris des controles sur le fond blanc.
		setBkMode.Call(wparam, fondTransparent)
		brosse, _, _ := getStockObject.Call(brosseBlanche)
		return brosse

	case wmClose:
		// Ni croix ni Alt+F4 : aucune ne se ferme d'elle-meme.
		return 0
	}
	r, _, _ := defWindowProcW.Call(fenetre, code, wparam, lparam)
	return r
}

func enregistrerClasse() {
	h, _, _ := getModuleHandleW.Call(0)
	instance = syscall.Handle(h)
	curseur, _, _ := loadCursorW.Call(0, uintptr(idcArrow))
	blanche, _, _ := getStockObject.Call(brosseBlanche)

	classe := classeFenetre{
		style:     0x0003, // CS_HREDRAW | CS_VREDRAW
		procedure: syscall.NewCallback(procedure),
		instance:  instance,
		curseur:   syscall.Handle(curseur),
		fond:      syscall.Handle(blanche),
		nom:       nomClasse,
	}
	classe.taille = uint32(unsafe.Sizeof(classe))

	if r, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(&classe))); r == 0 {
		panic(fmt.Sprintf("enregistrement de la classe impossible : %v", err))
	}

	// Segoe UI : la police des fenetres de Windows. La hauteur negative se
	// compte en pixels de caractere plutot qu'en cellule complete.
	police, _, _ = createFontW.Call(
		^uintptr(15), 0, 0, 0, 400, 0, 0, 0, // -16, poids normal
		1, 0, 0, 5, 0, // DEFAULT_CHARSET, qualite CLEARTYPE
		texte("Segoe UI"))
}

func controle(parent uintptr, classe, libelle string, style, x, y, l, h int32, id uintptr) uintptr {
	enfant, _, _ := createWindowExW.Call(
		0, texte(classe), texte(libelle),
		uintptr(wsChild|wsVisible|uint32(style)),
		uintptr(x), uintptr(y), uintptr(l), uintptr(h),
		parent, id, uintptr(instance), 0)
	sendMessageW.Call(enfant, wmSetFont, police, 1)
	return enfant
}

// creer pose une fenetre a la position donnee et renvoie sa zone utile, plus
// petite que la fenetre elle-meme a cause de la barre de titre et des bordures.
func creer(titre string, x, y, l, h int32) (uintptr, int32, int32) {
	fenetre, _, err := createWindowExW.Call(
		wsExTopMost|wsExDlgModalFrame,
		uintptr(unsafe.Pointer(nomClasse)), texte(titre),
		wsCaption, // ni reduire, ni agrandir, ni fermer
		uintptr(x), uintptr(y), uintptr(l), uintptr(h),
		0, 0, uintptr(instance), 0)
	if fenetre == 0 {
		panic(fmt.Sprintf("creation de la fenetre impossible : %v", err))
	}
	var zone rectangle
	getClientRect.Call(fenetre, uintptr(unsafe.Pointer(&zone)))
	return fenetre, zone.droite, zone.bas
}

func afficher(fenetre uintptr) {
	showWindow.Call(fenetre, swShow)
	setForegroundWindow.Call(fenetre)
	updateWindow.Call(fenetre)
}

// place tire une position au hasard, en laissant la fenetre entierement
// visible et en evitant la barre des taches.
func place(l, h int32) (int32, int32) {
	le, _, _ := getSystemMetrics.Call(largeurEcran)
	he, _, _ := getSystemMetrics.Call(hauteurEcran)

	libreX, libreY := int(le)-int(l), int(he)-int(h)-60
	if libreX < 1 {
		libreX = 1
	}
	if libreY < 1 {
		libreY = 1
	}
	return int32(rand.Intn(libreX)), int32(rand.Intn(libreY))
}

// ouvrir ajoute une fenetre. Les precedentes restent ou elles sont.
func ouvrir() {
	if maximum > 0 && ouvertes >= maximum {
		// Plus rien a ouvrir : on rappelle ou se trouve la sortie.
		if finale != 0 {
			afficher(finale)
		}
		return
	}
	ouvertes++
	if maximum > 0 && ouvertes == maximum {
		ouvrirFinale()
		return
	}

	titre := "Question"
	if ouvertes > 1 {
		titre = fmt.Sprintf("Question (%d)", ouvertes)
	}
	x, y := place(largeur, hauteur)
	fenetre, lc, hc := creer(titre, x, y, largeur, hauteur)

	const lb, hb, ecart = 104, 32, 16
	x0 := (lc - (2*lb + ecart)) / 2

	controle(fenetre, "STATIC", "Est-ce que ça t'embête ?", ssCenter, 20, 34, lc-40, 44, 0)
	controle(fenetre, "BUTTON", "Oui", bsDefPushButton|wsTabStop, x0, hc-hb-24, lb, hb, idOui)
	controle(fenetre, "BUTTON", "Non", bsDefPushButton|wsTabStop, x0+lb+ecart, hc-hb-24, lb, hb, idNon)

	afficher(fenetre)
}

// ouvrirFinale pose la derniere fenetre : plus de question, un seul bouton,
// mais qui se derobe.
func ouvrirFinale() {
	x, y := place(largeurF, hauteurF)
	finale, lcF, hcF = creer("D'accord, d'accord", x, y, largeurF, hauteurF)

	controle(finale, "STATIC", "Bon. Tu veux tout fermer ?", ssCenter, 20, 30, lcF-40, 30, 0)

	bx, by = (lcF-largeurB)/2, hcF-hauteurB-40
	bouton = controle(finale, "BUTTON", "Fermer tout",
		bsDefPushButton|wsTabStop, bx, by, largeurB, hauteurB, idFermer)

	if limiteEsquives != 0 {
		// Un minuteur plutot que WM_MOUSEMOVE : les deplacements de souris
		// au-dessus du bouton vont au bouton, pas a la fenetre, qui ne verrait
		// donc jamais la souris arriver sur lui.
		setTimer.Call(finale, minuteur, 30, 0)
	}
	afficher(finale)
}

// esquiver deplace le bouton si la souris s'en approche de trop pres.
func esquiver(fenetre uintptr) {
	if fenetre != finale || bouton == 0 {
		return
	}
	var p point
	getCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	screenToClient.Call(fenetre, uintptr(unsafe.Pointer(&p)))

	if p.x < bx-marge || p.x > bx+largeurB+marge || p.y < by-marge || p.y > by+hauteurB+marge {
		return
	}

	// On garde, parmi quelques positions tirees au hasard, celle qui eloigne le
	// plus le bouton du curseur — il fuit au lieu de sauter n'importe ou.
	meilleure := int32(-1)
	var mx, my int32
	for essai := 0; essai < 24; essai++ {
		cx := int32(rand.Intn(int(lcF-largeurB-20))) + 10
		cy := int32(rand.Intn(int(hcF-hauteurB-90))) + 80
		d := abs(cx+largeurB/2-p.x) + abs(cy+hauteurB/2-p.y)
		if d > meilleure {
			meilleure, mx, my = d, cx, cy
		}
		if d > ecartMin*2 {
			break
		}
	}
	bx, by = mx, my
	moveWindow.Call(bouton, uintptr(bx), uintptr(by), largeurB, hauteurB, 1)

	esquives++
	if limiteEsquives > 0 && esquives >= limiteEsquives {
		// Il s'avoue vaincu : sans cela, la seule issue serait le
		// gestionnaire des taches.
		killTimer.Call(fenetre, minuteur)
	}
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func main() {
	// Une fenetre Win32 doit vivre sur le fil qui l'a creee.
	runtime.LockOSThread()

	max := flag.Int("max", 10, "nombre de fenetres au total, la derniere fermant tout (0 : sans fin)")
	esq := flag.Int("esquives", 12, "nombre de fuites du bouton avant qu'il se laisse attraper (0 : il ne fuit pas)")
	flag.Parse()
	maximum, limiteEsquives = *max, *esq

	enregistrerClasse()
	ouvrir()

	// Une seule boucle pour toutes les fenetres.
	var m message
	for {
		r, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 { // 0 : WM_QUIT, -1 : erreur
			return
		}
		// Message poste au fil plutot qu'a une fenetre : il n'a pas de
		// destinataire a qui etre distribue, on le traite ici.
		if m.fenetre == 0 && m.code == wmOuvrir {
			ouvrir()
			continue
		}
		// Laisse le clavier atteindre le bouton fuyant : Tab pour l'atteindre,
		// Entree ou Espace pour l'actionner sans avoir a le rattraper.
		if finale != 0 {
			if r, _, _ := isDialogMessageW.Call(finale, uintptr(unsafe.Pointer(&m))); r != 0 {
				continue
			}
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
