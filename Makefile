# macklebox — a clean-room build of the tool specified in appspec/.
#
# The built binary is named `mackup`: appspec/02 fixes the observable command
# name, and `macklebox` is the project/package name only.

GO      ?= go
BIN     ?= bin/mackup
PKG     := ./cmd/mackup
VERSION ?=
LDFLAGS := $(if $(VERSION),-X github.com/promptctl/macklebox/internal/version.buildVersion=$(VERSION),)

.PHONY: all build test vet fmt check clean

all: check build

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

check: vet test

clean:
	rm -rf bin
