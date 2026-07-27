SHELL := /bin/bash
export GOTOOLCHAIN := go1.26.5

TOOLS_BIN := $(CURDIR)/.tools/bin
GOFUMPT := $(TOOLS_BIN)/gofumpt
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
NILAWAY := $(TOOLS_BIN)/nilaway
MODULE := http-services
BIN_NAME ?= http-services
BIN_DIR ?= $(CURDIR)/bin

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
DIST_DIR ?= dist
CGO_ENABLED ?= 0
PACKAGE_FILES ?= README.md config.yaml.example
PACKAGE_DIRS ?= static

.PHONY: help tools fmt fmt-check lint test test-race test-coverage test-integration pure-loc mod-verify build build-local build-cross build-check run dev migrate clean clean-dist version verify

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

tools:
	env GOTOOLCHAIN=$(GOTOOLCHAIN) GOBIN="$(TOOLS_BIN)" go install mvdan.cc/gofumpt@v0.10.0
	env GOTOOLCHAIN=$(GOTOOLCHAIN) GOBIN="$(TOOLS_BIN)" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	env GOTOOLCHAIN=$(GOTOOLCHAIN) GOBIN="$(TOOLS_BIN)" go install go.uber.org/nilaway/cmd/nilaway@v0.0.0-20260709164232-fd97c8330720

fmt:
	$(GOFUMPT) -w .
	env GOTOOLCHAIN=$(GOTOOLCHAIN) $(GOLANGCI_LINT) fmt

fmt-check:
	@set -o pipefail; output="$$($(GOFUMPT) -l .)" || exit $$?; test -z "$$output" || { printf '%s\n' "$$output"; exit 1; }
	env GOTOOLCHAIN=$(GOTOOLCHAIN) $(GOLANGCI_LINT) fmt --diff

lint:
	env GOTOOLCHAIN=$(GOTOOLCHAIN) $(GOLANGCI_LINT) run --timeout 5m ./...
	env GOTOOLCHAIN=$(GOTOOLCHAIN) $(NILAWAY) -include-pkgs="$(MODULE)" ./...
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go vet ./...

test: ## 运行 race detector、随机顺序且不使用测试缓存
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go test -race -shuffle=on -count=1 ./...

test-race: test ## test 的兼容别名

test-coverage: ## 运行覆盖率测试并生成 coverage.out
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go test -coverprofile=coverage.out -covermode=atomic ./...
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go tool cover -func=coverage.out

test-integration:
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go test -tags=integration -race -shuffle=on -count=1 ./...

pure-loc:
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go run ./tools/check-pure-loc -max 250

mod-verify:
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go mod verify

build:
	@if [ "$(CROSS)" = "1" ]; then $(MAKE) build-cross; else $(MAKE) build-local; fi

build-local:
	mkdir -p "$(BIN_DIR)"
	env GOTOOLCHAIN=$(GOTOOLCHAIN) go build -trimpath -ldflags "$(LDFLAGS)" -o "$(BIN_DIR)/$(BIN_NAME)" .

build-cross: clean-dist
	@mkdir -p "$(DIST_DIR)"
	@set -e; for platform in $(PLATFORMS); do \
		os="$${platform%/*}"; arch="$${platform#*/}"; extension=""; \
		[ "$$os" = "windows" ] && extension=".exe"; \
		output_dir="$(DIST_DIR)/$(BIN_NAME)_$(VERSION)_$${os}_$${arch}"; \
		mkdir -p "$$output_dir"; \
		GOOS="$$os" GOARCH="$$arch" CGO_ENABLED=$(CGO_ENABLED) GOTOOLCHAIN=$(GOTOOLCHAIN) \
			go build -trimpath -ldflags "$(LDFLAGS) -s -w" -o "$$output_dir/$(BIN_NAME)$$extension" .; \
		for file in $(PACKAGE_FILES); do [ -f "$$file" ] && cp "$$file" "$$output_dir/" || true; done; \
		for directory in $(PACKAGE_DIRS); do [ -d "$$directory" ] && cp -R "$$directory" "$$output_dir/" || true; done; \
		if command -v zip >/dev/null 2>&1 && [ "$$os" = "windows" ]; then \
			(cd "$(DIST_DIR)" && zip -q -r "$(BIN_NAME)_$(VERSION)_$${os}_$${arch}.zip" "$$(basename "$$output_dir")"); \
		else \
			(cd "$(DIST_DIR)" && tar -czf "$(BIN_NAME)_$(VERSION)_$${os}_$${arch}.tar.gz" "$$(basename "$$output_dir")"); \
		fi; \
		rm -rf "$$output_dir"; \
	done

build-check:
	@tmp_dir="$$(mktemp -d "$${TMPDIR:-/tmp}/$(BIN_NAME)-build.XXXXXX")"; trap 'rm -rf "$$tmp_dir"' EXIT INT TERM; env GOTOOLCHAIN=$(GOTOOLCHAIN) go build -trimpath -ldflags "$(LDFLAGS)" -o "$$tmp_dir/" ./...

run: build-local
	"$(BIN_DIR)/$(BIN_NAME)"

dev: build-local
	"$(BIN_DIR)/$(BIN_NAME)" --dev

migrate: build-local
	"$(BIN_DIR)/$(BIN_NAME)" --migrate

clean:
	rm -rf "$(BIN_DIR)"

clean-dist:
	rm -rf "$(DIST_DIR)"

version:
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

verify: fmt-check lint test pure-loc mod-verify build-check

.DEFAULT_GOAL := help
