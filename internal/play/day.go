package play

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/world"
)

// day is the time of day and the weather over the play (DESIGN D34): each
// level is a day, from sunrise, a step on with each rescue, to night with
// the last, under weather picked for the level. It is scenery: Play moves
// it in Update from the scenery's own random numbers, never the world's,
// so play is the same with it or without.
type day struct {
	level    int // the level whose day it is, 0 before the first
	from, to sky // the sky eased from, and to
	ease     int // ticks into easing from one to the other
	rng      *rand.Rand
}

// sky is a step of a day under a weather.
type sky struct {
	weather art.Weather
	step    int
}

// night reports whether s is the night.
func (s sky) night() bool { return s.step >= config.DaySteps }

func newDay(rng *rand.Rand) day { return day{rng: rng, ease: config.SkyEaseTicks} }

// step moves the day on for w: a new level starts a new day, at sunrise
// under new weather; a rescue moves it a step on, and the level's end
// brings night. Misses don't move it.
func (d *day) step(w *world.World) {
	if d.ease < config.SkyEaseTicks {
		d.ease++
	}
	weather := d.to.weather
	if w.Level != d.level {
		first := d.level == 0
		d.level, weather = w.Level, pickWeather(d.rng)
		if first {
			d.from, d.to, d.ease = sky{weather, 0}, sky{weather, 0}, config.SkyEaseTicks
			return
		}
	}
	at := sky{weather, min(w.Rescued, config.DaySteps)}
	if w.Ending > 0 {
		at.step = config.DaySteps
	}
	if at != d.to {
		// Eased again before it got there, it goes on from whichever
		// the sky looks more like.
		if 2*d.ease >= config.SkyEaseTicks {
			d.from = d.to
		}
		d.to, d.ease = at, 0
	}
}

// blend is how far the sky has eased from d.from to d.to, from 0 to 1,
// slow at either end.
func (d *day) blend() float64 {
	f := float64(d.ease) / config.SkyEaseTicks
	return f * f * (3 - 2*f)
}

// pickWeather is a level's weather, by config.WeatherOdds.
func pickWeather(rng *rand.Rand) art.Weather {
	total := 0
	for _, n := range config.WeatherOdds {
		total += n
	}
	r := rng.IntN(total)
	for i, n := range config.WeatherOdds {
		if r < n {
			return art.Weathers[i]
		}
		r -= n
	}
	return art.Clear
}

// skyCache keeps the two skies being eased between, each drawn once at the
// shown screen's scale, as they are slow to draw, and the night's moon and
// stars.
type skyCache struct {
	imgs  [2]*ebiten.Image
	keys  [2]skyKey
	night *ebiten.Image
	nkey  skyKey
}

// skyKey is what a cached sky was drawn for.
type skyKey struct {
	sky
	k    float64
	w, h int
}

// image is sky s drawn for p, from the cache if it is there, else drawn in
// the place of any but keep: in shapes, or in the pixel look in hard pixels
// banded to the palette (DESIGN D30).
func (c *skyCache) image(p pen, s, keep sky) *ebiten.Image {
	b := p.dst.Bounds()
	key := skyKey{s, p.k, b.Dx(), b.Dy()}
	for i, img := range c.imgs {
		if img != nil && c.keys[i] == key {
			return img
		}
	}
	i := 0
	if c.imgs[0] != nil && c.keys[0].sky == keep && c.keys[0].k == p.k {
		i = 1
	}
	c.imgs[i] = blank(c.imgs[i], b.Dx(), b.Dy())
	c.keys[i] = key
	fw, _ := config.Field()
	ap := p.art()
	ap.Dst = c.imgs[i]
	art.Sky(ap, float64(fw), s.weather, s.step, config.DaySteps)
	if ap.Hard {
		art.Banded(c.imgs[i])
	}
	return c.imgs[i]
}

// lights is the night's moon and stars under weather wx drawn for p, from
// the cache if it is there.
func (c *skyCache) lights(p pen, wx art.Weather) *ebiten.Image {
	b := p.dst.Bounds()
	key := skyKey{sky{wx, config.DaySteps}, p.k, b.Dx(), b.Dy()}
	if c.night != nil && c.nkey == key {
		return c.night
	}
	c.night, c.nkey = blank(c.night, b.Dx(), b.Dy()), key
	fw, _ := config.Field()
	ap := p.art()
	ap.Dst = c.night
	art.NightLights(ap, float64(fw), wx, 1)
	if ap.Hard {
		art.Banded(c.night)
	}
	return c.night
}

// blank is img cleared, or a new image if it is not w by h.
func blank(img *ebiten.Image, w, h int) *ebiten.Image {
	if img == nil || img.Bounds().Dx() != w || img.Bounds().Dy() != h {
		return ebiten.NewImage(w, h)
	}
	img.Clear()
	return img
}

// drawSky lays the sky behind the backdrop, eased from one to the next.
func (p *Play) drawSky(pn pen) {
	from, to, f := p.day.from, p.day.to, p.day.blend()
	if f < 1 {
		pn.dst.DrawImage(p.sky.image(pn, from, to), nil)
	}
	if f > 0 {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(float32(f))
		pn.dst.DrawImage(p.sky.image(pn, to, from), op)
	}
}

// drawLight lays the day's light over the whole field, under the score
// line (DESIGN D34): each sky's tints as far as it is eased, then at night
// the moon and stars over them.
func (p *Play) drawLight(pn pen) {
	from, to, f := p.day.from, p.day.to, p.day.blend()
	fw, fh := config.Field()
	w, h := float64(fw), float64(fh)
	ap := pn.art()
	dark, wx := 0.0, to.weather
	for _, s := range []struct {
		sky
		a float64
	}{{from, 1 - f}, {to, f}} {
		if s.a <= 0 {
			continue
		}
		for _, t := range art.Light(s.weather, s.step, config.DaySteps) {
			ap.Faded(t.A*s.a).Poly(t.C, 0, 0, w, 0, w, h, 0, h)
		}
		if s.night() {
			dark, wx = dark+s.a, s.weather
		}
	}
	if dark > 0 {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(float32(dark))
		pn.dst.DrawImage(p.sky.lights(pn, wx), op)
	}
}
