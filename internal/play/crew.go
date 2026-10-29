package play

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// crew is the rangers' walk and the sheet's give when it catches a koala,
// for drawing them (DESIGN §2): Play keeps it in Update from the world, so
// Draw only reads it, and the world never sees it.
type crew struct {
	man, woman walker
	give       int // ticks since the sheet last caught a koala
	zone       int // the zone it caught it in, 1 to 5, or 0 for none yet
}

// walker is one ranger's walk.
type walker struct {
	x      float64 // where it was last tick, in field pixels
	seen   bool    // x is set
	stride float64 // how far its legs swing, eased from 0 standing to 1
	step   float64 // how far through its walk it is, in radians
	left   bool    // facing left
}

// How the rangers walk and the sheet gives.
const (
	strideEase = 0.25 // how fast a stride starts and stops, a tick
	stepPerPx  = 0.45 // radians through a walk for each field pixel gone
	giveTicks  = 30   // how long the sheet gives after a catch
	giveDip    = 3.0  // how far it dips at the catch, in field pixels
)

// step counts a tick of w, with the events it gave.
func (c *crew) step(w *world.World, events []world.Event) {
	c.give++
	for _, e := range events {
		if e.Kind == world.Bounced {
			c.give, c.zone = 0, e.Zone
		}
	}
	left, right := rangerHomes(w)
	c.man.follow(left, true, false)
	if away, ok := rangerAway(w); ok {
		c.woman.follow(away, false, false)
	} else {
		c.woman.follow(right, true, true)
	}
}

// follow moves the walker to x: at home it faces the way it holds the
// sheet, faceLeft; away it faces the way it walks.
func (k *walker) follow(x float64, home, faceLeft bool) {
	dx := x - k.x
	if !k.seen {
		dx, k.seen = 0, true
	}
	k.x = x
	moving := math.Abs(dx) > 0.01
	target := 0.0
	if moving {
		target = 1
	}
	k.stride += (target - k.stride) * strideEase
	switch {
	case home:
		k.left = faceLeft
	case moving:
		k.left = dx < 0
	}
	if k.left {
		dx = -dx
	}
	k.step += dx * stepPerPx
}

// dip is how far the sheet is pushed down where it caught its last koala,
// springing back, and where across it that is, from its middle, in field
// pixels.
func (c *crew) dip() (x, d float64) {
	if c.zone == 0 || c.give >= giveTicks {
		return 0, 0
	}
	zones := []float64{config.ZoneSharp, config.ZoneSafe, config.ZoneCentre, config.ZoneSafe, config.ZoneSharp}
	at := -float64(config.SheetWidth) / 2
	for _, z := range zones[:c.zone-1] {
		at += z
	}
	at += zones[c.zone-1] / 2
	t := float64(c.give)
	return at, giveDip * math.Exp(-t/6) * math.Cos(t*0.6)
}
