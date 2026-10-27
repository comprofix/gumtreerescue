package world

import (
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
	w := New(session.New(normal, normal.Lives), 1, rand.New(rand.NewPCG(1, 2)))
	w.Starting, w.nextDrop = 0, 1
	return w
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
		w.Koalas, w.Waiting = append(w.Koalas, k), 0
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

// TestMissCrawlsBack lets a koala fall past the sheet: a life lost, then
// dazed, crawling and climbing back to wait again (DESIGN §5).
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
	waiting := w.Waiting
	run(t, w, func() Input { return away }, 2000, func([]Event) bool { return len(w.Koalas) == 0 })
	if w.Waiting != waiting+1 {
		t.Errorf("waiting %d, want %d: the koala did not climb back", w.Waiting, waiting+1)
	}
}

// TestTreeGrab bounces a koala left from the first station: it grabs the
// tree and climbs back, with no life lost (DESIGN §5).
func TestTreeGrab(t *testing.T) {
	for _, z := range []int{1, 2} {
		w := newWorld(t)
		w.Row.GapTicks = 1 << 20
		in := holdFor(StationX(0), z)
		caught := false
		run(t, w, func() Input {
			if caught {
				return Input{}
			}
			return in
		}, 1000, func(es []Event) bool {
			caught = caught || has(es, Bounced)
			return has(es, Grabbed)
		})
		if w.Koalas[0].State != Climbing || w.Session.Lives != normal.Lives || w.Misses != 0 {
			t.Errorf("zone %d: %v, lives %d, misses %d", z, w.Koalas[0].State, w.Session.Lives, w.Misses)
		}
	}
}

// TestHeadboardBonk sends a koala sharp from the station before the ute: it
// bonks the headboard and drops into the tray, rescued, with nothing for
// that bounce (DESIGN §5).
func TestHeadboardBonk(t *testing.T) {
	w := newWorld(t)
	from := w.Row.UteStation - 1
	w.Koalas = append(w.Koalas, Koala{State: Falling, X: StationX(from), Target: from, h0: 50 * config.SubPx, H: 50 * config.SubPx})
	w.Waiting = 0
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
	w.Rescued, w.Waiting = config.KoalaLevel-1, 0
	from := w.Row.UteStation - 1
	w.Koalas = append(w.Koalas, Koala{State: Falling, X: StationX(from), Target: from, h0: 50 * config.SubPx, H: 50 * config.SubPx})
	in := holdFor(StationX(from), 4)
	w.Sheet = in.Target
	run(t, w, func() Input { return in }, 1000, func(es []Event) bool { return has(es, LevelDone) })
	if want := config.SafePoints + config.RescuePoints + config.CleanBonus; w.Session.Score != want {
		t.Errorf("score %d, want %d with the clean bonus", w.Session.Score, want)
	}
	run(t, w, func() Input { return Input{} }, config.EndTicks+1, func([]Event) bool { return w.Level == 2 })
	if w.Waiting != config.KoalaLevel || w.Rescued != 0 || len(w.Koalas) != 0 || w.Starting == 0 {
		t.Errorf("level 2 starts with waiting %d rescued %d koalas %d starting %d", w.Waiting, w.Rescued, len(w.Koalas), w.Starting)
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
	w.Waiting = 0 // no koalas, so the game can't end
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

// TestWobble checks a koala wobbles on its branch for WobbleTicks before it
// drops (DESIGN §5).
func TestWobble(t *testing.T) {
	w := newWorld(t)
	if es := w.Update(Input{}); !has(es, Wobble) || w.Koalas[0].State != Wobbling {
		t.Fatalf("no wobble: %v", es)
	}
	k := &w.Koalas[0]
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

// TestFairHoldsDrops checks the scheduler won't drop a koala that lands
// when another, far off, is due on the sheet, and will when they land
// together.
func TestFairHoldsDrops(t *testing.T) {
	w := newWorld(t)
	w.Waiting = 0
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
	// Rangers far off hold a drop back, but only for DropWait.
	w.Koalas = nil
	_, w.Sheet = w.SheetRange()
	if koalas, rangers := w.fair(&k); !koalas || rangers {
		t.Errorf("rangers out of reach: fair %v, %v; want true, false", koalas, rangers)
	}
	w.Waiting, w.nextDrop = 1, 1
	for range config.DropWait {
		if has(w.Update(Input{}), Wobble) {
			t.Fatal("dropped with the rangers out of reach")
		}
	}
	run(t, w, func() Input { return Input{} }, 2*config.DropRetry, func(es []Event) bool { return has(es, Wobble) })
}

// perfectBot heads the sheet to the next catch safe play makes, or with
// none, to the first drop point.
func perfectBot(w *World) Input {
	best := catchAt{tick: -1, x: SafeSheetX(w.Row.Drops[0].Station)}
	for i := range w.Koalas {
		for _, c := range w.safeCatches(&w.Koalas[i], nil) {
			if best.tick < 0 || c.tick < best.tick {
				best = c
			}
		}
	}
	return Input{Aim: true, Target: best.x}
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
			if es := w.Update(perfectBot(w)); has(es, Missed) {
				t.Fatalf("seed %d: the perfect player missed on level %d at tick %d: sheet %d, koalas %+v", seed, w.Level, tick, w.Sheet, w.Koalas)
			}
		}
	}
}
