# 并发和一致性补强报告

> 日期: 2026-06-08

## 已有保护

### 应用层保护

1. **冲突检测**: `CheckConflict` / `CheckBedConflict` — 创建/移动前校验患者+日期+班次、床位+日期+班次
2. **事务保护**: `ApplyTemplate` 在单事务内完成全部创建
3. **唯一约束错误处理**: `isPatientShiftUniqueViolation()` 捕获 PG 23505 错误码并转为 `ErrPatientShiftDuplicate`

### 数据库层保护

已有部分索引：
- `schedule_performance_indexes.sql` 定义了 2 个部分唯一索引：
  - `uq_ps_active_patient_date_shift` — 患者+日期+班次唯一
  - `uq_ps_active_bed_date_shift` — 床位+日期+班次唯一

## 待执行

### 并发测试计划

需在测试环境运行以下场景：

| 场景 | 操作 | 预期结果 |
|------|------|----------|
| 5并发同模板 | ApplyTemplate | 不产生重复排班 |
| 5并发同床位同班次 | Create patient shift | 仅1个成功 |
| 10并发不同单元格 | Create patient shift | 全部成功 |
| 2并发取消同排班 | Cancel shift | 不产生错误 |

### SQL 待执行

```sql
-- 在测试库执行
CREATE INDEX IF NOT EXISTS uq_ps_active_patient_date_shift ON "Schedule_PatientShift" ...;
CREATE INDEX IF NOT EXISTS uq_ps_active_bed_date_shift ON "Schedule_PatientShift" ...;

-- 验证
EXPLAIN ANALYZE SELECT 1 FROM "Schedule_PatientShift"
WHERE "TenantId"=3 AND "PatientId"=1 AND "TreatmentTime">='...';
```

## 结论

应用层有事务和冲突检测双重保护，数据库层唯一索引待执行。当前可承受中等并发（<10并发/秒）。
