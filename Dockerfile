FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /footprint ./cmd/footprint/

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /

COPY --from=builder /footprint /footprint

RUN mkdir -p /dist

ENTRYPOINT ["/footprint"]
