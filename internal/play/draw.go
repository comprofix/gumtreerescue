package play

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/shapes"
	"github.com/comprofix/gumtreerescue/internal/ui"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// The play scene is drawn in simple shapes for now; the art proper comes
// in PLAN M4. Every size here is in field pixels.

// pen draws shapes given in field pixels at the shown screen's scale.
type pen struct {
	dst *ebiten.Image
	k   float64
}

func (p pen) rect(x, y, w, h float64, c int) {
	shapes.FillRect(p.dst, shapes.R(x*p.k, y*p.k, w*p.k, h*p.k), c)
}

func (p pen) round(x, y, w, h, r float64, c int) {
	shapes.FillRoundRect(p.dst, shapes.R(x*p.k, y*p.k, w*p.k, h*p.k), r*p.k, c)
}

func (p pen) circle(x, y, r float64, c int) { shapes.FillCircle(p.dst, x*p.k, y*p.k, r*p.k, c) }

func (p pen) line(x0, y0, x1, y1, w float64, c int) {
	shapes.Line(p.dst, x0*p.k, y0*p.k, x1*p.k, y1*p.k, w*p.k, c)
}

func (p pen) poly(c int, pts ...float64) {
	for i := range pts {
		pts[i] *= p.k
	}
	shapes.FillPoly(p.dst, c, pts...)
}

// px turns a fixed-point position into field pixels.
func px(v int) float64 { return float64(v) / config.SubPx }

// koalaY is the field y of a koala's feet.
func koalaY(k *world.Koala) float64 { return config.SheetY - px(k.H) }

// Draw implements scene.Scene.
func (p *Play) Draw(ctx *scene.Context, dst *ebiten.Image) {
	pn := pen{dst, scale(ctx)}
	w := p.world
	drawBackdrop(pn)
	drawTree(pn, w.Row.Drops)
	drawWaiting(pn, w.Waiting)
	drawUte(pn, px(w.UteX()))
	drawTray(pn, w)
	late := w.Session.Skill.LateMarker
	for i := range w.Koalas {
		k := &w.Koalas[i]
		// On Hard the marker shows only for a bounce's last third (DESIGN §5).
		if station, ticks, ok := w.Landing(k); ok && (!late || 3*ticks <= w.Row.BounceTicks) {
			drawMarker(pn, px(world.StationX(station)))
		}
	}
	for i := range w.Koalas {
		switch k := &w.Koalas[i]; k.State {
		case world.Wobbling:
			drawKoala(pn, px(k.X)+1.5*math.Sin(float64(k.Timer)*0.9), koalaY(k), false)
		case world.Falling, world.Arcing, world.Climbing:
			drawKoala(pn, px(k.X), koalaY(k), false)
		}
	}
	drawRangers(pn, px(w.Sheet))
	// Koalas on the ground pass in front of the rangers (DESIGN §5).
	for i := range w.Koalas {
		if k := &w.Koalas[i]; k.State == world.Dazed || k.State == world.Crawling {
			drawKoala(pn, px(k.X), koalaY(k), k.State == world.Crawling)
		}
	}
	drawHUD(ctx, dst, w)
	switch {
	case w.Starting > 0:
		banner(ctx, dst, fmt.Sprintf("LEVEL %d", w.Level), "")
	case w.Ending > 0:
		sub := ""
		if w.Misses == 0 {
			sub = fmt.Sprintf("NO MISSES  +%d", config.CleanBonus)
		}
		banner(ctx, dst, "WELL DONE!", sub)
	}
}

// drawBackdrop is the country bush: sky, clouds, ranges, paddocks, a fence
// and the red dirt track (DESIGN §3).
func drawBackdrop(p pen) {
	fw, fh := config.Field()
	w, h := float64(fw), float64(fh)
	p.rect(0, 0, w, h, resurrect64.Sky)
	for _, c := range [][3]float64{{250, 55, 1}, {470, 40, 1.3}, {590, 80, 0.8}} {
		x, y, s := c[0], c[1], c[2]
		p.circle(x, y, 12*s, resurrect64.White)
		p.circle(x+14*s, y-6*s, 14*s, resurrect64.White)
		p.circle(x+30*s, y, 11*s, resurrect64.White)
		p.round(x-10*s, y, 50*s, 10*s, 5*s, resurrect64.White)
	}
	p.poly(resurrect64.Cobalt, 0, 240, 120, 190, 230, 215, 360, 170, 480, 205, 590, 180, w, 200, w, 250, 0, 250)
	p.rect(0, 238, w, config.GroundY-238, resurrect64.Lime)
	p.rect(0, 270, w, config.GroundY-270, resurrect64.Moss)
	// A shed and a windmill in the distance.
	p.rect(560, 222, 46, 22, resurrect64.Fog)
	p.poly(resurrect64.Lilac, 554, 224, 583, 208, 612, 224)
	p.line(530, 244, 534, 196, 2, resurrect64.Lilac)
	p.line(538, 244, 534, 196, 2, resurrect64.Lilac)
	for _, d := range [][2]float64{{-12, 0}, {12, 0}, {0, -12}, {0, 12}} {
		p.line(534, 196, 534+d[0], 196+d[1], 2, resurrect64.Fog)
	}
	// The fence.
	for x := 10.0; x < w; x += 40 {
		p.rect(x, 252, 4, 26, resurrect64.Bark)
	}
	p.rect(0, 256, w, 3, resurrect64.Caramel)
	p.rect(0, 266, w, 3, resurrect64.Caramel)
	// The track.
	p.rect(0, config.GroundY, w, h-config.GroundY, resurrect64.Copper)
	p.rect(0, config.GroundY, w, 3, resurrect64.Caramel)
}

// drawTree is the gum tree: its trunk, canopy and a branch reaching out to
// each of the level's drop points (DESIGN §2, §5).
func drawTree(p pen, drops []config.Drop) {
	t := float64(config.TrunkX)
	p.poly(resurrect64.Sand, 14, config.GroundY, t, config.GroundY, t-6, 140, t-12, 60, 24, 60, 20, 140)
	p.line(44, 150, 100, 120, 6, resurrect64.Sand)
	p.line(34, 120, 4, 92, 6, resurrect64.Sand)
	for _, c := range [][3]float64{{10, 50, 34}, {50, 30, 38}, {95, 45, 30}, {130, 70, 24}, {-10, 90, 26}, {70, 75, 26}} {
		p.circle(c[0], c[1], c[2], resurrect64.Forest)
	}
	for _, c := range [][3]float64{{30, 25, 20}, {80, 30, 18}, {115, 55, 14}, {0, 70, 16}} {
		p.circle(c[0], c[1], c[2], resurrect64.Emerald)
	}
	for _, d := range drops {
		x0 := px(world.StationX(d.Station))
		y := float64(config.SheetY - d.Height)
		p.line(t-10, y+30, x0-20, y, 7, resurrect64.Sand)
		p.line(x0-22, y, x0+22, y, 5, resurrect64.Sand)
	}
}

// waitSpots are where waiting koalas sit in the tree: on branches, in the
// forks, holding the trunk and at its foot (DESIGN §5), as their feet.
var waitSpots = [config.KoalaLevel][2]float64{
	{96, 118}, {10, 88}, {40, 64}, {56, 150}, {24, 200},
	{58, 230}, {20, 262}, {8, 330}, {130, 64}, {70, 100},
}

func drawWaiting(p pen, n int) {
	for i := range min(n, len(waitSpots)) {
		drawKoala(p, waitSpots[i][0], waitSpots[i][1], false)
	}
}

// drawUte is the ranger ute, parked with its tray's station at x (DESIGN
// §2): tray and headboard behind, cab and bull bar in front.
func drawUte(p pen, x float64) {
	deck := float64(config.SheetY)
	hb := x + config.Headboard
	p.rect(x-config.TrayBack, deck+12, hb+90-(x-config.TrayBack), 8, resurrect64.Charcoal)
	// The headboard and its mesh rack.
	p.rect(hb-4, deck-50, 4, 50, resurrect64.Fog)
	for y := deck - 48; y < deck; y += 6 {
		p.line(hb-8, y, hb-1, y, 1, resurrect64.Fog)
	}
	// The cab, bonnet and bull bar.
	p.round(hb, deck-38, 60, 50, 6, resurrect64.White)
	p.round(hb+7, deck-32, 28, 18, 3, resurrect64.Sky)
	p.round(hb+56, deck-16, 32, 28, 5, resurrect64.White)
	p.rect(hb+86, deck-14, 6, 30, resurrect64.Charcoal)
	p.line(hb+50, deck-16, hb+50, deck-48, 4, resurrect64.Charcoal)
	p.circle(hb+25, deck-4, 7, resurrect64.Forest)
	for _, wx := range []float64{x + 10, hb + 68} {
		p.circle(wx, deck+22, 14, resurrect64.Charcoal)
		p.circle(wx, deck+22, 6, resurrect64.Fog)
	}
}

// drawTray is the ute's tray with the rescued koalas in it, drawn after
// the ute so they sit inside its low sides.
func drawTray(p pen, w *world.World) {
	x := px(w.UteX())
	deck := float64(config.SheetY)
	left := x - config.TrayBack
	n := 0
	for i := range w.Koalas {
		if k := &w.Koalas[i]; k.State == world.InTray {
			drawKoala(p, left+9+float64(n%5)*16, deck-float64(n/5)*6, false)
			n++
		}
	}
	p.rect(left, deck-6, config.TrayBack+config.Headboard-4, 18, resurrect64.Fog)
	p.line(left, deck-6, left+config.TrayBack+config.Headboard-4, deck-6, 1, resurrect64.Mist)
}

// drawMarker is a landing marker on the track below a station.
func drawMarker(p pen, x float64) {
	y := float64(config.GroundY + 6)
	p.poly(resurrect64.Gold, x-6, y, x+6, y, x, y+8)
}

// drawKoala draws a koala with its feet at (x, y): sitting, or on all
// fours when crawling.
func drawKoala(p pen, x, y float64, crawling bool) {
	if crawling {
		p.round(x-9, y-10, 18, 10, 5, resurrect64.Fog)
		p.circle(x-10, y-10, 6, resurrect64.Fog)
		p.circle(x-14, y-15, 3, resurrect64.Fog)
		p.circle(x-6, y-15, 3, resurrect64.Fog)
		p.circle(x-14, y-9, 2, resurrect64.Ink)
		return
	}
	p.round(x-7, y-12, 14, 12, 5, resurrect64.Fog)
	p.circle(x, y-16, 6, resurrect64.Fog)
	p.circle(x-6, y-20, 3.5, resurrect64.Fog)
	p.circle(x+6, y-20, 3.5, resurrect64.Fog)
	p.circle(x-6, y-20, 1.8, resurrect64.Mist)
	p.circle(x+6, y-20, 1.8, resurrect64.Mist)
	p.round(x-2, y-17, 4, 5, 2, resurrect64.Ink)
}

// drawRangers draws the two rangers holding the sheet, its middle at x,
// with faint marks between its zones.
func drawRangers(p pen, x float64) {
	half := float64(config.SheetWidth) / 2
	y := float64(config.SheetY)
	side := float64(config.RangerSide) / 2 // each ranger is RangerSide wide
	drawRanger(p, x-half-side, false)
	drawRanger(p, x+half+side, true)
	p.line(x-half-4, y, x+half+4, y, 3, resurrect64.White)
	at := x - half
	for _, z := range []float64{config.ZoneSharp, config.ZoneSafe, config.ZoneCentre, config.ZoneSafe} {
		at += z
		p.line(at, y-1.5, at, y+1.5, 1, resurrect64.Fog)
	}
}

// drawRanger is one ranger standing at x, in uniform (DESIGN §2): the
// woman has a ponytail.
func drawRanger(p pen, x float64, woman bool) {
	g := float64(config.GroundY)
	p.rect(x-5, g-6, 4, 6, resurrect64.Bark)
	p.rect(x+1, g-6, 4, 6, resurrect64.Bark)
	p.rect(x-4, g-14, 3, 8, resurrect64.Apricot)
	p.rect(x+1, g-14, 3, 8, resurrect64.Apricot)
	p.round(x-6, g-24, 12, 11, 2, resurrect64.Olive)
	p.round(x-6, g-42, 12, 20, 3, resurrect64.Khaki)
	p.circle(x, g-48, 6, resurrect64.Apricot)
	if woman {
		p.circle(x-6, g-46, 3, resurrect64.Bark)
	}
	p.round(x-10, g-53, 20, 3, 1.5, resurrect64.Khaki)
	p.round(x-5, g-58, 10, 6, 2, resurrect64.Khaki)
	p.rect(x-5, g-54, 10, 1.5, resurrect64.Pine)
}

// drawHUD is the score line across the top.
func drawHUD(ctx *scene.Context, dst *ebiten.Image, w *world.World) {
	sc, f := ctx.Scheme, ctx.Fonts.Small
	k := scale(ctx)
	y := int((config.HUDHeight*k - float64(f.LineHeight())) / 2)
	margin := int(8 * k)
	f.DrawShadowed(dst, fmt.Sprintf("SCORE %06d", w.Session.Score), margin, y, sc.Text, sc.Shadow)
	ui.TextCentred(dst, f, fmt.Sprintf("LEVEL %d", w.Level), y, sc.Text, sc.Shadow)
	lives := fmt.Sprintf("LIVES %d", w.Session.Lives)
	f.DrawShadowed(dst, lives, dst.Bounds().Dx()-margin-f.Width(lives), y, sc.Text, sc.Shadow)
}

// banner is a large message across the middle of the field, with an
// optional line under it.
func banner(ctx *scene.Context, dst *ebiten.Image, msg, sub string) {
	sc, f := ctx.Scheme, ctx.Fonts
	y := dst.Bounds().Dy()/2 - f.Large.LineHeight()
	ui.TextCentred(dst, f.Large, msg, y, sc.Title, sc.Shadow)
	if sub != "" {
		ui.TextCentred(dst, f.Small, sub, y+2*f.Large.LineHeight(), sc.Text, sc.Shadow)
	}
}
