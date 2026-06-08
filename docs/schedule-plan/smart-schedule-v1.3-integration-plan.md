# 透析排班 v1.3 一体化融合开发计划

## 1. 背景与目标

将 `F:\python\前后端代码\ai-hms_qhd\docs\排班功能说明\透析排班-backend-v1.3\backend` 中的智能排班 v1.3 功能完整融合进当前主系统：

- 后端融合进 `ai-hms-backend`，共用一个 Gin 进程、一个端口 `:8080`、一个数据库连接。
- 前端融合进 `ai-hms-frontend` 的现有菜单和路由。
- 不再单独启动 `:8081`。
- 不再使用反向代理。
- 不再使用 iframe。
- 不影响当前主系统已有 `/api/v1` 排班、患者、透析执行等功能。
- 尽量保留 v1.3 的算法、服务、生命周期、院感、CRRT、冲突队列等功能。

## 2. 当前系统关键约束

| 约束 | 说明 |
|---|---|
| 主系统已有 `Schedule_PatientShift` | 位于 `ai-hms-backend/internal/models/schedule.go`，不能被 v1.3 覆盖 |
| 主系统已有 `Schedule_PatientProfile` | 位于 `ai-hms-backend/internal/models/schedule_ext.go`，不能被 v1.3 覆盖 |
| 主系统已有 `/api/v1` | v1.3 API 不能继续挂载到 `/api/v1` |
| 主系统已有 `/health` | v1.3 的 `/health` 必须移除 |
| 主系统已有 `RequestLogger` | v1.3 的 `auditMiddleware()` 不再使用 |
| 主系统禁止 AutoMigrate | 不允许引入 v1.3 的 `db.Migrate()` / `AutoMigrate()` |
| 主系统数据库大小写敏感 | 继续使用主系统 GORM 配置：`SingularTable: true`、`NoLowerCase: true` |

## 3. v1.3 相比 v1.2 的关键增量

| 决策 | 变更 | 涉及位置 |
|---|---|---|
| 决策 23 | `PatientProfile` 新增 `WeeklyCount` 字段和校验 | `models.go`、`admin_service.go` |
| 决策 24 | 新增治疗模式 `HFD` / `HP`，扩展机器能力矩阵 | `constants.go`、`engine.go` |
| 决策 25 | `ScheduleTemplateItem` 新增 `DefaultMode`，`DecideMode` 增加 `baseMode` 参数 | `models.go`、`week.go` |
| 决策 26 | 新增院感状态和护士长豁免 | `models.go`、`lifecycle_service.go`、`api_admin.go` |
| 决策 27 | 新增出组/中途入组生命周期 | `models.go`、`lifecycle_service.go`、`schedule_service.go` |
| 决策 29 | 新增资料待补提示 | `lifecycle_service.go`、`api_admin.go` |

## 4. 最终目录结构

```text
ai-hms-backend/
├── cmd/server/main.go
├── internal/
│   └── smart_schedule/
│       ├── model/
│       │   └── models.go
│       ├── config/
│       │   └── config.go
│       ├── repo/
│       │   └── repo.go
│       ├── seed/
│       │   └── seed.go
│       ├── sched/
│       │   ├── board.go
│       │   ├── constants.go
│       │   ├── engine.go
│       │   ├── newpatient.go
│       │   ├── util.go
│       │   └── week.go
│       ├── service/
│       │   ├── admin_service.go
│       │   ├── crrt_service.go
│       │   ├── diff_service.go
│       │   ├── lifecycle_service.go
│       │   ├── makeup_service.go
│       │   ├── ops_service.go
│       │   ├── perturb_service.go
│       │   ├── quality_service.go
│       │   ├── schedule_service.go
│       │   ├── template_service.go
│       │   └── weekview.go
│       └── api/
│           ├── api.go
│           └── api_admin.go
```

```text
ai-hms-frontend/
├── src/pages/SmartSchedulePage.tsx
├── src/services/smartScheduleApi.ts
```

## 5. 包依赖设计

保持 v1.3 原有子包结构，不要把多个包合并成一个根包，否则会产生自导入和循环引用问题。

```text
model ← sched ← config
      ← repo ← sched
      ← seed ← sched
      ← service ← config + model + repo + sched
      ← api ← model + sched + seed + service
```

## 6. 表名隔离策略

所有 v1.3 表必须统一改为 `Schedule_v2_*`，不能使用 v1.3 原始的 `Schedule_*` 表名。

| 结构体 | 最终表名 |
|---|---|
| `Ward` | `Schedule_v2_Ward` |
| `Machine` | `Schedule_v2_Machine` |
| `MachineOutage` | `Schedule_v2_MachineOutage` |
| `Shift` | `Schedule_v2_Shift` |
| `Calendar` | `Schedule_v2_Calendar` |
| `PatientProfile` | `Schedule_v2_PatientProfile` |
| `PlanChange` | `Schedule_v2_PlanChange` |
| `PatientShift` | `Schedule_v2_PatientShift` |
| `CrrtSession` | `Schedule_v2_CrrtSession` |
| `ScheduleTemplate` | `Schedule_v2_ScheduleTemplate` |
| `ScheduleTemplateItem` | `Schedule_v2_ScheduleTemplateItem` |
| `ConflictQueue` | `Schedule_v2_ConflictQueue` |
| `Patient` | `Schedule_v2_Patient` |
| `TenantSetting` | `Schedule_v2_TenantSetting` |

必须全文搜索并修正硬编码 SQL 表名：

| 原硬编码 | 新硬编码 |
|---|---|
| `"Schedule_PatientShift"` | `"Schedule_v2_PatientShift"` |
| `"Schedule_CrrtSession"` | `"Schedule_v2_CrrtSession"` |
| `"Schedule_PatientProfile"` | `"Schedule_v2_PatientProfile"` |
| `"Schedule_ScheduleTemplate"` | `"Schedule_v2_ScheduleTemplate"` |
| `"Schedule_ScheduleTemplateItem"` | `"Schedule_v2_ScheduleTemplateItem"` |
| `"Schedule_ConflictQueue"` | `"Schedule_v2_ConflictQueue"` |
| `"Schedule_Machine"` | `"Schedule_v2_Machine"` |
| `"Schedule_Ward"` | `"Schedule_v2_Ward"` |
| `"Schedule_Shift"` | `"Schedule_v2_Shift"` |

重点检查：

```text
ai-hms-backend/internal/smart_schedule/service/crrt_service.go
ai-hms-backend/internal/smart_schedule/service/admin_service.go
ai-hms-backend/internal/smart_schedule/repo/repo.go
```

## 7. 分阶段执行计划

### 阶段 P0：清理旧残留

删除：

```text
ai-hms-backend/internal/smart_schedule/
ai-hms-backend/internal/api/v1/smart_schedule_proxy.go
```

从 `ai-hms-backend/cmd/server/main.go` 删除：

```go
v1api.RegisterSmartScheduleProxy(r, jwtManager)
```

### 阶段 P1：复制 v1.3 源码

源目录：

```text
F:\python\前后端代码\ai-hms_qhd\docs\排班功能说明\透析排班-backend-v1.3\backend
```

复制规则：

| v1.3 源路径 | 主系统目标路径 |
|---|---|
| `internal/model/models.go` | `internal/smart_schedule/model/models.go` |
| `internal/config/config.go` | `internal/smart_schedule/config/config.go` |
| `internal/repo/repo.go` | `internal/smart_schedule/repo/repo.go` |
| `internal/seed/seed.go` | `internal/smart_schedule/seed/seed.go` |
| `internal/sched/*.go` | `internal/smart_schedule/sched/*.go` |
| `internal/service/*.go` | `internal/smart_schedule/service/*.go` |
| `internal/api/api.go` | `internal/smart_schedule/api/api.go` |
| `internal/api/api_admin.go` | `internal/smart_schedule/api/api_admin.go` |

不复制：

```text
cmd/server/main.go
internal/db/db.go
*_test.go
web/index.html
```

### 阶段 P2：修正 import 路径

批量替换：

| 原 import | 新 import |
|---|---|
| `github.com/sdsph/dialysis-scheduling/internal/model` | `github.com/elliotxin/ai-hms-backend/internal/smart_schedule/model` |
| `github.com/sdsph/dialysis-scheduling/internal/config` | `github.com/elliotxin/ai-hms-backend/internal/smart_schedule/config` |
| `github.com/sdsph/dialysis-scheduling/internal/repo` | `github.com/elliotxin/ai-hms-backend/internal/smart_schedule/repo` |
| `github.com/sdsph/dialysis-scheduling/internal/sched` | `github.com/elliotxin/ai-hms-backend/internal/smart_schedule/sched` |
| `github.com/sdsph/dialysis-scheduling/internal/service` | `github.com/elliotxin/ai-hms-backend/internal/smart_schedule/service` |
| `github.com/sdsph/dialysis-scheduling/internal/seed` | `github.com/elliotxin/ai-hms-backend/internal/smart_schedule/seed` |

删除：

```go
"github.com/sdsph/dialysis-scheduling/internal/db"
```

### 阶段 P3：修正模型表名

在 `internal/smart_schedule/model/models.go` 中逐个修改 `TableName()`，必须全部使用 `Schedule_v2_*`。

特别注意：

- `PlanChange` 必须是 `Schedule_v2_PlanChange`。
- `Patient` 必须是 `Schedule_v2_Patient`。
- `TenantSetting` 必须是 `Schedule_v2_TenantSetting`。

### 阶段 P4：修正硬编码 SQL 表名

全文搜索 `Schedule_`，凡是属于智能排班 v1.3 的表名，全部替换为 `Schedule_v2_*`。

尤其检查 CRRT join：

```sql
JOIN "Schedule_PatientShift" ps
```

必须改为：

```sql
JOIN "Schedule_v2_PatientShift" ps
```

### 阶段 P5：新增索引创建函数

在 `internal/smart_schedule/repo/repo.go` 中新增：

```go
func EnsureIndexes(g *gorm.DB) error {
	stmts := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_v2_ps_patient_slot
		 ON "Schedule_v2_PatientShift" ("TenantId","PatientId","ScheduleDate","ShiftId")
		 WHERE "Status" NOT IN (70,80) AND "ShiftId" IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_v2_ps_machine_slot
		 ON "Schedule_v2_PatientShift" ("TenantId","MachineId","ScheduleDate","ShiftId")
		 WHERE "Status" NOT IN (70,80) AND "MachineId" IS NOT NULL AND "ShiftId" IS NOT NULL`,
	}
	for _, s := range stmts {
		if err := g.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}
```

要求：

- 禁止 AutoMigrate。
- 禁止 `DropTable`。
- 索引名必须带 `v2`，避免与主系统索引冲突。

### 阶段 P6：重构 `api/api.go`

目标：v1.3 API 只注册到主系统 `/api/v2` 下，不再管理全局 Engine。

修改项：

| 原逻辑 | 新逻辑 |
|---|---|
| `func (s *Server) Register(r *gin.Engine)` | `func (s *Server) Register(rg *gin.RouterGroup)` |
| `r.Use(auditMiddleware())` | 删除 |
| `r.GET("/health", s.health)` | 删除 |
| `r.StaticFile("/", "web/index.html")` | 删除 |
| `v1 := r.Group("/api/v1")` | 删除 |
| `v1.GET(...)` | 改为 `rg.GET(...)` |
| `s.registerAdmin(v1)` | 改为 `s.registerAdmin(rg)` |

新增构造函数：

```go
func NewServer(db *gorm.DB) *Server {
	return &Server{DB: db}
}
```

`Register()` 最终形态应接近：

```go
func (s *Server) Register(rg *gin.RouterGroup) {
	if err := repo.EnsureIndexes(s.DB); err != nil {
		log.Printf("[smart_schedule] ensure indexes failed: %v", err)
	}

	rg.Use(tenantMiddleware())

	rg.GET("/schedule/week", s.weekView)
	rg.GET("/schedule/board", s.board)
	rg.GET("/schedule/diffs", s.diffs)
	rg.GET("/schedule/quality", s.quality)
	rg.GET("/conflicts", s.listConflicts)

	rg.POST("/schedule/generate", guard(RoleHeadNurse, RoleChargeNurse), s.generate)
	rg.POST("/admin/seed", guard(RoleHeadNurse), s.seedDemo)

	rg.POST("/schedule/confirm-plan", guard(RoleHeadNurse), s.confirmPlan)
	rg.POST("/schedule/confirm-day", guard(RoleHeadNurse, RoleChargeNurse), s.confirmDay)

	rg.POST("/shifts/:id/cancel", guard(RoleHeadNurse, RoleChargeNurse), s.cancelShift)
	rg.POST("/shifts/:id/absent", guard(RoleHeadNurse, RoleChargeNurse), s.absentShift)
	rg.POST("/shifts/:id/move", guard(RoleHeadNurse, RoleChargeNurse), s.moveShift)

	rg.POST("/schedule/temporary", guard(RoleDoctor, RoleHeadNurse, RoleChargeNurse), s.insertTemporary)
	rg.POST("/schedule/crrt", guard(RoleDoctor, RoleHeadNurse, RoleChargeNurse), s.insertCrrt)
	rg.GET("/schedule/crrt", s.listCrrt)

	rg.POST("/machines/:id/outage", guard(RoleHeadNurse), s.machineOutage)
	rg.POST("/schedule/holiday", guard(RoleHeadNurse), s.setHoliday)
	rg.POST("/patients/:id/plan-change", guard(RoleDoctor, RoleHeadNurse), s.planChange)
	rg.POST("/patients/:id/makeup", guard(RoleHeadNurse, RoleChargeNurse), s.makeup)

	s.registerAdmin(rg)
}
```

### 阶段 P7：改造鉴权上下文

`tenantMiddleware()` 不再读取 Header，改为读取主系统 JWT context。

新增 import：

```go
"github.com/elliotxin/ai-hms-backend/internal/middleware"
```

目标实现：

```go
func tenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := middleware.GetTenantID(c)
		if tenant <= 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "error": "缺少有效租户"})
			return
		}

		c.Set("tenant", tenant)
		c.Set("role", mapRole(middleware.GetRoles(c)))
		c.Set("userId", middleware.GetUserID(c))
		c.Next()
	}
}
```

角色映射：

```go
func mapRole(roles []string) string {
	for _, r := range roles {
		switch r {
		case "ADMIN", "管理员", "安全管理员", "运维管理员":
			return RoleHeadNurse
		case "DOCTOR", "医生":
			return RoleDoctor
		case "HEAD_NURSE", "护士长":
			return RoleHeadNurse
		case "CHARGE_NURSE", "主班护士":
			return RoleChargeNurse
		case "NURSE", "护士":
			return RoleNurse
		}
	}
	return ""
}
```

`userOf()` 改为：

```go
func userOf(c *gin.Context) int64 {
	if v := c.GetString("userId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return 0
}
```

### 阶段 P8：注册 `/api/v2`

在 `ai-hms-backend/cmd/server/main.go` 中新增 import：

```go
smartapi "github.com/elliotxin/ai-hms-backend/internal/smart_schedule/api"
```

新增注册逻辑：

```go
smartSchedule := r.Group("/api/v2")
smartSchedule.Use(middleware.AuthMiddleware(jwtManager))
smartapi.NewServer(database.GetDB()).Register(smartSchedule)
```

最终 API 路径：

| 功能 | 路径 |
|---|---|
| 周视图 | `GET /api/v2/schedule/week` |
| 排班矩阵 | `GET /api/v2/schedule/board` |
| 生成排班 | `POST /api/v2/schedule/generate` |
| 整盘确认 | `POST /api/v2/schedule/confirm-plan` |
| 当日确认 | `POST /api/v2/schedule/confirm-day` |
| 移床 | `POST /api/v2/shifts/:id/move` |
| CRRT | `POST /api/v2/schedule/crrt` |
| 冲突列表 | `GET /api/v2/conflicts` |
| 出组 | `POST /api/v2/patients/:id/discharge` |
| 中途入组 | `POST /api/v2/patients/:id/place` |
| 院感状态 | `POST /api/v2/patients/:id/infection` |
| 院感豁免 | `POST /api/v2/patients/:id/infection-waive` |

### 阶段 P9：生成 DDL SQL

新增：

```text
docs/sql/v1.3_v2_tables.sql
```

要求：

- 使用 `CREATE TABLE IF NOT EXISTS`。
- 使用 `ALTER TABLE ADD COLUMN IF NOT EXISTS`。
- 使用 `CREATE UNIQUE INDEX IF NOT EXISTS`。
- 禁止 `DROP TABLE`。
- 禁止修改主系统旧表。
- 禁止 AutoMigrate。

必须包含 v1.3 新字段：

| 表 | 字段 |
|---|---|
| `Schedule_v2_PatientProfile` | `WeeklyCount` |
| `Schedule_v2_PatientProfile` | `PatientStatus` |
| `Schedule_v2_PatientProfile` | `DischargeReason` |
| `Schedule_v2_PatientProfile` | `DischargedAt` |
| `Schedule_v2_PatientProfile` | `DischargedBy` |
| `Schedule_v2_ScheduleTemplateItem` | `DefaultMode` |
| `Schedule_v2_Patient` | `InfectionStatus` |
| `Schedule_v2_Patient` | `InfectionWaivedBy` |
| `Schedule_v2_Patient` | `InfectionWaivedAt` |

### 阶段 P10：新增前端 API 服务

新增：

```text
ai-hms-frontend/src/services/smartScheduleApi.ts
```

封装 `/api/v2`：

| 函数 | API |
|---|---|
| `getBoard(date)` | `GET /api/v2/schedule/board` |
| `getWeek(date)` | `GET /api/v2/schedule/week` |
| `generateSchedule(payload)` | `POST /api/v2/schedule/generate` |
| `confirmPlan(payload)` | `POST /api/v2/schedule/confirm-plan` |
| `confirmDay(payload)` | `POST /api/v2/schedule/confirm-day` |
| `cancelShift(id, payload)` | `POST /api/v2/shifts/:id/cancel` |
| `moveShift(id, payload)` | `POST /api/v2/shifts/:id/move` |
| `insertTemporary(payload)` | `POST /api/v2/schedule/temporary` |
| `insertCrrt(payload)` | `POST /api/v2/schedule/crrt` |
| `listConflicts()` | `GET /api/v2/conflicts` |
| `getDiffs(params)` | `GET /api/v2/schedule/diffs` |
| `getQuality(params)` | `GET /api/v2/schedule/quality` |
| `listWards()` | `GET /api/v2/admin/wards` |
| `listMachines()` | `GET /api/v2/admin/machines` |
| `listPatients()` | `GET /api/v2/admin/patients` |
| `listProfiles()` | `GET /api/v2/admin/profiles` |
| `placePatient(id, payload)` | `POST /api/v2/patients/:id/place` |
| `dischargePatient(id, payload)` | `POST /api/v2/patients/:id/discharge` |
| `setInfection(id, payload)` | `POST /api/v2/patients/:id/infection` |
| `waiveInfection(id)` | `POST /api/v2/patients/:id/infection-waive` |
| `listIncompleteProfiles()` | `GET /api/v2/admin/incomplete-profiles` |

### 阶段 P11：重写 `SmartSchedulePage.tsx`

当前文件：

```text
ai-hms-frontend/src/pages/SmartSchedulePage.tsx
```

必须删除 iframe。

页面功能区域：

| 区域 | 功能 |
|---|---|
| 顶部工具栏 | 日期选择、生成排班、确认按钮 |
| 质量评分卡 | 调用 `/api/v2/schedule/quality` |
| 周排班矩阵 | 调用 `/api/v2/schedule/board` |
| 冲突队列 | 调用 `/api/v2/conflicts` |
| 差异检测 | 调用 `/api/v2/schedule/diffs` |
| CRRT 面板 | 调用 `/api/v2/schedule/crrt` |
| 管理面板 | 病区、机器、病人、骨架、模板 |
| 生命周期 | 出组、入组、院感状态、资料待补 |

第一阶段可先实现核心能力：

1. 日期选择。
2. 排班矩阵展示。
3. 生成排班。
4. 确认排班。
5. 冲突队列。
6. 质量评分。
7. 资料待补列表。

后续再补全拖拽移床、CRRT 管理、模板维护等高级交互。

## 8. 必须核对的 v1.3 生命周期代码

整合前必须读取并逐行核对：

```text
F:\python\前后端代码\ai-hms_qhd\docs\排班功能说明\透析排班-backend-v1.3\backend\internal\service\lifecycle_service.go
F:\python\前后端代码\ai-hms_qhd\docs\排班功能说明\透析排班-backend-v1.3\backend\internal\api\api_admin.go
```

核对函数签名：

| API 功能 | service 函数 |
|---|---|
| 出组 | `DischargePatient` |
| 中途入组 | `PlaceNewPatientService` |
| 设置院感 | `SetInfectionStatus` |
| 院感豁免 | `WaiveInfection` |
| 资料待补 | `ListIncompleteProfiles` |

## 9. 安全要求

| 风险 | 处理 |
|---|---|
| `POST /admin/seed` 无权限 | 必须加 `guard(RoleHeadNurse)` |
| `DEV_SUPERUSER=true` | 默认关闭，启动时打印 WARN |
| 无角色访问 | 返回 401 |
| 租户缺失 | 返回 403 |
| 自动 DDL | 禁止 AutoMigrate，仅保留 SQL 文件 |

如保留 `DEV_SUPERUSER`，启动时增加警告：

```go
if strings.EqualFold(os.Getenv("DEV_SUPERUSER"), "true") {
	log.Println("[WARN] DEV_SUPERUSER=true, 智能排班开发模式已启用（生产环境请勿设置）")
}
```

## 10. 验证步骤

### 后端验证

在 `ai-hms-backend` 执行：

```powershell
go build -o "$env:TEMP\ai-hms-check.exe" ./cmd/server
```

如果迁移了测试文件，再执行：

```powershell
go test ./internal/smart_schedule/sched ./internal/smart_schedule/service
```

### 前端验证

在 `ai-hms-frontend` 执行：

```powershell
npm run lint
npm run build
```

### 接口验证

| 请求 | 预期 |
|---|---|
| `GET /health` | 主系统健康检查正常 |
| `GET /api/v2/schedule/board` | 登录后返回排班矩阵 |
| `POST /api/v2/schedule/generate` | 护士长/主班可生成 |
| `GET /api/v2/admin/incomplete-profiles` | 返回资料待补列表 |
| `POST /api/v2/patients/:id/discharge` | 医生/护士长可出组 |
| `GET /api/v1/schedule/...` | 主系统原排班功能不受影响 |

## 11. 最终执行顺序

| 阶段 | 内容 |
|---|---|
| P0 | 删除旧 `smart_schedule` 和 `smart_schedule_proxy.go` |
| P1 | 复制 v1.3 源码到 7 个子包 |
| P2 | 批量替换 import |
| P3 | 全部表名改为 `Schedule_v2_*` |
| P4 | 修正硬编码 SQL 表名 |
| P5 | 新增 `repo.EnsureIndexes()` |
| P6 | 重构 `api.Register()` 为 `*gin.RouterGroup` |
| P7 | 改造 JWT 鉴权、角色映射、`userOf()` |
| P8 | 修改 `main.go` 注册 `/api/v2` |
| P9 | 生成 `docs/sql/v1.3_v2_tables.sql` |
| P10 | 前端新增 `smartScheduleApi.ts` |
| P11 | 重写 `SmartSchedulePage.tsx`，去除 iframe |
| P12 | 执行后端 build |
| P13 | 执行前端 lint/build |
| P14 | 手工验证主系统 `/api/v1` 与新系统 `/api/v2` |

## 12. 完成标准

- `ai-hms-backend` 编译通过。
- `ai-hms-frontend` lint/build 通过。
- `/api/v1` 主系统原功能不受影响。
- `/api/v2` 智能排班接口可通过主系统 JWT 鉴权访问。
- `Schedule_PatientShift`、`Schedule_PatientProfile` 等主系统旧表没有任何结构或数据变更。
- 所有 v1.3 智能排班数据写入 `Schedule_v2_*` 表。
- 前端 `/smart-schedule` 页面不再使用 iframe。
