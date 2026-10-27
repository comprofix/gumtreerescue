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
}

// Levels is the level table, from level 1. Past its end, the last row
// repeats.
var Levels = []Level{
	// Level 1 teaches: one koala in the air at a time.
	{BounceTicks: 96, GapTicks: 6 * TPS, UteStation: 3, Drops: []Drop{{Station: 0, Height: 200}}},
}

// MinUteSteps is how close the ute may park to a drop point, in stations:
// no nearer than 3, so no single catch reaches the tray and every rescue
// needs the rangers to move (DESIGN §5).
const MinUteSteps = 3

// LevelAt is level n's row, counting from 1.
func LevelAt(n int) Level {
	return Levels[min(max(n, 1), len(Levels))-1]
}
