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
- ✅ **可测试启动生命周期** - bootstrap 统一管理迁移、监听、PID 与反向清理
- ✅ **健康检查** - 单一健康检查端点
- ✅ **DTO 实体隔离** - 所有接口返回实体通过 DTO 与内部模型解耦，防止直接暴露数据库结构

## 目录结构

```
http-services/
├── bootstrap/            # 应用组装与生命周期：初始化、迁移、HTTP、PID、清理
├── api/                    # API 相关代码
│   ├── app/               # 业务处理（按版本与分组组织）
│   │   └── v1/
│   │       ├── open/
│   │       │   └── health/    # 健康检查模块（开放，/api/v1/open/health）
│   │       └── private/       # 私有接口预留（/api/v1/private）
│   ├── middleware/        # 中间件
│   │   ├── access-log.go     # 结构化访问日志
│   │   ├── cross-domain.go   # 跨域处理
│   │   ├── jwt.go            # JWT 验证
│   │   ├── page.go           # 分页处理
│   │   ├── params.go         # 参数验证
│   │   ├── rate-limit.go     # 限流中间件
│   │   ├── security.go        # 安全响应头
│   │   ├── trace-id.go        # 请求追踪 ID
│   │   └── recovery.go       # 统一 panic recovery
│   ├── response/          # 响应处理
│   │   ├── code.go           # 状态码定义
│   │   └── format.go         # 响应格式化
│   └── router.go          # 路由配置
├── common/               # 跨模块共享语义预留（模板中为占位目录）
├── domain/               # 领域模型与领域服务（核心业务规则）
├── db/                   # 持久化适配器、模型、数据库常量与迁移入口
│   ├── migrate.go        # 顶层迁移聚合入口（按业务表域注册迁移）
│   ├── msqldb/           # MySQL/GORM client、基础模型、业务表域子包
│   │   ├── client.go     # GORM client、连接池、GORM 日志配置
│   │   └── base.go       # GORM 基础模型 BaseModel
│   └── rdb/              # Redis client 与缓存/session 访问封装
│       └── client.go     # Redis 初始化、获取与关闭
├── services/             # 长驻服务与后台任务预留
│   └── cron/             # 定时任务调度预留
├── config/                # 配置管理
│   ├── config.go          # 配置变量定义
│   ├── load.go            # 配置加载
│   └── check.go           # 配置校验
├── utils/                 # 工具函数
│   ├── authentication/    # JWT 认证工具
│   ├── contextkey/        # Gin context key 常量
│   ├── encryption/        # 加密工具（BCrypt）
│   ├── id/               # ID 生成器（Sonyflake）
│   ├── log/              # 日志管理
│   ├── pathtool/         # 路径工具
│   ├── pidfile/          # pid 文件管理
│   ├── random/           # 随机字符串
│   ├── runmodel/         # 运行模式检测
│   └── taskgroup/        # 并发任务组：取消、panic 恢复与有序错误
├── log/                   # 日志文件目录
├── static/               # 静态资源目录
├── bin/                  # 构建输出目录
├── dist/                 # 跨平台打包产物（make build-cross）
├── vendor/               # 本地依赖镜像目录；默认不提交，通常直接使用 go.mod/go.sum
├── .env.example          # 环境变量配置示例
├── config.yaml           # 配置文件（不提交到 Git）
├── config.yaml.example   # 配置文件示例
├── go.mod                # Go module 定义
├── go.sum                # Go module 校验文件
├── main.go               # CLI、版本、信号 context 与进程退出边界
├── Makefile              # 构建脚本
└── README.md             # 项目文档

```

## 推荐代码分层

模板建议按下面的边界扩展。当前模板的路由样例位于 `api/app/v1/open/health`，私有接口预留在 `api/app/v1/private`。`go-template/http-services` 已内置基础持久化组件，`common/`、`services/` 仍主要作为扩展占位，真实项目可以在这些目录里继续扩展共享语义和后台任务实现。

- `api/`：传输层，负责 Gin 路由、中间件、请求 DTO、响应 DTO 与领域错误到接口响应的映射，不承载核心业务规则。
- `bootstrap/`：应用组装与生命周期层，负责初始化配置/日志、可选迁移、listener/PID、HTTP 运行和资源反向清理；业务规则不放这里。
- `domain/`：业务规则层，放状态流转、领域错误、跨模块流程编排等和 HTTP 无关的逻辑。
- `db/`：持久化适配层，放数据库客户端、模型、查询封装、数据库常量和迁移入口。模板已内置 MySQL/GORM 与 Redis 基础 client，真实项目可继续按 MySQL、Redis 等适配器拆分。
- `services/`：长驻服务和后台任务层，放 cron、消息队列 consumer/producer、worker 等运行期任务。模板当前预留 `services/cron/` 占位。
- `common/`：跨模块共享语义，适合放枚举、常量、跨模块 DTO、事件封装等业务共识，不替代 `utils/`。
- `utils/`：基础设施工具，保留认证、加密、ID、日志、`utils/pathtool`、`utils/runmodel`、`utils/taskgroup` 等通用能力，不放具体业务规则。
- 真实应用如需部署、接口文档、脚本、示例或流水线，可按需增加 `docs/`、`deploy/`、`scripts/`、`examples/`、`.workflow/`。这些属于真实项目的生产化扩展，不是当前模板的必需目录。

### 目录放置规则

- `api/` vs `domain/`：`api/` 只处理 HTTP 入参、DTO、响应和错误翻译；状态流转、事务流程、跨模块业务规则放到 `domain/`。
- `domain/` vs `services/`：`domain/` 放可被 handler、worker、cron 复用的业务规则；`services/` 放长驻任务的启动、调度、消费循环和生命周期管理。
- `db/` vs `domain/`：`db/` 只表达存储结构、数据库常量、查询 helper、迁移和外部存储适配；跨表业务流程、状态机和领域错误不要下沉到 `db/`。
- `common/` vs `utils/`：`common/` 放跨模块共享的业务语义，如枚举、业务常量、跨模块 DTO、事件定义；`utils/` 只放与业务无关的基础设施工具，如日志、认证、加密、ID、路径、pidfile。
- 数据库常量优先跟随对应数据库模块放在 `db/` 下；只有被多个业务模块作为业务语义共同使用时，才抽到 `common/`。

### db 目录约定

`db/` 是持久化适配层，只表达外部存储的连接、模型、查询、迁移和存储侧常量，不承载跨表业务流程、状态机、接口响应或用户提示文案。`domain/` 可以直接调用 `db/msqldb`、`db/rdb` 中的函数，但 `db/` 不应反向 import `domain/` 或 `api/`。

模板内置了 `db/msqldb/` 与 `db/rdb/` 的基础组件。落到真实项目时，常见扩展方式如下：

| 路径/文件 | 通常放什么 |
| --- | --- |
| `db/migrate.go` | 顶层数据库迁移聚合入口，按业务表域固定顺序调用各子包迁移。 |
| `db/msqldb/client.go` | GORM MySQL 单例 client、连接池、GORM logger、初始化选项等。 |
| `db/msqldb/base.go` | 表模型共享基础字段，例如 ID、创建时间、更新时间、软删除等；只保留 GORM 语义，不定义对外 JSON 契约。 |
| `db/msqldb/<module>/model.go` | 某个业务表域的 GORM model，只描述表结构和存储字段，优先只写 GORM tag。 |
| `db/msqldb/<module>/query.go` | 该表域的 Get/List/Create/Update/Delete、分页查询、条件查询等 helper。 |
| `db/msqldb/<module>/constants.go` | 与数据库字段强相关的状态、类型、来源等常量，避免调用方硬编码 magic number。 |
| `db/msqldb/<module>/migrate.go` | 子包内 `AutoMigrate` 聚合点，新增 model 时同步检查这里。 |
| `db/rdb/client.go` | Redis client 初始化、连接获取、关闭逻辑，以及必要的连接池配置。 |

`<module>` 建议按业务表域命名，而不是按技术动作命名。例如 `user/`、`order/`、`product/`、`notice/`、`area/` 可以分别代表用户、订单、商品、公告、地区树等表域；这些只是命名示例，不是模板必须内置的目录。

查询函数应返回 DB model、持久层结构或基础错误；对外 DTO、HTTP 状态、响应 envelope、中文接口提示应留在 `api/`，业务规则错误和跨表流程应留在 `domain/`。事务敏感函数可以提供 `InTx` 变体，但事务边界优先由 `domain/` 决定。

### 新增业务模块落地流程

以新增 `user` 模块为例，建议按下面顺序放代码：

```text
db/msqldb/user/
├── model.go       # user 相关表结构，嵌入 msqldb.BaseModel
├── constants.go   # 和 user 表字段强相关的状态/类型常量
├── query.go       # user 表的查询、写入、分页、事务 helper
└── migrate.go     # user 表域 AutoMigrate 聚合点

domain/user/
├── errors.go      # user 领域错误，不写 HTTP response
└── user.go        # user 业务规则、状态流转、跨表流程编排

api/app/v1/private/user/
├── dto.go         # 请求/响应 DTO，不直接暴露 DB model
├── errors.go      # user 领域错误到 response code/message 的映射
├── user.go        # handler：参数绑定、调用 domain、DTO 映射、返回响应
└── router.go      # 注册 /api/v1/private/user 路由
```

落地步骤：

1. 先在 `db/msqldb/<module>/model.go` 定义表结构；通用 ID、时间、软删除字段优先嵌入 `msqldb.BaseModel`。
2. 在同包 `query.go` 写 Get/List/Create/Update/InTx 等持久化 helper，不写 HTTP 语义和用户提示文案。
3. 在同包 `migrate.go` 提供 `Migrate(db *gorm.DB) error`，再到 `db/migrate.go` 的 `MigrateAll` 中按依赖顺序注册。
4. 在 `domain/<module>/` 写业务规则、领域错误和跨表流程；需要事务时由 domain 决定事务边界，再调用 db 层的 `InTx` helper。
5. 在 `api/app/v1/{open|private}/<module>/` 写路由、handler、DTO 和错误映射；handler 不直接返回 DB model。
6. 只有需要多个模块共享的业务语义才放 `common/<module>/`；定时任务、consumer、worker、外部 client 放 `services/`，并薄调用 domain。

### 放置决策速查

写业务时如果不确定代码放哪里，先按下面的规则判断：

| 你正在写的内容 | 放置位置 | 示例 |
| --- | --- | --- |
| HTTP 请求体、query 参数、响应字段 | `api/app/v1/{open|private}/<module>/dto.go` | `CreateUserRequest`、`UserDTO` |
| Gin handler、参数绑定、响应返回 | `api/app/v1/{open|private}/<module>/<module>.go` | `CreateUser(c *gin.Context)` |
| 路由注册 | `api/app/v1/{open|private}/<module>/router.go` | `group.POST("/user", CreateUser)` |
| 领域错误、状态流转、跨表流程 | `domain/<module>/` | `ErrUserDisabled`、`DisableUser` |
| 表结构、字段 tag、索引 | `db/msqldb/<module>/model.go` | `type User struct { msqldb.BaseModel ... }` |
| 单表或表域查询写入 | `db/msqldb/<module>/query.go` | `GetUserByID`、`CreateUser` |
| 和数据库字段强绑定的状态/类型 | `db/msqldb/<module>/constants.go` | `UserStatusEnabled = 1` |
| 多模块共享的业务语义 | `common/<module>/` 或 `common/<scene>/` | 登录上下文 key、跨模块事件结构 |
| Redis key 拼接和缓存/session 访问 | 优先 `domain/<module>/`，通过 `db/rdb.Client()` 获取 client | `userSessionRedisKey(userID)` |
| 定时任务、consumer、worker 启停 | `services/<name>/` | `services/cron`、`services/kafkax` |
| 通用且无业务含义的工具 | `utils/<name>/` | 日志、ID、加密、路径工具 |

默认不额外创建 `repository/`、`biz/`、`dao/`、`model/` 顶层目录；当前模板已经用 `api -> domain -> db/services` 表达边界。只有当某个抽象被多个模块真实复用，或外部依赖需要替换以便测试时，再引入更细的接口抽象。

### 完整示例：新增 user 模块

下面的示例不是模板内置功能，而是展示新增业务时各层应该如何协作。

#### 1. 持久层模型：`db/msqldb/user/model.go`

```go
package user

import "http-services/db/msqldb"

type User struct {
    msqldb.BaseModel
    Username string `gorm:"column:username;type:varchar(64);not null;uniqueIndex;comment:用户名"`
    Nickname string `gorm:"column:nickname;type:varchar(64);not null;default:'';comment:昵称"`
    Status   int    `gorm:"column:status;not null;default:1;index;comment:状态"`
}
```

DB model 不写对外 JSON 契约；接口字段统一在 API DTO 中声明。

#### 2. 持久层常量：`db/msqldb/user/constants.go`

```go
package user

const (
    UserStatusEnabled  = 1
    UserStatusDisabled = 2
)
```

这类常量和数据库字段强绑定，优先放在 `db/msqldb/user`。如果后来多个业务模块都把它当作业务语义使用，再考虑抽到 `common/user`。

#### 3. 持久层查询：`db/msqldb/user/query.go`

```go
package user

import (
    "errors"

    "http-services/db/msqldb"

    "gorm.io/gorm"
)

func GetByID(id uint) (*User, error) {
    db, err := msqldb.Client()
    if err != nil {
        return nil, err
    }

    var item User
    err = db.Where("id = ?", id).First(&item).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }
    if err != nil {
        return nil, err
    }
    return &item, nil
}

func Create(item *User) error {
    db, err := msqldb.Client()
    if err != nil {
        return err
    }
    return db.Create(item).Error
}
```

`query.go` 不返回中文提示，不拼 API response，也不理解 Gin context。它只描述怎么读写存储。

#### 4. 持久层迁移：`db/msqldb/user/migrate.go`

```go
package user

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
    return db.AutoMigrate(&User{})
}
```

然后在 `db/migrate.go` 注册：

```go
import (
    "fmt"

    "http-services/db/msqldb"
    userdb "http-services/db/msqldb/user"
)

func MigrateAll() error {
    database, err := msqldb.Client()
    if err != nil {
        return fmt.Errorf("init mysql client: %w", err)
    }
    sqlDB, err := database.DB()
    if err != nil {
        return fmt.Errorf("get mysql sql.DB: %w", err)
    }
    if err := sqlDB.Ping(); err != nil {
        return fmt.Errorf("ping mysql: %w", err)
    }

    return RunMigrators(database,
        Migrator{Name: "user", Migrate: userdb.Migrate},
    )
}
```

#### 5. 领域层：`domain/user/user.go`

```go
package user

import (
    "errors"

    userdb "http-services/db/msqldb/user"
)

var ErrUserDisabled = errors.New("user disabled")

type Profile struct {
    ID       uint
    Username string
    Nickname string
}

func GetProfile(id uint) (Profile, error) {
    item, err := userdb.GetByID(id)
    if err != nil {
        return Profile{}, err
    }
    if item.Status == userdb.UserStatusDisabled {
        return Profile{}, ErrUserDisabled
    }

    return Profile{
        ID:       item.ID,
        Username: item.Username,
        Nickname: item.Nickname,
    }, nil
}
```

领域层可以使用 db model 作为内部输入，但对外返回更稳定的领域结构或结果；不要把 `gin.Context`、HTTP code、response envelope 放进这里。

#### 6. API DTO：`api/app/v1/private/user/dto.go`

```go
package user

type ProfileDTO struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Nickname string `json:"nickname"`
}
```

DTO 是对外契约。即使当前字段和 DB model 一样，也建议显式映射，避免未来 DB 字段变化直接泄漏到接口。

#### 7. API 错误映射：`api/app/v1/private/user/errors.go`

```go
package user

import (
    "errors"

    "http-services/api/response"
    userdomain "http-services/domain/user"
    httplog "http-services/utils/log"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

func ReturnDomainError(c *gin.Context, err error) {
    httplog.WithRequest(c).Error("user domain error", zap.Error(err))
    switch {
    case errors.Is(err, userdomain.ErrUserDisabled):
        response.ReturnError(c, response.PERMISSION_DENIED, "用户已停用")
    default:
        response.ReturnError(c, response.INTERNAL, "用户操作失败")
    }
}
```

领域错误到用户可读文案的转换放在 API 层；这样同一个 domain 函数被 HTTP、cron、worker 调用时都不会携带接口语义。

#### 8. API handler 与路由

```go
package user

import (
    "strconv"

    "http-services/api/response"
    userdomain "http-services/domain/user"

    "github.com/gin-gonic/gin"
)

func GetProfile(c *gin.Context) {
    id64, err := strconv.ParseUint(c.Query("id"), 10, 64)
    if err != nil || id64 == 0 {
        response.ReturnError(c, response.INVALID_ARGUMENT, "用户 ID 不合法")
        return
    }

    profile, err := userdomain.GetProfile(uint(id64))
    if err != nil {
        ReturnDomainError(c, err)
        return
    }

    response.ReturnOk(c, ProfileDTO{
        ID:       profile.ID,
        Username: profile.Username,
        Nickname: profile.Nickname,
    })
}

func RegisterPrivateRoutes(private *gin.RouterGroup) {
    group := private.Group("/user")
    group.GET("/profile", GetProfile)
}
```

再到上一级聚合路由注册这个模块，例如在 `api/app/v1/private/router.go` 中调用 `user.RegisterPrivateRoutes(private)`。

### 事务、Redis 与后台任务约定

- 事务边界优先放在 `domain/`。如果一个流程要同时写用户表、余额表、流水表，由 domain 开启事务，再调用 db 层提供的 `InTx` helper；不要让单个 `query.go` 偷偷决定跨表事务。
- Redis key 拼接函数放在使用它的业务包附近，例如 `domain/user/session.go` 里的 `userSessionRedisKey(id)`；不要把业务 Redis key 放到 `utils/`。访问 Redis 时优先使用 `rdb.Client()` 并处理错误，`GetClient()` 只作为兼容便捷入口。
- `services/cron`、consumer、worker 的 handler 应尽量只有“解析任务参数 -> 调 domain -> 记录结果”，核心状态流转仍放在 `domain/`。
- 配置项新增时同时改 `config/config.go`、`config/load.go`、`config.yaml.example` 和配置测试；环境变量名遵循 `HTTP_SERVICES_<SECTION>_<KEY>`。
- 新增业务代码时优先补包内测试：db 层测查询条件和迁移注册，domain 层测状态流转和错误，api 层用 `httptest` 测响应 envelope。

### common、domain 与 services 补充约定

- `common/` 适合按端侧或跨模块语义拆包，例如 `auth/`、`tenant/`、`portal/`，放 context key、跨模块 DTO、事件结构或业务常量；不要放数据库 model，也不要替代 `utils/`。
- `domain/` 可以按业务入口或核心业务域拆包，例如认证会话、用户、订单、健康检查等；它可以编排多个 `db/` 查询和外部 client，但不返回 Gin context、HTTP response 或 API envelope。
- `services/` 放长期运行服务和外部系统适配，例如 `cron/`、Kafka consumer runtime、OpenAPI client、通知 client、worker 等；定时任务和 consumer handler 应尽量薄调用 `domain/`，不要把核心业务规则写成只能由后台任务复用的形态。

## DTO 返回规范与领域分层示例

为了避免直接向外暴露数据库 Model 或内部 Service 结构体，所有对外 API 的业务实体返回都应通过 DTO（Data Transfer Object，数据传输对象）进行一层映射：

- Handler 只负责将领域对象/模型转换为 DTO，再通过 `response.ReturnOk` / `response.ReturnOkWithTotal` 返回。
- DTO 可以放在具体模块目录下（例如：`api/app/v1/open/health/dto.go`），也可以放在 `common` 目录中供多个模块复用。
- DTO 中仅保留对外需要的字段，可对内部字段进行重命名、组合或过滤。

以健康检查接口为例，完整链路为：领域层模型 → DTO → 统一响应：

```go
// domain/health/status.go
type Status struct {
	Status    string
	Ready     bool
	Uptime    time.Duration
	Timestamp int64
}

func GetStatus() (Status, error) {
	return Status{
		Status:    "ok",
		Ready:     true,
		Uptime:    time.Since(startTime),
		Timestamp: time.Now().Unix(),
	}, nil
}

// api/app/v1/open/health/dto.go
type StatusDTO struct {
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`
	Uptime    string `json:"uptime"`
	Timestamp int64  `json:"timestamp"`
}

// api/app/v1/open/health/health.go
func Status(c *gin.Context) {
	status, err := healthDomain.GetStatus()
	if err != nil {
		health.ReturnDomainError(c, err)
		return
	}

	dto := StatusDTO{
		Status:    status.Status,
		Ready:     status.Ready,
		Uptime:    status.Uptime.String(),
		Timestamp: status.Timestamp,
	}
	response.ReturnOk(c, dto)
}
```

在实际业务开发中，禁止直接返回数据库实体（如 `db.User`）、ORM Model 或 Service 内部结构体，必须通过 DTO 显式挑选与组合需要暴露的字段。

### 领域错误与 API 层错误响应示例

错误码与错误信息分层：

- 全局错误码：在 `api/response/code.go` 定义，例如 `FAILED_PRECONDITION`、`UNAVAILABLE`、`INTERNAL` 等；
- 领域错误：在 `domain/<module>/errors.go` 定义具有业务语义的错误（不关心 HTTP 细节）；
- API 层错误响应：在对应模块的 `errors.go` 中，将领域错误映射为统一响应。

以健康检查模块为例：

```go
// domain/health/errors.go
var (
	ErrServiceNotReady  = errors.New("service not ready")
	ErrServiceUnhealthy = errors.New("service unhealthy")
)

// api/app/v1/open/health/errors.go
func ReturnDomainError(c *gin.Context, err error) {
	// 在统一错误映射前记录真实的领域错误和请求上下文，便于排查
	log.WithRequest(c).Error("健康检查领域错误", zap.Error(err))

	switch {
	case errors.Is(err, domain.ErrServiceNotReady):
		data := response.FAILED_PRECONDITION
		data.Code = CodeHealthServiceNotReady // 模块级自定义业务错误码
		response.ReturnError(c, data, "服务尚未就绪，请稍后重试")
	case errors.Is(err, domain.ErrServiceUnhealthy):
		data := response.UNAVAILABLE
		data.Code = CodeHealthServiceUnhealthy // 模块级自定义业务错误码
		response.ReturnError(c, data, "服务当前不可用，请稍后重试")
	default:
		response.ReturnError(c, response.INTERNAL, "服务内部错误")
	}
}
```

handler 中在拿到领域错误后，只需要调用 `ReturnDomainError(c, err)` 即可，既保证了错误码统一，又不会把领域层实现细节泄露到接口层。

### 统一响应 HTTP 状态策略

当前脚手架的 `api/response` 统一使用 HTTP 200 返回 JSON，业务成功或失败由响应体中的 `code` 与 `status` 表达。例如鉴权失败会返回 `{"code":401,"status":"UNAUTHENTICATED"}`，但 HTTP status 仍为 200。这样做便于某些前端网关统一解析业务错误；如果你的项目需要 REST 风格 HTTP status，应在项目初始化阶段统一调整 `api/response` 与相关测试，不要在单个 handler 中混用两套策略。

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

# 执行数据库迁移后退出
make migrate
# 或
./bin/http-services -m
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
  pid_file: "http-services.pid"   # pid 文件路径（支持相对路径）
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

database:
  mysql_dsn: ""                    # MySQL DSN；需要迁移或访问 db/msqldb 时填写

redis:
  host: "127.0.0.1:6379"           # Redis 地址；需要 session、验证码、缓存等能力时使用
  password: ""                     # Redis 密码；未设置时留空
  key_prefix: ""                   # Redis key 公共前缀；非空时必须以 : 结尾，例如 service:env:

log:
  max_size: 50                    # 单个日志文件最大大小（MB）
  max_age: 30                     # 保留旧日志文件的最大天数
  level: "info"                  # 业务日志级别: debug, info, warn, error
  gin_level: ""                  # Gin access/error 日志级别；为空时跟随 level
```

#### Redis 公共 key 前缀

`redis.key_prefix` 默认为空，因此不会改变现有 Redis key。YAML 和环境变量中的值会去除首尾空白；每个非空前缀都必须以 `:` 结尾，例如 `service:env:`。这是配置使用契约，加载器不会自动补充分隔符或校验格式。

Redis client 初始化时会捕获该前缀，因此修改后必须重启服务。切换前缀不会迁移或删除旧前缀下的 key；如需保留旧数据，请自行安排迁移或清理。

内置 hook 覆盖常用字符串与计数器、过期、Hash、List、Set、Sorted Set、多 key、重命名/复制、Lua 声明 key、`KEYS` 以及 `SCAN`/`HSCAN`/`SSCAN`/`ZSCAN` 等命令。启用公共前缀后，全局 `SCAN` 必须提供非空 `MATCH`（例如 `*`），否则会直接返回错误，避免扫描到其他命名空间。自定义命令和 Redis module 命令会原样透传；依赖公共前缀前，必须为其 key 参数位置显式增加 hook 支持和测试。

### HTTPS/TLS 部署

服务进程只提供 HTTP，不内置 ACME 自动证书签发或本地证书文件 TLS 热更新。生产环境建议在 Caddy、Nginx、Traefik、Kubernetes Ingress 或云负载均衡层终止 HTTPS，再将流量反向代理到本服务监听端口。

服务默认信任本机反向代理来源 `127.0.0.1` 和 `::1`，因此通过本机 Caddy/Nginx 反代时，Gin 的 `ClientIP()` 会从 `X-Forwarded-For` / `X-Real-IP` 获取真实客户端 IP。如果反向代理与服务不在同一主机或同一 loopback 来源，请在 `api/router.go` 中将 `SetTrustedProxies` 调整为实际代理 IP 或网段，避免直接信任所有来源。

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
export HTTP_SERVICES_LOG_MAX_AGE=60
export HTTP_SERVICES_LOG_LEVEL=warn
export HTTP_SERVICES_LOG_GIN_LEVEL=info

# 覆盖 MySQL / Redis 配置
export HTTP_SERVICES_DATABASE_MYSQL_DSN="user:pass@tcp(127.0.0.1:3306)/app?charset=utf8mb4&parseTime=True&loc=Local"
export HTTP_SERVICES_REDIS_HOST="127.0.0.1:6379"
export HTTP_SERVICES_REDIS_PASSWORD=""
export HTTP_SERVICES_REDIS_KEY_PREFIX="service:env:"

# 运行服务
./bin/http-services
```

如果你希望业务日志和 Gin 日志分开控级，可以这样配置：

```yaml
log:
  level: "warn"
  gin_level: "info"
```

说明：

- `log.level` 控制业务日志级别；
- `log.gin_level` 控制 Gin access/error 日志级别；
- 当 `log.gin_level` 为空时，Gin 日志默认跟随 `log.level`；
- 配置热重载后，logger 会自动刷新，无需重启服务。

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
- `trace_id`: 请求追踪 ID（由网关或中间件注入 `X-Trace-ID`）
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
pageQuery := middleware.ParsePageQuery(c)
if pageQuery.IsDisabled() {
    // page=-1 或 page_size=-1 表示取消分页
}

// 取消分页（获取全部数据）
// 请求参数：page=-1 或 page_size=-1
```

### 4. 路由组织（分层）

```go
// 顶层：api/router.go（仅初始化与挂载 /api，业务路由下沉到 app 层）
func InitApi() *gin.Engine {
    router := gin.New()
    router.Use(middleware.TraceID(), middleware.AccessLog(), middleware.Recovery())
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
import "http-services/utils/contextkey"

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
jwtData, exists := c.Get(contextkey.JWTData)
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

- 业务日志输出到文件 `log/<程序名>.log`（例如：`log/http-services.log`）
- Gin 框架日志输出到文件 `log/<程序名>.gin.log`（例如：`log/http-services.gin.log`）
- JSON 格式，便于日志分析
- Info 级别日志
- 自动轮转（可通过配置文件或环境变量自定义）：
  - 按天切割：每天 00:00（本地时间）主动轮转
  - 单文件最大大小兜底：默认 50MB（可配置 `log.max_size`），同一天写满会继续产生新文件（文件名带时间戳）
  - 保留天数：默认 30 天（可配置 `log.max_age`）

### 自定义日志配置

通过配置文件：

```yaml
log:
  max_size: 100      # 单个日志文件最大 100MB
  max_age: 60        # 保留 60 天
```

通过环境变量：

```bash
export HTTP_SERVICES_LOG_MAX_SIZE=100
export HTTP_SERVICES_LOG_MAX_AGE=60
```

### 使用示例

```go
import "go.uber.org/zap"

// 记录日志
zap.L().Info("业务事件", zap.String("action", "process"))
zap.L().Error("操作失败", zap.Error(err))
zap.L().Debug("调试信息", zap.Any("data", data))

// api 层获取日志实例（仅带基础上下文，如 trace_id、method、path）
log := log.FromContext(c)
log.Info("处理请求", zap.String("path", c.Request.URL.Path))

// 如需在排查问题时同时记录本次请求的参数（query / 表单 / 路径参数），
// 可以使用 WithRequest 获取带请求参数字段的 logger：
log := log.WithRequest(c)
log.Error("处理请求失败", zap.Error(err))
```

## 命令行参数

```bash
# 开发模式
./bin/http-services -d
./bin/http-services --dev

# 查看版本
./bin/http-services -v
./bin/http-services --version

# 执行数据库迁移后退出
./bin/http-services -m
./bin/http-services --migrate

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
make migrate   # 构建并执行数据库迁移
make clean     # 清理构建文件
make version   # 显示版本信息
make test      # 运行测试
make test-race # 使用 race detector 与随机顺序运行测试
make verify    # 依次执行格式化、静态检查、覆盖率测试和 race 测试
```

### 测试说明

- 运行 `make test`：
  - 输出中文汇总（包数量、通过/失败、用例通过/失败/跳过）
  - 生成覆盖率文件 `coverage.out` 并打印总覆盖率
- 运行 `make test-race`：使用 `-race -shuffle=on -count=1` 检查数据竞争和测试顺序依赖
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

### 2. 使用 Caddy/Nginx 反向代理

Caddy 示例：

```caddyfile
api.example.com {
    reverse_proxy 127.0.0.1:8080
}
```

Nginx 示例：

```nginx
upstream http_services {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    # 生产环境可在这里配置 listen 443 ssl，或由 Caddy/Ingress/负载均衡统一终止 HTTPS。
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
- `gorm.io/gorm` / `gorm.io/driver/mysql` - MySQL 持久化基础组件
- `redis/go-redis/v9` - Redis 客户端
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
