# Copyright 2025 The v3io-go Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	 http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO_VERSION := $(shell go version | cut -d " " -f 3)
GOPATH ?= $(shell go env GOPATH)
SHELL := /bin/bash

# Get default os / arch from go env
V3IO_DEFAULT_OS := $(shell go env GOOS)
V3IO_DEFAULT_ARCH := $(shell go env GOARCH || echo amd64)

V3IO_OS := $(if $(V3IO_OS),$(V3IO_OS),$(V3IO_DEFAULT_OS))
V3IO_ARCH := $(if $(V3IO_ARCH),$(V3IO_ARCH),$(V3IO_DEFAULT_ARCH))

# Version information
V3IO_VERSION_GIT_COMMIT = $(shell git rev-parse HEAD)
V3IO_PATH ?= $(shell pwd)

# Link flags
GO_LINK_FLAGS ?= -s -w

# Go test timeout
V3IO_GO_TEST_TIMEOUT ?= "30m"

#
# Must be first target
#
.PHONY: all
all:
	$(error "Please pick a target (run 'make targets' to view targets)")

#
# Linting and formatting
#
.PHONY: fmt
fmt: ensure-golangci-linter
	gofmt -s -w .
	$(GOLANGCI_LINT_BIN) run --fix

.PHONY: lint
lint: modules ensure-golangci-linter
	@echo Linting...
	$(GOLANGCI_LINT_BIN) run -v
	@echo Done.


GOLANGCI_LINT_VERSION := 2.3.0
GOLANGCI_LINT_BIN := $(CURDIR)/.bin/golangci-lint
GOLANGCI_LINT_INSTALL_COMMAND := GOBIN=$(CURDIR)/.bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)

.PHONY: ensure-golangci-linter
ensure-golangci-linter:
	@if ! command -v $(GOLANGCI_LINT_BIN) >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		$(GOLANGCI_LINT_INSTALL_COMMAND); \
	else \
		installed_version=$$($(GOLANGCI_LINT_BIN) version | awk '/version/ {gsub(/^v/, "", $$4); print $$4}'); \
		if [ "$$installed_version" != "$(GOLANGCI_LINT_VERSION)" ]; then \
			echo "golangci-lint version mismatch ($$installed_version != $(GOLANGCI_LINT_VERSION)). Reinstalling..."; \
			$(GOLANGCI_LINT_INSTALL_COMMAND); \
		fi \
	fi

#
# Environment validation
#
.PHONY: check-env
check-env:
ifndef V3IO_DATAPLANE_URL
		$(error V3IO_DATAPLANE_URL is undefined)
endif
ifndef V3IO_DATAPLANE_USERNAME
		$(error V3IO_DATAPLANE_USERNAME is undefined)
endif
ifndef V3IO_DATAPLANE_ACCESS_KEY
		$(error V3IO_DATAPLANE_ACCESS_KEY is undefined)
endif
ifndef V3IO_CONTROLPLANE_URL
		$(error V3IO_CONTROLPLANE_URL is undefined)
endif
ifndef V3IO_CONTROLPLANE_USERNAME
		$(error V3IO_CONTROLPLANE_USERNAME is undefined)
endif
ifndef V3IO_CONTROLPLANE_PASSWORD
		$(error V3IO_CONTROLPLANE_PASSWORD is undefined)
endif
ifndef V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD
		$(error V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD is undefined)
endif
	@echo "All required env vars populated"

#
# Schema generation
#
.PHONY: generate-capnp
generate-capnp:
	@echo "Generating Cap'n Proto schemas..."
	@cd ./pkg/dataplane/schemas/; ./build
	@echo "Schema generation complete."

.PHONY: clean-schemas
clean-schemas:
	@echo "Cleaning generated schemas..."
	@cd ./pkg/dataplane/schemas/; ./clean
	@echo "Schema cleanup complete."

#
# Building
#
.PHONY: build-lib
build-lib: modules
	@echo "Building v3io-go library..."
	@go build ./pkg/...
	@echo "Library build successful."

.PHONY: build
build: clean-schemas generate-capnp build-lib lint test
	@echo "Full build pipeline complete."

#
# Testing
#
.PHONY: test-unit
test-unit: modules ensure-gopath
	go test -race -tags=test_unit -v ./pkg/... -short

.PHONY: test-integration
test-integration: modules ensure-gopath
	go test -race -tags=test_integration -v ./pkg/... --timeout $(V3IO_GO_TEST_TIMEOUT)

.PHONY: test-coverage
test-coverage:
	go test -tags=test_unit -coverprofile=coverage.out ./pkg/... || go tool cover -html=coverage.out

.PHONY: test-controlplane
test-controlplane: modules ensure-gopath check-env
	go test -test.v=true -race -tags=test_unit -count=1 ./pkg/controlplane/...

.PHONY: test-dataplane
test-dataplane: modules ensure-gopath check-env
	go test -test.v=true -race -tags=test_unit -count=1 ./pkg/dataplane/...

.PHONY: test-dataplane-simple
test-dataplane-simple: modules ensure-gopath check-env
	go test -test.v=true -tags=test_unit -count=1 ./pkg/dataplane/...

.PHONY: test-system
test-system: test-controlplane test-dataplane-simple

.PHONY: test
test: test-unit

.PHONY: build-test-container
build-test-container:
	@echo "Building test container..."
	docker build \
		--file hack/test/docker/Dockerfile \
		--tag v3io-go-test:latest \
		.

.PHONY: test-system-in-docker
test-system-in-docker: build-test-container
	@echo "Running system test in docker container..."
	docker run --rm \
		--env V3IO_DATAPLANE_URL="$(V3IO_DATAPLANE_URL)" \
		--env V3IO_DATAPLANE_USERNAME="$(V3IO_DATAPLANE_USERNAME)" \
		--env V3IO_DATAPLANE_ACCESS_KEY="$(V3IO_DATAPLANE_ACCESS_KEY)" \
		--env V3IO_CONTROLPLANE_URL="$(V3IO_CONTROLPLANE_URL)" \
		--env V3IO_CONTROLPLANE_USERNAME="$(V3IO_CONTROLPLANE_USERNAME)" \
		--env V3IO_CONTROLPLANE_PASSWORD="$(V3IO_CONTROLPLANE_PASSWORD)" \
		--env V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD="$(V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD)" \
		v3io-go-test:latest make test-system
	@echo "Docker test complete."

#
# Go modules and environment
#
.PHONY: ensure-gopath
ensure-gopath:
ifndef GOPATH
	$(error GOPATH must be set)
endif

.PHONY: modules
modules: ensure-gopath
	@go mod download

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: clean
clean: clean-schemas
	@echo "Cleaning build artifacts..."
	@go clean -cache
	@rm -rf .bin/
	@rm -rf coverage.out
	@echo "Clean complete."

.PHONY: targets
targets:
	@awk -F: '/^[^ \t="]+:/ && !/PHONY/ {print $$1}' Makefile | sort -u

#
# Versioning
#
.PHONY: version
version:
	@echo "Git commit: $(V3IO_VERSION_GIT_COMMIT)"
	@echo "Go version: $(GO_VERSION)"
	@echo "OS/Arch: $(V3IO_OS)/$(V3IO_ARCH)"