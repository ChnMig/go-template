.PHONY: build run dev clean help test fmt lint verify

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

test: ## 运行测试并显示覆盖率（含中文提示与结果解析）
	@echo "运行测试（带覆盖率）..."
	@set -o pipefail; \
	OUT=$$(mktemp -t go-test-XXXXXX); \
	trap 'rm -f "$$OUT"' EXIT; \
	if go test -v -coverprofile=coverage.out -covermode=atomic ./... | tee "$$OUT"; then \
	  STATUS=0; \
	else \
	  STATUS=$$?; \
	fi; \
	PASS_PKGS=$$(grep -c '^ok[[:space:]]' "$$OUT" || true); \
	FAIL_PKGS=$$(grep -c '^FAIL[[:space:]]' "$$OUT" || true); \
	TOTAL_PKGS=$$((PASS_PKGS+FAIL_PKGS)); \
	PASS_TESTS=$$(grep -c '^--- PASS:' "$$OUT" || true); \
	FAIL_TESTS=$$(grep -c '^--- FAIL:' "$$OUT" || true); \
	SKIP_TESTS=$$(grep -c '^--- SKIP:' "$$OUT" || true); \
	if [ -f coverage.out ]; then \
	  TOTAL_COV=$$(go tool cover -func=coverage.out | awk '/^total:/ {print $$3}'); \
	else \
	  TOTAL_COV="N/A"; \
	fi; \
	echo "测试汇总：包 总数=$$TOTAL_PKGS 通过=$$PASS_PKGS 失败=$$FAIL_PKGS"; \
	echo "用例汇总：通过=$$PASS_TESTS 失败=$$FAIL_TESTS 跳过=$$SKIP_TESTS"; \
	echo "总覆盖率：$$TOTAL_COV"; \
	exit $$STATUS

fmt: ## 格式化代码
	@echo "格式化代码..."
	@gofmt -w $$(find . -name "*.go" -not -path "./vendor/*")
	@echo "✓ 格式化完成"

lint: ## 运行代码检查（go vet）
	@echo "运行 go vet..."
	@go vet ./...
	@echo "✓ 检查完成"

# 一键验证：先格式化，再静态检查，最后跑覆盖率测试
verify: ## 验证：fmt -> lint -> test
	@$(MAKE) fmt
	@$(MAKE) lint
	@$(MAKE) test

.DEFAULT_GOAL := help
