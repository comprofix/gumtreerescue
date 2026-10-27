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
// with a quick reaction (DESIGN §5).
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

// What happens to a missed koala, and a koala that grabs the tree (DESIGN
// §5).
const (
	DazedTicks   = 90            // sitting dazed on the ground
	CrawlSubPx   = 3 * SubPx / 4 // crawling back to the tree, per tick
	ClimbTicks   = 2 * TPS       // climbing back up to wait again
	BonkedTicks  = 2 * TPS       // dazed in the tray after a headboard bonk
	LevelTicks   = 2 * TPS       // the pause before a level starts
	FirstDrop    = TPS / 2       // after that pause, until the first drop
	WobbleTicks  = TPS / 2       // a koala wobbles at its drop point before it falls
	KoalaLevel   = 10            // koalas rescued to finish a level
	GameOverHold = 2 * TPS       // the finished game stays on screen
)

// Points (DESIGN §5).
const (
	SafePoints   = 10  // a safe-right bounce, zone 4
	SharpPoints  = 20  // a sharp-right bounce, zone 5
	RescuePoints = 100 // a koala in the tray
	CleanBonus   = 500 // a level with no misses
)
