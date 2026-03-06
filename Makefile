include .env
export

.PHONY: db
db:
	docker compose up -d postgres

.PHONY: db-stop
db-stop:
	docker compose down

.PHONY: migrate
migrate:
	cd services/api && goose -dir db/migrations postgres "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	cd services/api && goose -dir db/migrations postgres "$(DATABASE_URL)" down

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: api
api:
	cd services/api && go run ./cmd/server/main.go

.PHONY: frontend
frontend:
	cd apps/frontend && PUBLIC_API_URL=http://localhost:$(PORT) pnpm dev

.PHONY: dev
dev:
	docker compose up

.PHONY: install
install:
	cd services/api && go mod download
	cd apps/frontend && pnpm install

.PHONY: build-api
build-api:
	cd services/api && go build -o bin/retrosnack-api ./cmd/server/main.go

.PHONY: build-frontend
build-frontend:
	cd apps/frontend && PUBLIC_API_URL=http://localhost:$(PORT) pnpm build

.PHONY: test
test:
	cd services/api && go test ./...

.PHONY: typecheck
typecheck: 
	cd apps/frontend && pnpm check

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
