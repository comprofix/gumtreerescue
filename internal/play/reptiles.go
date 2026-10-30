package play

import (
	"math"
	"math/rand/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// reptiles are the lace monitor and the carpet python on the dirt track
// (DESIGN §3, D32): one of each, now and then, wandering back and forth in
// front of the rangers and the ute, the python curling up to sunbake, and
// both leaving the screen to come back later. Like the flock they are
// scenery, moved in Update from the scenery's random numbers.
type reptiles struct {
	monitor goanna
	python  python
	rng     *rand.Rand
}

// What a reptile is doing.
type reptileState int

const (
	offTrack  reptileState = iota // off the screen
	wandering                     // walking or slithering along the track
	resting                       // the monitor stopped to taste the air; the python coiled up, sunbaking
)

// goanna is the lace monitor.
type goanna struct {
	state   reptileState
	x, y    float64 // its feet on the track, in field pixels
	left    bool    // heading left
	leaving bool    // walking off the screen
	timer   int     // ticks left of what it is doing
	step    float64 // radians through its walk
	walk    float64 // how far its legs swing, eased
	t       int     // ticks it has been resting, for its tongue
}

// python is the carpet python.
type python struct {
	state   reptileState
	x, y    float64 // its middle on the track, in field pixels
	left    bool    // heading left
	leaving bool    // slithering off the screen
	timer   int     // ticks left of what it is doing
	along   float64 // how far it has slid, for the curves its body lies in
	coil    float64 // how far it is curled up, eased
	t       int     // ticks it has been resting, for its tongue
}

// How the reptiles behave, in ticks and field pixels.
const (
	trackTop     = 344 // the band of track they keep to, clear of the ute's wheels
	trackLow     = 353
	trackEdge    = 40  // how near the screen's edge they turn back, unless leaving
	monitorSpeed = 0.3 // how far the monitor walks a tick
	monitorEase  = 0.08
	monitorAway  = 25 * config.TPS // the shortest it stays off the screen
	monitorBack  = 70 * config.TPS // and the longest
	monitorGoMin = 3 * config.TPS  // the shortest it walks before stopping
	monitorGoMax = 9 * config.TPS  // and the longest
	monitorStop  = 2 * config.TPS  // the shortest it stops to taste the air
	monitorStay  = 6 * config.TPS  // and the longest
	monitorFlick = 50              // how often its tongue flicks while it stops
	pythonSpeed  = 0.18            // how far the python slides a tick
	pythonEase   = 0.012           // how fast it curls up or uncurls, a tick
	pythonAway   = 35 * config.TPS // the shortest it stays off the screen
	pythonBack   = 100 * config.TPS
	pythonGoMin  = 6 * config.TPS  // the shortest it slithers before deciding again
	pythonGoMax  = 16 * config.TPS // and the longest
	pythonBask   = 15 * config.TPS // the shortest it sunbakes
	pythonBake   = 40 * config.TPS // and the longest
	pythonFlick  = 90              // how often its tongue flicks while it sunbakes
	leaveChance  = 0.25            // the chance a reptile leaves when it decides again
	turnBack     = 0.4             // the chance it turns back the way it came
	baskChance   = 0.5             // the chance the python curls up to sunbake
	firstAway    = 40 * config.TPS // the longest before each first comes along
	flickTicks   = 16              // how long a tongue flick takes
)

// newReptiles starts the monitor and the python off the screen, each to
// come along within firstAway, placed by rng, and keeps rng for what they
// do next.
func newReptiles(rng *rand.Rand) reptiles {
	return reptiles{
		monitor: goanna{state: offTrack, timer: rng.IntN(firstAway)},
		python:  python{state: offTrack, timer: rng.IntN(firstAway)},
		rng:     rng,
	}
}

func (r *reptiles) between(lo, hi int) int { return lo + r.rng.IntN(hi-lo+1) }

// enter brings a reptile on from a side picked at random, reach past the
// screen's edge, on a line of the track, and returns where and which way.
func (r *reptiles) enter(reach float64) (x, y float64, left bool) {
	fw, _ := config.Field()
	left = r.rng.IntN(2) == 0
	x = -reach
	if left {
		x = float64(fw) + reach
	}
	return x, trackTop + r.rng.Float64()*(trackLow-trackTop), left
}

// step moves the reptiles on a tick.
func (r *reptiles) step() {
	r.stepMonitor()
	r.stepPython()
}

func (r *reptiles) stepMonitor() {
	m := &r.monitor
	fw, _ := config.Field()
	m.timer--
	switch m.state {
	case offTrack:
		if m.timer <= 0 {
			m.x, m.y, m.left = r.enter(art.MonitorAhead)
			m.state, m.leaving, m.timer = wandering, false, r.between(monitorGoMin, monitorGoMax)
		}
	case wandering:
		m.walk = min(1, m.walk+monitorEase)
		m.x += dirOf(m.left) * monitorSpeed
		m.step += 2 * math.Pi * monitorSpeed / art.MonitorStride
		if m.leaving && pastEdge(m.x, m.left, art.MonitorBehind) {
			m.state, m.timer = offTrack, r.between(monitorAway, monitorBack)
			return
		}
		if m.timer <= 0 || !m.leaving && nearEdge(m.x, m.left, float64(fw)) {
			m.state, m.timer, m.t = resting, r.between(monitorStop, monitorStay), 0
		}
	case resting:
		m.walk = max(0, m.walk-monitorEase)
		m.t++
		if m.timer > 0 {
			break
		}
		// Off again: away, back the way it came, or on.
		m.leaving = r.rng.Float64() < leaveChance
		if nearEdge(m.x, m.left, float64(fw)) && !m.leaving || r.rng.Float64() < turnBack {
			m.left = !m.left
		}
		m.state, m.timer = wandering, r.between(monitorGoMin, monitorGoMax)
	}
}

func (r *reptiles) stepPython() {
	s := &r.python
	fw, _ := config.Field()
	s.timer--
	switch s.state {
	case offTrack:
		if s.timer <= 0 {
			s.x, s.y, s.left = r.enter(art.PythonLength / 2)
			s.state, s.leaving, s.timer = wandering, false, r.between(pythonGoMin, pythonGoMax)
		}
	case wandering:
		// It sets off only once uncurled.
		if s.coil > 0 {
			s.coil = max(0, s.coil-pythonEase)
			break
		}
		s.x += dirOf(s.left) * pythonSpeed
		s.along += pythonSpeed
		if s.leaving && pastEdge(s.x, s.left, art.PythonLength/2+4) {
			s.state, s.timer = offTrack, r.between(pythonAway, pythonBack)
			return
		}
		if s.timer > 0 && (s.leaving || !nearEdge(s.x, s.left, float64(fw))) {
			break
		}
		// It decides again: away, sunbaking, or on. It only turns about
		// curled up, so it is never seen to flip.
		s.leaving = r.rng.Float64() < leaveChance
		s.timer = r.between(pythonGoMin, pythonGoMax)
		if !s.leaving && (r.rng.Float64() < baskChance || nearEdge(s.x, s.left, float64(fw))) {
			s.state, s.timer, s.t = resting, r.between(pythonBask, pythonBake), 0
		}
	case resting:
		s.coil = min(1, s.coil+pythonEase)
		s.t++
		if s.timer > 0 {
			break
		}
		if nearEdge(s.x, s.left, float64(fw)) || r.rng.Float64() < turnBack {
			s.left = !s.left
		}
		s.state, s.timer = wandering, r.between(pythonGoMin, pythonGoMax)
	}
}

// dirOf is -1 heading left, else 1.
func dirOf(left bool) float64 {
	if left {
		return -1
	}
	return 1
}

// pastEdge reports whether a reptile at x heading left or right, reaching
// back from x behind it, has gone wholly off the screen that way.
func pastEdge(x float64, left bool, behind float64) bool {
	fw, _ := config.Field()
	if left {
		return x+behind < 0
	}
	return x-behind > float64(fw)
}

// nearEdge reports whether a reptile at x heading left or right is near
// the screen's edge it is heading for.
func nearEdge(x float64, left bool, fw float64) bool {
	if left {
		return x < trackEdge
	}
	return x > fw-trackEdge
}

// flick is how far a tongue is out t ticks into a rest, flicking every
// every ticks.
func flick(t, every int) float64 {
	if f := t % every; f < flickTicks {
		return math.Sin(math.Pi * float64(f) / flickTicks)
	}
	return 0
}

// draw draws the reptiles on the track, the farther first.
func (r *reptiles) draw(p pen) {
	m, s := &r.monitor, &r.python
	monitor := func() {
		if m.state == offTrack {
			return
		}
		g := art.Goanna{Step: m.step, Walk: m.walk}
		if m.state == resting {
			g.Tongue = flick(m.t, monitorFlick)
		}
		art.LaceMonitor(p.art().At(m.x, m.y).Mirrored(m.left), g)
	}
	snake := func() {
		if s.state == offTrack {
			return
		}
		look := art.Snake{Wave: art.PythonWave(s.along), Coil: s.coil}
		if s.state == resting && s.coil == 1 {
			look.Tongue = flick(s.t, pythonFlick)
		}
		art.CarpetPython(p.art().At(s.x, s.y).Mirrored(s.left), look)
	}
	if m.y > s.y {
		snake()
		monitor()
		return
	}
	monitor()
	snake()
}
