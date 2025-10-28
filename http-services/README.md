# HTTP Services

基于 Gin 框架的 HTTP API 服务模板，提供了完整的项目结构、中间件支持、日志管理和配置管理功能。

## 项目特点

- ✅ **标准化项目结构** - 清晰的目录组织，易于维护和扩展
- ✅ **配置文件管理** - 基于 YAML 的配置文件，支持灵活配置
- ✅ **JWT 认证** - 完整的 Token 签发和验证机制
- ✅ **限流中间件** - 支持基于 IP 和 Token 的灵活限流配置
- ✅ **日志管理** - 开发/生产模式自动切换，支持日志轮转
- ✅ **命令行支持** - 基于 Kong 的命令行参数解析
- ✅ **响应规范化** - 统一的 API 响应格式，符合 Google API 设计指南
- ✅ **跨域支持** - 内置 CORS 中间件
- ✅ **优雅关闭** - 支持信号监听和优雅退出

## 目录结构

```
http-services/
├── api/                    # API 相关代码
│   ├── app/               # 业务处理
│   │   └── example/       # 示例业务模块
│   ├── middleware/        # 中间件
│   │   ├── cross-domain.go   # 跨域处理
│   │   ├── jwt.go            # JWT 验证
│   │   ├── page.go           # 分页处理
│   │   ├── params.go         # 参数验证
│   │   └── rate_limit.go     # 限流中间件
│   ├── response/          # 响应处理
│   │   ├── code.go           # 状态码定义
│   │   └── format.go         # 响应格式化
│   └── router.go          # 路由配置
├── config/                # 配置管理
│   ├── config.go          # 配置变量定义
│   ├── load.go            # 配置加载
│   └── check.go           # 配置校验
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

# 整理依赖
make tidy
```

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

### config.yaml

```yaml
server:
  port: 8080              # 服务监听端口

jwt:
  key: "YOUR_SECRET_KEY"  # JWT 签名密钥（必须修改！）
  expiration: "12h"       # Token 过期时间（如：12h, 24h, 30m）
```

### 环境变量

- `model`: 运行模式，可选值 `dev` 或 `release`（命令行参数优先级更高）

## 开发规范

### 1. API 响应格式

所有 API 响应遵循统一格式，符合 [Google API 设计指南](https://google-cloud.gitbook.io/api-design-guide/errors)：

```json
{
  "code": 200,
  "status": "OK",
  "description": "No error",
  "message": "可选的具体错误信息",
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
// 需要认证的路由组
privateRouter := router.Group("/api/private", middleware.TokenVerify)
{
    privateRouter.GET("/user", handler)
}
```

#### 限流配置

```go
// IP 限流（每秒10个请求，突发20个）
router.Group("/api/public", middleware.IPRateLimit(10, 20))

// Token 限流（基于用户，每秒100个请求，突发200个）
router.Group("/api/user",
    middleware.TokenVerify,
    middleware.TokenRateLimit(100, 200))

// 预定义限流级别
router.POST("/api/sensitive", middleware.StrictRateLimit(), handler)    // 严格（5/秒）
router.GET("/api/normal", middleware.ModerateRateLimit(), handler)       // 中等（50/秒）
router.GET("/api/read", middleware.RelaxedRateLimit(), handler)          // 宽松（100/秒）

// 自定义限流 Key
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

### 4. 路由组织

```go
// api/router.go
func openRouter(router *gin.RouterGroup) {
    // 开放接口：无需认证，严格限流
    exampleRouter := router.Group("/open/example", middleware.IPRateLimit(10, 20))
    {
        exampleRouter.GET("/pong", example.Pong)
        exampleRouter.POST("/token", example.CreateToken)
    }
}

func privateRouter(router *gin.RouterGroup) {
    // 私有接口：需要认证，宽松限流
    exampleRouter := router.Group("/private/example",
        middleware.TokenVerify,
        middleware.TokenRateLimit(100, 200))
    {
        exampleRouter.GET("/pong", example.Pong)
    }
}
```

### 5. JWT 使用

```go
// 签发 Token
token, err := authentication.JWTIssue("user_id")

// 验证 Token（中间件自动处理）
// 在 handler 中获取 JWT 数据
jwtData, exists := c.Get("jwtData")
```

## 日志管理

### 开发模式（`-d` 参数）
- 日志输出到控制台
- 彩色格式，易于阅读
- Debug 级别日志

### 生产模式（默认）
- 日志输出到文件 `log/http-services.log`
- JSON 格式，便于日志分析
- Info 级别日志
- 自动轮转：
  - 单文件最大 50MB
  - 最多保留 3 个备份文件
  - 保留 30 天

### 使用示例

```go
import "go.uber.org/zap"

// 记录日志
zap.L().Info("用户登录", zap.String("user", "admin"))
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
make tidy      # 整理依赖
```

## API 示例

### 获取 Token

```bash
curl -X POST http://localhost:8080/api/v1/open/example/token \
  -H "Content-Type: application/json" \
  -d '{"user":"admin","password":"pwd"}'
```

### 访问受保护接口

```bash
curl -X GET http://localhost:8080/api/v1/private/example/pong \
  -H "token: YOUR_JWT_TOKEN"
```

### 测试限流

```bash
# 快速发送多个请求测试限流
for i in {1..30}; do
  curl http://localhost:8080/api/v1/open/example/pong &
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
- `golang.org/x/time/rate` - 限流器
- `alecthomas/kong` - 命令行解析
- `goccy/go-yaml` - YAML 配置解析
- `sony/sonyflake` - 分布式 ID 生成

## 许可证

[MIT License]

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题，请提交 Issue 或联系维护者。
