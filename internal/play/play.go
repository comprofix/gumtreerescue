// Package play is Gum Tree Rescue's play scene: it runs the world from the
// player's input and draws it. It is native in both looks (PLAN §3): the
// world works in field pixels, and the scene draws them at the shown
// screen's scale, ×2 in the smooth look and ×1 in the pixel look.
package play

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/audio"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/input"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/session"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// Sounds are the game's sounds, for the sound test. They come with the
// audio (PLAN M5).
var Sounds = audio.Bank{}

// Play is a game in progress.
type Play struct {
	world *world.World
	mouse bool // the mouse is steering, until a key or the pad is used
}

// New starts a game at a skill, on the level the -level flag asks for.
func New(ctx *scene.Context, skill config.Skill) scene.Scene {
	ctx.Session = session.New(skill, config.StartLives)
	return &Play{world: world.New(ctx.Session, ctx.StartLevel, ctx.RNG)}
}

// Native implements scene.Native: the scene draws in either look.
func (*Play) Native() {}

// World is the game's state, for tests.
func (p *Play) World() *world.World { return p.world }

// Update implements scene.Scene.
func (p *Play) Update(ctx *scene.Context) error {
	in := ctx.Input
	switch {
	case in.JustPressed(input.Pause):
		ctx.Cut(scene.NewPause(p))
		return nil
	case in.JustPressed(input.Menu):
		ctx.Switch(scene.NewMainMenu())
		return nil
	case in.JustPressed(input.Restart):
		ctx.Switch(New(ctx, ctx.Session.Skill))
		return nil
	}
	for _, e := range p.world.Update(p.input(ctx)) {
		if e.Kind == world.GameOver {
			ctx.Cut(newEnded(p))
		}
	}
	return nil
}

// input is the player's steering this tick: Left and Right, or the mouse,
// which heads the sheet to the pointer at the rangers' own speed (DESIGN
// §4). Moving the mouse hands steering to it; a direction takes it back.
func (p *Play) input(ctx *scene.Context) world.Input {
	in := ctx.Input
	var dir int
	if in.Held(input.Left) {
		dir--
	}
	if in.Held(input.Right) {
		dir++
	}
	if in.Held(input.Left) || in.Held(input.Right) {
		p.mouse = false
		return world.Input{Dir: dir}
	}
	pt, ok := in.Pointer()
	if in.PointerMoved() && ok {
		p.mouse = true
	}
	if !p.mouse {
		return world.Input{}
	}
	return world.Input{Aim: true, Target: int(float64(pt.X) / scale(ctx) * config.SubPx)}
}

// scale is how many screen pixels the shown look has to a field pixel.
func scale(ctx *scene.Context) float64 {
	fw, _ := config.Field()
	return float64(ctx.Preset.ScreenWidth) / float64(fw)
}

// Ended holds the finished game on screen, then goes on to the high
// scores. Confirm or a click goes on sooner.
type Ended struct {
	play  *Play
	ticks int
}

func newEnded(p *Play) *Ended { return &Ended{play: p} }

// Native implements scene.Native.
func (*Ended) Native() {}

// Update implements scene.Scene.
func (g *Ended) Update(ctx *scene.Context) error {
	g.ticks++
	in := ctx.Input
	if g.ticks >= config.GameOverHold || (g.ticks > config.TPS/2 && (in.JustPressed(input.Confirm) || in.Clicked())) {
		s := ctx.Session
		ctx.Switch(scene.AfterGame(ctx, s.Skill, s.Score))
	}
	return nil
}

// Draw implements scene.Scene.
func (g *Ended) Draw(ctx *scene.Context, dst *ebiten.Image) {
	g.play.Draw(ctx, dst)
	scene.Dim(dst, ctx.Scheme.Shadow, 0.5)
	banner(ctx, dst, "GAME OVER", "")
}
