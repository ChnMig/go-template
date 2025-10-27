# 项目文档索引

## 后端文档
`backend/configuration.md` - 后端配置管理架构与使用说明，修改配置相关代码时必读
`backend/middleware.md` - 中间件架构与使用文档，开发API时必读

## 全局重要记忆
- 项目使用 YAML 配置文件管理配置项，配置文件位于 `http-services/config.yaml`
- 配置文件已加入 `.gitignore`，使用 `config.yaml.example` 作为模板
- 所有配置加载在程序启动时完成，并在 logger 初始化后进行校验
- 项目基于 art-design-pro-edge-go-server 框架，定期同步基础组件更新
- 使用标准JWT认证，简洁高效
