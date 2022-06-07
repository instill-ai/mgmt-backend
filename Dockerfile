FROM golang:1.18.2 AS build

WORKDIR /go/src
COPY . /go/src

RUN go get -d -v ./...

RUN --mount=type=cache,target=/root/.cache/go-build go build -o /mgmt-backend ./cmd/main
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /mgmt-backend-migrate ./cmd/migration
RUN --mount=type=cache,target=/root/.cache/go-build go build -o /mgmt-backend-init ./cmd/init

FROM gcr.io/distroless/base AS runtime

ENV GIN_MODE=release
WORKDIR /mgmt-backend

COPY --from=build /go/src/config ./config
COPY --from=build /go/src/release-please ./release-please
COPY --from=build /go/src/internal/db/migration ./internal/db/migration

COPY --from=build /mgmt-backend ./
COPY --from=build /mgmt-backend-migrate ./
COPY --from=build /mgmt-backend-init ./

ENTRYPOINT ["./mgmt-backend"]
