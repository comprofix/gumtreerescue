package main

import (
	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// treeBranch is the branch out to drop point d, as the play scene gives it
// to art.Tree: from the tree's middle, where it grows, out.
func treeBranch(d config.Drop) []float64 {
	b := config.Branch(d)
	base := b[len(b)-1].Y
	pts := []float64{float64(config.SpineX(base)), float64(base)}
	for j := len(b) - 1; j >= 0; j-- {
		pts = append(pts, float64(b[j].X), float64(b[j].Y))
	}
	return pts
}

func treeBoughs() [][]float64 {
	var out [][]float64
	for _, b := range config.Boughs {
		var pts []float64
		for _, s := range b {
			pts = append(pts, float64(s.X), float64(s.Y))
		}
		out = append(out, pts)
	}
	return out
}

func treeScene() scene {
	boughs := treeBoughs()
	base := func(p art.Pen) { art.Tree(p, nil, boughs) }
	// The drop points, as the levels bring them in: the hanging branch,
	// then a higher and a lower one.
	drops := []struct {
		name string
		d    config.Drop
	}{
		{"tree_hanging", config.Hanging},
		{"tree_higher", config.Higher},
		{"tree_lower", config.Lower},
	}
	parts := []part{{name: "tree", draw: base}}
	for _, d := range drops {
		br := [][]float64{treeBranch(d.d)}
		parts = append(parts, part{name: d.name, draw: func(p art.Pen) { art.Tree(p, br, boughs) }, over: base})
	}
	return scene{
		name: "tree",
		about: `; The gum tree in the pixel look (DESIGN §3, D30), generated from the smooth
; art: smooth bark, Sand shaded Clay and lit Blush, with Khaki and Mist
; patches where the old bark has shed; Sage leaves shaded Slate and lit
; Fern. tree is the trunk, limbs, canopy and boughs; each tree_ branch is
; the branch out to a drop point, laid over it on the levels that have it.
`,
		w: 230, h: 345, parts: parts,
	}
}

func backdropScene() scene {
	fw, fh := config.Field()
	w, h := float64(fw), float64(fh)
	return scene{
		name: "backdrop",
		about: `; The country bush behind the play in the pixel look (DESIGN §3, D27, D30),
; generated from the smooth art, its blends banded to the palette:
; backdrop_back is all behind the fence, the escarpments, paddocks, farm,
; background gums and cows (D35), with the sky left clear for the day's sky,
; drawn in code (D33, D34); backdrop_front is the fence, the near grass and the red
; dirt track, laid over what stands in the paddock behind the fence.
`,
		w: fw, h: fh,
		parts: []part{
			{name: "backdrop_back", draw: func(p art.Pen) { art.Backdrop(p, w, config.GroundY) }},
			{name: "backdrop_front", draw: func(p art.Pen) { art.BackdropFront(p, w, h, config.GroundY) }},
		},
	}
}

func scenes() []scene { return []scene{treeScene(), backdropScene()} }
