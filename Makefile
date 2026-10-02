.PHONY: dev test build migrate
dev:
	docker compose up --build
test:
	go test ./...
build:
	go build ./...
migrate:
	docker compose exec -T postgres psql -U flowpilot -d flowpilot -f /docker-entrypoint-initdb.d/001_initial.sql
