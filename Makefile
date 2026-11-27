BIN ?= $(HOME)/.local/bin/dj

.PHONY: build test install

build:
	go build -o dj ./cmd/dj

test:
	go test ./...

install:
	go build -o $(BIN) ./cmd/dj

.PHONY: demo
demo: build
	PATH="$(CURDIR):$$PATH" vhs demo/dj.tape
