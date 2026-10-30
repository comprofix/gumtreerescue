// Package play is Gum Tree Rescue's play scene: it runs the world from the
// player's input and draws it. It is native in both looks (PLAN §3): the
// world works in field pixels, and the scene draws them at the shown
// screen's scale, ×2 in the smooth look and ×1 in the pixel look.
package play

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/input"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/session"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// Play is a game in progress.
type Play struct {
	world *world.World
	mouse bool          // the mouse is steering, until a key or the pad is used
	back  backdropCache // the country bush, drawn once
	tree  treeCache     // the gum tree, drawn once a level
	ute   uteCache      // the ute, drawn once
	drive driving       // the ute's drive in or off, for its dust and smoke
	crew  crew          // the rangers' walk and the sheet's give
	birds flock         // the birds about the farm
	roos  mob           // the kangaroos behind the fence
	track reptiles      // the lace monitor and the carpet python on the track
	dizzy []ring        // the dazed effects showing last tick
	sound sounder       // picks the sounds

	demo *rand.Rand // the demo's own random numbers, nil in a real game
}

// New starts a game at a skill, on the level the -level flag asks for.
func New(ctx *scene.Context, skill config.Skill) scene.Scene {
	ctx.Session = session.New(skill, skill.Lives)
	p := newPlay(sceneryRNG(ctx), world.New(ctx.Session, ctx.StartLevel, ctx.RNG))
	ctx.Audio.PlayTune(LevelStartTune) // no music in play, so the title tune stops (DESIGN §3)
	return p
}

// NewDemo starts the demo behind the title (DESIGN §7): the rangers play
// by themselves at Normal from the first level, without sound, under the
// title tune. It has its own session and random numbers, from the
// scenery's, so the game's seed plays the same game after it.
func NewDemo(ctx *scene.Context) scene.Scene {
	scenery := sceneryRNG(ctx)
	p := newPlay(scenery, nil)
	p.demo = scene.NewRNG(scenery.Uint64())
	p.restart()
	return p
}

// sceneryRNG is ctx's random numbers for scenery, or with none, a test's
// own from seed 0.
func sceneryRNG(ctx *scene.Context) *rand.Rand {
	if ctx.Scenery == nil {
		return scene.NewSceneryRNG(0)
	}
	return ctx.Scenery
}

// newPlay is a play scene for w, its scenery placed by scenery.
func newPlay(scenery *rand.Rand, w *world.World) *Play {
	p := &Play{world: w, birds: newFlock(scenery), roos: newMob(scenery), track: newReptiles(scenery)}
	p.sound = newSounder(scenery)
	return p
}

// restart starts the demo's game again from the first level.
func (p *Play) restart() {
	skill := config.Skills[1] // Normal
	p.world = world.New(session.New(skill, skill.Lives), 1, p.demo)
	p.crew = crew{}
}

// Native implements scene.Native: the scene draws in either look.
func (*Play) Native() {}

// World is the game's state, for tests.
func (p *Play) World() *world.World { return p.world }

// Update implements scene.Scene.
func (p *Play) Update(ctx *scene.Context) error {
	in := ctx.Input
	switch {
	case p.demo != nil: // the title takes the keys
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
	steer := p.world.SafePlay
	if p.demo == nil {
		steer = func() world.Input { return p.input(ctx) }
	}
	events := p.world.Update(steer())
	for _, e := range events {
		switch {
		case e.Kind != world.GameOver:
		case p.demo != nil:
			p.restart()
		default:
			ctx.Cut(newEnded(p))
		}
	}
	p.drive.step(p.world)
	p.crew.step(p.world, events)
	p.birds.step()
	p.roos.step()
	p.track.step()
	now := rings(p.world, int(ctx.Tick))
	p.birds.release(p.dizzy, now)
	if p.demo != nil {
		p.dizzy = now
		return nil
	}
	tune, fx := p.sound.step(p.world, events, &p.crew, p.drive, &p.birds, p.dizzy, now)
	p.dizzy = now
	if tune != nil {
		ctx.Audio.PlayTune(tune)
	}
	for _, e := range fx {
		ctx.Audio.PlayEffect(e)
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
	banner(ctx, dst, "GAME OVER")
}
