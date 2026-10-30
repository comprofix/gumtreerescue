package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// A sandstone escarpment's colours (DESIGN D33): pale orange-pink cliffs,
// lit and shaded, dark bush capping the top, and forested slopes below.
const (
	cliff      = resurrect64.Apricot
	cliffLit   = resurrect64.Blush
	cliffShade = resurrect64.Clay
	cliffDeep  = resurrect64.Mauve
	capBush    = resurrect64.Pine
	slope      = resurrect64.Sage
	slopeDark  = resurrect64.Slate
	slopeLit   = resurrect64.Fern
)

// mesas draws the sandstone escarpments along the horizon (DESIGN D33): a
// far one on the left hazed blue by the air, a low ridge far off on the
// right, and a nearer one, each a flat top capped with bush over a band of
// sheer cliffs, then steep forested slopes down to the paddocks.
func mesas(p Pen, w float64) {
	ridge(p.Faded(0.5), rangeFar, Horizon, 480, Horizon, 560, 184, 640, 174, w+10, 172)
	escarpment(p, []float64{-10, 156, 40, 151, 120, 153, 200, 150, 280, 154, 330, 158}, 330, 400, 8, 0.6)
	escarpment(p, []float64{228, 172, 260, 167, 320, 165, 390, 166, 450, 163, 510, 167, 548, 171}, 548, 650, 12, 0.1)
	// Haze at their feet, and the bush along it.
	for y := 218.0; y < Horizon; y += 2 {
		haze(p, cloudLit, w, y, y+2, 0.12*(y-218)/14)
	}
	for i := 0; i < 110; i++ {
		x := float64(i)*w/110 + 6*noise(i)
		r := 2 + 3*noise(i+100)
		p.Faded(0.8).Ellipse(x, Horizon-r*0.5, r*1.3, r, 0, bushDark)
		p.Faded(0.8).Ellipse(x-0.4, Horizon-r*0.6-0.4, r*1.1, r*0.8, 0, bush)
	}
}

// escarpment draws one sandstone escarpment: its flat top along the x, y
// pairs, a cliff band cliffH deep under the bush capping it, a steep left
// end, a long slope down from its right end at x1 to the horizon by x2,
// and forested slopes, all hazed hz towards the sky.
func escarpment(p Pen, top []float64, x1, x2, cliffH, hz float64) {
	x0, y0 := top[0], top[1]
	// The outline: up the steep left end, along the top, then the long slope
	// down from its right end.
	edge := curve(top, false)
	y1 := top[len(top)-1]
	sil := []float64{x0 - 34, Horizon, x0 - 22, y0 + 14, x0 - 8, y0 + 5}
	sil = append(sil, edge...)
	sil = append(sil, x1+8, y1+6, (x1+x2)/2, y1+14, x2, Horizon-6, x2+20, Horizon)
	p.Poly(slope, sil...)
	at := func(x float64) float64 { return lerpTop(edge, x) }
	// The forested slopes: darker down the gullies, then clumps of trees,
	// lit and shaded.
	for g := x0 + 14; g < x1-6; g += 18 + 14*noise(int(g)) {
		top := at(g) + cliffH + 1
		p.Faded(0.35).Poly(slopeDark, g-0.5, top, g+0.8, top, g+3, Horizon-4, g-2.5, Horizon-4)
	}
	for i := 0; i < 2600; i++ {
		x := x0 - 30 + (x2-x0+40)*noise(i+300)
		y := Horizon - (Horizon-y0)*noise(i+5300)
		if !inside(sil, x, y) || y < at(x)+cliffH+1.6 && x > x0-4 && x < x1+4 {
			continue
		}
		c := slope
		switch n := noise(i + 2300); {
		case n < 0.3:
			c = slopeDark
		case n > 0.85:
			c = slopeLit
		}
		r := 0.6 + 0.7*noise(i+3300)
		p.Faded(0.8).Ellipse(x, y, r*1.2, r, 0, c)
	}
	// The cliff band under the bush along the top: pale sandstone, its
	// faces lit and shaded in vertical bands, with dark cracks.
	var band []float64
	for x := x0 - 6; x <= x1+4; x++ {
		band = append(band, x, at(x)+1.2)
	}
	for x := x1 + 4; x >= x0-6; x-- {
		d := cliffH * (0.8 + 0.4*noise(int(x)+500))
		band = append(band, x, at(x)+1.2+d)
	}
	p.Poly(cliff, band...)
	for x := x0 - 4; x < x1+3; x += 1.5 + 2.5*noise(int(x*3)) {
		t := at(x) + 1.4
		h := cliffH * (0.5 + 0.5*noise(int(x*7)))
		c := cliffShade
		if noise(int(x*5)) > 0.55 {
			c = cliffLit
		}
		p.Stroke(c, 0.7+0.6*noise(int(x*11)), x, t, x+0.3, t+h)
		if noise(int(x*13)) > 0.8 {
			p.Stroke(cliffDeep, 0.6, x+0.8, t, x+0.9, t+cliffH*0.9)
		}
	}
	// The overhang's shadow under the cap, then the bush capping the top.
	for x := x0 - 6; x < x1+4; x += 2 {
		p.Faded(0.6).Stroke(cliffDeep, 0.6, x, at(x)+1.5, x+2, at(x+2)+1.5)
	}
	for x := x0 - 6; x < x1+5; x += 1.4 {
		r := 0.8 + 0.7*noise(int(x*17))
		p.Circle(x, at(x)+0.6-r*0.4, r, capBush)
	}
	// Hazed by the air between.
	if hz > 0 {
		p.Faded(hz*0.6).Poly(rangeFar, sil...)
		p.Faded(hz*0.5).Poly(sky, sil...)
	}
}

// lerpTop is the y of the line through the x, y pairs at x, held level
// past its ends.
func lerpTop(xy []float64, x float64) float64 {
	if x <= xy[0] {
		return xy[1]
	}
	for i := 2; i+1 < len(xy); i += 2 {
		if x <= xy[i] {
			x0, y0, x1, y1 := xy[i-2], xy[i-1], xy[i], xy[i+1]
			return y0 + (y1-y0)*(x-x0)/math.Max(x1-x0, 1e-9)
		}
	}
	return xy[len(xy)-1]
}

// inside reports whether (x, y) is inside the polygon through the pairs.
func inside(xy []float64, x, y float64) bool {
	in := false
	n := len(xy) / 2
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi, xj, yj := xy[2*i], xy[2*i+1], xy[2*j], xy[2*j+1]
		if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
			in = !in
		}
	}
	return in
}
