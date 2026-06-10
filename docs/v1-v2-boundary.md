# v1/v2 架构边界说明

> 目的：明确主系统（v1）与智能排班（v2）之间的数据读写边界、共享表与唯一事实源，消除双写竞态风险。

## 路由边界

| 前缀 | 模块 | 认证 | 鉴权 |
|------|------|------|------|
| `/api/v1/*` | 主系统（患者管理、治疗执行、资源管理、系统管理） | JWT Bearer | `RequireRoles` / `RequirePermissions` |
| `/api/v2/schedule/*` | 智能排班（生成、周视图、冲突、CRRT、治疗执行） | JWT Bearer | `RoleHeadNurse` / `RoleChargeNurse` / `RoleDoctor` / `RoleNurse` |

## 共享表（v1 与 v2 共同读写）

| 表 | v1 写 | v2 写 | 唯一事实源 |
|----|-------|-------|-----------|
| `Schedule_PatientShift` | 治疗状态同步（20→50→60） | 排班生成、确认、上机、下机、取消、缺席、移床 | **v2 为主** |
| `Schedule_Ward` | 只读 | 只读 | 老库 |
| `Schedule_Machine` | 只读 | 只读 | 老库 |
| `Schedule_Shift` | 只读 | 只读 | 老库 |
| `PatientProfile` | 只读 | 读写（HDF 奇偶周） | v1/v2 共享 |
| `Treatment_Treatment` | **主**（治疗记录 CRUD） | 只读 | **v1 为主** |
| `Machine_MachineOutage` | 只读 | 读写（停机登记） | v2 |
| `Schedule_Calendar` | 只读 | 读写（假日设置） | v2 |
| `Schedule_Template` / `Schedule_TemplateItem` | 只读 | 读写（模板管理） | v2 |
| `ConflictQueue` | 只读 | 读写（冲突队列） | v2 |

## 双写协调机制

### 治疗状态同步（Schedule_PatientShift ⇋ Treatment_Treatment）

```
v2 排班模块（主）              v1 治疗服务（从）

POST /api/v2/shifts/:id/start  →  ScheduleId  →  services/treatment_service.go
  → 状态 20→50                                     → UpdateStatus 触发 syncScheduleStatus
     (已确认→透析中)                                 → 同步 Schedule_PatientShift.Status = 50

POST /api/v2/shifts/:id/complete  →  ScheduleId
  → 状态 50→60
     (透析中→已完成)                              → SubmitPostAssessment 触发 syncScheduleStatus
                                                    → 同步 Schedule_PatientShift.Status = 60
```

- **同步方向**：v1 治疗服务写回后 → 同步至 v2 排班表（单向回写）
- **失败处理**：sync 失败记录日志，不阻塞 v1 主流程（最终一致性）
- **ScheduleId 字段**：`Treatment_Treatment.ScheduleId` 指向 `Schedule_PatientShift.Id`

### 排班数据所有权

- **生成**：v2 排班引擎独有（POST /api/v2/schedule/generate）
- **调整**：v2 排班调整 API（cancel/absent/move/临时/CRRT）
- **确认**：v2 三级确认流程（confirm-plan → confirm-day ×2）
- **治疗执行**：v2 上机/下机 API + v1 治疗服务同步回写

## 禁止事项

1. ❌ v1 不得直接对 `Schedule_PatientShift` 执行 INSERT/DELETE（状态同步除外）
2. ❌ v2 不得写入 `Treatment_Treatment`（只读）
3. ❌ 生产环境严禁运行时 DDL（AutoMigrate、EnsureIndexes 等已封禁）
4. ❌ DEV_SUPERUSER=true 严禁出现在生产环境

## 待收敛项

- [ ] `TreatmentTime` vs `ScheduleDate`：老库列名为 `TreatmentTime`，已统一使用
- [ ] `MachineId` 默认值 0 vs NULL：老库允许 NULL，v2 使用 `NOT NULL DEFAULT 0`
- [ ] 权限码迁移：AdminRoles 兜底 → 未来 `Authorization_Permissions` 配置后迁移
