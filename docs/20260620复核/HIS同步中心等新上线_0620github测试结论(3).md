# AI-HMS 新上线内容 测试与评估报告

- **测试对象**：`pengpeng112/ai-hms_qhd` 远程 `master` 最新提交 **`b31161f`**
- **对比基线**：`8432f10`（此前主干）
- **测试日期**：2026-06-20
- **测试方式**：隔离 worktree 全量构建 + 单元测试 + 前端构建（未触碰本地工作区 WIP）
- **工具链**：Go 1.26.4 / Node 22.20.0 / Windows 11
- **结论**：✅ **构建、单测、前端构建全绿，质量高，可上线**；⚠️ 上线前**必须手工建新表**（见 §5）

---

## 1. 远程变更概览

远程已**收敛到单一 `master` 分支**（已删除 `fix/legacy-ui-restore`、`feat/ui-uplift`、`opencode/brave-moon` 等历史分支）。

`8432f10 → b31161f` 共 **10 个提交**，**156 文件，+17257 / −10609**。

| 提交 | 类别 | 内容 |
|---|---|---|
| `b31161f` | **全新** | HIS 同步中心与数据库文档整理 |
| `258abc6` | 新增 | 处方确认人当班门禁（仅当班医生可签发当日处方） |
| `2e09709` | 修复 | 质控榜显示医生姓名（Organ_Employee → Identity_Users 回退） |
| `c3338e5` | 整合交付 | 融合质控评分与接班门禁 |
| `a7f1c2d` | 整合交付 | 增加 RNa 处方与电子签医生墙 |
| `80bc5f1` | 整合交付 | 融合建档草稿门禁与今日驾驶舱 |
| `453cf25` | 安全 | 8 项安全与稳定性整改（竞态/分页/缓存/鉴权/DoS 防护） |
| `7270b9b` | UI/测试 | 全模块 UI 优化 + 测试修复 + 患者/治疗/字典/工作台整改 |
| `deabc1e` | 清理 | 清理 10 个无用 SQL 文件 |
| `5330676` | P0/P1 | 治疗状态映射/封存软删/医嘱处方 TenantId/监控 UF 数据源/导航索引化/v2 响应归一 |

> 其中 `c3338e5 / a7f1c2d / 80bc5f1` 为团队整合的既有后端交付（已核验与交付源逐字节一致）；**本轮真正全新的工程是 HIS 同步中心**。

---

## 2. 测试结果（全绿）

| 测试项 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `go build ./...` | ✅ EXIT 0（含新增 `sijms/go-ora` Oracle 驱动下载与编译） |
| 后端单测 | `go test ./...` | ✅ 全部包 `ok`，**0 失败** |
| 前端依赖 | `npm ci` | ✅ 419 包安装成功 |
| 前端构建 | `npm run build`（tsc -b && vite build） | ✅ EXIT 0，5615 模块，约 12s |

**后端新增/相关测试包，均通过：**

```
ok  internal/clinicalsafety
ok  internal/database               （含 health_test：新表健康检查）
ok  internal/integrations/hdis
ok  internal/middleware
ok  internal/services               （含以下新测试）
        ├─ qc_score_test            质控赋分
        ├─ qc_service_test          质控服务
        ├─ sign_service_test        电子签
        ├─ prescription_oncall_test 当班门禁（patientWardToday）
        ├─ order_sign_status_test   医嘱签署态
        └─ patient_service_test     患者服务
ok  internal/smart_schedule/{api,sched,service}
ok  internal/models/types, internal/utils, internal/utils/idgen
```

> 备注：`internal/integrations/his_oracle`、`cmd/his-sync` 原仓库零单测（依赖真实 Oracle）。本轮我**补写了 HIS 同步落库管线的端到端冒烟**（见 §2.1），填补该子系统的本地 DB 一侧覆盖。

### 2.1 端到端冒烟（新增，本轮补做）

针对三条新流做了"服务层 + 真实 DB 引擎（SQLite）"的端到端驱动，新增测试文件 `his_sync_pipeline_e2e_test.go`（随报告交付），**5 个用例全部通过**：

| 用例 | 覆盖 | 结果 |
|---|---|---|
| `UpsertExamReport_Idempotent` | 检查报告幂等 upsert：重复 `external_report_id` 不新增行、覆盖标题/结论、回填 `synced_at` | ✅ |
| `UpsertExamReport_RequiresExternalID` | 缺 `external_report_id` 报错 | ✅ |
| `ResolveLegacyPatientID` | 已确认映射可解析；候选/不存在映射不被解析 | ✅ |
| `AutoMatchByIDNo` | 按身份证号匹配本地患者并落确认映射、可被后续解析；无匹配返回 (nil,nil) | ✅ |
| `JobRunLifecycle` | 跑批记录 CreateRun→FinishRun：状态/计数落库、`duration_ms` 正向、成功后推进 config 游标 + 落 `last_run_at` | ✅ |

**两点离线测试结论（重要）：**

1. **HIS Oracle 实连查询不可离线测**：`client.QueryExamReports / QueryExamItems / FindPatientIDByIDNo` 需真实 Oracle。已验证其**下游落库逻辑**（upsert/映射/游标）正确；Oracle 取数→映射这一环须接通生产库后用 `his-sync --once` 灰度核对（见 §6 Checklist）。
2. **当班门禁 `LegacySign` 全流程只能在 PostgreSQL 上做 E2E**：该方法**读**走 GORM 模型（标识符 `Plan_PatientPrescription`）、**写**走 `Table("\"Plan_PatientPrescription\"")`（带引号标识符）——二者**仅在 PostgreSQL 的标识符规则下指向同一张表**；SQLite 下会分裂为两张名字不同的表，故无法离线整跑（**非生产 bug，全库 PG-only 时正确**）。其拆解部件（`patientWardToday`、`ResolveDuty`、`SignService`）已分别由单测覆盖通过。上线请按 §6 用真账号在 staging 验门禁拒签/放行。

**未执行项**：HTTP 全栈起服冒烟。原因：`server` 二进制硬编码 PostgreSQL（`postgres.Open`）+ HIS 需 Oracle，离线环境不具备，且生产库（10.10.8.83）不可触碰。已用服务层 SQLite E2E 覆盖可覆盖部分。

---

## 3. 新功能逐项评审

### 3.1 HIS 同步中心（本轮核心，全新）

**范围**：从院内 HIS（Oracle）增量同步**检验/检查报告**到本系统，供临床查看。

新增构件：
- `cmd/his-sync/main.go` —— 独立 CLI 二进制（`--job his_exam_report --once`）
- `internal/integrations/his_oracle/`（client / exam_report_mapper / types）—— Oracle 只读客户端
- `internal/services/his_exam_report_sync_service.go`（385 行）—— 同步主逻辑
- `internal/services/sync_job_service.go`、`external_patient_mapping_service.go`
- `internal/api/v1/`：`his_exam_sync_handler`、`his_oracle_config_handler`、`sync_job_handler`、`sync_patient_handler`

> ⚠️ 归类更正（实测核对）：`indicator_mapping_handler.go`、`internal/config/clinical_indicator_mapping.json`、`internal/integrations/hdis/` **不属于本轮 HIS Oracle 同步**，它们是更早提交 `a7f1c2d feat: 增加RNa处方与电子签医生墙` 引入的 **HDIS（医保/电子病历）对接**资产。原报告误将其列入 HIS 同步中心构件，特此纠正。HIS Oracle 同步本身不依赖 HDIS 指标映射。

**设计评审（结论：稳健）**：
- ✅ **患者匹配以身份证号 ID_NO 为准**（`PrematchAll` 先批量建映射；批同步内 `ResolveLegacyPatientID` 失败时再 `AutoMatchByIDNo` 兜底），仅 `MatchStatusConfirmed` 的映射参与同步。
- ✅ **游标增量**：以 HIS `CreateDate` 为游标，记录 `CursorBefore/After`，避免全量重拉。
- ✅ **幂等 upsert**：键 `(patient_id, source_system, external_report_id)`，已存在则更新，杜绝重复。
- ✅ **检查项目"删后插"事务**：分批删除（100）、分批插入（200），保证项目集一致。
- ✅ **对 Oracle 只读**：仅 `QueryExamReports / QueryExamItems / FindPatientIDByIDNo`，不写 HIS。
- ✅ **容错**：单患者/单条失败计入 `Failed` 不中断整批；`Failed>0` 落 `partial` 状态；错误日志超 10 条截断。
- ✅ **运行留痕**：每次跑批写 `sync_job_runs`（created/updated/skipped/failed/fetched 计数）。

### 3.2 处方确认人当班门禁（`258abc6`）

`prescription_service.LegacySign`：签发当日处方时，按患者今日所在病区解析当班医生。
- ✅ **安全 fail-open**：`ResolveDuty` 出错只记日志、**不拦截**；仅当"存在当班医生且 `StaffId ≠ 签发人`"时拒签 → 不会因排班数据缺口卡死临床流程。
- ✅ 患者今日无治疗（`wardID==0`）时跳过门禁。
- ✅ 签发只落 `ConfirmTime/ConfirmUserId`（**不改执行态**，护士仍可后续执行），并写 `sign_record` 统一留痕；重复签幂等。

### 3.3 质控榜医生姓名回退（`2e09709`）

`Organ_Employee` 取不到姓名时回退 `Identity_Users`，避免榜单显示空/ID。低风险修复。

---

## 4. 与契约/规划的一致性

- 当班门禁 + 电子签：契合 **契约02 待签线**（处方/方案/小结电子签 + 上机前核对的轻量留痕起步）。
- 质控赋分：契合 **⑤质控赋分** 规划（自动喂数、按责任医生聚合）。
- HIS 同步：契合 **ACTRS/外部系统接入**思路（平级外挂、只读取数、身份证号对人）。
- HIS 同步与 ACTRS（胸片 CTR/ACTR）为**同一类外部对接框架**，后续 ACTR 可复用 `external_patient_mappings` 映射与 `sync_job_*` 调度骨架。

---

## 5. ⚠️ 上线硬前置：必须手工建新表

`AutoMigrate 已永久禁用`（legacy-DB 模式，见 `database/migrate.go`、`cmd/server/main.go`）。
**应用运行时不会自动建表**；新表必须在部署阶段执行 SQL 建好，否则相关端点（同步中心 / 电子签 / 当班门禁 / 质控）会返回 **500**。

**部署需执行** `docs/sql/deploy_new_tables.sql`，包含 9 张新表：

| 表 | 用途 |
|---|---|
| `exam_reports` | 检验/检查报告主表 |
| `exam_report_items` | 报告项目明细 |
| `external_patient_mappings` | 外部系统↔本地患者映射（HIS/ACTRS 共用） |
| `sync_job_configs` | 同步任务配置 |
| `sync_job_runs` | 同步任务运行历史 |
| `sign_record` | 电子签名留痕 |
| `Schedule_StaffDuty` | 员工当班 |
| `Schedule_StaffDutyOverride` | 当班临时调整 |
| `Schedule_Patient` | 排班患者 |

> 利好：启动自检（`database/health.go`）会对缺失新表打印 `⚠️ 缺少新表 …（请在部署阶段执行 …建表）`。**部署后请检查启动日志确认无此告警。**

**HIS 同步另需配置**（`.env`，见 `.env.example` 新增项）：HIS Oracle 连接（Host/Port/Service/Username/Password）+ `LegacyTenantID`。

---

## 6. 上线 Checklist（建议）

- [ ] 部署库执行 `docs/sql/deploy_new_tables.sql`（9 张新表）
- [ ] 检查 server 启动日志：**无** `⚠️ 缺少新表` 告警
- [ ] 配置 `.env` 的 HIS Oracle 连接与 `LegacyTenantID`
- [ ] 先以 `his-sync --job his_exam_report --once` 小批跑通（确认 ID_NO 匹配率、created/updated 计数）
- [ ] 验证当班门禁：用非当班医生账号签发当日处方应被拒；当班医生应可签发
- [ ] 验证电子签留痕：签发后 `sign_record` 有记录、`ConfirmTime` 落库且不改执行态
- [ ] 质控榜核对：医生姓名正确显示（无空/ID）

---

## 7. 风险与待办

| 项 | 等级 | 说明 |
|---|---|---|
| 新表未建即上线 | **高** | 端点 500；靠启动自检兜底，部署须执行 SQL（§5） |
| HIS 同步无离线单测 | 中 | 建议接生产前用 `--once` 小批灰度，核对匹配率与计数 |
| ID_NO 匹配不到的患者 | 中 | 同步时计入 `skipped`，不报错；需人工核对未匹配清单（可经 `sync_patient_handler` 手工建映射） |
| 当班门禁依赖排班数据 | 低 | fail-open 设计，排班缺失不拦截；但意味着排班未维护时门禁形同虚设，需配套排班落地 |
| 前端包体积 | 低 | 主包 1.6MB（gzip 433KB），vite 提示可代码分割；非阻塞 |

---

*报告基于隔离 worktree 静态构建 + 单元测试得出，未做带真实 legacy 库的端到端验证；端到端项见 §6 Checklist。*

---

## 附：本地实测复核注记（2026-06-20）

针对本报告在真实工作区（`master @ b31161f`）做了二次核对，结论与修正如下：

1. **E2E 测试已合入并实跑通过**：将本报告随附的 `his_sync_pipeline_e2e_test.go` 合入 `ai-hms-backend/internal/services/`，执行 `go test ./internal/services -run TestHisSync_ -count=1 -v`，**5 个用例全部 PASS**（日志中红色 `record not found` 为 GORM 预期分支，非失败）。测试依赖 `github.com/glebarez/sqlite v1.11.0`（已在 go.mod）。方法签名/常量与 `his_exam_report_sync_service.go`、`external_patient_mapping_service.go`、`sync_job_service.go` 及 `models/sync_job.go` 全部吻合。该测试填补了 HIS 同步落库管线此前零单测的空白，作为资产保留。
2. **§3.1 归类已更正**：见上文 ⚠️。HDIS 资产归属 `a7f1c2d`，非本次 HIS 同步。
3. **9 张新表与健康检查告警**确认无误：`internal/database/health.go` 的 `RequiredNewTables` 恰为 9 张，`LogRequiredTablesStatus` 会打印 `⚠️ 缺少新表 …`（AutoMigrate 永久禁用，§5 前置结论成立）。
4. **go-ora 驱动**确认：`go.sum` 含 `sijms/go-ora/v2 v2.9.0`。
5. **关于 §4 "ACTRS / 契约02"**：此为本报告引用的外部规划语境，本仓库 `docs/` 无对应 ACTRS/契约文档；不影响技术结论，读者勿据此误以为仓库内存在 ACTRS 实现。

> 注：随附测试的 docs 副本（`his_sync_pipeline_e2e_test(1).go`）已删除，以合入位置 `internal/services/his_sync_pipeline_e2e_test.go` 为准，避免两份漂移。
