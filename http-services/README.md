# HTTP Services

基于 Gin 框架的 HTTP API 服务模板，提供了完整的项目结构、中间件支持、日志管理和配置管理功能。

## 项目特点

- ✅ **标准化项目结构** - 清晰的目录组织，易于维护和扩展
- ✅ **Viper 配置管理** - 支持 YAML 配置、环境变量覆盖和热重载
- ✅ **JWT 认证** - 灵活的 Token 签发和验证机制，支持自定义数据结构
- ✅ **限流中间件** - 支持基于 IP 和 Token 的灵活限流配置
- ✅ **日志管理** - 开发/生产模式自动切换，支持日志轮转
- ✅ **命令行支持** - 基于 Kong 的命令行参数解析
- ✅ **响应规范化** - 统一的 API 响应格式，符合 Google API 设计指南
- ✅ **跨域支持** - 内置 CORS 中间件
- ✅ **优雅关闭** - 支持信号监听和优雅退出，自动清理资源
- ✅ **健康检查** - 单一健康检查端点

## 目录结构

```
http-services/
├── api/                    # API 相关代码
│   ├── app/               # 业务处理（按版本与分组组织）
│   │   └── v1/
│   │       └── open/
│   │           └── health/    # 健康检查模块（开放）
│   ├── middleware/        # 中间件
│   │   ├── cross-domain.go   # 跨域处理
│   │   ├── jwt.go            # JWT 验证
│   │   ├── page.go           # 分页处理
│   │   ├── params.go         # 参数验证
│   │   └── rate-limit.go     # 限流中间件
│   ├── response/          # 响应处理
│   │   ├── code.go           # 状态码定义
│   │   └── format.go         # 响应格式化
│   └── router.go          # 路由配置
├── common/               # 跨服务共享代码（常量、DTO、公共逻辑）
├── services/             # 领域 Service 封装，承载核心业务流程
├── config/                # 配置管理
│   ├── config.go          # 配置变量定义
│   ├── load.go            # 配置加载
│   └── check.go           # 配置校验
├── db/                   # 数据库迁移、初始化脚本
├── utils/                 # 工具函数
│   ├── authentication/    # JWT 认证工具
│   ├── encryption/        # 加密工具（BCrypt）
│   ├── id/               # ID 生成器（Sonyflake）
│   ├── log/              # 日志管理
│   ├── path-tool/        # 路径工具
│   └── run-model/        # 运行模式检测
├── log/                   # 日志文件目录
├── static/               # 静态资源目录
├── bin/                  # 构建输出目录
├── dist/                 # 跨平台打包产物（make build-cross）
├── vendor/               # Go Modules 依赖镜像（vendor 模式）
├── config.yaml           # 配置文件（不提交到 Git）
├── config.yaml.example   # 配置文件示例
├── main.go               # 程序入口
├── Makefile              # 构建脚本
└── README.md             # 项目文档

```

## 快速开始

### 1. 准备配置文件

```bash
# 复制配置文件示例
cp config.yaml.example config.yaml

# 编辑配置文件，修改 JWT 密钥等敏感信息
vim config.yaml
```

### 2. 构建项目

```bash
# 查看所有可用命令
make help

# 构建项目
make build
```

### 2.1 跨平台打包

Makefile 内置跨平台构建与打包，产物位于 `dist/`，文件名包含版本、系统与架构。

基础用法：

```bash
# 一键跨平台构建与打包（Unix 平台 tar.gz，Windows 优先 zip）
make build CROSS=1
# 或显式使用
make build-cross
```

自定义平台矩阵（默认：`linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`）：

```bash
make build CROSS=1 \
  PLATFORMS="linux/amd64 linux/arm64 darwin/arm64 windows/amd64"
```

说明：
- 二进制内嵌版本信息：`Version`、`BuildTime`、`GitCommit`。
- 默认 `CGO_ENABLED=0`；如依赖 CGO 可覆盖该变量。
- 若存在将自动随包附带：`README.md`、`config.yaml.example`。
- Windows 平台优先使用 `zip`，其他平台使用 `.tar.gz`。

### 3. 运行服务

```bash
# 开发模式运行（日志输出到控制台，彩色格式）
make dev
# 或
./bin/http-services -d

# 生产模式运行（日志输出到文件，JSON 格式）
make run
# 或
./bin/http-services

# 查看版本信息
./bin/http-services -v
```

## 配置说明

项目使用 [Viper](https://github.com/spf13/viper) 进行配置管理，支持 YAML 配置文件、环境变量覆盖和配置热重载。

### 配置文件路径

配置文件 `config.yaml` 按以下优先级查找：
1. 当前工作目录
2. 程序所在目录
3. `/etc/http-services/` 目录

### config.yaml 完整配置

```yaml
server:
  port: 8080                      # 服务监听端口
  max_body_size: "10MB"           # 最大请求体大小
  max_header_bytes: 1048576       # 最大请求头大小（字节）
  shutdown_timeout: "10s"         # 优雅关闭超时时间
  read_timeout: "30s"             # 读取超时
  write_timeout: "30s"            # 写入超时
  idle_timeout: "120s"            # 空闲连接超时
  enable_rate_limit: false        # 是否启用全局限流
  global_rate_limit: 100          # 全局限流速率（每秒请求数）
  global_rate_burst: 200          # 全局限流突发数

jwt:
  key: "YOUR_SECRET_KEY"          # JWT 签名密钥（至少 32 字符，必须修改！）
  expiration: "12h"               # Token 过期时间（如：12h, 24h, 30m）

log:
  max_size: 50                    # 单个日志文件最大大小（MB）
  max_backups: 3                  # 保留的旧日志文件最大数量
  max_age: 30                     # 保留旧日志文件的最大天数
```

### 环境变量覆盖

所有配置项都可以通过环境变量覆盖，使用 `HTTP_SERVICES_` 前缀，配置路径用下划线分隔：

```bash
# 覆盖服务端口
export HTTP_SERVICES_SERVER_PORT=9090

# 覆盖 JWT 密钥
export HTTP_SERVICES_JWT_KEY="your-production-secret-key"

# 覆盖超时配置
export HTTP_SERVICES_SERVER_READ_TIMEOUT="60s"

# 启用全局限流
export HTTP_SERVICES_SERVER_ENABLE_RATE_LIMIT=true

# 覆盖日志配置
export HTTP_SERVICES_LOG_MAX_SIZE=100
export HTTP_SERVICES_LOG_MAX_BACKUPS=5
export HTTP_SERVICES_LOG_MAX_AGE=60

# 运行服务
./bin/http-services
```

**环境变量命名规则：**
- 前缀：`HTTP_SERVICES_`
- 嵌套路径：用下划线 `_` 替代点 `.`
- 示例：`server.port` → `HTTP_SERVICES_SERVER_PORT`

### 配置热重载

服务支持配置热重载功能。修改 `config.yaml` 后，服务会自动检测并重新加载配置，无需重启。

**注意：** 部分配置（如端口、超时等）需要重启服务才能生效，但大部分配置可以热重载。

### Docker 环境变量示例

```dockerfile
# Dockerfile
FROM alpine:latest
WORKDIR /app
COPY bin/http-services .
EXPOSE 8080
CMD ["./http-services"]
```

```bash
# 使用环境变量运行容器
docker run -d \
  -e HTTP_SERVICES_SERVER_PORT=8080 \
  -e HTTP_SERVICES_JWT_KEY="production-secret-key-min-32-chars" \
  -e HTTP_SERVICES_JWT_EXPIRATION="24h" \
  -e HTTP_SERVICES_SERVER_ENABLE_RATE_LIMIT=true \
  -p 8080:8080 \
  your-image:latest
```

### Kubernetes ConfigMap 示例

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: http-services-config
data:
  HTTP_SERVICES_SERVER_PORT: "8080"
  HTTP_SERVICES_JWT_EXPIRATION: "24h"
  HTTP_SERVICES_SERVER_ENABLE_RATE_LIMIT: "true"
---
apiVersion: v1
kind: Secret
metadata:
  name: http-services-secret
type: Opaque
stringData:
  HTTP_SERVICES_JWT_KEY: "your-production-secret-key-min-32-chars"
```

## 开发规范

### 1. API 响应格式

所有 API 响应遵循统一格式，符合 [Google API 设计指南](https://google-cloud.gitbook.io/api-design-guide/errors)：

```json
{
  "code": 200,
  "status": "OK",
  "description": "No error",
  "message": "可选的具体错误信息",
  "trace_id": "4b818aea2976c3d0a711e99c06ac3192",
  "timestamp": 1698765432,
  "detail": {},
  "total": 100
}
```

**字段说明：**
- `code`: HTTP 状态码
- `status`: 状态名称（如 OK, INVALID_ARGUMENT）
- `description`: 标准错误描述（符合 Google API 规范）
- `message`: 具体的业务错误信息（可选）
- `trace_id`: 请求追踪 ID（由网关或中间件注入 `X-Request-ID`）
- `timestamp`: 时间戳
- `detail`: 详细数据（可选）
- `total`: 分页总数（可选）

### 2. 错误处理

```go
// 返回错误
response.ReturnError(c, response.INVALID_ARGUMENT, "用户名不能为空")

// 返回成功
response.ReturnOk(c, data)

// 返回分页数据
response.ReturnOkWithTotal(c, 100, list)
```

**预定义错误码：**
- `OK` (200): 成功
- `INVALID_ARGUMENT` (400): 参数错误
- `UNAUTHENTICATED` (401): 未认证
- `PERMISSION_DENIED` (403): 权限不足
- `NOT_FOUND` (404): 资源不存在
- `ALREADY_EXISTS` (409): 资源已存在
- `RESOURCE_EXHAUSTED` (429): 超过限流
- `INTERNAL` (500): 内部错误
- `UNAVAILABLE` (503): 服务不可用

### 3. 中间件使用

#### JWT 认证

```go
// 建议在 v1 层为 private 分组统一附加认证（示意，不包含具体接口）
func RegisterRoutes(v1 *gin.RouterGroup) {
    // 私有分组聚合：按需开启 JWT 校验
    // privateGroup := v1.Group("/private", middleware.TokenVerify)
    // private.RegisterRoutes(privateGroup)
}
```

#### 限流配置

```go
// 在 v1 层或 open 聚合层就近添加限流
func RegisterRoutes(v1 *gin.RouterGroup) {
    // IP 限流（每秒10个请求，突发20个）
    openGroup := v1.Group("/open", middleware.IPRateLimit(10, 20))
    open.RegisterRoutes(openGroup)
}

// 预定义限流级别（示例：在 open 组下的具体接口上使用）
openGroup.POST("/sensitive", middleware.StrictRateLimit(), handler)   // 严格（5/秒）
openGroup.GET("/normal", middleware.ModerateRateLimit(), handler)     // 中等（50/秒）
openGroup.GET("/read", middleware.RelaxedRateLimit(), handler)        // 宽松（100/秒）

// 自定义限流 Key（示例：在 open 组的接口上）
middleware.RateLimitWithOptions(middleware.RateLimitOptions{
    Rate: 50,
    Burst: 100,
    KeyFunc: func(c *gin.Context) string {
        return c.GetHeader("X-API-Key")
    },
    Message: "API rate limit exceeded",
})
```

**限流参数说明：**
- `Rate`: 每秒请求数（令牌生成速率）
- `Burst`: 突发请求数（令牌桶容量）
- 建议 `Burst >= Rate`，通常设置为 `Burst = 2 × Rate`

#### 分页处理

```go
// 获取分页参数
page := middleware.GetPage(c)      // 默认 1
pageSize := middleware.GetPageSize(c)  // 默认 20

// 取消分页（获取全部数据）
// 请求参数：page=-1 或 pageSize=-1
```

### 4. 路由组织（分层）

```go
// 顶层：api/router.go（仅初始化与挂载 /api，业务路由下沉到 app 层）
func InitApi() *gin.Engine {
    router := gin.Default()
    // ... 全局中间件
    apiGroup := router.Group("/api")
    app.RegisterRoutes(apiGroup)
    return router
}

// app 层：api/app/router.go（在 /api 下挂载各版本）
func RegisterRoutes(api *gin.RouterGroup) {
    v1Group := api.Group("/v1")
    v1.RegisterRoutes(v1Group)
}

// v1 层：api/app/v1/router.go（在 /api/v1 下挂载 open / private 等分组）
func RegisterRoutes(v1 *gin.RouterGroup) {
    openGroup := v1.Group("/open")
    open.RegisterRoutes(openGroup)

    privateGroup := v1.Group("/private")
    private.RegisterRoutes(privateGroup)
}

// open 聚合层：api/app/v1/open/router.go（在 /api/v1/open 下注册各模块公开路由）
func RegisterRoutes(open *gin.RouterGroup) {
    health.RegisterOpenRoutes(open) // /api/v1/open/health
}
```

### 5. JWT 使用

JWT 使用 `map[string]interface{}` 存储自定义数据，支持灵活的数据结构。

```go
// 签发 Token - 使用 map 存储任意数据结构
userData := map[string]interface{}{
    "user_id":  "12345",
    "username": "admin",
    "role":     "admin",
    "email":    "admin@example.com",
}
token, err := authentication.JWTIssue(userData)
if err != nil {
    response.ReturnError(c, response.INTERNAL, "Token 生成失败")
    return
}

// 验证 Token（中间件自动处理）
// 在 handler 中获取 JWT 数据
jwtData, exists := c.Get("jwtData")
if !exists {
    response.ReturnError(c, response.UNAUTHENTICATED, "未找到认证信息")
    return
}

// 类型断言获取 map 数据
data, ok := jwtData.(map[string]interface{})
if !ok {
    response.ReturnError(c, response.INTERNAL, "认证数据格式错误")
    return
}

// 获取具体字段
userID := data["user_id"].(string)
username := data["username"].(string)
```

**JWT 最佳实践：**

- Token 中只存储必要的用户标识信息，不要存储敏感数据
- 根据实际业务需求设计 Token 数据结构
- 建议在项目中定义统一的 Token 数据结构规范

## 日志管理

### 开发模式（`-d` 参数）

- 日志输出到控制台
- 彩色格式，易于阅读
- Debug 级别日志

### 生产模式（默认）

- 日志输出到文件 `log/http-services.log`
- JSON 格式，便于日志分析
- Info 级别日志
- 自动轮转（可通过配置文件或环境变量自定义）：
  - 单文件最大大小：默认 50MB（可配置 `log.max_size`）
  - 最多保留备份数：默认 3 个（可配置 `log.max_backups`）
  - 保留天数：默认 30 天（可配置 `log.max_age`）

### 自定义日志配置

通过配置文件：

```yaml
log:
  max_size: 100      # 单个日志文件最大 100MB
  max_backups: 5     # 保留 5 个备份文件
  max_age: 60        # 保留 60 天
```

通过环境变量：

```bash
export HTTP_SERVICES_LOG_MAX_SIZE=100
export HTTP_SERVICES_LOG_MAX_BACKUPS=5
export HTTP_SERVICES_LOG_MAX_AGE=60
```

### 使用示例

```go
import "go.uber.org/zap"

// 记录日志
zap.L().Info("业务事件", zap.String("action", "process"))
zap.L().Error("操作失败", zap.Error(err))
zap.L().Debug("调试信息", zap.Any("data", data))
```

## 命令行参数

```bash
# 开发模式
./bin/http-services -d
./bin/http-services --dev

# 查看版本
./bin/http-services -v
./bin/http-services --version

# 查看帮助
./bin/http-services -h
./bin/http-services --help
```

## 常用 Make 命令

```bash
make help      # 显示所有可用命令
make build     # 构建项目
make run       # 构建并运行（生产模式）
make dev       # 构建并运行（开发模式）
make clean     # 清理构建文件
make version   # 显示版本信息
make test      # 运行测试
```

### 测试说明

- 运行 `make test`：
  - 输出中文汇总（包数量、通过/失败、用例通过/失败/跳过）
  - 生成覆盖率文件 `coverage.out` 并打印总覆盖率
- 如需仅查看覆盖率，也可直接使用 `go tool cover -func=coverage.out`

## API 示例

### 健康检查

```bash
# 健康检查（包含 ready 与 uptime 信息）
curl http://localhost:8080/api/v1/open/health

# 响应：{"status":"ok","ready":true,"uptime":"1h30m20s"}
```

### 访问受保护接口

当前模板未内置示例私有接口。可按需新增 `api/app/v1/private/<module>`，并在 `api/app/v1/private/router.go` 中注册。

### 测试限流

```bash
# 快速发送多个请求测试限流
for i in {1..30}; do
  curl http://localhost:8080/api/v1/open/health &
done
wait
```

## 部署建议

### 1. 使用 systemd 管理服务

创建 `/etc/systemd/system/http-services.service`：

```ini
[Unit]
Description=HTTP Services
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/http-services
ExecStart=/opt/http-services/bin/http-services
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### 2. 使用 Nginx 反向代理

```nginx
upstream http_services {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://http_services;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 3. Docker 部署

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -ldflags "-w -s" -o bin/http-services .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/http-services .
COPY --from=builder /app/config.yaml.example ./config.yaml.example
EXPOSE 8080
CMD ["./http-services"]
```

### 4. Kubernetes 部署示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: http-services
spec:
  replicas: 3
  selector:
    matchLabels:
      app: http-services
  template:
    metadata:
      labels:
        app: http-services
    spec:
      containers:
      - name: http-services
        image: your-registry/http-services:latest
        ports:
        - containerPort: 8080
        env:
        - name: HTTP_SERVICES_SERVER_PORT
          value: "8080"
        - name: HTTP_SERVICES_JWT_KEY
          valueFrom:
            secretKeyRef:
              name: http-services-secret
              key: HTTP_SERVICES_JWT_KEY
        - name: HTTP_SERVICES_JWT_EXPIRATION
          value: "24h"
        - name: HTTP_SERVICES_SERVER_ENABLE_RATE_LIMIT
          value: "true"
        # 健康检查配置（合并为单一端点 /api/v1/open/health）
        livenessProbe:
          httpGet:
            path: /api/v1/open/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /api/v1/open/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: http-services
spec:
  type: ClusterIP
  selector:
    app: http-services
  ports:
  - port: 8080
    targetPort: 8080
```

## 性能建议

1. **生产环境使用 Release 模式** - 日志写入文件，性能更好
2. **合理设置限流参数** - 根据服务器性能和业务需求调整
3. **启用 Gzip 压缩** - 减少网络传输
4. **使用连接池** - 数据库、Redis 等外部服务
5. **监控日志文件大小** - 定期清理旧日志

## 依赖项

主要依赖：

- `gin-gonic/gin` - Web 框架
- `golang-jwt/jwt` - JWT 认证
- `uber-go/zap` - 日志库
- `spf13/viper` - 配置管理
- `golang.org/x/time/rate` - 限流器
- `alecthomas/kong` - 命令行解析
- `sony/sonyflake` - 分布式 ID 生成
- `natefinch/lumberjack` - 日志轮转

## 许可证

[MIT License]

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题，请提交 Issue 或联系维护者。
