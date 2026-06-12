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

# Register-only command-group parents (pure cobra AddCommand wiring, no
# logic): exempt from the per-package coverage floor. Anything ELSE dropping
# to 0% must fail — keep this list explicit so a package that loses its tests
# cannot silently slip through.
COVER_EXEMPT := pkg/cmd/account pkg/cmd/billrun pkg/cmd/charge pkg/cmd/commitment \
	pkg/cmd/contact pkg/cmd/creditmemo pkg/cmd/debitmemo pkg/cmd/fulfillment \
	pkg/cmd/fulfillment-item pkg/cmd/invoice pkg/cmd/meter pkg/cmd/omnichannel \
	pkg/cmd/order pkg/cmd/order-action pkg/cmd/order-line-item pkg/cmd/payment \
	pkg/cmd/plan pkg/cmd/prepaid pkg/cmd/product pkg/cmd/ramp pkg/cmd/rateplan \
	pkg/cmd/subscription pkg/cmd/usage

# Per-package floor — RATCHET: 50% sits just under today's lowest tested
# package (pkg/cmd/factory, 54.8%); raise it as the lows improve. The total
# floor (73%) alone hid a dozen sub-floor packages behind the average.
COVER_PKG_FLOOR := 50.0

# Enforce the same coverage floors CI uses: total (73%) + per-package ratchet.
cover: test
	@total="$$(go tool cover -func=cov.out | awk '/^total:/ {sub(/%/, "", $$3); print $$3}')"; \
	echo "Total coverage: $$total%"; \
	if awk "BEGIN{exit !($$total < 73.0)}"; then \
		echo "FAIL: coverage $$total% is below the 73.0% threshold"; exit 1; \
	fi
	@awk -v floor="$(COVER_PKG_FLOOR)" -v exempt="$(COVER_EXEMPT)" '\
	BEGIN { n = split(exempt, e, /[ \t]+/); for (i = 1; i <= n; i++) ex["github.com/matsuzj/zuora-cli/" e[i]] = 1 } \
	NR > 1 { \
		colon = index($$1, ":"); file = substr($$1, 1, colon - 1); \
		pkg = file; sub(/\/[^\/]*$$/, "", pkg); \
		stmts[pkg] += $$2; if ($$3 > 0) cov[pkg] += $$2; \
	} \
	END { \
		bad = 0; \
		for (p in stmts) { \
			pct = 100 * cov[p] / stmts[p]; \
			if (pct < floor && !(p in ex)) { printf "FAIL: %s %.1f%% < %.1f%% per-package floor\n", p, pct, floor; bad = 1 } \
		} \
		if (!bad) print "Per-package coverage floor (" floor "%): OK"; \
		exit bad; \
	}' cov.out

# Scan for known vulnerabilities in deps and the stdlib toolchain (matches CI).
# Note: ./... covers code reachable from this module's packages only — go.mod
# `tool` deps (staticcheck) are NOT scanned. Acceptable: they never ship in the
# binary and only run on developer/CI machines.
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# E2E suites hit a real Zuora tenant — requires `zr auth login` first.
# Optionally run a subset by suite name (tests/e2e-<name>.sh): make e2e ARGS="local"
e2e: build
	./tests/run-all.sh $(ARGS)

# staticcheck/deadcode run via go.mod's `tool` directive, so local and CI
# always use the same pinned version (dependabot bumps it) — no separate
# install needed. deadcode -test must stay EMPTY: code reachable from neither
# the binary nor any test is deleted, not kept (P4-3 gate).
lint:
	go vet ./...
	go tool staticcheck ./...
	@dead="$$(go tool deadcode -test ./...)" || { echo "deadcode failed to run"; exit 1; }; \
	if [ -n "$$dead" ]; then \
		echo "deadcode found unreachable code (delete it or wire it):"; \
		echo "$$dead"; exit 1; \
	fi
	@bad="$$(grep -rln 'Examples:' pkg/cmd --include='*.go' --exclude='*_test.go' || true)"; \
	if [ -n "$$bad" ]; then \
		echo "example invocations belong in the cobra Example: field, not embedded in Long (P5-3):"; \
		echo "$$bad"; exit 1; \
	fi

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
