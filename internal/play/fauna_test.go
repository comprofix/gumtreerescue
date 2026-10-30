package play

import (
	"fmt"
	"testing"

	"github.com/comprofix/gumtreerescue/internal/art"
)

// Every bird, kangaroo and dazed kookaburra finds its sprite and frame in
// the sheet, centred on its origin (DESIGN D30).
func TestFaunaSprites(t *testing.T) {
	index := sheetIndex(t, "fauna")
	check := func(name string, frame int) {
		t.Helper()
		sp, ok := index[name]
		switch {
		case !ok:
			t.Errorf("no sprite %s", name)
		case frame < 0 || frame >= sp.Frames:
			t.Errorf("%s: frame %d of %d", name, frame, sp.Frames)
		case sp.W%2 != 0 || sp.H%2 != 0:
			t.Errorf("%s: %d×%d, not centred", name, sp.W, sp.H)
		}
	}
	for k := range art.NumSpecies {
		for _, s := range []float64{0.3, 0.7, 0.92, 1.2} {
			for _, w := range []float64{-1, -0.2, 0.6, 1} {
				check(birdSprite(k, s, true, 0, w))
			}
			check(birdSprite(k, s, false, 0, 0))
			check(birdSprite(k, s, false, 1, 0))
		}
	}
	for _, s := range []float64{0.5, 0.9, 1} {
		for _, pouch := range []bool{false, true} {
			for _, r := range []art.Roo{{Hop: -1}, {Hop: -1, Graze: 1, Chew: 4, Flick: 1}, {Hop: 0}, {Hop: 0.99}} {
				check(rooSprite(s, pouch, r))
			}
		}
	}
	for _, s := range []float64{0.2, 0.55, 1.05} {
		check(fmt.Sprintf("dizzy_%d", nearestSize(dizzySizes, s)), wingFrame(0))
	}
}

// Wings pick their frame from up to down, and sizes the nearest drawn.
func TestFaunaFrames(t *testing.T) {
	if wingFrame(1) != 0 || wingFrame(-1) != wingFrames-1 {
		t.Errorf("wings up %d, down %d", wingFrame(1), wingFrame(-1))
	}
	if n := nearestSize(birdSizes, 0.87); n != 80 {
		t.Errorf("size 0.87 is drawn at %d, want 80", n)
	}
	if name, f := rooSprite(0.5, true, art.Roo{Hop: 0.5}); name != "roo_joey_hop" || f != hopFrames/2 {
		t.Errorf("a joey half through a hop: %s frame %d", name, f)
	}
}
