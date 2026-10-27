package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The koala's colours (DESIGN §3): grey fur, shaded lilac and lit pale,
// with a cream chest and ear tufts, a dark nose and eyes.
const (
	ink       = resurrect64.Ink
	fur       = resurrect64.Fog
	furShade  = resurrect64.Lilac
	furLight  = resurrect64.Mist
	fluff     = resurrect64.White
	noseLight = resurrect64.Dusk
)

// JoeySize is a joey's size beside its mother's.
const JoeySize = 0.55

// Look is how a koala looks this frame, whatever it is doing.
type Look struct {
	Joey  bool    // a joey riding on its back
	Shut  bool    // its eyes closed: a blink, asleep or dazed
	Flick float64 // one ear flicked, from 0 to 1
	Lean  float64 // how far a sitting koala sways, in radians
}

// legs is how a sitting koala's legs are.
type legs int

const (
	feet    legs = iota // feet out in front, sitting on something flat
	hanging             // legs hanging over a limb, swinging
	tucked              // tucked up, curled asleep
)

// arm is what a sitting koala's arms are doing.
type arm int

const (
	resting arm = iota // both paws on its knees
	chewing            // the near paw holding a gum leaf to its mouth
	holding            // the far paw on its sore head
)

// Koala draws a koala sitting with its feet at the pen's origin, facing
// out. It is about 21 field pixels wide and 27 tall.
func Koala(p Pen, l Look) { sit(p, l, feet, resting, 0) }

// Perched draws a koala sitting on a limb, the limb's top at the pen's
// origin, its legs hanging over and swinging with swing (radians).
func Perched(p Pen, l Look, swing float64) { sit(p, l, hanging, resting, swing) }

// Asleep draws a koala curled up asleep in a fork, its feet at the pen's
// origin.
func Asleep(p Pen, l Look) {
	l.Shut = true
	sit(p, l, tucked, resting, 0)
}

// Chewing draws a sitting koala chewing a gum leaf, chew (radians) moving
// its jaw.
func Chewing(p Pen, l Look, chew float64) { sit(p, l, feet, chewing, chew) }

// Dazed draws a sitting koala holding its head, eyes shut (DESIGN §5).
func Dazed(p Pen, l Look) {
	l.Shut = true
	sit(p, l, feet, holding, 0)
}

// sit draws a sitting koala, its legs and arms as asked, t moving them.
func sit(p Pen, l Look, lg legs, a arm, t float64) {
	p = p.Turned(l.Lean, 0)
	headY, rx := -17.5, 6.5
	if lg == tucked {
		headY, rx = -14.5, 7.5 // hunched over, rounder
	}
	// The outline: body and legs grown by a line.
	p.Ellipse(0, -7.5, rx+Line, 7.5+Line, 0, ink)
	for _, s := range []float64{-1, 1} {
		switch lg {
		case feet:
			p.Ellipse(s*3.8, -1.2, 3+Line, 1.8+Line, 0, ink)
		case hanging:
			p.Stroke(ink, 3.4+2*Line, s*3.4, -2.5, s*3.6+0.7*math.Sin(t+s), 4.5)
		}
	}
	// The body, shaded below and to the right, lit above and to the left,
	// with a pale chest.
	p.Ellipse(0, -7.5, rx, 7.5, 0, furShade)
	p.Ellipse(-0.7, -8.4, rx-0.8, 6.6, 0, fur)
	p.Ellipse(-0.3, -6.6, 3.4, 4.4, 0, furLight)
	p.Curve(furLight, 1, -4.6, -10.5, -5, -7, -4.2, -4)
	// The legs, and the arms that rest on the knees.
	for _, s := range []float64{-1, 1} {
		switch lg {
		case feet:
			p.Ellipse(s*3.8, -1.2, 3, 1.8, 0, furShade)
			p.Ellipse(s*3.6, -1.5, 2.2, 1.1, 0, fur)
		case hanging:
			fx := s*3.6 + 0.7*math.Sin(t+s)
			p.Stroke(furShade, 3.4, s*3.4, -2.5, fx, 4.5)
			p.Stroke(fur, 1.8, s*3.1, -2.5, fx-0.2, 3.8)
		}
		if a == chewing && s == 1 || a == holding && s == -1 {
			continue // raised, drawn over the head
		}
		restingArm(p, s)
	}
	if l.Joey {
		head(p, 7.5, headY+5, JoeySize, Look{})
	}
	head(p, 0, headY, 1, l)
	switch a {
	case chewing:
		bob := 0.5 * math.Sin(t)
		furLimb(p, 3.2, 4.8, -11.5, 6.2, -13.5, 2.8, -12.6+bob)
		p.Ellipse(5.2, -15.6+bob, 1.3+Line, 3.6+Line, 0.9, ink)
		p.Ellipse(5.2, -15.6+bob, 1.3, 3.6, 0.9, resurrect64.Sage)
		p.Stroke(resurrect64.Fern, 0.4, 3.4, -13.6+bob, 7, -17.6+bob)
		paw(p, 2.8, -12.6+bob)
	case holding:
		furLimb(p, 3.2, -4.8, -11.5, -9, -17, -6, -23)
		paw(p, -6, -23)
	}
}

// restingArm draws an arm, s -1 left or 1 right, resting on the knee.
func restingArm(p Pen, s float64) {
	p.Curve(ink, 3.2+2*Line, s*4.8, -11.5, s*5, -7, s*2.6, -3.6)
	p.Curve(furShade, 3.2, s*4.8, -11.5, s*5, -7, s*2.6, -3.6)
	p.Curve(fur, 1.8, s*4.4, -11.5, s*4.5, -7.5, s*2.8, -4.4)
}

// furLimb draws an arm or a leg w wide, a smooth curve through the points,
// outlined, shaded and lit.
func furLimb(p Pen, w float64, xy ...float64) {
	p.Curve(ink, w+2*Line, xy...)
	p.Curve(furShade, w, xy...)
	p.Curve(fur, w*0.5, xy...)
}

// paw draws a paw's dark claws at (x, y).
func paw(p Pen, x, y float64) { p.Ellipse(x, y, 1.4, 1.1, 0, ink) }

// Clinging draws a koala holding on to the side of the trunk or a limb,
// seen side on and facing right, into the wood, its feet at the pen's
// origin and its arms round the wood about 7 field pixels to the right
// (DESIGN §5). reach, from 0 to 1, stretches it out as it bounds: at 0 it
// is gathered up, back arched, at 1 its arms reach up the wood and its legs
// push down. Backing down, it looks over its shoulder.
func Clinging(p Pen, l Look, reach float64, down bool) { cling(p, l, reach, down, false, 0) }

// ClingChewing draws a koala holding on to the side of a limb as Clinging
// does, still, its near paw holding a gum leaf to its mouth; chew
// (radians) moves its jaw.
func ClingChewing(p Pen, l Look, chew float64) { cling(p, l, 0, false, true, chew) }

// cling draws a koala holding on side on, stretched by reach, backing down
// with down, and with chew holding a gum leaf to its mouth, t moving its
// jaw.
func cling(p Pen, l Look, reach float64, down, chew bool, t float64) {
	arch := 1 - reach
	pawY, footY, knee := -18-4*reach, -0.5+1.5*reach, -5.5-1.8*arch
	// The far arm and leg, in shade.
	for _, xy := range [][]float64{
		{1.5, -14, 4.6, -15.8 - 2*reach, 6.4, pawY + 1.2},
		{0, -5, 5, knee - 0.6, 6.2, footY - 1.2},
	} {
		p.Curve(ink, 2.8+2*Line, xy...)
		p.Curve(furShade, 2.8, xy...)
	}
	// The body, its back to the left, rounder as it gathers up.
	bx, by, rx, ry := -1-0.6*arch, -9.5+0.5*arch, 5.2+0.7*arch, 8.4-0.6*arch
	p.Ellipse(bx, by, rx+Line, ry+Line, 0.12, ink)
	p.Ellipse(bx, by, rx, ry, 0.12, furShade)
	p.Ellipse(bx-0.6, by-0.8, rx-0.9, ry-1.4, 0.12, fur)
	p.Curve(furLight, 1, bx-3.8, by-4, bx-4.6, by, bx-3.8, by+4)
	// The near leg, its thigh round and its foot gripping the bark.
	p.Ellipse(0.6, -4.6, 3.4+Line, 3+Line, 0, ink)
	p.Ellipse(0.6, -4.6, 3.4, 3, 0, furShade)
	p.Ellipse(0.2, -5.1, 2.5, 2, 0, fur)
	furLimb(p, 2.8, 1.8, -3.4, 5.4, knee, 6.8, footY)
	paw(p, 7, footY)
	// The near arm, round the wood.
	if !chew {
		furLimb(p, 3, 1.4, -13, 4.6, -15.5-2*reach, 6.8, pawY)
		paw(p, 7, pawY)
	}
	if l.Joey {
		head(p, bx-4.5, by-5.5, JoeySize, Look{})
	}
	switch {
	case down:
		// Looking back and down past its shoulder.
		side(p.Mirrored(true).Turned(0.45, 19.5), 1, -19.5, l)
	case chew:
		side(p, 1.8, -20.5, l)
		bob := 0.5 * math.Sin(t)
		furLimb(p, 3, 1.4, -13, 4.4, -11.5, 4.8, -15.4+bob)
		p.Ellipse(6.6, -17+bob, 1.3+Line, 3.6+Line, 0.9, ink)
		p.Ellipse(6.6, -17+bob, 1.3, 3.6, 0.9, resurrect64.Sage)
		p.Stroke(resurrect64.Fern, 0.4, 4.8, -15+bob, 8.4, -19+bob)
		paw(p, 5, -15.4+bob)
	default:
		side(p, 1.8, -20.5-1.2*reach, l)
	}
}

// Thump draws a koala that has just hit the trunk, holding on to it side
// on and squashed flat against it by squash (from 1 at the hit to 0), its
// feet at the pen's origin.
func Thump(p Pen, l Look, squash float64) {
	Clinging(p.Squashed(1-0.25*squash, 1+0.08*squash), l, 0, false)
}

// Gripping draws a koala holding on along a sloping limb or branch, seen
// side on and facing right, up it, with the wood's top at the pen's origin
// (DESIGN §5). Turned, it lies along the slope. reach, from 0 to 1, takes
// its front paws ahead to grip further on; follow, from 0 to 1, brings its
// back feet up after them.
func Gripping(p Pen, l Look, reach, follow float64) {
	front, back := 10+2.5*reach, -3.5+2.5*follow
	// The far arm and leg, gripping round under the wood, in shade.
	for _, xy := range [][]float64{
		{5, -6.5, 8.6 + 1.5*reach, -2.6, front + 0.8, 1.2},
		{-4.6, -4, -5 + 1.5*follow, -0.8, back + 0.8, 1.6},
	} {
		p.Curve(ink, 2.8+2*Line, xy...)
		p.Curve(furShade, 2.8, xy...)
	}
	// The body, lying along the wood, lit along its back.
	p.Ellipse(0, -6.2, 8.2+Line, 5.2+Line, -0.05, ink)
	p.Ellipse(0, -6.2, 8.2, 5.2, -0.05, furShade)
	p.Ellipse(-0.6, -7, 7.2, 4, -0.05, fur)
	p.Curve(furLight, 1, -6, -9.2, -1, -10.8, 4, -10.4)
	// The near back leg, its thigh round, the foot gripping below.
	p.Ellipse(-4.6, -4.2, 3.4+Line, 3+Line, 0, ink)
	p.Ellipse(-4.6, -4.2, 3.4, 3, 0, furShade)
	p.Ellipse(-5, -4.7, 2.5, 2, 0, fur)
	furLimb(p, 2.8, -4.2, -2.6, -5.2+1.5*follow, -0.2, back, 1.8)
	paw(p, back, 2)
	// The near arm, reaching ahead and round.
	furLimb(p, 3, 5, -6.5, 8.6+1.5*reach, -3, front, 1.6)
	paw(p, front, 1.8)
	if l.Joey {
		head(p, -2, -12.5, JoeySize, Look{})
	}
	side(p, 10.5, -10.5, l)
}

// OnBough draws a koala sitting on a bough, seen side on and facing right,
// its bottom on the wood at the pen's origin, knees bent, feet over the
// front and a paw holding on.
func OnBough(p Pen, l Look) {
	p = p.Turned(l.Lean, 0)
	knee, foot := [2]float64{5, -3.6}, [2]float64{6, 4.5}
	// The far leg, in shade.
	far := []float64{0.6, -3, knee[0] + 1, knee[1] - 0.5, foot[0] + 1, foot[1] - 0.6}
	p.Curve(ink, 3+2*Line, far...)
	p.Curve(furShade, 3, far...)
	// The body, its back to the left, with a pale chest.
	p.Ellipse(-1, -8, 5.2+Line, 7.8+Line, -0.05, ink)
	p.Ellipse(-1, -8, 5.2, 7.8, -0.05, furShade)
	p.Ellipse(-1.6, -8.8, 4.3, 6.6, -0.05, fur)
	p.Ellipse(1.8, -7.6, 2.2, 4.2, -0.05, furLight)
	p.Curve(furLight, 1, -5, -12, -5.6, -8, -4.8, -4)
	// The near leg: the thigh forward, the knee bent.
	furLimb(p, 3.4, -0.4, -3, knee[0], knee[1], foot[0], foot[1])
	paw(p, foot[0]+0.4, foot[1])
	if l.Joey {
		head(p, -5.5, -13, JoeySize, Look{})
	}
	side(p, 2.5, -19, l)
	// The near paw, holding on.
	furLimb(p, 3, 1.2, -11.5, 4, -6.5, 4.6, -0.6)
	paw(p, 4.8, -0.2)
}

// Walking draws a koala on all fours with its feet at the pen's origin,
// facing right, along a branch or the ground, its claws gripping; phase
// (radians) swings its legs as it goes.
func Walking(p Pen, l Look, phase float64) {
	grip := func(x, y float64) { p.Ellipse(x, y, 1.2, 0.9, 0, ink) }
	sw := math.Sin(phase)
	step := func(x, s float64) (float64, float64) { return x + s*1.6*sw, -0.8 - max(0, s*sw)*1.2 }
	// The far legs, then the outline of the body and near legs.
	for _, g := range [][2]float64{{-4.5, 1}, {6.5, -1}} {
		fx, fy := step(g[0]-0.5, g[1])
		p.Stroke(ink, 2.6+2*Line, g[0], -5, fx, fy)
		p.Stroke(furShade, 2.6, g[0], -5, fx, fy)
		grip(fx, fy)
	}
	p.Ellipse(0, -7, 8.5+Line, 5.2+Line, -0.08, ink)
	p.Ellipse(0, -7, 8.5, 5.2, -0.08, furShade)
	p.Ellipse(-0.6, -8, 7.6, 4, -0.08, fur)
	p.Curve(furLight, 1, -6, -10, -1, -11.4, 4, -11)
	for _, g := range [][2]float64{{-6.5, -1}, {4.5, 1}} {
		fx, fy := step(g[0]+0.5, g[1])
		p.Stroke(ink, 2.8+2*Line, g[0], -5, fx, fy)
		p.Stroke(fur, 2.8, g[0], -5, fx, fy)
		grip(fx, fy)
	}
	if l.Joey {
		head(p, -1.5, -13, JoeySize, Look{})
	}
	side(p, 9.5, -9.5, l)
}

// Falling draws a koala in the air with its arms and legs spread, its feet
// at the pen's origin; flail, from -1 to 1, waves them. Turned about
// FallingMiddle, it tumbles.
func Falling(p Pen, l Look, flail float64) {
	type reach struct{ x0, y0, x1, y1 float64 }
	var spread []reach
	for _, s := range []float64{-1, 1} {
		spread = append(spread,
			reach{s * 4.6, -13, s * 11, -19 + flail*s*1.5},
			reach{s * 3.5, -4, s * 8.5, 0.5 - flail*s*1.2})
	}
	for _, r := range spread {
		p.Stroke(ink, 3.2+2*Line, r.x0, r.y0, r.x1, r.y1)
	}
	p.Ellipse(0, -8.5, 6.2+Line, 7.8+Line, 0, ink)
	for _, r := range spread {
		p.Stroke(furShade, 3.2, r.x0, r.y0, r.x1, r.y1)
		p.Stroke(fur, 1.6, r.x0, r.y0-0.4, r.x1, r.y1-0.4)
		paw(p, r.x1, r.y1)
	}
	p.Ellipse(0, -8.5, 6.2, 7.8, 0, furShade)
	p.Ellipse(-0.7, -9.3, 5.4, 6.8, 0, fur)
	p.Ellipse(-0.3, -7.6, 3.4, 4.6, 0, furLight)
	if l.Joey {
		head(p, 6.5, -14, JoeySize, Look{})
	}
	head(p, 0, -19, 1, l)
	p.Ellipse(0, -13.7, 0.9, 1.1, 0, ink) // an "oh"
}

// FallingMiddle is how far above its feet a falling koala's middle is, to
// turn it about.
const FallingMiddle = 10

// head draws a koala's face about (x, y), s times an adult's size: round
// fluffy ears, the face shaded and lit, the big dark nose and glinting
// eyes, or shut ones.
func head(p Pen, x, y, s float64, l Look) {
	q := func(v float64) float64 { return v * s }
	ear := func(e float64) (float64, float64) {
		if e > 0 {
			return x + q(6.5+0.6*l.Flick), y - q(5.5+1.2*l.Flick)
		}
		return x - q(6.5), y - q(5.5)
	}
	// The outline of the ears and head together.
	for _, e := range []float64{-1, 1} {
		ex, ey := ear(e)
		p.Circle(ex, ey, q(3.9)+Line, ink)
	}
	p.Ellipse(x, y, q(7)+Line, q(6)+Line, 0, ink)
	// The ears: fur round a pale, fluffy middle.
	for _, e := range []float64{-1, 1} {
		ex, ey := ear(e)
		p.Circle(ex, ey, q(3.9), furShade)
		p.Circle(ex-q(0.3), ey-q(0.3), q(3.3), fur)
		p.Circle(ex+e*q(0.5), ey+q(0.2), q(2.3), furLight)
		p.Circle(ex+e*q(1.1), ey+q(0.4), q(1.2), fluff)
	}
	// The face, shaded at the jaw, lit across the brow.
	p.Ellipse(x, y, q(7), q(6), 0, furShade)
	p.Ellipse(x-q(0.5), y-q(0.6), q(6.3), q(5.2), 0, fur)
	p.Curve(furLight, q(1.1), x-q(5), y-q(2.5), x-q(2.5), y-q(4.6), x+q(1), y-q(4.9))
	// The nose, with a soft shine.
	p.Ellipse(x, y+q(1.4), q(2.1), q(2.8), 0, ink)
	p.Ellipse(x-q(0.6), y+q(0.4), q(0.55), q(0.95), 0, noseLight)
	// The eyes, each with a glint, or shut.
	for _, e := range []float64{-1, 1} {
		ex, ey := x+e*q(3.3), y-q(1.2)
		if l.Shut {
			p.Curve(ink, q(0.7), ex-q(1.1), ey, ex, ey+q(0.6), ex+q(1.1), ey)
			continue
		}
		p.Circle(ex, ey, q(1.05), ink)
		p.Circle(ex-q(0.35), ey-q(0.35), q(0.38), fluff)
	}
}

// side draws a koala's face seen from the side about (x, y), facing right:
// one ear forward, one behind, the nose at the front.
func side(p Pen, x, y float64, l Look) {
	fx, fy := x+1.5+0.5*l.Flick, y-5.2-1.2*l.Flick
	p.Circle(x-3.5, y-4.6, 3.4+Line, ink)
	p.Circle(fx, fy, 3.4+Line, ink)
	p.Ellipse(x, y, 6+Line, 5.2+Line, 0, ink)
	p.Circle(x-3.5, y-4.6, 3.4, furShade)
	p.Circle(x-3.2, y-4.4, 2, furLight)
	p.Circle(fx, fy, 3.4, fur)
	p.Circle(fx+0.4, fy+0.3, 2, furLight)
	p.Circle(fx+0.9, fy+0.6, 1, fluff)
	p.Ellipse(x, y, 6, 5.2, 0, furShade)
	p.Ellipse(x-0.4, y-0.6, 5.3, 4.4, 0, fur)
	p.Ellipse(x+4.4, y+0.8, 1.9, 2.5, 0.2, ink)
	p.Ellipse(x+4, y-0.1, 0.5, 0.8, 0.2, noseLight)
	if l.Shut {
		p.Curve(ink, 0.7, x+0.6, y-1.4, x+1.6, y-0.9, x+2.6, y-1.4)
		return
	}
	p.Circle(x+1.6, y-1.4, 1, ink)
	p.Circle(x+1.3, y-1.7, 0.36, fluff)
}
