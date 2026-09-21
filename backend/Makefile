.PHONY: build lint test generate migrate-up migrate-down

build:
	go build -o bin/api ./cmd/api

lint:
	go vet ./...

test:
	go test -race ./...

generate:
	sqlc generate

migrate-up:
	goose -dir db/migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir db/migrations postgres "$$DATABASE_URL" down
