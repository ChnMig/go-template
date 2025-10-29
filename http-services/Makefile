.PHONY: build run dev clean help test test-cover fmt lint tidy init

# 版本信息
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建参数
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)
BINARY_NAME := http-services

# 工具与安装（本地不再安装 golangci-lint）

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 构建生产版本
	@echo "构建 $(BINARY_NAME)..."
	@go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .
	@echo "✓ 构建完成: bin/$(BINARY_NAME)"

run: build ## 构建并运行（生产模式）
	@echo "启动服务（生产模式）..."
	@./bin/$(BINARY_NAME)

dev: build ## 构建并运行（开发模式）
	@echo "启动服务（开发模式）..."
	@./bin/$(BINARY_NAME) --dev

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -f $(BINARY_NAME)
	@echo "✓ 清理完成"

version: ## 显示版本信息
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

test: ## 运行测试
	@go test -v ./...

test-cover: ## 运行测试并显示覆盖率
	@go test -cover ./...

fmt: ## 格式化代码
	@echo "格式化代码..."
	@gofmt -w $$(find . -name "*.go" -not -path "./vendor/*")
	@echo "✓ 格式化完成"

lint: ## 运行代码检查（go vet）
	@echo "运行 go vet..."
	@go vet ./...
	@echo "✓ 检查完成"



tidy: ## 整理依赖
	@go mod tidy
	@go mod vendor

# 初始化：整理依赖并 vendor
init: ## 初始化开发环境（整理并 vendor 依赖）
	@echo "初始化开发环境..."
	@$(MAKE) tidy
	@echo "✓ 初始化完成"

 

.DEFAULT_GOAL := help
