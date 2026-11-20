FROM oven/bun:1.3.10 AS dashboard
WORKDIR /workspace
COPY package.json bun.lock tsconfig.base.json ./
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY apps/browser-worker/package.json apps/browser-worker/package.json
COPY packages/ packages/
RUN bun install --frozen-lockfile --linker hoisted
WORKDIR /workspace/apps/dashboard
CMD ["bun", "--bun", "run", "dev", "--host", "0.0.0.0", "--strictPort"]

FROM golang:1.26.2-alpine AS build
WORKDIR /src
COPY apps/server/go.mod apps/server/go.sum ./
RUN go mod download
COPY apps/server/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/ ./cmd/...

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates
WORKDIR /app
USER 65532:65532

FROM runtime AS api
COPY --from=build /out/api /app/api
ENTRYPOINT ["/app/api"]

FROM runtime AS worker
COPY --from=build /out/worker /app/worker
ENTRYPOINT ["/app/worker"]

FROM runtime AS migrate
COPY --from=build /out/migrate /app/migrate
COPY apps/server/migrations/ /app/migrations/
ENTRYPOINT ["/app/migrate"]
