// Package world is Gum Tree Rescue's rules (DESIGN §5): koalas drop from
// the gum tree, the rangers' sheet bounces them from station to station,
// and the ute's tray takes them in. It has no graphics or audio: the play
// scene draws the state and plays sounds for the events Update returns.
//
// Every position is in fixed point (config.SubPx units to a field pixel)
// and every motion is worked out from the tick it began, never stepped, so
// bounces land exactly on stations and replays match (PLAN §3).
package world

import (
	"math/rand/v2"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/session"
)

// State is what a koala is doing.
type State int

const (
	Wobbling State = iota // on its branch, about to drop
	Falling               // dropped from a branch, or bonked and dropping into the tray
	Arcing                // bounced, on its way to a station
	Dazed                 // missed: sitting dazed on the ground
	Crawling              // crawling back to the tree
	Climbing              // climbing back up the tree to wait again
	InTray                // rescued
)

// Koala is one koala out of the tree.
type Koala struct {
	State State
	// X and H are where it is: across the field, and its height above the
	// sheet (below it, negative), in fixed point.
	X, H int
	// Target is the station it will land on: below it while it wobbles and
	// falls, where an arc ends while it arcs.
	Target int
	// Missed is set once it has passed the sheet uncaught; it falls on to
	// the ground.
	Missed bool
	// Bonked is set once it has hit the ute's headboard (DESIGN §5).
	Bonked bool
	// Timer counts down while it wobbles, is dazed, climbs or is dazed in
	// the tray.
	Timer int

	x0, x1 int // an arc's start and end, across
	h0     int // a fall's starting height
	t      int // ticks into the current fall or arc
}

// EventKind is something that happened during a tick.
type EventKind int

const (
	Wobble     EventKind = iota // a koala started to wobble on its branch
	Dropped                     // a koala dropped from a branch
	Bounced                     // caught on the sheet; Zone says where
	Rescued                     // a koala landed in the tray
	Bonked                      // a koala hit the headboard
	Missed                      // a koala hit the ground: a life lost
	Grabbed                     // a koala bounced into the tree and grabbed it
	ExtraLife                   // the score passed an extra life
	LevelDone                   // the tenth koala of the level is in the tray
	LevelStart                  // the next level's koalas start dropping
	GameOver                    // the last life was lost
)

// Event is something that happened, for sounds and effects.
type Event struct {
	Kind EventKind
	Zone int // for Bounced: the sheet's zone, 1 to 5
}

// Input is what the player is doing this tick: pushing left (-1) or right
// (1), or, with Aim, heading the sheet to Target, across the field in
// fixed point (the mouse).
type Input struct {
	Dir    int
	Aim    bool
	Target int
}

// World is one game of Gum Tree Rescue.
type World struct {
	Session *session.Session

	Level   int // counting from 1
	Row     config.Level
	Sheet   int // the sheet's middle, across the field in fixed point
	Koalas  []Koala
	Waiting int // koalas in the tree waiting to drop
	Rescued int // koalas in the tray this level
	Misses  int // misses this level
	Gained  int // extra lives this level
	Over    bool

	// Starting counts down the intro card before the level's koalas drop;
	// Ending, the drive-off and tally after its last rescue.
	Starting, Ending int

	nextDrop int // ticks until the next drop
	waited   int // how long the due drop has been held back
	nextLife int // the score that gives the next extra life
	rng      *rand.Rand
	events   []Event
	catches  []catchAt // reused by fair
}

// New starts a game on level (1 or more) with the lives of the session's
// skill.
func New(s *session.Session, level int, rng *rand.Rand) *World {
	s.Lives = s.Skill.Lives
	w := &World{Session: s, rng: rng, nextLife: config.ExtraLife}
	w.startLevel(max(level, 1))
	return w
}

// startLevel fills the tree with the level's koalas and puts the rangers
// under the first drop point (DESIGN §5).
func (w *World) startLevel(n int) {
	w.Level, w.Row = n, config.LevelAt(n)
	w.Koalas = w.Koalas[:0]
	w.Waiting, w.Rescued, w.Misses, w.Gained = config.KoalaLevel, 0, 0, 0
	w.Sheet = StationX(w.Row.Drops[0].Station)
	w.Starting, w.Ending = config.LevelTicks, 0
	w.nextDrop, w.waited = config.FirstDrop, 0
}

// StationX is station n's place across the field, in fixed point. Stations
// left of the first are past the trunk.
func StationX(n int) int {
	return (config.Station0X + n*config.StepPx) * config.SubPx
}

// SafeSheetX is where the sheet's middle is for a koala landing on station
// n to land in the middle of zone 4, a safe bounce right.
func SafeSheetX(n int) int {
	middle := config.ZoneSharp + config.ZoneSafe + config.ZoneCentre + config.ZoneSafe/2 - config.SheetWidth/2
	return StationX(n) - middle*config.SubPx
}

// UteX is where the ute is parked: its station, in fixed point.
func (w *World) UteX() int { return StationX(w.Row.UteStation) }

// HeadboardX is the ute's headboard, across the field in fixed point.
func (w *World) HeadboardX() int { return w.UteX() + config.Headboard*config.SubPx }

// SheetRange is how far the rangers can go, as their sheet's middle: from
// the trunk to the ute's tray, neither ranger walking into either (DESIGN
// §5).
func (w *World) SheetRange() (lo, hi int) {
	reach := (config.SheetWidth/2 + config.RangerSide) * config.SubPx
	return config.TrunkX*config.SubPx + reach, w.UteX() - config.TrayBack*config.SubPx - reach
}

// RangerSpeed is how far the rangers move in a tick, in fixed point: the
// skill's Reach hundredths of a step per bounce.
func (w *World) RangerSpeed() int { return w.speed(w.Session.Skill.Reach) }

func (w *World) speed(reach int) int {
	return reach * config.StepPx * config.SubPx / (100 * w.Row.BounceTicks)
}

// Update advances the game one tick and returns what happened. The slice
// is reused by the next call. Nothing changes once the game is over.
func (w *World) Update(in Input) []Event {
	w.events = w.events[:0]
	if w.Over {
		return w.events
	}
	w.moveSheet(in)
	switch {
	case w.Starting > 0:
		if w.Starting--; w.Starting == 0 {
			w.emit(LevelStart, 0)
		}
		return w.events
	case w.Ending > 0:
		if w.Ending--; w.Ending == 0 {
			w.startLevel(w.Level + 1)
		}
		return w.events
	}
	for i := range w.Koalas {
		w.step(&w.Koalas[i])
		if w.Over {
			return w.events
		}
	}
	w.Koalas = removeClimbed(w.Koalas, &w.Waiting)
	w.drop()
	if w.Rescued >= config.KoalaLevel {
		if w.Misses == 0 {
			w.score(config.CleanBonus)
		}
		w.Ending = config.EndTicks
		w.emit(LevelDone, 0)
	}
	return w.events
}

func (w *World) emit(k EventKind, zone int) { w.events = append(w.events, Event{k, zone}) }

// score adds points, and an extra life at every config.ExtraLife points
// up to config.MaxLives (DESIGN §5). Past the cap the threshold still
// moves on.
func (w *World) score(points int) {
	w.Session.AddScore(points)
	for w.Session.Score >= w.nextLife {
		w.nextLife += config.ExtraLife
		if w.Session.Lives < config.MaxLives {
			w.Session.Lives++
			w.Gained++
			w.emit(ExtraLife, 0)
		}
	}
}

// moveSheet moves the rangers by the player's input, no faster than their
// speed, within their range.
func (w *World) moveSheet(in Input) {
	speed := w.RangerSpeed()
	switch {
	case in.Dir != 0:
		w.Sheet += in.Dir * speed
	case in.Aim:
		w.Sheet += min(max(in.Target-w.Sheet, -speed), speed)
	}
	lo, hi := w.SheetRange()
	w.Sheet = min(max(w.Sheet, lo), hi)
}

// drop sends the next koala out to wobble on its branch when it is due and
// the scheduler finds it fair (PLAN §3); otherwise it tries again shortly.
func (w *World) drop() {
	if w.Waiting == 0 {
		return
	}
	if w.nextDrop--; w.nextDrop > 0 {
		return
	}
	d := w.Row.Drops[w.rng.IntN(len(w.Row.Drops))]
	k := newKoala(d)
	koalas, rangers := w.fair(&k)
	if !koalas || !rangers && w.waited < config.DropWait {
		w.nextDrop = config.DropRetry
		if koalas {
			w.waited += config.DropRetry
		}
		return
	}
	w.nextDrop, w.waited = w.Row.GapTicks+w.rng.IntN(config.DropJitter+1), 0
	w.Waiting--
	w.Koalas = append(w.Koalas, k)
	w.emit(Wobble, 0)
}

// newKoala is a koala on drop point d, starting its wobble.
func newKoala(d config.Drop) Koala {
	h := d.Height * config.SubPx
	return Koala{State: Wobbling, X: StationX(d.Station), H: h, h0: h, Target: d.Station, Timer: config.WobbleTicks}
}

// catchAt is a catch safe play will make: in how many ticks, and where the
// sheet's middle has to be, in fixed point.
type catchAt struct{ tick, x int }

// safeCatches adds the catches safe play makes of k from now on: each a
// safe bounce right, station by station until the tray.
func (w *World) safeCatches(k *Koala, out []catchAt) []catchAt {
	station, ticks, ok := w.Landing(k)
	if !ok || station < 0 {
		return out
	}
	for ; station < w.Row.UteStation; station++ {
		out = append(out, catchAt{ticks, SafeSheetX(station)})
		ticks += w.Row.BounceTicks
	}
	return out
}

// fair reports whether rangers at Normal's speed, playing safe, could
// catch k as well as every koala already out (DESIGN §5): each of k's
// catches must be reachable from, and reach, every other planned catch.
// rangers reports whether they can get to k's first catch from where they
// stand now; drop holds a koala back for that only up to config.DropWait,
// so walking away can't stop the game.
func (w *World) fair(k *Koala) (koalas, rangers bool) {
	others := w.catches[:0]
	for i := range w.Koalas {
		others = w.safeCatches(&w.Koalas[i], others)
	}
	w.catches = others
	speed := w.speed(config.RangerReach)
	reach := func(a, b catchAt) bool { return abs(a.x-b.x) <= speed*abs(a.tick-b.tick) }
	mine := w.safeCatches(k, nil)
	for _, a := range mine {
		for _, b := range others {
			if !reach(a, b) {
				return false, false
			}
		}
	}
	return true, len(mine) == 0 || reach(catchAt{0, w.Sheet}, mine[0])
}

func abs(n int) int { return max(n, -n) }

// step moves one koala on a tick.
func (w *World) step(k *Koala) {
	switch k.State {
	case Wobbling:
		if k.Timer--; k.Timer <= 0 {
			k.State, k.t = Falling, 0
			w.emit(Dropped, 0)
		}
	case Falling:
		k.t++
		k.H = k.h0 - w.fallen(k.t)
		w.landFall(k)
	case Arcing:
		k.t++
		w.arcAt(k)
		w.landArc(k)
	case Dazed:
		if k.Timer--; k.Timer <= 0 {
			k.State = Crawling
		}
	case Crawling:
		if k.X -= config.CrawlSubPx; k.X <= config.TrunkX*config.SubPx {
			k.X = config.TrunkX * config.SubPx
			k.State, k.Timer = Climbing, config.ClimbTicks
		}
	case Climbing:
		k.Timer--
	case InTray:
		k.Timer = max(k.Timer-1, 0)
	}
}

// fallen is how far anything falls in t ticks from rest, in fixed point.
// Gravity is set by the level's bounce: a bounce of T ticks rises PeakPx,
// so g = 8·PeakPx/T² and a fall from rest is g·t²/2 (PLAN §3).
func (w *World) fallen(t int) int {
	T := w.Row.BounceTicks
	return 4 * config.PeakPx * config.SubPx * t * t / (T * T)
}

// fallTicks is how many ticks a fall from rest at h0 takes to reach the
// sheet.
func (w *World) fallTicks(h0 int) int {
	t := 0
	for h0-w.fallen(t) > 0 {
		t++
	}
	return t
}

// arcAt places an arcing koala t ticks into its bounce: across in a
// straight line, and up and down on a parabola that peaks at PeakPx and
// comes back to the sheet after exactly BounceTicks.
func (w *World) arcAt(k *Koala) {
	T := w.Row.BounceTicks
	k.X = k.x0 + (k.x1-k.x0)*k.t/T
	k.H = 4 * config.PeakPx * config.SubPx * k.t * (T - k.t) / (T * T)
}

// landFall checks a falling koala against the sheet, the tray and the
// ground.
func (w *World) landFall(k *Koala) {
	switch {
	case k.Missed:
		w.ground(k)
	case k.H > 0:
	case k.Bonked:
		k.H, k.State, k.Timer = 0, InTray, config.BonkedTicks
		w.rescue()
	default:
		k.H = 0
		w.catch(k)
	}
}

// landArc checks an arcing koala against the tree, the headboard, the
// tray, the sheet and the ground.
func (w *World) landArc(k *Koala) {
	switch {
	case k.Missed:
		w.ground(k)
	case k.x1 < k.x0 && k.X <= config.TrunkX*config.SubPx:
		// Bounced back past the first station: it grabs the tree and
		// climbs back up (DESIGN §5).
		k.X, k.State, k.Timer = config.TrunkX*config.SubPx, Climbing, config.ClimbTicks
		w.emit(Grabbed, 0)
	case k.Target > w.Row.UteStation && k.X >= w.HeadboardX()-koalaHalf:
		// Aimed past the tray: it bonks the headboard and drops into the
		// tray (DESIGN §5).
		k.X = w.HeadboardX() - koalaHalf
		k.State, k.Bonked, k.h0, k.t = Falling, true, k.H, 0
		w.emit(Bonked, 0)
	case k.t < w.Row.BounceTicks:
	case k.Target == w.Row.UteStation:
		k.H, k.State = 0, InTray
		w.rescue()
	default:
		w.catch(k)
	}
}

// koalaHalf is where a bonked koala drops, short of the headboard: about
// half a koala's width.
const koalaHalf = 8 * config.SubPx

// catch bounces k off the sheet if the sheet is under it, by the zone it
// lands in; otherwise k has missed and falls on.
func (w *World) catch(k *Koala) {
	zone, ok := w.ZoneAt(k.X)
	if !ok {
		k.Missed = true
		return
	}
	from := k.Target
	to := from + config.ZoneSteps[zone-1]
	k.State, k.Target, k.t = Arcing, to, 0
	k.x0, k.x1 = StationX(from), StationX(to)
	k.X, k.H = k.x0, 0
	w.score(bouncePoints(zone, to > w.Row.UteStation))
	w.emit(Bounced, zone)
}

// bouncePoints is what a bounce from zone scores (DESIGN §5): only
// rightward bounces score, and not a sharp one that will bonk the
// headboard.
func bouncePoints(zone int, bonks bool) int {
	switch {
	case bonks:
		return 0
	case zone == 4:
		return config.SafePoints
	case zone == 5:
		return config.SharpPoints
	}
	return 0
}

// ZoneAt is the sheet's zone, 1 to 5, under x (fixed point), and whether
// the sheet is there at all.
func (w *World) ZoneAt(x int) (zone int, ok bool) {
	d := x - (w.Sheet - config.SheetWidth*config.SubPx/2)
	if d < 0 || d > config.SheetWidth*config.SubPx {
		return 0, false
	}
	edges := []int{config.ZoneSharp, config.ZoneSafe, config.ZoneCentre, config.ZoneSafe}
	at := 0
	for i, e := range edges {
		if at += e * config.SubPx; d < at {
			return i + 1, true
		}
	}
	return 5, true
}

// rescue counts a koala landed in the tray.
func (w *World) rescue() {
	w.Rescued++
	w.score(config.RescuePoints)
	w.emit(Rescued, 0)
}

// ground stops a missed koala on the ground: dazed, and a life lost.
func (w *World) ground(k *Koala) {
	floor := -config.SheetHeight * config.SubPx
	if k.H > floor {
		return
	}
	k.H, k.State, k.Timer = floor, Dazed, config.DazedTicks
	w.Misses++
	w.emit(Missed, 0)
	if w.Session.LoseLife() {
		w.Over = true
		w.emit(GameOver, 0)
	}
}

// removeClimbed takes out the koalas that have climbed back up the tree,
// adding them to those waiting.
func removeClimbed(ks []Koala, waiting *int) []Koala {
	out := ks[:0]
	for _, k := range ks {
		if k.State == Climbing && k.Timer <= 0 {
			*waiting++
			continue
		}
		out = append(out, k)
	}
	return out
}

// Landing is where and in how many ticks k next comes down to the sheet or
// the tray, for the landing marker; ok is false once it has nowhere to
// land (missed, on the ground, in the tray).
func (w *World) Landing(k *Koala) (station, ticks int, ok bool) {
	switch {
	case k.Missed:
		return 0, 0, false
	case k.State == Wobbling:
		return k.Target, k.Timer + w.fallTicks(k.h0), true
	case k.State == Arcing:
		return k.Target, w.Row.BounceTicks - k.t, true
	case k.State == Falling && !k.Bonked:
		t := k.t
		for k.h0-w.fallen(t) > 0 {
			t++
		}
		return k.Target, t - k.t, true
	}
	return 0, 0, false
}
