package play

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// driving is the ute's drive in at a level's start, or off at its end,
// for its dust and smoke (DESIGN §2): Play counts it in Update, so Draw
// only reads it. Every puff is worked out from it, never kept.
type driving struct {
	t   int  // ticks since the drive began
	off bool // driving off, not in
}

// step counts a tick of w, starting the count again as a drive begins.
func (d *driving) step(w *world.World) {
	switch {
	case w.Starting == config.LevelTicks:
		*d = driving{}
	case w.Ending == config.EndTicks:
		*d = driving{off: true}
	default:
		d.t++
	}
}

// shiftAt is how far right of its station the ute was t ticks into the
// drive, in field pixels, as uteShift has it: easing to a stop as it
// drives in, pulling away as it drives off.
func (d driving) shiftAt(t int) float64 {
	var f float64
	if d.off {
		f = min(float64(t)/(config.EndTicks/2), 1)
	} else {
		f = max(float64(config.LevelTicks-t), 0) / config.LevelTicks
	}
	return config.UteDrivePx * f * f
}

// speedAt is how far the ute went in tick t of the drive, in field pixels,
// right for more than 0.
func (d driving) speedAt(t int) float64 { return d.shiftAt(t+1) - d.shiftAt(t) }

// speed is how fast the ute is going now.
func (d driving) speed() float64 { return d.speedAt(d.t) }

// The dust off the ute's tyres and the smoke from its exhaust, in ticks
// and field pixels.
const (
	dustEvery  = 3              // a puff of dust off a tyre, each tyre in turn
	dustLife   = 48             // how long a puff of dust lasts
	cloudEvery = 3              // one puff in this many leaves a low cloud on the track,
	cloudLife  = 120            // which lingers this long
	dustFull   = 2.5            // the ute's speed, a tick, that kicks up the most dust
	dustStill  = 0.2            // and below which it kicks up none
	smokeEvery = 1              // a puff of exhaust smoke while the ute drives
	smokeLife  = 40             // how long a puff of smoke lasts
	smokeStill = 0.05           // the speed below which a driving ute makes no smoke,
	pullAway   = config.TPS / 3 // but for the burst as it pulls away to drive off
)

// drawDust draws the dust the ute kicks up from the track and leaves
// behind it as it drives in or off (DESIGN §2), in front of it.
func drawDust(p pen, w *world.World, d driving) {
	pn := p.art()
	park := px(w.UteX())
	ground := float64(config.SheetY + art.UteWheel)
	tyres := [2]float64{}
	tyres[0], tyres[1] = art.UteTyres(config.TrayBack, config.Headboard)
	for i := max(0, (d.t-cloudLife)/dustEvery); i*dustEvery <= d.t; i++ {
		t0 := i * dustEvery
		v := d.speedAt(t0)
		if math.Abs(v) < dustStill {
			continue
		}
		age := float64(d.t - t0)
		back := -math.Copysign(1, v) // the way dust is thrown, behind the tyre
		strong := min(1, math.Abs(v)/dustFull)
		x0 := park + d.shiftAt(t0) + tyres[i%2]
		if age < dustLife {
			f := age / dustLife
			x := x0 + back*(4+14*(1-math.Exp(-age/10))) + 3*jitter(i, 0)
			y := ground - 2 - 12*(1-math.Exp(-age/16)) + 1.5*jitter(i, 1)
			art.Dust(pn, x, y, 3+8*f+jitter(i, 2), 0.7*strong*(1-f)*(1-f))
		}
		if i%cloudEvery == 0 && age < cloudLife {
			f := age / cloudLife
			x := x0 + back*8 + 6*jitter(i, 3)
			art.Dust(pn, x, ground-5-6*f, 6+12*f, 0.4*strong*(1-f))
		}
	}
}

// drawSmoke draws the smoke from the ute's exhaust as it drives, with a
// burst as it pulls away to drive off (DESIGN §2), behind it.
func drawSmoke(p pen, w *world.World, d driving) {
	pn := p.art()
	park := px(w.UteX())
	ex, ey := art.UteExhaust(config.TrayBack, config.Headboard)
	for i := max(0, (d.t-smokeLife)/smokeEvery); i*smokeEvery <= d.t; i++ {
		t0 := i * smokeEvery
		burst := d.off && t0 < pullAway
		if math.Abs(d.speedAt(t0)) < smokeStill && !burst {
			continue
		}
		age := float64(d.t - t0)
		f := age / smokeLife
		x := park + d.shiftAt(t0) + ex - 2 - 0.3*age + 1.5*jitter(i, 4)
		y := config.SheetY + ey - 0.25*age + jitter(i, 5)
		r, a := 2.6+6*f, 0.3*(1-f)
		if burst {
			r, a = 2*r, 0.4*(1-f)
		}
		art.Smoke(pn, x, y, r, a, burst)
	}
}

// jitter is a fixed pseudo-random number from -1 to 1 for puff i's k-th
// wobble, the same every time it is drawn.
func jitter(i, k int) float64 {
	h := uint32(i)*2654435761 ^ uint32(k)*2246822519
	h ^= h >> 15
	h *= 2654435769
	h ^= h >> 13
	return float64(h%2001)/1000 - 1
}
