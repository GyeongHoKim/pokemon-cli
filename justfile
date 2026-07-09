default:
    just --list

build:
    go build -o bin/pokemon-cli ./cmd/pokemon-cli

run:
    go run ./cmd/pokemon-cli

test:
    go test ./...

fmt:
    golangci-lint fmt

lint:
    golangci-lint run

lint-fix:
    golangci-lint run --fix

tidy:
    go mod tidy

generate-sprites:
    go run ./cmd/sprite-gen

sprite-preview POKEMON="pikachu":
    go run ./cmd/sprite-gen -preview {{POKEMON}}

ci: fmt lint test

changelog:
    git-cliff --output CHANGELOG.md

release-dry:
    goreleaser release --snapshot --clean --skip=publish
