package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The kangaroo's colours (DESIGN §3): sandy grey-brown fur, shaded clay,
// with a pale chest and belly, and dark paws, nose and ear tips.
const (
	roo      = resurrect64.Sand
	rooShade = resurrect64.Clay
	rooLight = resurrect64.Blush
	rooDark  = resurrect64.Mauve
)

// Roo is how a kangaroo is this frame.
type Roo struct {
	Graze float64 // how far it is down grazing, from 0 sitting up to 1
	Hop   float64 // how far through a hop it is, from 0 landing to 1; -1 not hopping
	Chew  float64 // radians through its chewing
	Flick float64 // an ear flicked, from 0 to 1
	Pouch bool    // a joey looking out of its pouch
}

// The hop's parts, as parts of it: the feet on the ground from the landing
// to the push off, then in the air.
const (
	hopLand = 0.12 // the landing taken up, crouched
	hopPush = 0.24 // the push off, the legs out straight behind
)

// HopLift is how high a kangaroo f of the way through a hop is, from 0 on
// the ground to 1 at the top of its bound.
func HopLift(f float64) float64 {
	if f < hopPush {
		return 0
	}
	return math.Sin(math.Pi * (f - hopPush) / (1 - hopPush))
}

// HopAlong is how far along a hop's length a kangaroo f of the way
// through it is, from 0 to 1: a little while its feet are down, the rest
// in the air.
func HopAlong(f float64) float64 {
	if f < hopPush {
		return 0.12 * f / hopPush
	}
	return 0.12 + 0.88*(f-hopPush)/(1-hopPush)
}

// The kangaroo's frame at full size, in field pixels: the bones of its hind
// leg, and its body from hip to shoulder.
const (
	rooThigh = 6
	rooShin  = 6.6
	rooFoot  = 8
	rooBody  = 11
)

// legPose is the angles of a hind leg's thigh, shin and foot, in radians
// clockwise from pointing forward.
type legPose [3]float64

func deg(d float64) float64 { return d * math.Pi / 180 }

// The hind legs' poses: sitting, crouched at the landing, out straight
// behind at the push off, folded up under the body as they swing forward,
// and forward to land.
var (
	legSit    = legPose{deg(45), deg(150), deg(5)}
	legCrouch = legPose{deg(30), deg(160), deg(3)}
	legPush   = legPose{deg(118), deg(150), deg(160)}
	legFold   = legPose{deg(48), deg(172), deg(125)}
	legTuck   = legPose{deg(22), deg(142), deg(-6)}
)

func (a legPose) to(b legPose, f float64) legPose {
	f = f * f * (3 - 2*f)
	return legPose{lerp(a[0], b[0], f), lerp(a[1], b[1], f), lerp(a[2], b[2], f)}
}

// hopLegs is the hind legs' pose f of the way through a hop.
func hopLegs(f float64) legPose {
	switch {
	case f < hopLand:
		return legTuck.to(legCrouch, f/hopLand)
	case f < hopPush:
		return legCrouch.to(legPush, (f-hopLand)/(hopPush-hopLand))
	case f < 0.42:
		return legPush
	case f < 0.66:
		return legPush.to(legFold, (f-0.42)/0.24)
	case f < 0.9:
		return legFold.to(legTuck, (f-0.66)/0.24)
	}
	return legTuck
}

// along is the point l from (x, y) at angle a, clockwise from forward.
func along(x, y, a, l float64) (float64, float64) {
	s, c := math.Sincos(a)
	return x + c*l, y + s*l
}

// capsule is the outline round two circles, (x1, y1) radius r1 and (x2,
// y2) radius r2, and the straight sides between them: a teardrop when one
// is bigger.
func capsule(x1, y1, r1, x2, y2, r2 float64) []float64 {
	d := math.Hypot(x2-x1, y2-y1)
	a := math.Atan2(y2-y1, x2-x1)
	phi := math.Acos(min(max((r1-r2)/d, -1), 1))
	const n = 14
	var out []float64
	for i := 0; i <= n; i++ {
		t := a + phi + (2*math.Pi-2*phi)*float64(i)/n
		out = append(out, x1+r1*math.Cos(t), y1+r1*math.Sin(t))
	}
	for i := 0; i <= n; i++ {
		t := a - phi + 2*phi*float64(i)/n
		out = append(out, x2+r2*math.Cos(t), y2+r2*math.Sin(t))
	}
	return out
}

// rooLeg is a hind leg's knee, heel and toe from its hip at (x, y).
func rooLeg(x, y float64, a legPose) (kx, ky, hx, hy, tx, ty float64) {
	kx, ky = along(x, y, a[0], rooThigh)
	hx, hy = along(kx, ky, a[1], rooShin)
	tx, ty = along(hx, hy, a[2], rooFoot)
	return
}

// Kangaroo draws a kangaroo side on and facing right, its hind feet on
// the ground at the pen's origin (DESIGN §3, D27), built as one: a long
// thick tail; jointed hind legs, the thigh in the big haunch, the long
// shin back to the heel and the long foot; a body that narrows from the
// haunch to the small chest, with a pale belly and a joey in the pouch;
// short arms; a thin neck and a small head with long ears. Sitting up it
// leans on its tail; grazing it is level, its back arched, its forepaws
// and muzzle on the grass; hopping it is level too, its tail out behind,
// its legs driving back then swinging forward. Sitting up it is about 28
// field pixels tall; a joey is one drawn smaller.
func Kangaroo(p Pen, r Roo) {
	hopping := r.Hop >= 0
	legs := legSit.to(legCrouch, 0.6*r.Graze)
	if hopping {
		legs = hopLegs(r.Hop)
	}
	// The hips stand as high as the legs hold them, the lowest foot on the
	// ground.
	_, _, _, hy, _, ty := rooLeg(0, 0, legs)
	hipY := -max(hy, ty)
	hipX := -0.5

	// The body leans from upright to level as it grazes or bounds; in the
	// air it is longest, its head forward.
	lean := lerp(deg(18), deg(104), r.Graze)
	if hopping {
		lean = deg(68)
		if r.Hop >= hopPush {
			lean = deg(76)
		}
	}
	ux, uy := math.Sin(lean), -math.Cos(lean) // up the body
	nx, ny := -uy, ux                         // out of its front
	sx, sy := hipX+ux*rooBody, hipY+uy*rooBody
	const haunch, chest = 5.6, 3

	// The tail: down to the ground behind sitting or grazing, out straight
	// behind as a counterweight in a bound, lifting as the legs go back.
	tail := []float64{hipX - 2, hipY + 3, hipX - 10, -0.8, hipX - 20, -0.4}
	if hopping {
		up := 2.5 * math.Sin(2*math.Pi*r.Hop)
		tail = []float64{hipX - 2.5, hipY + 1.5, hipX - 9, hipY + 2.4 + up, hipX - 15, hipY + 2 + 1.5*up, hipX - 21, hipY + 0.4 + 2*up}
	}
	taper(p, ink, 4.6+2*Line, 0.9+2*Line, tail)
	taper(p, rooShade, 4.6, 0.9, tail)
	taper(p, roo, 2.6, 0.4, shift(tail, 0.3, -0.8))

	// The far hind leg, a little behind the near one, in shade.
	farLegs := legs
	if !hopping {
		farLegs[2] += deg(4)
	}
	drawRooLeg(p, hipX-1.2, hipY, farLegs, true)

	// The neck and head, down to the grass as it grazes, forward in a
	// bound, up and looking about when it sits.
	neckA := lerp(deg(-62), deg(12), r.Graze)
	headA := lerp(deg(14), deg(66), r.Graze)
	if hopping {
		neckA, headA = deg(-30), deg(10)
	}
	cx, cy := along(sx+nx*0.6, sy+ny*0.6, neckA, 4.4)
	jaw := 0.3 * math.Sin(r.Chew) * r.Graze
	mx, my := along(cx, cy, headA, 3.4)
	my += jaw

	// The arms: held in front of the chest, reaching down to the grass as
	// it grazes, tucked in a bound.
	paw := func(dx float64) (float64, float64, float64, float64) {
		ax, ay := sx+nx*1.6+dx, sy+ny*1.6+1
		if hopping {
			return ax + 1.4, ay + 2.2, ax + 3, ay + 1.4
		}
		gx, gy := ax+2+dx, -0.6
		ex, ey := lerp(ax+0.6, (ax+gx)/2+0.6, r.Graze), lerp(ay+2.6, (ay+gy)/2, r.Graze)
		return ex, ey, lerp(ax+2.4, gx, r.Graze), lerp(ay+2, gy, r.Graze)
	}
	arm := func(dx float64, c int) {
		ax, ay := sx+nx*1.6+dx, sy+ny*1.6+1
		ex, ey, px, py := paw(dx)
		p.Stroke(ink, 1.3+2*Line, ax, ay, ex, ey, px, py)
		p.Stroke(c, 1.3, ax, ay, ex, ey, px, py)
		p.Ellipse(px+0.3, py, 0.7, 0.5, 0, rooDark)
	}
	arm(-0.8, rooShade)
	ear := func(dx, a float64, c int) {
		ex, ey := along(cx+dx, cy-1.2, a, 2.4)
		p.Ellipse(ex, ey, 2.8+Line, 0.9+Line, a, ink)
		p.Ellipse(ex, ey, 2.8, 0.9, a, c)
		tx, ty := along(cx+dx, cy-1.2, a, 4.2)
		p.Ellipse(tx, ty, 0.9, 0.6, a, rooDark)
	}
	earA := lerp(deg(-112), deg(-150), r.Graze) + headA - deg(14)
	ear(-0.9, earA-0.12, rooShade)

	// The outline of the body, neck and head together, then their fur:
	// shaded below, lit along the back, the belly pale.
	body := capsule(hipX, hipY+0.6, haunch, sx, sy, chest)
	neck := capsule(sx+nx*0.6, sy+ny*0.6, 2.2, cx, cy, 1.6)
	skull := capsule(cx, cy, 2.3, mx, my, 1.25)
	for _, s := range [][]float64{body, neck, skull} {
		p.Stroke(ink, 2*Line, append(s, s[:2]...)...)
	}
	for _, s := range [][]float64{body, neck, skull} {
		p.Poly(rooShade, s...)
	}
	p.Poly(roo, capsule(hipX-0.5*nx-0.4*ux, hipY+0.6-0.5*ny-0.4*uy, haunch-1, sx-0.4*nx, sy-0.4*ny, chest-0.8)...)
	p.Poly(roo, capsule(sx+nx*0.4, sy+ny*0.4-0.3, 1.6, cx, cy-0.3, 1.1)...)
	p.Poly(roo, capsule(cx-0.3, cy-0.4, 1.8, mx-0.3, my-0.3, 0.8)...)
	p.Poly(rooLight, capsule(hipX+nx*2.8+ux*1.4, hipY+ny*2.8+uy*1.4, 2.4, sx+nx*1.6, sy+ny*1.6, 1.2)...)
	p.Curve(rooLight, 0.7, hipX-nx*4+ux*2, hipY-ny*4+uy*2, (hipX+sx)/2-nx*3.2, (hipY+sy)/2-ny*3.2, sx-nx*2, sy-ny*2)
	if r.Pouch {
		jx, jy := hipX+nx*4.6+ux*3.4, hipY+ny*4.6+uy*3.4
		p.Circle(jx, jy, 1.5+Line, ink)
		p.Circle(jx, jy, 1.5, roo)
		p.Ellipse(jx-0.6, jy-2, 0.5, 1.3, -0.4, rooShade)
		p.Ellipse(jx+0.3, jy-2.1, 0.5, 1.3, 0.1, roo)
		p.Circle(jx+0.5, jy-0.3, 0.35, ink)
		p.Circle(jx+1.5, jy+0.3, 0.35, ink)
	}
	// The face: the dark nose, the eye with its glint.
	p.Circle(mx+0.6, my-0.2, 0.7, ink)
	ex, ey := along(cx, cy, headA-deg(70), 0.9)
	p.Circle(ex+0.4, ey, 0.5, ink)
	p.Circle(ex+0.25, ey-0.2, 0.18, resurrect64.White)
	ear(0.2, earA+0.08+0.5*r.Flick, roo)

	drawRooLeg(p, hipX, hipY, legs, false)
	arm(0, roo)
}

// drawRooLeg draws a hind leg from its hip at (x, y), posed a, in shade
// if far: the thigh's line in the haunch, the shin and the long foot, the
// toes dark.
func drawRooLeg(p Pen, x, y float64, a legPose, far bool) {
	kx, ky, hx, hy, tx, ty := rooLeg(x, y, a)
	c := roo
	if far {
		c = rooShade
	}
	bone := []float64{kx, ky, hx, hy}
	foot := []float64{hx, hy, tx, ty}
	p.Stroke(ink, 2.6+2*Line, bone...)
	p.Stroke(ink, 1.8+2*Line, foot...)
	p.Stroke(c, 2.6, bone...)
	p.Stroke(c, 1.8, foot...)
	if far {
		p.Stroke(c, 4.4, x, y, kx, ky)
	} else {
		p.Curve(rooShade, 0.6, x+1, y+3.4, kx-0.4, ky-0.6, kx-1.6, ky+0.6)
	}
	p.Ellipse(tx, ty, 1, 0.6, a[2], rooDark)
}

// lerp is a to b, f of the way.
func lerp(a, b, f float64) float64 { return a + (b-a)*f }
