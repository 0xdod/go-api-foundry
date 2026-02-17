.PHONY: run run-with-migrate create-migration migrate generate-domain build tidy docker-build docker-run dev dev-migrate vendor	

-include .env

DB_URL?=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST-}:${POSTGRES_PORT}/${POSTGRES_DB_NAME}?sslmode=${POSTGRES_SSLMODE}
DB_MIGRATIONS_DIR?=migrations


run:
	go run ./cmd/server

run-with-migrate:
	go run ./cmd/server --auto-migrate

dev:
	$(MAKE) sqlc
	$(MAKE) swag
	air

dev-migrate:
	$(MAKE) migrate-up
	$(MAKE) dev

migrate:
	go run ./cmd/cli migrate

install-migrate:
	@if ! command -v migrate >/dev/null 2>&1; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi

install-sqlc:
	@if ! command -v sqlc >/dev/null 2>&1; then \
		echo "Installing sqlc..."; \
		go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
	fi

create-migration: install-migrate
	@if [ -z "$(name)" ]; then \
		read -p "Enter migration name: " name; \
	fi; \
	migrate create -ext sql -dir migrations -seq $${name}

migrate-up: install-migrate ## Migrate the database schema up to the latest version
	migrate -database $(DB_URL) -path $(DB_MIGRATIONS_DIR) up

migrate-down: install-migrate ## Migrate the database schema down to the latest version
	migrate -database $(DB_URL) -path $(DB_MIGRATIONS_DIR) down

migrate-force: install-migrate ## Force the database schema to a specific version
	@if [ -z "$(version)" ]; then \
		read -p "Enter migration version to force: " version; \
	fi; \
	migrate -database $(DB_URL) -path $(DB_MIGRATIONS_DIR) force $(version)

.PHONY: sqlc
sqlc:
	sqlc generate

install-swag:
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi

.PHONY: swag
swag: install-swag
	swag init -g cmd/server/main.go --parseInternal --parseDependency --parseDepth 2


.PHONY: mockgen-install
mockgen-install:
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "Installing mockgen..."; \
		go install go.uber.org/mock/mockgen@latest; \
	fi

.PHONY: mocks
mock: mockgen-install
	mockgen -source=domain/account/repository.go -destination=domain/account/mock_repository.go -package=account
	mockgen -source=domain/transaction/repository.go -destination=domain/transaction/mock_repository.go -package=transaction
	@echo "Mocks generated successfully!"

generate-domain:
	go run ./cmd/cli generate-domain

format:
	go fmt ./...

lint:
	go vet ./...

vendor:
	go mod vendor

test:
	go test ./...

.PHONY: integration-test
integration-test:
	RUN_INTEGRATION_TESTS=true go test -v -count=1 ./integration/...

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

docker-build:
	docker build -t go-api-foundry:dev .

docker-run:
	docker run --rm --env-file .env -p $${APP_PORT:-8080}:$${APP_PORT:-8080} go-api-foundry:dev
