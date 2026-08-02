package maj

import "testing"

func TestPlusRecente(t *testing.T) {
	cas := []struct {
		candidate, actuelle string
		attendu             bool
	}{
		{"v2.5.0", "2.4.0", true},
		{"2.4.1", "2.4.0", true},
		{"v3.0.0", "2.9.9", true},
		{"v2.4.0", "2.4.0", false},
		{"v2.3.9", "2.4.0", false},
		{"v2.5.0", "dev", false}, // build local : jamais de mise a jour
		{"n'importe", "2.4.0", false},
	}
	for _, c := range cas {
		if got := plusRecente(c.candidate, c.actuelle); got != c.attendu {
			t.Errorf("plusRecente(%q, %q) = %v, attendu %v", c.candidate, c.actuelle, got, c.attendu)
		}
	}
}
