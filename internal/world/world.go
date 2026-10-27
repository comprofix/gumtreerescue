// Package world is Gum Tree Rescue's rules (DESIGN §5): koalas drop from
// the gum tree, the rangers' sheet bounces them from station to station,
// and the ute's tray takes them in. It has no graphics or audio: the play
// scene draws the state and plays sounds for the events Update returns.
//
// Every position is in fixed point (config.SubPx units to a field pixel)
// and every motion is worked out from the tick it began, never stepped, so
// bounces land exactly on stations and replays match (PLAN §3). Koalas get
// about the tree along paths of straight legs, each taking a whole number
// of ticks, so they never jump from place to place.
package world

import (
	"math"
	"math/rand/v2"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/session"
)

// State is what a koala is doing.
type State int

const (
	Wobbling  State = iota // on its branch, about to drop
	Falling                // dropped from a branch, bonked, or slipped from the tree
	Arcing                 // bounced, on its way to a station
	Dazed                  // sitting dazed on the ground
	Crawling               // crawling back to the tree
	Climbing               // climbing the tree to its spot, or a lost joey to its perch
	InTray                 // rescued
	Walking                // leaving its spot for its drop point
	Searching              // a lost joey looking for its mother
	Lost                   // a lost joey waiting for its mother to come back to the tree
	Finding                // a lost joey climbing to its mother
)

// Koala is one koala out of its spot in the tree, or a lost joey.
type Koala struct {
	State State
	// X and H are where it is: across the field, and its height above the
	// sheet (below it, negative), in fixed point.
	X, H int
	// Target is the station it will land on: below it while it walks,
	// wobbles and falls, where an arc ends while it arcs.
	Target int
	// Missed is set once it has passed the sheet uncaught; it falls on to
	// the ground.
	Missed bool
	// Bonked is set once it has hit the ute's headboard (DESIGN §5).
	Bonked bool
	// Slipped is set once it has slipped from the tree; it falls to the
	// ground with no life lost (DESIGN §5).
	Slipped bool
	// Timer counts down while it wobbles, is dazed, is dazed in the tray,
	// or as a lost joey, searches.
	Timer int
	// Joey is set for a mother with a joey on her back: her bounces and
	// rescue score double (DESIGN §6).
	Joey bool
	// ID tells a level's koalas apart. A lost joey has its mother's.
	ID int
	// Baby is set for a lost joey, off its mother's back (DESIGN §6).
	Baby bool
	// Spot is where it is climbing to: an index into config.Spots, or for a
	// lost joey into config.JoeyPerches.
	Spot int

	x0, x1 int      // an arc's start and end, across
	h0     int      // a fall's starting height
	t      int      // ticks into the current fall, arc or leg of its path
	path   [4]point // where it walks or climbs: from path[0] to path[3]
	leg    int      // the leg of path it is on, from path[leg]
	gone   bool     // back in its spot, or on its mother's back
}

// point is a place on the field, in fixed point.
type point struct{ X, H int }

// Waiting is a koala sitting in its spot in the tree.
type Waiting struct {
	ID   int
	Spot int  // an index into config.Spots
	Joey bool // a joey on her back
}

// EventKind is something that happened during a tick.
type EventKind int

const (
	Leaving    EventKind = iota // a koala left its spot for its drop point
	Wobble                      // a koala started to wobble on its branch
	Dropped                     // a koala dropped from a branch
	Bounced                     // caught on the sheet; Zone says where
	Rescued                     // a koala landed in the tray
	Bonked                      // a koala hit the headboard
	Missed                      // a koala hit the ground: a life lost
	Grabbed                     // a koala bounced into the tree and grabbed it
	Slipped                     // a koala bounced into the tree and slipped
	JoeyLost                    // a joey tumbled off its mother on the ground
	JoeyFound                   // a lost joey climbed back on its mother
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
	Rescued int // koalas in the tray this level
	Misses  int // misses this level
	Gained  int // extra lives this level
	Over    bool

	// Tree is the koalas sitting in their spots, the next to drop first.
	// Koalas that climb back join the end.
	Tree []Waiting

	// Starting counts down the intro card before the level's koalas drop;
	// Ending, a ranger carrying any lost joeys to the ute, the drive-off and
	// the tally after its last rescue.
	Starting, Ending int

	nextDrop int     // ticks until the next koala is due to wobble
	point    int     // the next koala's drop point, in Row.Drops
	retry    int     // ticks until a held-back drop tries again
	waited   int     // how long the due drop has been held back
	nextLife int     // the score that gives the next extra life
	tumbled  []Koala // joeys lost this tick, joining Koalas after it
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

// startLevel sits the level's koalas in their spots, the seeded RNG picking
// which carry joeys, and puts the rangers under the first drop point
// (DESIGN §5, §6).
func (w *World) startLevel(n int) {
	w.Level, w.Row = n, config.LevelAt(n)
	w.Koalas = w.Koalas[:0]
	w.Tree = make([]Waiting, config.KoalaLevel)
	for i := range w.Tree {
		w.Tree[i] = Waiting{ID: i, Spot: i}
	}
	if w.Row.Joeys > 0 {
		for _, i := range w.rng.Perm(config.KoalaLevel)[:w.Row.Joeys] {
			w.Tree[i].Joey = true
		}
	}
	w.Rescued, w.Misses, w.Gained = 0, 0, 0
	w.Sheet = StationX(w.Row.Drops[0].Station)
	w.Starting, w.Ending = config.LevelTicks, 0
	w.nextDrop, w.retry, w.waited = config.FirstDrop, 0, 0
	w.point = w.rng.IntN(len(w.Row.Drops))
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

// SpotAt is where a koala's feet are in spot s, in fixed point.
func SpotAt(s config.Spot) (x, h int) {
	return s.X * config.SubPx, (config.SheetY - s.Y) * config.SubPx
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

// LostJoeys is how many lost joeys are out of their mothers' reach: at a
// level's end, a ranger carries them to the ute (DESIGN §6).
func (w *World) LostJoeys() int {
	n := 0
	for i := range w.Koalas {
		if w.Koalas[i].Baby {
			n++
		}
	}
	return n
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
	w.Koalas = append(removeGone(w.Koalas), w.tumbled...)
	w.tumbled = w.tumbled[:0]
	w.drop()
	if w.Rescued >= config.KoalaLevel {
		if w.Misses == 0 {
			w.score(config.CleanBonus)
		}
		w.Ending = config.EndTicks
		if w.LostJoeys() > 0 {
			w.Ending += config.CarryTicks
		}
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

// drop sends the koala at the front of the tree off to its drop point, in
// time to start its wobble when it is due, if the scheduler finds it fair
// (PLAN §3); otherwise it tries again shortly. It waits while a lost joey
// is climbing to that koala (DESIGN §6).
func (w *World) drop() {
	w.nextDrop = max(w.nextDrop-1, 0)
	w.retry = max(w.retry-1, 0)
	if len(w.Tree) == 0 || w.retry > 0 || w.finding(w.Tree[0].ID) {
		return
	}
	k := walker(w.Tree[0], w.Row.Drops[w.point])
	walk := k.pathLeft()
	if w.nextDrop > walk {
		return
	}
	koalas, rangers := w.fair(&k)
	if !koalas || !rangers && w.waited < config.DropWait {
		w.retry = config.DropRetry
		if koalas {
			w.waited += config.DropRetry
		}
		return
	}
	w.nextDrop, w.waited = walk+w.Row.GapTicks+w.rng.IntN(config.DropJitter+1), 0
	w.point = w.rng.IntN(len(w.Row.Drops))
	w.Tree = w.Tree[1:]
	w.Koalas = append(w.Koalas, k)
	w.emit(Leaving, 0)
}

// finding reports whether a lost joey is climbing to the koala id.
func (w *World) finding(id int) bool {
	for i := range w.Koalas {
		if k := &w.Koalas[i]; k.Baby && k.State == Finding && k.ID == id {
			return true
		}
	}
	return false
}

// newKoala is a koala on drop point d, starting its wobble.
func newKoala(d config.Drop) Koala {
	h := d.Height * config.SubPx
	return Koala{State: Wobbling, X: StationX(d.Station), H: h, h0: h, Target: d.Station, Timer: config.WobbleTicks}
}

// walker is the waiting koala s setting off for drop point d: across to
// the trunk, along it to the branch and out along the branch to its tip
// (DESIGN §5).
func walker(s Waiting, d config.Drop) Koala {
	k := newKoala(d)
	k.ID, k.Joey = s.ID, s.Joey
	tip := point{k.X, k.H}
	x, h := SpotAt(config.Spots[s.Spot])
	k.setPath(point{x, h}, tip.H-config.BranchRise*config.SubPx, tip)
	if k.pathLeft() > 0 {
		k.State = Walking
	} else {
		k.X, k.H = tip.X, tip.H
	}
	return k
}

// setPath starts k on a path from where it is: across to the trunk, up or
// down it to height via, and out to the end. Already at the end, it stays.
func (k *Koala) setPath(from point, via int, end point) {
	trunk := config.ClimbX * config.SubPx
	k.path = [4]point{from, {trunk, from.H}, {trunk, via}, end}
	if from == end {
		k.path = [4]point{end, end, end, end}
	}
	k.leg, k.t = 0, 0
	k.X, k.H = from.X, from.H
}

// climb moves k a tick along its path, and reports whether it has come to
// the end.
func (k *Koala) climb() bool {
	k.t++
	for ; k.leg+1 < len(k.path); k.leg++ {
		a, b := k.path[k.leg], k.path[k.leg+1]
		n := legTicks(a, b)
		if k.t < n {
			k.X, k.H = a.X+(b.X-a.X)*k.t/n, a.H+(b.H-a.H)*k.t/n
			return false
		}
		k.t -= n
	}
	end := k.path[len(k.path)-1]
	k.X, k.H = end.X, end.H
	return true
}

// pathLeft is how many ticks k has left on its path.
func (k *Koala) pathLeft() int {
	left := -k.t
	for i := k.leg; i+1 < len(k.path); i++ {
		left += legTicks(k.path[i], k.path[i+1])
	}
	return max(left, 0)
}

// legTicks is how many ticks a leg from a to b takes at config.ClimbSubPx
// a tick: the fewest that cover it, worked out in whole numbers so every
// machine agrees.
func legTicks(a, b point) int {
	dx, dh := b.X-a.X, b.H-a.H
	d2 := dx*dx + dh*dh
	n := int(math.Sqrt(float64(d2))) / config.ClimbSubPx
	for n*config.ClimbSubPx*n*config.ClimbSubPx < d2 {
		n++
	}
	return n
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
	case Walking:
		if k.climb() {
			k.State, k.Timer = Wobbling, config.WobbleTicks
			w.emit(Wobble, 0)
		}
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
			w.climbBack(k)
		}
	case Climbing:
		if k.climb() {
			w.settle(k)
		}
	case Searching:
		if k.Timer--; k.Timer <= 0 {
			k.State = Lost
			w.seek(k)
		}
	case Lost:
		w.seek(k)
	case Finding:
		if k.climb() {
			w.reunite(k)
		}
	case InTray:
		k.Timer = max(k.Timer-1, 0)
	}
}

// climbBack starts k up the tree from where it is: a koala to a free spot,
// a lost joey to a free perch, the seeded RNG picking which (DESIGN §5,
// §6).
func (w *World) climbBack(k *Koala) {
	k.Spot = w.freeSpot(k.Baby)
	spot := config.Spots[:]
	if k.Baby {
		spot = config.JoeyPerches[:]
	}
	x, h := SpotAt(spot[k.Spot])
	k.State = Climbing
	k.setPath(point{k.X, k.H}, h, point{x, h})
}

// freeSpot picks a spot no koala sits in or is climbing to, or for a lost
// joey, a perch no other joey has.
func (w *World) freeSpot(baby bool) int {
	n := len(config.Spots)
	if baby {
		n = len(config.JoeyPerches)
	}
	used := make([]bool, n)
	if !baby {
		for _, s := range w.Tree {
			used[s.Spot] = true
		}
	}
	for i := range w.Koalas {
		k := &w.Koalas[i]
		if k.Baby == baby && (k.State == Climbing || baby && (k.State == Searching || k.State == Lost)) {
			used[k.Spot] = true
		}
	}
	var free []int
	for i, u := range used {
		if !u {
			free = append(free, i)
		}
	}
	if len(free) == 0 {
		return 0
	}
	return free[w.rng.IntN(len(free))]
}

// settle ends a climb: a koala sits in its spot and joins the back of the
// queue; a lost joey starts looking for its mother, for a time the seeded
// RNG picks (DESIGN §6).
func (w *World) settle(k *Koala) {
	if k.Baby {
		k.State, k.Timer = Searching, config.JoeySearch+w.rng.IntN(config.JoeySearchUp+1)
		return
	}
	w.Tree = append(w.Tree, Waiting{ID: k.ID, Spot: k.Spot, Joey: k.Joey})
	k.gone = true
}

// seek sends a lost joey that has finished searching to its mother, if she
// is waiting in the tree; otherwise it waits for her (DESIGN §6).
func (w *World) seek(k *Koala) {
	for _, s := range w.Tree {
		if s.ID == k.ID {
			x, h := SpotAt(config.Spots[s.Spot])
			k.State = Finding
			k.setPath(point{k.X, k.H}, h, point{x, h})
			return
		}
	}
}

// reunite puts a lost joey back on its mother's back. drop holds her in
// her spot until it gets there, but if she has gone it climbs to a perch
// and looks again.
func (w *World) reunite(k *Koala) {
	for i := range w.Tree {
		if w.Tree[i].ID == k.ID {
			w.Tree[i].Joey, k.gone = true, true
			w.emit(JoeyFound, 0)
			return
		}
	}
	w.climbBack(k)
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
	case k.Missed || k.Slipped:
		w.ground(k)
	case k.H > 0:
	case k.Bonked:
		k.H, k.State, k.Timer = 0, InTray, config.BonkedTicks
		w.rescue(k)
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
		// Bounced back past the first station into the tree: the seeded
		// RNG decides whether it grabs on where it hit and climbs to a
		// spot, or slips and falls to the ground (DESIGN §5).
		k.X = config.TrunkX * config.SubPx
		if w.rng.IntN(100) < config.SlipChance {
			k.State, k.Slipped, k.h0, k.t = Falling, true, k.H, 0
			w.emit(Slipped, 0)
			return
		}
		w.emit(Grabbed, 0)
		w.climbBack(k)
	case k.Target > w.Row.UteStation && k.X >= w.HeadboardX()-koalaHalf:
		// Aimed past the tray: it bonks the headboard and drops into the
		// tray (DESIGN §5).
		k.X = w.HeadboardX() - koalaHalf
		k.State, k.Bonked, k.h0, k.t = Falling, true, k.H, 0
		w.emit(Bonked, 0)
	case k.t < w.Row.BounceTicks:
	case k.Target == w.Row.UteStation:
		k.H, k.State = 0, InTray
		w.rescue(k)
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
	w.score(worth(k, bouncePoints(zone, to > w.Row.UteStation)))
	w.emit(Bounced, zone)
}

// worth is what points are worth for k: double with a joey (DESIGN §5).
func worth(k *Koala, points int) int {
	if k.Joey {
		return 2 * points
	}
	return points
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

// rescue counts k landed in the tray.
func (w *World) rescue(k *Koala) {
	w.Rescued++
	w.score(worth(k, config.RescuePoints))
	w.emit(Rescued, 0)
}

// ground stops a koala on the ground, dazed. A joey on its back tumbles
// off, dazed too (DESIGN §6). A koala missed on the sheet costs a life;
// one that slipped from the tree does not (DESIGN §5).
func (w *World) ground(k *Koala) {
	floor := -config.SheetHeight * config.SubPx
	if k.H > floor {
		return
	}
	k.H, k.State, k.Timer = floor, Dazed, config.DazedTicks
	if k.Joey {
		k.Joey = false
		w.tumbled = append(w.tumbled, Koala{
			State: Dazed, X: k.X + config.JoeyTumblePx*config.SubPx, H: floor,
			Timer: config.DazedTicks, ID: k.ID, Baby: true,
		})
		w.emit(JoeyLost, 0)
	}
	if k.Slipped {
		return
	}
	w.Misses++
	w.emit(Missed, 0)
	if w.Session.LoseLife() {
		w.Over = true
		w.emit(GameOver, 0)
	}
}

// removeGone takes out the koalas back in their spots and the joeys back
// on their mothers.
func removeGone(ks []Koala) []Koala {
	out := ks[:0]
	for _, k := range ks {
		if !k.gone {
			out = append(out, k)
		}
	}
	return out
}

// Landing is where and in how many ticks k next comes down to the sheet or
// the tray, for the landing marker; ok is false once it has nowhere to
// land (missed, slipped, on the ground, in the tree, in the tray).
func (w *World) Landing(k *Koala) (station, ticks int, ok bool) {
	switch {
	case k.Missed || k.Slipped || k.Baby:
		return 0, 0, false
	case k.State == Walking:
		return k.Target, k.pathLeft() + config.WobbleTicks + w.fallTicks(k.h0), true
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
