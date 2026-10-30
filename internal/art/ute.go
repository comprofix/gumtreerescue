package art

import (
	"math"

	"github.com/comprofix/gumtreerescue/internal/gfx"
	"github.com/comprofix/gumtreerescue/internal/gfx/palettes/resurrect64"
)

// The ute's colours (DESIGN §2, §3): white paint shaded pale grey, a grey
// tray and steel wheels, navy-black bar, snorkel and trim, dark tyres,
// blue glass, amber indicators and red tail lights, and the dark green
// emblem and lettering.
const (
	paint      = resurrect64.White
	paintShade = resurrect64.Mist
	tray       = resurrect64.Fog
	trayShade  = resurrect64.Dusk
	trayLight  = resurrect64.Mist
	trim       = resurrect64.Navy
	tyre       = resurrect64.Charcoal
	glass      = resurrect64.Blue
	glassLight = resurrect64.Sky
	emblem     = resurrect64.Forest
	lettering  = resurrect64.Pine
)

// UteWheel is how far below the tray's deck the ute's tyres meet the
// track, in field pixels.
const UteWheel = 36

// UteFront is how far past the headboard the front of the ute's bull bar
// is, in field pixels.
const UteFront = 79

// UteWheelR is the radius of the ute's tyres, in field pixels.
const UteWheelR = 12.0

// axles is where the ute's rear and front wheels' middles are across, for
// a tray running back behind the pen's origin to the headboard, head in
// front: the rear a third of the way along the tray, the front under the
// bonnet.
func axles(back, head float64) (rear, front float64) {
	return -back + (back+head)/3, head + 54
}

// UteTyres is where the ute's tyres meet the track, across, from its
// station, for a tray running back field pixels behind it to the
// headboard, head in front: where dust kicks up as it drives.
func UteTyres(back, head float64) (rear, front float64) { return axles(back, head) }

// UteExhaust is where the ute's exhaust pipe ends, across and down from
// its station on the tray's deck: under the tray, behind the rear wheel.
func UteExhaust(back, head float64) (x, y float64) {
	rear, _ := axles(back, head)
	return rear - UteWheelR - 4.6, 11
}

// Ute draws the ranger ute side on, facing right (DESIGN §2): everything
// but its wheels, which UteWheels draws as they roll, and its tray's near
// drop side, which TraySide draws over the koalas in it. The pen's origin
// is on the tray's deck at its station; the tray runs back field pixels
// behind it to the headboard, head in front.
func Ute(p Pen, back, head float64) {
	h := head
	rear, front := axles(back, head)
	// Its shadow on the track.
	p.Ellipse((rear+front)/2, UteWheel, (front-rear)/2+30, 2.6, 0, resurrect64.Rust)
	// The chassis, the rear mudguard and the toolbox under the tray.
	p.Poly(ink, -back+2, 5, h+60, 5, h+60, 11, -back+2, 11)
	p.Poly(trim, rear-15, 5, rear+15, 5, rear+17, 13, rear-17, 13)
	p.Poly(ink, h-30, 5, h-5, 5, h-5, 19, h-30, 19)
	p.Poly(trim, h-29, 5.8, h-5.8, 5.8, h-5.8, 18.2, h-29, 18.2)
	p.Stroke(resurrect64.Lilac, 0.8, h-29, 8, h-5.8, 8)
	p.Ellipse(h-17.5, 12, 1.4, 1.8, 0, resurrect64.Fog)
	// The exhaust pipe, out from under the tray behind the rear wheel.
	ex, ey := UteExhaust(back, head)
	p.Stroke(ink, 1.8+2*Line, rear-6, ey-1.6, ex+1, ey)
	p.Stroke(resurrect64.Dusk, 1.8, rear-6, ey-1.6, ex+1, ey)
	p.Ellipse(ex+0.6, ey, 0.7, 1.1, 0, ink)
	// The tow hitch and the blank number plate at the back.
	p.Stroke(ink, 2.6, -back-1, 14, -back+6, 14)
	p.Poly(ink, -back+1, 7.2, -back+9, 7.2, -back+9, 12.4, -back+1, 12.4)
	p.Poly(paint, -back+1.8, 8, -back+8.2, 8, -back+8.2, 11.6, -back+1.8, 11.6)
	// The tray's deck and far side, behind the koalas.
	p.Poly(ink, -back, -7, h-4, -7, h-4, 7, -back, 7)
	p.Poly(trayShade, -back+0.8, -6.2, h-4.8, -6.2, h-4.8, 6.2, -back+0.8, 6.2)
	headboard(p, h)
	cab(p, h, front)
}

// UteWheels draws the ute's wheels, with the pen as for Ute, turned by
// roll radians clockwise as the ute drives and spinning spin radians a
// tick, and the front fender flare over its wheel.
func UteWheels(p Pen, back, head, roll, spin float64) {
	rear, front := axles(back, head)
	blur := math.Abs(spin) > blurSpin
	wheel(p, rear, UteWheel-UteWheelR, roll, blur)
	wheel(p, front, UteWheel-UteWheelR, roll, blur)
	flare := []float64{front - 15.4, 20, front - 14.4, 12.8, front - 8.6, 8.4, front + 8.6, 8.4, front + 14.4, 12.8, front + 15.4, 20}
	p.Stroke(ink, 2.6+2*Line, flare...)
	p.Stroke(trim, 2.6, flare...)
}

// UteRoll is how far the ute's wheels have turned, in radians clockwise,
// when it has driven x field pixels to the right.
func UteRoll(x float64) float64 { return x / UteWheelR }

// TraySide draws the ute's tray's near drop side, in front of the koalas
// in it, with the pen as for Ute: grey panels with a ridge, latches and a
// red tail light.
func TraySide(p Pen, back, head float64) {
	x0, x1 := -back, head-4
	p.Poly(ink, x0-Line, -6-Line, x1+Line, -6-Line, x1+Line, 6+Line, x0-Line, 6+Line)
	p.Poly(tray, x0, -6, x1, -6, x1, 6, x0, 6)
	p.Poly(trayShade, x0, 2.6, x1, 2.6, x1, 6, x0, 6)
	p.Stroke(trayLight, 1, x0+0.6, -5.2, x1-0.6, -5.2)
	p.Stroke(trayShade, 0.6, x0+0.5, -1.2, x1-0.5, -1.2)
	// Latches, holding the side up, one at each third.
	for i := 1; i <= 3; i++ {
		x := x0 + (x1-x0)*float64(i)/4
		p.Poly(ink, x-1.6, -1.2, x+1.6, -1.2, x+1.6, 4.2, x-1.6, 4.2)
		p.Ellipse(x, 1.4, 0.6, 1.2, 0, resurrect64.Fog)
	}
	// The tail light on the back corner.
	p.Poly(ink, x0-1.6, -0.6, x0+1.4, -0.6, x0+1.4, 5.6, x0-1.6, 5.6)
	p.Poly(resurrect64.Red, x0-0.9, 0.2, x0+0.7, 0.2, x0+0.7, 4.8, x0-0.9, 4.8)
}

// blurSpin is how fast, in radians a tick, a wheel spins before its tread
// and holes blur into rings, as they would to the eye, rather than seem
// to stand still or turn back.
const blurSpin = 0.12

// wheel draws a chunky off-road tyre on a grey steel wheel about (x, y),
// its tread, holes and valve turned by roll radians clockwise, or with
// blur, spinning too fast to see but for the valve; the light stays where
// it is.
func wheel(p Pen, x, y, roll float64, blur bool) {
	const r = UteWheelR
	p.Circle(x, y, r+Line, ink)
	p.Circle(x, y, r, tyre)
	// Tread blocks round the tyre's edge, or a blur of them.
	if blur {
		p.Circle(x, y, r-0.4, resurrect64.Shadow)
		p.Circle(x, y, r-1.4, tyre)
	}
	for i := range 16 {
		if blur {
			break
		}
		a := float64(i)*math.Pi/8 + roll
		s, c := math.Sincos(a)
		p.Ellipse(x+c*(r-0.4), y+s*(r-0.4), 1.4, 0.9, a, ink)
	}
	p.Curve(resurrect64.Shadow, 1.2, x-r*0.75, y-r*0.35, x-r*0.45, y-r*0.72, x, y-r*0.82)
	// The steel wheel: rim, dish with holes round it, and the hub.
	p.Circle(x, y, 7+Line, ink)
	p.Circle(x, y, 7, tray)
	p.Circle(x+0.4, y+0.4, 5.6, trayShade)
	p.Circle(x, y, 5.2, tray)
	if blur {
		p.Circle(x, y, 4.4, trayShade)
		p.Circle(x, y, 2.8, tray)
	}
	for i := range 6 {
		if blur {
			break
		}
		s, c := math.Sincos(float64(i)*math.Pi/3 + roll)
		p.Circle(x+c*3.6, y+s*3.6, 0.8, trayShade)
	}
	p.Circle(x, y, 1.8, trayShade)
	p.Circle(x-0.4, y-0.4, 0.8, trayLight)
	// The valve on the rim, which shows the wheel turn at any speed.
	s, c := math.Sincos(roll - math.Pi/2)
	p.Circle(x+c*6.2, y+s*6.2, 0.9, trayLight)
}

// headboard draws the tray's headboard behind the cab, h across: a tube
// frame with a mesh rack.
func headboard(p Pen, h float64) {
	frame := []float64{h - 7, -6, h - 7, -27, h - 5.4, -30, h, -30}
	p.Stroke(ink, 2+2*Line, frame...)
	for i := 0; i < 6; i++ {
		y := -26 + float64(i)*3.4
		p.Stroke(resurrect64.Dusk, 0.5, h-6.5, y, h-1, y+1.4)
	}
	p.Stroke(resurrect64.Dusk, 0.5, h-4, -28, h-4, -6)
	p.Stroke(tray, 2, frame...)
	p.Stroke(trayLight, 0.7, h-7.4, -10, h-7.4, -26.6)
}

// cab draws the ute's white cab and bonnet from h, the headboard, to the
// bull bar, over the front wheel at front: the window and door with the
// emblem and PARKS AND WILDLIFE, the snorkel, aerial, mirror, lights and
// bar.
func cab(p Pen, h, front float64) {
	body := []float64{
		h, 20, h, -24, h + 1.2, -27.4, h + 4.4, -28.6, h + 26, -28.6, h + 29, -27,
		h + 40, -12, h + 64, -9, h + 70, -6.6, h + 72, -2, h + 72, 14, h + 69, 20,
	}
	p.Stroke(ink, 2*Line, append(body, body[:2]...)...)
	p.Poly(paint, body...)
	// Shade along the bottom and the wheel arch cut round the tyre.
	p.Poly(paintShade, h+0.8, 12, h+71.2, 12, h+71.2, 14, h+68.6, 19.2, h+0.8, 19.2)
	arch := []float64{front - 14.2, 20, front - 13, 13.6, front - 8, 9.8, front + 8, 9.8, front + 13, 13.6, front + 14.2, 20}
	p.Poly(tyre, arch...)
	// The window, with the seat's back through it, and the door.
	win := []float64{h + 4, -25.4, h + 25.6, -25.4, h + 36.4, -12.4, h + 4, -12.4}
	p.Poly(ink, h+3.2, -26.2, h+26, -26.2, h+37.6, -11.6, h+3.2, -11.6)
	p.Poly(glass, win...)
	p.Poly(glassLight, h+4, -25.4, h+12, -25.4, h+6, -12.4, h+4, -12.4)
	p.Poly(trim, h+8, -21, h+12.6, -21.6, h+13.4, -12.4, h+8.2, -12.4)
	p.Stroke(resurrect64.Fog, 0.6, h+3, -12, h+3, 17, h+36, 17, h+40, -11)
	p.Stroke(trim, 1.4, h+6.4, -8.4, h+10, -8.4)
	// The mirror, the emblem and the lettering.
	p.Poly(ink, h+36.4, -15, h+41, -15.4, h+41, -10.6, h+36.4, -10.6)
	badge(p, h+19, 7.4)
	letter(p, "PARKS AND", h+19.5, -9.4)
	letter(p, "WILDLIFE", h+19.5, -4.6)
	// The snorkel up the windscreen pillar, the aerial and the bonnet's
	// shine.
	snork := []float64{front - 9, -2, h + 42.4, -12.4, h + 33, -28, h + 33, -32.6, h + 38, -33}
	p.Stroke(ink, 3+2*Line, snork...)
	p.Stroke(trim, 3, snork...)
	p.Stroke(ink, 0.7, h+52, -10, h+52, -34)
	p.Curve(resurrect64.White, 1, h+42, -11.2, h+54, -9.4, h+64, -7.8)
	// The headlight, the bull bar with its spotlights and indicator.
	p.Ellipse(h+68.4, -4, 2, 2.4, 0, ink)
	p.Ellipse(h+68.2, -4.2, 1.4, 1.8, 0, glassLight)
	p.Poly(ink, h+69, 2, h+79, 2, h+79, 18, h+69, 18)
	p.Poly(trim, h+69.8, 2.8, h+78.2, 2.8, h+78.2, 17.2, h+69.8, 17.2)
	p.Stroke(ink, 2+2*Line, h+70, 2, h+70, -7, h+76, -9, h+78, -3, h+78, 2)
	p.Stroke(trim, 2, h+70, 2, h+70, -7, h+76, -9, h+78, -3, h+78, 2)
	for _, x := range []float64{h + 72, h + 76.4} {
		p.Circle(x, -11.4, 2.2+Line, ink)
		p.Circle(x, -11.4, 2.2, trim)
		p.Circle(x+0.4, -11.4, 1.4, resurrect64.Lemon)
	}
	p.Poly(resurrect64.Amber, h+75, 8, h+78.2, 8, h+78.2, 10.6, h+75, 10.6)
}

// badge draws the emblem about (x, y): a round dark green badge ringed in
// white, with a white koala's head on it (DESIGN §2).
func badge(p Pen, x, y float64) {
	p.Circle(x, y, 4.6+Line, ink)
	p.Circle(x, y, 4.6, paint)
	p.Circle(x, y, 3.8, emblem)
	for _, e := range []float64{-1, 1} {
		p.Circle(x+e*1.7, y-1.2, 1, paint)
	}
	p.Ellipse(x, y+0.4, 1.8, 1.6, 0, paint)
	p.Ellipse(x, y+0.8, 0.5, 0.7, 0, emblem)
}

// letter writes s in the door's dark green lettering, centred on x with
// the top of its line at y, in field pixels.
func letter(p Pen, s string, x, y float64) {
	f := gfx.NewVectorFont(max(1, int(math.Round(4.4*p.K*p.size()))))
	sx, sy := p.pt(x, y)
	f.Draw(p.Dst, s, int(sx)-f.Width(s)/2, int(sy), lettering)
}
