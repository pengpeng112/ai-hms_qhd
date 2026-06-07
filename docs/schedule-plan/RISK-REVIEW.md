# 排班重构计划风险审查与补充

> 状态：执行前必读  
> 来源：按当前仓库代码复核 `docs/schedule-plan/README.md`、`P0-bug-and-gap.md`、`P2-contract-render.md`、`P3-log-and-import.md`  
> 结论：原计划方向可用，但不能直接从 P0 Step 1 开工；必须先做 P0-0 事实校准与接口闭环。

## 1. 最高优先级风险

### R1 · 前端已调用的排班聚合接口，当前后端未注册

证据：
- 前端 `Schedule.tsx` 通过 `restApi.getScheduleWeek` 调 `GET /api/v1/schedule/week`。
- 前端拖拽换床调用 `POST /api/v1/patient-shifts/:id/move`。
- 前端互换调用 `POST /api/v1/patient-shifts/swap`。
- 当前后端 `RegisterScheduleRoutes` 只注册 `/shifts`、`/patient-shifts`、`/patients/:id/shift`。

风险：
- 若直接执行 P0 Step 1/2/3，执行 AI 会在不存在的后端周视图/移动/互换逻辑上继续堆代码。
- 排班页可能仍无法加载核心数据，即使角标、PUT 等局部修改完成也无法验收。

补充要求：
- 在 P0 前新增 `P0-0 · 当前接口闭环与事实校准`。
- P0-0 必须让现有 `Schedule.tsx` 的当前调用闭环：`/schedule/week`、`/patient-shifts/:id/move`、`/patient-shifts/swap`。
- 如果决定不用 `/move` / `/swap`，必须同步改前端 `useScheduleDragDrop.ts` 和 `Schedule.tsx`，不能保留悬空客户端方法。

### R2 · 原计划引用的后端行号和函数与当前仓库不符

证据：
- 原计划说 `patient_shift_service.go` 约 1100+ 行，并引用 488-544 行周视图填充逻辑。
- 当前文件约 331 行，没有 `listPendingSchedulePatients`、周视图 DTO、模板应用、move/swap 实现。

风险：
- 执行 AI 会按错误行号改不到正确位置，或者自行新建不一致实现。

补充要求：
- 任何 Step 不得依赖原计划中的行号。
- 先 `grep`/读取当前文件，再确定真实落点。
- P0-0 要输出当前排班接口事实表，作为后续 P0/P2/P3 的依据。

### R3 · 状态枚举口径冲突

证据：
- 原计划写老库 `Schedule_PatientShift.Status`：`10 草稿 / 20 已确认 / 30 用户确认 / 40 用户取消 / 50 排班取消 / 60 转出`。
- 当前代码 `legacy_enum_maps.go` 映射为新状态 `0/1/2/3/4 -> 10/20/30/40/50`，并把 `40` 当作“已完成”。

风险：
- 报表、取消、历史保护、合约执行率会把“用户取消”和“已完成”混淆。
- P3 报表若直接用 `Status IN (20,30)` 可能与前端显示和治疗执行状态不一致。

补充要求：
- P0-0 必须确认并记录最终状态口径。
- 状态口径未确认前，不允许开发 P3 报表和任何依赖“完成/取消”的统计。
- 如确认当前代码错误，先修 `legacy_enum_maps.go` 和相关测试，再继续后续 Step。

### R4 · 计划引用了不存在的文档路径

证据：
- 当前仓库没有 `排班管理.md`。
- 当前仓库没有 `docs/legacy-migration-pending-confirmation.md` 和 `docs/legacy-migration-confirmed-completed.md`。
- 当前实际待确认文件是 `docs/legacy-migration-uncertain-field-checklist.md`。

风险：
- 执行 AI 会把待确认事项写到不存在文件，导致确认事项丢失。

补充要求：
- 排班结构事实统一以根目录 `老血透数据库表结构-合并版.md` 为准。
- 待确认事项统一写入 `docs/legacy-migration-uncertain-field-checklist.md`。
- DDL 已人工确认/DBA 执行结果，写入对应 SQL PR 描述或新增阶段记录，不再引用不存在的 completed 文档。

## 2. P0 风险与补充

### R5 · 修改排班改 PUT 时，当前后端 Update 入参不够

证据：
- 当前前端修改已排班时，如果更换患者，会先 `deletePatientShift(existing.id)` 再 `createPatientShift(...)`。
- 当前后端 `PatientShiftUpdateRequest` 只支持 `shiftId`、`bedId`、`wardId`、`status`、`notes`。
- 当前 Create 请求结构也未接收前端传入的 `patientPlanId`、`shiftTiming`、`dialysisMode`。

风险：
- 仅把前端改为 `PUT /patient-shifts/{id}` 后，无法保留“更换患者”语义。
- `PatientPlanId`、`ShiftTiming` 可能继续被后端丢弃，导致长期/临时和合约角标无法可靠显示。

补充要求：
- P0 Step 2 前先补后端 request：是否允许 PUT 修改 `patientId`、`scheduleDate`、`patientPlanId`、`shiftTiming` 必须明确。
- 若允许修改患者，必须重新做患者同日同班冲突校验、床位同格冲突校验、历史日期保护。
- 若不允许修改患者，前端弹窗应禁止在“编辑已有排班”时更换患者，只允许改床位/班次/模式等安全字段。

### R6 · batch-save 不存在

证据：
- 当前后端未搜到 `batch-save` / `BatchSave`。

风险：
- P0 Step 5 若按“后端已存在”执行，会失败或临时拼接口。

补充要求：
- P0 Step 5 必须先改为“确认是否做批量保存”。
- 若做，需要新增后端接口设计、部分成功/全部失败口径、冲突返回格式。
- 若不做，删除该 Step，避免拖大 P0 范围。

### R7 · 角标 sourceType 取值与当前前端类型不一致

证据：
- 当前 `RestScheduleWeekShift.sourceType` 类型是 `'manual' | 'template' | 'import'`。
- 原计划希望派生 `'contract' | 'temporary' | 'manual'`。
- 当前卡片仍判断 `sourceType === 'template'` 显示“模板”。

风险：
- 如果后端改为 `contract/temporary/manual`，现有模板角标会消失。
- 如果保留 `template/import`，临时/合约角标又无法准确表达。

补充要求：
- P0 Step 1 要先定义统一枚举：建议 `manual | contract | temporary | template | import`。
- 前端角标显示规则按优先级：`temporary` > `isManualAdjusted` > `contract/template/import`。
- 类型定义、后端 DTO、卡片渲染必须同时改。

## 3. P2 风险与补充

### R8 · Plan_PatientPlan 三字段只能表达“单一落位规则”

风险：
- `PreferredBedId/PreferredShiftId/PreferredWeekdayMask` 只能表达一个方案对应一组床位/班次/周几。
- 无法表达“周一 HD、周三 HDF 不同床/不同班次”或单双周周几完全不同。

补充要求：
- P2 文档必须明确 MVP 限制：只支持“同一方案默认透析模式 + 一组首选落位规则”。
- 若遇到多模式/多规则患者，记录到 `docs/legacy-migration-uncertain-field-checklist.md`，不要强行塞入三个字段。

### R9 · 周渲染幂等条件不够

原计划：以 `(TenantId, PatientId, TreatmentTime, ShiftId, BedId)` 判断已存在。

风险：
- 同一患者同日同班换床后，合约渲染可能再生成一条新记录。
- 同一床位同日同班已有他人占用时，若患者维度不存在但床位维度存在，必须返回床位冲突。

补充要求：
- 幂等先按患者维度判断：`TenantId + PatientId + DATE(TreatmentTime) + ShiftId + Status not in canceled`。
- 冲突再按床位维度判断：`TenantId + BedId + DATE(TreatmentTime) + ShiftId + Status not in canceled`。
- 两类结果分别返回：`skippedByPatient`、`conflictedByBed`，不要混成一个 skipped。

### R10 · 合约落位校验需确认床位/班次启用字段

风险：
- 老库大小写敏感，床位/班次启用字段是 `IsDisabled`。
- 任何 SQL 漏双引号或误用新库字段会导致线上查询失败。

补充要求：
- 所有 SQL 使用显式双引号：`"Schedule_Bed"."IsDisabled" = false`。
- 不使用 `Register_Ward`、`Register_Bed`、`Schedule_Bed.DeviceId`。

## 4. P3 风险与补充

### R11 · “换床日志”实际不可追溯

风险：
- 只加 `AdjustReason` 无法知道原床位、原班次、原操作人链路。
- UI 如果叫“换床日志”，用户会期待完整历史。

补充要求：
- 文案改为“调整记录标记”或“最近调整”。
- 弹窗说明必须写明：仅展示当前记录最近一次调整原因，更早历史无法追溯。

### R12 · 报表统计依赖状态口径，必须后置

风险：
- P3 报表中的 `Status IN (20,30)` 在状态口径未确认前不可靠。

补充要求：
- P3 Step 4 必须依赖 P0-0 状态口径确认结果。
- 报表 SQL 中需明确排除取消/转出/草稿，具体状态值不得硬编码在多个文件，建议集中常量。

### R13 · CSV 导入字段“周几掩码”描述前后不一致

证据：
- P3 CSV 表头写“周几掩码”，示例值又写 `1,3,5`。

风险：
- 用户以为要填数字 mask，例如 21；后端却按逗号周几解析。

补充要求：
- 表头改为“周几（1=周一,逗号分隔）”。
- 后端同时兼容 `1,3,5` 与 `21` 需要明确；若不兼容，模板必须写清楚。

## 5. 必须加入计划的 P0-0 验收

- [ ] `GET /api/v1/schedule/week` 与当前前端 `RestScheduleWeekResponse` 对齐。
- [ ] `patientShifts` 每条包含 `id/patientId/patientName/wardId/bedId/bedName/shiftId/patientPlanId/dialysisMode/oddWeekFrequency/evenWeekFrequency/shiftTiming/status/statusName/treatmentTime/lastModifyTime`。
- [ ] `pendingPatients` 口径明确，至少能返回未排满患者列表。
- [ ] `POST /api/v1/patient-shifts/:id/move` 可用，或前端不再调用它。
- [ ] `POST /api/v1/patient-shifts/swap` 可用，或前端不再调用它。
- [ ] 创建/更新排班不会丢 `PatientPlanId` 与 `ShiftTiming`。
- [ ] 状态映射有单测覆盖，并与计划文档一致。
- [ ] `cd ai-hms-backend && go test ./internal/services ./internal/api/v1` 通过。
- [ ] `cd ai-hms-frontend && npm run lint && npm run build` 通过。

## 6. 执行顺序补充

修订后的顺序：

```text
P0-0 当前接口闭环与事实校准
  -> P0 Bug 与缺口修复
  -> P2 合约自动渲染
  -> P3 调整原因 / 导入 / 报表
```

任何执行 AI 如果发现本风险文档与阶段文档冲突，以本文件为准；若冲突影响数据库字段或接口契约，停止并向用户确认。
