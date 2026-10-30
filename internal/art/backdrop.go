package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The backdrop's layout, in field pixels from the top (DESIGN §3): where
// the paddocks meet the escarpments, and the fence's posts and rails.
const (
	Horizon   = 232
	FenceTop  = 250 // the tops of the posts
	RailHigh  = 256 // the top rail's top
	RailLow   = 266 // the bottom rail's top
	FenceFoot = 278
)

// The backdrop's colours (DESIGN §3): a blue sky hazing pale at the
// horizon, blue haze on the far hills, golden paddocks, weathered grey
// timber and red dirt.
const (
	sky       = resurrect64.Sky
	skyHigh   = resurrect64.Blue
	cloudLit  = resurrect64.White
	cloudDark = resurrect64.Mist
	rangeFar  = resurrect64.Blue
	bush      = resurrect64.Sage
	bushDark  = resurrect64.Slate
	grass     = resurrect64.Honey
	grassDry  = resurrect64.Gold
	grassDeep = resurrect64.Caramel
	grassTip  = resurrect64.Lemon
	paddock   = resurrect64.Lime
	timber    = resurrect64.Sand
	timberDim = resurrect64.Clay
	timberLit = resurrect64.Mist
	grain     = resurrect64.Mauve
	dirt      = resurrect64.Copper
	dirtDark  = resurrect64.Rust
	dirtLit   = resurrect64.Caramel
)

// Backdrop draws the country bush behind the play, w by h field pixels
// with the ground's top at ground (DESIGN §3): the escarpments, the
// paddocks with a farm shed, a tank and a windmill, the koalas' home in
// the gums about the paddock and the golden grass; then BackdropFront
// draws the rest over the kangaroos behind the fence. The sky is left
// clear, for Sky behind it, as it changes through the day (DESIGN D34).
// It is slow to draw, so it is drawn once and kept.
func Backdrop(p Pen, w, ground float64) {
	mesas(p, w)
	paddocks(p, w, ground)
	farm(p.Faded(0.9))
	for _, g := range BackGums {
		BackGum(p.At(g.X, g.Y).Sized(g.Size).Mirrored(g.Flip), g.Shape)
	}
}

// BackdropFront draws the near part of the backdrop, over what is behind
// the fence, w by h field pixels with the ground's top at ground: the
// post-and-rail fence, the tufts of the near grass and the red dirt track.
func BackdropFront(p Pen, w, h, ground float64) {
	fence(p, w)
	tufts(p, w, ground)
	track(p, w, h, ground)
}

// noise is a fixed scatter from 0 to 1 for each i, the same every time,
// so the backdrop never takes from the game's random numbers.
func noise(i int) float64 {
	x := uint32(i)*2654435761 ^ 0x9e3779b9
	x ^= x >> 15
	x *= 2246822519
	x ^= x >> 13
	return float64(x%10000) / 10000
}

// strip fills the strip from y0 to y1 right across, w wide.
func strip(p Pen, c int, w, y0, y1 float64) { p.Poly(c, 0, y0, w, y0, w, y1, 0, y1) }

// haze fills the strip from y0 to y1 a as solid, from 0 to 1, so the
// colour blends over what is there; at 0 or less it draws nothing.
func haze(p Pen, c int, w, y0, y1, a float64) {
	if a > 0 {
		strip(p.Faded(a), c, w, y0, y1)
	}
}

// ridge fills from the smooth line through the x, y pairs down to bottom.
func ridge(p Pen, c int, bottom float64, xy ...float64) {
	pts := curve(xy, false)
	pts = append(pts, pts[len(pts)-2], bottom, pts[0], bottom)
	p.Poly(c, pts...)
}

// paddocks draws the fields from the horizon to the ground: a patchwork
// far off, with lines of bush between, then golden grass, deeper as it
// comes near.
func paddocks(p Pen, w, ground float64) {
	strip(p, grass, w, Horizon, ground)
	// The far patchwork: fields in turn green and gold, edged in bush.
	x := 0.0
	for i := 0; x < w; i++ {
		fw := 50 + 60*noise(i+200)
		c := paddock
		if i%2 == 1 {
			c = grassDry
		}
		depth := 9 + 6*noise(i+300)
		p.Faded(0.75).Poly(c, x, Horizon, x+fw, Horizon, x+fw+8, Horizon+depth, x+6, Horizon+depth)
		p.Faded(0.6).Stroke(bush, 1.2, x+fw, Horizon, x+fw+8, Horizon+depth)
		x += fw
	}
	// A darker, deeper gold coming near, in soft bands.
	for y := FenceFoot - 10.0; y < ground; y += 2 {
		haze(p, grassDeep, w, y, y+2, 0.45*(y-FenceFoot+10)/(ground-FenceFoot+10))
	}
	// Mown stripes and light across the near paddock.
	for i := 0; i < 5; i++ {
		y := FenceFoot + 4 + float64(i)*10
		p.Faded(0.18).Curve(grassTip, 2.4, -10, y+3, w*0.4, y, w+10, y+4)
	}
}

// farm draws the farm in the distance on the right: a shed with a
// corrugated roof and a water tank beside it, and a windmill.
func farm(p Pen) {
	// The shed, its wall and roof ribbed, its door open dark.
	p.Poly(ink, 557, 223, 609, 223, 609, 245, 557, 245)
	p.Poly(resurrect64.Fog, 558, 224, 608, 224, 608, 244, 558, 244)
	for x := 561.0; x < 607; x += 3 {
		p.Stroke(resurrect64.Lilac, 0.5, x, 225, x, 243)
	}
	p.Poly(resurrect64.Shadow, 572, 231, 584, 231, 584, 244, 572, 244)
	p.Poly(ink, 552, 225, 563, 211, 603, 211, 614, 225)
	p.Poly(resurrect64.Dusk, 554, 224.2, 564, 212, 602, 212, 612, 224.2)
	for x := 566.0; x < 601; x += 2.6 {
		p.Stroke(resurrect64.Lilac, 0.5, x, 212.6, x-2.4, 223.6)
	}
	p.Stroke(resurrect64.Mist, 0.8, 564, 212.4, 602, 212.4)
	// The round tank, ribbed.
	p.Poly(ink, 615, 229, 631, 229, 631, 245, 615, 245)
	p.Poly(resurrect64.Mist, 616, 230, 630, 230, 630, 244, 616, 244)
	for y := 232.0; y < 244; y += 2.4 {
		p.Stroke(resurrect64.Fog, 0.6, 616, y, 630, y)
	}
	p.Poly(resurrect64.Fog, 615, 229, 623, 226, 631, 229)
	// The windmill: a lattice tower, the wheel of blades and its tail.
	top, foot := 202.0, 246.0
	for _, s := range []float64{-1, 1} {
		p.Stroke(resurrect64.Dusk, 1.1, 534+s*8, foot, 534+s*1.4, top+4)
	}
	for i := 0; i < 5; i++ {
		f := float64(i) / 5
		g := float64(i+1) / 5
		y0, y1 := foot-(foot-top-4)*f, foot-(foot-top-4)*g
		w0, w1 := 8-6.6*f, 8-6.6*g
		p.Stroke(resurrect64.Dusk, 0.5, 534-w0, y0, 534+w1, y1)
		p.Stroke(resurrect64.Dusk, 0.5, 534+w0, y0, 534-w1, y1)
	}
	p.Poly(resurrect64.Fog, 537, top-1.5, 552, top-5, 552, top+3, 537, top+1.5)
	p.Stroke(resurrect64.Dusk, 0.8, 534, top, 548, top-1)
	for i := 0; i < 18; i++ {
		s, c := math.Sincos(float64(i) * math.Pi / 9)
		p.Stroke(resurrect64.Mist, 1.2, 534+c*3, top+s*3, 534+c*11, top+s*11)
	}
	p.Circle(534, top, 11.5, resurrect64.Ink)
	p.Faded(0.9).Circle(534, top, 11, resurrect64.Fog)
	for i := 0; i < 18; i++ {
		s, c := math.Sincos(float64(i)*math.Pi/9 + 0.08)
		p.Stroke(resurrect64.Mist, 1.1, 534+c*2.6, top+s*2.6, 534+c*10.6, top+s*10.6)
	}
	p.Circle(534, top, 2.4, resurrect64.Dusk)
}

// fence draws the post-and-rail fence across the paddock, weathered grey
// timber with grain along the rails and down the posts.
func fence(p Pen, w float64) {
	rail := func(y float64, seed int) {
		p.Poly(ink, 0, y-0.7, w, y-0.7, w, y+3.7, 0, y+3.7)
		p.Poly(timber, 0, y, w, y, w, y+3, 0, y+3)
		p.Poly(timberDim, 0, y+2, w, y+2, w, y+3, 0, y+3)
		p.Stroke(timberLit, 0.5, 0, y+0.4, w, y+0.4)
		for i := 0; i < 60; i++ {
			x := float64(i)*w/60 + 8*noise(seed+i)
			l := 4 + 8*noise(seed+i+500)
			gy := y + 0.9 + 1.2*noise(seed+i+900)
			p.Faded(0.7).Stroke(grain, 0.35, x, gy, x+l, gy+0.2)
		}
	}
	for x := 10.0; x < w; x += 40 {
		p.Poly(ink, x-0.7, FenceTop-0.4, x+4.7, FenceTop-1.4, x+4.7, FenceFoot, x-0.7, FenceFoot)
		p.Poly(timber, x, FenceTop+0.4, x+4, FenceTop-0.6, x+4, FenceFoot, x, FenceFoot)
		p.Poly(timberDim, x+2.6, FenceTop-0.3, x+4, FenceTop-0.6, x+4, FenceFoot, x+2.6, FenceFoot)
		for g := 0; g < 2; g++ {
			gx := x + 0.9 + 0.9*float64(g) + 0.4*noise(int(x)+g)
			p.Faded(0.7).Stroke(grain, 0.35, gx, FenceTop+2, gx+0.3, FenceFoot-2)
		}
	}
	rail(RailHigh, 1000)
	rail(RailLow, 2000)
}

// tufts draws grass tufts over the near paddock, more and taller coming
// near, and along the fence's foot.
func tufts(p Pen, w, ground float64) {
	for i := 0; i < 260; i++ {
		f := noise(i + 3000)
		y := FenceFoot + (ground-FenceFoot)*math.Sqrt(f)
		x := w * noise(i+4000)
		tall := 2 + 3.5*f
		c := []int{grassDry, grassDeep, resurrect64.Olive, grassTip}[i%4]
		for b := -1.0; b <= 1; b++ {
			p.Faded(0.85).Stroke(c, 0.6, x+b*0.8, y, x+b*1.8, y-tall*(1-0.25*math.Abs(b)))
		}
	}
	for i := 0; i < 120; i++ {
		x := w * noise(i+5000)
		p.Faded(0.8).Stroke(resurrect64.Olive, 0.6, x, FenceFoot+1, x+0.6, FenceFoot-2-2*noise(i+5100))
	}
}

// track draws the red dirt track from the ground's top: a lip of grass,
// the ruts the tyres have worn, pale patches and pebbles.
func track(p Pen, w, h, ground float64) {
	strip(p, dirt, w, ground, h)
	// The lip, the grass's edge ragged over it.
	p.Poly(dirtLit, 0, ground, w, ground, w, ground+2.4, 0, ground+2.4)
	p.Poly(dirtDark, 0, ground+2.4, w, ground+2.4, w, ground+3.4, 0, ground+3.4)
	for i := 0; i < 160; i++ {
		x := w * noise(i+6000)
		p.Stroke(grassDeep, 0.7, x, ground-1.5, x+0.8*(noise(i+6100)-0.5)*4, ground+1.4+1.6*noise(i+6200))
	}
	// The ruts the tyres have worn, and pale dust between them.
	for _, y := range []float64{ground + 7, ground + 22} {
		var rut []float64
		for x := -10.0; x <= w+10; x += 40 {
			rut = append(rut, x, y+1.4*(noise(int(x)+int(y)*7)-0.5))
		}
		p.Faded(0.55).Curve(dirtDark, 3.4, rut...)
		p.Faded(0.4).Curve(dirtLit, 0.8, shift(rut, 0, -2.2)...)
	}
	for i := 0; i < 14; i++ {
		x, y := w*noise(i+7000), ground+11+8*noise(i+7100)
		p.Faded(0.25).Ellipse(x, y, 14+16*noise(i+7200), 2.2, 0, dirtLit)
	}
	// Pebbles, each with its shadow.
	for i := 0; i < 70; i++ {
		x, y := w*noise(i+8000), ground+5+(h-ground-6)*noise(i+8100)
		r := 0.5 + 0.9*noise(i+8200)
		c := []int{resurrect64.Sand, resurrect64.Clay, dirtLit}[i%3]
		p.Faded(0.5).Ellipse(x+0.4, y+0.5, r*1.3, r*0.7, 0, dirtDark)
		p.Ellipse(x, y, r*1.2, r*0.8, 0, c)
	}
}
