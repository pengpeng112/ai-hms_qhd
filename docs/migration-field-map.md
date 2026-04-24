# 新血透 → 老血透 · 字段级迁移对照表

> 本表为 `docs/migration-plan-legacy.md` 的附录。每张新表列出字段级映射，Codex 据此重写 GORM 模型与 SQL。

## 映射类别

| 类别 | 含义 | 典型操作 |
|------|------|---------|
| `rename` | 仅表名不同，字段相近 | 改 `TableName()`、按映射补 `gorm:"column:..."` |
| `rename-fields` | 表名改且字段名差异较多 | 上 + 逐字段 column tag |
| `same-name` | 表名相同 | 核对字段类型与命名差异 |
| `rewrite` | 结构差异大（JSONB ↔ 扁平列） | 重写 service 读写逻辑 + 可能增加领域对象 |
| `rewrite+child` | 重写 + 拆出子表 | rewrite + 额外处理子表 CRUD |
| `fold` | 老库由多张表拼成新表 | service 层 join 或多次查询组装 |
| `split-to-many` | 新表被老库拆成多张专项表 | service 层分散写入/聚合读取 |
| `multi-join` | 新表通过老库多表 join | 已有实现 / 查询构造器 |
| `fold-to-parent` | 子表合并到父表（JSON 或多行） | 视情况处理 |
| `app-only` | 老库无对应，应用层自管 | **不改 TableName，保留新表**（若老库存在该表则不冲突） |

## 迁移类别汇总

| # | 新表 | 老表 | 类别 | 说明 |
|---|------|------|------|------|
| 1 | `users` | `—` | `app-only` | 老库无独立用户表，旧系统使用 Identity/Organ 体系；保留应用层表 |
| 2 | `patients` | `Register_PatientInfomation` | `rename` | 老库合并了档案扩展，字段最丰富的主表 |
| 3 | `patient_basic_infos` | `Register_PatientInfomation + Register_Hospitalization + Register_IDInfomation + Register_FamilyMember` | `fold` | 新表为独立 1:1 扩展，老库字段分散在 4 张 Register 表 |
| 4 | `medical_histories` | `Register_MedicalHistory + Register_Allergen + Register_Complication + Register_Diagnosis + Register_Pathology + Register_Protopathy + Register_Tumor` | `split-to-many` | 新表一张扁平 33 列；老库为 7 张专项表，每类病史一张 |
| 5 | `infection_infos` | `Register_Infection` | `rename-text-parse` | 老表 InfectionDesc/OtherDesc/Note 是自由文本，当前已在 patient_core_service 中做关键字解析 |
| 6 | `vascular_accesses` | `Register_VascularAccess` | `rename` | 已完成 TableName 切换，字段差异待对齐 |
| 7 | `vascular_access_interventions` | `Register_VascularAccessChange` | `rename` | 已完成 TableName 切换 |
| 8 | `outcome_records` | `Register_OutCome` | `rename` | 已完成 TableName 切换 |
| 9 | `hospitalizations` | `Register_Hospitalization` | `rename` | 字段几乎一一对应 |
| 10 | `treatment_plans` | `Plan_PatientPlan` | `rewrite` | 新表 4 列 JSONB（dialysisMode/anticoagulant/parameters/materials） vs 老表扁平多列；重点改造 |
| 11 | `orders` | `Order_PatientOrder` | `rename-fields` | patient_core_service.buildActiveOrders 已部分使用；字段别名与状态字段需统一 |
| 12 | `prescriptions` | `Plan_PatientPrescription + Plan_PatientPrescriptionMaterial` | `rewrite+child` | 新表 materials/orderItems 为 JSONB；老库 material 为子表 |
| 13 | `adjustment_records` | `Plan_PatientPlanPrescriptionAdjustment` | `rename` | 字段少，简单改名 |
| 14 | `wards` | `Schedule_Ward` | `rename` |  |
| 15 | `beds` | `Schedule_Bed` | `rename` |  |
| 16 | `shifts` | `Schedule_Shift` | `rename` |  |
| 17 | `patient_shifts` | `Schedule_PatientShift` | `rename` |  |
| 18 | `Treatment_Treatment` | `Treatment_Treatment` | `same-name` | 同名表，字段需核对 |
| 19 | `Treatment_BeforeCheck` | `Treatment_BeforeCheck` | `same-name` | 同名表 |
| 20 | `Treatment_BeforeSigns` | `Treatment_BeforeSigns` | `same-name` | 同名表 |
| 21 | `Treatment_DuringParam` | `Treatment_DuringParam` | `same-name` | 同名表 |
| 22 | `Treatment_AfterSigns` | `Treatment_AfterSigns` | `same-name` | 同名表 |
| 23 | `Treatment_Alarm` | `Treatment_Alarm` | `same-name` | 同名表 |
| 24 | `plan_templates` | `Plan_PlanTPL + Plan_PlanTPLMaterial` | `rewrite+child` | 新表 templateContent 为 JSONB，老库为父表 + 材料子表 |
| 25 | `material_catalogs` | `Auxiliary_MaterialInfomation` | `rename` |  |
| 26 | `drug_catalogs` | `Auxiliary_DrugInfomation` | `rename` |  |
| 27 | `order_templates` | `Order_OrderTPL` | `rename` | 老库可能仅一张表包含模板；需确认是否含子条目 |
| 28 | `order_template_items` | `Order_OrderTPL (同表)` | `fold-to-parent` | 需确认：若老库无子表，items 以 JSON/多行形式存在父表 |
| 29 | `dict_types` | `CodeDictionary_CodeDictionarys` | `rewrite` | 新表为 type/item 两表；老库单表树形（parent + category 分类） |
| 30 | `dict_items` | `CodeDictionary_CodeDictionarys` | `rewrite` | 同上 |
| 31 | `lab_reports` | `LIS_Examination` | `rename-fields` | 已部分在 patient_core_service.buildLabTrends 使用 |
| 32 | `lab_report_items` | `LIS_ExaminationItem` | `rename-fields` | 同上 |
| 33 | `exam_reports` | `—` | `app-only+sync` | 老库无本地表，数据来自 HDIS 同步；保留新表 |
| 34 | `patient_key_indicators` | `—` | `app-only+sync` | 老库无；来自 HDIS Record 同步；保留新表 |
| 35 | `integration_hdis_settings` | `—` | `app-only` | 新系统配置表，与老库无关 |
| 36 | `permissions` | `—` | `app-only` | 权限定义，应用层 |
| 37 | `role_permissions` | `—` | `app-only` | 角色-权限关联 |
| 38 | `clinical_tasks` | `—` | `app-only?` | 待评估：老库可能无对应；如纯应用层则保留 |
| 39 | `devices` | `Auxiliary_EquipmentInfomation + Schedule_BedEquipmentRel + Schedule_Bed + Schedule_Ward` | `multi-join` | 已由 device_service 通过多表 join 实现；model.TableName 仍为 devices，需修正 |
| 40 | `inventory_items` | `Stock_Stock + Stock_Storage` | `rewrite` | 老库库存拆为 Stock_Stock（总账） + Stock_Storage（仓库） |
| 41 | `stock_logs` | `Stock_InOutStorage + Stock_InOutStorageDetail` | `rewrite` | 出入库单 + 明细 |
| 42 | `label_tasks` | `—` | `app-only` | 条码标签任务，应用层 |

## `users` → `—（应用层保留）`

**类别：** `app-only`  
**说明：** 老库无独立用户表，旧系统使用 Identity/Organ 体系；保留应用层表

> 🟢 **此表保留为应用层独立表，无需迁移。**

## `patients` → `Register_PatientInfomation`

**类别：** `rename`  
**说明：** 老库合并了档案扩展，字段最丰富的主表

**新表字段数：** 18  **老表（主） `Register_PatientInfomation` 字段数：** 42

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `name` | varchar(50) | `Name` | character varying(256) | synonym | ✅ 患者姓名 |
| `age` | int | — |  | 老库无对应（需业务决策） | ❌ TODO 年龄 |
| `gender` | varchar(10) | `Gender` | character varying(64) | synonym | ✅ M / F (ISO 5218) |
| `bed_number` | varchar(20) | — |  | 老库无对应（需业务决策） | ❌ TODO 床位号 |
| `diagnosis` | text | `DiagnosisDesc` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 诊断 |
| `risk_level` | varchar(20) | — |  | 老库无对应（需业务决策） | ❌ TODO 高危 / 中危 / 低危 |
| `status` | varchar(20) | `Status` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 active / inactive / discharged |
| `patient_type` | varchar(50) | `PatientType` | character varying(64) | synonym | ✅ 门诊 / 住院 |
| `insurance_type` | varchar(50) | `ExpenseType` | character varying(64) | synonym | ✅ 医保类型 |
| `dry_weight` | decimal(5,2) | `PredictWeight` | numeric | synonym | ✅ 干体重 (kg) |
| `default_mode` | varchar(50) | `DialysisMethod` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 默认透析模式 |
| `doctor_id` | varchar(36) | `ResponsibilityDrId` | bigint | synonym | ✅ 主管医生 ID |
| `doctor_name` | varchar(50) | `AttendDr` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 主管医生姓名 |
| `admission_date` | timestamp | `FirstDialysisDate` | timestamp | synonym | ✅ 入院日期 |
| `discharge_date` | timestamp | — |  | 老库无对应（需业务决策） | ❌ TODO 出院日期 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_PatientInfomation` 有 32 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Spell` | character varying(256) | — |
| `Type` | character varying(64) | — |
| `TreatmentStatus` | character varying(64) | — |
| `OutComeStatus` | character varying(64) | — |
| `BirthDate` | timestamp | — |
| `Nation` | character varying(64) | — |
| `ABOType` | character varying(64) | — |
| `RHType` | character varying(64) | — |
| `Height` | numeric | — |
| `Weight` | numeric | — |
| `Occupation` | character varying(128) | — |
| `MaritalStatus` | character varying(64) | — |
| `EducationLevel` | character varying(64) | — |
| `Province` | character varying(64) | — |
| `City` | character varying(64) | — |
| `County` | character varying(64) | — |
| `Address` | character varying(256) | — |
| `PhoneNo` | character varying(32) | — |
| `SSN` | character varying(64) | — |
| `DialysisNo` | character varying(64) | — |
| `ResponsibilityNurseId` | bigint | — |
| `FirstDialysisHospital` | character varying(64) | — |
| `Note` | character varying(1024) | — |
| `WeChatNo` | character varying(128) | — |
| `ImportInfo` | character varying(1024) | — |
| `CreatorId` | bigint | — |
| `IDName` | character varying(256) | — |
| `HomePhoneNo` | character varying(32) | — |
| `ImageBase64String` | text | — |
| `Workunit` | character varying(256) | — |
| `OurHospitalFirstDialysisDate` | timestamp | — |

</details>

## `patient_basic_infos` → `Register_PatientInfomation + Register_Hospitalization + Register_IDInfomation + Register_FamilyMember`

**类别：** `fold`  
**说明：** 新表为独立 1:1 扩展，老库字段分散在 4 张 Register 表

**新表字段数：** 34  **老表（主） `Register_PatientInfomation` 字段数：** 42

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 关联 patients.id |
| `pinyin` | varchar(100) | `Spell` | character varying(256) | synonym | ✅ 姓名拼音 |
| `birthday` | timestamp | `BirthDate` | timestamp | synonym | ✅ 出生日期 |
| `ethnicity` | varchar(20) | `Nation` | character varying(64) | synonym | ✅ 民族 |
| `id_type` | varchar(20) | `IDType` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 身份证 / 护照 / 其他 |
| `id_number` | varchar(50) | `IDNo` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 证件号码 |
| `visit_category` | varchar(20) | — |  | 无匹配 | ❌ TODO 门诊 / 住院 / 急诊 |
| `admission_no` | varchar(50) | — |  | 无匹配 | ❌ TODO 住院号 |
| `visit_no` | varchar(50) | — |  | 无匹配 | ❌ TODO 就诊号 |
| `medical_record_no` | varchar(50) | `MedicalRecordNo` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 病历号 |
| `insurance_no` | varchar(50) | `SSN` | character varying(64) | synonym | ✅ 医保号 |
| `hdis_patient_id` | int | — |  | 无匹配 | ❌ TODO HDIS/LIS 外部系统 ID |
| `dialysis_no` | varchar(50) | `DialysisNo` | character varying(64) | synonym | ✅ 透析号 |
| `nurse_name` | varchar(50) | `ResponsibilityNurseId` | bigint | synonym | ✅ 责任护士 |
| `first_dialysis_date` | timestamp | `FirstDialysisDate` | timestamp | synonym | ✅ 首次透析日期 |
| `first_hospital_date` | timestamp | `OurHospitalFirstDialysisDate` | timestamp | synonym | ✅ 首次在本院透析日期 |
| `first_dialysis_hospital` | varchar(100) | `FirstDialysisHospital` | character varying(64) | synonym | ✅ 首次透析医院 |
| `height` | varchar(10) | `Height` | numeric | synonym | ✅ 身高 (cm) |
| `abo_blood_type` | varchar(10) | `ABOType` | character varying(64) | synonym | ✅ A / B / AB / O |
| `rh_blood_type` | varchar(10) | `RHType` | character varying(64) | synonym | ✅ Rh+ / Rh- |
| `education_level` | varchar(20) | `EducationLevel` | character varying(64) | synonym | ✅ 文化程度 |
| `occupation` | varchar(50) | `Occupation` | character varying(128) | synonym | ✅ 职业 |
| `marital_status` | varchar(20) | `MaritalStatus` | character varying(64) | synonym | ✅ 婚姻状况 |
| `workplace` | varchar(100) | `Workunit` | character varying(256) | synonym | ✅ 工作单位 |
| `phone` | varchar(20) | `PhoneNo` | character varying(32) | synonym | ✅ 手机号码 |
| `wechat` | varchar(50) | `WeChatNo` | character varying(128) | synonym | ✅ 微信号 |
| `landline` | varchar(20) | `HomePhoneNo` | character varying(32) | synonym | ✅ 固定电话 |
| `address` | text | `Address` | character varying(256) | synonym | ✅ 地址 |
| `district` | varchar(100) | — |  | 无匹配 | ❌ TODO 区域 |
| `contact_name` | varchar(50) | — |  | 无匹配 | ❌ TODO 紧急联系人 |
| `contact_phone` | varchar(20) | — |  | 无匹配 | ❌ TODO 紧急联系电话 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_PatientInfomation` 有 19 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Name` | character varying(256) | — |
| `Type` | character varying(64) | — |
| `TreatmentStatus` | character varying(64) | — |
| `OutComeStatus` | character varying(64) | — |
| `Gender` | character varying(64) | — |
| `Weight` | numeric | — |
| `Province` | character varying(64) | — |
| `City` | character varying(64) | — |
| `County` | character varying(64) | — |
| `ExpenseType` | character varying(64) | — |
| `ResponsibilityDrId` | bigint | — |
| `Note` | character varying(1024) | — |
| `ImportInfo` | character varying(1024) | — |
| `CreatorId` | bigint | — |
| `IDName` | character varying(256) | — |
| `ImageBase64String` | text | — |
| `PatientType` | character varying(64) | — |
| `PredictWeight` | numeric | — |

</details>

## `medical_histories` → `Register_MedicalHistory + Register_Allergen + Register_Complication + Register_Diagnosis + Register_Pathology + Register_Protopathy + Register_Tumor`

**类别：** `split-to-many`  
**说明：** 新表一张扁平 33 列；老库为 7 张专项表，每类病史一张

**新表字段数：** 35  **老表（主） `Register_MedicalHistory` 字段数：** 18

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `current_illness` | text | — |  | 无匹配 | ❌ TODO 现病史 |
| `past_history` | text | — |  | 无匹配 | ❌ TODO 既往史 |
| `transfusion_history` | text | — |  | 无匹配 | ❌ TODO 输血史 |
| `marital_history` | text | — |  | 无匹配 | ❌ TODO 婚育史 |
| `family_history` | text | `FamilyHistory` | text | pascal-exact | ✅ 家族史 |
| `disease_diagnosis` | text | — |  | 无匹配 | ❌ TODO 疾病诊断 |
| `primary_disease_name` | varchar(255) | — |  | 无匹配 | ❌ TODO 原发病名称 |
| `primary_disease_content` | text | — |  | 无匹配 | ❌ TODO 原发病详情 |
| `primary_disease_type` | varchar(255) | — |  | 无匹配 | ❌ TODO 原发病分类 |
| `primary_disease_check_time` | varchar(32) | — |  | 无匹配 | ❌ TODO 原发病检查时间 |
| `primary_disease_check_doc` | varchar(100) | — |  | 无匹配 | ❌ TODO 原发病检查医生 |
| `pathology_name` | varchar(255) | — |  | 无匹配 | ❌ TODO 病理诊断名称 |
| `pathology_content` | text | — |  | 无匹配 | ❌ TODO 病理诊断详情 |
| `pathology_type` | varchar(255) | — |  | 无匹配 | ❌ TODO 病理诊断分类 |
| `pathology_check_time` | varchar(32) | — |  | 无匹配 | ❌ TODO 病理检查时间 |
| `pathology_check_doc` | varchar(100) | — |  | 无匹配 | ❌ TODO 病理检查医生 |
| `allergen_name` | varchar(255) | — |  | 无匹配 | ❌ TODO 过敏信息名称 |
| `allergen_content` | text | — |  | 无匹配 | ❌ TODO 过敏信息详情 |
| `allergen_type` | varchar(255) | — |  | 无匹配 | ❌ TODO 过敏原分类 |
| `allergen_check_time` | varchar(32) | — |  | 无匹配 | ❌ TODO 过敏检查时间 |
| `allergen_check_doc` | varchar(100) | — |  | 无匹配 | ❌ TODO 过敏检查医生 |
| `tumor_history_name` | varchar(255) | — |  | 无匹配 | ❌ TODO 肿瘤病史名称 |
| `tumor_history_content` | text | — |  | 无匹配 | ❌ TODO 肿瘤病史详情 |
| `tumor_history_type` | varchar(255) | — |  | 无匹配 | ❌ TODO 肿瘤分类 |
| `tumor_history_check_time` | varchar(32) | — |  | 无匹配 | ❌ TODO 肿瘤检查时间 |
| `tumor_history_check_doc` | varchar(100) | — |  | 无匹配 | ❌ TODO 肿瘤检查医生 |
| `complication_name` | varchar(255) | — |  | 无匹配 | ❌ TODO 并发症名称 |
| `complication_content` | text | — |  | 无匹配 | ❌ TODO 并发症详情 |
| `complication_type` | varchar(255) | — |  | 无匹配 | ❌ TODO 并发症分类 |
| `complication_check_time` | varchar(32) | — |  | 无匹配 | ❌ TODO 并发症检查时间 |
| `complication_check_doc` | varchar(100) | — |  | 无匹配 | ❌ TODO 并发症检查医生 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_MedicalHistory` 有 13 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Complaints` | text | — |
| `PresentIllnessHistory` | text | — |
| `PastIllnessHistory` | text | — |
| `PersonalHistory` | text | — |
| `MaritalReproductiveHistory` | text | — |
| `DiagnosisDesc` | text | — |
| `PhysicalExamination` | text | — |
| `SpecialistExamination` | text | — |
| `AncillaryExamination` | text | — |
| `Note` | character varying(1024) | — |
| `CreatorId` | bigint | — |
| `Narrator` | character varying(64) | — |

</details>

## `infection_infos` → `Register_Infection`

**类别：** `rename-text-parse`  
**说明：** 老表 InfectionDesc/OtherDesc/Note 是自由文本，当前已在 patient_core_service 中做关键字解析

**新表字段数：** 10  **老表（主） `Register_Infection` 字段数：** 9

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `hbsag` | varchar(10) | — |  | 无匹配 | ❌ TODO 乙肝表面抗原 |
| `hcvab` | varchar(10) | — |  | 无匹配 | ❌ TODO 丙肝抗体 |
| `hivab` | varchar(10) | — |  | 无匹配 | ❌ TODO HIV 抗体 |
| `tpab` | varchar(10) | — |  | 无匹配 | ❌ TODO 梅毒抗体 |
| `tb` | varchar(10) | — |  | 无匹配 | ❌ TODO 结核 |
| `update_date` | timestamp | — |  | 无匹配 | ❌ TODO 检测更新日期 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_Infection` 有 5 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `InfectionDesc` | character varying(1024) | — |
| `OtherDesc` | character varying(1024) | — |
| `Note` | character varying(1024) | — |
| `CreatorId` | bigint | — |

</details>

## `vascular_accesses` → `Register_VascularAccess`

**类别：** `rename`  
**说明：** 已完成 TableName 切换，字段差异待对齐

**新表字段数：** 24  **老表（主） `Register_VascularAccess` 字段数：** 26

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `access_type` | varchar(50) | `AccessType` | character varying(64) | pascal-exact | ✅ AVF / AVG / TCC / NCC |
| `site` | varchar(100) | — |  | 无匹配 | ❌ TODO 通路部位 |
| `artery` | text (JSON) | `Artery` | character varying(128) | pascal-exact | ✅ 动脉（JSON 字符串数组） |
| `vein` | text (JSON) | — |  | 无匹配 | ❌ TODO 静脉（JSON 字符串数组） |
| `side` | varchar(10) | — |  | 无匹配 | ❌ TODO L / R |
| `hospital` | varchar(200) | — |  | 无匹配 | ❌ TODO 手术医院 |
| `surgeon` | varchar(100) | — |  | 无匹配 | ❌ TODO 手术医生 |
| `surgery_date` | timestamp | — |  | 无匹配 | ❌ TODO 手术时间 |
| `first_use_date` | timestamp | — |  | 无匹配 | ❌ TODO 首次使用时间 |
| `access_number` | int | — |  | 无匹配 | ❌ TODO 第几次血管通路 |
| `intervention_count` | int | — |  | 无匹配 | ❌ TODO 干预次数 |
| `intervention_date` | timestamp | — |  | 无匹配 | ❌ TODO 干预日期 |
| `catheter_method` | varchar(50) | — |  | 无匹配 | ❌ TODO 置管方法（导管） |
| `catheter_depth` | varchar(20) | `CatheterDepth` | numeric | pascal-exact | ✅ 导管深度 |
| `v_puncture_position` | text (JSON) | — |  | 无匹配 | ❌ TODO V侧穿刺位置（JSON 数组） |
| `a_puncture_position` | text (JSON) | — |  | 无匹配 | ❌ TODO A侧穿刺位置（JSON 数组） |
| `notes` | text | `Note` | character varying(1024) | synonym | ✅ 备注 |
| `images` | text (JSON) | — |  | 无匹配 | ❌ TODO 图片 URLs（JSON 数组） |
| `is_default` | bool | `IsDefault` | boolean | pascal-exact | ✅ 是否默认 |
| `is_disabled` | bool | `IsDisabled` | boolean | synonym | ✅ 是否禁用 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_VascularAccess` 有 16 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `AccessPosition` | character varying(128) | — |
| `Venous` | character varying(128) | — |
| `LeftAndRight` | character varying(128) | — |
| `OperationTime` | timestamp | — |
| `PictureIds` | bigint | — |
| `CreatorId` | bigint | — |
| `ASidePointCount` | character varying(256) | — |
| `OperationHospital` | character varying(128) | — |
| `VSidePointCount` | character varying(256) | — |
| `CatheterizeMethod` | character varying(128) | — |
| `FirstUseTime` | timestamp | — |
| `InterveneCount` | bigint | — |
| `AccessCount` | bigint | — |
| `InterveneTime` | timestamp | — |
| `OperationDr` | text | — |

</details>

## `vascular_access_interventions` → `Register_VascularAccessChange`

**类别：** `rename`  
**说明：** 已完成 TableName 切换

**新表字段数：** 13  **老表（主） `Register_VascularAccessChange` 字段数：** 14

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `vascular_access_id` | varchar(36) | `VascularAccessId` | bigint | pascal-exact | ✅ 关联 vascular_accesses.id |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 冗余患者 ID（便于查询） |
| `access_type` | varchar(50) | — |  | 无匹配 | ❌ TODO 通路类型（冗余） |
| `avg_blood_flow` | int | — |  | 无匹配 | ❌ TODO 平均血流量 |
| `usage_days` | int | — |  | 无匹配 | ❌ TODO 使用天数 |
| `surgery_type` | varchar(50) | — |  | 无匹配 | ❌ TODO 手术类型 |
| `intervention_reason` | text | — |  | 无匹配 | ❌ TODO 干预原因 |
| `doctor` | varchar(50) | — |  | 无匹配 | ❌ TODO 干预医生 |
| `intervention_date` | timestamp | — |  | 无匹配 | ❌ TODO 干预时间 |
| `description` | text | — |  | 无匹配 | ❌ TODO 干预描述 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_VascularAccessChange` 有 9 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `UseDuration` | integer | — |
| `AvgBF` | numeric | — |
| `ChangeTime` | timestamp | — |
| `ChangeReason` | character varying(512) | — |
| `ChangeDesc` | character varying(512) | — |
| `SketchMap` | bigint | — |
| `PhysicalMap` | bigint | — |
| `CreatorId` | bigint | — |

</details>

## `outcome_records` → `Register_OutCome`

**类别：** `rename`  
**说明：** 已完成 TableName 切换

**新表字段数：** 11  **老表（主） `Register_OutCome` 字段数：** 10

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `type` | varchar(20) | `Type` | character varying(64) | synonym | ✅ 转入 / 转出 |
| `reason` | varchar(255) | `Reason` | character varying(64) | pascal-exact | ✅ 原因 |
| `time` | timestamp | — |  | 无匹配 | ❌ TODO 转归时间 |
| `remarks` | text | — |  | 无匹配 | ❌ TODO 备注 |
| `registrar` | varchar(50) | — |  | 无匹配 | ❌ TODO 登记人 |
| `registration_time` | timestamp | — |  | 无匹配 | ❌ TODO 登记时间 |
| `is_door_rule` | bool | — |  | 无匹配 | ❌ TODO 是否门规 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_OutCome` 有 4 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `OutComeTime` | timestamp | — |
| `Note` | character varying(1024) | — |
| `CreatorId` | bigint | — |

</details>

## `hospitalizations` → `Register_Hospitalization`

**类别：** `rename`  
**说明：** 字段几乎一一对应

**新表字段数：** 19  **老表（主） `Register_Hospitalization` 字段数：** 16

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅ 租户 ID |
| `patient_id` | bigint | `PatientId` | bigint | synonym | ✅ 关联患者 |
| `case_no` | varchar(64) | `CaseNo` | character varying(64) | pascal-exact | ✅ 病案号 |
| `hosp_no` | varchar(64) | `HospNo` | character varying(64) | pascal-exact | ✅ 住院号 |
| `bar_code` | varchar(64) | `BarCode` | character varying(64) | pascal-exact | ✅ 条码 |
| `hosp_patient_type` | varchar(64) | `HospPatientType` | character varying(64) | pascal-exact | ✅ 门诊 / 住院 / 急诊 |
| `hosp_receive_dept` | varchar(64) | `HospReceiveDept` | character varying(64) | pascal-exact | ✅ 接收科室 |
| `hosp_ward` | varchar(64) | `HospWard` | character varying(64) | pascal-exact | ✅ 病房 |
| `hosp_bed` | varchar(64) | `HospBed` | character varying(64) | pascal-exact | ✅ 床位 |
| `attend_dr` | varchar(64) | `AttendDr` | character varying(64) | pascal-exact | ✅ 主治医生 |
| `reception_dr` | varchar(64) | `ReceptionDr` | character varying(64) | pascal-exact | ✅ 接诊医生 |
| `status` | int | `Status` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 1=在院, 0=出院 |
| `admission_date` | timestamp | `FirstDialysisDate` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 入院日期 |
| `discharge_date` | timestamp | — |  | 老库无对应（需业务决策） | ❌ TODO 出院日期 |
| `notes` | text | `Note` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 备注 |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅ 创建人 |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Register_Hospitalization` 有 1 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `MedicalRecordNo` | character varying(64) | — |

</details>

## `treatment_plans` → `Plan_PatientPlan`

**类别：** `rewrite`  
**说明：** 新表 4 列 JSONB（dialysisMode/anticoagulant/parameters/materials） vs 老表扁平多列；重点改造

**新表字段数：** 18  **老表（主） `Plan_PatientPlan` 字段数：** 48

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `weekly_frequency` | int | `OddWeekFrequency` | integer | synonym | ✅ 每周透析次数 |
| `biweekly_frequency` | int | `EvenWeekFrequency` | integer | synonym | ✅ 双周透析次数 |
| `duration` | int | `DialysisDuration` | numeric | synonym | ✅ 透析时长（小时） |
| `dry_weight` | decimal(5,2) | `PredictWeight` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 干体重 (kg) |
| `extra_weight` | decimal(5,2) | `ExtraWeight` | numeric | pascal-exact | ✅ 额外体重 (kg) |
| `status` | varchar(20) | `Status` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 启用 / 禁用 |
| `doctor_id` | varchar(36) | `ResponsibilityDrId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 开立医生 ID |
| `start_date` | timestamp | — |  | 无匹配 | ❌ TODO 方案开始日期 |
| `end_date` | timestamp | — |  | 无匹配 | ❌ TODO 方案结束日期 |
| `notes` | text | `Note` | text | synonym | ✅ 备注 |
| `dialysis_mode` | **jsonb** | — |  | 无匹配 | ❌ TODO 透析模式（嵌套对象） |
| `anticoagulant` | **jsonb** | — |  | 无匹配 | ❌ TODO 抗凝方案（嵌套对象） |
| `parameters` | **jsonb** | — |  | 无匹配 | ❌ TODO 透析参数（嵌套对象） |
| `materials` | **jsonb** | — |  | 无匹配 | ❌ TODO 材料清单（数组） |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Plan_PatientPlan` 有 39 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Name` | character varying(256) | — |
| `CreatorId` | bigint | — |
| `PlanTPLId` | bigint | — |
| `DialysisMethod` | character varying(256) | — |
| `DryWeight` | numeric | — |
| `AdjustQuantity` | numeric | — |
| `BF` | numeric | — |
| `BV` | numeric | — |
| `FirstAnticoagulant` | bigint | — |
| `FirstDosage` | numeric | — |
| `MaintainAnticoagulant` | bigint | — |
| `DilutionProportion` | numeric | — |
| `InjectionRate` | numeric | — |
| `InjectionDuration` | numeric | — |
| `InjectionVolume` | numeric | — |
| `VascularAccessId` | bigint | — |
| `Dialysate` | character varying(64) | — |
| `DialysateFlow` | numeric | — |
| `DialysateVolume` | numeric | — |
| `NaIonCon` | numeric | — |
| `CaIonCon` | numeric | — |
| `KIonCon` | numeric | — |
| `HCO3IonCon` | numeric | — |
| `Conductivity` | numeric | — |
| `DialysateTmp` | numeric | — |
| `SubstituateVolume` | numeric | — |
| `DilutionMnt` | character varying(64) | — |
| `IsDisabled` | boolean | — |
| `SalineQuantity` | numeric | — |
| `SealQuantity` | numeric | — |
| `ArterialQuantity` | numeric | — |
| `VenousQuantity` | numeric | — |
| `SealType` | character varying(64) | — |
| `Frequency` | character varying(128) | — |
| `GlucoseCon` | numeric | — |
| `DialysateGroupId` | bigint | — |
| `AutoConfirmPrescription` | character varying(64) | — |
| `SubstituateFlow` | numeric | — |

</details>

## `orders` → `Order_PatientOrder`

**类别：** `rename-fields`  
**说明：** patient_core_service.buildActiveOrders 已部分使用；字段别名与状态字段需统一

**新表字段数：** 27  **老表（主） `Order_PatientOrder` 字段数：** 27

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `type` | varchar(20) | `Type` | integer | synonym | ✅ 长期 / 临时 |
| `category` | varchar(50) | `Classification` | character varying(64) | synonym | ✅ 药品 / 检查 / 治疗 / 护理 / 饮食 |
| `name` | varchar(100) | `Name` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 医嘱名称 |
| `content` | text | `Content` | character varying(256) | synonym | ✅ 医嘱内容 |
| `dose` | varchar(50) | `Dose` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 剂量 |
| `unit` | varchar(20) | `Unit` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 单位 |
| `route` | varchar(50) | `Route` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 用法（静脉注射、口服等） |
| `timing` | varchar(50) | `Timing` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 使用时机 |
| `exec_timing` | varchar(50) | — |  | 无匹配 | ❌ TODO 执行时机 |
| `drug_id` | uint | `DrugId` | bigint | pascal-exact | ✅ 关联 drug_catalogs.id（可选） |
| `spec` | varchar(100) | — |  | 无匹配 | ❌ TODO 规格 |
| `group_id` | varchar(36) | — |  | 无匹配 | ❌ TODO 医嘱组号（同组医嘱联合执行） |
| `doctor_id` | varchar(36) | `ResponsibilityDrId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 开单医生 ID |
| `doctor_name` | varchar(50) | `AttendDr` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 开单医生姓名 |
| `status` | varchar(20) | `Status` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 待执行 / 执行中 / 已执行 / 已停止 |
| `start_time` | timestamp | `StartTime` | timestamp | synonym | ✅ 医嘱开始时间 |
| `end_time` | timestamp | `EndTime` | timestamp | synonym | ✅ 医嘱结束时间（临时医嘱必填） |
| `frequency` | varchar(50) | `Frequency` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 频次（qd/bid/tid/qod 等） |
| `priority` | varchar(20) | `Priority` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 普通 / 紧急 / 临急 |
| `notes` | text | `Note` | character varying(1024) | synonym | ✅ 备注 |
| `executed_at` | timestamp | `ExecuteTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 执行时间 |
| `executed_by` | varchar(36) | `ExecutorId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 执行人 ID |
| `stop_reason` | text | `StopReason` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 停止原因 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Order_PatientOrder` 有 16 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `OrderTPLId` | bigint | — |
| `OrderGroup` | bigint | — |
| `Dosage` | character varying(64) | — |
| `UseOpportunity` | character varying(128) | — |
| `UseMethod` | character varying(128) | — |
| `UseWay` | character varying(128) | — |
| `OperatorId` | bigint | — |
| `CreateDept` | character varying(1024) | — |
| `DealDept` | character varying(1024) | — |
| `IsDisabled` | boolean | — |
| `CreatorId` | bigint | — |
| `PatientPlanId` | bigint | — |
| `UseNum` | numeric | — |
| `ChargeItemId` | bigint | — |
| `AllDosage` | numeric | — |

</details>

## `prescriptions` → `Plan_PatientPrescription + Plan_PatientPrescriptionMaterial`

**类别：** `rewrite+child`  
**说明：** 新表 materials/orderItems 为 JSONB；老库 material 为子表

**新表字段数：** 20  **老表（主） `Plan_PatientPrescription` 字段数：** 49

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅ 关联 patients.id |
| `treatment_plan_id` | varchar(36) | — |  | 无匹配 | ❌ TODO 关联 treatment_plans.id |
| `prescription_date` | timestamp | — |  | 无匹配 | ❌ TODO 处方日期 |
| `doctor_id` | varchar(36) | `ResponsibilityDrId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 开方医生 ID |
| `doctor_name` | varchar(50) | `AttendDr` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 开方医生姓名 |
| `status` | varchar(20) | `Status` | integer | synonym | ✅ 待执行 / 执行中 / 已执行 / 已取消 |
| `duration` | int | `DialysisDuration` | numeric | synonym | ✅ 透析时长（小时） |
| `dry_weight` | decimal(5,2) | `PredictWeight` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 干体重 |
| `extra_weight` | decimal(5,2) | — |  | 无匹配 | ❌ TODO 额外体重 |
| `dialysis_mode` | **jsonb** | — |  | 无匹配 | ❌ TODO 透析模式（同 TreatmentPlan） |
| `anticoagulant` | **jsonb** | — |  | 无匹配 | ❌ TODO 抗凝方案 |
| `parameters` | **jsonb** | — |  | 无匹配 | ❌ TODO 透析参数 |
| `materials` | **jsonb** | — |  | 无匹配 | ❌ TODO 材料清单 |
| `order_items` | **jsonb** | — |  | 无匹配 | ❌ TODO 药品明细快照 |
| `notes` | text | `Note` | text | synonym | ✅ 备注 |
| `executed_at` | timestamp | `ExecuteTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 执行时间 |
| `executed_by` | varchar(36) | `ExecutorId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 执行人 ID |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Plan_PatientPrescription` 有 42 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `TreatmentId` | bigint | — |
| `PatientPlanId` | bigint | — |
| `CreatorId` | bigint | — |
| `ConfirmUserId` | bigint | — |
| `ConfirmTime` | timestamp | — |
| `CaseStatus` | character varying(64) | — |
| `DialysisMethod` | character varying(256) | — |
| `DryWeight` | numeric | — |
| `AdjustQuantity` | numeric | — |
| `BF` | numeric | — |
| `BV` | numeric | — |
| `FirstAnticoagulant` | bigint | — |
| `FirstDosage` | numeric | — |
| `MaintainAnticoagulant` | bigint | — |
| `DilutionProportion` | numeric | — |
| `InjectionRate` | numeric | — |
| `InjectionDuration` | numeric | — |
| `InjectionVolume` | numeric | — |
| `VascularAccessId` | bigint | — |
| `Dialysate` | character varying(64) | — |
| `DialysateFlow` | numeric | — |
| `DialysateVolume` | numeric | — |
| `NaIonCon` | numeric | — |
| `CaIonCon` | numeric | — |
| `KIonCon` | numeric | — |
| `HCO3IonCon` | numeric | — |
| `Conductivity` | numeric | — |
| `DialysateTmp` | numeric | — |
| `SubstituateVolume` | numeric | — |
| `DilutionMnt` | character varying(64) | — |
| `SalineQuantity` | numeric | — |
| `SealQuantity` | numeric | — |
| `ArterialQuantity` | numeric | — |
| `VenousQuantity` | numeric | — |
| `UFQuantity` | numeric | — |
| `SealType` | character varying(64) | — |
| `GlucoseCon` | numeric | — |
| `DialysateGroupId` | bigint | — |
| `SubstituateFlow` | numeric | — |
| `IsInduceDialysisPrescription` | boolean | — |
| `HeparinType` | integer | — |

</details>

## `adjustment_records` → `Plan_PatientPlanPrescriptionAdjustment`

**类别：** `rename`  
**说明：** 字段少，简单改名

**新表字段数：** 5  **老表（主） `Plan_PatientPlanPrescriptionAdjustment` 字段数：** 94

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 关联 patients.id |
| `content` | text | `Content` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 调整内容描述 |
| `operator` | varchar(50) | — |  | 无匹配 | ❌ TODO 调整人 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Plan_PatientPlanPrescriptionAdjustment` 有 92 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Type` | integer | — |
| `PatientPlanPrescriptionId` | bigint | — |
| `AdjustUserId` | bigint | — |
| `AdjustTime` | timestamp | — |
| `AdjustReason` | character varying(512) | — |
| `BeforeOddWeekFrequency` | integer | — |
| `AfterOddWeekFrequency` | integer | — |
| `BeforeEvenWeekFrequency` | integer | — |
| `AfterEvenWeekFrequency` | integer | — |
| `BeforeDialysisMethod` | character varying(512) | — |
| `AfterDialysisMethod` | character varying(512) | — |
| `BeforeDialysisDuration` | numeric | — |
| `AfterDialysisDuration` | numeric | — |
| `BeforeDryWeight` | numeric | — |
| `AfterDryWeight` | numeric | — |
| `BeforeExtraWeight` | numeric | — |
| `AfterExtraWeight` | numeric | — |
| `BeforeAdjustQuantity` | numeric | — |
| `AfterAdjustQuantity` | numeric | — |
| `BeforeBF` | numeric | — |
| `AfterBF` | numeric | — |
| `BeforeFirstAnticoagulant` | bigint | — |
| `AfterFirstAnticoagulant` | bigint | — |
| `BeforeFirstDosage` | numeric | — |
| `AfterFirstDosage` | numeric | — |
| `BeforeMaintainAnticoagulant` | bigint | — |
| `AfterMaintainAnticoagulant` | bigint | — |
| `BeforeDilutionProportion` | numeric | — |
| `AfterDilutionProportion` | numeric | — |
| `BeforeInjectionRate` | numeric | — |
| `AfterInjectionRate` | numeric | — |
| `BeforeInjectionDuration` | numeric | — |
| `AfterInjectionDuration` | numeric | — |
| `BeforeInjectionVolume` | numeric | — |
| `AfterInjectionVolume` | numeric | — |
| `BeforeVascularAccessId` | bigint | — |
| `AfterVascularAccessId` | bigint | — |
| `BeforeDialysate` | character varying(64) | — |
| `AfterDialysate` | character varying(64) | — |
| `BeforeDialysateFlow` | numeric | — |
| `AfterDialysateFlow` | numeric | — |
| `BeforeDialysateVolume` | numeric | — |
| `AfterDialysateVolume` | numeric | — |
| `BeforeNaIonCon` | numeric | — |
| `AfterNaIonCon` | numeric | — |
| `BeforeCaIonCon` | numeric | — |
| `AfterCaIonCon` | numeric | — |
| `BeforeKIonCon` | numeric | — |
| `AfterKIonCon` | numeric | — |
| `BeforeHCO3IonCon` | numeric | — |
| `AfterHCO3IonCon` | numeric | — |
| `BeforeConductivity` | numeric | — |
| `AfterConductivity` | numeric | — |
| `BeforeDialysateTmp` | numeric | — |
| `AfterDialysateTmp` | numeric | — |
| `BeforeSubstituateVolume` | numeric | — |
| `AfterSubstituateVolume` | numeric | — |
| `BeforeDilutionMnt` | character varying(64) | — |
| `AfterDilutionMnt` | character varying(64) | — |
| `BeforeSalineQuantity` | numeric | — |
| `AfterSalineQuantity` | numeric | — |
| `BeforeSealType` | character varying(64) | — |
| `AfterSealType` | character varying(64) | — |
| `BeforeSealQuantity` | numeric | — |
| `AfterSealQuantity` | numeric | — |
| `BeforeArterialQuantity` | numeric | — |
| `AfterArterialQuantity` | numeric | — |
| `BeforeVenousQuantity` | numeric | — |
| `AfterVenousQuantity` | numeric | — |
| `BeforeUFQuantity` | numeric | — |
| `AfterUFQuantity` | numeric | — |
| `Status` | integer | — |
| `DealUserId` | bigint | — |
| `DealTime` | timestamp | — |
| `DealContent` | character varying(512) | — |
| `CreatorId` | bigint | — |
| `LastModifyTime` | timestamp | — |
| `BeforeBV` | numeric | — |
| `AfterBV` | numeric | — |
| `BeforeFrequency` | character varying(128) | — |
| `AfterFrequency` | character varying(128) | — |
| `BeforeGlucoseCon` | numeric | — |
| `AfterGlucoseCon` | numeric | — |
| `BeforeSubstituateFlow` | numeric | — |
| `AfterSubstituateFlow` | numeric | — |
| `BeforeIsInduceDialysisPrescription` | boolean | — |
| `AfterIsInduceDialysisPrescription` | boolean | — |
| `BeforeHeparinType` | integer | — |
| `AfterHeparinType` | integer | — |
| `BeforeMaterial` | jsonb | — |
| `AfterMaterial` | jsonb | — |

</details>

## `wards` → `Schedule_Ward`

**类别：** `rename`  
**新表字段数：** 12  **老表（主） `Schedule_Ward` 字段数：** 12

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅ 租户 ID |
| `name` | varchar(128) | `Name` | character varying(256) | synonym | ✅ 病房名称 |
| `ward_type` | varchar(64) | — |  | 无匹配 | ❌ TODO HD / HDF / Isolation / VIP |
| `department` | varchar(128) | — |  | 无匹配 | ❌ TODO 科室 |
| `floor` | int | — |  | 无匹配 | ❌ TODO 楼层 |
| `is_disabled` | bool | `IsDisabled` | boolean | synonym | ✅  |
| `sort` | int | `Sort` | integer | pascal-exact | ✅ 排序 |
| `notes` | text | `Note` | character varying(512) | synonym | ✅  |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Schedule_Ward` 有 3 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `PatientType` | character varying(64) | — |
| `InfectionType` | character varying(64) | — |
| `ResponsibleUsers` | character varying(512) | — |

</details>

## `beds` → `Schedule_Bed`

**类别：** `rename`  
**新表字段数：** 12  **老表（主） `Schedule_Bed` 字段数：** 12

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅ 租户 ID |
| `ward_id` | bigint | `WardId` | bigint | synonym | ✅ 关联 wards.id |
| `name` | varchar(64) | `Name` | character varying(256) | synonym | ✅ 床位号 |
| `bed_type` | varchar(64) | — |  | 无匹配 | ❌ TODO Regular / ICU / VIP / Isolation |
| `status` | varchar(20) | `Status` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 available / occupied / reserved / maintenance |
| `is_disabled` | bool | `IsDisabled` | boolean | synonym | ✅  |
| `sort` | int | `Sort` | integer | pascal-exact | ✅  |
| `notes` | text | `Note` | character varying(512) | synonym | ✅  |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Schedule_Bed` 有 2 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `FEPId` | bigint | — |
| `AcquisiteConnectId` | bigint | — |

</details>

## `shifts` → `Schedule_Shift`

**类别：** `rename`  
**新表字段数：** 12  **老表（主） `Schedule_Shift` 字段数：** 12

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅ 租户 ID |
| `name` | varchar(64) | `Name` | character varying(256) | synonym | ✅ 班次名称 |
| `start_time` | varchar(10) | `StartTime` | timestamp | synonym | ✅ 开始时间 HH:MM |
| `end_time` | varchar(10) | `EndTime` | timestamp | synonym | ✅ 结束时间 HH:MM |
| `type` | varchar(64) | `Type` | integer | synonym | ✅ Morning / Afternoon / Night / Overtime |
| `is_disabled` | bool | `IsDisabled` | boolean | synonym | ✅  |
| `sort` | int | `Sort` | integer | pascal-exact | ✅  |
| `notes` | text | `Note` | character varying(512) | synonym | ✅  |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

## `patient_shifts` → `Schedule_PatientShift`

**类别：** `rename`  
**新表字段数：** 13  **老表（主） `Schedule_PatientShift` 字段数：** 13

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅ 租户 ID |
| `patient_id` | bigint | `PatientId` | bigint | synonym | ✅ 关联患者 |
| `schedule_date` | timestamp | `TreatmentTime` | timestamp | synonym | ✅ 排班日期 |
| `shift_id` | bigint | `ShiftId` | bigint | synonym | ✅ 关联 shifts.id |
| `bed_id` | bigint | `BedId` | bigint | synonym | ✅ 关联 beds.id |
| `ward_id` | bigint | `WardId` | bigint | synonym | ✅ 关联 wards.id |
| `status` | int | `Status` | integer | synonym | ✅ 0=待执行 / 1=已确认 / 2=进行中 / 3=已完成 / 4=已取消 |
| `is_disabled` | bool | `IsDisabled` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `notes` | text | `Note` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Schedule_PatientShift` 有 2 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `PatientPlanId` | bigint | — |
| `ShiftTiming` | integer | — |

</details>

## `Treatment_Treatment` → `Treatment_Treatment`

**类别：** `same-name`  
**说明：** 同名表，字段需核对

**新表字段数：** 21  **老表（主） `Treatment_Treatment` 字段数：** 34

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅ 租户 ID |
| `patient_id` | bigint | `PatientId` | bigint | synonym | ✅  |
| `treatment_date` | timestamp | — |  | 无匹配 | ❌ TODO 治疗日期 |
| `schedule_id` | bigint | `ScheduleId` | bigint | pascal-exact | ✅ 关联 patient_shifts.id |
| `reception_dr_id` | bigint | `ReceptionDrId` | bigint | pascal-exact | ✅ 接诊医生 |
| `sign_in_time` | timestamp | `SignInTime` | timestamp | pascal-exact | ✅ 签到时间 |
| `queue_no` | varchar(32) | `QueueNo` | character varying(32) | pascal-exact | ✅ 排队号 |
| `reception_time` | timestamp | `ReceptionTime` | timestamp | pascal-exact | ✅ 接诊时间 |
| `day_programme_id` | bigint | `DayProgrammeId` | bigint | pascal-exact | ✅ 日间治疗方案 ID |
| `ward_id` | bigint | `WardId` | bigint | synonym | ✅  |
| `ward_name` | varchar(256) | `WardName` | character varying(256) | pascal-exact | ✅ 病区名称 |
| `bed_id` | bigint | `BedId` | bigint | synonym | ✅  |
| `shift_id` | bigint | `ShiftId` | bigint | synonym | ✅  |
| `shift_timing` | int | — |  | 无匹配 | ❌ TODO 班次时段 |
| `type` | int | `Type` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 1=HD / 2=HDF / 3=HP / 4=HD+HP |
| `status` | int | `Status` | character varying(64) | synonym | ✅ 0=待开始 / 1=进行中 / 2=已完成 / 3=已取消 |
| `is_disabled` | bool | `IsDisabled` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Treatment_Treatment` 有 17 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `BedName` | character varying(256) | — |
| `EquipmentId` | character varying | — |
| `EquipmentName` | character varying(512) | — |
| `StartTime` | timestamp | — |
| `EndTime` | timestamp | — |
| `RealDuration` | numeric | — |
| `RealUFQuantity` | numeric | — |
| `NurseSummary` | character varying(1024) | — |
| `TreatmentSummary` | character varying(1024) | — |
| `CaseStatus` | character varying(64) | — |
| `ShiftName` | character varying(256) | — |
| `RealSubstituateVolume` | numeric | — |
| `TreatmentCount` | integer | — |
| `HospPatientType` | character varying(64) | — |
| `TmrPath` | character varying(1024) | — |
| `TmrTime` | timestamp | — |
| `TmrPages` | integer | — |

</details>

## `Treatment_BeforeCheck` → `Treatment_BeforeCheck`

**类别：** `same-name`  
**说明：** 同名表

**新表字段数：** 28  **老表（主） `Treatment_BeforeCheck` 字段数：** 18

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `treatment_id` | bigint | `TreatmentId` | bigint | pascal-exact | ✅ 关联 Treatment_Treatment.id |
| `weight` | float64 | — |  | 无匹配 | ❌ TODO 体重 (kg) |
| `temperature` | float64 | — |  | 无匹配 | ❌ TODO 体温 (°C) |
| `sbp` | int | — |  | 无匹配 | ❌ TODO 收缩压 (mmHg) |
| `dbp` | int | — |  | 无匹配 | ❌ TODO 舒张压 (mmHg) |
| `heart_rate` | int | — |  | 无匹配 | ❌ TODO 心率 (bpm) |
| `edema` | int | — |  | 无匹配 | ❌ TODO 水肿程度 |
| `consciousness` | varchar(32) | — |  | 无匹配 | ❌ TODO 意识状态 |
| `complication` | varchar(512) | — |  | 无匹配 | ❌ TODO 并发症 |
| `dry_weight` | float64 | `PredictWeight` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 干体重 |
| `pre_weight` | float64 | — |  | 无匹配 | ❌ TODO 预估体重 |
| `vascular_access` | varchar(256) | — |  | 无匹配 | ❌ TODO 血管通路 |
| `cannula_type` | varchar(64) | — |  | 无匹配 | ❌ TODO 穿刺类型 |
| `cannula_position` | varchar(256) | — |  | 无匹配 | ❌ TODO 穿刺部位 |
| `catheter` | varchar(512) | — |  | 无匹配 | ❌ TODO 导管情况 |
| `heparing_lock` | varchar(512) | — |  | 无匹配 | ❌ TODO 肝素封管 |
| `machine_no` | varchar(64) | — |  | 无匹配 | ❌ TODO 机器号 |
| `dialyzer` | varchar(256) | — |  | 无匹配 | ❌ TODO 透析器 |
| `dialysate` | varchar(256) | — |  | 无匹配 | ❌ TODO 透析液 |
| `calcium` | float64 | — |  | 无匹配 | ❌ TODO 钙浓度 |
| `sodium` | float64 | — |  | 无匹配 | ❌ TODO 钠浓度 |
| `bicarbonate` | float64 | — |  | 无匹配 | ❌ TODO 碳酸氢根 |
| `notes` | varchar(1024) | `Note` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 备注 |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Treatment_BeforeCheck` 有 12 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `BeforeSignsId` | bigint | — |
| `BeforeSymptomId` | bigint | — |
| `OperatorId` | bigint | — |
| `OperateTime` | timestamp | — |
| `MaterialsResult` | boolean | — |
| `MaterialsMistake` | character varying(1024) | — |
| `ParamResult` | boolean | — |
| `ParamMistake` | character varying(1024) | — |
| `VascularAccessResult` | boolean | — |
| `VascularAccessMistake` | character varying(1024) | — |
| `PipelineResult` | boolean | — |
| `PipelineMistake` | character varying(1024) | — |

</details>

## `Treatment_BeforeSigns` → `Treatment_BeforeSigns`

**类别：** `same-name`  
**说明：** 同名表

**新表字段数：** 15  **老表（主） `Treatment_BeforeSigns` 字段数：** 17

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `treatment_id` | bigint | `TreatmentId` | bigint | pascal-exact | ✅  |
| `sbp` | int | `SBP` | numeric | case-insensitive | ✅ 收缩压 |
| `dbp` | int | `DBP` | numeric | case-insensitive | ✅ 舒张压 |
| `heart_rate` | int | `HeartRate` | numeric | pascal-exact | ✅ 心率 |
| `sp_o2` | int | — |  | 无匹配 | ❌ TODO 血氧饱和度 (%) |
| `respiration` | int | `Respiration` | numeric | pascal-exact | ✅ 呼吸频率 |
| `temperature` | float64 | — |  | 无匹配 | ❌ TODO 体温 |
| `weight` | float64 | `Weight` | numeric | pascal-exact | ✅ 体重 |
| `symptoms` | varchar(1024) | — |  | 无匹配 | ❌ TODO 症状描述 |
| `notes` | varchar(1024) | `Note` | character varying(1024) | synonym | ✅ 备注 |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Treatment_BeforeSigns` 有 5 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `OperatorId` | bigint | — |
| `OperateTime` | timestamp | — |
| `ExtraWeight` | numeric | — |
| `BodyTemp` | numeric | — |
| `PressurePoint` | character varying(64) | — |

</details>

## `Treatment_DuringParam` → `Treatment_DuringParam`

**类别：** `same-name`  
**说明：** 同名表

**新表字段数：** 18  **老表（主） `Treatment_DuringParam` 字段数：** 22

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `treatment_id` | bigint | `TreatmentId` | bigint | pascal-exact | ✅  |
| `record_time` | timestamp | — |  | 无匹配 | ❌ TODO 记录时间 |
| `code` | varchar(32) | — |  | 无匹配 | ❌ TODO 参数代码 |
| `blood_flow` | float64 | — |  | 无匹配 | ❌ TODO 血流量 (ml/min) |
| `dialysate_flow` | float64 | — |  | 无匹配 | ❌ TODO 透析液流量 (ml/min) |
| `uf_volume` | float64 | — |  | 无匹配 | ❌ TODO 超滤量 (ml) |
| `venous_pressure` | float64 | `VenousPressure` | numeric | pascal-exact | ✅ 静脉压 (mmHg) |
| `arterial_pressure` | float64 | `ArterialPressure` | numeric | pascal-exact | ✅ 动脉压 (mmHg) |
| `tmp` | float64 | `TMP` | numeric | case-insensitive | ✅ 跨膜压 (mmHg) |
| `temperature` | float64 | — |  | 无匹配 | ❌ TODO 温度 (°C) |
| `conductivity` | float64 | `Conductivity` | numeric | pascal-exact | ✅ 电导度 (mS/cm) |
| `uf_rate` | float64 | — |  | 无匹配 | ❌ TODO 超滤率 (ml/h) |
| `notes` | varchar(512) | `Note` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 备注 |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Treatment_DuringParam` 有 12 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `OperatorId` | bigint | — |
| `OperateTime` | timestamp | — |
| `UFQuantity` | numeric | — |
| `MachineTmp` | numeric | — |
| `BF` | numeric | — |
| `SubstituateSpeed` | numeric | — |
| `HeparinPumpFlow` | numeric | — |
| `RelativeBloodVolume` | numeric | — |
| `RealBloodVolume` | numeric | — |
| `RealClearanceRate` | numeric | — |
| `ArterialBloodTemp` | numeric | — |
| `VenousBloodTemp` | numeric | — |

</details>

## `Treatment_AfterSigns` → `Treatment_AfterSigns`

**类别：** `same-name`  
**说明：** 同名表

**新表字段数：** 16  **老表（主） `Treatment_AfterSigns` 字段数：** 19

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `treatment_id` | bigint | `TreatmentId` | bigint | pascal-exact | ✅  |
| `sbp` | int | `SBP` | numeric | case-insensitive | ✅ 收缩压 |
| `dbp` | int | `DBP` | numeric | case-insensitive | ✅ 舒张压 |
| `heart_rate` | int | `HeartRate` | numeric | pascal-exact | ✅ 心率 |
| `sp_o2` | int | — |  | 无匹配 | ❌ TODO 血氧饱和度 |
| `weight` | float64 | `Weight` | numeric | pascal-exact | ✅ 体重 |
| `uf_volume` | float64 | — |  | 无匹配 | ❌ TODO 实际超滤量 (ml) |
| `dialysis_time` | int | — |  | 无匹配 | ❌ TODO 透析时长 (分钟) |
| `complication` | varchar(1024) | — |  | 无匹配 | ❌ TODO 并发症 |
| `symptoms` | varchar(1024) | — |  | 无匹配 | ❌ TODO 症状描述 |
| `notes` | varchar(1024) | `Note` | character varying(1024) | synonym | ✅ 备注 |
| `creator_id` | bigint | `CreatorId` | bigint | synonym | ✅  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Treatment_AfterSigns` 有 8 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `OperatorId` | bigint | — |
| `OperateTime` | timestamp | — |
| `ExtraWeight` | numeric | — |
| `LossWeight` | numeric | — |
| `RealIntake` | numeric | — |
| `BodyTemp` | numeric | — |
| `PressurePoint` | character varying(64) | — |
| `Respiration` | numeric | — |

</details>

## `Treatment_Alarm` → `Treatment_Alarm`

**类别：** `same-name`  
**说明：** 同名表

**新表字段数：** 14  **老表（主） `Treatment_Alarm` 字段数：** 13

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | bigint | `Id` | bigint | pascal-exact | ✅  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `treatment_id` | bigint | `TreatmentId` | bigint | pascal-exact | ✅  |
| `alarm_time` | timestamp | — |  | 无匹配 | ❌ TODO 报警时间 |
| `alarm_code` | varchar(64) | — |  | 无匹配 | ❌ TODO 报警代码 |
| `alarm_level` | int | — |  | 无匹配 | ❌ TODO 1=信息 / 2=警告 / 3=错误 / 4=严重 |
| `alarm_message` | varchar(512) | — |  | 无匹配 | ❌ TODO 报警信息 |
| `is_handled` | bool | — |  | 无匹配 | ❌ TODO 是否已处理 |
| `handled_by` | bigint | — |  | 无匹配 | ❌ TODO 处理人 ID |
| `handled_at` | timestamp | — |  | 无匹配 | ❌ TODO 处理时间 |
| `handle_note` | varchar(512) | — |  | 无匹配 | ❌ TODO 处理说明 |
| `creator_id` | bigint | `CreatorId` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `create_time` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `last_modify_time` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Treatment_Alarm` 有 8 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `Source` | integer | — |
| `Code` | character varying(64) | — |
| `Content` | text | — |
| `Levle` | integer | — |
| `EndTime` | timestamp | — |
| `HandleTime` | timestamp | — |
| `HandleId` | bigint | — |
| `HandleContent` | text | — |

</details>

## `plan_templates` → `Plan_PlanTPL + Plan_PlanTPLMaterial`

**类别：** `rewrite+child`  
**说明：** 新表 templateContent 为 JSONB，老库为父表 + 材料子表

**新表字段数：** 11  **老表（主） `Plan_PlanTPL` 字段数：** 35

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `name` | varchar(100) | `Name` | character varying(256) | synonym | ✅ 模板名称 |
| `description` | text | — |  | 无匹配 | ❌ TODO 描述 |
| `mode` | varchar(20) | — |  | 无匹配 | ❌ TODO HD / HDF / HP / HF / HFD |
| `is_default` | bool | — |  | 无匹配 | ❌ TODO  |
| `is_enabled` | bool | — |  | 无匹配 | ❌ TODO  |
| `category` | varchar(50) | `Classification` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 分类: 急性 / 慢性 / 导管等 |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `template_content` | **jsonb** | — |  | 无匹配 | ❌ TODO 完整模板内容 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Plan_PlanTPL` 有 30 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `CreatorId` | bigint | — |
| `IsDisabled` | boolean | — |
| `DialysisMethod` | character varying(256) | — |
| `DialysisDuration` | numeric | — |
| `AdjustQuantity` | numeric | — |
| `BF` | numeric | — |
| `BV` | numeric | — |
| `FirstAnticoagulant` | bigint | — |
| `FirstDosage` | numeric | — |
| `MaintainAnticoagulant` | bigint | — |
| `DilutionProportion` | numeric | — |
| `InjectionRate` | numeric | — |
| `InjectionDuration` | numeric | — |
| `InjectionVolume` | numeric | — |
| `VascularAccessId` | bigint | — |
| `Dialysate` | character varying(64) | — |
| `DialysateFlow` | numeric | — |
| `DialysateVolume` | numeric | — |
| `NaIonCon` | numeric | — |
| `CaIonCon` | numeric | — |
| `KIonCon` | numeric | — |
| `Conductivity` | numeric | — |
| `DialysateTmp` | numeric | — |
| `SubstituateVolume` | numeric | — |
| `DilutionMnt` | character varying(64) | — |
| `HCO3IonCon` | numeric | — |
| `GlucoseCon` | numeric | — |
| `DialysateGroupId` | bigint | — |
| `Note` | text | — |
| `SubstituateFlow` | numeric | — |

</details>

## `material_catalogs` → `Auxiliary_MaterialInfomation`

**类别：** `rename`  
**新表字段数：** 18  **老表（主） `Auxiliary_MaterialInfomation` 字段数：** 20

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | uint | `Id` | bigint | pascal-exact | ✅  |
| `code` | varchar(50) | `Code` | character varying(128) | pascal-exact | ✅ 材料编码 |
| `name` | varchar(100) | `Name` | character varying(256) | synonym | ✅ 材料名称 |
| `short_name` | varchar(50) | `ShortName` | character varying(128) | pascal-exact | ✅ 简称 |
| `mnemonic` | varchar(50) | — |  | 无匹配 | ❌ TODO 助记符 |
| `category` | varchar(50) | `Classification` | character varying(64) | synonym | ✅ 材料分类 |
| `spec` | varchar(100) | — |  | 无匹配 | ❌ TODO 规格 |
| `standard_type` | varchar(50) | — |  | 无匹配 | ❌ TODO 标准类型 |
| `brand` | varchar(100) | `Brand` | character varying(64) | pascal-exact | ✅ 品牌 |
| `unit` | varchar(20) | `Unit` | character varying(64) | synonym | ✅ 单位 |
| `packaging` | varchar(50) | — |  | 无匹配 | ❌ TODO 包装 |
| `manufacturer` | varchar(100) | `Manufacturer` | character varying(128) | pascal-exact | ✅ 生产厂家 |
| `sort_order` | int | — |  | 无匹配 | ❌ TODO 排序 |
| `is_enabled` | bool | — |  | 无匹配 | ❌ TODO  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `notes` | text | `Note` | character varying(512) | synonym | ✅  |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Auxiliary_MaterialInfomation` 有 8 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `Spell` | character varying(256) | — |
| `Specification` | character varying(64) | — |
| `Package` | character varying(64) | — |
| `Type` | character varying(64) | — |
| `IsDisabled` | boolean | — |
| `CreatorId` | bigint | — |
| `StdCat` | character varying(256) | — |
| `Sort` | numeric | — |

</details>

## `drug_catalogs` → `Auxiliary_DrugInfomation`

**类别：** `rename`  
**新表字段数：** 24  **老表（主） `Auxiliary_DrugInfomation` 字段数：** 23

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | uint | `Id` | bigint | pascal-exact | ✅  |
| `code` | varchar(50) | `Code` | character varying(128) | pascal-exact | ✅ 药品编码 |
| `name` | varchar(100) | `Name` | character varying(256) | synonym | ✅ 药品名称 |
| `short_name` | varchar(50) | `ShortName` | character varying(128) | pascal-exact | ✅ 简称 |
| `mnemonic` | varchar(50) | — |  | 无匹配 | ❌ TODO 助记符（拼音首字母） |
| `generic_name` | varchar(100) | — |  | 无匹配 | ❌ TODO 通用名 |
| `category` | varchar(50) | `Classification` | character varying(64) | synonym | ✅ 药品分类 |
| `spec` | varchar(100) | — |  | 无匹配 | ❌ TODO 规格 |
| `concentration` | varchar(50) | — |  | 无匹配 | ❌ TODO 浓度 |
| `spec_unit` | varchar(20) | — |  | 无匹配 | ❌ TODO 规格单位 |
| `min_unit_dose` | varchar(20) | — |  | 无匹配 | ❌ TODO 最小单位剂量 |
| `unit` | varchar(20) | `Unit` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 基本单位 |
| `brand` | varchar(100) | `Brand` | character varying(64) | pascal-exact | ✅ 品牌 |
| `packaging` | varchar(50) | — |  | 无匹配 | ❌ TODO 包装 |
| `manufacturer` | varchar(100) | `Manufacturer` | character varying(128) | pascal-exact | ✅ 生产厂家 |
| `standard_type` | varchar(50) | — |  | 无匹配 | ❌ TODO 标准类型 |
| `timing` | varchar(50) | `Timing` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 使用时机 |
| `tips` | varchar(200) | — |  | 无匹配 | ❌ TODO 提示信息 |
| `sort_order` | int | — |  | 无匹配 | ❌ TODO 排序 |
| `is_enabled` | bool | — |  | 无匹配 | ❌ TODO 是否启用 |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `notes` | text | `Note` | character varying(512) | synonym | ✅ 备注 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Auxiliary_DrugInfomation` 有 12 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `Specification` | character varying(64) | — |
| `Package` | character varying(64) | — |
| `IsDisabled` | boolean | — |
| `CreatorId` | bigint | — |
| `Spell` | character varying(256) | — |
| `BasicUnit` | character varying(64) | — |
| `SpecificationUnit` | character varying(64) | — |
| `Sort` | integer | — |
| `StdCat` | character varying(256) | — |
| `UseTips` | character varying(1024) | — |
| `MinUnitDosage` | bigint | — |
| `UseOpportunity` | character varying(64) | — |

</details>

## `order_templates` → `Order_OrderTPL`

**类别：** `rename`  
**说明：** 老库可能仅一张表包含模板；需确认是否含子条目

**新表字段数：** 12  **老表（主） `Order_OrderTPL` 字段数：** 18

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `name` | varchar(100) | `Name` | character varying(256) | synonym | ✅ 模板名称 |
| `type` | varchar(20) | `Type` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 长期 / 临时 |
| `category` | varchar(50) | `Classification` | character varying(64) | synonym | ✅ 药品 / 检查 / 治疗 / 护理 / 饮食 |
| `content` | text | `Content` | character varying(256) | synonym | ✅ 模板内容 |
| `frequency` | varchar(50) | `Frequency` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 默认频次 |
| `priority` | varchar(20) | `Priority` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 优先级 |
| `is_default` | bool | — |  | 无匹配 | ❌ TODO  |
| `is_enabled` | bool | — |  | 无匹配 | ❌ TODO  |
| `tenant_id` | bigint | `TenantId` | bigint | synonym | ✅  |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Order_OrderTPL` 有 11 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `OrderGroup` | bigint | — |
| `IsDisabled` | boolean | — |
| `DrugId` | bigint | — |
| `Dosage` | character varying(64) | — |
| `UseOpportunity` | character varying(128) | — |
| `UseMethod` | character varying(128) | — |
| `UseWay` | character varying(128) | — |
| `Note` | character varying(1024) | — |
| `CreatorId` | bigint | — |
| `UseNum` | numeric | — |
| `AllDosage` | numeric | — |

</details>

## `order_template_items` → `Order_OrderTPL (同表)`

**类别：** `fold-to-parent`  
**说明：** 需确认：若老库无子表，items 以 JSON/多行形式存在父表

**新表字段数：** 15  **老表（主） `Order_OrderTPL` 字段数：** 18

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID (BeforeCreate 自动生成) |
| `template_id` | varchar(36) | — |  | 无匹配 | ❌ TODO 关联 order_templates.id |
| `drug_id` | uint | `DrugId` | bigint | pascal-exact | ✅ 关联 drug_catalogs.id |
| `drug_name` | varchar(100) | — |  | 无匹配 | ❌ TODO 药品名称 |
| `spec` | varchar(100) | — |  | 无匹配 | ❌ TODO 规格 |
| `min_unit_dose` | varchar(20) | — |  | 无匹配 | ❌ TODO 最小单位剂量 |
| `dosage` | varchar(50) | `Dosage` | character varying(64) | pascal-exact | ✅ 用量 |
| `unit` | varchar(20) | `Unit` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 单位 |
| `route` | varchar(50) | `Route` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 用法 |
| `frequency` | varchar(50) | `Frequency` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 频次 |
| `timing` | varchar(50) | `Timing` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 使用时机 |
| `group_id` | varchar(36) | — |  | 无匹配 | ❌ TODO 组号 |
| `sort_order` | int | — |  | 无匹配 | ❌ TODO 排序 |
| `created_at` | timestamp | `CreateTime` | timestamp | synonym | ✅  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp | synonym | ✅  |

<details><summary>老表 `Order_OrderTPL` 有 13 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Name` | character varying(256) | — |
| `OrderGroup` | bigint | — |
| `IsDisabled` | boolean | — |
| `Classification` | character varying(64) | — |
| `Content` | character varying(256) | — |
| `UseOpportunity` | character varying(128) | — |
| `UseMethod` | character varying(128) | — |
| `UseWay` | character varying(128) | — |
| `Note` | character varying(1024) | — |
| `CreatorId` | bigint | — |
| `UseNum` | numeric | — |
| `AllDosage` | numeric | — |

</details>

## `dict_types` → `CodeDictionary_CodeDictionarys`

**类别：** `rewrite`  
**说明：** 新表为 type/item 两表；老库单表树形（parent + category 分类）

**新表字段数：** 9  **老表（主） `CodeDictionary_CodeDictionarys` 字段数：** 7

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | — |  | 无匹配 | ❌ TODO UUID (BeforeCreate 自动生成) |
| `code` | varchar(50) | `Code` | character varying(64) | pascal-exact | ✅ 字典类型编码 |
| `name` | varchar(100) | `Name` | character varying(64) | synonym | ✅ 字典类型名称 |
| `description` | varchar(500) | — |  | 无匹配 | ❌ TODO 描述 |
| `icon` | varchar(50) | — |  | 无匹配 | ❌ TODO 图标 |
| `sort_order` | int | — |  | 无匹配 | ❌ TODO  |
| `is_enabled` | bool | — |  | 无匹配 | ❌ TODO  |
| `created_at` | timestamp | `CreateTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `updated_at` | timestamp | `LastModifyTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |

<details><summary>老表 `CodeDictionary_CodeDictionarys` 有 5 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `Type` | character varying(64) | — |
| `OrganId` | bigint | — |
| `IsDisabled` | boolean | — |
| `Sort` | integer | — |
| `Builtin` | boolean | — |

</details>

## `dict_items` → `CodeDictionary_CodeDictionarys`

**类别：** `rewrite`  
**说明：** 同上

**新表字段数：** 11  **老表（主） `CodeDictionary_CodeDictionarys` 字段数：** 7

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | — |  | 无匹配 | ❌ TODO UUID (BeforeCreate 自动生成) |
| `type_code` | varchar(50) | — |  | 无匹配 | ❌ TODO 关联 dict_types.code |
| `code` | varchar(50) | `Code` | character varying(64) | pascal-exact | ✅ 字典项编码 |
| `name` | varchar(100) | `Name` | character varying(64) | synonym | ✅ 字典项名称 |
| `description` | varchar(500) | — |  | 无匹配 | ❌ TODO 描述 |
| `sort_order` | int | — |  | 无匹配 | ❌ TODO  |
| `is_enabled` | bool | — |  | 无匹配 | ❌ TODO  |
| `extra` | varchar(500) | — |  | 无匹配 | ❌ TODO 扩展字段（如颜色标识） |
| `parent_code` | varchar(50) | — |  | 无匹配 | ❌ TODO 父级代码（树形结构） |
| `created_at` | timestamp | `CreateTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `updated_at` | timestamp | `LastModifyTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |

<details><summary>老表 `CodeDictionary_CodeDictionarys` 有 5 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `Type` | character varying(64) | — |
| `OrganId` | bigint | — |
| `IsDisabled` | boolean | — |
| `Sort` | integer | — |
| `Builtin` | boolean | — |

</details>

## `lab_reports` → `LIS_Examination`

**类别：** `rename-fields`  
**说明：** 已部分在 patient_core_service.buildLabTrends 使用

**新表字段数：** 19  **老表（主） `LIS_Examination` 字段数：** 10

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `patient_id` | varchar(36) | `PatientId` | bigint | synonym | ✅  |
| `report_no` | varchar(64) | — |  | 无匹配 | ❌ TODO 报告号 |
| `item_code` | varchar(64) | `ItemCode` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 检验项目编码 |
| `item_name` | varchar(128) | `ItemName` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 检验项目名称 |
| `clinical_diagnosis` | text | — |  | 无匹配 | ❌ TODO 临床诊断 |
| `specimen_type` | varchar(64) | `SpecimenType` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 标本类型 |
| `urgency` | varchar(32) | `Urgency` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 常规 / 急诊 |
| `request_doctor` | varchar(64) | `RequestDoctor` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 申请医生 |
| `requested_at` | timestamp | `RequestTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 申请时间 |
| `sampled_at` | timestamp | `SampleTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 采样时间 |
| `received_at` | timestamp | `ReceiveTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 接收时间 |
| `reported_at` | timestamp | `ReportTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 报告时间 |
| `status` | varchar(32) | `Status` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `external_report_id` | varchar(128) | — |  | 老库无对应（需业务决策） | ❌ TODO 外部系统报告 ID |
| `source_system` | varchar(16) | — |  | 老库无对应（需业务决策） | ❌ TODO LOCAL / LIS / PACS / HDIS_EXAM / HDIS_RECORD |
| `synced_at` | timestamp | — |  | 老库无对应（需业务决策） | ❌ TODO 同步时间 |
| `created_at` | timestamp | `CreateTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp(6) | synonym | ✅  |

<details><summary>老表 `LIS_Examination` 有 7 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `Name` | character varying(128) | — |
| `Type` | character varying(64) | — |
| `ResultTime` | timestamp | — |
| `SyncUserId` | bigint | — |
| `SyncTime` | timestamp | — |
| `TestNO` | character varying(64) | — |

</details>

## `lab_report_items` → `LIS_ExaminationItem`

**类别：** `rename-fields`  
**说明：** 同上

**新表字段数：** 11  **老表（主） `LIS_ExaminationItem` 字段数：** 10

| 新字段 | 新类型 | 老字段 | 老类型 | 匹配方式 | 备注 |
|--------|--------|--------|--------|----------|------|
| `id` | varchar(36) | `Id` | bigint | pascal-exact | ✅ UUID |
| `lab_report_id` | varchar(36) | — |  | 无匹配 | ❌ TODO 关联 lab_reports.id (CASCADE DELETE) |
| `item_code` | varchar(64) | `ItemCode` | character varying(64) | synonym | ✅ 指标编码 |
| `item_name` | varchar(128) | `ItemName` | character varying(64) | synonym | ✅ 指标名称 |
| `result_value` | varchar(64) | `Result` | character varying(64) | synonym | ✅ 检验结果 |
| `unit` | varchar(32) | `Unit` | character varying(32) | synonym | ✅ 单位 |
| `reference_range` | varchar(128) | `Reference` | character varying(64) | synonym | ✅ 参考范围 |
| `abnormal_flag` | varchar(8) | `ResultSign` | character varying(16) | synonym | ✅ H=偏高 / L=偏低 / N=正常 |
| `tested_at` | timestamp | `ResultTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表 检测时间 |
| `created_at` | timestamp | `CreateTime` |  | synonym-不在目标表（可能跨表/计算得到） | ⚠️ 跨表  |
| `updated_at` | timestamp | `LastModifyTime` | timestamp(6) | synonym | ✅  |

<details><summary>老表 `LIS_ExaminationItem` 有 2 个未被新表消费的字段</summary>

| 老字段 | 类型 | 业务含义 |
|--------|------|----------|
| `TenantId` | bigint | — |
| `ExaminationId` | bigint | — |

</details>

## `exam_reports` → `—（应用层保留）`

**类别：** `app-only+sync`  
**说明：** 老库无本地表，数据来自 HDIS 同步；保留新表

> 🟢 **此表保留为应用层独立表，无需迁移。**
> 数据由外部系统（HDIS/LIS）同步写入；老库无本地副本。

## `patient_key_indicators` → `—（应用层保留）`

**类别：** `app-only+sync`  
**说明：** 老库无；来自 HDIS Record 同步；保留新表

> 🟢 **此表保留为应用层独立表，无需迁移。**
> 数据由外部系统（HDIS/LIS）同步写入；老库无本地副本。

## `integration_hdis_settings` → `—（应用层保留）`

**类别：** `app-only`  
**说明：** 新系统配置表，与老库无关

> 🟢 **此表保留为应用层独立表，无需迁移。**

## `permissions` → `—（应用层保留）`

**类别：** `app-only`  
**说明：** 权限定义，应用层

> 🟢 **此表保留为应用层独立表，无需迁移。**

## `role_permissions` → `—（应用层保留）`

**类别：** `app-only`  
**说明：** 角色-权限关联

> 🟢 **此表保留为应用层独立表，无需迁移。**

## `clinical_tasks` → `—（应用层保留）`

**类别：** `app-only?`  
**说明：** 待评估：老库可能无对应；如纯应用层则保留

> ⚠️ 未在 DATABASE_DESIGN.md 找到 `clinical_tasks` 字段定义，跳过字段级映射。

## `devices` → `Auxiliary_EquipmentInfomation + Schedule_BedEquipmentRel + Schedule_Bed + Schedule_Ward`

**类别：** `multi-join`  
**说明：** 已由 device_service 通过多表 join 实现；model.TableName 仍为 devices，需修正

> ⚠️ 未在 DATABASE_DESIGN.md 找到 `devices` 字段定义，跳过字段级映射。

## `inventory_items` → `Stock_Stock + Stock_Storage`

**类别：** `rewrite`  
**说明：** 老库库存拆为 Stock_Stock（总账） + Stock_Storage（仓库）

> ⚠️ 未在 DATABASE_DESIGN.md 找到 `inventory_items` 字段定义，跳过字段级映射。

## `stock_logs` → `Stock_InOutStorage + Stock_InOutStorageDetail`

**类别：** `rewrite`  
**说明：** 出入库单 + 明细

> ⚠️ 未在 DATABASE_DESIGN.md 找到 `stock_logs` 字段定义，跳过字段级映射。

## `label_tasks` → `—（应用层保留）`

**类别：** `app-only`  
**说明：** 条码标签任务，应用层

> 🟢 **此表保留为应用层独立表，无需迁移。**
