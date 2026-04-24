# AI-HMS 数据库设计文档

> 版本: 1.0 | 更新日期: 2026-03-05 | 数据库: PostgreSQL | ORM: GORM v2

---

## 1. 总览

### 1.1 系统说明

AI-HMS（AI Hemodialysis Management System）是一套透析中心管理系统，后端使用 Go + Gin + GORM，数据库为 PostgreSQL。

### 1.2 表清单（共 30 张表）

| 序号 | 表名 | 模块 | 说明 | PK 类型 |
|------|------|------|------|---------|
| 1 | `users` | 用户 | 系统用户与角色 | UUID (varchar 36) |
| 2 | `patients` | 患者 | 患者主表 | UUID |
| 3 | `patient_basic_infos` | 患者 | 患者档案扩展（1:1） | UUID |
| 4 | `medical_histories` | 患者 | 临床病史档案（1:1） | UUID |
| 5 | `infection_infos` | 患者 | 传染病标记（1:1） | UUID |
| 6 | `vascular_accesses` | 患者 | 血管通路（1:N） | UUID |
| 7 | `vascular_access_interventions` | 患者 | 血管通路干预记录（1:N） | UUID |
| 8 | `outcome_records` | 患者 | 治疗转归记录（1:N） | UUID |
| 9 | `hospitalizations` | 住院 | 住院信息 | bigint auto |
| 10 | `treatment_plans` | 治疗方案 | 透析治疗方案 | UUID |
| 11 | `orders` | 医嘱 | 长期/临时医嘱 | UUID |
| 12 | `prescriptions` | 处方 | 每日透析处方（医嘱单） | UUID |
| 13 | `adjustment_records` | 治疗方案 | 方案调整记录 | UUID |
| 14 | `wards` | 排班 | 病区/病房 | bigint auto |
| 15 | `beds` | 排班 | 床位 | bigint auto |
| 16 | `shifts` | 排班 | 班次定义 | bigint auto |
| 17 | `patient_shifts` | 排班 | 患者排班 | bigint auto |
| 18 | `Treatment_Treatment` | 透析执行 | 透析治疗主表 | bigint auto |
| 19 | `Treatment_BeforeCheck` | 透析执行 | 透前检查 | bigint auto |
| 20 | `Treatment_BeforeSigns` | 透析执行 | 透前体征 | bigint auto |
| 21 | `Treatment_DuringParam` | 透析执行 | 透析中参数 | bigint auto |
| 22 | `Treatment_AfterSigns` | 透析执行 | 透后体征 | bigint auto |
| 23 | `Treatment_Alarm` | 透析执行 | 报警记录 | bigint auto |
| 24 | `plan_templates` | 配置 | 治疗方案模板 | UUID |
| 25 | `material_catalogs` | 主数据 | 材料目录 | uint auto |
| 26 | `drug_catalogs` | 主数据 | 药品目录 | uint auto |
| 27 | `order_templates` | 配置 | 医嘱模板 | UUID |
| 28 | `order_template_items` | 配置 | 医嘱模板条目 | UUID |
| 29 | `dict_types` | 字典 | 字典类型 | UUID |
| 30 | `dict_items` | 字典 | 字典项 | UUID |
| 31 | `lab_reports` | 检验 | 检验报告主表 | UUID |
| 32 | `lab_report_items` | 检验 | 检验报告明细 | UUID |
| 33 | `exam_reports` | 检查 | 检查报告 | UUID |
| 34 | `patient_key_indicators` | 指标 | 患者关键指标 | UUID |
| 35 | `integration_hdis_settings` | 集成 | HDIS 对接配置 | UUID |

### 1.3 PK 策略

| 策略 | 适用表 | 说明 |
|------|--------|------|
| **UUID (varchar 36)** | 患者域、治疗方案域、字典域 | 应用层生成 `uuid.New()` |
| **bigint auto** | 排班域、透析执行域、住院 | 数据库自增 |
| **uint auto** | 主数据（药品/材料目录） | 数据库自增 |

### 1.4 迁移策略

- **开发环境**: `database.AutoMigrate()` 自动建表/加列，启动时执行
- **生产环境**: 跳过 AutoMigrate，使用 `scripts/` 目录下的 SQL 脚本手动迁移
- 外键约束在迁移时**禁用** (`DisableForeignKeyConstraintWhenMigrating: true`)，关联关系仅在 GORM 层维护

---

## 2. ER 关系图

```
┌─────────────────────────────────────────────────────────────────────┐
│                          患者域                                      │
│                                                                     │
│  patients ──1:1── patient_basic_infos                               │
│     │       ──1:1── medical_histories                               │
│     │       ──1:1── infection_infos                                  │
│     │       ──1:N── vascular_accesses ──1:N── vascular_access_      │
│     │       │                                  interventions        │
│     │       ──1:N── outcome_records                                  │
│     │       ──1:N── hospitalizations                                 │
│     │       ──1:N── lab_reports ──1:N── lab_report_items            │
│     │       ──1:N── exam_reports                                     │
│     │       ──1:N── patient_key_indicators                          │
│     │                                                               │
└─────┼───────────────────────────────────────────────────────────────┘
      │
┌─────┼───────────────────────────────────────────────────────────────┐
│     │              治疗方案域                                         │
│     │                                                               │
│     ├──1:N── treatment_plans ◄───── prescriptions (FK: treatmentPlanId)
│     │                                                               │
│     ├──1:N── orders                                                  │
│     │                                                               │
│     └──1:N── adjustment_records                                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                         排班域                                       │
│                                                                     │
│  wards ──1:N── beds                                                  │
│  shifts ──1:N── patient_shifts ──N:1── patients                     │
│                    │                                                 │
│                    └──N:1── beds, wards                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                      透析执行域                                      │
│                                                                     │
│  Treatment_Treatment ──1:1── Treatment_BeforeCheck                   │
│         │            ──1:1── Treatment_BeforeSigns                   │
│         │            ──1:1── Treatment_AfterSigns                    │
│         │            ──1:N── Treatment_DuringParam                   │
│         │            ──1:N── Treatment_Alarm                         │
│         │                                                           │
│         └──N:1── patients, patient_shifts, wards, beds, shifts      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    配置/主数据域                                      │
│                                                                     │
│  plan_templates          (治疗方案模板，含 JSONB templateContent)     │
│  drug_catalogs           (药品目录)                                  │
│  material_catalogs       (材料目录)                                  │
│  order_templates ──1:N── order_template_items                        │
│  dict_types ──1:N── dict_items                                       │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. 各表详细设计

### 3.1 用户模块

#### `users` — 系统用户

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| username | varchar(50) | UNIQUE, NOT NULL | 用户名 |
| password | varchar(255) | NOT NULL | 密码哈希（JSON 序列化时隐藏） |
| real_name | varchar(50) | | 真实姓名 |
| phone | varchar(20) | | 手机号 |
| email | varchar(100) | | 邮箱 |
| role | varchar(50) | NOT NULL | 角色代码 |
| status | varchar(20) | DEFAULT 'active' | active / inactive |
| department_id | varchar(36) | | 科室 ID |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**角色枚举值:**

| 代码 | 含义 |
|------|------|
| `ADMIN` | 系统管理员 |
| `DOCTOR_CHIEF` | 主任医师 |
| `DOCTOR_SUPERVISOR` | 主治医师 |
| `DOCTOR_DUTY` | 值班医师 |
| `NURSE_HEAD` | 护士长 |
| `NURSE_MANAGER` | 护理组长 |
| `NURSE_RESPONSIBLE` | 责任护士 |
| `ENGINEER` | 工程师 |

---

### 3.2 患者模块

#### `patients` — 患者主表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| name | varchar(50) | NOT NULL | 患者姓名 |
| age | int | NOT NULL | 年龄 |
| gender | varchar(10) | NOT NULL | M / F (ISO 5218) |
| bed_number | varchar(20) | | 床位号 |
| diagnosis | text | | 诊断 |
| risk_level | varchar(20) | DEFAULT '低危' | 高危 / 中危 / 低危 |
| status | varchar(20) | DEFAULT 'active' | active / inactive / discharged |
| patient_type | varchar(50) | | 门诊 / 住院 |
| insurance_type | varchar(50) | | 医保类型 |
| dry_weight | decimal(5,2) | | 干体重 (kg) |
| default_mode | varchar(50) | | 默认透析模式 |
| doctor_id | varchar(36) | | 主管医生 ID |
| doctor_name | varchar(50) | | 主管医生姓名 |
| admission_date | timestamp | | 入院日期 |
| discharge_date | timestamp | | 出院日期 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**GORM 关联:** `VascularAccesses` (1:N), `MedicalHistory` (1:1), `TreatmentPlan` (1:1)

---

#### `patient_basic_infos` — 患者档案扩展（1:1）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | UNIQUE, NOT NULL | 关联 patients.id |
| pinyin | varchar(100) | | 姓名拼音 |
| birthday | timestamp | | 出生日期 |
| ethnicity | varchar(20) | | 民族 |
| id_type | varchar(20) | DEFAULT '身份证' | 身份证 / 护照 / 其他 |
| id_number | varchar(50) | | 证件号码 |
| visit_category | varchar(20) | | 门诊 / 住院 / 急诊 |
| admission_no | varchar(50) | | 住院号 |
| visit_no | varchar(50) | | 就诊号 |
| medical_record_no | varchar(50) | | 病历号 |
| insurance_no | varchar(50) | | 医保号 |
| hdis_patient_id | int | UNIQUE | HDIS/LIS 外部系统 ID |
| dialysis_no | varchar(50) | | 透析号 |
| nurse_name | varchar(50) | | 责任护士 |
| first_dialysis_date | timestamp | | 首次透析日期 |
| first_hospital_date | timestamp | | 首次在本院透析日期 |
| first_dialysis_hospital | varchar(100) | | 首次透析医院 |
| height | varchar(10) | | 身高 (cm) |
| abo_blood_type | varchar(10) | | A / B / AB / O |
| rh_blood_type | varchar(10) | | Rh+ / Rh- |
| education_level | varchar(20) | | 文化程度 |
| occupation | varchar(50) | | 职业 |
| marital_status | varchar(20) | | 婚姻状况 |
| workplace | varchar(100) | | 工作单位 |
| phone | varchar(20) | | 手机号码 |
| wechat | varchar(50) | | 微信号 |
| landline | varchar(20) | | 固定电话 |
| address | text | | 地址 |
| district | varchar(100) | | 区域 |
| contact_name | varchar(50) | | 紧急联系人 |
| contact_phone | varchar(20) | | 紧急联系电话 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

---

#### `medical_histories` — 临床病史档案（1:1）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | UNIQUE, NOT NULL | 关联 patients.id |
| current_illness | text | | 现病史 |
| past_history | text | | 既往史 |
| transfusion_history | text | | 输血史 |
| marital_history | text | | 婚育史 |
| family_history | text | | 家族史 |
| disease_diagnosis | text | | 疾病诊断 |
| primary_disease_name | varchar(255) | | 原发病名称 |
| primary_disease_content | text | | 原发病详情 |
| primary_disease_type | varchar(255) | | 原发病分类 |
| primary_disease_check_time | varchar(32) | | 原发病检查时间 |
| primary_disease_check_doc | varchar(100) | | 原发病检查医生 |
| pathology_name | varchar(255) | | 病理诊断名称 |
| pathology_content | text | | 病理诊断详情 |
| pathology_type | varchar(255) | | 病理诊断分类 |
| pathology_check_time | varchar(32) | | 病理检查时间 |
| pathology_check_doc | varchar(100) | | 病理检查医生 |
| allergen_name | varchar(255) | | 过敏信息名称 |
| allergen_content | text | | 过敏信息详情 |
| allergen_type | varchar(255) | | 过敏原分类 |
| allergen_check_time | varchar(32) | | 过敏检查时间 |
| allergen_check_doc | varchar(100) | | 过敏检查医生 |
| tumor_history_name | varchar(255) | | 肿瘤病史名称 |
| tumor_history_content | text | | 肿瘤病史详情 |
| tumor_history_type | varchar(255) | | 肿瘤分类 |
| tumor_history_check_time | varchar(32) | | 肿瘤检查时间 |
| tumor_history_check_doc | varchar(100) | | 肿瘤检查医生 |
| complication_name | varchar(255) | | 并发症名称 |
| complication_content | text | | 并发症详情 |
| complication_type | varchar(255) | | 并发症分类 |
| complication_check_time | varchar(32) | | 并发症检查时间 |
| complication_check_doc | varchar(100) | | 并发症检查医生 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

---

#### `infection_infos` — 传染病标记（1:1）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | UNIQUE, NOT NULL | 关联 patients.id |
| hbsag | varchar(10) | DEFAULT '阴性' | 乙肝表面抗原 |
| hcvab | varchar(10) | DEFAULT '阴性' | 丙肝抗体 |
| hivab | varchar(10) | DEFAULT '阴性' | HIV 抗体 |
| tpab | varchar(10) | DEFAULT '阴性' | 梅毒抗体 |
| tb | varchar(10) | | 结核 |
| update_date | timestamp | | 检测更新日期 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**检测值:** `阴性` / `阳性`

---

#### `vascular_accesses` — 血管通路（1:N）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | 关联 patients.id |
| access_type | varchar(50) | NOT NULL | AVF / AVG / TCC / NCC |
| site | varchar(100) | | 通路部位 |
| artery | text (JSON) | | 动脉（JSON 字符串数组） |
| vein | text (JSON) | | 静脉（JSON 字符串数组） |
| side | varchar(10) | | L / R |
| hospital | varchar(200) | | 手术医院 |
| surgeon | varchar(100) | | 手术医生 |
| surgery_date | timestamp | | 手术时间 |
| first_use_date | timestamp | | 首次使用时间 |
| access_number | int | DEFAULT 1 | 第几次血管通路 |
| intervention_count | int | DEFAULT 0 | 干预次数 |
| intervention_date | timestamp | | 干预日期 |
| catheter_method | varchar(50) | | 置管方法（导管） |
| catheter_depth | varchar(20) | | 导管深度 |
| v_puncture_position | text (JSON) | | V侧穿刺位置（JSON 数组） |
| a_puncture_position | text (JSON) | | A侧穿刺位置（JSON 数组） |
| notes | text | | 备注 |
| images | text (JSON) | | 图片 URLs（JSON 数组） |
| is_default | bool | DEFAULT false | 是否默认 |
| is_disabled | bool | DEFAULT false | 是否禁用 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**通路类型枚举:**

| 代码 | 含义 |
|------|------|
| `自体动静脉内瘘AVF` | 自体动静脉内瘘 |
| `移植物动静脉内瘘AVG` | 移植物动静脉内瘘 |
| `带隧道和涤纶套的透析导管TCC` | 带隧道导管 |
| `无隧道和涤纶套的透析导管NCC` | 无隧道导管 |

**自定义类型 `StringSlice`:** 用于将 Go `[]string` 序列化为 JSON 存储到 `text` 列。实现了 `driver.Valuer` 和 `sql.Scanner` 接口。

---

#### `vascular_access_interventions` — 血管通路干预记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| vascular_access_id | varchar(36) | NOT NULL, INDEX | 关联 vascular_accesses.id |
| patient_id | varchar(36) | NOT NULL, INDEX | 冗余患者 ID（便于查询） |
| access_type | varchar(50) | | 通路类型（冗余） |
| avg_blood_flow | int | DEFAULT 0 | 平均血流量 |
| usage_days | int | DEFAULT 0 | 使用天数 |
| surgery_type | varchar(50) | NOT NULL | 手术类型 |
| intervention_reason | text | NOT NULL | 干预原因 |
| doctor | varchar(50) | | 干预医生 |
| intervention_date | timestamp | NOT NULL | 干预时间 |
| description | text | | 干预描述 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

---

#### `outcome_records` — 治疗转归记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | 关联 patients.id |
| type | varchar(20) | NOT NULL | 转入 / 转出 |
| reason | varchar(255) | | 原因 |
| time | timestamp | | 转归时间 |
| remarks | text | | 备注 |
| registrar | varchar(50) | | 登记人 |
| registration_time | timestamp | | 登记时间 |
| is_door_rule | bool | DEFAULT false | 是否门规 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

---

### 3.3 治疗方案模块

#### `treatment_plans` — 透析治疗方案

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | 关联 patients.id |
| weekly_frequency | int | DEFAULT 3 | 每周透析次数 |
| biweekly_frequency | int | DEFAULT 0 | 双周透析次数 |
| duration | int | DEFAULT 4 | 透析时长（小时） |
| dry_weight | decimal(5,2) | | 干体重 (kg) |
| extra_weight | decimal(5,2) | | 额外体重 (kg) |
| status | varchar(20) | DEFAULT '启用' | 启用 / 禁用 |
| doctor_id | varchar(36) | | 开立医生 ID |
| start_date | timestamp | | 方案开始日期 |
| end_date | timestamp | | 方案结束日期 |
| notes | text | | 备注 |
| dialysis_mode | **jsonb** | | 透析模式（嵌套对象） |
| anticoagulant | **jsonb** | | 抗凝方案（嵌套对象） |
| parameters | **jsonb** | | 透析参数（嵌套对象） |
| materials | **jsonb** | | 材料清单（数组） |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**JSONB 嵌套结构 — `dialysis_mode`:**

```json
{
  "mode": "HD",           // HD / HDF / HP / HD+HP
  "bloodFlow": 250,       // 血流量 (ml/min)
  "substituteInputMode": "前稀释",  // 置换液输入方式
  "substituteFlow": 0,    // 置换液流速 (ml/min)
  "substituteVolume": 0,  // 置换液总量 (L)
  "bv": "",               // 抗凝剂标识
  "frequencyDesc": "3次/周", // 频率描述
  "autoConfirm": false,   // 自动确认
  "status": "启用",
  "notes": ""
}
```

**JSONB 嵌套结构 — `anticoagulant`:**

```json
{
  "initialDrug": "普通肝素",    // 首剂药物
  "initialDose": "2000IU",     // 首剂量
  "maintenanceDrug": "普通肝素", // 维持量药物
  "infusionRate": "500IU/h",   // 输注速度
  "infusionTime": "3.5h",      // 输注时间
  "maintenanceDose": "1750IU", // 维持量
  "totalDose": "3750IU"        // 总量
}
```

**JSONB 嵌套结构 — `parameters` (DialysisParameters):**

```json
{
  "dialysateType": "碳酸氢盐",  // 透析液类型
  "dialysateGroup": "A组",     // 透析液组号
  "flowRate": 500,             // 透析液流量 (ml/min)
  "na": 138,                   // 钠 (mmol/L)
  "ca": 1.5,                   // 钙 (mmol/L)
  "k": 2.0,                    // 钾 (mmol/L)
  "hco3": 32,                  // 碳酸氢根 (mmol/L)
  "glucose": "含糖",           // 葡萄糖
  "conductivity": 14.0,        // 电导度 (mS/cm)
  "temp": 36.5,                // 温度 (°C)
  "volume": 120                // 透析液量 (L)
}
```

**JSONB 嵌套结构 — `materials` (MaterialList):**

```json
[
  {
    "id": "uuid",
    "name": "FX80 透析器",
    "category": "透析器",
    "count": 1,
    "code": "DLZ-001",
    "brand": "Fresenius",
    "spec": "1.8m²",
    "note": ""
  }
]
```

---

#### `orders` — 医嘱

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | 关联 patients.id |
| type | varchar(20) | NOT NULL | 长期 / 临时 |
| category | varchar(50) | | 药品 / 检查 / 治疗 / 护理 / 饮食 |
| name | varchar(100) | | 医嘱名称 |
| content | text | NOT NULL | 医嘱内容 |
| dose | varchar(50) | | 剂量 |
| unit | varchar(20) | | 单位 |
| route | varchar(50) | | 用法（静脉注射、口服等） |
| timing | varchar(50) | | 使用时机 |
| exec_timing | varchar(50) | | 执行时机 |
| drug_id | uint | INDEX | 关联 drug_catalogs.id（可选） |
| spec | varchar(100) | | 规格 |
| group_id | varchar(36) | | 医嘱组号（同组医嘱联合执行） |
| doctor_id | varchar(36) | NOT NULL | 开单医生 ID |
| doctor_name | varchar(50) | | 开单医生姓名 |
| status | varchar(20) | DEFAULT '待执行' | 待执行 / 执行中 / 已执行 / 已停止 |
| start_time | timestamp | | 医嘱开始时间 |
| end_time | timestamp | | 医嘱结束时间（临时医嘱必填） |
| frequency | varchar(50) | | 频次（qd/bid/tid/qod 等） |
| priority | varchar(20) | DEFAULT '普通' | 普通 / 紧急 / 临急 |
| notes | text | | 备注 |
| executed_at | timestamp | | 执行时间 |
| executed_by | varchar(36) | | 执行人 ID |
| stop_reason | text | | 停止原因 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**状态机:**

```
待执行 ──► 执行中 ──► 已执行
  │                      ▲
  │                      │ (临时医嘱过期，定时任务自动标记)
  └──► 已停止 ◄──── 执行中
```

**业务规则:**
- 医生身份 (`doctor_id`, `doctor_name`) 由后端 Auth 中间件自动注入，不接受前端传入
- 临时医嘱过期：定时任务每 5 分钟扫描 `type=临时 AND end_time < NOW() AND status IN (待执行, 执行中)`，条件更新为 `已执行`（幂等）
- 停用操作幂等：已停止的再次 Stop 不报错
- 查询默认排除过期临时医嘱，`includeExpired=true` 可包含

---

#### `prescriptions` — 每日透析处方（医嘱单）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | 关联 patients.id |
| treatment_plan_id | varchar(36) | NOT NULL | 关联 treatment_plans.id |
| prescription_date | timestamp | NOT NULL | 处方日期 |
| doctor_id | varchar(36) | NOT NULL | 开方医生 ID |
| doctor_name | varchar(50) | | 开方医生姓名 |
| status | varchar(20) | DEFAULT '待执行' | 待执行 / 执行中 / 已执行 / 已取消 |
| duration | int | | 透析时长（小时） |
| dry_weight | decimal(5,2) | | 干体重 |
| extra_weight | decimal(5,2) | | 额外体重 |
| dialysis_mode | **jsonb** | | 透析模式（同 TreatmentPlan） |
| anticoagulant | **jsonb** | | 抗凝方案 |
| parameters | **jsonb** | | 透析参数 |
| materials | **jsonb** | | 材料清单 |
| order_items | **jsonb** | | 药品明细快照 |
| notes | text | | 备注 |
| executed_at | timestamp | | 执行时间 |
| executed_by | varchar(36) | | 执行人 ID |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**JSONB 嵌套结构 — `order_items` (PrescriptionOrderItemList):**

```json
[
  {
    "orderId": "uuid",      // 来源医嘱 ID
    "name": "促红素",        // 药品名称
    "category": "促红素",    // 分类（用于分组显示）
    "dose": "10000IU",      // 剂量
    "unit": "IU",           // 单位
    "frequency": "每周一次", // 频次
    "route": "皮下注射",    // 用法
    "spec": "10000IU/支"    // 规格
  }
]
```

**业务规则:**
- `treatment_plan_id` 由后端自动填充：查询患者 `status=启用` 的 TreatmentPlan，按 `updated_at DESC` 取第一条；无启用方案返回 400
- **状态机约束:** Update 仅允许 `待执行` 状态；Execute/Cancel 幂等
- `order_items` 更新采用**全量替换**语义
- "提取长嘱"操作：从患者在用长期医嘱生成 `order_items` 快照，并复制启用方案的透析参数/抗凝/材料

---

#### `adjustment_records` — 方案调整记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | 关联 patients.id |
| content | text | NOT NULL | 调整内容描述 |
| operator | varchar(50) | | 调整人 |
| created_at | timestamp | | |

---

### 3.4 住院模块

#### `hospitalizations` — 住院信息

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | INDEX | 租户 ID |
| patient_id | bigint | NOT NULL, INDEX | 关联患者 |
| case_no | varchar(64) | | 病案号 |
| hosp_no | varchar(64) | | 住院号 |
| bar_code | varchar(64) | | 条码 |
| hosp_patient_type | varchar(64) | | 门诊 / 住院 / 急诊 |
| hosp_receive_dept | varchar(64) | | 接收科室 |
| hosp_ward | varchar(64) | | 病房 |
| hosp_bed | varchar(64) | | 床位 |
| attend_dr | varchar(64) | | 主治医生 |
| reception_dr | varchar(64) | | 接诊医生 |
| status | int | DEFAULT 1 | 1=在院, 0=出院 |
| admission_date | timestamp | | 入院日期 |
| discharge_date | timestamp | | 出院日期 |
| notes | text | | 备注 |
| creator_id | bigint | | 创建人 |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

---

### 3.5 排班模块

#### `wards` — 病区/病房

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | |
| tenant_id | bigint | INDEX | 租户 ID |
| name | varchar(128) | NOT NULL | 病房名称 |
| ward_type | varchar(64) | | HD / HDF / Isolation / VIP |
| department | varchar(128) | | 科室 |
| floor | int | | 楼层 |
| is_disabled | bool | DEFAULT false | |
| sort | int | | 排序 |
| notes | text | | |
| creator_id | bigint | | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `beds` — 床位

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | |
| tenant_id | bigint | INDEX | 租户 ID |
| ward_id | bigint | INDEX | 关联 wards.id |
| name | varchar(64) | NOT NULL | 床位号 |
| bed_type | varchar(64) | | Regular / ICU / VIP / Isolation |
| status | varchar(20) | DEFAULT 'available' | available / occupied / reserved / maintenance |
| is_disabled | bool | DEFAULT false | |
| sort | int | | |
| notes | text | | |
| creator_id | bigint | | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `shifts` — 班次定义

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | |
| tenant_id | bigint | INDEX | 租户 ID |
| name | varchar(64) | NOT NULL | 班次名称 |
| start_time | varchar(10) | NOT NULL | 开始时间 HH:MM |
| end_time | varchar(10) | NOT NULL | 结束时间 HH:MM |
| type | varchar(64) | | Morning / Afternoon / Night / Overtime |
| is_disabled | bool | DEFAULT false | |
| sort | int | | |
| notes | text | | |
| creator_id | bigint | | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `patient_shifts` — 患者排班

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | |
| tenant_id | bigint | INDEX | 租户 ID |
| patient_id | bigint | NOT NULL, UNIQUE(patient_id, schedule_date) | 关联患者 |
| schedule_date | timestamp | NOT NULL, UNIQUE(patient_id, schedule_date) | 排班日期 |
| shift_id | bigint | NOT NULL, INDEX | 关联 shifts.id |
| bed_id | bigint | INDEX | 关联 beds.id |
| ward_id | bigint | INDEX | 关联 wards.id |
| status | int | DEFAULT 0 | 0=待执行 / 1=已确认 / 2=进行中 / 3=已完成 / 4=已取消 |
| is_disabled | bool | DEFAULT false | |
| notes | text | | |
| creator_id | bigint | | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

**唯一约束:** `(patient_id, schedule_date)` — 同一患者同一天只能排一次班

---

### 3.6 透析治疗执行模块

> 注意: 该模块表名使用 `Treatment_` 前缀（历史原因），与其他模块的 snake_case 命名不同。

#### `Treatment_Treatment` — 透析治疗主表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | NOT NULL, INDEX | 租户 ID |
| patient_id | bigint | NOT NULL, INDEX(patient_id, treatment_date) | |
| treatment_date | timestamp | NOT NULL, INDEX(patient_id, treatment_date) | 治疗日期 |
| schedule_id | bigint | INDEX | 关联 patient_shifts.id |
| reception_dr_id | bigint | | 接诊医生 |
| sign_in_time | timestamp | | 签到时间 |
| queue_no | varchar(32) | | 排队号 |
| reception_time | timestamp | | 接诊时间 |
| day_programme_id | bigint | | 日间治疗方案 ID |
| ward_id | bigint | | |
| ward_name | varchar(256) | | 病区名称 |
| bed_id | bigint | INDEX | |
| shift_id | bigint | INDEX | |
| shift_timing | int | | 班次时段 |
| type | int | NOT NULL | 1=HD / 2=HDF / 3=HP / 4=HD+HP |
| status | int | NOT NULL, DEFAULT 0 | 0=待开始 / 1=进行中 / 2=已完成 / 3=已取消 |
| is_disabled | bool | DEFAULT false | |
| creator_id | bigint | NOT NULL | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `Treatment_BeforeCheck` — 透前检查（1:1）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | NOT NULL, INDEX | |
| treatment_id | bigint | UNIQUE, NOT NULL | 关联 Treatment_Treatment.id |
| weight | float64 | | 体重 (kg) |
| temperature | float64 | | 体温 (°C) |
| sbp | int | | 收缩压 (mmHg) |
| dbp | int | | 舒张压 (mmHg) |
| heart_rate | int | | 心率 (bpm) |
| edema | int | | 水肿程度 |
| consciousness | varchar(32) | | 意识状态 |
| complication | varchar(512) | | 并发症 |
| dry_weight | float64 | | 干体重 |
| pre_weight | float64 | | 预估体重 |
| vascular_access | varchar(256) | | 血管通路 |
| cannula_type | varchar(64) | | 穿刺类型 |
| cannula_position | varchar(256) | | 穿刺部位 |
| catheter | varchar(512) | | 导管情况 |
| heparing_lock | varchar(512) | | 肝素封管 |
| machine_no | varchar(64) | | 机器号 |
| dialyzer | varchar(256) | | 透析器 |
| dialysate | varchar(256) | | 透析液 |
| calcium | float64 | | 钙浓度 |
| sodium | float64 | | 钠浓度 |
| bicarbonate | float64 | | 碳酸氢根 |
| notes | varchar(1024) | | 备注 |
| creator_id | bigint | NOT NULL | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `Treatment_BeforeSigns` — 透前体征（1:1）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | INDEX | |
| treatment_id | bigint | UNIQUE, NOT NULL | |
| sbp | int | | 收缩压 |
| dbp | int | | 舒张压 |
| heart_rate | int | | 心率 |
| sp_o2 | int | | 血氧饱和度 (%) |
| respiration | int | | 呼吸频率 |
| temperature | float64 | | 体温 |
| weight | float64 | | 体重 |
| symptoms | varchar(1024) | | 症状描述 |
| notes | varchar(1024) | | 备注 |
| creator_id | bigint | NOT NULL | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `Treatment_DuringParam` — 透析中参数（1:N，时序记录）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | NOT NULL, INDEX | |
| treatment_id | bigint | NOT NULL, INDEX(treatment_id, record_time) | |
| record_time | timestamp | NOT NULL, INDEX(treatment_id, record_time) | 记录时间 |
| code | varchar(32) | NOT NULL | 参数代码 |
| blood_flow | float64 | | 血流量 (ml/min) |
| dialysate_flow | float64 | | 透析液流量 (ml/min) |
| uf_volume | float64 | | 超滤量 (ml) |
| venous_pressure | float64 | | 静脉压 (mmHg) |
| arterial_pressure | float64 | | 动脉压 (mmHg) |
| tmp | float64 | | 跨膜压 (mmHg) |
| temperature | float64 | | 温度 (°C) |
| conductivity | float64 | | 电导度 (mS/cm) |
| uf_rate | float64 | | 超滤率 (ml/h) |
| notes | varchar(512) | | 备注 |
| creator_id | bigint | NOT NULL | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `Treatment_AfterSigns` — 透后体征（1:1）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | INDEX | |
| treatment_id | bigint | UNIQUE, NOT NULL | |
| sbp | int | | 收缩压 |
| dbp | int | | 舒张压 |
| heart_rate | int | | 心率 |
| sp_o2 | int | | 血氧饱和度 |
| weight | float64 | | 体重 |
| uf_volume | float64 | | 实际超滤量 (ml) |
| dialysis_time | int | | 透析时长 (分钟) |
| complication | varchar(1024) | | 并发症 |
| symptoms | varchar(1024) | | 症状描述 |
| notes | varchar(1024) | | 备注 |
| creator_id | bigint | NOT NULL | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

#### `Treatment_Alarm` — 报警记录（1:N）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK auto | |
| tenant_id | bigint | NOT NULL, INDEX | |
| treatment_id | bigint | NOT NULL, INDEX | |
| alarm_time | timestamp | NOT NULL | 报警时间 |
| alarm_code | varchar(64) | NOT NULL | 报警代码 |
| alarm_level | int | NOT NULL | 1=信息 / 2=警告 / 3=错误 / 4=严重 |
| alarm_message | varchar(512) | NOT NULL | 报警信息 |
| is_handled | bool | DEFAULT false | 是否已处理 |
| handled_by | bigint | | 处理人 ID |
| handled_at | timestamp | | 处理时间 |
| handle_note | varchar(512) | | 处理说明 |
| creator_id | bigint | NOT NULL | |
| create_time | timestamp | | |
| last_modify_time | timestamp | | |

---

### 3.7 配置/主数据模块

#### `plan_templates` — 治疗方案模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| name | varchar(100) | NOT NULL | 模板名称 |
| description | text | | 描述 |
| mode | varchar(20) | NOT NULL | HD / HDF / HP / HF / HFD |
| is_default | bool | DEFAULT false | |
| is_enabled | bool | | |
| category | varchar(50) | | 分类: 急性 / 慢性 / 导管等 |
| tenant_id | bigint | INDEX | |
| template_content | **jsonb** | | 完整模板内容 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

`template_content` JSONB 包含: `weeklyFrequency`, `biweeklyFrequency`, `duration`, `dryWeight`, `dialysisMode`, `anticoagulant`, `parameters`, `materials`，结构与 TreatmentPlan 的 JSONB 字段一致。

---

#### `drug_catalogs` — 药品目录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK auto | |
| code | varchar(50) | UNIQUE, NOT NULL | 药品编码 |
| name | varchar(100) | NOT NULL | 药品名称 |
| short_name | varchar(50) | | 简称 |
| mnemonic | varchar(50) | | 助记符（拼音首字母） |
| generic_name | varchar(100) | | 通用名 |
| category | varchar(50) | NOT NULL, INDEX | 药品分类 |
| spec | varchar(100) | | 规格 |
| concentration | varchar(50) | | 浓度 |
| spec_unit | varchar(20) | | 规格单位 |
| min_unit_dose | varchar(20) | | 最小单位剂量 |
| unit | varchar(20) | | 基本单位 |
| brand | varchar(100) | | 品牌 |
| packaging | varchar(50) | | 包装 |
| manufacturer | varchar(100) | | 生产厂家 |
| standard_type | varchar(50) | | 标准类型 |
| timing | varchar(50) | | 使用时机 |
| tips | varchar(200) | | 提示信息 |
| sort_order | int | DEFAULT 0 | 排序 |
| is_enabled | bool | | 是否启用 |
| tenant_id | bigint | INDEX | |
| notes | text | | 备注 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**药品分类:** 抗凝剂 / 低分子肝素 / 柠檬酸钠 / 促红素 / 铁剂 / 钙剂 / 维生素D / 降压药 / 利尿剂 / 其他 / `METHOD`（方法类条目，非药品）

---

#### `material_catalogs` — 材料目录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | uint | PK auto | |
| code | varchar(50) | UNIQUE, NOT NULL | 材料编码 |
| name | varchar(100) | NOT NULL | 材料名称 |
| short_name | varchar(50) | | 简称 |
| mnemonic | varchar(50) | | 助记符 |
| category | varchar(50) | NOT NULL, INDEX | 材料分类 |
| spec | varchar(100) | | 规格 |
| standard_type | varchar(50) | | 标准类型 |
| brand | varchar(100) | | 品牌 |
| unit | varchar(20) | | 单位 |
| packaging | varchar(50) | | 包装 |
| manufacturer | varchar(100) | | 生产厂家 |
| sort_order | int | DEFAULT 0 | 排序 |
| is_enabled | bool | | |
| tenant_id | bigint | INDEX | |
| notes | text | | |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**材料分类:** 透析器 / 血路管 / 导管 / 穿刺针 / 注射器 / 输液器 / 消毒剂 / 敷料 / 手套 / 其他

---

#### `order_templates` — 医嘱模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| name | varchar(100) | NOT NULL | 模板名称 |
| type | varchar(20) | NOT NULL | 长期 / 临时 |
| category | varchar(50) | NOT NULL | 药品 / 检查 / 治疗 / 护理 / 饮食 |
| content | text | NOT NULL | 模板内容 |
| frequency | varchar(50) | | 默认频次 |
| priority | varchar(20) | DEFAULT '普通' | 优先级 |
| is_default | bool | DEFAULT false | |
| is_enabled | bool | | |
| tenant_id | bigint | INDEX | |
| created_at | timestamp | | |
| updated_at | timestamp | | |

#### `order_template_items` — 医嘱模板条目

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID (BeforeCreate 自动生成) |
| template_id | varchar(36) | NOT NULL, INDEX | 关联 order_templates.id |
| drug_id | uint | INDEX | 关联 drug_catalogs.id |
| drug_name | varchar(100) | NOT NULL | 药品名称 |
| spec | varchar(100) | | 规格 |
| min_unit_dose | varchar(20) | | 最小单位剂量 |
| dosage | varchar(50) | | 用量 |
| unit | varchar(20) | | 单位 |
| route | varchar(50) | | 用法 |
| frequency | varchar(50) | | 频次 |
| timing | varchar(50) | | 使用时机 |
| group_id | varchar(36) | | 组号 |
| sort_order | int | DEFAULT 0 | 排序 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

---

### 3.8 字典模块

#### `dict_types` — 字典类型

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID (BeforeCreate 自动生成) |
| code | varchar(50) | UNIQUE, NOT NULL | 字典类型编码 |
| name | varchar(100) | NOT NULL | 字典类型名称 |
| description | varchar(500) | | 描述 |
| icon | varchar(50) | | 图标 |
| sort_order | int | DEFAULT 0 | |
| is_enabled | bool | DEFAULT true | |
| created_at | timestamp | | |
| updated_at | timestamp | | |

#### `dict_items` — 字典项

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID (BeforeCreate 自动生成) |
| type_code | varchar(50) | NOT NULL, INDEX | 关联 dict_types.code |
| code | varchar(50) | NOT NULL | 字典项编码 |
| name | varchar(100) | NOT NULL | 字典项名称 |
| description | varchar(500) | | 描述 |
| sort_order | int | DEFAULT 0 | |
| is_enabled | bool | DEFAULT true | |
| extra | varchar(500) | | 扩展字段（如颜色标识） |
| parent_code | varchar(50) | | 父级代码（树形结构） |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**唯一索引:** `idx_dict_items_unique ON dict_items(type_code, code)`

**已使用的字典类型编码:**

| 编码 | 含义 |
|------|------|
| `DIALYSIS_MODE` | 透析方式 |
| `ANTICOAGULANT` | 抗凝剂类型 |
| `DIALYSATE_TYPE` | 透析液类型 |
| `DIALYSATE_GROUP` | 透析液组 |
| `DIALYSATE_FLOW` | 透析液流量 |
| `GLUCOSE` | 葡萄糖类型 |
| `MATERIAL_CATEGORY` | 材料分类 |
| `DRUG_CATEGORY` | 药品分类 |
| `ORDER_TYPE` | 医嘱类型 |
| `ORDER_CATEGORY` | 医嘱分类 |
| `ORDER_ROUTE` | 医嘱用法 |
| `ORDER_FREQUENCY` | 医嘱频次 |
| `ORDER_TIMING` | 使用时机 |
| `PATIENT_STATUS` | 患者状态 |
| `VASCULAR_ACCESS` | 血管通路类型 |
| `VASCULAR_SITE` | 血管通路部位 |
| `VEIN_TYPE` | 静脉类型 |
| `ARTERY_TYPE` | 动脉类型 |
| `INSURANCE_TYPE` | 医保类型 |
| `PATIENT_TYPE` | 患者类型 |
| `ID_TYPE` | 证件类型 |
| `VISIT_CATEGORY` | 就诊类别 |
| `DOCTOR` | 医生列表 |
| `HOSPITAL` | 手术医院 |
| `SURGERY_TYPE` | 手术类型 |
| `PRIMARY_DISEASE` | 原发病分类 |
| `COMPLICATION` | 并发症类型 |
| `PATHOLOGY` | 病理诊断分类 |
| `TUMOR` | 肿瘤分类 |
| `ALLERGEN` | 过敏原类型 |
| `OUTCOME` | 患者转归 |

---

### 3.9 检验检查模块

#### `lab_reports` — 检验报告主表

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | |
| report_no | varchar(64) | INDEX | 报告号 |
| item_code | varchar(64) | | 检验项目编码 |
| item_name | varchar(128) | | 检验项目名称 |
| clinical_diagnosis | text | | 临床诊断 |
| specimen_type | varchar(64) | | 标本类型 |
| urgency | varchar(32) | DEFAULT '常规' | 常规 / 急诊 |
| request_doctor | varchar(64) | | 申请医生 |
| requested_at | timestamp | | 申请时间 |
| sampled_at | timestamp | | 采样时间 |
| received_at | timestamp | | 接收时间 |
| reported_at | timestamp | | 报告时间 |
| status | varchar(32) | DEFAULT '已出报告' | |
| external_report_id | varchar(128) | INDEX | 外部系统报告 ID |
| source_system | varchar(16) | DEFAULT 'LOCAL' | LOCAL / LIS / PACS / HDIS_EXAM / HDIS_RECORD |
| synced_at | timestamp | | 同步时间 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

#### `lab_report_items` — 检验报告明细

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| lab_report_id | varchar(36) | NOT NULL, INDEX | 关联 lab_reports.id (CASCADE DELETE) |
| item_code | varchar(64) | NOT NULL, INDEX | 指标编码 |
| item_name | varchar(128) | NOT NULL | 指标名称 |
| result_value | varchar(64) | NOT NULL | 检验结果 |
| unit | varchar(32) | | 单位 |
| reference_range | varchar(128) | | 参考范围 |
| abnormal_flag | varchar(8) | DEFAULT 'N' | H=偏高 / L=偏低 / N=正常 |
| tested_at | timestamp | | 检测时间 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

#### `exam_reports` — 检查报告

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | |
| exam_date | timestamp | INDEX | 检查日期 |
| title | varchar(200) | NOT NULL | 检查标题 |
| conclusion | text | | 结论 |
| department | varchar(100) | | 检查科室 |
| external_report_id | varchar(128) | INDEX | |
| source_system | varchar(16) | DEFAULT 'LOCAL' | |
| synced_at | timestamp | | |
| created_at | timestamp | | |
| updated_at | timestamp | | |

#### `patient_key_indicators` — 患者关键指标

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| patient_id | varchar(36) | NOT NULL, INDEX | |
| external_record_id | varchar(128) | NOT NULL, UNIQUE(external_record_id, source_system) | 外部记录 ID |
| source_system | varchar(32) | NOT NULL, UNIQUE(external_record_id, source_system) | 来源系统 |
| index_name | varchar(200) | NOT NULL | 指标名称 |
| index_code | varchar(64) | INDEX | 指标编码 |
| result | varchar(128) | | 结果值 |
| unit | varchar(64) | | 单位 |
| reference | varchar(200) | | 参考范围 |
| result_sign | varchar(8) | | 结果标记 |
| test_time | timestamp | INDEX | 检测时间 |
| evaluation_result | varchar(64) | | 评估结果 |
| synced_at | timestamp | | |
| created_at | timestamp | | |
| updated_at | timestamp | | |

**唯一索引:** `idx_patient_key_indicators_unique ON (external_record_id, source_system)`

---

### 3.10 集成配置模块

#### `integration_hdis_settings` — HDIS 对接配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | varchar(36) | PK | UUID |
| webcmd_url | varchar(255) | NOT NULL | HDIS WebCmd URL |
| graphql_url | varchar(255) | NOT NULL | HDIS GraphQL URL |
| auth_url | varchar(255) | NOT NULL | 认证 URL |
| client_id | varchar(64) | NOT NULL | 客户端 ID |
| service_username | varchar(64) | NOT NULL | 服务账号 |
| service_password_encrypted | text | NOT NULL | 加密密码（JSON 隐藏） |
| access_token_encrypted | text | | 加密 Token（JSON 隐藏） |
| token_expires_at | timestamp | | Token 过期时间 |
| auto_refresh_enabled | bool | NOT NULL, DEFAULT true | 自动刷新 |
| refresh_lead_seconds | int | NOT NULL, DEFAULT 1800 | 提前刷新秒数 |
| last_error | text | | 最后错误信息 |
| updated_by | varchar(36) | | 最后更新人 |
| created_at | timestamp | | |
| updated_at | timestamp | | |

---

## 4. JSONB 自定义类型汇总

| Go 类型 | 数据库列类型 | 序列化方式 | 使用表 |
|---------|-------------|-----------|--------|
| `DialysisMode` | jsonb | GORM serializer:json | treatment_plans, prescriptions |
| `Anticoagulant` | jsonb | GORM serializer:json | treatment_plans, prescriptions |
| `DialysisParameters` | jsonb | GORM serializer:json | treatment_plans, prescriptions |
| `MaterialList` | jsonb | Scan/Value 接口 | treatment_plans, prescriptions |
| `PrescriptionOrderItemList` | jsonb | Scan/Value 接口 | prescriptions |
| `PlanTemplateContent` | jsonb | GORM serializer:json | plan_templates |
| `StringSlice` | text (JSON) | Scan/Value 接口 | vascular_accesses |

---

## 5. 索引策略

| 表 | 索引 | 类型 | 用途 |
|----|------|------|------|
| users | username | UNIQUE | 登录查找 |
| patient_basic_infos | patient_id | UNIQUE | 1:1 关联 |
| patient_basic_infos | hdis_patient_id | UNIQUE | 外部系统 ID |
| medical_histories | patient_id | UNIQUE | 1:1 关联 |
| infection_infos | patient_id | UNIQUE | 1:1 关联 |
| vascular_accesses | patient_id | INDEX | 按患者查询 |
| vascular_access_interventions | vascular_access_id | INDEX | 按通路查询 |
| vascular_access_interventions | patient_id | INDEX | 按患者查询 |
| orders | patient_id | INDEX | 按患者查询 |
| orders | drug_id | INDEX | 药品关联 |
| prescriptions | patient_id | INDEX | 按患者查询 |
| treatment_plans | patient_id | INDEX | 按患者查询 |
| patient_shifts | (patient_id, schedule_date) | UNIQUE | 防重排 |
| Treatment_Treatment | (patient_id, treatment_date) | INDEX | 按日期查询 |
| Treatment_DuringParam | (treatment_id, record_time) | INDEX | 时序查询 |
| Treatment_BeforeCheck | treatment_id | UNIQUE | 1:1 关联 |
| Treatment_BeforeSigns | treatment_id | UNIQUE | 1:1 关联 |
| Treatment_AfterSigns | treatment_id | UNIQUE | 1:1 关联 |
| dict_items | (type_code, code) | UNIQUE | 防重复 |
| drug_catalogs | code | UNIQUE | 药品唯一编码 |
| material_catalogs | code | UNIQUE | 材料唯一编码 |
| patient_key_indicators | (external_record_id, source_system) | UNIQUE | 去重 |
| lab_reports | external_report_id | INDEX | 外部同步查找 |

---

## 6. 设计约定

### 6.1 命名规范

| 维度 | 规范 | 示例 |
|------|------|------|
| 表名 | snake_case（透析执行模块除外） | `treatment_plans` |
| 列名 | snake_case | `patient_id` |
| Go 字段 | PascalCase | `PatientID` |
| JSON 序列化 | camelCase | `patientId` |
| GORM Tag | 显式指定类型与约束 | `gorm:"type:varchar(36);primaryKey"` |

### 6.2 时间字段

| 字段名 | 用途 | 域 |
|--------|------|-----|
| `created_at` / `updated_at` | GORM 自动维护 | 患者域、治疗域、字典域 |
| `create_time` / `last_modify_time` | GORM 自动维护 | 排班域、透析执行域 |

两种命名并存是历史原因（不同阶段开发），新表统一使用 `created_at` / `updated_at`。

### 6.3 软删除

当前系统**不使用 GORM 软删除**（无 `deleted_at` 字段）。医嘱不提供物理删除，统一为"停用"状态变更。

### 6.4 租户隔离

部分表（排班域、透析执行域、住院、配置模板、目录）包含 `tenant_id` 字段，为多租户预留。当前系统为单租户运行，`tenant_id` 的过滤逻辑未强制启用。

### 6.5 外键策略

- GORM 层定义 `foreignKey` 关联（用于 Preload/Join）
- 数据库层**不创建**物理外键约束 (`DisableForeignKeyConstraintWhenMigrating: true`)
- 唯一例外: `lab_report_items` 的 `lab_report_id` 设置了 `OnDelete:CASCADE`

---

## 7. 数据流关系图

### 7.1 医嘱 → 处方 数据流

```
drug_catalogs ──引用──► order_template_items
                              │
                        "从模板创建"
                              │
                              ▼
                         orders (长期/临时医嘱)
                              │
                        "提取长嘱"
                              │
                              ▼
treatment_plans ──复制──► prescriptions
   (透析参数)           (每日处方, orderItems 为医嘱快照)
```

### 7.2 外部系统数据同步

```
HDIS/LIS ───同步──► lab_reports + lab_report_items
                     (source_system = 'LIS')

HDIS ───同步──► exam_reports
                (source_system = 'HDIS_EXAM')

HDIS Record ──同步──► patient_key_indicators
                      (source_system = 'HDIS_RECORD')
```

---

## 8. 连接池配置

```go
sqlDB.SetMaxIdleConns(10)      // 最大空闲连接
sqlDB.SetMaxOpenConns(100)     // 最大打开连接
sqlDB.SetConnMaxLifetime(1h)   // 连接最大存活时间
```

---

## 9. 定时任务

| 任务 | 频率 | SQL | 说明 |
|------|------|-----|------|
| 过期临时医嘱标记 | 5 分钟 | `UPDATE orders SET status='已执行' WHERE type='临时' AND end_time < NOW() AND status IN ('待执行','执行中')` | 幂等，单实例前提 |

---

*文档生成自 `ai-hms-backend/internal/models/` 目录下的 GORM 模型定义。*
