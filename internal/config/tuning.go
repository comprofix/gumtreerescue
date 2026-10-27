package config

// Gum Tree Rescue's tuning (DESIGN §5). Distances are in field pixels: the
// field is the pixel look's screen, which the smooth look shows at twice
// the size, so play is the same in either look. Times are in ticks. The
// numbers are a starting point, tuned in play.

// Field is the size of the play field, in field pixels: the Modern
// preset's screen.
func Field() (w, h int) { return Modern.ScreenWidth, Modern.ScreenHeight }

// SubPx is how many fixed-point units make one field pixel. The world
// keeps every position in these, so arcs land exactly (DESIGN §5).
const SubPx = 256

// The ground, the tree and the stations koalas land on (DESIGN §5).
const (
	HUDHeight   = 24  // the score line across the top
	GroundY     = 330 // the top of the ground
	SheetHeight = 30  // the sheet, and the ute's tray, above the ground
	TrunkX      = 62  // the trunk's right side: a koala bounced into it grabs the tree
	Station0X   = 136 // the first station, under the hanging branch
	StepPx      = 80  // from one station to the next: a safe arc's length
	PeakPx      = 150 // every bounce rises this far above the sheet
)

// SheetY is the height of the sheet and the ute's tray, in field pixels
// from the top.
const SheetY = GroundY - SheetHeight

// The sheet and its five zones, left to right, in field pixels (DESIGN §5):
// sharp, safe, centre, safe, sharp. They add up to SheetWidth, a little
// under one step.
const (
	ZoneSharp  = 8
	ZoneSafe   = 18
	ZoneCentre = 12
	SheetWidth = 2*ZoneSharp + 2*ZoneSafe + ZoneCentre
)

// ZoneSteps is how far a catch in each zone sends a koala, in steps,
// zone 1 to zone 5.
var ZoneSteps = [5]int{-2, -1, 0, 1, 2}

// RangerReach is how far the rangers can move in one bounce on Normal, in
// hundredths of a step: just over 2 steps, so a sharp slip can be caught
// with a quick reaction (DESIGN §5). The drop scheduler plans for it on
// every skill, so skills change only help.
const RangerReach = 220

// RangerSide is how far each ranger reaches past the end of the sheet, in
// field pixels. The rangers stop where one would walk into the trunk or the
// ute's tray (DESIGN §5).
const RangerSide = 12

// The ute, measured from the station it is parked on, in field pixels
// (DESIGN §5). The tray runs from TrayBack behind the station to the
// headboard, which a koala aimed past the tray bonks.
const (
	TrayBack  = 14
	Headboard = 76
)

// Spot is a place in the tree, as a koala's feet, in field pixels from the
// top left, and Way, the bends of its limb on the way in to the trunk.
type Spot struct {
	X, Y int
	Way  []Spot
}

// limb1, limb0 and limb3 are the bends along the tree's limbs, along their
// tops: the right limb up into the canopy, the left one, and the bough to
// the left.
var (
	limb1 = []Spot{{X: 98, Y: 81}, {X: 68, Y: 113}, {X: 46, Y: 152}}
	limb0 = []Spot{{X: 20, Y: 110}, {X: 28, Y: 150}}
	limb3 = []Spot{{X: 26, Y: 122}}
)

// Boughs are short, level boughs off the right limb, out past the leaves,
// as the lines along their middles from the limb out (DESIGN §3): drawn
// over the canopy, so the koalas up there sit on wood, not leaves.
var Boughs = [][]Spot{
	{{X: 78, Y: 104}, {X: 100, Y: 106}, {X: 124, Y: 106}},
	{{X: 110, Y: 78}, {X: 132, Y: 72}, {X: 150, Y: 73}},
}

// Spots are where waiting koalas sit in the tree, one each: on its limbs
// and boughs, in the forks, holding the left sides of the middle limb and
// the trunk, clear of the climbers on their right, and at its foot (DESIGN
// §5). A level's koalas start in them in order, the first to drop in the
// first.
var Spots = [KoalaLevel]Spot{
	{X: 70, Y: 112, Way: limb1[2:]},
	{X: 12, Y: 104, Way: limb3},
	{X: 30, Y: 97},
	{X: 114, Y: 104, Way: append([]Spot{{X: 78, Y: 102}}, limb1[1:]...)},
	{X: 17, Y: 190},
	{X: 15, Y: 236},
	{X: 14, Y: 280},
	{X: 16, Y: 331},
	{X: 134, Y: 69, Way: append([]Spot{{X: 110, Y: 74}}, limb1...)},
	{X: 4, Y: 68, Way: limb0},
}

// JoeyPerches are where a lost joey sits while it looks for its mother
// (DESIGN §6): more than a level ever has joeys.
var JoeyPerches = [3]Spot{{X: 84, Y: 96, Way: limb1[1:]}, {X: 44, Y: 56}, {X: 28, Y: 150}}

// Spine is the tree's middle, from the foot of the trunk up it and on up
// the middle limb: the branches grow from it, and the ways in from left of
// it come round the back of the trunk.
var Spine = []Spot{{X: 42, Y: 334}, {X: 40, Y: 150}, {X: 35, Y: 128}, {X: 40, Y: 80}, {X: 46, Y: 40}}

// SpineX is where the Spine is across at height y.
func SpineX(y int) int { return along(Spine, y) }

// The trunk's sides: a koala holds on to its left side facing right, and to
// its right side facing left.
const (
	Left = iota
	Right
)

// ClimbLine is where a koala's feet are as it holds on to the trunk's right
// side, side on, and climbs it, on up the middle limb's right (DESIGN §5).
// Every way in the tree meets it, those from the left side round the back.
var ClimbLine = []Spot{{X: 66, Y: 334}, {X: 67, Y: 320}, {X: 63, Y: 286}, {X: 59, Y: 236}, {X: 55, Y: 186}, {X: 49, Y: 150}, {X: 44, Y: 130}, {X: 48, Y: 80}, {X: 53, Y: 40}}

// ClimbX is where a koala climbing the trunk is across at height y. A
// koala bounced into the tree grabs it there (DESIGN §5).
func ClimbX(y int) int { return along(ClimbLine, y) }

// SideOf is the side of the trunk a way in from x, y comes to.
func SideOf(x, y int) int {
	if x < SpineX(y) {
		return Left
	}
	return Right
}

// along is where a line going up through the spots, from the bottom, is
// across at height y; past either end it carries straight on up or down.
func along(line []Spot, y int) int {
	for i := 1; i < len(line); i++ {
		a, b := line[i-1], line[i]
		if y >= b.Y {
			return a.X + (b.X-a.X)*(a.Y-y)/(a.Y-b.Y)
		}
	}
	return line[len(line)-1].X
}

// BranchRise is how far a branch rises from the Spine to its tip, the drop
// point.
const BranchRise = 34

// Branch is drop point d's branch, as the line along its middle that
// koalas walk and the art follows: from its tip, sagging a little, to
// where it meets the trunk's right side. The art carries it on into the
// Spine.
func Branch(d Drop) []Spot {
	x, y := Station0X+d.Station*StepPx, SheetY-d.Height
	return []Spot{{X: x, Y: y}, {X: x - 20, Y: y + 1}, {X: x - 52, Y: y + 10}, {X: ClimbX(y + BranchRise), Y: y + BranchRise}}
}

// What happens to a missed koala, a koala bounced into the tree and a lost
// joey (DESIGN §5, §6).
const (
	DazedTicks   = 90            // sitting dazed on the ground
	CrawlSubPx   = 3 * SubPx / 4 // crawling back to the tree, per tick
	ClimbSubPx   = 3 * SubPx / 4 // climbing about the tree, per tick on average
	HopTicks     = 2 * TPS / 5   // a koala bounds up and down the trunk: twice ClimbSubPx a tick for half of this, then still for half
	SlipChance   = 50            // in a hundred: a koala bounced into the tree slips, or grabs on
	ThumpTicks   = TPS / 5       // a koala bounced into the tree is squashed against it this long
	JoeyTumblePx = 10            // a lost joey lands this far right of its mother
	JoeySearch   = 2 * TPS       // a lost joey looks for its mother at least this long,
	JoeySearchUp = 6 * TPS       // and up to this much longer, at random
	CarryTicks   = 3 * TPS       // at a level's end, a ranger carries lost joeys to the ute
	BonkedTicks  = 2 * TPS       // dazed in the tray after a headboard bonk
	LevelTicks   = 5 * TPS / 2   // the intro card, while the ute drives in
	EndTicks     = 4 * TPS       // the ute drives off, and the level's tally
	FirstDrop    = TPS / 2       // after the intro, until the first drop
	WobbleTicks  = TPS / 2       // a koala wobbles at its drop point before it falls
	DropRetry    = TPS / 10      // a drop held back by the scheduler tries again after this
	DropJitter   = TPS / 2       // up to this much added to a level's gap at random
	DropWait     = 2 * TPS       // a drop waits no longer than this for the rangers to come back
	KoalaLevel   = 10            // koalas rescued to finish a level
	GameOverHold = 2 * TPS       // the finished game stays on screen
)

// UteDrivePx is how far the ute drives in at a level's start and off at its
// end, in field pixels: from wholly off the right edge.
const UteDrivePx = 420

// Points and lives (DESIGN §5).
const (
	SafePoints   = 10    // a safe-right bounce, zone 4
	SharpPoints  = 20    // a sharp-right bounce, zone 5
	RescuePoints = 100   // a koala in the tray
	CleanBonus   = 500   // a level with no misses
	ExtraLife    = 5_000 // an extra life at this score, and every this many after
	MaxLives     = 5     // extra lives stop here
)
