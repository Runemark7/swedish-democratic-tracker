COMPOSE = docker compose -f docker-compose.dev.yml
DB_URL   = postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable

.PHONY: run stop migrate

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
