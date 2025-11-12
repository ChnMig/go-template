# 构建与打包规范（http-services）

本项目通过 `http-services/Makefile` 提供统一的本地构建、跨平台打包与验证流程。发布与交付一律优先使用 Makefile 目标，避免自定义脚本分散。

## 关键目标

- `make help`：查看可用命令与说明
- `make build`：生产构建（设置 `CROSS=1` 时转为跨平台打包）
- `make build-local`：仅在本机平台构建到 `bin/`
- `make build-cross`：跨平台构建并打包到 `dist/`
- `make run`：生产模式运行
- `make dev`：开发模式运行（带彩色日志）
- `make test`：运行单测并输出覆盖率与统计摘要
- `make fmt`：格式化代码
- `make lint`：`go vet` 静态检查
- `make verify`：格式化 → 检查 → 测试 的一键校验链
- `make clean`：清理 `bin/`
- `make clean-dist`：清理 `dist/`
- `make version`：打印版本信息（来自嵌入的编译变量）

## 版本信息注入

编译通过 `-ldflags` 注入三项元信息：
- `main.Version`：来自 `git describe --tags --always --dirty`（无标签时回退为 `dev`）
- `main.BuildTime`：UTC 构建时间 `YYYY-MM-DD_HH:MM:SS`
- `main.GitCommit`：短提交哈希（获取失败回退为 `unknown`）

业务代码可通过这三项变量对外展示版本。例如：`--version` 输出。

## 跨平台打包

- 默认平台矩阵（可通过 `PLATFORMS` 覆盖）：
  - `linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`
- 产物目录：`dist/`，文件名包含二进制名、版本、OS、ARCH
- 压缩格式：Windows 优先使用 `zip`（系统存在 `zip` 命令时），其余使用 `.tar.gz`
- 默认 `CGO_ENABLED=0`（可覆盖）
- 随包文件（存在才会复制）：`README.md`、`config.yaml.example`

示例：

```bash
# 全量平台打包（推荐发布流程）
make -C http-services build CROSS=1

# 指定平台集
make -C http-services build CROSS=1 \
  PLATFORMS="linux/amd64 linux/arm64 darwin/arm64 windows/amd64"

# 如需启用 CGO
make -C http-services build CROSS=1 CGO_ENABLED=1
```

## 使用约定（全局重要记忆）

- 构建与发版统一走 Makefile：`verify` 校验、`build`/`build-cross` 构建、`CROSS=1` 触发跨平台打包。
- 发布产物始终从 `dist/` 目录获取，包内至少包含可执行文件与 `config.yaml.example`。
- Windows 平台优先产出 `zip`，其余平台使用 `.tar.gz` 保持兼容。
- 新增随包文件时请通过 `PACKAGE_FILES` 变量扩展，避免手动复制。
- 在 CI/CD 中优先使用 `make -C http-services build CROSS=1`。
