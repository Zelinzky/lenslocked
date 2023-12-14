# makefile to build the project, run locally and run the tests

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
	@podman compose --file compose.yaml up --detach
	@sleep 1

db-down:
	@podman compose --file compose.yaml down

db-migrate-up:
	@goose -dir ./migrations postgres "host=localhost port=5432 user=baloo password=junglebook dbname=lenslocked sslmode=disable" up

db-migrate-down:
	@goose -dir ./migrations postgres "host=localhost port=5432 user=baloo password=junglebook dbname=lenslocked sslmode=disable" down

db-migrate-reset:
	@goose -dir ./migrations postgres "host=localhost port=5432 user=baloo password=junglebook dbname=lenslocked sslmode=disable" reset

db-migrate-status:
	@goose -dir ./migrations postgres "host=localhost port=5432 user=baloo password=junglebook dbname=lenslocked sslmode=disable" status

db-migrate-fix:
	@goose -dir ./migrations fix