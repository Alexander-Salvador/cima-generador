# CIMA Makefile
# Build, test, and install CIMA-specific Go binaries.
#
# CIMA uses OVAV's sheets bridge (see go.mod replace directive).
# This Makefile is minimal — CIMA has fewer targets than OVAV.

GO ?= go
CIMA_ROOT := $(shell git rev-parse --show-toplevel 2>/dev/null || pwd)
GO_RUNTIME := $(CIMA_ROOT)/go-runtime
OVAV_ROOT ?= /home/braka/Systems/ovav

# Targets
BIN_DIR := $(GO_RUNTIME)/build

.PHONY: help build test vet fmt clean install setup-vault

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build all CIMA binaries (sheets bridge + citrix_upload)
	@mkdir -p $(BIN_DIR)
	cd $(GO_RUNTIME) && $(GO) build -o $(BIN_DIR)/sheets ./cmd/sheets/
	cd $(GO_RUNTIME) && $(GO) build -o $(BIN_DIR)/citrix_upload ./cmd/citrix_upload/
	@echo "[ok] built sheets + citrix_upload in $(BIN_DIR)/"

test: ## Run tests for CIMA Go packages
	cd $(GO_RUNTIME) && $(GO) test ./cmd/sheets/... -v -count=1

vet: ## Run go vet
	cd $(GO_RUNTIME) && $(GO) vet ./cmd/sheets/... ./cmd/citrix_upload/...

fmt: ## Format Go source
	cd $(GO_RUNTIME) && $(GO) fmt ./...

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)

install: ## Install CIMA binaries to ~/.local/bin
	@mkdir -p $(HOME)/.local/bin
	cp $(BIN_DIR)/sheets $(HOME)/.local/bin/cima-sheets
	cp $(BIN_DIR)/citrix_upload $(HOME)/.local/bin/cima-citrix-upload
	@echo "[ok] installed to ~/.local/bin/ as cima-*"

setup-vault: ## Deploy Google OAuth credentials to .ovav/vault/
	@bash $(CIMA_ROOT)/scripts/setup-vault.sh