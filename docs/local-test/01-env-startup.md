# 阶段一：环境、构建、启动测试

> 测试时间：2026-06-07
> 测试环境：Windows x64, 本地开发机

## 1. Git 状态

| 项目 | 值 |
|------|-----|
| 分支 | `fix/legacy-ui-restore` |
| 提交号 | `4f375a7884727807a0ac87f100ea0c7d8b8848e9` |
| 未提交文件 | 9 个已修改 + 1 个未跟踪 |
| 已修改文件 | `.env.example`, `cmd/create_admin/main.go`, `internal/middleware/auth.go`, `internal/models/legacy/organ_employee.go`, `internal/services/auth_service.go`, `internal/services/auth_service_test.go`, `internal/services/order_service.go`, `internal/services/schedule_week_service.go`, `src/components/PermissionGuard.tsx`, `src/services/auth.ts` |
| 未跟踪文件 | `docs/ai_hms_qhd_segmented_test_plan.md` |

## 2. 版本信息

| 组件 | 版本 |
|------|------|
| Go | go1.26.1 windows/amd64 |
| Node.js | v25.2.1 |
| npm | 11.9.0 |
| PostgreSQL | 13.3 (openEuler) |

## 3. 测试数据库

| 配置项 | 值 |
|--------|-----|
| 数据库类型 | PostgreSQL |
| Host | 10.20.1.153:5432 |
| 数据库名 | `dialysis` |
| 用户 | postgres |
| 表总数 | 192 |
| 患者数 | 365 |
| 医嘱数 | 2,347 |
| 排班数 | 27,094 |
| 用户数 | 404 |
| 治疗方案数 | 719 |

> 确认为测试库（非空但有明显测试数据量级），连接正常。

## 4. 后端构建

| 项目 | 结果 |
|------|------|
| `go build` | 通过 |
| `go test ./...` | 全部通过（6 个包有测试，均 PASS） |

### go test 详情

| 包 | 结果 |
|----|------|
| internal/integrations/hdis | PASS (0.728s) |
| internal/middleware | PASS (1.859s) |
| internal/models/types | PASS (0.971s) |
| internal/services | PASS (3.500s) |
| internal/utils | PASS (2.106s) |
| internal/utils/idgen | PASS (2.277s) |

## 5. 前端构建

| 项目 | 结果 |
|------|------|
| `npm ci` | 通过 (338 packages) |
| `npm run lint` | 0 错误, 3 警告 |
| `npm run build` | 通过 (19.1s) |

### Lint 警告详情

```
src/pages/dialysis-processing/execution/Consumables.tsx:19:41
  react-hooks/exhaustive-deps: useEffect missing dependency 'loadItems'

src/pages/dialysis-processing/execution/PostAssessment.tsx:267:6
  react-hooks/exhaustive-deps: useEffect missing dependency 'form.postNetWeight'

src/pages/dialysis-processing/execution/PostAssessment.tsx:273:6
  react-hooks/exhaustive-deps: useEffect missing dependency 'form.lossWeight'
```

### npm 安全漏洞

- 共 13 个漏洞：5 moderate, 8 high
- 未在此次构建中修复，建议后续运行 `npm audit fix`

## 6. 健康检查

| 端点 | 状态 | 说明 |
|------|------|------|
| `GET /health` | 200 OK | 返回 `{"status":"ok","db":"ok","service":"ai-hms-backend","version":"1.0.0"}` |
| `GET /healthz` | 不存在 | 改进项 |
| `GET /readyz` | 不存在 | 改进项 |

## 7. 数据库表检查

### 测试计划要求表验证

| 表名 | 存在 | 备注 |
|------|------|------|
| Identity_Users | 是 | 404 条记录 |
| Identity_Roles | 是 | |
| Identity_UserRoles | 是 | |
| Authorization_Roles | 是 | |
| Authorization_RoleUsers | 是 | |
| Organ_Employee | 是 | |
| Register_PatientInfomation | 是 | 365 条记录 |
| Plan_PatientPlan | 是 | 719 条记录 |
| Order_PatientOrder | 是 | 2,347 条记录 |
| Order_PatientDayOrder | 是 | |
| Schedule_Ward | 是 | |
| Schedule_Bed | 是 | |
| Schedule_Shift | 是 | |
| Schedule_PatientShift | 是 | 27,094 条记录 |
| Schedule_WardExt | **否** | **缺失** |
| Schedule_BedMachineExt | **否** | **缺失** |
| Schedule_PatientProfile | **否** | **缺失** |
| Schedule_PatientShiftExt | **否** | **缺失** |
| Schedule_ScheduleTemplate | **否** | **缺失** |
| Schedule_ScheduleTemplateItem | **否** | **缺失** |
| Schedule_ConflictQueue | **否** | **缺失** |
| Schedule_Calendar | **否** | **缺失** |
| Schedule_MachineOutage | **否** | **缺失** |

### 缺失表影响分析

9 张 Schedule 扩展表缺失，直接影响：
- 阶段五（排班基础配置）：病区扩展、床位机器扩展、患者排班骨架
- 阶段六（排班模板、周排班）：模板 CRUD、模板应用、周排班查询
- 阶段七（透析执行）：患者排班扩展信息

### 其他关键表（存在）

| 表名 | 用途 |
|------|------|
| Treatment_Treatment | 透析治疗主表 |
| Treatment_BeforeSigns | 透前体征 |
| Treatment_BeforeCheck | 透前核对 |
| Treatment_DuringSigns | 透中体征 |
| Treatment_DuringParam | 透中参数 |
| Treatment_AfterSigns | 透后体征 |
| Treatment_AfterSymptom | 透后症状 |
| Treatment_MaterialTrace | 耗材追溯 |
| Auxiliary_EquipmentInfomation | 设备信息 |
| Auxiliary_EquipmentDisinfection | 设备消毒 |
| Auxiliary_MaterialInfomation | 耗材信息 |
| Stock_Stock | 库存 |
| Stock_InOutStorage | 出入库 |
| Stock_InOutStorageDetail | 出入库明细 |
| Auxiliary_HealthEducation | 健康宣教 |
| Auxiliary_PatientHealthEducation | 患者健康宣教 |

## 8. 启动日志异常

- 无 error 日志
- `[WARNING] You trusted all proxies, this is NOT safe.` — 改进项，建议配置 `GIN_TRUSTED_PROXIES`
- `[LEGACY-DB] startup in legacy database mode: AutoMigrate and startup seed initialization are disabled` — 符合预期
- `[CRON] order cron is disabled` — 符合预期（`ORDER_CRON_ENABLED` 未设为 true）

## 9. 阶段一结论

| 检查项 | 状态 |
|--------|------|
| 代码可构建 | 通过 |
| go test 通过 | 通过 |
| 前端 lint 通过 | 通过（3 warning） |
| 前端 build 通过 | 通过 |
| 数据库连接 | 通过 |
| 健康检查 | 通过 |
| 缺失表 | **9 张表缺失（P1）** |
| 启动异常 | 无阻断性异常 |

### 确认问题

**ISSUE-001**：9 张 Schedule 扩展表缺失，详见 `issues.md`。

**后续阶段阻塞风险**：阶段五（排班基础配置）和阶段六（排班模板）可能因缺失表而无法完整测试，需在进入对应阶段前确认后端是否已在这些表上实现 CRUD。
