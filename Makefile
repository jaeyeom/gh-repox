# gh-repox Makefile
# Precompiled Go GitHub CLI extension for repository policy management.

BINARY   := gh-repox
MODULE   := github.com/jaeyeom/gh-repox
GOFLAGS  ?=

# ── Aggregate targets ───────────────────────────────────────────────

.PHONY: all check

## all: full local workflow (format, lint-fix, test, build)
all: format fix test build

## check: CI-safe checks (no mutation)
check: check-format lint test build

# ── Build ───────────────────────────────────────────────────────────

.PHONY: build clean install

## build: compile the CLI binary
build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/gh-repox

## install: install the binary into GOPATH/bin
install:
	go install $(GOFLAGS) ./cmd/gh-repox

## clean: remove build artifacts and coverage files
clean:
	rm -f $(BINARY) coverage.out coverage.html

# ── Format ──────────────────────────────────────────────────────────

.PHONY: format check-format

## format: auto-format all Go source files
format:
	gofmt -w .

## check-format: verify formatting (fails on diff)
check-format:
	@test -z "$$(gofmt -l .)" || { echo "gofmt: files need formatting:"; gofmt -l .; exit 1; }

# ── Lint / Fix ──────────────────────────────────────────────────────

.PHONY: lint fix vet

## lint: run go vet (add golangci-lint when .golangci.yml exists)
lint: vet

## fix: format then vet (ordering ensures no concurrent writes)
fix: format vet

## vet: run go vet on all packages
vet:
	go vet ./...

# ── Test ────────────────────────────────────────────────────────────

.PHONY: test coverage coverage-html coverage-report clean-coverage

## test: run all unit tests
test:
	go test ./...

## coverage: generate coverage profile
coverage:
	go test -coverprofile=coverage.out ./...

## coverage-html: open HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

## coverage-report: print coverage summary to stdout
coverage-report: coverage
	go tool cover -func=coverage.out

## clean-coverage: remove coverage artifacts
clean-coverage:
	rm -f coverage.out coverage.html

# ── Module maintenance ──────────────────────────────────────────────

.PHONY: tidy verify

## tidy: tidy and verify go.mod / go.sum
tidy:
	go mod tidy

## verify: verify module checksums
verify:
	go mod verify
