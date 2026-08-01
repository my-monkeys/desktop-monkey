//go:build windows

// Package tray ajoute une icone dans la zone de notification de Windows, avec
// un petit menu : lancer au demarrage, ouvrir les reglages, quitter. Comme le
// reste du programme, tout passe par des appels syscall directs — pas de cgo.
//
// La fenetre cachee qui recoit les messages de l'icone vit sur le meme fil que
// la boucle principale ; celle-ci pompe deja tous les messages du fil (via
// PeekMessage), donc le menu fonctionne sans boucle dediee.
package tray

import (
	"image"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procCreatePopupMenu  = user32.NewProc("CreatePopupMenu")
	procAppendMenuW      = user32.NewProc("AppendMenuW")
	procTrackPopupMenu   = user32.NewProc("TrackPopupMenu")
	procDestroyMenu      = user32.NewProc("DestroyMenu")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procSetForegroundWin = user32.NewProc("SetForegroundWindow")
	procPostMessageW     = user32.NewProc("PostMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procLoadIconW        = user32.NewProc("LoadIconW")
	procCreateIconIndir  = user32.NewProc("CreateIconIndirect")
	procDestroyIcon      = user32.NewProc("DestroyIcon")

	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	procShellExecuteW    = shell32.NewProc("ShellExecuteW")

	procCreateDIBSection = gdi32.NewProc("CreateDIBSection")
	procCreateBitmap     = gdi32.NewProc("CreateBitmap")
	procDeleteObject     = gdi32.NewProc("DeleteObject")

	procRegCreateKeyExW = advapi32.NewProc("RegCreateKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW = advapi32.NewProc("RegDeleteValueW")
	procRegQueryValueEx = advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	wsOverlapped = 0x00000000

	nimAdd     = 0x0
	nimDelete  = 0x2
	nifMessage = 0x1
	nifIcon    = 0x2
	nifTip     = 0x4

	wmApp          = 0x8000
	wmTrayCallback = wmApp + 1
	wmRButtonUp    = 0x0205
	wmLButtonUp    = 0x0202
	wmContextMenu  = 0x007B
	wmCommand      = 0x0111
	wmNull         = 0x0000
	wmDestroy      = 0x0002

	mfString    = 0x0000
	mfSeparator = 0x0800
	mfChecked   = 0x0008

	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100

	idcArrow = 32512
	idiApp   = 32512

	hkeyCurrentUser = 0x80000000
	keyRead         = 0x20019
	keyWrite        = 0x20006
	regSZ           = 1
	errFileNotFound = 2

	idDemarrage = 1
	idReglages  = 2
	idQuitter   = 3
)

type point struct{ X, Y int32 }

type wndClassExW struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSm                             uintptr
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

type iconInfo struct {
	FIcon              int32
	XHotspot, YHotspot uint32
	HbmMask, HbmColor  uintptr
}

type bitmapInfoHeader struct {
	Size                         uint32
	Width, Height                int32
	Planes, BitCount             uint16
	Compression, SizeImage       uint32
	XPelsPerMeter, YPelsPerMeter int32
	ClrUsed, ClrImportant        uint32
}

// etat partage : il n'y a qu'un seul tray, la procedure de fenetre y accede.
var (
	gNID     notifyIconData
	gExe     string
	gConfig  string
	gAppName string
	gIcone   uintptr
)

// Nouvelle installe l'icone et son menu. exe est le chemin de l'executable (pour
// le lancement au demarrage), config le chemin du fichier de reglages a ouvrir,
// et icone une image 32x32 (l'icone du singe) — nil pour l'icone systeme.
func Nouvelle(nomApp, exe, config string, icone *image.RGBA) error {
	gAppName, gExe, gConfig = nomApp, exe, config

	instance, _, _ := procGetModuleHandleW.Call(0)
	classe, err := syscall.UTF16PtrFromString(nomApp + "TrayClasse")
	if err != nil {
		return err
	}
	wc := wndClassExW{
		Size:      uint32(unsafe.Sizeof(wndClassExW{})),
		WndProc:   syscall.NewCallback(procedure),
		Instance:  instance,
		ClassName: classe,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// fenetre cachee, seule receptrice des messages de l'icone
	hwnd, _, _ := procCreateWindowExW.Call(
		0, uintptr(unsafe.Pointer(classe)), uintptr(unsafe.Pointer(classe)),
		wsOverlapped, 0, 0, 0, 0, 0, 0, instance, 0)
	if hwnd == 0 {
		return syscall.EINVAL
	}

	gIcone = creerIcone(icone)
	if gIcone == 0 {
		gIcone, _, _ = procLoadIconW.Call(0, uintptr(idiApp))
	}

	gNID = notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: wmTrayCallback,
		HIcon:            gIcone,
	}
	copyUTF16(gNID.SzTip[:], nomApp)
	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&gNID)))
	return nil
}

// Fermer retire l'icone de la zone de notification.
func Fermer() {
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&gNID)))
	if gIcone != 0 {
		procDestroyIcon.Call(gIcone)
	}
	if gNID.HWnd != 0 {
		procDestroyWindow.Call(gNID.HWnd)
	}
}

func procedure(hwnd, msg, wparam, lparam uintptr) uintptr {
	switch msg {
	case wmTrayCallback:
		switch lparam & 0xFFFF {
		case wmRButtonUp, wmLButtonUp, wmContextMenu:
			afficherMenu(hwnd)
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, msg, wparam, lparam)
	return r
}

func afficherMenu(hwnd uintptr) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	drapeauDem := uintptr(mfString)
	if auDemarrage() {
		drapeauDem |= mfChecked
	}
	appendItem(menu, drapeauDem, idDemarrage, "Lancer au demarrage")
	appendItem(menu, mfString, idReglages, "Ouvrir les reglages...")
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	appendItem(menu, mfString, idQuitter, "Quitter")

	// requis pour que le menu se referme correctement en cliquant ailleurs
	procSetForegroundWin.Call(hwnd)

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	cmd, _, _ := procTrackPopupMenu.Call(menu,
		tpmRightButton|tpmReturnCmd, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	procPostMessageW.Call(hwnd, wmNull, 0, 0)

	switch cmd {
	case idDemarrage:
		basculerDemarrage()
	case idReglages:
		ouvrirReglages()
	case idQuitter:
		Fermer()
		procPostQuitMessage.Call(0)
	}
}

func appendItem(menu, flags, id uintptr, texte string) {
	p, _ := syscall.UTF16PtrFromString(texte)
	procAppendMenuW.Call(menu, flags, id, uintptr(unsafe.Pointer(p)))
}

func ouvrirReglages() {
	verbe, _ := syscall.UTF16PtrFromString("open")
	fichier, _ := syscall.UTF16PtrFromString(gConfig)
	procShellExecuteW.Call(0,
		uintptr(unsafe.Pointer(verbe)), uintptr(unsafe.Pointer(fichier)), 0, 0, 1)
}

// --- lancement au demarrage : cle HKCU ...\Run --------------------------------

var cheminRun = `Software\Microsoft\Windows\CurrentVersion\Run`

func ouvrirRun(acces uintptr) (uintptr, bool) {
	p, _ := syscall.UTF16PtrFromString(cheminRun)
	var h uintptr
	r, _, _ := procRegCreateKeyExW.Call(hkeyCurrentUser,
		uintptr(unsafe.Pointer(p)), 0, 0, 0, acces, 0,
		uintptr(unsafe.Pointer(&h)), 0)
	return h, r == 0
}

func auDemarrage() bool {
	h, ok := ouvrirRun(keyRead)
	if !ok {
		return false
	}
	defer procRegCloseKey.Call(h)
	nom, _ := syscall.UTF16PtrFromString(gAppName)
	r, _, _ := procRegQueryValueEx.Call(h,
		uintptr(unsafe.Pointer(nom)), 0, 0, 0, 0)
	return r == 0
}

func basculerDemarrage() {
	if auDemarrage() {
		h, ok := ouvrirRun(keyWrite)
		if !ok {
			return
		}
		defer procRegCloseKey.Call(h)
		nom, _ := syscall.UTF16PtrFromString(gAppName)
		procRegDeleteValueW.Call(h, uintptr(unsafe.Pointer(nom)))
		return
	}
	h, ok := ouvrirRun(keyWrite)
	if !ok {
		return
	}
	defer procRegCloseKey.Call(h)
	nom, _ := syscall.UTF16PtrFromString(gAppName)
	val := `"` + gExe + `"`
	v, _ := syscall.UTF16FromString(val)
	procRegSetValueExW.Call(h, uintptr(unsafe.Pointer(nom)), 0, regSZ,
		uintptr(unsafe.Pointer(&v[0])), uintptr(len(v)*2))
}

// --- icone construite depuis le sprite ---------------------------------------

func creerIcone(img *image.RGBA) uintptr {
	if img == nil {
		return 0
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	entete := bitmapInfoHeader{
		Size:     uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:    int32(w),
		Height:   -int32(h), // origine en haut, comme les images Go
		Planes:   1,
		BitCount: 32,
	}
	var bits unsafe.Pointer
	hbmColor, _, _ := procCreateDIBSection.Call(0,
		uintptr(unsafe.Pointer(&entete)), 0,
		uintptr(unsafe.Pointer(&bits)), 0, 0)
	if hbmColor == 0 {
		return 0
	}
	px := unsafe.Slice((*byte)(bits), w*h*4)
	for i := 0; i < w*h*4; i += 4 {
		px[i+0] = img.Pix[i+2] // B
		px[i+1] = img.Pix[i+1] // V
		px[i+2] = img.Pix[i+0] // R
		px[i+3] = img.Pix[i+3] // A
	}

	hbmMask, _, _ := procCreateBitmap.Call(uintptr(w), uintptr(h), 1, 1, 0)

	info := iconInfo{FIcon: 1, HbmMask: hbmMask, HbmColor: hbmColor}
	hIcon, _, _ := procCreateIconIndir.Call(uintptr(unsafe.Pointer(&info)))

	procDeleteObject.Call(hbmColor)
	procDeleteObject.Call(hbmMask)
	return hIcon
}

func copyUTF16(dst []uint16, s string) {
	src, _ := syscall.UTF16FromString(s)
	n := copy(dst, src)
	if n < len(dst) {
		dst[n] = 0
	}
}
