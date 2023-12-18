# makefile to build the project, run locally and run the tests

postgres_connection_string = "host=localhost port=5432 user=baloo password=junglebook dbname=lenslocked sslmode=disable"

default: build

build:
	@echo "Building the project"
	@go build -o bin/ ./...

run: db-up
	@echo "Running the project"
	@air

tidy:
	@echo "Tidying the project"
	@go mod tidy

clean: db-down

db-up:
	@docker compose --file compose.yaml up --detach
	@sleep 1

db-down:
	@docker compose --file compose.yaml down

db-migrate-up:
	@goose -dir ./migrations postgres $(postgres_connection_string) up

db-migrate-down:
	@goose -dir ./migrations postgres $(postgres_connection_string) down

db-migrate-reset:
	@goose -dir ./migrations postgres $(postgres_connection_string) reset

db-migrate-status:
	@goose -dir ./migrations postgres $(postgres_connection_string) status

db-migrate-create:
	@goose -dir ./migrations create $(name) sql

db-migrate-fix:
	@goose -dir ./migrations fix