# CI/CD 配置说明

本项目使用 GitHub Actions 进行持续集成，已针对中国国内网络环境进行优化。

## 配置文件

- `.github/workflows/test.yml` - CI/CD 主配置文件
- `http-services/.golangci.yml` - golangci-lint 代码检查配置

## CI 流程

### 1. 测试作业 (Test Job)

**触发条件：** push 到 main/develop 分支，或创建 PR

**执行步骤：**
- 支持多 Go 版本测试 (1.23.x, 1.24.x, 1.25.x)
- 配置国内 Go 模块代理加速（goproxy.cn, 阿里云镜像）
- 缓存 Go 模块和构建缓存
- 运行 `go vet` 静态分析
- 运行单元测试和集成测试（带 race 检测）
- 生成测试覆盖率报告
- 上传覆盖率到 Codecov（需配置 CODECOV_TOKEN）
- 覆盖率低于 50% 时发出警告

### 2. 代码检查作业 (Lint Job)

**执行步骤：**
- 使用 golangci-lint 进行代码质量检查
- 检查代码风格、潜在问题、代码重复等
- 配置国内镜像加速

### 3. 构建作业 (Build Job)

**执行步骤：**
- 构建可执行文件
- 注入版本信息（版本号、构建时间、Git commit）
- 上传构建产物（保留 7 天）

### 4. 安全扫描作业 (Security Job)

**执行步骤：**
- 使用 Gosec 扫描安全漏洞
- 生成 SARIF 报告
- 上传到 GitHub Security（可在 Security 标签页查看）

## 国内网络优化

为了解决国内访问 GitHub 和 Go 官方资源的网络问题，CI 配置中做了以下优化：

### Go 模块代理配置

```bash
GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
GOSUMDB=sum.golang.google.cn
```

**使用的镜像源：**
1. **goproxy.cn** - 七牛云维护，国内访问最快
2. **mirrors.aliyun.com/goproxy** - 阿里云镜像，稳定可靠
3. **goproxy.io** - 全球 CDN，备用选择
4. **direct** - 直连官方源（最后备选）

### 缓存策略

使用 GitHub Actions 缓存机制缓存：
- Go 模块目录 (`~/go/pkg/mod`)
- Go 构建缓存 (`~/.cache/go-build`)

缓存 key 基于 Go 版本和 `go.sum` 文件，确保依赖变化时重新下载。

## Codecov 配置（可选）

如果要启用代码覆盖率上传到 Codecov：

1. 访问 [Codecov](https://codecov.io/) 并登录
2. 添加你的 GitHub 仓库
3. 获取 Codecov Token
4. 在 GitHub 仓库设置中添加 Secret：
   - 名称：`CODECOV_TOKEN`
   - 值：你的 Codecov Token

如果不配置 `CODECOV_TOKEN`，覆盖率上传步骤会失败但不会影响 CI 通过。

## 本地验证

在提交代码前，建议先本地验证：

### 运行测试
```bash
cd http-services
go test -v -race ./...
```

### 运行代码检查
```bash
# 安装 golangci-lint（如果还没安装）
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 运行检查
cd http-services
golangci-lint run
```

### 构建项目
```bash
cd http-services
make build
```

### 运行安全扫描（可选）
```bash
# 安装 gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# 运行扫描
cd http-services
gosec ./...
```

## 故障排查

### 依赖下载失败

如果遇到 Go 模块下载失败：

1. **检查代理配置**
   ```bash
   go env GOPROXY
   go env GOSUMDB
   ```

2. **本地测试代理**
   ```bash
   export GOPROXY=https://goproxy.cn,direct
   go mod download
   ```

3. **清理模块缓存**
   ```bash
   go clean -modcache
   go mod download
   ```

### golangci-lint 超时

如果 golangci-lint 运行超时，可以：

1. 增加超时时间（已设置为 5 分钟）
2. 减少启用的 linter 数量
3. 在 `.golangci.yml` 中添加更多排除规则

### 测试失败

1. **检查本地环境**
   ```bash
   go test -v ./...
   ```

2. **检查 race 检测**
   ```bash
   go test -race ./...
   ```

3. **查看详细日志**
   在 GitHub Actions 页面查看完整的测试输出

## 配置自定义

### 修改 Go 版本

在 `.github/workflows/test.yml` 中修改：
```yaml
strategy:
  matrix:
    go-version: [ '1.23.x', '1.24.x', '1.25.x' ]  # 添加或删除版本
```

### 修改触发条件

```yaml
on:
  push:
    branches: [ main, develop, feature/* ]  # 添加分支
  pull_request:
    branches: [ main ]
```

### 调整 linter 规则

编辑 `http-services/.golangci.yml`，启用或禁用特定的 linter。

## 状态徽章

在 README.md 中添加状态徽章：

```markdown
[![Test](https://github.com/username/repo/workflows/Test/badge.svg)](https://github.com/username/repo/actions)
[![codecov](https://codecov.io/gh/username/repo/branch/main/graph/badge.svg)](https://codecov.io/gh/username/repo)
```

## 参考资料

- [GitHub Actions 文档](https://docs.github.com/cn/actions)
- [golangci-lint 文档](https://golangci-lint.run/)
- [Gosec 文档](https://github.com/securego/gosec)
- [Codecov 文档](https://docs.codecov.com/)
- [Goproxy.cn 文档](https://goproxy.cn/)
