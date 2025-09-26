FROM 42tokyo2508.azurecr.io/backend:base AS build
WORKDIR /usr/src/backend

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /usr/local/bin/server ./cmd

FROM 42tokyo2508.azurecr.io/backend:base AS production
WORKDIR /usr/local/bin

COPY --from=build /usr/local/bin/server /usr/local/bin/server

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
