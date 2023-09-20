.DEFAULT_GOAL:=help

#============================================================================

# Load environment variables for local development
include .env
export

#============================================================================

.PHONY: dev
dev:							## Run dev container
	@echo $(shell cat $(shell eval echo ${SYSTEM_CONFIG_PATH})/user_uid)
	@docker compose ls -q | grep -q "instill-base" && true || \
		(echo "Error: Run \"make latest PROFILE=mgmt ITMODE_ENABLED=true\" in base repository (https://github.com/instill-ai/base) in your local machine first." && exit 1)
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

.PHONY: logs
logs:					## Tail service container logs with -n 10
	@docker logs ${SERVICE_NAME} --follow --tail=10

.PHONY: stop
stop:							## Stop container
	@docker stop -t 1 ${SERVICE_NAME}

.PHONY: top
top:							## Display all running service processes
	@docker top ${SERVICE_NAME}

.PHONY: build
build:							## Build dev docker image
	@docker build \
		--build-arg SERVICE_NAME=${SERVICE_NAME} \
		--build-arg GOLANG_VERSION=${GOLANG_VERSION} \
		--build-arg K6_VERSION=${K6_VERSION} \
		-f Dockerfile.dev  -t instill/${SERVICE_NAME}:dev .

.PHONY: go-gen
go-gen:       					## Generate codes
	go generate ./...

.PHONY: unit-test
unit-test:       				## Run unit test
	@go test -v -race -coverpkg=./... -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out
	@rm coverage.out

.PHONY: integration-test
integration-test:				## Run integration test
	@TEST_FOLDER_ABS_PATH=${PWD} k6 run \
		-e API_GATEWAY_PROTOCOL=${API_GATEWAY_PROTOCOL} -e API_GATEWAY_URL=${API_GATEWAY_URL} \
		integration-test/grpc.js --no-usage-report --quiet
	@TEST_FOLDER_ABS_PATH=${PWD} k6 run \
		-e API_GATEWAY_PROTOCOL=${API_GATEWAY_PROTOCOL} -e API_GATEWAY_URL=${API_GATEWAY_URL} \
		integration-test/rest.js --no-usage-report --quiet

.PHONY: help
help:       	 				## Show this help
	@echo "\nMakefile for local development"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m (default: help)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
