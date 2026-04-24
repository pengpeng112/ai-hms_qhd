---
session: ses_28a8
updated: 2026-04-10T04:50:46.167Z
---

# Session Summary

## Goal
Identify all existing ID generation logic and user authentication/password verification entry points in the backend to support legacy-db-adapter-plan-v2 T0-6/T0-8.

## Constraints & Preferences
- Read-only analysis only; do not modify files.
- Must cover `internal/services`, `internal/models`, `internal/utils`, and API login route/handler.
- Search keywords: `uuid`, `snowflake`, `bcrypt`, `password`, `login`, `auth`, `token`.
- Distinguish “already reusable” vs “missing/new required.”
- Do not infer database schema beyond code evidence.
- Preserve exact file paths and be specific.

## Progress
### Done
- [x] Located login route/handler in `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/cmd/server/main.go`:
  - `public.POST("/auth/login", loginHandler(jwtManager, userService))`
  - `loginHandler(...)` binds `LoginRequest`, calls `userService.Authenticate(...)`, then `jwtManager.GenerateToken(...)`.
- [x] Located authentication/password verification flow in `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/user_service.go`:
  - `Authenticate(username, password string)` fetches user by username and verifies password with `utils.CheckPassword(...)`.
- [x] Located password hash utilities in `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/password.go`:
  - `HashPassword(...)` uses `bcrypt.GenerateFromPassword`
  - `CheckPassword(...)` uses `bcrypt.CompareHashAndPassword`
- [x] Located JWT generation/parsing in `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/jwt.go`:
  - `GenerateToken(...)`, `ParseToken(...)`, `ValidateToken(...)`
- [x] Mapped ID generation mechanisms:
  - `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/id.go`
    - `GenerateID()` returns `uuid.New().String()`
    - `GeneratePatientID(count int)` returns `PYYYYMMDD%03d`
  - `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/idgen/snowflake.go`
    - Snowflake generator exists (`NextID()`), but no usage found in service/model call sites.
- [x] Identified major UUID/string-ID call sites:
  - `internal/services/patient_service.go`
  - `internal/services/patient_basic_service.go`
  - `internal/services/order_service.go`
  - `internal/services/prescription_service.go`
  - `internal/services/vascular_access_service.go`
  - `internal/services/dict_service.go`
  - `internal/services/exam_report_sync_service.go`
  - `internal/services/key_indicator_service.go`
  - `internal/services/lab_report_service.go`
  - `internal/services/lis_sync_service.go`
  - `internal/services/medical_history_service.go`
  - `internal/services/treatment_config_service.go`
- [x] Identified DB auto-increment / legacy int64 ID models:
  - `internal/models/hospitalization.go`
  - `internal/models/schedule.go`
  - `internal/models/treatment.go`
  - `internal/models/treatment_config.go` (`MaterialCatalog`, `DrugCatalog`)
- [x] Identified string primary-key models:
  - `internal/models/patient.go`
  - `internal/models/user.go`
  - `internal/models/dict.go`
  - `internal/models/treatment_config.go` (`PlanTemplate`, `OrderTemplate`, `OrderTemplateItem`)
- [x] Verified there are deprecated/new-db indicators in some model files:
  - `internal/models/patient.go`
  - `internal/models/hospitalization.go`
  - `internal/models/schedule.go`
- [x] Encountered tooling issues during search:
  - `rg`/`glob` not available in PATH from the tool layer.
  - Some `background_output(...)` lookups returned `Task not found` before the final valid outputs were retrieved.

### In Progress
- [ ] Consolidating findings into a structured inventory of reusable vs missing/new ID/auth pieces for T0-6/T0-8.

### Blocked
- (none)

## Key Decisions
- **Use existing UUID helpers where available**: `utils.GenerateID()` is already a reusable entry point and matches the current UUID strategy.
- **Treat `internal/utils/idgen/snowflake.go` as unused capability**: Snowflake support exists in code, but no repository call sites were found, so it is not currently part of the implemented flow.
- **Treat `GeneratePatientID` as legacy business-ID logic needing review**: It is count-based and format differs from the patient creation logic, so it should not be assumed safe for legacy adapter work without further design.
- **Separate auth from token issuance**: Password verification happens in `UserService.Authenticate`, while token creation happens in `JWTManager.GenerateToken`, making these the two integration points.

## Next Steps
1. Build the final inventory table of all ID generation mechanisms with file/function/evidence and mark reusable vs missing.
2. Build the final auth inventory: login route, handler, `Authenticate`, bcrypt helper, JWT helper, middleware token parsing.
3. List adapter-relevant dependencies on old table structure:
   - string UUID PKs
   - int64 auto-increment PKs
   - patient business ID generation
4. Summarize actionable T0-6/T0-8 implications, especially where Snowflake could be inserted and where legacy schema must remain unchanged.

## Critical Context
- `cmd/server/main.go` registers `POST /api/v1/auth/login` and uses `loginHandler(jwtManager, userService)`.
- `loginHandler` calls `userService.Authenticate(username, req.Password)` then `jwtManager.GenerateToken(...)`.
- `user_service.go: Authenticate` uses `utils.CheckPassword(password, user.Password)`.
- `password.go` uses bcrypt for both hashing and verification.
- `id.go` provides `GenerateID()` (UUID) and `GeneratePatientID(count int)` (`PYYYYMMDD%03d`).
- `idgen/snowflake.go` exists with `NextID()` but was not found in active call sites.
- `patient_service.go` creates patient IDs manually as `PYYYYMMDD%04d` with retry logic, and also uses `utils.GenerateID()` for related records.
- Many services still directly call `uuid.New().String()` instead of `utils.GenerateID()`.
- Several models use `int64` auto-increment IDs, which likely depend on existing table/sequence behavior.

## File Operations
### Read
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/cmd/server/main.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/middleware/auth.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/models/dict.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/models/hospitalization.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/models/patient.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/models/schedule.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/models/treatment_config.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/models/user.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/exam_report_sync_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/key_indicator_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/order_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/patient_basic_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/patient_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/prescription_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/treatment_config_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/user_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/services/vascular_access_service.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/id.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/idgen/snowflake.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/idgen/snowflake_test.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/jwt.go`
- `F:/python/前后端代码/ai-hms_qhd/ai-hms-backend/internal/utils/password.go`

### Modified
- (none)
