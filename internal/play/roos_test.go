package play

import (
	"math"
	"testing"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/scene"
)

// TestMob runs the kangaroos for ten minutes: they keep to their stretch
// of paddock behind the fence, the joeys at foot never stray far from
// their mothers, and every grown one grazes, sits up and hops.
func TestMob(t *testing.T) {
	m := newMob(scene.NewSceneryRNG(11))
	grown, joeys, pouches := 0, 0, 0
	for _, r := range m.roos {
		switch {
		case r.mum >= 0:
			joeys++
		case r.pouch:
			pouches++
			grown++
		default:
			grown++
		}
	}
	if grown != 3 || joeys != 2 || pouches != 1 {
		t.Fatalf("%d grown, %d joeys at foot and %d in pouches, want 3, 2 and 1", grown, joeys, pouches)
	}
	seen := make([]map[rooState]bool, len(m.roos))
	for i := range seen {
		seen[i] = map[rooState]bool{}
	}
	behindGum := false
	for range 10 * 60 * config.TPS {
		m.step()
		for i, r := range m.roos {
			seen[i][r.state] = true
			behindGum = behindGum || r.x < 150
			if r.x < paddockLo-hopLen || r.x > paddockHi+hopLen || r.y < paddockTop-1 || r.y > paddockLow+3 {
				t.Fatalf("kangaroo %d at (%.1f, %.1f), out of its paddock", i, r.x, r.y)
			}
			if r.mum >= 0 {
				if d := math.Abs(r.x - m.roos[r.mum].x); d > 6*hopLen {
					t.Fatalf("joey %d strayed %.1f from its mother", i, d)
				}
			}
		}
	}
	if !behindGum {
		t.Error("no kangaroo hopped into the farmland behind the big gum")
	}
	for i, r := range m.roos {
		if r.mum < 0 && (!seen[i][grazing] || !seen[i][sittingUp] || !seen[i][hopping]) {
			t.Errorf("kangaroo %d only did %v", i, seen[i])
		}
		if r.mum >= 0 && !seen[i][hopping] {
			t.Errorf("joey %d never hopped after its mother", i)
		}
	}
}
