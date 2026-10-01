GO ?= go
BINARY ?= nth
BUILD_DIR ?= build
VERSION ?= dev
VERSION_PACKAGE = github.com/Hayao0819/nth/internal/version
VERSION_FLAGS = -X $(VERSION_PACKAGE).Version=$(VERSION)

.PHONY: build test race vet clean

build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath -ldflags='$(VERSION_FLAGS)' -o $(BUILD_DIR)/$(BINARY) .

test:
	$(GO) test ./...

race:
	CGO_ENABLED=1 $(GO) test -race ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out
