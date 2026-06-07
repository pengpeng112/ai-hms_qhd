# P0 · Bug 与缺口修复（1-2 天 / P0-0 + 6 个 Step）

> 阶段目标：先完成当前排班接口闭环与事实校准，再修干净现有 Bug、暴露缺 UI 的接口、加临时病人角标。
> 前置：无（不依赖 P2/P3）。
> 关键特点：**不动数据库表结构**，仅前后端代码改动；DBA 不参与。
> 完工标准：P0-0 + 6 个 Step 全部通过验收，主分支 `verify.sh` + `npm run build` 全绿。

## 在开工前必须先读

1. `docs/schedule-plan/README.md`（一页须知 + 老库红线 + 决策摘要）
2. 本文件全部内容
3. `docs/schedule-plan/RISK-REVIEW.md`（当前代码复核后的风险补充，优先级高于本文件旧假设）
4. 根目录 `老血透数据库表结构-合并版.md`（老库表结构权威文档）
5. `ai-hms-backend/internal/api/v1/schedule_handler.go`（确认当前实际注册路由）
6. `ai-hms-backend/internal/services/patient_shift_service.go`（先 Read 全文；不要相信旧计划中的行号）
7. `ai-hms-backend/internal/models/schedule.go`
8. `ai-hms-frontend/src/pages/Schedule.tsx`
9. `ai-hms-frontend/src/hooks/useScheduleModals.ts`、`useScheduleDragDrop.ts`

## P0 不要做的事

- ❌ 不要写任何 `ALTER TABLE` / `CREATE TABLE`
- ❌ 不要新建后端 `models/` 中的表实体
- ❌ 不要改老库列名 / 改字典 typeCode
- ❌ 不要顺手做 P2 的合约渲染
- ❌ 不要把 P0-0 与后续 6 个 Step 合并到一个 PR

---

## P0-0 · 当前接口闭环与事实校准（必须先做）

### 背景

当前前端排班页已经调用以下接口：

- `GET /api/v1/schedule/week`
- `POST /api/v1/patient-shifts/:id/move`
- `POST /api/v1/patient-shifts/swap`

但当前后端 `RegisterScheduleRoutes` 仅注册了 `/shifts`、`/patient-shifts`、`/patients/:id/shift`。在这些接口闭环前，后续角标、频次差集、拖拽、批量保存都缺少稳定基础。

### 改这些文件

- `ai-hms-backend/internal/api/v1/schedule_handler.go`
- `ai-hms-backend/internal/services/patient_shift_service.go`
- `ai-hms-backend/internal/models/schedule.go`（如需要补字段 tag）
- `ai-hms-backend/internal/services/legacy_enum_maps.go`（如状态口径确认后需修正）
- `ai-hms-frontend/src/services/restClient.ts`（只在接口契约改变时同步类型）
- `ai-hms-frontend/src/pages/Schedule.tsx`、`ai-hms-frontend/src/hooks/useScheduleDragDrop.ts`（如决定移除 `/move` / `/swap` 调用）
- `docs/legacy-migration-uncertain-field-checklist.md`（记录仍无法确认的状态/字段口径）

### 必须完成的事实表

在 PR 描述或本计划补丁中列出：

| 项目 | 当前事实 | 最终决定 |
|---|---|---|
| `/schedule/week` | 是否后端已实现/注册 | 实现或改前端 |
| `/patient-shifts/:id/move` | 是否后端已实现/注册 | 实现或改前端 |
| `/patient-shifts/swap` | 是否后端已实现/注册 | 实现或改前端 |
| `Schedule_PatientShift.Status` | 老库真实枚举 | 同步 `legacy_enum_maps.go` |
| `PatientPlanId` | 创建/更新是否落库 | 必须不丢字段 |
| `ShiftTiming` | 创建/更新是否落库 | 长期=20，临时=10 |

### 接口契约最低要求

`GET /api/v1/schedule/week` 必须返回并匹配当前前端类型 `RestScheduleWeekResponse`：

```ts
{
  wards: RestScheduleWard[]
  beds: RestScheduleBed[]
  shifts: RestShift[]
  patientShifts: RestScheduleWeekShift[]
  pendingPatients: RestSchedulePendingPatient[]
}
```

`patientShifts` 每条至少包含：

```text
id, patientId, patientName, wardId, bedId, bedName, shiftId,
patientPlanId, dialysisMode, oddWeekFrequency, evenWeekFrequency,
shiftTiming, status, statusName, treatmentTime, lastModifyTime
```

`pendingPatients` 口径至少要说明：

- 数据来源：活跃患者 + 当前方案 `Plan_PatientPlan`
- 频次来源：`OddWeekFrequency` / `EvenWeekFrequency`
- 是否按“频次差集”计算剩余次数；如果暂未做，必须在后续 Step 3 明确补齐

### 字段落库要求

当前前端创建排班会传 `patientPlanId`、`shiftTiming`、`dialysisMode`。P0-0 必须确认：

- `PatientPlanId` 写入 `Schedule_PatientShift.PatientPlanId`
- `ShiftTiming` 写入 `Schedule_PatientShift.ShiftTiming`
- `dialysisMode` 如老表无直接承接列，不得凭空写入；需要展示时从患者方案 `Plan_PatientPlan.DialysisMethod` 或其他已确认来源回填

如字段无法确认，记录到 `docs/legacy-migration-uncertain-field-checklist.md`。

### 状态口径要求

必须用 `老血透数据库表结构-合并版.md` 和当前业务确认 `Schedule_PatientShift.Status` 语义。确认前：

- 不开发依赖完成/取消口径的报表
- 不新增硬编码状态集合
- 不把 `40` 同时当“用户取消”和“已完成”

### 验收

- [ ] 当前排班页能成功加载周视图。
- [ ] 新建排班后 `PatientPlanId` 和 `ShiftTiming` 不丢失。
- [ ] 拖拽换床可用，或前端不再调用悬空 `/move` 接口。
- [ ] 互换排班可用，或前端不再调用悬空 `/swap` 接口。
- [ ] 状态映射有测试或明确文档记录。
- [ ] `cd ai-hms-backend && go test ./internal/services ./internal/api/v1` 通过。
- [ ] `cd ai-hms-frontend && npm run lint && npm run build` 通过。

---

## Step 1 · 后端补 sourceType / isManualAdjusted 派生字段

### 背景

老库 `Schedule_PatientShift` 没有 `sourceType` / `isManualAdjusted` 列，但**前端卡片已在判断这两字段**（位于 `Schedule.tsx`），所以角标永远不显示。执行本 Step 前，必须已完成 P0-0，并确认周视图 DTO 的真实填充位置。

解决方案：**不加列，后端在 DTO 层派生**：

- `sourceType`：根据 `ShiftTiming` 派生；枚举需兼容当前前端既有 `template/import`
  - `ShiftTiming=10`（临时） → `sourceType='temporary'`
  - `ShiftTiming=20`（长期） → `sourceType='contract'`（与 `Plan_PatientPlan` 关联即视为合约位）
  - 其余 → `sourceType='manual'`
- `isManualAdjusted`：`LastModifyTime - CreateTime > 60s` 视为人工调整过

### 改这些文件

- `ai-hms-backend/internal/services/patient_shift_service.go`（修改 P0-0 确认后的周视图 DTO 结构体 + 填值逻辑）
- `ai-hms-backend/internal/services/patient_shift_service_test.go`（新增/补充派生逻辑单测）
- `ai-hms-frontend/src/services/restClient.ts`（前端类型 `RestScheduleWeekShift` 同步加字段）

### 怎么改

**A. 后端 DTO 加字段**

```go
type PatientShiftWeekItem struct {
    // ...保留原有字段...
    SourceType       string `json:"sourceType"`        // contract | temporary | manual
    IsManualAdjusted bool   `json:"isManualAdjusted"`  // 是否人工调整过
}
```

**B. 在查询函数填充逻辑**（不要相信旧行号；以 P0-0 确认的真实周视图填充循环为准）：

```go
sourceType := "manual"
switch r.ShiftTiming {
case 10:
    sourceType = "temporary"
case 20:
    if r.PatientPlanID > 0 {
        sourceType = "contract"
    }
}

isAdjusted := false
if !r.CreateTime.IsZero() && r.LastModifyTime.Sub(r.CreateTime) > time.Minute {
    isAdjusted = true
}

items = append(items, PatientShiftWeekItem{
    // ...原字段...
    SourceType:       sourceType,
    IsManualAdjusted: isAdjusted,
})
```

> ⚠️ 如当前查询 SELECT 列表未包含 `CreateTime`，需补 `ps."CreateTime"` 进 SELECT 与 row 结构体。先读 P0-0 确认后的真实 SELECT 列表，不要按旧行号修改。

**C. 前端类型同步**（不改逻辑）：

`RestScheduleWeekShift` 中如缺 `sourceType` / `isManualAdjusted`，加上；如已有 `template/import`，不要删，合并为完整枚举：

```ts
export interface RestScheduleWeekShift {
  // ...原字段...
  sourceType?: 'contract' | 'temporary' | 'manual' | 'template' | 'import'
  isManualAdjusted?: boolean
}
```

### 自检 grep

```bash
grep -n "SourceType\|IsManualAdjusted" ai-hms-backend/internal/services/patient_shift_service.go
grep -n "sourceType\|isManualAdjusted" ai-hms-frontend/src/services/restClient.ts
```

两条都应有命中。

### 验收

- [ ] 周视图卡片右上角"调"角标在被人工修改过的卡片上显示
- [ ] 临时排班（`ShiftTiming=10`）`sourceType` 为 `temporary`
- [ ] 长期合约位 `sourceType='contract'`
- [ ] 后端单测覆盖三种 sourceType 与 isAdjusted 的派生逻辑
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 2 · 修改排班改用 PUT /patient-shifts/{id}

### 背景

当前 Schedule.tsx 的"修改排班"流程是**先 DELETE 再 POST**，导致 `patient_shift_id` 漂移，与 `treatment` 表的关联断链，治疗记录查不到原排班来源。

后端 `PUT /api/v1/patient-shifts/{id}` 已经实现，但当前入参可能不支持更换患者、日期、`PatientPlanId`、`ShiftTiming`。本 Step 不能只改前端；必须先确认 P0-0 的字段落库结论。

### 改这些文件

- `ai-hms-frontend/src/pages/Schedule.tsx`（修改弹窗保存逻辑）
- `ai-hms-frontend/src/hooks/useScheduleModals.ts`（如保存 handler 在这里）
- `ai-hms-frontend/src/services/restClient.ts`（确认 `restApi.updatePatientShift` 存在；不存在则补一个 PUT 包装）

### 怎么改

**A. 先 grep 确认 PUT 接口的客户端方法是否已有**

```bash
grep -n "updatePatientShift\|PUT.*patient-shifts" ai-hms-frontend/src/services
grep -n "PUT.*patient-shifts\|PutPatientShift\|UpdatePatientShift" ai-hms-backend/internal/api/v1
```

如前端无包装，新增：

```ts
async updatePatientShift(id: number, body: UpdatePatientShiftRequest) {
  return await restClient.put(`/api/v1/patient-shifts/${id}`, body)
}
```

**B. 修改 Schedule.tsx 中"修改排班"保存路径**

旧逻辑伪代码（找出现 `restApi.deletePatientShift` 紧接 `restApi.createPatientShift` 的地方）：

```ts
// 旧（错）
await restApi.deletePatientShift(oldId)
await restApi.createPatientShift(newBody)

// 新：仅在后端 PUT 入参已支持完整语义时使用
await restApi.updatePatientShift(oldId, newBody)
```

如果后端决定不允许 PUT 更换患者，则前端编辑已有排班时必须禁止更换患者，只允许改床位/班次等安全字段。

**注意**：仅"修改排班"走 PUT；"取消排班"仍走原"软删 Status=50"接口。

### 自检 grep

```bash
grep -rn "deletePatientShift" ai-hms-frontend/src/pages/Schedule.tsx ai-hms-frontend/src/hooks
```

`Schedule.tsx` 修改弹窗代码路径中应**没有** `deletePatientShift` 紧跟 `createPatientShift` 的组合。

### 验收

- [ ] 修改排班后，`patient_shift_id` 不变
- [ ] 已存在的治疗记录关联仍可访问（手动验证：先建排班→在透析执行页生成治疗记录→改排班→检查治疗页关联是否仍在）
- [ ] 历史日期保护、班次冲突等校验仍生效（PUT 后端校验已有）
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 3 · 待排班队列改"频次差集"

### 背景

当前 `listPendingSchedulePatients`（`patient_shift_service.go:548`）的逻辑：

```
本周该患者有任意 1 条排班 → 从队列移除
```

频次 3 次/周的患者排了 1 次后，剩余 2 次排不到队列，导致漏排。

### 改这一个文件

`ai-hms-backend/internal/services/patient_shift_service.go`

### 怎么改

**A. 计算每个患者本周"应排次数"**

应排次数 = `OddWeekFrequency`（单周）或 `EvenWeekFrequency`（双周），按 ISO 周号奇偶决定：

```go
// 计算 ISO 周号奇偶
_, isoWeek := mondayDate.ISOWeek()
isOddWeek := isoWeek%2 == 1

func expectedTimes(plan struct{ Odd, Even int }, isOdd bool) int {
    if isOdd {
        return plan.Odd
    }
    return plan.Even
}
```

**B. 改差集逻辑**

```go
// 旧
scheduledIDs := map[int64]bool{}
for _, item := range scheduled {
    scheduledIDs[item.PatientID] = true
}

// 新：计每患者已排次数
scheduledCount := map[int64]int{}
for _, item := range scheduled {
    // 仅长期排班计入"应排"，临时不计
    if item.ShiftTiming == 20 {
        scheduledCount[item.PatientID]++
    }
}
```

**C. 队列项展示"剩余次数"**

```go
type SchedulePatientDTO struct {
    // ...原字段...
    ExpectedTimes  int `json:"expectedTimes"`   // 本周应排
    ScheduledTimes int `json:"scheduledTimes"`  // 已排
    RemainingTimes int `json:"remainingTimes"`  // 剩余 = expected - scheduled
}
```

筛除条件改为 `RemainingTimes <= 0` 才出队列；前端可据此显示"剩 2 次"标签。

### 自检 grep

```bash
grep -n "scheduledIDs\|RemainingTimes" ai-hms-backend/internal/services/patient_shift_service.go
```

新逻辑应可见 `RemainingTimes`，老的 `scheduledIDs[r.ID]` 已被替换。

### 验收

- [ ] 单元测试：频次 3 次/周患者，本周已排 1 次，待排队列仍显示该患者，`remainingTimes=2`
- [ ] 双周频次场景验证（单周 3 次 / 双周 2 次）
- [ ] 前端右侧待排队列文字增加"剩 N 次"小标签（小改动，可放本 Step）
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 4 · 班次管理 UI（CRUD）

### 背景

后端 `/api/v1/shifts` CRUD 已存在，前端无入口。本 Step 给"系统配置"添加班次管理页。

### 改这些文件

- `ai-hms-frontend/src/pages/ShiftConfig.tsx`（**新增**，< 350 行）
- `ai-hms-frontend/src/router.tsx`（注册路由 `/shift-config`）
- `ai-hms-frontend/src/layouts/Sidebar.tsx`（系统配置组下加菜单项）
- `ai-hms-frontend/src/layouts/routeMeta.ts`（注册面包屑）
- `ai-hms-frontend/src/i18n/locales/zh/nav.json`（新增 `nav.shiftConfig` key）
- `ai-hms-frontend/src/services/restClient.ts`（确认 `getShiftList` / `createShift` / `updateShift` / `deleteShift` 客户端方法存在；不存在则补）

### 怎么改

**A. 页面布局**：参考 `BedManagement.tsx` / `WardManagement.tsx` 的现有样式

```tsx
// 顶部：[新建班次] 按钮 + 搜索
// 表格列：班次名称 / 开始时间 / 结束时间 / 排序 / 启用状态 / 操作（编辑/启用-禁用）
// 编辑用 antd Modal，字段对应 Schedule_Shift 表
```

**B. 字段映射**（对应根目录 `老血透数据库表结构-合并版.md` 中 `Schedule_Shift`）：

| 前端字段 | 老库列 | 类型 |
|---|---|---|
| name | Name | string(256) |
| sort | Sort | int |
| startTime | StartTime | DateTime |
| endTime | EndTime | DateTime |
| isDisabled | IsDisabled | bool |
| note | Note | string(512) |

**C. 表单校验**

- StartTime < EndTime
- Name 必填
- 启用/禁用走 PATCH 或 PUT（看后端实现），禁用前提示"该班次已有 N 条历史排班，禁用不删除历史"

### 验收

- [ ] 4 角色（admin / scheduler / nurse_head / doctor）权限正确：仅 admin / scheduler 可见
- [ ] 班次新建、编辑、启用、禁用 4 项交互通过
- [ ] 禁用后排班页选择班次时该项不出现
- [ ] 截图：`docs/ui-snapshots/schedule-P0/step-4-shift-config-{list,edit,disabled}.png`
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 5 · 批量保存按钮接 batch-save

### 背景

原计划假设后端 `POST /api/v1/patient-shifts/batch-save` 已存在；当前复核未发现该接口。此 Step 先做取舍：确认要新增后端批量接口，还是暂时删除批量保存范围。

### 改这些文件

- 如决定新增接口：`ai-hms-backend/internal/api/v1/schedule_handler.go`、`ai-hms-backend/internal/services/patient_shift_service.go`
- 如决定前端接入：`ai-hms-frontend/src/pages/Schedule.tsx`
- `ai-hms-frontend/src/services/restClient.ts`（确认或新增 `batchSavePatientShifts` 客户端方法）

### 怎么改

**A. 顶部新增"批量保存"按钮**

放在"应用模板"按钮旁，仅当用户存在"草稿态变更"（如刚拖拽未保存）时高亮。

**B. 触发条件**

- 拖拽暂存模式：拖完不立即调 `/move`，先存到本地 state，按"批量保存"统一提交
- 沿用现有 `useScheduleDragDrop` 的状态机

> ⚠️ 如果当前拖拽是即时调用 `/move` 的，本 Step 改造为"批量提交"会改动多处。**先 Read `useScheduleDragDrop.ts` 全文**，评估是否能在不影响即时拖拽的前提下，仅给"新建/修改"操作加批量入口。如果改动 > 200 行，记到 `docs/legacy-migration-uncertain-field-checklist.md` 拆 Step 5a / 5b。

**C. 接口契约**：若后端不存在，先设计入参和部分失败返回格式，不允许前端先接空方法。

### 验收

- [ ] 多格新建排班（如同时给 5 张床位排班）→ 一次提交成功
- [ ] 批量提交失败时，前端逐项标红冲突格子（不要"all-or-nothing"，看后端是否支持部分成功；不支持则提示用户解决冲突重试）
- [ ] `verify.sh` + `npm run build` 通过

---

## Step 6 · 临时病人角标

### 背景

用户答复"临时病人要有标注"。老库 `Schedule_PatientShift.ShiftTiming=10` 即临时，已在 Step 1 派生为 `sourceType='temporary'`。

### 改这一个文件

`ai-hms-frontend/src/pages/Schedule.tsx`

### 怎么改

**A. 卡片渲染**

```tsx
{item.sourceType === 'temporary' && (
  <span className="absolute -top-1 -left-1 rounded bg-orange-500 px-1.5 py-0.5 text-density-strict font-bold text-white">
    临
  </span>
)}
```

**B. 待排班队列**

如该患者本次为临时病人（来自外院/插单），队列项前加"临"小标签。

> 临时病人来源识别：当前老库无独立"临时病人"字段。本 Step 仅根据 `ShiftTiming=10` 显示。如业务还想标"非本中心患者"作为患者级属性，记到 `docs/legacy-migration-uncertain-field-checklist.md`，等用户确认后再做。

**C. 临时排班创建入口**

新建排班弹窗增加"长期 / 临时"选择（默认长期=ShiftTiming=20）。临时则后端 ShiftTiming=10。

### 验收

- [ ] 临时排班卡片左上角显示橙色"临"角标
- [ ] 新建弹窗能选"临时"，保存后 ShiftTiming=10
- [ ] 长期排班无角标
- [ ] 截图：`docs/ui-snapshots/schedule-P0/step-6-temporary-{card,modal}.png`
- [ ] `verify.sh` + `npm run build` 通过

---

## P0 完工核对

- [ ] 6 个 Step 各 1 个 PR 合并到 `feat/schedule-uplift` 分支
- [ ] 前端周视图：合约角标、调整角标、临时角标 三类可见
- [ ] 修改排班 ID 不漂移
- [ ] 频次 3 次/周患者排 1 次后队列仍可见，"剩 2 次"标签显示
- [ ] 班次管理页可 CRUD
- [ ] 批量保存按钮可用
- [ ] 数据库未做任何 DDL，老库现存数据零变化

通过后开 `docs/schedule-plan/P2-contract-render.md`。
