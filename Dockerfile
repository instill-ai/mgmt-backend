ARG GOLANG_VERSION=1.25.6
FROM golang:${GOLANG_VERSION} AS build

WORKDIR /build

ARG SERVICE_NAME SERVICE_VERSION TARGETOS TARGETARCH

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.serviceVersion=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}" \
    -o /${SERVICE_NAME} ./cmd/main

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.serviceVersion=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}-migrate" \
    -o /${SERVICE_NAME}-migrate ./cmd/migration

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.serviceVersion=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}-init" \
    -o /${SERVICE_NAME}-init ./cmd/init

RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags "-X main.serviceVersion=${SERVICE_VERSION} -X main.serviceName=${SERVICE_NAME}-worker" \
    -o /${SERVICE_NAME}-worker ./cmd/worker

FROM golang:${GOLANG_VERSION}

USER nobody:nogroup

ARG SERVICE_NAME SERVICE_VERSION

WORKDIR /${SERVICE_NAME}

COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME} ./
COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME}-migrate ./
COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME}-init ./
COPY --from=build --chown=nonroot:nonroot /${SERVICE_NAME}-worker ./

COPY --chown=nonroot:nonroot ./config ./config
COPY --chown=nonroot:nonroot ./pkg/db/migration ./pkg/db/migration

ENV SERVICE_NAME=${SERVICE_NAME}
ENV SERVICE_VERSION=${SERVICE_VERSION}
