# 老血透数据库表结构（合并版）

> **来源合并**
> - `数据库表结构.md` — 字段级权威（字段名、类型、主键、非空、默认值）
> - `数据库表设计.pdf` — 模块分组、表顺序与业务含义（**中文注释因 PDF 字体缺少 ToUnicode CMap，`pdftotext` 提取丢失**，需人工/OCR 回填）
> - `数据库表ER图.pdf` — 外键多重度标记（`*FieldId` → 指向父表主键）
>
> **本文件用途**：作为后端代码对齐老血透数据库的权威参考。后续新血透 `DATABASE_DESIGN.md` 将以此做字段级对照。
>
> **统计**：102 张表；设计 PDF 覆盖 80 张；ER 图候选外键字段 1 个。

## 模块索引

- [患者档案（Register）（19 张）](#register)
- [治疗计划（Plan）（7 张）](#plan)
- [排班（Schedule）（9 张）](#schedule)
- [医嘱（Order）（4 张）](#order)
- [治疗记录（Treatment）（16 张）](#treatment)
- [基础数据 / 辅助资料（Auxiliary）（11 张）](#auxiliary)
- [检验接口（LIS）（5 张）](#lis)
- [设备日志（Device）（12 张）](#device)
- [库存（Stock）（6 张）](#stock)
- [费用（Cost）（3 张）](#cost)
- [质控评估（QualityEvaluation）（1 张）](#qualityevaluation)
- [通知（Notify）（2 张）](#notify)
- [消息（Message）（1 张）](#message)
- [留言板（MessageBoard）（1 张）](#messageboard)
- [系统日志（Log）（1 张）](#log)
- [应用配置（Applications）（1 张）](#applications)
- [代码字典（CodeDictionary）（1 张）](#codedictionary)
- [租户配置（TenantConfig）（1 张）](#tenantconfig)
- [用户（User）（1 张）](#user)

## 患者档案（Register）

### `Register_PatientInfomation`

- 字段数：42
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(256) |  | ✓ |  | — 待补 |
| 4 | `Spell` | character varying(256) |  |  |  | — 待补 |
| 5 | `Type` | character varying(64) |  |  |  | — 待补 |
| 6 | `TreatmentStatus` | character varying(64) |  |  |  | — 待补 |
| 7 | `OutComeStatus` | character varying(64) |  |  |  | — 待补 |
| 8 | `Gender` | character varying(64) |  |  |  | — 待补 |
| 9 | `BirthDate` | timestamp |  |  |  | — 待补 |
| 10 | `Nation` | character varying(64) |  |  |  | — 待补 |
| 11 | `ABOType` | character varying(64) |  |  |  | — 待补 |
| 12 | `RHType` | character varying(64) |  |  |  | — 待补 |
| 13 | `Height` | numeric |  |  |  | — 待补 |
| 14 | `Weight` | numeric |  |  |  | — 待补 |
| 15 | `Occupation` | character varying(128) |  |  |  | — 待补 |
| 16 | `MaritalStatus` | character varying(64) |  |  |  | — 待补 |
| 17 | `EducationLevel` | character varying(64) |  |  |  | — 待补 |
| 18 | `Province` | character varying(64) |  |  |  | — 待补 |
| 19 | `City` | character varying(64) |  |  |  | — 待补 |
| 20 | `County` | character varying(64) |  |  |  | — 待补 |
| 21 | `Address` | character varying(256) |  |  |  | — 待补 |
| 22 | `PhoneNo` | character varying(32) |  |  |  | — 待补 |
| 23 | `ExpenseType` | character varying(64) |  |  |  | — 待补 |
| 24 | `SSN` | character varying(64) |  |  |  | — 待补 |
| 25 | `DialysisNo` | character varying(64) |  |  |  | — 待补 |
| 26 | `ResponsibilityDrId` | bigint |  |  |  | ID 外键（待核对） |
| 27 | `ResponsibilityNurseId` | bigint |  |  |  | ID 外键（待核对） |
| 28 | `FirstDialysisDate` | timestamp |  |  |  | — 待补 |
| 29 | `FirstDialysisHospital` | character varying(64) |  |  |  | — 待补 |
| 30 | `Note` | character varying(1024) |  |  |  | 备注 |
| 31 | `WeChatNo` | character varying(128) |  |  |  | — 待补 |
| 32 | `ImportInfo` | character varying(1024) |  |  |  | — 待补 |
| 33 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 34 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 35 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 36 | `IDName` | character varying(256) |  |  |  | — 待补 |
| 37 | `HomePhoneNo` | character varying(32) |  |  |  | — 待补 |
| 38 | `ImageBase64String` | text |  |  |  | 图像 Base64 |
| 39 | `PatientType` | character varying(64) |  |  |  | — 待补 |
| 40 | `Workunit` | character varying(256) |  |  |  | — 待补 |
| 41 | `PredictWeight` | numeric |  |  |  | — 待补 |
| 42 | `OurHospitalFirstDialysisDate` | timestamp |  |  |  | — 待补 |

### `Register_Allergen`

- 字段数：11
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `Name` | character varying(1024) |  |  |  | — 待补 |
| 6 | `ExamineTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `ExamineDr` | character varying(64) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_Complication`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `Name` | character varying(512) |  |  |  | — 待补 |
| 6 | `Description` | character varying(1024) |  |  |  | — 待补 |
| 7 | `ExamineTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 8 | `ExamineDr` | character varying(64) |  |  |  | — 待补 |
| 9 | `Note` | character varying(1024) |  |  |  | 备注 |
| 10 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 11 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 12 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_Diagnosis`

- 字段数：11
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `DiagnosisDate` | timestamp |  |  |  | — 待补 |
| 5 | `Complaints` | character varying(512) |  |  |  | — 待补 |
| 6 | `DiagnosisDesc` | character varying(1024) |  |  |  | — 待补 |
| 7 | `ComplicationDesc` | character varying(1024) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_ElectronicDocument`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  |  |  | — 待补 |
| 5 | `Name` | character varying(256) |  |  |  | — 待补 |
| 6 | `AttachmentIds` | character varying |  |  |  | — 待补 |
| 7 | `PatientSign` | character varying(64) |  |  |  | — 待补 |
| 8 | `SignTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `Note` | character varying(1024) |  |  |  | 备注 |
| 10 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 11 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 12 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_FamilyMember`

- 字段数：11
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Name` | character varying(64) |  |  |  | — 待补 |
| 5 | `Kinship` | character varying(64) |  |  |  | — 待补 |
| 6 | `PhoneNo` | character varying(32) |  |  |  | — 待补 |
| 7 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 11 | `Type` | character varying(64) |  |  |  | — 待补 |

### `Register_Hospitalization`

- 字段数：16
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `CaseNo` | character varying(64) |  |  |  | — 待补 |
| 5 | `HospNo` | character varying(64) |  |  |  | — 待补 |
| 6 | `BarCode` | character varying(64) |  |  |  | — 待补 |
| 7 | `HospPatientType` | character varying(64) |  |  |  | — 待补 |
| 8 | `HospReceiveDept` | character varying(64) |  |  |  | — 待补 |
| 9 | `HospWard` | character varying(64) |  |  |  | — 待补 |
| 10 | `HospBed` | character varying(64) |  |  |  | — 待补 |
| 11 | `AttendDr` | character varying(64) |  |  |  | — 待补 |
| 12 | `ReceptionDr` | character varying(64) |  |  |  | — 待补 |
| 13 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 14 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 15 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 16 | `MedicalRecordNo` | character varying(64) |  |  |  | — 待补 |

### `Register_IDInfomation`

- 字段数：9
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `IDType` | character varying(64) |  |  |  | — 待补 |
| 5 | `IDNo` | character varying(64) |  |  |  | — 待补 |
| 6 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 7 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 8 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 9 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_Image`

- 字段数：13
- 设计文档覆盖：✅
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint |  | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `ImageName` | character varying(128) |  |  |  | — 待补 |
| 5 | `ImageBase64String` | text |  |  |  | 图像 Base64 |
| 6 | `FeatureCode` | text |  |  |  | — 待补 |
| 7 | `Type` | character varying(64) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `Sort` | numeric |  | ✓ |  | — 待补 |
| 13 | `BizId` | bigint |  |  |  | ID 外键（待核对） |

### `Register_Infection`

- 字段数：9
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `InfectionDesc` | character varying(1024) |  |  |  | — 待补 |
| 5 | `OtherDesc` | character varying(1024) |  |  |  | — 待补 |
| 6 | `Note` | character varying(1024) |  |  |  | 备注 |
| 7 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 8 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 9 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_MedicalHistory`

- 字段数：18
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Complaints` | text |  |  |  | — 待补 |
| 5 | `PresentIllnessHistory` | text |  |  |  | — 待补 |
| 6 | `PastIllnessHistory` | text |  |  |  | — 待补 |
| 7 | `PersonalHistory` | text |  |  |  | — 待补 |
| 8 | `MaritalReproductiveHistory` | text |  |  |  | — 待补 |
| 9 | `FamilyHistory` | text |  |  |  | — 待补 |
| 10 | `DiagnosisDesc` | text |  |  |  | — 待补 |
| 11 | `PhysicalExamination` | text |  |  |  | — 待补 |
| 12 | `SpecialistExamination` | text |  |  |  | — 待补 |
| 13 | `AncillaryExamination` | text |  |  |  | — 待补 |
| 14 | `Note` | character varying(1024) |  |  |  | 备注 |
| 15 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 16 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 17 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 18 | `Narrator` | character varying(64) |  |  |  | — 待补 |

### `Register_OrderSheet`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `OrderTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 5 | `Status` | integer |  | ✓ |  | — 待补 |
| 6 | `Age` | integer |  | ✓ |  | — 待补 |
| 7 | `ImageBase64String` | text |  |  |  | 图像 Base64 |
| 8 | `DigitalSignature` | text |  |  |  | — 待补 |
| 9 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `ContentJsonb` | jsonb |  |  |  | — 待补 |

### `Register_OutCome`

- 字段数：10
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  |  |  | — 待补 |
| 5 | `Reason` | character varying(64) |  |  |  | — 待补 |
| 6 | `OutComeTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `Note` | character varying(1024) |  |  |  | 备注 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_Pathology`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `Name` | character varying(512) |  |  |  | — 待补 |
| 6 | `ExamineTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `ExamineDr` | character varying(64) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `Description` | character varying(1024) |  |  |  | — 待补 |

### `Register_Protopathy`

- 字段数：11
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `Name` | character varying(512) |  |  |  | — 待补 |
| 6 | `ExamineTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `ExamineDr` | character varying(64) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_Tumor`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `TreatmentDesc` | character varying(1024) |  |  |  | — 待补 |
| 6 | `Name` | character varying(512) |  |  |  | — 待补 |
| 7 | `ExamineTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 8 | `ExamineDr` | character varying(64) |  |  |  | — 待补 |
| 9 | `Note` | character varying(1024) |  |  |  | 备注 |
| 10 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 11 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 12 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_VascularAccess`

- 字段数：26
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `AccessType` | character varying(64) |  |  |  | — 待补 |
| 5 | `AccessPosition` | character varying(128) |  |  |  | — 待补 |
| 6 | `Artery` | character varying(128) |  |  |  | — 待补 |
| 7 | `Venous` | character varying(128) |  |  |  | — 待补 |
| 8 | `LeftAndRight` | character varying(128) |  |  |  | — 待补 |
| 9 | `OperationTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 10 | `Note` | character varying(1024) |  |  |  | 备注 |
| 11 | `PictureIds` | bigint |  |  |  | — 待补 |
| 12 | `IsDefault` | boolean |  | ✓ | false | — 待补 |
| 13 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 14 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 15 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 16 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 17 | `ASidePointCount` | character varying(256) |  |  |  | — 待补 |
| 18 | `OperationHospital` | character varying(128) |  |  |  | — 待补 |
| 19 | `VSidePointCount` | character varying(256) |  |  |  | — 待补 |
| 20 | `CatheterizeMethod` | character varying(128) |  |  |  | — 待补 |
| 21 | `CatheterDepth` | numeric |  |  |  | — 待补 |
| 22 | `FirstUseTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 23 | `InterveneCount` | bigint |  |  |  | — 待补 |
| 24 | `AccessCount` | bigint |  |  |  | — 待补 |
| 25 | `InterveneTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 26 | `OperationDr` | text |  |  |  | — 待补 |

### `Register_VascularAccessChange`

- 字段数：14
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `VascularAccessId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `UseDuration` | integer |  | ✓ |  | — 待补 |
| 6 | `AvgBF` | numeric |  | ✓ |  | — 待补 |
| 7 | `ChangeTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 8 | `ChangeReason` | character varying(512) |  |  |  | — 待补 |
| 9 | `ChangeDesc` | character varying(512) |  |  |  | — 待补 |
| 10 | `SketchMap` | bigint |  |  |  | — 待补 |
| 11 | `PhysicalMap` | bigint |  |  |  | — 待补 |
| 12 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 13 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 14 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Register_VascularAccessImage`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `VascularAccessId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `ImageName` | character varying(128) |  |  |  | — 待补 |
| 5 | `ImageBase64String` | text |  |  |  | 图像 Base64 |
| 6 | `Note` | character varying(1024) |  |  |  | 备注 |
| 7 | `Sort` | numeric |  | ✓ |  | — 待补 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

## 治疗计划（Plan）

### `Plan_PatientPlan`

- 字段数：48
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Name` | character varying(256) |  | ✓ |  | — 待补 |
| 5 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 6 | `CreateTime` | timestamp |  |  |  | 创建时间 |
| 7 | `PlanTPLId` | bigint |  |  |  | ID 外键（待核对） |
| 8 | `OddWeekFrequency` | integer |  |  |  | — 待补 |
| 9 | `EvenWeekFrequency` | integer |  |  |  | — 待补 |
| 10 | `DialysisMethod` | character varying(256) |  | ✓ |  | — 待补 |
| 11 | `DialysisDuration` | numeric |  |  |  | — 待补 |
| 12 | `DryWeight` | numeric |  |  |  | — 待补 |
| 13 | `ExtraWeight` | numeric |  |  |  | — 待补 |
| 14 | `AdjustQuantity` | numeric |  |  |  | — 待补 |
| 15 | `BF` | numeric |  |  |  | — 待补 |
| 16 | `BV` | numeric |  |  |  | — 待补 |
| 17 | `FirstAnticoagulant` | bigint |  |  |  | — 待补 |
| 18 | `FirstDosage` | numeric |  |  |  | — 待补 |
| 19 | `MaintainAnticoagulant` | bigint |  |  |  | — 待补 |
| 20 | `DilutionProportion` | numeric |  |  |  | — 待补 |
| 21 | `InjectionRate` | numeric |  |  |  | — 待补 |
| 22 | `InjectionDuration` | numeric |  |  |  | — 待补 |
| 23 | `InjectionVolume` | numeric |  |  |  | — 待补 |
| 24 | `VascularAccessId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 25 | `Dialysate` | character varying(64) |  |  |  | — 待补 |
| 26 | `DialysateFlow` | numeric |  |  |  | — 待补 |
| 27 | `DialysateVolume` | numeric |  |  |  | — 待补 |
| 28 | `NaIonCon` | numeric |  |  |  | — 待补 |
| 29 | `CaIonCon` | numeric |  |  |  | — 待补 |
| 30 | `KIonCon` | numeric |  |  |  | — 待补 |
| 31 | `HCO3IonCon` | numeric |  |  |  | — 待补 |
| 32 | `Conductivity` | numeric |  |  |  | — 待补 |
| 33 | `DialysateTmp` | numeric |  |  |  | — 待补 |
| 34 | `SubstituateVolume` | numeric |  |  |  | — 待补 |
| 35 | `DilutionMnt` | character varying(64) |  |  |  | — 待补 |
| 36 | `IsDisabled` | boolean |  | ✓ |  | 是否禁用/软删除 |
| 37 | `LastModifyTime` | timestamp |  |  |  | 最近修改时间 |
| 38 | `SalineQuantity` | numeric |  |  |  | — 待补 |
| 39 | `SealQuantity` | numeric |  |  |  | — 待补 |
| 40 | `ArterialQuantity` | numeric |  |  |  | — 待补 |
| 41 | `VenousQuantity` | numeric |  |  |  | — 待补 |
| 42 | `SealType` | character varying(64) |  |  |  | — 待补 |
| 43 | `Frequency` | character varying(128) |  |  |  | — 待补 |
| 44 | `GlucoseCon` | numeric |  |  |  | — 待补 |
| 45 | `DialysateGroupId` | bigint |  |  |  | ID 外键（待核对） |
| 46 | `AutoConfirmPrescription` | character varying(64) |  |  | `'10'::charactervarying` | — 待补 |
| 47 | `Note` | text |  |  |  | 备注 |
| 48 | `SubstituateFlow` | numeric |  |  |  | — 待补 |

### `Plan_PatientPlanMaterial`

- 字段数：8
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientPlanId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `MaterialId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `MaterialGroup` | bigint |  | ✓ |  | — 待补 |
| 6 | `Num` | numeric |  | ✓ |  | — 待补 |
| 7 | `Note` | character varying(512) |  | ✓ |  | 备注 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Plan_PatientPlanPrescriptionAdjustment`

- 字段数：94
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ | `nextval('public."Plan_PatientPlanPrescriptionAdjustment_Id_seq"'::regclass)` | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Type` | integer |  | ✓ |  | — 待补 |
| 4 | `PatientPlanPrescriptionId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `AdjustUserId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `AdjustTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `AdjustReason` | character varying(512) |  |  |  | — 待补 |
| 8 | `BeforeOddWeekFrequency` | integer |  |  |  | — 待补 |
| 9 | `AfterOddWeekFrequency` | integer |  |  |  | — 待补 |
| 10 | `BeforeEvenWeekFrequency` | integer |  |  |  | — 待补 |
| 11 | `AfterEvenWeekFrequency` | integer |  |  |  | — 待补 |
| 12 | `BeforeDialysisMethod` | character varying(512) |  |  |  | — 待补 |
| 13 | `AfterDialysisMethod` | character varying(512) |  |  |  | — 待补 |
| 14 | `BeforeDialysisDuration` | numeric |  |  |  | — 待补 |
| 15 | `AfterDialysisDuration` | numeric |  |  |  | — 待补 |
| 16 | `BeforeDryWeight` | numeric |  |  |  | — 待补 |
| 17 | `AfterDryWeight` | numeric |  |  |  | — 待补 |
| 18 | `BeforeExtraWeight` | numeric |  |  |  | — 待补 |
| 19 | `AfterExtraWeight` | numeric |  |  |  | — 待补 |
| 20 | `BeforeAdjustQuantity` | numeric |  |  |  | — 待补 |
| 21 | `AfterAdjustQuantity` | numeric |  |  |  | — 待补 |
| 22 | `BeforeBF` | numeric |  |  |  | — 待补 |
| 23 | `AfterBF` | numeric |  |  |  | — 待补 |
| 24 | `BeforeFirstAnticoagulant` | bigint |  |  |  | — 待补 |
| 25 | `AfterFirstAnticoagulant` | bigint |  |  |  | — 待补 |
| 26 | `BeforeFirstDosage` | numeric |  |  |  | — 待补 |
| 27 | `AfterFirstDosage` | numeric |  |  |  | — 待补 |
| 28 | `BeforeMaintainAnticoagulant` | bigint |  |  |  | — 待补 |
| 29 | `AfterMaintainAnticoagulant` | bigint |  |  |  | — 待补 |
| 30 | `BeforeDilutionProportion` | numeric |  |  |  | — 待补 |
| 31 | `AfterDilutionProportion` | numeric |  |  |  | — 待补 |
| 32 | `BeforeInjectionRate` | numeric |  |  |  | — 待补 |
| 33 | `AfterInjectionRate` | numeric |  |  |  | — 待补 |
| 34 | `BeforeInjectionDuration` | numeric |  |  |  | — 待补 |
| 35 | `AfterInjectionDuration` | numeric |  |  |  | — 待补 |
| 36 | `BeforeInjectionVolume` | numeric |  |  |  | — 待补 |
| 37 | `AfterInjectionVolume` | numeric |  |  |  | — 待补 |
| 38 | `BeforeVascularAccessId` | bigint |  |  |  | ID 外键（待核对） |
| 39 | `AfterVascularAccessId` | bigint |  |  |  | ID 外键（待核对） |
| 40 | `BeforeDialysate` | character varying(64) |  |  |  | — 待补 |
| 41 | `AfterDialysate` | character varying(64) |  |  |  | — 待补 |
| 42 | `BeforeDialysateFlow` | numeric |  |  |  | — 待补 |
| 43 | `AfterDialysateFlow` | numeric |  |  |  | — 待补 |
| 44 | `BeforeDialysateVolume` | numeric |  |  |  | — 待补 |
| 45 | `AfterDialysateVolume` | numeric |  |  |  | — 待补 |
| 46 | `BeforeNaIonCon` | numeric |  |  |  | — 待补 |
| 47 | `AfterNaIonCon` | numeric |  |  |  | — 待补 |
| 48 | `BeforeCaIonCon` | numeric |  |  |  | — 待补 |
| 49 | `AfterCaIonCon` | numeric |  |  |  | — 待补 |
| 50 | `BeforeKIonCon` | numeric |  |  |  | — 待补 |
| 51 | `AfterKIonCon` | numeric |  |  |  | — 待补 |
| 52 | `BeforeHCO3IonCon` | numeric |  |  |  | — 待补 |
| 53 | `AfterHCO3IonCon` | numeric |  |  |  | — 待补 |
| 54 | `BeforeConductivity` | numeric |  |  |  | — 待补 |
| 55 | `AfterConductivity` | numeric |  |  |  | — 待补 |
| 56 | `BeforeDialysateTmp` | numeric |  |  |  | — 待补 |
| 57 | `AfterDialysateTmp` | numeric |  |  |  | — 待补 |
| 58 | `BeforeSubstituateVolume` | numeric |  |  |  | — 待补 |
| 59 | `AfterSubstituateVolume` | numeric |  |  |  | — 待补 |
| 60 | `BeforeDilutionMnt` | character varying(64) |  |  |  | — 待补 |
| 61 | `AfterDilutionMnt` | character varying(64) |  |  |  | — 待补 |
| 62 | `BeforeSalineQuantity` | numeric |  |  |  | — 待补 |
| 63 | `AfterSalineQuantity` | numeric |  |  |  | — 待补 |
| 64 | `BeforeSealType` | character varying(64) |  |  |  | — 待补 |
| 65 | `AfterSealType` | character varying(64) |  |  |  | — 待补 |
| 66 | `BeforeSealQuantity` | numeric |  |  |  | — 待补 |
| 67 | `AfterSealQuantity` | numeric |  |  |  | — 待补 |
| 68 | `BeforeArterialQuantity` | numeric |  |  |  | — 待补 |
| 69 | `AfterArterialQuantity` | numeric |  |  |  | — 待补 |
| 70 | `BeforeVenousQuantity` | numeric |  |  |  | — 待补 |
| 71 | `AfterVenousQuantity` | numeric |  |  |  | — 待补 |
| 72 | `BeforeUFQuantity` | numeric |  |  |  | — 待补 |
| 73 | `AfterUFQuantity` | numeric |  |  |  | — 待补 |
| 74 | `Status` | integer |  |  |  | — 待补 |
| 75 | `DealUserId` | bigint |  |  |  | ID 外键（待核对） |
| 76 | `DealTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 77 | `DealContent` | character varying(512) |  |  |  | — 待补 |
| 78 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 79 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 80 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 81 | `BeforeBV` | numeric |  |  |  | — 待补 |
| 82 | `AfterBV` | numeric |  |  |  | — 待补 |
| 83 | `BeforeFrequency` | character varying(128) |  |  |  | — 待补 |
| 84 | `AfterFrequency` | character varying(128) |  |  |  | — 待补 |
| 85 | `BeforeGlucoseCon` | numeric |  |  |  | — 待补 |
| 86 | `AfterGlucoseCon` | numeric |  |  |  | — 待补 |
| 87 | `BeforeSubstituateFlow` | numeric |  |  |  | — 待补 |
| 88 | `AfterSubstituateFlow` | numeric |  |  |  | — 待补 |
| 89 | `BeforeIsInduceDialysisPrescription` | boolean |  |  |  | — 待补 |
| 90 | `AfterIsInduceDialysisPrescription` | boolean |  |  |  | — 待补 |
| 91 | `BeforeHeparinType` | integer |  |  |  | — 待补 |
| 92 | `AfterHeparinType` | integer |  |  |  | — 待补 |
| 93 | `BeforeMaterial` | jsonb |  |  |  | — 待补 |
| 94 | `AfterMaterial` | jsonb |  |  |  | — 待补 |

### `Plan_PatientPrescription`

- 字段数：49
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `PatientPlanId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 7 | `CreateTime` | timestamp |  |  |  | 创建时间 |
| 8 | `ConfirmUserId` | bigint |  |  |  | ID 外键（待核对） |
| 9 | `ConfirmTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 10 | `Status` | integer |  |  |  | — 待补 |
| 11 | `CaseStatus` | character varying(64) |  |  |  | — 待补 |
| 12 | `DialysisMethod` | character varying(256) |  | ✓ |  | — 待补 |
| 13 | `DialysisDuration` | numeric |  |  |  | — 待补 |
| 14 | `DryWeight` | numeric |  |  |  | — 待补 |
| 15 | `AdjustQuantity` | numeric |  |  |  | — 待补 |
| 16 | `BF` | numeric |  |  |  | — 待补 |
| 17 | `BV` | numeric |  |  |  | — 待补 |
| 18 | `FirstAnticoagulant` | bigint |  |  |  | — 待补 |
| 19 | `FirstDosage` | numeric |  |  |  | — 待补 |
| 20 | `MaintainAnticoagulant` | bigint |  |  |  | — 待补 |
| 21 | `DilutionProportion` | numeric |  |  |  | — 待补 |
| 22 | `InjectionRate` | numeric |  |  |  | — 待补 |
| 23 | `InjectionDuration` | numeric |  |  |  | — 待补 |
| 24 | `InjectionVolume` | numeric |  |  |  | — 待补 |
| 25 | `VascularAccessId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 26 | `Dialysate` | character varying(64) |  |  |  | — 待补 |
| 27 | `DialysateFlow` | numeric |  |  |  | — 待补 |
| 28 | `DialysateVolume` | numeric |  |  |  | — 待补 |
| 29 | `NaIonCon` | numeric |  |  |  | — 待补 |
| 30 | `CaIonCon` | numeric |  |  |  | — 待补 |
| 31 | `KIonCon` | numeric |  |  |  | — 待补 |
| 32 | `HCO3IonCon` | numeric |  |  |  | — 待补 |
| 33 | `Conductivity` | numeric |  |  |  | — 待补 |
| 34 | `DialysateTmp` | numeric |  |  |  | — 待补 |
| 35 | `SubstituateVolume` | numeric |  |  |  | — 待补 |
| 36 | `DilutionMnt` | character varying(64) |  |  |  | — 待补 |
| 37 | `LastModifyTime` | timestamp |  |  |  | 最近修改时间 |
| 38 | `SalineQuantity` | numeric |  |  |  | — 待补 |
| 39 | `SealQuantity` | numeric |  |  |  | — 待补 |
| 40 | `ArterialQuantity` | numeric |  |  |  | — 待补 |
| 41 | `VenousQuantity` | numeric |  |  |  | — 待补 |
| 42 | `UFQuantity` | numeric |  | ✓ | 0 | — 待补 |
| 43 | `SealType` | character varying(64) |  |  |  | — 待补 |
| 44 | `GlucoseCon` | numeric |  |  |  | — 待补 |
| 45 | `DialysateGroupId` | bigint |  |  |  | ID 外键（待核对） |
| 46 | `Note` | text |  |  |  | 备注 |
| 47 | `SubstituateFlow` | numeric |  |  |  | — 待补 |
| 48 | `IsInduceDialysisPrescription` | boolean |  |  | false | — 待补 |
| 49 | `HeparinType` | integer |  |  | 0 | — 待补 |

### `Plan_PatientPrescriptionMaterial`

- 字段数：9
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientPrescriptionId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `MaterialId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `MaterialGroup` | bigint |  | ✓ |  | — 待补 |
| 6 | `Num` | numeric |  | ✓ |  | — 待补 |
| 7 | `Note` | character varying(512) |  |  |  | 备注 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 9 | `ChargeItemId` | bigint |  |  | 0 | ID 外键（待核对） |

### `Plan_PlanTPL`

- 字段数：35
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 5 | `CreateTime` | timestamp |  |  |  | 创建时间 |
| 6 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 7 | `DialysisMethod` | character varying(256) |  | ✓ |  | — 待补 |
| 8 | `DialysisDuration` | numeric |  |  |  | — 待补 |
| 9 | `AdjustQuantity` | numeric |  |  |  | — 待补 |
| 10 | `BF` | numeric |  |  |  | — 待补 |
| 11 | `BV` | numeric |  |  |  | — 待补 |
| 12 | `FirstAnticoagulant` | bigint |  |  |  | — 待补 |
| 13 | `FirstDosage` | numeric |  |  |  | — 待补 |
| 14 | `MaintainAnticoagulant` | bigint |  |  |  | — 待补 |
| 15 | `DilutionProportion` | numeric |  |  |  | — 待补 |
| 16 | `InjectionRate` | numeric |  |  |  | — 待补 |
| 17 | `InjectionDuration` | numeric |  |  |  | — 待补 |
| 18 | `InjectionVolume` | numeric |  |  |  | — 待补 |
| 19 | `VascularAccessId` | bigint |  |  |  | ID 外键（待核对） |
| 20 | `Dialysate` | character varying(64) |  |  |  | — 待补 |
| 21 | `DialysateFlow` | numeric |  |  |  | — 待补 |
| 22 | `DialysateVolume` | numeric |  |  |  | — 待补 |
| 23 | `NaIonCon` | numeric |  |  |  | — 待补 |
| 24 | `CaIonCon` | numeric |  |  |  | — 待补 |
| 25 | `KIonCon` | numeric |  |  |  | — 待补 |
| 26 | `Conductivity` | numeric |  |  |  | — 待补 |
| 27 | `DialysateTmp` | numeric |  |  |  | — 待补 |
| 28 | `SubstituateVolume` | numeric |  |  |  | — 待补 |
| 29 | `DilutionMnt` | character varying(64) |  |  |  | — 待补 |
| 30 | `LastModifyTime` | timestamp |  |  |  | 最近修改时间 |
| 31 | `HCO3IonCon` | numeric |  |  |  | — 待补 |
| 32 | `GlucoseCon` | numeric |  |  |  | — 待补 |
| 33 | `DialysateGroupId` | bigint |  |  |  | ID 外键（待核对） |
| 34 | `Note` | text |  |  |  | 备注 |
| 35 | `SubstituateFlow` | numeric |  |  |  | — 待补 |

### `Plan_PlanTPLMaterial`

- 字段数：8
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PlanTPLId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `MaterialId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `MaterialGroup` | bigint |  | ✓ |  | — 待补 |
| 6 | `Num` | numeric |  | ✓ |  | — 待补 |
| 7 | `Note` | character varying(512) |  | ✓ |  | 备注 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

## 排班（Schedule）

### `Schedule_Bed`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `Name` | character varying(256) |  | ✓ |  | — 待补 |
| 4 | `Sort` | integer |  |  |  | — 待补 |
| 5 | `WardId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `IsDisabled` | boolean |  |  | false | 是否禁用/软删除 |
| 7 | `Note` | character varying(512) |  |  |  | 备注 |
| 8 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 11 | `FEPId` | bigint |  |  |  | ID 外键（待核对） |
| 12 | `AcquisiteConnectId` | bigint |  |  |  | ID 外键（待核对） |

### `Schedule_BedEquipmentRel`

- 字段数：10
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `EquipmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `Sort` | integer |  |  |  | — 待补 |
| 5 | `BedId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `IsDefault` | boolean |  |  |  | — 待补 |
| 7 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 8 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |
| 9 | `Type` | integer |  | ✓ | 1, note: '1 业务设备管理表  2:云中心用户表  3：待扩展...' | — 待补 |
| 10 | `ParameterS` | jsonb |  |  |  | — 待补 |

### `Schedule_BedEquipmentRelChange`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `EquipmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `Sort` | integer |  |  |  | — 待补 |
| 5 | `BedId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `IsDefault` | boolean |  |  |  | — 待补 |
| 7 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 8 | `Type` | integer |  | ✓ | 1, note: '1 业务设备管理表  2:云中心用户表  3：待扩展...' | — 待补 |
| 9 | `ParameterS` | jsonb |  |  |  | — 待补 |
| 10 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 11 | `CreateTime` | timestamp(6) |  | ✓ | `now()` | 创建时间 |
| 12 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |

### `Schedule_BedFEPChange`

- 字段数：8
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `BedId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `BeforeFEPId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `AfterFEPId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 7 | `CreateTime` | timestamp(6) |  | ✓ | `now()` | 创建时间 |
| 8 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |

### `Schedule_CheckIn`

- 字段数：13
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `LastHandoverIds` | character varying(64) |  |  |  | — 待补 |
| 4 | `ShiftId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `WardId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `ClockInTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `OperatorType` | bigint |  |  |  | — 待补 |
| 8 | `Type` | bigint |  |  |  | — 待补 |
| 9 | `Note` | text |  |  |  | 备注 |
| 10 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 11 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Schedule_Handover`

- 字段数：15
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `ShiftId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `TreatmentTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 5 | `WardId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `ClockInTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `OperatorType` | bigint |  |  |  | — 待补 |
| 8 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 9 | `CheckInId` | bigint |  |  |  | ID 外键（待核对） |
| 10 | `Note` | text |  |  |  | 备注 |
| 11 | `Content` | character varying(1024) |  |  |  | — 待补 |
| 12 | `AcceptContent` | character varying(1024) |  |  |  | — 待补 |
| 13 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 14 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 15 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Schedule_PatientShift`

- 字段数：13
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TreatmentTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 5 | `ShiftId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `WardId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 7 | `BedId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 8 | `PatientPlanId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 9 | `ShiftTiming` | integer |  | ✓ |  | — 待补 |
| 10 | `Status` | integer |  | ✓ |  | — 待补 |
| 11 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Schedule_Shift`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `Sort` | integer |  |  |  | — 待补 |
| 5 | `StartTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `EndTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `Type` | integer |  |  |  | — 待补 |
| 8 | `Note` | character varying(512) |  |  |  | 备注 |
| 9 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 10 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 11 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 12 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Schedule_Ward`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `Sort` | integer |  | ✓ |  | — 待补 |
| 5 | `PatientType` | character varying(64) |  |  |  | — 待补 |
| 6 | `InfectionType` | character varying(64) |  |  |  | — 待补 |
| 7 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 8 | `Note` | character varying(512) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `ResponsibleUsers` | character varying(512) |  |  |  | — 待补 |

## 医嘱（Order）

### `Order_OrderTPL`

- 字段数：18
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `OrderGroup` | bigint |  |  |  | — 待补 |
| 5 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 6 | `Classification` | character varying(64) |  |  |  | — 待补 |
| 7 | `DrugId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 8 | `Content` | character varying(256) |  |  |  | — 待补 |
| 9 | `Dosage` | character varying(64) |  |  |  | — 待补 |
| 10 | `UseOpportunity` | character varying(128) |  |  |  | — 待补 |
| 11 | `UseMethod` | character varying(128) |  |  |  | — 待补 |
| 12 | `UseWay` | character varying(128) |  |  |  | — 待补 |
| 13 | `Note` | character varying(1024) |  |  |  | 备注 |
| 14 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 15 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 16 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 17 | `UseNum` | numeric |  |  |  | — 待补 |
| 18 | `AllDosage` | numeric |  |  |  | — 待补 |

### `Order_PatientDayOrder`

- 字段数：26
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TreatmentTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 5 | `PatientOrderId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `OrderGroup` | bigint |  |  |  | — 待补 |
| 7 | `Status` | integer |  | ✓ |  | — 待补 |
| 8 | `CaseStatus` | character varying(64) |  |  |  | — 待补 |
| 9 | `Classification` | character varying(64) |  |  |  | — 待补 |
| 10 | `DrugId` | bigint |  |  |  | ID 外键（待核对） |
| 11 | `Content` | character varying(256) |  |  |  | — 待补 |
| 12 | `Dosage` | character varying(64) |  |  |  | — 待补 |
| 13 | `UseOpportunity` | character varying(128) |  |  |  | — 待补 |
| 14 | `UseMethod` | character varying(128) |  |  |  | — 待补 |
| 15 | `UseWay` | character varying(128) |  |  |  | — 待补 |
| 16 | `Note` | character varying(1024) |  |  |  | 备注 |
| 17 | `OperatorId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 18 | `CreateDept` | character varying(64) |  |  |  | — 待补 |
| 19 | `DealDept` | character varying(64) |  |  |  | — 待补 |
| 20 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 21 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 22 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 23 | `UseNum` | numeric |  |  |  | — 待补 |
| 24 | `DealOpportunity` | character varying(64) |  |  |  | — 待补 |
| 25 | `ChargeItemId` | bigint |  |  | 0 | ID 外键（待核对） |
| 26 | `AllDosage` | numeric |  |  |  | — 待补 |

### `Order_PatientDayOrderDeal`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientDayOrderId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `OperatorId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `Status` | integer |  | ✓ |  | — 待补 |
| 8 | `CaseStatus` | character varying(64) |  |  |  | — 待补 |
| 9 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |
| 10 | `Note` | character varying(1024) |  |  |  | 备注 |

### `Order_PatientOrder`

- 字段数：27
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `OrderTPLId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OrderGroup` | bigint |  |  |  | — 待补 |
| 6 | `Type` | integer |  |  |  | — 待补 |
| 7 | `DrugId` | bigint |  |  |  | ID 外键（待核对） |
| 8 | `Classification` | character varying(64) |  |  |  | — 待补 |
| 9 | `Content` | character varying(256) |  |  |  | — 待补 |
| 10 | `Dosage` | character varying(64) |  |  |  | — 待补 |
| 11 | `UseOpportunity` | character varying(128) |  |  |  | — 待补 |
| 12 | `UseMethod` | character varying(128) |  |  |  | — 待补 |
| 13 | `UseWay` | character varying(128) |  |  |  | — 待补 |
| 14 | `Note` | character varying(1024) |  |  |  | 备注 |
| 15 | `OperatorId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 16 | `StartTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 17 | `EndTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 18 | `CreateDept` | character varying(1024) |  |  |  | — 待补 |
| 19 | `DealDept` | character varying(1024) |  |  |  | — 待补 |
| 20 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 21 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 22 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 23 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 24 | `PatientPlanId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 25 | `UseNum` | numeric |  |  |  | — 待补 |
| 26 | `ChargeItemId` | bigint |  |  | 0 | ID 外键（待核对） |
| 27 | `AllDosage` | numeric |  |  |  | — 待补 |

## 治疗记录（Treatment）

### `Treatment_Action`

- 字段数：10
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `Name` | character varying(256) |  | ✓ |  | — 待补 |
| 5 | `OperatorId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 8 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 9 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 10 | `Code` | character varying(32) |  |  |  | — 待补 |

### `Treatment_AfterSigns`

- 字段数：19
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Weight` | numeric |  |  |  | — 待补 |
| 7 | `ExtraWeight` | numeric |  |  |  | — 待补 |
| 8 | `LossWeight` | numeric |  |  |  | — 待补 |
| 9 | `RealIntake` | numeric |  |  |  | — 待补 |
| 10 | `BodyTemp` | numeric |  |  |  | — 待补 |
| 11 | `SBP` | numeric |  |  |  | — 待补 |
| 12 | `DBP` | numeric |  |  |  | — 待补 |
| 13 | `PressurePoint` | character varying(64) |  |  |  | — 待补 |
| 14 | `HeartRate` | numeric |  |  |  | — 待补 |
| 15 | `Respiration` | numeric |  |  |  | — 待补 |
| 16 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 17 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 18 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 19 | `Note` | character varying(1024) |  |  |  | 备注 |

### `Treatment_AfterSymptom`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Code` | character varying(64) |  |  |  | — 待补 |
| 7 | `Value` | character varying(1024) |  |  |  | — 待补 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_Alarm`

- 字段数：13
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 3 | `Source` | integer |  | ✓ | 1 | — 待补 |
| 4 | `Code` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `Content` | text |  |  |  | — 待补 |
| 6 | `Levle` | integer |  | ✓ | 1 | — 待补 |
| 7 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 8 | `EndTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 9 | `HandleTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 10 | `HandleId` | bigint |  | ✓ | 0 | ID 外键（待核对） |
| 11 | `HandleContent` | text |  |  |  | — 待补 |
| 12 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 13 | `LastModifyTime` | timestamp |  | ✓ | `now()` | 最近修改时间 |

### `Treatment_BeforeCheck`

- 字段数：18
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `BeforeSignsId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `BeforeSymptomId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 7 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 8 | `MaterialsResult` | boolean |  |  |  | — 待补 |
| 9 | `MaterialsMistake` | character varying(1024) |  |  |  | — 待补 |
| 10 | `ParamResult` | boolean |  |  |  | — 待补 |
| 11 | `ParamMistake` | character varying(1024) |  |  |  | — 待补 |
| 12 | `VascularAccessResult` | boolean |  |  |  | — 待补 |
| 13 | `VascularAccessMistake` | character varying(1024) |  |  |  | — 待补 |
| 14 | `PipelineResult` | boolean |  |  |  | — 待补 |
| 15 | `PipelineMistake` | character varying(1024) |  |  |  | — 待补 |
| 16 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 17 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 18 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_BeforeSigns`

- 字段数：17
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Weight` | numeric |  |  |  | — 待补 |
| 7 | `ExtraWeight` | numeric |  |  |  | — 待补 |
| 8 | `BodyTemp` | numeric |  | ✓ |  | — 待补 |
| 9 | `SBP` | numeric |  | ✓ |  | — 待补 |
| 10 | `DBP` | numeric |  | ✓ |  | — 待补 |
| 11 | `PressurePoint` | character varying(64) |  | ✓ |  | — 待补 |
| 12 | `HeartRate` | numeric |  | ✓ |  | — 待补 |
| 13 | `Respiration` | numeric |  | ✓ |  | — 待补 |
| 14 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 15 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 16 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 17 | `Note` | character varying(1024) |  |  |  | 备注 |

### `Treatment_BeforeSymptom`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Code` | character varying(64) |  |  |  | — 待补 |
| 7 | `Value` | character varying(1024) |  |  |  | — 待补 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_DiseaseCourse`

- 字段数：13
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `DiseaseCourseTPLId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `Content` | text |  | ✓ |  | — 待补 |
| 7 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 8 | `Note` | character varying(1024) |  | ✓ |  | 备注 |
| 9 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 10 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 11 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_DuringOther`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Code` | character varying(64) |  |  |  | — 待补 |
| 7 | `Value` | character varying(1024) |  |  |  | — 待补 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_DuringParam`

- 字段数：22
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `TMP` | numeric |  |  |  | — 待补 |
| 7 | `UFQuantity` | numeric |  |  |  | — 待补 |
| 8 | `MachineTmp` | numeric |  |  |  | — 待补 |
| 9 | `VenousPressure` | numeric |  |  |  | — 待补 |
| 10 | `BF` | numeric |  |  |  | — 待补 |
| 11 | `Conductivity` | numeric |  |  |  | — 待补 |
| 12 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 13 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 14 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 15 | `ArterialPressure` | numeric |  |  |  | — 待补 |
| 16 | `SubstituateSpeed` | numeric |  |  |  | — 待补 |
| 17 | `HeparinPumpFlow` | numeric |  |  |  | — 待补 |
| 18 | `RelativeBloodVolume` | numeric |  |  |  | — 待补 |
| 19 | `RealBloodVolume` | numeric |  |  |  | — 待补 |
| 20 | `RealClearanceRate` | numeric |  |  |  | — 待补 |
| 21 | `ArterialBloodTemp` | numeric |  |  |  | — 待补 |
| 22 | `VenousBloodTemp` | numeric |  |  |  | — 待补 |

### `Treatment_DuringSigns`

- 字段数：14
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `SBP` | numeric |  |  |  | — 待补 |
| 7 | `DBP` | numeric |  |  |  | — 待补 |
| 8 | `HeartRate` | numeric |  |  |  | — 待补 |
| 9 | `BodyTemp` | numeric |  |  |  | — 待补 |
| 10 | `Respiration` | numeric |  |  |  | — 待补 |
| 11 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 14 | `SpO2` | numeric |  |  |  | — 待补 |

### `Treatment_DuringSymptom`

- 字段数：13
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `OperateTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Symptom` | character varying(1024) |  |  |  | — 待补 |
| 7 | `SymptomType` | character varying(1024) |  |  |  | — 待补 |
| 8 | `HandleDrId` | bigint |  |  |  | ID 外键（待核对） |
| 9 | `HandleContent` | character varying(1024) |  |  |  | — 待补 |
| 10 | `HandleResult` | character varying(1024) |  |  |  | — 待补 |
| 11 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_MaterialTrace`

- 字段数：12
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `ChargeItemId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 3 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `Num` | numeric |  |  |  | — 待补 |
| 5 | `Unit` | character varying(8) |  |  |  | — 待补 |
| 6 | `Batch` | character varying(64) |  |  |  | — 待补 |
| 7 | `SerialNo` | character varying(64) |  |  |  | — 待补 |
| 8 | `Note` | character varying(512) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `TenantId` | bigint |  |  |  | 租户 ID |
| 12 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Treatment_Others`

- 字段数：9
- 设计文档覆盖：—
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint |  | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `PredictWeight` | numeric |  |  |  | — 待补 |
| 6 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 7 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 9 | `PredictUFQ` | numeric |  |  |  | — 待补 |

### `Treatment_Treatment`

- 字段数：34
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `ScheduleId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `ReceptionDrId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `SignInTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 7 | `QueueNo` | character varying(32) |  |  |  | — 待补 |
| 8 | `ReceptionTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 9 | `DayProgrammeId` | bigint |  |  |  | ID 外键（待核对） |
| 10 | `WardId` | bigint |  |  |  | ID 外键（待核对） |
| 11 | `WardName` | character varying(256) |  |  |  | — 待补 |
| 12 | `BedId` | bigint |  |  |  | ID 外键（待核对） |
| 13 | `BedName` | character varying(256) |  |  |  | — 待补 |
| 14 | `EquipmentId` | character varying |  |  |  | ID 外键（待核对） |
| 15 | `EquipmentName` | character varying(512) |  |  |  | — 待补 |
| 16 | `StartTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 17 | `EndTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 18 | `RealDuration` | numeric |  |  |  | — 待补 |
| 19 | `RealUFQuantity` | numeric |  |  |  | — 待补 |
| 20 | `NurseSummary` | character varying(1024) |  | ✓ |  | — 待补 |
| 21 | `TreatmentSummary` | character varying(1024) |  | ✓ |  | — 待补 |
| 22 | `Status` | character varying(64) |  | ✓ |  | — 待补 |
| 23 | `CaseStatus` | character varying(64) |  | ✓ |  | — 待补 |
| 24 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 25 | `CreateTime` | timestamp |  |  |  | 创建时间 |
| 26 | `LastModifyTime` | timestamp |  |  |  | 最近修改时间 |
| 27 | `ShiftId` | bigint |  |  |  | ID 外键（待核对） |
| 28 | `ShiftName` | character varying(256) |  |  |  | — 待补 |
| 29 | `RealSubstituateVolume` | numeric |  |  |  | — 待补 |
| 30 | `TreatmentCount` | integer |  |  | 0 | — 待补 |
| 31 | `HospPatientType` | character varying(64) |  |  |  | — 待补 |
| 32 | `TmrPath` | character varying(1024) |  |  |  | — 待补 |
| 33 | `TmrTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 34 | `TmrPages` | integer |  |  | 0 | — 待补 |

### `Treatment_TreatmentMonthSummarySheet`

- 字段数：14
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Gender` | character varying(64) |  |  |  | — 待补 |
| 5 | `Age` | integer |  | ✓ |  | — 待补 |
| 6 | `TreatmentTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `Year` | integer |  | ✓ |  | — 待补 |
| 8 | `Month` | integer |  | ✓ |  | — 待补 |
| 9 | `ImageBase64String` | text |  |  |  | 图像 Base64 |
| 10 | `DigitalSignature` | text |  |  |  | — 待补 |
| 11 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 14 | `ContentJsonb` | jsonb |  |  |  | — 待补 |

## 基础数据 / 辅助资料（Auxiliary）

### `Auxiliary_AcquisiteConnect`

- 字段数：7
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(32) |  |  |  | — 待补 |
| 4 | `Devices` | character varying(128) |  |  |  | — 待补 |
| 5 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Auxiliary_DialysateGroup`

- 字段数：12
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `NaIonCon` | numeric |  |  |  | — 待补 |
| 5 | `CaIonCon` | numeric |  |  |  | — 待补 |
| 6 | `KIonCon` | numeric |  |  |  | — 待补 |
| 7 | `HCO3IonCon` | numeric |  |  |  | — 待补 |
| 8 | `GlucoseCon` | numeric |  |  |  | — 待补 |
| 9 | `Note` | character varying(1024) |  |  |  | 备注 |
| 10 | `Sort` | numeric |  |  |  | — 待补 |
| 11 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 12 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |

### `Auxiliary_DrugInfomation`

- 字段数：23
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `Classification` | character varying(64) |  |  |  | — 待补 |
| 5 | `Code` | character varying(128) |  |  |  | — 待补 |
| 6 | `Brand` | character varying(64) |  |  |  | — 待补 |
| 7 | `Specification` | character varying(64) |  |  |  | — 待补 |
| 8 | `Package` | character varying(64) |  |  |  | — 待补 |
| 9 | `Manufacturer` | character varying(128) |  |  |  | — 待补 |
| 10 | `Note` | character varying(512) |  |  |  | 备注 |
| 11 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 12 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 13 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 14 | `Spell` | character varying(256) |  |  |  | — 待补 |
| 15 | `BasicUnit` | character varying(64) |  |  |  | — 待补 |
| 16 | `SpecificationUnit` | character varying(64) |  |  |  | — 待补 |
| 17 | `Sort` | integer |  |  |  | — 待补 |
| 18 | `StdCat` | character varying(256) |  |  |  | — 待补 |
| 19 | `LastModifyTime` | timestamp |  | ✓ | `'2023-02-24 12:31:28.753248+08'::timestampwithtimezone` | 最近修改时间 |
| 20 | `ShortName` | character varying(128) |  |  | `''::charactervarying` | — 待补 |
| 21 | `UseTips` | character varying(1024) |  |  |  | — 待补 |
| 22 | `MinUnitDosage` | bigint |  |  |  | — 待补 |
| 23 | `UseOpportunity` | character varying(64) |  |  |  | — 待补 |

### `Auxiliary_EquipmentDisinfection`

- 字段数：16
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `EquipmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `DisinfectUserId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `DisinfectWay` | character varying(256) |  |  |  | — 待补 |
| 6 | `StartTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 7 | `Description` | character varying(1024) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 13 | `Status` | integer |  |  |  | — 待补 |
| 14 | `EndTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 15 | `Type` | character varying(64) |  |  |  | — 待补 |
| 16 | `Disinfectant` | character varying(64) |  |  |  | — 待补 |

### `Auxiliary_EquipmentInfomation`

- 字段数：20
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Name` | character varying(256) |  | ✓ |  | — 待补 |
| 4 | `IDNo` | character varying(128) |  |  |  | — 待补 |
| 5 | `SerialNo` | character varying(128) |  |  |  | — 待补 |
| 6 | `Brand` | character varying(128) |  |  |  | — 待补 |
| 7 | `ModelNo` | character varying(128) |  |  |  | — 待补 |
| 8 | `DialysisMethod` | character varying(512) |  |  |  | — 待补 |
| 9 | `Type` | character varying |  |  |  | — 待补 |
| 10 | `ManufactureDate` | timestamp |  |  |  | — 待补 |
| 11 | `Manufacturer` | character varying(128) |  |  |  | — 待补 |
| 12 | `InstallDate` | timestamp |  |  |  | — 待补 |
| 13 | `Maintenance` | bigint |  | ✓ |  | — 待补 |
| 14 | `MaintenanceCycle` | character varying(64) |  |  |  | — 待补 |
| 15 | `Note` | character varying(1024) |  |  |  | 备注 |
| 16 | `IsDisabled` | boolean |  |  | false | 是否禁用/软删除 |
| 17 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 18 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 19 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 20 | `Flux` | character varying(64) |  |  |  | — 待补 |

### `Auxiliary_FepServiceConfig`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `ServiceName` | character varying(32) |  |  |  | — 待补 |
| 4 | `Device` | character varying(32) |  |  |  | — 待补 |
| 5 | `Interval` | integer |  | ✓ |  | — 待补 |
| 6 | `UploadRowCount` | integer |  | ✓ |  | — 待补 |
| 7 | `Version` | character varying(16) |  | ✓ |  | — 待补 |
| 8 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Auxiliary_HdGoods`

- 字段数：10
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `ChargeItemId` | bigint |  |  |  | ID 外键（待核对） |
| 3 | `Type` | bigint |  |  |  | — 待补 |
| 4 | `Sort` | bigint |  |  |  | — 待补 |
| 5 | `Name` | character varying(256) |  |  |  | — 待补 |
| 6 | `Code` | character varying(256) |  |  |  | — 待补 |
| 7 | `Spell` | character varying(64) |  |  |  | — 待补 |
| 8 | `BasicUnit` | character varying(32) |  |  |  | — 待补 |
| 9 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 10 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |

### `Auxiliary_HealthEducation`

- 字段数：13
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `Sort` | numeric |  |  |  | — 待补 |
| 4 | `Name` | character varying(1024) |  | ✓ |  | — 待补 |
| 5 | `Description` | text |  |  |  | — 待补 |
| 6 | `Type` | character varying(64) |  |  |  | — 待补 |
| 7 | `AttachmentIds` | character varying(64) |  |  |  | — 待补 |
| 8 | `Note` | character varying(1024) |  |  |  | 备注 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 13 | `Classify` | character varying(64) |  |  |  | — 待补 |

### `Auxiliary_JsonData`

- 字段数：9
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `Code` | character varying(64) |  | ✓ |  | — 待补 |
| 6 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 7 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 9 | `Value` | jsonb |  |  |  | — 待补 |

### `Auxiliary_MaterialInfomation`

- 字段数：20
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `Name` | character varying(256) |  |  |  | — 待补 |
| 4 | `Spell` | character varying(256) |  |  |  | — 待补 |
| 5 | `Classification` | character varying(64) |  |  |  | — 待补 |
| 6 | `Code` | character varying(128) |  |  |  | — 待补 |
| 7 | `Brand` | character varying(64) |  |  |  | — 待补 |
| 8 | `Specification` | character varying(64) |  |  |  | — 待补 |
| 9 | `Package` | character varying(64) |  |  |  | — 待补 |
| 10 | `Manufacturer` | character varying(128) |  |  |  | — 待补 |
| 11 | `Note` | character varying(512) |  |  |  | 备注 |
| 12 | `Type` | character varying(64) |  |  |  | — 待补 |
| 13 | `IsDisabled` | boolean |  |  |  | 是否禁用/软删除 |
| 14 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 15 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 16 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 17 | `StdCat` | character varying(256) |  |  |  | — 待补 |
| 18 | `Unit` | character varying(64) |  | ✓ | `''::charactervarying` | — 待补 |
| 19 | `ShortName` | character varying(128) |  |  | `''::charactervarying` | — 待补 |
| 20 | `Sort` | numeric |  |  |  | — 待补 |

### `Auxiliary_PatientHealthEducation`

- 字段数：15
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `HealthEducationId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `OperatorId` | bigint |  |  |  | ID 外键（待核对） |
| 6 | `EducationTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 7 | `EducationType` | character varying(256) |  |  |  | — 待补 |
| 8 | `EducationResult` | character varying(512) |  |  |  | — 待补 |
| 9 | `NurseSign` | character varying(32) |  |  |  | — 待补 |
| 10 | `PatientSign` | character varying(32) |  |  |  | — 待补 |
| 11 | `FinishTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 12 | `Note` | character varying(1024) |  |  |  | 备注 |
| 13 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 14 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 15 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

## 检验接口（LIS）

### `LIS_Examination`

- 字段数：10
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `Name` | character varying(128) |  |  |  | — 待补 |
| 5 | `Type` | character varying(64) |  |  |  | — 待补 |
| 6 | `ResultTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 7 | `SyncUserId` | bigint |  |  |  | ID 外键（待核对） |
| 8 | `SyncTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |
| 10 | `TestNO` | character varying(64) |  |  |  | — 待补 |

### `LIS_ExaminationApply`

- 字段数：22
- 设计文档覆盖：✅
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint |  | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `TestNO` | character varying(64) |  | ✓ |  | — 待补 |
| 5 | `Priority` | integer |  |  |  | — 待补 |
| 6 | `TestCase` | character varying(64) |  |  |  | — 待补 |
| 7 | `ClinicalDiagnosisDesc` | character varying(1024) |  |  |  | — 待补 |
| 8 | `Specimen` | character varying(64) |  |  |  | — 待补 |
| 9 | `SpecimenDesc` | character varying(64) |  |  |  | — 待补 |
| 10 | `SpecimenReceivedTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 11 | `SpecimenSampleTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 12 | `ApplyTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 13 | `ApplyDept` | character varying(64) |  |  |  | — 待补 |
| 14 | `ApplyUserName` | character varying(64) |  |  |  | — 待补 |
| 15 | `DealDept` | character varying(64) |  |  |  | — 待补 |
| 16 | `ResultStatus` | integer |  |  |  | — 待补 |
| 17 | `ResultRPTTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 18 | `ResultRPTUser` | character varying(64) |  |  |  | — 待补 |
| 19 | `ResultVerifyUser` | character varying(64) |  |  |  | — 待补 |
| 20 | `PrintCount` | integer |  |  |  | — 待补 |
| 21 | `ContainerCode` | character varying(64) |  |  |  | — 待补 |
| 22 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `LIS_ExaminationItem`

- 字段数：10
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `ExaminationId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `ItemName` | character varying(64) |  |  |  | — 待补 |
| 5 | `ItemCode` | character varying(64) |  |  |  | — 待补 |
| 6 | `Result` | character varying(64) |  |  |  | — 待补 |
| 7 | `Unit` | character varying(32) |  |  |  | — 待补 |
| 8 | `Reference` | character varying(64) |  |  |  | — 待补 |
| 9 | `ResultSign` | character varying(16) |  |  |  | — 待补 |
| 10 | `LastModifyTime` | timestamp(6) |  | ✓ | `now()` | 最近修改时间 |

### `LIS_ExaminationItem_Config`

- 字段数：10
- 设计文档覆盖：—
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint |  | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `RetItemName` | character varying(64) |  |  |  | — 待补 |
| 4 | `Spell` | character varying(256) |  |  |  | — 待补 |
| 5 | `RetExaminationName` | character varying(128) |  |  |  | — 待补 |
| 6 | `Unit` | character varying(32) |  |  |  | — 待补 |
| 7 | `Reference` | character varying(64) |  |  |  | — 待补 |
| 8 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `LIS_ExaminationItem_Ret`

- 字段数：13
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  |  |  | 租户 ID |
| 3 | `ItemName` | character varying(64) |  |  |  | — 待补 |
| 4 | `ItemCode` | character varying(64) |  |  |  | — 待补 |
| 5 | `ExaminationName` | character varying(128) |  |  |  | — 待补 |
| 6 | `ExaminationType` | character varying(64) |  |  |  | — 待补 |
| 7 | `RetItemName` | character varying(64) |  |  |  | — 待补 |
| 8 | `RetExaminationName` | character varying(128) |  |  |  | — 待补 |
| 9 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `Sort` | integer |  |  |  | — 待补 |
| 13 | `ItemSort` | integer |  |  |  | — 待补 |

## 设备日志（Device）

### `Device_AlarmInfo`

- 字段数：9
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `DeviceId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `AlarmCode` | integer |  | ✓ |  | — 待补 |
| 6 | `AlarmType` | integer |  | ✓ |  | — 待补 |
| 7 | `AlarmTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 8 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 9 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Device_BODYLog`

- 字段数：12
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MDC_LEN_BODY_ACTUAL` | numeric |  |  |  | — 待补 |
| 10 | `MDC_MASS_BODY_ACTUAL` | numeric |  |  |  | — 待补 |
| 11 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |
| 12 | `RawData` | jsonb |  |  |  | — 待补 |

### `Device_BPHRLog`

- 字段数：15
- 设计文档覆盖：—
- 主键：`Id`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `DeviceSerialNo` | character varying(64) |  | ✓ |  | — 待补 |
| 3 | `MDC_PRESS_BLD_ART_ABP_SYS` | numeric |  |  |  | — 待补 |
| 4 | `MDC_PRESS_BLD_ART_ABP_MEAN` | numeric |  |  |  | — 待补 |
| 5 | `MDC_PRESS_BLD_ART_ABP_DIA` | numeric |  |  |  | — 待补 |
| 6 | `MDC_ECG_CARD_BEAT_RATE` | numeric |  |  |  | — 待补 |
| 7 | `FactoryNo` | character varying(64) |  |  |  | — 待补 |
| 8 | `DeviceType` | character varying(64) |  |  |  | — 待补 |
| 9 | `DeviceModel` | character varying(64) |  |  |  | — 待补 |
| 10 | `SIMCardNO` | character varying(64) |  |  |  | — 待补 |
| 11 | `PacketType` | integer |  |  |  | — 待补 |
| 12 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 13 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 14 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 15 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

### `Device_BPLog`

- 字段数：13
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MDC_PRESS_BLD_ART_ABP_SYS` | numeric |  |  |  | — 待补 |
| 10 | `MDC_PRESS_BLD_ART_ABP_MEAN` | numeric |  |  |  | — 待补 |
| 11 | `MDC_PRESS_BLD_ART_ABP_DIA` | numeric |  |  |  | — 待补 |
| 12 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |
| 13 | `BP_Location` | character varying(32) |  |  |  | — 待补 |

### `Device_DMLog`

- 字段数：34
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `TMP` | numeric |  |  |  | — 待补 |
| 10 | `UFVolume` | numeric |  |  |  | — 待补 |
| 11 | `VenousPressure` | numeric |  |  |  | — 待补 |
| 12 | `ArterialPressure` | numeric |  |  |  | — 待补 |
| 13 | `BF` | numeric |  |  |  | — 待补 |
| 14 | `Conductivity` | numeric |  |  |  | — 待补 |
| 15 | `APumpSpeedDeviation` | numeric |  |  |  | — 待补 |
| 16 | `BPumpSpeedDeviation` | numeric |  |  |  | — 待补 |
| 17 | `HeparinPumpFlow` | numeric |  |  |  | — 待补 |
| 18 | `DialysateInFlow` | numeric |  |  |  | — 待补 |
| 19 | `DialysateOutFlow` | numeric |  |  |  | — 待补 |
| 20 | `AConductivity` | numeric |  |  |  | — 待补 |
| 21 | `DialysateTemp` | numeric |  |  |  | — 待补 |
| 22 | `TreatmentTime` | numeric |  |  |  | 时间字段（待补语义） |
| 23 | `UFSetVolume` | numeric |  |  |  | — 待补 |
| 24 | `UFQuantity` | numeric |  |  |  | — 待补 |
| 25 | `BConductivity` | numeric |  |  |  | — 待补 |
| 26 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |
| 27 | `SubstituateVolume` | numeric |  |  |  | — 待补 |
| 28 | `HeparinVolume` | numeric |  |  |  | — 待补 |
| 29 | `SubstituateSpeed` | numeric |  |  |  | — 待补 |
| 30 | `RelativeBloodVolume` | numeric |  |  |  | — 待补 |
| 31 | `RealBloodVolume` | numeric |  |  |  | — 待补 |
| 32 | `RealClearanceRate` | numeric |  |  |  | — 待补 |
| 33 | `ArterialBloodTemp` | numeric |  |  |  | — 待补 |
| 34 | `VenousBloodTemp` | numeric |  |  |  | — 待补 |

### `Device_FEPLog`

- 字段数：13
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `FEPId` | bigint |  |  |  | ID 外键（待核对） |
| 3 | `LogTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 4 | `CPUTemp` | numeric(8,3) |  |  |  | — 待补 |
| 5 | `CPULoad` | numeric(8,3) |  |  |  | — 待补 |
| 6 | `MemLoad` | numeric(8,3) |  |  |  | — 待补 |
| 7 | `FSLoad` | numeric(8,3) |  |  |  | — 待补 |
| 8 | `TenantId` | bigint |  |  |  | 租户 ID |
| 9 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 10 | `TreatmentId` | bigint |  |  |  | ID 外键（待核对） |
| 11 | `CreateTime` | timestamp |  |  |  | 创建时间 |
| 12 | `LastModifyTime` | timestamp |  |  |  | 最近修改时间 |
| 13 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

### `Device_HRLog`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MDC_ECG_CARD_BEAT_RATE` | numeric |  |  |  | — 待补 |
| 10 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

### `Device_PULSLog`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MDC_PULS_OXIM_PULS_RATE` | numeric |  |  |  | — 待补 |
| 10 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

### `Device_RESPLog`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MNDRY_BOA_RESP_SCORE` | numeric |  |  |  | — 待补 |
| 10 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

### `Device_SPO2HLog`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MDC_PULS_OXIM_SAT_O2` | numeric |  |  |  | — 待补 |
| 10 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

### `Device_TEMPLog`

- 字段数：11
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `MDC_TEMP` | numeric |  |  |  | — 待补 |
| 10 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |
| 11 | `TEMP_Location` | character varying(32) |  |  |  | — 待补 |

### `Device_UREALog`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `FEPId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 5 | `TreatmentId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `LogTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 9 | `Concentration` | numeric |  |  |  | — 待补 |
| 10 | `DeviceId` | integer |  | ✓ |  | ID 外键（待核对） |

## 库存（Stock）

### `Stock_ChargeItem`

- 字段数：15
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `Name` | character varying(64) |  | ✓ |  | — 待补 |
| 3 | `CatalogId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `HisItemId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `Price` | numeric |  |  |  | — 待补 |
| 6 | `CatalogType` | integer |  | ✓ |  | — 待补 |
| 7 | `Note` | text |  |  |  | 备注 |
| 8 | `CreatorId` | bigint |  | ✓ |  | 创建人 ID |
| 9 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 10 | `IsDisabled` | boolean |  | ✓ |  | 是否禁用/软删除 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 13 | `Unit` | character varying(64) |  |  |  | — 待补 |
| 14 | `Spell` | character varying(256) |  |  |  | — 待补 |
| 15 | `Sort` | bigint |  |  | 10000 | — 待补 |

### `Stock_ChargeItemDetail`

- 字段数：9
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ | `nextval('public."Stock_ ChargeItemDetail_Id_seq"'::regclass)` | 主键 ID（snowflake / bigint） |
| 2 | `ChargeItemId` | bigint |  |  |  | ID 外键（待核对） |
| 3 | `ChargeItemDetailId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `Num` | numeric |  |  |  | — 待补 |
| 5 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 9 | `TenantId` | bigint |  | ✓ |  | 租户 ID |

### `Stock_InOutStorage`

- 字段数：12
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `BillNo` | character varying(64) |  |  |  | — 待补 |
| 3 | `BillType` | bigint |  |  |  | — 待补 |
| 4 | `CategoryNum` | bigint |  |  |  | — 待补 |
| 5 | `TotalNum` | bigint |  |  |  | — 待补 |
| 6 | `Note` | character varying(128) |  |  |  | 备注 |
| 7 | `HandlerId` | bigint |  |  |  | ID 外键（待核对） |
| 8 | `Status` | bigint |  |  |  | — 待补 |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `TenantId` | bigint |  | ✓ |  | 租户 ID |

### `Stock_InOutStorageDetail`

- 字段数：12
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `InOutStorageId` | bigint |  |  |  | ID 外键（待核对） |
| 3 | `ChargeItemId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `Num` | numeric |  |  |  | — 待补 |
| 5 | `BeginNum` | numeric |  |  |  | — 待补 |
| 6 | `Batch` | character varying(64) |  | ✓ | `''::charactervarying` | — 待补 |
| 7 | `SerialNo` | text |  |  |  | — 待补 |
| 8 | `StorageId` | bigint |  |  |  | ID 外键（待核对） |
| 9 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 10 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 11 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 12 | `TenantId` | bigint |  | ✓ |  | 租户 ID |

### `Stock_Stock`

- 字段数：8
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `ChargeItemId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 3 | `Batch` | character varying(64) |  | ✓ | `''::charactervarying` | — 待补 |
| 4 | `Price` | numeric |  | ✓ | 0 | — 待补 |
| 5 | `Num` | numeric |  |  |  | — 待补 |
| 6 | `StorageId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `TenantId` | bigint |  | ✓ |  | 租户 ID |

### `Stock_Storage`

- 字段数：12
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `Position` | character varying(256) |  |  |  | — 待补 |
| 3 | `Contacts` | character varying(64) |  |  |  | — 待补 |
| 4 | `Telephone` | character varying(64) |  |  |  | — 待补 |
| 5 | `Sort` | bigint |  |  |  | — 待补 |
| 6 | `Note` | character varying(256) |  |  |  | 备注 |
| 7 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 8 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 9 | `IsDisabled` | boolean |  | ✓ | false | 是否禁用/软删除 |
| 10 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 11 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 12 | `Name` | character varying |  |  |  | — 待补 |

## 费用（Cost）

### `Cost_PatientBalance`

- 字段数：5
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `PatientId` | bigint |  | ✓ |  | 患者 ID → `Register_PatientInfomation.Id` |
| 3 | `Balance` | numeric |  |  |  | — 待补 |
| 4 | `TenantId` | bigint |  |  |  | 租户 ID |
| 5 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

### `Cost_PatientBillFlow`

- 字段数：15
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `BillNo` | character varying(64) |  |  |  | — 待补 |
| 3 | `BillType` | bigint |  |  |  | — 待补 |
| 4 | `SourceBillId` | bigint |  |  |  | ID 外键（待核对） |
| 5 | `PatientId` | bigint |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 6 | `Sum` | numeric |  |  |  | — 待补 |
| 7 | `LastBalance` | numeric |  |  |  | — 待补 |
| 8 | `Balance` | numeric |  |  |  | — 待补 |
| 9 | `Note` | character varying(128) |  |  |  | 备注 |
| 10 | `SettleStatus` | bigint |  |  |  | — 待补 |
| 11 | `SettleTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 12 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 13 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 14 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 15 | `TenantId` | bigint |  | ✓ |  | 租户 ID |

### `Cost_PatientBillFlowDetail`

- 字段数：9
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `PatientBillFlowId` | bigint |  |  |  | ID 外键（待核对） |
| 3 | `ChargeItemId` | bigint |  |  |  | ID 外键（待核对） |
| 4 | `Num` | numeric |  |  |  | — 待补 |
| 5 | `Price` | numeric |  |  |  | — 待补 |
| 6 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 7 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 8 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 9 | `TenantId` | bigint |  | ✓ |  | 租户 ID |

## 质控评估（QualityEvaluation）

### `QualityEvaluation_Record`

- 字段数：17
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | numeric |  |  |  | 租户 ID |
| 3 | `PatientId` | numeric |  |  |  | 患者 ID → `Register_PatientInfomation.Id` |
| 4 | `IndexName` | character varying(64) |  |  |  | — 待补 |
| 5 | `IndexCode` | character varying(64) |  |  |  | — 待补 |
| 6 | `Type` | bigint |  |  |  | — 待补 |
| 7 | `Result` | character varying(64) |  |  |  | — 待补 |
| 8 | `Unit` | character varying(32) |  |  |  | — 待补 |
| 9 | `Reference` | character varying(64) |  |  |  | — 待补 |
| 10 | `ResultSign` | character varying(16) |  |  |  | — 待补 |
| 11 | `TestTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 12 | `EvaluationResult` | character varying(16) |  |  |  | — 待补 |
| 13 | `ItemScore` | bigint |  |  |  | — 待补 |
| 14 | `EvaluationClass` | character varying(16) |  |  |  | — 待补 |
| 15 | `CreatorId` | numeric(32,0) |  |  |  | 创建人 ID |
| 16 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 17 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |

## 通知（Notify）

### `Notify_Data`

- 字段数：15
- 设计文档覆盖：✅
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `Title` | character varying(512) |  | ✓ |  | — 待补 |
| 3 | `Content` | jsonb |  |  |  | — 待补 |
| 4 | `AppCode` | integer |  |  |  | — 待补 |
| 5 | `BizObjId` | character varying(32) |  |  |  | ID 外键（待核对） |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `CreateUserId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 8 | `HandleUserId` | bigint |  |  |  | ID 外键（待核对） |
| 9 | `HandleTime` | timestamp |  |  |  | 时间字段（待补语义） |
| 10 | `HandleResult` | text |  |  |  | — 待补 |
| 11 | `IsCompleted` | boolean |  | ✓ | false | — 待补 |
| 12 | `Scene` | character varying(1024) |  |  |  | — 待补 |
| 13 | `SessionId` | character varying(32) |  | ✓ |  | ID 外键（待核对） |
| 14 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 15 | `LastModifyTime` | timestamp |  | ✓ | `now()` | 最近修改时间 |

### `Notify_Data_ReadRecord`

- 字段数：8
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `DataId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 3 | `ReadTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 4 | `ReaderId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `Status` | integer |  | ✓ | 0 | — 待补 |
| 6 | `Note` | text |  |  |  | 备注 |
| 7 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 8 | `LastModifyTime` | timestamp |  | ✓ | `now()` | 最近修改时间 |

## 消息（Message）

### `Message_Messages`

- 字段数：9
- 设计文档覆盖：—
- 主键：`Id`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `Type` | text |  |  |  | — 待补 |
| 3 | `Receiver` | text |  |  |  | — 待补 |
| 4 | `Content` | text |  |  |  | — 待补 |
| 5 | `Time` | timestamp(6) |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Status` | text |  |  |  | — 待补 |
| 7 | `ReturnUrl` | text |  |  |  | — 待补 |
| 8 | `HandleUrl` | text |  |  |  | — 待补 |
| 9 | `SessionId` | text |  |  |  | ID 外键（待核对） |

## 留言板（MessageBoard）

### `MessageBoard_MessageAndReply`

- 字段数：10
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 3 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 4 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 5 | `UserId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 6 | `Content` | text |  | ✓ |  | — 待补 |
| 7 | `ProcessRole` | integer |  | ✓ |  | — 待补 |
| 8 | `MessageId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 9 | `ThreadId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 10 | `ReadStatus` | boolean |  | ✓ |  | — 待补 |

## 系统日志（Log）

### `Log_Logs`

- 字段数：17
- 设计文档覆盖：—
- 主键：`Id`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `LogName` | character varying(128) |  |  |  | — 待补 |
| 3 | `Level` | character varying(64) |  |  |  | — 待补 |
| 4 | `TraceId` | character varying(64) |  |  |  | ID 外键（待核对） |
| 5 | `OperationTime` | timestamp |  | ✓ |  | 时间字段（待补语义） |
| 6 | `Duration` | character varying(64) |  |  |  | — 待补 |
| 7 | `Ip` | character varying(64) |  |  |  | — 待补 |
| 8 | `Host` | character varying(64) |  |  |  | — 待补 |
| 9 | `ThreadId` | character varying(64) |  |  |  | ID 外键（待核对） |
| 10 | `Browser` | character varying(4096) |  |  |  | — 待补 |
| 11 | `Url` | character varying(4096) |  |  |  | — 待补 |
| 12 | `UserId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 13 | `Content` | jsonb |  |  |  | — 待补 |
| 14 | `Exception` | jsonb |  |  |  | — 待补 |
| 15 | `Signature` | text |  |  |  | — 待补 |
| 16 | `AreaId` | bigint |  | ✓ | 0 | ID 外键（待核对） |
| 17 | `OrganId` | bigint |  | ✓ | 0 | ID 外键（待核对） |

## 应用配置（Applications）

### `Applications_PromptConfigure`

- 字段数：18
- 设计文档覆盖：—
- 主键：`Id`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `AppCode` | character varying(128) |  | ✓ |  | — 待补 |
| 2 | `OrganId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 3 | `PromptClass` | character varying(32) |  | ✓ |  | — 待补 |
| 4 | `Code` | character varying(32) |  | ✓ |  | — 待补 |
| 5 | `Name` | character varying(64) |  | ✓ |  | — 待补 |
| 6 | `Scenes` | text |  |  |  | — 待补 |
| 7 | `EnableTimedTask` | boolean |  | ✓ |  | — 待补 |
| 8 | `TaskInterval` | bigint |  | ✓ |  | — 待补 |
| 9 | `SendWay` | character varying(32) |  |  |  | — 待补 |
| 10 | `SendInterval` | bigint |  | ✓ |  | — 待补 |
| 11 | `Execute` | text |  | ✓ |  | — 待补 |
| 12 | `ExecuteType` | character varying(32) |  | ✓ |  | — 待补 |
| 13 | `PermissionCodes` | character varying(1024) |  |  |  | — 待补 |
| 14 | `RoleIds` | character varying(1024) |  |  |  | — 待补 |
| 15 | `Form` | character varying(64) |  |  |  | — 待补 |
| 16 | `Sort` | integer |  | ✓ | 0 | — 待补 |
| 17 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 18 | `LastModifyTime` | timestamp |  |  |  | 最近修改时间 |

## 代码字典（CodeDictionary）

### `CodeDictionary_CodeDictionarys`

- 字段数：7
- 设计文档覆盖：—

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Code` | character varying(64) |  | ✓ |  | — 待补 |
| 2 | `Type` | character varying(64) |  | ✓ |  | — 待补 |
| 3 | `Name` | character varying(64) |  | ✓ |  | — 待补 |
| 4 | `OrganId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 5 | `IsDisabled` | boolean |  | ✓ |  | 是否禁用/软删除 |
| 6 | `Sort` | integer |  | ✓ |  | — 待补 |
| 7 | `Builtin` | boolean |  | ✓ | false | — 待补 |

## 租户配置（TenantConfig）

### `TenantConfig`

- 字段数：12
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | integer | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `InternetAddress` | character varying(2048) |  | ✓ |  | — 待补 |
| 4 | `InternetPort` | integer |  | ✓ |  | — 待补 |
| 5 | `IntranetAddress` | character varying(2048) |  | ✓ |  | — 待补 |
| 6 | `IntranetPort` | integer |  | ✓ |  | — 待补 |
| 7 | `ClientId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 8 | `TenantName` | character varying(512) |  | ✓ | `'医院'::charactervarying` | — 待补 |
| 9 | `Longitude` | numeric |  | ✓ | 0, note: '经度' | — 待补 |
| 10 | `Latitude` | numeric |  | ✓ | 0, note: '纬度' | — 待补 |
| 11 | `Enable2faTime` | timestamp |  | ✓ | `'9999-12-31 00:00:00+08'::timestampwithtimezone` | 时间字段（待补语义） |
| 12 | `LoginTypes` | character varying(64) |  | ✓ | `'pwd'::charactervarying` | — 待补 |

## 用户（User）

### `User_Image`

- 字段数：8
- 设计文档覆盖：—
- 主键：`Id`
- 外键候选（来自 ER 图 `*` 标记）：`TenantId`

| # | 字段 | 类型 | 主键 | 非空 | 默认值 | 业务含义（待补） |
|---|------|------|------|------|--------|------------------|
| 1 | `Id` | bigint | ✓ | ✓ |  | 主键 ID（snowflake / bigint） |
| 2 | `TenantId` | bigint |  | ✓ |  | 租户 ID |
| 3 | `UserId` | bigint |  | ✓ |  | ID 外键（待核对） |
| 4 | `ImageBase64String` | text |  |  |  | 图像 Base64 |
| 5 | `CreatorId` | bigint |  |  |  | 创建人 ID |
| 6 | `CreateTime` | timestamp |  | ✓ |  | 创建时间 |
| 7 | `LastModifyTime` | timestamp |  | ✓ |  | 最近修改时间 |
| 8 | `Type` | integer |  |  |  | — 待补 |

## 附录 A：设计 PDF 覆盖但 md 未收录的表

- `Register_Comorbidity`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Register_InformedConsent`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Schedule_PatientShiftTPL`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_Puncture`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_DiseaseCourseTPL`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_Nursing`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_FeedBack`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_ShiftHandover`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_BeforeSymptomJsonData`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_DuringOtherJsonData`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Treatment_AfterSymptomJsonData`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Auxiliary_VascularAccessMaterial`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Auxiliary_EquipmentUsageLog`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Auxiliary_EquipmentMaintenance`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Auxiliary`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Auxiliary_DataTPL`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `JsonAuxiliary_JsonData`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Report_Reports`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `LIS_ExaminationItem_Tr`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `Idint8`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `ItemIdint8Id`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `LastTreatmentIdint8ID`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `NextTreatmentIdint8ID`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `LastIntervalint8`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `NextIntervalint8`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `TenantIdint8ID`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `CreatorIdint8ID`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `CreateTimedatetime`（仅见于设计 PDF，字段需从 PDF 人工提取）
- `User_JsonData`（仅见于设计 PDF，字段需从 PDF 人工提取）

## 附录 B：md 收录但设计 PDF 未覆盖的表

- `Register_Protopathy`
- `Register_VascularAccessImage`
- `Applications_PromptConfigure`
- `Auxiliary_AcquisiteConnect`
- `Auxiliary_FepServiceConfig`
- `Auxiliary_HealthEducation`
- `Auxiliary_JsonData`
- `CodeDictionary_CodeDictionarys`
- `Cost_PatientBalance`
- `Cost_PatientBillFlow`
- `Cost_PatientBillFlowDetail`
- `Device_AlarmInfo`
- `Device_BODYLog`
- `Device_BPHRLog`
- `Device_BPLog`
- `Device_DMLog`
- `Device_FEPLog`
- `Device_HRLog`
- `Device_PULSLog`
- `Device_RESPLog`
- `Device_SPO2HLog`
- `Device_TEMPLog`
- `Device_UREALog`
- `LIS_ExaminationItem_Config`
- `Log_Logs`
- `MessageBoard_MessageAndReply`
- `Message_Messages`
- `Notify_Data_ReadRecord`
- `Order_PatientDayOrder`
- `Order_PatientDayOrderDeal`
- `Order_PatientOrder`
- `Plan_PatientPlanPrescriptionAdjustment`
- `Schedule_BedFEPChange`
- `Schedule_CheckIn`
- `Stock_ChargeItem`
- `Stock_ChargeItemDetail`
- `Stock_InOutStorage`
- `Stock_InOutStorageDetail`
- `Stock_Stock`
- `Stock_Storage`
- `TenantConfig`
- `Treatment_AfterSigns`
- `Treatment_AfterSymptom`
- `Treatment_Alarm`
- `Treatment_BeforeSigns`
- `Treatment_BeforeSymptom`
- `Treatment_DuringOther`
- `Treatment_DuringSigns`
- `Treatment_MaterialTrace`
- `Treatment_Others`
- `User_Image`

## 附录 C：中文业务注释回填建议

PDF 中文字形未嵌入 ToUnicode CMap，`pdftotext` 提取后字段注释列全部为空。建议回填方式（任选其一）：

1. 使用 OCR（`tesseract` + 中文简体模型）对 `数据库表设计.pdf` 每页扫描，结合字段序号回填。
2. 将 PDF 导出为图片后人工补注。
3. 对接老系统仓库或字典表（如 `CodeDictionary_CodeDictionarys`）反查业务枚举值含义。
