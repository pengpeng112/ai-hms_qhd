# P1 配置和主数据页面 — 字段级比对核查报告

> 核查日期：2026-05-31
> 核查范围：7 个配置/主数据页面 + 对应后端服务
> 核查依据：`老血透数据库表结构-合并版.md`、`LEGACY_TABLE_FIELD_MAPPING.md`

---

## 一、总览

| 页面 | 路由 | 老库主表 | 后端服务 | 整体状态 |
|------|------|---------|---------|---------|
| 诊疗配置-方案模板 | `/treatment-config` | `Plan_PlanTPL` + `Plan_PlanTPLMaterial` | `treatment_config_service.go` | ✅ 基本对齐 |
| 诊疗配置-医嘱模板 | `/treatment-config` | `Order_OrderTPL` | `treatment_config_service.go` | ✅ 基本对齐 |
| 诊疗配置-材料目录 | `/treatment-config` | `Auxiliary_MaterialInfomation` | `treatment_config_service.go` | ✅ 基本对齐 |
| 诊疗配置-药品目录 | `/treatment-config` | `Auxiliary_DrugInfomation` | `treatment_config_service.go` | ✅ 基本对齐 |
| 病区管理 | `/ward-management` | `Schedule_Ward` | `ward_service.go` | ⚠️ 部分差异 |
| 床位管理 | `/bed-management` | `Schedule_Bed` + `Schedule_BedEquipmentRel` | `bed_service.go` | ⚠️ 部分差异 |
| 班次配置 | `/shift-config` | `Schedule_Shift` | `shift_service.go` | ⚠️ 部分差异 |
| 设备管理 | `/device-binding` | `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` | `device_service.go` | ⚠️ 部分差异 |
| 字典配置 | `/dict-config` | `CodeDictionary_CodeDictionarys` | `dict_service.go` | ⚠️ 写操作需确认 |
| 宣教管理 | `/education-management` | `Auxiliary_HealthEducation` | `health_education_service.go` | ✅ 基本对齐 |

---

## 二、逐页详细核查

### 2.1 诊疗配置 — 方案模板 (PlanTab)

**前端文件**: `src/pages/TreatmentConfig/tabs/PlanTab.tsx`
**前端 API**: `src/services/treatmentConfigApi.ts:426-461` → `planTemplateApi`
**后端**: `internal/services/treatment_config_service.go:22-826` (`PlanTemplateService`)
**后端 Handler**: `internal/api/v1/treatment_config_handler.go`

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 1.1 | P0 | 列表查询 | search/mode/isEnabled | GET `/api/v1/treatment-templates` | `Plan_PlanTPL.Name/DialysisMethod/IsDisabled` | `Plan_PlanTPL` 35字段 | LegacyList 正确映射 | ✅ | — |
| 1.2 | P0 | 创建模板 | name/mode/templateContent | POST `/api/v1/treatment-templates` | `Plan_PlanTPL` 全字段 + `Plan_PlanTPLMaterial` | 老库字段一一对应 | LegacyCreate 正确写入 | ✅ | — |
| 1.3 | P0 | 模板材料 | templateContent.materials[] | 同上 | `Plan_PlanTPLMaterial.PlanTPLId/MaterialId/MaterialGroup/Num/Note` | 同老库 | syncLegacyPlanTemplateMaterials 正确 | ✅ | — |
| 1.4 | P1 | 前端字段差异 | `PlanTemplate.mode` 类型 `'HD'\|'HFD'\|'HP'\|'HF'\|'HDF'` | — | 老库 `DialysisMethod` 支持 `HD+HP` | `character varying(256)` | 前端 mode 枚举缺 `HD+HP` | ⚠️ | 前端 mode 类型需补 `HD+HP` |
| 1.5 | P1 | 前端嵌套结构 | `templateContent` JSON 嵌套 | — | 老库为扁平列 | 35个扁平列 | 后端做展平映射，正确 | ✅ | — |
| 1.6 | P2 | 前端缺字段 | 无 `DryWeight` 独立输入 | — | `Plan_PlanTPL` 无 DryWeight 列 | 老库 `Plan_PlanTPL` 确实无此列 | 后端 dto 写死 0 | ✅ | 正确，模板层不需要 |
| 1.7 | P2 | 前端字段映射 | `anticoagulant.initialDrug` (名称) | — | `FirstAnticoagulant` (bigint 药品ID) | `bigint` | 后端通过 `findLegacyDrugIDByName` 转换 | ✅ | — |
| 1.8 | P2 | 软删除 | 删除操作 | DELETE | `IsDisabled=true` | `boolean` | LegacyDelete 正确 | ✅ | — |

### 2.2 诊疗配置 — 医嘱模板 (OrderTab)

**前端文件**: `src/pages/TreatmentConfig/tabs/OrderTab.tsx`
**前端 API**: `src/services/treatmentConfigApi.ts:540-575` → `orderTemplateApi`
**后端**: `internal/services/treatment_config_service.go` (`OrderTemplateService`)

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 2.1 | P0 | 列表查询 | search/type/category/isEnabled | GET `/api/v1/order-templates` | `Order_OrderTPL` | 18字段 | LegacyList 正确 | ✅ | — |
| 2.2 | P0 | 创建/更新 | name/type/category/content/items[] | POST/PUT | `Order_OrderTPL` | 同老库 | LegacyCreate/LegacyUpdate 正确 | ✅ | — |
| 2.3 | P1 | 前端 `OrderTemplate.type` | `'长期'\|'临时'` | — | 老库 `Order_PatientOrder.Type` (integer) | `integer` | 后端 `toOrderTemplateType` 从 Content/Note 文本推断 | ⚠️ | Type 推断逻辑脆弱，需人工确认 |
| 2.4 | P1 | 前端 `items[].route` | 用法/途径 | — | `Order_OrderTPL.UseWay` | `varchar(128)` | 映射正确 | ✅ | — |
| 2.5 | P1 | 前端 `items[].frequency` | 频次 | — | `Order_OrderTPL.UseMethod` | `varchar(128)` | 映射正确 | ✅ | — |
| 2.6 | P1 | 前端 `items[].timing` | 使用时机 | — | `Order_OrderTPL.UseOpportunity` | `varchar(128)` | 映射正确 | ✅ | — |
| 2.7 | P2 | 老库 `AllDosage` 字段 | 前端无 | — | `Order_OrderTPL.AllDosage` | `numeric` | 后端写入 0 | ⚠️ | 需确认是否需要持久化 |

### 2.3 诊疗配置 — 材料目录 (MaterialTab)

**前端文件**: `src/pages/TreatmentConfig/tabs/MaterialTab.tsx`
**前端 API**: `src/services/treatmentConfigApi.ts:464-499` → `materialCatalogApi`
**后端**: `internal/services/treatment_config_service.go:900-1544` (`MaterialCatalogService`)

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 3.1 | P0 | CRUD | code/name/category/spec/brand/unit... | `/api/v1/materials/catalog` | `Auxiliary_MaterialInfomation` | 20字段 | Legacy CRUD 正确映射 | ✅ | — |
| 3.2 | P1 | 字段映射 | `shortName` | — | `ShortName` | `varchar(128)` | ✅ | ✅ | — |
| 3.3 | P1 | 字段映射 | `mnemonic` | — | `Spell` | `varchar(256)` | ✅ | ✅ | — |
| 3.4 | P1 | 字段映射 | `standardType` | — | `StdCat` + `Type` (双写) | `varchar(256)` + `varchar(64)` | LegacyUpdate 双写 `StdCat` 和 `Type` | ⚠️ | 老库 `Type` 含义为材料分类，非标准分类，双写可能语义冲突 |
| 3.5 | P1 | 字段映射 | `packaging` | — | `Package` | `varchar(64)` | ✅ | ✅ | — |
| 3.6 | P2 | 前端缺字段 | 无 `UseTips` | — | `Auxiliary_MaterialInfomation.UseTips` | `varchar(1024)` | 后端未读写 | ⚠️ | 老库有此字段但当前未暴露 |
| 3.7 | P2 | 前端缺字段 | 无 `MinUnitDosage` | — | `Auxiliary_DrugInfomation.MinUnitDosage` | `bigint` | 材料表无此字段，药品表有 | ✅ | 正确，材料不需要 |

### 2.4 诊疗配置 — 药品目录 (DrugTab)

**前端文件**: `src/pages/TreatmentConfig/tabs/DrugTab.tsx`
**前端 API**: `src/services/treatmentConfigApi.ts:502-537` → `drugCatalogApi`
**后端**: `internal/services/treatment_config_service.go` (`DrugCatalogService`)

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 4.1 | P0 | CRUD | code/name/category/spec/brand/unit... | `/api/v1/drugs/catalog` | `Auxiliary_DrugInfomation` | 23字段 | Legacy CRUD 正确映射 | ✅ | — |
| 4.2 | P1 | 前端字段 | `genericName` (通用名) | — | 老库无直接字段 | — | 前端新增字段，老库无对应 | ⚠️ | 需确认是否持久化到老库某列 |
| 4.3 | P1 | 前端字段 | `concentration` | — | 老库无直接字段 | — | 前端新增字段 | ⚠️ | 同上 |
| 4.4 | P1 | 前端字段 | `specUnit` | — | `SpecificationUnit` | `varchar(64)` | 药品表无此字段，材料表有 | ⚠️ | 需确认映射关系 |
| 4.5 | P1 | 前端字段 | `minUnitDose` | — | `MinUnitDosage` | `bigint` | 映射正确 | ✅ | — |
| 4.6 | P1 | 前端字段 | `baseUnit` | — | `BasicUnit` | `varchar(64)` | 映射正确 | ✅ | — |
| 4.7 | P1 | 前端字段 | `timing` | — | `UseOpportunity` | `varchar(64)` | 药品表有此字段 | ✅ | — |
| 4.8 | P2 | 前端字段 | `tips` | — | `UseTips` | `varchar(1024)` | 映射正确 | ✅ | — |

### 2.5 病区管理 (WardManagement)

**前端文件**: `src/pages/WardManagement.tsx`
**前端 API**: `src/services/managementApi.ts:118-136` → `wardManagementApi`
**后端**: `internal/services/ward_service.go`
**后端 Handler**: `internal/api/v1/ward_handler.go`

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 5.1 | P0 | 列表 | name/sort/patientType/infectionType/isDisabled/note/responsibleUsers | GET `/api/v1/wards` | `Schedule_Ward` 全字段 | 12字段 | ✅ 正确读取 `Schedule_Ward` | ✅ | — |
| 5.2 | P0 | 创建/更新 | 同上 | POST/PUT `/api/v1/wards` | `Schedule_Ward` | 同老库 | ✅ 正确写入 | ✅ | — |
| 5.3 | P1 | `ResponsibleUsers` 存储 | 前端多选→逗号拼接字符串 | — | `ResponsibleUsers varchar(512)` | `varchar(512)` | 逗号拼接存储 | ✅ | — |
| 5.4 | P1 | 前端 `patientType` 枚举 | `'普通患者'\|'隔离患者'` 硬编码 | — | 老库 `PatientType varchar(64)` 期望 `'10'\|'20'` | `varchar(64)` | ⚠️ 前端传中文，老库存代码 | ❌ | 前端应使用字典值 `10/20`，或后端做映射 |
| 5.5 | P1 | 前端 `infectionType` 枚举 | `'乙肝'\|'丙肝'\|'梅毒'\|'HIV'` 硬编码 | — | 老库 `InfectionType varchar(64)` 期望 `'普通'\|'乙肝'\|'丙肝'` | `varchar(64)` | ⚠️ 前端多了 `梅毒/HIV`，老库设计仅 `普通/乙肝/丙肝` | ❌ | 需确认老库实际数据，对齐枚举 |
| 5.6 | P1 | `bedCount` 字段 | 前端展示列 | — | 后端 `List` 未计算 bedCount | 老库 `Schedule_Ward` 无此列 | ⚠️ 后端 WardDTO 有 `BedCount` 但 List 查询未填充 | ❌ | 后端需 JOIN `Schedule_Bed` 统计 |
| 5.7 | P1 | 删除 | DELETE `/api/v1/wards/:id` | 物理删除 | `Schedule_Ward` | 老库有 `IsDisabled` | ⚠️ 物理删除，非软删除 | ⚠️ | 应改为 `IsDisabled=true` 软删除 |
| 5.8 | P2 | 前端 `patientTypeLabel` | 展示列 | — | 后端未填充 | — | 前端有此字段但后端未返回 | ⚠️ | 后端需从字典翻译 |

### 2.6 床位管理 (BedManagement)

**前端文件**: `src/pages/BedManagement.tsx`
**前端 API**: `src/services/managementApi.ts:138-156` → `bedManagementApi`
**后端**: `internal/services/bed_service.go`

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 6.1 | P0 | 列表 | name/wardId/wardName/sort/note/isDisabled/fepId/acquisiteConnectId | GET `/api/v1/beds` | `Schedule_Bed` + `Schedule_BedEquipmentRel` + `Auxiliary_EquipmentInfomation` | 12字段 + 关联表 | ✅ 正确 | ✅ | — |
| 6.2 | P0 | 设备关联 | equipments[] / defaultEquipmentName / equipmentCount | 同上 | `Schedule_BedEquipmentRel` | 10字段 | ✅ JOIN 查询正确 | ✅ | — |
| 6.3 | P0 | 创建/更新 | 同上 | POST/PUT `/api/v1/beds` | `Schedule_Bed` + `Schedule_BedEquipmentRel` | 同老库 | ✅ `syncBedEquipments` 正确 | ✅ | — |
| 6.4 | P1 | `FEPId` 字段 | 前端 InputNumber | — | `Schedule_Bed.FEPId` | `bigint` | ✅ 正确映射 | ✅ | — |
| 6.5 | P1 | `AcquisiteConnectId` 字段 | 前端 InputNumber | — | `Schedule_Bed.AcquisiteConnectId` | `bigint` | ✅ 正确映射 | ✅ | — |
| 6.6 | P1 | 删除 | DELETE | 物理删除 | `Schedule_Bed` + `Schedule_BedEquipmentRel` | 老库有 `IsDisabled` | ⚠️ 物理删除 | ⚠️ | 应改为软删除 |

### 2.7 班次配置 (ShiftConfig)

**前端文件**: `src/pages/ShiftConfig.tsx`
**前端 API**: `src/services/restClient.ts` → `restApi.getShifts/createShift/updateShift/deleteShift`
**后端**: `internal/services/shift_service.go`
**模型**: `internal/models/schedule.go:83-104` (`Shift`)

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 7.1 | P0 | 列表 | name/startTime/endTime/sort/isDisabled/notes | GET `/api/v1/shifts` | `Schedule_Shift` | 12字段 | ✅ 正确 | ✅ | — |
| 7.2 | P0 | 创建 | name/startTime/endTime/sort/notes | POST | `Schedule_Shift` | 同老库 | ✅ 正确 | ✅ | — |
| 7.3 | P0 | 更新 | 同上 + isDisabled | PUT | `Schedule_Shift` | 同老库 | ⚠️ `Notes→Note` 映射 | ⚠️ | 见下方说明 |
| 7.4 | P1 | 字段名差异 | 前端 `notes` | — | 后端 `shift.Notes` → DB `Note` | `Schedule_Shift.Note varchar(512)` | 模型 tag `gorm:"column:Note"`，JSON `notes` | ✅ | 正确，但 Update 时 `req.Notes→updates["Note"]` |
| 7.5 | P1 | `StartTime/EndTime` 类型 | 前端 `HH:mm` 字符串 | — | 模型 `string` tag `varchar(32)` | 老库 `timestamp` | ⚠️ 老库物理列为 `timestamp`，模型用 `string` 兼容 | ⚠️ | 需确认老库是否接受 `HH:mm` 字符串写入 timestamp 列 |
| 7.6 | P1 | `Type` 字段 | 前端未展示/编辑 | — | `Schedule_Shift.Type integer` | `integer` (10=长期/20=临时) | 模型 `string`，后端 `req.Type` 传空 | ⚠️ | 需确认 Type 是否需要默认值 |
| 7.7 | P1 | 删除 | — | 软删除 `IsDisabled=true` | `Schedule_Shift` | 老库有 `IsDisabled` | ✅ Delete 正确改为软删除 | ✅ | — |

### 2.8 设备管理 (DeviceManagement)

**前端文件**: `src/pages/DeviceManagement.tsx`
**前端 API**: `src/services/equipment.ts`
**后端**: `internal/services/device_service.go`

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 8.1 | P0 | 列表 | Name/Brand/ModelNo/SerialNo/Status/BedNumber/WardName | GET `/api/v1/devices` | `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` + `Schedule_Bed` + `Schedule_Ward` | 20字段 + 关联 | ✅ LEGACY_TABLE_FIELD_MAPPING 已记录 | ✅ | — |
| 8.2 | P1 | 前端字段 `IDNo` | 设备编号 | — | `Auxiliary_EquipmentInfomation.IDNo` | `varchar(128)` | ✅ 正确 | ✅ | — |
| 8.3 | P1 | 前端字段 `Manufacturer` | 生产厂家 | — | `Auxiliary_EquipmentInfomation.Manufacturer` | `varchar(128)` | ✅ 正确 | ✅ | — |
| 8.4 | P1 | 前端字段 `DialysisMethod` | 透析方式 | — | `Auxiliary_EquipmentInfomation.DialysisMethod` | `varchar(512)` | ✅ 正确 | ✅ | — |
| 8.5 | P1 | `Status` 存储 | 前端展示 normal/warning/error/offline | — | `Schedule_BedEquipmentRel.ParameterS` (jsonb) | `jsonb` | ⚠️ 借用 ParameterS 存状态字符串 | ⚠️ | 非老库原生字段，属于兼容实现 |
| 8.6 | P1 | `Notes` 存储 | 前端展示 | — | 映射 `ward.Name` → `Device.Notes` | 老库 `EquipmentInfomation.Note` | ⚠️ 当前 Notes 映射为病区名而非设备备注 | ⚠️ | 应映射到 `EquipmentInfomation.Note` |
| 8.7 | P1 | 前端额外字段 | DeviceType/InstallDate/MaintenanceCycle/Flux | — | `Auxiliary_EquipmentInfomation` 对应列 | 各列存在 | ✅ 正确读取 | ✅ | — |
| 8.8 | P2 | 设备详情弹窗 | 消毒记录/维护记录/使用记录 | GET `/api/v1/devices/:id/disinfections` 等 | `Auxiliary_EquipmentDisinfection` 等 | 对应老库表 | ✅ 正确 | ✅ | — |

### 2.9 字典配置 (DictConfig)

**前端文件**: `src/pages/DictConfig.tsx`
**前端 API**: `src/services/dictApi.ts`
**后端**: `internal/services/dict_service.go`

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 9.1 | P0 | 类型列表 | code/name/sortOrder/source | GET `/api/v1/dict/types` | `CodeDictionary_CodeDictionarys` 按 Type 聚合 | 7字段 (Code/Type/Name/OrganId/IsDisabled/Sort/Builtin) | ✅ listLegacyCodeDictionaryTypes 正确 | ✅ | — |
| 9.2 | P0 | 字典项列表 | code/name/description/sortOrder/isEnabled/parentCode | GET `/api/v1/dict/items/:typeCode` | `CodeDictionary_CodeDictionarys` | 同上 | ✅ listLegacyCodeDictionaryItems 正确 | ✅ | — |
| 9.3 | P0 | 写操作同步 | 新增/编辑/删除字典项 | POST/PUT/DELETE | 新表 `dict_items` + 老库 `CodeDictionary_CodeDictionarys` | 老库表 | ⚠️ `legacyDictUpsert` 同步写老库 | ⚠️ | 见下方详细说明 |
| 9.4 | P1 | 前端 `source` 标记 | `'legacy'\|'local'` | — | 后端返回 source 字段 | — | ✅ 前端正确识别老库来源并禁用编辑 | ✅ | — |
| 9.5 | P1 | 前端禁用老库编辑 | 老库字典项按钮 disabled | — | — | — | ✅ 正确 | ✅ | — |
| 9.6 | P1 | 新表字典 CRUD | 全字段 | `/api/v1/dict/items` | `dict_items` (新表) | 老库无对应表 | ⚠️ 新表有 `parent_code` 支持树形，老库无此概念 | ⚠️ | 树形结构仅在新表维护，老库同步时丢失层级 |
| 9.7 | P1 | `legacyDictUpsert` 逻辑 | — | — | 写 `CodeDictionary_CodeDictionarys` | 老库 | ⚠️ upsert 使用 `Type+Code` 匹配，正确 | ✅ | — |
| 9.8 | P2 | 前端 `parentCode` | 上级编码 | — | 新表 `parent_code` | 老库无此列 | ⚠️ 老库同步时 parentCode 丢失 | ⚠️ | 需确认是否需要老库回写 |

### 2.10 宣教管理 (EducationManagement)

**前端文件**: `src/pages/EducationManagement.tsx`
**前端 API**: `src/services/managementApi.ts:158-176` → `educationManagementApi`
**后端**: `internal/services/health_education_service.go`

| 编号 | 优先级 | 功能点 | 前端字段 | API路径 | 后端字段/表 | 老库标准字段 | 类型/内容核查 | 结论 | 建议 |
|------|--------|--------|---------|---------|------------|------------|-------------|------|------|
| 10.1 | P0 | 列表 | name/description/sort/type/classify/note/isDisabled | GET `/api/v1/health-educations` | `Auxiliary_HealthEducation` | 13字段 | ✅ ListContents 正确读取 | ✅ | — |
| 10.2 | P0 | 创建 | 同上 | POST | `Auxiliary_HealthEducation` | 同老库 | ✅ CreateContent 正确写入 | ✅ | — |
| 10.3 | P0 | 更新 | 同上 | PUT | `Auxiliary_HealthEducation` | 同老库 | ✅ UpdateContent 正确 | ✅ | — |
| 10.4 | P0 | 删除 | DELETE | 物理删除 | `Auxiliary_HealthEducation` | 老库有 `IsDisabled` | ⚠️ 物理删除 | ⚠️ | 应改为软删除 |
| 10.5 | P1 | `Sort` 类型 | 前端 `number` | — | 后端 `Sort string` | 老库 `numeric` | ⚠️ 后端 DTO Sort 为 string 类型 | ⚠️ | 应改为 int/numeric |
| 10.6 | P1 | 前端缺字段 | 无 `attachmentIds` | — | `Auxiliary_HealthEducation.AttachmentIds` | `varchar(64)` | 后端有读取但前端未展示 | ⚠️ | 前端应支持附件管理 |
| 10.7 | P2 | 前端缺功能 | 无 `IsDisabled` 过滤 | — | 后端查询未过滤 | — | 列表返回全部 | ✅ | 可接受 |

---

## 三、关键不一致/待确认项汇总

### ❌ 必须修复

| 编号 | 页面 | 问题 | 影响 | 建议 |
|------|------|------|------|------|
| F-1 | 病区管理 | `patientType` 前端传中文 `'普通患者'`/`'隔离患者'`，老库期望 `'10'`/`'20'` | 写入数据不一致 | 后端加映射或前端用字典值 |
| F-2 | 病区管理 | `infectionType` 前端枚举多了 `梅毒/HIV`，老库设计仅 `普通/乙肝/丙肝` | 数据不一致 | 确认老库实际支持范围 |
| F-3 | 病区管理 | `bedCount` 后端 WardDTO 有此字段但 List 查询未计算 | 前端显示 0 | 后端 JOIN 统计 |
| F-4 | 病区管理/床位管理/宣教管理 | 删除为物理删除，非软删除 | 数据丢失风险 | 改为 `IsDisabled=true` |

### ⚠️ 需人工确认

| 编号 | 页面 | 问题 | 影响 |
|------|------|------|------|
| C-1 | 诊疗配置-方案模板 | 前端 mode 类型缺 `HD+HP` | 无法创建 HD+HP 模板 |
| C-2 | 诊疗配置-医嘱模板 | OrderTPL.Type 推断逻辑从 Content/Note 文本匹配 `'临时'` | 推断不准确 |
| C-3 | 诊疗配置-材料目录 | `standardType` 双写 `StdCat` 和 `Type`，`Type` 在老库含义为材料分类 | 语义冲突 |
| C-4 | 班次配置 | `StartTime/EndTime` 模型用 string，老库物理列为 timestamp | 写入兼容性 |
| C-5 | 班次配置 | `Type` 字段前端未编辑，老库为 integer (10/20) | 需默认值 |
| C-6 | 设备管理 | `Status` 借用 `ParameterS` (jsonb) 存储 | 非原生字段 |
| C-7 | 设备管理 | `Notes` 映射为 `ward.Name` 而非设备备注 | 语义错误 |
| C-8 | 字典配置 | 新表 `parent_code` 树形结构老库无法表达 | 同步丢失层级 |
| C-9 | 药品目录 | `genericName/concentration/specUnit` 老库无直接对应列 | 需确认映射 |
| C-10 | 宣教管理 | 后端 `Sort` 为 string 类型，老库为 numeric | 类型不一致 |

---

## 四、各页面 API 路径总览

| 页面 | API路径 | HTTP方法 | 后端表 |
|------|---------|---------|--------|
| 方案模板 | `/api/v1/treatment-templates` | GET/POST/PUT/DELETE | `Plan_PlanTPL` + `Plan_PlanTPLMaterial` |
| 医嘱模板 | `/api/v1/order-templates` | GET/POST/PUT/DELETE | `Order_OrderTPL` |
| 材料目录 | `/api/v1/materials/catalog` | GET/POST/PUT/DELETE | `Auxiliary_MaterialInfomation` |
| 药品目录 | `/api/v1/drugs/catalog` | GET/POST/PUT/DELETE | `Auxiliary_DrugInfomation` |
| 病区管理 | `/api/v1/wards` | GET/POST/PUT/DELETE | `Schedule_Ward` |
| 床位管理 | `/api/v1/beds` | GET/POST/PUT/DELETE | `Schedule_Bed` + `Schedule_BedEquipmentRel` |
| 班次配置 | `/api/v1/shifts` | GET/POST/PUT/DELETE | `Schedule_Shift` |
| 设备管理 | `/api/v1/devices` | GET/POST/PUT/DELETE | `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` |
| 字典配置 | `/api/v1/dict/types`, `/api/v1/dict/items` | GET/POST/PUT/DELETE | `CodeDictionary_CodeDictionarys` + `dict_items`(新表) |
| 宣教管理 | `/api/v1/health-educations` | GET/POST/PUT/DELETE | `Auxiliary_HealthEducation` |

---

## 五、老库表字段覆盖度

| 老库表 | 总字段数 | 当前已使用 | 未使用字段 |
|--------|---------|-----------|-----------|
| `Plan_PlanTPL` | 35 | 32 | `DryWeight`(模板层不需要), `Frequency`, `AutoConfirmPrescription` |
| `Plan_PlanTPLMaterial` | 8 | 8 | 无 |
| `Order_OrderTPL` | 18 | 16 | `AllDosage`(写0), `Type`(推断) |
| `Auxiliary_MaterialInfomation` | 20 | 18 | `UseTips`, `MinUnitDosage` |
| `Auxiliary_DrugInfomation` | 23 | 20+ | `SpecificationUnit` 映射待确认 |
| `Schedule_Ward` | 12 | 12 | 无 |
| `Schedule_Bed` | 12 | 12 | 无 |
| `Schedule_BedEquipmentRel` | 10 | 10 | 无 |
| `Schedule_Shift` | 12 | 10 | `Type` 未默认, `Builtin` 未使用 |
| `Auxiliary_EquipmentInfomation` | 20 | 20 | 无 |
| `CodeDictionary_CodeDictionarys` | 7 | 6 | `Builtin` 未使用 |
| `Auxiliary_HealthEducation` | 13 | 13 | 无 |
| `Auxiliary_PatientHealthEducation` | 15 | 15 | 无 |

---

## 六、结论与优先修复建议

1. **P0 修复**（影响数据正确性）：
   - 病区管理 `patientType`/`infectionType` 枚举对齐
   - 物理删除改软删除（病区/床位/宣教）

2. **P1 确认**（影响功能完整性）：
   - 方案模板 mode 补 `HD+HP`
   - 设备管理 Notes 映射修正
   - 班次 StartTime/EndTime 类型兼容性

3. **P2 优化**（体验改善）：
   - 字典树形结构老库同步策略
   - 宣教附件管理
   - 材料 UseTips 字段暴露
