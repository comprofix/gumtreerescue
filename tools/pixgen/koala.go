package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The koala's frames (DESIGN §3, D30). Slopes are turned in steps of a
// 32nd of a turn, the somersault in 16ths.
const (
	slope   = 2 * math.Pi / 32
	tumbles = 16
	walks   = 6
)

// Which slopes each pose is drawn at, in steps: gripping lies up a limb
// from level to steep, walking tips a little either way.
var (
	gripSlopes = []int{0, -1, -2, -3, -4, -5, -6, -7}
	walkSlopes = []int{-2, -1, 0, 1, 2}
)

// look is a variant's look: as it is, blinking, or flicking an ear.
type look struct {
	suffix string
	l      art.Look
}

var looks = []look{{"", art.Look{}}, {"_blink", art.Look{Shut: true}}, {"_flick", art.Look{Flick: 1}}}

// The faces stamped over the generated frames, as the smooth art's eyes
// are too small to come through: a front face (the hand-drawn sample's
// eyes and nose) about the middle of the head, or a side face's eye.
func frontFace(p art.Pen, x, y, s float64, shut bool) mark {
	q := p.Rel(x, y)
	return mark{kind: front, x: q.OX - p.OX, y: q.OY - p.OY, shut: shut, small: s < 1}
}

func sideFace(p art.Pen, x, y float64, shut bool) mark {
	q := p.Rel(x+1.6, y-1.4)
	return mark{kind: sideEye, x: q.OX - p.OX, y: q.OY - p.OY, shut: shut}
}

// koalaSprites is every pose for a koala of size s, with a joey on its back
// if joey; still poses also blink and flick an ear.
func koalaSprites(prefix string, s float64, joey, baby bool) []sprite {
	var out []sprite
	add := func(pose string, still bool, fr func(l art.Look) []frame) {
		for i, v := range looks {
			if i > 0 && !still {
				break
			}
			l := v.l
			l.Joey = joey
			name := prefix + "_" + pose
			if joey {
				name += "_joey"
			}
			out = append(out, sprite{name: name + v.suffix, frames: fr(l)})
		}
	}
	sz := func(p art.Pen) art.Pen { return p.Sized(s) }
	add("sit", true, func(l art.Look) []frame {
		return []frame{{func(p art.Pen) { art.Koala(sz(p), l) },
			func(p art.Pen) []mark { return []mark{frontFace(sz(p), 0, -17.5, s, l.Shut)} }}}
	})
	cling := func(down bool) func(l art.Look) []frame {
		return func(l art.Look) (fs []frame) {
			for i := range 4 {
				r := float64(i) / 3
				fs = append(fs, frame{func(p art.Pen) { art.Clinging(sz(p), l, r, down) },
					func(p art.Pen) []mark {
						if down {
							return []mark{sideFace(sz(p).Mirrored(true).Turned(0.45, 19.5), 1, -19.5, l.Shut)}
						}
						return []mark{sideFace(sz(p), 1.8, -20.5-1.2*r, l.Shut)}
					}})
			}
			return fs
		}
	}
	add("cling", true, cling(false))
	add("clingdown", false, cling(true))
	// Gripping: at each slope, still, reaching, then following.
	add("grip", true, func(l art.Look) (fs []frame) {
		for _, k := range gripSlopes {
			for _, rf := range [][2]float64{{0, 0}, {1, 0}, {0, 1}} {
				turned := func(p art.Pen) art.Pen { return sz(p).Turned(float64(k)*slope, 0) }
				fs = append(fs, frame{func(p art.Pen) { art.Gripping(turned(p), l, rf[0], rf[1]) },
					func(p art.Pen) []mark { return []mark{sideFace(turned(p), 10.5, -10.5, l.Shut)} }})
			}
		}
		return fs
	})
	add("walk", false, func(l art.Look) (fs []frame) {
		for _, k := range walkSlopes {
			for i := range walks {
				turned := func(p art.Pen) art.Pen { return sz(p).Turned(float64(k)*slope, 0) }
				fs = append(fs, frame{func(p art.Pen) { art.Walking(turned(p), l, 2*math.Pi*float64(i)/walks) },
					func(p art.Pen) []mark { return []mark{sideFace(turned(p), 9.5, -9.5, l.Shut)} }})
			}
		}
		return fs
	})
	if baby {
		return out
	}
	add("asleep", false, func(l art.Look) []frame {
		return []frame{{func(p art.Pen) { art.Asleep(p, l) },
			func(p art.Pen) []mark { return []mark{frontFace(p, 0, -14.5, 1, true)} }}}
	})
	add("dazed", false, func(l art.Look) []frame {
		return []frame{{func(p art.Pen) { art.Dazed(p, l) },
			func(p art.Pen) []mark { return []mark{frontFace(p, 0, -17.5, 1, true)} }}}
	})
	add("onbough", true, func(l art.Look) []frame {
		return []frame{{func(p art.Pen) { art.OnBough(p, l) },
			func(p art.Pen) []mark { return []mark{sideFace(p, 2.5, -19, l.Shut)} }}}
	})
	add("chew", true, func(l art.Look) (fs []frame) {
		for i := range 4 {
			fs = append(fs, frame{func(p art.Pen) { art.Chewing(p, l, math.Pi/2*float64(i)) },
				func(p art.Pen) []mark { return []mark{frontFace(p, 0, -17.5, 1, l.Shut)} }})
		}
		return fs
	})
	add("clingchew", true, func(l art.Look) (fs []frame) {
		for i := range 4 {
			fs = append(fs, frame{func(p art.Pen) { art.ClingChewing(p, l, math.Pi/2*float64(i)) },
				func(p art.Pen) []mark { return []mark{sideFace(p, 1.8, -20.5, l.Shut)} }})
		}
		return fs
	})
	add("thump", false, func(l art.Look) (fs []frame) {
		for i := range 3 {
			sq := float64(i) / 2
			squashed := func(p art.Pen) art.Pen { return p.Squashed(1-0.25*sq, 1+0.08*sq) }
			fs = append(fs, frame{func(p art.Pen) { art.Thump(p, l, sq) },
				func(p art.Pen) []mark { return []mark{sideFace(squashed(p), 1.8, -20.5, l.Shut)} }})
		}
		return fs
	})
	add("fall", false, func(l art.Look) (fs []frame) {
		for _, f := range []float64{-1, 0, 1} {
			fs = append(fs, frame{func(p art.Pen) { art.Falling(p, l, f) },
				func(p art.Pen) []mark { return []mark{frontFace(p, 0, -19, 1, l.Shut)} }})
		}
		return fs
	})
	add("tumble", false, func(l art.Look) (fs []frame) {
		for i := range tumbles {
			fs = append(fs, frame{draw: func(p art.Pen) {
				art.Falling(p.Turned(2*math.Pi*float64(i)/tumbles, art.FallingMiddle), l, 0)
			}})
		}
		return fs
	})
	return out
}

// The hand-drawn sitting koala (PLAN M6), kept as the sitting pose, with
// its feet on its bottom row and the middle of its 22 columns over them.
const handSit = `
..kkkk..........kkkk..
.kwwwwk........kwwwwk.
kwwwwwkkkkkkkkkkwwwwwk
kwwwwfllllllllllfwwwwk
kswwfffllllllllfffwwsk
.kswffffffffffffffwsk.
..kkffffffffffffffkk..
.kffffffffffffffffffk.
.kffffkwffffffkwffffk.
.kffffkkfkkkkfkkffffk.
.kllfffffkkkkfffffllk.
..kllffffkdkkffffllk..
...kkfffffkkfffffkk...
.....kkkffffffkkk.....
....kssfllllllfssk....
...ksskfllllllfkssk...
..kfsskfllllllfkssfk..
..kfsskfllllllfkssfk..
..kffsskfllllfkssffk..
..kfffsskfllfkssfffk..
..kfffkkffllffkkfffk..
..ksfffffllllfffffsk..
..kssffffffffffffssk..
..ksssssffffffsssssk..
.kkfffkkkkkkkkkkfffkk.
kfffffffkkkkkkfffffffk
kssffffsk....ksffffssk
.kkkkkkk......kkkkkkk.`

// The hand-drawn sitting joey, in the same colours, its feet on its bottom
// row and the middle of its 12 columns over them.
const handJoey = `
.kkk....kkk.
kwwfk..kfwwk
kwfffkkfffwk
.kffffffffk.
kffkwffkwffk
kffkkffkkffk
.kfffkkfffk.
.kllfkkfllk.
..kkffffkk..
..ksllllsk..
.ksfllllfsk.
.ksfllllfsk.
.kssffffssk.
kfffkkkkfffk
kssfk..kfssk
.kkk....kkk.`

// hand is a hand-drawn sitting figure: its rows, the row of its eyes'
// glints and their columns, and the columns and rows of its right ear.
type hand struct {
	rows          string
	eyeRow        int
	eyeCols       []int
	earCol, earTo int
}

var (
	koalaHand = hand{handSit, 8, []int{6, 7, 14, 15}, 15, 6}
	joeyHand  = hand{handJoey, 4, []int{3, 4, 7, 8}, 7, 3}
)

// grid is h as a grid about Reach, blinking if shut, or with its right ear
// flicked up a row.
func (h hand) grid(shut, flick bool) grid {
	keys := map[byte]int{'k': resurrect64.Ink, 's': resurrect64.Lilac, 'f': resurrect64.Fog,
		'l': resurrect64.Mist, 'w': resurrect64.White, 'd': resurrect64.Dusk}
	rows := strings.Split(strings.TrimSpace(h.rows), "\n")
	g := make(grid, 2*Reach)
	for y := range g {
		g[y] = make([]int, 2*Reach)
		for x := range g[y] {
			g[y][x] = -1
		}
	}
	w := len(rows[0])
	top, left := Reach-len(rows), Reach-w/2
	for y, r := range rows {
		for x := range len(r) {
			if r[x] != '.' {
				g[top+y][left+x] = keys[r[x]]
			}
		}
	}
	if shut {
		// The eyes' top row, with its glints, becomes fur.
		for _, x := range h.eyeCols {
			g[top+h.eyeRow][left+x] = resurrect64.Fog
		}
	}
	if flick {
		for y := range h.earTo {
			for x := h.earCol; x < w; x++ {
				g[top+y-1][left+x] = g[top+y][left+x]
			}
		}
	}
	return g
}

func koalaFile() file {
	about := fmt.Sprintf(`; The koalas in the pixel look (DESIGN §3, D30), generated from the smooth
; art, with the faces stamped on from the hand-drawn sitting koala: an Ink
; outline, Fog fur shaded Lilac and lit Mist. Each sprite is as wide
; either side of the koala's feet, so it mirrors about them. Slopes are in
; steps of a 32nd of a turn, the somersault (tumble) in %d steps; walking
; has %d frames a stride.
`, tumbles, walks)
	var s []sprite
	s = append(s, koalaSprites("koala", 1, false, false)...)
	s = append(s, koalaSprites("koala", 1, true, false)...)
	s = append(s, koalaSprites("joey", art.JoeySize, false, true)...)
	// The sitting koala and joey are drawn by hand.
	for i := range s {
		shut, flick := strings.HasSuffix(s[i].name, "_blink"), strings.HasSuffix(s[i].name, "_flick")
		switch strings.TrimSuffix(strings.TrimSuffix(s[i].name, "_blink"), "_flick") {
		case "koala_sit":
			s[i].hand = []grid{koalaHand.grid(shut, flick)}
		case "joey_sit":
			s[i].hand = []grid{joeyHand.grid(shut, flick)}
		}
	}
	return file{name: "koala", about: about, sprites: s}
}

func files() []file { return []file{koalaFile(), rangerFile(), faunaFile(), reptileFile()} }
