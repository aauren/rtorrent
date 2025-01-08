.DEFAULT_GOAL := all
BUILD_IN_DOCKER?=true
IS_ROOT=$(filter 0,$(shell id -u))
IN_DOCKER_GROUP=$(filter docker,$(shell groups))
DOCKER=$(if $(or $(IN_DOCKER_GROUP),$(IS_ROOT),$(OSX)),docker,sudo docker)
GO_MOD_CACHE?=$(shell go env GOMODCACHE)
GO_CACHE?=$(shell go env GOCACHE)
DOCKER_LINT_IMAGE?=golangci/golangci-lint:v1.63.4
DOCKER_BUILD_IMAGE?=golang:1.23.4-alpine3.21

.PHONY: all test lint genmoqs instdeps gofmt gofmt-fix build

gofmt:
	gofmt -l -s $(shell find . -not \( \( -wholename '*/vendor/*' \) -prune \) -name '*.go')

gofmt-fix:
	goimports -w $(shell find . -not \( \( -wholename '*/vendor/*' \) -prune \) -name '*.go')
	gofmt -s -w $(shell find . -not \( \( -wholename '*/vendor/*' \) -prune \) -name '*.go')

instdeps:
	go get ./...
	go install go.uber.org/mock/mockgen@latest

updatedeps:
	go get -u ./...

genmoqs: instdeps
	rm -f rtorrent/rtorrent_moq.go
	mockgen -source=rtorrent/rtorrent.go -destination=rtorrent/rtorrent_moq.go -mock_names=New=NewMockClient -package=rtorrent

lint: gofmt
ifeq "$(BUILD_IN_DOCKER)" "true"
	$(DOCKER) run -v $(PWD):/go/src/github.com/aauren/rtorrent \
		-v $(GO_CACHE):/root/.cache/go-build \
		-v $(GO_MOD_CACHE):/go/pkg/mod \
		-w /go/src/github.com/aauren/rtorrent $(DOCKER_LINT_IMAGE) \
		bash -c \
		'golangci-lint run ./...'
else
	golangci-lint run ./...
endif

test: gofmt ## Runs code quality pipelines (gofmt, tests, coverage, etc)
ifeq "$(BUILD_IN_DOCKER)" "true"
	$(DOCKER) run -v $(PWD):/go/src/github.com/aauren/rtorrent \
		-v $(GO_CACHE):/root/.cache/go-build \
		-v $(GO_MOD_CACHE):/go/pkg/mod \
		-w /go/src/github.com/aauren/rtorrent $(DOCKER_BUILD_IMAGE) \
		sh -c \
		'CGO_ENABLED=0 go test -v -timeout 30s github.com/aauren/rtorrent/rtorrent/...'
else
	go test -v -timeout 30s github.com/aauren/rtorrent/rtorrent/...
endif

build:
ifeq "$(BUILD_IN_DOCKER)" "true"
	$(DOCKER) run -v $(PWD):/go/src/github.com/aauren/rtorrent \
		-v $(GO_CACHE):/root/.cache/go-build \
		-v $(GO_MOD_CACHE):/go/pkg/mod \
		-w /go/src/github.com/aauren/rtorrent $(DOCKER_BUILD_IMAGE) \
		sh -c \
		'CGO_ENABLED=0 go build -v github.com/aauren/rtorrent/rtorrent/...'
else
	go build -v github.com/aauren/rtorrent/rtorrent/...
endif

all: lint test build
