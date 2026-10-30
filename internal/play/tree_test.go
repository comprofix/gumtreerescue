package play

import (
	"testing"

	"github.com/comprofix/gumtreerescue/internal/config"
)

// Every drop point any level has has its branch in the tree sheet
// (DESIGN D30).
func TestTreeBranches(t *testing.T) {
	index := sheetIndex(t, "tree")
	if _, ok := index["tree"]; !ok {
		t.Fatal("no tree sprite")
	}
	for n := 1; n <= len(config.Levels)+1; n++ {
		for _, d := range config.LevelAt(n).Drops {
			found := false
			for _, b := range treeBranches {
				if b.drop == d {
					found = true
					if _, ok := index[b.name]; !ok {
						t.Errorf("no sprite %s", b.name)
					}
				}
			}
			if !found {
				t.Errorf("level %d: no branch for drop %+v", n, d)
			}
		}
	}
}
