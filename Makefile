.PHONY: all test clean

PACKAGE_NAME ?=theo-agent

PACKAGE_NAMESPACE=github.com/theoapp/$(PACKAGE_NAME)
COMMON_PACKAGE_NAMESPACE=$(PACKAGE_NAMESPACE)/common

VERSION := $(shell ./ci/version)
REVISION := $(shell git rev-parse --short=8 HEAD || echo unknown)
BRANCH := $(shell git show-ref | grep "$(REVISION)" | grep -v HEAD | awk '{print $$2}' | sed 's|refs/remotes/origin/||' | sed 's|refs/heads/||' | sort | head -n 1)
BUILT := $(shell date -u +%Y-%m-%dT%H:%M:%S%z)

GO_LDFLAGS ?= -X $(COMMON_PACKAGE_NAMESPACE).NAME=$(PACKAGE_NAME) -X $(COMMON_PACKAGE_NAMESPACE).VERSION=$(VERSION) \
              -X $(COMMON_PACKAGE_NAMESPACE).REVISION=$(REVISION) -X $(COMMON_PACKAGE_NAMESPACE).BUILT=$(BUILT) \
              -X $(COMMON_PACKAGE_NAMESPACE).BRANCH=$(BRANCH) \
              -s -w

all: test build

build: test
	mkdir -p build
	go build -ldflags "$(GO_LDFLAGS)" -o build/theo-agent .

test:
	go test ./...

clean:
	go clean ./...