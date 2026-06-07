# 排班 & 班次配置 — 字段级核查报告

> 核查日期：2026-05-31
> 核查范围：`/schedule`、`/shift-config`、排班模板、拖拽逻辑
> 核查依据：`老血透数据库表结构-合并版.md`、`排班管理.md`、`LEGACY_TABLE_FIELD_MAPPING.md`

---

## 一、总览

| 模块 | 后端目标表 | 核查结论 |
|------|-----------|---------|
| 班次配置 | `Schedule_Shift` | ⚠️ StartTime/EndTime 类型不匹配；Type 字段缺失 |
| 排班管理-周视图 | `Schedule_PatientShift` + `Schedule_Ward` + `Schedule_Bed` + `Schedule_Shift` | ⚠️ 多处类型/可空性不匹配 |
| 排班管理-创建/更新 | `Schedule_PatientShift` | ⚠️ Notes/IsDisabled 不落库；NN 字段可空 |
| 排班管理-换床/互换 | `Schedule_PatientShift` | ✅ 逻辑正确，字段映射基本到位 |
| 排班管理-模板 | `Schedule_PatientShift` (Status=60) | ❌ 前端调用路径/参数不匹配 |
| 排班管理-状态枚举 | Status 映射 | ⚠️ 前端历史弹窗显示旧值未映射 |
| 排班管理-拖拽 | `Schedule_PatientShift` | ✅ 逻辑正确 |

---

## 二、详细核查表

### 2.1 班次配置 (`/shift-config` → `ShiftConfig.tsx`)

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| S-1 | P0 | 班次配置-表格 | ShiftConfig.tsx:101 | 班次名称列 | `name` | GET /api/v1/shifts | schedule_handler.go:33 | `Name` varchar(256) | `Name` varchar(256) | ✅ 一致 | 通过 | - | - | 否 |
| S-2 | P0 | 班次配置-表格 | ShiftConfig.tsx:102 | 开始时间列 | `startTime` | GET /api/v1/shifts | schedule_handler.go:33 | `StartTime` varchar(32) | `StartTime` timestamp | ⚠️ 类型不匹配：老库为 timestamp，模型用 varchar(32) 存 HH:MM 文本 | 需改造 | schedule.go:87 `gorm:"column:StartTime;type:varchar(32)"` vs 老库 timestamp | 统一为 timestamp 或保持 varchar 但确认老库实际存储格式 | **是** |
| S-3 | P0 | 班次配置-表格 | ShiftConfig.tsx:103 | 结束时间列 | `endTime` | GET /api/v1/shifts | schedule_handler.go:33 | `EndTime` varchar(32) | `EndTime` timestamp | ⚠️ 同 S-2 | 需改造 | 同上 | 同上 | **是** |
| S-4 | P0 | 班次配置-表格 | ShiftConfig.tsx:104 | 排序列 | `sort` | GET /api/v1/shifts | schedule_handler.go:33 | `Sort` integer | `Sort` integer | ✅ 一致 | 通过 | - | - | 否 |
| S-5 | P0 | 班次配置-表格 | ShiftConfig.tsx:106-109 | 状态开关 | `isDisabled` | GET /api/v1/shifts | schedule_handler.go:33 | `IsDisabled` boolean | `IsDisabled` boolean | ✅ 一致 | 通过 | - | - | 否 |
| S-6 | P1 | 班次配置-编辑弹窗 | ShiftConfig.tsx:144-159 | 表单字段 | name/startTime/endTime/sort/notes | POST /api/v1/shifts | shift_service.go:70-76 | `Name/StartTime/EndTime/Sort/Note` | `Name/StartTime/EndTime/Sort/Note` | ⚠️ 前端未发送 `type` 字段；老库 `Type` 为 integer (10=长期/20=临时) | 缺失字段 | shift_service.go:74 `Type string` 存在但前端不传；老库 Schedule_Shift.Type 为 integer | 前端增加班次类型选择（长期/临时） | **是** |
| S-7 | P1 | 班次配置-编辑弹窗 | ShiftConfig.tsx:158 | 备注 | `notes` | POST/PUT /api/v1/shifts | shift_service.go:93,154 | `Notes` → `Note` 列 | `Note` varchar(512) | ✅ 一致（模型 tag 映射 column:Note） | 通过 | schedule.go:92 `gorm:"column:Note"` | - | 否 |
| S-8 | P0 | 班次配置-创建 | ShiftConfig.tsx:63-69 | 创建班次 | name/startTime/endTime/sort/notes | POST /api/v1/shifts | shift_service.go:80 | 写入 Schedule_Shift | Schedule_Shift | ⚠️ `Type` 字段未传，老库 Type 为 integer 非空 | 缺失 | shift_service.go:90 `Type: req.Type` 但 req.Type 为空 | 创建时必须指定 Type | **是** |
| S-9 | P0 | 班次配置-删除 | ShiftConfig.tsx:90-97 | 删除班次 | id | DELETE /api/v1/shifts/:id | shift_service.go:170 | `IsDisabled=true` 软删除 | `IsDisabled` boolean | ✅ 一致 | 通过 | shift_service.go:178 `Update("IsDisabled", true)` | - | 否 |

### 2.2 排班管理-周视图 (`/schedule` → `Schedule.tsx`)

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| W-1 | P0 | 周视图-病区列表 | Schedule.tsx:134 | 病区筛选 | `wards[].id/name/bedCount` | GET /api/v1/schedule/week | schedule_week_service.go:109-114 | 查询 Schedule_Ward | Schedule_Ward (12字段) | ✅ 查询字段匹配老库 | 通过 | schedule_week_service.go:110-112 Select Id,Name,Sort,PatientType,InfectionType,ResponsibleUsers | - | 否 |
| W-2 | P0 | 周视图-床位列表 | Schedule.tsx:137-140 | 床位渲染 | `beds[].id/name/wardId` | GET /api/v1/schedule/week | schedule_week_service.go:128-132 | 查询 Schedule_Bed | Schedule_Bed (12字段) | ✅ 查询字段匹配老库 | 通过 | schedule_week_service.go:129-130 Select Id,Name,WardId,Sort | - | 否 |
| W-3 | P0 | 周视图-班次列表 | Schedule.tsx:141 | 班次表头 | `shifts[].id/name/startTime/endTime` | GET /api/v1/schedule/week | schedule_week_service.go:176-181 | 查询 Schedule_Shift | Schedule_Shift (12字段) | ⚠️ 查询 StartTime/EndTime 用 ::text 转文本，老库为 timestamp | 兼容 | schedule_week_service.go:178-179 `"StartTime"::text` 做了类型转换 | 确认老库实际存储格式 | **是** |
| W-4 | P0 | 周视图-排班卡片 | Schedule.tsx:484-507 | 已排班卡片 | `patientName/dialysisMode/statusName/sourceType` | GET /api/v1/schedule/week | schedule_week_service.go:214-224 | JOIN Patient+Plan+Bed | Schedule_PatientShift+Register_PatientInfomation+Plan_PatientPlan | ✅ 多表 JOIN 字段正确 | 通过 | schedule_week_service.go:217-227 完整 JOIN 链 | - | 否 |
| W-5 | P1 | 周视图-排班卡片 | Schedule.tsx:498-506 | 来源标签 | `sourceType` | GET /api/v1/schedule/week | schedule_week_service.go:256-262 | 由 ShiftTiming 推导 | Schedule_PatientShift.ShiftTiming (10=临时/20=长期) | ⚠️ sourceType 逻辑：ShiftTiming=10→temporary, =20+有planId→contract, 其他→manual；老库无 sourceType 字段 | 兼容 | schedule_week_service.go:257-262 纯内存推导 | 确认是否需要"template"来源标记 | **是** |
| W-6 | P0 | 周视图-待排班队列 | Schedule.tsx:571-604 | 患者列表 | `name/gender/dialysisMode/oddWeekFrequency/evenWeekFrequency/remainingTimes` | GET /api/v1/schedule/week | schedule_week_service.go:299-318 | JOIN Plan_PatientPlan+Register_PatientInfomation | Plan_PatientPlan.OddWeekFrequency/EvenWeekFrequency/DialysisMethod + Register_PatientInfomation.Name/Gender/Spell | ✅ 字段映射正确 | 通过 | schedule_week_service.go:312-317 | - | 否 |
| W-7 | P1 | 周视图-待排班 | Schedule.tsx:598-599 | 剩余次数 | `remainingTimes` | GET /api/v1/schedule/week | schedule_week_service.go:335-338 | 本周频次 - 已排次数 | Plan_PatientPlan.OddWeekFrequency/EvenWeekFrequency | ⚠️ 计算逻辑用 ISO 周号判断单双周，老库是否有此约定 | 需确认 | schedule_week_service.go:297 `isoWeek%2==1` 判断奇偶 | 确认老系统的单双周算法 | **是** |
| W-8 | P1 | 周视图-病区标签 | Schedule.tsx:364 | 床数显示 | `w.bedCount` | GET /api/v1/schedule/week | schedule_week_service.go:154 | 统计 bedCount | Schedule_Bed | ✅ 内存统计 | 通过 | schedule_week_service.go:142-143 wardBedCount map | - | 否 |

### 2.3 排班管理-创建排班 (`Schedule.tsx` 弹窗)

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| C-1 | P0 | 创建排班-保存 | Schedule.tsx:236 | 创建排班 | `patientId/scheduleDate/shiftId/bedId/wardId/dialysisMode/patientPlanId/shiftTiming/status` | POST /api/v1/patient-shifts | patient_shift_service.go:171 | 写入 Schedule_PatientShift | Schedule_PatientShift (13字段) | ⚠️ 见下述子项 | 需改造 | - | - | - |
| C-2 | P0 | 创建排班-字段 | Schedule.tsx:236 | shiftTiming | `shiftTiming: 20` | POST /api/v1/patient-shifts | patient_shift_service.go:194 | `ShiftTiming *int` | `ShiftTiming` integer NN | ⚠️ 模型为可空指针 `*int`，老库 NOT NULL | 需改造 | patient_shift_service.go:194 `ShiftTiming: req.ShiftTiming`；schedule.go:124 `gorm:"column:ShiftTiming"` | 模型改为非指针或设默认值 | 否 |
| C-3 | P0 | 创建排班-字段 | Schedule.tsx:236 | status | `status: 1` | POST /api/v1/patient-shifts | patient_shift_service.go:181-184 | `Status` → 老库 20 | `Status` integer NN | ✅ status=1 → MapNewToLegacy → 20(已确认) | 通过 | legacy_enum_maps.go:37 `1: 20` | - | 否 |
| C-4 | P0 | 创建排班-字段 | Schedule.tsx:236 | bedId/wardId | `bedId/wardId` | POST /api/v1/patient-shifts | patient_shift_service.go:191-192 | `BedId *int64/WardId *int64` | `BedId` bigint NN / `WardId` bigint NN | ⚠️ 模型可空，老库 NOT NULL | 需改造 | schedule.go:121-122 注释说"legacy schema标注NN；为兼容历史请求暂保留可空" | 创建时强制要求 bedId/wardId | **是** |
| C-5 | P0 | 创建排班-字段 | Schedule.tsx:236 | patientPlanId | `patientPlanId` | POST /api/v1/patient-shifts | patient_shift_service.go:193 | `PatientPlanId *int64` | `PatientPlanId` bigint NN | ⚠️ 模型可空，老库 NOT NULL | 需改造 | schedule.go:123 同上 | 创建时强制要求或设默认值 | **是** |
| C-6 | P1 | 创建排班-字段 | Schedule.tsx:236 | dialysisMode | `dialysisMode` | POST /api/v1/patient-shifts | patient_shift_service.go:158-168 | 请求体有此字段但不写入 Schedule_PatientShift | Schedule_PatientShift 无此列 | ✅ 正确：dialysisMode 来自 Plan_PatientPlan.DialysisMethod，不存排班表 | 通过 | schedule_week_service.go:220 JOIN Plan_PatientPlan 获取 | - | 否 |
| C-7 | P2 | 创建排班-字段 | Schedule.tsx:236 | notes | `notes` | POST /api/v1/patient-shifts | patient_shift_service.go:197 | `Notes string gorm:"-"` | Schedule_PatientShift 无 Note 列 | ⚠️ Notes 字段 gorm:"-" 不落库，前端传了但丢弃 | 不落库 | schedule.go:127 `gorm:"-" json:"notes"` | 老库无此列，前端可不传 | 否 |

### 2.4 排班管理-修改排班

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| U-1 | P0 | 修改排班 | Schedule.tsx:226-233 | 更新排班 | `bedId/wardId/shiftId/treatmentTime/patientPlanId/shiftTiming` | PUT /api/v1/patient-shifts/:id | patient_shift_service.go:257 | 更新 Schedule_PatientShift | Schedule_PatientShift | ✅ 字段映射正确 | 通过 | patient_shift_service.go:302-323 updates map | - | 否 |
| U-2 | P0 | 修改排班 | Schedule.tsx:232 | shiftTiming: 20 | 硬编码 20 | PUT /api/v1/patient-shifts/:id | patient_shift_service.go:316 | `ShiftTiming` → 老库 20 | `ShiftTiming` integer | ✅ 20=长期，正确 | 通过 | - | - | 否 |
| U-3 | P0 | 修改排班-校验 | Schedule.tsx:244 | 历史保护 | `isDateLocked` 前端判断 | DELETE /api/v1/patient-shifts/:id | patient_shift_service.go:498-508 | ValidateShiftEdit | - | ✅ 前后端双重校验 | 通过 | Schedule.tsx:244 + patient_shift_service.go:505 | - | 否 |

### 2.5 排班管理-取消排班

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| D-1 | P0 | 取消排班 | Schedule.tsx:243-248 | 删除排班 | `id` | DELETE /api/v1/patient-shifts/:id | patient_shift_service.go:347-364 | `Status=50` (排班取消) | `Status` integer | ✅ 软删除，Status→50 正确 | 通过 | patient_shift_service.go:355 `MapPatientShiftStatusNewToLegacy(PatientShiftStatusCancelled)` → 50 | - | 否 |

### 2.6 排班管理-换床

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| M-1 | P0 | 换床弹窗 | Schedule.tsx:281 | 换床 | `bedId/wardId` | POST /api/v1/patient-shifts/:id/move | schedule_handler.go:455 | 更新 BedId/WardId | Schedule_PatientShift.BedId/WardId | ✅ 正确 | 通过 | schedule_handler.go:523-528 updateReq | - | 否 |
| M-2 | P0 | 换床弹窗 | Schedule.tsx:706-725 | 换床-目标床位选择 | `groupedBeds` 按病区分组 | POST /api/v1/patient-shifts/:id/move | - | - | - | ✅ 前端按病区分组展示床位，合理 | 通过 | - | - | 否 |

### 2.7 排班管理-拖拽

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| DR-1 | P0 | 拖拽到空格 | useScheduleDragDrop.ts:92 | 换床+换日期+换班次 | `bedId/wardId/treatmentTime/shiftId` | POST /api/v1/patient-shifts/:id/move | schedule_handler.go:455 | Move 全字段更新 | Schedule_PatientShift | ✅ 支持跨日期/跨班次/跨床位 | 通过 | schedule_handler.go:492-508 | - | 否 |
| DR-2 | P0 | 拖拽互换 | useScheduleDragDrop.ts:129 | 互换排班 | `sourceId/targetId` | POST /api/v1/patient-shifts/swap | schedule_handler.go:545 | Swap 事务交换 | Schedule_PatientShift | ✅ 事务交换 WardId/BedId/ShiftId/TreatmentTime | 通过 | patient_shift_service.go:372-406 | - | 否 |

### 2.8 排班管理-模板

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| T-1 | P0 | 模板列表页 | ScheduleTemplateList.tsx:15 | 加载模板列表 | fetch `/api/v1/schedule/template` | GET /api/v1/schedule/template | schedule_handler.go:692 | 路由不存在 | - | ❌ **致命**：后端只有 POST /schedule/template/save 和 POST /schedule/template/apply，无 GET 列表接口 | 不通过 | schedule_handler.go:692-696 只有 POST 路由 | 后端增加 GET /schedule/template 接口（从 Schedule_PatientShift Status=60 查询） | 否 |
| T-2 | P0 | 模板列表页 | ScheduleTemplateList.tsx:30-35 | 模板字段 | `name/cycleWeeks/isDefault/isEnabled` | - | - | - | Schedule_PatientShift 无这些字段 | ❌ **致命**：模板类型定义(ScheduleTemplate)与实际数据结构(WeekShiftItem)完全不匹配 | 不通过 | scheduleTemplate.ts:5-17 定义了不存在的字段 | 模板列表应使用 WeekShiftItem 结构或新建模板主表 | **是** |
| T-3 | P0 | 模板编辑器 | ScheduleTemplateEditor.tsx:14-16 | 编辑器 | "TODO: 待实现" | - | - | - | - | ❌ 未实现 | 不通过 | ScheduleTemplateEditor.tsx:15 | 需实现拖拽模板编辑器 | 否 |
| T-4 | P0 | 应用模板弹窗 | ApplyTemplateModal.tsx:18-22 | 应用模板 | `fetch POST {}` 空 body | POST /api/v1/schedule/template/apply | schedule_handler.go:636-653 | 需要 `targetDate` 字段 | - | ❌ **致命**：前端发空 body `{}`，后端需要 `targetDate` 字段 | 不通过 | ApplyTemplateModal.tsx:22 `body: JSON.stringify({})` vs schedule_handler.go:638 `TargetDate string` | 前端传入 targetDate（当前周起始日） | 否 |
| T-5 | P0 | 保存模板 | restClient.ts:2141 | 保存模板 | `items: {patientId,shiftId,wardId,bedId,patientPlanId,weekday}` | POST /api/v1/schedule/template/save | schedule_handler.go:618-633 | `items: []PatientShiftCreateRequest` | Schedule_PatientShift (Status=60) | ⚠️ 前端传 `weekday` 字段，后端 PatientShiftCreateRequest 无此字段 | 不通过 | restClient.ts:2141 vs patient_shift_service.go:158-168 | 前端去掉 weekday 或后端增加处理 | 否 |
| T-6 | P0 | 应用模板-后端 | patient_shift_service.go:656-707 | 应用模板逻辑 | 读 Status=60 记录 → 创建 Status=20 记录 | POST /api/v1/schedule/template/apply | patient_shift_service.go:656 | Status=60 → Status=20 | Schedule_PatientShift.Status | ✅ 逻辑正确：模板排班(60)→已确认(20) | 通过 | patient_shift_service.go:698 `Status: 20` | - | 否 |
| T-7 | P1 | 保存模板-后端 | patient_shift_service.go:623-654 | 保存模板 | 先删 Status=60 再批量创建 | POST /api/v1/schedule/template/save | patient_shift_service.go:628 | `Status: 60` 直接写入 | Schedule_PatientShift.Status | ⚠️ 写入 Status=60 未经 MapNewToLegacy 转换 | 需确认 | patient_shift_service.go:645 `Status: 60` 直接赋值 | 60 在老库中就是 60（转出人员），映射一致但语义不同 | **是** |

### 2.9 排班管理-状态枚举

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| ST-1 | P0 | 周视图-状态名 | schedule_week_service.go:83-90 | statusNames map | `{10:"草稿", 20:"已确认", 30:"用户确认", 40:"用户取消", 50:"排班取消", 60:"转出"}` | - | - | StatusName | Status | ⚠️ statusNames 使用老库原值（10/20/30/40/50/60），但 WeekShiftItem.Status 经过 LegacyToNew 转换为新值（0/1/3/5/4/6） | 不一致 | schedule_week_service.go:281 `Status: MapPatientShiftStatusLegacyToNew(r.Status)` 但 282 行 `StatusName: statusNames[r.Status]` 用原始值 | Status 和 StatusName 来源不一致，但显示正确（StatusName 用老值查 map） | 否 |
| ST-2 | P1 | 历史弹窗-状态 | Schedule.tsx:799-802 | 排班/换床记录状态 | `{10:"待确认", 20:"已确认", 30:"已完成", 50:"已取消", 60:"转出"}` | GET /api/v1/patient-shifts | patient_shift_service.go:112-114 | Status 经 LegacyToNew 转换 | Status | ⚠️ **不一致**：后端 List 返回 Status 已转为新值(0/1/3/4/6)，但前端用老值(10/20/30/50/60)显示 | 不通过 | patient_shift_service.go:113 `items[i].Status = MapPatientShiftStatusLegacyToNew(...)` 但 Schedule.tsx:800 用 `{10:...}` 映射 | 前端改用新值映射 `{0:"待确认", 1:"已确认", 3:"已完成", 4:"已取消", 6:"转出"}` | 否 |
| ST-3 | P0 | 周视图-状态映射 | schedule_week_service.go:281-282 | WeekShiftItem 状态 | Status 新值 + StatusName 老值 | GET /api/v1/schedule/week | schedule_week_service.go:281 | Status: 新值, StatusName: 老值名 | Status | ⚠️ Status 和 StatusName 来源不同但语义对应正确 | 兼容 | - | 统一使用新值或老值 | **是** |

### 2.10 排班管理-冲突校验

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| V-1 | P0 | 创建前校验 | - | 患者冲突 | - | POST /api/v1/patient-shifts | patient_shift_service.go:436-460 | CheckConflict: PatientId+Date+ShiftId | - | ✅ 排除 Status 50/40/60 | 通过 | patient_shift_service.go:444-448 | - | 否 |
| V-2 | P0 | 创建前校验 | - | 床位冲突 | - | POST /api/v1/patient-shifts | patient_shift_service.go:462-487 | CheckBedConflict: BedId+Date+ShiftId | - | ✅ 排除 Status 50/40/60 | 通过 | - | - | 否 |
| V-3 | P0 | 编辑前校验 | - | 历史保护+已过班次+治疗中 | - | PUT/DELETE /api/v1/patient-shifts/:id | patient_shift_service.go:498-551 | ValidateShiftEdit | - | ✅ 三重校验完整 | 通过 | - | - | 否 |

### 2.11 前端遗留服务 (`schedule.ts`)

| 编号 | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|-------------|-----------|
| L-1 | P2 | schedule.ts | schedule.ts:23 | SHIFT_FIELDS | `Id,TenantId,Name,Sort,StartTime,EndTime,Type,Note` | GraphQL | - | - | Schedule_Shift | ⚠️ 此文件使用 GraphQL 接口（fetchPaginatedData），与 REST API 并行存在 | 存疑 | schedule.ts:44 `fetchPaginatedData<Shift>('Shift', SHIFT_FIELDS, ...)` | 确认 GraphQL 接口是否仍可用；如已废弃则清理此文件 | **是** |
| L-2 | P2 | schedule.ts | schedule.ts:26-28 | PATIENT_SHIFT_FIELDS | `Id,TenantId,PatientId,TreatmentTime,ShiftId,WardId,BedId,PatientPlanId` | GraphQL | - | - | Schedule_PatientShift | ⚠️ 缺少 ShiftTiming/Status/CreatorId 等字段 | 存疑 | schedule.ts:26-28 | 补齐字段或废弃此文件 | **是** |

### 2.12 老库字段名差异汇总

| 编号 | 优先级 | 模块 | 后端模型字段 | 老库物理列 | 差异说明 | 影响 | 需人工确认 |
|------|--------|------|------------|----------|---------|------|-----------|
| F-1 | P0 | Schedule_Shift | `StartTime varchar(32)` | `StartTime timestamp` | 类型不匹配 | GORM 可能自动转换，但需确认老库存储格式 | **是** |
| F-2 | P0 | Schedule_Shift | `EndTime varchar(32)` | `EndTime timestamp` | 类型不匹配 | 同上 | **是** |
| F-3 | P0 | Schedule_Shift | `Type varchar(64)` | `Type integer` | 类型不匹配：老库为整数(10=长期/20=临时) | 前端不传此字段，创建时为空 | **是** |
| F-4 | P0 | Schedule_PatientShift | `BedId *int64` (可空) | `BedId bigint NN` | 可空性不匹配 | 创建时可能写入 NULL | **是** |
| F-5 | P0 | Schedule_PatientShift | `WardId *int64` (可空) | `WardId bigint NN` | 可空性不匹配 | 同上 | **是** |
| F-6 | P0 | Schedule_PatientShift | `PatientPlanId *int64` (可空) | `PatientPlanId bigint NN` | 可空性不匹配 | 同上 | **是** |
| F-7 | P0 | Schedule_PatientShift | `ShiftTiming *int` (可空) | `ShiftTiming integer NN` | 可空性不匹配 | 同上 | **是** |
| F-8 | P2 | Schedule_PatientShift | `IsDisabled bool gorm:"-"` | 无此列 | 模型有但老库无 | 不落库，无影响 | 否 |
| F-9 | P2 | Schedule_PatientShift | `Notes string gorm:"-"` | 无此列 | 模型有但老库无 | 不落库，无影响 | 否 |
| F-10 | P1 | Schedule_Bed | 后端文档 `EquipmentConnectId` | 老库 `AcquisiteConnectId` | 文档与实际代码不一致 | 代码已用 `AcquisiteConnectId`，仅文档需修正 | 否 |

---

## 三、不一致/缺失/待确认项汇总

### 3.1 ❌ 致命问题（阻塞使用）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 1 | 模板列表页 GET 接口不存在 | ScheduleTemplateList.tsx:15 vs schedule_handler.go:692 | 模板列表页无法加载数据 |
| 2 | 模板数据类型完全不匹配 | scheduleTemplate.ts:5-17 vs 实际 WeekShiftItem | 前端模板类型定义与后端返回结构不符 |
| 3 | 应用模板弹窗发送空 body | ApplyTemplateModal.tsx:22 vs schedule_handler.go:638 | 后端需要 targetDate 但前端不传，必报错 |
| 4 | 保存模板 weekday 字段不存在 | restClient.ts:2141 vs patient_shift_service.go:158 | PatientShiftCreateRequest 无 weekday 字段 |
| 5 | 模板编辑器未实现 | ScheduleTemplateEditor.tsx:14-16 | 功能缺失 |

### 3.2 ⚠️ 高优问题（影响数据正确性）

| # | 问题 | 位置 | 影响 |
|---|------|------|------|
| 6 | Schedule_Shift StartTime/EndTime 类型 mismatch | schedule.go:87-88 vs 老库 timestamp | 存储格式可能不一致 |
| 7 | Schedule_Shift.Type 整数/字符串 mismatch | schedule.go:89 vs 老库 integer | 前端不传 Type，创建班次无类型 |
| 8 | Schedule_PatientShift NN 字段可空 | schedule.go:121-124 | BedId/WardId/PatientPlanId/ShiftTiming 可写入 NULL |
| 9 | 历史弹窗状态映射错误 | Schedule.tsx:799-802 vs patient_shift_service.go:113 | 后端返回新值(0/1/3/4/6)但前端用老值(10/20/30/50/60)显示 |

### 3.3 📋 待确认项

| # | 问题 | 需确认内容 |
|---|------|-----------|
| 10 | StartTime/EndTime 实际存储格式 | 老库 timestamp 列实际存的是完整时间戳还是仅 HH:MM？ |
| 11 | Type 字段语义 | 老库 Type=10(长期)/20(临时) 是否仍在使用？前端是否需要展示？ |
| 12 | 单双周算法 | schedule_week_service.go:297 ISO 周号取模是否与老系统一致？ |
| 13 | 模板 Status=60 语义 | 老库 60=转出人员，新系统用于模板排班，语义是否冲突？ |
| 14 | schedule.ts GraphQL 服务 | 是否仍可用？是否需要废弃清理？ |

---

## 四、改造建议优先级排序

1. **P0-紧急**：修复应用模板弹窗（传 targetDate）+ 后端增加 GET 模板列表接口
2. **P0-紧急**：统一 Schedule_Shift StartTime/EndTime 存储类型
3. **P0-紧急**：Schedule_PatientShift NN 字段设默认值或强制非空
4. **P0-紧急**：修复历史弹窗状态映射（前端改用新值）
5. **P1-重要**：前端增加班次类型选择（长期/临时）
6. **P1-重要**：保存模板接口参数对齐
7. **P1-重要**：实现模板编辑器
8. **P2-低**：清理 schedule.ts GraphQL 遗留代码
