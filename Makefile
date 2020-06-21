BINDIR := $(shell pwd)/bin


install_deps:
	go mod download
	go get -u github.com/gobuffalo/packr/packr


build: install_deps
	mkdir -p $(BINDIR)
	$(shell go env GOPATH)/bin/packr
	go build -o $(BINDIR)/spotify-cli ./cmd/spotify-cli
	$(shell go env GOPATH)/bin/packr clean

packr:
	$(shell go env GOPATH)/bin/packr

clean:
	rm -rf $(BINDIR)


test:
	go test -v ./...


images:
	plantuml -tpng img/components.puml
	plantuml -tpng img/workflow.puml


release:
	goreleaser release

