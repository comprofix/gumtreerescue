package main

import (
	"fmt"
	"math"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// The rangers' frames (DESIGN §3, D30): standing, then rangerSteps frames
// of a stride, for each way they hold their arms.
const rangerSteps = 8

// rangerArms are where a ranger's hands are, as the play scene puts them:
// holding up the sheet, let down at its sides, carrying joeys on its
// forearms, and picking them up.
var rangerArms = []struct {
	name   string
	hx, hy float64
}{
	{"hold", float64(config.RangerSide) / 2, art.RangerHold - 1.5},
	{"down", 2, -29},
	{"carry", 5.5, -33},
	{"pick", 6, -20},
}

func rangerSprites() []sprite {
	var out []sprite
	for _, who := range []string{"man", "woman"} {
		for _, a := range rangerArms {
			for _, shut := range []bool{false, true} {
				var fs []frame
				for i := range rangerSteps + 1 {
					s := art.Stance{Woman: who == "woman", HandX: a.hx, HandY: a.hy, Shut: shut}
					if i > 0 {
						s.Stride, s.Step = 1, 2*math.Pi*float64(i-1)/rangerSteps
					}
					// The eye, as the art puts it, bobbing with the walk.
					bob := 0.6 * s.Stride * math.Abs(math.Cos(s.Step))
					fs = append(fs, frame{func(p art.Pen) { art.Ranger(p, s) },
						func(art.Pen) []mark {
							if shut {
								return nil
							}
							return []mark{{kind: dot, x: 3.8, y: -51.4 - bob}}
						}})
				}
				name := "ranger_" + who + "_" + a.name
				if shut {
					name += "_blink"
				}
				out = append(out, sprite{name: name, frames: fs})
			}
		}
	}
	return out
}

func rangerFile() file {
	about := fmt.Sprintf(`; The rangers in the pixel look (DESIGN §2, §3, D30), generated from the
; smooth art: khaki hat and shirt, a dark green band, olive shorts, brown
; belt and boots. Each sprite is as wide either side of the ranger's feet,
; facing right; frame 0 stands, then %d frames of a stride. The arms hold
; up the sheet (hold), hang down (down), carry joeys (carry) or pick them
; up (pick).
`, rangerSteps)
	// A ranger is slim, so its outlines count only as much as its colours.
	return file{name: "ranger", about: about, sprites: rangerSprites(), ink: 1}
}
