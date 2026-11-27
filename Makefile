BIN ?= $(HOME)/.local/bin/dj

.PHONY: build test install

build:
	go build -o dj .

test:
	go test ./...

install:
	go build -o $(BIN) .
