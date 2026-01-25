FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /footprint ./cmd/footprint/

FROM alpine:3.21

# Install ca-certificates and git
RUN apk --no-cache add ca-certificates git

WORKDIR /

# Copy the binary from builder and entrypoint script
COPY --from=builder /footprint /footprint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
