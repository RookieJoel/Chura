.PHONY: build lint test generate generate-proto migrate-up migrate-down

-include .env
export DATABASE_URL

build:
	go build -o bin/project ./cmd/project
	go build -o bin/sprint ./cmd/sprint
	go build -o bin/analytics ./cmd/analytics
	go build -o bin/workitem ./cmd/workitem

lint:
	go vet ./...

test:
	go test -race ./...

generate:
	sqlc generate

generate-proto:
	protoc --go_out=. --go_opt=module=github.com/RookieJoel/Chura/backend --go-grpc_out=. --go-grpc_opt=module=github.com/RookieJoel/Chura/backend internal/adapter/handler/grpc/proto/workitem.proto

migrate-up:
	goose -dir internal/adapter/db/postgres/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir internal/adapter/db/postgres/migrations postgres "$(DATABASE_URL)" down
