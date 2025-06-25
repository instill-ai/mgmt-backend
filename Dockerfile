ARG GOLANG_VERSION=1.24.2
FROM golang:${GOLANG_VERSION} AS build

ARG SERVICE_NAME

WORKDIR /src

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

ARG SERVICE_NAME SERVICE_VERSION TARGETOS TARGETARCH

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.version=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}" \
    -o /${SERVICE_NAME} ./cmd/main

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.version=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}" \
    -o /${SERVICE_NAME}-migrate ./cmd/migration

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.version=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}" \
    -o /${SERVICE_NAME}-init ./cmd/init

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.version=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}" \
    -o /${SERVICE_NAME}-worker ./cmd/worker

FROM golang:${GOLANG_VERSION}

USER nobody:nogroup

ARG SERVICE_NAME SERVICE_VERSION

WORKDIR /${SERVICE_NAME}

COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME} ./
COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME}-migrate ./
COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME}-init ./
COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME}-worker ./

COPY --from=build --chown=nonroot:nonroot /src/config ./config
COPY --from=build --chown=nonroot:nonroot /src/pkg/db/migration ./pkg/db/migration

ENV SERVICE_NAME=${SERVICE_NAME}
ENV SERVICE_VERSION=${SERVICE_VERSION}
