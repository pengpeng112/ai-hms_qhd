# 后端性能整改报告

> 整改日期: 2026-06-08

## 改动清单

### 1. DATE(TreatmentTime) → 范围查询（索引友好）

| 文件 | 函数 | 改动 |
|------|------|------|
| `patient_shift_service.go:86-93` | List | `DATE() >= DATE(?)` → `"TreatmentTime" >= ?` |
| `patient_shift_service.go:439` | GetByPatientAndDate | `DATE()=DATE(?)` → `>=? AND <?` |
| `patient_shift_service.go:463-464` | CheckConflict | `DATE()=DATE(?)` → `>=? AND <?` |
| `patient_shift_service.go:490-491` | CheckBedConflict | `DATE()=DATE(?)` → `>=? AND <?` |
| `schedule_template_service.go:423-424` | checkConflictTx | `DATE()=DATE(?)` → `>=? AND <?` |
| `schedule_template_service.go:437-438` | checkBedConflictTx | `DATE()=DATE(?)` → `>=? AND <?` |
| `schedule_board_service.go:186-187` | LoadBoard | `DATE()>=DATE(?)` → `"TreatmentTime" >= ?` |
| `schedule_generate_service.go:190` | buildBoard | `DATE()>=DATE(?)` → `"TreatmentTime" >= ?` |
| `dashboard_service.go:105-106` | GetStats | `DATE()=?` → `>=? AND <?` |

**影响**: 所有排班查询均可利用 `TreatmentTime` 上的索引。

### 2. 待排患者限制返回

`schedule_week_service.go:322` — `Limit(200)`，避免一次性返回 685 患者（412KB → ~120KB）

### 3. 索引脚本

`docs/sql/schedule_performance_indexes.sql` — 7 个组合索引 + 2 个部分唯一索引

### 4. 待验证 SQL（需在测试库执行）

```sql
EXPLAIN ANALYZE
SELECT * FROM "Schedule_PatientShift"
WHERE "TenantId" = 3
  AND "TreatmentTime" >= '2026-06-08 00:00:00'
  AND "TreatmentTime" <  '2026-06-15 00:00:00'
  AND "Status" NOT IN (40, 50, 60);
```

## 预期效果

| 指标 | 整改前 | 整改后 |
|------|--------|--------|
| /schedule/week 响应大小 | 412KB | ~120KB |
| 冲突检查 SQL | 全表扫描(DATE函数) | 索引查找 |
| 并发排班保护 | 仅应用层 | 部分唯一索引 |
