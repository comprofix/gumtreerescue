package play

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// mob is the kangaroos in the paddock behind the fence (DESIGN §3, D27):
// grown ones grazing, sitting up to look about and now and then hopping a
// few bounds along the paddock, one with a joey in her pouch, and joeys at
// foot keeping close to their mothers. Like the flock it is scenery, moved
// in Update from the scenery's random numbers.
type mob struct {
	roos []kangaroo
	rng  *rand.Rand
}

// rooState is what a kangaroo is doing.
type rooState int

const (
	grazing rooState = iota
	sittingUp
	hopping
)

// kangaroo is one of the mob.
type kangaroo struct {
	x, y  float64 // where its feet are on the grass, in field pixels
	size  float64
	left  bool // facing left
	pouch bool // a joey in its pouch
	mum   int  // for a joey at foot, its mother in the mob; -1 for a grown one
	id    int  // its own time for chewing and flicking its ears

	state rooState
	timer int     // ticks left grazing or sitting up
	graze float64 // how far its head is down, eased
	hops  int     // hops left to go
	dir   float64 // which way it hops, -1 or 1
	t     int     // ticks into this hop
	from  float64 // where this hop began
}

// How the kangaroos behave, in ticks and field pixels.
const (
	grazeMin   = 3 * config.TPS  // the shortest a kangaroo grazes
	grazeMax   = 10 * config.TPS // and the longest
	upMin      = 1 * config.TPS  // the shortest it sits up to look about
	upMax      = 3 * config.TPS  // and the longest
	hopChance  = 0.45            // the chance it hops off after looking about
	hopTicks   = 26              // how long a grown kangaroo's hop takes
	hopLen     = 26              // how far it goes in a hop, about its length
	hopHigh    = 8               // how high it bounds
	headEase   = 0.06            // how fast its head goes down or up, a tick
	paddockLo  = 160             // the stretch of paddock they keep to, clear of the big gum
	paddockHi  = 625
	paddockTop = 249 // the band behind the fence they stand in, the rails across their legs
	paddockLow = 256
	joeyAt     = 10  // how far behind its mother a joey at foot keeps
	joeyNear   = 6   // and how far from there it strays before it hops after her
	joeySize   = 0.5 // a joey at foot beside a grown kangaroo
)

// newMob puts three grown kangaroos about the paddock, one with a joey in
// her pouch, and two joeys at foot with their mothers, placed by rng.
func newMob(rng *rand.Rand) mob {
	m := mob{rng: rng}
	for i := range 3 {
		lo := paddockLo + float64(i)*(paddockHi-paddockLo)/3
		m.roos = append(m.roos, kangaroo{
			x: lo + rng.Float64()*(paddockHi-paddockLo)/3, y: paddockTop + rng.Float64()*(paddockLow-paddockTop),
			size: 0.9 + 0.1*rng.Float64(), left: rng.IntN(2) == 0, pouch: i == 1, mum: -1, id: i,
			timer: rng.IntN(grazeMax), graze: 1,
		})
	}
	for _, mum := range []int{1, 2} {
		r := m.roos[mum]
		m.roos = append(m.roos, kangaroo{
			x: r.x + m.behind(mum), y: r.y + 2, size: joeySize, left: r.left, mum: mum, id: len(m.roos), graze: 1,
		})
	}
	return m
}

// behind is how far across from its mother a joey at foot keeps.
func (m *mob) behind(mum int) float64 {
	if m.roos[mum].left {
		return joeyAt
	}
	return -joeyAt
}

func (m *mob) between(lo, hi int) int { return lo + m.rng.IntN(hi-lo+1) }

// step moves the mob on a tick.
func (m *mob) step() {
	for i := range m.roos {
		r := &m.roos[i]
		if r.state == hopping {
			r.hop()
			continue
		}
		if r.mum >= 0 {
			m.follow(r)
			continue
		}
		r.timer--
		switch r.state {
		case grazing:
			r.graze = min(1, r.graze+headEase)
			if r.timer <= 0 {
				r.state, r.timer = sittingUp, m.between(upMin, upMax)
			}
		case sittingUp:
			r.graze = max(0, r.graze-headEase)
			if r.timer > 0 {
				break
			}
			if m.rng.Float64() < hopChance {
				// Off a few bounds, back into the paddock if near its end.
				dir := float64(1 - 2*m.rng.IntN(2))
				if r.x < paddockLo+40 {
					dir = 1
				} else if r.x > paddockHi-40 {
					dir = -1
				}
				r.start(dir, m.between(2, 6))
			} else {
				r.state, r.timer = grazing, m.between(grazeMin, grazeMax)
			}
		}
	}
}

// follow keeps a joey at foot by its mother: grazing when she grazes,
// looking up when she does, and hopping after her when she goes.
func (m *mob) follow(r *kangaroo) {
	mum := &m.roos[r.mum]
	spot := min(max(mum.x+m.behind(r.mum), paddockLo), paddockHi)
	r.y += (mum.y + 2 - r.y) * 0.05
	if d := spot - r.x; math.Abs(d) > joeyNear {
		dir := math.Copysign(1, d)
		r.start(dir, max(1, int(math.Ceil(math.Abs(d)/r.hopLen()))))
		return
	}
	r.left = mum.left
	if mum.state == grazing {
		r.graze = min(1, r.graze+headEase)
	} else {
		r.graze = max(0, r.graze-headEase)
	}
}

// hopLen and hopTime are how far one hop takes this kangaroo, and how
// long: a joey's are shorter and quicker, so it keeps up.
func (r *kangaroo) hopLen() float64 { return hopLen * (0.5 + 0.5*r.size) }
func (r *kangaroo) hopTime() int    { return int(hopTicks * (0.55 + 0.45*r.size)) }

// start sets the kangaroo hopping n bounds dir.
func (r *kangaroo) start(dir float64, n int) {
	r.state, r.dir, r.hops, r.t, r.from, r.left = hopping, dir, n, 0, r.x, dir < 0
}

// hop moves a hopping kangaroo on a tick, stopping at the paddock's ends
// or when its bounds are done, to sit up and look about.
func (r *kangaroo) hop() {
	r.graze = max(0, r.graze-4*headEase)
	r.t++
	r.x = r.from + r.dir*r.hopLen()*art.HopAlong(float64(r.t)/float64(r.hopTime()))
	if r.t < r.hopTime() {
		return
	}
	r.hops--
	r.t, r.from = 0, r.x
	out := r.x+r.dir*r.hopLen() < paddockLo || r.x+r.dir*r.hopLen() > paddockHi
	if r.hops <= 0 || out && r.mum < 0 {
		r.state, r.timer = sittingUp, upMin
	}
}

// draw draws the mob, the farther first, each grazing, looking about or
// in a bound, chewing as it grazes and flicking an ear now and then.
func (m *mob) draw(p pen, tick int) {
	order := make([]int, len(m.roos))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return int(math.Copysign(1, m.roos[a].y-m.roos[b].y))
	})
	for _, i := range order {
		r := &m.roos[i]
		look := art.Roo{Graze: r.graze, Hop: -1, Chew: float64(tick) * 0.35, Pouch: r.pouch}
		if f := (tick + r.id*113) % (260 + r.id*41); f < 14 {
			look.Flick = math.Sin(math.Pi * float64(f) / 14)
		}
		y := r.y
		if r.state == hopping {
			look.Hop = float64(r.t) / float64(r.hopTime())
			y -= hopHigh * r.size * art.HopLift(look.Hop)
		}
		art.Kangaroo(art.Pen{Dst: p.dst, K: p.k}.At(r.x, y).Sized(r.size).Mirrored(r.left), look)
	}
}
