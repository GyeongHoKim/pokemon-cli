# pokemon-cli

A Pokémon companion for your terminal — pick any of the 1,025 species and watch it sit, react, and talk, in a TUI built with [bubbletea](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss).

> Status: the sprite pipeline (`internal/sprite`) is implemented — every species resolves to real terminal-renderable art. The bubbletea UI/state machine itself isn't wired up yet (`main.go` is still a placeholder).

## Install

```bash
# npm
npm install -g @gyeonghokim/pokemon-cli

# Homebrew
brew install GyeongHoKim/tap/pokemon-cli

# Scoop
scoop bucket add GyeongHoKim https://github.com/GyeongHoKim/scoop-bucket
scoop install pokemon-cli
```

## Usage

```bash
pokemon-cli
```

Resize your terminal — Pikachu reacts. Requires a truecolor-capable terminal for the best visuals (most modern terminal emulators support this out of the box).

## Development

Toolchain versions are pinned via [mise](https://mise.jdx.dev):

```bash
mise install
```

Common tasks (see `justfile`):

```bash
just          # list all tasks
just run      # go run .
just build    # build a local binary into bin/
just test     # go test ./...
just fmt      # golangci-lint fmt
just lint     # golangci-lint run
just ci       # fmt + lint + test, same as CI
just release-dry  # local goreleaser snapshot build, no publishing

just generate-sprites     # regenerate internal/sprite's embedded dataset from PokeAPI/sprites
just sprite-preview eevee # print a species' rendered sprite art to the terminal
```

### Commit messages

Commits must follow [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `chore:`, ...) — enforced locally by a husky `commit-msg` hook and re-checked in CI on pull requests. Run `npm install` once after cloning so the hook is installed.

### Release process

Pushing a `v*` tag triggers two GitHub Actions workflows:

- `changelog.yml` regenerates `CHANGELOG.md` from commit history via [git-cliff](https://git-cliff.org) and commits it to `main`.
- `release.yml` builds and publishes binaries via [goreleaser](https://goreleaser.com) (GitHub Release, Homebrew tap, Scoop bucket) and publishes the npm packages via OIDC trusted publishing (no stored npm token).

## Legal

Pokémon and Pokémon character names are trademarks of Nintendo. This is an unaffiliated fan project — see [NOTICE](NOTICE) for sprite asset attribution.

## License

MIT
