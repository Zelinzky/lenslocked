# makefile to build the project, run locally and run the tests

default: build

build:
	@echo "Building the project"
	@go build -o bin/ ./...

run:
	@echo "Running the project"
	@air