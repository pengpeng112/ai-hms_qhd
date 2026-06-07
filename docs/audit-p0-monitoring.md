# 监控页面字段级比对审计报告

> 审计时间：2026-05-31
> 审计范围：`/monitoring` 页面及子组件、hooks、后端 service/handler
> 核查依据：`老血透数据库表结构-合并版.md`、`LEGACY_TABLE_FIELD_MAPPING.md`

---

## 一、架构概要

| 项目 | 说明 |
|------|------|
| 页面入口 | `ai-hms-frontend/src/pages/Monitoring.tsx` |
| 数据 hooks | `useMonitoringData.ts` — 一次性 REST 加载，**无轮询、无 WebSocket** |
| 设备列表 API | `GET /api/v1/devices` → `device_handler.go:List` → `device_service.go:List` |
| 患者列表 API | `GET /api/v1/patients` → 用于构建 `patientName/bedNumber` 关联 |
| 医嘱 API | `GET /api/v1/patients/:id/orders` → `restApi.getPatientOrders` |
| 数据实时性 | **非实时**：仅页面加载时拉取一次，之后不再刷新 |
| 设备状态存储 | `Schedule_BedEquipmentRel.ParameterS`（jsonb 字段，老库标注"待人工补注"） |

---

## 二、核查明细表

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|-----------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| M-01 | P0 | StatusGrid | StatusGrid.tsx:106-206 | 设备网格渲染 | `device.id`, `device.bedNumber`, `device.patientName`, `device.status`, `device.mode`, `device.timeRemaining` | `GET /api/v1/devices` | device_service.go:396-431 | `e.Id`, `bed.Name`, `COALESCE(NULLIF(rel.ParameterS,''),'normal')` | `Auxiliary_EquipmentInfomation.Id`, `Schedule_Bed.Name`, `Schedule_BedEquipmentRel.ParameterS` | **部分不一致**：`ParameterS` 类型为 jsonb，老库标注"待人工补注"，当前用作状态字符串存储属于兼容实现 | ⚠️ 需确认 | LEGACY_TABLE_FIELD_MAPPING.md:141, device_service.go:419 | 确认 `ParameterS` jsonb 字段原始用途；如需正式状态字段应扩展 `Schedule_BedEquipmentRel` 或新增列 | ✅ |
| M-02 | P0 | StatusGrid | StatusGrid.tsx:150-163 | 干体重/增重%/通路展示 | `dryWeight`, `gainPct`, `vascularAccess` — **全部硬编码显示 `"--"`** | 无 | 无 | 无 | `Register_PatientInfomation.PredictWeight`(干体重), `Register_VascularAccess.AccessType`(通路) | **完全缺失**：前端占位字段未接入任何数据源 | 🔴 需改造 | StatusGrid.tsx:151-162 干体重/增重/通路三列全部 `"--"` | 需从患者详情或治疗方案 API 获取 `dryWeight`、`vascularAccess` 数据并映射到 MonitorDevice | ✅ |
| M-03 | P0 | StatusGrid | StatusGrid.tsx:165-185 | 血压/心率趋势图 | `device.vitals.sbp`, `device.vitals.dbp`, `device.vitals.hr`, `cachedGraphData` | 无专用 API | 无 | 无 | `Treatment_DuringSigns.SBP/DBP/HeartRate`, `Treatment_DuringParam.BF/TMP/UFQuantity/Conductivity/MachineTmp` | **完全缺失**：`toMonitorDevice` 硬编码 vitals 全为 0，`cachedGraphData` 始终为空数组 | 🔴 需改造 | types.ts:89-100, types.ts:26-27 | 需新增实时监测数据 API（从 `Treatment_DuringParam` + `Treatment_DuringSigns` 读取最新记录），或对接设备数据采集接口 | ✅ |
| M-04 | P0 | StatusGrid | StatusGrid.tsx:187-198 | 超滤进度条 | `device.vitals.ufGoal`, `device.vitals.ufVolume` | 无专用 API | 无 | 无 | `Treatment_Treatment.RealUFQuantity`(实际超滤), `Plan_PatientPlan`/处方(目标超滤) | **完全缺失**：`ufGoal` 和 `ufVolume` 始终为 0，进度条永远显示 0% | 🔴 需改造 | types.ts:96-97 | 需从治疗记录 + 处方获取超滤目标和实际超滤量 | ✅ |
| M-05 | P0 | StatusGrid | StatusGrid.tsx:132 | 透析模式 | `device.mode` | 间接：患者列表 → `defaultMode` | restClient.ts:77 | `RestPatient.defaultMode` | `Plan_PatientPlan.DialysisMethod` | **间接映射**：通过患者列表 `defaultMode` 字段关联，但患者列表 API 不一定返回该字段 | ⚠️ 需确认 | useMonitoringData.ts:29-30, types.ts:66 | 确认患者列表 API 是否返回 `defaultMode`；若不返回需从 `Plan_PatientPlan` 补充 | ✅ |
| M-06 | P0 | StatusGrid | StatusGrid.tsx:134 | 剩余时间 | `device.timeRemaining` | 无 | 无 | 无 | `Treatment_Treatment.StartTime`/`EndTime`/`RealDuration` | **完全缺失**：硬编码 `"--"` | 🔴 需改造 | types.ts:89 | 需从当前进行中的治疗记录计算剩余时间（`StartTime + 预设时长 - now`） | ✅ |
| M-07 | P0 | AlertList | AlertList.tsx:11-60 | 报警列表 | `device.status === 'alarm' \|\| 'warning'`, `device.mode`, `device.timeRemaining`, `device.bedNumber`, `device.patientName` | 同设备列表 | 同上 | 同上 | 同上 | **依赖 M-01**：状态字段来源 `ParameterS`（兼容实现）；`mode`/`timeRemaining` 同 M-05/M-06 | ⚠️ 需确认 | AlertList.tsx:14 | 同 M-01/M-05/M-06 | ✅ |
| M-08 | P1 | Monitoring.tsx 弹窗 | Monitoring.tsx:116-261 | 综合监测弹窗（血压/压力/超滤趋势图） | `HistoryPoint: sbp, dbp, hr, ap, vp, tmp, bf, uf` | 无专用 API | 无 | 无 | `Treatment_DuringParam`(BF/TMP/UFQuantity/VenousPressure/ArterialPressure), `Treatment_DuringSigns`(SBP/DBP/HeartRate) | **完全缺失**：`cachedHistoryData` 始终为空数组，图表无数据 | 🔴 需改造 | types.ts:7-17, types.ts:27 | 需新增透中参数趋势 API，按治疗记录查询 `DuringParam` + `DuringSigns` 并返回时间序列 | ✅ |
| M-09 | P1 | Monitoring.tsx 弹窗 | Monitoring.tsx:264-532 | 处方调整弹窗 | 多个 `PrescriptionInput`：透析时间、标准血流量、肝素类型/剂量、干体重、超滤量、通路、透析液 | 无（全部硬编码/TODO） | 无 | 无 | `Plan_PatientPlan`(DialysisDuration/BF/DryWeight/DialysisMethod), `Plan_PatientPrescription`(肝素方案) | **完全缺失**：所有字段硬编码，保存按钮标注"未开放" | 🔴 需改造 | Monitoring.tsx:60-61 | 需对接处方管理 API，从 `Plan_PatientPlan` + `Plan_PatientPrescription` 读取 | ✅ |
| M-10 | P1 | OrderListModal | Monitoring.tsx:544-713 | 医嘱列表 | `order.content`, `order.frequency`, `order.doctorName`, `order.startTime`, `order.status` | `GET /api/v1/patients/:id/orders` | restClient.ts:1389-1404 | `Order_PatientOrder.Content/Classification/Type/StartTime/OperatorId` | `Order_PatientOrder` 表 | **已对接**：后端读取 `Order_PatientOrder` 表，字段映射基本正确 | ✅ 基本通过 | treatment_service.go 中未直接涉及，但 `patient_core_service.go` 有 `buildActiveOrders` | 确认 `frequency` 字段在 `Order_PatientOrder` 中的来源 | ✅ |
| M-11 | P1 | SummaryModal | Monitoring.tsx:716-805 | 填写小结弹窗 | `summary`(textarea), `device.vitals.sbp/dbp`, `device.vitals.ufVolume` | `PUT /api/v1/treatments/:id/summary` | treatment_service.go:1439+ | `Treatment_Treatment.NurseSummary/TreatmentSummary` | `Treatment_Treatment.NurseSummary`, `Treatment_Treatment.TreatmentSummary` | **部分对接**：小结保存 API 存在，但展示的血压/超滤数据来源为 vitals（始终为 0） | ⚠️ 需改造 | Monitoring.tsx:774-781 | 血压/超滤数据需从治疗记录实时获取 | ✅ |
| M-12 | P1 | StatusGrid | StatusGrid.tsx:139-146 | 操作按钮（医嘱/小结） | `onOpenModal(device, 'ORDERS')`, `onOpenModal(device, 'SUMMARY')` | 无直接 API | 无 | 无 | 无 | **纯前端交互**：打开弹窗，无数据字段问题 | ✅ 通过 | — | — | ❌ |
| M-13 | P0 | useMonitoringData | useMonitoringData.ts:16-18 | 数据加载策略 | `restApi.getDeviceList({pageSize:200})`, `restApi.getPatientList({page:1,pageSize:500})` | `GET /api/v1/devices`, `GET /api/v1/patients` | device_service.go:434-481 | `Auxiliary_EquipmentInfomation` + 关联表, `Register_PatientInfomation` | 对应老库表 | **已对接**：设备列表读取老库多表关联；患者列表读取 `Register_PatientInfomation` | ✅ 通过 | useMonitoringData.ts:16-18 | 无 | ❌ |
| M-14 | P0 | useMonitoringData | useMonitoringData.ts:45 | 无实时刷新 | `useEffect` 仅挂载时执行一次，无 `setInterval`/WebSocket | 无 | 无 | 无 | 无 | **架构问题**：监控页面应有实时数据推送能力（轮询或 WebSocket） | 🔴 需改造 | useMonitoringData.ts:45 | 需添加轮询（如 30s 间隔）或 WebSocket 接入设备实时数据 | ✅ |
| M-15 | P0 | device_service | device_service.go:396-431 | 设备列表 SQL | JOIN `Schedule_BedEquipmentRel` ON `IsDefault=true` + `ParameterS` 作为状态 | — | — | `Schedule_BedEquipmentRel.ParameterS`(jsonb) | 老库定义：jsonb 类型，标注"待人工补注" | **兼容实现**：`ParameterS` 原始用途不明，当前用作设备状态字符串存储 | ⚠️ 需确认 | device_service.go:419, 老库表:874 | 确认 `ParameterS` jsonb 原始设计意图；长期方案应新增专用状态列 | ✅ |
| M-16 | P0 | device_service | device_service.go:390-402 | 设备表关联 | LEFT JOIN `Schedule_Bed`, `Schedule_Ward`, `Auxiliary_EquipmentMaintenance` | — | — | 多表关联 | 老库表结构定义 | **正确**：关联条件符合老库外键关系 | ✅ 通过 | device_service.go:396-401 | 无 | ❌ |
| M-17 | P1 | types.ts | types.ts:73-103 | `toMonitorDevice` 映射 | 将 `RestDevice` 映射为 `MonitorDevice`，vitals 全部硬编码为 0 | — | — | — | — | **数据丢失**：设备原始字段（serialNo/brand/model/manufacturer）在映射时被丢弃 | ⚠️ 需改造 | types.ts:89-102 | 若监控页需展示设备详情，应保留原始字段 | ✅ |
| M-18 | P1 | types.ts | types.ts:48-71 | `buildDeviceAssignments` | 用 `patient.bedNumber` 与 `device.bedNumber` 做字符串匹配关联 | — | — | — | — | **脆弱关联**：依赖床位名称完全匹配，若床位名称格式不一致则关联失败 | ⚠️ 需确认 | types.ts:52-63 | 确认床位名称在患者和设备表中格式一致；或改用 `bedId` 数字关联 | ✅ |

---

## 三、汇总统计

| 类别 | 数量 |
|------|------|
| P0 级问题（必须修复） | 8 |
| P1 级问题（应该修复） | 6 |
| 已通过核查 | 4 |

---

## 四、核心问题总结

### 4.1 🔴 监控数据完全缺失（M-03, M-04, M-06, M-08）

`MonitorDevice.vitals` 中的 9 个生命体征字段（sbp/dbp/hr/bf/tmp/ufGoal/ufVolume/conductivity/temp）在 `toMonitorDevice` 中全部硬编码为 0。前端图表数据 `cachedGraphData` 和 `cachedHistoryData` 始终为空数组。

**根本原因**：后端 `GET /api/v1/devices` 仅返回设备档案信息（名称/床位/状态），不包含任何透中实时监测数据。透中数据存储在 `Treatment_DuringParam` 和 `Treatment_DuringSigns` 表中，但监控页面未调用相关 API。

**数据源对照**：

| 前端字段 | 老库来源表 | 老库字段 |
|----------|-----------|----------|
| sbp | `Treatment_DuringSigns` | `SBP` (numeric) |
| dbp | `Treatment_DuringSigns` | `DBP` (numeric) |
| hr | `Treatment_DuringSigns` | `HeartRate` (numeric) |
| bf | `Treatment_DuringParam` | `BF` (numeric) |
| tmp | `Treatment_DuringParam` | `TMP` (numeric) |
| ufVolume | `Treatment_DuringParam` | `UFQuantity` (numeric) |
| conductivity | `Treatment_DuringParam` | `Conductivity` (numeric) |
| temp | `Treatment_DuringParam` | `MachineTmp` (numeric) |
| ufGoal | `Plan_PatientPlan` + 处方 | 需从处方计算 |

### 4.2 🔴 无实时数据推送（M-14）

`useMonitoringData` 仅在组件挂载时执行一次 `useEffect`，无后续刷新。作为"监控"页面，应具备以下能力之一：
- 短轮询（30s-60s 间隔重新拉取）
- WebSocket 推送设备状态变更
- SSE（Server-Sent Events）推送透中参数

### 4.3 ⚠️ 设备状态字段来源存疑（M-01, M-15）

设备状态存储在 `Schedule_BedEquipmentRel.ParameterS`（jsonb 字段），老库文档标注该字段"待人工补注"。当前后端将其用作字符串类型的状态存储（normal/warning/alarm/offline），属于兼容实现而非老库原生设计。

### 4.4 ⚠️ 前端占位字段未接入（M-02）

`StatusGrid` 中干体重、增重百分比、血管通路三列全部显示 `"--"`，这些数据需要从患者详情或治疗方案中获取。

---

## 五、建议改造方向

1. **新增实时监测数据 API**：创建 `GET /api/v1/monitoring/vitals` 接口，从当前进行中的治疗记录（`Treatment_Treatment.Status=30`）关联查询 `Treatment_DuringParam` + `Treatment_DuringSigns` 最新记录
2. **添加数据刷新机制**：在 `useMonitoringData` 中添加 30s 间隔轮询，或接入 WebSocket
3. **关联患者方案数据**：通过 `Schedule_PatientShift.PatientPlanId` → `Plan_PatientPlan` 获取干体重、透析模式、超滤目标等
4. **确认 `ParameterS` 用途**：与老系统维护方确认 jsonb 字段原始设计，评估是否需要新增专用状态列

---

## 六、需人工确认项

| 编号 | 确认项 | 涉及文件 |
|------|--------|----------|
| R-01 | `Schedule_BedEquipmentRel.ParameterS`（jsonb）在老系统中的原始用途是什么？ | device_service.go:419 |
| R-02 | 设备状态（normal/warning/alarm/offline）在老系统中如何判断？是否有其他字段或逻辑？ | device_service.go:312-325 |
| R-03 | 患者列表 API 是否返回 `defaultMode` 字段？ | restClient.ts:77, useMonitoringData.ts:29 |
| R-04 | `Order_PatientOrder` 表中 `frequency` 字段是否存在？来源是什么？ | restClient.ts:1389 |
| R-05 | 是否有现成的设备实时数据采集接口（如前置机 FEP 接口）可对接？ | Schedule_Bed.FEPId 字段 |
| R-06 | 老系统监控页面的数据刷新频率和机制是什么？ | — |
