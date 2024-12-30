# Build stage
FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gosss ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

RUN mkdir -p /app/data && chown -R appuser:appuser /app/data

WORKDIR /app

COPY --from=builder /app/gosss .

COPY --from=builder /app/.env.example ./.env

USER appuser

EXPOSE 8191

# Set environment variables
ENV PORT=8191

# Run the application
CMD ["./gosss"]