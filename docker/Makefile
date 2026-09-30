define migrate-service
	@docker compose run --rm migrate \
		-path /migrations-$1 \
		-database "postgres://postgres:postgres@postgres-$1:5432/$1?sslmode=disable" \
		$2
endef

migrate-user-service-create: ## PostgreSQL: Создать новую версию схемы данных
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-user-create name=init"; \
		exit 1; \
	fi; \
	docker compose run --rm migrate create -ext sql -dir /migrations-user -seq "$(name)"

migrate-user-service-up: ## PostgreSQL: Накатить миграции
	$(call migrate-service,user,up)

migrate-user-service-down: ## PostgreSQL: Откатить миграции
	$(call migrate-service,user,down 1)