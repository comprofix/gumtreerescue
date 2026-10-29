package play

import (
	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// ring is one dazed effect showing: whose it is, and its mini kookaburras
// where they are, in field pixels, and how big.
type ring struct {
	id    int
	baby  bool
	birds [3]struct{ x, y, s float64 }
}

// rings are the dazed effects showing over w's koalas at tick, as drawKoala
// and drawTray draw them: on koalas and joeys dazed on the ground, and on
// koalas dazed in the tray after a headboard bonk.
func rings(w *world.World, tick int) []ring {
	var out []ring
	add := func(k *world.Koala, x, y, size float64, t int) {
		r := ring{id: k.ID, baby: k.Baby}
		for i, b := range art.DizzyBirds(t) {
			r.birds[i].x, r.birds[i].y, r.birds[i].s = x+b.X*size, y+b.Y*size, b.Size*size
		}
		out = append(out, r)
	}
	ute := px(w.UteX()) + uteShift(w)
	n := 0
	for i := range w.Koalas {
		k := &w.Koalas[i]
		switch {
		case k.State == world.Dazed && !(k.Baby && w.Ending > 0):
			size := 1.0
			if k.Baby {
				size = art.JoeySize
			}
			add(k, px(k.X), koalaY(k), size, tick)
		case k.State == world.InTray:
			if k.Bonked && k.Timer > 0 {
				x, y := traySeat(ute, n)
				add(k, x, y, 1, k.Timer)
			}
			n++
		}
	}
	return out
}

// release sends off the mini kookaburras of every dazed effect in was
// that is not in now, to perch about the farm (DESIGN D24): the daze is
// over, and they never just vanish.
func (f *flock) release(was, now []ring) {
	for _, r := range was {
		over := true
		for _, n := range now {
			if n.id == r.id && n.baby == r.baby {
				over = false
			}
		}
		if !over {
			continue
		}
		for _, b := range r.birds {
			f.arrive(art.Kooka, b.x, b.y, b.s)
		}
	}
}
