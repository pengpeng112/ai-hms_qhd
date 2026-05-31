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

## 三、后端 API 缺失（审查发现前端调用路径与后端路由已对齐）

> **重要更新（2026-05-31）**: 经核查，以下所有模块的后端 handler/service/route 均已实现并注册，
> 之前的"缺失"判断是因为 handler 文件不在预期位置（user_handler.go/role_management_handler.go/health_education_handler.go）。

| 模块 | 状态 | 路由文件 |
|------|------|----------|
| 用户管理 CRUD（9个端点） | ✅ | `user_handler.go:202-217` |
| 角色管理 CRUD（5个端点） | ✅ | `role_management_handler.go:95-102` |
| 健康宣教（3个端点） | ✅ | `health_education_handler.go:59-65` |
| 消毒登记 | ✅ | `treatment_handler.go:466` |
| 治疗小结 | ✅ | `treatment_handler.go:467` |

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
| 4 | 硬编码占位 | `DialysisSummary.tsx` | 凝血分级已改为从 symptomItems 读取真实数据 | 🟡中 | ✅ |
| 5 | 硬编码占位 | `MidMonitoring.tsx:336-337` | "机位状态"、"同步间隔" 仍硬编码 | 🟡中 | ❌ |
| 6 | 硬编码占位 | `PostAssessment.tsx` | 内瘘护理 checkbox 已改为可交互 | 🟡中 | ✅ |
| 7 | 按钮禁用 | `PreAssessment.tsx` | "暂存草稿" 按钮已移除 | 🟡中 | ✅ |
| 8 | 按钮禁用 | `Verification.tsx` | 消毒登记按钮已启用（后端已实现） | 🟡中 | ✅ |
| 9 | 按钮禁用 | `DialysisSummary.tsx` | 保存小结按钮已启用（后端已实现） | 🟡中 | ✅ |
| 10 | 占位提示 | `MedicalOrders.tsx` | "从模板组调取" 仍显示占位提示 | 🟡中 | ❌ |
| 11 | 抗凝剂表 | `MedicalOrders.tsx:410` | 抗凝剂表格仍显示"暂无数据" | 🟡中 | ❌ |

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
4. ✅ 用户列表 SQL 查询降级（已修复）
5. ✅ 后端健康宣教模块（已实现，handler 位于 `health_education_handler.go`）

### P1 — 高优（系统核心功能无法使用）
6. ✅ 后端用户管理 CRUD（已实现，handler 位于 `user_handler.go`）
7. ✅ 后端角色管理 CRUD（已实现，handler 位于 `role_management_handler.go`）
8. ❌ `ScheduleTemplateEditor.tsx` — 空白页面，需实现排班模板编辑

### P2 — 中等（功能受限但不阻断）
9. ✅ 治疗消毒登记（已实现，`treatment_handler.go:466`）
10. ✅ 治疗小结保存（已实现，`treatment_handler.go:467`）
11. ❌ 医嘱模板调取 — `MedicalOrders.tsx` 仍提示占位
12. ❌ 抗凝剂表 — `MedicalOrders.tsx:410` 显示"暂无数据"
13. ❌ PostAssessment 下机图片按钮 disabled
14. ❌ `ScheduleTemplateEditor.tsx` 空白
15. ❌ `EducationManagement.tsx` 增删改缺失
16. ❌ `MasterData.tsx` 宣教/收费 Tab 硬编码

### P3 — 低优（代码质量与工程改进）
17. ❌ GraphQL 残留服务清理
18. ❌ 重复 API 函数合并
19. ❌ 类型定义冗余

### ✅ 本次会话已修复
- 患者管理：默认在科患者 + 统计数据显示（新增 `/patients/stats` 接口）
- 病区概览：切换为 Dashboard Stats 真实数据
- 患者详情：右侧面板改为悬浮按钮+抽屉/弹窗
- 透前评估：目标超滤量自动计算、A/V端位点下拉、内瘘默认标签、皮肤记录、备注默认空、接诊医生+评估人
- 当日处方：数据来源切换为治疗方案
- 双人核对：用户查询SQL降级、手术/时间默认计算、消毒/人员默认值
- 透后评估：透后净重自动计算、凝血分级可选、透析事件是/否切换、血压/心率来源提示
- 后端：新增 `schedule_week_service.go`，`user_service.go` 列缺失降级查询