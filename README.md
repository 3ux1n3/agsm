# AGSM

AGSM is a terminal UI for discovering, browsing, and resuming coding-agent sessions from one place.

Current public scope:
- OpenCode support
- unified session list across OpenCode storage
- resume, refresh, rename, delete, and new-session launch flow
- adaptive terminal theming with a full-screen TUI

## Status

AGSM is early `v0.x` software.

The current release focuses on OpenCode first. Claude Code and Codex support are planned, but not included yet.

## Features

- Full-screen TUI for OpenCode sessions
- Session discovery from both:
  - OpenCode JSON session storage
  - OpenCode SQLite-backed session storage
- Search and quick navigation
- Resume selected session in OpenCode
- Rename sessions with AGSM-managed metadata
- Delete sessions from the UI
- Start a new OpenCode session from a chosen directory

## Requirements

- Go installed locally if building from source
- `opencode` installed and available on `PATH`
- macOS or Linux

## Install

```bash
go install github.com/3ux1n3/agsm@latest
```

## Run

If installed with `go install`:

```bash
agsm
```

From this repository:

```bash
make run
```

## Build

```bash
make build
```

## Test

```bash
make test
```

## Configuration

Runtime config lives in `~/.config/agsm/`.

- config file: `~/.config/agsm/config.toml`
- metadata file: `~/.config/agsm/metadata.json`

This repository also includes a local `.config/` directory for project-owned examples and future development config.

Example config:

```toml
sort_by = "last_active"
sort_order = "desc"

[agents.opencode]
enabled = true

[ui]
nerd_fonts = false
color_scheme = "auto"
```

See `.config/agsm.example.toml` for the committed example file.

## Notes

- OpenCode session discovery reads both legacy JSON storage and current DB-backed sessions.
- AGSM stores custom session names in its own metadata file instead of modifying OpenCode session data.
- AGSM currently targets OpenCode only.

## Keybindings

- `↑` / `↓`: move selection
- `Enter`: resume selected session
- `/`: search
- `Esc`: clear search or cancel modal
- `Ctrl+N`: new session
- `Ctrl+R`: rename session
- `Ctrl+D`: delete session
- `Ctrl+L`: refresh
- `q`: quit

## Development

Useful commands:

```bash
make fmt
make test
make build
make run
```

## Roadmap

- [ ] Add GitHub Actions CI
- [ ] Add `CONTRIBUTING.md`
- [ ] Add issue templates
- [ ] Add pull request template
- [ ] Add release automation with GoReleaser
- [ ] Add tagged release workflow
- [ ] Add Homebrew distribution
- [ ] Add Claude Code adapter
- [ ] Add Codex adapter
- [ ] Add screenshots and demo GIF to README

## License

MIT. See [`LICENSE`](./LICENSE).
