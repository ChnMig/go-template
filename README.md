# go-template

Golang project templates

## Not necessarily for everyone.

The goal of go-template is to improve productivity, not simplicity.😉

So there's going to be some third-party modules that are good enough😊, and of course, my have to make sure that they're good enough😋. 

## Download templates with gonew

These templates were designed to work and be downloaded with 
[gonew](https://pkg.go.dev/golang.org/x/tools/cmd/gonew).

## directory

### http-service

Suitable for use as a http-api service template.

## Configuration

The project uses a YAML configuration file for managing settings.

### Setup

1. Copy the example configuration file:

   ```bash
   cd http-service
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
