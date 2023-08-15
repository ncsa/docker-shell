# Default prefix (can be overridden by the user)
PREFIX ?= /usr/local

# Name of the resulting binary after building
BIN_NAME = docker-shell

# Go build command
GO_BUILD = go build

all: build

build:
	$(GO_BUILD) -o $(BIN_NAME) docker-shell.go

install: build
	install -m 755 $(BIN_NAME) $(PREFIX)/bin

clean:
	rm -f $(BIN_NAME)

.PHONY: all build install clean
