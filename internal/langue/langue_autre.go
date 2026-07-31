//go:build !windows

package langue

import (
	"os"
	"strings"
)

// Francais indique si l'utilisateur travaille en francais.
func Francais() bool {
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if s := os.Getenv(v); s != "" {
			return strings.HasPrefix(strings.ToLower(s), "fr")
		}
	}
	return false
}
