package config

// Drop is a place koalas drop from: above a station, at a height above the
// sheet in field pixels.
type Drop struct {
	Station int
	Height  int
}

// Level is one row of the level table (DESIGN §5).
type Level struct {
	BounceTicks int    // one bounce's time in the air: the level's speed
	GapTicks    int    // between one koala dropping and the next
	UteStation  int    // where the ute parks
	Drops       []Drop // where koalas drop from
	Joeys       int    // koalas carrying a joey, worth double (DESIGN §6)
	Intro       string // the level's one new thing, on its intro card
}

// The drop points: the hanging branch, and the higher and lower branches
// that join it from levels 3 and 6. All are above the first station, so
// the ute is always far enough away (MinUteSteps).
var (
	Hanging = Drop{Station: 0, Height: 200}
	Higher  = Drop{Station: 0, Height: 250}
	Lower   = Drop{Station: 0, Height: 160}
)

// Levels is the level table, from level 1. Each level adds one new thing
// (DESIGN §5). Past its end, the last row repeats ever faster (LevelAt).
var Levels = []Level{
	// Level 1 teaches: one koala in the air at a time.
	{BounceTicks: 96, GapTicks: 6 * TPS, UteStation: 3, Drops: []Drop{Hanging}, Intro: "CATCH THE KOALAS!"},
	{BounceTicks: 90, GapTicks: 5 * TPS, UteStation: 3, Drops: []Drop{Hanging}, Intro: "A LITTLE FASTER"},
	{BounceTicks: 90, GapTicks: 5 * TPS, UteStation: 3, Drops: []Drop{Hanging, Higher}, Intro: "A HIGHER BRANCH!"},
	{BounceTicks: 90, GapTicks: 5 * TPS, UteStation: 4, Drops: []Drop{Hanging, Higher}, Intro: "THE UTE MOVES BACK"},
	{BounceTicks: 86, GapTicks: 9 * TPS / 2, UteStation: 4, Drops: []Drop{Hanging, Higher}, Joeys: 1, Intro: "A JOEY! WORTH DOUBLE"},
	{BounceTicks: 86, GapTicks: 9 * TPS / 2, UteStation: 4, Drops: []Drop{Hanging, Higher, Lower}, Joeys: 1, Intro: "A LOWER BRANCH!"},
	{BounceTicks: 84, GapTicks: 5 * TPS, UteStation: 5, Drops: []Drop{Hanging, Higher, Lower}, Joeys: 2, Intro: "THE UTE MOVES BACK AGAIN"},
}

// The endless levels past the table (DESIGN §5): each is SpeedNudge ticks
// a bounce quicker than the one before, down to FastestBounce.
const (
	SpeedNudge    = 2
	FastestBounce = 60
	EndlessIntro  = "FASTER!"
)

// MinUteSteps is how close the ute may park to a drop point, in stations:
// no nearer than 3, so no single catch reaches the tray and every rescue
// needs the rangers to move (DESIGN §5).
const MinUteSteps = 3

// LevelAt is level n's row, counting from 1. Past the table the last row
// repeats with its bounce nudged quicker each level.
func LevelAt(n int) Level {
	n = max(n, 1)
	if n <= len(Levels) {
		return Levels[n-1]
	}
	l := Levels[len(Levels)-1]
	l.BounceTicks = max(l.BounceTicks-(n-len(Levels))*SpeedNudge, FastestBounce)
	l.Intro = EndlessIntro
	return l
}
