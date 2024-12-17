# Include variables from the .envrc file
include .env

# =============================================================================== #
# HELPERS
# =============================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && \
	if [ -z "$$ans" ]; then \
		echo "You haven't entered anything. Please try again."; \
		false; \
	elif [ "$${ans:-N}" = "y" ]; then \
		echo "Proceeding..."; \
	else \
		echo "No? Thanks, bye!"; \
		false; \
	fi

# ============================================================================== #
# DEVELOPMENT
# ============================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo "creating migration files for ${name}..."
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "running up migratons..."
	migrate -path ./migrations -database ${DB_DSN} up


# ============================================================================== #
# QUALITY CONTROL
# ============================================================================== # 

## audit: tidy dependancies and format,vet and test all codes

.PHONY: audit
audit: vendor
	@echo 'Tidying and verfying module dependancies...'
	go mod tidy
	go mod verify

	@echo 'Formatting code...'
	go fmt ./...

	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...

	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verfying module dependancies...'
	go mod tidy
	go mod verify

	@echo 'Vendering dependancies'
	go mod vendor