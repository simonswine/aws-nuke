TRAVIS_COMMIT ?= unknown
TRAVIS_TAG    ?= unknown

.PHONY: all test verify

all: verify build

verify: go_verify

build: go_build

go_verify: go_fmt go_vet go_test

go_build:
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -a -tags netgo -ldflags '-w -X main.version=$(TRAVIS_TAG) -X main.commit=$(TRAVIS_COMMIT) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)' -o aws-nuke

go_verify: go_fmt go_vet go_test

go_test:
	go test $$(go list ./pkg/... ./cmd/...)

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

go_vet:
	go vet $$(go list ./pkg/... ./cmd/...)
