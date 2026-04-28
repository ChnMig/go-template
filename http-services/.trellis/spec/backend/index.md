# Backend Development Guidelines

> 本项目后端开发规范索引，供 Trellis implement/check 子代理加载。

---

## Overview

本仓库是 Go HTTP API 服务模板，主要技术栈为 Gin、Viper、Zap、Lumberjack、JWT 与标准库测试。后端开发应优先遵循本目录中的真实项目规范，而不是套用通用脚手架习惯。

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | 模块组织、路由分层、工具包位置 | Filled |
| [Database Guidelines](./database-guidelines.md) | 当前数据库状态、未来接入约束 | Filled |
| [Error Handling](./error-handling.md) | 领域错误、HTTP 统一响应、错误映射 | Filled |
| [Quality Guidelines](./quality-guidelines.md) | 代码质量、测试、验证命令 | Filled |
| [Logging Guidelines](./logging-guidelines.md) | Zap 日志、Gin 独立日志、请求上下文日志 | Filled |

---

## 使用要求

1. 修改后端代码前读取本索引，并按任务类型读取对应规范文件。
2. 以现有代码模式为准：路由在 `api/` 下分层注册，业务语义在 `domain/` 下表达，通用能力放在 `utils/`，配置集中在 `config/`。
3. 文档与代码注释使用中文；保留 Go、Gin、Zap 等英文专有名词。
4. 新增功能必须补充贴近变更风险的测试，并至少运行 `make fmt`、`make lint`、`make test` 或等价验证。
