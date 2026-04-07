# ─────────────────────────────────────────────────────────────────────────────
# Go Project Template — Makefile
# Usage: make setup              (first time — auto-detects name from directory)
#        make build / test / run / docker-build
# ─────────────────────────────────────────────────────────────────────────────

# Auto-detect project identity from the working directory and git remote.
# When this template is used via "Use this template" on GitHub, the new repo
# will be cloned under its own name — PROJECT_NAME and GIT_OWNER pick that up
# automatically without any manual override.
PROJECT_NAME  := $(notdir $(CURDIR))
GIT_OWNER     := $(shell git remote get-url origin 2>/dev/null \
                   | sed -E 's|.*[:/]([^/]+)/[^/.]+.*|\1|' \
                   || echo "unknown")

BINARY_NAME   ?= $(PROJECT_NAME)
SERVER_BINARY  = server
TUI_BINARY     = tui
AGENT_BINARY   = agent
# ghcr.io/<owner>/<repo> — no double-segment; owner comes from the git remote.
REPO          ?= ghcr.io/$(GIT_OWNER)/$(PROJECT_NAME)
GIT_SHA       := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

.DEFAULT_GOAL := all
.PHONY: all lint build build-linux run run-server run-tui test test-unit \
        test-integration migrate-up migrate-down migrate-create \
        frontend-install frontend-build frontend-dev \
        docker-build docker-tag docker-push docker-clean \
        playwright-install setup clean vendor

# ── Default ────────────────────────────────────────────────────────────────────
all: lint test build

# ── Setup (idempotent bootstrap) ───────────────────────────────────────────────
# Usage: make setup
# Auto-detects PROJECT_NAME from the directory name and GIT_OWNER from the
# git remote. When this repo is created from the GitHub template feature the
# directory will already have the correct name — no manual override needed.
setup:
	@echo "==> Setting up project…"
	@# Copy .env if missing
	@[ -f .env ] || (cp .env.example .env && echo "  Created .env from .env.example — fill in secrets before running")
	@[ -f frontend/.env ] || (cp frontend/.env.example frontend/.env && echo "  Created frontend/.env")
	@# ── Rename go-project-template → PROJECT_NAME (idempotent) ──────────────
	@# Read current module owner+name from go.mod, e.g. "github.com/dathan/go-project-template"
	@CURR_MODULE=$$(grep '^module ' go.mod | awk '{print $$2}'); \
	CURR_OWNER=$$(echo "$$CURR_MODULE" | sed -E 's|github.com/([^/]+)/.*|\1|'); \
	CURR_NAME=$$(echo "$$CURR_MODULE" | sed -E 's|.*/([^/]+)$$|\1|'); \
	NEW_NAME=$(PROJECT_NAME); \
	NEW_OWNER=$(GIT_OWNER); \
	if [ "$$CURR_NAME" != "$$NEW_NAME" ] || [ "$$CURR_OWNER" != "$$NEW_OWNER" ]; then \
		echo "  Renaming: github.com/$$CURR_OWNER/$$CURR_NAME → github.com/$$NEW_OWNER/$$NEW_NAME"; \
		SED_I="sed -i ''"; \
		sed -i '' 's/x/' /dev/null 2>/dev/null || SED_I="sed -i"; \
		find . -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' \
			-type f \( -name '*.go' -o -name '*.mod' \) \
			-exec $$SED_I "s|github.com/$$CURR_OWNER/$$CURR_NAME|github.com/$$NEW_OWNER/$$NEW_NAME|g" {} +; \
		find . -not -path './.git/*' -not -path './vendor/*' -not -path './node_modules/*' \
			-type f \( -name '*.md' -o -name '*.yaml' -o -name '*.yml' \
			           -o -name '*.sh' -o -name 'Makefile' -o -name 'Dockerfile' -o -name '*.toml' \) \
			-exec $$SED_I "s/$$CURR_NAME/$$NEW_NAME/g" {} +; \
		[ -d "cmd/$$CURR_NAME" ] && mv "cmd/$$CURR_NAME" "cmd/$$NEW_NAME" || true; \
		echo "  Done. Run: go mod tidy"; \
	else \
		echo "  Nothing to rename (already $$CURR_MODULE)"; \
	fi
	@# Install Go tools (idempotent — only installs if missing)
	@which golangci-lint >/dev/null 2>&1 || (echo "  Installing golangci-lint…" && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@which migrate >/dev/null 2>&1 || (echo "  Installing golang-migrate…" && \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	@which ginkgo >/dev/null 2>&1 || (echo "  Installing ginkgo…" && \
		go install github.com/onsi/ginkgo/v2/ginkgo@latest)
	@# Install frontend dependencies
	@[ -d frontend/node_modules ] || $(MAKE) frontend-install
	@echo "==> Setup complete. Edit .env then run: make run-server"

# ── Go Build ──────────────────────────────────────────────────────────────────
build:
	go build -o ./bin ./cmd/...

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" \
		-o bin/$(SERVER_BINARY) ./cmd/server
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" \
		-o bin/$(TUI_BINARY) ./cmd/tui
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" \
		-o bin/$(AGENT_BINARY) ./cmd/agent

# ── Run ───────────────────────────────────────────────────────────────────────
run: run-server

run-server: build
	./bin/$(SERVER_BINARY)

run-tui: build
	./bin/$(TUI_BINARY)

# ── Testing ───────────────────────────────────────────────────────────────────
test: test-unit

test-unit:
	go test -v -race ./... 2>&1 | tee test/coverage.out

test-integration:
	@echo "==> Integration tests (requires Postgres via testcontainers)"
	go test -v -tags integration ./test/integration/...

test-e2e:
	@echo "==> E2E tests (requires server running on :8080)"
	cd frontend && npx playwright test

# ── Lint ──────────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ── Database Migrations ───────────────────────────────────────────────────────
DB_URL ?= postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSLMODE)

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create:
	@[ -n "$(NAME)" ] || (echo "Usage: make migrate-create NAME=<migration_name>" && exit 1)
	migrate create -ext sql -dir ./migrations -seq $(NAME)

# ── Frontend ──────────────────────────────────────────────────────────────────
frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

frontend-dev:
	cd frontend && npm run dev

# ── Playwright ────────────────────────────────────────────────────────────────
playwright-install:
	go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps

# ── Docker ────────────────────────────────────────────────────────────────────
docker-build: frontend-build
	docker build -t $(BINARY_NAME)-release .

docker-tag:
	docker tag $$(docker image ls --filter 'reference=$(BINARY_NAME)-release' -q) \
		$(REPO):$(GIT_SHA)

docker-push: docker-build docker-tag
	docker push $(REPO):$(GIT_SHA)

docker-clean:
	docker rmi $$(docker image ls --filter 'reference=$(BINARY_NAME)-*' -q) 2>/dev/null || true

# ── Misc ──────────────────────────────────────────────────────────────────────
vendor:
	go mod vendor

clean:
	go clean
	rm -rf bin/ frontend/dist
	find . -type d -name '.tmp_*' -prune -exec rm -rvf {} \;
