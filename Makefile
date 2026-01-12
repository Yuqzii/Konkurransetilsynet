PROFILE ?= dev
DETACHED ?= # Set to -d to start as detached (used when deploying).
PULL ?= # Set to --pull to pull images instead of using cache.

default: build run

run:
	docker-compose --profile $(PROFILE) up $(DETACHED)

build:
	docker compose --profile $(PROFILE) build $(PULL)

migrate-down:
	docker compose --profile $(PROFILE) run --rm dev-migrate-down $(COUNT)

db-shell:
	docker-compose exec db psql -U postgres bot_data

dev-db-shell:
	docker-compose exec dev-db psql -U postgres bot_data
