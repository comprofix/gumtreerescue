package art

import "github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"

// The landing shadow: how wide and deep it is when the koala is down on
// it, in field pixels, its colour, and how solid it is drawn then.
const (
	shadowRX   = 11
	shadowRY   = 3.2
	shadowTone = resurrect64.Shadow
	shadowFill = 0.75
)

// LandingShadow is the shadow of a koala coming down on the spot at the
// pen's origin, near from 0, just bounced, to 1, landing. It grows and
// darkens from a small spot as the koala comes down, so it reads by size,
// not colour.
func LandingShadow(p Pen, near float64) {
	near = min(max(near, 0), 1)
	f := 0.15 + 0.85*near*near
	p.Faded(shadowFill*(0.3+0.7*near)).Ellipse(0, 0, shadowRX*f, shadowRY*f, 0, shadowTone)
}
