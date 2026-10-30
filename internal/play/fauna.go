package play

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/gfx"
)

// The pixel wildlife's frames (DESIGN D30), as assets/sprites/fauna.txt has
// them: each centred on its origin, at a few sizes.
var (
	birdSizes  = []float64{0.6, 0.8, 1}
	dizzySizes = []float64{0.4, 0.6, 0.8, 1}
)

const (
	wingFrames  = 4 // wings from up to down
	hopFrames   = 8 // through a hop, from landing
	rooGrazes   = 3 // sitting up, half down, grazing
	rooChews    = 2
	rooJoeySize = 0.7 // smaller than this is a joey at foot
)

// speciesNames is each bird's name in its sprites' names.
var speciesNames = [art.NumSpecies]string{art.Kooka: "kooka", art.Cockatoo: "cocky", art.Magpie: "magpie"}

// nearestSize is the size in sizes nearest s, as a whole percentage.
func nearestSize(sizes []float64, s float64) int {
	best := sizes[0]
	for _, z := range sizes {
		if math.Abs(z-s) < math.Abs(best-s) {
			best = z
		}
	}
	return int(math.Round(best * 100))
}

// wingFrame is the frame for wings at wing, from up at 1 to down at -1.
func wingFrame(wing float64) int {
	return clamp(int(math.Round((1-wing)/2*(wingFrames-1))), 0, wingFrames-1)
}

// drawCentred draws frame i of the sprite named name with its middle at
// (x, y), facing left if left.
func drawCentred(dst *ebiten.Image, sh *gfx.Sheet, name string, i int, x, y float64, left bool) {
	sp := sh.Sprite(name)
	sp.Draw(dst, i, math.Round(x)-float64(sp.W/2), math.Round(y)-float64(sp.H/2), left)
}

// birdSprite is the sprite and frame for a bird of kind k at size s:
// perched, its head dipped by bob from 0 to 1, or flying, its wings at wing.
func birdSprite(k art.Species, s float64, flying bool, bob, wing float64) (string, int) {
	if flying {
		return fmt.Sprintf("bird_%s_fly_%d", speciesNames[k], nearestSize(birdSizes, s)), wingFrame(wing)
	}
	return fmt.Sprintf("bird_%s_perch_%d", speciesNames[k], nearestSize(birdSizes, s)), int(math.Round(min(max(bob, 0), 1)))
}

// rooSprite is the sprite and frame for a kangaroo of size s, with a joey
// in its pouch if pouch, as r has it.
func rooSprite(s float64, pouch bool, r art.Roo) (string, int) {
	name := "roo"
	switch {
	case s < rooJoeySize:
		name = "roo_joey"
	case pouch:
		name = "roo_pouch"
	}
	if r.Hop >= 0 {
		return name + "_hop", clamp(int(r.Hop*hopFrames), 0, hopFrames-1)
	}
	chew := 0
	if math.Sin(r.Chew) < 0 {
		chew = 1
	}
	name += "_sit"
	if r.Flick >= 0.5 {
		name += "_flick"
	}
	return name, clamp(int(math.Round(r.Graze*(rooGrazes-1))), 0, rooGrazes-1)*rooChews + chew
}

// drawDizzy draws the dazed effect over a dazed koala of size s with its
// feet at (x, y): the mini kookaburras at tick, as sprites in the pixel
// look and in shapes in the smooth one (DESIGN D24).
func (p pen) drawDizzy(x, y, s float64, tick int) {
	if p.fauna == nil {
		a := art.Pen{Dst: p.dst, K: p.k}.At(x, y)
		if s != 1 {
			a = a.Sized(s)
		}
		art.Dizzy(a, tick)
		return
	}
	for _, b := range art.DizzyBirds(tick) {
		name := fmt.Sprintf("dizzy_%d", nearestSize(dizzySizes, b.Size*s))
		drawCentred(p.dst, p.fauna, name, wingFrame(b.Wing), x+b.X*s, y+b.Y*s, b.Left)
	}
}
