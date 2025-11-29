# go-template

Golang project templates

## Not necessarily for everyone.

The goal of go-template is to improve productivity, not simplicity.😉

So there's going to be some third-party modules that are good enough😊, and of course, my have to make sure that they're good enough😋. 

## Download templates with gonew

These templates were designed to work and be downloaded with 
[gonew](https://pkg.go.dev/golang.org/x/tools/cmd/gonew).

## directory

### http-services

Suitable for use as a http-api service template.

## Quick Start

### Setup

1. Copy the example configuration file:

   ```bash
   cd http-services
   cp config.yaml.example config.yaml
   ```

2. Edit `config.yaml` and update the values, especially:
   - `jwt.key`: **必须修改为至少32字符的强密钥** (服务启动时会进行安全检查)
   - `jwt.expiration`: Set token expiration time (e.g., "12h", "24h", "30m")

3. Build and run:

   ```bash
   # 显示帮助
   make help

   # 构建
   make build

   # 运行（生产模式）
   make run

   # 运行（开发模式）
   make dev
   ```

### Cross-Platform Packaging

Use the Makefile to package binaries for multiple platforms. Artifacts are placed under `dist/` with version, OS and ARCH in the filename.

Basic usage:

```bash
cd http-services

# Cross-compile and package (tar.gz on Unix, zip on Windows)
make build CROSS=1
# or explicitly
make build-cross
```

Customize target platforms via `PLATFORMS` (defaults: `linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64`):

```bash
make build CROSS=1 \
  PLATFORMS="linux/amd64 linux/arm64 darwin/arm64 windows/amd64"
```

Notes:
- Binaries embed version metadata: `Version`, `BuildTime`, `GitCommit`.
- `CGO_ENABLED=0` by default; override if you depend on CGO.
- Extra files included alongside binaries if present: `README.md`, `config.yaml.example`.
- Windows packages are zipped when `zip` is available; others use `.tar.gz`.

### Command Line Options

```bash
# 开发模式
./bin/http-services --dev

# 显示版本信息
./bin/http-services --version

# 显示帮助
./bin/http-services --help
```

## Configuration

The project uses a YAML configuration file for managing settings.

### Configuration File Structure

```yaml
server:
  port: 8080

jwt:
  key: "YOUR_SECRET_KEY_HERE"
  expiration: "12h"
```

**Important**: The `config.yaml` file is ignored by git to prevent sensitive data from being committed. Always use `config.yaml.example` as a template.

### Configuration Reload & Restart

当前模板虽然使用 Viper 支持监控 `config.yaml` 变更，但大部分配置（例如 `server.port`、TLS/ACME 开关、全局限流、超时配置等）只在进程启动时读取并应用，运行中的 HTTP 服务器和路由不会自动根据新配置重新构建。

因此在生产环境中，**修改配置后建议始终重启服务进程**，以确保所有配置项都按预期生效；不要依赖“热更新配置”来切换是否启用 TLS、修改端口或调整全局限流策略。

## Features

### Core Components

- **JWT Authentication**: Standard JWT authentication with security validation
- **CORS**: Cross-Origin Resource Sharing middleware
- **Password Encryption**: BCrypt-based secure password hashing
- **Pagination**: Built-in pagination support with configurable defaults
- **Graceful Shutdown**: Proper HTTP server graceful shutdown with 10s timeout
- **Health Checks**: `/health` and `/ready` endpoints for monitoring

### Middleware

- `RequestID`: Request ID tracking for distributed tracing
- `SecurityHeaders`: Security response headers (X-Content-Type-Options, X-Frame-Options, etc.)
- `BodySizeLimit`: Request body size limit (default 10MB)
- `TokenVerify`: JWT authentication middleware
- `CorssDomainHandler`: CORS middleware
- `IPRateLimit`: IP-based rate limiting with token bucket algorithm
- `TokenRateLimit`: Token-based rate limiting for authenticated users

### Utilities

- **Authentication** (`util/authentication`):
  - JWT token generation and parsing
  - HS256 signing and verification
  - Standard claims support

- **Encryption** (`util/encryption`):
  - BCrypt password hashing
  - Password verification

- **ID Generation** (`util/id`):
  - Sonyflake-based distributed unique ID generation
  - MD5-based unique ID generation

### Dependencies

Key dependencies include:

- `github.com/gin-gonic/gin` - Web framework
- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/goccy/go-yaml` - YAML parser
- `github.com/alecthomas/kong` - Command line parser
- `golang.org/x/crypto/bcrypt` - Password encryption
- `go.uber.org/zap` - Structured logging

## Project Structure

```text
http-services/
├── api/
│   ├── app/              # API handlers
│   │   ├── example/      # Example API endpoints
│   │   └── health/       # Health check endpoints
│   ├── middleware/       # Middleware components
│   └── response/         # Response formatters
├── config/               # Configuration management
├── utils/                # Utility packages
│   ├── authentication/   # JWT utilities (with tests)
│   ├── encryption/       # Password encryption (with tests)
│   ├── id/              # ID generation (with tests)
│   ├── log/             # Logging
│   ├── path-tool/       # Path utilities
│   └── run-model/       # Runtime mode utilities
├── db/                   # Database layer (placeholder)
├── services/             # Business logic layer (placeholder)
├── common/               # Common utilities (placeholder)
└── main.go              # Application entry point
```

## Testing

Run all tests:

```bash
make test
```

Run tests with coverage:

```bash
go test -cover ./...
```

Current test coverage includes:
- JWT authentication and token handling
- BCrypt password hashing and verification
- ID generation (Sonyflake + MD5)

## Security Features

- **JWT Key Validation**: Server refuses to start with weak or default JWT keys
- **Request Body Size Limit**: Prevents DoS attacks from large payloads
- **Security Headers**: Automatic security headers for all responses
- **Rate Limiting**: Configurable rate limiting per IP or authenticated user
- **Request ID Tracking**: Distributed tracing support

## Development Notes

This template is based on [art-design-pro-edge-go-server](https://github.com/ChnMig/art-design-pro-edge-go-server) and includes regularly synchronized updates to core components.
