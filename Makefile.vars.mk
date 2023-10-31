IMG_TAG ?= latest

GOUCRT_MAIN_GO ?= cmd/rtintg-denonavr/rtintg-denonavr.go
GOUCRT_GOOS ?= linux
GOUCRT_GOARCH ?= amd64
GOUCRT_GOARCH_ARM64 ?= arm64


CURDIR ?= $(shell pwd)
BIN_FILENAME ?= $(CURDIR)/$(PROJECT_ROOT_DIR)/rtintg-denonavr-amd64
BIN_FILENAME_ARM64 ?= $(CURDIR)/$(PROJECT_ROOT_DIR)/rtintg-denonavr-arm64
WORK_DIR = $(CURDIR)/.work

go_bin ?= $(PWD)/.work/bin
$(go_bin):
	@mkdir -p $@

golangci_bin = $(go_bin)/golangci-lint


# Image URL to use all building/pushing image targets
GOUCRT_GHCR_IMG ?= ghcr.io/splattner/remotetwo-integration-denonavr:$(IMG_TAG)

