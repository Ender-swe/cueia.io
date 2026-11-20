FROM golang:1.26.2-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
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
COPY backend/migrations/ /app/migrations/
ENTRYPOINT ["/app/migrate"]
