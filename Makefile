
all: migrate test build run

migrate:
	migrate -path internal/migrations -database "postgres://root:root@localhost:5432/taskmanager?sslmode=disable" up

.PHONY: test
test:
	go test -v -race ./...

.PHONY: build
build:
	go build -v ./cmd/app/ -o taskmanager

.PHONY: run
run: 
	./taskmanager 