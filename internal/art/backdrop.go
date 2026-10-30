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
// the gums about the paddock, cows and the golden grass; then BackdropFront
// draws the rest over the kangaroos behind the fence. The sky is left
// clear, for Sky behind it, as it changes through the day (DESIGN D34).
// It is slow to draw, so it is drawn once and kept.
func Backdrop(p Pen, w, ground float64) {
	mesas(p, w)
	paddocks(p, w, ground)
	farm(p)
	for _, g := range BackGums {
		BackGum(p.At(g.X, g.Y).Sized(g.Size).Mirrored(g.Flip), g.Shape)
	}
	for _, c := range Cows {
		DrawCow(p.At(c.X, c.Y).Sized(c.Size).Mirrored(c.Flip), c)
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

// farm draws the farm in the distance on the right (DESIGN D35): the
// windmill with its trough, the shed and the water tank, each with its
// shadow on the grass.
func farm(p Pen) {
	p.Faded(0.4).Poly(grassDeep, 553, 247, 588, 247, 636, 244, 640, 249, 590, 251, 556, 250)
	shed(p)
	tank(p)
	windmill(p)
}

// shed draws the farm shed from its near corner, lit from the left: the
// gable end with its open door, hay inside and a louvred vent; the long wall
// going back in shade, with windows; and the roof's far-reaching slope, its
// iron ribbed and rusting, with a ridge cap and a roof vent.
func shed(p Pen) {
	// The long wall, in shade, darker under the eave.
	p.Poly(ink, 587, 224, 618, 222, 618, 245, 587, 248)
	p.Poly(resurrect64.Fog, 588, 225, 617, 223, 617, 244, 588, 247)
	for x := 590.5; x < 617; x += 2.5 {
		p.Stroke(resurrect64.Lilac, 0.5, x, 226, x, 246)
	}
	p.Poly(resurrect64.Dusk, 588, 225, 617, 223, 617, 226, 588, 229)
	// Two windows on it, showing the sky.
	for _, x := range []float64{593, 605} {
		p.Poly(ink, x-0.7, 231.2, x+7.7, 230.5, x+7.7, 238.5, x-0.7, 239.2)
		p.Poly(resurrect64.Sky, x, 231.8, x+7, 231.2, x+7, 237.9, x, 238.5)
		p.Stroke(ink, 0.6, x+3.5, 231.5, x+3.5, 238.2)
		p.Stroke(ink, 0.6, x, 235.1, x+7, 234.5)
	}
	// The gable end, in the light, its iron ribbed.
	p.Poly(ink, 555, 226, 572, 210, 589, 226, 589, 248, 555, 248)
	p.Poly(resurrect64.Mist, 556, 226, 572, 211, 588, 226, 588, 247, 556, 247)
	for x := 558.0; x < 588; x += 2.4 {
		y := 226 - 15*(1-abs(x-572)/16)
		p.Stroke(resurrect64.Fog, 0.5, x, y+0.6, x, 246.5)
	}
	// Its door open dark, hay stacked inside, framed in timber.
	p.Poly(ink, 561, 229, 583, 229, 583, 247, 561, 247)
	p.Poly(resurrect64.Shadow, 562, 230, 582, 230, 582, 247, 562, 247)
	for i, b := range [][2]float64{{563, 242}, {570, 242}, {566, 237.5}} {
		c := []int{resurrect64.Gold, resurrect64.Honey, resurrect64.Gold}[i]
		p.Poly(resurrect64.Caramel, b[0], b[1], b[0]+7, b[1], b[0]+7, b[1]+4.6, b[0], b[1]+4.6)
		p.Poly(c, b[0], b[1], b[0]+6.4, b[1], b[0]+6.4, b[1]+3.8, b[0], b[1]+3.8)
	}
	p.Stroke(resurrect64.Clay, 1.2, 561, 229.4, 583, 229.4)
	p.Stroke(resurrect64.Clay, 1.2, 561.4, 229.4, 561.4, 247)
	p.Stroke(resurrect64.Clay, 1.2, 582.6, 229.4, 582.6, 247)
	// The louvred vent up in the gable.
	p.Poly(ink, 568, 217, 576, 217, 576, 224, 568, 224)
	for y := 218.2; y < 224; y += 1.6 {
		p.Stroke(resurrect64.Fog, 0.7, 568.6, y, 575.4, y)
	}
	// The roof's slope, going back, ribbed, a little rust where it is old.
	p.Poly(ink, 572, 208.6, 603, 206.6, 621, 223.6, 591, 228)
	p.Poly(resurrect64.Dusk, 572.6, 209.8, 602.6, 207.8, 619.4, 223.2, 590.6, 227)
	for i := 1; i < 11; i++ {
		f := float64(i) / 11
		p.Stroke(resurrect64.Lilac, 0.7, lerp(572.6, 602.6, f), lerp(209.8, 207.8, f), lerp(590.6, 619.4, f), lerp(227, 223.2, f))
	}
	p.Faded(0.55).Poly(resurrect64.Rust, 594, 217, 600, 216.4, 604, 220.8, 597.4, 221.6)
	p.Faded(0.55).Poly(resurrect64.Rust, 583, 214, 587, 213.6, 590, 217, 586, 217.4)
	p.Stroke(resurrect64.Mist, 1.4, 572, 209.2, 603, 207.2)
	// The barge boards along the gable's edge, and the gutter.
	p.Stroke(ink, 2.2, 552.6, 228, 572, 209.4, 591, 228)
	p.Stroke(resurrect64.White, 1.2, 552.8, 227.8, 572, 209.6, 590.8, 227.8)
	p.Stroke(resurrect64.Fog, 1, 591, 227.4, 620, 223.4)
	// The roof vent on the ridge.
	p.Poly(ink, 586, 204, 591, 204, 591.6, 208.6, 585.4, 208.6)
	p.Poly(resurrect64.Fog, 586.8, 204.8, 590.2, 204.8, 590.6, 208, 586.4, 208)
	p.Ellipse(588.5, 204, 3.4, 1.8, 0, ink)
	p.Ellipse(588.5, 204, 2.8, 1.3, 0, resurrect64.Mist)
}

// tank draws the round water tank on its stone base beside the shed, lit
// from the left and shaded round to the right, its ribs curving about it, a
// downpipe from the gutter into its domed top, and a tap at its foot.
func tank(p Pen) {
	const cx, rx, top, foot = 628.0, 9.0, 229.0, 245.0
	p.Ellipse(cx, foot+1.2, rx+2, 2.6, 0, resurrect64.Clay)
	p.Ellipse(cx, foot+0.8, rx+1.4, 2, 0, resurrect64.Sand)
	p.Poly(ink, cx-rx-0.8, top, cx+rx+0.8, top, cx+rx+0.8, foot, cx-rx-0.8, foot)
	p.Ellipse(cx, foot, rx+0.8, 2.3, 0, ink)
	// The wall, in vertical bands from lit to shade.
	bands := []struct {
		f0, f1 float64
		c      int
	}{{-1, -0.35, resurrect64.Mist}, {-0.35, 0.3, resurrect64.Fog}, {0.3, 0.75, resurrect64.Lilac}, {0.75, 1, resurrect64.Dusk}}
	for _, b := range bands {
		p.Poly(b.c, cx+b.f0*rx, top, cx+b.f1*rx, top, cx+b.f1*rx, foot, cx+b.f0*rx, foot)
	}
	p.Ellipse(cx, foot, rx, 1.6, 0, resurrect64.Lilac)
	// The ribs curve round it, as the eye looks down on them.
	for y := top + 3.2; y < foot; y += 3.4 {
		p.Curve(resurrect64.Dusk, 0.8, cx-rx, y, cx-rx*0.5, y+1, cx, y+1.3, cx+rx*0.5, y+1, cx+rx, y)
	}
	// The domed top, and the pipe from the gutter.
	p.Ellipse(cx, top, rx+0.8, 3.4, 0, ink)
	p.Ellipse(cx, top, rx, 2.7, 0, resurrect64.Fog)
	p.Poly(resurrect64.Mist, cx-rx+1, top-0.2, cx, top-2.8, cx+2, top-1, cx-3, top+1.6)
	p.Stroke(ink, 1.6, 620, 224, 622, 227, 625, 227.6)
	p.Stroke(resurrect64.Fog, 0.8, 620, 224, 622, 227, 625, 227.6)
	// The tap.
	p.Stroke(ink, 1.2, cx-rx-0.4, foot-3, cx-rx-2.4, foot-3, cx-rx-2.4, foot-1.2)
}

// windmill draws the windmill and its trough: a four-legged lattice tower on
// concrete feet, its far legs paler, braced and with a ladder up to a
// platform; the wheel a little side on, its blades lit in turn, with its rim
// and hub, the gearbox and the tail vane behind; and the pipe to the trough.
func windmill(p Pen) {
	const cx, top, foot, plat = 534.0, 202.0, 246.0, 209.0
	leg := func(fx, fy, tx float64, c int, w float64) { p.Stroke(c, w, cx+fx, foot+fy, cx+tx, plat) }
	// The far legs and their braces.
	leg(-4.5, -1.6, -1, resurrect64.Fog, 1)
	leg(5.5, -1.6, 1.6, resurrect64.Fog, 1)
	// The near legs' braces, level and crossed, tier by tier.
	const tiers = 4
	half := func(f, foot float64) float64 { return foot + (1.8-foot)*f }
	y := func(f float64) float64 { return foot + (plat-foot)*f }
	for i := range tiers {
		f, g := float64(i)/tiers, float64(i+1)/tiers
		w0, w1 := half(f, 9), half(g, 9)
		p.Stroke(resurrect64.Dusk, 1, cx-w0, y(f), cx+w1, y(g))
		p.Stroke(resurrect64.Dusk, 1, cx+w0, y(f), cx-w1, y(g))
		p.Stroke(resurrect64.Dusk, 1, cx-w1, y(g), cx+w1, y(g))
	}
	// The pipe down the middle, to the trough.
	p.Stroke(resurrect64.Mist, 1, cx+0.4, plat+1, cx+0.4, foot-2, cx+8, foot-2)
	// The near legs, on their feet.
	for _, s := range []float64{-1, 1} {
		leg(s*9, 0, s*1.8, ink, 2)
		leg(s*9, 0, s*1.8, resurrect64.Clay, 1.2)
		p.Poly(ink, cx+s*9-2.4, foot-0.8, cx+s*9+2.4, foot-0.8, cx+s*9+2.6, foot+1.4, cx+s*9-2.6, foot+1.4)
		p.Poly(resurrect64.Mist, cx+s*9-1.8, foot-0.2, cx+s*9+1.8, foot-0.2, cx+s*9+2, foot+0.8, cx+s*9-2, foot+0.8)
	}
	// The ladder up the right leg.
	for f := 0.1; f < 0.95; f += 0.16 {
		x := cx + half(f, 9)
		p.Stroke(resurrect64.Mist, 0.8, x-1.4, y(f), x+0.6, y(f))
	}
	// The platform at the tower's top.
	p.Poly(ink, cx-5.4, plat-1.6, cx+5.4, plat-1.6, cx+5.4, plat+1, cx-5.4, plat+1)
	p.Poly(resurrect64.Clay, cx-4.8, plat-1, cx+4.8, plat-1, cx+4.8, plat+0.4, cx-4.8, plat+0.4)
	// The tail vane on its boom, and the gearbox.
	p.Poly(ink, cx+5, top-1.8, cx+20.6, top-6.4, cx+20.6, top+3.4, cx+5, top+1.8)
	p.Poly(resurrect64.Mist, cx+6, top-1.1, cx+19.8, top-5.2, cx+19.8, top+2.4, cx+6, top+1.1)
	p.Poly(resurrect64.Rust, cx+13, top-3.3, cx+16, top-4.2, cx+16, top+1.5, cx+13, top+0.9)
	p.Poly(ink, cx-1.8, top-2.4, cx+6, top-2.4, cx+6, top+3.2, cx-1.8, top+3.2)
	p.Poly(resurrect64.Dusk, cx-1.2, top-1.8, cx+5.4, top-1.8, cx+5.4, top+2.6, cx-1.2, top+2.6)
	p.Stroke(resurrect64.Dusk, 1.4, cx+1, top+3.2, cx+1, plat-1.6)
	// The wheel, a little side on: its shadow behind, then the blades.
	const squash = 0.86
	wheel := func(r, a float64) (float64, float64) {
		s, c := math.Sincos(a)
		return cx + c*r*squash, top + s*r
	}
	p.Ellipse(cx+1.6, top+0.4, 12.4*squash, 12.4, 0, resurrect64.Shadow)
	const blades = 12
	for i := range blades {
		a := float64(i) * 2 * math.Pi / blades
		x0, y0 := wheel(3.4, a-0.12)
		x1, y1 := wheel(12, a-0.22)
		x2, y2 := wheel(12, a+0.22)
		x3, y3 := wheel(3.4, a+0.12)
		p.Poly(ink, x0, y0, x1, y1, x2, y2, x3, y3)
		c := resurrect64.Mist
		switch {
		case i%2 == 1:
			c = resurrect64.Fog
		case math.Cos(a)+math.Sin(a) < -0.4: // the top left catches the light
			c = resurrect64.White
		}
		x0, y0 = wheel(3.8, a-0.08)
		x1, y1 = wheel(11.5, a-0.18)
		x2, y2 = wheel(11.5, a+0.18)
		x3, y3 = wheel(3.8, a+0.08)
		p.Poly(c, x0, y0, x1, y1, x2, y2, x3, y3)
	}
	// The rims that hold the blades.
	ring := func(r float64, c int, w float64) {
		var xy []float64
		for i := 0; i <= 36; i++ {
			x, y := wheel(r, float64(i)*math.Pi/18)
			xy = append(xy, x, y)
		}
		p.Stroke(c, w, xy...)
	}
	ring(12.2, ink, 0.9)
	ring(7.6, resurrect64.Dusk, 0.6)
	p.Circle(cx, top, 3+Line, ink)
	p.Circle(cx, top, 3, resurrect64.Dusk)
	p.Circle(cx-0.7, top-0.7, 1.2, resurrect64.Fog)
	// The trough at its foot, its water showing.
	const tx, ty = 542.0, 245.0
	p.Poly(ink, tx-0.8, ty-0.8, tx+11.8, ty-0.8, tx+11.8, ty+4, tx-0.8, ty+4)
	p.Poly(resurrect64.Fog, tx, ty, tx+11, ty, tx+11, ty+3.2, tx, ty+3.2)
	p.Poly(resurrect64.Sky, tx+0.6, ty+0.3, tx+10.4, ty+0.3, tx+10.4, ty+1.2, tx+0.6, ty+1.2)
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
