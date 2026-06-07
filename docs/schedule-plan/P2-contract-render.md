# P2 · 合约自动渲染（2-3 周 / 5 个 Step）

> 阶段目标：把 `Plan_PatientPlan` 升级为完整合约（频次 + 首选床位/班次/周几），实现"医生开方→护士落床位→自动渲染未来周排班"。
> 前置：P0-0 + P0 全部 Step 通过验收。
> **关键约束：仅扩展字段，不影响原有数据。** 三个新字段全部 `NULLABLE` 不带 `DEFAULT`；老数据保持 NULL，业务表现等于"未启用合约渲染"，与 P0 现状一致。

## 在开工前必须先读

1. `docs/schedule-plan/README.md`（老库红线 + 决策摘要）
2. 本文件全部内容
3. `docs/schedule-plan/RISK-REVIEW.md`（P2 风险补充，优先级高于本文件旧假设）
4. 根目录 `老血透数据库表结构-合并版.md`（核对 `Plan_PatientPlan` 与 `Schedule_*` 字段）
5. `ai-hms-backend/internal/services/patient_service.go`（看 `legacyPatientPlan` 结构体，约 24-75 行）
6. `ai-hms-backend/internal/services/patient_shift_service.go`（看 P0-0 后的周查询与待排班队列）
7. `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`（确认是否已有"首选床位"近似字段）

## P2 不要做的事

- ❌ 不要新建任何老库表
- ❌ 不要改 `Plan_PatientPlan` 现有字段（仅 ADD COLUMN）
- ❌ 不要在合约新字段为 NULL 时尝试渲染（NULL = 未启用，跳过该患者）
- ❌ 不要做 P3 的换床日志（那是另一份 SQL）
- ❌ 不要把 5 个 Step 合并到一个 PR

---

## Step 1 · DDL · 给 Plan_PatientPlan 加 3 个可空字段

### 改这些文件

- `ai-hms-backend/scripts/legacy_alter_plan_patientplan_add_preferred_fields.sql`（**新增**）
- `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`（追加映射段）
- SQL PR 描述或阶段记录（DBA 审核通过后记录执行结果；当前仓库无 `legacy-migration-confirmed-completed.md`）

### 怎么改

**A. 先核查老库是否已有近似字段**

```bash
# 在测试库执行（用户操作）：
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'Plan_PatientPlan'
  AND (column_name ILIKE '%bed%' OR column_name ILIKE '%shift%' OR column_name ILIKE '%weekday%')
ORDER BY ordinal_position;
```

把结果粘到 PR 描述中。如已有近似字段，**改本 Step 为 GORM tag 对齐**而不是新增字段，记到 `docs/legacy-migration-uncertain-field-checklist.md` 等用户确认。

**B. SQL 文件内容**

`ai-hms-backend/scripts/legacy_alter_plan_patientplan_add_preferred_fields.sql`：

```sql
-- 排班合约扩展字段（关联：docs/schedule-plan/P2-contract-render.md Step 1）
-- 全部为 NULLABLE 且不带 DEFAULT，老数据保持 NULL，业务表现 = "未启用合约渲染"
-- 必须由 DBA 审核执行，应用启动时禁止 AutoMigrate（已锁，参见 internal/database/migrate.go）

BEGIN;

-- 首选床位：护士落床位时回写
ALTER TABLE "Plan_PatientPlan"
  ADD COLUMN IF NOT EXISTS "PreferredBedId" BIGINT NULL;

-- 首选班次
ALTER TABLE "Plan_PatientPlan"
  ADD COLUMN IF NOT EXISTS "PreferredShiftId" BIGINT NULL;

-- 周几位掩码（bit0=周一 ... bit6=周日）
-- 例：0b0010101 = 21 → 周一三五；NULL → 未启用合约渲染
ALTER TABLE "Plan_PatientPlan"
  ADD COLUMN IF NOT EXISTS "PreferredWeekdayMask" SMALLINT NULL;

-- 索引（仅给非空值建，避免老数据全 NULL 影响）
CREATE INDEX IF NOT EXISTS "ix_Plan_PatientPlan_PreferredBedId"
  ON "Plan_PatientPlan" ("PreferredBedId")
  WHERE "PreferredBedId" IS NOT NULL;

COMMIT;
```

**C. LEGACY_TABLE_FIELD_MAPPING.md 追加**

```markdown
### Plan_PatientPlan（排班合约扩展）
- PreferredBedId    → 首选床位，BIGINT NULL，关联 Schedule_Bed.Id
- PreferredShiftId  → 首选班次，BIGINT NULL，关联 Schedule_Shift.Id
- PreferredWeekdayMask → 周几位掩码，SMALLINT NULL，bit0=周一 ... bit6=周日
- 三字段同时非空 → 启用合约自动渲染；任一为空 → 不渲染（与 P0 现状一致）
```

### 自检

- [ ] SQL 文件**只有 ALTER ADD COLUMN 与 CREATE INDEX**，无 DELETE/UPDATE/DROP
- [ ] 全部字段 `NULL`，无 `DEFAULT` 子句
- [ ] PR 标题 `[P2-1] DDL · Plan_PatientPlan 加首选字段`，影响面 `backend / deploy`
- [ ] PR 描述明确写明："本 PR 不部署应用，仅给 DBA 审核 SQL"
- [ ] 用户确认 + DBA 审核通过后才能合并

### 验收

- [ ] DBA 在测试库执行 SQL 成功
- [ ] `SELECT COUNT(*) FROM "Plan_PatientPlan" WHERE "PreferredBedId" IS NOT NULL;` 应为 0（老数据未受影响）
- [ ] 应用未做改动，跑 `verify.sh` 通过
- [ ] DBA 执行结果记录在 SQL PR 描述或阶段记录中

---

## Step 2 · 后端 GORM model 与查询字段对齐

### 前置

Step 1 SQL 已在测试库落地。

### 改这些文件

- `ai-hms-backend/internal/services/patient_service.go`（`legacyPatientPlan` 结构体加字段）
- `ai-hms-backend/internal/services/patient_shift_service.go`（如周查询需读首选字段，加 SELECT）
- `ai-hms-backend/internal/services/patient_core_service.go`（同理）

### 怎么改

**A. 给 `legacyPatientPlan` 加 3 个 GORM 字段**

```go
type legacyPatientPlan struct {
    // ...原字段...
    PreferredBedID         *int64 `gorm:"column:PreferredBedId"`
    PreferredShiftID       *int64 `gorm:"column:PreferredShiftId"`
    PreferredWeekdayMask   *int   `gorm:"column:PreferredWeekdayMask"`
}
```

**全部用指针类型**，区分"NULL"与"0"。

**B. 不要在现有 `getPatientList` / `getPatientDetail` 接口增加这三字段返回**（避免影响前端）。仅在 Step 3 的合约接口中读取。

**C. MVP 业务限制必须写进接口/页面说明**

P2 只支持“同一患者方案的一组首选床位 + 首选班次 + 周几规则”。如果出现同一方案需要按周几切换不同模式、不同床位或不同班次，不要强行塞入这三个字段，记录到 `docs/legacy-migration-uncertain-field-checklist.md`。

### 自检 grep

```bash
grep -n "PreferredBedID\|PreferredShiftID\|PreferredWeekdayMask" ai-hms-backend/internal/services
```

应仅在 `legacyPatientPlan` 与新合约接口出现。

### 验收

- [ ] 现有 patient API 返回结构无变化
- [ ] `verify.sh` 通过
- [ ] 单测：构造一条 `Plan_PatientPlan` 三字段全 NULL 数据，能正常 Find 不报错
- [ ] 单测：构造三字段全填数据，能正常 Find 并解出指针值

---

## Step 3 · 合约 API · 查询、修改首选字段

### 改这些文件

- `ai-hms-backend/internal/services/schedule_contract_service.go`（**新增**）
- `ai-hms-backend/internal/api/v1/schedule_contract_handler.go`（**新增**）
- `ai-hms-backend/cmd/server/main.go`（注册 `RegisterScheduleContractRoutes`）
- `ai-hms-frontend/src/services/contractApi.ts`（**新增** 客户端包装）

### 怎么改

**A. 服务层接口**

```go
type ScheduleContractService struct{ db *gorm.DB }

// 列出"待落位合约"：有 Plan_PatientPlan 但 Preferred* 任一为 NULL
func (s *ScheduleContractService) ListPendingAssignment(tenantID int64) ([]ContractDTO, error)

// 列出"已落位合约"：Preferred* 三字段全部非 NULL
func (s *ScheduleContractService) ListAssigned(tenantID int64, wardID *int64) ([]ContractDTO, error)

// 单个查询
func (s *ScheduleContractService) Get(planID, tenantID int64) (*ContractDTO, error)

// 落位（护士填首选床位/班次/周几）
func (s *ScheduleContractService) AssignPreferences(planID, tenantID int64, req AssignPreferencesRequest) error

// 取消落位（清空 Preferred*）
func (s *ScheduleContractService) UnassignPreferences(planID, tenantID int64) error
```

**B. 路由**

```
GET    /api/v1/schedule/contracts                     ?status=pending|assigned&wardId=
GET    /api/v1/schedule/contracts/:id
PUT    /api/v1/schedule/contracts/:id/preferences     落位
DELETE /api/v1/schedule/contracts/:id/preferences     取消落位
```

> 沿用项目响应封装 `pkg/response`。

**C. 校验**

`AssignPreferences` 请求体：

```go
type AssignPreferencesRequest struct {
    PreferredBedID       int64 `json:"preferredBedId" binding:"required,gt=0"`
    PreferredShiftID     int64 `json:"preferredShiftId" binding:"required,gt=0"`
    PreferredWeekdayMask int   `json:"preferredWeekdayMask" binding:"required,min=1,max=127"` // 至少选1天
}
```

校验：床位/班次启用状态、床位归属病区、频次（OddWeek + EvenWeek）与 mask 中"1"的位数一致或差异在 ±1（请假补排容差）。差异超过容差 → 警告（前端弹确认），不强制拒绝。

**D. CSV 导入暂不做**（P3 处理）。

### 验收

- [ ] `GET /api/v1/schedule/contracts?status=pending` 返回所有"医生已开方但护士未落位"的患者
- [ ] `PUT /api/v1/schedule/contracts/:id/preferences` 写回三字段成功
- [ ] DELETE 端点清空三字段
- [ ] 单测覆盖：落位成功、参数缺失拒绝、mask 与频次不一致返回警告
- [ ] `verify.sh` 通过

---

## Step 4 · 前端 · 待落位侧栏 + 落位弹窗

### 改这些文件

- `ai-hms-frontend/src/pages/Schedule.tsx`（右侧栏增加"合约待落位" Tab）
- `ai-hms-frontend/src/components/schedule/AssignPreferencesModal.tsx`（**新增**，< 300 行）
- `ai-hms-frontend/src/services/contractApi.ts`（Step 3 已建）

### 怎么改

**A. 右侧栏增加 Tab**

当前已有"待排班队列"。新增第二个 Tab："合约待落位"。

```
[ 待排班 (12) ]  [ 合约待落位 (3) ]   ← 数字徽章
```

数据源：`GET /api/v1/schedule/contracts?status=pending&wardId={当前筛选病区}`

**B. 待落位项展示**

每项显示：患者姓名 + 频次（如"单3/双2"）+ 模式 + "落位"按钮。

**C. 落位弹窗 `AssignPreferencesModal`**

字段：

- 首选床位（按当前筛选病区分组的下拉，排除停用）
- 首选班次（启用班次下拉）
- 周几（7 个 Checkbox，按频次约束）
  - 频次 = 3 则建议勾 3 天，prevent 勾少于 2 / 多于 4
  - 频次 = 2 则建议 2 天，限制 1-3 天
  - 双周频次不一致时显示提示
- 保存按钮

实现要点：

- 周几用 `weekdayMask = (周一?1:0) | (周二?2:0) | (周三?4:0) | ... | (周日?64:0)` 计算
- 校验失败显示警告 banner，按"仍要保存"按钮才提交

**D. 落位完成后**

待落位列表自动减 1；周视图本周如已渲染过，提示"已落位，下周起自动生成排班"。

### 验收

- [ ] 待落位 Tab 数量准确（与后端 `status=pending` 一致）
- [ ] 落位弹窗 7 个 Checkbox 与频次校验正确
- [ ] 保存后患者从待落位列表消失
- [ ] 截图：`docs/ui-snapshots/schedule-P2/step-4-{tab,modal,saved}.png`
- [ ] `npm run lint && npm run build` 通过

---

## Step 5 · 后端 · 周渲染服务（幂等）

### 改这些文件

- `ai-hms-backend/internal/services/schedule_render_service.go`（**新增**）
- `ai-hms-backend/internal/api/v1/schedule_handler.go`（增加 `POST /api/v1/patient-shifts/render-week` 端点）
- `ai-hms-frontend/src/pages/Schedule.tsx`（顶部增加"生成本周排班"按钮）
- `ai-hms-backend/cmd/server/main.go`（如需注册定时任务，**本 Step 不做**，留 P3 评估）

### 怎么改

**A. 服务层逻辑**

```go
// 渲染指定周的合约位（幂等）
// 入参：mondayDate（周一日期）+ tenantID + 可选 wardID
// 行为：
//   1. 查所有"已落位"合约（PreferredBedId/ShiftId/WeekdayMask 全非空）
//   2. 计算本周 ISO 周号奇偶 → 取 OddWeek/EvenWeek 频次
//   3. 按 WeekdayMask 投影到具体日期，先做患者维度幂等检查
//      患者同日同班已存在未取消记录 → skippedByPatient
//   4. 再做床位维度冲突检查
//      同床同日同班被其他患者占用 → conflictedByBed
//   5. 均无冲突 → 插入 ShiftTiming=20 长期 Status=10 草稿
//   6. 返回 {created:N, skippedByPatient:N, conflictedByBed:[...]}
func (s *ScheduleRenderService) RenderWeek(req RenderWeekRequest) (*RenderWeekResult, error)
```

**关键**：

- 幂等：以 `TenantId + PatientId + DATE(TreatmentTime) + ShiftId + Status not canceled` 判断患者同日同班已存在
- 床位冲突：以 `TenantId + BedId + DATE(TreatmentTime) + ShiftId + Status not canceled` 判断同一格是否被其他患者占用
- 不修改已存在记录（即使状态为草稿）

**B. 路由**

```
POST /api/v1/patient-shifts/render-week
Body: { mondayDate: "2026-06-01", wardId?: 12 }
Resp: { success: true, data: { created: 23, skippedByPatient: 5, conflictedByBed: [...] } }
```

**C. 前端按钮**

Schedule 顶部"生成本周排班"按钮：

- 点击 → 弹确认对话框（"将根据合约自动生成 N 条草稿排班，已存在的不会修改。继续？"）
- 调 `POST /render-week`
- 成功 toast：`新建 23 条 / 患者已存在跳过 5 条 / 3 条床位冲突`
- 冲突列表展示在 Toast 下方"查看详情"

**D. 不做**

- 定时任务自动渲染（留 P3 评估，先手动触发）
- 节假日特殊周覆盖（YAGNI，等业务真的提出）
- 模式按周几切换（你答复"分两步走"，等查测试库后决定）

### 验收

- [ ] 给 5 个测试患者落位（不同频次/模式/床位），调一次 `render-week`，正确插入 13-15 条排班
- [ ] 再调一次同样的请求 → `created=0, skippedByPatient=N`（幂等）
- [ ] 故意制造床位冲突 → `conflictedByBed` 数组返回正确
- [ ] 老数据（Preferred* 全 NULL 的患者）不被渲染、不影响
- [ ] 单测覆盖：奇偶周频次切换、mask 解码、幂等、冲突
- [ ] 截图：`docs/ui-snapshots/schedule-P2/step-5-{button,confirm,result-toast}.png`
- [ ] `verify.sh` + `npm run build` 通过

---

## P2 完工核对

- [ ] 5 个 Step 全部合并
- [ ] `Plan_PatientPlan` 仅新增 3 个 NULL 字段，老数据零变化（再次 SELECT COUNT 确认）
- [ ] 待落位流程闭环（医生开方 → Plan 出现在待落位 → 护士落位 → 渲染未来周）
- [ ] 现有手动排班流程不受影响（不落位的患者仍按原方式排班）
- [ ] 截图归档：`docs/ui-snapshots/schedule-P2/`

通过后开 `docs/schedule-plan/P3-log-and-import.md`。
