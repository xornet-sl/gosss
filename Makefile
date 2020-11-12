BINDIR        := $(CURDIR)/bin
INSTALL_PATH  ?= /usr/local/bin
DIST_DIRS     := find * -type d -exec
TARGETS       := linux/amd64 linux/386 linux/arm linux/arm64 darwin/amd64 windows/amd64
BINNAME       ?= gosss

# Required for globs to work correctly
SHELL         = /usr/bin/env bash

WHICH_GO      = $(shell which go > /dev/null; echo $$?)
GOBIN         =
ifeq ($(WHICH_GO),0)
GOBIN         = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN         = $(shell go env GOPATH)/bin
endif
endif
GOX           = $(GOBIN)/gox

ARCH          = $(shell uname -p)

TESTS         := .
TESTFLAGS     := -race -v
PKG           := ./...
LDFLAGS       := -s -w
GOFLAGS       :=
GOXFLAGS      := -parallel 8
SRC           := $(shell find cmd pkg -type f -iname '*.go' -print)

GIT_COMMIT    = $(shell git rev-parse HEAD)
GIT_SHA       = $(shell git rev-parse --short HEAD)
GIT_TAG       = $(shell git describe --tags --abbrev=0 --match='v*' --candidates=1 2>/dev/null)
GIT_DIRTY     = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

ifdef VERSION
BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}
VERSION ?= ${BINARY_VERSION}

ifneq ($(BINARY_VERSION),)
LDFLAGS += -X main.version=${BINARY_VERSION}
endif
LDFLAGS += -X main.commit=${GIT_SHA}


###################################3
# build

.PHONY: all
all: build

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o '$(BINDIR)/$(BINNAME)' ./cmd/gosss


###################################3
# tests

.PHONY: test
test: build
#test: test-style
test: test-unit

.PHONY: test-unit
test-unit:
	go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)


###################################3
# cross-build

$(GOX):
	(cd /; go get -u github.com/mitchellh/gox)

.PHONY: build-cross
build-cross: LDFLAGS += -extldflags "-static"
build-cross: $(GOX)
	CGO_ENABLED=0 $(GOX) $(GOXFLAGS) -output="_dist/{{.OS}}-{{.Arch}}/$(BINNAME)" -osarch='$(TARGETS)' -tags '$(TAGS)' -ldflags '$(LDFLAGS)' ./cmd/gosss

.PHONY: dist
dist: build-cross
	@echo Creating dist tarballs...
	@( \
		cd _dist && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf gosss-${VERSION}-{}.tar.gz {} \; && \
		$(DIST_DIRS) zip -qr gosss-${VERSION}-{}.zip {} \; && \
		for f in $$(ls ./*.{gz,zip} 2>/dev/null); do \
			shasum -a 256 "$${f}" > "$${f}.sha256sum" ; \
		done; \
	)


###################################
# install

.PHONY: install
install: $(BINDIR)/$(BINNAME)
	@install -v "$(BINDIR)/$(BINNAME)" "$(INSTALL_PATH)/$(BINNAME)"

.PHONY: uninstall
uninstall:
	@rm -fv $(INSTALL_PATH)/$(BINNAME)


###################################


.PHONY: clean
clean:
	@rm -rf '$(BINDIR)' ./_dist


.PHONY: info
info:
	 @echo "Version:           ${VERSION}"
	 @echo "Git Tag:           ${GIT_TAG}"
	 @echo "Git Commit:        ${GIT_COMMIT}"
	 @echo "Git Tree State:    ${GIT_DIRTY}"
