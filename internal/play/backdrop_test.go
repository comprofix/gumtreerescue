package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/config"
)

// The pixel backdrop spans the field: its back from the top, its front
// down to the bottom, over the fence (DESIGN D30).
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
	if back.H < config.GroundY {
		t.Errorf("backdrop_back is %d tall, short of the ground at %d", back.H, config.GroundY)
	}
}
