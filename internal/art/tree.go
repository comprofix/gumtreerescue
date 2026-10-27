package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The gum tree's colours (DESIGN §3): smooth bark, sandy and shaded clay,
// lit cream, in patches of grey and apricot; grey-green leaves, shaded
// slate and lit fern.
const (
	bark      = resurrect64.Sand
	barkShade = resurrect64.Clay
	barkLight = resurrect64.Blush
	leaf      = resurrect64.Sage
	leafShade = resurrect64.Slate
	leafLight = resurrect64.Fern
)

// limb is a branch of the tree: a curve through its points, tapering from
// w0 wide at the first to w1 at the last, in field pixels.
type limb struct {
	w0, w1 float64
	pts    []float64
}

// The tree's limbs, from the trunk up into the canopy, and a short bough
// to the left.
var limbs = []limb{
	{13, 5, []float64{30, 160, 22, 112, 6, 72, -8, 44}},
	{12, 4, []float64{44, 160, 68, 118, 98, 86, 128, 62}},
	{10, 4, []float64{34, 130, 40, 80, 48, 26}},
	{6, 2.5, []float64{28, 128, 14, 108, 0, 96}},
}

// clumps are the canopy's clumps of leaves: x, y and radius.
var clumps = [][3]float64{
	{-6, 50, 22}, {16, 28, 24}, {46, 12, 22}, {78, 24, 22}, {108, 42, 20},
	{134, 60, 15}, {30, 60, 17}, {64, 50, 18}, {96, 70, 13}, {-10, 86, 15},
	{4, 70, 14}, {120, 38, 14}, {58, 28, 16}, {-14, 28, 16}, {30, 36, 15},
	{90, 52, 14},
}

// trunk is the trunk's outline, flaring at its roots; its right side is
// the field's TrunkX at the ground, where koalas grab and climb.
var trunk = []float64{
	-4, 334, 12, 322, 18, 286, 20, 236, 22, 186, 22, 150, 20, 120,
	30, 128, 38, 136, 44, 150, 50, 186, 54, 236, 58, 286, 62, 320, 78, 334,
}

// trunkLit is the lit part of the trunk: all but a band of shade down its
// right side.
var trunkLit = []float64{
	-2, 333, 13, 322, 19, 286, 21, 236, 23, 186, 23, 150, 22, 124,
	30, 132, 36, 142, 40, 156, 44, 190, 47, 236, 50, 286, 53, 318, 64, 333,
}

// Tree draws the gum tree in field coordinates, with the pen's origin at
// the field's: the trunk and limbs, the canopy, short boughs out over the
// leaves, and the branches out to the drop points. Each bough and branch is
// the x, y pairs along its middle from where it leaves the tree to its
// tip.
func Tree(p Pen, branches, boughs [][]float64) {
	for _, l := range limbs {
		taper(p, ink, l.w0+2*Line, l.w1+2*Line, l.pts)
	}
	for _, l := range limbs {
		taper(p, barkShade, l.w0, l.w1, l.pts)
		taper(p, bark, l.w0*0.6, l.w1*0.6, shift(l.pts, -l.w0*0.15, -l.w0*0.1))
	}
	p.OutlinedBlob(barkShade, trunk...)
	p.Blob(bark, trunkLit...)
	// The light down the trunk's left side.
	p.Blob(barkLight, 15, 320, 20, 290, 22, 240, 24, 190, 24, 150, 27, 150,
		28, 190, 27, 240, 25, 290, 21, 322)
	// Patches where the old bark has shed, grey-green, long and soft.
	for _, b := range [][5]float64{
		{38, 300, 3.5, 9, 0.08}, {46, 252, 2.6, 8, -0.06}, {32, 214, 3, 7, 0.05},
		{42, 176, 2.2, 6, 0}, {50, 286, 1.8, 5, 0.1}, {30, 262, 2.2, 5, -0.05},
	} {
		p.Ellipse(b[0], b[1], b[2], b[3], b[4], resurrect64.Khaki)
		p.Ellipse(b[0]-0.5, b[1]-1, b[2]*0.55, b[3]*0.7, b[4], resurrect64.Mist)
	}
	for _, b := range [][3]float64{{36, 236, 1.4}, {46, 206, 1.1}, {38, 160, 1.2}} {
		p.Ellipse(b[0], b[1], b[2], b[2]*1.8, 0, barkShade)
		p.Ellipse(b[0]-0.3, b[1]-0.4, b[2]*0.5, b[2], 0, resurrect64.Shadow)
	}
	canopy(p)
	for _, b := range boughs {
		branch(p, b, 5)
	}
	for _, b := range branches {
		branch(p, b, 7)
	}
}

// canopy draws the leaves in clumps: an outline round them all, then each
// clump shaded below and lit above, with leaves hanging from its edges.
func canopy(p Pen) {
	for _, c := range clumps {
		p.Circle(c[0], c[1], c[2]+Line, ink)
	}
	for _, c := range clumps {
		p.Circle(c[0], c[1], c[2], leafShade)
	}
	for _, c := range clumps {
		p.Circle(c[0]-c[2]*0.12, c[1]-c[2]*0.16, c[2]*0.8, leaf)
	}
	for _, c := range clumps {
		// Leaves across the clump: dark below, lit towards the top left.
		for j := range 12 {
			a := float64(j) * 2.4
			r := c[2] * 0.75 * math.Sqrt((float64(j)+0.5)/12)
			x, y := c[0]+r*math.Cos(a), c[1]+r*math.Sin(a)
			col := leafShade
			if x-c[0]+y-c[1] < -c[2]*0.2 {
				col = leafLight
			}
			p.Ellipse(x, y, 1.1, 3, a, col)
		}
		// Long leaves hanging below the clump, as gum leaves do.
		for i, a := range []float64{-0.6, -0.15, 0.3, 0.75} {
			x := c[0] + a*c[2]
			y := c[1] + c[2]*0.8 - abs(a)*c[2]*0.3
			p.Ellipse(x, y+3, 1.3, 4, a*0.4, ink)
			p.Ellipse(x, y+3, 0.8, 3.4, a*0.4, []int{leafShade, leaf}[i%2])
		}
	}
}

// branch draws a branch w wide where it leaves the tree, through the x, y
// pairs along its middle, out to its last, reaching a little past it, with
// a twig and leaves at its tip.
func branch(p Pen, mid []float64, w float64) {
	n := len(mid)
	x, y := mid[n-2], mid[n-1]
	pts := append(append([]float64(nil), mid...), x+14, y-1)
	tip := w * 3 / 7
	taper(p, ink, w+2*Line, tip+2*Line, pts)
	p.Stroke(ink, 1.6+2*Line, x+6, y-1, x+14, y-8)
	taper(p, barkShade, w, tip, pts)
	p.Stroke(barkShade, 1.6, x+6, y-1, x+14, y-8)
	taper(p, bark, w*4/7, tip/2, shift(pts, 0, -1))
	p.Curve(barkLight, 1, shift(pts[2:n], 0, -1.6)...)
	y--
	for i, a := range []float64{0.5, -0.4, 1.2} {
		lx, ly := x+14+a*3, y-7+float64(i)
		p.Ellipse(lx, ly+3, 1.3, 4, a*0.5, ink)
		p.Ellipse(lx, ly+3, 0.8, 3.4, a*0.5, leaf)
	}
}

// taper draws a smooth line through pts in palette colour c, from w0 wide
// at its start to w1 at its end.
func taper(p Pen, c int, w0, w1 float64, pts []float64) {
	cv := curve(pts, false)
	n := len(cv)/2 - 1
	for i := range n {
		w := w0 + (w1-w0)*float64(i)/float64(max(n-1, 1))
		p.Stroke(c, w, cv[2*i], cv[2*i+1], cv[2*i+2], cv[2*i+3])
	}
}

// shift is the x, y pairs moved by (dx, dy).
func shift(xy []float64, dx, dy float64) []float64 {
	out := make([]float64, len(xy))
	for i := 0; i+1 < len(xy); i += 2 {
		out[i], out[i+1] = xy[i]+dx, xy[i+1]+dy
	}
	return out
}

func abs(v float64) float64 { return max(v, -v) }
