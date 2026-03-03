.PHONY: dev dev-backend dev-frontend docker-up docker-down docker-prod-up docker-prod-down build test lint swagger wire clean

# Development
dev: docker-up dev-backend dev-frontend

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && pnpm dev

# Docker (development infrastructure only)
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v

# Docker (production: full stack with backend services)
docker-prod-up:
	docker-compose -f docker-compose.prod.yml up -d --build

docker-prod-down:
	docker-compose -f docker-compose.prod.yml down

docker-prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

# Build
build-backend:
	cd backend && go build -o bin/server ./cmd/server

build-frontend:
	cd frontend && pnpm build

build: build-backend build-frontend

# Test
test-backend:
	cd backend && go test ./... -v -cover

test-frontend:
	cd frontend && pnpm test

test: test-backend test-frontend

# Lint
lint-backend:
	cd backend && golangci-lint run ./...

lint-frontend:
	cd frontend && pnpm lint

lint: lint-backend lint-frontend

# Code generation
swagger:
	cd backend && swag init -g cmd/server/main.go -o docs

wire:
	cd backend && wire ./...

# Clean
clean:
	rm -rf backend/bin frontend/dist
