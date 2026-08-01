//go:build windows

// Package calque affiche une image a canal alpha directement sur le bureau.
//
// Pourquoi pas un moteur de jeu : sur Windows, aucune combinaison n'est fiable.
// Ebitengine refuse DirectX des qu'on demande une fenetre transparente et
// bascule sur OpenGL, absent de beaucoup de machines (l'application ne demarre
// alors pas du tout) ou incapable de composer la transparence (fond noir
// opaque). Et l'astuce inverse — DirectX plus fenetre a couleur-cle — ne marche
// pas non plus : la presentation "flip" de DXGI est incompatible avec les
// fenetres en couche.
//
// On utilise donc le mecanisme prevu par Windows pour exactement cet usage :
// une fenetre en couche mise a jour par UpdateLayeredWindow, qui accepte une
// image 32 bits avec alpha par pixel. Le rendu est logiciel, ne depend d'aucun
// pilote 3D, et donne une vraie transparence progressive — les bords lisses des
// bulles sont donc parfaitement rendus.
package calque

import (
	"fmt"
	"image"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procPeekMessageW     = user32.NewProc("PeekMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procGetDC            = user32.NewProc("GetDC")
	procReleaseDC        = user32.NewProc("ReleaseDC")
	procUpdateLayered    = user32.NewProc("UpdateLayeredWindow")

	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procDeleteDC           = gdi32.NewProc("DeleteDC")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	wsPopup = 0x80000000

	wsExLayered    = 0x00080000 // mise a jour par UpdateLayeredWindow
	wsExToolWindow = 0x00000080 // absente de la barre des taches et d'Alt+Tab
	wsExTopMost    = 0x00000008 // au-dessus des autres fenetres
	wsExNoActivate = 0x08000000 // ne prend jamais le focus

	swShowNoActivate = 4
	pmRemove         = 0x0001
	wmQuit           = 0x0012

	errClasseExiste = 1410 // ERROR_CLASS_ALREADY_EXISTS

	biRGB        = 0
	dibRGBColors = 0
	acSrcOver    = 0x00
	acSrcAlpha   = 0x01
	ulwAlpha     = 0x00000002
	cwUseDefault = ^uintptr(0x7FFFFFFF) // 0x80000000
)

type pointL struct{ X, Y int32 }
type sizeL struct{ Cx, Cy int32 }

type blendFunction struct {
	BlendOp, BlendFlags, SourceConstantAlpha, AlphaFormat byte
}

type bitmapInfoHeader struct {
	Size                         uint32
	Width, Height                int32
	Planes, BitCount             uint16
	Compression, SizeImage       uint32
	XPelsPerMeter, YPelsPerMeter int32
	ClrUsed, ClrImportant        uint32
}

type wndClassExW struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSm                             uintptr
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      pointL
	Private uint32
}

// Calque est une fenetre sans bordure, toujours visible et traversante, dont le
// contenu est une image a canal alpha.
type Calque struct {
	hwnd       uintptr
	memDC      uintptr
	bitmap     uintptr
	ancien     uintptr
	pixels     []byte // pointe directement dans la DIB : ecriture sans copie
	larg, haut int
}

// Ouvrir cree la fenetre. Sa taille est fixee ici ; le contenu affiche peut
// occuper une zone plus petite.
//
// A appeler depuis un fil verrouille (runtime.LockOSThread) : sous Windows, une
// fenetre appartient au fil qui l'a creee et seul celui-ci peut traiter ses
// messages.
func Ouvrir(nom string, larg, haut int) (*Calque, error) {
	runtime.LockOSThread()

	instance, _, _ := procGetModuleHandleW.Call(0)
	classe, err := syscall.UTF16PtrFromString(nom + "Classe")
	if err != nil {
		return nil, err
	}
	titre, err := syscall.UTF16PtrFromString(nom)
	if err != nil {
		return nil, err
	}

	wc := wndClassExW{
		Size:      uint32(unsafe.Sizeof(wndClassExW{})),
		WndProc:   syscall.NewCallback(procedureFenetre),
		Instance:  instance,
		ClassName: classe,
	}
	// Plusieurs fenetres peuvent partager la meme classe (une par crotte, par
	// exemple) : reenregistrer une classe deja connue echoue avec
	// ERROR_CLASS_ALREADY_EXISTS, ce qui est ici sans consequence.
	if ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		if errno, ok := err.(syscall.Errno); !ok || errno != errClasseExiste {
			return nil, fmt.Errorf("RegisterClassExW : %w", err)
		}
	}

	// WS_EX_TRANSPARENT est volontairement absent. Sans lui, Windows teste les
	// clics au pixel pres sur une fenetre en couche : ils traversent la ou
	// l'image est transparente, et sont absorbes par le singe la ou il est
	// dessine. Avec ce style, tout traverserait — y compris les clics destines
	// au singe, qui atteindraient aussi le bureau en dessous.
	hwnd, _, err := procCreateWindowExW.Call(
		wsExLayered|wsExToolWindow|wsExTopMost|wsExNoActivate,
		uintptr(unsafe.Pointer(classe)),
		uintptr(unsafe.Pointer(titre)),
		wsPopup,
		0, 0, uintptr(larg), uintptr(haut),
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return nil, fmt.Errorf("CreateWindowExW : %w", err)
	}

	c := &Calque{hwnd: hwnd, larg: larg, haut: haut}
	if err := c.creerSurface(larg, haut); err != nil {
		procDestroyWindow.Call(hwnd)
		return nil, err
	}

	procShowWindow.Call(hwnd, swShowNoActivate)
	return c, nil
}

// creerSurface alloue la zone memoire partagee avec GDI dans laquelle on peint.
func (c *Calque) creerSurface(larg, haut int) error {
	dc, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, dc)

	memDC, _, err := procCreateCompatibleDC.Call(dc)
	if memDC == 0 {
		return fmt.Errorf("CreateCompatibleDC : %w", err)
	}

	entete := bitmapInfoHeader{
		Size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:       int32(larg),
		Height:      -int32(haut), // negatif : origine en haut, comme les images Go
		Planes:      1,
		BitCount:    32,
		Compression: biRGB,
	}

	var bits unsafe.Pointer
	bitmap, _, err := procCreateDIBSection.Call(
		memDC, uintptr(unsafe.Pointer(&entete)), dibRGBColors,
		uintptr(unsafe.Pointer(&bits)), 0, 0,
	)
	if bitmap == 0 {
		procDeleteDC.Call(memDC)
		return fmt.Errorf("CreateDIBSection : %w", err)
	}

	ancien, _, _ := procSelectObject.Call(memDC, bitmap)

	c.memDC, c.bitmap, c.ancien = memDC, bitmap, ancien
	c.pixels = unsafe.Slice((*byte)(bits), larg*haut*4)
	return nil
}

// Afficher place la fenetre en (x, y) et y peint l'image.
//
// L'image doit etre a alpha premultiplie, ce qui est justement la convention de
// image.RGBA : elle peut donc etre transferee sans conversion de couleurs, au
// seul echange des canaux rouge et bleu (GDI attend l'ordre BGRA).
func (c *Calque) Afficher(img *image.RGBA, x, y int) error {
	b := img.Bounds()
	larg, haut := b.Dx(), b.Dy()
	if larg > c.larg || haut > c.haut {
		return fmt.Errorf("image %dx%d plus grande que le calque %dx%d", larg, haut, c.larg, c.haut)
	}

	for ligne := 0; ligne < haut; ligne++ {
		src := img.Pix[ligne*img.Stride : ligne*img.Stride+larg*4]
		dst := c.pixels[ligne*c.larg*4 : ligne*c.larg*4+larg*4]
		for i := 0; i < larg*4; i += 4 {
			dst[i+0] = src[i+2] // B
			dst[i+1] = src[i+1] // V
			dst[i+2] = src[i+0] // R
			dst[i+3] = src[i+3] // A
		}
	}
	// efface la portion de la surface non couverte par l'image
	for ligne := 0; ligne < haut; ligne++ {
		reste := c.pixels[ligne*c.larg*4+larg*4 : (ligne+1)*c.larg*4]
		for i := range reste {
			reste[i] = 0
		}
	}
	for ligne := haut; ligne < c.haut; ligne++ {
		reste := c.pixels[ligne*c.larg*4 : (ligne+1)*c.larg*4]
		for i := range reste {
			reste[i] = 0
		}
	}

	dcEcran, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, dcEcran)

	position := pointL{int32(x), int32(y)}
	origine := pointL{0, 0}
	taille := sizeL{int32(c.larg), int32(c.haut)}
	melange := blendFunction{acSrcOver, 0, 255, acSrcAlpha}

	ret, _, err := procUpdateLayered.Call(
		c.hwnd, dcEcran,
		uintptr(unsafe.Pointer(&position)),
		uintptr(unsafe.Pointer(&taille)),
		c.memDC,
		uintptr(unsafe.Pointer(&origine)),
		0,
		uintptr(unsafe.Pointer(&melange)),
		ulwAlpha,
	)
	if ret == 0 {
		return fmt.Errorf("UpdateLayeredWindow : %w", err)
	}
	return nil
}

// Traversant n'a rien a faire sous Windows : une fenetre en couche teste deja
// les clics au pixel pres (ils traversent la ou l'image est transparente).
func (c *Calque) Traversant(oui bool) {}

// TraiterMessages vide la file de messages de la fenetre. A appeler a chaque
// tour de boucle : sans cela Windows considere l'application comme figee.
// Renvoie false quand la fenetre doit se fermer.
func (c *Calque) TraiterMessages() bool {
	var m msg
	for {
		ret, _, _ := procPeekMessageW.Call(
			uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
		if ret == 0 {
			return true
		}
		if m.Message == wmQuit {
			return false
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// Fermer libere la fenetre et ses ressources graphiques.
func (c *Calque) Fermer() {
	if c.memDC != 0 {
		procSelectObject.Call(c.memDC, c.ancien)
		procDeleteObject.Call(c.bitmap)
		procDeleteDC.Call(c.memDC)
	}
	if c.hwnd != 0 {
		procDestroyWindow.Call(c.hwnd)
	}
}

func procedureFenetre(hwnd uintptr, message uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wparam, lparam)
	return ret
}
