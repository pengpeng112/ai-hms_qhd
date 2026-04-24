---
session: ses_279b
updated: 2026-04-13T10:05:34.487Z
---

# Session Summary

## Goal
Find and document all sources in the ai-hms-backend repo that reference the legacy / real database table structures (table name constants, quoted legacy table names, field references, raw SQL, and model-to-table mappings) with priority on dashboard, patient, treatment, schedule, equipment areas — success = a file-grouped inventory of where legacy tables and key fields (TenantId, IsDisabled, TreatmentTime / StartTime, ShiftId) are referenced and whether the known legacy table names are present/verified.

## Constraints & Preferences
- Explore-only (no file modifications).
- Maximize search effort within the local repository; focus on internal/services, internal/models, internal/api.
- Identify quoted old table names (e.g. "Register_PatientInfomation").
- Identify occurrences/sources of TenantId, IsDisabled, TreatmentTime, ShiftId.
- Avoid searching or reporting unrelated frontend/UI files.
- Do not implement or change code.
- Tools used: file listing + direct file reads (PowerShell Get-ChildItem + file reads). (Note: ripgrep (rg) was unavailable.)

## Progress
### Done
- [x] Located and listed .go files under internal/services, internal/models, internal/api and other backend directories (PowerShell listing). Key files enumerated (dashboard_service.go, patient/service files, models files, etc.).
- [x] Read contents of multiple relevant files (listed exactly under "File Operations - Read" below). Extracted notable references from those files:
  - internal/services/dashboard_service.go: defines legacy table name constants (some quoted) and runs raw-table queries using GORM Table(...).Key observations:
    - Constants include: legacyPatientTable = `"Register_PatientInfomation"`, legacyPatientShiftTable = `"Schedule_PatientShift"`, legacyShiftTable = `"Schedule_Shift"`, legacyTreatmentTable = `"Treatment_Treatment"`, legacyEquipmentTable = `"Auxiliary_EquipmentInfomation"`.
    - Uses quoted column names and raw SQL fragments: DATE("StartTime"), EXTRACT(hour FROM "StartTime"), DATE("TreatmentTime") — i.e., queries reference StartTime and TreatmentTime column names directly.
    - Where clauses use a symbol legacyTenantID (used but not defined in this file).
    - Inventory alerts query uses a local inventory table (inventory_items) (non-legacy) and plain WHERE conditions (is_disabled; min_stock/stock).
  - internal/models/patient.go: model Patient maps TableName() -> "Register_PatientInfomation" (verifies Register_PatientInfomation).
    - Contains TenantID struct tag gorm:"column:TenantId" and other fields; VascularAccess and other nested types contain TenantID and IsDisabled fields; IsDisabled appears on VascularAccess struct and others.
  - internal/models/treatment.go: large model file; the Treatment struct maps TableName() -> "Treatment_Treatment" (verifies Treatment_Treatment). Treatment struct contains TenantId, ShiftId, IsDisabled, many time fields (TreatmentDate, SignInTime, ReceptionTime, etc.). Dashboard service queries reference StartTime (raw legacy column) while the model uses TreatmentDate / other fields — potential mismatch to verify.
  - internal/models/schedule.go: PatientShift model TableName() -> "patient_shifts" (lowercase), not `"Schedule_PatientShift"`; Shift TableName -> "shifts" (not `"Schedule_Shift"`). PatientShift struct includes ShiftId, ScheduleDate, TenantId, IsDisabled.
  - internal/models/legacy/*: legacy package contains models mapping some legacy tables (e.g., OrganEmployee.TableName() -> "Organ_Employee"; IdentityUser.TableName() -> "Identity_Users") — confirms a legacy model package exists for older schema mappings.
  - internal/models/device.go and internal/services/device_service.go: device model TableName() -> "devices" (non-legacy) but dashboard_service references legacyEquipmentTable = `"Auxiliary_EquipmentInfomation"` (legacy equipment table). Device model contains IsDisabled field; DeviceService ensures schema and uses is_disabled filter.
  - internal/services/treatment_service.go: uses models.Treatment struct, filters and updates on columns named in Go model (patient_id, status, type, treatment_date), uses Preload of related associations (Shift, Schedule, etc.). Create/Update operations use snake_case DB column names (e.g., shift_id, is_disabled) consistent with GORM conventions.
  - internal/services/patient_shift_service.go and internal/services/shift_service.go: use models.PatientShift and models.Shift (which map to patient_shifts and shifts) and reference ShiftId, IsDisabled, TenantId in code and queries.
  - internal/database/migrate.go: AutoMigrate is disabled in legacy mode (logged "[LEGACY-DB] AutoMigrate call blocked: permanently disabled") — explicit decision to avoid running DDL on legacy DB.
  - internal/services/legacy_id_helper.go: functions to parse/produce legacy ID values (LegacyID type conversions) — indicates special handling for legacy numeric IDs.
- [x] Verified presence of the user's previously-known real/legacy tables in code:
  - Register_PatientInfomation: verified via models.Patient.TableName() returning "Register_PatientInfomation".
  - Treatment_Treatment: verified via models.Treatment.TableName() returning "Treatment_Treatment".
  - Schedule_PatientShift: found used as a quoted constant in dashboard_service (`"Schedule_PatientShift"`) but models.PatientShift.TableName() returns "patient_shifts" (lowercase) — mismatch that needs reconciliation.
  - Schedule_Shift: dashboard_service uses `"Schedule_Shift"` constant; models.Shift.TableName() returns "shifts" — mismatch.
  - Auxiliary_EquipmentInfomation: present as legacyEquipmentTable constant in dashboard_service (`"Auxiliary_EquipmentInfomation"`).
- [x] Captured the main issues encountered while searching:
  - rg (ripgrep) was not available in the environment; grep attempts via provided tools failed with "Executable not found in $PATH: 'rg'".
  - Some PowerShell Get-ChildItem commands produced warnings/errors for multi-filter usage; adjusted to list files and read directly.

### In Progress
- [ ] Consolidating a file-grouped inventory mapping observed legacy table name constants, quoted table names, model TableName() returns, field references (TenantId, IsDisabled, ShiftId, StartTime/TreatmentTime) and raw SQL fragments from service files (dashboard queries). The raw extraction and reconciliation (e.g., "Schedule_PatientShift" constant vs models.patient_shifts mapping) is being assembled but not yet presented in grouped form.
- [ ] Locating definition(s) of legacyTenantID used in internal/services/dashboard_service.go (used in many Where clauses) — currently referenced but its declaration/location not yet found in the read files.

### Blocked
- Missing ripgrep (rg) tool prevented fast full-repo regex searches; required switching to manual file listing + targeted file reads. This slows exhaustive discovery.  
- (none) other than the above.

## Key Decisions
- **Search approach change (rg unavailable)**: used PowerShell file listings and targeted file reads of the directories internal/services, internal/models, internal/api. Rationale: rg was not available in the execution environment; reading key files directly still provided the required information and kept to the "explore-only" constraint.
- **Prioritize internal/services, internal/models, internal/api**: focused reads on these directories as user mandated; rationale: these are the layers most likely to contain legacy table constants, raw SQL usages, and model->table mappings.

## Next Steps
1. Produce the requested exhaustive, file-grouped report mapping (for each file): file path, detected legacy table names (constants and quoted strings), TableName() return values in models, direct column name usages in raw SQL fragments or Table(...) queries, and the key fields referenced (TenantId, IsDisabled, TreatmentTime/StartTime, ShiftId). Highlight the dashboard/patient/treatment/schedule/equipment items first.
2. Search for the definition(s) of legacyTenantID (and any other legacyTenant constants) across the repository (if rg still unavailable, iterate targeted file reads for config files, constants files, or package-level vars).
3. Reconcile mismatches: list places where dashboard_service references quoted legacy table names (e.g., "Schedule_PatientShift" / "Schedule_Shift") but model TableName() returns different (lowercase) table names (patient_shifts / shifts) — provide suggestions where verification is required.
4. Continue exhaustive scan of internal/models and internal/services to find all occurrences of TenantId, IsDisabled, StartTime/TreatmentTime, ShiftId, and report file + line excerpt for each occurrence.
5. If deeper grep is required, request permission to install or enable a code search tool (rg) or allow a repository-wide grep run; otherwise continue manual file-read passes.

## Critical Context
- dashboard_service.go is a central consumer of legacy tables in analytics/dashboards. It defines constants with quoted legacy table names:
  - legacyPatientTable = `"Register_PatientInfomation"`
  - legacyPatientShiftTable = `"Schedule_PatientShift"`
  - legacyShiftTable = `"Schedule_Shift"`
  - legacyTreatmentTable = `"Treatment_Treatment"`
  - legacyEquipmentTable = `"Auxiliary_EquipmentInfomation"`
  - inventoryTable = "inventory_items"
- models confirm some legacy mappings:
  - internal/models/patient.go: TableName() returns "Register_PatientInfomation" — direct match to legacyPatientTable (good).
  - internal/models/treatment.go: TableName() returns "Treatment_Treatment" — direct match to legacyTreatmentTable (good).
  - internal/models/schedule.go: PatientShift.TableName() returns "patient_shifts" (lowercase) while dashboard uses `"Schedule_PatientShift"` (possible legacy name vs app-mapped name mismatch).
  - internal/models/legacy/*: contains legacy models (Organ_Employee, Identity_Users) which explicitly return legacy table names.
- Key fields seen in models and services:
  - TenantId / TenantID: appears across legacy and modern models (VascularAccess, OrganEmployee, many tables). Service queries often filter by TenantId (dashboard_service uses legacyTenantID in Where).
  - IsDisabled / is_disabled: present in many models (Device.IsDisabled, PatientShift.IsDisabled, Treatment.IsDisabled, etc.) and used in queries (COALESCE("IsDisabled", false) in dashboard_service; where clauses with is_disabled in DeviceService and others).
  - ShiftId / shift_id: present in PatientShift and Treatment models; used in queries and Preloads (Preload("Shift")).
  - TreatmentTime / StartTime / TreatmentDate / StartTime: Services and dashboard queries use DATE("StartTime") and DATE("TreatmentTime") (raw legacy column names). Models use TreatmentDate and other time fields — requires verification of mapping between StartTime/TreatmentTime in legacy DB and model fields (e.g., TreatmentDate/Treatment's other time fields).
- AutoMigrate policy: internal/database/migrate.go explicitly blocks AutoMigrate in legacy mode (AutoMigrate returns a permanent-disabled error), so code avoids DDL against legacy DB.
- Unresolved / to-verify items:
  - Definition and value of legacyTenantID referenced in dashboard_service.go (not found in the files read so far).
  - Exact mapping between legacy StartTime/TreatmentTime column names referenced in raw SQL and the model fields (e.g., TreatmentDate). Confirm whether models include gorm:"column:StartTime" tags (we saw some column tags mapping TenantId etc., but not StartTime in the portions read).
  - Whether schedule legacy table names used in dashboard_service ("Schedule_PatientShift", "Schedule_Shift") correspond to actual DB table names in production; models map to snake_case names ("patient_shifts", "shifts") suggesting the app may map downcased names or maintain compatibility via TableName overrides in other packages (or different modules).

## File Operations
### Read
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\database\migrate.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\device.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\legacy\doc.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\legacy\identity_user.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\legacy\organ_employee.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\patient.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\patient_basic_info.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\schedule.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\models\treatment.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\services\dashboard_service.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\services\device_service.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\services\legacy_id_helper.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\services\patient_shift_service.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\services\shift_service.go`
- `F:\python\前后端代码\ai-hms_qhd\ai-hms-backend\internal\services\treatment_service.go`

### Modified
- (none)

IMPORTANT NOTES / ERRORS ENCOUNTERED
- Attempted repository-wide regex searches using ripgrep (rg) via provided tools; failed with error: "Executable not found in $PATH: 'rg'". Switched to PowerShell file listings + targeted file reads.
- PowerShell Get-ChildItem had an argument/Filter multi-pattern issue in one command; adjusted listing strategy and read files directly.

If you want, I will now:
- produce the detailed file-by-file grouped inventory the user requested (table name constants, TableName() mappings, key fields occurrences and raw SQL fragments), sorted with dashboard/patient/treatment/schedule/equipment prioritized; and
- continue searching for legacyTenantID and any remaining references (using direct file reads if rg remains unavailable).
