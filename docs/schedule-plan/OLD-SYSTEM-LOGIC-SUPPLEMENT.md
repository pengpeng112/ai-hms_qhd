# 排班计划补充 · 老系统逻辑对齐

> 状态：执行前必读  
> 依据：根目录 `排班管理.md`  
> 目的：把老系统已明确的排班业务规则补入当前重构计划，减少执行 AI 主观猜测。

## 1. 老系统已明确的业务事实

### 1.1 核心表与字段

| 表 | 字段 | 老系统含义 | 当前计划落点 |
|---|---|---|---|
| `Schedule_PatientShift` | `TreatmentTime` | 治疗日期 | 周视图、日视图、历史保护、周渲染 |
| `Schedule_PatientShift` | `ShiftId` | 班次 ID | 班次列、患者/床位冲突判断 |
| `Schedule_PatientShift` | `WardId` | 病区 ID | 病区分组、感染匹配 |
| `Schedule_PatientShift` | `BedId` | 床位 ID | 床位行、同床同班唯一性 |
| `Schedule_PatientShift` | `PatientPlanId` | 患者方案 ID | 必须写入，后续透析模式/频次来源 |
| `Schedule_PatientShift` | `ShiftTiming` | 10 临时 / 20 长期 | 临时角标、待排班频次差集、合约渲染 |
| `Schedule_PatientShift` | `Status` | 10 草稿 / 20 已确认 / 30 用户确认 / 40 用户取消 / 50 排班取消 / 60 转出人员 | 必须修正当前状态映射 |
| `Schedule_Ward` | `PatientType` | 适用患者类型：长期10 / 临时20 | 病区匹配规则，不能只做展示 |
| `Schedule_Ward` | `InfectionType` | 适用传染病：普通 / 乙肝 / 丙肝 | 感染匹配规则 |
| `Schedule_Bed` | `IsDisabled` | 床位启用/禁用 | 排班时只展示/选择启用床位 |
| `Schedule_Shift` | `Sort` | 班次排序 | 当天已过班次保护 |
| `Schedule_Shift` | `StartTime` / `EndTime` | 班次开始/结束时间 | 班次管理、当前班次判断 |
| `Plan_PatientPlan` | `DialysisMethod` | 患者透析方式 | 排班卡片模式显示、设备匹配 |
| `Plan_PatientPlan` | `OddWeekFrequency` / `EvenWeekFrequency` | 单/双周频次 | 待排班队列、合约执行率 |

## 2. 必须补入 P0 的规则

### P0-0 追加：状态映射必须按老系统文档修正

当前计划已要求 P0-0 做状态口径确认；根据 `排班管理.md`，应优先按下表修正：

| 老库值 | 名称 | 当前前端建议显示 |
|---|---|---|
| 10 | 草稿 | 草稿 |
| 20 | 已确认 | 已确认 |
| 30 | 用户确认 | 用户确认 |
| 40 | 用户取消 | 用户取消 |
| 50 | 排班取消 | 排班取消 |
| 60 | 转出人员 | 转出 |

执行要求：
- 后端 `legacy_enum_maps.go` 不应再把 `40` 当“已完成”。
- 取消排班接口应写 `Status=50`。
- 用户取消若未来有入口，应写 `Status=40`。
- `Status=60` 用于转出人员或软删除模板排班，不应混入普通取消。

### P0-0 追加：创建/更新必须保留老系统保存结构

老系统 `subSave` 保存结构包含：

```text
Id, WardId, BedId, PatientId, PatientPlanId, ShiftId, ShiftTiming, Status, TreatmentTime, LastModifyTime
```

执行要求：
- `POST /api/v1/patient-shifts` 请求体必须接收并落库 `patientPlanId`、`shiftTiming`。
- `PUT /api/v1/patient-shifts/:id` 如用于修改排班，也必须支持 `patientPlanId`、`shiftTiming`、`treatmentTime`，是否允许改 `patientId` 见不确定项。
- 如前端传 `dialysisMode`，后端不能写入不存在字段；展示时从 `Plan_PatientPlan.DialysisMethod` 回填。

### P0-1 追加：sourceType 派生应以老系统含义为准

建议派生规则：

| 条件 | `sourceType` |
|---|---|
| `ShiftTiming=10` | `temporary` |
| `ShiftTiming=20` 且 `PatientPlanId > 0` | `contract` |
| 其他 | `manual` |

若未来保留模板/导入来源，扩展为：`manual | contract | temporary | template | import`，但不得覆盖临时标识。

### P0-3 追加：待排班队列必须用频次差集

老系统文档明确排班要显示“已排/应排/差异提示”。因此待排班队列不能用“本周排过一次就移除”。

执行口径：
- `expectedTimes`：根据 ISO 周奇偶取 `OddWeekFrequency` 或 `EvenWeekFrequency`。
- `scheduledTimes`：统计本周 `ShiftTiming=20` 且 `Status NOT IN (40,50,60)` 的排班数。
- `remainingTimes = expectedTimes - scheduledTimes`。
- `remainingTimes > 0` 时进入待排班队列。
- 临时排班 `ShiftTiming=10` 不抵扣长期应排次数。

### P0 追加：四类校验必须先做最小实现

`排班管理.md` 第五章给出校验规则，当前 P0 需要最小闭环：

| 规则 | P0 最小实现 | 后续增强 |
|---|---|---|
| 历史数据保护 | `TreatmentTime < 今天` 禁止修改/取消/拖拽 | 管理员例外待确认 |
| 已过班次保护 | 当天 `Schedule_Shift.Sort < 当前班次 Sort` 禁止修改 | 需要按当前时间换算当前班次 |
| 治疗中保护 | 删除/取消前检查治疗记录是否已开始 | 需确认具体治疗动作表 |
| 唯一性规则 | 同一床位同一日期同一班次只能一人；同一患者同一日期同一班次只能一条 | 批量保存时逐项返回冲突 |

## 3. P2 合约自动渲染需要按老系统模板逻辑补强

老系统支持模板管理与复制模板：
- 多个模板，如正式模板、草稿模板。
- 模板启用，可配置默认模板。
- 支持 2 周 / 4 周循环模板。
- 查看未来周时，若没有排班计划，可提示从模板复制。

当前 P2 的“Plan_PatientPlan 加首选字段 + 手动生成本周排班”是 MVP，不等价于老系统完整模板。

### P2-MVP 保留

继续保留：
- `Plan_PatientPlan.PreferredBedId`
- `Plan_PatientPlan.PreferredShiftId`
- `Plan_PatientPlan.PreferredWeekdayMask`
- 手动 `render-week`

适用场景：简单长期患者，一组床位/班次/周几规则。

### 新增 P2B：模板复制兼容阶段

如果要对齐老系统，需要在 P2 后新增 P2B，而不是把 P2 做爆：

| Step | 内容 | 是否需要 DDL |
|---|---|---|
| P2B-1 | 调研老库是否已有模板表/模板字段，确认 `SchedulePatientShiftTPL` 是否真实存在 | 否 |
| P2B-2 | 若模板表存在，补模板查询/编辑/启用接口 | 否或少量字段 |
| P2B-3 | 周视图切换未来周时，如果本周无排班，提示“是否从模板复制” | 否 |
| P2B-4 | 复制模板到实际 `Schedule_PatientShift`，新记录 `Id=0` 语义改为后端生成 ID | 否 |
| P2B-5 | 模板周数支持 2/4 周循环，按 `weekNumber % templateWeekCount` 取模板周 | 待确认 |

注意：当前仓库曾记录 `SchedulePatientShiftTPL` 只是“设计出现但结构未收录”，不能假设生产表存在。P2B-1 必须先查库或让你确认。

## 4. P3 需要调整的口径

### 调整原因不是完整换床日志

老系统调整流程是拖放后保存数据，但 `排班管理.md` 没有给出换床日志表。当前计划只加 `AdjustReason` 是“最近调整分类”，不能追溯完整历史。

执行要求：
- UI 文案用“最近调整”或“调整标记”，不要写“完整换床日志”。
- 如果必须展示原床位 -> 新床位，需要新增日志表或找到老库现有日志表；当前计划不支持。

### 报表要按老系统状态口径计算

有效排班建议口径：
- 草稿：`10` 是否计入由用户确认，默认不计入完成率，可计入草稿工作量。
- 已确认/用户确认：`20/30` 可计入有效排班。
- 用户取消/排班取消/转出：`40/50/60` 不计入有效排班。

## 5. 已从 docs 文档中找到答案的确认项

以下内容通过分析仓库现有文档（`docs/legacy-db-schema.md`、`docs/migration-field-map.md`、`docs/treatment-execution-legacy-dev-record-2026-04-21.md`、`docs/legacy-migration-session-summary-2026-04-21.md` 等）已找到具体依据，可直接作为开发基线：

### S-1 治疗中保护 — 已从数据库确认

| 已确认 | 来源 |
|---|---|
| 治疗动作存储在 `Treatment_Action` 表（`Code` + `Name` + `TreatmentId` + `OperateTime`） | `docs/legacy-db-schema.md` 1141-1161 |
| 治疗主表 `Treatment_Treatment` 有 `ScheduleId` 字段关联排班，`Status` 为 varchar 类型 | 直接查库确认 |
| 老系统 `Treatment_actions[20]` 指 `Treatment_Action.Code=20` = "透前评估" | 直接查库确认 |

**数据库实际 Code 值**：
| Code | Name |
|---|---|
| 10 | 签到 / 已签到 |
| **20** | **透前评估** |
| 30 | 透中监测 |
| 40 | 透后评估 |
| 50 | 取消治疗 |
| 60 | 结束 |
| 70 | 首次核对 |
| 80 | 设备消毒 |
| 90 | 耗材登记/核对 |
| 100 | 健康宣教 |
| 110 | 保存并确认/保存草稿/调整并确认 |
| 120 | 透析医嘱系列 |
| 140 | 获取设备开始治疗时间 |
| 150 | 二次核对 |
| 170 | 治疗小结/医生小结 |

**P0-7 实现规则**：删除/取消排班前，通过 `ScheduleId` 查找 `Treatment_Treatment`，若存在记录且该治疗的 `Treatment_Action` 中存在 `Code >= 20`（即已完成透前评估），则禁止删除。

**可直接编码**。

### S-2 已过班次保护 — 已从数据库确认

**数据库实际班次数据**：
| Name | Sort | StartTime | EndTime |
|---|---|---|---|
| 上午 | 10 | 08:00 | 13:30 |
| 下午 | 20 | 13:30 | 18:30 |
| 夜班 | 30 | 18:30 | 23:00 |

**P0-7 实现规则**：当前时间落在哪个班次用 `StartTime::time <= NOW()::time < EndTime::time` 判断，取其 `Sort` 值，禁止修改 `Sort < 当前班次 Sort` 的当天排班行。

**可直接编码**。

### S-3 设备字段 — 已从数据库确认枚举值

**数据库实际设备数据**：
| Type (类型) | Flux (通量) | 支持治疗方式 |
|---|---|---|
| 血滤机 | 高 | HD, HDF, HF |
| 血透机 | 高 | HD, HFD |
| 血透机 | 低 | HD |
| 血滤机 | (空) | HD, HDF, HF（按血滤机默认） |

字段均为中文文本。`DialysisMethod` 在所有记录中均为空。

**P0-7 实现规则**：
- `Type='血滤机'`：允许排班任意模式，不拦截
- `Type='血透机' AND Flux='高'`：拒绝 HDF、HF 模式
- `Type='血透机' AND Flux='低'`：只允许 HD 模式

**可直接编码**。

### S-4 感染匹配 — 已从数据库确认字段格式

**数据库实际感染数据**：
```
InfectionDesc = "无"    OtherDesc=""   Note=""
InfectionDesc = "阴性"  OtherDesc=""   Note=""
```

**数据库实际病区数据**：
```
InfectionType = (空/NULL)
PatientType = "10" 或 (空/NULL)
```

测试数据库中感染和病区数据极少。`InfectionDesc` 是短文本（"无"/"阴性"），无结构化标记。

**P0-7 实现规则**：仅当 `InfectionDesc` 包含明确感染关键词（如"乙肝"/"丙肝"）时做匹配告警，否则不拦截。

**可直接编码（警告模式）**。

### S-5 普通患者排传染病区 — 已从数据库确认

测试数据库中 `Schedule_Ward.InfectionType` 全部为空或 NULL，说明当前环境未配置感染分区。但规则已从 `排班管理.md` 确认："默认不能，可配置"。

**P0-7 实现规则**：默认拒绝。`InfectionType` 为 NULL 或 "普通" 时视为普通病区，不限制；非空且非"普通"时，检验患者感染是否匹配。

**可直接编码**。

### S-7 模板表 — 已从数据库确认不存在

**数据库查询结果**：排班相关含 TPL/template 的表仅有 `Plan_PlanTPL`（方案模板）和 `Order_OrderTPL`（医嘱模板）。**没有 `SchedulePatientShiftTPL` 或任何排班模板表**。

老系统排班模板功能可能是通过 `Plan_PlanTPL` + 固定周循环逻辑实现的，并非独立排班模板表。

**可直接编码**：P2B 暂不做。P2 MVP 合约落位用 `Plan_PatientPlan` 加首选字段。

### S-10 修改排班换患者 — 已确认方案

**可直接编码**：编辑已有排班时选择患者的 select 下拉控件的 onChange 禁用，不允许换人，只允许改床位、班次、ShiftTiming、PatientPlanId。

### S-11 草稿进入治疗 — 已从数据库确认

**数据库排班状态分布**：
| Status | 含义 | 数量 |
|---|---|---|
| 10 | 草稿 | 1317 |
| 20 | 已确认 | 24081 |
| 50 | 排班取消 | 1508 |
| 60 | 转出人员 | 187 |

Status=30（用户确认）和 Status=40（用户取消）在数据中不存在，说明实际业务未启用这两个状态。

**可直接编码**：透析执行创建治疗记录时，只允许 Status=20（已确认）的排班。

---

## 6. 仍未确定需要你回复的功能清单

以下内容在数据库和文档中均未找到答案：

| 编号 | 不确定项 | 阻塞原因 | 当前阻塞的开发任务 |
|---|---|---|---|
| S-6 | 多传染病优先级顺序 | 病区 `InfectionType` 可能有多个值如"乙肝"/"丙肝"，患者如果同时有乙肝和丙肝排到哪个区？ | P0-7 感染匹配 |
| S-8 | 模板周数配置 `patientShiftWeekTPL` | 数据库中没有排班模板表，老系统源码变量 `patientShiftWeekTPL` 的值来源未知 | P2B 模板复制 |
| S-9 | "重新制定模板"按钮真实行为 | 数据库无模板数据可验证，老系统代码不可用 | P2B 模板管理 UI |
| S-12 | 批量保存是否需要 | 取决于护士实际操作习惯 | P0-5 批量保存取舍 |

**已从数据库确认、可直接编码的 8 项**：S-1、S-2、S-3、S-4、S-5、S-7、S-10、S-11。

**仍需你回复的 4 项**：S-6、S-8、S-9、S-12。

## 7. 修订后的开发顺序

```text
P0-0 接口闭环与状态/字段校准
P0-1 周视图角标与 DTO 派生
P0-2 修改排班改 PUT 或限制编辑语义
P0-3 待排班队列频次差集
P0-4 班次 CRUD 页面
P0-5 批量保存取舍与最小实现
P0-6 临时排班入口与角标
P0-7 老系统四类校验最小闭环（历史、已过班次、治疗中、唯一性）
P2 简单合约落位 + 手动周渲染 MVP
P2B 老系统模板复制兼容（需先确认模板表/配置）
P3 最近调整标记 + 导入 + 报表
```

执行 AI 必须按以上顺序推进。若与旧阶段文档冲突，以本补充为准。
