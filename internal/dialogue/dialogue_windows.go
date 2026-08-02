//go:build windows

// Sous Windows, la fenetre de reglages est une fenetre native portant une
// WebView2 (le rendu web d'Edge, present sur les Windows modernes), chargee
// sur la page locale — la meme interface que sous macOS. Elle vit dans sa
// propre goroutine a fil fixe : chaque fil Windows peut porter ses fenetres et
// sa boucle de messages, le singe continue de vivre pendant les reglages.
package dialogue

import (
	"os/exec"
	"runtime"
	"sync"

	webview2 "github.com/jchv/go-webview2"
)

var (
	dvMu sync.Mutex
	dvW  webview2.WebView
)

// Disponible indique que la plateforme sait afficher la fenetre.
func Disponible() bool { return true }

// Ouvrir montre la fenetre sur l'URL donnee. Si le runtime WebView2 manque
// (vieux Windows sans Edge), la page s'ouvre dans le navigateur.
func Ouvrir(url, titre string, larg, haut int) {
	dvMu.Lock()
	deja := dvW != nil
	dvMu.Unlock()
	if deja {
		return // deja ouverte
	}

	go func() {
		runtime.LockOSThread()
		w := webview2.NewWithOptions(webview2.WebViewOptions{
			AutoFocus: true,
			WindowOptions: webview2.WindowOptions{
				Title:  titre,
				Width:  uint(larg),
				Height: uint(haut),
				Center: true,
			},
		})
		if w == nil {
			// pas de runtime WebView2 : le navigateur fera l'affaire
			_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			return
		}
		dvMu.Lock()
		dvW = w
		dvMu.Unlock()

		w.Navigate(url)
		w.Run() // boucle de messages du fil, jusqu'a la fermeture

		dvMu.Lock()
		dvW = nil
		dvMu.Unlock()
	}()
}

// Fermer ferme la fenetre si elle est ouverte.
func Fermer() {
	dvMu.Lock()
	w := dvW
	dvMu.Unlock()
	if w != nil {
		w.Dispatch(func() { w.Terminate() })
	}
}
