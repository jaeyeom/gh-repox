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

.PHONY: build clean install release-check

## build: compile the CLI binary (local development only)
build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/gh-repox

## install: install the binary into GOPATH/bin
install:
	go install $(GOFLAGS) ./cmd/gh-repox

## release-check: cross-compile for all release platforms to verify release readiness
release-check:
	@echo "Verifying cross-compilation for all release platforms..."
	@for platform in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="dist/gh-repox-$${os}-$${arch}"; \
		if [ "$$os" = "windows" ]; then output="$${output}.exe"; fi; \
		echo "  Building $${os}/$${arch}..."; \
		GOOS=$$os GOARCH=$$arch go build $(GOFLAGS) -o "$$output" ./cmd/gh-repox || exit 1; \
	done
	@echo "All platforms built successfully in dist/"

## clean: remove build artifacts and coverage files
clean:
	rm -rf $(BINARY) dist coverage.out coverage.html

# ── Format ──────────────────────────────────────────────────────────

.PHONY: format check-format

## format: auto-format all Go source files
format:
	gofmt -w .

## check-format: verify formatting (fails on diff)
check-format:
	@test -z "$$(gofmt -l .)" || { echo "gofmt: files need formatting:"; gofmt -l .; exit 1; }

# ── Lint / Fix ──────────────────────────────────────────────────────

.PHONY: lint fix vet lint-golangci fix-golangci

## lint: run go vet and golangci-lint
lint: vet lint-golangci

## fix: format, vet, and golangci-lint auto-fix
fix: format vet fix-golangci

## vet: run go vet on all packages
vet:
	go vet ./...

## lint-golangci: run golangci-lint
lint-golangci:
	golangci-lint run ./...

## fix-golangci: run golangci-lint with auto-fix
fix-golangci:
	golangci-lint run --fix ./...

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
