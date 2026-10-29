package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// Species is a kind of bird about the farm (DESIGN §3, D27).
type Species int

const (
	Kooka      Species = iota // a kookaburra
	Cockatoo                  // a white cockatoo with its yellow crest
	Magpie                    // a black and white magpie
	NumSpecies                // how many there are
)

// The cockatoo's and the magpie's colours; the kookaburra's are with its
// mini one's.
const (
	cockyWhite = resurrect64.White
	cockyShade = resurrect64.Mist
	cockyCrest = resurrect64.Gold
	cockyBill  = resurrect64.Shadow
	pieBlack   = resurrect64.Charcoal
	pieWhite   = resurrect64.White
	pieBill    = resurrect64.Fog
	pieEye     = resurrect64.Brick
)

// Perch is a place a bird can sit, its feet there, and how big a bird
// sitting there is drawn: smaller in the far gums than on the fence.
type Perch struct {
	X, Y, Size float64
	Fence      bool
}

// Perches are the places birds sit: on top of the background gums' clumps
// of leaves that no other clump covers, and on the fence's posts beyond
// the big gum (DESIGN D27).
func Perches() []Perch {
	var out []Perch
	for _, g := range BackGums {
		cl := gumShapes[g.Shape].clumps
		for i, c := range cl {
			tx, ty := c[0], c[1]-c[2]*0.9
			covered := false
			for j, o := range cl {
				if j != i && math.Hypot(tx-o[0], ty-o[1]) < o[2] {
					covered = true
				}
			}
			if covered {
				continue
			}
			if g.Flip {
				tx = -tx
			}
			out = append(out, Perch{X: g.X + tx*g.Size, Y: g.Y + ty*g.Size, Size: 0.6 + 0.3*g.Size})
		}
	}
	for x := 170.0; x < 640; x += 40 {
		out = append(out, Perch{X: x + 2, Y: FenceTop - 0.6, Size: 1, Fence: true})
	}
	return out
}

// Perched draws a bird of species s sitting with its feet at the pen's
// origin, side on and facing right; bob, from 0 to 1, dips its head as it
// looks about. A bird sitting is about 9 field pixels tall.
func Perched(p Pen, s Species, bob float64) {
	switch s {
	case Cockatoo:
		perchedCocky(p, bob)
	case Magpie:
		perchedMagpie(p, bob)
	default:
		perchedKooka(p, bob)
	}
}

// birdFeet draws a sitting bird's two short legs and claws.
func birdFeet(p Pen) {
	p.Stroke(ink, 0.5, -0.4, -1.4, -0.6, 0)
	p.Stroke(ink, 0.5, 0.8, -1.4, 0.8, 0)
}

// perchedKooka is a kookaburra sitting upright, its barred tail hanging
// behind, its big head with the stripe through its eye and the big bill.
func perchedKooka(p Pen, bob float64) {
	hy := -8 + 0.6*bob
	birdFeet(p)
	p.Stroke(ink, 1.6+2*birdLine, -1.6, -3, -3.4, 1.6)
	p.Stroke(birdTail, 1.6, -1.6, -3, -3.4, 1.6)
	p.Stroke(birdBrown, 0.35, -2.4, -1.6, -1.6, -1.2)
	p.Stroke(birdBrown, 0.35, -2.9, -0.4, -2.1, 0)
	p.Ellipse(0, -4, 2.5+birdLine, 3.2+birdLine, -0.25, ink)
	p.Circle(0.8, hy, 2.3+birdLine, ink)
	p.Ellipse(0, -4, 2.5, 3.2, -0.25, birdCream)
	// The brown wing folded over its side, the blue patch at the shoulder.
	p.Ellipse(-0.8, -4.2, 1.6, 2.8, -0.35, birdBrown)
	p.Ellipse(-0.3, -5.6, 0.8, 0.5, -0.4, birdBlue)
	p.Circle(0.8, hy, 2.3, birdCream)
	p.Ellipse(0.4, hy-1.2, 1.8, 0.9, 0.1, birdCrown)
	p.Stroke(birdBrown, 0.6, -0.8, hy, 2.2, hy+0.1)
	p.Circle(1.4, hy-0.2, 0.42, ink)
	p.Poly(ink, 2.4, hy-0.8, 5.8, hy+0.6, 2.4, hy+1.2)
	p.Poly(birdBill, 2.6, hy+0.4, 5.2, hy+0.6, 2.6, hy+1)
}

// perchedCocky is a white cockatoo sitting, shaded underneath, its yellow
// crest raised, a dark hooked bill and a pale ring round its eye.
func perchedCocky(p Pen, bob float64) {
	hy := -8 + 0.6*bob
	birdFeet(p)
	p.Stroke(ink, 1.6+2*birdLine, -1.4, -3, -2.6, 0.6)
	p.Stroke(cockyShade, 1.6, -1.4, -3, -2.6, 0.6)
	p.Ellipse(0, -4.2, 2.6+birdLine, 3.4+birdLine, -0.2, ink)
	// The crest, a fan of yellow plumes up and back from its crown.
	for i, a := range []float64{-2.4, -2.0, -1.6} {
		s, c := math.Sincos(a)
		l := 3.2 - 0.4*float64(i)
		p.Stroke(ink, 0.9+2*birdLine, 0.4, hy-1.2, 0.4+c*l, hy-1.2+s*l)
	}
	p.Circle(0.8, hy, 2.1+birdLine, ink)
	for i, a := range []float64{-2.4, -2.0, -1.6} {
		s, c := math.Sincos(a)
		l := 3.2 - 0.4*float64(i)
		p.Stroke(cockyCrest, 0.9, 0.4, hy-1.2, 0.4+c*l, hy-1.2+s*l)
	}
	p.Ellipse(0, -4.2, 2.6, 3.4, -0.2, cockyShade)
	p.Ellipse(-0.3, -4.8, 2.1, 2.6, -0.2, cockyWhite)
	p.Curve(cockyShade, 0.6, -1.4, -5.4, -1.2, -3.4, -0.2, -2)
	p.Circle(0.8, hy, 2.1, cockyWhite)
	p.Circle(1.5, hy-0.2, 0.7, resurrect64.Sky)
	p.Circle(1.5, hy-0.2, 0.4, ink)
	p.Poly(ink, 2.4, hy-0.6, 4, hy+0.2, 3.4, hy+1.8, 2.4, hy+1.2)
	p.Poly(cockyBill, 2.6, hy-0.3, 3.7, hy+0.3, 3.2, hy+1.4, 2.6, hy+1)
}

// perchedMagpie is a magpie sitting, black with a white nape, a white
// patch on its folded wing, a pale bill and a dark red eye.
func perchedMagpie(p Pen, bob float64) {
	hy := -7.8 + 0.6*bob
	birdFeet(p)
	p.Stroke(ink, 1.5+2*birdLine, -1.6, -3, -3.4, 0.8)
	p.Stroke(pieBlack, 1.5, -1.6, -3, -3.4, 0.8)
	p.Stroke(pieWhite, 0.6, -1.5, -2.6, -2.2, -1.4)
	p.Ellipse(0, -4, 2.4+birdLine, 3.1+birdLine, -0.25, ink)
	p.Circle(0.8, hy, 1.9+birdLine, ink)
	p.Ellipse(0, -4, 2.4, 3.1, -0.25, pieBlack)
	p.Ellipse(-0.8, -4.6, 1.3, 2.2, -0.35, pieWhite)
	p.Ellipse(-0.7, -3.4, 1, 1.4, -0.35, pieBlack)
	p.Circle(0.8, hy, 1.9, pieBlack)
	p.Ellipse(-0.4, hy+0.4, 1.1, 1.8, 0.3, pieWhite)
	p.Circle(1.4, hy-0.3, 0.42, pieEye)
	p.Poly(ink, 2.3, hy-0.7, 5.2, hy+0.3, 2.3, hy+0.9)
	p.Poly(pieBill, 2.5, hy-0.4, 4.4, hy+0.2, 2.5, hy+0.5)
}

// Flying draws a bird of species s in flight, side on and facing right,
// its middle at the pen's origin, wings up at wing 1 and down at -1.
func Flying(p Pen, s Species, wing float64) {
	switch s {
	case Cockatoo:
		flier(p, wing, cockyWhite, cockyShade, cockyWhite, func(hx, hy float64) {
			for _, a := range []float64{-2.5, -2.1} {
				sn, c := math.Sincos(a)
				p.Stroke(cockyCrest, 0.7, hx-0.4, hy-1, hx-0.4+c*2.4, hy-1+sn*2.4)
			}
			p.Circle(hx+0.5, hy-0.2, 0.35, ink)
			p.Poly(cockyBill, hx+1.2, hy-0.5, hx+2.6, hy+0.2, hx+1.4, hy+1)
		})
	case Magpie:
		flier(p, wing, pieBlack, pieWhite, pieBlack, func(hx, hy float64) {
			p.Ellipse(hx-1, hy+0.2, 0.8, 1.2, 0.4, pieWhite)
			p.Circle(hx+0.5, hy-0.2, 0.35, pieEye)
			p.Poly(pieBill, hx+1.2, hy-0.5, hx+3.4, hy+0.1, hx+1.2, hy+0.4)
		})
	default:
		Kookaburra(p, wing)
	}
}

// flier draws a bird flying: body and head in colour body, its tail, and
// wings in colour wings with tips in tips, and face drawing its face about
// the head's middle.
func flier(p Pen, wing float64, body, wings, tips int, face func(hx, hy float64)) {
	fw := func(up float64, c int) {
		s, co := math.Sincos(0.25 + 1.1*up)
		tx, ty := 0.2-4.4*co, -0.9-4.4*s
		xy := []float64{1.2, -0.9, tx, ty, -1.3, -0.7}
		p.Poly(ink, outlined(xy, birdLine)...)
		p.Poly(c, xy...)
		p.Stroke(tips, 0.7, (1.2+tx)/2, (-0.9+ty)/2, tx, ty)
	}
	fw(-0.6*wing, wings)
	p.Stroke(ink, 1.3+2*birdLine, -2, 0, -4.4, 0.6)
	p.Stroke(body, 1.3, -2, 0, -4.4, 0.6)
	p.Ellipse(0, 0, 2.5+birdLine, 1.6+birdLine, 0, ink)
	p.Circle(2, -1.2, 1.4+birdLine, ink)
	p.Ellipse(0, 0, 2.5, 1.6, 0, body)
	p.Circle(2, -1.2, 1.4, body)
	face(2, -1.2)
	fw(wing, wings)
}
