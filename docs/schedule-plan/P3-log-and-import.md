# P3 · 调整原因标记 + CSV 导入 + 报表（1-2 周 / 4 个 Step）

> 阶段目标：给 `Schedule_PatientShift` 加 `AdjustReason` 字段，让最近一次调整可分类；提供 CSV 批量导入合约；输出"床位利用率/合约执行率"两张报表。
> 前置：P0 + P2 全部 Step 通过验收。
> **关键约束：仅扩展 1 个字段，老数据保持 NULL。** AdjustReason NULL = 历史数据未分类，业务表现等同于 P0/P2 现状。

## 在开工前必须先读

1. `docs/schedule-plan/README.md`（老库红线 + 决策摘要）
2. 本文件全部内容
3. `docs/schedule-plan/RISK-REVIEW.md`（P3 风险补充，优先级高于本文件旧假设）
4. 根目录 `老血透数据库表结构-合并版.md`（核对 `Schedule_PatientShift` 表结构）
5. `ai-hms-backend/internal/services/patient_shift_service.go`（看 P0-0 后的 `/move` `/swap` 端点）
6. `ai-hms-backend/internal/models/schedule.go`

## P3 不要做的事

- ❌ 不要新建 `Schedule_PatientShiftLog` 表（用户答复"只加 AdjustReason 字段"）
- ❌ 不要把 4 个 Step 合并到一个 PR
- ❌ 不要做"历史排班反推合约"（P4 项，超出本计划）

---

## Step 1 · DDL · Schedule_PatientShift 加 AdjustReason

### 改这些文件

- `ai-hms-backend/scripts/legacy_alter_schedule_patientshift_add_adjustreason.sql`（**新增**）
- `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`（追加映射段）
- SQL PR 描述或阶段记录（DBA 审核通过后记录执行结果；当前仓库无 `legacy-migration-confirmed-completed.md`）

### 怎么改

**A. 先核查老库是否已有近似字段**

```sql
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'Schedule_PatientShift'
  AND (column_name ILIKE '%reason%' OR column_name ILIKE '%adjust%' OR column_name ILIKE '%change%');
```

如已有 `Reason` / `AdjustType` 等近似字段，**改为 GORM tag 对齐**，记到 `docs/legacy-migration-uncertain-field-checklist.md` 让用户裁定。

**B. SQL 文件**

```sql
-- 排班调整原因字段（关联：docs/schedule-plan/P3-log-and-import.md Step 1）
-- NULLABLE 不带 DEFAULT；老数据保持 NULL，业务表现 = "未分类"，与 P0/P2 现状一致

BEGIN;

ALTER TABLE "Schedule_PatientShift"
  ADD COLUMN IF NOT EXISTS "AdjustReason" VARCHAR(64) NULL;

-- 仅给非空记录建索引
CREATE INDEX IF NOT EXISTS "ix_Schedule_PatientShift_AdjustReason"
  ON "Schedule_PatientShift" ("AdjustReason")
  WHERE "AdjustReason" IS NOT NULL;

COMMIT;
```

**C. 字典枚举**（不入库，应用层常量）

```go
// internal/services/schedule_constants.go
const (
    AdjustReasonBedMove    = "bed_move"     // 换床
    AdjustReasonDayChange  = "day_change"   // 换日
    AdjustReasonModeChange = "mode_change"  // 换模式
    AdjustReasonMakeup     = "makeup"       // 请假补排
    AdjustReasonSwap       = "swap"         // 互换
    AdjustReasonOther      = "other"        // 其他
)
```

**D. LEGACY_TABLE_FIELD_MAPPING.md 追加**

```markdown
### Schedule_PatientShift（调整原因扩展）
- AdjustReason  → 排班调整原因，VARCHAR(64) NULL
- 取值：bed_move / day_change / mode_change / makeup / swap / other / NULL（未分类）
- NULL 不视为错误，仅影响"最近调整"等基于此字段的过滤
```

### 验收

- [ ] SQL 文件**只有 ALTER ADD + CREATE INDEX**，无 DELETE/UPDATE/DROP
- [ ] DBA 审核通过后落地测试库
- [ ] `SELECT COUNT(*) FROM "Schedule_PatientShift" WHERE "AdjustReason" IS NOT NULL;` = 0（老数据未受影响）
- [ ] DBA 执行结果记录在 SQL PR 描述或阶段记录中

---

## Step 2 · /move /swap 写 AdjustReason + 最近调整展示

### 改这些文件

- `ai-hms-backend/internal/services/patient_shift_service.go`（`/move` `/swap` 写入 `AdjustReason`）
- `ai-hms-backend/internal/api/v1/patient_shift_handler.go`（如有需要，扩展请求体）
- `ai-hms-backend/internal/models/schedule.go`（GORM model 加字段）
- `ai-hms-frontend/src/pages/Schedule.tsx`（最近调整展示改为按 AdjustReason 过滤）

### 怎么改

**A. GORM model**

```go
type PatientShift struct {
    // ...原字段...
    AdjustReason *string `gorm:"column:AdjustReason;type:varchar(64)" json:"adjustReason,omitempty"`
}
```

**B. /move 端点**

```go
// 现有 Move 函数尾部 Update 时加：
updates["AdjustReason"] = AdjustReasonBedMove
updates["LastModifyTime"] = time.Now()
```

**C. /swap 端点**

互换两条记录时各写 `AdjustReason = AdjustReasonSwap`。

**D. 前端"最近调整"弹窗**

旧：拉全量 PatientShift 历史显示。
新：

```ts
restApi.getPatientShiftHistory(patientId, { adjustReason: ['bed_move', 'swap'] })
```

后端如无 `adjustReason` 过滤参数，新增一个 query 参数 `?adjustReason=bed_move,swap`。

**E. UI 标签**

不要承诺完整换床历史。每行只能显示当前床位/班次、调整原因、`LastModifyTime` 等当前行可得信息。弹窗顶部必须加注："仅显示当前记录最近一次调整原因，更早的调整历史无法追溯"。

### 验收

- [ ] 拖拽换床后，对应 PatientShift 行 `AdjustReason='bed_move'`、`LastModifyTime` 更新
- [ ] /swap 后两条 `AdjustReason='swap'`
- [ ] 最近调整弹窗只显示有 AdjustReason 的记录
- [ ] 单测：构造 NULL AdjustReason 的老数据，不进 /move 路径，仍可正常显示
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 3 · 合约 CSV 批量导入

### 改这些文件

- `ai-hms-backend/internal/services/schedule_contract_service.go`（增加 `ImportContracts`）
- `ai-hms-backend/internal/api/v1/schedule_contract_handler.go`（增加 `POST /api/v1/schedule/contracts/import`）
- `ai-hms-frontend/src/pages/Schedule.tsx` 或新页 `ContractImport.tsx`（导入入口）
- `docs/schedule-contract-csv-template.md`（**新增** CSV 模板说明）

### 怎么改

**A. CSV 模板**

8 列：

```
病历号,姓名,频次单周,频次双周,首选床位编码,首选班次编码,周几（1=周一，逗号分隔）,默认透析模式
P0001,张三,3,3,A-101,早班,"1,3,5",HD
P0002,李四,2,2,A-102,中班,"1,4",HDF
```

`周几（1=周一，逗号分隔）` 填字符串 `1,3,5`（更直观），后端解析为整型 mask。不要把模板列名写成“周几掩码”，否则用户会误以为应填写 `21` 这类二进制掩码值。

**B. 后端解析**

```go
type ImportContractsRequest struct {
    Items []ContractImportRow `json:"items" binding:"required,min=1,max=500"`
}

type ContractImportRow struct {
    PatientCode      string `json:"patientCode"`
    PatientName      string `json:"patientName"`
    OddWeekFrequency int    `json:"oddWeekFrequency"`
    EvenWeekFrequency int   `json:"evenWeekFrequency"`
    PreferredBedCode  string `json:"preferredBedCode"`
    PreferredShiftCode string `json:"preferredShiftCode"`
    Weekdays          string `json:"weekdays"`            // "1,3,5"
    DialysisMethod    string `json:"dialysisMethod"`
}

func (s *ScheduleContractService) ImportContracts(tenantID int64, req ImportContractsRequest) (*ImportResult, error)
```

返回结果：

```go
type ImportResult struct {
    Total     int                  `json:"total"`
    Succeeded int                  `json:"succeeded"`
    Failed    []ImportFailureItem  `json:"failed"`
}

type ImportFailureItem struct {
    Row    int    `json:"row"`     // CSV 行号
    Reason string `json:"reason"`  // 失败原因
}
```

**逐行处理**，失败的行不影响其他行（不开事务包整个导入）。每行内部开小事务保证原子。

**C. 失败原因枚举**

- 患者编码不存在 / 多匹配
- 床位编码不存在 / 已停用
- 班次编码不存在
- weekdays 与频次不一致
- 患者已有合约（同一 Plan_PatientPlan 已落位）→ 跳过 + 记录"已存在"

**D. 前端入口**

待落位 Tab 顶部按钮"批量导入"：

- 上传 CSV
- 前端解析（用 `xlsx` 已有依赖）→ 显示预览表格 → 用户点"确认导入"
- 调 `POST /import` → 显示结果 Toast：`成功 N / 失败 M（点击查看详情）`

### 验收

- [ ] CSV 模板可下载
- [ ] 部分失败时已成功的行已落库，失败的行可重试
- [ ] 单测覆盖：所有失败原因
- [ ] 截图：`docs/ui-snapshots/schedule-P3/step-3-{template,preview,result}.png`
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 4 · 报表 · 床位利用率 + 合约执行率

### 改这些文件

- `ai-hms-backend/internal/services/schedule_report_service.go`（**新增**）
- `ai-hms-backend/internal/api/v1/schedule_report_handler.go`（**新增**）
- `ai-hms-frontend/src/pages/ScheduleReport.tsx`（**新增**，挂在统计报表下）
- `ai-hms-frontend/src/router.tsx`、`Sidebar.tsx`、`routeMeta.ts`、`nav.json`

### 怎么改

**A. 接口**

```
GET /api/v1/schedule/reports/bed-utilization?from=2026-05-01&to=2026-05-31&wardId=
  Resp: [{ bedId, bedName, wardName, totalSlots, usedSlots, utilizationRate }, ...]

GET /api/v1/schedule/reports/contract-execution?from=...&to=...&wardId=
  Resp: [{ patientId, patientName, expectedTimes, actualTimes, executionRate }, ...]
```

**B. 计算口径**

**床位利用率**：
- `totalSlots` = 该日期范围内 `Schedule_Bed.IsDisabled=false` 的床位 × 启用班次数 × 日历天数
- `usedSlots` = `Schedule_PatientShift` 在该范围内的有效占用去重计数；具体状态集合必须以 P0-0 确认后的状态口径为准
- `utilizationRate` = used / total

**合约执行率**：
- `expectedTimes` = 按合约 OddWeek/EvenWeek 频次推算的应排次数
- `actualTimes` = 实际有效排班次数；具体状态集合必须以 P0-0 确认后的 `Schedule_PatientShift.Status` 口径为准
- `executionRate` = actual / expected

**注意**：
- 不计 `ShiftTiming=10` 临时排班到 actualTimes（合约执行率只看长期）
- 床位利用率分子分母都按"床位 × 班次"计算，不按时长

**C. 前端展示**

报表页用 `recharts`：

- 床位利用率：横向条形图（按床位排序）+ 表格
- 合约执行率：表格 + 异常筛选（执行率 < 80% 标红）

**D. 性能**

如老库 PatientShift 行数过大，加：

- 强制日期范围 ≤ 90 天
- 使用索引 `(TenantId, TreatmentTime, Status)`（如不存在加进 P3 SQL，但 Step 1 已合并，需要在本 Step 加另一份独立 DDL PR）

### 验收

- [ ] P0-0 已确认 `Schedule_PatientShift.Status` 状态口径，报表中完成/取消/草稿集合与口径一致
- [ ] 5 月 1 日-31 日报表能正常计算（验证总应排数 vs 总实际数）
- [ ] 利用率数据与人工抽查 3 个床位的样本一致
- [ ] 合约执行率为 100% 的患者排在前；异常患者标红
- [ ] 截图：`docs/ui-snapshots/schedule-P3/step-4-{bed-utilization,contract-execution}.png`
- [ ] `verify.sh` + `npm run build` 通过

---

## P3 完工核对

- [ ] 4 个 Step 全部合并
- [ ] `Schedule_PatientShift` 仅新增 1 个 NULL 字段，老数据零变化
- [ ] 最近调整可按 AdjustReason 区分
- [ ] CSV 批量导入跑通
- [ ] 两张报表上线

至此排班重构 P0 + P2 + P3 全部完工。

## 后续（不在本计划，仅记录）

- P4 历史反推合约：扫过去 N 周稳定排班自动建议合约
- 节假日特殊周覆盖：业务真有需求再做
- 模式按周几切换：等查测试库 OddWeek/EvenWeek 是否常出现患者频次切换模式后决定
- 自动定时渲染未来 N 周：当前手动触发，业务接受度高再加定时任务
