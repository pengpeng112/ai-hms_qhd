# 智能排班系统完整交接文档

> 目标读者：接手此项目的 AI/开发者  
> 生成时间：2026-06-08  
> 仓库：`pengpeng112/ai-hms_qhd`  
> 状态：v1.3 代码已完成融合，但与老数据库尚未对接

---

## 1. 仓库与分支

| 项目 | 值 |
|------|-----|
| GitHub | `https://github.com/pengpeng112/ai-hms_qhd.git` |
| 包含 v1.3 代码的分支 | `opencode/brave-moon`（已推送 origin） |
| 用户实际工作分支 | `fix/legacy-ui-restore`（尚无 v1.3 代码） |
| 两分支差距 | `opencode/brave-moon` 领先 2 个提交 |
| 当前环境 worktree | `C:\Users\Administrator\.local\share\opencode\worktree\e467b55f745b230c1550c74277856634892086b9\brave-moon` |
| 后端入口 | `ai-hms-backend/cmd/server/main.go` |
| 前端入口 | `ai-hms-frontend/`（Vite + React 19 + TypeScript + Ant Design 6） |

### 分支合并方式

```bash
cd 你的项目根目录
git checkout fix/legacy-ui-restore
git merge opencode/brave-moon
```

---

## 2. 项目架构

```
brave-moon/
├── ai-hms-backend/          # Go 后端（Gin + GORM + PostgreSQL）
│   ├── cmd/server/main.go   # 入口，注册所有路由
│   ├── internal/
│   │   ├── smart_schedule/  # ★ v1.3 智能排班模块（核心交付物）
│   │   │   ├── api/         # /api/v2 所有端点（api.go + api_admin.go）
│   │   │   ├── config/      # 配置常量
│   │   │   ├── model/       # 14 张 v2 表 GORM 模型
│   │   │   ├── repo/        # 数据访问 + 唯一索引
│   │   │   ├── sched/       # 调度引擎（两轮分配/HDF/奇偶周/新病人）
│   │   │   ├── seed/        # 演示数据写入
│   │   │   └── service/     # 11 个业务 service
│   │   ├── api/v1/          # 老系统 /api/v1 路由
│   │   ├── services/        # 老系统业务层
│   │   ├── models/          # 老系统 GORM 模型
│   │   ├── middleware/      # JWT/CORS/日志中间件
│   │   └── database/        # 数据库连接层（AutoMigrate 已永久禁用）
│   └── docs/sql/            # DDL 文件
│       ├── v1.3_v2_tables.sql   # 14 张 Schedule_v2_* 表 DDL
│       └── v1.2_v2_tables.sql   # 旧版
├── ai-hms-frontend/         # React 前端
│   ├── src/pages/SmartSchedulePage.tsx   # ★ 智能排班页面
│   ├── src/services/smartScheduleApi.ts  # ★ /api/v2 API 调用层
│   ├── src/router.tsx                    # 路由注册（包含 /smart-schedule）
│   └── src/layouts/Sidebar.tsx           # 侧边栏菜单（含智能排班入口）
├── docs/
│   ├── schedule-plan/                    # 排班方案文档
│   │   ├── smart-schedule-v1.3-integration-plan.md    # v1.3 集成方案详情
│   │   └── smart-schedule-legacy-data-sync-plan.md    # 老库对接方案（★待开发）
│   └── sql/v1.3_v2_tables.sql
├── 老血透数据库表结构-合并版.md           # 老库 102 张表完整定义
└── 新数据库表结构.md                      # 新表结构（参考）
```

---

## 3. 当前已提交的内容（opencode/brave-moon HEAD: 5ad3054）

### 3.1 后端 — 23 个 Go 源文件

| 文件 | 行数 | 说明 |
|------|------|------|
| `cmd/server/main.go` | 修改 | 第 95-97 行注册 `/api/v2` 路由组 |
| `internal/smart_schedule/api/api.go` | ~736 | 主 Handler：42 个端点 + JWT 鉴权 + responseWrapper |
| `internal/smart_schedule/api/api_admin.go` | ~300 | 管理端点：Ward/Machine/Patient/Profile/Template CRUD |
| `internal/smart_schedule/config/config.go` | ~30 | AnchorMonday、DraftWeeks 等 4 个配置函数 |
| `internal/smart_schedule/model/models.go` | 258 | 14 张 v2 表 GORM 模型 |
| `internal/smart_schedule/repo/repo.go` | ~130 | LoadBoard、SaveDrafts、SaveConflicts、EnsureIndexes |
| `internal/smart_schedule/sched/engine.go` | ~400 | **核心引擎**：两轮分配、HDF 溢出、排满顺延 |
| `internal/smart_schedule/sched/board.go` | ~200 | Board 内存快照构建 |
| `internal/smart_schedule/sched/week.go` | ~200 | 周序号、奇偶周、频率判定 |
| `internal/smart_schedule/sched/newpatient.go` | ~100 | 新病人中途入组 |
| `internal/smart_schedule/sched/constants.go` | ~80 | 全部常量/枚举 |
| `internal/smart_schedule/sched/util.go` | ~40 | 工具函数 |
| `internal/smart_schedule/seed/seed.go` | ~100 | 7 个假病人演示数据 |
| `internal/smart_schedule/service/schedule_service.go` | ~90 | GenerateSchedule 主流程 |
| `internal/smart_schedule/service/ops_service.go` | ~200 | 三级确认、取消、缺席、移床 |
| `internal/smart_schedule/service/perturb_service.go` | ~350 | 临时透析、停机迁移、假日、方案变更 |
| `internal/smart_schedule/service/diff_service.go` | ~80 | 应排/已排差异检测 |
| `internal/smart_schedule/service/makeup_service.go` | ~60 | 补透 |
| `internal/smart_schedule/service/quality_service.go` | ~80 | 质量评分 |
| `internal/smart_schedule/service/weekview.go` | ~100 | 周视图聚合矩阵 |
| `internal/smart_schedule/service/admin_service.go` | ~300 | 资源 CRUD + 模板重建 |
| `internal/smart_schedule/service/template_service.go` | ~80 | 模板相关 |
| `internal/smart_schedule/service/crrt_service.go` | ~100 | CRRT 插入与列表 |
| `internal/smart_schedule/service/lifecycle_service.go` | ~120 | 出组、入组、院感、豁免 |

### 3.2 前端 — 关键文件

| 文件 | 说明 |
|------|------|
| `src/pages/SmartSchedulePage.tsx` | 完整页面：周板矩阵、拖拽移动、右键菜单、质量评分、CRRT 面板、冲突列表、管理弹窗 |
| `src/services/smartScheduleApi.ts` | 40+ 个 API 函数，全部指向 `/api/v2` |
| `src/router.tsx` | 已添加 `/smart-schedule` 路由 |
| `src/layouts/Sidebar.tsx` | 已添加"智能排班"菜单项 |
| `src/i18n/locales/*/nav.json` | 已添加中英双语 |

### 3.3 SQL/文档

| 文件 | 说明 |
|------|------|
| `docs/sql/v1.3_v2_tables.sql` | 14 张 `Schedule_v2_*` 表的 DDL（IF NOT EXISTS，幂等） |
| `docs/schedule-plan/smart-schedule-v1.3-integration-plan.md` | 完整集成方案 |

---

## 4. 未提交内容（需提交）

| 文件 | 说明 | 重要性 |
|------|------|--------|
| `api/api.go`（修改） | `userOf()` 默认值 0→1 | 低：修复审计人 ID |
| `SmartSchedulePage.tsx`（修改） | 前端增强（拖拽/模态框等） | 中 |
| `sched/engine_test.go`（新） | 3 个引擎单元测试 | **高**：核心算法验证 |
| `sched/week_test.go`（新） | 5 个时间/频率单元测试 | **高** |
| `service/integration_test.go`（新） | 11 个集成测试（需 DB） | **高** |
| `docs/.../smart-schedule-legacy-data-sync-plan.md`（新） | 老库对接方案 | **高**：下一步工作依据 |

**提交命令**：
```bash
git add -A
git commit -m "feat: 补充测试文件 + userOf修复 + 老库对接方案文档"
git push origin opencode/brave-moon
```

---

## 5. 运行时验证结果

### 5.1 编译状态

| 项目 | 命令 | 结果 |
|------|------|------|
| 后端编译 | `cd ai-hms-backend && go build -o check.exe ./cmd/server` | ✅ 通过 |
| 前端编译 | `cd ai-hms-frontend && npm run build` | ✅ 通过（tsc + vite） |
| 单元测试 | `cd ai-hms-backend && go test ./internal/smart_schedule/sched/` | ✅ 8/8 PASS |
| 集成测试 | `cd ai-hms-backend && go test ./internal/smart_schedule/service/` | ✅ 11 SKIP（需 TEST_DATABASE_URL） |

### 5.2 运行状态

| 服务 | 地址 | 状态 |
|------|------|------|
| 后端 | `http://localhost:8080` | ✅ `/health` 返回 200 |
| 前端 | `http://localhost:5173` | ✅ 页面可访问 |
| 登录 | `test_admin` / `Test@123456` | ✅ 可用 |
| `/api/v2` 路由 | 42+ 个端点 | ✅ 全部注册 |

---

## 6. 数据库

### 6.1 连接信息（.env）

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=admin123
DB_NAME=ai_hms_db
```

### 6.2 老库表（主系统使用，v2 模块不读）

| 表 | 用途 |
|----|------|
| `Register_PatientInfomation` | 患者主档（42 字段） |
| `Plan_PatientPlan` | 透析方案（48 字段，含 OddWeekFrequency/EvenWeekFrequency/DialysisMethod） |
| `Schedule_Ward` | 病区 |
| `Schedule_Bed` | 床位 |
| `Schedule_Shift` | 班次 |
| `Schedule_PatientShift` | 患者排班记录 |
| `Schedule_BedEquipmentRel` | 床位-设备关联 |
| `Auxiliary_EquipmentInfomation` | 设备档案（含 DialysisMethod/Flux/Type） |
| `Register_Infection` | 传染病信息 |

### 6.3 v2 表（智能排班独用，与老库零耦合）

14 张 `Schedule_v2_*` 表，需手动执行 DDL 创建：

| 表 | 说明 |
|----|------|
| `Schedule_v2_Ward` | 病区（含 ZoneType A/B/C） |
| `Schedule_v2_Machine` | 机位（含 MachineType HD/HDF/CRRT） |
| `Schedule_v2_MachineOutage` | 机器停机 |
| `Schedule_v2_Shift` | 班次（含 ShiftCode） |
| `Schedule_v2_Calendar` | 机构日历/假日 |
| `Schedule_v2_Patient` | 患者轻量主档 |
| `Schedule_v2_PatientProfile` | 患者排班骨架（★核心：Frequency/Mode/Hdf*/Zone） |
| `Schedule_v2_PlanChange` | 方案变更记录 |
| `Schedule_v2_PatientShift` | 排班记录（★核心输出表） |
| `Schedule_v2_CrrtSession` | CRRT 会话 |
| `Schedule_v2_ScheduleTemplate` | 模板头 |
| `Schedule_v2_ScheduleTemplateItem` | 模板项 |
| `Schedule_v2_ConflictQueue` | 冲突队列 |
| `Schedule_v2_TenantSetting` | 租户配置 |

**重要**：数据库 `AutoMigrate` 已被 `internal/database/migrate.go` 永久禁用。创建表必须在 PostgreSQL 中手动执行 `docs/sql/v1.3_v2_tables.sql`。

唯一索引（`repo.EnsureIndexes()` 在 `/api/v2` 路由注册时自动创建）：
- `uq_v2_ps_patient_slot`：同一患者同日同班次不重复
- `uq_v2_ps_machine_slot`：同一机位同日同班次不被双占

---

## 7. 老库对接 — 待开发工作

### 7.1 核心问题

**v2 智能排班系统与老数据库完全隔离**：
- v2 只读写 `Schedule_v2_*` 表
- 不读取任何老表（`Register_PatientInfomation`、`Plan_PatientPlan`、`Schedule_Ward` 等）
- 当前只能用 `POST /api/v2/admin/seed` 写入 7 个假病人演示排班
- 无法基于真实患者数据生成排班

### 7.2 需要开发的 5 个同步组件

| # | 组件 | 老表来源 | v2 目标表 | 复杂度 | 关键推断 |
|----|------|---------|----------|--------|---------|
| 1 | `SyncWards` | `Schedule_Ward` | `Schedule_v2_Ward` | 低 | ZoneType（A/B/C） |
| 2 | `SyncShifts` | `Schedule_Shift` | `Schedule_v2_Shift` | 低 | ShiftCode（MORNING/AFTERNOON/NIGHT） |
| 3 | `SyncMachines` | `Schedule_Bed` + `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` | `Schedule_v2_Machine` | **高** | MachineType、Code 组装 |
| 4 | `SyncPatients` | `Register_PatientInfomation` + `Plan_PatientPlan` + `Register_Infection` | `Schedule_v2_Patient` + `Schedule_v2_PatientProfile` | **最高** | FreqPattern、DefaultMode、HdfWeekday、ZoneTag、HomeWardId |
| 5 | `RebuildTemplates` | `Schedule_v2_PatientProfile`（已有） | `Schedule_v2_ScheduleTemplate` + `Schedule_v2_ScheduleTemplateItem` | 低 | 直接复制（已有函数） |

### 7.3 待新增文件

| 文件 | 说明 |
|------|------|
| `internal/smart_schedule/service/sync_service.go` | 核心同步逻辑（~400 行） |
| `internal/smart_schedule/api/api_sync.go` | `POST /api/v2/sync/legacy` 端点（~100 行） |
| `internal/smart_schedule/service/sync_service_test.go` | 同步逻辑测试（~200 行） |
| `cmd/sync/main.go` | CLI 独立同步工具（~80 行） |

### 7.4 8 个待人工决策的问题

| # | 问题 | 建议默认值 | 影响 |
|---|------|-----------|------|
| 1 | 奇数周/偶数周频率不同时取大值还是平均值？ | 取大值 | `WeeklyCount` |
| 2 | 每周3次是否总是一三五？ | 默认一三五 | `FreqPattern` |
| 3 | ZoneType 推断规则是否正确？ | 长期+普通→A，传染病→C，临时→B | `Ward.ZoneType` |
| 4 | HDF 日默认周三是否合理？ | 默认3（周三） | `HdfWeekday` |
| 5 | "HD+HP"复合模式如何映射？ | 取主模式 HD | `DefaultMode` |
| 6 | 是否从历史推算固定机位？ | 不同步（NULL） | `FixedHdMachineId` |
| 7 | 老排班记录是否迁移到 v2？ | 不迁移 | 从头生成 |
| 8 | 同步所有租户还是指定？ | 按请求指定，默认 1 | `TenantId` |

### 7.5 执行顺序

```
SyncWards() → SyncShifts() → SyncMachines() → SyncPatients() → RebuildTemplates()
                                                                          ↓
                                                              POST /api/v2/schedule/generate
                                                                  （用户手动触发）
```

---

## 8. 与 v1.3 原始代码的差异对照

原始代码位于用户本地：`F:\python\前后端代码\ai-hms_qhd\docs\排班功能说明\透析排班-backend-v1.3\backend\internal\`

| 层面 | 兼容性 | 说明 |
|------|--------|------|
| model（14 张表） | ✅ 100% | 仅 TableName 加 `v2_` 前缀 |
| sched（调度引擎） | ✅ 100% | 算法完全一致 |
| service（11 个） | ✅ 100% | 业务逻辑一致，SQL 中表名已改 |
| api（路由注册） | ✅ 95% | Register 从 *Engine→*RouterGroup；新增 responseWrapper；删除了 auditMiddleware（主系统已有） |
| 测试文件 | ⚠️ 0% → **已补** | 已迁移 engine_test.go + week_test.go + integration_test.go |

---

## 9. 已知问题

| 问题 | 严重度 | 状态 |
|------|--------|------|
| v2 不读取老库数据 → 无法基于真实患者生成排班 | **最高** | 待开发（见第 7 节） |
| 未合并到用户工作分支 `fix/legacy-ui-restore` | **高** | 待合并 |
| `Schedule_v2_*` 表可能未在用户数据库创建 | **高** | 需手动执行 DDL |
| 测试文件未提交 | 中 | 已写，待 commit |
| 前端 SmartSchedulePage 未提交 | 中 | 已修改，待 commit |

---

## 10. 快速启动命令

```powershell
# 1. 创建 v2 数据库表（在 PostgreSQL 中执行一次）
psql -h localhost -U admin -d ai_hms_db -f docs/sql/v1.3_v2_tables.sql

# 2. 启动后端
cd ai-hms-backend
go run ./cmd/server

# 3. 启动前端（新终端窗口）
cd ai-hms-frontend
npm run dev

# 4. 浏览器访问
# http://localhost:5173
# 登录：test_admin / Test@123456
# 左侧菜单 → "智能排班"

# 5. 跑测试
cd ai-hms-backend
go test ./internal/smart_schedule/sched/         # 单元测试（无需 DB）
go test ./internal/smart_schedule/service/       # 集成测试（需 TEST_DATABASE_URL）
```

---

## 11. 关键绝对路径

```
# 后端核心
C:\Users\Administrator\.local\share\opencode\worktree\e467b55f745b230c1550c74277856634892086b9\brave-moon\ai-hms-backend\internal\smart_schedule\

# 前端页面
C:\Users\Administrator\.local\share\opencode\worktree\e467b55f745b230c1550c74277856634892086b9\brave-moon\ai-hms-frontend\src\pages\SmartSchedulePage.tsx

# DDL
C:\Users\Administrator\.local\share\opencode\worktree\e467b55f745b230c1550c74277856634892086b9\brave-moon\docs\sql\v1.3_v2_tables.sql

# 老库对接方案（待复核）
C:\Users\Administrator\.local\share\opencode\worktree\e467b55f745b230c1550c74277856634892086b9\brave-moon\docs\schedule-plan\smart-schedule-legacy-data-sync-plan.md

# 老库表结构参考
C:\Users\Administrator\.local\share\opencode\worktree\e467b55f745b230c1550c74277856634892086b9\brave-moon\老血透数据库表结构-合并版.md

# 原始 v1.3 独立版（对比用）
F:\python\前后端代码\ai-hms_qhd\docs\排班功能说明\透析排班-backend-v1.3\backend\internal\
```
