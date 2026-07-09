# AGENTS.md

This file provides guidance to AI Agents when working with code in this repository.

## Code Quality Gate (mandatory)

Any time you modify a file in this repo, before considering the task done you MUST run `just fmt` and `just lint` and confirm both pass with no errors (`just ci` also runs `go test ./...` — run that too when Go code changed). Do not report work as complete with unformatted code or outstanding lint errors.

## What this is

`pokemon-cli` (`@gyeonghokim/pokemon-cli` on npm) is a terminal app, built with [bubbletea](https://github.com/charmbracelet/bubbletea)/[lipgloss](https://github.com/charmbracelet/lipgloss), that animates Pikachu running around and resting in the user's terminal, reacting to terminal resize. **As of now only the project scaffolding/tooling exists** — `main.go` is a placeholder (`package main; func main() { fmt.Println(...) }`). The bubbletea state machine, sprite rendering, and resize handling are not implemented yet.

## Commands

Toolchain (go, node, just, golangci-lint, goreleaser, git-cliff) is pinned via [mise](https://mise.jdx.dev) in `mise.toml`. Run `mise install` once per clone/CI run before anything else.

All day-to-day commands go through `justfile`, not raw `go`/`golangci-lint` invocations:

```bash
just run          # go run .
just build        # go build -o bin/pokemon-cli .
just test         # go test ./...
just fmt          # golangci-lint fmt (gofmt + goimports)
just lint         # golangci-lint run
just lint-fix      # golangci-lint run --fix
just tidy         # go mod tidy
just ci           # fmt + lint + test — mirrors what CI runs, run this before pushing
just changelog    # git-cliff --output CHANGELOG.md (local preview)
just release-dry  # goreleaser release --snapshot --clean --skip=publish — full local release build, no publishing
```

To run a single Go test: `go test ./... -run TestName` (no test files exist yet).

## Commit conventions

Commits must be [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore:`, ...) — a husky `commit-msg` hook (`.husky/commit-msg`, config in `commitlint.config.js`) rejects non-conforming messages locally, and `wagoid/commitlint-github-action` re-checks on pull requests. Run `npm install` once after cloning so the git hook is wired up (`package.json`'s `prepare` script installs it).

## Release architecture

This is the part that spans multiple files and isn't obvious from any single one:

- **`.goreleaser.yaml`** builds Go binaries for linux/darwin/windows × amd64/arm64, archives them, and publishes: a GitHub Release, a Homebrew Cask to the separate `GyeongHoKim/homebrew-tap` repo, and a Scoop manifest to the separate `GyeongHoKim/scoop-bucket` repo. Note it uses `homebrew_casks:`, not the deprecated `brews:` pipe — goreleaser now recommends Casks even for plain CLI binaries.
- **npm distribution is hand-rolled**, not goreleaser-native (goreleaser OSS has no npm publisher). The pattern (same one esbuild/turbo use) lives under `npm/`:
  - `npm/pokemon-cli/` is the public wrapper package (`@gyeonghokim/pokemon-cli`). Its `bin.js` resolves the actually-installed platform-specific optional dependency and `spawnSync`s the real binary — it ships no binary itself.
  - `npm/platforms/<os>-<arch>/` are five thin per-platform packages (`@gyeonghokim/pokemon-cli-{darwin,linux}-{amd64,arm64}` and `-windows-amd64`), each just an `os`/`cpu`-constrained `package.json` — the binary is copied in at release time, not committed.
  - **`scripts/prepare-npm-release.mjs`** is the bridge: given a version, it scans goreleaser's `dist/` output for each platform's binary (matched by `_<goos>_<goarch>` substring in the build dir name), copies it into the matching `npm/platforms/<dir>/`, and stamps the version into all six `package.json` files (five platform packages + the wrapper's own version and its `optionalDependencies`).
  - npm publishing uses **OIDC trusted publishing**, not a stored `NPM_TOKEN` — see `permissions.id-token: write` in `.github/workflows/release.yml`. Each of the six packages needs a Trusted Publisher configured on npmjs.com pointing at this repo + `release.yml` before OIDC publish works.
- **Changelog generation is a separate concern from goreleaser's changelog**: `cliff.toml` + `.github/workflows/changelog.yml` (using `git-cliff`) generate and commit an actual `CHANGELOG.md` to `main` on tag push. goreleaser's own `changelog:` block in `.goreleaser.yaml` only shapes the GitHub Release notes body — the two are intentionally not merged into one mechanism.
- Three separate workflow files fire on different triggers: `ci.yml` (push to main / PR), `changelog.yml` and `release.yml` (both on `v*` tag push, independent jobs — `changelog.yml`'s commit lands on `main` *after* the tag, so the tagged tree itself won't contain that changelog update).

## Editor config

`.editorconfig` is the portable baseline all editors pick up (tabs for `.go`, 2-space for json/yaml/md). `.vscode/` and `.zed/` carry editor-specific gopls/golangci-lint wiring on top of that; there's intentionally no LazyVim-specific file since LazyVim has no project-local config mechanism that doesn't require explicit user opt-in.
