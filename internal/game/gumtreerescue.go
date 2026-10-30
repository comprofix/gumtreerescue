package game

import "github.com/comprofix/gumtreerescue/internal/play"

// Gum Tree Rescue's place in the game: its play scene, drawn for both
// looks (PLAN §3), and its sounds.
func init() {
	content = Content{
		NewPlay:    play.New,
		NativePlay: true,
		Sounds:     play.Sounds,
		Sheets:     []string{"koala"},
		TitleTune:  play.TitleTune,
		ScoresTune: play.ScoresTune,
	}
}
