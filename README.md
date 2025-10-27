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

## Configuration

The project uses a YAML configuration file for managing settings.

### Setup

1. Copy the example configuration file:

   ```bash
   cd http-services
   cp config.yaml.example config.yaml
   ```

2. Edit `config.yaml` and update the values, especially:
   - `jwt.key`: Change this to a secure random string
   - `jwt.expiration`: Set token expiration time (e.g., "12h", "24h", "30m")

### Configuration File Structure

```yaml
server:
  port: 8080

jwt:
  key: "YOUR_SECRET_KEY_HERE"
  expiration: "12h"
```

**Important**: The `config.yaml` file is ignored by git to prevent sensitive data from being committed. Always use `config.yaml.example` as a template.

## Features

### Core Components

- **JWT Authentication**: Supports both single-app and multi-tenant authentication
- **CORS**: Cross-Origin Resource Sharing middleware
- **Password Encryption**: BCrypt-based secure password hashing
- **Pagination**: Built-in pagination support with configurable defaults

### Middleware

- `TokenVerify`: Basic JWT authentication for single-app
- `MultiTenantTokenVerify`: JWT authentication with tenant support
- `SuperAdminVerify`: Super admin permission verification
- `TenantAdminVerify`: Tenant admin permission verification
- `CorssDomainHandler`: CORS middleware
- `CheckParam`: Request parameter validation

### Utilities

- **Authentication** (`util/authentication`):
  - JWT token generation and parsing
  - HS256 signing and verification
  - Multi-tenant JWT support via `api/auth`

- **Encryption** (`util/encryption`):
  - BCrypt password hashing
  - Password verification

- **ID Generation** (`util/id`):
  - MD5-based unique ID generation

### Dependencies

Key dependencies include:

- `github.com/gin-gonic/gin` - Web framework
- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/goccy/go-yaml` - YAML parser
- `golang.org/x/crypto/bcrypt` - Password encryption
- `go.uber.org/zap` - Structured logging

## Project Structure

```text
http-services/
├── api/
│   ├── app/              # API handlers
│   ├── auth/             # Multi-tenant authentication
│   ├── middleware/       # Middleware components
│   └── response/         # Response formatters
├── config/               # Configuration management
├── util/                 # Utility packages
│   ├── authentication/   # JWT utilities
│   ├── encryption/       # Password encryption
│   ├── id/              # ID generation
│   ├── log/             # Logging
│   └── path-tool/       # Path utilities
└── main.go              # Application entry point
```

## Development Notes

This template is based on [art-design-pro-edge-go-server](https://github.com/ChnMig/art-design-pro-edge-go-server) and includes regularly synchronized updates to core components.
