package art

import "github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"

// cowColours are a breed's colours (DESIGN D35): its body, lit and shaded,
// and its dark parts, the patches and head.
type cowColours struct{ light, shade, dark int }

var (
	holstein = cowColours{resurrect64.White, resurrect64.Mist, resurrect64.Charcoal}
	jersey   = cowColours{resurrect64.Copper, resurrect64.Rust, resurrect64.Bark}
)

// The cows' other colours: a Holstein's pink muzzle, a Jersey's pale ring
// round its dark nose, and the udder.
const (
	cowPink = resurrect64.Pink
	cowRing = resurrect64.Blush
)

// Cow is one of the cows scattered about the paddocks, part of the backdrop
// (DESIGN D35), smaller with distance.
type Cow struct {
	X, Y   float64 // where its feet are, in field pixels
	Size   float64 // how big it is beside a cow of full size, about 19 tall
	Flip   bool    // facing left
	Graze  bool    // its head down to the grass
	Jersey bool    // a plain fawn Jersey, or else a black and white Holstein
	Coat   int     // which of cowCoats a Holstein has
}

// Cows are the cows in the paddocks, Holsteins and Jerseys mixed, the farther
// first, clear of the background gums' trunks and the windmill: some behind
// the big gum, the rest on towards the farm.
var Cows = []Cow{
	{X: 122, Y: 238, Size: 0.5, Flip: true, Graze: true, Jersey: true},
	{X: 208, Y: 238, Size: 0.5, Coat: 2},
	{X: 478, Y: 238, Size: 0.45, Flip: true, Coat: 0},
	{X: 392, Y: 239, Size: 0.5, Graze: true, Jersey: true},
	{X: 266, Y: 240, Size: 0.55, Flip: true, Graze: true, Coat: 0},
	{X: 186, Y: 246, Size: 0.75, Graze: true, Jersey: true},
	{X: 340, Y: 246, Size: 0.75, Flip: true, Jersey: true},
	{X: 100, Y: 247, Size: 0.8, Coat: 2},
	{X: 420, Y: 247, Size: 0.8, Flip: true, Graze: true, Coat: 1},
}

// cowCoats are the black patches a Holstein can have on its white body, each
// x, y, and its radii across and down, about its feet.
var cowCoats = [][][4]float64{
	{{-7, -13, 4, 3}, {4, -10, 4.4, 3.4}},
	{{-1, -13, 6, 2.8}, {-8.5, -8.5, 2.4, 2.4}, {9, -13, 2.6, 2.6}},
	{{-8, -11, 3, 4.4}, {1, -14, 3, 2}, {7.5, -9.5, 3.2, 3}},
}

// DrawCow draws a cow side on and facing right, its feet at the pen's
// origin: a deep body, a Holstein's with its black patches, legs, a tail with
// a dark switch, its udder, and its dark head, up or down to the grass, with
// a white blaze on a Holstein and a pale ringed nose on a Jersey.
func DrawCow(p Pen, c Cow) {
	col := holstein
	if c.Jersey {
		col = jersey
	}
	p.Faded(0.35).Ellipse(0, 0.3, 15, 1.6, 0, grassDeep)
	leg := func(x float64, fill int) {
		p.Stroke(ink, 2.4+2*Line, x, -8, x, -0.4)
		p.Stroke(fill, 2.4, x, -8, x, -0.4)
		p.Stroke(ink, 2.4, x, -1, x, -0.2)
	}
	leg(-6, col.shade)
	leg(9.5, col.shade)
	tail := []float64{-12, -14, -14.4, -11, -14.2, -5}
	p.Stroke(ink, 1+2*Line, tail...)
	p.Stroke(col.light, 1, tail...)
	p.Ellipse(-14.2, -4, 1, 1.8, 0, col.dark)

	hx, hy := 15.5, -16.0 // the head's middle
	if c.Graze {
		hx, hy = 16.5, -3.5
	}
	neck := capsule(9, -13, 4, hx-1, hy, 2.4)
	body := capsule(-7, -11.5, 5.6, 7.5, -12, 5.4)
	p.Stroke(ink, 2*Line, append(neck, neck[:2]...)...)
	p.Stroke(ink, 2*Line, append(body, body[:2]...)...)
	p.Ellipse(-5.6, -5.8, 2.6+Line, 1.8+Line, 0, ink)
	p.Ellipse(-5.6, -5.8, 2.6, 1.8, 0, cowPink)
	p.Poly(col.shade, neck...)
	p.Poly(col.shade, body...)
	p.Poly(col.light, capsule(-7, -12.3, 4.7, 7.5, -12.8, 4.5)...)
	if !c.Jersey {
		for _, b := range cowCoats[c.Coat] {
			p.Ellipse(b[0], b[1], b[2], b[3], 0, col.dark)
		}
	}
	leg(-3.4, col.light)
	leg(6.8, col.light)

	p.Ellipse(hx-2.6, hy-2.4, 1.9+Line, 0.9+Line, -0.4, ink)
	p.Ellipse(hx-2.6, hy-2.4, 1.9, 0.9, -0.4, col.dark)
	skull := capsule(hx-0.6, hy-0.6, 2.8, hx+2.4, hy+1.4, 1.9)
	p.Stroke(ink, 2*Line, append(skull, skull[:2]...)...)
	p.Poly(col.dark, skull...)
	if c.Jersey {
		p.Ellipse(hx+2.8, hy+1.8, 1.8, 1.5, 0.5, cowRing)
		p.Ellipse(hx+3.2, hy+2, 1, 0.9, 0.5, ink)
	} else {
		p.Stroke(resurrect64.White, 1.1, hx-0.8, hy-2.4, hx+1.6, hy+0.4)
		p.Ellipse(hx+2.8, hy+1.8, 1.7, 1.4, 0.5, cowPink)
	}
	p.Circle(hx+0.2, hy+0.2, 0.5, resurrect64.White)
}
