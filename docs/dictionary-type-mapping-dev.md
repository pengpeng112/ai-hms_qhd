# 字典配置开发说明（CodeDictionary_CodeDictionarys）

更新时间：2026-04-23  
适用范围：`字典配置`菜单、以及所有依赖 `/api/v1/dict/*` 的下拉选项页面。

## 1. 数据源与规则

- 后端字典服务已改为**优先读取老库**：`CodeDictionary_CodeDictionarys`。
- 字典类型判断以 `Type` 字段为准。
- 对于可识别 `Type`，后端映射到统一业务字典编码（`DICT_TYPES`）。
- 对于未识别 `Type`，后端直接使用原 `Type` 作为 `code` 返回；前端字典配置页会自动归入“其他”分组。

## 2. Type -> 统一字典编码映射

| 老库 Type | 统一 code（前端 DICT_TYPES） | 用途 |
|---|---|---|
| `DialysisMethod` | `DIALYSIS_MODE` | 透析方式 |
| `HeparinType` | `ANTICOAGULANT` | 抗凝剂类型 |
| `Dialysate` | `DIALYSATE_TYPE` / `DIALYSATE_GROUP` | 透析液类型/组 |
| `DialysateFlow` | `DIALYSATE_FLOW` | 透析液流量 |
| `GlucoseConOptions` | `GLUCOSE` | 葡萄糖类型 |
| `MaterialType` | `MATERIAL_CATEGORY` | 材料分类 |
| `DrugType` | `DRUG_CATEGORY` | 药品分类 |
| `UseMethodType` | `ORDER_TYPE` | 医嘱类型 |
| `CatalogType` | `ORDER_CATEGORY` | 医嘱分类 |
| `UseWayType` | `ORDER_ROUTE` | 医嘱用法 |
| `FrequencyType` | `ORDER_FREQUENCY` | 医嘱频次 |
| `UseOpportunityType` | `ORDER_TIMING` | 医嘱时机 |
| `AccessType` | `VASCULAR_ACCESS` | 血管通路类型 |
| `AccessPosition` | `VASCULAR_SITE` | 血管通路部位 |
| `VenousType` | `VEIN_TYPE` | 静脉类型 |
| `ArteryType` | `ARTERY_TYPE` | 动脉类型 |
| `ExpenseType` | `INSURANCE_TYPE` | 医保类型 |
| `PatientType` | `PATIENT_TYPE` | 患者类型 |
| `HospPatientType` | `VISIT_CATEGORY` | 就诊类别 |
| `IDType` | `ID_TYPE` | 证件类型 |
| `ABOType` | `BLOOD_TYPE_ABO` | ABO血型 |
| `RHType` | `BLOOD_TYPE_RH` | Rh血型 |
| `EducationLevel` | `EDUCATION_LEVEL` | 文化程度 |
| `MaritalStatus` | `MARITAL_STATUS` | 婚姻状况 |
| `OutComeType` + `OutComeReason` | `OUTCOME` | 转归（树形） |

## 3. OUTCOME 特殊处理

- `OUTCOME` 由两个老库 `Type` 合成：
  - 一级：`OutComeType`（父节点）
  - 二级：`OutComeReason`（子节点）
- `OutComeReason.Name` 若为 `父码|名称` 格式，会自动解析为 `parentCode + name`。

## 4. 接口约定（供页面统一调用）

- 字典类型列表：`GET /api/v1/dict/types`
- 字典项列表：`GET /api/v1/dict/items/{typeCode}`
- 字典项树：`GET /api/v1/dict/items/{typeCode}/tree`

> 后续功能页面的选项（下拉/级联）应统一通过以上接口获取，不再硬编码枚举。

## 5. 未匹配 Type 的处理

- 未出现在映射表的 `Type`：
  - 保留原 `Type` 作为 `code`
  - 返回到字典配置页后归类到“其他”
- 如后续业务确认要纳入固定分组，请补充本文件映射表并同步后端映射代码。

