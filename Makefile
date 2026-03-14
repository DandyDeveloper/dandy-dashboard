.PHONY: dev dev-backend dev-frontend build build-backend build-frontend docker-build docker-up docker-down test lint clean install-deps

# ── Development ────────────────────────────────────────────────────────────────

dev:
	@echo "Start dev-backend and dev-frontend in separate terminals."
	@echo "  make dev-backend"
	@echo "  make dev-frontend"

dev-backend:
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air -c .air.toml

dev-frontend:
	cd web && npm run dev

# ── Local binary builds ────────────────────────────────────────────────────────

build: build-backend build-frontend

build-backend:
	go build -ldflags="-s -w" -o bin/server ./cmd/server

build-frontend:
	cd web && npm install && npm run build

# ── Docker ─────────────────────────────────────────────────────────────────────

docker-build:
	docker compose build

# Start without Redis (embedded bolt store)
docker-up:
	docker compose up --build

# Start with external Redis store (set STORE_URL=redis://redis:6379 in .env)
docker-up-redis:
	docker compose --profile redis up --build

docker-down:
	docker compose down

# ── Test & Lint ────────────────────────────────────────────────────────────────

test:
	go test ./...

lint:
	@which golangci-lint > /dev/null || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	cd web && npm run check

# ── Utilities ──────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ web/dist/

install-deps:
	cd web && npm install
