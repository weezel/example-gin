APP		?= my-app
name		?= you-didnt-define-migration-name

GO		?= go
DOCKER		?= docker
DOCKER_BUILDKIT ?= 1
VERSION		?= $(shell git log --pretty=format:%h -n 1)
BUILD_TIME	?= $(shell date)
# -s removes symbol table and -ldflags -w debugging symbols
LDFLAGS		?= -asmflags -trimpath -ldflags \
		   "-s -w -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'"
GOARCH		?= amd64
GOOS		?= linux
# CGO_ENABLED=0 == static by default
CGO_ENABLED	?= 0

PSQL_CLIENT	?= psql
PG_DUMP		?= pg_dump
POSTGRES_VER	?= 14.4-alpine
DB_HOST		?= $(shell awk -F '=' '/^DB_HOST/ { print $$NF }' .env)
DB_PORT		?= $(shell awk -F '=' '/^DB_PORT/ { print $$NF }' .env)
DB_NAME		?= $(shell awk -F '=' '/^DB_NAME/ { print $$NF }' .env)
DB_USERNAME	?= $(shell awk -F '=' '/^DB_USERNAME/ { print $$NF }' .env)
DB_PASSWORD	?= $(shell awk -F '=' '/^DB_PASSWORD/ { print $$NF }' .env)
COMPOSE_FILE	?= docker-compose.yml


all: test lint build-webserver

build-webserver:
	-rm -rf cmd/webserver/schemas
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build $(LDFLAGS) \
		-o target/webserver \
		cmd/webserver/main.go

build-dbmigrate:
	-rm -rf cmd/dbmigrate/schemas
	# Embed cannot travel to parent directories, hence copy
	# migration files here.
	cp -R sql/schemas cmd/dbmigrate/
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build $(LDFLAGS) \
		-o target/dbmigrate \
		cmd/dbmigrate/main.go

.PHONY: clean
clean:
	rm -rf target/

install-dependencies:
	@go get -d -v ./...

lint:
	@golangci-lint run ./...

vulncheck:
	@govulncheck -v ./...

escape-analysis:
	$(GO) build -gcflags="-m" 2>&1

docker-build:
	@DOCKER_BUILDKIT=$(DOCKER_BUILDKIT) $(DOCKER) \
			build --rm --target app -t $(APP)-build .

docker-get-artifact:
	-mkdir -p target/webserver
	$(DOCKER) create -ti --name tmp $(APP)-builder /bin/bash
	$(DOCKER) cp tmp:/go/src/app/target/webserver target/webserver/main
	$(DOCKER) rm -f tmp

build-artifact: docker-build docker-get-artifact

migrate-add:
	@echo "Creating a new database migration"
	@goose -dir sql/schemas/ create $(name) sql

migrate-status: build-dbmigrate
	@echo "Status of database migrations"
	@./target/dbmigrate -s

migrate-all: build-dbmigrate
	@echo "Performing all database migrations"
	@./target/dbmigrate -m

create-db:
	-@$(PSQL_CLIENT) postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/ \
		-q -c "CREATE DATABASE $(DB_NAME) OWNER postgres ENCODING UTF8;"

db-dump:
	$(PG_DUMP) postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME) \
		> $(DB_NAME)_dump_$(shell date "+%Y-%m-%d_%H:%M:%S").sql

db-restore:
	$(PSQL_CLIENT) postgresql://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME) \
		-q -f $(RESTORE_FILE)

postgresql:
	@$(DOCKER) compose -f $(COMPOSE_FILE) up -d
	echo "Waiting database to start up..."
	@sleep 1

start-db: postgresql create-db migrate-all

stop-db:
	@$(DOCKER) compose down

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test:
	go test ./...

# This runs all tests, including integration tests
test-integration: start-db
	-@go test -tags=integration ./...
	@docker compose down

.PHONY: sqlc
sqlc:
	sqlc generate
