package play

import (
	"math"
	"testing"

	"github.com/comprofix/gumtreerescue/internal/art"
)

// Every ranger, arms and stride finds its sprite and frame in the sheet,
// with its feet inside it (DESIGN D30).
func TestRangerSprites(t *testing.T) {
	index := sheetIndex(t, "ranger")
	for _, woman := range []bool{false, true} {
		for _, arms := range []rangerArms{armsHold, armsDown, armsCarry, armsPick} {
			for _, s := range []art.Stance{{}, {Shut: true}, {Stride: 1, Step: 7}, {Stride: 0.8, Step: -2, Shut: true}} {
				s.Woman = woman
				name, frame := rangerSprite(s, arms)
				sp, ok := index[name]
				if !ok {
					t.Fatalf("%+v %s: no sprite %s", s, arms, name)
				}
				if frame < 0 || frame >= sp.Frames || rangerFeet > sp.H || sp.W%2 != 0 {
					t.Errorf("%s: frame %d of %d, feet %d in %d×%d", name, frame, sp.Frames, rangerFeet, sp.W, sp.H)
				}
			}
		}
	}
}

// A ranger standing shows frame 0; walking, the stride's nearest frame.
func TestRangerFrames(t *testing.T) {
	for _, c := range []struct {
		s     art.Stance
		frame int
	}{
		{art.Stance{Stride: 0.3, Step: 2}, 0},
		{art.Stance{Stride: 1}, 1},
		{art.Stance{Stride: 1, Step: math.Pi}, 1 + rangerSteps/2},
		{art.Stance{Stride: 1, Step: -2 * math.Pi / rangerSteps}, rangerSteps},
	} {
		if _, f := rangerSprite(c.s, armsHold); f != c.frame {
			t.Errorf("%+v: frame %d, want %d", c.s, f, c.frame)
		}
	}
}
