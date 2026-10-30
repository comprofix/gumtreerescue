package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
)

// The cows stand in the paddocks, between the horizon and the fence, the
// farmland behind the big gum has them too, and they are Holsteins and
// Jerseys mixed (DESIGN D35).
func TestCows(t *testing.T) {
	fw, _ := config.Field()
	behindGum, jerseys := 0, 0
	for _, c := range art.Cows {
		if c.Jersey {
			jerseys++
		}
		if c.X < 0 || c.X > float64(fw) || c.Y <= art.Horizon || c.Y >= art.FenceTop {
			t.Errorf("a cow at (%.0f, %.0f) is out of the paddocks", c.X, c.Y)
		}
		if c.X < 160 {
			behindGum++
		}
	}
	if behindGum == 0 {
		t.Error("no cows in the farmland behind the big gum")
	}
	if jerseys == 0 || jerseys == len(art.Cows) {
		t.Errorf("%d Jerseys of %d cows, want a mix", jerseys, len(art.Cows))
	}
}

// The farmland behind the big gum has a tree for the birds to perch in.
func TestPerchBehindGum(t *testing.T) {
	for _, p := range art.Perches() {
		if !p.Fence && p.X < 180 {
			return
		}
	}
	t.Error("no perch in the farmland behind the big gum")
}

// The pixel backdrop spans the field: its back from below the clear sky
// down past the ground's top, its front down to the bottom, over the
// fence (DESIGN D30, D34).
func TestBackdropSprites(t *testing.T) {
	index := sheetIndex(t, "backdrop")
	fw, fh := config.Field()
	back, ok := index["backdrop_back"]
	if !ok || back.W != fw {
		t.Errorf("backdrop_back %+v, want %d wide", back, fw)
	}
	front, ok := index["backdrop_front"]
	if !ok || front.W != fw || backdropFrontY+front.H != fh {
		t.Errorf("backdrop_front %+v from %d, want %d wide down to %d", front, backdropFrontY, fw, fh)
	}
	if backdropBackY+back.H < config.GroundY {
		t.Errorf("backdrop_back from %d is %d tall, short of the ground at %d", backdropBackY, back.H, config.GroundY)
	}
}
