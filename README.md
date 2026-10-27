# Gum Tree Rescue

Created from [gameskeleton](https://github.com/comprofix/gameskeleton) v0.5.1.

## Requirements

- Go 1.27 or later
- **Linux only:** cgo plus the X11/GL/ALSA development packages:
  `libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`

## Build and run

```sh
go run ./cmd/gumtreerescue                 # run the game
go run -tags debug ./cmd/gumtreerescue     # with the debug overlay
go generate ./assets               # rebuild sprite sheets
go vet ./...
go test ./...
go build -o bin/ ./cmd/gumtreerescue
```

## Personal documents

`DESIGN.md` and `PLAN.md` are personal working documents. `.gitignore` keeps
them out of git, so back them up somewhere else. `CLAUDE.md` is tracked, so
anyone working on the game with Claude Code gets the same guidance; put
personal instructions in `CLAUDE.local.md`, which Claude Code also reads and
`.gitignore` keeps out of git.

## License

MIT. See [LICENSE](LICENSE).
