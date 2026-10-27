package config

// Skill is a difficulty level the player picks before a game. Gum Tree
// Rescue's skills change only how much help the player gets, never the
// rules, so there is one level table (DESIGN §5).
type Skill struct {
	ID   string // stable key, used in saved files
	Name string // shown in menus
	Desc string // one line shown under the menu

	Reach      int  // how far the rangers move in a bounce, in hundredths of a step
	Lives      int  // lives at the start
	LateMarker bool // the landing marker shows only for the last third of a bounce
}

// Skills are the skill levels, easiest first.
var Skills = []Skill{
	{ID: "easy", Name: "Easy", Desc: "Quicker rangers and five lives.", Reach: 320, Lives: 5},
	{ID: "normal", Name: "Normal", Desc: "The game as it is meant to be played.", Reach: RangerReach, Lives: 3},
	{ID: "hard", Name: "Hard", Desc: "Landing markers show only at the last moment.", Reach: RangerReach, Lives: 3, LateMarker: true},
}
