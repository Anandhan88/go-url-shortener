# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener .

# ---

# Run stage
FROM alpine:3.21

WORKDIR /app

# Copy binary and assets from builder
COPY --from=builder /app/url-shortener .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./url-shortener"]
