package play

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// flock is the birds about the farm (DESIGN §3, D27): kookaburras,
// cockatoos and magpies perched in the background gums and on the fence,
// now and then flying to another perch or off the screen, and back later.
// It is scenery: Play moves it in Update from the scenery's own random
// numbers, never the world's, so play is the same with it or without.
type flock struct {
	perches []art.Perch
	taken   []bool // a bird sits on the perch or is flying to it
	birds   []bird
	rng     *rand.Rand
}

// birdState is what a bird is doing.
type birdState int

const (
	perched birdState = iota
	flying
	away // off the screen
)

// bird is one bird of the flock.
type bird struct {
	kind  art.Species
	state birdState
	at    int  // its perch, or the one it flies to; -1 flying off
	timer int  // ticks left perched, or away
	sat   int  // ticks it has sat on its perch
	gone  bool // flying off for good, to make room
	left  bool // facing left
	id    int  // its own time for looking about

	// Its flight, from (x0, y0) at size s0 to (x1, y1) at s1, t ticks of
	// dur through, rising lift above the line between them.
	x0, y0, s0, x1, y1, s1 float64
	t, dur                 int
	lift                   float64
}

// How the birds behave, in ticks and field pixels.
const (
	birdCount  = 8
	perchMin   = 4 * config.TPS  // the shortest a bird sits
	perchMax   = 16 * config.TPS // and the longest
	awayMin    = 6 * config.TPS  // the shortest it stays off the screen
	awayMax    = 20 * config.TPS // and the longest
	offChance  = 0.3             // the chance a bird leaving its perch flies off the screen
	turnChance = 1.0 / 400       // the chance a sitting bird turns about, a tick
	flySpeed   = 1.3             // how far a bird flies a tick
	flyMin     = 30              // the shortest flight
	birdMiddle = 4               // how far above its feet a bird's middle is, at size 1
	offScreen  = 14              // how far past the field's edge a bird leaves or comes in
	skyLow     = 60              // the band of sky birds leave and come in by
	skyHigh    = 170
)

// newFlock sits the birds on perches picked by rng, and keeps rng for
// what they do next.
func newFlock(rng *rand.Rand) flock {
	f := flock{perches: art.Perches(), rng: rng}
	f.taken = make([]bool, len(f.perches))
	for i := range min(birdCount, len(f.perches)) {
		b := bird{kind: art.Species(rng.IntN(int(art.NumSpecies))), id: i, left: rng.IntN(2) == 0}
		b.at = f.free(-1)
		f.taken[b.at] = true
		// Staggered, so they do not all go at once.
		b.timer = rng.IntN(perchMax)
		f.birds = append(f.birds, b)
	}
	return f
}

// free is a free perch other than not, picked at random, or -1 if there
// is none.
func (f *flock) free(not int) int {
	var free []int
	for i, t := range f.taken {
		if !t && i != not {
			free = append(free, i)
		}
	}
	if len(free) == 0 {
		return -1
	}
	return free[f.rng.IntN(len(free))]
}

func (f *flock) between(lo, hi int) int { return lo + f.rng.IntN(hi-lo+1) }

// step moves the flock on a tick.
func (f *flock) step() {
	for i := range f.birds {
		b := &f.birds[i]
		switch b.state {
		case perched:
			b.sat++
			if f.rng.Float64() < turnChance {
				b.left = !b.left
			}
			if b.timer--; b.timer <= 0 {
				f.leave(b)
			}
		case flying:
			if b.t++; b.t < b.dur {
				break
			}
			if b.at >= 0 {
				b.state, b.timer, b.sat = perched, f.between(perchMin, perchMax), 0
			} else {
				b.state, b.timer = away, f.between(awayMin, awayMax)
			}
		case away:
			if b.timer--; b.timer > 0 {
				break
			}
			fw, _ := config.Field()
			x := -offScreen + float64(f.rng.IntN(2))*(float64(fw)+2*offScreen)
			b.x1, b.y1, b.s1 = x, float64(f.between(skyLow, skyHigh)), 1
			f.flyTo(b, f.room())
		}
	}
	// Birds that flew off to make room are gone once off the screen.
	f.birds = slices.DeleteFunc(f.birds, func(b bird) bool { return b.gone && b.state == away })
}

// arrive brings a new bird of kind into the flock, flying in from (x, y)
// at size s to a perch, making room if the perches are full.
func (f *flock) arrive(kind art.Species, x, y, s float64) {
	f.birds = append(f.birds, bird{kind: kind, id: len(f.birds), x1: x, y1: y, s1: s})
	b := &f.birds[len(f.birds)-1]
	f.flyTo(b, f.room())
}

// room is a free perch for a bird coming in: if the perches are all full,
// the bird that has sat longest flies off the screen to make room, gone
// for good if the flock is bigger than it starts (DESIGN D27).
func (f *flock) room() int {
	if to := f.free(-1); to >= 0 {
		return to
	}
	var oldest *bird
	for i := range f.birds {
		if b := &f.birds[i]; b.state == perched && (oldest == nil || b.sat > oldest.sat) {
			oldest = b
		}
	}
	if oldest == nil { // every perch is a bird's to land on: fly on by
		return -1
	}
	to := oldest.at
	p := f.perches[to]
	oldest.x1, oldest.y1, oldest.s1 = p.X, p.Y-birdMiddle*p.Size, p.Size
	oldest.gone = len(f.birds) > birdCount
	f.flyTo(oldest, -1)
	return to
}

// leave takes a bird off its perch: to another free one, or off the
// screen by the nearer side.
func (f *flock) leave(b *bird) {
	from := b.at
	f.taken[from] = false
	p := f.perches[from]
	b.x1, b.y1, b.s1 = p.X, p.Y-birdMiddle*p.Size, p.Size
	to := -1
	if f.rng.Float64() >= offChance {
		to = f.free(from)
	}
	f.flyTo(b, to)
}

// flyTo starts b flying from where it is, (x1, y1) at s1, to perch to, or
// off the screen if to is -1.
func (f *flock) flyTo(b *bird, to int) {
	b.x0, b.y0, b.s0 = b.x1, b.y1, b.s1
	if to >= 0 {
		p := f.perches[to]
		f.taken[to] = true
		b.x1, b.y1, b.s1 = p.X, p.Y-birdMiddle*p.Size, p.Size
	} else {
		fw, _ := config.Field()
		b.x1 = -offScreen
		if b.x0 > float64(fw)/2 {
			b.x1 = float64(fw) + offScreen
		}
		b.y1, b.s1 = float64(f.between(skyLow, skyHigh)), 1
	}
	d := math.Hypot(b.x1-b.x0, b.y1-b.y0)
	b.state, b.at, b.t = flying, to, 0
	b.dur = max(flyMin, int(d/flySpeed))
	b.lift = min(30, d*0.25) + 6
	b.left = b.x1 < b.x0
}

// where is where a flying bird's middle is, and its size.
func (b *bird) where() (x, y, s float64) {
	f := float64(b.t) / float64(b.dur)
	e := f * f * (3 - 2*f) // easing off and in to land
	return lerp(b.x0, b.x1, e), lerp(b.y0, b.y1, e) - b.lift*math.Sin(math.Pi*f), lerp(b.s0, b.s1, e)
}

// draw draws the birds on the screen: sitting, now and then dipping their
// heads to look about, or flying, flapping as they go.
func (f *flock) draw(p pen, tick int) {
	for i := range f.birds {
		b := &f.birds[i]
		a := art.Pen{Dst: p.dst, K: p.k}
		switch b.state {
		case perched:
			at := f.perches[b.at]
			bob := 0.0
			if t := (tick + b.id*71) % (150 + b.id*23); t < 24 {
				bob = math.Sin(math.Pi * float64(t) / 24)
			}
			art.Perched(a.At(at.X, at.Y).Sized(at.Size).Mirrored(b.left), b.kind, bob)
		case flying:
			x, y, s := b.where()
			art.Flying(a.At(x, y).Sized(s).Mirrored(b.left), b.kind, math.Sin(float64(b.t)*0.55+float64(b.id)))
		}
	}
}
