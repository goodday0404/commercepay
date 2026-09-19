DB_USER := $(POSTGRES_USER)
DB_PASSWORD := $(POSTGRES_PASSWORD)
DB := $(POSTGRES_DB)
DB_TEST := $(POSTGRES_DB_TEST)
TEST_API := $(TEST_DATABASE_URL)

MIGRATIONS_DIR := ./migrations

GOOSE_TEST := \
	GOOSE_DRIVER=postgres \
	GOOSE_DBSTRING="$(TEST_DATABASE_URL)" \
	GOOSE_MIGRATION_DIR="$(MIGRATIONS_DIR)" \
	go tool goose

.PHONY: \
	test-db-status \
	test-db-version \
	test-db-up \
	test-db-up-one \
	test-db-down \
	test-db-redo \
	test-db-reset \
	test-db-validate \
	check-test-db-url \
	exec-test-db \
	test-api-run

exec-test-db:
	docker compose exec postgres \
	psql -U $(DB_USER) -d $(DB_TEST)

check-test-db-url:
	@test -n "$(TEST_DATABASE_URL)" || \
		(echo "ERROR: TEST_DATABASE_URL is not set"; exit 1)

test-db-status: check-test-db-url
	$(GOOSE_TEST) status

test-db-version: check-test-db-url
	$(GOOSE_TEST) version

test-db-up: check-test-db-url
	$(GOOSE_TEST) up

test-db-up-one: check-test-db-url
	$(GOOSE_TEST) up-by-one

test-db-down: check-test-db-url
	$(GOOSE_TEST) down

test-db-redo: check-test-db-url
	$(GOOSE_TEST) redo

test-db-reset: check-test-db-url
	$(GOOSE_TEST) reset

test-db-validate:
	go tool goose -dir "$(MIGRATIONS_DIR)" validate

test-api-run: 
	DATABASE_URL="$(TEST_API)" \
	go run ./cmd/api