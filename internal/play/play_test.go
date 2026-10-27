package play

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/input"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/scene/scenetest"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// start begins a game on a harness, past the fade in.
func start(t *testing.T, seed uint64) (*scenetest.Harness, *Play) {
	h := scenetest.New(t, seed)
	h.Ctx.NewPlay = New
	h.Ctx.NativePlay = true
	p := New(h.Ctx, config.Skills[1]).(*Play) // Normal
	h.Start(p)
	h.Ticks(scene.FadeTicks)
	return h, p
}

// next is the koala that comes down to the sheet soonest, if any.
func next(w *world.World) (*world.Koala, bool) {
	var best *world.Koala
	soonest := 0
	for i := range w.Koalas {
		k := &w.Koalas[i]
		station, ticks, ok := w.Landing(k)
		if !ok || station == w.Row.UteStation {
			continue
		}
		if best == nil || ticks < soonest {
			best, soonest = k, ticks
		}
	}
	return best, best != nil
}

// botKey is the arrow key a steady player holds: towards where the next
// koala lands in zone 4, a safe bounce right, or with none to catch, back
// towards the drop point to be ready.
func botKey(w *world.World) []ebiten.Key {
	station := w.Row.Drops[0].Station
	if k, ok := next(w); ok {
		station = k.Target
	}
	middle := (config.ZoneSharp + config.ZoneSafe + config.ZoneCentre + config.ZoneSafe/2 - config.SheetWidth/2) * config.SubPx
	want := world.StationX(station) - middle
	switch d := want - w.Sheet; {
	case d > w.RangerSpeed():
		return []ebiten.Key{ebiten.KeyArrowRight}
	case d < -w.RangerSpeed():
		return []ebiten.Key{ebiten.KeyArrowLeft}
	}
	return nil
}

// summary is the game's state as text, for the golden files.
func summary(p *Play) string {
	w := p.World()
	var b strings.Builder
	fmt.Fprintf(&b, "level %d  score %d  lives %d  over %v\n", w.Level, w.Session.Score, w.Session.Lives, w.Over)
	fmt.Fprintf(&b, "waiting %d  rescued %d  misses %d  sheet %d\n", w.Waiting, w.Rescued, w.Misses, w.Sheet)
	for i, k := range w.Koalas {
		fmt.Fprintf(&b, "koala %d: state %d x %d h %d target %d missed %v bonked %v\n", i, k.State, k.X, k.H, k.Target, k.Missed, k.Bonked)
	}
	return b.String()
}

// TestGoldenReplay plays three levels' worth of ticks with the steady
// player, by keys, and checks the game ends exactly as recorded. A change
// to the rules, tuning or randomness shows up here; if it is intended, run
// the test with -update.
func TestGoldenReplay(t *testing.T) {
	h, p := start(t, 42)
	for range 3 * config.KoalaLevel * config.LevelAt(1).GapTicks {
		h.Run(input.Step{Keys: botKey(p.World()), Ticks: 1})
	}
	w := p.World()
	if w.Misses != 0 || w.Session.Lives != w.Session.Skill.Lives || w.Level < 2 {
		t.Errorf("the steady player missed: level %d, lives %d, misses %d", w.Level, w.Session.Lives, w.Misses)
	}
	scenetest.Golden(t, "replay", summary(p))
}

// TestWalkAwayGameOver plays a game with the rangers walked away to the
// right: every koala is missed, and the game ends and goes on to the high
// scores. (Standing still would bounce the first koala straight up for
// ever, as the rangers start under the drop.)
func TestWalkAwayGameOver(t *testing.T) {
	h, p := start(t, 7)
	for range 100 * config.LevelAt(1).GapTicks {
		if p.World().Over {
			break
		}
		h.Run(scenetest.Hold(1, ebiten.KeyArrowRight))
	}
	scenetest.Current[*Ended](h)
	scenetest.Golden(t, "walk_away", summary(p))
	h.Ticks(config.GameOverHold + 2*scene.FadeTicks)
	scenetest.Current[*scene.HighScoreTable](h)
}

// TestMouseSteers checks the sheet heads for the pointer at the rangers'
// speed, in either look, and that a key takes steering back.
func TestMouseSteers(t *testing.T) {
	for _, look := range []config.Preset{config.Smooth, config.Modern} {
		t.Run(look.Name, func(t *testing.T) {
			h, p := start(t, 1)
			h.Ctx.Preset = look
			w := p.World()
			w.Row.GapTicks = 1 << 20 // one koala, so the game can't end
			k := float64(look.ScreenWidth) / float64(config.Modern.ScreenWidth)
			_, hi := w.SheetRange()
			target := hi / config.SubPx
			before := w.Sheet
			h.Run(input.MoveTo(int(float64(target)*k), 100))
			if got := w.Sheet - before; got != w.RangerSpeed() {
				t.Fatalf("the sheet moved %d in a tick towards the pointer, want %d", got, w.RangerSpeed())
			}
			h.Run(input.Idle(1000))
			if w.Sheet != hi {
				t.Errorf("sheet at %d, want the pointer's %d", w.Sheet, hi)
			}
			h.Run(scenetest.Hold(10, ebiten.KeyArrowLeft), input.Idle(10))
			if w.Sheet >= hi {
				t.Error("a key did not take steering back from the mouse")
			}
		})
	}
}

func TestPause(t *testing.T) {
	h, _ := start(t, 1)
	h.Run(input.Press(ebiten.KeyEscape))
	scenetest.Current[*scene.Pause](h)
}
