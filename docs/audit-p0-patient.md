# 患者模块字段级核查审计报告

> 审计时间：2026-05-31
> 审计范围：患者列表 (`/patients`)、患者详情 (`/patients/:id`) 及其 9 个子 Tab + 5 个附属组件
> 核查依据：`老血透数据库表结构-合并版.md`、`LEGACY_TABLE_FIELD_MAPPING.md`、`字典类型对照表(1).md`

---

## 一、总览

| 模块 | 一致 | 不一致 | 缺失 | 只读未闭环 | 待确认 | 不适用 |
|------|------|--------|------|-----------|--------|--------|
| 患者列表 PatientList | 6 | 2 | 3 | 0 | 1 | 0 |
| 患者详情概览 OverviewTab | 4 | 3 | 2 | 0 | 0 | 0 |
| 基本信息 BasicInfoTab | 12 | 5 | 4 | 0 | 2 | 0 |
| 治疗方案 TreatmentPlanTab | 8 | 4 | 6 | 0 | 2 | 0 |
| 医嘱管理 SchemeOrderTab | 5 | 3 | 4 | 0 | 1 | 0 |
| 检查检验 LabsExamsTab | 3 | 2 | 1 | 0 | 0 | 0 |
| 临床病史 MedicalRecordTab | 4 | 2 | 3 | 0 | 1 | 0 |
| 血管通路 VascularTab | 6 | 3 | 2 | 0 | 1 | 0 |
| 治疗历史 HistoryTab | 3 | 1 | 2 | 0 | 0 | 0 |
| 月份小结 MonthlySummaryTab | 0 | 0 | 0 | 14 | 0 | 0 |
| 附属组件 (Header/Drawer等) | 5 | 1 | 1 | 0 | 0 | 0 |
| **合计** | **56** | **26** | **28** | **14** | **8** | **0** |

---

## 二、问题项详细清单

### P0 — 关键数据映射/存储问题

| # | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|---|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|------------|-----------|
| 1 | P0 | PatientList | PatientList.tsx:310 | 列表行-年龄 | `patient.age` | `GET /api/v1/patients` | patient_service.go:436 | `Age`(GORM `gorm:"-"`) | `Register_PatientInfomation.BirthDate`(timestamp) | 前端显示age(int)，老库存BirthDate(timestamp)。后端Age字段为计算字段(gorm:"-")，不在DB列映射中 | **不一致** | `patient.go:65` Age gorm:"-"; 老库BirthDate为timestamp | 后端从BirthDate计算age传给前端，但列表接口需确认是否已计算 | 否 |
| 2 | P0 | PatientList | PatientList.tsx:345-348 | 列表行-床号 | `patient.bedNumber` | `GET /api/v1/patients` | patient_service.go:436 | `BedNumber`(GORM `gorm:"-"`) | 老库`Register_PatientInfomation`无`bed_number`字段 | 前端展示bedNumber，但老库患者主表无此字段。映射文档明确说"legacy 主表无 bed_number" | **缺失** | LEGACY_TABLE_FIELD_MAPPING.md:23; patient.go:66 BedNumber gorm:"-" | 需从排班表`Schedule_PatientShift`/`Schedule_Bed`关联获取，或标记为暂不支持 | 否 |
| 3 | P0 | PatientList | PatientList.tsx:326 | 列表行-透析模式 | `patient.defaultMode` | `GET /api/v1/patients` | patient_service.go | `DefaultMode`(GORM `gorm:"-"`) | `Plan_PatientPlan.DialysisMethod`(varchar 256) | 前端展示defaultMode，需跨表查询Plan_PatientPlan。列表接口性能隐患 | **不一致** | patient.go:69 DefaultMode gorm:"-"; 老库DialysisMethod在Plan_PatientPlan | 列表接口应JOIN Plan_PatientPlan取最新方案的DialysisMethod，或冗余到主查询 | 否 |
| 4 | P0 | PatientList | PatientList.tsx:331 | 列表行-干体重 | `patient.dryWeight` | `GET /api/v1/patients` | patient.go:62 | `Weight`(GORM → `Register_PatientInfomation.Weight`) | `Register_PatientInfomation.Weight`(numeric) | 前端字段名dryWeight映射到老库Weight列。含义：Weight在老库是"体重"非"干体重"，干体重在Plan_PatientPlan.DryWeight | **不一致** | patient.go:62 DryWeight→Weight; 老库Weight="体重", DryWeight在Plan表 | Patient.DryWeight不应映射到Register.Weight。应从Plan_PatientPlan.DryWeight读取 | 否 |
| 5 | P0 | PatientList | PatientList.tsx:309 | 列表行-诊断 | `patient.diagnosis` (间接) | `GET /api/v1/patients` | patient.go:61 | `Note`(GORM → `Register_PatientInfomation.Note`) | `Register_PatientInfomation.Note`(varchar 1024) | 前端Patient.diagnosis映射到老库Note字段。老库Note是"备注"而非"诊断"，诊断在Register_Diagnosis.DiagnosisDesc | **不一致** | patient.go:61 Diagnosis→Note; 老库Note="备注" | diagnosis字段应从Register_Diagnosis表读取，Note作为fallback | 否 |
| 6 | P0 | OverviewTab | OverviewTab.tsx:156-163 | 感染监控 | `infection.hbsag/hcvab/hivab/tpab/tb` | `GET /api/v1/patients/:id/core` | patient_core_service.go:228-247 | `Register_Infection.InfectionDesc+OtherDesc+Note`(合并文本) | `Register_Infection.InfectionDesc`(varchar 1024) | 后端将InfectionDesc+OtherDesc+Note合并后做关键词匹配(hbsag/hbv/乙肝等)检测阳性/阴性。非结构化存储，准确度依赖文本内容 | **不一致** | patient_core_service.go:239 合并3字段做关键词检测 | 老库存自由文本，新系统应结构化存储。当前为兼容实现，检测逻辑需覆盖更多关键词变体 | 否 |
| 7 | P0 | OverviewTab | OverviewTab.tsx:127-153 | 核心方案 | `treatmentPlan.dialysisMode/dryWeight/bloodFlow` | `GET /api/v1/patients/:id/core` | patient_core_service.go:251-289 | `Plan_PatientPlan`多字段 | `Plan_PatientPlan.DialysisMethod/DryWeight/BF`等 | buildCurrentPlan仅读取部分字段(OddWeekFrequency, EvenWeekFrequency, DialysisMethod, DialysisDuration, DryWeight, BF, FirstAnticoagulant, MaintainAnticoagulant, Note)。缺失：ExtraWeight, BV, DilutionProportion, 多数透析液参数 | **缺失** | patient_core_service.go:28 声明了部分字段; 老库Plan_PatientPlan有48字段 | buildCurrentPlan应扩展读取更多Plan_PatientPlan字段以支持完整方案展示 | 否 |
| 8 | P0 | BasicInfoTab | BasicInfoTab.tsx:197-241 | 保存请求体 | `personalInfo.gender` (中文→M/F) | `PUT /api/v1/patients/:id/basic-info` | patient_basic_service.go:151-153 | `Gender`(varchar 64) | `Register_PatientInfomation.Gender`(varchar 64) | 前端chineseToGender将"男"→"M","女"→"F"写入。老库存中文还是代码取决于原始数据，可能不一致 | **待确认** | BasicInfoTab.tsx:58-66 chineseToGender; patient_basic_service.go:152 | 需确认老库Gender列实际存储的是"M/F"还是"男/女"。如是中文则chineseToGender写入会破坏数据 | **是** |
| 9 | P0 | BasicInfoTab | BasicInfoTab.tsx:207-216 | 医疗登记-保存 | `medicalInfo.visitCategory/admissionNo/visitNo/medicalRecordNo` | `PUT /api/v1/patients/:id/basic-info` | patient_basic_service.go:251-271 | 写入`Register_Hospitalization` | `Register_Hospitalization.HospPatientType/HospNo/CaseNo/MedicalRecordNo` | 后端将visitCategory→HospPatientType, admissionNo→HospNo, visitNo→CaseNo, medicalRecordNo→MedicalRecordNo。字段名映射正确 | **一致** | patient_basic_service.go:253-263 | — | 否 |
| 10 | P0 | BasicInfoTab | BasicInfoTab.tsx:202 | 证件类型 | `personalInfo.idType` | `PUT /api/v1/patients/:id/basic-info` | patient_basic_service.go:274-276 | 写入`Register_IDInfomation.IDType` | `Register_IDInfomation.IDType`(varchar 64) | 前端传字典code(如"居民身份证")，后端直接写入IDType。字典值与老库枚举需一致 | **待确认** | dictApi.ts:207 ID_TYPE; 老库IDType枚举值 | 需确认前端字典ID_TYPE的code值与老库Register_IDInfomation.IDType存储的枚举值是否完全匹配 | **是** |
| 11 | P0 | BasicInfoTab | BasicInfoTab.tsx:214-215 | 医保类型 | `medicalInfo.insuranceType` | `PUT /api/v1/patients/:id/basic-info` | patient_basic_service.go:166-168 | 写入`Register_PatientInfomation.ExpenseType` | `Register_PatientInfomation.ExpenseType`(varchar 64) | 前端传字典code，后端写入ExpenseType。映射文档：INSURANCE_TYPE→ExpenseType | **一致** | patient_basic_service.go:167; 固定映射文档 | — | 否 |
| 12 | P0 | BasicInfoTab | BasicInfoTab.tsx:225 | 干体重写入 | `vitalSocialInfo.dryWeight` | `PUT /api/v1/patients/:id/basic-info` | patient_basic_service.go:187-189 | 写入`Register_PatientInfomation.Weight` | `Register_PatientInfomation.Weight`(numeric) | 前端dryWeight写入老库Weight列。Weight在老库语义为"体重"，不是"干体重"。dryWeight应写入Plan_PatientPlan.DryWeight | **不一致** | patient_basic_service.go:188 Weight=*dryWeight; 老库Weight="体重" | dryWeight不应写入Register.Weight。应单独更新Plan_PatientPlan.DryWeight | 否 |
| 13 | P0 | TreatmentPlanTab | TreatmentPlanTab.tsx:501-700+ | 治疗方案保存 | 透析参数/抗凝剂/透析液参数 | `PUT /api/v1/patients/:id/treatment-plan` | patient_handler.go:291+ → treatment_plan_service.go | 写入`Plan_PatientPlan` | `Plan_PatientPlan`48字段 | 方案保存涉及大量Plan_PatientPlan字段。前端发送dialysisMode/anticoagulant/parameters/materials嵌套结构，后端需展平写入Plan_PatientPlan各列 | **一致** | patientApi.ts:64-89 TreatmentPlan类型; Plan_PatientPlan字段 | — | 否 |
| 14 | P0 | SchemeOrderTab | SchemeOrderTab.tsx:51-71 | 医嘱列表字段 | `order.name/dose/unit/route/frequency/timing` | `GET /api/v1/patients/:id/orders` | order_handler.go → order_service.go | 读取`Order_PatientOrder` | `Order_PatientOrder.Content/Dosage/UseMethod/UseWay`等 | 前端Order类型有name/content/dose/unit/route/timing/frequency等字段。老库Order_PatientOrder的字段名不同(Content/Dosage/UseMethod等) | **不一致** | orderApi.ts:54-82 Order类型; 老库Content/Dosage/UseMethod | 后端service需做字段名映射：Content→name/content, Dosage→dose, UseMethod→route, UseWay→timing等 | 否 |
| 15 | P0 | SchemeOrderTab | SchemeOrderTab.tsx:65 | 医嘱类型 | `order.type` ('长期'/'临时') | `GET /api/v1/patients/:id/orders` | order_service.go | `Order_PatientOrder.Type/Classification` | `Order_PatientOrder.Type`(varchar), `Classification`(varchar) | 后端buildActiveOrders优先用Classification，fallback到Type。老库存储的值需确认是"长期/临时"还是其他编码 | **待确认** | patient_core_service.go:305-308 | 需确认老库Order_PatientOrder.Type/Classification实际存储的枚举值 | **是** |
| 16 | P0 | VascularTab | VascularTab.tsx:21-46 | 血管通路字段映射 | `accessType/site/artery/vein/side` | `GET /api/v1/patients/:id/vascular-accesses` | patient.go:110-137 VascularAccess模型 | `Register_VascularAccess`各字段 | `Register_VascularAccess.AccessType/AccessPosition/Artery/Venous/LeftAndRight` | 模型映射：AccessType→accessType, AccessPosition→site, Artery→artery, Venous→vein, LeftAndRight→side | **一致** | patient.go:114-118 gorm column映射 | — | 否 |
| 17 | P0 | VascularTab | VascularTab.tsx:369-370 | 动脉/静脉显示 | `artery[]`/`vein[]` (数组) | GET api | patient.go:116-117 | `Artery`(varchar 128)/`Venous`(varchar 128) | `Register_VascularAccess.Artery`(varchar 128)/`Venous`(varchar 128) | 前端artery/vein为string[]数组，老库存单个varchar。需确认后端是否做了逗号分隔↔数组转换 | **待确认** | VascularTab.tsx:369 dictArrayNames处理数组; patient.go:116-117 单string | 需确认后端VascularAccess service是否将老库逗号分隔字符串拆分为数组 | **是** |

### P1 — 字典/枚举一致性

| # | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|---|--------|----------|-------------|--------|-------------|------|------|------------|-----------|
| 18 | P1 | PatientList | PatientList.tsx:316-320 | 患者类型显示 | `patient.patientType` 通过`DICT_TYPES.PATIENT_TYPE`映射 | **一致** | dictApi.ts:206 PATIENT_TYPE; legacy_enum_maps.go:279-285 门诊→10/住院→20 | — | 否 |
| 19 | P1 | PatientList | PatientList.tsx:320 | 医保类型显示 | `patient.insuranceType` 通过`DICT_TYPES.INSURANCE_TYPE`映射 | **一致** | dictApi.ts:205 INSURANCE_TYPE; 后端fallback映射文档 | — | 否 |
| 20 | P1 | BasicInfoTab | BasicInfoTab.tsx:38-49 | 字典加载 | ID_TYPE/PATIENT_TYPE/VISIT_CATEGORY/BLOOD_TYPE_ABO/RH/EDUCATION_LEVEL/MARITAL_STATUS/INSURANCE_TYPE/RELATIONSHIP_OPTIONS | **一致** | dictApi.ts:207-213 对应字典类型代码; 字典类型对照表(1).md | — | 否 |
| 21 | P1 | BasicInfoTab | BasicInfoTab.tsx:368 | 性别显示 | `genderToChinese(info.gender)` M→男/F→女 | **一致** | BasicInfoTab.tsx:52-56 | — | 否 |
| 22 | P1 | TreatmentPlanTab | TreatmentPlanTab.tsx:204-211 | 透析模式/透析液/葡萄糖字典 | DIALYSIS_MODE/DIALYSATE_TYPE/DIALYSATE_GROUP/DIALYSATE_FLOW/GLUCOSE/DRUG_CATEGORY | **一致** | dictApi.ts:189-197 | — | 否 |
| 23 | P1 | OverviewTab | OverviewTab.tsx:159-163 | 感染值判断 | `value === '阳性' \|\| String(value) === 'Positive'` | **不一致** | OverviewTab.tsx:165 | 后端buildInfection返回的值可能是"阴性"或包含关键词的原文。前端判断逻辑需与后端返回格式对齐 | 否 |
| 24 | P1 | VascularTab | VascularTab.tsx:156-164 | 通路类型/部位/动静脉/手术类型字典 | VASCULAR_ACCESS/VASCULAR_SITE/ARTERY_TYPE/VEIN_TYPE/HOSPITAL/DOCTOR/SURGERY_TYPE | **一致** | dictApi.ts:201-204, 214-217 | — | 否 |

### P1 — 接口传输与后端存储问题

| # | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|---|--------|----------|-------------|--------|-------------|---------|-------------|----------|--------------|-------------|------|------|------------|-----------|
| 25 | P1 | PatientList | PatientList.tsx:353 | 列表行-医生名 | `patient.doctorName` | `GET /api/v1/patients` | patient_basic_service.go:498-509 | `ResponsibilityDrId`(bigint) → 解析为名字 | `Register_PatientInfomation.ResponsibilityDrId`(bigint) | 后端通过resolveEmployeeName将DrId转为名字。列表接口需确认是否已做此转换 | **不一致** | patient.go:68 DoctorName gorm:"-"; 需JOIN Organ_Employee | 列表接口应JOIN Organ_Employee解析医生名，或在列表service中单独查询 | 否 |
| 26 | P1 | PatientList | PatientList.tsx:338-344 | 列表行-状态 | `patient.status` | `GET /api/v1/patients` | patient.go:53 | `TreatmentStatus`(varchar 64) | `Register_PatientInfomation.TreatmentStatus`(varchar 64) | 前端展示"透析中/候诊/居家"等状态。老库TreatmentStatus存储"在院/出院/死亡"。后端buildHeader做了active→透析中/治疗中转换 | **不一致** | patient_core_service.go:137-149; 老库TreatmentStatus="在院/出院/死亡" | 需确认TreatmentStatus与前端status值的完整映射关系 | **是** |
| 27 | P1 | OverviewTab | OverviewTab.tsx:272-283 | 活跃医嘱 | `orders[].content/type/startTime/doctor` | `GET /api/v1/patients/:id/core` | patient_core_service.go:293-317 | `Order_PatientOrder.Content/Type/Classification/StartTime/OperatorId` | `Order_PatientOrder`对应字段 | 医嘱仅有5个字段(ID/Content/Type/StartTime/Doctor)。缺失：Dosage/UseMethod/UseWay/Note等 | **缺失** | patient_core_service.go:309-315; 老库Order_PatientOrder有更多字段 | buildActiveOrders应扩展返回更多医嘱字段 | 否 |
| 28 | P1 | BasicInfoTab | BasicInfoTab.tsx:143-169 | 基本信息读取 | 调用`restApi.getPatientBasicInfo()` | `GET /api/v1/patients/:id/basic-info` | patient_basic_service.go:77-124 | 读取4张表: Register_PatientInfomation + Register_Hospitalization + Register_IDInfomation + Register_FamilyMember | 老库4张表 | 读取逻辑正确，从4张legacy表聚合数据。但PatientBasicInfo模型(gorm:"-")中的字段为计算字段 | **一致** | patient_basic_service.go:97-112 查询4张表 | — | 否 |
| 29 | P1 | BasicInfoTab | BasicInfoTab.tsx:228-230 | 职业/婚姻/文化程度 | `occupation/maritalStatus/educationLevel` | `PUT /api/v1/patients/:id/basic-info` | patient_basic_service.go:196-204 | 写入`Register_PatientInfomation.Occupation/MaritalStatus/EducationLevel` | 老库对应字段(varchar 64/64/64) | 字段映射正确，存code值。前端通过字典显示名称 | **一致** | patient_basic_service.go:196-204 | — | 否 |
| 30 | P1 | LabsExamsTab | LabsExamsTab.tsx:116-131 | 检验报告 | `labReports[].items[].itemCode/itemName/resultValue/unit/referenceRange/abnormalFlag` | `GET /api/v1/patients/:id/lab-reports` | lab_report_service.go | `LIS_ExaminationItem.ItemCode/ItemName/Result/Unit/Reference/ResultSign` + `LIS_Examination` | `LIS_ExaminationItem`对应字段 | 后端JOIN查询LIS_ExaminationItem+LIS_Examination。字段映射：ItemCode→itemCode, Result→resultValue, ResultSign→abnormalFlag | **一致** | patient_core_service.go:343-344 SELECT语句 | — | 否 |
| 31 | P1 | LabsExamsTab | LabsExamsTab.tsx:132-145 | 关键指标 | `keyIndicators[].indexCode/indexName/result/unit/reference/resultSign` | `GET /api/v1/patients/:id/key-indicators` | key_indicator_service.go | 关键指标存储表 | 待确认关键指标存储表 | 关键指标API独立于检验报告，数据源待确认 | **待确认** | LabsExamsTab.tsx:132 | 需确认关键指标API的后端数据源表 | **是** |
| 32 | P1 | MedicalRecordTab | MedicalRecordTab.tsx:89-119 | 临床病史读取 | `medicalHistory.current/past/transfusion/marital/family/diagnosis/primary/pathology/allergen/tumor/complication` | `GET /api/v1/patients/:id/medical-history` | medical_history_service.go | 聚合多表: Register_MedicalHistory + Register_Protopathy + Register_Pathology + Register_Allergen + Register_Tumor + Register_Complication + Register_Diagnosis | 老库对应表 | 后端需从7+张表聚合病史数据。当前medicalHistory响应结构为嵌套对象 | **一致** | MedicalRecordTab.tsx:55-67 EMPTY_MEDICAL_HISTORY结构 | — | 否 |
| 33 | P1 | MedicalRecordTab | MedicalRecordTab.tsx:85-86, 156-189 | 转归记录 | `outcomeRecords[].type/reason/time/remarks` | `GET/POST /api/v1/patients/:id/outcome-records` | outcome_service.go | `Register_OutCome.Type/Reason/OutComeTime/Note` | `Register_OutCome.Type/Reason/OutComeTime/Note` | 字段映射：Type→type, Reason→reason, OutComeTime→time, Note→remarks | **一致** | 老库Register_OutCome字段 | — | 否 |
| 34 | P1 | HistoryTab | HistoryTab.tsx:42-67 | 治疗历史 | `treatments[].treatmentDate/timeRange/treatmentType/durationMinutes/weightLossKg/startBp/endBp/complications/doctorName` | `GET /api/v1/treatments?patientId=xxx` | treatment_service.go → treatment_handler.go | 读取`treatment_treatment`相关表 | `Treatment_TreatmentRecord`等表 | 治疗历史从Treatment相关表读取。字段映射需确认 | **待确认** | HistoryTab.tsx:46-63; 老库Treatment表结构 | 需确认Treatment表结构与前端字段的完整映射 | **是** |
| 35 | P1 | MonthlySummaryTab | MonthlySummaryTab.tsx:14-226 | 月份评估小结 | 所有字段(自理能力/睡眠/饮食/营养/尿量/用药依从/血压监测/血糖监测/血流量/干体重/间期增重/透析充分性/iPTH/钙/磷等) | 无API调用 | 无后端实现 | 无对应后端 | `Treatment_TreatmentMonthSummarySheet` | 整个Tab为纯前端硬编码/静态UI，无API调用，无后端实现，无数据持久化。老库有Treatment_TreatmentMonthSummarySheet表 | **只读未闭环** | MonthlySummaryTab.tsx无任何API调用; 老库有对应表 | 需实现月份小结的读写API，对接Treatment_TreatmentMonthSummarySheet | 否 |

### P2 — 组件/UI层级问题

| # | 优先级 | 页面/模块 | 前端文件:行号 | 功能点 | 前端字段/操作 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|---|--------|----------|-------------|--------|-------------|------|------|------------|-----------|
| 36 | P2 | PatientHeader | PatientHeader.tsx:45-57 | 患者头部信息 | name/gender/age/insuranceType/doctorName/defaultMode/status/bedNumber/riskLevel | **一致** | 所有字段来自Patient模型 | — | 否 |
| 37 | P2 | ClinicalFocusDrawer | ClinicalFocusDrawer.tsx:34-46 | 临床焦点抽屉 | diagnosis/riskLevel/infection/allergies + dialysisParams | **一致** | 数据来自Patient模型 | — | 否 |
| 38 | P2 | FocusPanel | FocusPanel.tsx:14-26 | 临床焦点面板 | 与ClinicalFocusDrawer相同字段 | **一致** | 同上 | — | 否 |
| 39 | P2 | FloatingActionButtons | FloatingActionButtons.tsx:1-59 | 悬浮按钮 | hasRisk/todoCount | **不一致** | todoCount硬编码为3(TodoPopover.tsx:19-44为MOCK数据) | 待办任务应从后端API获取真实数据 | 否 |
| 40 | P2 | TodoPopover | TodoPopover.tsx:19-44 | 待办任务列表 | 3条硬编码MOCK数据(保存后同步处方/目标超滤待复核/院感数据待同步) | **缺失** | TodoPopover.tsx:19 MOCK_TASKS | 需实现待办任务API，替换MOCK数据 | 否 |

---

## 三、重点问题详解

### 3.1 Patient.DryWeight → Register.Weight 映射错误 (P0 #4, #12)

**问题**：`models.Patient.DryWeight` 映射到 `Register_PatientInfomation.Weight`。

- 老库 `Weight` 字段语义为"体重"（当前体重）
- 老库"干体重"存储在 `Plan_PatientPlan.DryWeight`
- 前端 `dryWeight` 含义为"干体重"

**影响**：
1. 患者列表显示的"干体重"实际是老库的"体重"
2. BasicInfoTab 保存干体重时写入了 Weight 列，覆盖了原始体重数据

**修复方向**：
- `Patient.DryWeight` 应从 `Plan_PatientPlan.DryWeight` 读取
- 保存时应更新 `Plan_PatientPlan.DryWeight` 而非 `Register_PatientInfomation.Weight`
- 可考虑增加 `Patient.Weight` 字段映射到老库 Weight

### 3.2 Patient.Diagnosis → Register.Note 映射语义不匹配 (P0 #5)

**问题**：`models.Patient.Diagnosis` 映射到 `Register_PatientInfomation.Note`。

- 老库 `Note` 语义为"备注"
- 老库存储诊断信息的表为 `Register_Diagnosis.DiagnosisDesc`

**修复方向**：
- 读取时：优先从 `Register_Diagnosis` 表取最新诊断，fallback 到 Note
- 保存时：诊断写入 `Register_Diagnosis` 表

### 3.3 感染信息非结构化存储 (P0 #6)

**问题**：后端 `buildInfection` 将 `Register_Infection` 的 `InfectionDesc + OtherDesc + Note` 合并后做关键词匹配。

- 存储方式：自由文本（如"HBsAg阳性，HCV-Ab阴性"）
- 检测方式：关键词匹配（"hbsag"/"hbv"/"乙肝"等）
- 准确度依赖文本内容，可能漏检或误判

**修复方向**：
- 新系统应结构化存储感染信息（独立字段或关联表）
- 当前兼容实现需扩充关键词库

### 3.4 Gender 存储格式待确认 (P0 #8)

**问题**：前端保存时通过 `chineseToGender` 将"男"→"M"、"女"→"F"写入老库 Gender 列。

- 如果老库存储的是中文（"男"/"女"），则写入"M"/"F"会破坏数据
- 如果老库存储的是代码（"M"/"F"），则映射正确

**需人工确认**：查询老库 `Register_PatientInfomation.Gender` 列的实际数据样本。

### 3.5 月份小结完全未实现 (P1 #35)

**问题**：`MonthlySummaryTab.tsx` 整个组件为纯前端静态 UI，无任何 API 调用。

- 老库有 `Treatment_TreatmentMonthSummarySheet` 表
- 前端字段包括：自理能力、睡眠、饮食、营养、尿量、血压监测、血糖监测、血流量、干体重、透析充分性、iPTH、钙、磷、转归等
- 所有 RadioGroup/SmallInput 组件的值未持久化

**修复方向**：实现 `/api/v1/patients/:id/monthly-summaries` CRUD API。

---

## 四、一致项速查

以下字段/模块经核查确认一致，无需改造：

| 模块 | 一致字段 |
|------|---------|
| PatientList | patientType(字典映射), insuranceType(字典映射), name, gender, id |
| BasicInfoTab | name→Name, pinyin→Spell, birthday→BirthDate, ethnicity→Nation, aboBloodType→ABOType, rhBloodType→RHType, height→Height, educationLevel→EducationLevel, occupation→Occupation, maritalStatus→MaritalStatus, workplace→Workunit, phone→PhoneNo, wechat→WeChatNo, landline→HomePhoneNo, address→Address, visitCategory→HospPatientType, admissionNo→HospNo, visitNo→CaseNo, medicalRecordNo→MedicalRecordNo, insuranceType→ExpenseType, insuranceNo→SSN, dialysisNo→DialysisNo, firstDialysisDate→FirstDialysisDate, firstDialysisHospital→FirstDialysisHospital, idType→Register_IDInfomation.IDType, idNumber→Register_IDInfomation.IDNo |
| TreatmentPlanTab | dialysisMode→DialysisMethod, duration→DialysisDuration, dryWeight→DryWeight, bloodFlow→BF, 频次→OddWeekFrequency/EvenWeekFrequency, 透析液参数→NaIonCon/CaIonCon/KIonCon/HCO3IonCon/Conductivity/DialysateTmp/DialysateFlow |
| VascularTab | accessType→AccessType, site→AccessPosition, artery→Artery, vein→Venous, side→LeftAndRight, hospital→OperationHospital, surgeon→OperationDr, surgeryDate→OperationTime, firstUseDate→FirstUseTime, accessNumber→AccessCount, interventionCount→InterveneCount, catheterMethod→CatheterizeMethod, catheterDepth→CatheterDepth, notes→Note, isDefault→IsDefault, isDisabled→IsDisabled |
| LabsExamsTab | labReport items: ItemCode/ItemName/Result/Unit/Reference/ResultSign |
| MedicalRecordTab | outcomeRecords: Type→type, Reason→reason, OutComeTime→time, Note→remarks |
| OverviewTab | infection: InfectionDesc+OtherDesc+Note→hbsag/hcvab/hivab/tpab; plan: DialysisMethod→dialysisMode, DryWeight→dryWeight, BF→bloodFlow |

---

## 五、待人工确认清单

| # | 确认项 | 涉及文件 | 确认方式 |
|---|--------|---------|---------|
| 1 | 老库 `Register_PatientInfomation.Gender` 存储的是"M/F"还是"男/女" | patient_basic_service.go:152 | SQL查询: `SELECT DISTINCT "Gender" FROM "Register_PatientInfomation" LIMIT 20` |
| 2 | 老库 `Register_PatientInfomation.TreatmentStatus` 的实际枚举值 | patient_core_service.go:137-149 | SQL查询: `SELECT DISTINCT "TreatmentStatus" FROM "Register_PatientInfomation"` |
| 3 | 老库 `Order_PatientOrder.Type/Classification` 的实际枚举值 | patient_core_service.go:305-308 | SQL查询: `SELECT DISTINCT "Type","Classification" FROM "Order_PatientOrder"` |
| 4 | 前端字典 `ID_TYPE` 的 code 值与老库 `Register_IDInfomation.IDType` 枚举是否一致 | BasicInfoTab.tsx:204 | 对比字典API返回值与老库: `SELECT DISTINCT "IDType" FROM "Register_IDInfomation"` |
| 5 | 前端字典 `PATIENT_TYPE` 的 code 值与老库 `PatientType` 存储值的映射是否完整 | legacy_enum_maps.go:279-285 | 确认"10→门诊/20→住院"是否覆盖所有老库值 |
| 6 | `VascularAccess.artery/vein` 老库存储格式（逗号分隔？单值？）及后端转换逻辑 | patient.go:116-117 | SQL查询: `SELECT "Artery","Venous" FROM "Register_VascularAccess" LIMIT 20` |
| 7 | 关键指标API (`/key-indicators`) 的后端数据源表 | LabsExamsTab.tsx:132 | 查看 key_indicator_service.go |
| 8 | 治疗历史API的后端数据源表结构 | HistoryTab.tsx:42 | 查看 treatment_service.go |

---

## 六、改造优先级建议

### 第一批（P0，数据正确性）
1. 修复 `DryWeight` 映射：从 `Plan_PatientPlan.DryWeight` 读取，保存到 Plan 表
2. 修复 `Diagnosis` 映射：从 `Register_Diagnosis` 读取
3. 确认 `Gender` 存储格式，修正 `chineseToGender` 逻辑
4. 确认 `TreatmentStatus` 枚举映射完整性

### 第二批（P1，功能完整性）
5. 扩展 `buildActiveOrders` 返回字段
6. 扩展 `buildCurrentPlan` 读取字段
7. 实现月份小结 CRUD API
8. 替换待办任务 MOCK 数据

### 第三批（P2，体验优化）
9. 列表接口优化：JOIN 查询 bedNumber/doctorName/defaultMode
10. 血管通路 artery/vein 数组 ↔ 字符串转换标准化
