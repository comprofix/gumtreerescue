package play

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/comprofix/gumtreerescue/internal/audio"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/input"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/scene/scenetest"
	"github.com/comprofix/gumtreerescue/internal/session"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// testWorld is a game on level 1 at Normal, still at its intro card.
func testWorld() *world.World {
	normal := config.Skills[1]
	return world.New(session.New(normal, normal.Lives), 1, rand.New(rand.NewPCG(1, 2)))
}

func ticksOf(notes []audio.Note) int {
	return int(math.Round(audio.Length(notes).Seconds() * config.TPS))
}

func TestSounds(t *testing.T) {
	if err := audio.Join(scene.Sounds, Sounds).Check(); err != nil {
		t.Fatal(err)
	}
}

// TestBoingPitch checks the boing is pitched by zone: sharp highest,
// straight up lowest, the same either way (DESIGN §3, §8).
func TestBoingPitch(t *testing.T) {
	hz := func(z int) float64 { return Boings[z-1].Part.Notes[0].Hz }
	if !(hz(1) > hz(2) && hz(2) > hz(3)) || hz(1) != hz(5) || hz(2) != hz(4) {
		t.Errorf("boing pitches by zone: %v %v %v %v %v", hz(1), hz(2), hz(3), hz(4), hz(5))
	}
}

// TestWhistle checks the slide whistle lasts exactly the fall and only
// ever falls in pitch.
func TestWhistle(t *testing.T) {
	for _, n := range []int{1, 2, 37, 60, 95} {
		notes := whistle(n).Part.Notes
		if got := ticksOf(notes); got != n {
			t.Errorf("whistle for %d ticks lasts %d", n, got)
		}
		for i := 1; i < len(notes); i++ {
			if notes[i].Hz > notes[i-1].Hz {
				t.Errorf("whistle for %d ticks rises at note %d", n, i)
			}
		}
	}
}

// TestEventSounds checks each event the world reports has its sound.
func TestEventSounds(t *testing.T) {
	w := testWorld()
	cases := []struct {
		e    world.Event
		want []*audio.Effect
	}{
		{world.Event{Kind: world.Wobble}, []*audio.Effect{Creak}},
		{world.Event{Kind: world.Rescued}, []*audio.Effect{Clunk, Rattle, Chime}},
		{world.Event{Kind: world.Bonked}, []*audio.Effect{Bonk, Mesh}},
		{world.Event{Kind: world.Grabbed}, []*audio.Effect{Rustle}},
		{world.Event{Kind: world.Slipped}, []*audio.Effect{Slip}},
		{world.Event{Kind: world.ExtraLife}, []*audio.Effect{Fanfare}},
	}
	for z := 1; z <= 5; z++ {
		cases = append(cases, struct {
			e    world.Event
			want []*audio.Effect
		}{world.Event{Kind: world.Bounced, Zone: z}, []*audio.Effect{Boings[z-1]}})
	}
	for _, c := range cases {
		s := newSounder(scene.NewSceneryRNG(1))
		_, fx := s.step(w, []world.Event{c.e}, &crew{}, driving{t: 1000}, &flock{}, nil, nil)
		if !slices.Equal(fx, c.want) {
			t.Errorf("event %v: sounds %v, want %v", c.e, names(fx), names(c.want))
		}
	}
	s := newSounder(scene.NewSceneryRNG(1))
	_, fx := s.step(w, []world.Event{{Kind: world.Dropped, Ticks: 44}}, &crew{}, driving{t: 1000}, &flock{}, nil, nil)
	if len(fx) != 1 || fx[0].Name != "fall whistle" || ticksOf(fx[0].Part.Notes) != 44 {
		t.Errorf("drop: sounds %v, want a 44-tick whistle", names(fx))
	}
	if tune, _ := s.step(w, []world.Event{{Kind: world.GameOver}}, &crew{}, driving{t: 1000}, &flock{}, nil, nil); tune != GameOverTune {
		t.Errorf("game over: tune %v", tune)
	}
}

func names(fx []*audio.Effect) []string {
	var out []string
	for _, e := range fx {
		out = append(out, e.Name)
	}
	return out
}

// TestDazedSounds checks a daze starts with a chuckle, a thud too on the
// ground, then tweets while it shows.
func TestDazedSounds(t *testing.T) {
	w := testWorld()
	step := func(s *sounder, was, now []ring) []*audio.Effect {
		_, fx := s.step(w, nil, &crew{}, driving{t: 1000}, &flock{}, was, now)
		return fx
	}
	s := newSounder(scene.NewSceneryRNG(1))
	ground := []ring{{id: 3}}
	if fx := step(&s, nil, ground); !slices.Equal(fx, []*audio.Effect{Thud, Thump, Chuckle}) {
		t.Errorf("on the ground: %v", names(fx))
	}
	var tweets []int
	for i := 1; i <= tweetFirst+2*tweetEvery; i++ {
		if fx := step(&s, ground, ground); slices.Contains(fx, Tweet) {
			tweets = append(tweets, i)
		}
	}
	if want := []int{tweetFirst, tweetFirst + tweetEvery, tweetFirst + 2*tweetEvery}; !slices.Equal(tweets, want) {
		t.Errorf("tweets on ticks %v, want %v", tweets, want)
	}
	tray := []ring{{id: 4, tray: true}}
	if fx := step(&s, ground, append(ground, tray...)); !slices.Equal(fx, []*audio.Effect{Chuckle}) {
		t.Errorf("in the tray: %v", names(fx))
	}
	if fx := step(&s, ground, nil); len(fx) != 0 {
		t.Errorf("daze over: %v", names(fx))
	}
}

// TestUteSounds checks the ute's engine and tyres as it drives in and off,
// its reverse beeper only as it backs in, after the jingle, and silence
// once it is parked or gone.
func TestUteSounds(t *testing.T) {
	w := testWorld() // at its intro card: no bird calls
	count := func(off bool, ticks int) (engine, gravel, beep int) {
		s := newSounder(scene.NewSceneryRNG(1))
		for i := range ticks {
			d := driving{t: i, off: off}
			_, fx := s.step(w, nil, &crew{}, d, &flock{}, nil, nil)
			for _, e := range fx {
				switch e {
				case Beep:
					beep++
					if off || i < jingleTicks || math.Abs(d.speed()) <= beepStill {
						t.Errorf("beep on tick %d off %v at speed %.2f", i, off, d.speed())
					}
				case Gravel:
					gravel++
				default:
					if e.Name == "ute engine" {
						engine++
					}
				}
			}
		}
		return
	}
	if e, g, b := count(false, config.LevelTicks); e == 0 || g == 0 || b == 0 {
		t.Errorf("backing in: engine %d, gravel %d, beeps %d", e, g, b)
	}
	if e, g, b := count(true, config.EndTicks/2); e == 0 || g == 0 || b != 0 {
		t.Errorf("driving off: engine %d, gravel %d, beeps %d", e, g, b)
	}
	s := newSounder(scene.NewSceneryRNG(1))
	for _, d := range []driving{{t: config.LevelTicks + 10}, {t: config.EndTicks, off: true}} {
		if _, fx := s.step(w, nil, &crew{}, d, &flock{}, nil, nil); len(fx) != 0 {
			t.Errorf("parked or gone %+v: %v", d, names(fx))
		}
	}
}

// TestFootsteps checks a footstep for each footfall of the rangers' walk,
// and none standing still.
func TestFootsteps(t *testing.T) {
	w := testWorld()
	var c crew
	s := newSounder(scene.NewSceneryRNG(1))
	steps := 0
	tick := func() {
		c.step(w, nil)
		_, fx := s.step(w, nil, &c, driving{t: 1000}, &flock{}, nil, nil)
		for _, e := range fx {
			if e == Steps[0] || e == Steps[1] {
				steps++
			}
		}
	}
	for range 60 {
		tick()
	}
	if steps != 0 {
		t.Fatalf("%d footsteps standing still", steps)
	}
	start := c.man.step
	for range 120 {
		w.Sheet += w.RangerSpeed()
		tick()
	}
	falls := int(math.Abs(c.man.step-start) / footEvery)
	if steps < falls-2 || steps > falls+1 {
		t.Errorf("%d footsteps for %d footfalls", steps, falls)
	}
}

// TestBirdCalls checks the birds call now and then while a level is under
// way, only birds sitting on a perch, and never just as a koala lands.
func TestBirdCalls(t *testing.T) {
	w := testWorld()
	w.Starting = 0
	f := newFlock(scene.NewSceneryRNG(2))
	s := newSounder(scene.NewSceneryRNG(1))
	const ticks = 5 * 60 * config.TPS
	calls := 0
	for range ticks {
		s.birds(w, &f)
		for _, e := range s.fx {
			if !slices.Contains(Calls[:], e) {
				t.Fatalf("not a bird call: %s", e.Name)
			}
			calls++
		}
		s.fx = s.fx[:0]
	}
	if calls < ticks/callMax || calls > ticks/callMin+1 {
		t.Errorf("%d calls in %d ticks", calls, ticks)
	}

	away := f
	away.birds = slices.Clone(f.birds)
	for i := range away.birds {
		away.birds[i].state = flying
	}
	s.call = 1
	if s.birds(w, &away); len(s.fx) != 0 {
		t.Errorf("a call with no bird perched: %v", names(s.fx))
	}

	w = testWorld()
	for range 10 * config.TPS {
		w.Update(world.Input{})
		landing := false
		for i := range w.Koalas {
			if _, n, ok := w.Landing(&w.Koalas[i]); ok && n <= birdHush {
				landing = true
			}
		}
		if landing {
			s.fx, s.call = s.fx[:0], 1
			if s.birds(w, &f); len(s.fx) != 0 {
				t.Errorf("a call just as a koala lands: %v", names(s.fx))
			}
			return
		}
	}
	t.Fatal("no koala came down to land")
}

// TestPlayWithSound plays a game with a synth: the level start jingle
// replaces the title tune, and the game plays the same as without sound.
func TestPlayWithSound(t *testing.T) {
	run := func(synth *audio.Synth) string {
		h := scenetest.New(t, 42)
		h.Ctx.Audio = synth
		h.Ctx.NewPlay, h.Ctx.NativePlay = New, true
		synth.PlayTune(TitleTune)
		p := New(h.Ctx, config.Skills[1]).(*Play)
		if synth != nil && synth.Tune() != LevelStartTune {
			t.Errorf("tune %v as play starts", synth.Tune())
		}
		h.Start(p)
		h.Ticks(scene.FadeTicks)
		for range 2 * config.KoalaLevel * config.LevelAt(1).GapTicks {
			h.Run(input.Step{Keys: botKey(p.World()), Ticks: 1})
		}
		return summary(p)
	}
	if with, without := run(audio.NewSynth(config.Chip4)), run(nil); with != without {
		t.Errorf("the game differs with sound:\n%s\nwithout:\n%s", with, without)
	}
}
