package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The rangers' colours (DESIGN §2, §3): khaki hat and shirt, a dark green
// band, name tag and emblem, olive green shorts, a brown belt and boots
// with a silver buckle, dark socks and brown hair.
const (
	khaki      = resurrect64.Khaki
	khakiShade = resurrect64.Fern
	khakiLight = resurrect64.Mist
	band       = resurrect64.Pine
	shorts     = resurrect64.Olive
	shortsDark = resurrect64.Slate
	shortsLit  = resurrect64.Moss
	skin       = resurrect64.Apricot
	skinShade  = resurrect64.Salmon
	skinLight  = resurrect64.Blush
	hair       = resurrect64.Bark
	hairLight  = resurrect64.Clay
	leather    = resurrect64.Bark
	leatherLit = resurrect64.Sand
	sock       = resurrect64.Charcoal
	buckle     = resurrect64.Mist
)

// Stance is how a ranger stands this frame.
type Stance struct {
	Woman        bool    // the woman, with her ponytail; else the man
	HandX, HandY float64 // where both hands are, from its feet
	Stride       float64 // how far its legs swing as it walks, from 0 to 1
	Step         float64 // how far through its walk it is, in radians
	Shut         bool    // its eyes closed: a blink
}

// The ranger's frame, from its feet, in field pixels: where its hips and
// shoulders are, and how long its legs and arms are.
const (
	hipY      = -26
	thigh     = 10
	shin      = 10
	shoulderY = -43.5
	upperArm  = 7.6
	forearm   = 7.6
)

// RangerHold is where a ranger holding up the sheet has its hands, from its
// feet: in front of it, at the height of the sheet.
const RangerHold = -30

// Ranger draws a ranger in uniform standing with its feet at the pen's
// origin, seen side on and facing right (DESIGN §2): a wide-brimmed hat
// with a band and badge, a short-sleeved shirt with a pocket, name tag and
// shoulder patch, cargo shorts, a belt with a buckle and a radio, socks and
// work boots. It is about 12 field pixels wide and 60 tall.
func Ranger(p Pen, s Stance) {
	sw := math.Sin(s.Step)
	// It bobs as it walks, highest as its legs pass.
	p = p.At(p.OX, p.OY-0.6*s.Stride*math.Abs(math.Cos(s.Step)))
	near, far := 0.38*s.Stride*sw, -0.38*s.Stride*sw
	// The knee bends as the leg swings through, lifting the boot.
	bend := func(sgn float64) float64 { return 0.7 * s.Stride * max(0, sgn*math.Cos(s.Step)) }
	rangerArm(p, s.HandX-0.6, s.HandY+0.4, -0.6, true)
	rangerLeg(p, -0.4, far, bend(-1), true)
	// The shorts round its hips, the near leg in front.
	p.OutlinedBlob(shorts, -4.2, -29, 4, -29, 4.4, -24, 2, -21.5, -2.4, -21.5, -4.6, -24)
	p.Blob(shortsDark, -4.2, -27, -2.2, -27.4, -1.6, -22.2, -3.8, -22.4)
	rangerLeg(p, 0.4, near, bend(1), false)
	if s.Woman {
		ponytail(p, sw*s.Stride)
	}
	shirt(p)
	rangerHead(p, s)
	rangerArm(p, s.HandX, s.HandY, 0.2, false)
}

// rangerLeg draws a leg from its hip at hx, swung a radians forward and its knee
// bent by b, in shade if far: the cargo shorts over the thigh, the knee
// and shin, a dark sock and a work boot.
func rangerLeg(p Pen, hx, a, b float64, far bool) {
	sk, cl := skin, shorts
	if far {
		sk, cl = skinShade, shortsDark
	}
	kx, ky := hx+thigh*math.Sin(a), hipY+thigh*math.Cos(a)
	ax, ay := kx+shin*math.Sin(a-b), ky+shin*math.Cos(a-b)
	p.Stroke(ink, 2.8+2*Line, kx, ky-2, ax, ay)
	p.Stroke(sk, 2.8, kx, ky-2, ax, ay)
	if !far {
		p.Stroke(skinLight, 0.8, kx+0.6, ky-1, ax+0.6, ay-3)
	}
	// The sock, then the boot, its toe forward and its sole dark.
	sx, sy := ax-(ax-kx)*0.25, ay-(ay-ky)*0.25
	p.Stroke(sock, 3, sx, sy, ax, ay)
	b0 := []float64{ax - 2, ay - 1.4, ax + 1.6, ay - 1.6, ax + 2.6, ay + 2.2, ax + 5, ay + 3.4, ax + 5.2, ay + 6, ax - 2.4, ay + 6}
	p.Stroke(ink, 2*Line, append(b0, b0[:2]...)...)
	p.Poly(leather, b0...)
	p.Poly(ink, ax-2.6, ay+5, ax+5.4, ay+5, ax+5.4, ay+6.4, ax-2.6, ay+6.4)
	if !far {
		p.Stroke(leatherLit, 0.6, ax-0.6, ay-0.4, ax+1.6, ay+2.4, ax+4, ay+3.6)
	}
	// The shorts' leg over the thigh, with a cargo pocket on the near one.
	tx, ty := hx+0.5*thigh*math.Sin(a), hipY+0.5*thigh*math.Cos(a)
	p.Stroke(ink, 5.4+2*Line, hx, hipY, tx, ty)
	p.Stroke(cl, 5.4, hx, hipY, tx, ty)
	if !far {
		mx, my := hx+0.4*thigh*math.Sin(a), hipY+0.4*thigh*math.Cos(a)
		p.Ellipse(mx-0.6, my, 1.4, 1.7, a, shortsDark)
		p.Stroke(shortsLit, 0.5, mx-1.8, my-1.6, mx+0.6, my-1.8)
		p.Stroke(shortsLit, 0.7, hx+2, hipY-1, tx+2.2, ty)
	}
}

// shirt draws the khaki work shirt, shaded at the back, its chest pocket,
// dark green name tag and collar, and the belt with its silver buckle and
// the radio on it.
func shirt(p Pen) {
	p.OutlinedBlob(khaki, -3.6, -29.6, 3.8, -29.6, 4.7, -36, 4.4, -41, 2.6, -45, -1, -46, -3.8, -44.5, -4.5, -39, -4.2, -33)
	p.Blob(khakiShade, -3.4, -30, -1.4, -30.2, -2, -38, -3, -44, -4.2, -40, -4, -33.4)
	p.Curve(khakiLight, 0.9, 0.6, -44.8, 3.2, -42.6, 4.2, -38.4)
	// The pocket with its buttoned flap, and the name tag over it.
	p.Poly(khakiShade, 0.8, -40.4, 3.9, -40.4, 3.9, -37, 0.8, -37)
	p.Poly(ink, 0.6, -40.8, 4.1, -40.8, 4.1, -39.6, 0.6, -39.6)
	p.Poly(khaki, 0.9, -40.6, 3.8, -40.6, 3.8, -39.8, 0.9, -39.8)
	p.Circle(2.35, -39.6, 0.4, khakiShade)
	p.Poly(band, 1, -43, 3.6, -43, 3.6, -41.8, 1, -41.8)
	// The collar.
	p.Poly(khakiShade, -0.4, -45.4, 2.6, -45, 1.4, -43.2)
	// The belt, its buckle at the front and the radio at the back.
	p.Poly(ink, -4.4, -30.8, 4.4, -30.8, 4.4, -28.2, -4.4, -28.2)
	p.Poly(leather, -4, -30.2, 4, -30.2, 4, -28.8, -4, -28.8)
	p.Poly(ink, 2.4, -31, 4.2, -31, 4.2, -28, 2.4, -28)
	p.Poly(buckle, 2.8, -30.6, 3.8, -30.6, 3.8, -28.4, 2.8, -28.4)
	p.Stroke(ink, 0.6, -3.8, -31, -3.8, -34.4)
	p.Poly(ink, -5.4, -31.6, -2.6, -31.6, -2.6, -25.6, -5.4, -25.6)
	p.Poly(resurrect64.Charcoal, -5, -31.2, -3, -31.2, -3, -26, -5, -26)
	p.Circle(-4, -30, 0.45, resurrect64.Red)
}

// rangerArm draws an arm from the shoulder at sx to a hand at (hx, hy), its
// elbow bent down and back to fit, in shade if far: the short sleeve, with
// the shoulder patch on the near one, the forearm and the hand.
func rangerArm(p Pen, hx, hy, sx float64, far bool) {
	sl, sk := khaki, skin
	if far {
		sl, sk = khakiShade, skinShade
	}
	sy := shoulderY
	dx, dy := hx-sx, hy-sy
	l := math.Hypot(dx, dy)
	bend := math.Sqrt(max(0, upperArm*upperArm-l*l/4))
	ex, ey := sx+dx/2-dy/l*bend, sy+dy/2+dx/l*bend
	if ex > sx+dx/2 { // bend the elbow back, never forward
		ex, ey = sx+dx/2+dy/l*bend, sy+dy/2-dx/l*bend
	}
	p.Curve(ink, 2.6+2*Line, ex, ey, hx, hy)
	p.Curve(sk, 2.6, ex, ey, hx, hy)
	p.Circle(hx, hy, 1.6+Line, ink)
	p.Circle(hx, hy, 1.6, sk)
	if !far {
		p.Circle(hx-0.4, hy-0.5, 0.6, skinLight)
	}
	cx, cy := sx+0.8*(ex-sx), sy+0.8*(ey-sy)
	p.Stroke(ink, 4.4+2*Line, sx, sy, cx, cy)
	p.Stroke(sl, 4.4, sx, sy, cx, cy)
	if !far {
		p.Stroke(khakiLight, 0.7, sx+0.4, sy-1.6, cx+0.6, cy-1.2)
		px, py := sx+0.35*(ex-sx), sy+0.35*(ey-sy)
		p.Circle(px, py, 1.2, band)
		p.Circle(px, py, 0.45, khakiLight)
	}
}

// ponytail draws the woman's brown ponytail, out from under the back of
// her hat and down past her collar, swung by swing as she walks.
func ponytail(p Pen, swing float64) {
	tail := []float64{-2.6, -51.2, -5, -50, -6.2 - 0.6*swing, -47, -5.6 - 0.9*swing, -44}
	p.Curve(ink, 3+2*Line, tail...)
	p.Curve(hair, 3, tail...)
	p.Curve(hairLight, 0.8, -3.4, -51.4, -5.4, -50, -6.2-0.6*swing, -47.4)
	p.Circle(-3, -51.2, 1.1, band)
}

// rangerHead draws a ranger's head in profile, facing right, with its hair, face
// and wide-brimmed hat with its band and badge.
func rangerHead(p Pen, s Stance) {
	x, y := 1.2, -50.6
	// The neck, then the head, shaded at the back and the jaw.
	p.Stroke(ink, 2.6+2*Line, 0.8, -44.6, 1.2, -47)
	p.Stroke(skinShade, 2.6, 0.8, -44.6, 1.2, -47)
	p.Circle(x+4.8, y+0.6, 1.3+Line, ink) // the nose
	p.Ellipse(x, y, 4.6+Line, 5+Line, 0, ink)
	p.Ellipse(x, y, 4.6, 5, 0, skinShade)
	p.Ellipse(x+0.5, y-0.4, 4, 4.4, 0, skin)
	p.Circle(x+4.8, y+0.6, 1.3, skin)
	p.Circle(x+4.6, y+0.2, 0.5, skinLight)
	// The hair at the back of the head, under the hat: short for the man,
	// tied back for the woman.
	locks := []float64{-3.6, -52.8, -1, -53.8, 0.8, -53, 0.2, -50.8, -1, -48.8, -2.6, -47.6, -3.8, -49.8}
	if s.Woman {
		locks = []float64{-3.6, -52.8, -1, -53.8, 1.6, -53.2, 0.6, -51.8, -1.2, -50.6, -3, -50, -3.8, -51}
	}
	p.OutlinedBlob(hair, locks...)
	p.Curve(hairLight, 0.6, -2.6, -52.4, -0.8, -53, 0.4, -52.8)
	// The ear, the eye with its glint, or shut, the brow, a cheek and a
	// smile.
	p.Ellipse(x-1.4, y+0.4, 0.9, 1.3, 0, skinShade)
	p.Ellipse(x-1.2, y+0.4, 0.4, 0.7, 0, skin)
	ex, ey := x+2.6, y-0.8
	if s.Shut {
		p.Curve(ink, 0.6, ex-0.8, ey, ex, ey+0.5, ex+0.8, ey)
	} else {
		p.Ellipse(ex, ey, 0.75, 0.95, 0, ink)
		p.Circle(ex-0.25, ey-0.35, 0.3, resurrect64.White)
	}
	p.Stroke(hair, 0.7, ex-1, ey-1.8, ex+1, ey-2)
	if s.Woman {
		p.Stroke(ink, 0.4, ex+0.6, ey-0.8, ex+1.3, ey-1.3) // lashes
	}
	p.Circle(x+2.4, y+1.6, 0.6, resurrect64.Salmon)
	p.Curve(ink, 0.5, x+2.6, y+2.8, x+3.4, y+3.4, x+4.2, y+3)
	hat(p)
}

// hat draws the wide-brimmed khaki hat, its dented crown with the dark
// green band and the round badge.
func hat(p Pen) {
	p.Ellipse(1.2, -54.4, 7.8+Line, 1.5+Line, -0.03, ink)
	p.Ellipse(1.2, -54.4, 7.8, 1.5, -0.03, khakiShade)
	p.Ellipse(1.2, -54.8, 7.4, 1, -0.03, khaki)
	p.OutlinedBlob(khaki, -3.4, -54.6, -3.2, -58, -1.4, -59.8, 1.2, -59.2, 3.8, -59.8, 5.4, -58, 5.6, -54.6)
	p.Blob(khakiShade, -3.2, -54.8, -1.6, -55, -1.8, -58.6, -3, -57.6)
	p.Curve(khakiLight, 0.7, -1.2, -59.2, 1.2, -58.6, 3.6, -59.2)
	p.Poly(band, -3.4, -54.8, 5.6, -54.8, 5.5, -56.4, -3.3, -56.4)
	p.Circle(1.6, -57.6, 1.3+Line*0.6, ink)
	p.Circle(1.6, -57.6, 1.3, emblem)
	p.Circle(1.6, -57.6, 0.5, resurrect64.White)
}

// Sheet draws the rangers' sheet, held at its ends at x0 and x1 across the
// pen's origin, sagging sag field pixels below the ends in its middle and
// pushed down a further dip at x, where it has just caught a koala: white
// canvas, shaded underneath, with a seam across it at each of seams, and
// gathered at the ends where the rangers grip it.
func Sheet(p Pen, x0, x1, sag, x, dip float64, seams []float64) {
	const n = 20
	thick := 2.8
	top := func(at float64) float64 {
		u := (at - x0) / (x1 - x0)
		d := (at - x) / 7
		return 4*sag*u*(1-u) + dip*math.Exp(-d*d)
	}
	var rim, under []float64
	for i := 0; i <= n; i++ {
		at := x0 + (x1-x0)*float64(i)/n
		u := float64(i) / n
		rim = append(rim, at, top(at))
		under = append([]float64{at, top(at) + thick + 1.2*4*u*(1-u)}, under...)
	}
	cloth := append(append([]float64{}, rim...), under...)
	p.Stroke(ink, 2*Line, append(cloth, cloth[:2]...)...)
	p.Poly(resurrect64.White, cloth...)
	var fold []float64
	for i := 0; i+1 < len(under); i += 2 {
		fold = append(fold, under[i], under[i+1]-1.1)
	}
	p.Stroke(resurrect64.Mist, 1.4, fold...)
	for _, sx := range seams {
		y := top(sx)
		p.Stroke(resurrect64.Fog, 0.5, sx, y+0.4, sx, y+thick+0.4)
	}
	// Gathered up in the rangers' hands at each end.
	for _, e := range []float64{x0, x1} {
		p.Ellipse(e, 1.2, 1.8+Line, 2.6+Line, 0, ink)
		p.Ellipse(e, 1.2, 1.8, 2.6, 0, resurrect64.Mist)
		p.Ellipse(e-0.3, 0.6, 1, 1.6, 0, resurrect64.White)
	}
}
