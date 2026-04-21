VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build install clean test lint

build:
	go build $(LDFLAGS) -o kachlan .

install:
	go install $(LDFLAGS) .

clean:
	rm -f kachlan
	rm -rf dist/

test:
	go test ./...

lint:
	golangci-lint run
