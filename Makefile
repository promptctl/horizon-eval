# macklebox — a clean-room mackup reimplementation driven by appspec/.
#
# The binary is named mackup: the observable surface is the spec's, and
# macklebox is the project and module name only (appspec/00-overview.md).

GO      ?= go
BIN     := bin/mackup
PKG     := ./cmd/mackup

.PHONY: all build test vet fmt check clean

all: check

build: $(BIN)

$(BIN): $(shell find . -name '*.go' -not -path './bin/*')
	$(GO) build -o $(BIN) $(PKG)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

# fmt fails rather than rewrites, so an unformatted tree cannot pass check.
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed:"; echo "$$unformatted"; exit 1; \
	fi

check: fmt vet test

clean:
	rm -rf bin
