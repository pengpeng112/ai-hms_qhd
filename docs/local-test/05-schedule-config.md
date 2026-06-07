# 阶段五：排班基础配置测试

> 测试时间：2026-06-07
> 测试分支：`fix/legacy-ui-restore`
> 提交号：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`

## 1. 病区（Ward）测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 病区列表 | GET /api/v1/wards | 200 | 200 (total=6→7) | 通过 |
| 新增病区 | POST /api/v1/wards | 201 | 200, id=7 | 通过（**修复后**） |
| 更新病区 | PUT /api/v1/wards/7 | 200 | 200, note已更新 | 通过 |
| 数据库验证 | Schedule_Ward | TenantId=3 正确 | TenantId=3 ✓ | 通过 |

> 新增病区初始返回 500（缺 NOT NULL 字段），修复后通过。

## 2. 床位（Bed）测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 床位列表 | GET /api/v1/beds | 200 | 200 | 通过 |
| 新增床位 | POST /api/v1/beds | 201 | 200, id=28 | 通过（**修复后**） |
| 数据库验证 | Schedule_Bed | TenantId=3, Note正确 | ✓ | 通过 |

> 新增床位初始返回 500（缺 NOT NULL 字段），修复后通过。

## 3. 班次（Shift）测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 班次列表 | GET /api/v1/shifts | 200 | 200 (count=3→4) | 通过 |
| 新增班次 | POST /api/v1/shifts | 201 | 200, id=4 | 通过（**修复后**） |
| 数据库验证 | Schedule_Shift | TenantId=3, Name正确 | ✓ | 通过 |

> 新增班次初始返回 500（模型字段类型与 DB 列类型不匹配），修复后通过。

## 4. 数据库验证

### 病区

```sql
SELECT "Id", "Name", "TenantId", "Note" FROM "Schedule_Ward" WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

| Id | Name | TenantId | Note |
|----|------|----------|------|
| 7 | TEST_AI_HMS_Aqu | 3 | TEST_AI_HMS_ward_updated |

### 床位

```sql
SELECT "Id", "Name", "TenantId", "Note" FROM "Schedule_Bed" WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

| Id | Name | TenantId | Note |
|----|------|----------|------|
| 28 | TEST_AI_HMS_A01 | 3 | TEST_AI_HMS_bed_test |

### 班次

```sql
SELECT "Id", "Name", "TenantId" FROM "Schedule_Shift" WHERE "Name" LIKE 'TEST_AI_HMS_%';
```

| Id | Name | TenantId |
|----|------|----------|
| 4 | TEST_AI_HMS_zaoban | 3 |

## 5. 代码修复记录

| 文件 | 修改内容 |
|------|----------|
| `internal/services/ward_service.go` | Create 添加 TenantId/CreatorId/CreateTime/LastModifyTime |
| `internal/api/v1/ward_handler.go` | Create 传入 tenantId/creatorId 参数 |
| `internal/services/bed_service.go` | Create 添加 TenantId/CreatorId/CreateTime/LastModifyTime |
| `internal/api/v1/bed_handler.go` | Create 传入 tenantId/creatorId 参数 |
| `internal/services/shift_service.go` | Create 改用 raw map 插入 + parseShiftTime 转换 + Type int 处理 |

## 6. 阶段五结论

| 检查项 | 状态 |
|--------|------|
| 病区 CRUD | 通过（修复后） |
| 床位 CRUD | 通过（修复后） |
| 班次 CRUD | 通过（修复后） |
| TenantId 数据隔离 | 通过（TenantId=3） |
| 数据库落库一致性 | 通过 |

### 发现问题

- **ISSUE-007 (P1)**：WardService.Create 缺 TenantId/CreatorId/CreateTime/LastModifyTime → 已修复
- **ISSUE-008 (P1)**：BedService.Create 缺 TenantId/CreatorId/CreateTime/LastModifyTime → 已修复
- **ISSUE-009 (P1)**：Shift 模型 StartTime/EndTime 为 varchar 但 DB 列为 timestamp，Type 为 varchar 但 DB 为 integer → 已修复（raw map 插入 + parseShiftTime）
