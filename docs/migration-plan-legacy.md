# 新血透 → 老血透 · 后端代码迁移计划

> **本计划的读者是 Codex**（AI 编程助手）。每个任务都设计成独立、可验证、可按顺序推进。
>
> - **不要**改动数据库。所有改造仅限 Go 代码侧（`ai-hms-backend/`）。
> - **不要**恢复 `AutoMigrate` 或任何 DDL 代码路径。
> - 每完成一个任务，先跑 `./scripts/verify.sh` 再进入下一个。
> - 每个任务的 `Deliverable` 段列出产出文件；**不要**产出计划之外的文件。
>
> **配套资料**
> - 字段级映射表：`docs/migration-field-map.md`（Codex 查字段必读）
> - 老库字段权威：`老血透数据库表结构-合并版.md`（根目录）
> - 当前已有部分映射笔记：`ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`
> - 新表模型源头：`DATABASE_DESIGN.md`（根目录）
> - Legacy DB 红线：`ai-hms-backend/CLAUDE.md`

---

## 0. 背景与目标

### 0.1 项目形态

- **后端**（`ai-hms-backend`）用 Go + Gin + GORM，按**新血透设计**（`DATABASE_DESIGN.md`，35 张 snake_case 表，多处 JSONB + UUID）建立了 GORM 模型。
- **生产数据库**是**老血透 PostgreSQL**（102 张 PascalCase 表，`Register_*` / `Plan_*` / `Treatment_*` / `Schedule_*` 等前缀，`bigint` 主键）。
- 当前已完成**部分模型切换**到老库（见 `LEGACY_TABLE_FIELD_MAPPING.md`）：`Patient`、`VascularAccess`、`VascularAccessIntervention`、`OutcomeRecord`、`InfectionInfo`、`Treatment*` 系列。**其余仍指向新表名**，跑起来会因目标表不存在而查询失败。

### 0.2 本计划的目标

把所有**新表模型**一步步改造成**能在老血透库上运行**。每一步产出都：

1. 通过 `go build ./cmd/server`、`go vet ./...`、`go test ./...`
2. 不删除数据库中任何东西（只改 Go 代码）
3. 保留新血透前端的 API 契约（handler 输入输出不变）

### 0.3 不在本计划范围

- 数据库 DDL、表迁移、数据导入
- 前端改动
- 新 feature
- HDIS/LIS 外部同步逻辑（`patient_key_indicators`、`exam_reports` 走 HDIS 同步，保持现状）

---

## 1. 术语与文件定位

| 术语 | 含义 |
|------|------|
| **新表** | 在 `DATABASE_DESIGN.md` 定义的 35 张 snake_case 表 |
| **老表 / Legacy** | 生产库（老血透）里 102 张 PascalCase 表 |
| **应用层表** | 老库没有、新系统自管的表（如 `integration_hdis_settings`、`permissions`）；**不迁移**，继续用原表名 |
| **映射类别** | 见 `docs/migration-field-map.md` 开头；`rename` / `rewrite` / `fold` / `split-to-many` / `app-only` 等 |

**Codex 每个任务里最高频改动的文件夹**：

```
ai-hms-backend/internal/
├── models/          ← TableName、column tag、字段类型（最集中）
├── services/        ← SQL / GORM 查询、DTO 转换
└── api/v1/          ← handler（一般只动 Request/Response 映射）
```

**绝对不改**：
- `internal/database/migrate.go`（AutoMigrate 守卫）
- `internal/database/database.go` 的连接逻辑
- `config/` 配置加载

---

## 2. 红线与约束（每个任务都适用）

1. **只读老库 schema**，绝不 `CREATE/ALTER/DROP`。
2. **GORM 建模时写死 column tag**：老库列名是 PascalCase（`Register_PatientInfomation.PatientType`），Go 侧用 snake_case 属性的必须加 `gorm:"column:PatientType"`。
3. **主键**：老库大多用 `bigint`；新代码用 UUID (`varchar(36)`) 的地方，**切换到 bigint + 雪花/`nextLegacyNumericID`**；保留 `modeltypes.LegacyID` 抽象。
4. **时间字段**：老库统一 `CreateTime` / `LastModifyTime`（非 `created_at/updated_at`）。GORM 靠 tag 补齐：

   ```go
   CreateTime     time.Time `gorm:"column:CreateTime;autoCreateTime"`
   LastModifyTime time.Time `gorm:"column:LastModifyTime;autoUpdateTime"`
   ```

5. **JSON 字段处理**：老库凡是有"Dialysis方案"等 JSON 化存储的，用 `Plan_JsonData`/`Auxiliary_JsonData` 衔接（参见对照表）；无 JSONB 时字段拆散。
6. **软删除**：新代码几乎不用 `deleted_at`；老库统一用 `IsDisabled boolean`。service 的 Delete 实际执行 `UPDATE ... SET IsDisabled=true`。
7. **枚举值**：老库枚举为中文/代号混合（如 `PatientType='10'` vs 新代码 `patient_type='门诊'`）——每个 rewrite 任务需要一张**枚举映射表**在 service 层做双向转换。
8. **TenantId**：老库所有表含 `TenantId bigint`，每条读写 SQL 都要带 `WHERE TenantId = ?`。从 Auth 中间件取。

---

## 3. 策略总纲

### 3.1 按类别划分

| 类别 | 表数 | 策略 | 本计划 Phase |
|------|------|------|---------------|
| `rename` | 10 | 仅改 TableName 和字段 tag | Phase 1 |
| `rename-fields` | 3 | 上 + 逐字段 column tag | Phase 2 |
| `same-name` | 6 | 核对字段差异、补 tag | Phase 2 |
| `rewrite` | 4 | 重写 service 读写逻辑 | Phase 3 |
| `rewrite+child` | 2 | Phase 3 + 子表 CRUD | Phase 3 |
| `fold` | 1 | service 层 join 多张老表 | Phase 4 |
| `split-to-many` | 1 | service 分散写入多张老表 | Phase 4 |
| `multi-join` | 1 | 已实现，仅修 TableName | Phase 5 |
| `app-only` | 10 | 不改 | Phase 6 |

### 3.2 Phase 推进原则

- **先简单后复杂**：先把 `rename` 类跑通，积累 Tag 规范，再攻 `rewrite`。
- **每个 Phase 独立可合并 PR**：不相互依赖代码路径上的更改。
- **Phase 内并行**：一个 Phase 里的任务之间基本独立（同一 Phase 改到同一文件的任务合并）。

---

## 4. Phase 分解

### Phase 0 — 前置准备（半小时）

#### Task 0.1 创建工作分支

**What**: 创建 `feat/legacy-schema-alignment` 分支。

**Why**: 整个迁移跨多 PR，需要在独立分支串起来。

**How**:
```bash
git checkout -b feat/legacy-schema-alignment
git push -u origin feat/legacy-schema-alignment
```

**Verify**: `git branch --show-current` 输出 `feat/legacy-schema-alignment`。

**Deliverable**: 无文件产出。

---

#### Task 0.2 读通关键参考

**What**: 顺序通读下列文件，形成全局心智模型：

1. `docs/migration-plan-legacy.md`（本文件）
2. `docs/migration-field-map.md`
3. `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`
4. `ai-hms-backend/CLAUDE.md` 的 Legacy DB 守则
5. `ai-hms-backend/internal/services/legacy_id_helper.go`（`nextLegacyNumericID` 实现）
6. `ai-hms-backend/internal/services/schema_helper.go`（注意现在是 no-op）
7. `ai-hms-backend/internal/models/types/id.go`（`LegacyID` 类型）

**Verify**: 能口答这些问题：
- 主键类型从 UUID→bigint 时，GORM tag 怎么写？
- Create 时新 id 怎么来？
- 枚举字段在哪里做中→老映射？
- 字典缺表 fallback 的触发时机？

**Deliverable**: 无。

---

#### Task 0.3 建立枚举映射集中位置

**What**: 在 `internal/services/legacy_enum_maps.go` 建立一个**空白**集中枚举转换模块。各 Phase 会往里填。

**骨架**：

```go
package services

// legacy_enum_maps.go
// 集中收纳"新血透枚举 → 老血透字段值"的双向映射，避免散落在多 service。
// 每个枚举：NewToLegacyXxx / LegacyToNewXxx 一对；未命中返回原值。

// PatientType: 新（门诊/住院）→ 老（10/20）
var patientTypeNewToLegacy = map[string]string{
    "门诊": "10",
    "住院": "20",
}
var patientTypeLegacyToNew = map[string]string{
    "10": "门诊",
    "20": "住院",
}

func MapPatientTypeNewToLegacy(v string) string {
    if m, ok := patientTypeNewToLegacy[v]; ok { return m }
    return v
}
func MapPatientTypeLegacyToNew(v string) string {
    if m, ok := patientTypeLegacyToNew[v]; ok { return m }
    return v
}

// TODO: 后续 Phase 补充 DialysisMode / OrderStatus / OrderType / ...
```

**Verify**: `go build ./...` 通过。

**Deliverable**: `ai-hms-backend/internal/services/legacy_enum_maps.go`。

---

### Phase 1 — `rename` 类改造（2-3 小时）

#### Task 1.1 `hospitalizations` → `Register_Hospitalization`

**对照**: `docs/migration-field-map.md` 的 `hospitalizations` 段。

**目标文件**:
- `ai-hms-backend/internal/models/hospitalization.go`
- `ai-hms-backend/internal/services/hospitalization_service.go`（如存在）
- 相关 handler 不动

**改造要点**:
1. `TableName()` 改为 `"Register_Hospitalization"`
2. 主键 `Id` 改为 bigint（已是 bigint 则不动），补 `gorm:"column:Id;primaryKey"`
3. 字段 column tag 逐项按对照表补：
   - `case_no` → `CaseNo`
   - `hosp_no` → `HospNo`
   - `bar_code` → `BarCode`
   - `hosp_patient_type` → `HospPatientType`
   - `hosp_receive_dept` → `HospReceiveDept`
   - `hosp_ward` → `HospWard`
   - `hosp_bed` → `HospBed`
   - `attend_dr` → `AttendDr`
   - `reception_dr` → `ReceptionDr`
   - `medical_record_no` → `MedicalRecordNo`
   - `tenant_id` → `TenantId`
   - `patient_id` → `PatientId`
   - `creator_id` → `CreatorId`
   - `create_time` → `CreateTime` + `autoCreateTime`
   - `last_modify_time` → `LastModifyTime` + `autoUpdateTime`
4. 新表独有字段（`status/admission_date/discharge_date/notes`）**老库没有**：
   - `status` → 老库无，用 Hospitalization 是否存在开放记录（`discharge_date IS NULL`）判断
   - `admission_date`/`discharge_date` → 用 `CreateTime` + service 逻辑维护
   - `notes` → 老库无直接字段，需要决策：`NULL` 返回或用 `Register_Diagnosis.Note`
5. Service 查询：所有 `SELECT` 带 `TenantId`；软删除用 `IsDisabled = false`（老表可能没有 IsDisabled，核对 `老血透数据库表结构-合并版.md` 里 `Register_Hospitalization`）

**Verify**:
```bash
cd ai-hms-backend
./scripts/verify.sh
go test ./internal/services -run Hospitalization
```

**Deliverable**: 以上 2 个文件的 PR 级变更。

---

#### Task 1.2 `wards/beds/shifts/patient_shifts` → `Schedule_*`

**目标文件**:
- `internal/models/schedule.go`
- `internal/services/shift_service.go`
- `internal/services/patient_shift_service.go`

**改造要点**:

| 模型 | 新 TableName | 老 TableName | 字段重点 |
|------|-------------|-------------|----------|
| `Ward` | `wards` | `Schedule_Ward` | 全字段 PascalCase，`ward_type` → `WardType` |
| `Bed` | `beds` | `Schedule_Bed` | `bed_type` → `BedType`，`ward_id` → `WardId` |
| `Shift` | `shifts` | `Schedule_Shift` | `start_time`/`end_time` 在老库可能是 `StartTime`/`EndTime` varchar；核对合并 md |
| `PatientShift` | `patient_shifts` | `Schedule_PatientShift` | `schedule_date` → `TreatmentTime` |

**状态枚举转换**（写到 `legacy_enum_maps.go`）：
- `patient_shifts.status` 新值 `0/1/2/3/4` 对应 `待执行/已确认/进行中/已完成/已取消`
- 老库 `Schedule_PatientShift.Status` 取值需核对（查 `老血透数据库表结构-合并版.md` 对应字段）

**Verify**:
```bash
./scripts/verify.sh
go test ./internal/services -run "Shift|Ward|Bed"
```

**Deliverable**: `internal/models/schedule.go` + 2 个 service。

---

#### Task 1.3 `adjustment_records` → `Plan_PatientPlanPrescriptionAdjustment`

**目标文件**:
- `internal/models/treatment.go`（`AdjustmentRecord` 部分）
- `internal/services/patient_service.go` 的 `GetAdjustmentRecords/CreateAdjustmentRecord`

**改造要点**:
1. `TableName()` 改为 `"Plan_PatientPlanPrescriptionAdjustment"`
2. 字段对齐见对照表
3. 注意 `content` 在老库可能是 `AdjustmentContent` 或 `Note`，务必核对

**Verify**: `./scripts/verify.sh` + 手工调用 `GET /patients/:id/adjustment-records`。

**Deliverable**: model + service 的最小改动。

---

#### Task 1.4 `material_catalogs` → `Auxiliary_MaterialInfomation`

**目标文件**:
- `internal/models/treatment_config.go`（`MaterialCatalog` 部分）
- `internal/services/treatment_config_service.go`

**改造要点**:
1. `TableName()` 改为 `"Auxiliary_MaterialInfomation"`
2. 主键 `uint` 改 `bigint`
3. 字段对齐（见对照表）
4. `code` 在老库可能是 `Code` 或 `MaterialCode`；`category` 可能是 `Type`/`Classification`

**Verify**: `./scripts/verify.sh`，列表/创建接口手测。

---

#### Task 1.5 `drug_catalogs` → `Auxiliary_DrugInfomation`

**目标文件**:
- `internal/models/treatment_config.go`（`DrugCatalog` 部分）
- `internal/services/treatment_config_service.go`

**改造要点**: 同 1.4 逻辑。

---

### Phase 2 — `rename-fields` / `same-name` 字段对齐（3-4 小时）

#### Task 2.1 `orders` → `Order_PatientOrder`（已部分完成）

**状态**: `patient_core_service.buildActiveOrders` 已读老表，但 `Order` model 的 `TableName()` 还是 `orders`。

**目标文件**:
- `internal/models/treatment.go`（`Order` 部分）
- `internal/services/order_service.go`
- `internal/services/order_cron.go`（过期定时任务的 SQL）

**改造要点**:
1. `TableName()` 改为 `"Order_PatientOrder"`
2. 字段 column tag 全量补齐（参照 `LEGACY_TABLE_FIELD_MAPPING.md` 3.1.3 已有字段）
3. `status` 映射（枚举表）：新 `待执行/执行中/已执行/已停止` ↔ 老库 `Status`（bigint 枚举）——查老库字段含义
4. `type` 映射：新 `长期/临时` ↔ 老 `Type`（可能是 `1/2` 或中文）
5. **过期任务 SQL 更新**：`UPDATE "Order_PatientOrder" SET "Status"=? WHERE "Type"=? AND "EndTime" < NOW() AND "Status" IN (?, ?)`
6. 条件查询默认排除过期临时医嘱的逻辑保留

**Verify**: `go test ./internal/services -run Order`；手工触发一次 cron。

**Deliverable**: model + 2 service。

---

#### Task 2.2 Treatment 系列字段核对（6 张同名表）

**状态**: `Treatment_Treatment/BeforeCheck/BeforeSigns/DuringParam/AfterSigns/Alarm` 6 张表名一致，但字段命名存在新旧差异。

**目标文件**: `internal/models/treatment.go` 的后半段（行 321 之后）

**改造要点**（对每张表）：
1. 字段 PascalCase 化（如 `record_time` → `RecordTime`）
2. `creator_id` → `CreatorId`，`create_time` → `CreateTime`，`last_modify_time` → `LastModifyTime`
3. 时间字段用 `autoCreateTime`/`autoUpdateTime` tag
4. `Treatment.type` (int) 的值对应透析模式：新代码 1=HD / 2=HDF / 3=HP / 4=HD+HP —— 核对老库是否一致
5. `Treatment.status` (int) 同上

**特别注意**:
- `Treatment_BeforeCheck` 在老库对应 `Treatment_BeforeSymptom` + `Treatment_BeforeCheck`（老库 `BeforeSymptom` 是症状，`BeforeCheck` 是核对），**不要合并**，代码侧保持 `TreatmentBeforeCheck` 只映射老库 `Treatment_BeforeCheck`
- 若老库某字段不存在（如新代码 `Treatment_BeforeCheck.dry_weight`），用 `gorm:"-"` 标注为非持久化字段，并在 service 层从别处来源补齐

**Verify**: `go test ./internal/services -run Treatment`。

---

#### Task 2.3 `lab_reports` + `lab_report_items` → `LIS_Examination` + `LIS_ExaminationItem`

**状态**: 已在 `patient_core_service.buildLabTrends` 部分使用，但 `LabReport`/`LabReportItem` model 的 TableName 还是新表名。

**目标文件**:
- `internal/models/lab_report.go`
- `internal/services/lab_report_service.go`
- `internal/services/lis_sync_service.go`

**改造要点**:
1. `TableName()` 改为 `"LIS_Examination"` / `"LIS_ExaminationItem"`
2. 字段映射（对照表）：
   - `item_code` → `ItemCode`
   - `item_name` → `ItemName`
   - `clinical_diagnosis` → 老库可能无；查合并 md
   - `specimen_type` → `SpecimenType`（如存在）
   - `request_doctor` → `RequestDoctor`
   - `requested_at`/`sampled_at`/`received_at`/`reported_at` → `RequestTime`/`SampleTime`/`ReceiveTime`/`ReportTime`
   - `external_report_id`/`source_system`/`synced_at` → 老库**无**，在 model 加 `gorm:"-"`；LIS 同步服务仍保留这些字段在应用层用。
3. `abnormal_flag` (`H/L/N`) → 老库 `ResultSign`，映射表补齐：
   ```go
   var abnormalFlagNewToLegacy = map[string]string{"H":"H", "L":"L", "N":""}
   ```

**Verify**: `go test ./internal/services -run Lab`；手工调用 `GET /patients/:id/lab-reports`。

---

### Phase 3 — `rewrite` 结构重构（最复杂，预计 1-2 天）

#### Task 3.1 `treatment_plans` → `Plan_PatientPlan`

**复杂度**: ⭐⭐⭐⭐

**核心差异**:
- 新表 `treatment_plans` 用 4 个 JSONB 列（`dialysis_mode`/`anticoagulant`/`parameters`/`materials`）
- 老表 `Plan_PatientPlan` 是扁平列（`DialysisMethod`/`DialysisDuration`/`DryWeight`/`BF`/`FirstAnticoagulant`/`MaintainAnticoagulant`/`OddWeekFrequency`/`EvenWeekFrequency`/...）
- 材料清单在老库由 `Plan_PatientPlanMaterial` 子表承载

**目标文件**:
- `internal/models/treatment.go`（`TreatmentPlan` 部分）
- `internal/services/treatment_service.go`（或 `prescription_service.go`）
- 新增 `internal/models/plan_patient_plan_material.go`（或放在 `treatment.go` 里）

**改造步骤**:

1. **保留** `TreatmentPlan` 作为**对外 DTO**（handler/前端契约不变），新增 `LegacyPlanPatientPlan` struct 对应 `Plan_PatientPlan` 物理表。
2. 新增 `LegacyPlanPatientPlanMaterial` 对应 `Plan_PatientPlanMaterial`。
3. **读取路径**：service 查两张老表，组装 JSON 字段：
   ```go
   func toTreatmentPlan(legacy LegacyPlanPatientPlan, mats []LegacyPlanPatientPlanMaterial) *TreatmentPlan {
       return &TreatmentPlan{
           ID: legacy.ID,
           PatientID: legacy.PatientId,
           WeeklyFrequency: legacy.OddWeekFrequency,
           BiweeklyFrequency: legacy.EvenWeekFrequency,
           Duration: legacy.DialysisDuration,
           DryWeight: legacy.DryWeight,
           DialysisMode: DialysisMode{
               Mode: MapDialysisMethodLegacyToNew(legacy.DialysisMethod),
               BloodFlow: legacy.BF,
               // ...
           },
           Anticoagulant: Anticoagulant{
               InitialDrug: legacy.FirstAnticoagulant,
               MaintenanceDrug: legacy.MaintainAnticoagulant,
               // ...
           },
           Parameters: DialysisParameters{...},
           Materials: materialsFromLegacy(mats),
       }
   }
   ```
4. **写入路径**：service 接受 `TreatmentPlan`（前端传入），**反向拆解**到 `Plan_PatientPlan` 扁平列 + `Plan_PatientPlanMaterial` 子表：
   - 事务：先 upsert `Plan_PatientPlan`，再按 `PlanId` delete + insert 所有 material
5. **新增枚举映射**（补 `legacy_enum_maps.go`）：
   - `DialysisMode`: 新 `HD/HDF/HP/HD+HP` ↔ 老 `DialysisMethod`（值待查老库实际数据）
   - `Anticoagulant 药物名`: 新"普通肝素"等文本 ↔ 老 `FirstAnticoagulant`（可能是 dict 编码）
6. **无处安放的字段**（新表有、老表无）：
   - `dialysis_mode.substituteInputMode/substituteFlow/substituteVolume/bv/frequencyDesc/autoConfirm/status/notes` → 用 `Plan_PatientPlan.Note`（JSON 拼装）或在 `Auxiliary_JsonData` 建一条关联记录。选择**用 Note 存 JSON 子段**（最小改动）。
   - `parameters` 中 `dialysateType/group/flowRate/na/ca/k/hco3/glucose/conductivity/temp/volume` → 老库 `Plan_PatientPlan` 有部分字段（`DialysateType/FlowRate/Na/Ca/K/HCO3/Glucose/Conductivity/Temp/DialysateVolume`），逐字段核对；缺失的走 Note JSON
7. **删除**：老库统一软删（`IsDisabled=true`），不物理删。
8. **状态**：新 `status='启用'/'禁用'` ↔ 老 `IsDisabled=false/true`。

**Verify**:
```bash
./scripts/verify.sh
go test ./internal/services -run "TreatmentPlan|Plan"
# 手工测：GET /patients/:id/treatment-plan、POST 同路径
```

**Deliverable**:
- `internal/models/treatment.go` 更新（`TreatmentPlan` 保留为 DTO，增加 legacy struct）
- `internal/services/treatment_service.go` 重写读写方法
- `internal/services/legacy_enum_maps.go` 补充 dialysis/anticoagulant 映射

---

#### Task 3.2 `prescriptions` → `Plan_PatientPrescription` + `Plan_PatientPrescriptionMaterial`

**复杂度**: ⭐⭐⭐⭐

**模式**: 与 Task 3.1 同型 —— 新表 JSONB 多列 → 老库父 + 子表。

**目标文件**:
- `internal/models/treatment.go`（`Prescription` 部分）
- `internal/services/prescription_service.go`

**改造步骤**:

1. 类比 3.1 建 `LegacyPlanPatientPrescription` + `LegacyPlanPatientPrescriptionMaterial`
2. `order_items`（新） —— 老库 **没有**独立存储，要么：
   - (A) 写入 `Plan_PatientPrescription.OrderItemsJson`（若存在 text 字段）
   - (B) 每条 `orderItem` 在 `Auxiliary_JsonData` 存一条，类型 `PRESCRIPTION_ORDER`
   - **推荐** (A) —— 找老库 `Plan_PatientPrescription` 中的 `Note/Content/JsonContent` 字段（查合并 md 确认）
3. "提取长嘱"操作：从 `Order_PatientOrder` 查患者长期医嘱，生成 `order_items` 快照，同时从 `Plan_PatientPlan` 复制参数
4. 状态机约束（新）保留在 service 层：Update 仅允许 `待执行`

**Verify**:
```bash
./scripts/verify.sh
go test ./internal/services -run Prescription
# 手工：POST /patients/:id/prescriptions、提取长嘱按钮
```

---

#### Task 3.3 `plan_templates` → `Plan_PlanTPL` + `Plan_PlanTPLMaterial`

**复杂度**: ⭐⭐⭐

**目标文件**:
- `internal/models/treatment_config.go`（`PlanTemplate` 部分）
- `internal/services/treatment_config_service.go`

**改造步骤**:

1. 类比 Task 3.1 的思路，但操作的是**模板**（即 `Plan_PlanTPL` + `Plan_PlanTPLMaterial`）
2. `template_content` JSONB → 老库对应字段拆散填到 `Plan_PlanTPL` 扁平列
3. CRUD 走事务：父表 upsert + 子表 delete/insert

**Verify**: `./scripts/verify.sh` + 模板管理界面手测。

---

#### Task 3.4 `dict_types` + `dict_items` → `CodeDictionary_CodeDictionarys`

**复杂度**: ⭐⭐⭐

**核心差异**:
- 新表：type + item 两张表
- 老表：单表树形（自引用 parent，`Category` 字段区分字典类型）

**目标文件**:
- `internal/models/dict.go`
- `internal/services/dict_service.go`

**改造步骤**:

1. 保留 `DictType`/`DictItem` 作为 DTO
2. 新增 `LegacyCodeDictionary` struct → `CodeDictionary_CodeDictionarys`
3. `DictType.code` → 老库 `Category` 字段（查合并 md 的 `CodeDictionary_CodeDictionarys` 确认）
4. `DictItem.type_code` → `Category`；`DictItem.code` → `Code`；`DictItem.name` → `Name`；`DictItem.parent_code` → `ParentId` 或 `ParentCode`
5. `ListTypes`：`SELECT DISTINCT Category FROM CodeDictionary_CodeDictionarys WHERE IsDisabled=false`
6. `GetItemsByTypeCode(typeCode)`：`SELECT * FROM CodeDictionary_CodeDictionarys WHERE Category=?`
7. **fallback 保留**：如果 `CodeDictionary_CodeDictionarys` 表本身不存在（遗留环境），仍走 `legacyFallbackDictTypes/Items`

**Verify**: `./scripts/verify.sh` + `GET /dict/types`、`GET /dict/items/DIALYSIS_MODE`。

---

### Phase 4 — `fold` / `split-to-many`（2-3 天，按需拆分）

#### Task 4.1 `patient_basic_infos` → 4 张 Register 表合并

**复杂度**: ⭐⭐⭐⭐⭐（最复杂）

**核心差异**:
- 新表 `patient_basic_infos`（1:1 扩展，36 列）
- 老库字段分散在：`Register_PatientInfomation` / `Register_Hospitalization` / `Register_IDInfomation` / `Register_FamilyMember`

**目标文件**:
- `internal/models/patient_basic_info.go`（保留 DTO）
- `internal/services/patient_basic_service.go`
- `internal/services/patient_basic_types.go`

**参考**: `docs/basic-info-legacy-gap-analysis.md` 已做完整字段差异分析，**直接按照该文档执行**。

**改造步骤**（高层）:

1. `PatientBasicInfo` 结构**不改 TableName**，但加 `gorm:"-"` 标记（不落库）
2. `GetBasicInfo(patientID)`:
   - 从 `Register_PatientInfomation` 取姓名/性别/拼音/民族/血型/职业/婚姻/教育/地址/电话等（`buildPersonalInfo`）
   - 从 `Register_Hospitalization` 最近一条取就诊类别/住院号/病房等（`buildMedicalInfo`）
   - 从 `Register_IDInfomation` 取最近一条非禁用证件（`buildMedicalInfo` 补充）
   - 从 `Register_FamilyMember` 取主要联系人（`buildContactInfo`）
3. `UpdateBasicInfo(patientID, req)`: 反向拆解，按目标表 UPSERT
4. 写入时注意**枚举转换**（`patient_type/id_type/marital_status`）
5. 对 `docs/basic-info-legacy-gap-analysis.md` 第 6 节列出的每个"当前返回但未从数据库取"的字段，必须补齐取数

**Verify**:
```bash
./scripts/verify.sh
go test ./internal/services -run PatientBasic
# 手工：GET /patients/:id/basic-info，核对所有 tab 字段都不空（有真实测试数据时）
```

**Deliverable**: 3 个目标文件 + 可能新增的辅助方法。

---

#### Task 4.2 `medical_histories` → 7 张 Register 分表

**复杂度**: ⭐⭐⭐⭐⭐

**核心差异**:
- 新表 33 列扁平
- 老库 7 张：`Register_MedicalHistory` / `_Allergen` / `_Complication` / `_Diagnosis` / `_Pathology` / `_Protopathy` / `_Tumor`

**目标文件**:
- `internal/models/patient.go`（`MedicalHistory` 部分）
- `internal/services/medical_history_service.go`

**改造步骤**:

1. `MedicalHistory` 保留为 DTO（handler 不变）
2. 新增 7 个 legacy struct，每个对应一张 Register 分表
3. `GetMedicalHistory(patientID)`:
   - `Register_MedicalHistory` 读基础病史（current_illness/past_history/family_history/...）
   - `Register_Allergen` 读 allergen_*（**多条时取最新或合并**——需要业务决策；推荐**取最新**）
   - `Register_Tumor` 读 tumor_*
   - `Register_Complication` 读 complication_*
   - `Register_Pathology` 读 pathology_*
   - `Register_Protopathy` 读 primary_disease_*
   - `Register_Diagnosis` 读 disease_diagnosis
4. `UpdateMedicalHistory(patientID, req)`:
   - 事务：7 张表分别处理：
     - `MedicalHistory` UPSERT（WHERE PatientId=?）
     - 其它 6 张表采用 "delete existing + insert new"（因老库一对多）
5. **注意**：新表每类病史只有一条记录（name/content/type/check_time/check_doc），老库可能有多条；按"最新 1 条"规则简化

**Verify**:
```bash
./scripts/verify.sh
go test ./internal/services -run MedicalHistory
# 手工：GET /patients/:id/medical-history、PUT 同路径
```

---

### Phase 5 — `multi-join` 与目录表收尾（半天）

#### Task 5.1 修正 `devices` model 的 TableName

**状态**: `device_service` 已实现老库多表 join，但 `Device.TableName()` 还是 `"devices"`。这导致 model 被其它 service 误用时会查错表。

**改法**: 直接删掉 `Device` 的 `TableName()` 覆盖，让 model 仅作 DTO（`gorm:"-"`），或标注 `TableName() → "Auxiliary_EquipmentInfomation"` 但所有字段加 `gorm:"-"`（因字段组合自多表）。

**推荐**: 把 `Device` struct 移出 `models/` 到 `services/device_types.go`，彻底断开 GORM 绑定。

**Verify**: `grep -rn "models.Device" ai-hms-backend/` 无残留，`./scripts/verify.sh` 通过。

---

#### Task 5.2 `inventory_items` + `stock_logs` 决策

**状态**: 老库 `Stock_*` 族存在，新代码 `inventory.go` 独立。

**决策**:
- 如果系统实际不用 inventory 模块（看 `internal/api/v1/inventory_handler.go` 是否被注册），**保留 app-only**，不迁移，给 `Inventory*` 加 `gorm:"-"`。
- 如果必须对接 `Stock_*`，按 rewrite 做（类似 Task 3.1）。

**判定**: 检查 `cmd/server/main.go` 是否调 `RegisterInventoryRoutes`。当前 main.go 里**有**注册，但若生产环境不用，仍可放入"Phase 6 不迁移"。

**本任务产出**: 在本迁移计划末尾追加决策结果（一段备注）。codex 阅读后若选择"不迁移"，标 DONE；若选择"迁移"，新建 Task 5.3 展开。

---

#### Task 5.3 `order_templates` + `order_template_items` → `Order_OrderTPL`

**核心差异**: 新代码为模板 + 条目两张；老库目前只见到 `Order_OrderTPL`（父表是否含子条目需确认）。

**改造步骤**:

1. 核对 `老血透数据库表结构-合并版.md` 中 `Order_OrderTPL` 字段，找是否有 `TemplateId` 区分或 `ItemsJson` 存条目
2. 若老库**确实只有一张表**：
   - 每条 `OrderTemplate` + 其 `OrderTemplateItem` 合并为 `Order_OrderTPL` 里一组行（按 `TemplateId` 分组）
   - "模板"层级的属性（name/type/category）冗余到每行，或用代码里第一行表示 header
3. Service 读：`GROUP BY TemplateId`
4. Service 写：事务删原组 + 插新组

**Verify**: `./scripts/verify.sh` + 模板管理 UI 手测。

---

### Phase 6 — `app-only` 表确认不改（半小时）

#### Task 6.1 明确保留的应用层表清单

**What**: 检查以下表**不改**，在代码里保留独立 schema。但因为老库模式禁止 `AutoMigrate`，这些表**在老库中不存在**，运行时会报 "relation does not exist"。

| 表 | 处理策略 |
|----|---------|
| `users` | **外部管理**——在 startup SQL 手动建表（生产 DBA 执行）；或改造为读老库 `Identity/Organ` 体系（超出本计划）。先**建表策略**：在 `ai-hms-backend/scripts/app_only_tables.sql` 放 CREATE TABLE 脚本，由 DBA 在生产老库执行。 |
| `permissions` / `role_permissions` | 同上 |
| `dict_types` / `dict_items` | **不需要**建表，因为 3.4 已改为读 `CodeDictionary_CodeDictionarys`；且 service 有 fallback 到内置常量 |
| `integration_hdis_settings` | 同 `users` 方式 |
| `exam_reports` / `patient_key_indicators` | 同上（若业务需要落库）；或用 HDIS 实时同步，不落库（改 service 层） |
| `clinical_tasks` | 同 `users` |
| `label_tasks` | 同 `users`（如果 inventory 决策不迁移） |
| `inventory_items` / `stock_logs` | 看 Task 5.2 决策 |

**Deliverable**:
- `ai-hms-backend/scripts/app_only_tables.sql`（新文件）—— 包含 `users`、`permissions`、`role_permissions`、`integration_hdis_settings`、`exam_reports`、`patient_key_indicators`、`clinical_tasks`、`label_tasks`（以及 inventory 相关如果决定保留）的 `CREATE TABLE IF NOT EXISTS` 语句。**仅供 DBA 手动执行**，代码运行时不触发。

**Verify**: 脚本语法正确 (`psql -h ... -U ... -c "EXPLAIN ..."` 不可行；改用 `pg_query_go` 或简单 `psql --dry-run` 一次）。

---

### Phase 7 — 回归与验证（1 天）

#### Task 7.1 全量单测

```bash
cd ai-hms-backend
./scripts/verify.sh
go test ./... -v | tee /tmp/full-test.log
```

**要求**: 全绿，或失败项在本计划已标为 TODO。

---

#### Task 7.2 冒烟测试

```bash
cd ai-hms-backend
# 前置：确保连到真老血透库、JWT_SECRET 等已配
./scripts/smoke_test.sh http://localhost:8080
```

**要求**: 登录、患者列表、患者详情、治疗方案、处方、排班 全部 200。

---

#### Task 7.3 前端手工走查

按 `docs/basic-info-legacy-gap-analysis.md` 的字段清单，在前端走查：

- 患者列表分页排序
- 患者详情 basic-info 每一栏都填充
- 医疗历史 tab 每类记录（过敏/并发症/肿瘤等）能读写
- 排班能创建、修改、取消
- 治疗方案能新建、启用、禁用
- 每日处方能创建、提取长嘱、执行
- 字典 fallback + 老库 `CodeDictionary_CodeDictionarys` 两种路径都可用

**Deliverable**: `thoughts/migration-verification-YYYYMMDD.md`，记录走查截图/问题。

---

## 5. 任务依赖图

```
Phase 0 (准备)
   │
   ▼
Phase 1 (rename)             —— 独立，各任务并行
   ├── 1.1 Hospitalization
   ├── 1.2 Schedule*
   ├── 1.3 AdjustmentRecord
   ├── 1.4 MaterialCatalog
   └── 1.5 DrugCatalog
         │
         ▼
Phase 2 (rename-fields/same-name)
   ├── 2.1 Order ─────┐
   ├── 2.2 Treatment* │  2.1/2.2 互相独立；但 2.1 被 3.2 依赖
   └── 2.3 LabReport  │
         │            │
         ▼            │
Phase 3 (rewrite)     │
   ├── 3.1 TreatmentPlan  (依赖 2.2 Treatment)
   ├── 3.2 Prescription   (依赖 3.1 TreatmentPlan 和 2.1 Order)
   ├── 3.3 PlanTemplate   (独立)
   └── 3.4 Dict           (独立)
         │
         ▼
Phase 4 (fold/split)
   ├── 4.1 PatientBasicInfo
   └── 4.2 MedicalHistory
         │
         ▼
Phase 5 (multi-join + catalog)
   ├── 5.1 Device (fix model)
   ├── 5.2 Inventory decision
   └── 5.3 OrderTemplate (可能)
         │
         ▼
Phase 6 (app-only 表脚本)
         │
         ▼
Phase 7 (回归)
```

---

## 6. 每任务执行模板（Codex 必读）

**完成每一个 Task 都按此模板操作**：

```markdown
### Task {N.M}: {标题}

**Status**: pending → in_progress → completed

**Pre**:
- 工作分支已切换到 `feat/legacy-schema-alignment`
- `./scripts/verify.sh` 当前状态绿

**Do**:
1. 对照 `docs/migration-field-map.md` 里本表段落
2. 对照 `老血透数据库表结构-合并版.md` 确认字段类型与业务注释
3. 按本 Task 的 "目标文件" 清单修改
4. 同步更新 `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`（追加一节说明本 Task 的映射）
5. 涉及新枚举，补 `internal/services/legacy_enum_maps.go`

**Verify**:
- `./scripts/verify.sh` 绿
- `go test ./... -count=1` 绿
- 若有 handler 影响，写最小单测（service 层）或跑 smoke script 的相关段落

**Commit**:
- 提交信息：`feat(legacy): migrate {新表名} to {老表名}`
- 包含变更：model + service + mapping doc

**After**:
- 标记 Task 为 completed
- 进入下一个 Task；同 Phase 任务可并行 PR
```

---

## 7. 风险与回滚

| 风险 | 影响 | 缓解 |
|------|------|------|
| 老库字段名与 `数据库表结构.md` 不符 | 查询失败 | 改动前 `psql` 连老库 `\d "表名"` 确认 |
| 枚举值实际分布与设计不符（如 PatientType 不是 10/20 而是中文） | 数据错位 | 每个 rewrite 任务先 `SELECT DISTINCT X FROM 老表 LIMIT 20` 采样 |
| 老库有 NOT NULL 约束但新代码没填 | INSERT 失败 | 每个 Create 逻辑核对老表 NN 列表（合并 md 已标） |
| 并发下 `MAX(Id)+1` 冲突 | INSERT 失败 | 用 snowflake；`nextLegacyNumericID` 已有事务保护，压力大时改为序列 |
| `app-only` 表未建表 | startup 报错 | Phase 6 脚本必须在生产老库执行 |
| 前端契约 drift | UI 报错 | 所有 handler Request/Response 结构**不改**；DTO 变化都吸收在 service 内部 |

**回滚策略**: 每个 PR 独立 merge，单 PR 回滚即可；不跨表涉及数据修改，无需 DB 回滚。

---

## 8. 附录索引

| 附录 | 用途 |
|------|------|
| `docs/migration-field-map.md` | 新→老字段逐行对照（1600+ 行） |
| `老血透数据库表结构-合并版.md` | 老库字段级权威 |
| `DATABASE_DESIGN.md` | 新血透表设计源头 |
| `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md` | 本次迁移的持续更新日志（每 Task 追加） |
| `docs/basic-info-legacy-gap-analysis.md` | Task 4.1 专用详尽分析 |

---

## 9. 完成判定（Definition of Done）

整个迁移 **DONE** 的判据：

- [ ] `grep -rn "TableName" ai-hms-backend/internal/models/*.go | grep -vE "Register_|Schedule_|Plan_|Treatment_|Order_|Auxiliary_|LIS_|CodeDictionary_"` 只剩 `app-only` 白名单表
- [ ] `./scripts/verify.sh` 全绿
- [ ] `./scripts/smoke_test.sh` 全绿（连真老库）
- [ ] 前端所有页面手工走查无 500/404
- [ ] `LEGACY_TABLE_FIELD_MAPPING.md` 已包含所有迁移表的映射记录
- [ ] `scripts/app_only_tables.sql` 已交付 DBA 并在老库执行成功

---

*本计划生成于 2026-04-17，基于代码快照（branch: master, HEAD: c919145）。每次结构大变时请重新校对 `docs/migration-field-map.md`。*
