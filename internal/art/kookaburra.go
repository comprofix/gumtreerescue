package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// A kookaburra's colours: a cream head and chest, a brown back and crown
// with a dark stripe through the eye, a patch of blue on the wing, a rusty
// barred tail, and a big bill, dark above and pale below.
const (
	birdCream = resurrect64.Blush
	birdBrown = resurrect64.Bark
	birdCrown = resurrect64.Clay
	birdBlue  = resurrect64.Blue
	birdTail  = resurrect64.Rust
	birdBill  = resurrect64.Sand
)

// Dizzy draws the dazed effect over a dazed koala, its feet at the pen's
// origin (DESIGN D24): three mini kookaburras flying round above its head,
// tick moving them on and flapping their wings. The near ones are drawn a
// little bigger than the far ones.
func Dizzy(p Pen, tick int) {
	const (
		cy, rx, ry = -31.0, 11.0, 2.6 // the ring they fly round
		speed      = 0.07             // radians a tick
		flap       = 0.6              // radians a tick
	)
	for i := range 3 {
		a := float64(tick)*speed + float64(i)*2*math.Pi/3
		s, c := math.Sincos(a)
		near := 0.85 + 0.2*s
		// Round the front they fly right, round the back left.
		bird := p.At(p.OX+rx*c*p.size(), p.OY+(cy+ry*s)*p.size()).Sized(near).Mirrored(s < 0)
		Kookaburra(bird, math.Sin(float64(tick)*flap+float64(i)))
	}
}

// birdLine is a mini kookaburra's outline, finer than a koala's to suit
// its size.
const birdLine = 0.45

// Kookaburra draws a mini kookaburra flying, side on and facing right, its
// middle at the pen's origin, wings up at wing 1 and down at -1. It is
// about 10 field pixels long.
func Kookaburra(p Pen, wing float64) {
	// The far wing, flapping against the near one, then the outline: tail,
	// body and head grown by a line.
	birdWing(p, -0.6*wing, birdCrown)
	p.Stroke(ink, 1.4+2*birdLine, -2, 0, -4.6, 0.8)
	p.Ellipse(0, 0, 2.4+birdLine, 1.8+birdLine, 0, ink)
	p.Circle(1.9, -1.5, 1.5+birdLine, ink)
	// The tail, rusty with dark bars.
	p.Stroke(birdTail, 1.4, -2, 0, -4.6, 0.8)
	p.Stroke(birdBrown, 0.35, -3.2, -0.2, -3.4, 1.1)
	p.Stroke(birdBrown, 0.35, -4, 0, -4.2, 1.3)
	// The body: cream below, brown back.
	p.Ellipse(0, 0, 2.4, 1.8, 0, birdCream)
	p.Ellipse(-0.5, -0.6, 2, 1.1, -0.15, birdBrown)
	// The head: cream, with a brown crown and a dark stripe through the
	// eye.
	p.Circle(1.9, -1.5, 1.5, birdCream)
	p.Ellipse(1.6, -2.4, 1.2, 0.55, 0.1, birdCrown)
	p.Stroke(birdBrown, 0.5, 0.9, -1.6, 2.5, -1.5)
	p.Circle(2.2, -1.65, 0.38, ink)
	// The big bill, dark above and pale below.
	p.Poly(ink, 2.7, -2.4, 5.8, -1.3, 2.7, -0.4)
	p.Poly(birdBill, 2.9, -1.3, 5.3, -1.3, 2.9, -0.6)
	// The near wing.
	birdWing(p, wing, birdBrown)
}

// birdWing draws a kookaburra's wing from its shoulder, raised at up 1 and
// lowered at -1, in colour c with the blue patch near the shoulder.
func birdWing(p Pen, up float64, c int) {
	const sx, sy, long = 0.2, -0.9, 3.6
	s, co := math.Sincos(0.25 + 1.1*up)
	tx, ty := sx-long*co, sy-long*s
	// The leading and trailing edges meet at the tip.
	xy := []float64{sx + 1, sy, tx, ty, sx - 1.3, sy + 0.2}
	p.Poly(ink, outlined(xy, birdLine)...)
	p.Poly(c, xy...)
	p.Ellipse(sx-0.25*long*co, sy-0.25*long*s, 0.6, 0.4, 0.25+1.1*up, birdBlue)
}

// outlined is triangle xy grown by w about its middle, for its outline.
func outlined(xy []float64, w float64) []float64 {
	mx, my := (xy[0]+xy[2]+xy[4])/3, (xy[1]+xy[3]+xy[5])/3
	out := make([]float64, len(xy))
	for i := 0; i < len(xy); i += 2 {
		dx, dy := xy[i]-mx, xy[i+1]-my
		d := math.Hypot(dx, dy)
		out[i], out[i+1] = xy[i]+dx/d*w*2, xy[i+1]+dy/d*w*2
	}
	return out
}
