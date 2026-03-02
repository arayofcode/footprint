-include .env
export

run:
	go run ./cmd/footprint/main.go

test:
	go test -count=1 ./...