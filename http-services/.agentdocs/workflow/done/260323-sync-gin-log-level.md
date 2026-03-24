# 脚手架同步：独立 Gin 日志级别配置

## 任务背景

源仓库 `/Users/chenming/work/fenxiangwuxian/catering/fuli-services` 中，日志基础设施已经支持：

- 业务日志与 Gin 日志使用独立 logger / 独立文件；
- 业务日志级别由 `log.level` 控制；
- Gin access/error 日志级别由 `log.gin_level` 控制；
- 当 `log.gin_level` 为空时，Gin 日志默认跟随 `log.level`；
- 配置热重载后会重新执行 `log.SetLogger()`，使日志级别即时生效。

目标仓库 `/Users/chenming/work/go-template/http-services` 是脚手架项目，适合同步这类通用、底层、可复用的运行时能力。

## 本次目标

1. 将独立 Gin 日志级别配置同步到脚手架仓库；
2. 保持默认行为兼容：未配置 `log.gin_level` 时继续跟随 `log.level`；
3. 补齐脚手架侧测试、示例配置与 README 说明；
4. 不引入任何业务域逻辑。

## 非目标

- 不同步 OpenCustomer / Kafka / notice / order 等业务代码；
- 不同步源仓库的业务文档；
- 不改动 Gin 日志与业务日志的文件拆分方案；
- 不扩展到 DB/gorm 日志级别独立控制。

## 实施阶段

### 阶段 0：梳理
- [x] 对比源仓库与脚手架仓库的 config / log / main 结构
- [x] 确认最小迁移面

### 阶段 1：实现
- [x] 配置层新增 `log.level` / `log.gin_level`
- [x] logger 初始化支持业务与 Gin 独立级别
- [x] 配置热重载时重新执行 `log.SetLogger()`

### 阶段 2：验证与文档
- [x] 新增/更新 logger 单元测试
- [x] 更新 `config.yaml.example` 与 `README.md`
- [x] 执行 `make verify`

## 关键约束

- 这是脚手架仓库，只同步通用能力；
- 任何迁移都必须以“模板最小化、通用化”为第一原则；
- 默认行为必须兼容旧配置，不能要求已有模板用户立即调整配置。

## 完成结果

- 已在脚手架仓库增加独立 Gin 日志级别配置：`log.gin_level`；
- 当 `log.gin_level` 为空时，Gin 日志默认跟随 `log.level`；
- 已在配置热重载时重新执行 `log.SetLogger()`，保证日志级别变更能即时生效；
- 已补齐 `utils/log/log_test.go`，覆盖“独立 Gin 日志级别”和“跟随业务级别回退”两种行为；
- 已更新 `config.yaml.example` 与 `README.md`；
- 已在脚手架仓库执行 `go test ./utils/log` 与 `make verify` 并通过。
