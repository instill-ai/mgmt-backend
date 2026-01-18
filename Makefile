.DEFAULT_GOAL:=help

#============================================================================

# Load environment variables for local development
include .env
export

# Integration test configuration
# - From host: make integration-test (uses localhost:8080)
# - In container: make integration-test API_GATEWAY_URL=api-gateway:8080 DB_HOST=pg_sql
API_GATEWAY_PROTOCOL ?= http
API_GATEWAY_URL ?= localhost:8080
DB_HOST ?= localhost

GOTEST_FLAGS := CFG_DATABASE_HOST=${TEST_DBHOST} CFG_DATABASE_NAME=${TEST_DBNAME}
ifeq (${DBTEST}, true)
	GOTEST_TAGS := -tags=dbtest
endif

#============================================================================

.PHONY: dev
dev:							## Run dev container
	@docker compose ls -q | grep -q "instill-core" && true || \
		(echo "Error: Run \"make latest\" in core repository (https://github.com/instill-ai/core) in your local machine first and run \"docker rm -f ${SERVICE_NAME}\" " && exit 1)
	@docker inspect --type container ${SERVICE_NAME} >/dev/null 2>&1 && echo "A container named ${SERVICE_NAME} is already running." || \
		echo "Run dev container ${SERVICE_NAME}. To stop it, run \"make stop\"."
	@docker run -d --rm \
		-v $(PWD):/${SERVICE_NAME} \
		-p ${PUBLIC_SERVICE_PORT}:${PUBLIC_SERVICE_PORT} \
		-p ${PRIVATE_SERVICE_PORT}:${PRIVATE_SERVICE_PORT} \
		-e CFG_SERVER_DEFAULTUSERUID=$(shell cat $(shell eval echo ${SYSTEM_CONFIG_PATH})/user_uid) \
		--network instill-network \
		--name ${SERVICE_NAME} \
		instill/${SERVICE_NAME}:dev >/dev/null 2>&1

.PHONY: latest
latest: ## Run latest container
	@docker compose ls -q | grep -q "instill-core" && true || \
		(echo "Error: Run \"make latest\" in instill-core repository (https://github.com/instill-ai/instill-core) in your local machine first and run \"docker rm -f ${SERVICE_NAME}\"." && exit 1)
	@docker inspect --type container ${SERVICE_NAME} >/dev/null 2>&1 && echo "A container named ${SERVICE_NAME} is already running." || \
		echo "Run latest container ${SERVICE_NAME}. To stop it, run \"make stop\"."
	@docker run --network=instill-network \
		--name ${SERVICE_NAME} \
		-d instill/${SERVICE_NAME}:latest \
		/bin/sh -c "\
		./${SERVICE_NAME}-migrate && \
		./${SERVICE_NAME}-init && \
		./${SERVICE_NAME} \
		"

.PHONY: logs
logs:					## Tail service container logs with -n 10
	@docker logs ${SERVICE_NAME} --follow --tail=10

.PHONY: stop
stop:							## Stop container
	@docker stop -t 1 ${SERVICE_NAME}

.PHONY: rm
rm:								## Remove container
	@docker rm -f ${SERVICE_NAME}

.PHONY: top
top:							## Display all running service processes
	@docker top ${SERVICE_NAME}

.PHONY: build-dev
build-dev: ## Build dev docker image
	@docker build \
		--build-arg SERVICE_NAME=${SERVICE_NAME} \
		--build-arg K6_VERSION=${K6_VERSION} \
		-f Dockerfile.dev -t instill/${SERVICE_NAME}:dev .

.PHONY: build-latest
build-latest: ## Build latest docker image
	@docker build \
		--build-arg SERVICE_NAME=${SERVICE_NAME} \
		--build-arg SERVICE_VERSION=dev \
		-t instill/${SERVICE_NAME}:latest .

.PHONY: go-gen
go-gen:       					## Generate codes
	go generate ./...


.PHONY: dbtest-pre
dbtest-pre:
	@DBTEST=true ${GOTEST_FLAGS} go run ./cmd/migration

.PHONY: coverage
coverage:
	@if [ "${DBTEST}" = "true" ]; then  make dbtest-pre; fi
	@${GOTEST_FLAGS} go test -v -race ${GOTEST_TAGS} -coverpkg=./... -coverprofile=coverage.out -covermode=atomic ./...
	@if [ "${HTML}" = "true" ]; then  \
		go tool cover -func=coverage.out && \
		go tool cover -html=coverage.out && \
		rm coverage.out; \
	fi

.PHONY: integration-test
integration-test:				## Run integration test
	@echo "✓ Running tests via API Gateway: ${API_GATEWAY_URL}"
	@echo "  DB_HOST: ${DB_HOST}"
	@rm -f /tmp/mgmt-integration-test.log
	@TEST_FOLDER_ABS_PATH=${PWD} k6 run --address="" \
		-e API_GATEWAY_PROTOCOL=${API_GATEWAY_PROTOCOL} \
		-e API_GATEWAY_URL=${API_GATEWAY_URL} \
		-e DB_HOST=${DB_HOST} \
		integration-test/grpc.js --no-usage-report 2>&1 | tee -a /tmp/mgmt-integration-test.log
	@TEST_FOLDER_ABS_PATH=${PWD} k6 run --address="" \
		-e API_GATEWAY_PROTOCOL=${API_GATEWAY_PROTOCOL} \
		-e API_GATEWAY_URL=${API_GATEWAY_URL} \
		-e DB_HOST=${DB_HOST} \
		integration-test/rest.js --no-usage-report 2>&1 | tee -a /tmp/mgmt-integration-test.log

.PHONY: help
help:       	 				## Show this help
	@echo "\nMakefile for local development"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m (default: help)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
