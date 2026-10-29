package play

import (
	"fmt"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/shapes"
	"github.com/comprofix/gumtreerescue/internal/ui"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// The play scene's pictures come from package art as they are drawn; the
// rest are still simple shapes until theirs are (PLAN M4). Every size here
// is in field pixels.

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
	tick := int(ctx.Tick)
	// Koalas going round the back of the trunk pass behind it (DESIGN §5).
	for i := range w.Koalas {
		if k := &w.Koalas[i]; k.Behind() {
			drawKoala(pn, w, k, tick)
		}
	}
	p.tree.draw(dst, pn.k, w.Row.Drops)
	drawWaiting(pn, w.Tree, tick)
	ute := px(w.UteX()) + uteShift(w)
	drawSmoke(pn, w, p.drive)
	p.ute.draw(dst, pn.k, ute, p.drive.speed())
	drawTray(pn, w, ute)
	drawDust(pn, w, p.drive)
	late := w.Session.Skill.LateMarker
	for i := range w.Koalas {
		k := &w.Koalas[i]
		// On Hard the marker shows only for a bounce's last third (DESIGN §5).
		if station, ticks, ok := w.Landing(k); ok && (!late || 3*ticks <= w.Row.BounceTicks) {
			drawMarker(pn, px(world.StationX(station)))
		}
	}
	// Koalas on the ground pass in front of the rangers (DESIGN §5). Lost
	// joeys are the rangers' once the level is over (DESIGN §6).
	ground := func(k *world.Koala) bool { return k.State == world.Dazed || k.State == world.Crawling }
	for i := range w.Koalas {
		if k := &w.Koalas[i]; k.State != world.InTray && !ground(k) && !k.Behind() && !(k.Baby && w.Ending > 0) {
			drawKoala(pn, w, k, tick)
		}
	}
	drawRangers(pn, w, &p.crew, tick)
	for i := range w.Koalas {
		if k := &w.Koalas[i]; ground(k) && !(k.Baby && w.Ending > 0) {
			drawKoala(pn, w, k, tick)
		}
	}
	drawHUD(ctx, dst, w)
	switch {
	case w.Starting > 0:
		banner(ctx, dst, fmt.Sprintf("LEVEL %d", w.Level), w.Row.Intro)
	case w.Ending > 0 && w.Ending <= config.EndTicks/2:
		banner(ctx, dst, "WELL DONE!", tally(w)...)
	}
}

// uteShift is how far right of its station the ute is drawn: it drives in
// under the intro card, and off with the tray full before the tally
// (DESIGN §7).
func uteShift(w *world.World) float64 {
	var f float64 // 0 parked, 1 off the screen
	half := config.EndTicks / 2
	switch {
	case w.Starting > 0:
		f = float64(w.Starting) / config.LevelTicks
	case w.Ending > config.EndTicks: // a ranger carries lost joeys to it
	case w.Ending > half:
		f = 1 - float64(w.Ending-half)/float64(half)
	case w.Ending > 0:
		f = 1
	}
	return config.UteDrivePx * f * f
}

// tally is the level-end tally's lines: the rescues, the no-miss bonus or
// the misses, and any extra life (DESIGN §7).
func tally(w *world.World) []string {
	lines := []string{fmt.Sprintf("KOALAS RESCUED  %d", w.Rescued)}
	if w.Misses == 0 {
		lines = append(lines, fmt.Sprintf("NO MISSES  +%d", config.CleanBonus))
	} else {
		lines = append(lines, fmt.Sprintf("MISSES  %d", w.Misses))
	}
	if w.Gained > 0 {
		lines = append(lines, "EXTRA LIFE!")
	}
	return lines
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

// treeCache is the gum tree drawn once for a level's drop points at the
// shown screen's scale, as it is slow to draw (DESIGN §3); Draw lays it.
type treeCache struct {
	img   *ebiten.Image
	k     float64
	drops []config.Drop
}

// draw lays the tree with a branch out to each drop point on dst, drawing
// it first if the scale or the drop points have changed.
func (c *treeCache) draw(dst *ebiten.Image, k float64, drops []config.Drop) {
	b := dst.Bounds()
	if c.img == nil || c.k != k || !slices.Equal(c.drops, drops) || c.img.Bounds() != b {
		if c.img == nil || c.img.Bounds() != b {
			c.img = ebiten.NewImage(b.Dx(), b.Dy())
		}
		c.img.Clear()
		c.k, c.drops = k, slices.Clone(drops)
		// Each branch along the line koalas walk, from the tree's middle,
		// where it grows, out.
		branches := make([][]float64, len(drops))
		for i, d := range drops {
			b := config.Branch(d)
			base := b[len(b)-1].Y
			branches[i] = append(branches[i], float64(config.SpineX(base)), float64(base))
			for j := len(b) - 1; j >= 0; j-- {
				branches[i] = append(branches[i], float64(b[j].X), float64(b[j].Y))
			}
		}
		boughs := make([][]float64, len(config.Boughs))
		for i, b := range config.Boughs {
			for _, s := range b {
				boughs[i] = append(boughs[i], float64(s.X), float64(s.Y))
			}
		}
		art.Tree(art.Pen{Dst: c.img, K: k}, branches, boughs)
	}
	dst.DrawImage(c.img, nil)
}

// sitPose is how a koala waits in its spot (DESIGN §5).
type sitPose int

const (
	onBough     sitPose = iota // sitting side on on a bough, feet over its front
	alongLimb                  // holding on along a sloping limb
	inFork                     // curled up asleep in a fork
	eating                     // chewing a gum leaf
	onTrunk                    // holding on to a side of the trunk
	clingEating                // holding on to a limb's side, chewing a gum leaf
	atFoot                     // sitting at the foot of the tree
)

// spotPoses is the pose for each of config.Spots, to suit the wood there.
var spotPoses = [config.KoalaLevel]sitPose{
	alongLimb, inFork, clingEating, onBough, onTrunk,
	onTrunk, onTrunk, atFoot, eating, inFork,
}

// limbSlope is how far a koala gripping the limb at spot s turns to lie
// along it, facing up it, and whether it faces left.
func limbSlope(s config.Spot) (float64, bool) {
	down := s.Way[0]
	ux, up := float64(s.X-down.X), float64(down.Y-s.Y)
	return -math.Atan2(up, math.Abs(ux)), ux < 0
}

// onLeft reports whether a koala at x, y on the trunk holds on to its left
// side, facing right; on its right side it faces left.
func onLeft(x, y int) bool { return config.SideOf(x, y) == config.Left }

// idle is a koala's look as it sits about: a blink every few seconds, now
// and then a flick of an ear, and a slow sway, each koala in its own time.
func idle(id, tick int) art.Look {
	var l art.Look
	blink, flick := 170+id*37%130, 400+id*53%250
	l.Shut = (tick+id*67)%blink < 6
	if f := (tick + id*97) % flick; f < 12 {
		l.Flick = math.Sin(math.Pi * float64(f) / 12)
	}
	l.Lean = 0.035 * math.Sin(float64(tick)/70+float64(id))
	return l
}

// drawWaiting draws the koalas waiting in the tree, each in its own spot
// and the pose that goes with it (DESIGN §5).
func drawWaiting(p pen, tree []world.Waiting, tick int) {
	for _, s := range tree {
		at := config.Spots[s.Spot]
		a := art.Pen{Dst: p.dst, K: p.k}.At(float64(at.X), float64(at.Y))
		l := idle(s.ID, tick)
		l.Joey = s.Joey
		switch spotPoses[s.Spot] {
		case onBough:
			art.OnBough(a, l)
		case alongLimb:
			turn, left := limbSlope(at)
			art.Gripping(a.Turned(turn, 0).Mirrored(left), l, 0, 0)
		case inFork:
			art.Asleep(a, l)
		case clingEating:
			art.ClingChewing(a.Mirrored(!onLeft(at.X, at.Y)), l, float64(tick)/5)
		case eating:
			art.Chewing(a, l, float64(tick)/5)
		case onTrunk:
			art.Clinging(a.Mirrored(!onLeft(at.X, at.Y)), l, 0, false)
		default:
			art.Koala(a, l)
		}
	}
}

// drawKoala draws a koala out of its spot, or a lost joey, as it is doing:
// bounding up or down the tree, walking a branch or the ground, teetering
// at its drop point, falling with its limbs spread, tumbling over in a
// bounce, dazed, or a joey looking about for its mother (DESIGN §5, §6).
func drawKoala(p pen, w *world.World, k *world.Koala, tick int) {
	a := art.Pen{Dst: p.dst, K: p.k}.At(px(k.X), koalaY(k))
	l := idle(k.ID, tick)
	l.Joey = k.Joey
	if k.Baby {
		a = a.Sized(art.JoeySize)
		l = idle(k.ID+config.KoalaLevel, tick)
	}
	switch k.State {
	case world.Crawling:
		art.Walking(a.Mirrored(true), l, px(k.X)*0.7)
	case world.Walking, world.Climbing, world.Finding:
		drawMoving(a, k, l)
	case world.Wobbling:
		art.Walking(a.Turned(0.14*math.Sin(float64(k.Timer)*0.9), 0), l, 0)
	case world.Falling:
		art.Falling(a, l, math.Sin(float64(tick)*0.5))
	case world.Arcing:
		// One somersault a bounce, the way it is going (DESIGN §5).
		turn := 0.0
		if dx, _ := k.Heading(); dx != 0 {
			turn = 2 * math.Pi * float64(k.Since()) / float64(w.Row.BounceTicks)
			if dx < 0 {
				turn = -turn
			}
		}
		art.Falling(a.Turned(turn, art.FallingMiddle), l, 0)
	case world.Thumped:
		art.Thump(a.Mirrored(true), l, float64(k.Timer)/config.ThumpTicks)
	case world.Dazed:
		l.Lean = 0.08 * math.Sin(float64(tick)*0.15)
		art.Dazed(a, l)
		art.Dizzy(a, tick)
	case world.Searching:
		l.Lean = 0.2 * math.Sin(float64(tick)*0.06)
		art.Koala(a, l)
	default:
		art.Koala(a, l)
	}
}

// gripStride is how far across a koala gripping along a sloping limb or
// branch goes for each reach of its paws, in field pixels.
const gripStride = 8

// drawMoving draws a koala going about the tree along its path (DESIGN
// §5): up or down the trunk or a steep limb it clings to its side and
// bounds, arms then legs, backing down bottom first; round the back of the
// trunk it shimmies, still clinging and facing the wood; along a sloping limb
// or branch it grips the wood facing up it, reaching ahead then bringing
// its back feet up, or backing down feet first; along a near-level branch
// it walks on top.
func drawMoving(a art.Pen, k *world.Koala, l art.Look) {
	dx, dh := k.Heading()
	f := 0.0 // how far through the moving half of its bound, 0 when still
	if h := k.Hop(); 2*h < config.HopTicks {
		f = 2 * float64(h) / config.HopTicks
	}
	slope := -math.Atan2(float64(dh), float64(abs(dx)))
	switch {
	case k.Behind():
		left := onLeft(k.X/config.SubPx, int(koalaY(k)))
		art.Clinging(a.Mirrored(!left), l, math.Sin(math.Pi*f), false)
	case abs(dh) > 2*abs(dx) || dx == 0:
		left := onLeft(k.X/config.SubPx, int(koalaY(k)))
		art.Clinging(a.Mirrored(!left), l, math.Sin(math.Pi*f), dh < 0)
	case 3*abs(dh) > abs(dx):
		// A steady grip, paws then feet, one every gripStride across.
		g := px(k.X) / gripStride
		if dx < 0 {
			g = -g
		}
		g -= math.Floor(g)
		first, then := math.Sin(math.Pi*min(2*g, 1)), math.Sin(math.Pi*max(2*g-1, 0))
		up, reach, follow := dx > 0, first, then
		if dh < 0 {
			up, reach, follow = dx < 0, then, first
		}
		art.Gripping(a.Turned(-math.Atan2(float64(abs(dh)), float64(abs(dx))), 0).Mirrored(!up), l, reach, follow)
	default:
		art.Walking(a.Turned(slope, 0).Mirrored(dx < 0), l, px(k.X)*0.7)
	}
}

func abs(n int) int { return max(n, -n) }

// The carry at a level's end, as parts of config.CarryTicks (DESIGN §6):
// lost joeys climb down to the foot of the tree while the ranger walks
// over, she picks them up, carries them to the ute and puts them in.
const (
	carryDown = 0.3
	carryPick = 0.4
	carryPut  = 0.9
)

// carrying is how far through the carry the level's end is, from 0 to 1,
// and whether it is under way.
func carrying(w *world.World) (float64, bool) {
	if w.Ending <= config.EndTicks {
		return 0, false
	}
	return 1 - float64(w.Ending-config.EndTicks)/config.CarryTicks, true
}

// joeysInTray reports whether the level's lost joeys are in the ute.
func joeysInTray(w *world.World) bool {
	c, ok := carrying(w)
	return w.Ending > 0 && (!ok || c >= carryPut)
}

// rangerAway is where the ranger carrying lost joeys is, if she is away
// from the sheet: to the tree, to the ute, and back while it drives off.
func rangerAway(w *world.World) (float64, bool) {
	half := config.EndTicks / 2
	if w.LostJoeys() == 0 || w.Ending <= half {
		return 0, false
	}
	_, home := rangerHomes(w)
	tree := float64(config.TrunkX + 12)
	ute := px(w.UteX()) - config.TrayBack - 6
	c, ok := carrying(w)
	switch {
	case !ok:
		return lerp(ute, home, float64(config.EndTicks-w.Ending)/float64(config.EndTicks-half)), true
	case c < carryDown:
		return lerp(home, tree, c/carryDown), true
	case c < carryPick:
		return tree, true
	case c < carryPut:
		return lerp(tree, ute, (c-carryPick)/(carryPut-carryPick)), true
	}
	return ute, true
}

func lerp(a, b, f float64) float64 { return a + (b-a)*min(max(f, 0), 1) }

// rangerHomes is where the two rangers stand, either side of the sheet.
func rangerHomes(w *world.World) (left, right float64) {
	x := px(w.Sheet)
	off := float64(config.SheetWidth)/2 + float64(config.RangerSide)/2
	return x - off, x + off
}

// drawCarried draws the lost joeys during the carry: climbing down to the
// foot of the tree, waiting there, then in the arms of the ranger at x,
// facing left if left.
func drawCarried(p pen, w *world.World, x float64, left bool, tick int) {
	c, ok := carrying(w)
	if !ok || c >= carryPut {
		return
	}
	foot := float64(config.TrunkX + 4)
	n := 0
	for i := range w.Koalas {
		k := &w.Koalas[i]
		if !k.Baby {
			continue
		}
		l := idle(k.ID+config.KoalaLevel, tick)
		at := art.Pen{Dst: p.dst, K: p.k}
		switch {
		case c < carryDown:
			f := c / carryDown
			y := lerp(koalaY(k), config.GroundY, f)
			reach := math.Abs(math.Sin(y * 0.3))
			art.Clinging(at.At(lerp(px(k.X), foot+float64(n)*7, f), y).Sized(art.JoeySize).Mirrored(true), l, reach, true)
		case c < carryPick:
			art.Koala(at.At(foot+float64(n)*7, config.GroundY).Sized(art.JoeySize), l)
		default:
			// Sitting on her forearms, in front of her.
			f := 1.0
			if left {
				f = -1
			}
			art.Koala(at.At(x+f*(4+float64(n)*4), config.GroundY+carryHold).Sized(art.JoeySize), l)
		}
		n++
	}
}

// uteCache is the ranger ute drawn once at the shown screen's scale, but
// for its wheels and its tray's near side, as it is slow to draw (DESIGN
// §2); Draw lays it wherever the ute is and rolls its wheels.
type uteCache struct {
	img *ebiten.Image
	k   float64
}

// The ute's picture reaches this far behind its station and above its
// tray's deck, and is this big, in field pixels.
const (
	uteBehind = config.TrayBack + 4
	uteAbove  = 38
	uteWide   = uteBehind + config.Headboard + art.UteFront + 4
	uteTall   = uteAbove + art.UteWheel + 6
)

// draw lays the ute on dst with its tray's station at x, going speed field
// pixels a tick, drawing it first if the scale has changed.
func (c *uteCache) draw(dst *ebiten.Image, k, x, speed float64) {
	if c.img == nil || c.k != k {
		c.k = k
		c.img = ebiten.NewImage(int(math.Ceil(uteWide*k)), int(math.Ceil(uteTall*k)))
		art.Ute(art.Pen{Dst: c.img, K: k}.At(uteBehind, uteAbove), config.TrayBack, config.Headboard)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(math.Round((x-uteBehind)*k), math.Round((config.SheetY-uteAbove)*k))
	dst.DrawImage(c.img, op)
	// The wheels roll as it drives, turned by how far it has come.
	art.UteWheels(art.Pen{Dst: dst, K: k}.At(math.Round(x*k)/k, config.SheetY), config.TrayBack, config.Headboard, art.UteRoll(x), art.UteRoll(speed))
}

// drawTray is the ute's tray, its station at x, with the rescued koalas in
// it, drawn after the ute so they sit inside its low sides.
func drawTray(p pen, w *world.World, x float64) {
	deck := float64(config.SheetY)
	left := x - config.TrayBack
	n := 0
	for i := range w.Koalas {
		if k := &w.Koalas[i]; k.State == world.InTray {
			a := art.Pen{Dst: p.dst, K: p.k}.At(left+9+float64(n%5)*16, deck-float64(n/5)*6)
			l := art.Look{Joey: k.Joey}
			if k.Bonked && k.Timer > 0 {
				art.Dazed(a, l)
				art.Dizzy(a, k.Timer)
			} else {
				art.Koala(a, l)
			}
			n++
		}
	}
	// Lost joeys a ranger has carried to the back of the ute (DESIGN §6).
	if joeysInTray(w) {
		for j := range w.LostJoeys() {
			art.Koala(art.Pen{Dst: p.dst, K: p.k}.At(left+6+float64(j)*8, deck-4).Sized(art.JoeySize), art.Look{})
		}
	}
	art.TraySide(art.Pen{Dst: p.dst, K: p.k}.At(x, deck), config.TrayBack, config.Headboard)
}

// drawMarker is a landing marker on the track below a station.
func drawMarker(p pen, x float64) {
	y := float64(config.GroundY + 6)
	p.poly(resurrect64.Gold, x-6, y, x+6, y, x, y+8)
}

// How the rangers hold the sheet and the joeys, in field pixels: the sheet
// sags this far in its middle, and a ranger carrying joeys holds them on
// her forearms this far above the ground, or picks them up this low.
const (
	sheetSag  = 1.5
	carryHold = -33
	carryLow  = -20
)

// drawRangers draws the two rangers, the man on the left and the woman on
// the right, facing each other and holding up the sheet between them, a
// seam across it between each of its zones (DESIGN §2, §5); at a level's
// end she may be away carrying lost joeys to the ute (DESIGN §6).
func drawRangers(p pen, w *world.World, c *crew, tick int) {
	x := px(w.Sheet)
	half := float64(config.SheetWidth) / 2
	side := float64(config.RangerSide) / 2 // each ranger is RangerSide wide
	ground := art.Pen{Dst: p.dst, K: p.k}
	left, right := rangerHomes(w)
	hold := float64(art.RangerHold) - sheetSag
	stance := func(k walker, woman bool, id int) art.Stance {
		return art.Stance{
			Woman: woman, HandX: side, HandY: hold,
			Stride: k.stride, Step: k.step, Shut: (tick+id*89)%230 < 6,
		}
	}
	man := stance(c.man, false, 0)
	if away, ok := rangerAway(w); ok {
		// Without her, he lets his arms down to his sides.
		man.HandX, man.HandY = 2, -29
		art.Ranger(ground.At(left, config.GroundY), man)
		woman := stance(c.woman, true, 1)
		woman.HandX, woman.HandY = 5.5, carryHold
		if cc, ok := carrying(w); ok && cc >= carryDown && cc < carryPick {
			woman.HandX, woman.HandY = 6, carryLow
		}
		art.Ranger(ground.At(away, config.GroundY).Mirrored(c.woman.left), woman)
		drawCarried(p, w, away, c.woman.left, tick)
		return
	}
	var seams []float64
	at := x - half
	for _, z := range []float64{config.ZoneSharp, config.ZoneSafe, config.ZoneCentre, config.ZoneSafe} {
		at += z
		seams = append(seams, at)
	}
	dx, dip := c.dip()
	art.Sheet(ground.At(0, config.SheetY-sheetSag), x-half, x+half, sheetSag, x+dx, dip, seams)
	art.Ranger(ground.At(left, config.GroundY), man)
	art.Ranger(ground.At(right, config.GroundY).Mirrored(true), stance(c.woman, true, 1))
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

// banner is a large message across the middle of the field, with any lines
// under it.
func banner(ctx *scene.Context, dst *ebiten.Image, msg string, lines ...string) {
	sc, f := ctx.Scheme, ctx.Fonts
	y := dst.Bounds().Dy()/2 - f.Large.LineHeight()
	ui.TextCentred(dst, f.Large, msg, y, sc.Title, sc.Shadow)
	y += 2 * f.Large.LineHeight()
	for _, l := range lines {
		ui.TextCentred(dst, f.Small, l, y, sc.Text, sc.Shadow)
		y += f.Small.LineHeight() * 3 / 2
	}
}
