FROM golang:1.25-alpine AS builder  

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o /app/bin/sub-service \
  ./cmd/sub-service/main.go
  
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o /app/bin/migrate \
  ./cmd/migrate/main.go

FROM alpine:latest AS runtime
RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 appusers && adduser -D -u 1000 -G appusers appuser
WORKDIR /app  

COPY --from=builder /app/bin/sub-service /app/sub-service
COPY --from=builder /app/bin/migrate /app/migrate

COPY --from=builder /app/migrations /app/migrations

RUN chown -R appuser:appusers /app
USER appuser

FROM runtime AS sub-service
EXPOSE 8083
ENTRYPOINT ["/app/sub-service"]

FROM runtime AS migrate
ENTRYPOINT ["/app/migrate"]