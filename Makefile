
all: migrate test build

migrate:
	migrate -path internal/migrations -database "postgres://root:root@localhost:5432/taskmanager?sslmode=disable" up

.PHONY: build
build:
	go build -v ./cmd/app/

.PHONY: test
test:
	go test -v -race ./...
