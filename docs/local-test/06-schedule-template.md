# 阶段六：排班扩展、模板、周排班测试

> 测试时间：2026-06-07
> 测试分支：`fix/legacy-ui-restore`
> 提交号：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`

## 1. 病区扩展配置

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 读取 | GET /api/v1/schedule/config/wards | 200 | 200 (空表) | 通过 |
| 写入(admin) | PUT /api/v1/schedule/config/wards | 200 | 200 | 通过 |
| 写入(doctor) | PUT /api/v1/schedule/config/wards | 403 | 403 | 通过 |

### 验证数据

| Id | TenantId | WardId | ZoneType | Note |
|----|----------|--------|----------|------|
| 1 | 3 | 1 | A | TEST_AI_HMS_ward_ext |

## 2. 床位机器扩展

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 读取 | GET /api/v1/schedule/config/machines | 200 | 200 (空表) | 通过 |
| 写入 | PUT /api/v1/schedule/config/machines | 200 | 200 | 通过 |

### 验证数据

| Id | TenantId | BedId | MachineType | SupportedModes | MachineCode |
|----|----------|-------|-------------|----------------|-------------|
| 1 | 3 | 28 | HDF | HD,HDF,HF | TEST_AI_HMS_M001 |

## 3. 患者排班骨架

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 读取 | GET /api/v1/schedule/config/patient-profiles | 200 | 200 (空表) | 通过 |
| 写入 | PUT /api/v1/schedule/config/patient-profiles | 200 | 200 | 通过 |

### 验证数据

| Id | TenantId | PatientId | ZoneTag | FreqPattern | DefaultMode | HdfEnabled |
|----|----------|-----------|---------|-------------|-------------|------------|
| 1 | 3 | 300410 | A | 30 | HDF | true |

## 4. 奇偶周设置

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 读取 | GET /api/v1/schedule/config/settings | 200 | 200 | 通过 |
| 写入 | PUT /api/v1/schedule/config/settings | 200 | 200 | 通过 |

### 验证数据

| Id | TenantId | SettingKey | SettingValue | SettingType |
|----|----------|------------|-------------|-------------|
| 1 | 3 | OddEvenWeekAnchorMonday | 2026-06-01 | date |

## 5. 周排班查询

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 查询一周排班 | GET /api/v1/schedule/week?startDate=2026-06-01&endDate=2026-06-07 | 200 | 200 | 通过 |

## 6. 模板保存

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 保存模板 | POST /api/v1/schedule/template/save | 200 | 200 | 通过 |

### 请求体

```json
{
  "name": "TEST_AI_HMS_Aqu_template",
  "scope": "A",
  "wardId": 1,
  "items": [{
    "patientId": 300410, "zoneTag": "A", "wardId": 1, "shiftId": 1,
    "freqPattern": 30, "fixedHdBedId": 28,
    "hdfEnabled": true, "hdfWeekday": 3, "hdfWeekParity": 1
  }]
}
```

### 验证数据

**Schedule_ScheduleTemplate**

| Id | Name | Scope | WardId | IsActive |
|----|------|-------|--------|----------|
| 1 | TEST_AI_HMS_Aqu_template | A | 1 | true |

**Schedule_ScheduleTemplateItem**

| Id | TemplateId | PatientId | ShiftId | FreqPattern | FixedHdBedId |
|----|------------|-----------|---------|-------------|-------------|
| 1 | 1 | 300410 | 1 | 30 | 28 |

## 7. 模板应用

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 首次应用 | POST /api/v1/schedule/template/apply | 200 | 200（**修复后**） | 通过 |
| 重复应用 | POST /api/v1/schedule/template/apply | 400 拒绝 | 400 | 通过 |

> 首次应用初始返回 500（PatientPlanId NOT NULL），修复后通过。

### 请求体

```json
{"templateId": 1, "targetDate": "2026-06-01", "wardId": 1}
```

### 生成数据

**Schedule_PatientShift (Status=10 草稿)**

| Id | PatientId | ShiftId | BedId | PatientPlanId | Status |
|----|-----------|---------|-------|---------------|--------|
| 28169 | 300410 | 1 | 28 | 0 | 10 |

**Schedule_PatientShiftExt**

| Id | PatientShiftId | DialysisMode | SourceType | RuleStatus | SourceTemplateItemId |
|----|----------------|--------------|------------|------------|---------------------|
| 1 | 28169 | HDF | 10 | 10 | 1 |

## 8. 重复排班检测

| 测试场景 | 预期 | 实际 | 状态 |
|----------|------|------|------|
| 同日同班同患者再次应用 | 400 拒绝 | 400 | 通过 |

### SQL 验证

```sql
-- 患者同日同班重复检测
SELECT "TenantId", "PatientId", DATE("TreatmentTime") AS dt, "ShiftId", COUNT(*) AS cnt
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (40, 50, 60)
GROUP BY "TenantId", "PatientId", DATE("TreatmentTime"), "ShiftId"
HAVING COUNT(*) > 1;
-- 结果：0 行 ✓

-- 同床同日同班重复检测
SELECT "TenantId", "BedId", DATE("TreatmentTime") AS dt, "ShiftId", COUNT(*) AS cnt
FROM "Schedule_PatientShift"
WHERE "Status" NOT IN (40, 50, 60) AND "BedId" IS NOT NULL
GROUP BY "TenantId", "BedId", DATE("TreatmentTime"), "ShiftId"
HAVING COUNT(*) > 1;
-- 结果：0 行 ✓
```

## 9. 代码修复记录

| 文件 | 修改内容 |
|------|----------|
| `internal/services/schedule_template_service.go:367` | 添加 `PatientPlanId: &[]int64{0}[0]` 修复 NOT NULL 约束 |

## 10. 阶段六结论

| 检查项 | 状态 |
|--------|------|
| 病区扩展 CRUD | 通过 |
| 机器扩展 CRUD | 通过 |
| 患者排班骨架 CRUD | 通过 |
| 奇偶周设置 | 通过 |
| 周排班查询 | 通过 |
| 模板保存 | 通过 |
| 模板应用（修复后） | 通过 |
| 重复排班检测 | 通过 |
| 同日同班无重复 | 通过（SQL 验证） |
| 同床同日同班无重复 | 通过（SQL 验证） |
| 管理员写权限保护 | 通过 (403) |

### 发现问题

- **ISSUE-010 (P1)**：ApplyTemplate 不设置 PatientPlanId 导致 NOT NULL 冲突 → 已修复（添加默认值 0）
