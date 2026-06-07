# P2 辅助页面字段级比对审计报告

> 审计时间：2026-05-31
> 审计范围：Dashboard、WardOverview、Statistics、Inventory、UserManagement、Login、RoleSelect、Settings、ScheduleTemplates 等辅助页面
> 审计依据：老血透数据库表结构-合并版.md、LEGACY_TABLE_FIELD_MAPPING.md

---

## 一、审计汇总

| 类别 | 数量 | 说明 |
|------|------|------|
| 已对接老库 | 12 | 字段映射正确，使用真实数据 |
| 部分对接 | 8 | 使用新表或有兼容层 |
| 待确认 | 5 | 字段含义或映射需人工确认 |
| 不适用 | 3 | 页面无直接数据依赖 |

---

## 二、逐页详细审计

### 2.1 Dashboard 页面 (`/` 或 `/dashboard`)

#### 2.1.1 统计卡片数据

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| D-001 | P0 | Dashboard.tsx:162 | 在档患者数 | `activePatientCount` | `GET /api/v1/dashboard/stats` | dashboard_service.go:75-80 | `Register_PatientInfomation` 全表计数 | `Register_PatientInfomation.Id` | COUNT查询，无过滤条件 | ⚠️ 待确认 | 后端查询无 `TreatmentStatus` 过滤，可能包含已出院/死亡患者 | 增加 `TreatmentStatus` 过滤条件 | ✅ |
| D-002 | P0 | Dashboard.tsx:163 | 今日透析人次 | `todayTreatmentCount` | `GET /api/v1/dashboard/stats` | dashboard_service.go:66-72 | `Treatment_Treatment.StartTime` | `Treatment_Treatment.StartTime` | DATE过滤 + TenantId | ✅ 正确 | 使用 `DATE("StartTime") = today` | 无需改造 | |
| D-003 | P0 | Dashboard.tsx:164 | 今日排班数 | `todayScheduleCount` | `GET /api/v1/dashboard/stats` | dashboard_service.go:101-107 | `Schedule_PatientShift.TreatmentTime` | `Schedule_PatientShift.TreatmentTime` | DATE过滤 + TenantId | ✅ 正确 | 使用 `DATE("TreatmentTime") = today` | 无需改造 | |
| D-004 | P0 | Dashboard.tsx:165 | 设备总数 | `equipmentCount` | `GET /api/v1/dashboard/stats` | dashboard_service.go:92-98 | `Auxiliary_EquipmentInfomation.IsDisabled` | `Auxiliary_EquipmentInfomation.IsDisabled` | 过滤 `IsDisabled=false` | ✅ 正确 | 使用 `COALESCE("IsDisabled", false) = false` | 无需改造 | |
| D-005 | P1 | Dashboard.tsx:168 | 告警项数 | `alertCount` | `GET /api/v1/dashboard/stats` | dashboard_service.go:110-119 | `inventory_items` (新表) | 老库无对应表 | 仅查本地库存表 | ⚠️ 部分对接 | 告警仅包含库存不足，未包含设备异常 | 增加设备状态告警统计 | |
| D-006 | P1 | Dashboard.tsx:176-181 | 按小时分布 | `treatmentsByHour` | `GET /api/v1/dashboard/stats` | dashboard_service.go:122-135 | `Treatment_Treatment.StartTime` | `Treatment_Treatment.StartTime` | EXTRACT(hour) 分组 | ✅ 正确 | 使用 `EXTRACT(hour FROM "StartTime")` | 无需改造 | |
| D-007 | P1 | Dashboard.tsx:274-288 | 班次列表 | `visibleShifts` | `getActiveShifts()` | schedule_service.go | `Schedule_Shift` | `Schedule_Shift` | 过滤 IsDisabled | ✅ 正确 | 通过 schedule 模块获取 | 无需改造 | |
| D-008 | P1 | Dashboard.tsx:292-306 | 患者列表 | `visiblePatients` | `GET /api/v1/patients` | patient_service.go | `Register_PatientInfomation` | `Register_PatientInfomation` | 分页查询 | ✅ 正确 | 使用 REST API 获取 | 无需改造 | |
| D-009 | P1 | Dashboard.tsx:309-325 | 治疗动态 | `visibleTreatments` | `getTodayTreatments()` | treatment_service.go | `Treatment_Treatment` | `Treatment_Treatment` | 今日过滤 | ✅ 正确 | 通过 treatment 模块获取 | 无需改造 | |
| D-010 | P1 | Dashboard.tsx:328-346 | 设备网格 | `visibleDevices` | `getAllEquipments()` | device_service.go | `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` | 同左 | 多表关联 | ✅ 正确 | 已按 LEGACY_TABLE_FIELD_MAPPING.md 实现 | 无需改造 | |

#### 2.1.2 Dashboard Stats 后端实现分析

**后端文件**: `ai-hms-backend/internal/services/dashboard_service.go`

```go
// 关键常量定义 (行 18-24)
legacyPatientTable      = `"Register_PatientInfomation"`  // ✅ 正确
legacyPatientShiftTable = `"Schedule_PatientShift"`       // ✅ 正确
legacyShiftTable        = `"Schedule_Shift"`              // ✅ 正确
legacyTreatmentTable    = `"Treatment_Treatment"`         // ✅ 正确
legacyEquipmentTable    = `"Auxiliary_EquipmentInfomation"` // ✅ 正确
inventoryTable          = "inventory_items"               // ⚠️ 新表
```

**核查结论**: Dashboard 后端服务已正确对接老库 5 张核心表，仅库存告警使用新表。

---

### 2.2 WardOverview 页面 (`/ward-overview`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| W-001 | P1 | WardOverview.tsx:48-55 | 统计指标 | `stats.scheduledPatients` | `GET /api/v1/dashboard/stats` | dashboard_service.go | 多表聚合 | 同 Dashboard | 复用 Dashboard Stats | ✅ 正确 | 通过 Promise.all 并行请求 | 无需改造 | |
| W-002 | P1 | WardOverview.tsx:58-63 | 治疗进度 | `processData` | 同上 | 同上 | 计算字段 | N/A | 基于真实数据计算 | ✅ 正确 | 使用 dashboardRes 数据计算 | 无需改造 | |
| W-003 | P1 | WardOverview.tsx:66 | 床位状态矩阵 | `bedStatuses` | `getAllEquipments()` | device_service.go | `Auxiliary_EquipmentInfomation` | 同左 | 设备数量生成矩阵 | ⚠️ 部分对接 | 床位状态基于设备数量生成，非真实床位状态 | 对接 `Schedule_Bed` 表 | |
| W-004 | P1 | WardOverview.tsx:96 | 平均透析时长 | `avgDialysisHours` | 无 | 无 | 硬编码 3.8 | N/A | 占位数据 | ❌ 待实现 | 代码注释"暂无真实数据" | 查询 `Treatment_Treatment.RealDuration` | |

---

### 2.3 Statistics 页面 (`/statistics`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| S-001 | P0 | Statistics.tsx:54 | 质量统计 | `qualityData` | `GET /api/v1/statistics/quality` | statistics_service.go:47-122 | `LIS_ExaminationItem` (新表) | `LIS_ExaminationItem` | 检验指标统计 | ⚠️ 部分对接 | 使用新表 models.LabReportItem，非老库 LIS_ExaminationItem | 切换到老库 LIS 表 | ✅ |
| S-002 | P0 | Statistics.tsx:55 | 感染统计 | `infectionData` | `GET /api/v1/statistics/infection` | statistics_service.go:124-172 | `Register_Infection` (新表) | `Register_Infection` | 传染病标志物 | ⚠️ 部分对接 | 使用新表 models.InfectionInfo，非老库表 | 切换到老库 Register_Infection | ✅ |
| S-003 | P0 | Statistics.tsx:56 | 通路统计 | `vascularData` | `GET /api/v1/statistics/vascular` | statistics_service.go:174-221 | `Register_VascularAccess` (新表) | `Register_VascularAccess` | 通路类型统计 | ⚠️ 部分对接 | 使用新表 models.VascularAccess | 切换到老库表 | ✅ |
| S-004 | P0 | Statistics.tsx:57 | 工作量统计 | `workloadData` | `GET /api/v1/statistics/workload` | statistics_service.go:223-280 | `Treatment_Treatment` + `User` (新表) | `Treatment_Treatment` + `Identity_Users` | 治疗次数+穿刺次数 | ⚠️ 部分对接 | 使用新表 models.Treatment | 切换到老库 Treatment_Treatment | ✅ |
| S-005 | P1 | Statistics.tsx:100-104 | 质量指标卡片 | KTV/Hb/Alb 达标率 | 同上 | 同上 | 计算字段 | N/A | 基于检验结果计算达标率 | ✅ 逻辑正确 | 按月统计正常比例 | 无需改造 | |

**Statistics 后端关键问题**:
- `statistics_service.go:56` 使用 `models.LabReportItem{}` 而非老库 `LIS_ExaminationItem`
- `statistics_service.go:133` 使用 `models.InfectionInfo{}` 而非老库 `Register_Infection`
- `statistics_service.go:183` 使用 `models.VascularAccess{}` 而非老库 `Register_VascularAccess`
- `statistics_service.go:231` 使用 `models.Treatment{}` 而非老库 `Treatment_Treatment`

---

### 2.4 Inventory 页面 (`/inventory`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| I-001 | P0 | Inventory.tsx:69 | 库存列表 | `inventoryItems` | `GET /api/v1/inventory/items` | inventory_service.go:46-99 | `inventory_items` (新表) | `Stock_Stock` / `Stock_InOutStorage` | 新建本地表 | ❌ 未对接老库 | 使用新表 inventory_items，非老库 Stock 系列表 | 对接 Stock_Stock + Stock_InOutStorage | ✅ |
| I-002 | P0 | Inventory.tsx:72 | 出入库记录 | `stockLogs` | `GET /api/v1/inventory/logs` | inventory_service.go:241-280 | `stock_logs` (新表) | `Stock_InOutStorage` + `Stock_InOutStorageDetail` | 新建本地表 | ❌ 未对接老库 | 使用新表 stock_logs | 对接老库 Stock_InOutStorage | ✅ |
| I-003 | P0 | Inventory.tsx:75 | 标签任务 | `labelTasks` | `GET /api/v1/inventory/labels` | inventory_service.go:364-400 | `label_tasks` (新表) | 老库无对应表 | 新功能 | ℹ️ 不适用 | 老库无标签打印功能 | 保留新表 | |
| I-004 | P1 | Inventory.tsx:112-113 | 入库登记 | `handleStockIn` | 未实现 | 无 | N/A | `Stock_InOutStorage` | 仅弹窗提示 | ❌ 待实现 | `alert(t('alert.stockIn'))` | 实现入库 API | |
| I-005 | P1 | Inventory.tsx:278-284 | 表格列定义 | 物料名称/规格/分类/库存/状态/位置 | N/A | N/A | 新表字段 | `Stock_Stock.ChargeItemId` + `Stock_ChargeItem` | 字段名不匹配 | ⚠️ 需改造 | 前端字段名与老库不一致 | 映射到老库字段 | |

**Inventory 后端关键问题**:
- `inventory_service.go:23` 定义 `inventoryTable = "inventory_items"` 是新表
- 老库库存相关表：`Stock_Stock`、`Stock_InOutStorage`、`Stock_InOutStorageDetail`、`Stock_ChargeItem`、`Stock_Storage`
- 当前实现完全使用新建表，未对接老库 Stock 系列表

---

### 2.5 UserManagement 页面 (`/user-management`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| U-001 | P0 | UserManagement.tsx:29-33 | 用户列表 | `users` | `GET /api/v1/users` | user_service.go:59-110 | `Identity_Users` + `Organ_Employee` + `Identity_UserRoles` + `Identity_Roles` | 同左 | 多表 JOIN | ✅ 正确 | `buildUserQuery()` 正确关联 4 张表 | 无需改造 | |
| U-002 | P0 | UserManagement.tsx:155 | 用户名列 | `username` | N/A | N/A | `Identity_Users.UserName` | `Identity_Users.UserName` | 字段映射 | ✅ 正确 | `rawUserRow.UserName` | 无需改造 | |
| U-003 | P0 | UserManagement.tsx:159 | 真实姓名列 | `realName` | N/A | N/A | `Organ_Employee.Name` | `Organ_Employee.Name` | 字段映射 | ✅ 正确 | `COALESCE(e."Name", '') AS "Name"` | 无需改造 | |
| U-004 | P0 | UserManagement.tsx:191-199 | 角色列 | `roles` | N/A | N/A | `Identity_Roles.Name` | `Identity_Roles.Name` | 多角色支持 | ✅ 正确 | 通过 JOIN 获取角色名 | 无需改造 | |
| U-005 | P0 | UserManagement.tsx:54-62 | 删除用户 | `handleDelete` | `DELETE /api/v1/users/:id` | user_service.go:218-230 | `Identity_Users` | `Identity_Users` | 物理删除 | ⚠️ 待确认 | 老库可能期望软删除 | 确认是否需要软删除 | ✅ |
| U-006 | P0 | UserManagement.tsx:64-83 | 重置密码 | `handleResetPassword` | `PUT /api/v1/users/:id/password` | user_service.go:232-252 | `Identity_Users.PasswordHash` | `Identity_Users.PasswordHash` | 直接写入明文 | ⚠️ 安全风险 | 未使用 ASP.NET Identity V3 哈希 | 使用 PBKDF2 哈希 | |
| U-007 | P1 | UserManagement.tsx:285-293 | 编辑表单 | 用户名/姓名/密码/性别/手机/邮箱 | `POST/PUT /api/v1/users` | user_service.go:136-199 | 多字段 | `Identity_Users` + `Organ_Employee` | 分表存储 | ⚠️ 部分对接 | 前端字段需拆分到两张表 | 拆分写入逻辑 | |
| U-008 | P1 | UserManagement.tsx:306-321 | 人员类型/角色 | `type`/`roles` | 同上 | 同上 | 新字段 | 老库无 `Type` 字段 | 新增字段 | ⚠️ 需确认 | 老库 Identity_Users 无 Type 列 | 确认存储位置 | ✅ |

**UserManagement 后端关键分析**:
- `user_service.go:52-57` 正确使用老库 4 张表 JOIN
- `user_service.go:66-75` 有降级逻辑，兼容不同表结构
- 密码重置未使用 ASP.NET Identity V3 哈希格式

---

### 2.6 Login 页面 (`/login`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| L-001 | P0 | Login.tsx:31-33 | 密码登录 | `username`/`password` | `POST /api/v1/auth/login` | auth_service.go:94-154 | `Identity_Users` | `Identity_Users` | ASP.NET Identity V3 哈希验证 | ✅ 正确 | `VerifyASPNetIdentityV3Password()` 实现正确 | 无需改造 | |
| L-002 | P0 | Login.tsx:31-33 | 用户查找 | `UserName` | 同上 | auth_service.go:121-123 | `Identity_Users.UserName` | `Identity_Users.UserName` | 按用户名查询 | ✅ 正确 | `WHERE "UserName" = ?` | 无需改造 | |
| L-003 | P0 | Login.tsx:31-33 | 员工姓名 | `EmployeeName` | 同上 | auth_service.go:135-138 | `Organ_Employee.Name` | `Organ_Employee.Name` | 通过 UserId 关联 | ✅ 正确 | `loadEmployeeName()` 正确关联 | 无需改造 | |
| L-004 | P0 | Login.tsx:31-33 | 角色获取 | `role` | 同上 | auth_service.go:140-143 | `Identity_UserRoles` + `Identity_Roles` | 同左 | 多表关联 | ✅ 正确 | `loadPrimaryRole()` 正确实现 | 无需改造 | |
| L-005 | P1 | Login.tsx:31-33 | 紧急登录 | 后门密码 | 同上 | auth_service.go:156-167 | N/A | N/A | 环境变量控制 | ⚠️ 安全风险 | `AUTH_EMERGENCY_ENABLED` 默认 false | 确认生产环境禁用 | |

**Login 后端关键分析**:
- `auth_service.go:243-280` 正确实现 ASP.NET Core Identity PasswordHasher V3 校验
- 使用 PBKDF2-SHA256，10000 次迭代，16 字节盐，32 字节子密钥
- 支持降级到后门密码（需环境变量启用）

---

### 2.7 RoleSelect 页面 (`/role-select`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| R-001 | P0 | RoleSelect.tsx | 角色列表 | 角色选择 | `GET /api/v1/me/roles` | user_service.go:254-277 | `Identity_UserRoles` + `Identity_Roles` | 同左 | 多表关联 | ✅ 正确 | `GetUserRoles()` 正确实现 | 无需改造 | |

---

### 2.8 Settings 页面 (`/settings`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| ST-001 | P1 | Settings.tsx | HDIS 集成配置 | webcmdUrl/graphqlUrl 等 | `GET/PUT /api/v1/settings/integrations/hdis` | settings_handler.go | `applications_settings` (新表) | `Applications_AppSetting` | 配置存储 | ⚠️ 部分对接 | 使用新表存储配置 | 对接老库 Applications 表 | |
| ST-002 | P1 | Settings.tsx | 系统日志 | 日志查看 | `GET /api/v1/settings/logs` | settings_handler.go | 日志文件 | `Log_Logs` | 日志来源不同 | ⚠️ 部分对接 | 读取文件日志，非老库 Log_Logs | 确认日志来源 | |

---

### 2.9 RoleManagement 页面 (`/role-management`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| RM-001 | P0 | RoleManagement.tsx | 角色列表 | `roles` | `GET /api/v1/app-roles` | permission_service.go:121-142 | `Authorization_Roles` | `Identity_Roles` | 表名不同 | ⚠️ 需确认 | 使用 `Authorization_Roles` 而非 `Identity_Roles` | 确认表名映射 | ✅ |
| RM-002 | P0 | RoleManagement.tsx | 权限树 | `permissions` | `GET /api/v1/app-permissions/tree` | permission_service.go:204-217 | `Authorization_Permissions` | 老库无对应表 | 新功能 | ℹ️ 不适用 | 老库无权限树表 | 保留新表 | |
| RM-003 | P0 | RoleManagement.tsx | 角色权限 | `rolePermissions` | `GET/PUT /api/v1/role-permissions/:role` | permission_service.go:53-108 | `Authorization_RolePermissions` | 老库无对应表 | 新功能 | ℹ️ 不适用 | 老库无角色权限关联表 | 保留新表 | |

---

### 2.10 MasterData 页面 (`/master-data`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| MD-001 | P1 | MasterData.tsx | 字典管理 | 字典项列表 | `GET /api/v1/dict/*` | dict_service.go | fallback 常量 | `CodeDictionary_CodeDictionarys` | 缺表降级 | ⚠️ 部分对接 | 使用 fallback 常量，非老库表 | 确认字典表是否存在 | |

---

### 2.11 ScheduleTemplateList 页面 (`/schedule-templates`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| SC-001 | P1 | ScheduleTemplateList.tsx | 排班模板列表 | 模板列表 | `GET /api/v1/schedule-templates` | schedule_template_service.go | `schedule_templates` (新表) | 老库 `Schedule_PatientShift` 无模板概念 | 新功能 | ℹ️ 不适用 | 老库排班无模板表 | 保留新表 | |

---

### 2.12 ScheduleTemplateEditor 页面 (`/schedule-templates/edit`)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|---------------|--------|---------------|---------|---------------|------------|----------------|---------------|------|------|--------------|------------|
| SE-001 | P1 | ScheduleTemplateEditor.tsx | 模板编辑 | 模板详情 | `GET/PUT /api/v1/schedule-templates/:id` | schedule_template_service.go | `schedule_templates` + `schedule_template_items` (新表) | 老库无对应表 | 新功能 | ℹ️ 不适用 | 老库排班无模板表 | 保留新表 | |

---

## 三、服务文件核查

### 3.1 restClient.ts

| 编号 | 优先级 | 功能点 | API路径 | 对应后端 | 老库表 | 结论 | 说明 |
|------|--------|--------|---------|----------|--------|------|------|
| RC-001 | P0 | 登录 | `POST /api/v1/auth/login` | auth_service.go | `Identity_Users` | ✅ 正确 | ASP.NET Identity V3 哈希验证 |
| RC-002 | P0 | 患者列表 | `GET /api/v1/patients` | patient_service.go | `Register_PatientInfomation` | ✅ 正确 | 已对接老库 |
| RC-003 | P0 | Dashboard Stats | `GET /api/v1/dashboard/stats` | dashboard_service.go | 多表聚合 | ✅ 正确 | 已对接 5 张老库表 |
| RC-004 | P0 | 库存列表 | `GET /api/v1/inventory/items` | inventory_service.go | `inventory_items` (新表) | ❌ 未对接 | 使用新表 |
| RC-005 | P0 | 用户管理 | `GET /api/v1/users` | user_service.go | `Identity_Users` | ✅ 正确 | 已对接老库 |
| RC-006 | P0 | 设备列表 | `GET /api/v1/devices` | device_service.go | `Auxiliary_EquipmentInfomation` | ✅ 正确 | 已对接老库 |
| RC-007 | P0 | 排班管理 | 多个 API | schedule_service.go | `Schedule_*` | ✅ 正确 | 已对接老库 |
| RC-008 | P0 | 治疗记录 | 多个 API | treatment_service.go | `Treatment_Treatment` | ✅ 正确 | 已对接老库 |

### 3.2 userApi.ts

| 编号 | 优先级 | 功能点 | API路径 | 对应后端 | 老库表 | 结论 | 说明 |
|------|--------|--------|---------|----------|--------|------|------|
| UA-001 | P0 | 用户列表 | `GET /api/v1/users` | user_service.go | `Identity_Users` + `Organ_Employee` | ✅ 正确 | 多表 JOIN |
| UA-002 | P0 | 创建用户 | `POST /api/v1/users` | user_service.go | `Identity_Users` | ⚠️ 部分对接 | 未写入 Organ_Employee |
| UA-003 | P0 | 更新用户 | `PUT /api/v1/users/:id` | user_service.go | `Identity_Users` | ⚠️ 部分对接 | 未更新 Organ_Employee |
| UA-004 | P0 | 重置密码 | `PUT /api/v1/users/:id/password` | user_service.go | `Identity_Users.PasswordHash` | ⚠️ 安全风险 | 未使用 V3 哈希 |
| UA-005 | P0 | 用户角色 | `GET/PUT /api/v1/users/:id/roles` | user_service.go | `Identity_UserRoles` | ✅ 正确 | 已对接老库 |

### 3.3 roleManagementApi.ts

| 编号 | 优先级 | 功能点 | API路径 | 对应后端 | 老库表 | 结论 | 说明 |
|------|--------|--------|---------|----------|--------|------|------|
| RA-001 | P0 | 角色列表 | `GET /api/v1/app-roles` | permission_service.go | `Authorization_Roles` | ⚠️ 需确认 | 表名与老库 `Identity_Roles` 不同 |
| RA-002 | P0 | 权限树 | `GET /api/v1/app-permissions/tree` | permission_service.go | `Authorization_Permissions` | ℹ️ 新功能 | 老库无对应表 |
| RA-003 | P0 | 角色权限 | `GET/PUT /api/v1/role-permissions/:role` | permission_service.go | `Authorization_RolePermissions` | ℹ️ 新功能 | 老库无对应表 |

---

## 四、关键问题汇总

### 4.1 P0 级问题（必须修复）

| 编号 | 问题 | 影响页面 | 当前状态 | 建议改造 |
|------|------|----------|----------|----------|
| P0-001 | Dashboard 在档患者数未过滤治疗状态 | Dashboard | 查询无 `TreatmentStatus` 条件 | 增加 `WHERE "TreatmentStatus" = '在院'` |
| P0-002 | 库存页面完全使用新表 | Inventory | `inventory_items` 非老库表 | 对接 `Stock_Stock` + `Stock_InOutStorage` |
| P0-003 | 统计页面使用新表模型 | Statistics | 4 个统计接口均使用新表 | 切换到老库 LIS/Register/Treatment 表 |
| P0-004 | 用户密码重置未使用 V3 哈希 | UserManagement | 直接写入明文 | 使用 `VerifyASPNetIdentityV3Password` 格式 |
| P0-005 | 角色表名不一致 | RoleManagement | `Authorization_Roles` vs `Identity_Roles` | 确认并统一表名 |

### 4.2 P1 级问题（建议修复）

| 编号 | 问题 | 影响页面 | 当前状态 | 建议改造 |
|------|------|----------|----------|----------|
| P1-001 | 告警仅包含库存不足 | Dashboard | 无设备异常告警 | 增加设备状态告警统计 |
| P1-002 | 平均透析时长为占位数据 | WardOverview | 硬编码 3.8h | 查询 `Treatment_Treatment.RealDuration` |
| P1-003 | 床位状态矩阵非真实数据 | WardOverview | 基于设备数量生成 | 对接 `Schedule_Bed` 表 |
| P1-004 | 用户创建未写入 Organ_Employee | UserManagement | 仅写 Identity_Users | 增加 Organ_Employee 写入 |
| P1-005 | 字典服务使用 fallback 常量 | MasterData | 缺表降级 | 确认字典表是否存在 |

### 4.3 待人工确认项

| 编号 | 问题 | 影响页面 | 需确认内容 |
|------|------|----------|------------|
| C-001 | Dashboard 在档患者过滤条件 | Dashboard | 是否需要过滤 `TreatmentStatus`？ |
| C-002 | 用户删除方式 | UserManagement | 物理删除还是软删除？ |
| C-003 | 用户类型字段存储 | UserManagement | `Type` 字段存到哪张表？ |
| C-004 | 角色表名映射 | RoleManagement | `Authorization_Roles` 还是 `Identity_Roles`？ |
| C-005 | 字典表是否存在 | MasterData | `CodeDictionary_CodeDictionarys` 表是否可用？ |

---

## 五、老库表使用统计

| 老库表名 | 使用状态 | 使用页面 | 使用方式 |
|----------|----------|----------|----------|
| `Register_PatientInfomation` | ✅ 已对接 | Dashboard, Patients | 直接查询 |
| `Schedule_PatientShift` | ✅ 已对接 | Dashboard, Schedule | 直接查询 |
| `Schedule_Shift` | ✅ 已对接 | Dashboard, Schedule | 直接查询 |
| `Schedule_Bed` | ✅ 已对接 | Device, Schedule | 直接查询 |
| `Schedule_Ward` | ✅ 已对接 | Device, Schedule | 直接查询 |
| `Treatment_Treatment` | ✅ 已对接 | Dashboard, Treatment | 直接查询 |
| `Auxiliary_EquipmentInfomation` | ✅ 已对接 | Dashboard, Device | 直接查询 |
| `Identity_Users` | ✅ 已对接 | Login, UserManagement | 直接查询 |
| `Organ_Employee` | ✅ 已对接 | Login, UserManagement | JOIN 查询 |
| `Identity_UserRoles` | ✅ 已对接 | Login, UserManagement | JOIN 查询 |
| `Identity_Roles` | ✅ 已对接 | Login, UserManagement | JOIN 查询 |
| `LIS_ExaminationItem` | ⚠️ 新表替代 | Statistics | 使用新表模型 |
| `Register_Infection` | ⚠️ 新表替代 | Statistics | 使用新表模型 |
| `Register_VascularAccess` | ⚠️ 新表替代 | Statistics | 使用新表模型 |
| `Stock_Stock` | ❌ 未对接 | Inventory | 使用新表 |
| `Stock_InOutStorage` | ❌ 未对接 | Inventory | 使用新表 |
| `CodeDictionary_CodeDictionarys` | ⚠️ fallback | MasterData | 缺表降级 |
| `Authorization_Roles` | ⚠️ 需确认 | RoleManagement | 表名可能不同 |
| `Authorization_Permissions` | ℹ️ 新功能 | RoleManagement | 新表 |
| `Authorization_RolePermissions` | ℹ️ 新功能 | RoleManagement | 新表 |

---

## 六、改造优先级建议

### 第一优先级（P0）
1. **Inventory 库存模块对接老库** - 对接 `Stock_Stock` + `Stock_InOutStorage`
2. **Statistics 统计模块对接老库** - 切换到老库 LIS/Register/Treatment 表
3. **用户密码哈希修复** - 使用 ASP.NET Identity V3 格式
4. **Dashboard 患者过滤条件** - 增加 `TreatmentStatus` 过滤

### 第二优先级（P1）
1. **WardOverview 真实数据** - 对接 `Schedule_Bed` 和 `Treatment_Treatment.RealDuration`
2. **Dashboard 告警增强** - 增加设备状态告警
3. **用户创建完善** - 增加 `Organ_Employee` 写入
4. **角色表名确认** - 统一 `Authorization_Roles` / `Identity_Roles`

### 第三优先级（待确认）
1. 字典表是否存在
2. 用户删除方式确认
3. 用户类型字段存储位置

---

## 七、代码定位索引

### 前端文件
- `ai-hms-frontend/src/pages/Dashboard.tsx` - 工作台页面
- `ai-hms-frontend/src/pages/WardOverview.tsx` - 病区概览
- `ai-hms-frontend/src/pages/Statistics.tsx` - 统计页面
- `ai-hms-frontend/src/pages/Inventory.tsx` - 库存管理
- `ai-hms-frontend/src/pages/UserManagement.tsx` - 用户管理
- `ai-hms-frontend/src/pages/Login.tsx` - 登录页面
- `ai-hms-frontend/src/pages/RoleManagement.tsx` - 角色管理
- `ai-hms-frontend/src/pages/MasterData.tsx` - 基础数据
- `ai-hms-frontend/src/pages/Settings.tsx` - 系统设置
- `ai-hms-frontend/src/services/restClient.ts` - REST API 客户端
- `ai-hms-frontend/src/services/userApi.ts` - 用户 API
- `ai-hms-frontend/src/services/roleManagementApi.ts` - 角色 API

### 后端文件
- `ai-hms-backend/internal/services/dashboard_service.go` - 看板统计服务
- `ai-hms-backend/internal/services/statistics_service.go` - 统计服务
- `ai-hms-backend/internal/services/inventory_service.go` - 库存服务
- `ai-hms-backend/internal/services/user_service.go` - 用户服务
- `ai-hms-backend/internal/services/auth_service.go` - 认证服务
- `ai-hms-backend/internal/services/permission_service.go` - 权限服务
- `ai-hms-backend/internal/services/device_service.go` - 设备服务
- `ai-hms-backend/internal/services/schedule_service.go` - 排班服务
- `ai-hms-backend/internal/services/treatment_service.go` - 治疗服务

### 参考文档
- `老血透数据库表结构-合并版.md` - 老库权威结构
- `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md` - 已有映射记录

---

*审计完成。共发现 5 个 P0 级问题、5 个 P1 级问题、5 个待人工确认项。*
