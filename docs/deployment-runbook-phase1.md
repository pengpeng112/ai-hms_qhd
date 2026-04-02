# AI-HMS Phase 1 部署操作手册 (Runbook)

本手册详细描述了 AI-HMS 系统第一阶段（Phase 1）在 openEuler 服务器上的部署流程。Phase 1 核心目标是实现登录、患者列表及患者详情功能的上线，并对接真实的 PostgreSQL 数据库。

## Section 1: 环境概述

### 1.1 服务器信息
- **数据库服务器 (DB Server)**: `10.10.8.83`
  - 数据库类型：PostgreSQL
  - 数据库名称：`Postgre`
  - 用户名：`amdin`
  - 密码：由运维人员根据安全策略配置（占位符：`<fill in>`）
- **应用服务器 (App Server)**: `10.10.8.84`
  - 操作系统：openEuler 22.03 (LTS-SP2)

### 1.2 第一阶段范围
- **功能点**：登录 (Login) + 患者列表 (Patient List) + 患者详情 (Patient Detail)
- **数据源**：对接 10.10.8.83 上的真实 PostgreSQL 数据，禁用 Mock 数据。

### 1.3 系统架构
```text
[浏览器 (Browser)] 
       ↓
[前端 (Nginx:3000 @ 10.10.8.84)] 
       ↓
[后端 (Go:8080 @ 10.10.8.84)] 
       ↓
[数据库 (PostgreSQL:5432 @ 10.10.8.83)]
```

---

## Section 2: 前置条件

### 2.1 数据库服务器 (10.10.8.83) 要求
- PostgreSQL 服务已正常运行。
- 已创建名为 `Postgre` 的数据库。
- 已创建用户 `amdin` 并配置了正确的密码。
- **种子数据**：已执行 `ai-hms-backend/scripts/seed_phase1.sql`。该脚本包含 `test_admin` 初始用户及示例患者数据。

### 2.2 应用服务器 (10.10.8.84) 要求
- **Docker 部署（推荐）**：已安装 Docker 和 docker-compose。
- **Systemd 部署（备选）**：已安装 Go 1.24+ 和 Node.js 22+。
- **网络连通性**：应用服务器必须能够通过 5432 端口访问数据库服务器。

---

## Section 3: Docker 部署方式（推荐）

### 3.1 传输代码
将项目代码完整传输至应用服务器 `10.10.8.84`。

### 3.2 配置环境变量
在项目根目录下执行：
```bash
cp .env.production.template .env
```
编辑 `.env` 文件，填入以下真实值：

| 变量名 | 示例/说明 |
| :--- | :--- |
| `DB_HOST` | 10.10.8.83 |
| `DB_PORT` | 5432 |
| `DB_USER` | amdin |
| `DB_PASSWORD` | <fill in> |
| `DB_NAME` | Postgre |
| `DB_SSL_MODE` | disable |
| `JWT_SECRET` | 运行 `openssl rand -hex 32` 生成 |
| `JWT_EXPIRATION_HOURS` | 24 |
| `APP_SECRET` | 运行 `openssl rand -hex 32` 生成 |
| `CORS_ALLOWED_ORIGINS` | http://10.10.8.84:3000 |
| `SERVER_HOST` | 0.0.0.0 |
| `SERVER_PORT` | 8080 |
| `GIN_MODE` | release |
| `VITE_API_BASE_URL` | http://10.10.8.84:8080 |

### 3.3 初始化种子数据
如果尚未在数据库中执行种子数据脚本，请运行：
```bash
# 确保已在 .env 中设置了正确的 DB 变量
export DB_HOST=10.10.8.83
export DB_USER=amdin
export DB_PASS=<fill in>
export DB_NAME=Postgre
bash ai-hms-backend/scripts/seed_phase1.sh
```

### 3.4 构建与启动
```bash
docker compose build
docker compose up -d
```

### 3.5 验证启动状态
```bash
docker compose ps
docker compose logs backend
# 运行健康检查
curl http://10.10.8.84:8080/health
```

---

## Section 4: Systemd 部署方式（备选）

### 4.1 本地/流水线构建
1. **构建后端**：
   ```bash
   cd ai-hms-backend
   GOOS=linux GOARCH=amd64 go build -o ai-hms-backend ./cmd/server/
   ```
2. **构建前端**：
   ```bash
   cd ai-hms-frontend
   VITE_API_BASE_URL=http://10.10.8.84:8080 npm run build
   ```

### 4.2 传输文件
将后端二进制文件 `ai-hms-backend` 和前端构建产物 `dist` 目录传输至服务器。

### 4.3 配置环境
在服务器上创建运行用户及目录，并准备 `.env` 文件（变量列表详见 Section 3.2）。

### 4.4 安装后端服务
参考 `deploy/ai-hms-backend.service` 创建并安装 systemd 服务单元。
```bash
systemctl daemon-reload
systemctl enable ai-hms-backend
systemctl start ai-hms-backend
```

### 4.5 配置 Nginx 前端
参考 `ai-hms-frontend/nginx.conf` 配置 Nginx。确保 Nginx 监听 3000 端口，并正确指向 `dist` 目录。
```bash
systemctl restart nginx
```

---

## Section 5: 部署后验证

### 5.1 基础健康检查
执行：
```bash
curl http://10.10.8.84:8080/health
```
预期返回：`{"status":"UP"}` (或类似 200 OK 响应)。

### 5.2 运行冒烟测试
执行项目提供的自动化测试脚本：
```bash
bash ai-hms-backend/scripts/smoke_test.sh http://10.10.8.84:8080
```
该脚本将执行 8 项测试，涵盖：登录鉴权、获取患者列表、获取患者详情、未授权访问拦截等。

### 5.3 浏览器手动验证
1. 打开浏览器访问：`http://10.10.8.84:3000`。
2. 使用测试账号登录：
   - 用户名：`test_admin`
   - 密码：`Test@123456`
3. 验证进入患者列表页，且显示的是真实数据库数据（非 Mock）。
4. 点击任意患者，验证可正常跳转并显示患者详情。

---

## Section 6: 上线前检查清单

在宣布部署完成前，请逐一核对：

- [ ] DB server 10.10.8.83 可达 (ping/telnet 5432)
- [ ] DB `Postgre` 存在，用户 `amdin` 可连接
- [ ] `seed_phase1.sql` 已执行 (包含 `test_admin` 用户 + 样例患者)
- [ ] `.env` 所有必填项已填写（无 `<REQUIRED:` 或 `<generate-` 占位符残留）
- [ ] `JWT_SECRET` 已用 `openssl rand -hex 32` 生成强随机密钥
- [ ] `VITE_API_BASE_URL` 指向正确的后端地址
- [ ] `CORS_ALLOWED_ORIGINS` 包含前端访问地址
- [ ] Docker/systemd 服务已启动且 health check 返回 200
- [ ] `smoke_test.sh` 8/8 测试项全部通过
- [ ] 浏览器可访问 http://10.10.8.84:3000 并完成：登录 → 患者列表 → 患者详情闭环
- [ ] 患者详情页不再显示 Mock 数据（确认 `MOCK_PATIENTS` 已从主路径移除）
- [ ] 验证：使用错误密码登录被拒绝（返回 401）
- [ ] 验证：未携带 token 访问 `/api/v1/patients` 被拒绝（返回 401/403）

---

## Section 7: 回滚与停机

### 7.1 Docker 方式
- **停止服务**：`docker compose down`
- **回滚代码**：检出上一个稳定版本的代码/镜像，重新执行 `docker compose up -d`。

### 7.2 Systemd 方式
- **停止后端**：`systemctl stop ai-hms-backend`
- **启动后端**：`systemctl start ai-hms-backend`

### 7.3 数据库回滚
Phase 1 的种子数据脚本采用增量方式，不包含破坏性变更。如需清理环境，请手动清理对应的表数据。

---

## Section 8: 已知限制 (Phase 1)

以下功能在第一阶段尚未实现或存在局限性：
- 不包含 HDIS/LIS/设备系统集成。
- **数据兼容性问题**：由于 `Patient.ID` (varchar/UUID) 与 `Treatment.PatientId` (bigint) 类型不兼容，治疗记录关联功能暂不可用。
- **Mock 页面**：Dashboard、Schedule、Monitoring、Inventory、MasterData、DialysisProcessing 等页面仍维持 Mock 数据。
- **依赖项**：第一阶段无需安装 `chromedp/Chrome`。
- **自动化测试**：前端尚未集成 Playwright 自动化 e2e 测试，目前依赖手动浏览器验证。
