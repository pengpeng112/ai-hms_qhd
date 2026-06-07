# 最终回归测试报告

> 提交号：`349802e1df104b2dd07cb6374421437c93580d55`
> 回归时间：2026-06-07 18:30
> 测试分支：`fix/legacy-ui-restore`

## 1. 构建与静态检查

| 检查项 | 结果 |
|--------|------|
| `go test ./...` | **全部 PASS**（6 包 tested，4 包 no test files） |
| `npm run lint` | **0 错误**，3 warnings（透析执行页面 react-hooks/exhaustive-deps，已知项） |
| `npm run build` | **通过**（27.16s） |

## 2. API 回归测试

| 测试项 | 接口 | 预期 | 实际 | 状态 |
|--------|------|------|------|------|
| 锁定用户拒绝登录 | POST /api/v1/auth/login | 401 | 401 | **通过** |
| 血管通路 notes 落库 | POST /patients/{id}/vascular-accesses | 200，Note 正确 | 200，Note=REGRESSION_349802e_note_test | **通过** |
| Dashboard 统计 | GET /api/v1/dashboard/stats | 200 | 200，activePatients=365 | **通过** |
| 模板应用 PatientPlanId | POST /schedule/template/apply | 200 | 200 | **通过** |
| 透析 7 步闭环 | 完整流程 | 全部 200 | 全部 PASS | **通过** |

## 3. 重点 Issue 回归确认

| Issue | 描述 | 回归结果 |
|-------|------|----------|
| ISSUE-001 | 14 张 Schedule 表缺失 | 已创建，阶段五/六测试通过 |
| ISSUE-002 | 锁定用户仍可登录 | 锁定用户→401 ✓ |
| ISSUE-003 | MedicalHistory 表不存在 | Preload 已移除 ✓ |
| ISSUE-004 | TreatmentPlan 表不存在 | Preload 已移除 ✓ |
| ISSUE-005 | VascularAccessId NOT NULL | 已添加默认值 0 ✓ |
| ISSUE-006 | 血管通路 Note 未落库 | 字段名 notes（非 note），已澄清 ✓ |
| ISSUE-007 | WardService 缺 NOT NULL | 已修复 ✓ |
| ISSUE-008 | BedService 缺 NOT NULL | 已修复 ✓ |
| ISSUE-009 | Shift varchar vs timestamp | raw map + parseShiftTime ✓ |
| ISSUE-010 | ApplyTemplate 缺 PatientPlanId | 默认值 0 ✓ |
| ISSUE-011 | upsert 缺 OperateTime | 已添加 ✓ |
| ISSUE-012 | upsert 缺 OperatorId | 已添加 ✓ |
| ISSUE-013 | SaveDisinfection 缺字段 | 重写插入逻辑 ✓ |
| ISSUE-014 | jsonb ParameterS 类型 | ::text 转换 ✓ |
| ISSUE-015 | DISTINCT + ORDER BY | 重建 listQuery ✓ |
| ISSUE-016 | 缺 Equipment 表 | 已创建 ✓ |
| ISSUE-017 | Dashboard Status int→text | "60" 字符串 ✓ |

## 4. 测试数据概览

| 表 | TEST_AI_HMS_% 行数 |
|----|-------------------|
| Identity_Users | 4 |
| Identity_Roles | 3 |
| Organ_Employee | 4 |
| Identity_UserRoles | 4 |
| Register_VascularAccess | ~5 |
| Plan_PatientPlan | ~3 |
| Order_PatientOrder | ~2 |
| Schedule_Ward | 1 |
| Schedule_Bed | 1 |
| Schedule_Shift | 1 |
| Schedule_WardExt | 1 |
| Schedule_BedMachineExt | 1 |
| Schedule_PatientProfile | 1 |
| Schedule_TenantSetting | 1 |
| Schedule_ScheduleTemplate | 1 |
| Schedule_ScheduleTemplateItem | 1 |
| Schedule_PatientShift | ~2 |
| Treatment_Treatment | ~5 |
| Auxiliary_EquipmentInfomation | ~2 |

> 测试数据未清理，保留供后续审查使用。**正式试运行前必须清理。**

## 5. 结论

- **P0 = 0**，**P1 全部已修复**（14/14），**P2 全部已澄清/修复**（3/3）
- **go test 全部通过**，**前端 lint 0 错误**，**npm build 通过**
- **5 项 API 回归全部通过**
- **建议进入受控小范围科室试运行**
