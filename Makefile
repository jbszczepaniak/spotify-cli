BINDIR := $(shell pwd)/bin

.PHONY: install_deps
install_deps:
	go mod download

.PHONY: build
build: install_deps
	mkdir -p $(BINDIR)
	packr -o $(BINDIR)/spotify-cli ./cmd/spotify-cli

.PHONY: clean
clean:
	rm -rf $(BINDIR)

.PHONY: test
test:
	go test ./...

.DEFAULT_GOAL := build