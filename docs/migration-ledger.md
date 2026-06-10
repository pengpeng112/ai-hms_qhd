# AI-HMS 迁移台账

> 版本: 1.0 | 更新: 2026-06-10
> 用途: 追踪老血透系统→新系统的全部迁移工作，作为唯一索引。
> 每项数据迁移完成后打 ✓，记录变更日期和责任人。

---

## 一、已完成的迁移

### 1.1 排班模块 v2 迁移

| 项目 | 状态 | 文档 |
|------|------|------|
| 老库 Schedule_* 表直接复用 | ✅ 完成 | `docs/排班模块交接文档.md` |
| v2 数据模型→老库字段映射 | ✅ 完成 | `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md` |
| 唯一索引（DDL 从运行时移除） | ✅ 完成 | `ai-hms-backend/scripts/schedule_unique_indexes.sql` |
| 治疗状态同步（v1⇋v2） | ✅ 完成 | `docs/v1-v2-boundary.md` |
| 排班 API（/api/v2/schedule/*） | ✅ 完成 | `docs/排班模块交接文档.md` |

### 1.2 患者管理

| 项目 | 状态 | 说明 |
|------|------|------|
| 老库 PatientProfile 复用 | ✅ 完成 | 字段映射见 LEGACY_TABLE_FIELD_MAPPING.md |
| 住院信息聚合 | ✅ 完成 | Registration_* → API |
| 检验报告同步 | ✅ 完成 | HDIS/LIS 外部同步 |

### 1.3 治疗管理

| 项目 | 状态 | 说明 |
|------|------|------|
| Treatment_Treatment 读写 | ✅ 完成 | v1 治疗服务主控 |
| Schedule_PatientShift 状态同步 | ✅ 完成 | v2 上机/下机 ⇌ v1 治疗服务 |
| 临床病史迁移 | ✅ 完成 | `ai-hms-backend/scripts/migrate_medical_history_fields.sql` |
| 转归记录迁移 | ✅ 完成 | `ai-hms-backend/internal/database/migrate_outcome.sql` |

### 1.4 身份认证

| 项目 | 状态 | 说明 |
|------|------|------|
| ASP.NET Identity V3 密码哈希兼容 | ✅ 完成 | `auth_service.go` |
| 老库 Identity_Users 直接认证 | ✅ 完成 | 不在新表存储用户 |
| JWT Token 签发 | ✅ 完成 | HMAC-SHA256 |

---

## 二、待完成的迁移

### 2.1 排班模块

| 项目 | 优先级 | 阻塞原因 |
|------|--------|---------|
| MachineId 语义统一（0 vs NULL） | P2 | 待 DBA 确认老库数据分布 |
| 权限码迁移（角色→权限码） | P2 | 待老库 Authorization_Permissions 配置 |
| 前端测试完善 | P2 | 已引入 Vitest，待增加覆盖 |

### 2.2 人员管理

| 项目 | 优先级 | 说明 |
|------|--------|------|
| 员工数据迁移（Organ_Employee） | 已复用 | 直接在老库读写 |
| 角色权限完整映射 | P2 | 部分角色在新老系统间未映射 |

---

## 三、不确定项（需业务确认）

详见 `docs/legacy-migration-uncertain-field-checklist.md`

| 编号 | 字段 | 不确定原因 |
|------|------|-----------|
| U-1 | TreatmentTime vs ScheduleDate | 老库列名为 TreatmentTime，已统一使用 |
| U-2 | MachineId 默认值 | 老库允许 NULL，v2 使用 NOT NULL DEFAULT 0 |
| U-3 | ZoneType 推断逻辑 | 根据 Zone Tag 自动推断 A/B/C 三区 |
| U-4 | WardId 业务含义 | 单病区系统（WardId=3），CRRT 跨区需特殊处理 |

---

## 四、文档索引

### 核心迁移文档

| 文档 | 位置 | 内容 |
|------|------|------|
| 老数据库表结构 | `老血透数据库表结构-合并版.md` | 102 张老库表完整 DDL |
| 新数据库表结构 | `新数据库表结构.md` | 新系统表设计 |
| 老库字段映射 | `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md` | v2 模型→老库列一对一映射 |
| 不确定项清单 | `docs/legacy-migration-uncertain-field-checklist.md` | 需业务确认的字段 |
| 排班模块交接 | `docs/排班模块交接文档.md` | 排班子系统全貌 |
| v1/v2 边界 | `docs/v1-v2-boundary.md` | 主系统与排班读写边界 |
| 老数据库设计 | `老数据库表设计.md` | 原始设计说明 |
| 数据库设计 | `DATABASE_DESIGN.md` | 新系统设计说明 |

### SQL 迁移脚本

| 脚本 | 位置 | 说明 |
|------|------|------|
| 排班唯一索引 | `ai-hms-backend/scripts/schedule_unique_indexes.sql` | DBA 建索引（含重复探测） |
| 转归记录 | `ai-hms-backend/scripts/migrate_outcome_record_door_rule.sql` | 转归门规则迁移 |
| 病史字段 | `ai-hms-backend/scripts/migrate_medical_history_fields.sql` | 病史字段迁移 |
| v2 合并 | `docs/sql/v2_merge_legacy.sql` | v2 合并老库 SQL |
| 转归记录 | `ai-hms-backend/internal/database/migrate_outcome.sql` | 转归记录迁移 |

### 评估与改进

| 文档 | 位置 | 说明 |
|------|------|------|
| 系统评估报告 | `docs/20260610ai 评估/系统评估报告(1).md` | 2026-06-10 AI 评估 |
| 系统改进计划 | `docs/20260610ai 评估/系统改进计划(1).md` | 三阶段改进计划 |
| 审计说明 | `docs/20260610ai 评估/审计说明.md` | 审计方法与约定 |

---

## 五、迁移纪律

1. **严禁运行时 DDL** — `database/migrate.go` 永久封锁 AutoMigrate/DropTables
2. **唯一真值源** — 每项数据只在一个模块写入，交叉读写通过明确 API 同步
3. **字段映射不猜测** — 不确定字段记录到 `legacy-migration-uncertain-field-checklist.md`
4. **废弃数据不删除** — 老库历史数据保留，不清除、不做 DROP
5. **每次迁移打 ✓** — 本台账持续更新，避免"已做但没说"的信息孤岛

---

*最后更新: 2026-06-10 — 安全红线修复 + 角色权限码迁移 + 前端测试引入*
