package play

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// rangerArms is how a ranger holds its arms, as the pixel ranger's sprites
// have them (DESIGN D30); the smooth art takes the hands' place itself.
type rangerArms string

const (
	armsHold  rangerArms = "hold"  // holding up the sheet
	armsDown  rangerArms = "down"  // let down at its sides
	armsCarry rangerArms = "carry" // carrying joeys on its forearms
	armsPick  rangerArms = "pick"  // picking joeys up
)

// The pixel ranger's sprites: how many frames its stride has, after its
// standing frame, and how far down them its feet are.
const (
	rangerSteps = 8
	rangerFeet  = 61
)

// drawRanger draws a ranger in stance s with its feet at x on the ground,
// facing left if left: as a sprite in the pixel look, holding its arms as
// arms, and in shapes in the smooth one.
func (p pen) drawRanger(x float64, left bool, s art.Stance, arms rangerArms) {
	if p.rangers == nil {
		art.Ranger(p.art().At(x, config.GroundY).Mirrored(left), s)
		return
	}
	name, frame := rangerSprite(s, arms)
	sp := p.rangers.Sprite(name)
	sp.Draw(p.dst, frame, math.Round(x)-float64(sp.W/2), config.GroundY-rangerFeet, left)
}

// rangerSprite is the pixel ranger's sprite and frame for stance s: frame
// 0 standing, or the stride's frame nearest its step.
func rangerSprite(s art.Stance, arms rangerArms) (string, int) {
	name := "ranger_man_"
	if s.Woman {
		name = "ranger_woman_"
	}
	name += string(arms)
	if s.Shut {
		name += "_blink"
	}
	if s.Stride < 0.5 {
		return name, 0
	}
	return name, 1 + wrap(int(math.Round(s.Step/(2*math.Pi/rangerSteps))), rangerSteps)
}
