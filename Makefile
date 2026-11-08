.PHONY: dev server-mock server web test live-test gen build

dev: ## mock API (synthetic fixture) + Vite, no PowerSchool calls
	$(MAKE) -j2 server-mock web

server-mock:
	cd server && MOCK=1 go run ./cmd/server

server: ## real school; reads SCHOOL_BASE_URL (and optional settings) from .env
	cd server && set -a && . ../.env && set +a && go run ./cmd/server

web:
	cd web && pnpm dev

test:
	cd server && go test ./...
	cd web && pnpm test

live-test: ## hits the real school with the .env account; prints counts only
	cd server && go test -tags live -run Live -v -count=1 ./internal/powerschool/

gen: ## regenerate web/src/api/*.gen.ts from the Go DTOs
	cd server && go tool tygo generate

build:
	cd server && go build -o ../bin/server ./cmd/server
	cd web && pnpm build
