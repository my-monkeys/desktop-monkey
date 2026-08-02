// Package maj garde le singe a jour tout seul : il interroge la derniere
// release GitHub du projet, compare a la version compilee, telecharge le
// binaire de sa plateforme et remplace l'executable en place. L'appelant n'a
// plus qu'a redemarrer.
//
// Le remplacement se fait par renommage : l'executable courant devient
// <exe>.ancien (un fichier ouvert se renomme tres bien, meme sous Windows),
// le nouveau prend sa place, et le .ancien est balaye au prochain demarrage.
package maj

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const depot = "my-monkeys/desktop-monkey"

// nomAsset renvoie le nom de l'asset de release pour cette plateforme.
func nomAsset() string {
	switch runtime.GOOS {
	case "windows":
		return "monkey.exe"
	case "darwin":
		return "monkey-mac"
	}
	return ""
}

type releaseGitHub struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Verifier interroge la derniere release. Elle renvoie le tag et l'URL du
// binaire de la plateforme si une version plus recente que l'actuelle existe.
func Verifier(versionActuelle string) (tag, url string, ok bool) {
	asset := nomAsset()
	if asset == "" {
		return "", "", false
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/" + depot + "/releases/latest")
	if err != nil {
		return "", "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", false
	}
	var r releaseGitHub
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&r); err != nil {
		return "", "", false
	}
	if !plusRecente(r.TagName, versionActuelle) {
		return "", "", false
	}
	for _, a := range r.Assets {
		if a.Name == asset {
			return r.TagName, a.URL, true
		}
	}
	return "", "", false
}

// plusRecente compare deux versions "x.y.z" (le prefixe v est tolere).
func plusRecente(candidate, actuelle string) bool {
	c, okC := parseVersion(candidate)
	a, okA := parseVersion(actuelle)
	if !okC || !okA {
		return false
	}
	for i := 0; i < 3; i++ {
		if c[i] != a[i] {
			return c[i] > a[i]
		}
	}
	return false
}

func parseVersion(v string) ([3]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	morceaux := strings.SplitN(v, ".", 3)
	if len(morceaux) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, m := range morceaux {
		n, err := strconv.Atoi(m)
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}

// Appliquer telecharge le binaire et le met a la place de l'executable
// courant. Au retour sans erreur, il ne reste qu'a redemarrer.
func Appliquer(url string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("telechargement : HTTP %d", resp.StatusCode)
	}

	neuf := exe + ".neuf"
	f, err := os.OpenFile(neuf, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(neuf)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(neuf)
		return err
	}

	ancien := exe + ".ancien"
	os.Remove(ancien)
	if err := os.Rename(exe, ancien); err != nil {
		os.Remove(neuf)
		return err
	}
	if err := os.Rename(neuf, exe); err != nil {
		// on remet l'ancien en place : mieux vaut un vieux singe qu'aucun
		os.Rename(ancien, exe)
		os.Remove(neuf)
		return err
	}
	return nil
}

// NettoyerAncien balaye le reste d'une mise a jour passee. A appeler au
// demarrage ; l'echec est sans gravite (l'ancien processus vit peut-etre
// encore), le prochain demarrage reessaiera.
func NettoyerAncien() {
	if exe, err := os.Executable(); err == nil {
		os.Remove(exe + ".ancien")
	}
}
