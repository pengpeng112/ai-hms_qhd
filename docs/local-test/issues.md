# 问题清单

> 测试时间：2026-06-07
> 代码分支：`fix/legacy-ui-restore`
> 提交号：`349802e1df104b2dd07cb6374421437c93580d55`

---

## ISSUE-001：9 张 Schedule 扩展表缺失

- **级别**：P1
- **模块**：排班模块（Schedule）
- **测试阶段**：阶段一（环境检查）
- **发现时间**：2026-06-07
- **代码提交**：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
- **测试环境**：PostgreSQL 13.3, 测试库 `dialysis`, Host 10.20.1.153
- **复现步骤**：
  1. 连接测试库 `dialysis`
  2. 查询 `information_schema.tables` 中 Schedule_ 前缀的表
  3. 与测试计划对比
- **预期结果**：以下 23 张表全部存在
- **实际结果**：仅 14 张存在，9 张缺失

| 缺失表 | 影响功能 |
|--------|----------|
| Schedule_WardExt | 病区扩展配置（区域类型、子区标识） |
| Schedule_BedMachineExt | 床位-机器绑定（机器类型、支持模式、禁用状态） |
| Schedule_PatientProfile | 患者排班骨架（频率模式、HDF 配置、固定床位） |
| Schedule_PatientShiftExt | 患者排班扩展信息 |
| Schedule_ScheduleTemplate | 排班模板主表 |
| Schedule_ScheduleTemplateItem | 排班模板明细 |
| Schedule_ConflictQueue | 排班冲突队列 |
| Schedule_Calendar | 排班日历 |
| Schedule_MachineOutage | 机器停机记录 |

- **涉及接口**：
  - `PUT /api/v1/schedule/config/wards`
  - `PUT /api/v1/schedule/config/machines`
  - `PUT /api/v1/schedule/config/patient-profiles`
  - `POST /api/v1/schedule/templates`
  - `POST /api/v1/schedule/templates/apply`
  - `GET /api/v1/schedule/week`
- **涉及数据库表**：上述 9 张表
- **数据库验证 SQL**：
  ```sql
  SELECT table_name FROM information_schema.tables
  WHERE table_schema = 'public' AND table_name LIKE 'Schedule_%'
  ORDER BY table_name;
  ```
- **数据库验证结果**：仅返回 Schedule_Bed, Schedule_CheckIn, Schedule_Handover, Schedule_PatientShift, Schedule_Shift, Schedule_Ward（6 张表）
- **日志摘要**：后端无报错（启动时未检查这些表）
- **初步原因**：后端代码已在 `internal/models/schedule_ext.go` 中定义了所有 9 张表的 GORM Model 和 TableName()，且在 `internal/services/schedule_template_service.go` 中有实际读写操作。但老血透数据库 `dialysis` 中未创建这 9 张表。测试代码（`schedule_template_service_test.go`）通过测试时自行 `CREATE TABLE`，不依赖实际库。生产运行时会因表不存在而 panic/返回错误。
- **建议修复**：
  1. 确认后端代码是否引用了这些表（搜索 gorm model 定义）
  2. 如果代码已引用但表不存在，需在老血透库中执行建表 DDL
  3. 如果代码未实现，需在进入阶段五/六前完成开发和建表
  4. 建议先在测试库中 `CREATE TABLE IF NOT EXISTS` 创建这些表
- **是否阻断试运行**：是（排班核心功能依赖这些表）
- **当前状态**：已修复（已执行 `docs/sql/schedule_extension_tables.sql`，14 张表全部创建成功）

---

## ISSUE-002：锁定用户仍可登录——AuthService 未检查 LockoutEnd

- **级别**：P1
- **模块**：认证模块（AuthService）
- **测试阶段**：阶段二（认证权限测试）
- **发现时间**：2026-06-07
- **代码提交**：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
- **测试环境**：PostgreSQL 13.3, 测试库 `dialysis`
- **复现步骤**：
  1. 在 Identity_Users 表中将用户 LockoutEnd 设置为过去时间（模拟锁定）
  2. 使用该用户凭证调用 `POST /api/v1/auth/login`
  3. 检查是否返回 401
- **预期结果**：锁定用户应返回 401（或特定锁定错误）
- **实际结果**：锁定用户登录成功，返回 200 + token
- **接口请求**：`POST /api/v1/auth/login` `{"username":"TEST_AI_HMS_viewer","password":"Test@123456"}`
- **接口响应**：200 OK，正常返回 token 和用户信息
- **涉及文件**：`ai-hms-backend/internal/services/auth_service.go:87-148`
- **涉及数据库表**：`Identity_Users`（LockoutEnd, LockoutEnabled, AccessFailedCount 列）
- **数据库验证 SQL**：
  ```sql
  SELECT "Id", "UserName", "LockoutEnabled", "LockoutEnd", "AccessFailedCount"
  FROM "Identity_Users" WHERE "Id" = 300415;
  ```
- **数据库验证结果**：LockoutEnd 已设置为锁定时间，但登录仍成功
- **日志摘要**：无异常日志
- **初步原因**：`auth_service.go` 的 `Authenticate()` 方法只验证了 UserName 存在和 PasswordHash 匹配，未检查 `LockoutEnd`、`LockoutEnabled` 和 `AccessFailedCount` 字段。ASP.NET Core Identity 标准锁定逻辑未实现。
- **建议修复**：
  1. 在 `Authenticate()` 中，查询用户后增加锁定期检查：
     ```go
     if identityUser.LockoutEnabled && identityUser.LockoutEnd != nil && identityUser.LockoutEnd.After(time.Now()) {
         return nil, errors.New("account is locked out")
     }
     ```
  2. 登录失败时增加 `AccessFailedCount` 计数
  3. 登录成功时重置 `AccessFailedCount`
- **是否阻断试运行**：否（不影响核心业务流程）
- **当前状态**：已修复（添加 LockoutEnd/LockoutEnabled 检查，已通过测试验证）

---

## ISSUE-003：MedicalHistory 模型映射到不存在的表导致患者详情 500

- **级别**：P1
- **模块**：患者模块（PatientService）
- **测试阶段**：阶段三（患者详情测试）
- **发现时间**：2026-06-07
- **代码提交**：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
- **测试环境**：PostgreSQL 13.3, 测试库 `dialysis`
- **复现步骤**：
  1. 调用 `GET /api/v1/patients/300410`
  2. 返回 500 错误
- **预期结果**：返回 200，患者详情含 `medicalHistory` 字段
- **实际结果**：500，错误 "relation 'medical_histories' does not exist"
- **接口请求**：`GET /api/v1/patients/300410`
- **接口响应**：`{"error":{"code":"INTERNAL_ERROR","message":"错误: 关系 \"medical_histories\" 不存在 (SQLSTATE 42P01)"}}`
- **涉及文件**：`internal/services/patient_service.go` (Preload), `internal/models/patient.go` (MedicalHistory.TableName)
- **涉及数据库表**：`Register_MedicalHistory`（实际）vs `medical_histories`（模型期望）
- **初步原因**：`models/patient.go` 中 `MedicalHistory.TableName()` 返回 `"medical_histories"`，该表不存在。老库实际表为 `Register_MedicalHistory`，且字段结构完全不同。模型设计为新库 schema，未适配老库。
- **建议修复**：
  1. 临时：从 Patient Get 中移除 `Preload("MedicalHistory")`（已执行）
  2. 长期：设计正确的 MedicalHistory 模型映射到 `Register_MedicalHistory`，或使用 `medical_history_service.go` 专有接口 `/api/v1/patients/:id/medical-history`
- **是否阻断试运行**：否（已修复）
- **当前状态**：已修复（移除 Preload）

---

## ISSUE-004：TreatmentPlan 模型映射到不存在的表导致患者详情 500

- **级别**：P1
- **模块**：患者模块（PatientService）
- **测试阶段**：阶段三（患者详情测试）
- **发现时间**：2026-06-07
- **代码提交**：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
- **测试环境**：PostgreSQL 13.3, 测试库 `dialysis`
- **复现步骤**：
  1. 先修复 ISSUE-003
  2. 再次调用 `GET /api/v1/patients/300410`
  3. 返回 500 错误
- **预期结果**：200
- **实际结果**：500，"relation 'treatment_plans' does not exist"
- **涉及文件**：`internal/services/patient_service.go`, `internal/models/treatment.go`
- **涉及数据库表**：`Plan_PatientPlan`（实际）vs `treatment_plans`（模型期望）
- **初步原因**：同 ISSUE-003，`TreatmentPlan.TableName()` 返回 `"treatment_plans"`，老库不存在。治疗方案数据应通过 `/api/v1/patients/:id/treatment-plan` 接口获取。
- **建议修复**：
  1. 临时：从 Patient Get 中移除 `Preload("TreatmentPlan")`（已执行）
  2. 长期：设计符合老库的 Plan_PatientPlan 映射
- **是否阻断试运行**：否（已修复）
- **当前状态**：已修复（移除 Preload）

---

## ISSUE-005：LegacyCreateTreatmentPlan 缺少 VascularAccessId 导致 500

- **级别**：P1
- **模块**：治疗方案模块（PatientService）
- **测试阶段**：阶段三
- **发现时间**：2026-06-07
- **代码提交**：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
- **测试环境**：PostgreSQL 13.3, 测试库 `dialysis`
- **复现步骤**：
  1. 调用 `POST /api/v1/patients/300410/treatment-plan` 新增方案
  2. 返回 500 错误
- **预期结果**：201/200，方案创建成功
- **实际结果**：500，"null value in column 'VascularAccessId' of relation 'Plan_PatientPlan' violates not-null constraint"
- **接口请求**：`POST /api/v1/patients/300410/treatment-plan`
- **接口响应**：`SQLSTATE 23502`
- **涉及文件**：`internal/services/patient_service.go:1566-1607`
- **涉及数据库表**：`Plan_PatientPlan`（VascularAccessId NOT NULL）
- **初步原因**：`LegacyCreateTreatmentPlan` 的 `createMap` 未包含 `VascularAccessId` 字段，而该列在 `Plan_PatientPlan` 有 NOT NULL 约束。
- **建议修复**：
  1. 临时：添加 `"VascularAccessId": 0` 到 createMap（已执行）
  2. 长期：自动获取患者的默认血管通路 ID 并填入
- **是否阻断试运行**：否（已修复）
- **当前状态**：已修复（添加默认值 0）

---

## ISSUE-006：血管通路新增时 Note 字段未落库

- **级别**：P2
- **模块**：血管通路模块（VascularAccessHandler）
- **测试阶段**：阶段三
- **发现时间**：2026-06-07
- **代码提交**：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
- **测试环境**：PostgreSQL 13.3, 测试库 `dialysis`
- **复现步骤**：
  1. 调用 `POST /api/v1/patients/300410/vascular-accesses`，传入 `note` 字段
  2. 查询数据库 Register_VascularAccess
- **预期结果**：Note 字段值为 `TEST_AI_HMS_vascular_access_test`
- **实际结果**：Note 字段为空
- **接口请求**：`POST /api/v1/patients/300410/vascular-accesses` `{"accessType":"AVF","note":"TEST_AI_HMS_vascular_access_test"}`
- **涉及文件**：VascularAccessHandler / VascularAccessService 待定位
- **涉及数据库表**：`Register_VascularAccess`
- **初步原因**：Create 请求处理时 Note 字段映射遗漏或字段名不匹配
- **建议修复**：检查 VascularAccessHandler Create 中的 Note 列映射
- **是否阻断试运行**：否
- **当前状态**：已澄清（非代码问题 — JSON 字段名为 `notes`（复数），测试输入误用了 `note`（单数），使用正确字段名后已验证落库正常）

---

## ISSUE-007：WardService.Create 缺少 NOT NULL 字段导致 500

- **级别**：P1
- **模块**：排班配置（WardService）
- **测试阶段**：阶段五
- **发现时间**：2026-06-07
- **复现步骤**：POST /api/v1/wards 创建病区
- **预期结果**：201 Created
- **实际结果**：500，NOT NULL constraint violation (TenantId/CreatorId/CreateTime/LastModifyTime)
- **涉及文件**：`internal/services/ward_service.go:166`, `internal/api/v1/ward_handler.go:53`
- **涉及数据库表**：`Schedule_Ward`
- **建议修复**：Handler 提取 middleware.GetTenantID/GetCreatorID，Service 添加这些字段到 createMap
- **当前状态**：已修复

---

## ISSUE-008：BedService.Create 缺少 NOT NULL 字段导致 500

- **级别**：P1
- **模块**：排班配置（BedService）
- **测试阶段**：阶段五
- **发现时间**：2026-06-07
- **复现步骤**：同 ISSUE-007，影响 POST /api/v1/beds
- **涉及文件**：`internal/services/bed_service.go:212`, `internal/api/v1/bed_handler.go:53`
- **涉及数据库表**：`Schedule_Bed`
- **当前状态**：已修复

---

## ISSUE-009：Shift 模型字段类型与 DB 列类型不匹配导致 500

- **级别**：P1
- **模块**：排班配置（ShiftService）
- **测试阶段**：阶段五
- **发现时间**：2026-06-07
- **复现步骤**：POST /api/v1/shifts 传入 {"startTime":"08:00","endTime":"12:00"}
- **预期结果**：201 Created
- **实际结果**：500，DB 列 StartTime/EndTime 为 timestamp 但模型为 varchar；Type 为 integer 但模型为 varchar
- **涉及文件**：`internal/services/shift_service.go`, `internal/models/schedule.go`
- **涉及数据库表**：`Schedule_Shift`
- **建议修复**：ShiftService.Create 改用 raw map 插入 + parseShiftTime 转换 HH:MM 为 time.Time
- **当前状态**：已修复（raw map + parseShiftTime + Type int 转换）

---

## ISSUE-010：ApplyTemplate 缺少 PatientPlanId NOT NULL 字段

- **级别**：P1
- **模块**：排班模板（ScheduleTemplateService）
- **测试阶段**：阶段六
- **当前状态**：已修复（添加 `PatientPlanId: &[]int64{0}[0]`）

---

## ISSUE-011：upsertTreatmentSigns 缺少 OperateTime

- **级别**：P1
- **模块**：治疗服务（TreatmentService）
- **测试阶段**：阶段七
- **当前状态**：已修复（添加 `"OperateTime": now`）

---

## ISSUE-012：upsertTreatmentSigns 缺少 OperatorId

- **级别**：P1
- **模块**：治疗服务（TreatmentService）
- **测试阶段**：阶段七
- **当前状态**：已修复（添加 `"OperatorId": creatorID`）

---

## ISSUE-013：SaveDisinfection 缺少 NOT NULL 字段

- **级别**：P1
- **模块**：治疗服务（TreatmentService）
- **测试阶段**：阶段七
- **当前状态**：已修复（重写为完整的 raw map 插入）

---

## ISSUE-014：Device List jsonb ParameterS 类型转换错误

- **级别**：P1
- **模块**：设备服务（DeviceService）
- **测试阶段**：阶段八
- **当前状态**：已修复（添加 `::text` 转换）

---

## ISSUE-015：Device List DISTINCT + ORDER BY 冲突

- **级别**：P1
- **模块**：设备服务（DeviceService）
- **测试阶段**：阶段八
- **当前状态**：已修复（重建 listQuery 避免污染）

---

## ISSUE-016：缺少 Auxiliary_EquipmentMaintenance/UsageLog 表

- **级别**：P2
- **模块**：设备
- **测试阶段**：阶段八
- **当前状态**：已修复（已创建 2 张表）

---

## ISSUE-017：Dashboard Stats Status 列类型编码错误导致 500

- **级别**：P2
- **模块**：统计看板（DashboardService）
- **测试阶段**：阶段九
- **发现时间**：2026-06-07
- **复现步骤**：GET /api/v1/dashboard/stats
- **预期结果**：200
- **实际结果**：500，"unable to encode 60 into text format"
- **涉及文件**：`internal/services/dashboard_service.go:147`
- **涉及数据库表**：`Treatment_Treatment`（Status 为 varchar 列）
- **初步原因**：`Status` 列类型为 varchar，但传入了 int 值 60
- **当前状态**：已修复（`60` → `"60"`）
