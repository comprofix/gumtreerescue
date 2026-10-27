# CLAUDE.md

Guidance for Claude Code when working in this repo. This file is tracked in git
and comes with every game made from the skeleton. Personal instructions that
should not be shared go in `CLAUDE.local.md` beside it, which Claude Code also
reads and `.gitignore` keeps out of git.

## 1. Project

**Gum Tree Rescue:** an original single-screen arcade game with an
Australian theme. Koalas fall from a gum tree on the left, and two rangers
with a sheet bounce them across the screen into the tray of a ute parked on
the right. Built from the Game Skeleton (Go and Ebitengine).

- Remote: `git@github.com:comprofix/gumtreerescue.git`
- `DESIGN.md` and `PLAN.md` are local working documents, **not tracked in git**.

Fixed decisions (DESIGN.md §10 has them all):

- **A brand-new game** in the catch-and-bounce genre, not a remake. No
  other game, decompile or clone is used as a reference (§9, DESIGN §0).
- **Both looks:** Smooth (1280×720), which players start on, and Modern
  pixel art (640×360), switched in Options. Resurrect 64 and `Chip4`. Game
  coordinates are those of the 640×360 screen.
- **Bounces:** real gravity, true free fall from the branch, and a fixed
  rebound peak, so every bounce takes the same time. The sheet has five
  zones (−2, −1, 0, +1, +2 steps) and koalas land on stations one step
  apart.
- **Single player**, with Easy, Normal and Hard skills and a high score
  table for each.
- **CI/CD is set up:** release builds by GitHub Actions when a version tag
  is pushed (§5).

## 2. Working documents

DESIGN.md and PLAN.md are local and not tracked in git. If either is missing
at the start of a session, ask the user for their copies before doing
anything else. Don't write new ones unless the user asks. This file is
tracked; personal instructions go in CLAUDE.local.md, which is not.

## 3. Phases and decisions

- **Design phase:** only DESIGN.md, PLAN.md and CLAUDE.md change. Do not write
  code, install anything or create build files until the user explicitly says
  to start a milestone.
- **Build phase:** one PLAN.md milestone at a time. Each milestone ends with a
  playable build and passing tests. Tick its checkboxes when it's done.
- **Decisions** live in the DESIGN.md decisions table. Never pick an answer to
  an open decision yourself; bring the question to the user. Once it's decided,
  update DESIGN.md and PLAN.md so they stay the single source of truth.
- Rules or facts that aren't confirmed yet are tagged **[verify]** in
  DESIGN.md. Flag them rather than guessing.

## 4. Working with the user

- **Confirm before** creating, changing or deleting files, installing modules,
  or making commits or pushes. Reading and inspecting are fine. Planning and
  discussion are not a go-ahead.
- **Never commit or push** unless the user says so for that specific commit or
  push.
- **One commit per change.** When a milestone's checks pass, report back and
  wait. Once the user approves, commit each distinct change on its own (one
  fix, feature or milestone per commit, never bundled).
- **Commit messages follow Conventional Commits 1.0.0:**
  `type(scope): description`, with an optional body and footers.
  - Types: `feat` (something new for the player), `fix` (a bug fix), `docs`,
    `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.
  - The scope is optional and names the area or package, e.g. `feat(audio):`,
    `fix(world):`.
  - The description is short, lower case and imperative, with no full stop.
  - An incompatible change adds `!` after the type or a `BREAKING CHANGE:`
    footer.
  - Milestone names (`K1`, `M3`) stay in PLAN.md and never appear in commit
    messages.
- **Versions follow SemVer**, worked out from the commits since the last tag:
  any breaking change makes a major bump, any `feat` a minor bump, otherwise
  any `fix` or `perf` a patch bump. A release is an annotated tag `vX.Y.Z`
  with the message `<Title> vX.Y.Z`.
- Commits have no `Co-Authored-By: Claude` or `Claude-Session:` trailers. The
  user is the only author.
- **CI/CD** is set up only as the user chose at setup (§2), or later when
  they ask, and then only as in §5. Do not add or change other GitHub
  workflows unless the user explicitly asks.
- **Do not track** DESIGN.md, PLAN.md, CLAUDE.local.md or `.claude/` in git,
  and do not remove them from `.gitignore`.
- Report results truthfully: if tests fail, show the output; if a step was
  skipped, say so.

## 5. Stack and commands

- Go 1.27 and Ebitengine v2.10.5.
- Settings and scores are saved as `settings.json` and `scores.json` next to
  the game.

```sh
go run ./cmd/gumtreerescue                 # run the game
go run -tags debug ./cmd/gumtreerescue     # with the debug overlay
go run ./cmd/gumtreerescue -seed 42        # with a fixed random seed
go run ./cmd/gumtreerescue -level 5        # start games on level 5, for a game with levels (no scores)
go generate ./assets                      # rebuild sprite sheets from assets/sprites/*.txt
go run ./tools/spritegen -in assets/sprites -out assets/gen -preview <dir>
go build -o bin/ ./cmd/gumtreerescue
go vet ./...
go test ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/gumtreerescue.exe ./cmd/gumtreerescue
```

Linux builds need cgo and the X11/GL/ALSA development packages:
`libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev
libxxf86vm-dev libasound2-dev pkg-config`. Build them in WSL.

### CI/CD and releases

`tools/release/` holds the templates (its README says what each does). A tag
`vX.Y.Z` pushed to GitHub runs vet and tests on Ubuntu 22.04 and Windows,
builds a Windows zip and a Linux tarball on Ubuntu, and publishes a GitHub
release with checksums and that version's CHANGELOG.md section as its notes.
There is no workflow on every push.

**Setting it up**, each step its own commit once the user agrees:

1. Copy `tools/release/release.yml` to `.github/workflows/`, and
   `tools/release/cliff.toml` and `tools/release/CHANGELOG.md` to the root;
   then delete `tools/release/`.
2. Add a Download section to the README, after the game's description:

   ```
   ## Download

   Builds for Windows and Linux (64-bit) are on the
   [Releases](https://<repo>/releases) page, with `SHA256SUMS.txt` to check
   them against.

   - **Windows:** unzip `<game>-vX.Y.Z-windows-amd64.zip` and run `<game>.exe`.
   - **Linux:** extract `<game>-vX.Y.Z-linux-amd64.tar.gz` and run `./<game>`.
     It needs OpenGL, X11 and ALSA, which desktop systems already have.
   ```
3. Note in §1 that CI/CD is set up, and in the decisions table of DESIGN.md.

The release fails while the program icon is still the skeleton's
placeholder (`icon.Placeholder`, §7), so draw the game's own icon before the
first release.

**Releasing vX.Y.Z**, only when the user says, the version worked out by
SemVer (§4) from the commits since the last tag. Needs
[git-cliff](https://git-cliff.org).

1. `git cliff --unreleased --tag vX.Y.Z --prepend CHANGELOG.md`.
2. Tidy the new section's wording, since it becomes the release notes.
   Anything players should know from `docs`, `test`, `build`, `ci` or
   `chore` commits, which git-cliff leaves out, is added by hand.
3. Commit it as `chore(release): vX.Y.Z`, then
   `git tag -a vX.Y.Z -m "<Title> vX.Y.Z"`. The tag's message is the
   release's title.
4. Push the commit, then `git push origin vX.Y.Z`.

Never regenerate the whole CHANGELOG.md: the tidied wording would be lost.

## 6. Code conventions

- Standard Go style: `gofmt`, small packages under `internal/`, no global
  variables that change.
- **Tick-based (60 ticks per second) and deterministic.**
  - No `time.Now()` in gameplay.
  - All randomness comes from the RNG in `Context`, created from the seed.
  - Use whole-number or fixed-point positions where exact results matter.
- `Update` changes state; `Draw` only reads it. Never change state in `Draw`.
- **Game logic runs headless:** no GPU calls in `Update`. Game-logic packages
  must not import graphics or audio packages. Scenes and audio react to game
  state and events.
- **Project settings** are chosen only in `internal/config/project.go`:
  - the title and tick rate;
  - `Project`, the preset, which sets the resolution, scale modes and
    palette;
  - `Second`, the second look in a game with two, or nil (§7).

  Presets are defined in `internal/config/preset.go`. The game starts from
  `config.Project`; scenes read the screen size from `Context.Preset`, the
  look shown, and the drawing helpers from the image they draw on. Never
  hard-code a screen size or palette anywhere else.
- Colours are named with the constants of the chosen palette's package (e.g.
  `resurrect64.Gold`) and turned into RGBA with `gfx.RGBA`.
- Tuning values (speeds, timings, scores, levels, skill tables) live in
  `internal/config`. No magic numbers in game code.
- Colours come from the palette, by name. No hex codes outside the palette
  files, including in sprite files.
- Assets are built into the game with `go:embed` from `assets/`.
- A game's own code cites its DESIGN sections (e.g. "DESIGN §5") to explain
  why it does something. Shared skeleton code does not, because each game's
  DESIGN.md is its own; it is explained by its package docs instead.
- Alt+Enter, F9, F10 and F11 are reserved for the window and never trigger
  game actions.
- **The mouse** works in both styles. `Input.Pointer` gives it in screen
  pixels, through the letterbox in every scale mode; `Clicked`, `MouseHeld`
  and `Wheel` give the buttons and wheel, and a right click is Back and
  Pause, as Esc is. A `ui.Menu` works by mouse once placed (`Place` or
  `PlaceCentred`, before its `Update`, with what its `Draw` is given).
  Scripted input drives the mouse too (`input.MoveTo`, `input.Click`).
- The game plugs into the shared scenes through `game.Content` (its play
  scene, sounds and sprite sheets; optionally a demo behind the title and
  its own title and high score tunes; in a game with two looks,
  `NativePlay` once its play scene is drawn for both, §7), set in one file
  such as `internal/game/<game>.go`. Until it is set, Play is disabled in
  the main menu:

  ```go
  func init() {
  	content = Content{NewPlay: play.New, Sounds: play.Sounds, Sheets: []string{"<game>"}}
  }
  ```

### Package map

```
cmd/<game>/        main.go: flags, saved files, window, run
cmd/winicon/       writes the program icon as a .ico for the Windows exe
internal/
  game/            ebiten.Game: scenes, window keys, scale modes, presentation
  scene/           scene manager, fades, shake, shared scenes
  scene/scenetest/ headless test harness, golden files
  ui/              list menu (keys and mouse), panel, text helpers
  shapes/          smooth shapes, panels and buttons in palette colours
  config/          project.go, presets and looks, settings, tuning values,
                   skills, levels
  session/         one game in progress
  score/           high score tables
  gfx/             palette, schemes, sheets, animation, fonts (bitmap and
                   Go Mono Bold), scaler
  gfx/palettes/    ready-made palettes
  audio/           synth, tunes, sound effects, scheduling, output
  input/           actions, keyboard, gamepad, mouse, scripted input
  spritegrid/      text-grid sprite parser and packer
  bitfont/         bitmap fonts without a GPU (game and tools)
  debug/           debug.Enabled (build tag: debug)
  icon/            the program icon at every size, from assets/sprites/icon.txt
  <game logic>/    the game's own world, with no graphics or audio imports
tools/spritegen/   sprites → PNG + JSON (+ preview)
assets/            assets.go, sprites/, gen/, font/
```

## 7. Art

The game's art style, chosen at setup (§2), is pixel art, smooth, or both.
Either way colours come from the palette by name, and the program icon is a
sprite (below).

### Pixel art

- Claude makes all the art as **text-grid sprites** in `assets/sprites/*.txt`
  (format in the `spritegrid` package docs):
  - `.` is a see-through pixel.
  - With palettes of up to 16 colours, each pixel can be a hex digit.
  - Larger palettes use a `colours` key in each file that names the palette
    colours.
  - `recolour` makes colour-swapped copies.
- Facing left is made by mirroring when drawn. Do not draw separate
  left-facing art.
- After changing a sprite, run `go generate ./assets`. The regenerated
  `assets/gen/*` is part of the change.
- Check new art by eye on the preview sheet before using it. Previews go in the
  scratchpad or `/preview/`, never in `assets/gen`.
- Effects (lighting, particles, weather) are drawn by code at runtime, not
  stored as sprites.
- The game only loads PNG and JSON, so a sprite drawn by hand can replace a
  generated one without code changes.

### Smooth

- Claude draws the art in code with `internal/shapes`: rectangles, rounded
  rectangles, circles, lines, polygons and arcs, anti-aliased, and panels
  and buttons styled from the colour scheme (`PanelFor`, `ButtonFor`).
  Pictures that are needed often (a card, a chip) are drawn by one function
  each, in the game's own package.
- The screen is 1280×720 and is scaled smoothly to the window, so lay
  things out by the screen and font sizes, not by whole pixels.
- Text is Go Mono Bold (`gfx.SmoothFonts`), monospaced like the bitmap
  fonts, so menus line up in character cells.
- Sprites still work, for anything better drawn as an image.
- Check new art by eye from a screenshot before using it. Preview programs
  go in the scratchpad, never in the repo.

### Both looks

- `config.Project` is the first look, which players start on, and
  `config.Second` the other; Options opens on Look, and the choice is saved.
  `-style both` gives Smooth first and Modern second.
- The two must pass `Looks.Check`, run at start-up: one pixel, one smooth,
  the same palette and audio style, the smooth screen a whole multiple of
  the pixel one. Smooth is Modern ×2. For Classic PC's 320×200 at ×4, add a
  1280×800 smooth preset and a 320×200 pixel one with the same palette.
- Game coordinates stay those of the pixel screen, so the gameplay is the
  same in both looks; only the drawing differs.
- **The shared scenes are native**: they draw at the shown screen's size in
  either look. A game's own scene is pixel art until it says otherwise:
  while the smooth look is shown it is made, updated and drawn with the
  pixel look's preset and fonts, then enlarged with sharp pixels. The
  pointer is still in the shown screen's pixels, so such a scene divides it
  by the multiple.
- Redraw scenes for the smooth look one at a time. A scene drawn for both
  has a `Native()` method and reads `Context.Preset` to choose how to draw.
  Once the play scene (and demo) are native, set `Content.NativePlay`; the
  game panics if the two disagree.
- Check both looks by screenshot whenever a scene's drawing changes.

### Program icon

`assets/sprites/icon.txt` holds the window's and the exe's icon, drawn at
16 and 32 pixels, in both styles. The skeleton's is a placeholder, and
`icon.Placeholder` says so. Once the game is designed and built, draw the
game's own in its place and set `icon.Placeholder` to false. Until then
`cmd/winicon` refuses to write the exe's icon, so a release fails.

## 8. Audio

- All sound is made in code by the synth in `internal/audio`. No audio files
  unless the user agrees.
- Tunes and sound effects are note tables in Go. A nil `*audio.Synth` ignores
  every call, so tests run without sound.

## 9. Originality and reference material

- All art, music, sound and text is original. Never add or reference another
  game's assets: no ripped sprites, no transcribed music, no copied text.
- **For a brand-new game "in the style of" a classic:** do not use the
  original's name, characters, art, music or level layouts.
- **For a remake, follow the facts-only rule.** Disassemblies, source code or
  documents about the original may be read **for facts only**: rules,
  formulas, counts, thresholds and difficulty tables.
  - Record each fact in DESIGN.md, tagged **[ref]**, before using it.
  - Never copy or translate its code, sprites, fonts, sound tables or text.
  - Write Go code from DESIGN.md, not from the reference.
  - Clone reference repos into the scratchpad, never into this repo.
- Reference videos and screenshots go in `ref/`, which is git-ignored.
  - Extract frames into the scratchpad.
  - Use them for layout, behaviour and timing only. Never trace sprites from
    them.

## 10. Testing

- Every change to game logic comes with unit tests that run without a window
  or sound.
- Prefer golden replay tests (seed plus inputs → expected final state) for
  gameplay sequences.
- `internal/scene/scenetest` runs scenes headless from a seed and scripted
  input, and `scenetest.Golden` compares results with `testdata/*.golden`.
  After an intended change, rewrite them with `go test ./<package> -update`.
- Run `go vet ./...` and `go test ./...` before saying work is done.
- Visual and audio work: check by hand against a short checklist at each
  milestone, and offer screenshots for the user to review.
