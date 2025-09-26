# k6
FROM grafana/k6:0.51.0 AS k6stage

# 共通ベース
FROM golang:1.24.5 AS builder-base
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# server 用ビルド
FROM builder-base AS builder-worker
RUN CGO_ENABLED=0 go build -o /out/worker .

# runner 用ビルド
FROM builder-base AS builder-runner
RUN CGO_ENABLED=0 go build -o /out/runner ./cmd/runner

# ===== server（ベンチマーカ側） =====
FROM alpine:3.22 AS worker
RUN apk update && apk add --no-cache bash ca-certificates sed jq curl
WORKDIR /app
COPY --from=builder-worker /out/worker ./worker
COPY --from=k6stage /usr/bin/k6 /usr/bin/k6
COPY run_k6.sh /app/run_k6.sh
RUN chmod +x /app/run_k6.sh
COPY scenarios/ ./scenarios/
RUN mkdir -p logs scores
CMD ["./worker"]

# ===== ローカル実行用（k6 + スコア） =====
FROM alpine:3.22 AS local-runner
RUN apk update && apk add --no-cache bash ca-certificates sed jq curl
WORKDIR /app
COPY --from=builder-runner /out/runner ./runner
COPY --from=k6stage /usr/bin/k6 /usr/bin/k6
COPY run_k6.sh /app/run_k6.sh
RUN chmod +x /app/run_k6.sh
COPY scenarios/ ./scenarios/
RUN mkdir -p logs scores
ENTRYPOINT ["./runner"]
