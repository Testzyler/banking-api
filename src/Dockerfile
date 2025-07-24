# Stage 1: Build
FROM golang:1.24.5-alpine3.21 AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o banking-api .

# Stage 2: Minimal Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/banking-api .

COPY --from=builder /app/config.yaml ./config.yaml

# Copy database seeds for migrations
COPY --from=builder /app/database/seeds ./database/seeds

CMD ["./banking-api"]