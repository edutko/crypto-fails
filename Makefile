REPO ?= $(shell grep -E '^module ' go.mod | cut -d ' ' -f 2)
VERSION ?= $(shell git describe --tags)
GO_INTERNAL_FILES=$(shell find internal -name '*.go')
GO_PKG_FILES=$(shell find pkg -name '*.go')
TEMPL_FILES=$(shell find internal -name '*.templ')
OUT=out

server: templ $(OUT)/server
license-tool: $(OUT)/license-tool

test:
	go test -race ./...

clean:
	rm -rf out/*

image: templ build/server/Dockerfile cmd/server/main.go $(GO_INTERNAL_FILES) $(GO_PKG_FILES)
	docker build --platform linux/amd64,linux/arm64 --build-arg VERSION=$(VERSION) \
		-t 5-dollar-wrench:$(VERSION) -f build/server/Dockerfile .

templ: $(TEMPL_FILES)
	go tool github.com/a-h/templ/cmd/templ generate ./internal/view/...

$(OUT)/license-tool: tools/license-tool $(GO_INTERNAL_FILES) $(GO_PKG_FILES)
	CGO_ENABLED=0 go build -o $@ ./tools/license-tool

$(OUT)/server: cmd/server/main.go $(GO_INTERNAL_FILES) $(GO_PKG_FILES)
	CGO_ENABLED=0 go build -ldflags "-X $(REPO)/internal/app.Version=$(VERSION)" -o $@ ./cmd/server
