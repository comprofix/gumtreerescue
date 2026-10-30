package play

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/comprofix/gumtreerescue/assets"
	"github.com/comprofix/gumtreerescue/internal/art"
)

// koalaSheet is the koala sheet's index: each sprite's size and frames.
func koalaSheet(t *testing.T) map[string]struct{ W, H, Frames int } {
	t.Helper()
	data, err := assets.FS.ReadFile("gen/koala.json")
	if err != nil {
		t.Fatal(err)
	}
	var index map[string]struct{ W, H, Frames int }
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatal(err)
	}
	return index
}

// Every pose, in every look, finds its sprite and frame in the sheet, with
// its feet inside it (DESIGN D30).
func TestKoalaSprites(t *testing.T) {
	index := koalaSheet(t)
	looks := []art.Look{{}, {Shut: true}, {Flick: 1}, {Joey: true}, {Joey: true, Shut: true}}
	for kind := kSit; kind <= kFall; kind++ {
		for _, l := range looks {
			for _, baby := range []bool{false, true} {
				for _, k := range []koalaPose{
					{}, {reach: 1}, {follow: 1}, {down: true}, {squash: 1},
					{flail: -1}, {flail: 1}, {phase: 5}, {phase: -2},
					{turn: -1.2}, {turn: 0.4}, {turn: -3}, {turn: 3, tumble: true}, {turn: -0.3, tumble: true},
				} {
					k.kind, k.look, k.baby = kind, l, baby
					if baby {
						k.look.Joey = false
					}
					base, look, frame := k.sprite()
					s, ok := index[base+look]
					if !ok {
						t.Fatalf("%+v: no sprite %s", k, base+look)
					}
					if frame < 0 || frame >= s.Frames {
						t.Errorf("%+v: frame %d of %s, which has %d", k, frame, base+look, s.Frames)
					}
					if feet, ok := koalaFeet[base]; !ok || feet > s.H || s.W%2 != 0 {
						t.Errorf("%s: feet %d (known %v) in %d×%d", base, feet, ok, s.W, s.H)
					}
				}
			}
		}
	}
}

// Turned and moving poses pick the frame nearest to them.
func TestKoalaFrames(t *testing.T) {
	for _, c := range []struct {
		k     koalaPose
		base  string
		frame int
	}{
		{koalaPose{kind: kGrip, turn: -2 * koalaSlope, reach: 1}, "koala_grip", 2*3 + 1},
		{koalaPose{kind: kGrip, turn: -9 * koalaSlope}, "koala_grip", koalaGripSteps * 3},
		{koalaPose{kind: kWalk, turn: koalaSlope, phase: 2 * math.Pi / koalaStrides}, "koala_walk", 3*koalaStrides + 1},
		{koalaPose{kind: kFall, tumble: true, turn: -math.Pi / 2}, "koala_tumble", koalaTumbles * 3 / 4},
		{koalaPose{kind: kCling, reach: 0.6, down: true}, "koala_clingdown", 2},
		{koalaPose{kind: kThump, squash: 1}, "koala_thump", 2},
		{koalaPose{kind: kChew, phase: math.Pi}, "koala_chew", 2},
		{koalaPose{kind: kAsleep, baby: true}, "joey_sit", 0},
	} {
		base, _, frame := c.k.sprite()
		if base != c.base || frame != c.frame {
			t.Errorf("%+v: %s frame %d, want %s frame %d", c.k, base, frame, c.base, c.frame)
		}
	}
}
