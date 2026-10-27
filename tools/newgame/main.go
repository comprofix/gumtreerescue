// Command newgame turns a fresh copy of the skeleton into a new game.
// It does the steps that must be exact:
//
//	go run ./tools/newgame -module github.com/comprofix/<game> -name <game> -title "<Game Title>" [-style pixel|smooth|both] [-keep-demo]
//
// It changes the module path in go.mod and every import, and the skeleton's
// name in code, in CLAUDE.md and in the release templates; renames
// cmd/gameskeleton to cmd/<game>; sets config.Title; for the smooth art
// style sets config.Project to the Smooth preset (pixel art keeps Modern,
// to be changed by hand for another pixel preset); for both looks sets
// config.Project to Smooth and config.Second to Modern; removes the demo
// unless -keep-demo is given, and always for smooth or both, as the demo is
// pixel art;
// writes a new README saying which skeleton version it
// came from; removes itself; and runs go build, go vet and go test to check
// everything still works.
//
// It refuses to run where DESIGN.md exists, so it can't damage the skeleton
// itself, whose working documents live beside it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SkeletonVersion is the skeleton release this copy is. Keep it in step with
// the release tag.
const SkeletonVersion = "0.5.1"

const (
	oldModule = "github.com/comprofix/gameskeleton"
	oldName   = "gameskeleton"
)

// The demo's files and folders, relative to the root.
var demoPaths = []string{
	"internal/demo",
	"internal/game/demo.go",
	"assets/sprites/demo.txt",
	"assets/gen/demo.png",
	"assets/gen/demo.json",
}

// Folders the rewrite never enters.
var skipDirs = map[string]bool{".git": true, ".claude": true, "bin": true, "dist": true, "ref": true, "preview": true}

type options struct {
	dir, module, name, title, style string
	keepDemo, check                 bool
}

// The art styles, as -style takes them, and the preset each sets: for both
// looks the first, with secondLook as the second.
var stylePresets = map[string]string{"pixel": "Modern", "smooth": "Smooth", "both": "Smooth"}

// secondLook is the second look -style both sets.
const secondLook = "Modern"

func main() {
	var o options
	flag.StringVar(&o.dir, "dir", ".", "the copy of the skeleton to change")
	flag.StringVar(&o.module, "module", "", "the new module path, e.g. github.com/comprofix/mygame")
	flag.StringVar(&o.name, "name", "", "the command's name, e.g. mygame (lower case letters, digits and -)")
	flag.StringVar(&o.title, "title", "", `the game's title, e.g. "My Game"`)
	flag.StringVar(&o.style, "style", "pixel", "the art style: pixel, smooth, or both (Smooth first, Modern second, switched in Options)")
	flag.BoolVar(&o.keepDemo, "keep-demo", false, "keep the Dot Dash demo (pixel art only)")
	flag.BoolVar(&o.check, "check", true, "run go build and go test afterwards")
	flag.Parse()
	if err := run(o, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "newgame:", err)
		os.Exit(1)
	}
}

var validName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func run(o options, log io.Writer) error {
	switch {
	case o.module == "" || strings.ContainsAny(o.module, " \t\"") || o.module == oldModule:
		return fmt.Errorf("-module %q: give the new game's module path", o.module)
	case !validName.MatchString(o.name) || o.name == oldName:
		return fmt.Errorf("-name %q: use lower case letters, digits and -, starting with a letter", o.name)
	case strings.TrimSpace(o.title) == "":
		return errors.New("-title: give the game's title")
	case stylePresets[o.style] == "":
		return fmt.Errorf("-style %q: use pixel, smooth or both", o.style)
	case o.style != "pixel" && o.keepDemo:
		return errors.New("-keep-demo: the demo is pixel art, so only a pixel game can keep it")
	}
	root := o.dir
	if _, err := os.Stat(filepath.Join(root, "DESIGN.md")); err == nil {
		return errors.New("DESIGN.md exists here: newgame runs on a fresh copy before DESIGN.md is written (move it aside to go on)")
	}
	gomod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(gomod), "module "+oldModule+"\n") {
		return errors.New("go.mod is not the skeleton's: has this copy already been set up?")
	}

	// Remove first, so nothing is rewritten only to be deleted.
	remove := []string{"tools/newgame"}
	if !o.keepDemo {
		remove = append(remove, demoPaths...)
	}
	for _, p := range remove {
		if err := os.RemoveAll(filepath.Join(root, filepath.FromSlash(p))); err != nil {
			return err
		}
		fmt.Fprintln(log, "removed", p)
	}

	n, err := rewrite(root, o)
	if err != nil {
		return err
	}
	fmt.Fprintf(log, "renamed the module and %s in %d files\n", oldName, n)

	if err := os.Rename(filepath.Join(root, "cmd", oldName), filepath.Join(root, "cmd", o.name)); err != nil {
		return err
	}
	fmt.Fprintf(log, "moved cmd/%s to cmd/%s\n", oldName, o.name)

	project := filepath.Join(root, "internal", "config", "project.go")
	if err := setTitle(project, o.title); err != nil {
		return err
	}
	fmt.Fprintf(log, "set the title to %q\n", o.title)
	if err := setPreset(project, stylePresets[o.style]); err != nil {
		return err
	}
	if o.style == "both" {
		if err := setSecond(project, secondLook); err != nil {
			return err
		}
		fmt.Fprintf(log, "set the looks to %s and %s\n", stylePresets[o.style], secondLook)
	} else {
		fmt.Fprintf(log, "set the preset to %s for %s art\n", stylePresets[o.style], o.style)
	}

	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(readme(o)), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(log, "wrote README.md")

	if o.check {
		for _, args := range [][]string{{"build", "./..."}, {"vet", "./..."}, {"test", "./..."}} {
			fmt.Fprintln(log, "go", strings.Join(args, " "))
			cmd := exec.Command("go", args...)
			cmd.Dir, cmd.Stdout, cmd.Stderr = root, log, log
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("go %s: %w", args[0], err)
			}
		}
	}
	fmt.Fprintln(log, "done")
	return nil
}

// rewrite changes the module path, then the skeleton's name, in every Go
// file, go.mod and Markdown file, and returns how many it changed.
func rewrite(root string, o options) (int, error) {
	changed := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".mod", ".md", ".yml", ".toml":
		default:
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		s := strings.ReplaceAll(string(data), oldModule, o.module)
		s = strings.ReplaceAll(s, oldName, o.name)
		if s == string(data) {
			return nil
		}
		changed++
		return os.WriteFile(path, []byte(s), 0o644)
	})
	return changed, err
}

var titleLine = regexp.MustCompile(`(?m)^const Title = ".*"$`)

// setTitle sets config.Title in project.go.
func setTitle(path, title string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !titleLine.Match(data) {
		return fmt.Errorf("%s: no `const Title = \"...\"` line", path)
	}
	out := titleLine.ReplaceAllLiteral(data, []byte("const Title = "+strconv.Quote(title)))
	return os.WriteFile(path, out, 0o644)
}

var presetLine = regexp.MustCompile(`(?m)^var Project = \w+$`)

// setPreset sets config.Project in project.go to the named preset.
func setPreset(path, preset string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !presetLine.Match(data) {
		return fmt.Errorf("%s: no `var Project = ...` line", path)
	}
	out := presetLine.ReplaceAllLiteral(data, []byte("var Project = "+preset))
	return os.WriteFile(path, out, 0o644)
}

var secondLine = regexp.MustCompile(`(?m)^var Second \*Preset$`)

// setSecond sets config.Second in project.go to the named preset.
func setSecond(path, preset string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !secondLine.Match(data) {
		return fmt.Errorf("%s: no `var Second *Preset` line", path)
	}
	out := secondLine.ReplaceAllLiteral(data, []byte("var Second = &"+preset))
	return os.WriteFile(path, out, 0o644)
}

// readme is the new game's starting README.
func readme(o options) string {
	return fmt.Sprintf(`# %s

Created from [gameskeleton](https://github.com/comprofix/gameskeleton) v%s.

## Requirements

- Go 1.27 or later
- **Linux only:** cgo plus the X11/GL/ALSA development packages:
  `+"`libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`"+`

## Build and run

`+"```sh"+`
go run ./cmd/%[3]s                 # run the game
go run -tags debug ./cmd/%[3]s     # with the debug overlay
go generate ./assets               # rebuild sprite sheets
go vet ./...
go test ./...
go build -o bin/ ./cmd/%[3]s
`+"```"+`

## Personal documents

`+"`DESIGN.md` and `PLAN.md`"+` are personal working documents. `+"`.gitignore`"+` keeps
them out of git, so back them up somewhere else. `+"`CLAUDE.md`"+` is tracked, so
anyone working on the game with Claude Code gets the same guidance; put
personal instructions in `+"`CLAUDE.local.md`"+`, which Claude Code also reads and
`+"`.gitignore`"+` keeps out of git.

## License

MIT. See [LICENSE](LICENSE).
`, o.title, SkeletonVersion, o.name)
}
