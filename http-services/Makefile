SHELL := /bin/bash
export GOTOOLCHAIN := go1.25.5
export GOFLAGS ?= -mod=readonly

.PHONY: help build build-local build-cross build-check run dev migrate clean clean-dist version test test-race test-coverage test-integration fmt lint verify

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)
BIN_NAME ?= http-services
BIN_DIR ?= bin

PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
DIST_DIR ?= dist
CGO_ENABLED ?= 0
PACKAGE_FILES ?= README.md config.yaml.example
PACKAGE_DIRS ?= static

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 构建生产版本（CROSS=1 启用跨平台打包，BIN_NAME=xxx 指定二进制名称）
	@if [ "$(CROSS)" = "1" ]; then \
		$(MAKE) build-cross; \
	else \
		$(MAKE) build-local; \
	fi

build-local: ## 本地构建
	@echo "构建 $(BIN_NAME)..."
	@mkdir -p "$(BIN_DIR)"
	@go build -trimpath -ldflags "$(LDFLAGS)" -o "$(BIN_DIR)/$(BIN_NAME)" .
	@echo "✓ 构建完成: $(BIN_DIR)/$(BIN_NAME)"

build-cross: clean-dist ## 跨平台构建并打包到 dist/
	@echo "开始跨平台构建与打包: $(PLATFORMS)"
	@mkdir -p "$(DIST_DIR)"
	@set -e; \
	for platform in $(PLATFORMS); do \
		os="$${platform%/*}"; arch="$${platform#*/}"; extension=""; \
		[ "$$os" = "windows" ] && extension=".exe"; \
		output_dir="$(DIST_DIR)/$(BIN_NAME)_$(VERSION)_$${os}_$${arch}"; \
		echo "- 构建 $$os/$$arch"; \
		mkdir -p "$$output_dir"; \
		GOOS="$$os" GOARCH="$$arch" CGO_ENABLED=$(CGO_ENABLED) \
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
	@echo "✓ 打包完成，产物位于 $(DIST_DIR)/"

build-check: ## 在临时目录验证构建，不保留产物
	@temp_dir="$$(mktemp -d "$${TMPDIR:-/tmp}/$(BIN_NAME)-build.XXXXXX")"; \
	trap 'rm -rf "$$temp_dir"' EXIT INT TERM; \
	go build -trimpath -ldflags "$(LDFLAGS)" -o "$$temp_dir/$(BIN_NAME)" .

run: build-local ## 构建并以生产模式运行
	@"$(BIN_DIR)/$(BIN_NAME)"

dev: build-local ## 构建并以开发模式运行
	@"$(BIN_DIR)/$(BIN_NAME)" --dev

migrate: build-local ## 构建并执行数据库迁移
	@"$(BIN_DIR)/$(BIN_NAME)" --migrate

clean: ## 清理本地构建文件
	@rm -rf "$(BIN_DIR)"

clean-dist: ## 清理跨平台打包产物
	@rm -rf "$(DIST_DIR)"

version: ## 显示版本信息
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

test: ## 使用 race detector 与随机顺序运行全量测试
	@go test -race -shuffle=on -count=1 ./...

test-race: test ## test 的兼容别名

test-coverage: ## 生成覆盖率报告
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out

test-integration: ## 运行可选外部依赖集成测试
	@go test -tags=integration -race -shuffle=on -count=1 ./db/msqldb ./db/rdb

fmt: ## 格式化 Go 代码
	@gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

lint: ## 运行 go vet
	@go vet ./...

verify: lint test build-check ## 验证代码、测试和构建且不保留构建产物
	@go mod verify

.DEFAULT_GOAL := help
