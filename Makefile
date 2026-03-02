VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X github.com/nusii/nusii-cli/cmd.version=$(VERSION) \
           -X github.com/nusii/nusii-cli/cmd.commit=$(COMMIT) \
           -X github.com/nusii/nusii-cli/cmd.date=$(DATE)

.PHONY: build test lint install integration-test clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/nusii .

test:
	go test ./... -v -count=1

lint:
	go vet ./...

install:
	go install -ldflags "$(LDFLAGS)" .

integration-test:
	NUSII_INTEGRATION_TEST=1 go test ./test/integration/ -v -count=1

clean:
	rm -rf bin/
