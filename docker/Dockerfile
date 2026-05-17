FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o skill-sync-server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/skill-sync-server .
EXPOSE 8080
VOLUME ["/app/server-data"]
ENTRYPOINT ["./skill-sync-server"]
CMD ["--addr", ":8080", "--db", "/app/server-data/sync.db"]