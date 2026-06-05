# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o jia-server ./cmd/server/main.go

# Final light stage
FROM alpine:3.19

WORKDIR /app

# Copy SSL certificates and binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/jia-server .

EXPOSE 3000

ENV PORT=3000
ENV ENV=production

CMD ["./jia-server"]
