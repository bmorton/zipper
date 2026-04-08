.PHONY: test bench vet lint build generate all

test:
	go test -race -count=1 ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

lint: vet
	go mod tidy -diff
	golangci-lint run ./...

build:
	go build -o bin/zipper ./cmd/zipper

generate:
	go run ./cmd/generate --output data/zipcodes.csv.gz

all: lint test bench
