//go:build windows

// La fenetre native de reglages, version Windows : une fenetre a onglets
// (SysTabControl32) avec curseurs (msctls_trackbar32), listes et cases, batie
// en appels syscall directs comme le reste du programme — pas de cgo. Elle est
// construite depuis la meme description JSON que la version macOS.
package dialogue

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"log"
	"syscall"
	"unsafe"
)

var (
	duser32   = syscall.NewLazyDLL("user32.dll")
	dgdi32    = syscall.NewLazyDLL("gdi32.dll")
	dcomctl32 = syscall.NewLazyDLL("comctl32.dll")
	dkernel32 = syscall.NewLazyDLL("kernel32.dll")

	pRegisterClassExW   = duser32.NewProc("RegisterClassExW")
	pCreateWindowExW    = duser32.NewProc("CreateWindowExW")
	pDefWindowProcW     = duser32.NewProc("DefWindowProcW")
	pDestroyWindow      = duser32.NewProc("DestroyWindow")
	pShowWindow         = duser32.NewProc("ShowWindow")
	pSetWindowTextW     = duser32.NewProc("SetWindowTextW")
	pSendMessageW       = duser32.NewProc("SendMessageW")
	pSetForegroundWin   = duser32.NewProc("SetForegroundWindow")
	pGetSystemMetrics   = duser32.NewProc("GetSystemMetrics")
	pGetSysColor        = duser32.NewProc("GetSysColor")
	pGetSysColorBrush   = duser32.NewProc("GetSysColorBrush")
	pSetBkMode          = dgdi32.NewProc("SetBkMode")
	pSetTextColor       = dgdi32.NewProc("SetTextColor")
	pGetStockObject     = dgdi32.NewProc("GetStockObject")
	pCreateDIBSection   = dgdi32.NewProc("CreateDIBSection")
	pDeleteObject       = dgdi32.NewProc("DeleteObject")
	pInitCommonControls = dcomctl32.NewProc("InitCommonControlsEx")
	pGetModuleHandleW   = dkernel32.NewProc("GetModuleHandleW")
)

const (
	dwsVisible      = 0x10000000
	dwsChild        = 0x40000000
	dwsCaption      = 0x00C00000
	dwsSysMenu      = 0x00080000
	dwsClipSiblings = 0x04000000
	dwsTabStop      = 0x00010000

	dssRight  = 0x0002
	dssBitmap = 0x000E

	dbsDefPush   = 0x0001
	dbsAutoCheck = 0x0003

	dcbsDropList = 0x0003
	dcbAddString = 0x0143
	dcbSetCurSel = 0x014E
	dcbGetCurSel = 0x0147

	dbmSetCheck = 0x00F1
	dbmGetCheck = 0x00F0

	dtbmGetPos   = 0x0400
	dtbmSetPos   = 0x0405
	dtbmSetRange = 0x0406

	dtcmInsertItem = 0x133E
	dtcmGetCurSel  = 0x130B
	dtcnSelChange  = 0xFFFFFDD9 // TCN_FIRST(-550) - 1, en uint32

	dwmCommand        = 0x0111
	dwmHScroll        = 0x0114
	dwmNotify         = 0x004E
	dwmClose          = 0x0010
	dwmDestroy        = 0x0002
	dwmSetFont        = 0x0030
	dwmCtlColorStatic = 0x0138

	dstmSetImage = 0x0172

	dSwHide = 0
	dSwShow = 5

	couleurFace  = 15 // COLOR_BTNFACE
	couleurGrise = 17 // COLOR_GRAYTEXT

	idEnregistrer = 1000
	idAnnuler     = 1001
	idTab         = 1002
)

type dwndClassExW struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSm                             uintptr
}

type dtcItemW struct {
	Mask, State, StateMask uint32
	_                      uint32
	PszText                *uint16
	CchTextMax, IImage     int32
	LParam                 uintptr
}

type dnmHdr struct {
	HwndFrom, IDFrom uintptr
	Code             uint32
}

type dbitmapInfoHeader struct {
	Size                         uint32
	Width, Height                int32
	Planes, BitCount             uint16
	Compression, SizeImage       uint32
	XPelsPerMeter, YPelsPerMeter int32
	ClrUsed, ClrImportant        uint32
}

type dinitCommonControlsEx struct {
	Size, ICC uint32
}

// description JSON (miroir de la version macOS)
type dChamp struct {
	Type, Cle, Nom, Aide string
	Min, Max, Pas        float64
	Valeur               float64
	Options, Libelles    []string
	Texte                string
	Coche                bool
	Liee, Png            string
	BaseL, BaseH         int
}

type dOnglet struct {
	Titre  string
	Champs []dChamp
}

type dDesc struct {
	Titre, Enregistrer, Annuler string
	Onglets                     []dOnglet
}

// un controle construit, pour la collecte et les mises a jour
type dCtl struct {
	champ dChamp
	h     uintptr // trackbar, combo ou case
	hVal  uintptr // label de la valeur d'un curseur
}

var (
	dClasseOK  bool
	dFen       uintptr
	dResultat  string
	dPret      bool
	dPolice    uintptr
	dCtls      map[string]*dCtl
	dAides     map[uintptr]bool // labels a peindre en gris
	dGroupes   [][]uintptr      // handles par onglet, pour montrer/cacher
	dTab       uintptr
	dApercu    uintptr // STATIC SS_BITMAP de l'apercu
	dApercuBmp uintptr // HBITMAP courant (a liberer)
	dApercuImg *image.RGBA
	dApercuLie string
	dApercuT   float64 // derniere taille dessinee
)

// Disponible indique que la plateforme offre le dialogue natif.
func Disponible() bool { return true }

// Resultat renvoie le JSON des valeurs enregistrees, une seule fois.
func Resultat() (string, bool) {
	if !dPret {
		return "", false
	}
	dPret = false
	return dResultat, true
}

// Ouvrir construit et montre la fenetre ; si elle existe deja, la ramene
// simplement au premier plan.
func Ouvrir(descJSON string) {
	if dFen != 0 {
		pSetForegroundWin.Call(dFen)
		return
	}
	var d dDesc
	if err := json.Unmarshal([]byte(descJSON), &d); err != nil {
		log.Printf("description du dialogue : %v", err)
		return
	}

	icc := dinitCommonControlsEx{Size: 8, ICC: 0x4 | 0x8} // barres + onglets
	pInitCommonControls.Call(uintptr(unsafe.Pointer(&icc)))
	dPolice, _, _ = pGetStockObject.Call(17) // DEFAULT_GUI_FONT

	instance, _, _ := pGetModuleHandleW.Call(0)
	classe, _ := syscall.UTF16PtrFromString("SingeReglagesClasse")
	if !dClasseOK {
		wc := dwndClassExW{
			Size:       uint32(unsafe.Sizeof(dwndClassExW{})),
			WndProc:    syscall.NewCallback(dProcedure),
			Instance:   instance,
			Background: uintptr(couleurFace + 1),
			ClassName:  classe,
		}
		pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		dClasseOK = true
	}

	// hauteur du plus grand onglet
	contenuH := 0
	for _, o := range d.Onglets {
		h := 12
		for _, ch := range o.Champs {
			h += dHauteurChamp(ch) + 8
		}
		if h > contenuH {
			contenuH = h
		}
	}
	// le bandeau d'onglets seul en haut (pas de "page" : les champs vivent sur
	// le fond de la fenetre, ce qui evite toute guerre d'ordre Z avec le tab)
	largFen := 500
	const hautChamps = 52
	basChamps := hautChamps + contenuH
	fenH := basChamps + 96 // boutons + barre de titre

	titre, _ := syscall.UTF16PtrFromString(d.Titre)
	ecranL, _, _ := pGetSystemMetrics.Call(0)
	ecranH, _, _ := pGetSystemMetrics.Call(1)
	dFen, _, _ = pCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(classe)), uintptr(unsafe.Pointer(titre)),
		dwsCaption|dwsSysMenu|dwsClipSiblings,
		(ecranL-uintptr(largFen))/2, (ecranH-uintptr(fenH))/2,
		uintptr(largFen), uintptr(fenH), 0, 0, instance, 0)
	if dFen == 0 {
		return
	}

	dCtls = map[string]*dCtl{}
	dAides = map[uintptr]bool{}
	dGroupes = make([][]uintptr, len(d.Onglets))

	dTab = dEnfant("SysTabControl32", "", dwsChild|dwsVisible|dwsClipSiblings|dwsTabStop,
		8, 8, largFen-32, 30, idTab)
	for i, o := range d.Onglets {
		txt, _ := syscall.UTF16PtrFromString(o.Titre)
		it := dtcItemW{Mask: 0x1, PszText: txt} // TCIF_TEXT
		pSendMessageW.Call(dTab, dtcmInsertItem, uintptr(i), uintptr(unsafe.Pointer(&it)))
	}

	// les champs de chaque onglet, sous le bandeau
	for i, o := range d.Onglets {
		y := hautChamps
		for _, ch := range o.Champs {
			hs := dConstruireChamp(ch, 22, y, largFen-60)
			dGroupes[i] = append(dGroupes[i], hs...)
			y += dHauteurChamp(ch) + 8
		}
	}
	dMontrerOnglet(0)

	// le champ ferme d'une liste choisie alors qu'elle etait cachee reste
	// parfois vide : on re-affirme la selection une fois les controles montres
	for _, c := range dCtls {
		if c.champ.Type == "choix" {
			i, _, _ := pSendMessageW.Call(c.h, dcbGetCurSel, 0, 0)
			pSendMessageW.Call(c.h, dcbSetCurSel, i, 0)
		}
	}

	// boutons
	dEnfant("BUTTON", d.Enregistrer, dwsChild|dwsVisible|dbsDefPush,
		largFen-206, basChamps+8, 178, 30, idEnregistrer)
	dEnfant("BUTTON", d.Annuler, dwsChild|dwsVisible,
		largFen-300, basChamps+8, 86, 30, idAnnuler)

	pShowWindow.Call(dFen, dSwShow)
	pSetForegroundWin.Call(dFen)
}

// dEnfant cree un controle enfant avec la police standard.
func dEnfant(classe, texte string, style uintptr, x, y, l, h, id int) uintptr {
	c, _ := syscall.UTF16PtrFromString(classe)
	t, _ := syscall.UTF16PtrFromString(texte)
	instance, _, _ := pGetModuleHandleW.Call(0)
	hw, _, _ := pCreateWindowExW.Call(0,
		uintptr(unsafe.Pointer(c)), uintptr(unsafe.Pointer(t)), style,
		uintptr(x), uintptr(y), uintptr(l), uintptr(h),
		dFen, uintptr(id), instance, 0)
	pSendMessageW.Call(hw, dwmSetFont, dPolice, 1)
	return hw
}

func dHauteurChamp(ch dChamp) int {
	h := 30 // une ligne : libelle + controle
	if ch.Type == "curseur" || ch.Type == "entier" {
		h = 52
	}
	if ch.Type == "apercu" {
		return 20 + ch.BaseH*2
	}
	if ch.Aide != "" {
		h += 17
	}
	return h
}

// dConstruireChamp cree les controles d'un champ (caches : l'onglet actif les
// montrera) et renvoie leurs handles.
func dConstruireChamp(ch dChamp, x, y, larg int) []uintptr {
	var hs []uintptr
	st := uintptr(dwsChild | dwsClipSiblings)

	nom := dEnfant("STATIC", ch.Nom, st, x, y, larg-90, 17, 0)
	hs = append(hs, nom)

	switch ch.Type {
	case "curseur", "entier":
		pas := ch.Pas
		if pas <= 0 {
			pas = 1
		}
		lv := dEnfant("STATIC", dFormatValeur(ch.Valeur, pas), st|dssRight,
			x+larg-86, y, 86, 17, 0)
		tb := dEnfant("msctls_trackbar32", "", st|dwsTabStop, x, y+20, larg, 26, 0)
		max := int((ch.Max - ch.Min) / pas)
		pSendMessageW.Call(tb, dtbmSetRange, 1, uintptr(max)<<16)
		pSendMessageW.Call(tb, dtbmSetPos, 1, uintptr(int((ch.Valeur-ch.Min)/pas+0.5)))
		c := &dCtl{champ: ch, h: tb, hVal: lv}
		dCtls[ch.Cle] = c
		hs = append(hs, lv, tb)
		if ch.Aide != "" {
			a := dEnfant("STATIC", ch.Aide, st, x, y+48, larg, 15, 0)
			dAides[a] = true
			hs = append(hs, a)
		}
	case "choix":
		cb := dEnfant("COMBOBOX", "", st|dcbsDropList|dwsTabStop, x+larg-180, y-3, 180, 200, 0)
		sel := 0
		for i, lib := range ch.Libelles {
			t, _ := syscall.UTF16PtrFromString(lib)
			pSendMessageW.Call(cb, dcbAddString, 0, uintptr(unsafe.Pointer(t)))
			if i < len(ch.Options) && ch.Options[i] == ch.Texte {
				sel = i
			}
		}
		pSendMessageW.Call(cb, dcbSetCurSel, uintptr(sel), 0)
		dCtls[ch.Cle] = &dCtl{champ: ch, h: cb}
		hs = append(hs, cb)
		if ch.Aide != "" {
			a := dEnfant("STATIC", ch.Aide, st, x, y+26, larg, 15, 0)
			dAides[a] = true
			hs = append(hs, a)
		}
	case "case":
		bt := dEnfant("BUTTON", "", st|dbsAutoCheck|dwsTabStop, x+larg-24, y, 20, 20, 0)
		if ch.Coche {
			pSendMessageW.Call(bt, dbmSetCheck, 1, 0)
		}
		dCtls[ch.Cle] = &dCtl{champ: ch, h: bt}
		hs = append(hs, bt)
		if ch.Aide != "" {
			a := dEnfant("STATIC", ch.Aide, st, x, y+26, larg, 15, 0)
			dAides[a] = true
			hs = append(hs, a)
		}
	case "apercu":
		dAides[nom] = true
		if img := dDecoderPng(ch.Png); img != nil {
			dApercuImg = img
			dApercuLie = ch.Liee
			dApercu = dEnfant("STATIC", "", st|dssBitmap,
				x+(larg-ch.BaseL*2)/2, y+18, ch.BaseL*2, ch.BaseH*2, 0)
			hs = append(hs, dApercu)
			t := 1.0
			if c, ok := dCtls[ch.Liee]; ok {
				t = c.valeur()
			}
			dPeindreApercu(t)
		}
	}
	for _, h := range hs {
		pShowWindow.Call(h, dSwHide)
	}
	return hs
}

func dDecoderPng(b64 string) *image.RGBA {
	brut, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(brut))
	if err != nil {
		return nil
	}
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			out.Set(x, y, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return out
}

// valeur lit la position d'un curseur en unites reelles.
func (c *dCtl) valeur() float64 {
	pas := c.champ.Pas
	if pas <= 0 {
		pas = 1
	}
	pos, _, _ := pSendMessageW.Call(c.h, dtbmGetPos, 0, 0)
	return c.champ.Min + float64(pos)*pas
}

func dFormatValeur(v, pas float64) string {
	if pas >= 1 {
		return fmtF(v, 0)
	}
	return fmtF(v, 2)
}

func fmtF(v float64, dec int) string {
	b, _ := json.Marshal(v)
	s := string(b)
	if dec == 0 {
		if i := indexOf(s, '.'); i >= 0 {
			return s[:i]
		}
		return s
	}
	i := indexOf(s, '.')
	if i < 0 {
		return s + ".00"
	}
	for len(s) < i+1+dec {
		s += "0"
	}
	return s[:i+1+dec]
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// dPeindreApercu regenere le bitmap de l'apercu : le singe a l'echelle t,
// pose en bas au centre d'un canevas fixe (sa taille maximale).
func dPeindreApercu(t float64) {
	if dApercu == 0 || dApercuImg == nil {
		return
	}
	dApercuT = t
	base := dApercuImg.Bounds()
	cl, chh := base.Dx()*2, base.Dy()*2 // canevas = taille maximale (x2)
	sl, sh := int(float64(base.Dx())*t), int(float64(base.Dy())*t)

	entete := dbitmapInfoHeader{
		Size:     uint32(unsafe.Sizeof(dbitmapInfoHeader{})),
		Width:    int32(cl),
		Height:   -int32(chh),
		Planes:   1,
		BitCount: 32,
	}
	var bits unsafe.Pointer
	bmp, _, _ := pCreateDIBSection.Call(0,
		uintptr(unsafe.Pointer(&entete)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 {
		return
	}
	px := unsafe.Slice((*byte)(bits), cl*chh*4)

	// fond : la couleur de la boite de dialogue
	face, _, _ := pGetSysColor.Call(couleurFace)
	fr, fg, fb := byte(face), byte(face>>8), byte(face>>16)
	for i := 0; i < len(px); i += 4 {
		px[i+0], px[i+1], px[i+2], px[i+3] = fb, fg, fr, 0xFF
	}

	// le singe, plus proche voisin, ancre bas-centre
	x0, y0 := (cl-sl)/2, chh-sh
	for y := 0; y < sh; y++ {
		sy := y * base.Dy() / sh
		for x := 0; x < sl; x++ {
			sx := x * base.Dx() / sl
			o := dApercuImg.PixOffset(sx, sy)
			if dApercuImg.Pix[o+3] < 128 {
				continue
			}
			i := ((y0+y)*cl + x0 + x) * 4
			px[i+0] = dApercuImg.Pix[o+2]
			px[i+1] = dApercuImg.Pix[o+1]
			px[i+2] = dApercuImg.Pix[o+0]
		}
	}

	ancien, _, _ := pSendMessageW.Call(dApercu, dstmSetImage, 0, bmp)
	if ancien != 0 {
		pDeleteObject.Call(ancien)
	}
	if dApercuBmp != 0 && dApercuBmp != ancien {
		pDeleteObject.Call(dApercuBmp)
	}
	dApercuBmp = bmp
}

func dMontrerOnglet(n int) {
	for i, hs := range dGroupes {
		mode := uintptr(dSwHide)
		if i == n {
			mode = dSwShow
		}
		for _, h := range hs {
			pShowWindow.Call(h, mode)
		}
	}
}

func dProcedure(hwnd, msg, wparam, lparam uintptr) uintptr {
	switch msg {
	case dwmHScroll:
		// un curseur a bouge : maj de sa valeur affichee (et de l'apercu)
		for cle, c := range dCtls {
			if c.h == lparam && c.hVal != 0 {
				v := c.valeur()
				pas := c.champ.Pas
				if pas <= 0 {
					pas = 1
				}
				t, _ := syscall.UTF16PtrFromString(dFormatValeur(v, pas))
				pSetWindowTextW.Call(c.hVal, uintptr(unsafe.Pointer(t)))
				if cle == dApercuLie && v != dApercuT {
					dPeindreApercu(v)
				}
			}
		}
		return 0

	case dwmNotify:
		h := (*dnmHdr)(unsafe.Pointer(lparam))
		if h.HwndFrom == dTab && h.Code == dtcnSelChange {
			sel, _, _ := pSendMessageW.Call(dTab, dtcmGetCurSel, 0, 0)
			dMontrerOnglet(int(sel))
		}
		return 0

	case dwmCtlColorStatic:
		// aides en gris, fond de boite pour tous les libelles
		if dAides[lparam] {
			pSetTextColor.Call(wparam, dCouleur(couleurGrise))
		}
		pSetBkMode.Call(wparam, 1) // TRANSPARENT
		br, _, _ := pGetSysColorBrush.Call(couleurFace)
		return br

	case dwmCommand:
		switch wparam & 0xFFFF {
		case idEnregistrer:
			dCollecter()
			pDestroyWindow.Call(hwnd)
			return 0
		case idAnnuler:
			pDestroyWindow.Call(hwnd)
			return 0
		}

	case dwmClose:
		pDestroyWindow.Call(hwnd)
		return 0

	case dwmDestroy:
		if dApercuBmp != 0 {
			pDeleteObject.Call(dApercuBmp)
			dApercuBmp = 0
		}
		dFen, dTab, dApercu = 0, 0, 0
		dCtls, dAides, dGroupes = nil, nil, nil
		dApercuImg, dApercuLie = nil, ""
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, msg, wparam, lparam)
	return r
}

func dCouleur(idx uintptr) uintptr {
	c, _, _ := pGetSysColor.Call(idx)
	return c
}

// dCollecter lit tous les controles et serialise le resultat.
func dCollecter() {
	vals := map[string]any{}
	for cle, c := range dCtls {
		switch c.champ.Type {
		case "curseur":
			vals[cle] = c.valeur()
		case "entier":
			vals[cle] = int(c.valeur() + 0.5)
		case "case":
			r, _, _ := pSendMessageW.Call(c.h, dbmGetCheck, 0, 0)
			vals[cle] = r == 1
		case "choix":
			i, _, _ := pSendMessageW.Call(c.h, dcbGetCurSel, 0, 0)
			if int(i) >= 0 && int(i) < len(c.champ.Options) {
				vals[cle] = c.champ.Options[i]
			}
		}
	}
	brut, err := json.Marshal(vals)
	if err != nil {
		return
	}
	dResultat, dPret = string(brut), true
}
