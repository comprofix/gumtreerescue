package play

import (
	"math"
	"testing"
)

// The ute's sprites are in the sheet, with a wheel frame for each quarter
// of the way between its holes (DESIGN D30).
func TestUteSprites(t *testing.T) {
	index := sheetIndex(t, "ute")
	for _, name := range []string{"ute_body", "ute_side", "ute_flare"} {
		if _, ok := index[name]; !ok {
			t.Errorf("no sprite %s", name)
		}
	}
	if w := index["ute_wheel"]; w.Frames != uteWheelFrames || w.W != 2*uteWheelMid+1 || w.H != 2*uteWheelMid+1 {
		t.Errorf("ute_wheel: %d×%d, %d frames", w.W, w.H, w.Frames)
	}
}

// The wheel steps a frame each quarter of the way between its holes, and
// comes round again at the next hole, either way.
func TestUteWheelFrame(t *testing.T) {
	step := math.Pi / 3 / uteWheelFrames
	for i, want := range []int{0, 1, 2, 3, 0, 1} {
		if got := uteWheelFrame(float64(i) * step); got != want {
			t.Errorf("turned %d steps: frame %d, want %d", i, got, want)
		}
	}
	if got := uteWheelFrame(-step); got != uteWheelFrames-1 {
		t.Errorf("turned back a step: frame %d", got)
	}
}
