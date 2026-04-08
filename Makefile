.PHONY: test bench vet build generate all

test:
	go test -race -count=1 ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

build:
	go build -o bin/zipper ./cmd/zipper

generate:
	go run ./cmd/generate --output data/zipcodes.csv.gz

all: vet test bench
