# AI-HMS 本地开发、真实环境模拟与内网 Linux 迁移指南

> 适用目录：`ai-hms_qhd`  
> 子项目：
> - `ai-hms-backend`（Go/Gin + PostgreSQL）
> - `ai-hms-frontend`（React + Vite）

---

## 1. 项目代码现状快速解析

## 1.1 后端（ai-hms-backend）

- 技术栈：Go 1.24.x、Gin、GORM、PostgreSQL
- 启动入口：`cmd/server/main.go`
- 配置加载：`config/config.go`（读取 `.env`）
- 数据库连接：`internal/database/database.go`
- 自动迁移：`internal/database/migrate.go`
- 业务路由：`internal/api/v1/*`
- HDIS 对接：`internal/integrations/hdis/*`

关键事实：

1. 服务启动时会尝试连库，连不上也会继续启动（只打印 warning）。
2. `GIN_MODE=debug` 会执行 AutoMigrate；`GIN_MODE=release` 不会自动迁移。
3. 当前登录是演示硬编码账号：`admin / admin123`。
4. 健康检查地址是：`/health`。

## 1.2 前端（ai-hms-frontend）

- 技术栈：React 19 + TS + Vite + Antd
- 启动：`npm run dev`
- API 客户端：`src/services/restClient.ts`

关键事实：

1. 前端实际读取的后端地址变量是：`VITE_API_BASE_URL`。
2. `.env` 里出现了 `VITE_BACKEND_API_URL`，但当前代码并未使用它。
3. 所以部署前必须确保 `VITE_API_BASE_URL` 正确。

---

## 2. Windows 本地开发环境搭建（可直接执行）

## 2.1 安装依赖

- Go 1.24.x
- Node.js 20 LTS+
- npm 10+
- PostgreSQL 15+

## 2.2 创建数据库

```sql
CREATE DATABASE ai_hms_db;