COMPOSE = docker compose -f docker-compose.dev.yml
DB_URL   = postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable

.PHONY: run stop migrate seed-elections

run:
	$(COMPOSE) down --remove-orphans
	$(COMPOSE) up --build

migrate:
	docker run --rm --network host \
		-v $(PWD)/backend/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database "$(DB_URL)" \
		up

seed-elections:
	docker run --rm --network host \
		-e DATABASE_URL="$(DB_URL)" \
		-v $(PWD)/backend:/app -w /app \
		golang:1.26.1 go run ./cmd/seed-elections
