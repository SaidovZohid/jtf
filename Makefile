.SILENT:
username=$(shell whoami)

cache:
	@go clean -testcache

build:
	@go build -o bin/main ./cmd/main.go

run: build
	./bin/main

tidy:
	@go mod tidy
	@go mod vendor

docker-build:
	docker build --platform linux/amd64 -t zohiddev/jtf:latest .

docker-push: docker-build
	docker push zohiddev/jtf:latest

up: build
	@docker compose --env-file ./.env.docker up -d 

down:
	@docker compose down

ssh:
	@ssh-keygen -f "/Users/"${username}"/.ssh/known_hosts" -R "[localhost]:2222"


.PHONY: run cache build 