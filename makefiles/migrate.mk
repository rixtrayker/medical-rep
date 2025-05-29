.PHONY: migrate-up migrate-down migrate-create migrate-drop migrate-reset

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-drop:
	$(MIGRATE) drop -f

migrate-reset:
	$(MAKE) migrate-drop
	$(MAKE) migrate-up
	$(MAKE) db-seed

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=init_users"; \
		exit 1; \
	fi
	migrate create -ext sql -dir db/migrations -seq $(name)

db-seed:
	psql $(DB_URL) -f db/seed.sql
