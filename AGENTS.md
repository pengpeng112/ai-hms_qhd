# Repository Guidelines

## Project Structure & Module Organization
This repository is split into `ai-hms-backend` and `ai-hms-frontend`.

- `ai-hms-backend/cmd/server` contains the Go entrypoint.
- `ai-hms-backend/internal/api/v1`, `services`, `models`, `middleware`, and `database` hold handlers, business logic, data models, cross-cutting middleware, and persistence code.
- `ai-hms-backend/scripts` includes verification, seed, import, and smoke-test helpers.
- `ai-hms-frontend/src/pages`, `components`, `layouts`, `services`, `contexts`, and `i18n` hold the React UI, shared building blocks, API clients, app state, and locale files.
- `docs` and `deploy` contain environment, deployment, and systemd guidance.

## Build, Test, and Development Commands
Run commands from the relevant subproject directory.

- `cd ai-hms-backend && go run ./cmd/server` starts the API locally.
- `cd ai-hms-backend && go test ./...` runs backend unit tests.
- `cd ai-hms-backend && ./scripts/verify.sh` checks formatting, `go vet`, and `go build ./cmd/server`.
- `cd ai-hms-frontend && npm install` installs frontend dependencies.
- `cd ai-hms-frontend && npm run dev` starts the Vite dev server.
- `cd ai-hms-frontend && npm run lint` runs ESLint.
- `cd ai-hms-frontend && npm run build` performs the TypeScript and production build check.
- `docker compose up -d` runs the packaged stack with the root `docker-compose.yml`.

## Coding Style & Naming Conventions
Go code should stay `gofmt`-formatted and follow standard package layout. Keep exported identifiers in `PascalCase`; unexported helpers use `camelCase`. React and TypeScript files use 2-space indentation, `PascalCase` for components such as `PatientDetail.tsx`, and `camelCase` for hooks and utilities such as `useOutcomeDict.ts`. Use the existing ESLint config in `ai-hms-frontend/eslint.config.js`. Do not commit generated artifacts.

## Testing Guidelines
Backend tests live beside the code as `*_test.go`; follow that pattern for new service, utility, and integration coverage. Frontend has no committed app test suite yet, so contributors should treat `npm run lint`, `npm run build`, and manual UI checks as the minimum gate. Run both verification and build checks before opening a PR when touching both applications.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit prefixes such as `feat:`, `fix:`, and `chore:`. Keep subjects imperative and scoped to the behavior changed. PRs should include a short summary, affected areas (`backend`, `frontend`, `deploy`, or `docs`), validation performed, and any schema or environment changes. Add screenshots for UI work and note new env keys such as `VITE_API_BASE_URL`, `JWT_SECRET`, or database settings.

## Security & Configuration Tips
Use `.env.production.template` and `docs/environment-contract.md` as the baseline for local or deployment config. Never commit real secrets, production hostnames, or database credentials.

## Git Push 认证配置

本机已安装 GitHub Desktop，使用 Windows 凭据管理器（Credential Manager）进行 GitHub 认证，**无需手动输入用户名密码**。

每次推送前执行以下命令激活凭据管理器：

```bash
git config --global credential.helper manager
```

验证是否生效：

```bash
git push
```

如果推送失败（`fatal: Authentication failed`），说明凭据管理器未生效，执行：

```bash
git config --global credential.helper manager-core
git push
```

> **注意**：若上述命令仍失败，则使用 `gh auth login`（需先安装 GitHub CLI）或提示用户手动推送。

## 确认优先原则
当任务需求存在不确定、上下文不足、目标不明确或可能产生误解时，必须先向用户确认关键细节；在未确认前，不要自行假设、猜测或执行可能影响结果的操作。

## Dictionary Update Workflow
Before changing any dropdown, cascader, or dictionary-backed field, verify the existing dictionary configuration first.

- Check frontend dictionary constants in `ai-hms-frontend/src/services/dictApi.ts`.
- Check backend dictionary mappings and fallback items in `ai-hms-backend/internal/services/dict_service.go`.
- For patient basic information fields, confirm these mappings before coding:
  - `ID_TYPE` maps to legacy `IDType`.
  - `VISIT_CATEGORY` maps to legacy `HospPatientType`.
  - `INSURANCE_TYPE` maps to legacy `ExpenseType`.
- If the local system is running, query `/api/v1/dict/items/{typeCode}?isEnabled=true` with a valid token and tell the user whether the requested items already exist.
- Preserve frontend API contracts and prefer backend dictionary compatibility or fallback updates over hardcoded page-only changes.
- If a requested dictionary item cannot be mapped to a legacy dictionary type or its business meaning is unclear, record it in `docs/legacy-migration-pending-confirmation.md` before proceeding.

---

# Legacy Migration Rules for Hemodialysis System

## Goal
Migrate backend data access from the **new hemodialysis database** to the **legacy hemodialysis database** with **no front-end behavior change whenever possible**.

The frontend should remain stable in:
- API path
- response structure
- field names
- page behavior
- interaction flow

## Required Schema References
Use the following files as the source of truth during migration:

- `新数据库表结构.md`  
  Current database schema used by the project.

- `老血透数据库表结构-合并版.md`  
  Target legacy database schema for migration.

Do not guess schema mappings. Always verify against these files.

## Core Migration Principles
1. **Frontend must remain unchanged whenever possible**
   - Do not change frontend API contracts unless explicitly required.
   - Keep response field names and response structure stable.
   - Perform compatibility work in the backend.

2. **Migrate one page or one module at a time**
   - Do not perform a full-system replacement in one step.
   - Scope each task to one page, one API, or one feature module.
   - Keep each migration step reviewable, testable, and reversible.

3. **Field mapping must be explicit**
   - Map legacy database fields to current API fields explicitly.
   - Confirm all key fields carefully, especially:
     - primary keys
     - patient identifiers
     - visit identifiers
     - status fields
     - datetime fields
     - foreign keys
     - business keys

4. **Prefer backend adaptation over frontend modification**
   - Rewrite SQL and service logic as needed.
   - Preserve the existing frontend-facing data contract.

## Required Workflow for Every Migration Task
For every page, API, or module migration, follow this order:

1. Read the current backend API and query logic.
2. Compare:
   - `新数据库表结构.md`
   - `老血透数据库表结构-合并版.md`
3. Identify table mapping and field mapping.
4. Rewrite SQL or repository logic to use the legacy database.
5. Keep the API response structure unchanged.
6. Validate locally, especially:
   - null handling
   - field type compatibility
   - datetime parsing and formatting
   - status field meaning
   - pagination and sorting behavior

## Never Do the Following
- Do not assume table mappings without checking schema documents.
- Do not assume field meaning from field names alone.
- Do not silently change frontend response fields.
- Do not delete old logic directly if rollback or comparison may still be needed.
- Do not mark a migration as complete if key fields are still uncertain.

## Uncertainty Handling Rules
If any of the following occurs:

- a table cannot be mapped directly
- a field meaning is unclear
- a multi-table merge relationship is unclear
- the legacy database does not contain a field used by the new database
- the business meaning differs between new and legacy schema

Then you must do all of the following:

1. Clearly mark the uncertainty in code comments if relevant.
2. Do not make a subjective or silent assumption.
3. Record the issue in:

- `docs/legacy-migration-pending-confirmation.md`

## Required Record Format for Uncertain Items
Each recorded item should include:

- page or module name
- API name or route
- related tables
- related fields
- current judgment
- unclear point
- pending confirmation item
- temporary handling, if any

## Code and Documentation Requirements
- Keep migration-related code changes clearly commented.
- For important SQL changes, state the mapping direction, for example:
  - new schema table -> legacy schema table
  - new field -> legacy field
- Keep the old logic traceable when necessary for rollback or comparison.
- Keep changes minimal and focused on the current migration scope.

## Completion Standard
A migration task is considered complete only when:

- backend queries have switched to the legacy database for the scoped page/module
- frontend behavior remains unchanged or acceptable
- key field mappings are verified
- local validation is completed
- uncertain items, if any, are recorded in `docs/legacy-migration-pending-confirmation.md`

## Priority Rule
When there is a conflict between general repository guidance and legacy migration work, follow this priority for migration-related tasks:

1. keep frontend behavior stable
2. use legacy schema as the backend data source
3. preserve response compatibility
4. avoid subjective schema assumptions
5. record uncertainties clearly

## Instruction to Codex
For any hemodialysis migration task, you must strictly follow the rules in this section.

Always verify schema mappings before coding.
Never assume unclear mappings.
Record resolved, manually confirmed items in `docs/legacy-migration-confirmed-completed.md`.
Record unresolved items in `docs/legacy-migration-pending-confirmation.md`.
