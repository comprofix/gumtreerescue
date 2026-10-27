package art

import "github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"

// The colours of a puff of red dust off the track, and of exhaust smoke.
const (
	dust      = resurrect64.Honey
	dustShade = resurrect64.Caramel
	smoke     = resurrect64.Mist
	smokeDark = resurrect64.Fog
)

// Dust draws a soft puff of red dust about (x, y), r field pixels across
// its middle, a solid (from 0 to 1) so the scene shows through.
func Dust(p Pen, x, y, r, a float64) {
	puff(p.Faded(a), x, y, r, dustShade, dust)
}

// Smoke draws a soft puff of exhaust smoke about (x, y), r across its
// middle, a solid; dark puffs are greyer.
func Smoke(p Pen, x, y, r, a float64, dark bool) {
	shade, c := smokeDark, smoke
	if dark {
		shade, c = resurrect64.Dusk, smokeDark
	}
	puff(p.Faded(a), x, y, r, shade, c)
}

// puff draws a soft, lumpy cloud: a faint shade below, then three
// overlapping rounds, thickest where they meet.
func puff(p Pen, x, y, r float64, shade, c int) {
	p.Faded(0.5).Ellipse(x, y+r*0.3, r*0.9, r*0.55, 0, shade)
	p.Circle(x-r*0.35, y, r*0.7, c)
	p.Circle(x+r*0.35, y+r*0.05, r*0.65, c)
	p.Circle(x, y-r*0.3, r*0.6, c)
}
