package art

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/gfx"
)

// bayer is a 4×4 ordered dither's thresholds, in 16ths.
var bayer = [4][4]int{{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5}}

// Banded turns what is drawn on img into the palette's own colours, as the
// pixel look's sprites are (DESIGN D30): each blended colour becomes the
// two palette colours it lies between, in an ordered dither. See-through
// pixels stay see-through. It reads the image back, so it is slow: for
// pictures drawn once and kept, such as the pixel look's sky.
func Banded(img *ebiten.Image) {
	b := img.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(pix)
	pal := gfx.Palette().Colors
	type pair struct {
		a, b int
		t    float64
	}
	pairs := map[color.RGBA]pair{}
	vec := func(c color.RGBA) [3]float64 { return [3]float64{float64(c.R), float64(c.G), float64(c.B)} }
	// between is the pair of palette colours c lies nearest the line
	// between, and not too far apart, and how far along it c is.
	between := func(c color.RGBA) pair {
		if pr, ok := pairs[c]; ok {
			return pr
		}
		cv := vec(c)
		var best pair
		cost := math.Inf(1)
		for i := range pal {
			for j := i; j < len(pal); j++ {
				a, d := vec(pal[i]), vec(pal[j])
				var ab, ac [3]float64
				l2, dot := 0.0, 0.0
				for k := range 3 {
					ab[k], ac[k] = d[k]-a[k], cv[k]-a[k]
					l2 += ab[k] * ab[k]
					dot += ab[k] * ac[k]
				}
				t := 0.0
				if l2 > 0 {
					t = min(max(dot/l2, 0), 1)
				}
				e := 0.0
				for k := range 3 {
					f := cv[k] - (a[k] + t*ab[k])
					e += f * f
				}
				if e += l2 / 40; e < cost {
					best, cost = pair{i, j, t}, e
				}
			}
		}
		pairs[c] = best
		return best
	}
	for y := range b.Dy() {
		for x := range b.Dx() {
			o := 4 * (y*b.Dx() + x)
			a := pix[o+3]
			if a < 128 {
				pix[o], pix[o+1], pix[o+2], pix[o+3] = 0, 0, 0, 0
				continue
			}
			// Near colours share their pair, so few are worked out.
			un := func(v byte) byte { return byte(min(255, int(v)*255/int(a))) &^ 3 }
			pr := between(color.RGBA{un(pix[o]), un(pix[o+1]), un(pix[o+2]), 255})
			c := pal[pr.a]
			if pr.t > (float64(bayer[y%4][x%4])+0.5)/16 {
				c = pal[pr.b]
			}
			pix[o], pix[o+1], pix[o+2], pix[o+3] = c.R, c.G, c.B, 255
		}
	}
	img.WritePixels(pix)
}
