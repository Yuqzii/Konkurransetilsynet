PROFILE ?= dev

default: build run

run:
	docker-compose --profile $(PROFILE) up

build:
	docker compose --profile $(PROFILE) build

migrate-down:
	docker compose --profile $(PROFILE) run --rm dev-migrate-down $(COUNT)

db-shell:
	docker-compose exec db psql -U postgres bot_data

dev-db-shell:
	docker-compose exec dev-db psql -U postgres bot_data
