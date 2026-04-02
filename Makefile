APP=agsm

.PHONY: fmt test build run tidy

fmt:
	gofmt -w .

test:
	go test ./...

build:
	go build ./...

run:
	go run .

tidy:
	go mod tidy
