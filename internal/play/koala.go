package play

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/gfx"
)

// koalaKind is what a koala is doing, as the art draws it.
type koalaKind int

const (
	kSit       koalaKind = iota // sitting, facing out
	kAsleep                     // curled up asleep in a fork
	kChew                       // sitting, chewing a gum leaf
	kDazed                      // sitting, holding its head
	kCling                      // holding on to the side of the wood
	kClingChew                  // holding on, chewing a gum leaf
	kThump                      // squashed against the trunk
	kGrip                       // holding on along sloping wood
	kOnBough                    // sitting on a bough
	kWalk                       // on all fours
	kFall                       // in the air, limbs spread
)

// koalaPose is how a koala is drawn this frame, in either look: what it is
// doing, its look, whether it is a joey on its own, facing left, turned,
// and how far through its movement it is.
type koalaPose struct {
	kind   koalaKind
	look   art.Look
	baby   bool
	flip   bool
	turn   float64 // radians clockwise: about its feet, or its middle if tumble
	tumble bool    // turned about its middle, somersaulting in a bounce
	reach  float64 // clinging or gripping: how far its paws reach, 0 to 1
	follow float64 // gripping: how far its back feet follow, 0 to 1
	down   bool    // clinging: backing down
	phase  float64 // walking: its stride; chewing: its jaw (radians)
	squash float64 // thumped: how squashed, 1 at the hit to 0
	flail  float64 // falling: how its limbs wave, -1 to 1
}

// drawKoalaPose draws k with its feet at (x, y), as sprites in the pixel
// look and in shapes in the smooth one.
func (p pen) drawKoalaPose(x, y float64, k koalaPose) {
	if p.koalas != nil {
		k.drawPixel(p.dst, p.koalas, x, y)
		return
	}
	k.drawSmooth(art.Pen{Dst: p.dst, K: p.k}.At(x, y))
}

func (k koalaPose) drawSmooth(a art.Pen) {
	if k.baby {
		a = a.Sized(art.JoeySize)
	}
	a = a.Mirrored(k.flip)
	switch {
	case k.tumble:
		a = a.Turned(k.turn, art.FallingMiddle)
	case k.turn != 0:
		a = a.Turned(k.turn, 0)
	}
	l := k.look
	switch k.kind {
	case kAsleep:
		art.Asleep(a, l)
	case kChew:
		art.Chewing(a, l, k.phase)
	case kDazed:
		art.Dazed(a, l)
	case kCling:
		art.Clinging(a, l, k.reach, k.down)
	case kClingChew:
		art.ClingChewing(a, l, k.phase)
	case kThump:
		art.Thump(a, l, k.squash)
	case kGrip:
		art.Gripping(a, l, k.reach, k.follow)
	case kOnBough:
		art.OnBough(a, l)
	case kWalk:
		art.Walking(a, l, k.phase)
	case kFall:
		art.Falling(a, l, k.flail)
	default:
		art.Koala(a, l)
	}
}

// The pixel koala's frames (DESIGN D30), as assets/sprites/koala.txt has
// them: slopes in steps of a 32nd of a turn, gripping from level down to
// koalaGripSteps steps up, walking koalaWalkSteps either way with
// koalaStrides frames a stride, and the somersault in koalaTumbles.
const (
	koalaSlope     = 2 * math.Pi / 32
	koalaGripSteps = 7
	koalaWalkSteps = 2
	koalaStrides   = 6
	koalaTumbles   = 16
)

// koalaFeet is how far down each koala sprite its feet are, from its top;
// across, they are in its middle.
var koalaFeet = map[string]int{
	"joey_cling": 18, "joey_clingdown": 16, "joey_grip": 14, "joey_sit": 17, "joey_walk": 12,
	"koala_asleep": 25, "koala_asleep_joey": 25, "koala_chew": 29, "koala_chew_joey": 29,
	"koala_cling": 32, "koala_cling_joey": 32, "koala_clingchew": 31, "koala_clingchew_joey": 31,
	"koala_clingdown": 29, "koala_clingdown_joey": 29, "koala_dazed": 28, "koala_dazed_joey": 28,
	"koala_fall": 29, "koala_fall_joey": 29, "koala_grip": 25, "koala_grip_joey": 25,
	"koala_onbough": 30, "koala_onbough_joey": 30, "koala_sit": 29, "koala_sit_joey": 29,
	"koala_thump": 32, "koala_thump_joey": 32, "koala_tumble": 31, "koala_tumble_joey": 31,
	"koala_walk": 22, "koala_walk_joey": 22,
}

func (k koalaPose) drawPixel(dst *ebiten.Image, sh *gfx.Sheet, x, y float64) {
	base, look, frame := k.sprite()
	sp := sh.Sprite(base + look)
	sp.Draw(dst, frame, math.Round(x)-float64(sp.W/2), math.Round(y)-float64(koalaFeet[base]), k.flip)
}

// sprite is k's sprite, as its pose (with its blink or flick added) and
// frame.
func (k koalaPose) sprite() (base, look string, frame int) {
	pose, still := "sit", true
	steps := func(v, step float64) int { return int(math.Round(v / step)) }
	switch k.kind {
	case kAsleep:
		pose, still = "asleep", false
	case kChew:
		pose, frame = "chew", wrap(steps(k.phase, math.Pi/2), 4)
	case kDazed:
		pose, still = "dazed", false
	case kCling:
		pose, frame = "cling", clamp(steps(k.reach, 1.0/3), 0, 3)
		if k.down {
			pose, still = "clingdown", false
		}
	case kClingChew:
		pose, frame = "clingchew", wrap(steps(k.phase, math.Pi/2), 4)
	case kThump:
		pose, still, frame = "thump", false, clamp(steps(k.squash, 0.5), 0, 2)
	case kGrip:
		// Still, reaching ahead, or bringing its back feet up.
		part := 0
		switch {
		case k.reach > 1.0/3 && k.reach >= k.follow:
			part = 1
		case k.follow > 1.0/3:
			part = 2
		}
		pose, frame = "grip", 3*clamp(-steps(k.turn, koalaSlope), 0, koalaGripSteps)+part
	case kOnBough:
		pose = "onbough"
	case kWalk:
		tip := clamp(steps(k.turn, koalaSlope), -koalaWalkSteps, koalaWalkSteps) + koalaWalkSteps
		pose, still, frame = "walk", false, tip*koalaStrides+wrap(steps(k.phase, 2*math.Pi/koalaStrides), koalaStrides)
	case kFall:
		pose, still, frame = "fall", false, clamp(steps(k.flail, 1), -1, 1)+1
		if k.tumble {
			frame = wrap(steps(k.turn, 2*math.Pi/koalaTumbles), koalaTumbles)
			pose = "tumble"
		}
	}
	base = "koala_" + pose
	switch {
	case k.baby:
		// A joey on its own has fewer poses; it sits for the rest.
		switch pose {
		case "sit", "cling", "clingdown", "grip", "walk":
		default:
			pose, still, frame = "sit", true, 0
		}
		base = "joey_" + pose
	case k.look.Joey:
		base += "_joey"
	}
	switch {
	case !still:
	case k.look.Shut:
		look = "_blink"
	case k.look.Flick >= 0.5:
		look = "_flick"
	}
	return base, look, frame
}

func wrap(n, m int) int { return (n%m + m) % m }

func clamp(n, lo, hi int) int { return min(max(n, lo), hi) }
