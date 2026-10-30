package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// Weather is a level's weather (DESIGN D34).
type Weather int

const (
	Clear  Weather = iota // a clear blue sky
	Fair                  // light white clouds
	Cloudy                // grey-blue cloud
	Storm                 // dark storm cloud, with no rain
)

// Weathers is every weather, in order.
var Weathers = []Weather{Clear, Fair, Cloudy, Storm}

func (wx Weather) String() string { return [...]string{"clear", "fair", "cloudy", "storm"}[wx] }

// A day is n+1 steps (DESIGN D34): sunrise at step 0, the sun climbing to
// midday and sinking to sunset at step n-1, then night at step n.

// warmth is how near sunrise or sunset step s of n is: 1 there, 0 by the
// third step from it, and 0 at night.
func warmth(s, n int) float64 {
	if s >= n {
		return 0
	}
	return max(0, 1-float64(min(s, n-1-s))/2)
}

// dawn reports whether step s of n is in the first half of the day.
func dawn(s, n int) bool { return 2*s < n-1 }

// sunAt is where the sun is at step s of n: it goes east to west, right to
// left, rising over the far ridge and setting behind the far escarpment.
func sunAt(s, n int) (x, y float64) {
	t := float64(min(s, n-1)) / float64(n-1)
	return 600 - 560*t, 168 - 22*t - 132*math.Sin(math.Pi*t)
}

// Sky draws the sky at step s of n under weather wx, w wide, from the top
// down to the horizon, behind the backdrop.
func Sky(p Pen, w float64, wx Weather, s, n int) {
	if s >= n {
		nightBands(p, w)
		switch wx {
		case Cloudy:
			greyClouds(p, w, 14, resurrect64.Indigo, resurrect64.Navy, resurrect64.Lilac)
		case Storm:
			greyClouds(p, w, 26, resurrect64.Navy, resurrect64.Shadow, resurrect64.Indigo)
		}
		return
	}
	warm := warmth(s, n)
	switch wx {
	case Clear, Fair:
		skyBands(p, w)
		glow(p, w, s, n, warm)
		sunX, sunY := sunAt(s, n)
		sun(p, sunX, sunY, warm, 1)
		if wx == Fair {
			lit, dark := cloudLit, cloudDark
			switch {
			case warm > 0.4 && dawn(s, n):
				lit, dark = resurrect64.Blush, resurrect64.Lavender
			case warm > 0.4:
				lit, dark = resurrect64.Apricot, resurrect64.Lilac
			}
			for _, c := range [][3]float64{{232, 62, 1}, {455, 44, 1.25}, {350, 118, 0.6}, {120, 104, 0.8}} {
				cloud(p, c[0], c[1], c[2], lit, dark)
			}
		}
	case Cloudy:
		moodyBands(p, w)
		glow(p, w, s, n, warm*0.5)
		sunX, sunY := sunAt(s, n)
		sun(p, sunX, sunY, warm, 0.4)
		greyClouds(p, w, 18, resurrect64.Fog, resurrect64.Navy, resurrect64.Mist)
	case Storm:
		stormBands(p, w)
		glow(p, w, s, n, warm*0.25)
		greyClouds(p, w, 30, resurrect64.Lilac, resurrect64.Shadow, resurrect64.Fog)
	}
}

// NightLights draws the moon and the stars under weather wx, w wide, a as
// solid, over the night's tint so they stay bright, and above the
// escarpments: a clear night's stars and Milky Way, a full moon, dimmed
// through cloud, and nothing through a storm.
func NightLights(p Pen, w float64, wx Weather, a float64) {
	if a <= 0 || wx == Storm {
		return
	}
	p = p.Faded(a)
	if wx == Cloudy {
		moon(p.Faded(0.45), 584, 44)
		return
	}
	for i := 0; i < 220; i++ {
		x, y := w*noise(i+9500), 146*math.Pow(noise(i+9600), 1.3)
		r := 0.25 + 0.55*noise(i+9700)*noise(i+9800)
		c := cloudLit
		if noise(i+9900) > 0.8 {
			c = resurrect64.Lemon
		}
		p.Faded(0.5+0.5*noise(i+9950)).Circle(x, y, r, c)
	}
	for i := 0; i < 160; i++ {
		f := float64(i) / 160
		x, y := -20+(w+40)*f, 130-110*f+12*math.Sin(f*7)
		p.Faded(0.012).Ellipse(x, y, 30, 10, -0.4, resurrect64.Lavender)
	}
	moon(p, 584, 44)
}

// Tint is a colour laid over the whole field, A as solid.
type Tint struct {
	C int
	A float64
}

// Light is the tints over the field at step s of n under weather wx
// (DESIGN D34): cool blue moonlight at night, light enough that the
// koalas, the ute and the track stay clear; warm and faint at sunrise and
// sunset; and dulled under cloud.
func Light(wx Weather, s, n int) []Tint {
	if s >= n {
		return []Tint{{resurrect64.Navy, 0.26}, {resurrect64.Cobalt, 0.12}}
	}
	var t []Tint
	if warm := warmth(s, n); warm > 0 {
		c := resurrect64.Orange
		if dawn(s, n) {
			c = resurrect64.Salmon
		}
		t = append(t, Tint{c, 0.14 * warm})
	}
	switch wx {
	case Cloudy:
		t = append(t, Tint{resurrect64.Dusk, 0.12})
	case Storm:
		t = append(t, Tint{resurrect64.Shadow, 0.24})
	}
	return t
}

// skyBands draws the day's sky: deep blue at the top, blending to sky
// blue, then to a pale haze at the horizon.
func skyBands(p Pen, w float64) {
	strip(p, sky, w, 0, Horizon)
	const step = 2.0
	for y := 0.0; y < Horizon; y += step {
		if f := 1 - y/130; f > 0 {
			haze(p, skyHigh, w, y, y+step, 0.75*f*f)
		}
		if f := (y - 140) / (Horizon - 140); f > 0 {
			haze(p, cloudLit, w, y, y+step, 0.55*f*f)
		}
	}
}

// moodyBands is an overcast sky: slate grey-blue, paler and greyer down to
// the horizon.
func moodyBands(p Pen, w float64) {
	strip(p, resurrect64.Fog, w, 0, Horizon)
	for y := 0.0; y < Horizon; y += 2 {
		if f := 1 - y/150; f > 0 {
			haze(p, resurrect64.Indigo, w, y, y+2, 0.75*f)
		}
		if f := (y - 120) / (Horizon - 120); f > 0 {
			haze(p, resurrect64.Mist, w, y, y+2, 0.6*f*f)
		}
	}
}

// stormBands is a storm's sky: dark grey-blue, a little lighter at the
// horizon.
func stormBands(p Pen, w float64) {
	strip(p, resurrect64.Lilac, w, 0, Horizon)
	for y := 0.0; y < Horizon; y += 2 {
		if f := 1 - y/190; f > 0 {
			haze(p, resurrect64.Navy, w, y, y+2, 0.85*f)
		}
		if f := (y - 150) / (Horizon - 150); f > 0 {
			haze(p, resurrect64.Fog, w, y, y+2, 0.4*f*f)
		}
	}
}

// nightBands is a night sky: near black at the top to deep blue at the
// horizon.
func nightBands(p Pen, w float64) {
	strip(p, resurrect64.Navy, w, 0, Horizon)
	for y := 0.0; y < Horizon; y += 2 {
		if f := 1 - y/170; f > 0 {
			haze(p, resurrect64.Ink, w, y, y+2, 0.85*f)
		}
		if f := (y - 150) / (Horizon - 150); f > 0 {
			haze(p, resurrect64.Indigo, w, y, y+2, 0.6*f)
		}
	}
}

// glow is the colours of sunrise or sunset at step s of n, warm as strong:
// pink at sunrise and orange at sunset low down, and deeper blue above.
func glow(p Pen, w float64, s, n int, warm float64) {
	if warm <= 0 {
		return
	}
	low, mid := resurrect64.Orange, resurrect64.Amber
	if dawn(s, n) {
		low, mid = resurrect64.Salmon, resurrect64.Pink
	}
	for y := 0.0; y < Horizon; y += 2 {
		if f := 1 - y/110; f > 0 {
			haze(p, resurrect64.Indigo, w, y, y+2, 0.5*warm*f)
		}
		if f := (y - 50) / (Horizon - 50); f > 0 {
			haze(p, mid, w, y, y+2, 0.45*warm*f)
		}
		if f := (y - 130) / (Horizon - 130); f > 0 {
			haze(p, low, w, y, y+2, 0.7*warm*f*f)
		}
	}
}

// sun draws the sun at (x, y) with its glow, a as solid, golden as it is
// warm, near sunrise or sunset.
func sun(p Pen, x, y, warm, a float64) {
	p = p.Faded(a)
	halo, rim, disc := cloudLit, resurrect64.Lemon, cloudLit
	if warm > 0.4 {
		halo, rim, disc = resurrect64.Amber, resurrect64.Orange, resurrect64.Lemon
	}
	for r := 46.0; r > 16; r -= 6 {
		p.Faded(0.08).Circle(x, y, r, halo)
	}
	p.Faded(0.5).Circle(x, y, 15, rim)
	p.Circle(x, y, 12.5, rim)
	p.Circle(x, y, 10.5, disc)
}

// moon draws a full moon at (x, y), its seas and glow.
func moon(p Pen, x, y float64) {
	for r := 58.0; r > 19; r -= 6 {
		p.Faded(0.06).Circle(x, y, r, resurrect64.Lavender)
	}
	p.Circle(x, y, 17, cloudDark)
	p.Circle(x-1.2, y-1.2, 15.5, cloudLit)
	for _, s := range [][3]float64{{-5, -3, 4.5}, {4.5, 1.5, 3.6}, {-1.5, 6, 3}, {6, -6, 2.1}, {-7, 4, 1.6}} {
		p.Faded(0.35).Circle(x+s[0], y+s[1], s[2], resurrect64.Fog)
	}
}

// cloud draws a white cloud about (x, y), s times full size: round puffs
// over a flat base, lit and shaded underneath dark.
func cloud(p Pen, x, y, s float64, lit, dark int) {
	puffs := [][3]float64{{-22, 2, 9}, {-9, -5, 13}, {7, -9, 15}, {22, -2, 11}, {33, 3, 7}}
	q := func(v float64) float64 { return v * s }
	for _, f := range puffs {
		p.Circle(x+q(f[0]), y+q(f[1])+q(1.6), q(f[2]), dark)
	}
	p.Poly(dark, x-q(28), y+q(4), x+q(38), y+q(4), x+q(36), y+q(10.6), x-q(26), y+q(10.6))
	for _, f := range puffs {
		p.Circle(x+q(f[0]), y+q(f[1]), q(f[2]), lit)
	}
	p.Poly(lit, x-q(27), y+q(3), x+q(36), y+q(3), x+q(34), y+q(8.6), x-q(25), y+q(8.6))
	p.Faded(0.5).Curve(dark, q(1.2), x-q(14), y+q(1), x-q(4), y-q(1.5), x+q(6), y)
}

// greyClouds heaps n soft clouds across the sky, w wide: body their
// colour, under their undersides, top their lit tops.
func greyClouds(p Pen, w float64, n, body, under, top int) {
	for i := 0; i < n; i++ {
		x := -40 + (w+80)*noise(i+9000)
		y := 20 + 110*noise(i+9100)
		s := 0.7 + 0.9*noise(i+9200)
		greyCloud(p, x, y, s, y/130, body, under, top)
	}
}

// greyCloud is a soft cloud about (x, y), s times full size, darker
// underneath higher up (light from 0 high to 1 low): many faint puffs
// heaped up, so the layers build softly, a dark underside first, then
// the body a little higher, then a pale top where light gets through.
func greyCloud(p Pen, x, y, s, light float64, body, under, top int) {
	seed := int(x*7 + y*13)
	for i := 0; i < 16; i++ {
		dx := (noise(seed+i) - 0.5) * 80 * s
		dy := (noise(seed+i+50) - 0.5) * 12 * s
		r := (8 + 10*noise(seed+i+100)) * s
		bulge := 1 - math.Abs(dx)/(45*s)
		p.Faded(0.3-0.12*light).Ellipse(x+dx, y+dy+4*s, r*1.5, r*0.6, 0, under)
		p.Faded(0.35).Ellipse(x+dx, y+dy-bulge*5*s, r*1.3, r*0.75, 0, body)
		p.Faded(0.12+0.1*light).Ellipse(x+dx, y+dy-bulge*9*s, r*0.9, r*0.4, 0, top)
	}
}
