.DEFAULT_GOAL := build

GO := go

PKG_NAME := ssl-pairgen

all: build

clean:
	@ $(GO) clean
	@ rm -rf target/

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

os = $(word 1,$(subst -, ,$@))
arch = $(word 2,$(subst -, ,$@))
platform = $(word 2,$(subst _, ,$@))

$(PLATFORMS): deps
	GOOS=$(os) GOARCH=$(arch) $(GO) build \
		-ldflags "-w -s" \
		-o target/$(PKG_NAME)_$@

TARGETS = $(addprefix target/$(PKG_NAME)_,$(PLATFORMS))

$(TARGETS): main.go cert.go
	make $(platform)

linux: target/$(PKG_NAME)_linux-amd64 target/$(PKG_NAME)_linux-arm64

darwin: target/$(PKG_NAME)_darwin-amd64 target/$(PKG_NAME)_darwin-arm64

fmt:
	$(GO) fmt

deps:

build: fmt $(TARGETS)

.PHONY: deps build linux darwin
