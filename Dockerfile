FROM golang:1.24-alpine as builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bookshop-api ./cmd/api

# Final image
FROM alpine:3.19

WORKDIR /app

# Install dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bookshop-api .
COPY --from=builder /app/config /app/config

# Export port
EXPOSE 8080

# Run application
CMD ["./bookshop-api"]
