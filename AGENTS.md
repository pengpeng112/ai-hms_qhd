# Agent Notes

## Repo Shape
- Two active apps: `ai-hms-backend/` (Go module `github.com/elliotxin/ai-hms-backend`, Gin, GORM) and `ai-hms-frontend/` (React 19, Vite, TS, Ant Design 6, Tailwind 4).
- Everything else (`ai-hms-v1.3-透析执行/`, `gorm/`, `old_system/`) is reference material, not the active codebase.
- Backend entry: `cmd/server/main.go`. Routes registered via `v1api.RegisterXxxRoutes(group)` in `main.go`. Handlers in `internal/api/v1`, services in `internal/services`, models in `internal/models`.
- Frontend entry: Vite app. Pages in `src/pages/`, axios facade in `src/services/restClient.ts`, all service exports through `src/services/index.ts`.
- Legacy schema truth: `老血透数据库表结构-合并版.md` and `新数据库表结构.md` at repo root. Field mapping doc: `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`.
- `.env` is NOT committed. Copy from `.env.example` or ask the user for connection details.

## Commands
- Backend run: `cd ai-hms-backend && go run ./cmd/server`
- Backend test: `cd ai-hms-backend && go test ./internal/services ./internal/api/v1`
- Backend single test: `cd ai-hms-backend && go test ./internal/services -run TestOrderService_Create`
- Backend verify (Unix): `cd ai-hms-backend && ./scripts/verify.sh` — gofmt + go vet + go build + go test
- Backend Windows build check: `cd ai-hms-backend && go build -o "$env:TEMP\check.exe" ./cmd/server`
- Frontend install/run: `cd ai-hms-frontend && npm install && npm run dev`
- Frontend gate: `cd ai-hms-frontend && npm run lint && npm run build` — build includes `tsc -b`
- Frontend unit tests: `cd ai-hms-frontend && npm run test` (vitest, jsdom)
- Smoke test (needs seed data): `cd ai-hms-backend && ./scripts/smoke_test.sh`, creds `test_admin / Test@123456`
- Docker: root `docker-compose.yml`; PostgreSQL is external (host set in `.env`).

## Database Hard Rules
- Backend connects only to the legacy dialysis PostgreSQL (`LEGACY_DB_*` env vars, falls back to `DB_*`).
- `internal/database/migrate.go` permanently blocks `AutoMigrate` and `DropTables` — never re-enable.
- GORM config: `SingularTable: true`, `NoLowerCase: true`. Legacy table/column names are case-sensitive; raw SQL must use double quotes (e.g. `WHERE "Id" = ?`).
- DB change boundary: old-table `ALTER TABLE`, old-table default changes, column renames, and old-table unique indexes require DBA manual scripts/review. Do not run them from app startup or request paths.
- Independent new tables may be created by deployment-stage idempotent SQL only; app runtime still must not execute DDL. Current catalog and rationale live in `docs/database-change-maintenance.md`.
- New independent tables currently include: `exam_reports`, `exam_report_items`, `external_patient_mappings`, `sync_job_configs`, `sync_job_runs`, `sign_record`, `Schedule_StaffDuty`, `Schedule_StaffDutyOverride`, `Schedule_Patient`.
- `JWT_SECRET`, `APP_SECRET`, `CORS_ALLOWED_ORIGINS` are required env vars — server exits without them.
- `AUTH_EMERGENCY_ENABLED=false` by default. Do not introduce default-credential fallbacks.

## Legacy DB Quirks (high trip risk)
- **TenantId filter**: all queries must filter `"TenantId" = 3` (the `LegacyTenantID` constant in `internal/services`). Forgetting it risks cross-tenant data or empty results.
- **Treatment date matching**: never use `DATE("StartTime")` alone — many records have NULL `StartTime` (not yet on-machine). Use `DATE(COALESCE("StartTime", "SignInTime", "ReceptionTime", "CreateTime"))`. This pattern is already defined as `treatmentBusinessDate` in `dashboard_service.go`.
- **Patient active/in-dept status**: NOT determined by `TreatmentStatus` column (mostly empty). Determined by latest `Register_OutCome.Type`: `'10'` = active/in-dept, `'20'` = transferred. See `patient_service.go` List/GetStats for the `DISTINCT ON` subquery pattern.
- **`CodeDictionary_CodeDictionarys` has NO `Id` column**. Columns are only: `Code, Type, Name, Sort, IsDisabled, OrganId, Builtin`. The dict service struct `legacyCodeDictionaryRow` correctly omits `Id`.
- **Dict dual-code system**: frontend uses unified codes (`DIALYSIS_MODE`, `INSURANCE_TYPE`), backend maps to/from legacy types (`DialysisMethod`, `ExpenseType`) via `legacyTypeToUnifiedCode` / `unifiedCodeToLegacyTypes` maps in `dict_service.go`. `ListTypes` returns unified codes; `CreateItem` converts unified→legacy before writing.
- **Fixed patient-info mappings**: `ID_TYPE → IDType`, `VISIT_CATEGORY → HospPatientType`, `INSURANCE_TYPE → ExpenseType`.

## Frontend Hard Rules
- TypeScript strict: `strict`, `noUnusedLocals`, `noUnusedParameters`, `erasableSyntaxOnly` all on. Type errors block `npm run build`.
- ESLint (see `eslint.config.js`): prohibits bare `rounded-xl/2xl/3xl`, `text-[10px]`/`text-[11px]`, and `!important`. Many page files are exempted (see the exemption list in config); `src/pages/dialysis-processing/**` is one of many.
- React Hooks rules enforced: no synchronous `setState` inside `useEffect`.
- Vite aliases: `@` → `src`, plus `@components`, `@layouts`, `@pages`, `@hooks`, `@services`, `@utils`, `@types`.
- Do not commit `dist/`, `.env*`, logs, HAR files, or binary artifacts.

## Migration Conventions
- Change one page/interface/module at a time. Never do a full-system replacement.
- Preserve existing frontend API paths, response shapes, and field names. Adapt on the backend side.
- If a field meaning is uncertain, record it in `docs/legacy-migration-uncertain-field-checklist.md` — do not guess.
- Before altering any dropdown/cascade/dictionary field, check `src/services/dictApi.ts` and `internal/services/dict_service.go`.

## Git
- Only commit/push when explicitly asked. Before pushing: `git status`, `git diff`, `git log --oneline -10`.
- Windows Credential Manager for GitHub auth: `git config --global credential.helper manager`.
- Conventional Commits (`feat:` / `fix:` / `chore:` / `refactor:`).
