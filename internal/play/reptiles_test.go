package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/art"
	"github.com/comprofix/gumtreerescue/internal/config"
	"github.com/comprofix/gumtreerescue/internal/scene"
)

// TestReptiles runs the monitor and the python for half an hour: they keep
// to the track and near the screen, each wanders, rests and leaves, and
// comes back; the python only moves uncurled (DESIGN D32).
func TestReptiles(t *testing.T) {
	r := newReptiles(scene.NewSceneryRNG(5))
	fw, _ := config.Field()
	mon, py := map[reptileState]bool{}, map[reptileState]bool{}
	visits := [2]int{}
	was := [2]reptileState{offTrack, offTrack}
	for range 30 * 60 * config.TPS {
		x := r.python.x
		r.step()
		m, s := &r.monitor, &r.python
		mon[m.state], py[s.state] = true, true
		for i, st := range []reptileState{m.state, s.state} {
			if was[i] == offTrack && st != offTrack {
				visits[i]++
			}
			was[i] = st
		}
		for _, a := range []struct {
			name  string
			state reptileState
			x, y  float64
			reach float64
		}{{"monitor", m.state, m.x, m.y, art.MonitorBehind}, {"python", s.state, s.x, s.y, art.PythonLength}} {
			if a.state == offTrack {
				continue
			}
			if a.y < trackTop || a.y > trackLow || a.x < -a.reach-1 || a.x > float64(fw)+a.reach+1 {
				t.Fatalf("the %s at (%.1f, %.1f), off the track", a.name, a.x, a.y)
			}
		}
		if s.coil > 0 && s.x != x && s.state != offTrack {
			t.Fatalf("the python moved curled up: coil %.2f", s.coil)
		}
	}
	for _, st := range []reptileState{offTrack, wandering, resting} {
		if !mon[st] || !py[st] {
			t.Errorf("state %d: monitor %v, python %v", st, mon[st], py[st])
		}
	}
	if visits[0] < 3 || visits[1] < 3 {
		t.Errorf("the monitor came along %d times and the python %d, want 3 or more each", visits[0], visits[1])
	}
}
