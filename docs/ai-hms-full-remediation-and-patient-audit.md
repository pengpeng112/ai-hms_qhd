# AI-HMS 完整修复与患者管理老库核查计划

## TL;DR

> **Quick Summary**: 本计划覆盖交接审查中发现的 P0/P1/P2 风险，并新增对患者管理九个模块的真实 PostgreSQL 老库数据核查。执行目标是先消除数据安全/治疗流程风险，再补齐接口口径、回显一致性、认证安全、自动化 QA 与待确认记录。
>
> **Deliverables**:
> - 修复字典配置读写口径、透析执行串患者、治疗时间覆盖、透后字段回显等 P0 问题
> - 按实际 PostgreSQL 返回内容核查患者管理九个模块
> - 将无法确认字段/接口写回待确认记录文档
> - 建立最小可持续 QA：后端 Go 测试、字典/透析/患者管理 smoke、前端 E2E 基础
>
> **Estimated Effort**: XL
> **Parallel Execution**: YES - 5 implementation waves + final verification
> **Critical Path**: DB 只读审计基线 → P0 修复 → 患者管理模块核查 → P1 稳定化 → 最终 QA

---

## Context

### Original Request
用户提供 `docs/codex-handoff-summary-2026-04-24.md`，要求审查当前程序是否还有需要改进和完善的地方；后续明确要求生成完整修复计划，覆盖 P0/P1/P2，按波次并行执行。

### New Database-Driven Audit Scope
用户提供本机可连接的 PostgreSQL 老库信息：

- 数据库类型：PostgreSQL
- Host：`10.20.1.153`
- Port：`5432`
- Database：`dialysis`
- User：`postgres`
- Password：由用户提供；**不得写入代码、文档或提交历史**

执行时统一使用环境变量，不在计划或代码中明文保存密码：

```powershell
$env:AI_HMS_AUDIT_DB_HOST="10.20.1.153"
$env:AI_HMS_AUDIT_DB_PORT="5432"
$env:AI_HMS_AUDIT_DB_NAME="dialysis"
$env:AI_HMS_AUDIT_DB_USER="postgres"
$env:PGPASSWORD="<由用户在本机临时设置，不提交>"
```

### Interview / Review Summary
**Key Findings**:
- 字典配置：页面读老库 `CodeDictionary_CodeDictionarys`，但增删改启停主要写新表或因 synthetic ID 失败，存在误维护风险。
- 透析执行：切换患者可能短暂显示上一患者数据；透中/透后操作可能覆盖治疗开始/结束时间；透后字段无法完整回显。
- 认证：代码中存在交接文档未提及的内置管理员、`DEFAULT_PASSWORD` 回退口令、JWT `tenant_id > 0` 强校验。
- QA：前端无测试/E2E，后端缺真实老库 PostgreSQL 集成测试，CI/verify 门禁不足。
- 患者管理：需要基于真实老库返回内容逐一核查全息视图、基本信息、方案、病史、长期医嘱、检验、血管通路、治疗历史、月度评估。

### Metis Review
Metis 调用超时，已按以下保守原则自动补齐规划护栏：
- 不明文保存数据库密码。
- 所有 DB 核查先只读抽样，再决定修复。
- 字典写老库口径、认证安全口径、透后提交事务口径均作为需明确落地的高风险决策点。
- 无法确认字段必须写回待确认记录，不允许主观猜测。

---

## Work Objectives

### Core Objective
把当前“老库迁移 + 透析执行重构 + 字典配置改造 + 患者管理核查”的不确定状态收敛为可验证、可回归、可交接的稳定版本。

### Concrete Deliverables
- 修复 P0：字典读写口径、患者切换隔离、治疗时间语义、透后字段回显。
- 完成患者管理九模块 DB 实际返回核查。
- 修复 P1：认证安全文档/开关、错误态区分、透后提交原子性、字典 Type 统一、机构过滤口径。
- 清理 P2：占位功能、固定错误文案、环境模板编码、待确认文档。
- 新增自动化 QA 基线。

### Definition of Done
- [ ] P0 问题均有代码修复、自动化或代理可执行 QA 证据。
- [ ] 患者管理九模块每项都有“实际 DB 数据 → API 返回 → 前端展示/回显”的核查结论。
- [ ] 无法确认字段写入待确认记录。
- [ ] 前后端构建、后端测试、关键 smoke/E2E 通过。
- [ ] 无数据库密码进入仓库。

### Must Have
- 使用真实 PostgreSQL 数据做只读核查。
- 对 P0/P1/P2 分层修复，不跳过 P0。
- 所有 QA 可由执行代理完成，无需人工手工点击确认。
- 遇到字段不确定：记录，不猜测。

### Must NOT Have
- 不提交数据库密码、连接串明文、生产账号口令。
- 不直接覆盖当前大量未提交改动；执行前必须保存/确认工作区状态。
- 不把演示占位数据写入真实治疗记录。
- 不以“构建通过”替代真实接口/回显核查。

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** - 所有验收项必须能由执行代理运行命令、调用 API、查询 DB、使用 Playwright 或读取文件验证。

### Test Decision
- **Infrastructure exists**: PARTIAL
- **Automated tests**: Tests-after + targeted smoke/E2E
- **Framework**: Go `go test`; frontend build/lint; Playwright to be introduced for critical UI flows
- **DB Integration**: PostgreSQL read-only audit using environment variables

### QA Policy
- Evidence path: `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`
- API evidence: response JSON + status code
- DB evidence: sanitized query output, no passwords
- UI evidence: Playwright screenshots/traces
- Unknown fields: append to documented uncertainty checklist

---

## Execution Strategy

### Parallel Execution Waves

```text
Wave 1 - Safety baseline + audit scaffolding:
├── T1: Worktree safety + secret guardrails
├── T2: PostgreSQL read-only audit harness
├── T3: QA/CI minimum gate design
├── T4: Dialysis execution state isolation test harness
├── T5: Dictionary contract audit matrix

Wave 2 - P0 remediation:
├── T6: Fix dictionary legacy read/write contract
├── T7: Fix dialysis patient context isolation
├── T8: Fix treatment start/end time semantics
├── T9: Add complete post-assessment echo contract

Wave 3 - Patient management actual-data audit:
├── T10: Holographic view + basic info audit
├── T11: Treatment plan management audit
├── T12: Clinical history audit
├── T13: Long-term plan/order audit
├── T14: Lab/imaging report audit
├── T15: Vascular access assessment audit
├── T16: Treatment detail history audit
├── T17: Monthly assessment summary audit

Wave 4 - P1 stabilization:
├── T18: Auth hardening and documentation alignment
├── T19: Error-state separation across treatment APIs
├── T20: Atomic post-assessment submit
├── T21: Dictionary type unification + OrganId decision

Wave 5 - P2 cleanup + documentation:
├── T22: Remove/label placeholders and non-persisted UI
├── T23: Improve frontend/backend error messages
├── T24: Fix environment template encoding and docs
├── T25: Update uncertain-field records

Wave FINAL:
├── F1: Plan compliance audit
├── F2: Code quality review
├── F3: Real QA replay
└── F4: Scope fidelity check
```

### Dependency Matrix
- T1: blocks all implementation tasks
- T2: blocks T10-T17, supports T6/T21
- T3: blocks final QA and CI integration
- T4: supports T7-T9/T19/T20
- T5: supports T6/T21
- T6: depends T1/T5; blocks T21 and dictionary QA
- T7: depends T1/T4; blocks treatment E2E
- T8: depends T1/T4; blocks T20 and treatment E2E
- T9: depends T1/T4; blocks T20 and treatment E2E
- T10-T17: depend T1/T2; can run in parallel
- T18: depends T1
- T19: depends T7/T9
- T20: depends T8/T9/T19
- T21: depends T5/T6/T2
- T22-T25: depend related findings from T10-T21

### Agent Dispatch Summary
- Wave 1: 5 agents, mostly `quick` / `unspecified-high`
- Wave 2: 4 agents, `deep` for treatment semantics, `unspecified-high` for dictionary
- Wave 3: 8 agents, one per patient-management module, `deep`
- Wave 4: 4 agents, `deep` / `unspecified-high`
- Wave 5: 4 agents, `quick` / `writing`
- Final: 4 review agents

---

## TODOs

- [ ] 1. Worktree safety + secret guardrails

  **What to do**:
  - Capture `git status --short` before edits.
  - Add execution guidance: no force reset, no password committed, no `.env` secrets staged.
  - Ensure DB password is read only from local environment variable `PGPASSWORD`.

  **Must NOT do**:
  - Do not commit `.env`, password, raw connection string, or production credentials.

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small safety/config audit task.
  - **Skills**: [`git-master`]
    - `git-master`: needed for safe git status/diff handling.
  - **Skills Evaluated but Omitted**: `fullstack-dev` not needed; no implementation.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: All implementation tasks
  - **Blocked By**: None

  **References**:
  - `docs/codex-handoff-summary-2026-04-24.md:142-146` - warns large uncommitted changes must not be overwritten.
  - `.env.production.template` - known credential/documentation area needing careful handling.
  - `ai-hms-backend/.env.example` - compare safe env documentation.

  **Acceptance Criteria**:
  - [ ] Evidence file contains git status summary with no secrets printed.
  - [ ] Evidence file confirms no password-containing files are staged.

  **QA Scenarios**:
  ```text
  Scenario: Secret guardrail check
    Tool: Bash
    Preconditions: Repository checked out; no implementation started.
    Steps:
      1. Run `git status --short`.
      2. Run a staged/untracked filename scan for `.env`, `password`, `PGPASSWORD`, `admin@123` without printing file contents.
      3. Save sanitized output.
    Expected Result: No secret file is staged; if secret-like file appears, task blocks implementation.
    Evidence: .sisyphus/evidence/task-1-secret-guardrail.txt
  ```

  **Commit**: NO

- [ ] 2. PostgreSQL read-only audit harness

  **What to do**:
  - Create or document a reusable read-only DB audit command/script pattern using environment variables.
  - Verify connectivity to `10.20.1.153:5432/dialysis` without writing data.
  - Produce table/column discovery outputs for patient, treatment, orders, labs, vascular access, monthly evaluation, dictionary tables.

  **Must NOT do**:
  - Do not store password in files.
  - Do not run INSERT/UPDATE/DELETE/DDL.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Requires DB connectivity and careful read-only evidence capture.
  - **Skills**: []
  - **Skills Evaluated but Omitted**: `minimax-xlsx` not needed unless exporting spreadsheet is requested.

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T10-T17, supports T6/T21
  - **Blocked By**: T1 recommended

  **References**:
  - `docs/legacy-db-schema.md` - structured legacy DB schema summary.
  - `docs/migration-field-map.md` - new-to-old field mapping.
  - `老血透数据库表结构-合并版.md` - original old DB schema reference.

  **Acceptance Criteria**:
  - [ ] Connectivity proven using `SELECT current_database(), current_user`.
  - [ ] Read-only role/transaction guard used where possible.
  - [ ] Discovered table list saved without credentials.

  **QA Scenarios**:
  ```text
  Scenario: Read-only DB connectivity
    Tool: Bash / psql
    Preconditions: User has set PGPASSWORD locally; network can reach DB host.
    Steps:
      1. Run `psql -h $AI_HMS_AUDIT_DB_HOST -p $AI_HMS_AUDIT_DB_PORT -U $AI_HMS_AUDIT_DB_USER -d $AI_HMS_AUDIT_DB_NAME -c "BEGIN READ ONLY; SELECT current_database(), current_user; ROLLBACK;"`.
      2. Confirm output database is `dialysis` and user is `postgres`.
      3. Save sanitized command output.
    Expected Result: Query succeeds; no write statements executed.
    Evidence: .sisyphus/evidence/task-2-db-connectivity.txt
  ```

  **Commit**: YES
  - Message: `test(audit): add readonly database audit harness`

- [ ] 3. QA/CI minimum gate design

  **What to do**:
  - Move/define effective root-level CI or documented local gate.
  - Include frontend build/lint, backend `go vet`, `go build`, `go test ./...`.
  - Add placeholders or scripts for dictionary, dialysis execution, patient audit smoke.

  **Must NOT do**:
  - Do not keep `|| true` or `continue-on-error` for required gates.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: Cross-project CI and verification design.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: Final verification
  - **Blocked By**: T1

  **References**:
  - `ai-hms-backend/scripts/verify.sh` - currently does not run `go test`.
  - `ai-hms-backend/scripts/smoke_test.sh` - existing smoke style to extend.
  - `ai-hms-frontend/package.json` - frontend build/lint commands.

  **Acceptance Criteria**:
  - [ ] A single documented gate command exists.
  - [ ] Gate fails on backend test/build/lint failures.

  **QA Scenarios**:
  ```text
  Scenario: Local gate command executes
    Tool: Bash
    Preconditions: Dependencies installed.
    Steps:
      1. Run the documented local gate command.
      2. Capture exit code and output summary.
    Expected Result: Exit code 0 on clean tree, non-zero if any required check fails.
    Evidence: .sisyphus/evidence/task-3-local-gate.txt
  ```

  **Commit**: YES
  - Message: `ci: add effective verification gate`

- [ ] 4. Dialysis execution state isolation test harness

  **What to do**:
  - Add Playwright or agent-executable UI scenario to reproduce fast patient switching.
  - Define selectors/data for two patients with existing treatment records.
  - Capture before-fix failing evidence if possible, then use for T7 verification.

  **Must NOT do**:
  - Do not rely on manual visual confirmation only.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Browser UI state and loading behavior.
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T7-T9/T19/T20
  - **Blocked By**: T1

  **References**:
  - `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx` - selected patient/current treatment state owner.
  - `ai-hms-frontend/src/pages/dialysis-processing/execution/PostAssessment.tsx` - tab where stale treatment is visible.

  **Acceptance Criteria**:
  - [ ] Scenario can switch Patient A → Patient B under delayed API.
  - [ ] Evidence identifies whether stale data appears.

  **QA Scenarios**:
  ```text
  Scenario: Fast patient switch does not show stale data
    Tool: Playwright
    Preconditions: Two test patients exist with distinct treatment notes/monitoring values.
    Steps:
      1. Navigate to `/dialysis-processing`.
      2. Select Patient A and open `透后评估`.
      3. Select Patient B while treatment API is delayed by 2s.
      4. Assert Patient A-specific text/value is not visible after Patient B is selected.
    Expected Result: Loading/empty state appears until Patient B data loads; no stale Patient A values are visible.
    Evidence: .sisyphus/evidence/task-4-patient-switch.png
  ```

  **Commit**: YES
  - Message: `test(dialysis): add patient switch state regression scenario`

- [ ] 5. Dictionary contract audit matrix

  **What to do**:
  - Build a matrix for dictionary type/item behavior: legacy read, unified code, CRUD target, toggle target, tree behavior, OrganId handling.
  - Include examples: `DialysisMethod`, `AccessType`, `OutComeType`, `OutComeReason`, `Dialysate`.

  **Must NOT do**:
  - Do not change dictionary code before contract matrix is written.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Requires code, DB, API, and documentation alignment.
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1
  - **Blocks**: T6/T21
  - **Blocked By**: T1/T2

  **References**:
  - `ai-hms-backend/internal/services/dict_service.go` - dictionary service read/write behavior.
  - `ai-hms-frontend/src/pages/DictConfig.tsx` - frontend category and CRUD behavior.
  - `docs/dictionary-type-mapping-dev.md` - intended Type mapping.

  **Acceptance Criteria**:
  - [ ] Matrix lists each tested type and current behavior.
  - [ ] Matrix marks expected behavior for T6/T21.

  **QA Scenarios**:
  ```text
  Scenario: Dictionary CRUD target audit
    Tool: Bash / curl / psql
    Preconditions: Auth token available; DB read-only access available.
    Steps:
      1. Call `GET /api/v1/dict/types`.
      2. Call `GET /api/v1/dict/items/DialysisMethod` and one unified equivalent if present.
      3. Query matching legacy rows read-only.
      4. Save type/item ID formats and table source inference.
    Expected Result: Evidence clearly states whether item IDs are legacy synthetic or new table IDs.
    Evidence: .sisyphus/evidence/task-5-dict-contract.json
  ```

  **Commit**: YES
  - Message: `docs(dict): document dictionary contract matrix`

- [ ] 6. Fix dictionary legacy read/write contract

  **What to do**:
  - Choose and implement a safe dictionary write contract:
    1. legacy dictionaries read-only with UI disabling edit/delete/toggle, OR
    2. backend writes legacy table for legacy-sourced items, OR
    3. explicit merged overlay with visible source labels.
  - Default plan recommendation: **legacy-sourced items must display source and disallow unsafe write until write mapping is proven**.
  - Make create/update/delete/toggle behavior explicit in API responses.

  **Must NOT do**:
  - Do not silently write new `dict_items` when user is editing a legacy-sourced type unless UI labels it as local overlay.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Data integrity and backend/frontend contract change.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T21
  - **Blocked By**: T5

  **References**:
  - `ai-hms-backend/internal/services/dict_service.go` - implement source-aware item IDs/permissions.
  - `ai-hms-frontend/src/pages/DictConfig.tsx` - disable/label unsafe operations.
  - `docs/dictionary-type-mapping-dev.md` - mapping rationale.

  **Acceptance Criteria**:
  - [ ] Legacy-sourced items cannot be falsely edited as new-table items.
  - [ ] UI clearly shows source or disables unsupported actions.
  - [ ] API returns predictable error for unsupported legacy writes.

  **QA Scenarios**:
  ```text
  Scenario: Legacy item edit is safe
    Tool: Playwright + curl
    Preconditions: A legacy dictionary item exists for `DialysisMethod`.
    Steps:
      1. Open dictionary config and select the type.
      2. Attempt edit/toggle/delete on a legacy-sourced row.
      3. Assert UI either blocks action with exact text `老库字典暂不支持直接维护` or API updates the actual legacy row according to chosen contract.
    Expected Result: No silent write to wrong table; behavior is explicit.
    Evidence: .sisyphus/evidence/task-6-legacy-dict-safe.png
  ```

  **Commit**: YES
  - Message: `fix(dict): make legacy dictionary maintenance contract explicit`

- [ ] 7. Fix dialysis patient context isolation

  **What to do**:
  - Clear `currentTreatment` immediately when patient/date changes.
  - Propagate loading/empty/error state to every execution tab.
  - Prevent child tabs from rendering stale treatment data for a newly selected patient.

  **Must NOT do**:
  - Do not leave only `PreAssessment` loading-aware.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: UI state and user safety.
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: treatment E2E
  - **Blocked By**: T4

  **References**:
  - `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx` - parent state.
  - `execution/*.tsx` - child tabs receiving treatment props.

  **Acceptance Criteria**:
  - [ ] Patient switch clears stale treatment within same render cycle.
  - [ ] All tabs show loading/empty state until current patient data loads.

  **QA Scenarios**:
  ```text
  Scenario: No stale treatment after switching patient
    Tool: Playwright
    Preconditions: Patient A and B have visibly different treatment values.
    Steps:
      1. Select Patient A, open `透中监测`.
      2. Switch to Patient B under delayed API.
      3. Assert Patient A values disappear immediately and loading text appears.
      4. Assert Patient B values appear only after API completes.
    Expected Result: No cross-patient stale display.
    Evidence: .sisyphus/evidence/task-7-no-stale-patient.png
  ```

  **Commit**: YES
  - Message: `fix(dialysis): isolate treatment state on patient switch`

- [ ] 8. Fix treatment start/end time semantics

  **What to do**:
  - Prevent status updates from blindly overwriting `StartTime`/`EndTime`.
  - Preserve user-entered post-assessment end time.
  - Define exact semantics: start time set only at treatment start/sign-in; end time set from post assessment or explicit finish command.

  **Must NOT do**:
  - Do not call `time.Now()` in status update if a trusted time already exists.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Core clinical time data integrity.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T20
  - **Blocked By**: T4

  **References**:
  - `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx:ensureTodayTreatment` - caller behavior.
  - `ai-hms-backend/internal/services/treatment_service.go` - status/time update logic.
  - `docs/treatment-execution-legacy-dev-record-2026-04-21.md` - treatment execution migration record.

  **Acceptance Criteria**:
  - [ ] Updating during parameters does not modify treatment start time.
  - [ ] Post submit preserves user-entered end time.
  - [ ] Backend tests cover both behaviors.

  **QA Scenarios**:
  ```text
  Scenario: During monitoring does not reset start time
    Tool: Bash / curl
    Preconditions: Treatment exists with known `startTime`.
    Steps:
      1. Capture treatment detail JSON.
      2. Create or update during monitoring record.
      3. Fetch treatment detail again.
      4. Assert `startTime` equals original value.
    Expected Result: Start time unchanged.
    Evidence: .sisyphus/evidence/task-8-start-time-preserved.json

  Scenario: Post submit preserves filled end time
    Tool: Bash / curl
    Preconditions: Treatment exists and post form can be submitted.
    Steps:
      1. Submit post assessment with `endTime` = `2026-04-24T15:30:00+08:00`.
      2. Submit finish status.
      3. Fetch detail.
      4. Assert `endTime` remains `2026-04-24T15:30:00+08:00`, not current time.
    Expected Result: User-entered end time preserved.
    Evidence: .sisyphus/evidence/task-8-end-time-preserved.json
  ```

  **Commit**: YES
  - Message: `fix(treatment): preserve clinical start and end times`

- [ ] 9. Add complete post-assessment echo contract

  **What to do**:
  - Add backend response shape for post-assessment/afterSigns snapshot.
  - Include actual UF, substitute volume, post weight, extra weight, intake, pressure site, vitals, symptoms, notes.
  - Update frontend `RestTreatment` and `PostAssessment.mapTreatmentToForm`.

  **Must NOT do**:
  - Do not infer unknown legacy fields without documenting uncertainty.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Backend/frontend data contract.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: T20
  - **Blocked By**: T4

  **References**:
  - `ai-hms-frontend/src/pages/dialysis-processing/execution/PostAssessment.tsx` - form fields and mapper.
  - `ai-hms-frontend/src/services/restClient.ts` - `RestTreatment` contract.
  - `ai-hms-backend/internal/services/treatment_service.go` - `buildTreatmentRealtimeResponse`.

  **Acceptance Criteria**:
  - [ ] Save → refresh preserves all supported post-assessment fields.
  - [ ] Unsupported/uncertain fields are documented, not silently dropped.

  **QA Scenarios**:
  ```text
  Scenario: Post fields persist and echo
    Tool: Playwright
    Preconditions: Treatment exists for selected patient.
    Steps:
      1. Fill post assessment: actual UF `2.3`, post weight `61.5`, pressure site `左上肢`, heart rate `82`.
      2. Save.
      3. Reload page and reopen `透后评估`.
      4. Assert each exact value is visible in its field.
    Expected Result: All supported fields round-trip.
    Evidence: .sisyphus/evidence/task-9-post-echo.png
  ```

  **Commit**: YES
  - Message: `fix(dialysis): return complete post assessment snapshot`

- [ ] 10. Patient management audit: 全息视图档案 + 基本信息档案

  **What to do**:
  - Query actual old DB patient/person/register tables for representative patients.
  - Compare API response and frontend display for holographic view and basic info.
  - Record missing/mismatched fields.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Real DB → API → UI field mapping audit.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `docs/basic-info-legacy-gap-analysis.md` - existing gap analysis.
  - `docs/migration-field-map.md` - patient field mappings.
  - Patient management frontend/backend files to be located by executor.

  **Acceptance Criteria**:
  - [ ] At least 3 patient samples audited.
  - [ ] Each displayed field has source table/column or uncertainty record.

  **QA Scenarios**:
  ```text
  Scenario: Basic info field audit
    Tool: psql + curl
    Preconditions: Auth token and read-only DB access available.
    Steps:
      1. Select 3 patient IDs with complete demographic data from DB.
      2. Call patient detail/basic info APIs for each.
      3. Compare name, gender, age/birthday, phone, diagnosis, insurance/id fields.
    Expected Result: Mismatches listed with DB value, API value, UI field name.
    Evidence: .sisyphus/evidence/task-10-basic-info-audit.md
  ```

  **Commit**: YES
  - Message: `test(patient): audit holographic and basic info mappings`

- [ ] 11. Patient management audit: 治疗方案管理

  **What to do**:
  - Compare old DB dialysis scheme/prescription-like tables with treatment plan management API/UI.
  - Verify dialysis mode, anticoagulant, dialyzer, dialysate, frequency, duration, UF target.
  - Record uncertain old fields.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `docs/treatment-execution-legacy-dev-record-2026-04-21.md`
  - `docs/migration-field-map.md`
  - `老血透数据库表结构-合并版.md`

  **Acceptance Criteria**:
  - [ ] Actual returned plan fields match old DB samples or are recorded as uncertain.

  **QA Scenarios**:
  ```text
  Scenario: Treatment plan actual-data comparison
    Tool: psql + curl
    Preconditions: 3 patients with active plans exist.
    Steps:
      1. Query old DB plan/prescription rows.
      2. Call treatment plan API.
      3. Compare dialysis mode, anticoagulant, frequency, duration, dialysate, materials.
    Expected Result: Every mismatch has source value and API value.
    Evidence: .sisyphus/evidence/task-11-treatment-plan-audit.md
  ```

  **Commit**: YES

- [ ] 12. Patient management audit: 临床病史档案

  **What to do**:
  - Audit diagnosis, primary disease, comorbidities, allergies, infectious markers, past history.
  - Compare old DB content with API/UI.
  - Record code/value dictionary mismatches.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `docs/patient-management-dictionary-uncertain-2026-04-23.md`
  - `docs/legacy-migration-uncertain-field-checklist.md`

  **Acceptance Criteria**:
  - [ ] Clinical-history fields are mapped, fixed, or recorded uncertain.

  **QA Scenarios**:
  ```text
  Scenario: Clinical history mapping audit
    Tool: psql + curl
    Preconditions: Sample patients have diagnosis/history data.
    Steps:
      1. Query old DB clinical/history rows.
      2. Fetch patient clinical history API.
      3. Compare diagnosis, allergy, infection, past history fields.
    Expected Result: No silent missing high-value fields; uncertain fields documented.
    Evidence: .sisyphus/evidence/task-12-clinical-history-audit.md
  ```

  **Commit**: YES

- [ ] 13. Patient management audit: 长期方案/医嘱

  **What to do**:
  - Audit long-term plans/orders against old DB order/template/order execution tables.
  - Verify medicine/material/order text, dose, frequency, start/end, status.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `docs/legacy-migration-session-summary-2026-04-21.md`
  - `docs/migration-field-map.md`

  **Acceptance Criteria**:
  - [ ] Long-term orders display correct old DB content for sampled patients.

  **QA Scenarios**:
  ```text
  Scenario: Long-term order audit
    Tool: psql + curl
    Preconditions: Patients with long-term orders exist.
    Steps:
      1. Query old DB long-term order rows.
      2. Call long-term order API.
      3. Compare order name, dose, frequency, status, dates.
    Expected Result: API/UI aligns with DB or uncertainty is recorded.
    Evidence: .sisyphus/evidence/task-13-long-term-orders-audit.md
  ```

  **Commit**: YES

- [ ] 14. Patient management audit: 检验检查报告

  **What to do**:
  - Audit lab/imaging report APIs and UI against actual returned DB content.
  - Verify report time, item name, result, unit, reference range, abnormal flag.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `ai-hms-backend/internal/services/lab_report_service_test.go` - existing test style.
  - `docs/migration-field-map.md`

  **Acceptance Criteria**:
  - [ ] Sample lab reports round-trip from DB to API/UI.

  **QA Scenarios**:
  ```text
  Scenario: Lab report actual-data audit
    Tool: psql + curl
    Preconditions: Patient has lab report rows.
    Steps:
      1. Query latest lab records from old DB.
      2. Call lab report API.
      3. Compare item/result/unit/reference/abnormal flag.
    Expected Result: Values match or field gap is documented.
    Evidence: .sisyphus/evidence/task-14-lab-report-audit.md
  ```

  **Commit**: YES

- [ ] 15. Patient management audit: 血管通路评估

  **What to do**:
  - Audit vascular access type/site/status/assessment date/complications against old DB.
  - Verify dictionary display mapping.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `docs/patient-management-dictionary-uncertain-2026-04-23.md`
  - `docs/dictionary-type-mapping-dev.md`

  **Acceptance Criteria**:
  - [ ] Access type/site/status are correct for sampled patients.

  **QA Scenarios**:
  ```text
  Scenario: Vascular access audit
    Tool: psql + curl
    Preconditions: Patient has vascular access records.
    Steps:
      1. Query old DB access rows.
      2. Call vascular access API/UI endpoint.
      3. Compare access type, side/site, status, assessment fields.
    Expected Result: Dictionary display matches source codes.
    Evidence: .sisyphus/evidence/task-15-vascular-access-audit.md
  ```

  **Commit**: YES

- [ ] 16. Patient management audit: 治疗详情历史

  **What to do**:
  - Audit historical treatment list/detail against old DB treatment tables.
  - Verify date, status, prescription, before/during/after signs, summaries.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2/T8/T9 recommended

  **References**:
  - `docs/treatment-execution-legacy-dev-record-2026-04-21.md`
  - `ai-hms-backend/internal/services/treatment_service.go`

  **Acceptance Criteria**:
  - [ ] Historical treatment detail matches actual old DB rows.

  **QA Scenarios**:
  ```text
  Scenario: Historical treatment detail audit
    Tool: psql + curl
    Preconditions: Patient has multiple historical treatments.
    Steps:
      1. Query old DB treatment rows for 3 dates.
      2. Call treatment history/detail APIs.
      3. Compare status, times, prescription, monitoring, after assessment fields.
    Expected Result: Critical treatment history fields match.
    Evidence: .sisyphus/evidence/task-16-treatment-history-audit.md
  ```

  **Commit**: YES

- [ ] 17. Patient management audit: 月度评估小结

  **What to do**:
  - Determine whether monthly assessment summary has existing old DB source/API.
  - If source exists, audit fields; if not, record as confirmed gap.
  - Avoid inventing mappings.

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: T25
  - **Blocked By**: T2

  **References**:
  - `docs/codex-handoff-summary-2026-04-24.md:125-126` - monthly summary still uncertain.
  - `docs/legacy-migration-uncertain-field-checklist.md`

  **Acceptance Criteria**:
  - [ ] Monthly summary source is confirmed or recorded as unresolved.
  - [ ] If API is missing, implementation task is clearly scoped.

  **QA Scenarios**:
  ```text
  Scenario: Monthly assessment source verification
    Tool: psql + code search + curl
    Preconditions: Read-only DB access available.
    Steps:
      1. Search old DB schema for monthly/evaluation summary candidate tables.
      2. Search backend API for monthly assessment endpoints.
      3. If both exist, compare sample rows; if not, write confirmed gap.
    Expected Result: No guessed mapping; source status is explicit.
    Evidence: .sisyphus/evidence/task-17-monthly-summary-audit.md
  ```

  **Commit**: YES

- [ ] 18. Auth hardening and documentation alignment

  **What to do**:
  - Review built-in admin and `DEFAULT_PASSWORD` fallback.
  - Make emergency auth behavior configurable, disabled by default in production if appropriate.
  - Align `.env.example`, `.env.production.template`, and deployment docs.

  **Must NOT do**:
  - Do not leave hardcoded production-capable credentials undocumented.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Security-sensitive backend behavior.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: final security review
  - **Blocked By**: T1

  **References**:
  - `ai-hms-backend/internal/services/auth_service.go` - built-in admin/default password logic.
  - `ai-hms-backend/internal/middleware/auth.go` - `tenant_id` requirement.
  - `.env.production.template`, `ai-hms-backend/.env.example` - documentation drift.

  **Acceptance Criteria**:
  - [ ] Production behavior is safe and documented.
  - [ ] Tests cover fallback enabled/disabled and missing tenant behavior.

  **QA Scenarios**:
  ```text
  Scenario: Emergency auth disabled in production mode
    Tool: Bash / go test
    Preconditions: Test config sets production-like env.
    Steps:
      1. Run auth service tests for built-in admin/default password disabled case.
      2. Attempt login with fallback password.
      3. Assert login fails unless explicit env flag enables fallback.
    Expected Result: No undocumented emergency login works by default.
    Evidence: .sisyphus/evidence/task-18-auth-hardening.txt
  ```

  **Commit**: YES
  - Message: `fix(auth): harden emergency login configuration`

- [ ] 19. Error-state separation across treatment APIs

  **What to do**:
  - Separate “no treatment record” from 401/403/404/500/network errors.
  - Frontend should show actionable messages and not silently create records after real API failure.

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: API client and UI error handling refinements.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: T20
  - **Blocked By**: T7/T9

  **References**:
  - `ai-hms-frontend/src/services/restClient.ts#getErrorMessage`
  - `ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx#loadTodayTreatment`
  - `ai-hms-backend/internal/api/v1/treatment_handler.go#GetByPatientAndDate`

  **Acceptance Criteria**:
  - [ ] 404 no-treatment is treated differently from 401/500.
  - [ ] UI messages include exact failure category.

  **QA Scenarios**:
  ```text
  Scenario: API failure does not look like empty treatment
    Tool: Playwright / mocked API
    Preconditions: Treatment endpoint can return 500.
    Steps:
      1. Force treatment detail endpoint to return 500.
      2. Open `/dialysis-processing`.
      3. Assert visible error text contains `治疗记录加载失败` and no auto-create action is triggered.
    Expected Result: Error is visible and not confused with empty treatment.
    Evidence: .sisyphus/evidence/task-19-error-state.png
  ```

  **Commit**: YES

- [ ] 20. Atomic post-assessment submit

  **What to do**:
  - Replace two-step frontend submit with backend transactional endpoint, or implement robust compensation/retry with explicit state.
  - Preferred: backend endpoint that saves after-signs and finishes treatment in one transaction.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Transactional clinical workflow.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4
  - **Blocks**: final treatment E2E
  - **Blocked By**: T8/T9/T19

  **References**:
  - `DialysisExecution.tsx#handleSubmitPostAssessment`
  - `treatment_handler.go` and `treatment_service.go` after-signs/status methods.

  **Acceptance Criteria**:
  - [ ] Save+finish is atomic or has explicit recoverable state.
  - [ ] Failure leaves no ambiguous “saved but not finished” state without visible warning.

  **QA Scenarios**:
  ```text
  Scenario: Post submit transaction failure rolls back or reports split state
    Tool: Bash / integration test
    Preconditions: Test can force status update failure.
    Steps:
      1. Submit post assessment with forced failure after after-signs save step.
      2. Fetch treatment detail.
      3. Assert transaction rolled back OR API returns explicit recoverable state marker.
    Expected Result: No silent split state.
    Evidence: .sisyphus/evidence/task-20-atomic-post-submit.json
  ```

  **Commit**: YES

- [ ] 21. Dictionary type unification + OrganId decision

  **What to do**:
  - Align backend `ListTypes`, frontend categories, and docs on unified vs raw legacy Type.
  - Decide and implement `OrganId` filtering if old DB contains multi-organization rows.
  - Make `OUTCOME` and `Dialysate` aggregation/splitting deterministic.

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: Cross-cutting dictionary contract and legacy data semantics.
  - **Skills**: [`fullstack-dev`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4
  - **Blocks**: final dictionary QA
  - **Blocked By**: T5/T6/T2

  **References**:
  - `docs/dictionary-type-mapping-dev.md`
  - `dict_service.go#legacyTypeToUnifiedCode`
  - `DictConfig.tsx#TYPE_CATEGORY_MAP`
  - old DB `CodeDictionary_CodeDictionarys.OrganId`

  **Acceptance Criteria**:
  - [ ] `/api/v1/dict/types` contract is documented and implemented consistently.
  - [ ] Unknown types still safely fall to “其他字典”.
  - [ ] OrganId behavior confirmed by DB evidence.

  **QA Scenarios**:
  ```text
  Scenario: Unified dictionary types are deterministic
    Tool: curl + psql
    Preconditions: Dictionary rows exist for DialysisMethod, Dialysate, OutComeType/Reason.
    Steps:
      1. Call `/api/v1/dict/types`.
      2. Assert expected unified/raw contract exactly matches documentation.
      3. Query DB for OrganId distribution and verify API filter/merge behavior.
    Expected Result: No duplicate/conflicting types or cross-organization ambiguity.
    Evidence: .sisyphus/evidence/task-21-dict-type-unification.json
  ```

  **Commit**: YES

- [ ] 22. Remove/label placeholders and non-persisted UI

  **What to do**:
  - Replace `新增症状 N` behavior with real dictionary/free-text workflow or disable with label.
  - Label HealthEducation as placeholder or implement source if available.
  - Disable or clarify `暂存`, `消毒登记`, unpersisted fields.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5
  - **Blocks**: final UI QA
  - **Blocked By**: T7/T9/T17 findings

  **References**:
  - `execution/PreAssessment.tsx` - placeholder symptom add.
  - `execution/HealthEducation.tsx` - placeholder module.
  - `execution/Verification.tsx` - static disinfection display.

  **Acceptance Criteria**:
  - [ ] No demo placeholder text can be saved as real clinical data.
  - [ ] Non-persisted UI elements are disabled or clearly labeled.

  **QA Scenarios**:
  ```text
  Scenario: Placeholder symptom cannot be saved silently
    Tool: Playwright
    Preconditions: Treatment page available.
    Steps:
      1. Open pre-assessment.
      2. Click symptom add.
      3. Assert UI requires real text/dictionary selection or shows disabled explanation.
      4. Save and verify no `新增症状` appears in API response.
    Expected Result: Demo placeholder not persisted.
    Evidence: .sisyphus/evidence/task-22-no-placeholder-symptom.png
  ```

  **Commit**: YES

- [ ] 23. Improve frontend/backend error messages

  **What to do**:
  - Use `getErrorMessage` consistently in dialysis/dictionary/patient pages.
  - Preserve backend validation details where safe.
  - Add user-actionable messages for auth, network, validation, conflict, not-found.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5
  - **Blocks**: final QA
  - **Blocked By**: T19

  **References**:
  - `ai-hms-frontend/src/services/restClient.ts#getErrorMessage`
  - dialysis execution components
  - `DictConfig.tsx`

  **Acceptance Criteria**:
  - [ ] Common 401/403/404/409/500 cases show distinct messages.

  **QA Scenarios**:
  ```text
  Scenario: Validation error displays backend message
    Tool: Playwright / mocked API
    Preconditions: API returns 400 with message `字段 X 不能为空`.
    Steps:
      1. Submit invalid form.
      2. Assert toast or inline message contains `字段 X 不能为空`.
    Expected Result: User sees specific backend validation text.
    Evidence: .sisyphus/evidence/task-23-error-message.png
  ```

  **Commit**: YES

- [ ] 24. Fix environment template encoding and docs

  **What to do**:
  - Repair `.env.production.template` encoding/readability.
  - Align `.env.example`, deployment docs, auth emergency switches, DB audit env docs.
  - Ensure no real secrets are inserted.

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5
  - **Blocks**: final docs check
  - **Blocked By**: T18

  **References**:
  - `.env.production.template`
  - `ai-hms-backend/.env.example`
  - `docs/environment-contract.md`
  - `docs/docker-migration-deploy-upgrade-guide.md`

  **Acceptance Criteria**:
  - [ ] Template text readable UTF-8.
  - [ ] Security-sensitive env vars documented without real values.

  **QA Scenarios**:
  ```text
  Scenario: Env docs contain placeholders not secrets
    Tool: Bash
    Preconditions: Docs/templates updated.
    Steps:
      1. Search templates/docs for actual DB password and emergency passwords.
      2. Assert only placeholders are present.
      3. Verify files render readable UTF-8 text.
    Expected Result: No secrets; no mojibake.
    Evidence: .sisyphus/evidence/task-24-env-docs.txt
  ```

  **Commit**: YES

- [ ] 25. Update uncertain-field records

  **What to do**:
  - Consolidate unknowns from T10-T17 and T21 into existing uncertainty docs.
  - For each unknown: module, UI field, API field, old DB candidate columns, observed sample values, required human decision.
  - Do not delete prior uncertain items unless proven resolved by evidence.

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 5 final doc consolidation
  - **Blocks**: final verification
  - **Blocked By**: T10-T17/T21

  **References**:
  - `docs/legacy-migration-uncertain-field-checklist.md`
  - `docs/patient-management-dictionary-uncertain-2026-04-23.md`
  - `docs/legacy-migration-session-summary-2026-04-21.md`

  **Acceptance Criteria**:
  - [ ] Every unresolved field discovered in audit appears in a checklist.
  - [ ] Every resolved field has evidence reference.

  **QA Scenarios**:
  ```text
  Scenario: Unknown field checklist completeness
    Tool: Bash / file read
    Preconditions: Audit evidence from T10-T17 exists.
    Steps:
      1. Extract all `UNCERTAIN` markers from audit evidence.
      2. Search uncertainty docs for matching module+field.
      3. Assert every uncertain item is represented.
    Expected Result: No unrecorded uncertainty remains.
    Evidence: .sisyphus/evidence/task-25-uncertain-checklist.txt
  ```

  **Commit**: YES
  - Message: `docs(migration): update patient management uncertainty records`

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit okay before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. Verify every Must Have and Must NOT Have. Check evidence files exist. Confirm no password/secret entered repo. Output: `VERDICT: APPROVE/REJECT`.

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run frontend lint/build, backend `go vet ./...`, `go build ./...`, `go test ./...`. Review changed files for AI slop, unsafe defaults, hidden credentials, broad catches, stale placeholders. Output: `VERDICT`.

- [ ] F3. **Real QA Replay** — `unspecified-high` (+ `playwright` if UI)
  Execute every QA scenario from T1-T25. Save DB/API/UI evidence under `.sisyphus/evidence/final-qa/`. Output pass/fail matrix.

- [ ] F4. **Scope Fidelity Check** — `deep`
  Compare implementation diff with this plan. Ensure all P0/P1/P2 items addressed, no unrelated scope creep, all uncertainties documented. Output: `VERDICT`.

---

## Commit Strategy

- T1-T5: baseline and audit scaffolding commits, no feature behavior changes unless required.
- T6-T9: P0 fixes, each commit scoped to one risk domain.
- T10-T17: patient audit evidence/docs/tests, grouped by module.
- T18-T21: P1 stabilization commits.
- T22-T25: P2 cleanup and documentation commits.
- Final: no commit until all verification passes; do not push unless user explicitly requests.

---

## Success Criteria

### Verification Commands
```bash
# Backend
cd ai-hms-backend
go vet ./...
go build ./...
go test ./...

# Frontend
cd ai-hms-frontend
npm run lint
npm run build

# Smoke/E2E commands to be added or documented by implementation tasks
./scripts/smoke_test.sh http://localhost:8080
```

### Final Checklist
- [ ] P0 dictionary maintenance risk removed or explicitly blocked/labeled.
- [ ] Dialysis patient switch cannot show stale patient data.
- [ ] Treatment start/end times are preserved correctly.
- [ ] Post-assessment fields round-trip.
- [ ] Patient management nine-module audit completed.
- [ ] Unknown fields written to uncertainty records.
- [ ] Auth emergency behavior safe and documented.
- [ ] CI/local gate covers frontend build/lint and backend test/build.
- [ ] No secrets committed.

---

## Execution Prompt for Other Model

将下面提示词交给执行模型使用。执行模型必须按本计划执行，不得跳过安全护栏。

```text
你现在负责执行 AI-HMS 修复计划：

计划文件：.sisyphus/plans/ai-hms-full-remediation-and-patient-audit.md

执行原则：
1. 严格按计划中的 Wave 1 → Wave 2 → Wave 3 → Wave 4 → Wave 5 → Final Verification 顺序推进。
2. 每个任务必须先阅读该任务的 References，再实施；不得凭空猜字段、表名或业务口径。
3. PostgreSQL 老库核查只允许使用只读查询。连接信息从本机环境变量读取：
   - AI_HMS_AUDIT_DB_HOST=10.20.1.153
   - AI_HMS_AUDIT_DB_PORT=5432
   - AI_HMS_AUDIT_DB_NAME=dialysis
   - AI_HMS_AUDIT_DB_USER=postgres
   - PGPASSWORD 由用户在本机临时设置，禁止写入任何文件、日志、提交、计划或文档。
4. 禁止提交或打印数据库密码、生产账号、JWT、连接串明文。
5. 当前仓库有大量未提交改动。开始前必须执行并保存 `git status --short` 证据，不得 reset/checkout 覆盖他人改动。
6. P0 优先：字典维护口径、患者切换串数据、治疗开始/结束时间覆盖、透后字段回显必须先修。
7. 患者管理九模块核查必须基于真实 DB 返回内容：全息视图档案、基本信息档案、治疗方案管理、临床病史档案、长期方案/医嘱、检验检查报告、血管通路评估、治疗详情历史、月度评估小结。
8. 对无法确定的字段或接口口径，不允许主观猜测；必须写回 `docs/legacy-migration-uncertain-field-checklist.md` 或相关待确认文档，包含模块、字段、候选表/列、实际样例值、需要人工确认的问题。
9. 每个任务完成后必须执行计划内 QA Scenario，并把证据保存到 `.sisyphus/evidence/` 对应路径。
10. 最终必须运行 Final Verification Wave：计划符合性、代码质量、真实 QA、范围一致性四项全部通过后，向用户汇报并等待用户明确同意。

请先读取完整计划文件，然后从 T1 开始执行。
```
