package art

import (
	"math"
	"testing"
)

// The python's body follows the same curves in the dirt as it slides: a
// full turn of Wave for each curve's length along.
func TestPythonWave(t *testing.T) {
	if w := PythonWave(PythonLength / pythonWaves); math.Abs(w-2*math.Pi) > 1e-9 {
		t.Errorf("a curve's length along: %.3f radians, want 2π", w)
	}
}

// hash scatters the same way every time, from 0 to 1.
func TestHash(t *testing.T) {
	for n := range 200 {
		if h := hash(n); h < 0 || h >= 1 || h != hash(n) {
			t.Fatalf("hash(%d) = %v", n, h)
		}
	}
}
