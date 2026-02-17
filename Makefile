.PHONY: build test lint run clean vet fmt

BINARY := forg

build:
	go build -o $(BINARY) .

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

vet:
	go vet ./...

fmt:
	gofmt -w .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
