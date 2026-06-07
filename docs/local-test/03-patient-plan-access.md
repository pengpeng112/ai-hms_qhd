# 阶段三：患者、治疗方案、血管通路测试

> 测试时间：2026-06-07
> 测试分支：`fix/legacy-ui-restore`
> 提交号：`4f375a7884727807a0ac87f100ea0c7d8b8848e9`
> 测试患者：300410（支英俊）

## 1. 患者列表测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 分页查询 | GET /api/v1/patients?page=1&pageSize=3 | 200, 返回分页数据 | 200, 返回患者列表 | 通过 |
| 姓名模糊查询 | GET /api/v1/patients?name=test | 200 | 200 | 通过 |
| 在透过滤 | GET /api/v1/patients?onlyActive=true | 200 | 200 | 通过 |

## 2. 患者详情测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 患者详情 | GET /api/v1/patients/300410 | 200 | 200（**修复后**） | 通过 |
| 核心信息 | GET /api/v1/patients/300410/core | 200 | 200 | 通过 |
| 基本信息档案 | GET /api/v1/patients/300410/basic-info | 200 | 200 | 通过 |
| 不存在患者 | GET /api/v1/patients/999999 | 404 | 404 | 通过 |
| 非法ID | GET /api/v1/patients/abc | 400 | 400 | 通过 |

> 患者详情接口初始返回 500，原因见 ISSUE-003 / ISSUE-004。

## 3. 治疗方案测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 方案列表 | GET /api/v1/patients/300410/treatment-plans | 200, 返回方案列表 | 200 (count=2→3) | 通过 |
| 获取单个方案 | GET /api/v1/patients/300410/treatment-plan | 200 | 200 | 通过 |
| 新增方案 | POST /api/v1/patients/300410/treatment-plan | 201 | 200（**修复后**） | 通过 |
| 数据库验证 | Plan_PatientPlan 查询 | 新增记录存在 | ID=2063551146777120768, HDF | 通过 |

> 新增方案初始返回 500（VascularAccessId NOT NULL），修复后通过。详见 ISSUE-005。

### 新增方案输入

```json
{
  "weeklyFrequency": 3,
  "duration": 4,
  "dryWeight": 65.5,
  "status": "启用",
  "dialysisMode": {"mode": "HDF", "bloodFlow": 250},
  "anticoagulant": {"name": "TEST_AI_HMS_dilowfenzigansu"},
  "notes": "TEST_AI_HMS_treatment_plan_create_test"
}
```

### 数据库落库验证

```sql
SELECT "Id", "Name", "DialysisMethod", "VascularAccessId", "IsDisabled"
FROM "Plan_PatientPlan"
WHERE "TenantId" = 3 AND "PatientId" = 300410
ORDER BY "Id" DESC;
```

| Id | Name | DialysisMethod | VascularAccessId | IsDisabled |
|----|------|----------------|------------------|------------|
| 2063551146777120768 | 支英俊 | HDF | 0 | false |
| 792 | 支英俊 | HD | 471 | — |
| 790 | 支英俊 | HF | 471 | — |

## 4. 血管通路测试

| 测试场景 | 接口 | 预期 | 实际 | 状态 |
|----------|------|------|------|------|
| 通路列表 | GET /api/v1/patients/300410/vascular-accesses | 200 | 200 (count=1→2) | 通过 |
| 新增通路 | POST /api/v1/patients/300410/vascular-accesses | 201 | 200, id=2063551147133636608 | 通过 |
| 数据库验证 | Register_VascularAccess 查询 | 新增记录存在 | Id=2063551147133636608, AVF | 通过 |

> 注：Note 字段未正确落库（实际存储为空），见 ISSUE-006。

### 新增通路输入

```json
{
  "accessType": "AVF",
  "buildDate": "2025-01-01",
  "status": "正常",
  "note": "TEST_AI_HMS_vascular_access_test"
}
```

### 数据库落库验证

| Id | PatientId | AccessType | Note |
|----|-----------|------------|------|
| 2063551147133636608 | 300410 | AVF | (空) |
| 471 | 300410 | 带隧道和涤纶套的透析导管TCC | (空) |

## 5. 代码修复记录

| 文件 | 修改内容 |
|------|----------|
| `internal/services/patient_service.go:639` | 移除 `Preload("MedicalHistory")` — 表不存在 |
| `internal/services/patient_service.go:982` | 移除 `Preload("MedicalHistory")` + `Preload("TreatmentPlan")` — 表不存在 |
| `internal/services/patient_service.go:1607` | 新增 `"VascularAccessId": 0` — 修复 NOT NULL 约束 |

## 6. 阶段三结论

| 检查项 | 状态 |
|--------|------|
| 患者列表分页 | 通过 |
| 患者姓名模糊查询 | 通过 |
| 患者详情（修复后） | 通过 |
| 核心信息 | 通过 |
| 基本信息档案 | 通过 |
| 404/400 错误处理 | 通过 |
| 治疗方案列表 | 通过 |
| 治疗方案新增（修复后） | 通过 |
| 血管通路列表 | 通过 |
| 血管通路新增 | 通过 |
| 数据落库一致性 | 通过（Note 字段除外） |

### 发现问题

- **ISSUE-003 (P1)**：`MedicalHistory` 模型映射到不存在的 `medical_histories` 表，导致患者详情 500。已移除 Preload。
- **ISSUE-004 (P1)**：`TreatmentPlan` 模型映射到不存在的 `treatment_plans` 表，导致患者详情 500。已移除 Preload。
- **ISSUE-005 (P1)**：`LegacyCreateTreatmentPlan` 缺少 `VascularAccessId` 字段，导致 NOT NULL 约束失败。已添加默认值 0。
- **ISSUE-006 (P2)**：血管通路新增时 Note 字段未正确落库（存储为空）。
