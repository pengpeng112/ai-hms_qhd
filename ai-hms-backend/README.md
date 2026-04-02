# AI HMS Backend

透析中心管理系统后端服务

## 技术栈

- **语言**: Go 1.24.0
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL 15+
- **ORM**: GORM
- **认证**: JWT (JSON Web Tokens)

## 项目概述

本系统是为透析中心设计的业务管理后端服务，提供患者管理、排班管理、治疗记录等核心功能。

### 核心功能模块

- ✅ **用户认证**: JWT 认证授权
- ✅ **患者管理**: 患者信息、血管通路、病史记录
- ✅ **住院管理**: 住院记录管理
- ✅ **排班管理**: 病房、床位、班次管理
- ✅ **治疗管理**: 透析治疗记录、透前/透后体征、治疗参数
- 🚧 **设备监控**: (待实现)
- 🚧 **医嘱管理**: (待实现)

## 项目结构

```
.
├── cmd/
│   └── server/              # 应用程序入口
│       └── main.go        # 主函数、路由注册
├── config/                  # 配置文件管理
│       └── config.go       # 环境变量加载
├── internal/
│   ├── models/              # 数据库模型定义
│   │   ├── user.go         # 用户模型
│   │   ├── patient.go      # 患者模型
│   │   ├── hospitalization.go
│   │   ├── schedule.go      # 排班相关模型
│   │   └── treatment.go     # 治疗相关模型
│   ├── database/            # 数据库连接管理
│   │   ├── database.go      # 数据库初始化
│   │   └── migrate.go       # 自动迁移
│   ├── api/                 # HTTP API 层
│   │   └── v1/              # API v1 版本
│   │       ├── patient_handler.go
│   │       ├── hospitalization_handler.go
│   │       ├── schedule_handler.go
│   │       └── treatment_handler.go
│   ├── services/            # 业务逻辑层
│   │   ├── patient_service.go
│   │   ├── hospitalization_service.go
│   │   ├── shift_service.go
│   │   ├── patient_shift_service.go
│   │   └── treatment_service.go
│   ├── middleware/          # HTTP 中间件
│   │   ├── cors.go          # CORS 处理
│   │   └── auth.go          # JWT 认证
│   └── utils/               # 工具函数
│       ├── jwt.go           # JWT Token 管理
│       └── password.go      # 密码哈希
├── pkg/                     # 公共库
│   └── response/           # HTTP 响应封装
├── docs/                    # 项目文档
├── scripts/                 # 构建和部署脚本
├── .env                     # 环境变量配置
└── go.mod                   # Go 模块依赖
```

## 数据库表设计

当前已创建 **19 张表**，分为以下模块：

### 1. 用户模块 (1张表)
- `users` - 用户表

### 2. 患者管理模块 (4张表)
- `patients` - 患者基本信息
- `vascular_accesses` - 血管通路
- `medical_histories` - 病史记录
- `infection_infos` - 感染信息

### 3. 住院管理模块 (1张表)
- `hospitalizations` - 住院信息

### 4. 治疗方案模块 (3张表)
- `treatment_plans` - 透析治疗方案
- `prescriptions` - 每日处方
- `orders` - 医嘱

### 5. 排班管理模块 (4张表)
- `wards` - 病房
- `beds` - 床位
- `shifts` - 班次
- `patient_shifts` - 患者排班

### 6. 透析治疗模块 (6张表)
- `Treatment_Treatment` - 透析治疗主表
- `Treatment_BeforeCheck` - 透前检查
- `Treatment_BeforeSigns` - 透前体征
- `Treatment_DuringParam` - 透析中参数
- `Treatment_AfterSigns` - 透后体征
- `Treatment_Alarm` - 报警记录

详细的数据库设计请参考: `docs/database-schema.md`

## 快速开始

### 1. 环境要求

- Go 1.24.0+
- PostgreSQL 15+
- Git

### 2. 安装依赖

```bash
cd ai-hms-backend
go mod download
```

### 3. 配置数据库

#### 3.1 创建数据库

```sql
CREATE DATABASE ai_hms_db;
```

#### 3.2 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置数据库连接信息：

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=ai_hms_db
DB_SSL_MODE=disable

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=debug

# JWT 配置
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24
```

### 4. 启动服务

#### 方式一: 直接运行

```bash
go run cmd/server/main.go
```

#### 方式二: 使用 Air (热重载开发)

```bash
# 安装 Air
go install github.com/cosmtrek/air@latest

# 运行
air
```

### 5. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 登录测试
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

服务将在 `http://localhost:8080` 上运行

## API 接口文档

### 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON

### 认证接口

#### 登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "userId": "1",
  "username": "admin",
  "realName": "系统管理员",
  "role": "ADMIN"
}
```

### 患者管理接口

#### 获取患者列表
```http
GET /api/v1/patients?page=1&pageSize=20&name=&status=
Authorization: Bearer {token}
```

#### 获取患者详情
```http
GET /api/v1/patients/{id}
Authorization: Bearer {token}
```

#### 创建患者
```http
POST /api/v1/patients
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "张三",
  "age": 45,
  "gender": "男",
  "diagnosis": "慢性肾功能衰竭",
  "bedNumber": "01-02",
  "riskLevel": "高危"
}
```

### 住院管理接口

#### 获取住院列表
```http
GET /api/v1/hospitalizations?page=1&pageSize=20
Authorization: Bearer {token}
```

#### 获取患者当前住院信息
```http
GET /api/v1/patients/{patientId}/hospitalization
Authorization: Bearer {token}
```

### 排班管理接口

#### 获取班次列表
```http
GET /api/v1/shifts
Authorization: Bearer {token}
```

#### 获取患者当日排班
```http
GET /api/v1/patients/{patientId}/shift?date=2025-01-28
Authorization: Bearer {token}
```

### 治疗管理接口

#### 获取治疗记录列表
```http
GET /api/v1/treatments?page=1&pageSize=20&patientId=1&status=1
Authorization: Bearer {token}
```

#### 获取患者指定日期治疗
```http
GET /api/v1/patients/{patientId}/treatment?date=2025-01-28
Authorization: Bearer {token}
```

#### 更新治疗状态
```http
PUT /api/v1/treatments/{id}/status
Authorization: Bearer {token}
Content-Type: application/json

{
  "status": 1
}
```

状态说明:
- `0` - 待开始
- `1` - 进行中
- `2` - 已完成
- `3` - 已取消

## 开发指南

### 代码规范

- 遵循 Go 语言编码规范
- 使用 `golint` 和 `go fmt` 格式化代码
- 所有 `export` 的函数需要添加注释
- 复杂业务逻辑需要添加单元测试

### 提交代码

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 运行测试
go test ./...

# 提交
git add .
git commit -m "your commit message"
```

### 数据库迁移

所有表结构变更通过 GORM AutoMigrate 自动同步：

```go
// internal/database/migrate.go
func AutoMigrate() error {
    return DB.AutoMigrate(
        &models.User{},
        &models.Patient{},
        // ... 其他模型
    )
}
```

## 配置说明

### JWT 配置

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `JWT_SECRET` | JWT 签名密钥 | `your-secret-key` |
| `JWT_EXPIRATION_HOURS` | Token 有效期(小时) | 24 |

### 数据库配置

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户 | `postgres` |
| `DB_NAME` | 数据库名 | `ai_hms_db` |
| `DB_SSL_MODE` | SSL模式 | `disable` |

### 服务器配置

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `SERVER_HOST` | 监听地址 | `0.0.0.0` |
| `SERVER_PORT` | 监听端口 | `8080` |
| `GIN_MODE` | 运行模式 | `debug` |

## 更新日志

### v1.0.0 (2025-01-28)

**新增功能**:
- ✅ 项目初始化 (Gin + GORM + PostgreSQL)
- ✅ JWT 认证授权
- ✅ 用户管理模块
- ✅ 患者管理模块 (含基本信息、血管通路、病史记录)
- ✅ 住院管理模块
- ✅ 排班管理模块 (病房、床位、班次、患者排班)
- ✅ 透析治疗模块 (治疗记录、透前/透后体征、治疗参数、报警)

**数据库表**: 19 张

**API 端点**: 30+ 个

**已知问题**:
- 登录使用硬编码测试账号 (待完善)
- 需要添加更完善的请求参数验证
- 需要添加单元测试

## 文档

详细的开发日志请参考: `docs/2025-01-28-backend-development.md`

## License

MIT License
