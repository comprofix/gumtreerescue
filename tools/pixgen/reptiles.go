package main

import (
	"fmt"
	"math"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The reptiles' frames (DESIGN D30, D32), centred on their origins so they
// mirror about them.
const (
	monitorSteps  = 8 // through its walk
	monitorFlicks = 3 // its tongue in to out, resting
	pythonWaves   = 8 // through one curve of its slither
	pythonCoils   = 4 // stretched out to nearly curled up
	pythonFlicks  = 3 // its tongue in to out, coiled
)

// monitorSpeckle puts back the lace monitor's fine gold speckles over its
// blue-black hide.
var monitorSpeckle = &speckle{base: resurrect64.Navy, with: resurrect64.Gold}

func reptileSprites() []sprite {
	var walk, rest, slither, coiled []frame
	for i := range monitorSteps {
		g := art.Goanna{Step: 2 * math.Pi * float64(i) / monitorSteps, Walk: 1}
		walk = append(walk, frame{draw: func(p art.Pen) { art.LaceMonitor(p, g) }})
	}
	for i := range monitorFlicks {
		g := art.Goanna{Tongue: float64(i) / (monitorFlicks - 1)}
		rest = append(rest, frame{draw: func(p art.Pen) { art.LaceMonitor(p, g) }})
	}
	for c := range pythonCoils {
		for i := range pythonWaves {
			s := art.Snake{Wave: 2 * math.Pi * float64(i) / pythonWaves, Coil: float64(c) / pythonCoils}
			slither = append(slither, frame{draw: func(p art.Pen) { art.CarpetPython(p, s) }})
		}
	}
	for i := range pythonFlicks {
		s := art.Snake{Coil: 1, Tongue: float64(i) / (pythonFlicks - 1)}
		coiled = append(coiled, frame{draw: func(p art.Pen) { art.CarpetPython(p, s) }})
	}
	return []sprite{
		{name: "monitor_walk", frames: walk, centred: true, speckle: monitorSpeckle},
		{name: "monitor_rest", frames: rest, centred: true, speckle: monitorSpeckle},
		{name: "python", frames: slither, centred: true},
		{name: "python_coiled", frames: coiled, centred: true},
	}
}

func reptileFile() file {
	about := fmt.Sprintf(`; The reptiles on the track in the pixel look (DESIGN D30, D32), generated
; from the smooth art: the lace monitor walking (%d frames through its
; walk) and resting with its tongue flicking (%d frames); the carpet python
; slithering (%d frames through a curve, for each of %d steps from
; stretched out to nearly curled up) and coiled with its tongue flicking (%d
; frames). Each is centred on its origin, the monitor's feet and the
; python's middle on the track.
`, monitorSteps, monitorFlicks, pythonWaves, pythonCoils, pythonFlicks)
	return file{name: "reptiles", about: about, sprites: reptileSprites(), accents: map[int]float64{
		resurrect64.Lemon: 0.3,  // the python's spots
		resurrect64.Wine:  0.25, // tongues
	}}
}
