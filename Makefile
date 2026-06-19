-include .env
export

.PHONY: run build test migrate-up migrate-down lint tidy docker-up docker-down frontend-dev

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -v -cover

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

frontend-dev:
	cd frontend && npm run dev
