package world

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/session"
)

// normal is the skill the tests play at.
var normal = config.Skills[1]

// newWorld is a game on level 1 at Normal, past the level's opening pause,
// with the first koala about to start its wobble.
func newWorld(t *testing.T) *World {
	t.Helper()
	return seeded(t, 1)
}

// seeded is newWorld with its own seed.
func seeded(t *testing.T, seed uint64) *World {
	t.Helper()
	w := New(session.New(normal, normal.Lives), 1, rand.New(rand.NewPCG(seed, 2)))
	w.Starting, w.nextDrop = 0, 1
	return w
}

// spotsDistinct fails t if two koalas in the tree share a spot.
func spotsDistinct(t *testing.T, w *World) {
	t.Helper()
	seen := map[int]bool{}
	for _, s := range w.Tree {
		if seen[s.Spot] {
			t.Fatalf("two koalas in spot %d: %+v", s.Spot, w.Tree)
		}
		seen[s.Spot] = true
	}
}

// zoneMiddle is the offset of zone z's middle from the sheet's middle, in
// fixed point.
func zoneMiddle(z int) int {
	widths := []int{config.ZoneSharp, config.ZoneSafe, config.ZoneCentre, config.ZoneSafe, config.ZoneSharp}
	at := 0
	for i := range z - 1 {
		at += widths[i]
	}
	return (2*at + widths[z-1] - config.SheetWidth) * config.SubPx / 2
}

// holdFor is the input that keeps the sheet where a koala at x lands in
// zone z.
func holdFor(x, z int) Input { return Input{Aim: true, Target: x - zoneMiddle(z)} }

// run updates w with in until done reports true, failing after limit ticks.
func run(t *testing.T, w *World, in func() Input, limit int, done func([]Event) bool) {
	t.Helper()
	for range limit {
		if done(w.Update(in())) {
			return
		}
	}
	t.Fatalf("not done after %d ticks", limit)
}

func has(es []Event, k EventKind) bool {
	for _, e := range es {
		if e.Kind == k {
			return true
		}
	}
	return false
}

func TestZonesAddUp(t *testing.T) {
	if got := 2*config.ZoneSharp + 2*config.ZoneSafe + config.ZoneCentre; got != config.SheetWidth {
		t.Errorf("zones add up to %d, sheet is %d", got, config.SheetWidth)
	}
	if config.SheetWidth >= config.StepPx {
		t.Errorf("sheet %d is not under one step %d (DESIGN §5)", config.SheetWidth, config.StepPx)
	}
}

func TestZoneAt(t *testing.T) {
	w := newWorld(t)
	w.Sheet = StationX(1)
	for z := 1; z <= 5; z++ {
		if got, ok := w.ZoneAt(w.Sheet + zoneMiddle(z)); !ok || got != z {
			t.Errorf("middle of zone %d: got %d, %v", z, got, ok)
		}
	}
	half := config.SheetWidth * config.SubPx / 2
	for _, x := range []int{w.Sheet - half - 1, w.Sheet + half + 1} {
		if _, ok := w.ZoneAt(x); ok {
			t.Errorf("x %d off the sheet counted as on it", x)
		}
	}
}

// TestArcsLandOnStations bounces a koala from station 1 in each zone and
// checks it comes down exactly on its station after exactly one bounce's
// time (DESIGN §5, PLAN §3).
func TestArcsLandOnStations(t *testing.T) {
	for z := 1; z <= 5; z++ {
		w := newWorld(t)
		w.Row.UteStation = 4 // room to land anywhere
		k := Koala{State: Falling, X: StationX(1), Target: 1, h0: 100 * config.SubPx, H: 100 * config.SubPx}
		w.Koalas, w.Tree = append(w.Koalas, k), nil
		in := holdFor(StationX(1), z)
		w.Sheet = in.Target
		run(t, w, func() Input { return in }, 200, func(es []Event) bool { return has(es, Bounced) })
		got := &w.Koalas[0]
		want := 1 + config.ZoneSteps[z-1]
		if got.Target != want {
			t.Fatalf("zone %d: aimed at station %d, want %d", z, got.Target, want)
		}
		if want < 0 {
			continue // past the tree: see TestTreeGrab
		}
		w.Sheet = StationX(want) + config.SheetWidth*config.SubPx // out of the way
		for range w.Row.BounceTicks - 1 {
			w.Update(Input{})
			if got.H <= 0 {
				t.Fatalf("zone %d: came down early, at tick %d", z, got.t)
			}
		}
		w.Update(Input{})
		if got.X != StationX(want) || got.H != 0 && !got.Missed {
			t.Errorf("zone %d: after one bounce at x %d h %d, want x %d h 0", z, got.X, got.H, StationX(want))
		}
	}
}

func TestBouncePoints(t *testing.T) {
	for _, c := range []struct {
		zone  int
		bonks bool
		want  int
	}{
		{1, false, 0}, {2, false, 0}, {3, false, 0},
		{4, false, config.SafePoints}, {5, false, config.SharpPoints},
		{5, true, 0},
	} {
		if got := bouncePoints(c.zone, c.bonks); got != c.want {
			t.Errorf("zone %d bonks %v: %d points, want %d", c.zone, c.bonks, got, c.want)
		}
	}
}

// TestSafeRescue catches the first koala in zone 4 every time: it goes
// station by station into the tray.
func TestSafeRescue(t *testing.T) {
	w := newWorld(t)
	w.Row.GapTicks = 1 << 20 // one koala
	in := func() Input {
		for i := range w.Koalas {
			if k := &w.Koalas[i]; k.State == Falling || k.State == Arcing {
				return holdFor(StationX(k.Target), 4)
			}
		}
		return Input{}
	}
	run(t, w, in, 1000, func(es []Event) bool { return has(es, Rescued) })
	want := w.Row.UteStation*config.SafePoints + config.RescuePoints
	if w.Session.Score != want || w.Rescued != 1 || w.Session.Lives != normal.Lives {
		t.Errorf("score %d rescued %d lives %d, want %d, 1, %d", w.Session.Score, w.Rescued, w.Session.Lives, want, normal.Lives)
	}
}

// TestBackAndForthScoresOnce bounces a koala back and forth before sending
// it on: only stations new to it score, so it earns what a straight trip
// does (DESIGN §5).
func TestBackAndForthScoresOnce(t *testing.T) {
	w := newWorld(t)
	w.Row.GapTicks = 1 << 20 // one koala
	back := 20
	in := func() Input {
		for i := range w.Koalas {
			if k := &w.Koalas[i]; k.State == Falling || k.State == Arcing {
				z := 4
				if k.Target == 1 && back > 0 && k.t == w.Row.BounceTicks-1 {
					back--
				}
				if k.Target == 1 && back > 0 {
					z = 2
				}
				return holdFor(StationX(k.Target), z)
			}
		}
		return Input{}
	}
	run(t, w, in, 100000, func(es []Event) bool { return has(es, Rescued) })
	if back != 0 {
		t.Fatalf("%d trips back left to make", back)
	}
	if want := w.Row.UteStation*config.SafePoints + config.RescuePoints; w.Session.Score != want {
		t.Errorf("score %d, want %d, a straight trip's", w.Session.Score, want)
	}
}

// TestFurthestLastsTheLevel: a koala back from the tree scores only past
// where it had reached, and the next level starts afresh (DESIGN §5).
func TestFurthestLastsTheLevel(t *testing.T) {
	w := newWorld(t)
	w.Row.GapTicks = 1 << 20 // one koala
	w.furthest[0] = 2
	in := func() Input {
		for i := range w.Koalas {
			if k := &w.Koalas[i]; k.State == Falling || k.State == Arcing {
				return holdFor(StationX(k.Target), 4)
			}
		}
		return Input{}
	}
	run(t, w, in, 1000, func(es []Event) bool { return has(es, Rescued) })
	if want := config.SafePoints + config.RescuePoints; w.Session.Score != want {
		t.Errorf("score %d, want %d, one new station and the rescue", w.Session.Score, want)
	}
	w.startLevel(2)
	if w.furthest[0] != 0 {
		t.Errorf("furthest %d on a new level", w.furthest[0])
	}
}

// TestMissCrawlsBack lets a koala fall past the sheet: a life lost, then
// dazed, crawling and climbing back to a free spot to wait again (DESIGN
// §5).
func TestMissCrawlsBack(t *testing.T) {
	w := newWorld(t)
	w.Row.GapTicks = 1 << 20 // only one koala
	away := Input{Dir: 1}
	run(t, w, func() Input { return away }, 1000, func(es []Event) bool { return has(es, Missed) })
	if w.Session.Lives != normal.Lives-1 || w.Misses != 1 {
		t.Fatalf("lives %d misses %d after a miss", w.Session.Lives, w.Misses)
	}
	k := &w.Koalas[0]
	if k.State != Dazed || k.H != -config.SheetHeight*config.SubPx {
		t.Fatalf("missed koala is %v at h %d", k.State, k.H)
	}
	waiting := len(w.Tree)
	run(t, w, func() Input { return away }, 2000, func([]Event) bool { return len(w.Koalas) == 0 })
	if len(w.Tree) != waiting+1 {
		t.Errorf("waiting %d, want %d: the koala did not climb back", len(w.Tree), waiting+1)
	}
	spotsDistinct(t, w)
}

// TestTreeHit bounces a mother with a joey left from the first station, on
// many seeds: she grabs the tree and climbs to a spot with her joey, or
// slips to the ground and loses it until it finds her. No life is lost
// either way, and both happen (DESIGN §5, §6).
func TestTreeHit(t *testing.T) {
	grabs, slips := 0, 0
	for seed := range uint64(20) {
		for _, z := range []int{1, 2} {
			w := seeded(t, seed)
			w.Row.GapTicks = 1 << 20
			w.Tree[0].Joey = true
			in := holdFor(StationX(0), z)
			caught, lost, found := false, false, false
			var hit Event
			run(t, w, func() Input {
				if caught {
					return Input{}
				}
				return in
			}, 1000, func(es []Event) bool {
				caught = caught || has(es, Bounced)
				for _, e := range es {
					if e.Kind == Grabbed || e.Kind == Slipped {
						hit = e
						return true
					}
				}
				return false
			})
			k := w.Koalas[0]
			if k.State != Thumped || k.X != grabX(k.H) || k.H <= 0 {
				t.Fatalf("seed %d zone %d: hit the tree, then %v at %d,%d", seed, z, k.State, k.X, k.H)
			}
			for range config.ThumpTicks {
				w.Update(Input{})
			}
			k = w.Koalas[0]
			switch hit.Kind {
			case Grabbed:
				grabs++
				if k.State != Climbing || !k.Joey {
					t.Fatalf("seed %d zone %d: grabbed, then %v joey %v", seed, z, k.State, k.Joey)
				}
			case Slipped:
				slips++
				if k.State != Falling || !k.Slipped {
					t.Fatalf("seed %d zone %d: slipped, then %v", seed, z, k.State)
				}
			}
			run(t, w, func() Input { return Input{} }, 3000, func(es []Event) bool {
				lost = lost || has(es, JoeyLost)
				found = found || has(es, JoeyFound)
				return len(w.Koalas) == 0
			})
			if lost != (hit.Kind == Slipped) || found != lost {
				t.Errorf("seed %d zone %d: %v, joey lost %v found %v", seed, z, hit.Kind, lost, found)
			}
			if back := w.Tree[len(w.Tree)-1]; back.ID != 0 || !back.Joey {
				t.Errorf("seed %d zone %d: back of the queue %+v, want koala 0 with her joey", seed, z, back)
			}
			if w.Session.Lives != normal.Lives || w.Misses != 0 {
				t.Errorf("seed %d zone %d: lives %d misses %d", seed, z, w.Session.Lives, w.Misses)
			}
			spotsDistinct(t, w)
		}
	}
	if grabs == 0 || slips == 0 {
		t.Errorf("grabs %d slips %d: want both", grabs, slips)
	}
}

// TestHeadboardBonk sends a koala sharp from the station before the ute: it
// bonks the headboard and drops into the tray, rescued, with nothing for
// that bounce (DESIGN §5).
func TestHeadboardBonk(t *testing.T) {
	w := newWorld(t)
	from := w.Row.UteStation - 1
	w.Koalas = append(w.Koalas, Koala{State: Falling, X: StationX(from), Target: from, h0: 50 * config.SubPx, H: 50 * config.SubPx})
	w.Tree = nil
	in := holdFor(StationX(from), 5)
	w.Sheet = in.Target
	bonked := false
	run(t, w, func() Input { return in }, 1000, func(es []Event) bool {
		bonked = bonked || has(es, Bonked)
		return has(es, Rescued)
	})
	if !bonked {
		t.Error("rescued without bonking the headboard")
	}
	k := &w.Koalas[0]
	if k.State != InTray || !k.Bonked || k.Timer != config.BonkedTicks {
		t.Errorf("koala %v bonked %v timer %d", k.State, k.Bonked, k.Timer)
	}
	if w.Session.Score != config.RescuePoints {
		t.Errorf("score %d, want just the rescue, %d", w.Session.Score, config.RescuePoints)
	}
}

// TestLevelDone puts nine koalas in the tray and rescues the tenth: the
// clean bonus, then the next level.
func TestLevelDone(t *testing.T) {
	w := newWorld(t)
	w.Rescued, w.Tree = config.KoalaLevel-1, nil
	from := w.Row.UteStation - 1
	w.Koalas = append(w.Koalas, Koala{State: Falling, X: StationX(from), Target: from, h0: 50 * config.SubPx, H: 50 * config.SubPx})
	in := holdFor(StationX(from), 4)
	w.Sheet = in.Target
	run(t, w, func() Input { return in }, 1000, func(es []Event) bool { return has(es, LevelDone) })
	if want := config.SafePoints + config.RescuePoints + config.CleanBonus; w.Session.Score != want {
		t.Errorf("score %d, want %d with the clean bonus", w.Session.Score, want)
	}
	run(t, w, func() Input { return Input{} }, config.EndTicks+1, func([]Event) bool { return w.Level == 2 })
	if len(w.Tree) != config.KoalaLevel || w.Rescued != 0 || len(w.Koalas) != 0 || w.Starting == 0 {
		t.Errorf("level 2 starts with waiting %d rescued %d koalas %d starting %d", len(w.Tree), w.Rescued, len(w.Koalas), w.Starting)
	}
}

func TestGameOver(t *testing.T) {
	w := newWorld(t)
	away := Input{Dir: 1}
	run(t, w, func() Input { return away }, 20000, func(es []Event) bool { return has(es, GameOver) })
	if !w.Over || w.Session.Lives != 0 || w.Misses != normal.Lives {
		t.Errorf("over %v lives %d misses %d", w.Over, w.Session.Lives, w.Misses)
	}
	if es := w.Update(away); len(es) != 0 {
		t.Errorf("events after game over: %v", es)
	}
}

func TestSheetRangeAndSpeed(t *testing.T) {
	w := newWorld(t)
	w.Tree = nil // no koalas, so the game can't end
	lo, hi := w.SheetRange()
	for range 1000 {
		w.Update(Input{Dir: -1})
	}
	if w.Sheet != lo {
		t.Errorf("sheet at %d pushing left, want %d", w.Sheet, lo)
	}
	w.Update(Input{Aim: true, Target: hi * 2})
	if got := w.Sheet - lo; got != w.RangerSpeed() {
		t.Errorf("the mouse moved the sheet %d in a tick, faster than the rangers' %d", got, w.RangerSpeed())
	}
	// Two steps in one bounce, a little more (DESIGN §5).
	if reach := w.RangerSpeed() * w.Row.BounceTicks; reach < 2*config.StepPx*config.SubPx {
		t.Errorf("rangers reach %d in a bounce, under two steps", reach)
	}
}

// TestZonesReachable checks the rangers, kept out of the trunk and the
// tray, can still catch a koala at the first station in every zone that
// sends it on, and at the station before the ute in every zone but sharp
// left, on every level.
func TestZonesReachable(t *testing.T) {
	for n := range config.Levels {
		w := newWorld(t)
		w.startLevel(n + 1)
		lo, hi := w.SheetRange()
		check := func(station, zone int) {
			if s := StationX(station) - zoneMiddle(zone); s < lo || s > hi {
				t.Errorf("level %d: zone %d at station %d needs the sheet at %d, outside %d–%d", n+1, zone, station, s, lo, hi)
			}
		}
		for z := 2; z <= 5; z++ {
			check(0, z)
			check(w.Row.UteStation-1, z)
		}
	}
}

// TestUteParking checks every level parks the ute far enough from every
// drop point that no single catch reaches the tray, with the whole tray on
// screen; the cab may go off the edge (DESIGN §5).
func TestUteParking(t *testing.T) {
	fw, _ := config.Field()
	for n, row := range config.Levels {
		for _, d := range row.Drops {
			if row.UteStation-d.Station < config.MinUteSteps {
				t.Errorf("level %d: ute at station %d, %d from the drop at %d", n+1, row.UteStation, row.UteStation-d.Station, d.Station)
			}
		}
		if end := StationX(row.UteStation)/config.SubPx + config.Headboard; end > fw {
			t.Errorf("level %d: the tray ends at %d, off the %d-wide field", n+1, end, fw)
		}
	}
}

func TestLanding(t *testing.T) {
	w := newWorld(t)
	w.Update(Input{})
	k := &w.Koalas[0]
	station, ticks, ok := w.Landing(k)
	if !ok || station != k.Target {
		t.Fatalf("landing %d %d %v", station, ticks, ok)
	}
	for range ticks - 1 {
		w.Update(Input{})
	}
	if k.H <= 0 {
		t.Fatalf("down a tick early")
	}
	w.Update(Input{})
	if k.State == Falling && !k.Missed {
		t.Errorf("not down when Landing said, h %d", k.H)
	}
}

// TestWobble checks a koala walks from its spot out to the tip of its
// branch, then wobbles there for WobbleTicks before it drops (DESIGN §5).
func TestWobble(t *testing.T) {
	w := newWorld(t)
	spot := w.Tree[0].Spot
	if es := w.Update(Input{}); !has(es, Leaving) || w.Koalas[0].State != Walking {
		t.Fatalf("did not leave its spot: %v", es)
	}
	k := &w.Koalas[0]
	if x, h := SpotAt(config.Spots[spot]); k.X != x || k.H != h {
		t.Fatalf("set off from %d,%d, not its spot %d,%d", k.X, k.H, x, h)
	}
	run(t, w, func() Input { return Input{} }, 1000, func(es []Event) bool { return has(es, Wobble) })
	d := w.Row.Drops[0]
	if k.State != Wobbling || k.X != StationX(d.Station) || k.H != d.Height*config.SubPx {
		t.Fatalf("wobbling %v at %d,%d, not the drop point", k.State, k.X, k.H)
	}
	h := k.H
	for range config.WobbleTicks - 1 {
		if has(w.Update(Input{}), Dropped) {
			t.Fatal("dropped before the wobble was over")
		}
	}
	if es := w.Update(Input{}); !has(es, Dropped) || k.State != Falling || k.H != h {
		t.Errorf("after the wobble: %v, %v at h %d, want Falling at %d", es, k.State, k.H, h)
	}
}

// TestDropTicks checks a drop says how long its fall to the sheet takes,
// for the slide whistle (DESIGN §3): the catch comes exactly then, from
// every branch.
func TestDropTicks(t *testing.T) {
	w := newWorld(t)
	w.Row = config.LevelAt(6) // three branches
	for range 3 {
		var fall int
		run(t, w, func() Input { return Input{} }, 2000, func(es []Event) bool {
			for _, e := range es {
				if e.Kind == Dropped {
					fall = e.Ticks
				}
			}
			return fall > 0
		})
		k := &w.Koalas[len(w.Koalas)-1]
		in := func() Input { return holdFor(k.X, 3) }
		for range fall - 1 {
			if has(w.Update(in()), Bounced) {
				t.Fatalf("caught before the %d ticks the drop said", fall)
			}
		}
		if !has(w.Update(in()), Bounced) {
			t.Fatalf("not caught after the %d ticks the drop said", fall)
		}
		w.Koalas, w.nextDrop = w.Koalas[:0], 1
	}
}

// TestBranches checks drop points join the level table as DESIGN §5 says:
// one branch on levels 1 and 2, a higher one from 3, a lower one from 6,
// and the drops come from all of them.
func TestBranches(t *testing.T) {
	for n, want := range []int{1, 1, 2, 2, 2, 3, 3} {
		if got := len(config.LevelAt(n + 1).Drops); got != want {
			t.Errorf("level %d: %d drop points, want %d", n+1, got, want)
		}
	}
	w := newWorld(t)
	w.startLevel(6)
	seen := map[int]bool{}
	for range 200 {
		k := newKoala(w.Row.Drops[w.rng.IntN(len(w.Row.Drops))])
		seen[k.H] = true
	}
	if len(seen) != 3 {
		t.Errorf("drops came from %d heights, want 3", len(seen))
	}
}

// TestEndless checks the levels past the table repeat the last row, each a
// bounce quicker, down to the fastest (DESIGN §5).
func TestEndless(t *testing.T) {
	last := config.Levels[len(config.Levels)-1]
	n := len(config.Levels)
	if got := config.LevelAt(n + 1); got.BounceTicks != last.BounceTicks-config.SpeedNudge || got.Intro != config.EndlessIntro || got.UteStation != last.UteStation {
		t.Errorf("level %d: %+v", n+1, got)
	}
	if got := config.LevelAt(n + 3).BounceTicks; got != last.BounceTicks-3*config.SpeedNudge {
		t.Errorf("level %d: bounce %d", n+3, got)
	}
	if got := config.LevelAt(1000).BounceTicks; got != config.FastestBounce {
		t.Errorf("level 1000: bounce %d, want %d", got, config.FastestBounce)
	}
}

// TestExtraLives checks a life at every ExtraLife points, up to the cap,
// with the threshold moving on past it (DESIGN §5).
func TestExtraLives(t *testing.T) {
	w := newWorld(t)
	w.Session.Score = config.ExtraLife - 10
	w.score(10)
	if w.Session.Lives != normal.Lives+1 || w.Gained != 1 || !has(w.events, ExtraLife) {
		t.Fatalf("at %d: lives %d gained %d", w.Session.Score, w.Session.Lives, w.Gained)
	}
	w.score(10)
	if w.Session.Lives != normal.Lives+1 {
		t.Errorf("a second life just past the threshold: %d", w.Session.Lives)
	}
	w.Session.Lives = config.MaxLives
	w.score(config.ExtraLife)
	if w.Session.Lives != config.MaxLives || w.Gained != 1 {
		t.Errorf("past the cap: lives %d gained %d", w.Session.Lives, w.Gained)
	}
	w.Session.Lives = 1
	w.score(config.ExtraLife)
	if w.Session.Lives != 2 {
		t.Errorf("after the cap, the next threshold gave lives %d, want 2", w.Session.Lives)
	}
}

// TestSkills checks the skills start with their own lives and ranger speed
// and Easy's rangers are quicker (DESIGN §5).
func TestSkills(t *testing.T) {
	for _, sk := range config.Skills {
		w := New(session.New(sk, sk.Lives), 1, rand.New(rand.NewPCG(1, 2)))
		if w.Session.Lives != sk.Lives {
			t.Errorf("%s: lives %d, want %d", sk.Name, w.Session.Lives, sk.Lives)
		}
		if w.RangerSpeed() < w.speed(config.RangerReach) {
			t.Errorf("%s: rangers slower than the scheduler plans for", sk.Name)
		}
	}
	easy := New(session.New(config.Skills[0], 5), 1, rand.New(rand.NewPCG(1, 2)))
	if easy.RangerSpeed() <= newWorld(t).RangerSpeed() {
		t.Error("Easy's rangers are no quicker than Normal's")
	}
}

// TestJoeysInTheTree checks each level's tree holds as many joey carriers
// as the table says, none before level 5, and that the seed picks which
// (DESIGN §6).
func TestJoeysInTheTree(t *testing.T) {
	patterns := map[string]bool{}
	for seed := range uint64(20) {
		w := New(session.New(normal, normal.Lives), 1, rand.New(rand.NewPCG(seed, 3)))
		for n := 1; n <= 10; n++ {
			w.startLevel(n)
			got := 0
			for _, s := range w.Tree {
				if s.Joey {
					got++
				}
			}
			if got != w.Row.Joeys || n < 5 && got != 0 {
				t.Fatalf("seed %d level %d: %d joeys, want %d", seed, n, got, w.Row.Joeys)
			}
			if n == 5 {
				patterns[fmt.Sprint(w.Tree)] = true
			}
		}
	}
	if config.LevelAt(5).Joeys == 0 {
		t.Error("no joeys on level 5")
	}
	if len(patterns) < 2 {
		t.Error("every seed put the joey on the same koala")
	}
}

// TestJoeyDouble rescues a mother with a joey with a safe last bounce:
// both score double (DESIGN §5).
func TestJoeyDouble(t *testing.T) {
	w := newWorld(t)
	w.Tree = nil
	from := w.Row.UteStation - 1
	w.Koalas = append(w.Koalas, Koala{State: Falling, X: StationX(from), Target: from, h0: 50 * config.SubPx, H: 50 * config.SubPx, Joey: true})
	in := holdFor(StationX(from), 4)
	w.Sheet = in.Target
	run(t, w, func() Input { return in }, 1000, func(es []Event) bool { return has(es, Rescued) })
	if want := 2 * (config.SafePoints + config.RescuePoints); w.Session.Score != want {
		t.Errorf("score %d, want %d", w.Session.Score, want)
	}
}

// TestJoeyTumbles misses a mother with a joey: it tumbles off, dazed, and
// both climb the tree on their own. Once its search is over it climbs to
// her, and she drops with it again (DESIGN §6).
func TestJoeyTumbles(t *testing.T) {
	w := newWorld(t)
	w.Tree = []Waiting{{ID: 4, Spot: 0, Joey: true}}
	w.Row.GapTicks = 1 << 20
	away := Input{Dir: 1}
	run(t, w, func() Input { return away }, 2000, func(es []Event) bool { return has(es, Missed) })
	if len(w.Koalas) != 2 || w.Koalas[0].Joey || !w.Koalas[1].Baby || w.Koalas[1].ID != 4 || w.Koalas[1].State != Dazed {
		t.Fatalf("after the miss: %+v", w.Koalas)
	}
	run(t, w, func() Input { return away }, 3000, func(es []Event) bool { return has(es, JoeyFound) })
	if len(w.Tree) != 1 || !w.Tree[0].Joey || len(w.Koalas) != 0 {
		t.Fatalf("tree %+v koalas %+v after the joey found her", w.Tree, w.Koalas)
	}
	w.nextDrop = 0
	run(t, w, func() Input { return away }, 2000, func(es []Event) bool { return has(es, Leaving) })
	if !w.Koalas[0].Joey {
		t.Error("she dropped again without her joey")
	}
}

// TestLostJoeyWaits gives a joey whose search is over no mother in the
// tree: it waits, and once she is back, climbs to her, holding her there
// until it arrives (DESIGN §6).
func TestLostJoeyWaits(t *testing.T) {
	w := newWorld(t)
	x, h := SpotAt(config.JoeyPerches[0])
	w.Tree = nil
	w.Koalas = []Koala{{State: Searching, X: x, H: h, Timer: 1, ID: 3, Baby: true}}
	for range 200 {
		w.Update(Input{})
	}
	if k := w.Koalas[0]; k.State != Lost || k.X != x || k.H != h {
		t.Fatalf("joey with no mother: %v at %d,%d", k.State, k.X, k.H)
	}
	w.Tree, w.nextDrop = []Waiting{{ID: 3, Spot: 2}}, 0
	run(t, w, func() Input { return Input{} }, 2000, func(es []Event) bool {
		if has(es, JoeyFound) {
			return true
		}
		if has(es, Leaving) {
			t.Fatal("she left before her joey reached her")
		}
		return false
	})
	// She was due, so she leaves as soon as it is on her back.
	if len(w.Tree) != 0 || len(w.Koalas) != 1 || w.Koalas[0].ID != 3 || !w.Koalas[0].Joey {
		t.Errorf("tree %+v koalas %+v", w.Tree, w.Koalas)
	}
}

// TestLostJoeyCarried ends a level with a joey still in the tree: the ending
// takes longer, for a ranger to carry it to the ute, and it is gone when
// the next level starts (DESIGN §6).
func TestLostJoeyCarried(t *testing.T) {
	w := newWorld(t)
	w.Rescued, w.Tree = config.KoalaLevel-1, nil
	from := w.Row.UteStation - 1
	x, h := SpotAt(config.JoeyPerches[1])
	w.Koalas = []Koala{
		{State: Falling, X: StationX(from), Target: from, h0: 50 * config.SubPx, H: 50 * config.SubPx},
		{State: Lost, X: x, H: h, ID: 7, Baby: true, Spot: 1},
	}
	in := holdFor(StationX(from), 4)
	w.Sheet = in.Target
	run(t, w, func() Input { return in }, 1000, func(es []Event) bool { return has(es, LevelDone) })
	if w.Ending != config.EndTicks+config.CarryTicks || w.LostJoeys() != 1 {
		t.Fatalf("ending %d lost joeys %d", w.Ending, w.LostJoeys())
	}
	run(t, w, func() Input { return Input{} }, config.EndTicks+config.CarryTicks+1, func([]Event) bool { return w.Level == 2 })
	if w.LostJoeys() != 0 {
		t.Error("the lost joey is still about on level 2")
	}
}

// TestSpotsKept checks the koalas waiting in the tree stay in their spots
// as others leave (DESIGN §5).
func TestSpotsKept(t *testing.T) {
	w := newWorld(t)
	spotsDistinct(t, w)
	before := append([]Waiting(nil), w.Tree...)
	run(t, w, func() Input { return Input{} }, 10, func(es []Event) bool { return has(es, Leaving) })
	for i, s := range w.Tree {
		if s != before[i+1] {
			t.Fatalf("after the first left, the tree is %+v, was %+v", w.Tree, before)
		}
	}
}

// TestNoJumps plays games with a careless player, who bounces koalas every
// way and misses some, and checks no koala ever jumps from place to place:
// in the tree and on the ground each moves at most a bound's step a tick,
// each sets off from its spot, and each comes to a stop in its spot
// (DESIGN §5).
func TestNoJumps(t *testing.T) {
	// A bound goes twice ClimbSubPx a tick; across and up together, a
	// slanting one goes up to about 2.9 times it.
	const step = 3 * config.ClimbSubPx
	type id struct {
		n    int
		baby bool
	}
	slow := func(s State) bool {
		return s == Walking || s == Wobbling || s == Dazed || s == Crawling || s == Climbing || s == Searching || s == Lost || s == Finding || s == Thumped
	}
	near := func(k Koala, s Waiting) bool {
		x, h := SpotAt(config.Spots[s.Spot])
		return abs(k.X-x)+abs(k.H-h) <= 2*config.ClimbSubPx
	}
	for seed := range uint64(10) {
		w := seeded(t, seed)
		r := rand.New(rand.NewPCG(seed, 5))
		zone, level := 4, 0
		prev, seats := map[id]Koala{}, map[int]Waiting{}
		for tick := 0; tick < 30000 && !w.Over; tick++ {
			var in Input
			if k, ok := soonest(w); ok {
				in = holdFor(StationX(k.Target), zone)
			}
			es := w.Update(in)
			if has(es, Bounced) {
				zone = 1 + r.IntN(5)
			}
			if w.Level != level {
				level, prev, seats = w.Level, map[id]Koala{}, map[int]Waiting{}
			}
			now := map[id]Koala{}
			for _, k := range w.Koalas {
				i := id{k.ID, k.Baby}
				now[i] = k
				p, was := prev[i]
				switch {
				case was && slow(p.State) && slow(k.State):
					if abs(k.X-p.X)+abs(k.H-p.H) > step {
						t.Fatalf("seed %d tick %d: koala %v jumped from %d,%d (%v) to %d,%d (%v)", seed, tick, i, p.X, p.H, p.State, k.X, k.H, k.State)
					}
				case !was && !k.Baby:
					if s, ok := seats[k.ID]; ok && k.State == Walking && !near(k, s) {
						t.Fatalf("seed %d tick %d: koala %d set off from %d,%d, not its spot %+v", seed, tick, k.ID, k.X, k.H, s)
					}
				}
			}
			for _, s := range w.Tree {
				if _, sat := seats[s.ID]; !sat {
					if p, ok := prev[id{s.ID, false}]; ok && !near(p, s) {
						t.Fatalf("seed %d tick %d: koala %d sat in spot %d from %d,%d", seed, tick, s.ID, s.Spot, p.X, p.H)
					}
				}
			}
			spotsDistinct(t, w)
			prev, seats = now, map[int]Waiting{}
			for _, s := range w.Tree {
				seats[s.ID] = s
			}
		}
	}
}

// TestClimbBounds walks a koala from the foot of the tree, on its left
// side, out to the branch: it shimmies round the back of the trunk to the
// right side and climbs it, both in bounds, moving then still, goes out
// along the branch without stopping, and gets there when pathLeft said
// (DESIGN §5).
func TestClimbBounds(t *testing.T) {
	w := newWorld(t)
	k := walker(Waiting{Spot: 7}, w.Row.Drops[0])
	left := k.pathLeft()
	moving, still, behind := 0, 0, false
	for tick := 1; ; tick++ {
		x, h, leg := k.X, k.H, k.leg
		done := k.climb()
		behind = behind || k.Behind()
		switch {
		case k.X != x || k.H != h:
			moving++
		case done:
		case !steep(k.path[leg], k.path[leg+1]) && leg != k.round[0] && leg != k.round[1]:
			t.Fatalf("tick %d: held still at %d,%d, off the trunk", tick, k.X, k.H)
		default:
			still++
		}
		if done {
			if tick != left {
				t.Fatalf("got there after %d ticks, pathLeft said %d", tick, left)
			}
			break
		}
	}
	if moving < left/2 || still < left/5 {
		t.Errorf("moved %d ticks and held still %d of %d: want bounds", moving, still, left)
	}
	if !behind {
		t.Error("never went round the back of the trunk")
	}
}

// TestClimbersPassNoSitters checks that no route between spots passes over
// a koala holding the trunk's left side: climbers keep to the right side
// and go round the back (DESIGN §5).
func TestClimbersPassNoSitters(t *testing.T) {
	const wide, tall = 10 * config.SubPx, 24 * config.SubPx // about a koala holding on
	const trunkTop = 150                                    // where the trunk forks

	for i, a := range config.Spots {
		for j, b := range config.Spots {
			path, round := route(spotWay(a), spotWay(b))
			for l := 0; l+1 < len(path); l++ {
				if l == round[0] || l == round[1] {
					continue
				}
				p, q := path[l], path[l+1]
				for s, sit := range config.Spots {
					if s == i || s == j || config.SideOf(sit.X, sit.Y) != config.Left || sit.Y < trunkTop {
						continue
					}
					x, h := SpotAt(sit)
					for f := 0; f <= 10; f++ {
						px, ph := p.X+(q.X-p.X)*f/10, p.H+(q.H-p.H)*f/10
						if abs(px-x) < wide && abs(ph-h) < tall {
							t.Fatalf("spot %d to %d passes over spot %d at %d,%d", i, j, s, px, ph)
						}
					}
				}
			}
		}
	}
}

// soonest is the koala that comes down to the sheet soonest, if any.
func soonest(w *World) (*Koala, bool) {
	var best *Koala
	first := 0
	for i := range w.Koalas {
		k := &w.Koalas[i]
		station, ticks, ok := w.Landing(k)
		if !ok || station == w.Row.UteStation {
			continue
		}
		if best == nil || ticks < first {
			best, first = k, ticks
		}
	}
	return best, best != nil
}

// TestFairHoldsDrops checks the scheduler won't drop a koala that lands
// when another, far off, is due on the sheet, and will when they land
// together.
func TestFairHoldsDrops(t *testing.T) {
	w := newWorld(t)
	w.Tree = nil
	d := w.Row.Drops[0]
	k := newKoala(d)
	_, ticks, _ := w.Landing(&k)
	for _, c := range []struct {
		station int
		want    bool
	}{{0, true}, {2, false}} {
		other := Koala{State: Arcing, Target: c.station, t: w.Row.BounceTicks - ticks}
		w.Koalas = []Koala{other}
		w.Sheet = SafeSheetX(d.Station)
		if got, _ := w.fair(&k); got != c.want {
			t.Errorf("another koala landing on station %d with it: fair %v, want %v", c.station, got, c.want)
		}
	}
	// Rangers far off hold a drop back, but only for DropWait. The koala
	// waits on the drop point itself, with no walk to give them time.
	was := config.Spots[0]
	t.Cleanup(func() { config.Spots[0] = was })
	config.Spots[0] = config.Spot{X: StationX(d.Station) / config.SubPx, Y: config.SheetY - d.Height}
	w.Koalas = nil
	_, w.Sheet = w.SheetRange()
	if koalas, rangers := w.fair(&k); !koalas || rangers {
		t.Errorf("rangers out of reach: fair %v, %v; want true, false", koalas, rangers)
	}
	w.Tree, w.nextDrop = []Waiting{{}}, 1
	for range config.DropWait {
		if has(w.Update(Input{}), Leaving) {
			t.Fatal("dropped with the rangers out of reach")
		}
	}
	run(t, w, func() Input { return Input{} }, 2*config.DropRetry, func(es []Event) bool { return has(es, Leaving) })
}

// TestSchedulerIsFair plays many seeds through the table and into the
// endless levels with a perfect safe player at Normal: the scheduler must
// never drop a koala it can't catch (DESIGN §5, PLAN §4).
func TestSchedulerIsFair(t *testing.T) {
	const seeds, levels = 100, 12
	for seed := range uint64(seeds) {
		w := New(session.New(normal, normal.Lives), 1, rand.New(rand.NewPCG(seed, 99)))
		for tick := 0; w.Level <= levels; tick++ {
			if tick > 10000*levels {
				t.Fatalf("seed %d: stuck on level %d", seed, w.Level)
			}
			if es := w.Update(w.SafePlay()); has(es, Missed) {
				t.Fatalf("seed %d: the perfect player missed on level %d at tick %d: sheet %d, koalas %+v", seed, w.Level, tick, w.Sheet, w.Koalas)
			}
		}
	}
}
