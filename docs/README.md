# Docs 文档索引

> 生成时间：2026-06-09 | 系统：ai-hms_qhd  
> 目的：为 AI 解析提供文档导航，说明每份文档的用途和有效状态。

---

## 根目录文档

| 文件 | 状态 | 用途 |
|------|------|------|
| `排班模块交接文档.md` | ★ 最新 | **核心交接文档**：数据库结构、API 清单、状态机、前端对照、环境变量、硬规则、已知问题 |
| `README.md` | ★ 最新 | 本索引文件 |
| `legacy-migration-uncertain-field-checklist.md` | 持续参考 | 老库字段映射不确定项清单，排班模块与老表对接时的字段语义参考 |

---

## `sql/` — SQL 脚本参考

| 文件 | 状态 | 用途 |
|------|------|------|
| `v2_merge_legacy.sql` | ★ 已执行 | **核心迁移脚本**：ALTER TABLE 添加 v2 新列 + 数据从旧表同步到 Schedule_* |
| `v1.3_v2_tables.sql` | 参考 | v1.3 阶段 v2 表 DDL 定义 |
| `v1.2_v2_tables.sql` | 参考 | v1.2 阶段 v2 表 DDL 定义 |
| `schedule_extension_tables.sql` | 参考 | 扩展表 (PatientShiftExt 等) DDL |
| `schedule_conflict_queue.sql` | 参考 | 冲突队列表 DDL |
| `schedule_patient_shift_unique_safety_net.sql` | 参考 | 唯一索引安全网 SQL |
| `schedule_performance_indexes.sql` | 参考 | 性能索引 SQL |
| `schedule_status60_template_audit.sql` | 参考 | Status=60 模板审计 SQL |

---

## `排班功能说明/` — 排班设计参考

### `透析排班-backend-v1.4/` — v1.4 参考后端代码 ★

完整的独立 Go 后端参考实现，包含以下模块：

| 路径 | 说明 |
|------|------|
| `cmd/server/main.go` | 入口 |
| `internal/api/api.go` + `api_admin.go` | API 路由 (含上机/下机/审计) |
| `internal/config/config.go` | 配置管理 |
| `internal/db/db.go` | 数据库连接 + 迁移 + 唯一索引 |
| `internal/model/models.go` | GORM 模型 (14 张表) |
| `internal/repo/repo.go` | 数据访问层 |
| `internal/sched/` | 排班引擎 (board/constants/engine/newpatient/util/week + tests) |
| `internal/seed/seed.go` | 演示数据 |
| `internal/service/` | 13 个业务服务 (admin/crrt/diff/lifecycle/makeup/ops/perturb/quality/schedule/template/treatment/weekview + integration_test) |
| `web/index.html` | 独立 Web UI (React 内嵌) |

**v1.4 与当前系统的主要差异**：
- 使用 `ScheduleDate` 列名 (当前用 `TreatmentTime`)
- 使用 `Schedule_Machine` 表 (当前用 `Schedule_Bed`)
- 使用 `*int64` 可空字段 (当前用 `int64 NOT NULL DEFAULT 0`)
- 有独立的 `StartTreatment/CompleteTreatment` 实现 (当前已补全)
- 有 `auditMiddleware` 请求审计 (当前由主系统中间件负责)
- 有 `GET /health` 系统探活 (当前有 `GET /schedule/health` 数据健康检查)

---

## 不在本文档索引中的目录/文件

以下目录/文件**已清理**，不再保留：

- `local-test/` — 本地测试报告（过时）
- `schedule-plan/` — 旧版排班开发计划（过时）
- `ui-mockups/` — UI 线框图（过时）
- `remediation-skills/` — 空目录
- `排班功能说明/透析排班-backend/` — v1.0 参考代码（已删除）
- `排班功能说明/透析排班-backend-v1.1(1)/` — v1.1 参考代码（已删除）
- `排班功能说明/透析排班-backend-v1.2/` — v1.2 参考代码（已删除）
- `排班功能说明/透析排班-backend-v1.3/` — v1.3 参考代码（已删除）
- `排班功能说明/*.docx` — Word 文档（已删除）
- `排班功能说明/*.zip` — 压缩包（已删除）
- 根目录 33 个过期 .md 文件（已删除）

---

## AI 解析时推荐阅读顺序

1. **`排班模块交接文档.md`** → 全面了解当前系统状态
2. **`sql/v2_merge_legacy.sql`** → 理解表结构和数据迁移
3. **`legacy-migration-uncertain-field-checklist.md`** → 理解字段映射不确定项
4. **`排班功能说明/透析排班-backend-v1.4/`** → 参考 v1.4 独立后端实现
5. 当前源代码：`ai-hms-backend/internal/smart_schedule/` → 实际运行代码
