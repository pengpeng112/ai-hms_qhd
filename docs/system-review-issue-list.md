# AI-HMS 系统审查问题清单

> 生成时间：2026-05-31 | 最后审查：2026-05-31 20:50 | 审查范围：前端 + 后端全栈功能检查
> 前端 lint + build 通过 ✓ | 后端 go build 通过 ✓

---

## 一、前端语法/构建 Bug（已修复）

| # | 文件 | 问题 | 状态 |
|---|------|------|------|
| 1 | `execution/PreAssessment.tsx` | `handleSave` 中 `onSave({...})` 代码块重复了一次（约 10 行），导致解析错误 | ✅ 已修复 |
| 2 | `execution/PreAssessment.tsx` | `useEffect` 内同步 `setState`（`setForm` 在 `useEffect` 内直接调用）违反 React Hooks 规则 | ✅ 已修复，改为派生计算 `displayTargetUf` |
| 3 | `execution/PreAssessment.tsx` | `useEffect` + `useCallback` 模式调用 `loadVascularAccesses` 触发 `set-state-in-effect` 规则 | ✅ 已修复，改为标准 `useEffect` 内 async + cancel 模式 |
| 4 | `execution/HealthEducation.tsx` | "宣教人"和"签名护士"同绑 `form.nurseSign`，改一个另一个跟着变 | ✅ 已修复，删除重复的"签名护士"字段 |
| 5 | `execution/PostAssessment.tsx` | 下机护士硬编码 "本地预览管理员" | ✅ 已修复，改为 `--` 占位 |

---

## 二、前端页面功能完整度

### 2.1 占位/未实现页面

| # | 路由 | 文件 | 状态 | 说明 |
|---|------|------|------|------|
| 1 | `/dialysis-processing` | `DialysisProcessing.tsx` (pages 根) | stub | 仅透传重导出到 `./dialysis-processing/DialysisExecution`，无自有逻辑，属正常路由桥接 |
| 2 | — | `execution/Consumables.tsx` | stub | 只渲染标题+患者姓名，无任何表单/接口。对应旧系统 tab 缺失（无 2.4 目录），不在还原范围 |
| 3 | — | `execution/Disinfection.tsx` | stub | 同上，仅占位卡片。功能已在 `Verification.tsx` 机表消毒区域部分实现但按钮禁用 |
| 4 | — | `execution/FirstCheck.tsx` | stub | 功能已合并到 `Verification.tsx` 首次核对列 |
| 5 | — | `execution/SecondCheck.tsx` | stub | 功能已合并到 `Verification.tsx` 二次核对列 |
| 6 | `/schedule-templates/edit` | `ScheduleTemplateEditor.tsx` | stub | 仅"待实现"文字，无任何编辑功能 |
| 7 | `/education-management` | `EducationManagement.tsx` | partial | 有 API 加载和 Ant Table 渲染，但只有 3 列（标题/类型/状态），无增删改操作、无分页 |

### 2.2 部分实现页面

| # | 路由 | 文件 | 缺失功能 |
|---|------|------|----------|
| 8 | `/master-data` | `MasterData.tsx` | 药品/耗材 Tab 已接 REST API；宣教和收费 Tab 使用硬编码本地数据，无后端接口 |
| 9 | `/schedule-templates` | `ScheduleTemplateList.tsx` | 仅有 4 列表格，无新增/编辑/删除/设默认操作，无分页 |

---

## 三、后端 API 缺失（前端有调用但后端无路由）

### 3.1 完全缺失的模块

| # | 前端调用 | HTTP 方法 | 路径 | 严重程度 | 状态 | 说明 |
|---|---------|-----------|------|----------|------|------|
| 1 | `restApi.getHealthEducationContents()` | GET | `/api/v1/health-educations` | 🔴高 | ❌ | 后端无任何 health-education handler/service/route |
| 2 | `restApi.getPatientHealthEducations()` | GET | `/api/v1/patients/:id/health-educations` | 🔴高 | ❌ | 同上 |
| 3 | `restApi.createPatientHealthEducation()` | POST | `/api/v1/patients/:id/health-educations` | 🔴高 | ❌ | 同上 |

### 3.2 治疗模块缺失的端点

| # | 前端调用 | HTTP 方法 | 路径 | 严重程度 | 状态 | 说明 |
|---|---------|-----------|------|----------|------|------|
| 4 | `restApi.saveTreatmentDisinfection()` | PUT | `/api/v1/treatments/:id/disinfection` | 🟡中 | ❌ | Verification.tsx 机表消毒按钮已标注"功能待后端接口就绪"并禁用 |
| 5 | `restApi.saveTreatmentSummary()` | PUT | `/api/v1/treatments/:id/summary` | 🟡中 | ❌ | DialysisSummary.tsx 保存小结按钮已禁用 |

### 3.3 用户管理模块缺失端点

| # | 前端调用 | HTTP 方法 | 路径 | 严重程度 | 状态 | 说明 |
|---|---------|-----------|------|----------|------|------|
| 6 | `userApi.create()` | POST | `/api/v1/users` | 🔴高 | ❌ | 创建用户 |
| 7 | `userApi.getById()` | GET | `/api/v1/users/:id` | 🔴高 | ❌ | 获取用户详情 |
| 8 | `userApi.update()` | PUT | `/api/v1/users/:id` | 🔴高 | ❌ | 更新用户 |
| 9 | `userApi.remove()` | DELETE | `/api/v1/users/:id` | 🔴高 | ❌ | 删除用户 |
| 10 | `userApi.updateStatus()` | PUT | `/api/v1/users/:id/status` | 🔴高 | ❌ | 更新用户状态 |
| 11 | `userApi.resetPassword()` | PUT | `/api/v1/users/:id/password` | 🟡中 | ❌ | 重置密码 |
| 12 | `userApi.getRoles()` | GET | `/api/v1/users/:id/roles` | 🟡中 | ❌ | 获取用户角色 |
| 13 | `userApi.setRoles()` | PUT | `/api/v1/users/:id/roles` | 🟡中 | ❌ | 设置用户角色 |
| 14 | `userApi.getMyRoles()` | GET | `/api/v1/me/roles` | 🟡中 | ❌ | 获取当前用户角色 |

### 3.4 角色权限管理缺失端点

| # | 前端调用 | HTTP 方法 | 路径 | 严重程度 | 状态 | 说明 |
|---|---------|-----------|------|----------|------|------|
| 15 | `roleManagementApi.getRoleList()` | GET | `/api/v1/app-roles` | 🔴高 | ❌ | 角色列表 |
| 16 | `roleManagementApi.createRole()` | POST | `/api/v1/app-roles` | 🔴高 | ❌ | 创建角色 |
| 17 | `roleManagementApi.updateRole()` | PUT | `/api/v1/app-roles/:code` | 🔴高 | ❌ | 更新角色 |
| 18 | `roleManagementApi.deleteRole()` | DELETE | `/api/v1/app-roles/:code` | 🔴高 | ❌ | 删除角色 |
| 19 | `roleManagementApi.getPermissionTree()` | GET | `/api/v1/app-permissions/tree` | 🟡中 | ❌ | 权限树（后端仅有 `GET /permissions` 扁平列表） |

---

## 四、前端已调用但后端有路由可用的接口（无需修改）

以下前端调用与后端路由已对齐，功能应正常工作：

- 用户列表 `GET /api/v1/users` ✅
- 健康宣教（前端调用但后端无路由，见 §3.1）🔴
- 治疗全量 CRUD（含 before-signs/after-signs/during-params/first-check/second-check）✅
- 处方 CRUD + extract/execute/cancel ✅
- 医嘱 CRUD + from-template/stop/copy/group/ungroup ✅
- 字典管理全量 CRUD ✅
- 患者核心信息 ✅
- 血管通路 CRUD ✅
- 排班管理（班次+排班+周视图）✅
- Dashboard 统计 ✅
- 设备管理（列表+详情+消毒+维护+使用日志）✅
- 库存管理（品目CRUD+调整+日志+标签）✅
- 统计报表（质量+感染+血管+工作量）✅
- HDIS 集成设置 + Token 刷新 ✅
- 系统日志 ✅

---

## 五、后端已有路由但前端未使用的（潜在开发机会）

| # | 后端路由 | 说明 | 状态 |
|---|---------|------|------|
| 1 | `GET /api/v1/patients/stats` | 患者统计 — 本次会话已新增，PatientList 页面已调用 | ✅ |
| 2 | `DELETE /api/v1/patients/:id` | 删除患者 — PatientList 已使用 | ✅ |
| 3 | `POST /api/v1/patients` | 创建患者 |
| 4 | `GET/PUT /api/v1/patients/:id/basic-info` | 患者基本信息档案 |
| 5 | `POST /api/v1/patients/:id/adjustment-records` | 方案调整记录 |
| 6 | `GET/PUT /api/v1/patients/:id/medical-history` | 临床病史 |
| 7 | `GET/POST/PUT/DELETE /api/v1/patients/:id/outcome-records/*` | 转归记录 |
| 8 | `GET/POST /api/v1/patients/:id/lab-reports` | 检验报告 |
| 9 | `POST /api/v1/patients/:id/exam-reports/sync` | 检查报告同步 |
| 10 | `GET/POST /api/v1/patients/:id/key-indicators/*` | 关键指标 |
| 11 | `GET /api/v1/hospitalizations` + CRUD | 住院信息 |
| 12 | `GET/POST /api/v1/clinical-tasks` + `PUT status` | 临床任务 |
| 13 | `GET/POST/PUT/DELETE /api/v1/treatment-templates/*` + toggle/set-default | 诊疗方案模板 |
| 14 | `GET/POST/PUT/DELETE /api/v1/materials/catalog/*` + toggle + categories | 材料目录 |
| 15 | `GET/POST/PUT/DELETE /api/v1/drugs/catalog/*` + toggle + categories | 药品目录 |
| 16 | `GET/POST/PUT/DELETE /api/v1/order-templates/*` + toggle/set-default | 医嘱模板 |
| 17 | `POST /api/v1/dict/import` + `POST /api/v1/dict/items/init` | 字典批量导入/初始化 |

---

## 六、代码质量与一致性问题

| # | 类别 | 文件 | 问题描述 | 严重程度 | 状态 |
|---|------|------|----------|----------|------|
| 1 | GraphQL 残留 | `services/schedule.ts`, `treatment.ts`, `order.ts`, `vitals.ts`, `examination.ts` | 这 5 个文件使用 GraphQL 查询（通过 HDIS 兼容层），不是真实后端 REST API。当前后端无 GraphQL 端点，这些函数会 404 或返回空数据 | 🟡中 | ❌ |
| 2 | API 重复 | `services/restClient.ts` vs `services/orderApi.ts` | `restApi.getPatientOrders()` 与 `orderApi.list()` 都调 `GET /patients/:id/orders`，功能重复 | 🟢低 | ❌ |
| 3 | 类型定义冗余 | `services/types/api.ts` | 大量 GraphQL 相关类型与 REST API 类型并行存在 | 🟢低 | ❌ |
| 4 | 硬编码占位 | `DialysisSummary.tsx:104,147,148,174` | 凝血分级显示硬编码为"未记录"，未接入后端数据 | 🟡中 | ❌ |
| 5 | 硬编码占位 | `MidMonitoring.tsx:336-337` | "机位状态：正常传输中"、"同步间隔：60分钟/点" 硬编码 | 🟡中 | ❌ |
| 6 | 硬编码占位 | `PostAssessment.tsx:431` | "是否进行内瘘/导管护理健康指导" checkbox 只读且硬编码 checked | 🟡中 | ❌ |
| 7 | 按钮禁用 | `PreAssessment.tsx` 底部操作栏 | "暂存草稿" 按钮始终 disabled，无后端接口 | 🟡中 | ❌ |
| 8 | 按钮禁用 | `Verification.tsx:462` | "功能待后端接口就绪" 按钮始终 disabled（消毒登记），对应后端无 `/treatments/:id/disinfection` 路由 | 🟡中 | ❌ |
| 9 | 按钮禁用 | `DialysisSummary.tsx:121` | "保存小结" 按钮始终 disabled，对应后端无 `/treatments/:id/summary` 路由 | 🟡中 | ❌ |
| 10 | 医嘱模板 | `MedicalOrders.tsx:433` | "从模板组调取" 模式仅显示占位提示，无实际功能 | 🟡中 | ❌ |
| 11 | 抗凝剂表 | `MedicalOrders.tsx:410` | 抗凝剂表格始终显示"暂无抗凝剂接口数据"，无后端专属接口 | 🟡中 | ❌ |

---

## 七、安全与配置注意事项

| # | 类别 | 说明 | 严重程度 |
|---|------|------|----------|
| 1 | 环境变量 | 后端 `JWT_SECRET`、`APP_SECRET`、`CORS_ALLOWED_ORIGINS` 必填，缺少则启动失败 | 🔴高 |
| 2 | 认证绕过 | `AUTH_EMERGENCY_ENABLED=false` 为安全默认值，不要重新引入默认凭证 | 🔴高 |
| 3 | GORM 大小写 | `SingularTable: true` + `NoLowerCase: true`，SQL 查询老库表列必须双引号 | 🔴高 |
| 4 | AutoMigrate | 永久阻断，不要恢复自动迁移 | 🔴高 |

---

## 八、建议整改优先级

### P0 — 阻断级（必须先修）
1. ✅ PreAssessment.tsx 语法错误（已修复）
2. ✅ HealthEducation.tsx 重复字段（已修复）
3. ✅ PostAssessment.tsx 硬编码管理员（已修复）
4. ✅ 用户列表 SQL 查询 `UserId` 列缺失降级（`user_service.go` 已修复）
5. ❌ 后端健康宣教模块缺失 — 需要新建 handler + service + route

### P1 — 高优（系统核心功能无法使用）
6. ❌ 后端用户管理 CRUD 缺失 — `UserManagement.tsx` 页面增删改用户将全部 404
7. ❌ 后端角色管理 CRUD 缺失 — `RoleManagement.tsx` 页面增删改角色将全部 404
8. ❌ `ScheduleTemplateEditor.tsx` — 完全空白，需要实现排班模板编辑功能

### P2 — 中等（功能受限但不阻断）
9. ❌ 治疗消毒登记后端接口 — `treatments/:id/disinfection`
10. ❌ 治疗小结保存后端接口 — `treatments/:id/summary`
11. ❌ 医嘱模板调取 — 前端已占位，需对接 `POST /patients/:id/orders/from-template`
12. ❌ `EducationManagement.tsx` — 增删改功能缺失
13. ❌ `MasterData.tsx` 宣教/收费 Tab — 硬编码数据需对接后端
14. ❌ 凝血分级等透后评估只读字段 — 已改为可选 CoagSelect，但 DialysisSummary 凝血摘要仍硬编码
15. ❌ PostAssessment 内瘘护理 checkbox 只读 — 未修复

### P3 — 低优（代码质量与工程改进）
16. ❌ GraphQL 残留服务清理 — 5 个文件引用 HDIS GraphQL 端点
17. ❌ 重复 API 函数合并 — `restApi.getPatientOrders` vs `orderApi.list`
18. ❌ 类型定义冗余 — GraphQL 与 REST 类型并行存在

### ✅ 本次会话已修复
- 患者管理：默认在科患者 + 统计数据显示（新增 `/patients/stats` 接口）
- 病区概览：切换为 Dashboard Stats 真实数据
- 患者详情：右侧面板改为悬浮按钮+抽屉/弹窗
- 透前评估：目标超滤量自动计算、A/V端位点下拉、内瘘默认标签、皮肤记录、备注默认空、接诊医生+评估人
- 当日处方：数据来源切换为治疗方案
- 双人核对：用户查询SQL降级、手术/时间默认计算、消毒/人员默认值
- 透后评估：透后净重自动计算、凝血分级可选、透析事件是/否切换、血压/心率来源提示
- 后端：新增 `schedule_week_service.go`，`user_service.go` 列缺失降级查询