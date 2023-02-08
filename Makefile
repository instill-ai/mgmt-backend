.DEFAULT_GOAL:=help

#============================================================================

# Load environment variables for local development
include .env
export

#============================================================================

.PHONY: dev
dev:							## Run dev container
	@docker compose ls -q | grep -q "instill-vdp" && true || \
		(echo "Error: Run \"make dev PROFILE=mgmt ITMODE=true\" in vdp repository (https://github.com/instill-ai/vdp) in your local machine first." && exit 1)
	@docker inspect --type container ${SERVICE_NAME}-public >/dev/null 2>&1 && echo "A container named ${SERVICE_NAME}-public is already running." || \
		echo "Run dev container ${SERVICE_NAME}-public. To stop it, run \"make stop\"."
	@docker run -d --rm \
		-u $(id -u):$(id -g) \
		-v $(PWD):/${SERVICE_NAME} \
		-p ${PUBLIC_SERVICE_PORT}:${PUBLIC_SERVICE_PORT} \
		--network instill-network \
		--name ${SERVICE_NAME}-public \
		instill/${SERVICE_NAME}:dev >/dev/null 2>&1
	@docker inspect --type container ${SERVICE_NAME}-admin >/dev/null 2>&1 && echo "A container named ${SERVICE_NAME}-admin is already running." || \
		echo "Run dev container ${SERVICE_NAME}-admin. To stop it, run \"make stop\"."
	@docker run -d --rm \
		-u $(id -u):$(id -g) \
		-v $(PWD):/${SERVICE_NAME} \
		--network instill-network \
		--name ${SERVICE_NAME}-admin \
		instill/${SERVICE_NAME}:dev >/dev/null 2>&1

.PHONY: logs-public
logs-public:					## Tail public service container logs with -n 10
	@docker logs ${SERVICE_NAME}-public --follow --tail=10

.PHONY: logs-admin
logs-admin:						## Tail admin service container logs with -n 10
	@docker logs ${SERVICE_NAME}-admin --follow --tail=10

.PHONY: stop
stop:							## Stop container
	@docker stop -t 1 ${SERVICE_NAME}-public
	@docker stop -t 1 ${SERVICE_NAME}-admin

.PHONY: top
top:							## Display all running service processes
	@docker top ${SERVICE_NAME}-public
	@docker top ${SERVICE_NAME}-admin

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
	@TEST_FOLDER_ABS_PATH=${PWD} k6 run -e MODE=$(MODE) integration-test/rest.js --no-usage-report --quiet

.PHONY: help
help:       	 				## Show this help
	@echo "\nMakefile for locel development"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m (default: help)\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
