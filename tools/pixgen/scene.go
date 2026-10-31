package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// scene is a sprite file of pictures fixed in the field, such as the tree
// and the backdrop: each drawn in field coordinates within the rect at
// (x0, y0), w by h, cropped to what it covers, its top left printed.
type scene struct {
	name, about string
	x0, y0      int
	w, h        int
	parts       []part
}

// part is one sprite of a scene. With over set, it is only what draw
// changes from over, to lay over that sprite.
type part struct {
	name       string
	draw, over func(art.Pen)
}

func writeScene(s scene) {
	img := ebiten.NewImage(s.w*Scale, s.h*Scale)
	type done struct {
		name           string
		g              grid
		x0, x1, y0, y1 int
	}
	var all []done
	used := map[int]bool{}
	for _, pt := range s.parts {
		g := reduce(img, pt.draw, s.w, s.h, -s.x0, -s.y0)
		if pt.over != nil {
			base := reduce(img, pt.over, s.w, s.h, -s.x0, -s.y0)
			for y := range g {
				for x := range g[y] {
					if g[y][x] == base[y][x] {
						g[y][x] = -1
					}
				}
			}
		}
		d := done{name: pt.name, g: g, x0: s.w, y0: s.h}
		for y, row := range g {
			for x, c := range row {
				if c >= 0 {
					used[c] = true
					d.x0, d.x1 = min(d.x0, x), max(d.x1, x+1)
					d.y0, d.y1 = min(d.y0, y), max(d.y1, y+1)
				}
			}
		}
		all = append(all, d)
	}
	key := keyFor(used)
	var b strings.Builder
	b.WriteString(s.about)
	b.WriteString("colours")
	for i := range resurrect64.Palette.Colors {
		if c, ok := key[i]; ok {
			fmt.Fprintf(&b, " %c=%s", c, resurrect64.Palette.Names[i])
		}
	}
	b.WriteString("\n")
	for _, d := range all {
		fx, fy := s.x0+d.x0, s.y0+d.y0
		fmt.Printf("\t%s\t%d %d\n", d.name, fx, fy)
		fmt.Fprintf(&b, "\n; Its top left is field (%d, %d).\nsprite %s %d %d\n", fx, fy, d.name, d.x1-d.x0, d.y1-d.y0)
		for y := d.y0; y < d.y1; y++ {
			for x := d.x0; x < d.x1; x++ {
				if c := d.g[y][x]; c < 0 {
					b.WriteByte('.')
				} else {
					b.WriteByte(key[c])
				}
			}
			b.WriteByte('\n')
		}
	}
	if err := os.WriteFile(filepath.Join(*out, s.name+".txt"), []byte(b.String()), 0o644); err != nil {
		log.Fatal(err)
	}
}
