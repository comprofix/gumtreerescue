package main

import (
	"fmt"
	"math"

	"github.com/comprofix/gumtreerescue/internal/art"
)

// The wildlife's frames (DESIGN §3, D24, D27, D30). Each is drawn at a few
// sizes, as the farther ones are smaller, and centred on its origin, so it
// mirrors and lines up about it.
var (
	birdSizes  = []float64{0.6, 0.8, 1}
	dizzySizes = []float64{0.4, 0.6, 0.8, 1}
)

const (
	wingFrames = 4 // wings from up to down
	hopFrames  = 8 // through a hop, from landing
	rooGrazes  = 3 // sitting up, half down, grazing
	rooChews   = 2
)

var species = []struct {
	name string
	s    art.Species
}{{"kooka", art.Kooka}, {"cocky", art.Cockatoo}, {"magpie", art.Magpie}}

// pct is a size as a whole percentage, for a sprite's name.
func pct(s float64) int { return int(math.Round(s * 100)) }

func faunaSprites() []sprite {
	var out []sprite
	for _, sp := range species {
		for _, s := range birdSizes {
			var perch, fly []frame
			for _, bob := range []float64{0, 1} {
				perch = append(perch, frame{draw: func(p art.Pen) { art.Perched(p.Sized(s), sp.s, bob) }})
			}
			for i := range wingFrames {
				wing := 1 - 2*float64(i)/(wingFrames-1)
				fly = append(fly, frame{draw: func(p art.Pen) { art.Flying(p.Sized(s), sp.s, wing) }})
			}
			out = append(out,
				sprite{name: fmt.Sprintf("bird_%s_perch_%d", sp.name, pct(s)), frames: perch, centred: true},
				sprite{name: fmt.Sprintf("bird_%s_fly_%d", sp.name, pct(s)), frames: fly, centred: true})
		}
	}
	for _, s := range dizzySizes {
		var fs []frame
		for i := range wingFrames {
			wing := 1 - 2*float64(i)/(wingFrames-1)
			fs = append(fs, frame{draw: func(p art.Pen) { art.Kookaburra(p.Sized(s), wing) }})
		}
		out = append(out, sprite{name: fmt.Sprintf("dizzy_%d", pct(s)), frames: fs, centred: true})
	}
	roo := func(name string, s float64, pouch bool) {
		for _, flick := range []bool{false, true} {
			var sit []frame
			for g := range rooGrazes {
				for c := range rooChews {
					r := art.Roo{Graze: float64(g) / (rooGrazes - 1), Hop: -1, Chew: math.Pi/2 + math.Pi*float64(c), Pouch: pouch}
					if flick {
						r.Flick = 1
					}
					sit = append(sit, frame{draw: func(p art.Pen) { art.Kangaroo(p.Sized(s), r) }})
				}
			}
			n := name + "_sit"
			if flick {
				n += "_flick"
			}
			out = append(out, sprite{name: n, frames: sit, centred: true})
		}
		var hop []frame
		for i := range hopFrames {
			r := art.Roo{Hop: float64(i) / hopFrames, Pouch: pouch}
			hop = append(hop, frame{draw: func(p art.Pen) { art.Kangaroo(p.Sized(s), r) }})
		}
		out = append(out, sprite{name: name + "_hop", frames: hop, centred: true})
	}
	roo("roo", 0.95, false)
	roo("roo_pouch", 0.95, true)
	roo("roo_joey", 0.5, false)
	return out
}

func faunaFile() file {
	about := fmt.Sprintf(`; The wildlife in the pixel look (DESIGN §3, D24, D27, D30), generated from
; the smooth art: kookaburras, cockatoos and magpies perched (2 frames, the
; head dipping) and flying (%d frames, wings up to down) at sizes from far
; to near; the dazed effect's mini kookaburras; kangaroos sitting up to
; grazing (%d steps of %d chewing frames) and hopping (%d frames), one with
; a joey in her pouch, and joeys at foot. Each is centred on its origin,
; its feet for those sitting, its middle for those flying.
`, wingFrames, rooGrazes, rooChews, hopFrames)
	return file{name: "fauna", about: about, sprites: faunaSprites()}
}
