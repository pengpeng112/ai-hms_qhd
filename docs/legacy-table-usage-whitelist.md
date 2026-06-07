# 老库表使用白名单

每模块列出允许使用的老库权威表、读写属性、迁移状态。
不在本白名单中的表名 `TableName()` 均视为"待清理/已禁用"。

## 患者管理

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Register_PatientInfomation` | R/W | 已对接 | 患者主表，Patient model；创建时同时写入扩展字段（Spell/Nation/Height/ABOType/RHType/EducationLevel/Occupation/MaritalStatus/Workunit/PhoneNo/WeChatNo/HomePhoneNo/Address/Province/City/County/SSN/DialysisNo/FirstDialysisDate/FirstDialysisHospital/OurHospitalFirstDialysisDate） |
| `Register_Hospitalization` | R/W | 已对接 | 住院信息，创建/更新时写入 HospPatientType/HospNo/CaseNo/MedicalRecordNo |
| `Register_IDInfomation` | R/W | 已对接 | 证件信息，创建/更新时写入 IDType/IDNo |
| `Register_FamilyMember` | R/W | 已对接 | 家属信息，创建/更新时写入 Name/PhoneNo |
| `Register_Diagnosis` | R | 已对接 | DiagnosisDesc 由服务层填充 |
| `Register_VascularAccess` | R/W | 已对接 | |
| `Register_VascularAccessImage` | R/W | 已对接 | |
| `Register_VascularAccessChange` | R/W | 已对接 | 血管通路干预 |
| `Register_Infection` | R | 已对接 | |
| `Plan_PatientPlan` | R/W | 已对接 | DryWeight 来源 |
| `Schedule_Ward` | R | 已对接 | |
| `Schedule_Bed` | R | 已对接 | |
| `Schedule_BedEquipmentRel` | R | 已对接 | 设备绑定 |

**禁用表**：`patient_basic_infos` — 不再读写，`insertPatientBasicInfo()` 已移除，`createBasicInfo` 改为写入 `Register_PatientInfomation` + `Register_Hospitalization` + `Register_IDInfomation` + `Register_FamilyMember`。`PatientBasicInfo` model 仅作为内存 DTO 使用（`GetBasicInfo` 通过 `applyLegacyFallbacks` 从老库填充）。HDIS 同步的 `getHDISPatientID` 已返回禁用错误（老库无 HdisPatientID 对应列）。

## 医嘱管理

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Order_PatientOrder` | R/W | 已对接 | 医嘱主表 |
| `Order_PatientDayOrder` | W | 已对接 | 日医嘱记录 |

**禁用表**：`orders` — 新表已不再写入，`models.Order.TableName()` 返回 `"orders"` 但写路径已全部改为 `Order_PatientOrder`。`Copy/Revise/CreateFromTemplate` 操作暂不支持。

## 治疗管理

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Treatment_Treatment` | R/W | 已对接 | |
| `Treatment_DuringParam` | R | 已对接 | Note/Symptoms/Complication 确认老库无列，禁止写入 |
| `Treatment_AfterSigns` | R | 已对接 | |
| `Treatment_MaterialTrace` | R | 已对接 | 耗材记录 |
| `Treatment_TreatmentMonthSummarySheet` | R/W | 已对接 | ContentJsonb 承载 JSON |

## 排班管理

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Schedule_PatientShift` | R/W | 已对接 | Status=60 表示模板 |
| `Schedule_Timeslot` | R | 已对接 | |
| `Schedule_Ward` | R | 已对接 | |
| `Schedule_Bed` | R | 已对接 | |

## 库存管理

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Stock_Stock` | R | 已对接 | 库存主表 |
| `Stock_ChargeItem` | R | 已对接 | 收费项目 |
| `Stock_Storage` | R | 已对接 | 仓库信息 |
| `Stock_InOutStorage` | R | 已对接 | 出入库主单 |
| `Stock_InOutStorageDetail` | R | 已对接 | 出入库明细 |

**限制**：库存不支持直接新增/删除/调整，只读查询。

## 字典管理

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `CodeDictionary_CodeDictionarys` | R/W | 已对接 | 字典条目，唯一权威来源；CreateItem/UpdateItem/DeleteItem/ToggleItemEnabled 仅写此表；CreateType/UpdateType/DeleteType 已禁用（老库无独立类型表） |

**禁用表**：`dict_types`、`dict_items` — 新表已不再写入，`ListTypes/Items` fallback 逻辑已隔离，仅在老库查询错误时降级。

## 认证/用户/角色

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Identity_Users` | R/W | 已对接 | ASP.NET Identity 用户表 |
| `Identity_UserRoles` | R/W | 已对接 | 用户-角色关联 |
| `Identity_Roles` | R | 已对接 | 仅作为登录角色来源，不用于业务权限管理 |
| `Authorization_Roles` | R/W | 已对接 | 业务角色主表，应用角色管理和权限分配权威来源 |
| `Authorization_RoleUsers` | R | 已对接 | 用户-业务角色关联 |
| `Authorization_RolePermissions` | R/W | 已对接 | 角色-权限关联 |
| `Authorization_Permissions` | R/W | 已对接 | 权限定义表 |
| `Organ_Employee` | R | 已对接 | 员工信息，登录时获取姓名 |

## 健康宣教

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Auxiliary_HealthEducation` | R/W | 已对接 | |
| `Auxiliary_PatientHealthEducation` | R/W | 已对接 | |

## 其他

| 老库表 | 读写 | 状态 | 备注 |
|--------|------|------|------|
| `Treatment_DuringParam_Template` | R | 已对接 | 诊疗参数模板 |
| `Treatment_AfterSigns_Template` | R | 已对接 | 体征模板 |
| `Schedule_PatientShift_Template` | R | 已对接 | 排班模板 |

## 待确认/禁用表

| 新表名 | 状态 | 处理方式 |
|--------|------|----------|
| `patient_basic_infos` | 已迁移 | P3-01 完成：`insertPatientBasicInfo` 已移除，`createBasicInfo` 改写入老库四张表（`Register_PatientInfomation`/`Register_Hospitalization`/`Register_IDInfomation`/`Register_FamilyMember`）；`getHDISPatientID` 返回禁用错误 |
| `orders` | 禁用 | 写路径已改到 Order_PatientOrder，部分操作暂不支持 |
| `dict_types` | 禁用 | 仅作为 fallback 降级路径保留读 |
| `dict_items` | 禁用 | 仅作为 fallback 降级路径保留读 |
| `clinical_tasks` | 禁用 | clinical_task_service 读可派生，写返回"暂不支持" |
| `integration_hdis_settings` | 待确认 | 若需保留，用户须明确批准新配置存储 |
| `order_templates` | 待确认 | CreateFromTemplate 已禁用 |
