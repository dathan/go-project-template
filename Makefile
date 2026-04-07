# ─────────────────────────────────────────────────────────────────────────────
# Go Project Template — Makefile
# Usage: make setup              (first time — auto-detects name from directory)
#        make dev                (build + run Go server + Vite, Ctrl-C to stop)
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
.PHONY: all lint build build-linux run dev dev-stop dev-restart run-server run-tui test test-unit \
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

# dev: build the Go server, start it in the background, then start the Vite
# dev server in the foreground. Ctrl-C kills both via the EXIT trap.
# The Go server is ready when it logs "server starting"; we poll /healthz
# so the frontend doesn't proxy to a port that isn't listening yet.
dev-start: build
	@echo "==> Starting Postgres, Go server + Vite dev server (Ctrl-C to stop all)"
	@docker rm -f dev-postgres 2>/dev/null || true
	@docker run -d --name dev-postgres \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		-p 5432:5432 \
		postgres:latest
	@trap 'echo; echo "==> Stopping..."; kill $$(cat /tmp/.dev-server.pid) 2>/dev/null; docker rm -f dev-postgres 2>/dev/null || true; exit 0' INT TERM EXIT; \
	echo "  Waiting for Postgres on :5432..."; \
	for i in $$(seq 1 40); do \
		docker exec dev-postgres pg_isready -q 2>/dev/null && break; \
		sleep 0.5; \
	done; \
	docker exec dev-postgres pg_isready -q 2>/dev/null \
		|| (echo "ERROR: Postgres did not start"; docker rm -f dev-postgres 2>/dev/null; exit 1); \
	./bin/$(SERVER_BINARY) & echo $$! > /tmp/.dev-server.pid; \
	echo "  Waiting for Go server on :8080..."; \
	for i in $$(seq 1 30); do \
		curl -sf http://127.0.0.1:8080/healthz >/dev/null 2>&1 && break; \
		sleep 0.5; \
	done; \
	curl -sf http://127.0.0.1:8080/healthz >/dev/null 2>&1 \
		|| (echo "ERROR: Go server did not start"; kill $$(cat /tmp/.dev-server.pid) 2>/dev/null; docker rm -f dev-postgres 2>/dev/null; exit 1); \
	echo "  Go server ready. Starting Vite at http://127.0.0.1:5173 ..."; \
	cd frontend && npm run dev

# dev-stop: kill the background Go server and tear down the Postgres container.
# Safe to run even if nothing is running.
dev-stop:
	@echo "==> Stopping Go server..."
	@kill $$(cat /tmp/.dev-server.pid 2>/dev/null) 2>/dev/null || true
	@rm -f /tmp/.dev-server.pid
	@pkill -x $(SERVER_BINARY) 2>/dev/null && echo "  Server stopped" || echo "  Server was not running"
	@echo "==> Stopping Postgres..."
	@docker rm -f dev-postgres 2>/dev/null && echo "  Postgres stopped" || echo "  Postgres was not running"

# dev-restart: rebuild the Go server and hot-swap it without touching Postgres
# or the Vite dev server. Run this in a second terminal while `make dev` is
# running in the first.
dev-restart: build
	@echo "==> Restarting Go server..."
	@kill $$(cat /tmp/.dev-server.pid 2>/dev/null) 2>/dev/null || true
	@pkill -x $(SERVER_BINARY) 2>/dev/null || true
	@rm -f /tmp/.dev-server.pid
	@./bin/$(SERVER_BINARY) & echo $$! > /tmp/.dev-server.pid
	@echo "  Waiting for Go server on :8080..."; \
	for i in $$(seq 1 30); do \
		curl -sf http://127.0.0.1:8080/healthz >/dev/null 2>&1 && break; \
		sleep 0.5; \
	done; \
	curl -sf http://127.0.0.1:8080/healthz >/dev/null 2>&1 \
		&& echo "  Go server ready (pid $$(cat /tmp/.dev-server.pid))" \
		|| echo "ERROR: Go server did not start"

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
