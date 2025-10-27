# 命令行参数文档

## 概述
项目使用 kong 库来处理命令行参数解析，提供简洁的 CLI 接口。

## 可用参数

### --dev, -d
开发模式运行

**用法**:
```bash
./bin/http-services --dev
# 或
./bin/http-services -d
```

**效果**:
- 设置 `config.RunModel = "dev"`
- 日志输出更详细
- Gin 框架运行在开发模式

### --version, -v
显示版本信息

**用法**:
```bash
./bin/http-services --version
# 或
./bin/http-services -v
```

**输出示例**:
```
Version:    v1.0.0
Build Time: 2025-01-15_10:30:45
Git Commit: abc1234
```

### --help, -h
显示帮助信息

**用法**:
```bash
./bin/http-services --help
# 或
./bin/http-services -h
```

## 运行模式

项目支持两种运行模式：

### 1. 开发模式 (dev)
- 通过 `--dev` 参数或 `model=dev` 环境变量启用
- 日志输出详细
- Gin 框架显示调试信息

### 2. 生产模式 (release)
- 默认模式
- 日志输出简洁
- Gin 框架静默运行
- 性能优化

## 版本信息

版本信息在编译时通过 `-ldflags` 注入：

```bash
go build -ldflags "\
  -X main.Version=v1.0.0 \
  -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') \
  -X main.GitCommit=$(git rev-parse --short HEAD)" \
  -o bin/http-services .
```

推荐使用 Makefile 进行构建，它会自动处理版本信息。

## Makefile 使用

### 可用命令

```bash
make help      # 显示帮助信息
make build     # 构建生产版本
make run       # 构建并运行（生产模式）
make dev       # 构建并运行（开发模式）
make clean     # 清理构建文件
make version   # 显示版本信息
make test      # 运行测试
make tidy      # 整理依赖
```

### 常用场景

#### 开发时快速运行
```bash
make dev
```

#### 构建生产版本
```bash
make build
./bin/http-services
```

#### 使用环境变量控制运行模式
```bash
# 开发模式
model=dev ./bin/http-services

# 生产模式（默认）
./bin/http-services
```

## 实现细节

### main.go 结构

```go
var CLI struct {
    Dev     bool `help:"Run in development mode" short:"d"`
    Version bool `help:"Show version information" short:"v"`
}

var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)

func main() {
    // 1. 解析命令行参数
    ctx := kong.Parse(&CLI, ...)

    // 2. 处理版本信息
    if CLI.Version { ... }

    // 3. 加载配置
    config.LoadConfig()

    // 4. 设置运行模式
    if CLI.Dev {
        config.RunModel = config.RunModelDevValue
    } else {
        // 从环境变量检测
    }

    // 5. 初始化日志
    log.GetLogger()

    // 6. 启动服务
    r := api.InitApi()
    r.Run(...)
}
```

### 模式检测优先级

1. 命令行参数 `--dev` （最高优先级）
2. 环境变量 `model=dev`
3. 默认为生产模式 `release`

## 添加新参数

要添加新的命令行参数：

1. 在 `CLI` 结构体中添加字段：
```go
var CLI struct {
    Dev     bool   `help:"Run in development mode" short:"d"`
    Version bool   `help:"Show version information" short:"v"`
    Config  string `help:"Config file path" type:"path" default:"config.yaml"`
}
```

2. 在 `main()` 函数中处理参数：
```go
if CLI.Config != "" {
    // 使用指定的配置文件
}
```

## 参考资料

- [Kong 文档](https://github.com/alecthomas/kong)
- Kong 支持丰富的参数类型和验证
- 可以定义子命令、环境变量映射等高级功能
