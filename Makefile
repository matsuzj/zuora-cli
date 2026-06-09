BINARY_NAME := zr
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w \
	-X github.com/matsuzj/zuora-cli/internal/build.Version=$(VERSION) \
	-X github.com/matsuzj/zuora-cli/internal/build.Commit=$(COMMIT) \
	-X github.com/matsuzj/zuora-cli/internal/build.Date=$(DATE)

.PHONY: build test e2e lint vuln cover clean fmt fmtcheck modverify check ci

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/zr/

test:
	go test -race -count=1 -coverprofile=cov.out -covermode=atomic ./...

# Enforce the same coverage floor CI uses (73%).
cover: test
	@total="$$(go tool cover -func=cov.out | awk '/^total:/ {sub(/%/, "", $$3); print $$3}')"; \
	echo "Total coverage: $$total%"; \
	if awk "BEGIN{exit !($$total < 73.0)}"; then \
		echo "FAIL: coverage $$total% is below the 73.0% threshold"; exit 1; \
	fi

# Scan for known vulnerabilities in deps and the stdlib toolchain (matches CI).
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# E2E suites hit a real Zuora tenant — requires `zr auth login` first.
e2e: build
	./tests/run-all.sh

lint:
	go vet ./...
	staticcheck ./...

clean:
	rm -rf bin/

fmt:
	gofmt -w .

# Fail if any Go file is not gofmt-formatted (matches the CI Gofmt step).
fmtcheck:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-formatted (run 'make fmt'):"; \
		echo "$$unformatted"; exit 1; \
	fi

# Verify module dependencies (matches the CI "Verify dependencies" step).
modverify:
	go mod verify

# Quick local pre-commit gate: a SUBSET of `ci` (no mod-verify/vuln/build).
check: fmtcheck lint cover

# Full local mirror of the CI gate (.github/workflows/ci.yml): run before
# pushing to catch everything CI checks. (E2E is a separate manual gate — `make e2e`.)
ci: modverify fmtcheck lint vuln cover build
	@echo "ci: all checks passed (matches .github/workflows/ci.yml)"
