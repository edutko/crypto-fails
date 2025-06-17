REPO ?= $(shell grep -E '^module ' go.mod | cut -d ' ' -f 2)
VERSION ?= $(shell git describe --tags)
GO_INTERNAL_FILES=$(shell find internal -name '*.go')
TEMPL_FILES=$(shell find internal -name '*.templ')

server: out/server

test:
	go test -race ./...

clean:
	rm -rf out/*

image: templ build/server/Dockerfile cmd/server/main.go $(GO_INTERNAL_FILES)
	docker build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VERSION) \
		-t 5-dollar-wrench:$(VERSION) -f build/server/Dockerfile .

templ: $(TEMPL_FILES)
	go tool github.com/a-h/templ/cmd/templ generate ./internal/view/...

out/server: templ cmd/server/main.go $(GO_INTERNAL_FILES)
	CGO_ENABLED=0 go build -ldflags "-X $(REPO)/internal/app.Version=$(VERSION)" -o $@ ./cmd/server
