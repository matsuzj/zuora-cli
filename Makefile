BINARY_NAME := zr
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w \
	-X github.com/matsuzj/zuora-cli/internal/build.Version=$(VERSION) \
	-X github.com/matsuzj/zuora-cli/internal/build.Commit=$(COMMIT) \
	-X github.com/matsuzj/zuora-cli/internal/build.Date=$(DATE)

.PHONY: build test lint clean fmt check

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/zr/

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	staticcheck ./...

clean:
	rm -rf bin/

fmt:
	gofmt -w .

check: lint test

# --- AI Orchestration ---
.PHONY: ai ai-plan ai-impl ai-review ai-test ai-pr ai-quick-review ai-auth ai-status

ISSUE ?=
_ISSUE_ARG := $(if $(ISSUE),--issue $(ISSUE))

ai:
	./ai-orchestrator.sh $(_ISSUE_ARG) --stage all

ai-plan:
	./ai-orchestrator.sh $(_ISSUE_ARG) --stage plan

ai-impl:
	./ai-orchestrator.sh $(_ISSUE_ARG) --stage implement

ai-review:
	./ai-orchestrator.sh $(_ISSUE_ARG) --stage review

ai-test:
	./ai-orchestrator.sh $(_ISSUE_ARG) --stage test

ai-pr:
	./ai-orchestrator.sh $(_ISSUE_ARG) --stage pr

ai-quick-review:
	./scripts/ai-cross-review.sh

ai-auth:
	@echo "=== ANTHROPIC_API_KEY ===" && \
	if [ -n "$${ANTHROPIC_API_KEY:-}" ]; then echo "⚠️  設定済み（API課金優先）"; else echo "✅ 未設定（サブスク優先）"; fi
	@echo "=== Claude Code ===" && if command -v claude >/dev/null 2>&1; then claude auth status 2>&1 || echo "⚠️  未認証"; else echo "未インストール"; fi
	@echo "=== Codex CLI ===" && if command -v codex >/dev/null 2>&1; then codex login status 2>&1 || echo "⚠️  未認証"; else echo "未インストール"; fi

ai-status:
	@echo "Branch: $$(git rev-parse --abbrev-ref HEAD)"
	@echo "Worktrees:"; git worktree list
	@echo "最新ログ:"; log_dir=$$(ls -td logs/ai-orchestrator/*/ 2>/dev/null | head -1); \
	if [ -n "$${log_dir}" ] && [ -f "$${log_dir}run.log" ]; then cat "$${log_dir}run.log"; else echo "ログなし"; fi
