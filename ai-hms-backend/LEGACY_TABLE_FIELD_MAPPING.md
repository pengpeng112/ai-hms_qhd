# Legacy 表替换与字段更新记录

> 用途：记录本轮为兼容 legacy 数据库所做的“新表 -> legacy 表”替换及字段级映射，便于后续持续开发。

## 1. 背景约束

- 运行模式：legacy DB 模式
- `AutoMigrate` 已禁用，`internal/services/schema_helper.go` 为 no-op
- 需以真实 legacy 表为准，不依赖新 schema 自动建表

---

## 2. 总体替换清单（新表 -> legacy）

### 2.1 患者核心聚合（`internal/services/patient_core_service.go`）

| 原新表（逻辑） | 当前使用 legacy 表 | 备注 |
|---|---|---|
| `infection_infos`（感染信息逻辑） | `Register_Infection` | 通过 `legacyCoreInfection` 读取 |
| `treatment_plans` | `Plan_PatientPlan` | 通过 `legacyCorePlan` 读取 |
| `orders` | `Order_PatientOrder` | 通过 `legacyCoreOrder` 读取 |
| `lab_reports` + `lab_report_items` | `LIS_Examination` + `LIS_ExaminationItem` | 检验趋势/临床焦点通过 join 查询 |
| 导航依赖 `bed_number` 排序 | 改为 `Register_PatientInfomation` 对应患者主键排序（`Id`） | 因 legacy 主表无 `bed_number` |

### 2.2 设备模块（`internal/services/device_service.go`）

| 原新表（逻辑） | 当前使用 legacy 表 | 备注 |
|---|---|---|
| `devices`（列表/详情） | `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` + `Schedule_Bed` + `Schedule_Ward` | 设备、床位、病区需多表关联 |
| `devices`（写入：增删改） | 同上 | Create/Update/Delete/UpdateStatus 均已切到 legacy |

### 2.3 字典模块（`internal/services/dict_service.go`）

| 原新表（逻辑） | 当前策略 | 备注 |
|---|---|---|
| `dict_types` | 缺表时 fallback 到内置常量 | 非 legacy 表替换，而是运行时降级 |
| `dict_items` | 缺表时 fallback 到内置常量 | 覆盖透析方式/医保类型/患者类型 |

---

## 3. 字段级映射明细

## 3.1 patient_core_service.go

### 3.1.1 感染信息（`buildInfection`）

- 来源表：`Register_Infection`
- 使用字段：
  - `PatientId`
  - `InfectionDesc`
  - `OtherDesc`
  - `Note`
  - `LastModifyTime`
- 逻辑：合并 `InfectionDesc/OtherDesc/Note` 文本并关键词识别映射到 `HbsAg/HcvAb/HivAb/TpAb`

### 3.1.2 当前方案（`buildCurrentPlan`）

- 来源表：`Plan_PatientPlan`
- 使用字段：
  - `PatientId`
  - `LastModifyTime` / `CreateTime`（排序）
  - `IsDisabled`
  - `OddWeekFrequency` / `EvenWeekFrequency`
  - `DialysisMethod`
  - `DialysisDuration`
  - `DryWeight`
  - `BF`
  - `FirstAnticoagulant` / `MaintainAnticoagulant`
  - `Note`

### 3.1.3 活跃医嘱（`buildActiveOrders`）

- 来源表：`Order_PatientOrder`
- 过滤条件字段：
  - `PatientId`
  - `TenantId`
  - `IsDisabled`
- 返回映射字段：
  - `Id` -> `PatientCoreOrder.ID`
  - `Content` -> `PatientCoreOrder.Content`
  - `Classification`/`Type` -> `PatientCoreOrder.Type`
  - `StartTime` -> `PatientCoreOrder.StartTime`
  - `OperatorId` -> `PatientCoreOrder.Doctor`（经 `legacyOperatorName`）

### 3.1.4 检验趋势（`buildLabTrends`）

- 来源表：`LIS_ExaminationItem`（i）+ `LIS_Examination`（e）
- 关联：`e.Id = i.ExaminationId`
- 条件字段：
  - `e.PatientId`
  - `e.TenantId`
  - `e.ResultTime`
  - `i.ItemCode` / `i.ItemName`
- 映射字段：
  - `i.ItemCode` -> `item_code`
  - `i.ItemName` -> `item_name`
  - `i.Result` -> `result_value`
  - `i.Unit` -> `unit`
  - `i.Reference` -> `reference_range`
  - `i.ResultSign` -> `result_sign`
  - `COALESCE(e.ResultTime, i.LastModifyTime)` -> `tested_at`

### 3.1.5 临床焦点（`buildClinicalFocus`）

- 来源表：同上（`LIS_ExaminationItem` + `LIS_Examination`）
- 异常判断来源：`i.ResultSign`（不再使用不存在的 `AbnormalFlag`）

### 3.1.6 导航（`buildNavigation`）

- 读取模型：`models.Patient`（legacy 患者主表）
- 排序字段由旧逻辑 `bed_number` 改为 `"Id" ASC`

---

## 3.2 device_service.go

### 3.2.1 列表/详情（`List` / `getLegacyDeviceByID`）

- 主表：`Auxiliary_EquipmentInfomation`（e）
- 关联表：
  - `Schedule_BedEquipmentRel`（rel）
  - `Schedule_Bed`（bed）
  - `Schedule_Ward`（ward）
- 关联条件核心字段：
  - `rel.EquipmentId = e.Id`
  - `rel.IsDefault = true`
  - `rel.IsDisabled = false`
  - `bed.Id = rel.BedId`
  - `ward.Id = bed.WardId`

#### 输出字段映射

- `e.Id` -> `Device.ID`（cast 为 text）
- `e.TenantId` -> `Device.TenantId`
- `e.Name` -> `Device.Name`
- `e.SerialNo` -> `Device.SerialNo`
- `e.ModelNo` -> `Device.Model`
- `e.Brand` -> `Device.Manufacturer`
- `bed.Name` -> `Device.BedNumber`
- `ward.Id` -> `Device.WardId`
- `COALESCE(NULLIF(rel.ParameterS, ''), 'normal')` -> `Device.Status`
- `e.ManufactureDate` -> `Device.PurchaseDate`
- `COALESCE(ward.Name, '')` -> `Device.Notes`（当前仅展示映射）
- `COALESCE(e.Type,'') <> '__DELETED__'` 作为软删除过滤

### 3.2.2 创建设备（`Create`）

- 写入 `Auxiliary_EquipmentInfomation` 字段：
  - `Id`（由 `nextLegacyNumericID` 生成）
  - `TenantId`
  - `Name`
  - `SerialNo`
  - `Brand`
  - `ModelNo`
  - `DialysisMethod`（当前写空字符串）
  - `Type`（当前写空字符串）
  - `ManufactureDate`（当前写入 `time.Now()`）
- 如可解析床位，写入/更新 `Schedule_BedEquipmentRel`：
  - `Id`（自增策略）
  - `TenantId`
  - `EquipmentId`
  - `BedId`
  - `Sort`（当前固定 10）
  - `IsDefault=true`
  - `IsDisabled=false`
  - `LastModifyTime=now`
  - `Type=1`
  - `ParameterS=status`

### 3.2.3 更新设备（`Update`）

- 更新 `Auxiliary_EquipmentInfomation` 字段：
  - `Name`
  - `SerialNo`
  - `ModelNo`
  - `Brand`
- 床位/状态更新逻辑：
  - 根据 `BedNumber/WardId` 解析 `Schedule_Bed.Id`
  - 有床位：`upsertLegacyBedRelation(...)`
  - 无床位：`disableLegacyBedRelation(...)`
  - `Status` 通过 `ParameterS` 持久化

### 3.2.4 删除设备（`Delete`）

- 非物理删除：
  - `Auxiliary_EquipmentInfomation.Type = '__DELETED__'`
- 同时禁用关系：
  - `Schedule_BedEquipmentRel.IsDisabled = true`
  - `Schedule_BedEquipmentRel.LastModifyTime = now`

### 3.2.5 更新状态（`UpdateStatus`）

- 状态白名单：
  - `normal`
  - `warning`
  - `alarm`
  - `offline`
  - `maintenance`
- 持久化字段：
  - `Schedule_BedEquipmentRel.ParameterS = status`
  - `Schedule_BedEquipmentRel.LastModifyTime = now`
  - `Schedule_BedEquipmentRel.IsDisabled = false`

---

## 3.3 dict_service.go（缺表 fallback 记录）

### 3.3.1 缺表判断

- 函数：`isMissingRelationError(err)`
- 判定条件：错误信息包含 `relation` 且包含 `does not exist`

### 3.3.2 触发 fallback 的函数

- `ListTypes`
- `GetTypeByCode`
- `GetItemsByTypeCode`
- `GetItemsByTypeCodeTree`

### 3.3.3 fallback 覆盖项

- `DIALYSIS_MODE`
  - `HD` / `HDF` / `HP` / `HD+HP`
- `INSURANCE_TYPE`
  - `市职工普通` / `异地居民医保` / `城乡居民医保` / `自费`
- `PATIENT_TYPE`
  - `10=门诊` / `20=住院`

---

## 4. 已知限制（后续开发注意）

1. `Device.Notes`
   - 当前在 legacy 设备链路没有明确落库字段，列表/详情仅映射 `ward.Name`。
2. `Device.Status`
   - 当前借用 `Schedule_BedEquipmentRel.ParameterS` 持久化，属于兼容实现而非 legacy 原生状态字段。
3. ID 生成策略
   - `nextLegacyNumericID` 使用 `MAX("Id") + 1`，并发高峰下需评估冲突风险（当前依赖事务与业务并发场景较低）。

---

## 5. 代码定位（便于回查）

- `internal/services/patient_core_service.go`
  - `buildInfection`
  - `buildCurrentPlan`
  - `buildActiveOrders`
  - `buildLabTrends`
  - `buildClinicalFocus`
  - `buildNavigation`

- `internal/services/device_service.go`
  - `List`
  - `Create`
  - `Update`
  - `Delete`
  - `UpdateStatus`
  - `getLegacyDeviceByID`
  - `upsertLegacyBedRelation`
  - `disableLegacyBedRelation`

- `internal/services/dict_service.go`
  - `ListTypes`
  - `GetTypeByCode`
  - `GetItemsByTypeCode`
  - `GetItemsByTypeCodeTree`
  - `isMissingRelationError`
  - `legacyFallbackDictTypes`
  - `legacyFallbackDictItems`

---

## 6. Phase 0 追加记录

### 6.1 Task 0.3 枚举映射集中模块初始化

- 新增文件：`internal/services/legacy_enum_maps.go`
- 当前已落地映射：
  - `PatientType`：
    - 新 → 老：`门诊 -> 10`、`住院 -> 20`
    - 老 → 新：`10 -> 门诊`、`20 -> 住院`
- 导出函数：
  - `MapPatientTypeNewToLegacy(v string) string`
  - `MapPatientTypeLegacyToNew(v string) string`
- 约定：未命中映射时返回原值，避免服务层硬失败。
- 后续待补：`DialysisMode`、`OrderStatus`、`OrderType` 等枚举映射。

### 6.2 Task 1.1 `hospitalizations` → `Register_Hospitalization`

- 修改文件：
  - `internal/models/hospitalization.go`
  - `internal/services/hospitalization_service.go`
  - `internal/api/v1/hospitalization_handler.go`
- 变更要点：
  - `Hospitalization.TableName()` 切换为 `Register_Hospitalization`。
  - 模型字段补充 legacy 列映射：`Id/TenantId/PatientId/CaseNo/HospNo/BarCode/HospPatientType/HospReceiveDept/HospWard/HospBed/AttendDr/ReceptionDr/CreatorId/CreateTime/LastModifyTime`。
  - 新表语义字段 `status/admission_date/discharge_date/notes` 改为非持久化字段（`gorm:"-"`），在 service 层做兼容派生。
  - service 查询条件由 snake_case 改为 legacy PascalCase 列；补充 `TenantId` 过滤；列表排序改为 `CreateTime DESC`。
  - `GetByPatientId` 按 `CreateTime DESC` 取最近住院记录。
  - handler 将认证上下文 `tenantId` 传入服务层，保证按租户隔离。
- 兼容策略说明：
  - `Register_Hospitalization` 表结构中无 `Status/AdmissionDate/DischargeDate/Note` 对应持久列；
    - `status` 默认为在院（1）
    - `admissionDate` 派生为 `CreateTime`
    - `dischargeDate/notes` 仅接受请求回显，不入库。
  - 删除操作因目标表无 `IsDisabled` 字段，沿用物理删除，但增加 `TenantId` 过滤。

### 6.3 Task 1.2 `wards/beds/shifts/patient_shifts` → `Schedule_*`

- 修改文件：
  - `internal/models/schedule.go`
  - `internal/services/shift_service.go`
  - `internal/services/patient_shift_service.go`
  - `internal/api/v1/schedule_handler.go`
  - `internal/services/legacy_enum_maps.go`
- 变更要点：
  - `TableName()` 切换：
    - `Ward` -> `Schedule_Ward`
    - `Bed` -> `Schedule_Bed`
    - `Shift` -> `Schedule_Shift`
    - `PatientShift` -> `Schedule_PatientShift`
  - 模型字段补齐 legacy column tag，并将 `ScheduleDate` 持久化映射到 `TreatmentTime`。
  - service 查询/排序/更新条件由 snake_case 改为 legacy PascalCase 列名，统一补充 `TenantId` 过滤。
  - `ShiftService.Delete` 改为软删除（`IsDisabled=true`）。
  - `PatientShiftService.Delete` 改为状态取消（写入 legacy `Status=50`），保持按租户隔离。
  - `schedule_handler` 全链路透传 `tenantId` 到 service（List/Get/Update/Delete/冲突检查/按日查询）。
- 枚举映射追加：
  - 新增 `PatientShiftStatus` 双向映射：`0/1/2/3/4 <-> 10/20/30/40/50`
  - 导出函数：
    - `MapPatientShiftStatusNewToLegacy(v int) int`
    - `MapPatientShiftStatusLegacyToNew(v int) int`
- 兼容策略说明：
  - `Ward.Department/Floor`、`Bed.BedType/Status` 在目标老表无直接持久列，暂以 `gorm:"-"` 兼容保留。
  - `PatientShift.Notes/IsDisabled` 在目标老表无直接持久列，暂以 `gorm:"-"` 兼容保留。
