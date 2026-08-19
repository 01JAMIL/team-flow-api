.PHONY: run build docker-up docker-down

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/api ./cmd/api

docker-up:
	docker compose up -d

docker-build:
	docker compose up --build -d

docker-down:
	docker compose down