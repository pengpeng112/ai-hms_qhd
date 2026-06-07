# 排班功能重构执行计划 · 总索引

> 状态：早期保守路线，可执行前需与扩展表融合路线二选一
> 适用范围：`ai-hms-backend/`、`ai-hms-frontend/`
> 关联：根目录 `老血透数据库表结构-合并版.md`（老库表结构权威文档）
> 执行前补充：必须先读 `docs/schedule-plan/RISK-REVIEW.md`
> 老系统逻辑补充：必须先读 `docs/schedule-plan/OLD-SYSTEM-LOGIC-SUPPLEMENT.md`
> 扩展表融合路线：见 `docs/schedule-plan/NEXT-ROADMAP.md`

## 路线选择提示

当前仓库已存在另一条“老表主记录 + `Schedule_*` 扩展表”的排班融合路线，并已完成阶段 0 到阶段 3。后续执行前必须先确认采用哪条路线：

- 保守路线：本 README 描述的“只扩老表字段、不新建表”。
- 扩展表融合路线：`docs/schedule-plan/NEXT-ROADMAP.md` 描述的“老表主记录 + 新增扩展表”。

两条路线不能混用。若继续当前阶段 0-3 的成果，应以 `NEXT-ROADMAP.md` 为后续计划入口。

## 给执行 AI 的"一页须知"

**先读本 README、`RISK-REVIEW.md` 与 `OLD-SYSTEM-LOGIC-SUPPLEMENT.md`，再每次只读一个阶段文件**。每个 Step 即一个 PR，结构固定：
```
Step X · 标题
├─ 改哪些文件（精确路径）
├─ 怎么改（含代码 / SQL 片段）
├─ 自检 grep（可粘贴命令）
└─ 验收 Checklist
```

做完一个 Step 必须停下来：

1. 后端 `cd ai-hms-backend && ./scripts/verify.sh` 通过
2. 前端 `cd ai-hms-frontend && npm run lint && npm run build` 通过
3. 提交 PR（≤ 800 行 diff），等用户书面 ✅ 再开下一个 Step

不允许把多个 Step 合并到一个 PR；不允许跨阶段顺手改其他文件。

---

## 老库现状摘要（背景方案核心结论，必须先理解再开工）

### 老库已有抽象（**不要重新发明**）

| 老库表 / 字段 | 用途 | 决定 |
|---|---|---|
| `Plan_PatientPlan` | 患者方案，含 `OddWeekFrequency`（单周频次）+ `EvenWeekFrequency`（双周频次）+ `DialysisMethod` | **就是合约表**，仅扩 3 字段，不新建 |
| `Plan_PatientPlanPrescription` | 处方版本表 | 已支持版本化，复用 |
| `Schedule_PatientShift.ShiftTiming` | 10=临时 / 20=长期 | **就是临时病人标识**，复用 |
| `Schedule_PatientShift.Status` | 10 草稿/20 已确认/30 用户确认/40 用户取消/50 排班取消/60 转出 | 取消/转出走 Status，不加软删字段 |
| `Schedule_Bed.IsDisabled` / `Schedule_Shift.IsDisabled` | 床位 / 班次启用位 | **就是床位时间表的最小实现**，不新建 BedTimetable |

### 4 处误判已修正（用户已确认）

1. ❌ 旧方案"新建 PatientContract" → ✅ 改为 `Plan_PatientPlan` 加 3 个可空字段
2. ❌ 旧方案"新建 BedTimetable" → ✅ 中心不需要"某班次某周几不开"粒度，用现有启用位
3. ❌ 旧方案"新建 BedMoveLog" → ✅ `Schedule_PatientShift` 加 `AdjustReason` 字段即可
4. ❌ 旧方案"加 deleted_at 软删" → ✅ 复用 `Status=40/50/60`

### 用户确认决策（不要再问）

- 合约录入：医生开 `Plan_PatientPlan`（频次+模式），护士在排班页填首选床位/班次/周几
- 模式按周几切换（如周一HD/周三HDF）：**分两步走**——P2 先做"默认+备选"，等查测试库后决定是否加 weekdayMode
- 床位时间表维度：全院级（周几×班次），用现有启用位
- 换床记录：只加 `AdjustReason` 字段，不新建 Log 表（接受"哪次换床不可追溯"）
- 临时病人：用 `ShiftTiming=10` 派生，前端加角标
- **数据安全**：只扩展字段不影响原有数据（用户硬性要求）

---

## 老库守则（红线 · 违反直接打回）

参见根 `CLAUDE.md` "老血透数据库守则"节。本计划再加 5 条强约束：

1. **禁止 `DROP` / `TRUNCATE` / `DELETE` / `UPDATE` 任何老库现存数据**。本计划只做 `ALTER TABLE ADD COLUMN`。
2. **所有 DDL 必须放进 `ai-hms-backend/scripts/legacy_alter_*.sql` 文件 + DBA 审核**，禁止 `AutoMigrate`、禁止应用启动时自动执行。
3. **新增字段必须 `NULLABLE` 且不带 `DEFAULT`**（NULL = 未填写，老数据保持原状）。
4. **不新建任何老库表**。本计划全部改 `Plan_PatientPlan` 与 `Schedule_PatientShift` 已有表。
5. **GORM tag 对齐列名优先于改库**（按 `LEGACY_TABLE_FIELD_MAPPING.md`）；写 SQL 前先 grep 老库是否已有近似字段。

---

## 阶段顺序（严格串行）

```
P0-0 ─→ P0 ─→ P0-7 ─→ P2 ─→ P2B ─→ P3
```

- **P1（床位时间表）已删除**——用户答复"用现有启用位足够"。
- **P2B（模板复制兼容）** 是老系统模板逻辑补充，必须先确认模板表/配置真实来源后再做。
- **P4（历史反推合约 / 节假日特殊周 / 自动定时渲染）**为上线后迭代项，本计划不含。

| 阶段 | 文件 | 内容 | 工期 | 是否需 DBA |
|---|---|---|---|---|
| P0-0 | `P0-bug-and-gap.md` | 当前接口闭环与事实校准：`/schedule/week`、`/move`、`/swap`、状态口径、字段落库 | 1-2 天 | ❌ 仅前后端 |
| P0 | `P0-bug-and-gap.md` | 修角标 / PUT 修改 / 频次差集 / 班次 CRUD / 批量保存评估 / 临时病人角标 | 1-2 天 | ❌ 仅前后端 |
| P0-7 | `OLD-SYSTEM-LOGIC-SUPPLEMENT.md` | 老系统四类校验最小闭环：历史、已过班次、治疗中、唯一性 | 1-2 天 | ❌ 仅前后端 |
| P2 | `P2-contract-render.md` | `Plan_PatientPlan` 加 3 字段 + 待落位侧栏 + 周渲染服务 | 2-3 周 | ✅ 1 份 SQL |
| P2B | `OLD-SYSTEM-LOGIC-SUPPLEMENT.md` | 模板复制兼容：正式/草稿模板、2/4 周循环、未来周无排班提示复制 | 待确认 | 待确认 |
| P3 | `P3-log-and-import.md` | `Schedule_PatientShift` 加 1 字段 + 换床日志 + CSV 导入 + 报表 | 1-2 周 | ✅ 1 份 SQL |

### 数据库变更总览（共 4 个字段，全部 NULLABLE，不带 DEFAULT）

| 阶段 | 表 | 字段 | 类型 | 用途 |
|---|---|---|---|---|
| P2 | Plan_PatientPlan | PreferredBedId | BIGINT NULL | 护士落位的首选床位 |
| P2 | Plan_PatientPlan | PreferredShiftId | BIGINT NULL | 首选班次 |
| P2 | Plan_PatientPlan | PreferredWeekdayMask | SMALLINT NULL | 周几位掩码（bit0=周一…bit6=周日） |
| P3 | Schedule_PatientShift | AdjustReason | VARCHAR(64) NULL | 调整原因（bed_move/swap/...） |

**保证**：上述字段 NULL 时业务表现 = "未启用新功能"，与现状完全一致。即使新功能上线后某患者未落位，他依然按现有手动排班流程工作。

---

## 全局红线（每个 Step 都适用）

1. 不动现存字典 typeCode 与映射（`AGENTS.md` 字典工作流）
2. 不引入新依赖，除非该阶段明确允许
3. 注释一律中文，与现有代码库一致
4. 触碰旧文件先 Read 全文，禁止凭关键字 sed 全局替换
5. 修改 GORM model 必须保留现有 `gorm:"column:XXX"` tag，不动列名
6. 不确定的事一律记到 `docs/legacy-migration-uncertain-field-checklist.md` 等用户确认，**不要靠猜测推进**
7. 遇到与本计划冲突的需求，停下来问用户，不要"灵活变通"

## 提交规范

- 提交格式：`feat(schedule): xxx (P0-1)` / `refactor(schedule): xxx (P2-3)`
- PR 标题前缀：`[P0-1]` / `[P2-3]` 等
- PR 影响面：`backend` / `frontend` / 双向
- DDL PR 单独提，标 `[P2-1] DDL` / `[P3-1] DDL`，需 DBA 审批通过才合并
- DDL PR 描述必须写："本 PR 不部署应用代码，仅给 DBA 审核 SQL"
- DDL 落地后用 `SELECT COUNT(*) FROM 表 WHERE 新字段 IS NOT NULL` 验证 = 0

下一步：阅读 `docs/schedule-plan/RISK-REVIEW.md`、`docs/schedule-plan/OLD-SYSTEM-LOGIC-SUPPLEMENT.md` 与 `docs/schedule-plan/P0-bug-and-gap.md`，从 **P0-0** 开始。不要直接从原 Step 1 开工。
