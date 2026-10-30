package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The reptiles' colours (DESIGN D32): the lace monitor blue-black with a
// sheen of indigo, finely speckled yellow all over and ringed with speckles
// down its tail; the carpet python black with pale yellow rosettes and
// spots, a cream edge to its belly and yellow lips; both with dark red
// tongues.
const (
	monitorHide  = resurrect64.Navy
	monitorSheen = resurrect64.Indigo
	monitorSpeck = resurrect64.Gold
	monitorLip   = resurrect64.Khaki
	pythonBase   = resurrect64.Charcoal
	pythonSpot   = resurrect64.Lemon
	pythonBelly  = resurrect64.Khaki
	tongue       = resurrect64.Wine
)

// Goanna is how the lace monitor is this frame.
type Goanna struct {
	Step   float64 // radians through its walk
	Walk   float64 // how far its legs swing, from 0 standing to 1 walking
	Tongue float64 // how far its tongue is out, from 0 to 1
}

// How far the lace monitor reaches behind and ahead of its origin, in
// field pixels: its tail's tip and its snout.
const (
	MonitorBehind = 50
	MonitorAhead  = 23
)

// MonitorStride is how far the lace monitor walks in one step of its walk,
// a full turn of Step.
const MonitorStride = 9

// The lace monitor's body side on, hips to snout, as its top and its
// underside: x, then y, in field pixels above its feet.
var (
	monitorTop    = []float64{-11, -4.7, -6, -6.3, 0, -6.7, 5, -6.3, 8.5, -5.5, 11, -5.5, 14, -6.1, 16.5, -6.4, 19.5, -6, 22.5, -5.1}
	monitorUnder  = []float64{22.5, -5.1, 19.5, -4.3, 16.5, -4.1, 13, -4.2, 9, -3.7, 5, -2.9, 0, -2.6, -6, -2.8, -11, -3.5}
	monitorTailXY = []float64{-9.5, -4.1, -20, -3, -32, -1.7, -43, -0.9, -MonitorBehind, -0.9}
)

// LaceMonitor draws a lace monitor facing right, its feet on the ground at
// the pen's origin, side on (DESIGN D32): a long slim tail on the ground
// behind it, its body low and slender between splayed legs that walk in
// diagonal pairs, a long neck and a narrow pointed head, and its forked
// tongue flicking.
func LaceMonitor(p Pen, g Goanna) {
	// Its shadow on the dirt.
	p.Faded(0.5).Ellipse(-8, 0.4, 28, 1.6, 0, resurrect64.Rust)
	// The legs in diagonal pairs: near front with far hind, far front with
	// near hind, each swinging forward lifted and pushing back down.
	swing := func(phase float64) (reach, lift float64) {
		return 2 * g.Walk * math.Sin(g.Step+phase), 1.2 * g.Walk * max(0, math.Cos(g.Step+phase))
	}
	r1, l1 := swing(0)
	r2, l2 := swing(math.Pi)
	monitorLeg(p, 7.5, 2.5+r2, l2, 0.85, true) // far front
	monitorLeg(p, -7, -2.5+r1, l1, 1, true)    // far hind
	// The tail, long and slim, swaying a little as it walks, ringed with
	// speckles.
	sway := 0.5 * g.Walk * math.Sin(g.Step)
	tail := make([]float64, len(monitorTailXY))
	copy(tail, monitorTailXY)
	tail[3] += sway * 0.4
	tail[5] -= sway * 0.6
	pts := curve(tail, false)
	n := len(pts) / 2
	width := func(i int) float64 { return 3 - 2.4*float64(i)/float64(n-1) }
	for i := 1; i < n; i++ {
		p.Stroke(ink, width(i)+2*Line, pts[2*i-2], pts[2*i-1], pts[2*i], pts[2*i+1])
	}
	for i := 1; i < n; i++ {
		p.Stroke(monitorHide, width(i), pts[2*i-2], pts[2*i-1], pts[2*i], pts[2*i+1])
	}
	for i := 2; i < n-1; i += 2 {
		x, y, w := pts[2*i], pts[2*i+1], width(i)
		for _, f := range []float64{-0.3, 0, 0.3} {
			p.Circle(x, y+f*w, 0.32*w/3+0.12, monitorSpeck)
		}
	}
	// The body, neck and head, bobbing a little with each step: a sheen
	// along its back, its yellow lips, and speckles all over.
	q := p.Rel(0, -0.3*g.Walk*math.Abs(math.Sin(g.Step)))
	q.OutlinedBlob(monitorHide, append(append([]float64{}, monitorTop...), monitorUnder[2:]...)...)
	q.Curve(monitorSheen, 1.1, -8, -5.4, -2, -6, 4, -5.8, 8, -5.1, 12, -5.1, 16, -5.7)
	q.Curve(monitorLip, 0.5, 14, -4.5, 17, -4.5, 20.5, -4.8)
	// Speckles scattered over it, a few left out, finer on the neck.
	for row, y := range []float64{-5.9, -5, -4.1, -3.2} {
		for i, x := 0, -10.4+0.7*float64(row%2); x < 13; i, x = i+1, x+1.4 {
			h := hash(row*31 + i)
			sx, sy := x+0.45*(h-0.5), y+0.4*(hash(row*17+i*7)-0.5)
			top, under := lerpAt(monitorTop, sx), lerpAt(monitorUnder, sx)
			if h < 0.15 || sy < top+0.45 || sy > under-0.35 {
				continue
			}
			r := 0.36
			if sx > 9 {
				r = 0.27
			}
			q.Circle(sx, sy, r, monitorSpeck)
		}
	}
	// The eye, ringed yellow, and the nostril near the snout's tip.
	q.Circle(17.6, -5.5, 0.6, monitorSpeck)
	q.Circle(17.6, -5.5, 0.36, ink)
	q.Circle(21.2, -5.2, 0.22, ink)
	// The near legs.
	monitorLeg(p, 6, -1.5+r1, l1, 0.85, false)
	monitorLeg(p, -8.5, 3+r2, l2, 1, false)
	// The forked tongue, flicking out.
	if g.Tongue > 0 {
		x := 22.5 + 3*g.Tongue
		q.Stroke(tongue, 0.45, 22.3, -4.9, x, -4.8)
		q.Stroke(tongue, 0.35, x, -4.8, x+1.4*g.Tongue, -5.5)
		q.Stroke(tongue, 0.35, x, -4.8, x+1.4*g.Tongue, -4.1)
	}
}

// monitorLeg draws one of the lace monitor's legs from its shoulder or hip
// at x, its foot reach ahead and lifted lift, s its size: the upper leg out
// to a splayed elbow or knee, the lower leg down to the foot, and four long
// clawed toes. A far leg is in shade; a near one is speckled.
func monitorLeg(p Pen, x, reach, lift, s float64, far bool) {
	c := monitorHide
	if far {
		c = ink
	}
	ex, ey := x+reach*0.4, -2.6-lift/2
	fx, fy := x+reach, -0.4-lift
	p.Stroke(ink, 2*s+2*Line, x, -4.2, ex, ey)
	p.Stroke(ink, 1.4*s+2*Line, ex, ey, fx, fy)
	p.Stroke(c, 2*s, x, -4.2, ex, ey)
	p.Stroke(c, 1.4*s, ex, ey, fx, fy)
	for i, a := range []float64{-0.45, -0.1, 0.25, 0.6} {
		l := (2 + 0.5*float64(i%2)) * s
		p.Stroke(ink, 0.5, fx, fy, fx+l*math.Cos(a), fy+0.35+0.45*l*math.Sin(a))
	}
	if !far {
		p.Circle(x+(ex-x)*0.5, (-4.2+ey)/2, 0.34, monitorSpeck)
		p.Circle(ex+(fx-ex)*0.5, (ey+fy)/2, 0.3, monitorSpeck)
	}
}

// hash is a fixed number from 0 to 1 for n, to scatter marks the same way
// every frame.
func hash(n int) float64 {
	_, f := math.Modf(math.Abs(math.Sin(float64(n)*12.9898) * 43758.5453))
	return f
}

// lerpAt is the y of the line through the x, y pairs at x, or 0 past its
// ends; the pairs run either way along x.
func lerpAt(xy []float64, x float64) float64 {
	for i := 2; i+1 < len(xy); i += 2 {
		x0, y0, x1, y1 := xy[i-2], xy[i-1], xy[i], xy[i+1]
		if (x-x0)*(x-x1) <= 0 && x0 != x1 {
			return y0 + (y1-y0)*(x-x0)/(x1-x0)
		}
	}
	return 0
}

// Snake is how the carpet python is this frame.
type Snake struct {
	Wave   float64 // radians: where along the ground's curves its body lies
	Coil   float64 // how far it is curled up, from 0 stretched out to 1
	Tongue float64 // how far its tongue is out, from 0 to 1
}

// The carpet python, in field pixels: how long it is stretched out, how
// far its curves reach to either side and how many there are along it,
// how much of their depth is seen, and how high it lifts its head.
const (
	PythonLength = 76
	pythonSwing  = 8
	pythonWaves  = 1.6
	pythonDepth  = 0.55
	pythonLift   = 3.5
	pythonSteps  = 90
)

// PythonWave is the Wave of a python that has slid x field pixels along
// the ground: its body follows the same curves in the dirt as it goes, so
// it slithers rather than wriggling on the spot.
func PythonWave(x float64) float64 { return 2 * math.Pi * pythonWaves * x / PythonLength }

// CarpetPython draws a carpet python heading right, its belly on the
// ground about the pen's origin, seen from a little above (DESIGN D32):
// tail to head, its body in curves in the dirt and its neck lifted to lead
// the way, or curled up in a coil to sunbake with its head resting on top.
func CarpetPython(p Pen, s Snake) {
	coil := s.Coil * s.Coil * (3 - 2*s.Coil)
	type pt struct{ x, y, ground, w float64 }
	pts := make([]pt, pythonSteps)
	for i := range pts {
		t := float64(i) / (pythonSteps - 1)
		// Stretched out, along the curves in the dirt: deepest along its
		// middle, its tail trailing and its head leading straighter, lifted.
		x := -PythonLength/2 + PythonLength*t
		env := 0.35 + 0.65*math.Sin(math.Pi*min(t/0.9, 1))
		z := pythonSwing * env * math.Sin(2*math.Pi*pythonWaves*t-s.Wave)
		lift := pythonLift * math.Pow(max(0, (t-0.82)/0.18), 2)
		// Curled up, round and round from its tail outside to its head in
		// the middle, which rests on top.
		a := 2 * math.Pi * 2 * (1 - t)
		r := 2.5 + 10*(1-t)
		cx, cz := r*math.Cos(a), r*math.Sin(a)
		cy := cz*0.55 - 1.6*max(0, (t-0.85)/0.15)
		w := 3.2
		switch {
		case t < 0.3: // the tail tapers to a point
			w = 0.8 + 2.4*t/0.3
		case t > 0.9: // the neck
			w = 2.5
		}
		ground := z * pythonDepth
		pts[i] = pt{x + (cx-x)*coil, ground - lift + (cy-ground+lift)*coil, ground + (cy-ground)*coil, w}
	}
	// Its shadow on the dirt, under it wherever it lifts.
	for i := 0; i < len(pts); i += 3 {
		q := pts[i]
		p.Faded(0.25).Circle(q.x, q.ground+0.9+0.8*coil, q.w/2+0.4, resurrect64.Rust)
	}
	// Its body, tail to head, each length drawn over the last so the coils
	// lie on one another: the outline, then the body filled in again back
	// to where the outline reached, with the cream edge of its belly.
	fill := func(a, b int) {
		a = max(a, 0)
		xy := make([]float64, 0, 2*(b-a+1))
		belly := make([]float64, 0, 2*(b-a+1))
		for _, q := range pts[a : b+1] {
			xy = append(xy, q.x, q.y)
			belly = append(belly, q.x, q.y+q.w*0.36)
		}
		w := pts[b].w
		p.Stroke(pythonBase, w, xy...)
		p.Stroke(pythonBelly, w*0.2, belly...)
	}
	const back = 4 // how far back an outline reaches over the body, in steps
	const late = 8 // how long a mark waits for the outlines past it
	// Its marks: pale yellow spots in small clusters down its back, the
	// clusters on alternate sides, with odd spots scattered between.
	mark := func(i int) {
		q := pts[i]
		a := math.Atan2(pts[i+1].y-pts[i-1].y, pts[i+1].x-pts[i-1].x)
		sn, cs := math.Sincos(a)
		at := func(along, across, r float64) {
			p.Circle(q.x+along*cs-across*q.w*sn, q.y+along*sn+across*q.w*cs, r, pythonSpot)
		}
		side := 0.18
		if i%8 == 4 {
			side = -side
		}
		if i%4 == 0 {
			at(0, side, 0.3)
			at(0.7, side+0.12, 0.24)
			at(-0.6, side+0.1, 0.24)
			at(0.1, side-0.22, 0.22)
		}
		if h := hash(i); h > 0.4 {
			at(1.6*(h-0.5), -side*1.4, 0.2)
		}
	}
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		p.Stroke(ink, b.w+2*Line, a.x, a.y, b.x, b.y)
		fill(i-1-back, i-1)
		fill(i-1, i)
		if j := i - late; j > 14 && j%2 == 0 && j < pythonSteps-6 {
			mark(j)
		}
	}
	// The head: black above with spots, yellow lips, an eye and its tongue.
	h := pts[len(pts)-1]
	n := pts[len(pts)-4]
	// Curled up, its head rests level on its coils.
	a := math.Atan2(h.y-n.y, h.x-n.x) * (1 - coil)
	hp := p.Rel(h.x, h.y).Turned(a, 0)
	// The head, a little wider than the neck it runs on from: its outline,
	// then the neck filled over the join.
	hp.OutlinedBlob(pythonBase, -2, -1.2, 0.8, -1.4, 3.2, -0.9, 4, 0, 3.2, 0.75, 0.8, 1.2, -2, 1.1)
	fill(len(pts)-5, len(pts)-1)
	hp.Curve(pythonSpot, 0.45, -1.5, 0.95, 1, 1, 3.4, 0.55)
	for _, d := range [][2]float64{{-1.6, -0.5}, {0, -0.8}, {1.6, -0.6}} {
		hp.Circle(d[0], d[1], 0.25, pythonSpot)
	}
	hp.Circle(2, -0.2, 0.45, pythonBelly)
	hp.Circle(2.05, -0.2, 0.28, ink)
	if s.Tongue > 0 {
		x := 4.2 + 2.4*s.Tongue
		hp.Stroke(tongue, 0.4, 4.1, 0.2, x, 0.3)
		hp.Stroke(tongue, 0.3, x, 0.3, x+s.Tongue, -0.3)
		hp.Stroke(tongue, 0.3, x, 0.3, x+s.Tongue, 0.9)
	}
}
