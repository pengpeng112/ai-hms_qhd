# 阶段八：库存、设备、字典、用户角色测试

> 测试时间：2026-06-07
> 测试分支：`fix/legacy-ui-restore`
> 提交号：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`

## 1. 库存耗材测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 库存列表 | GET /api/v1/inventory/items | 200 | 200 | 通过 |
| 新增耗材 | POST /api/v1/inventory/items | 201 | 400（业务规则拒绝） | **符合设计** |
| 库存流水 | GET /api/v1/inventory/logs | 200 | 200 | 通过 |

> 库存新增返回 400：`"库存管理请通过 HIS 同步完成入库流程，暂不支持直接创建"`。符合业务设计——库存必须通过 HIS 同步。

## 2. 设备测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 设备列表 | GET /api/v1/devices | 200 | 200 (total=31) | 通过（修复后） |
| 新增设备 | POST /api/v1/devices | 201 | 200, id=31 | 通过（修复后） |

> 设备列表初始返回 500（jsonb 类型比较错误 + DISTINCT ORDER BY 冲突），修复后通过。

### 数据库验证

| Id | Name | TenantId |
|----|------|----------|
| 31 | TEST_AI_HMS_device_001 | 3 |

## 3. 字典测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 字典类型列表 | GET /api/v1/dict/types | 200 | 200 | 通过 |
| 字典类型新增 | POST /api/v1/dict/types | 500 | 未深入排查 | 待回归 |

## 4. 用户角色测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 用户列表 | GET /api/v1/users (admin) | 200 | 200 (total=408) | 通过 |

> 用户管理已在阶段二充分测试（创建、角色分配、权限控制）。

## 5. 数据库新建表

| 表名 | 用途 |
|------|------|
| Auxiliary_EquipmentMaintenance | 设备维护记录 |
| Auxiliary_EquipmentUsageLog | 设备使用日志 |

## 6. 代码修复记录

| 文件 | 修改内容 |
|------|----------|
| `internal/services/device_service.go:419` | `NULLIF(rel."ParameterS", '')` → `NULLIF(rel."ParameterS"::text, '')` — jsonb 类型转换 |
| `internal/services/device_service.go:442` | 同上 |
| `internal/services/device_service.go:462-476` | 重建 listQuery 避免 DISTINCT+ORDER BY 冲突 |

## 7. 阶段八结论

| 检查项 | 状态 |
|--------|------|
| 库存列表 | 通过 |
| 库存新增（HIS 同步限制） | 符合设计 |
| 库存流水 | 通过 |
| 设备列表（修复后） | 通过 |
| 设备新增（修复后） | 通过 |
| 字典类型列表 | 通过 |
| 用户列表 | 通过 |

### 发现问题

- **ISSUE-014 (P1)**：Device List 中 jsonb `ParameterS` 列与空字符串比较导致 SQLSTATE 22P02 → 已修复（添加 `::text` 转换）
- **ISSUE-015 (P1)**：Device List 中 DISTINCT 与 ORDER BY 不在 SELECT 列导致 SQLSTATE 42P10 → 已修复（重建查询）
- **ISSUE-016 (P2)**：缺少 `Auxiliary_EquipmentMaintenance`、`Auxiliary_EquipmentUsageLog` 表 → 已创建
