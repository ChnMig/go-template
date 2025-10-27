# 后端配置管理架构

## 概述
项目采用 YAML 配置文件方式管理配置，参考 art-design-pro-edge-go-server 项目实现。

## 配置文件结构

### 位置
- 配置文件：`http-service/config.yaml`（不提交到 git）
- 示例文件：`http-service/config.yaml.example`（提交到 git）

### 格式
```yaml
server:
  port: 8080

jwt:
  key: "密钥"
  expiration: "12h"
```

## 代码架构

### 核心文件
1. **config/config.go** - 配置变量定义
   - 静态配置：端口、路径、日志等不需要修改的配置
   - 动态配置：从 YAML 加载的配置变量（如 JWTKey, JWTExpiration）

2. **config/load.go** - 配置加载逻辑
   - `YamlConfig` 结构体：映射 YAML 文件结构
   - `LoadConfig()` 函数：读取并解析配置文件，应用到全局变量

3. **config/check.go** - 配置校验
   - `CheckConfig()` 函数：校验必需配置项是否存在
   - 缺失配置会导致程序 fatal 退出

4. **main.go** - 配置初始化流程
   - 加载配置 → 初始化 logger → 校验配置 → 启动服务

## 使用规范

### 添加新配置项
1. 在 `config.yaml` 和 `config.yaml.example` 中添加配置项
2. 在 `config/load.go` 的 `YamlConfig` 结构体中添加对应字段
3. 在 `config/config.go` 中定义全局变量
4. 在 `config/load.go` 的 `LoadConfig()` 函数中添加加载逻辑
5. 如果是必需配置，在 `config/check.go` 中添加校验

### 访问配置
在代码中直接使用 `config.XXX` 访问配置变量，例如：
```go
import "go-services/config"

func example() {
    key := config.JWTKey
    exp := config.JWTExpiration
}
```

## 依赖
- `github.com/goccy/go-yaml` - YAML 解析库

## 注意事项
- 配置文件包含敏感信息，已通过 `.gitignore` 排除，避免提交到版本控制
- 修改配置需要重启服务才能生效
- 配置校验在 logger 初始化后进行，确保错误信息能正确记录
