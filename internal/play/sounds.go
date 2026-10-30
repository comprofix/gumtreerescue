package play

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/audio"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// Gum Tree Rescue's sounds (DESIGN §3, D21, D29), all original. In Chip4
// a tune's part i plays on channel i: 0 lead, 1 harmony, 2 bass, 3 noise.
// The jingles in play use only the lead and harmony, so the ute's engine
// and tyres can have the other two. Effects keep to a channel each:
//
//   - 0, the cues the player needs: the boing, the extra-life fanfare, and
//     under them the slide whistle;
//   - 1, chimes and voices: the rescue chime, the creak, the dazed sounds,
//     the reverse beeper, and lowest the bird calls;
//   - 2, low bodies: the clunks, bonk and thud, and the ute's engine;
//   - 3, noise: rattles, rustles, the gravel and lowest the footsteps.
//
// Priorities keep the scenery's sounds from cutting a game's.
var (
	// TitleTune is a bush band two-step in G, for the title and menus:
	// boom-chick bass, a lagerphone shuffle and a twangy lead.
	TitleTune = &audio.Tune{Name: "two-step title", Loop: true, Parts: twoStep(
		audio.Notes(
			"D5:12 E5:12 G5:24 G5:12 A5:12 B5:24",
			"C6:24 B5:12 A5:12 G5:12 E5:12 G5:24",
			"A5:2 B5:22 A5:12 G5:12 E5:12 D5:12 E5:24",
			"D5:12 F#5:12 A5:24 F#5:12 E5:12 D5:24",
			"G5:12 G5:12 B5:12 D6:12 D6:24 B5:12 G5:12",
			"C6:12 E6:12 D6:12 C6:12 B5:12 A5:12 G5:24",
			"A5:12 B5:12 A5:12 F#5:12 E5:12 F#5:12 A5:24",
			"F#5:2 G5:34 -:12 D5:12 E5:12 F#5:24",
		),
		chordG, chordC, chordG, chordD, chordG, chordC, chordD, chordG,
	)}

	// LevelStartTune is a two-step pickup as a level begins.
	LevelStartTune = &audio.Tune{Name: "level start", Parts: []audio.Part{
		{Voice: twang, Notes: audio.Notes("D5:6 E5:6 G5:12 B5:12 D6:12 G6:24 -:24")},
		{Voice: chick, Notes: audio.Notes("G4:6 A4:6 B4:12 D5:12 G5:12 B5:24 -:24")},
	}}

	// LevelEndTune is the ute's two-tone horn as it drives off, then a
	// short jingle.
	LevelEndTune = &audio.Tune{Name: "level end", Parts: []audio.Part{
		{Voice: horn, Notes: audio.Notes("A4:10 -:4 A4:22 -:12", "G5:8 A5:8 B5:8 D6:8 B5:8 D6:8 G6:32")},
		{Voice: horn, Notes: audio.Notes("C#5:10 -:4 C#5:22 -:12", "D5:8 F#5:8 G5:8 B5:8 G5:8 B5:8 D6:32")},
	}}

	// GameOverTune is a short bush ballad in 3/4, falling to rest.
	GameOverTune = &audio.Tune{Name: "game over", Parts: ballad(
		audio.Notes(
			"B5:72 A5:18 G5:18",
			"G5:36 E5:72",
			"C5:36 E5:36 G5:36",
			"F#5:54 E5:18 D5:36",
			"G5:108",
		),
		chordG, chordEm, chordC, chordD, chordG,
	)}

	// ScoresTune is the ballad for the high scores, looping.
	ScoresTune = &audio.Tune{Name: "ballad", Loop: true, Parts: ballad(
		audio.Notes(
			"D5:36 G5:54 A5:18",
			"G5:36 E5:72",
			"D5:36 B4:36 D5:36",
			"F#5:54 E5:18 D5:36",
			"D5:36 G5:54 B5:18",
			"C6:36 B5:36 A5:36",
			"G5:36 F#5:36 A5:36",
			"G5:90 -:18",
		),
		chordG, chordC, chordG, chordD, chordG, chordC, chordD, chordG,
	)}
)

// Effects.
var (
	// Boings is the bounce off the sheet, zone 1 to 5: sharp highest,
	// straight up lowest, so a slip is heard before it is seen (DESIGN §8).
	Boings = [5]*audio.Effect{
		boing("boing sharp left", boingSharp),
		boing("boing safe left", boingSafe),
		boing("boing straight up", boingCentre),
		boing("boing safe right", boingSafe),
		boing("boing sharp right", boingSharp),
	}
	// Fanfare is an extra life.
	Fanfare = effect("extra life", 0, 6, twang, audio.Notes("G5:5 B5:5 D6:5 G6:5 D6:5 G6:20"))

	// Chime is a koala rescued.
	Chime = effect("rescue chime", 1, 4, bell, audio.Notes("E6:4 G6:4 C7:14"))
	// Creak is the branch as a koala wobbles on it.
	Creak = effect("branch creak", 1, 3, creakV, creak())
	// Chuckle is a daze starting: a little kookaburra laugh (DESIGN D24).
	Chuckle = effect("dazed chuckle", 1, 2, chirp, audio.Notes("A6:2 -:1 C7:2 -:1 A6:2 -:1 C7:3 -:2 E7:2 -:1 C7:4"))
	// Tweet is the mini kookaburras circling a dazed koala.
	Tweet = effect("dazed tweet", 1, 2, chirp, audio.Notes("E7:2 G7:2 -:4 E7:2 G7:3"))
	// Beep is the ute's reverse beeper as it backs in.
	Beep = effect("reverse beeper", 1, 2, beeper, audio.Notes("B5:14"))

	// Clunk and Rattle are a koala landing in the tray.
	Clunk  = effect("tray clunk", 2, 4, clunkV, []audio.Note{tone(196, 2), tone(147, 6)})
	Rattle = effect("tray rattle", 3, 4, hit, []audio.Note{tone(2500, 4)})
	// Bonk and Mesh are a koala hitting the headboard's mesh.
	Bonk = effect("headboard bonk", 2, 4, clunkV, []audio.Note{tone(660, 2), tone(440, 3), tone(330, 8)})
	Mesh = effect("mesh rattle", 3, 4, hit, noise(9000, 5000, 9000, 5000, 9000, 5000, 8000, 4000))
	// Thud and Thump are a koala landing on the ground.
	Thud  = effect("ground thud", 2, 4, thudV, []audio.Note{tone(110, 3), tone(70, 5), tone(55, 6)})
	Thump = effect("ground thump", 3, 4, hit, []audio.Note{tone(600, 6)})

	// Rustle is a koala grabbing the tree, and Slip one slipping from it.
	Rustle = effect("leaf rustle", 3, 3, leaves, noise(9000, 13000, 7000, 11000, 8000, 12000, 6000, 10000))
	Slip   = effect("bark scrabble", 3, 3, leaves, noise(5000, 9000, 4000, 8000, 3500, 7000, 3000, 6000, 2500, 5000, 2000))
	// Steps are a ranger's feet on the dirt, left and right.
	Steps = [2]*audio.Effect{
		effect("footstep left", 3, 0, step, []audio.Note{tone(1400, 3)}),
		effect("footstep right", 3, 0, step, []audio.Note{tone(1900, 3)}),
	}
	// Gravel is the ute's tyres on the track.
	Gravel = effect("tyres on gravel", 3, 1, step, []audio.Note{tone(5000, 1), tone(3000, 2)})

	// Calls are the birds' calls, by art.Species (DESIGN §3).
	Calls = [art.NumSpecies]*audio.Effect{
		art.Kooka:    effect("kookaburra laugh", 1, 0, laugh, kookaLaugh()),
		art.Cockatoo: effect("cockatoo screech", 1, 0, screechV, screech()),
		art.Magpie:   effect("magpie warble", 1, 0, flute, audio.Notes("G5:4 D6:3 B5:5 -:2 E6:3 C6:2 G6:6 -:3 D6:3 F6:2 E6:5 -:2 A5:3 B5:3 G5:8")),
	}
)

// Sounds are the game's sounds, for the sound test: a fall's whistle and
// the ute's engine are made to measure, so one of each stands in.
var Sounds = audio.Bank{
	Tunes: []*audio.Tune{TitleTune, LevelStartTune, LevelEndTune, GameOverTune, ScoresTune},
	Effects: append(Boings[:], whistle(60), Fanfare, Chime, Creak, Chuckle, Tweet, Beep,
		Clunk, Rattle, Bonk, Mesh, Thud, Thump, Rustle, Slip, Steps[0], Steps[1], Gravel,
		engine(engineLevels/2), Calls[art.Kooka], Calls[art.Cockatoo], Calls[art.Magpie]),
}

// Voices.
var (
	twang    = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(8), Sustain: 0.45, Release: audio.Ticks(3), Volume: 0.8}
	chick    = audio.Voice{Wave: audio.Pulse12, Decay: audio.Ticks(4), Sustain: 0.2, Release: audio.Ticks(2), Volume: 0.45}
	boom     = audio.Voice{Wave: audio.Triangle, Decay: audio.Ticks(10), Sustain: 0.6, Release: audio.Ticks(3)}
	brush    = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(3), Release: audio.Ticks(1), Volume: 0.35}
	croon    = audio.Voice{Wave: audio.Pulse25, Attack: audio.Ticks(2), Decay: audio.Ticks(20), Sustain: 0.6, Release: audio.Ticks(8), Volume: 0.7}
	pah      = audio.Voice{Wave: audio.Pulse12, Decay: audio.Ticks(6), Sustain: 0.3, Release: audio.Ticks(3), Volume: 0.35}
	horn     = audio.Voice{Wave: audio.Pulse50, Decay: audio.Ticks(10), Sustain: 0.7, Release: audio.Ticks(2), Volume: 0.6}
	boingV   = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(10), Sustain: 0.3, Release: audio.Ticks(3), Volume: 0.9}
	whistleV = audio.Voice{Wave: audio.Triangle, Attack: audio.Ticks(1), Release: audio.Ticks(2), Volume: 0.55}
	bell     = audio.Voice{Wave: audio.Pulse50, Decay: audio.Ticks(12), Sustain: 0.2, Release: audio.Ticks(6), Volume: 0.7}
	creakV   = audio.Voice{Wave: audio.Pulse12, Release: audio.Ticks(2), Volume: 0.4}
	chirp    = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(2), Sustain: 0.5, Release: audio.Ticks(1), Volume: 0.4}
	beeper   = audio.Voice{Wave: audio.Pulse50, Release: audio.Ticks(1), Volume: 0.4}
	clunkV   = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(5), Release: audio.Ticks(2), Volume: 0.9}
	thudV    = audio.Voice{Wave: audio.Triangle, Decay: audio.Ticks(10), Release: audio.Ticks(3)}
	hit      = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(4), Release: audio.Ticks(1), Volume: 0.6}
	leaves   = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(2), Sustain: 0.3, Release: audio.Ticks(1), Volume: 0.4}
	step     = audio.Voice{Wave: audio.Noise, Decay: audio.Ticks(2), Release: audio.Ticks(1), Volume: 0.2}
	engineV  = audio.Voice{Wave: audio.Pulse12, Release: audio.Ticks(3), Volume: 0.45}
	laugh    = audio.Voice{Wave: audio.Pulse25, Decay: audio.Ticks(3), Sustain: 0.4, Release: audio.Ticks(1), Volume: 0.4}
	screechV = audio.Voice{Wave: audio.Pulse12, Attack: audio.Ticks(1), Release: audio.Ticks(2), Volume: 0.35}
	flute    = audio.Voice{Wave: audio.Triangle, Attack: audio.Ticks(1), Decay: audio.Ticks(4), Sustain: 0.7, Release: audio.Ticks(2), Volume: 0.5}
)

// How the sounds are pitched, in Hz, and how often the scene plays them,
// in ticks.
const (
	boingSharp  = 523.25 // the boings' lowest notes, by zone
	boingSafe   = 392.0
	boingCentre = 293.66

	whistleHigh = 1568.0 // the slide whistle falls from here
	whistleLow  = 523.25 // to here as the koala reaches the sheet
	whistleStep = 2      // in steps this long

	engineIdle   = 50.0 // the engine's pitch standing still,
	engineRise   = 12.0 // and how much it rises with each field pixel a tick
	engineLevels = 16   // the engine's pitches, from idle up
	engineFull   = 8.0  // the ute's speed, a tick, at the top one
	engineEvery  = 4    // the engine's note is played again this often
	revPeak      = 3.0  // the speed the engine revs to as the ute pulls away
	gravelMost   = 8    // the tyres crunch at most this often,
	gravelLeast  = 2    // at least this often at full speed
	beepEvery    = 30   // the reverse beeper
	beepStill    = 0.15 // and the speed below which the ute is parked

	footEvery  = math.Pi // a footfall each time a walk goes this far, in radians
	footStride = 0.5     // and its stride is at least this

	tweetFirst = 24 // the first tweet after a daze's chuckle,
	tweetEvery = 40 // then this often while a daze shows

	callMin   = 5 * config.TPS  // the shortest time between bird calls
	callMax   = 14 * config.TPS // and the longest
	callRetry = config.TPS      // a call held back tries again after this
	birdHush  = 20              // no call starts this close to a koala landing
)

// chord is a bar's chord for the backing: the bass's root and fifth, and
// the two notes of its stabs.
type chord struct{ root, fifth, low, high string }

var (
	chordG  = chord{"G2", "D2", "B4", "D5"}
	chordC  = chord{"C3", "G2", "C5", "E5"}
	chordD  = chord{"D3", "A2", "A4", "D5"}
	chordEm = chord{"E2", "B2", "G4", "B4"}
)

// twoStep backs a lead of 4/4 bars with boom-chick: the bass on 1 and 3,
// stabs on 2 and 4, and a shuffle on the noise like a lagerphone.
func twoStep(lead []audio.Note, bars ...chord) []audio.Part {
	var bass, stabs []audio.Note
	for _, c := range bars {
		bass = append(bass, audio.Notes(fmt.Sprintf("%s:16 -:32 %s:16 -:32", c.root, c.fifth))...)
		stabs = append(stabs, audio.Notes(fmt.Sprintf("-:24 %s:12 -:36 %s:12 -:12", c.low, c.high))...)
	}
	const (
		jingle = 15000
		snare  = 5000
	)
	var shuffle []audio.Note
	for range bars {
		for _, hz := range []float64{jingle, jingle, snare, jingle, jingle, jingle, snare, jingle} {
			shuffle = append(shuffle, tone(hz, 12))
		}
	}
	return []audio.Part{{Voice: twang, Notes: lead}, {Voice: chick, Notes: stabs}, {Voice: boom, Notes: bass}, {Voice: brush, Notes: shuffle}}
}

// ballad backs a lead of 3/4 bars with oom-pah-pah: the bass on 1, the
// chord on 2 and 3.
func ballad(lead []audio.Note, bars ...chord) []audio.Part {
	var bass, pahs []audio.Note
	for _, c := range bars {
		bass = append(bass, audio.Notes(fmt.Sprintf("%s:30 -:78", c.root))...)
		pahs = append(pahs, audio.Notes(fmt.Sprintf("-:36 %s:24 -:12 %s:24 -:12", c.low, c.high))...)
	}
	return []audio.Part{{Voice: croon, Notes: lead}, {Voice: pah, Notes: pahs}, {Voice: boom, Notes: bass}}
}

func effect(name string, ch, prio int, v audio.Voice, notes []audio.Note) *audio.Effect {
	return &audio.Effect{Name: name, Channel: ch, Priority: prio, Part: audio.Part{Voice: v, Notes: notes}}
}

// tone is a note at hz held for ticks.
func tone(hz float64, ticks int) audio.Note { return audio.Note{Hz: hz, Dur: audio.Ticks(ticks)} }

// noise is noise changing at each rate in turn, two ticks each.
func noise(rates ...float64) []audio.Note {
	var out []audio.Note
	for _, r := range rates {
		out = append(out, tone(r, 2))
	}
	return out
}

// boing is a quick rising run up an octave from base, with a wobble at
// the top.
func boing(name string, base float64) *audio.Effect {
	return effect(name, 0, 5, boingV, []audio.Note{
		tone(base, 1), tone(base*1.25, 1), tone(base*1.5, 1), tone(base*2, 4), tone(base*1.89, 3), tone(base*2, 6),
	})
}

// whistle is a slide whistle falling for ticks, the length of a koala's
// fall to the sheet.
func whistle(ticks int) *audio.Effect {
	var notes []audio.Note
	for t := 0; t < ticks; t += whistleStep {
		f := float64(t) / float64(max(ticks-1, 1))
		notes = append(notes, tone(whistleHigh*math.Pow(whistleLow/whistleHigh, f), min(whistleStep, ticks-t)))
	}
	return effect("fall whistle", 0, 1, whistleV, notes)
}

// engine is one of the ute's engine notes, level 0 idling up to
// engineLevels-1 flat out: a low buzz with a putter.
func engine(level int) *audio.Effect {
	hz := engineIdle + engineRise*engineFull*float64(level)/(engineLevels-1)
	return effect("ute engine", 2, 1, engineV, []audio.Note{tone(hz, 2), tone(hz*0.92, 2)})
}

// creak is a branch bending: a rough buzz rising and falling.
func creak() []audio.Note {
	var out []audio.Note
	for i := range 24 {
		hz := 150 + 70*math.Sin(math.Pi*float64(i)/24)
		if i%2 == 1 {
			hz *= 1.06
		}
		out = append(out, tone(hz, 1))
	}
	return out
}

// kookaLaugh is a kookaburra's laugh: low hoots building to a rising,
// rattling cackle, then falling away.
func kookaLaugh() []audio.Note {
	return audio.Notes(
		"C5:3 -:3 C5:3 -:3 D5:3 -:3",
		"G5:2 E5:2 G5:2 E5:2 A5:2 F5:2 A5:2 F5:2",
		"C6:2 A5:2 C6:2 A5:2 D6:2 B5:2 D6:2 B5:2",
		"C6:3 -:1 C6:3 -:1 A5:3 -:2 G5:3 -:2 E5:4 -:4 D5:6",
	)
}

// screech is a cockatoo's two harsh screeches: a rough trill falling a
// little each time.
func screech() []audio.Note {
	var out []audio.Note
	for range 2 {
		for i := range 24 {
			hz := 1500 - 350*float64(i)/24
			if i%2 == 1 {
				hz *= 0.88
			}
			out = append(out, tone(hz, 1))
		}
		out = append(out, audio.Rest(8))
	}
	return out
}

// sounder picks the play scene's sounds each tick from the world's events
// and the scenery (DESIGN §3): it only listens, so play is the same with
// sound or without. Its bird calls take their timing from random numbers
// of its own, from the scenery's.
type sounder struct {
	rng      *rand.Rand
	whistles map[int]*audio.Effect // by length
	engines  [engineLevels]*audio.Effect
	foot     int // the footfall the heard ranger is on
	call     int // ticks to the next bird call
	tweet    int // ticks to the next dazed tweet
	fx       []*audio.Effect
}

func newSounder(scenery *rand.Rand) sounder {
	s := sounder{rng: rand.New(rand.NewPCG(scenery.Uint64(), scenery.Uint64())), whistles: map[int]*audio.Effect{}}
	for i := range s.engines {
		s.engines[i] = engine(i)
	}
	s.call = s.between(callMin, callMax)
	return s
}

func (s *sounder) between(lo, hi int) int { return lo + s.rng.IntN(hi-lo+1) }

// whistle is the slide whistle for a fall of ticks, made once.
func (s *sounder) whistle(ticks int) *audio.Effect {
	e, ok := s.whistles[ticks]
	if !ok {
		e = whistle(ticks)
		s.whistles[ticks] = e
	}
	return e
}

// step picks this tick's sounds: a tune to start, if any, and the effects
// to play, in order. was and now are the dazed effects showing last tick
// and this one. The slice is reused by the next call.
func (s *sounder) step(w *world.World, events []world.Event, c *crew, d driving, f *flock, was, now []ring) (*audio.Tune, []*audio.Effect) {
	s.fx = s.fx[:0]
	var tune *audio.Tune
	for _, e := range events {
		switch e.Kind {
		case world.Wobble:
			s.fx = append(s.fx, Creak)
		case world.Dropped:
			if e.Ticks > 0 {
				s.fx = append(s.fx, s.whistle(e.Ticks))
			}
		case world.Bounced:
			s.fx = append(s.fx, Boings[e.Zone-1])
		case world.Rescued:
			s.fx = append(s.fx, Clunk, Rattle, Chime)
		case world.Bonked:
			s.fx = append(s.fx, Bonk, Mesh)
		case world.Grabbed:
			s.fx = append(s.fx, Rustle)
		case world.Slipped:
			s.fx = append(s.fx, Slip)
		case world.ExtraLife:
			s.fx = append(s.fx, Fanfare)
		case world.GameOver:
			tune = GameOverTune
		}
	}
	switch {
	case tune != nil, w.Over:
	case w.Starting == config.LevelTicks:
		tune = LevelStartTune
	case w.Ending == config.EndTicks: // the ute drives off
		tune = LevelEndTune
	}
	s.dazed(was, now)
	s.walk(w, c)
	s.drive(d)
	s.birds(w, f)
	return tune, s.fx
}

// dazed is a chuckle as a daze starts, with a thud if it is on the
// ground, then tweets while any daze shows.
func (s *sounder) dazed(was, now []ring) {
	var fresh, ground bool
	for _, r := range now {
		if !r.in(was) {
			fresh = true
			ground = ground || !r.tray
		}
	}
	switch {
	case fresh:
		if ground {
			s.fx = append(s.fx, Thud, Thump)
		}
		s.fx = append(s.fx, Chuckle)
		s.tweet = tweetFirst
	case len(now) > 0:
		if s.tweet--; s.tweet <= 0 {
			s.fx = append(s.fx, Tweet)
			s.tweet = tweetEvery
		}
	}
}

// walk is a footstep each time the heard ranger's foot comes down: the
// man's, as the two walk in step with the sheet, or the woman's as she
// walks off on her own with lost joeys.
func (s *sounder) walk(w *world.World, c *crew) {
	k := &c.man
	if _, away := rangerAway(w); away {
		k = &c.woman
	}
	foot := int(math.Floor(k.step / footEvery))
	if foot != s.foot && k.stride >= footStride {
		s.fx = append(s.fx, Steps[foot&1])
	}
	s.foot = foot
}

// drive is the ute's engine and tyres as it drives in or off, its reverse
// beeper as it backs in once the level's jingle is over, and its engine
// revving as it pulls away.
func (s *sounder) drive(d driving) {
	speed := math.Abs(d.speed())
	rev := d.off && d.t < pullAway
	if rev {
		speed = max(speed, revPeak*float64(d.t)/pullAway)
	}
	if speed <= dustStill && !rev {
		return
	}
	if d.t%engineEvery == 0 {
		level := int(math.Round(min(speed/engineFull, 1) * (engineLevels - 1)))
		s.fx = append(s.fx, s.engines[level])
	}
	if every := max(gravelLeast, min(gravelMost, int(gravelMost/speed))); d.speed() != 0 && d.t%every == 0 {
		s.fx = append(s.fx, Gravel)
	}
	if !d.off && d.t >= jingleTicks && speed > beepStill && d.t%beepEvery == 0 {
		s.fx = append(s.fx, Beep)
	}
}

// jingleTicks is how long the level start jingle lasts: the reverse
// beeper waits for it.
var jingleTicks = int(math.Round(audio.Length(LevelStartTune.Parts[0].Notes).Seconds() * config.TPS))

// birds is a call now and then, while a level is under way, from a bird
// perched about the farm, held back while a koala is about to land so the
// boing is heard clearly (DESIGN §3).
func (s *sounder) birds(w *world.World, f *flock) {
	if w.Starting > 0 || w.Ending > 0 || w.Over {
		return
	}
	if s.call--; s.call > 0 {
		return
	}
	s.call = callRetry
	for i := range w.Koalas {
		if _, ticks, ok := w.Landing(&w.Koalas[i]); ok && ticks <= birdHush {
			return
		}
	}
	var sitting []art.Species
	for _, b := range f.birds {
		if b.state == perched {
			sitting = append(sitting, b.kind)
		}
	}
	if len(sitting) == 0 {
		return
	}
	s.fx = append(s.fx, Calls[sitting[s.rng.IntN(len(sitting))]])
	s.call = s.between(callMin, callMax)
}
