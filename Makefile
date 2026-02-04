# Google Cloud Spot Price History — build and run targets
# Requires: Go 1.21+, git. Optional: golint, gofmt in PATH for lint/fmt.

.PHONY: all build test vet fmt lint run help
.PHONY: collect-pricing-data clean

# Default target: run checks and build both binaries
all: test vet fmt lint build

help:
	@echo "Google Cloud Spot Price History"
	@echo ""
	@echo "Targets:"
	@echo "  all                  Run test, vet, fmt, lint, then build (default)"
	@echo "  build                Build bin/dataprocessing and bin/api"
	@echo "  test                 Run tests for all packages"
	@echo "  vet                  Run go vet on all packages"
	@echo "  fmt                  Check that all Go files are formatted (gofmt -l)"
	@echo "  lint                 Run golint on all packages"
	@echo "  collect-pricing-data Clone pricing repo and extract pricing.yml history to /tmp/pricing-data"
	@echo "  run                  Build, collect data, then run dataprocessing (DB: /tmp/history.sqlite3)"
	@echo "  clean                Remove bin/ and cloned pricing repo"
	@echo ""
	@echo "Examples:"
	@echo "  make                 # validate and build"
	@echo "  make build           # build only"
	@echo "  make run             # full pipeline: build, collect data, import into SQLite"
	@echo "  make collect-pricing-data && ./bin/dataprocessing -data /tmp/pricing-data -dbpath ./history.sqlite3"

# Run all package tests
test:
	go test ./...

# Run static analysis
vet:
	go vet ./...

# Fail if any Go file is not formatted (CI check)
fmt:
	@test -z "$$(gofmt -l .)" || (echo "Run: gofmt -w ."; gofmt -l .; exit 1)

# Run golint on all packages (requires golint: go install golang.org/x/lint/golint@latest)
lint:
	go list ./... | xargs -L1 golint -set_exit_status

# Build both binaries into bin/
build:
	@mkdir -p bin
	go build -ldflags="-w -s" -o bin/dataprocessing ./cmd/dataprocessing
	go build -ldflags="-w -s" -o bin/api ./cmd/api

# Clone pricing repo (if missing) or pull; then extract every revision of pricing.yml into /tmp/pricing-data.
collect-pricing-data:
	./scripts/extract_git_history.sh pricing.yml google-cloud-pricing-cost-calculator

# Full pipeline: build, collect pricing history, import into SQLite at /tmp/history.sqlite3
run: build collect-pricing-data
	./bin/dataprocessing -data /tmp/pricing-data -dbpath /tmp/history.sqlite3

# Remove build artifacts and cloned repo
clean:
	rm -rf bin google-cloud-pricing-cost-calculator
