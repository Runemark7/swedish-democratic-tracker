COMPOSE = docker compose -f docker-compose.dev.yml
DB_URL   = postgres://riksdagskollen:localdev@localhost:5432/riksdagskollen?sslmode=disable

.PHONY: run stop migrate seed-elections verify-values

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

# Runs the value-verification audit against the local dev DB (postgres on
# host :5432 via --network host) and writes a dated report into the repo's
# docs/superpowers/audits/. The whole repo is mounted at /repo and the binary
# is executed from /repo so the report's relative path resolves to repo root.
verify-values:
	docker run --rm --network host \
		-e DATABASE_URL="$(DB_URL)" \
		-v $(PWD):/repo -w /repo \
		golang:1.26.1 \
		sh -c 'cd backend && go build -buildvcs=false -o /tmp/verifyvalues ./cmd/verifyvalues && cd /repo && /tmp/verifyvalues'
