default:
    just --list

build:
    go build -o bin/pokemon-cli .

run:
    go run .

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

ci: fmt lint test

changelog:
    git-cliff --output CHANGELOG.md

release-dry:
    goreleaser release --snapshot --clean --skip=publish
