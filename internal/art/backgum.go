package art

import "math"

// A background gum is one of the gums about the paddock that make the farm
// the koalas' home (DESIGN §3, D27): smaller with distance, standing
// behind the fence, apart from each other so the paddocks, the farm and
// the ranges still show between them.
type Gum struct {
	X, Y  float64 // the foot of its trunk, in field pixels
	Size  float64 // how big it is beside a gum of full size, about 70 tall
	Shape int     // which of gumShapes it grows in
	Flip  bool    // mirrored, so no two look the same
}

// BackGums are the gums about the paddock: a small one in the farmland behind
// the big gum, one tall and one small to its right, one in the middle and
// one short of the windmill.
var BackGums = []Gum{
	{X: 150, Y: 246, Size: 0.75, Shape: 1},
	{X: 236, Y: 247, Size: 1.05, Shape: 0},
	{X: 292, Y: 246, Size: 0.7, Shape: 1, Flip: true},
	{X: 452, Y: 245, Size: 0.9, Shape: 1},
	{X: 500, Y: 244, Size: 0.6, Shape: 0, Flip: true},
}

// gumShape is how a background gum grows: its trunk and limbs, from the
// foot up, and its clumps of leaves, x, y and radius, about the foot.
type gumShape struct {
	limbs  []limb
	clumps [][3]float64
}

// gumShapes are the two ways the background gums grow: tall and spreading,
// its crown in three, or leaning, its crown in two.
var gumShapes = []gumShape{
	{
		limbs: []limb{
			{5, 3, []float64{0, 0, 1, -18, -1, -30}},
			{3, 1.2, []float64{-1, -28, -8, -40, -16, -50}},
			{3.2, 1.2, []float64{0, -30, 8, -42, 18, -48}},
			{2.6, 1, []float64{-1, -30, 2, -46, 0, -60}},
		},
		clumps: [][3]float64{{-17, -52, 9}, {-8, -59, 9}, {2, -64, 10}, {12, -58, 9}, {20, -50, 8}, {-3, -50, 7}, {9, -48, 6}},
	},
	{
		limbs: []limb{
			{4.6, 2.6, []float64{0, 0, -2, -16, 2, -28}},
			{3, 1, []float64{2, -26, 12, -36, 16, -50}},
			{2.6, 1, []float64{1, -26, -6, -38, -8, -48}},
		},
		clumps: [][3]float64{{16, -54, 9}, {6, -59, 8}, {-7, -52, 8}, {-12, -44, 6}, {23, -46, 6}, {2, -47, 6}},
	},
}

// BackGum draws a background gum in shape s with the foot of its trunk at
// the pen's origin, in the big gum's style but further off: its shadow on
// the grass, the pale trunk forking into limbs, and open clumps of leaves,
// the sky through the gaps, all outlined softly.
func BackGum(p Pen, s int) {
	g := gumShapes[s]
	soft := p.Faded(0.55)
	p.Faded(0.35).Ellipse(2, 0.4, 16, 2, 0, grassDeep)
	for _, l := range g.limbs {
		taper(soft, ink, l.w0+2*Line, l.w1+2*Line, l.pts)
	}
	for _, l := range g.limbs {
		taper(p, barkShade, l.w0, l.w1, l.pts)
		taper(p, bark, l.w0*0.55, l.w1*0.55, shift(l.pts, -l.w0*0.15, 0))
	}
	p.Stroke(barkLight, 0.8, -1.4, -2, -1, -16, -2, -26)
	for _, c := range g.clumps {
		soft.Circle(c[0], c[1], c[2]+Line, ink)
	}
	for _, c := range g.clumps {
		p.Circle(c[0], c[1], c[2], leafShade)
	}
	for _, c := range g.clumps {
		p.Circle(c[0]-c[2]*0.12, c[1]-c[2]*0.16, c[2]*0.8, leaf)
	}
	for _, c := range g.clumps {
		// Leaves across the clump, lit to the top left, and long ones
		// hanging below.
		for j := range 7 {
			a := float64(j) * 2.4
			r := c[2] * 0.7 * math.Sqrt((float64(j)+0.5)/7)
			x, y := c[0]+r*math.Cos(a), c[1]+r*math.Sin(a)
			col := leafShade
			if x-c[0]+y-c[1] < -c[2]*0.2 {
				col = leafLight
			}
			p.Ellipse(x, y, 0.9, 2.4, a, col)
		}
		for i, a := range []float64{-0.5, 0.1, 0.6} {
			x := c[0] + a*c[2]
			y := c[1] + c[2]*0.8 - abs(a)*c[2]*0.3
			soft.Ellipse(x, y+2.4, 1.1, 3.2, a*0.4, ink)
			p.Ellipse(x, y+2.4, 0.7, 2.8, a*0.4, []int{leafShade, leaf}[i%2])
		}
	}
}
