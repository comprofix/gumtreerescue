package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/scene"
	"github.com/comprofix/gumtreerescue/internal/session"
	"github.com/comprofix/gumtreerescue/internal/world"
)

func newWorld() *world.World {
	skill := config.Skills[1]
	return world.New(session.New(skill, skill.Lives), 1, scene.NewRNG(1))
}

// settle steps d until its sky has eased all the way.
func settle(d *day, w *world.World) {
	for range config.SkyEaseTicks + 1 {
		d.step(w)
	}
}

// A level's day starts at sunrise, each rescue moves it a step on, misses
// don't, and the second last rescue brings night, lasting through the
// level's end (DESIGN D34).
func TestDayFollowsRescues(t *testing.T) {
	w := newWorld()
	d := newDay(scene.NewSceneryRNG(3))
	d.step(w)
	if d.to.step != 0 || d.blend() != 1 {
		t.Fatalf("first level starts at step %d blend %v, want sunrise, settled", d.to.step, d.blend())
	}
	w.Rescued = 3
	settle(&d, w)
	if d.to.step != 3 || d.from.step != 0 || d.blend() != 1 {
		t.Errorf("after 3 rescues: from %d to %d blend %v, want to 3, settled", d.from.step, d.to.step, d.blend())
	}
	w.Misses = 2
	settle(&d, w)
	if d.to.step != 3 {
		t.Errorf("misses moved the day to step %d", d.to.step)
	}
	w.Rescued = config.KoalaLevel - 2
	settle(&d, w)
	if d.to.night() {
		t.Errorf("%d rescues: night already", w.Rescued)
	}
	w.Rescued = config.KoalaLevel - 1
	settle(&d, w)
	if !d.to.night() {
		t.Errorf("%d rescues: step %d, want night for the last koala", w.Rescued, d.to.step)
	}
	w.Rescued, w.Ending = config.KoalaLevel, config.EndTicks
	settle(&d, w)
	if !d.to.night() {
		t.Errorf("level done: step %d, want night", d.to.step)
	}
}

// The next level starts a new day at sunrise, its weather picked again.
func TestNewLevelSunrise(t *testing.T) {
	w := newWorld()
	d := newDay(scene.NewSceneryRNG(3))
	d.step(w)
	w.Rescued, w.Ending = config.KoalaLevel, config.EndTicks
	settle(&d, w)
	w.Level, w.Rescued, w.Ending = 2, 0, 0
	d.step(w)
	if d.level != 2 || d.to.step != 0 || !d.from.night() || d.blend() == 1 {
		t.Errorf("level 2: level %d from %+v to %+v blend %v, want easing from night to sunrise", d.level, d.from, d.to, d.blend())
	}
	settle(&d, w)
	if d.blend() != 1 {
		t.Errorf("not eased after %d ticks", config.SkyEaseTicks)
	}
}

// Every weather comes up, by its odds, from the scenery's numbers.
func TestPickWeather(t *testing.T) {
	rng := scene.NewSceneryRNG(9)
	seen := map[art.Weather]int{}
	for range 1000 {
		seen[pickWeather(rng)]++
	}
	for i, wx := range art.Weathers {
		if seen[wx] == 0 {
			t.Errorf("%v never picked", wx)
		}
		if config.WeatherOdds[i] > 0 && seen[wx] < 1000*config.WeatherOdds[i]/20/2 {
			t.Errorf("%v picked %d times in 1000, too few for odds %d", wx, seen[wx], config.WeatherOdds[i])
		}
	}
}
