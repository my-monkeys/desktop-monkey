package main

import (
	"testing"

	"github.com/my-monkeys/desktop-monkey/internal/planche"
)

// Le singe defeque souvent au meme endroit s'il ne bouge pas : ecarterCrotte
// doit garantir qu'aucun tas ne se pose sur un autre, meme demandes au meme
// point.
func TestEcarterCrotteEviteSuperposition(t *testing.T) {
	s := &scene{ecranL: 1000, pc: &planche.Planche{Largeur: 88}}

	for i := 0; i < crotteMax; i++ {
		x := s.ecarterCrotte(500)
		s.crottes = append(s.crottes, &crotte{solX: int(x)})
	}

	for i := range s.crottes {
		for j := i + 1; j < len(s.crottes); j++ {
			d := s.crottes[i].solX - s.crottes[j].solX
			if d < 0 {
				d = -d
			}
			if d < s.pc.Largeur {
				t.Errorf("crottes %d (x=%d) et %d (x=%d) trop proches : %d < %d",
					i, s.crottes[i].solX, j, s.crottes[j].solX, d, s.pc.Largeur)
			}
		}
	}
}

// Sans voisin, la crotte reste exactement ou le singe l'a demandee.
func TestEcarterCrotteLaisseLibre(t *testing.T) {
	s := &scene{ecranL: 1000, pc: &planche.Planche{Largeur: 88}}
	if got := s.ecarterCrotte(320); got != 320 {
		t.Errorf("sans voisin, x devrait rester 320, obtenu %.0f", got)
	}
}
