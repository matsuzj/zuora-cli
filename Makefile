BINARY_NAME := zr
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w \
	-X github.com/matsuzj/zuora-cli/internal/build.Version=$(VERSION) \
	-X github.com/matsuzj/zuora-cli/internal/build.Commit=$(COMMIT) \
	-X github.com/matsuzj/zuora-cli/internal/build.Date=$(DATE)

.PHONY: build test e2e lint vuln cover clean fmt check

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

check: lint cover
