# 透析排班子程序 — 后端

AI 透析管理系统的「患者透析周排班」子程序后端。技术栈:**Go + Gin + GORM + PostgreSQL**。
为每位透析病人安排未来 2~4 周的透析治疗(在哪个区、哪台机器、哪个班次),并处理三级确认、
临时透析、停机迁移、假日挪班、方案变更、差异检测、补透、CRRT 等全套临床场景。

> 本 README 面向接手测试的工程师。业务规则的权威说明见上层目录的
> `透析排班子程序规则规范_v1.md` 与 `透析排班设计_数据模型与算法_v1.md`。

---

## 1. 环境要求

| 组件 | 版本 |
|------|------|
| Go | **1.25+**(见 go.mod) |
| PostgreSQL | **14+** |

数据库连接通过环境变量 `DATABASE_URL` 提供,**密码不写在代码里**。

---

## 2. 快速开始

### 2.1 拉依赖

```bash
cd backend
go mod tidy
```

### 2.2 跑单元测试(不需要数据库)

验证排班算法核心(周/奇偶周、HDF 替换判定、两轮分配、机型能力等):

```bash
go test ./internal/sched/ -v
```

预期:全部 PASS(8 个测试)。

### 2.3 建库 + 启动服务

先在 PostgreSQL 里建一个**空库**(表由程序自动创建,无需手动建表):

```sql
CREATE DATABASE aihms;
```

设置连接串并启动(Windows PowerShell):

```powershell
$env:DATABASE_URL="host=127.0.0.1 user=postgres password=你的密码 dbname=aihms port=5432 sslmode=disable TimeZone=Asia/Shanghai"
go run ./cmd/server
```

Linux/macOS:

```bash
export DATABASE_URL="host=127.0.0.1 user=postgres password=你的密码 dbname=aihms port=5432 sslmode=disable TimeZone=Asia/Shanghai"
go run ./cmd/server
```

看到 `🚀 透析排班服务监听 :8080` 即启动成功(程序会先自动迁移所有 `Schedule_*` 表)。

可选:`LISTEN_ADDR` 改监听地址(默认 `:8080`)。

### 2.4 浏览器测试

打开 **http://localhost:8080**:

1. 点「写演示数据」→ 写入 3 区 / 3 班 / 14 台机 / 7 病人 / 1 模板。
2. 起始周一选 `2025-01-06`(或任意周一)→ 点「生成排班」→ 矩阵里出现排班。
3. 试各功能:三级确认、设为假日、方案变更、＋临时透析、＋CRRT、切角色;
   点格子弹菜单(取消/缺席/移动)、**拖动格子移床**、差异面板「一键补排」。

---

## 3. 目录结构

```
backend/
├── cmd/server/main.go          入口:连库 + 自动迁移 + 启动 Gin
├── web/index.html              前端(React+Tailwind via CDN,同源托管)
└── internal/
    ├── model/models.go         13 张 Schedule_* 表(GORM 模型)
    ├── db/db.go                连接 + AutoMigrate
    ├── config/config.go        租户级配置(奇偶周基准周一等)
    ├── sched/                  ★排班算法核心(纯逻辑,可单测)
    │   ├── constants.go        枚举/机型能力/频率→透析日
    │   ├── week.go             周序号/奇偶周/HDF替换判定
    │   ├── board.go            内存快照(资源+占用)
    │   ├── engine.go           主流程:两轮分配/HDF双固定/挑机/顺延
    │   ├── newpatient.go       新病人初始机位
    │   └── *_test.go           单元测试
    ├── repo/repo.go            持久化:加载快照 + 回写草稿/冲突
    ├── service/                业务编排
    │   ├── schedule_service.go 生成排班
    │   ├── weekview.go         周视图聚合
    │   ├── ops_service.go      三级确认/取消/缺席/移床
    │   ├── perturb_service.go  临时透析/停机迁移/假日/方案变更
    │   ├── diff_service.go     差异检测(应排vs已排)
    │   ├── makeup_service.go   补透
    │   └── crrt_service.go     CRRT 落位
    ├── api/api.go              HTTP 路由 + 租户/角色中间件
    └── seed/seed.go            演示数据
```

---

## 4. HTTP 接口一览(前缀 `/api/v1`)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/seed` | 写演示数据(空库时) |
| POST | `/schedule/generate` | 生成 2/4 周草稿 `{startDate, weeks}` |
| GET  | `/schedule/board?date=` | 周视图矩阵(区→机器→班×日) |
| GET  | `/schedule/week?date=` | 某周原始排班记录 |
| GET  | `/schedule/diffs?date=&weeks=` | 应排/已排差异 |
| GET  | `/conflicts?status=0` | 冲突/待处理队列 |
| POST | `/schedule/confirm-plan` | ① 整盘确认(护士长)`{weekStart, weeks}` |
| POST | `/schedule/confirm-day` | ②③ 次日/当日确认 `{date, level}` |
| POST | `/shifts/:id/cancel` | 取消(请假)`{reason}` |
| POST | `/shifts/:id/absent` | 当日缺席 `{reason}` |
| POST | `/shifts/:id/move` | 移床/换班 `{machineId, date, shiftId}` |
| POST | `/schedule/temporary` | 临时透析 `{patientId, wardId, date, mode}` |
| POST | `/schedule/crrt` | CRRT `{patientId, wardId, startAt, endAt}` |
| GET  | `/schedule/crrt?date=` | CRRT 占用列表 |
| POST | `/machines/:id/outage` | 登记停机 `{start, end, type, reason}` |
| POST | `/schedule/holiday` | 设为假日 `{date, mode}` |
| POST | `/patients/:id/plan-change` | 方案变更 `{changeType, newValue, effectiveDate}` |
| POST | `/patients/:id/makeup` | 一键补透 `{weekStart, weeks}` |

curl 示例:

```bash
curl -X POST localhost:8080/api/v1/admin/seed
curl -X POST localhost:8080/api/v1/schedule/generate \
     -H "Content-Type: application/json" -d '{"startDate":"2025-01-06","weeks":2}'
curl "localhost:8080/api/v1/schedule/board?date=2025-01-06"
```

---

## 5. 多租户与权限

- **租户**:请求头 `X-Tenant-Id`(缺省 `1`)。
- **角色**:请求头 `X-Role`,取值 `doctor` / `head_nurse` / `charge_nurse` / `nurse`。
  - **不传 X-Role = 超级用户,放行所有**(便于联调)。
  - 守卫规则(规范 §11):整盘确认=护士长;次/当日确认、取消/缺席/移床=护士长/主班;
    临时透析/CRRT=医生/护士长/主班;停机/假日=护士长;方案变更=医生/护士长;补透=护士长/主班。
- 网页右上角「角色」下拉可切换,用于测试权限拦截(返回 403)。

> 这是演示用的轻量鉴权。接入正式系统时应替换为 `ai-hms` 的鉴权中间件,
> 并把本地 `Schedule_Patient` 病人主档对接到老库 `Register_PatientInfomation`。

---

## 6. 数据库说明

- 所有表在 `Schedule_` 命名空间,PostgreSQL 双引号标识符。
- 启动时 GORM **AutoMigrate** 自动建表/补列,**只需一个空库**。
- 核心表 `Schedule_PatientShift`:一行=一位病人一次透析槽位;状态机
  `0待排/10草稿/20已确认/50透析中/60已完成/70已取消/80缺席`,三级确认用
  `Confirm1/2/3At` 时间戳;`SourceType` 区分常规/临时;`DialysisMode` 按次存 HD/HDF/CRRT;
  模板独立存表(不复用状态值)。
- 配置项(`Schedule_TenantSetting`):`OddEvenWeekAnchorMonday` 奇偶周基准周一
  (默认 2025-01-06,服务 HDF「每两周一次」错峰)。

---

## 7. 常见问题

- **`go` 不是命令**:新开一个终端窗口(刚装的 Go,PATH 需新窗口生效)。
- **连不上数据库**:确认 PostgreSQL 服务在跑、库已建、`DATABASE_URL` 的用户/密码/端口正确。
- **页面空白**:前端用 CDN 加载 React/Tailwind,需要能访问外网;或检查浏览器控制台报错。
- **想要干净数据**:删库重建(`DROP DATABASE aihms; CREATE DATABASE aihms;`)再启动 + 重新「写演示数据」。
- **时间显示**:CRRT 等时间按服务器本地时区解释;DSN 里建议带 `TimeZone=Asia/Shanghai`。

---

## 8. 验证状态

- `go build ./...` / `go vet ./...` 通过;`go test ./internal/sched/` 8/8 通过。
- 全部接口已在真实 PostgreSQL 上端到端验证(生成、三级确认、取消/移床、临时透析、
  停机迁移、假日挪班、方案变更、差异检测、补透、CRRT、权限拦截)。
