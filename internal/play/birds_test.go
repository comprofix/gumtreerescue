package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/scene"
)

// TestFlock runs the birds for ten minutes: no two ever sit on or fly to
// the same perch, they stay in the field's sky but for the edges they
// leave by, and they do fly about, off the screen and back.
func TestFlock(t *testing.T) {
	f := newFlock(scene.NewSceneryRNG(7))
	if len(f.birds) != birdCount {
		t.Fatalf("%d birds, want %d", len(f.birds), birdCount)
	}
	fw, _ := config.Field()
	flights, aways := 0, 0
	for range 10 * 60 * config.TPS {
		was := make([]birdState, len(f.birds))
		for i, b := range f.birds {
			was[i] = b.state
		}
		f.step()
		held := make(map[int]bool)
		for i, b := range f.birds {
			if b.state != away && b.at >= 0 {
				if held[b.at] {
					t.Fatalf("two birds on perch %d", b.at)
				}
				held[b.at] = true
				if !f.taken[b.at] {
					t.Fatalf("perch %d held but not taken", b.at)
				}
			}
			if b.state == flying {
				x, y, _ := b.where()
				if x < -offScreen-1 || x > float64(fw)+offScreen+1 || y < config.HUDHeight || y > config.GroundY {
					t.Fatalf("bird %d flew to (%.1f, %.1f)", i, x, y)
				}
			}
			if was[i] != flying && b.state == flying {
				flights++
			}
			if was[i] != away && b.state == away {
				aways++
			}
		}
		for i, taken := range f.taken {
			if taken && !held[i] {
				t.Fatalf("perch %d taken but no bird holds it", i)
			}
		}
	}
	if flights < 50 || aways < 10 {
		t.Errorf("%d flights and %d off the screen, want more", flights, aways)
	}
}

// TestFlockLeavesPlay checks the birds take nothing from the world's
// random numbers: a game with them plays the same as one without.
func TestFlockLeavesPlay(t *testing.T) {
	ha, a := start(t, 3)
	hb, b := start(t, 3)
	b.birds = flock{}
	ha.Ticks(40 * config.TPS)
	hb.Ticks(40 * config.TPS)
	if len(a.birds.birds) == 0 {
		t.Fatal("no birds")
	}
	wa, wb := a.World(), b.World()
	if wa.Session.Score != wb.Session.Score || wa.Session.Lives != wb.Session.Lives || len(wa.Koalas) != len(wb.Koalas) {
		t.Fatal("the birds changed the game")
	}
	for i := range wa.Koalas {
		if ka, kb := wa.Koalas[i], wb.Koalas[i]; ka.X != kb.X || ka.H != kb.H || ka.State != kb.State {
			t.Fatalf("koala %d differs with the birds", i)
		}
	}
}
