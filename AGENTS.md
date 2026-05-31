# Agent Notes

## Repo Shape
- Two apps: `ai-hms-backend/` (Go, Gin, GORM) and `ai-hms-frontend/` (React 19, Vite, TS, Ant Design 6, Tailwind 4).
- Everything else (`ai-hms-v1.3-透析执行/`, `gorm/`, `old_system/`) is reference material, not the active codebase.
- Backend entry: `ai-hms-backend/cmd/server/main.go`. Routes in `internal/api/v1`, services in `internal/services`, models in `internal/models`.
- Frontend entry: Vite app; pages in `src/pages/`, API facade in `src/services/restClient.ts`, new API modules in `src/services/orderApi.ts` and `src/services/userApi.ts` etc. All service exports go through `src/services/index.ts`.
- Legacy schema truth: `老血透数据库表结构-合并版.md` and `新数据库表结构.md` at repo root. Field mapping doc: `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`.

## Commands
- Backend run: `cd ai-hms-backend && go run ./cmd/server`
- Backend test: `cd ai-hms-backend && go test ./internal/services ./internal/api/v1`
- Backend verify (needs Unix shell): `cd ai-hms-backend && ./scripts/verify.sh` — runs gofmt, go vet, go build, go test
- Windows build check (no exe in repo): `cd ai-hms-backend && go build -o "$env:TEMP\check.exe" ./cmd/server`
- Frontend install/run: `cd ai-hms-frontend && npm install && npm run dev`
- Frontend gate: `cd ai-hms-frontend && npm run lint && npm run build` — build includes `tsc -b`
- Docker: root `docker-compose.yml` pulls `ai-hms-backend:latest` + `ai-hms-frontend:latest`; PostgreSQL at `10.10.8.83` is external.

## Database & Backend Hard Rules
- Backend connects only to the legacy dialysis PostgreSQL (`LEGACY_DB_*` env vars, falls back to `DB_*`).
- `internal/database/migrate.go` permanently blocks `AutoMigrate` and `DropTables`. Never re-enable auto-migration, seed, or DDL.
- GORM config: `SingularTable: true`, `NoLowerCase: true`. Legacy table/column names are case-sensitive; SQL must use double quotes.
- `JWT_SECRET`, `APP_SECRET`, `CORS_ALLOWED_ORIGINS` are required env vars — server exits without them.
- `AUTH_EMERGENCY_ENABLED=false` by default. Do not introduce default-credential fallbacks.

## Frontend Hard Rules
- TypeScript strict: `strict`, `noUnusedLocals`, `noUnusedParameters`, `erasableSyntaxOnly` all on. Type errors block `npm run build`.
- ESLint (see `eslint.config.js`): prohibits bare `rounded-xl/2xl/3xl`, `text-[10px]`/`text-[11px]`, and `!important`. Files in `src/pages/dialysis-processing/**` are exempted.
- React Hooks rules enforced: no synchronous `setState` inside `useEffect`.
- Vite aliases: `@` → `src`, plus `@components`, `@layouts`, `@pages`, `@hooks`, `@services`, `@utils`, `@types`.
- Do not commit `dist/`, `.env*`, logs, HAR files, or binary artifacts.

## Migration Conventions
- Change one page/interface/module at a time. Never do a full-system replacement.
- Preserve existing frontend API paths, response shapes, and field names. Adapt on the backend side.
- If a field meaning is uncertain, record it in `docs/legacy-migration-uncertain-field-checklist.md` — do not guess.
- Before altering any dropdown/cascade/dictionary field, check `src/services/dictApi.ts` and `ai-hms-backend/internal/services/dict_service.go`.
- Fixed patient-info mappings: `ID_TYPE → IDType`, `VISIT_CATEGORY → HospPatientType`, `INSURANCE_TYPE → ExpenseType`.

## Git
- Only commit/push when explicitly asked. Before pushing: `git status`, `git diff`, `git log --oneline -10`.
- Windows Credential Manager for GitHub auth: `git config --global credential.helper manager`.