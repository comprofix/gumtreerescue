// Command pixgen generates the pixel look's first frames from the smooth art
// (DESIGN §3, D30): each frame is drawn with hard edges at Scale screen
// pixels to a field pixel, and each Scale×Scale block becomes one sprite
// pixel of its main colour, or see-through if mostly empty. The output is a
// starting point, checked by eye and touched up by hand in assets/sprites.
//
//	go run ./tools/pixgen -out <dir> [-only koala]
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

var (
	out  = flag.String("out", ".", "where to write the sprite files")
	only = flag.String("only", "", "only this file's sprites (e.g. koala)")
)

// Scale is how many screen pixels a field pixel is drawn at before each
// block is reduced to one sprite pixel; Reach is how far from its origin a
// frame may reach, in field pixels.
const (
	Scale = 8
	Reach = 64
)

// frame draws one frame about the pen's origin, then stamps the marks on
// it.
type frame struct {
	draw  func(p art.Pen)
	marks func(p art.Pen) []mark
}

// mark is a face stamped over a generated frame, at x, y in field pixels
// from the origin: a front face's middle, a side face's eye, or a dot of
// Ink right there.
type mark struct {
	kind        int
	x, y        float64
	shut, small bool
}

const (
	front = iota
	sideEye
	dot
)

// frontRows is the hand-drawn sitting koala's eyes and nose, from the row
// above its eyes' middle and the column left of its middle; . leaves the
// frame's own pixel.
var frontRows = []string{
	".kw......kw.",
	".kkfkkkkfkk.",
	"...fkkkkf...",
	"...fkdkkf...",
	"....fkkf....",
}

var frontShut = []string{
	".ff......ff.",
	".kkfkkkkfkk.",
	"...fkkkkf...",
	"...fkdkkf...",
	"....fkkf....",
}

var stampColour = map[byte]int{'k': resurrect64.Ink, 'w': resurrect64.White, 'f': resurrect64.Fog, 'd': resurrect64.Dusk}

// stamp puts m's face on g, only over the figure.
func stamp(g grid, m mark) {
	set := func(x, y, c int) {
		if y >= 0 && y < len(g) && x >= 0 && x < len(g) && g[y][x] >= 0 {
			g[y][x] = c
		}
	}
	switch m.kind {
	case front:
		cb := Reach + int(math.Round(m.x))
		if m.small {
			er := int(math.Floor(Reach + m.y - 0.66))
			set(cb-2, er, resurrect64.Ink)
			set(cb+1, er, resurrect64.Ink)
			set(cb-1, er+1, resurrect64.Ink)
			set(cb, er+1, resurrect64.Ink)
			return
		}
		er := int(math.Floor(Reach + m.y - 1.2))
		rows := frontRows
		if m.shut {
			rows = frontShut
		}
		for dy, r := range rows {
			for dx := range len(r) {
				if c, ok := stampColour[r[dx]]; ok {
					set(cb-6+dx, er-1+dy, c)
				}
			}
		}
	case dot:
		set(int(math.Floor(Reach+m.x)), int(math.Floor(Reach+m.y)), resurrect64.Ink)
	case sideEye:
		ex, ey := int(math.Floor(Reach+m.x)), int(math.Floor(Reach+m.y))
		set(ex, ey, resurrect64.Ink)
		if !m.shut {
			set(ex, ey-1, resurrect64.Ink)
		}
	}
}

type sprite struct {
	name   string
	frames []frame
	hand   []grid // hand-drawn frames, used as they are in place of frames
	// centred crops it as tall above its origin as below too, so its
	// origin is its middle.
	centred bool
	speckle *speckle // fine speckles too small to survive reducing, put back
}

// speckle spots colour with over the pixels of colour base, one in every
// few in a fixed scatter about the origin, as a hide finely speckled all
// over.
type speckle struct{ base, with int }

func (s *speckle) apply(g grid) {
	for y, row := range g {
		for x, c := range row {
			if c == s.base && ((x-Reach)+3*(y-Reach)+500)%5 == 0 {
				row[x] = s.with
			}
		}
	}
}

// file is one sprite file: its comment and its sprites.
type file struct {
	name, about string
	sprites     []sprite
	ink         int // how much Ink counts for in a block, 0 for twice
	accents     map[int]float64
}

// grid is a frame's pixels: palette indices, -1 see-through, Reach*2 square
// with the origin at (Reach, Reach).
type grid [][]int

var index = func() map[color.RGBA]int {
	m := map[color.RGBA]int{}
	for i, c := range resurrect64.Palette.Colors {
		m[c] = i
	}
	return m
}()

// weight is how much each colour counts for in a block: Ink outlines and
// the darkest details count for more, so thin lines stay.
var weight = func() map[int]int {
	m := map[int]int{}
	for i := range resurrect64.Palette.Colors {
		m[i] = 1
	}
	m[resurrect64.Ink], m[resurrect64.Shadow] = 2, 2
	return m
}()

// unblend is a pixel read back, its colour times its alpha a, as its own
// colour.
func unblend(r, g, b, a byte) color.RGBA {
	if a == 255 {
		return color.RGBA{r, g, b, 255}
	}
	un := func(v byte) byte { return byte(min(255, int(v)*255/int(a))) }
	return color.RGBA{un(r), un(g), un(b), 255}
}

// exactIndex is the palette's colours, as they are.
var exactIndex = maps.Clone(index)

// distance is how far apart two colours look.
func distance(c, p color.RGBA) int {
	dr, dg, db := int(c.R)-int(p.R), int(c.G)-int(p.G), int(c.B)-int(p.B)
	return 3*dr*dr + 4*dg*dg + 2*db*db
}

// nearest is the palette colour nearest c: c itself, if it is one.
func nearest(c color.RGBA) int {
	if i, ok := index[c]; ok {
		return i
	}
	best := 0
	for i, p := range resurrect64.Palette.Colors {
		if distance(c, p) < distance(c, resurrect64.Palette.Colors[best]) {
			best = i
		}
	}
	index[c] = best
	return best
}

// bayer is the 4×4 ordered dither's thresholds, in 16ths.
var bayer = [4][4]int{{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5}}

// blend is how a blended colour is made of two palette colours: a, b and
// how far it is from a to b.
type blend struct {
	a, b int
	t    float64
}

var blends = map[color.RGBA]blend{}

// between is the pair of palette colours c lies between, nearest to the
// line from one to the other, and not too far apart, as a faded shape
// blends its colour over another.
func between(c color.RGBA) blend {
	if bl, ok := blends[c]; ok {
		return bl
	}
	pal := resurrect64.Palette.Colors
	f := func(p color.RGBA) [3]float64 { return [3]float64{float64(p.R), float64(p.G), float64(p.B)} }
	cv := f(c)
	best, bestCost := blend{a: nearest(c)}, math.Inf(1)
	for i := range pal {
		for j := i + 1; j < len(pal); j++ {
			a, b := f(pal[i]), f(pal[j])
			var ab, ac [3]float64
			l2, dot := 0.0, 0.0
			for k := range 3 {
				ab[k], ac[k] = b[k]-a[k], cv[k]-a[k]
				l2 += ab[k] * ab[k]
				dot += ab[k] * ac[k]
			}
			t := min(max(dot/l2, 0), 1)
			d := 0.0
			for k := range 3 {
				e := cv[k] - (a[k] + t*ab[k])
				d += e * e
			}
			if cost := d + l2/40; cost < bestCost {
				best, bestCost = blend{i, j, t}, cost
			}
		}
	}
	blends[c] = best
	return best
}

// dither is the palette colour for a blended block of colour c at (x, y):
// of the two colours it lies between, the one the ordered dither picks
// there, as far as c is from one to the other.
func dither(c color.RGBA, x, y int) int {
	bl := between(c)
	if bl.t > (float64(bayer[y%4][x%4])+0.5)/16 {
		return bl.b
	}
	return bl.a
}

// inkWeight is the file being written's own weight for Ink, 0 for weight's.
var inkWeight int

// accents is the file being written's small marks that would be lost to a
// block's majority: each colour wins a block it covers that share of.
var accents map[int]float64

func render(img *ebiten.Image, f frame) grid {
	g := reduce(img, f.draw, 2*Reach, 2*Reach, Reach, Reach)
	if f.marks != nil {
		for _, m := range f.marks(art.Pen{}.At(0, 0)) {
			stamp(g, m)
		}
	}
	return g
}

// reduce draws f with its origin at (ox, oy) and reduces it to w by h
// sprite pixels.
func reduce(img *ebiten.Image, f func(art.Pen), w, h, ox, oy int) grid {
	img.Clear()
	f(art.Pen{Dst: img, K: Scale, Hard: true}.At(float64(ox), float64(oy)))
	b := img.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(pix)
	g := make(grid, h)
	inked := make([][]bool, h) // the outline covers a fair part of the block
	for by := range h {
		inked[by] = make([]bool, w)
		g[by] = make([]int, w)
		for bx := range w {
			count := map[int]int{}
			solid, exact := 0, 0
			var sum [3]int // the blended pixels' colours, added up
			for y := by * Scale; y < (by+1)*Scale; y++ {
				for x := bx * Scale; x < (bx+1)*Scale; x++ {
					o := 4 * (y*b.Dx() + x)
					if pix[o+3] < 128 {
						continue
					}
					solid++
					c := unblend(pix[o], pix[o+1], pix[o+2], pix[o+3])
					if _, ok := exactIndex[c]; ok {
						exact++
					}
					sum[0], sum[1], sum[2] = sum[0]+int(c.R), sum[1]+int(c.G), sum[2]+int(c.B)
					count[nearest(c)]++
				}
			}
			g[by][bx] = -1
			inked[by][bx] = 5*count[resurrect64.Ink] >= Scale*Scale
			if 2*solid < Scale*Scale {
				continue
			}
			if 2*exact < solid {
				// Mostly blended by faded shapes: dither between the two
				// palette colours nearest the block's own.
				g[by][bx] = dither(color.RGBA{uint8(sum[0] / solid), uint8(sum[1] / solid), uint8(sum[2] / solid), 255}, bx, by)
				continue
			}
			// Outlines and dark details are thin, so they count for more.
			for i := range count {
				count[i] *= weight[i]
				if i == resurrect64.Ink && inkWeight > 0 {
					count[i] = count[i] / weight[i] * inkWeight
					if 5*count[i] < 3*solid {
						count[i] = 0 // only where it mostly covers the block
					}
				}
			}
			best := -1
			for c, share := range accents {
				if n := count[c] / weight[c]; float64(n) >= share*Scale*Scale && (best < 0 || n > count[best]/weight[best]) {
					best = c
				}
			}
			if best >= 0 {
				g[by][bx] = best
				continue
			}
			for i, c := range count {
				if best < 0 || c > count[best] || c == count[best] && i < best {
					best = i
				}
			}
			g[by][bx] = best
		}
	}
	// The outline round the edge: a see-through pixel beside the figure
	// that the outline covers a fair part of becomes Ink.
	var edge [][2]int
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			if g[y][x] >= 0 || !inked[y][x] {
				continue
			}
			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if c := g[y+d[1]][x+d[0]]; c >= 0 && c != resurrect64.Ink {
					edge = append(edge, [2]int{x, y})
					break
				}
			}
		}
	}
	for _, e := range edge {
		g[e[1]][e[0]] = resurrect64.Ink
	}
	return g
}

// write renders the file's sprites and writes them, each cropped to what
// its frames cover, as wide either side of its origin so it mirrors about
// it. It returns each sprite's origin, from its top left.
func write(img *ebiten.Image, f file) map[string][2]int {
	inkWeight, accents = f.ink, f.accents
	type done struct {
		name   string
		frames []grid
		x0, x1 int
		y0, y1 int
	}
	var all []done
	used := map[int]bool{}
	for _, s := range f.sprites {
		d := done{name: s.name, x0: Reach, x1: Reach, y0: Reach, y1: Reach}
		for i, fr := range s.frames {
			var g grid
			if s.hand != nil {
				g = s.hand[i]
			} else {
				g = render(img, fr)
			}
			if s.speckle != nil {
				s.speckle.apply(g)
			}
			for y, row := range g {
				for x, c := range row {
					if c < 0 {
						continue
					}
					used[c] = true
					d.x0, d.x1 = min(d.x0, x), max(d.x1, x+1)
					d.y0, d.y1 = min(d.y0, y), max(d.y1, y+1)
				}
			}
			d.frames = append(d.frames, g)
		}
		all = append(all, d)
	}
	// A sprite and its blink and flick share one box, so they swap in
	// place.
	family := func(name string) string {
		return strings.TrimSuffix(strings.TrimSuffix(name, "_blink"), "_flick")
	}
	box := map[string][4]int{}
	for _, d := range all {
		b, ok := box[family(d.name)]
		if !ok {
			b = [4]int{d.x0, d.x1, d.y0, d.y1}
		}
		box[family(d.name)] = [4]int{min(b[0], d.x0), max(b[1], d.x1), min(b[2], d.y0), max(b[3], d.y1)}
	}
	for i := range all {
		d := &all[i]
		b := box[family(d.name)]
		half := max(Reach-b[0], b[1]-Reach)
		d.x0, d.x1, d.y0, d.y1 = Reach-half, Reach+half, b[2], b[3]
		if f.sprites[i].centred {
			half := max(Reach-b[2], b[3]-Reach)
			d.y0, d.y1 = Reach-half, Reach+half
		}
	}
	key := keyFor(used)
	var b strings.Builder
	b.WriteString(f.about)
	b.WriteString("colours")
	for i := range resurrect64.Palette.Colors {
		if c, ok := key[i]; ok {
			fmt.Fprintf(&b, " %c=%s", c, resurrect64.Palette.Names[i])
		}
	}
	b.WriteString("\n")
	origins := map[string][2]int{}
	for _, d := range all {
		fmt.Fprintf(&b, "\nsprite %s %d %d\n", d.name, d.x1-d.x0, d.y1-d.y0)
		origins[d.name] = [2]int{Reach - d.x0, Reach - d.y0}
		for i, g := range d.frames {
			if i > 0 {
				b.WriteString("\n")
			}
			for y := d.y0; y < d.y1; y++ {
				for x := d.x0; x < d.x1; x++ {
					if c := g[y][x]; c < 0 {
						b.WriteByte('.')
					} else {
						b.WriteByte(key[c])
					}
				}
				b.WriteByte('\n')
			}
		}
	}
	if err := os.WriteFile(filepath.Join(*out, f.name+".txt"), []byte(b.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	// The compare sheet: each sprite's middle frame, as the smooth art draws
	// it and as sprite pixels, side by side at Scale.
	var tiles []struct {
		smooth, pixel *ebiten.Image
	}
	for i, s := range f.sprites {
		d := all[i]
		if strings.HasSuffix(d.name, "_blink") || strings.HasSuffix(d.name, "_flick") {
			continue
		}
		m := len(s.frames) / 2
		w, h := (d.x1-d.x0)*Scale, (d.y1-d.y0)*Scale
		sm := ebiten.NewImage(w, h)
		sm.Fill(color.RGBA{0x55, 0x6d, 0x8a, 0xff})
		if draw := s.frames[m].draw; draw != nil {
			draw(art.Pen{Dst: sm, K: Scale}.At(float64(Reach-d.x0), float64(Reach-d.y0)))
		}
		px := ebiten.NewImage(w, h)
		px.Fill(color.RGBA{0x55, 0x6d, 0x8a, 0xff})
		for y := d.y0; y < d.y1; y++ {
			for x := d.x0; x < d.x1; x++ {
				if c := d.frames[m][y][x]; c >= 0 {
					px.SubImage(image.Rect((x-d.x0)*Scale, (y-d.y0)*Scale, (x-d.x0+1)*Scale, (y-d.y0+1)*Scale)).(*ebiten.Image).Fill(resurrect64.Palette.Colors[c])
				}
			}
		}
		tiles = append(tiles, struct{ smooth, pixel *ebiten.Image }{sm, px})
	}
	const across, gap = 4, 16
	cw, ch := 0, 0
	for _, t := range tiles {
		cw, ch = max(cw, 2*t.smooth.Bounds().Dx()+gap), max(ch, t.smooth.Bounds().Dy())
	}
	rows := (len(tiles) + across - 1) / across
	sheet := ebiten.NewImage(across*(cw+2*gap), rows*(ch+gap))
	sheet.Fill(color.RGBA{0x20, 0x20, 0x28, 0xff})
	for i, t := range tiles {
		x, y := float64(i%across*(cw+2*gap)), float64(i/across*(ch+gap))
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		sheet.DrawImage(t.smooth, op)
		op.GeoM.Translate(float64(t.smooth.Bounds().Dx()+gap/2), 0)
		sheet.DrawImage(t.pixel, op)
	}
	savePNG(filepath.Join(*out, f.name+"_compare.png"), sheet)
	return origins
}

func savePNG(name string, img *ebiten.Image) {
	fh, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	if err := png.Encode(fh, img); err != nil {
		log.Fatal(err)
	}
}

// keyFor gives each used colour a character: its name's first letter, in
// lower then upper case, else any free letter.
func keyFor(used map[int]bool) map[int]byte {
	key := map[int]byte{}
	taken := map[byte]bool{'.': true}
	try := func(i int, c byte) bool {
		if taken[c] {
			return false
		}
		key[i], taken[c] = c, true
		return true
	}
	for i := range resurrect64.Palette.Colors {
		if !used[i] {
			continue
		}
		n := resurrect64.Palette.Names[i]
		if try(i, strings.ToLower(n)[0]) || try(i, strings.ToUpper(n)[0]) {
			continue
		}
		for _, c := range []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
			if try(i, c) {
				break
			}
		}
	}
	return key
}

type gen struct{ frames int }

func (g *gen) Update() error {
	if g.frames++; g.frames < 3 {
		return nil
	}
	img := ebiten.NewImage(2*Reach*Scale, 2*Reach*Scale)
	for _, s := range scenes() {
		if *only != "" && s.name != *only {
			continue
		}
		fmt.Printf("%s:\n", s.name)
		writeScene(s)
	}
	for _, f := range files() {
		if *only != "" && f.name != *only {
			continue
		}
		fmt.Printf("%s:\n", f.name)
		origins := write(img, f)
		names := slices.Sorted(maps.Keys(origins))
		for _, name := range names {
			if o := origins[name]; !strings.HasSuffix(name, "_blink") && !strings.HasSuffix(name, "_flick") {
				fmt.Printf("\t%s\t%d %d\n", name, o[0], o[1])
			}
		}
	}
	return ebiten.Termination
}

func (g *gen) Draw(*ebiten.Image)             {}
func (g *gen) Layout(int, int) (w int, h int) { return 64, 36 }

func main() {
	flag.Parse()
	ebiten.SetWindowSize(64, 36)
	ebiten.SetWindowTitle("pixgen")
	if err := ebiten.RunGame(&gen{}); err != nil {
		log.Fatal(err)
	}
}
