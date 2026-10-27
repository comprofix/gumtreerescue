// Package art draws Gum Tree Rescue's pictures in code (DESIGN §3): each
// figure in layers, an Ink outline, its colour, a shade and a highlight,
// then the details, in curves. Every picture is drawn in field pixels with
// a Pen at any scale, so the same art serves both looks.
package art

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/gumtreerescue/internal/gfx"
	"github.com/comprofix/gumtreerescue/internal/shapes"
)

// Line is how far an outline reaches past a shape, in field pixels.
const Line = 0.8

// Pen draws shapes given in field pixels, about an origin at (OX, OY), at
// K screen pixels to a field pixel. With Flip set it mirrors about the
// origin, so a figure drawn facing right faces left. Sized and Turned
// draw a figure smaller or turned.
type Pen struct {
	Dst    *ebiten.Image
	K      float64
	OX, OY float64
	Flip   bool

	s      float64 // the figure's size, 0 for full size
	sx, sy float64 // how much wider and taller it is drawn, 0 for as it is
	turn   float64 // radians clockwise, about the pivot
	pivot  float64 // how far above the origin the figure turns, at its size
}

// Squashed is the pen drawing figures w times as wide and h times as tall
// about its origin, as when one hits something.
func (p Pen) Squashed(w, h float64) Pen {
	p.sx, p.sy = w, h
	return p
}

// At is the pen with its origin at (x, y) in field pixels.
func (p Pen) At(x, y float64) Pen {
	p.OX, p.OY = x, y
	return p
}

// Mirrored is the pen facing the other way when flip is set.
func (p Pen) Mirrored(flip bool) Pen {
	p.Flip = p.Flip != flip
	return p
}

// Sized is the pen drawing figures s times their size about its origin.
func (p Pen) Sized(s float64) Pen {
	p.s = p.size() * s
	return p
}

// Turned is the pen turning figures by a radians clockwise about the point
// pivot field pixels above the origin, at the figure's size.
func (p Pen) Turned(a, pivot float64) Pen {
	p.turn, p.pivot = p.turn+a, pivot
	return p
}

func (p Pen) size() float64 {
	if p.s == 0 {
		return 1
	}
	return p.s
}

// pt is the screen position of (x, y) about the origin.
func (p Pen) pt(x, y float64) (float32, float32) {
	s := p.size()
	if p.sx != 0 {
		x, y = x*p.sx, y*p.sy
	}
	if p.turn != 0 {
		y += p.pivot
		sn, cs := math.Sincos(p.turn)
		x, y = x*cs-y*sn, x*sn+y*cs
		y -= p.pivot
	}
	x, y = x*s, y*s
	if p.Flip {
		x = -x
	}
	return float32((p.OX + x) * p.K), float32((p.OY + y) * p.K)
}

func (p Pen) path(xy []float64, closed bool) *vector.Path {
	var path vector.Path
	path.MoveTo(p.pt(xy[0], xy[1]))
	for i := 2; i+1 < len(xy); i += 2 {
		path.LineTo(p.pt(xy[i], xy[i+1]))
	}
	if closed {
		path.Close()
	}
	return &path
}

// Poly fills the polygon through the x, y pairs in palette colour c.
func (p Pen) Poly(c int, xy ...float64) {
	if len(xy) >= 6 {
		shapes.FillPath(p.Dst, p.path(xy, true), c)
	}
}

// Blob fills the smooth closed curve through the x, y pairs.
func (p Pen) Blob(c int, xy ...float64) { p.Poly(c, curve(xy, true)...) }

// OutlinedBlob fills the smooth closed curve through the x, y pairs in
// palette colour c, inside an Ink outline Line wide.
func (p Pen) OutlinedBlob(c int, xy ...float64) {
	pts := curve(xy, true)
	p.Stroke(ink, 2*Line, pts...)
	p.Poly(c, pts...)
}

// Ellipse fills an ellipse with radii rx and ry about (x, y), turned by
// tilt radians.
func (p Pen) Ellipse(x, y, rx, ry, tilt float64, c int) {
	const n = 28
	s, co := math.Sin(tilt), math.Cos(tilt)
	pts := make([]float64, 0, 2*n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / n
		ex, ey := rx*math.Cos(t), ry*math.Sin(t)
		pts = append(pts, x+ex*co-ey*s, y+ex*s+ey*co)
	}
	p.Poly(c, pts...)
}

// Circle fills a circle of radius r about (x, y).
func (p Pen) Circle(x, y, r float64, c int) { p.Ellipse(x, y, r, r, 0, c) }

// Stroke draws a line w field pixels wide through the points, rounded at
// its ends and corners.
func (p Pen) Stroke(c int, w float64, xy ...float64) {
	if len(xy) < 4 {
		return
	}
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(gfx.RGBA(c))
	vector.StrokePath(p.Dst, p.path(xy, false), &vector.StrokeOptions{
		Width: float32(w * p.K * p.size()), LineJoin: vector.LineJoinRound, LineCap: vector.LineCapRound,
	}, op)
}

// Curve draws a smooth line w wide through the points.
func (p Pen) Curve(c int, w float64, xy ...float64) { p.Stroke(c, w, curve(xy, false)...) }

// curve is a Catmull-Rom curve through the x, y pairs, closed back to the
// first when closed is set.
func curve(xy []float64, closed bool) []float64 {
	const steps = 6
	m := len(xy) / 2
	if m < 3 {
		return xy
	}
	at := func(i int) (float64, float64) {
		if closed {
			i = (i%m + m) % m
		} else {
			i = min(max(i, 0), m-1)
		}
		return xy[2*i], xy[2*i+1]
	}
	segs := m - 1
	if closed {
		segs = m
	}
	out := make([]float64, 0, 2*(segs*steps+1))
	x, y := at(0)
	out = append(out, x, y)
	for i := range segs {
		x0, y0 := at(i - 1)
		x1, y1 := at(i)
		x2, y2 := at(i + 1)
		x3, y3 := at(i + 2)
		for s := 1; s <= steps; s++ {
			t := float64(s) / steps
			t2, t3 := t*t, t*t*t
			f := func(a, b, c, d float64) float64 {
				return 0.5 * (2*b + (c-a)*t + (2*a-5*b+4*c-d)*t2 + (3*b-a-3*c+d)*t3)
			}
			out = append(out, f(x0, x1, x2, x3), f(y0, y1, y2, y3))
		}
	}
	return out
}
