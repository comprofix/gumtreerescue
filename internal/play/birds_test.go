package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/art"
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

// TestFlockMakesRoom fills every perch, then brings in more birds: each
// time the bird that has sat longest flies off to make room, the
// newcomer takes its perch, and once they are off the screen the birds
// that made room are gone, leaving one bird a perch.
func TestFlockMakesRoom(t *testing.T) {
	f := newFlock(scene.NewSceneryRNG(5))
	for f.free(-1) >= 0 {
		f.arrive(art.Cockatoo, 300, 100, 1)
	}
	settle := func() {
		for range 60 * config.TPS {
			f.step()
			for i := range f.birds {
				f.birds[i].timer = max(f.birds[i].timer, 2) // no bird leaves of itself
			}
		}
	}
	settle()
	n := len(f.perches)
	if len(f.birds) != n {
		t.Fatalf("%d birds for %d perches", len(f.birds), n)
	}
	for range 5 {
		oldest := -1
		for i, b := range f.birds {
			if b.state == perched && (oldest < 0 || b.sat > f.birds[oldest].sat) {
				oldest = i
			}
		}
		perch, id := f.birds[oldest].at, f.birds[oldest].id
		f.arrive(art.Magpie, 300, 100, 1)
		if b := f.birds[oldest]; b.state != flying || b.at != -1 || !b.gone {
			t.Fatalf("the longest sitting bird did not fly off for good: %+v", b)
		}
		if nb := f.birds[len(f.birds)-1]; nb.at != perch {
			t.Fatalf("the newcomer flies to perch %d, not %d, the one made free", nb.at, perch)
		}
		settle()
		for _, b := range f.birds {
			if b.id == id && b.kind != art.Magpie {
				t.Fatal("the bird that made room is still here")
			}
		}
		if len(f.birds) != n {
			t.Fatalf("%d birds for %d perches after making room", len(f.birds), n)
		}
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
