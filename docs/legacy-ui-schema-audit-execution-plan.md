# 前端界面与老血透库表结构核查执行计划

生成日期：2026-05-31

## 目标

按当前前端每个页面/弹窗/Tab/表单字段，逐项比对老库 `老血透数据库表结构-合并版.md`，核查前端展示、接口传输、后端存储表字段、字段类型、枚举/字典、默认值、关联关系是否准确一致。

本轮只做核查，不做改造。最终只输出一个问题清单，供人工确认后再进入代码改造。

## 核查依据

- 老库权威结构：`老血透数据库表结构-合并版.md`
- 老库业务补充：`老数据库表设计.md`
- 字典参考：`字典类型对照表(1).md`
- 已有映射记录：`ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`
- 待确认字段：`docs/legacy-migration-uncertain-field-checklist.md`
- 前端页面：`ai-hms-frontend/src/pages/**`
- 前端服务：`ai-hms-frontend/src/services/**`
- 后端路由：`ai-hms-backend/cmd/server/main.go`、`ai-hms-backend/internal/api/v1/**`
- 后端服务：`ai-hms-backend/internal/services/**`
- 后端模型：`ai-hms-backend/internal/models/**`

## 硬性规则

- 不修改任何代码、表结构、配置或数据。
- 不新增 AutoMigrate、seed、DDL 或新表依赖。
- 后端只能以老库真实表字段为准。
- 表名/字段名大小写敏感，核查 SQL/代码时关注双引号和 GORM `column` 标签。
- 如果字段含义不确定，只登记到清单，不猜测。
- 如果前端字段暂存在 JSON/Note/Value 中，必须标明 JSON key、承接字段、类型风险。
- 如果使用字典，必须核对 `CodeDictionary_CodeDictionarys.Type/Code/Name/Sort/IsDisabled`。

## 总体核查流程

1. 页面入口梳理：从 `ai-hms-frontend/src/router.tsx` 和菜单配置列出所有路由、页面、Tab、弹窗。
2. 前端字段抽取：对每个页面列出表格列、详情展示字段、筛选项、表单输入项、下拉选项、按钮触发的提交字段。
3. 前端接口定位：追踪页面调用的 `restClient.ts`、`patientApi.ts`、`orderApi.ts`、`managementApi.ts` 等服务方法，记录请求路径、请求体、响应字段。
4. 后端链路定位：追踪对应 handler/service/model/SQL，记录真实读取/写入的表、字段、where 条件、join、排序、状态过滤。
5. 老库结构比对：用 `老血透数据库表结构-合并版.md` 核对表名、字段名、物理类型、非空、默认值、业务含义、枚举/字典来源。
6. 一致性判断：比较前端字段名/含义、接口 DTO、后端存储字段、老库字段类型和内容是否一致。
7. 数据闭环判断：区分只读、可编辑、新增、修改、删除/禁用、批量、导入、模板应用等场景，分别核查落库与回显。
8. 清单输出：只输出问题清单，不输出改造代码。

## 每个字段必须核查的内容

- 前端字段：显示名称、组件类型、是否必填、默认值、单位、格式化规则。
- API 字段：请求字段、响应字段、TypeScript 类型、空值处理。
- 后端字段：DTO 字段、服务字段、GORM/SQL 字段、事务处理、错误处理。
- 老库字段：表名、字段名、类型、是否非空、默认值、业务含义。
- 字典/枚举：前端选项是否来自字典，后端是否保存 Code 还是 Name，老库是否同口径。
- ID 类型：前端 string/number 与老库 bigint 是否一致，是否存在字符串化或精度风险。
- 时间类型：日期/时间/时区/格式，是否保存为 timestamp，是否覆盖已有时间。
- 状态字段：启用/禁用、草稿/正式/取消、执行状态等是否与老库值一致。
- 关联关系：PatientId、TreatmentId、ScheduleId、WardId、BedId、EquipmentId、OperatorId 等是否准确关联。
- JSON 承接：若写入 `Auxiliary_JsonData.Value`、`Note`、`ContentJsonb`，必须记录 Code/key/schema。

## 页面分组与优先级

### P0 直接影响临床闭环

| 页面/路由 | 前端入口 | 重点老库表 |
|---|---|---|
| 患者列表 `/patients` | `PatientList.tsx` | `Register_PatientInfomation`, `Register_Infection`, `Schedule_PatientShift`, `Plan_PatientPlan` |
| 患者详情 `/patients/:id` | `PatientDetail.tsx`, `patient-detail/tabs/**` | `Register_*`, `Plan_*`, `Order_PatientOrder`, `LIS_*`, `Treatment_TreatmentMonthSummarySheet` |
| 排班 `/schedule` | `Schedule.tsx` | `Schedule_PatientShift`, `Schedule_Shift`, `Schedule_Bed`, `Schedule_Ward`, `Schedule_BedEquipmentRel` |
| 透析执行 `/dialysis-processing` | `DialysisProcessing.tsx`, `dialysis-processing/execution/**` | `Treatment_Treatment`, `Treatment_BeforeSigns`, `Treatment_DuringParam`, `Treatment_DuringSigns`, `Treatment_AfterSigns`, `Treatment_Action`, `Order_PatientOrder`, `Auxiliary_JsonData`, `Auxiliary_PatientHealthEducation` |
| 监控 `/monitoring` | `Monitoring.tsx`, `monitoring/**` | `Treatment_Treatment`, `Treatment_DuringParam`, `Schedule_*`, `Device_*`, `Auxiliary_EquipmentInfomation` |

### P1 影响配置和主数据

| 页面/路由 | 前端入口 | 重点老库表 |
|---|---|---|
| 诊疗配置 `/treatment-config` | `TreatmentConfig.tsx`, `TreatmentConfig/tabs/**` | `Plan_PlanTPL`, `Order_OrderTPL`, `Order_Material`, `Order_Drug`, `CodeDictionary_CodeDictionarys` |
| 病区管理 `/ward-management` | `WardManagement.tsx` | `Schedule_Ward`, `Schedule_Bed`, 用户/员工表 |
| 床位管理 `/bed-management` | `BedManagement.tsx` | `Schedule_Bed`, `Schedule_Ward`, `Schedule_BedEquipmentRel`, `Schedule_BedEquipmentRelChange`, `Auxiliary_EquipmentInfomation` |
| 班次配置 `/shift-config` | `ShiftConfig.tsx` | `Schedule_Shift` |
| 设备绑定 `/device-binding` | `DeviceManagement.tsx` | `Auxiliary_EquipmentInfomation`, `Schedule_BedEquipmentRel`, `Schedule_BedEquipmentRelChange`, `Device_*` |
| 字典配置 `/dict-config` | `DictConfig.tsx` | `CodeDictionary_CodeDictionarys` |
| 宣教管理 `/education-management` | `EducationManagement.tsx` | `Auxiliary_HealthEducation`, `Auxiliary_PatientHealthEducation` |

### P2 统计、权限、辅助页面

| 页面/路由 | 前端入口 | 重点老库表 |
|---|---|---|
| 首页 `/dashboard` | `Dashboard.tsx`, `dashboard/cards/**` | `Register_*`, `Schedule_*`, `Treatment_*`, `Device_*`, `Order_*` |
| 病区概览 `/ward-overview` | `WardOverview.tsx` | `Schedule_Ward`, `Schedule_Bed`, `Treatment_Treatment`, `Schedule_PatientShift` |
| 统计 `/statistics` | `Statistics.tsx` | `Treatment_*`, `Register_*`, `Schedule_*`, `LIS_*` |
| 库存 `/inventory` | `Inventory.tsx` | `Stock_*`, `Order_Material`, `Order_Drug` |
| 基础资料 `/master-data` | `MasterData.tsx` | `Auxiliary_*`, `CodeDictionary_CodeDictionarys` |
| 用户管理 `/user-management` | `UserManagement.tsx` | `Identity_Users`, `Organ_Employee` |
| 角色管理 `/role-management` | `RoleManagement.tsx` | 权限/角色相关 legacy 或当前表，需核实 |
| 登录/角色选择 | `Login.tsx`, `RoleSelect.tsx` | `Identity_Users`, `Organ_Employee` |
| 设置 `/settings` | `Settings.tsx` | 应用配置、租户配置、集成配置相关表 |
| 排班模板 | `ScheduleTemplateList.tsx`, `ScheduleTemplateEditor.tsx` | `Schedule_PatientShift` 状态模板记录、或老系统模板表需核实 |

## 单页面核查步骤模板

对每个页面按以下顺序执行：

1. 读取页面文件，列出所有用户可见字段。
2. 读取页面使用的 hooks/components/types，补齐弹窗和子组件字段。
3. 追踪服务层方法，记录接口路径、请求方法、请求/响应类型。
4. 追踪后端路由、handler、service，记录真实 SQL/GORM 表字段。
5. 到 `老血透数据库表结构-合并版.md` 查对应表字段。
6. 对每个字段判断：一致、不一致、缺失、只读未闭环、待确认、不适用。
7. 对每个保存动作判断：新增、修改、删除/禁用、批量、导入、模板应用是否真实落老库。
8. 对每个列表/详情判断：是否真实从老库读，是否存在 mock、静态数据、fallback、错误字段。
9. 对每个下拉判断：是否来自 `CodeDictionary_CodeDictionarys` 或明确静态枚举。
10. 输出问题清单，不输出修复代码。

## 输出清单格式

最终只生成一个 Markdown 表格，建议命名：`docs/legacy-ui-schema-audit-checklist.md`。

| 编号 | 优先级 | 页面/模块 | 前端文件 | 功能点 | 前端字段/操作 | API | 后端文件 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| AUD-001 | P0 | 示例 | 示例文件 | 示例功能 | 示例字段 | 示例接口 | 示例服务 | 当前字段 | 老库字段 | 类型/枚举/内容 | 不一致/缺失/待确认 | 文件:行号 | 只写方向，不写代码 | 是/否 |

结论枚举：

- 一致：前端、接口、后端、老库字段语义和类型均匹配。
- 不一致：字段名、类型、值域、字典、状态、单位、格式、关联关系任一不匹配。
- 缺失：前端有字段但后端/老库未存，或老库关键字段前端未展示/未保存。
- 只读未闭环：能展示但不能保存，或能保存但不能回显。
- 待确认：老库字段含义、字典值、JSON schema、关联关系无法从文档确认。
- 不适用：页面为纯导航、占位或不涉及老库数据。

## 典型风险检查点

- 前端字段名沿用新系统，但后端实际老库字段不同。
- 前端保存的是中文 Name，老库/字典要求保存 Code。
- 前端 number/string 与老库 bigint/numeric/timestamp 不一致。
- 患者 ID、治疗 ID、排班 ID 混用。
- `TenantId`、`IsDisabled`、`Status` 条件遗漏导致脏数据或历史数据混入。
- 编辑操作走了删除重建，导致老库主键/关联记录漂移。
- 使用 mock、静态数组、fallback 数据但页面看起来正常。
- 新增/编辑接口写入当前新表，而不是老库表。
- 老库字段不存在却被 DTO 暴露或前端填写。
- JSON 字段没有固定 schema，导致回显或统计不可控。
- 老库时间字段被默认 `now()` 覆盖已有值。
- 床位、设备、病区绑定缺少变更历史。

## 给其他 AI 的统一提示词

```text
你现在只做核查，不做任何代码改造、格式化、提交、数据库写入或 DDL。

任务：按当前前端界面逐个页面/Tab/弹窗/表单，与老血透数据库结构进行字段级比对，最终只生成一个问题清单。

必须阅读并以这些文件为依据：
- 老库权威结构：老血透数据库表结构-合并版.md
- 老库业务补充：老数据库表设计.md
- 字典参考：字典类型对照表(1).md
- 现有映射记录：ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md
- 待确认字段：docs/legacy-migration-uncertain-field-checklist.md
- 执行计划：docs/legacy-ui-schema-audit-execution-plan.md

核查范围：
- 前端路由：ai-hms-frontend/src/router.tsx
- 前端页面：ai-hms-frontend/src/pages/**
- 前端服务：ai-hms-frontend/src/services/**
- 后端路由/handler/service/model：ai-hms-backend/cmd/server/main.go、ai-hms-backend/internal/api/v1/**、ai-hms-backend/internal/services/**、ai-hms-backend/internal/models/**

逐页面核查内容：
1. 列出页面所有表格列、详情字段、筛选项、表单字段、弹窗字段、按钮保存动作。
2. 追踪每个字段对应的前端 service/API 请求和响应字段。
3. 追踪后端 handler/service/model/SQL，确认真实读取/写入的表和字段。
4. 用老血透数据库表结构-合并版.md 核对表名、字段名、物理类型、非空、默认值、业务含义。
5. 核查字段类型、字典/枚举、单位、时间格式、ID 类型、状态值、TenantId/IsDisabled 过滤、关联字段是否一致。
6. 对新增、编辑、删除/禁用、批量、导入、模板应用分别核查是否真实落老库并能回显。
7. 发现不确定字段只标记“待确认”，不要猜测。

最终输出要求：
- 只输出一个 Markdown 清单，不要写改造代码。
- 建议输出文件：docs/legacy-ui-schema-audit-checklist.md
- 表格列必须包含：编号、优先级、页面/模块、前端文件、功能点、前端字段/操作、API、后端文件、当前表字段、老库标准表字段、类型/内容核查、结论、证据、建议改造方向、需人工确认。
- 结论只能使用：一致、不一致、缺失、只读未闭环、待确认、不适用。
- 对“一致”的项目可以简写；重点详细列出“不一致/缺失/只读未闭环/待确认”。
- 每条问题必须给出文件路径和行号证据，至少包含前端文件和后端或老库文档证据。
- 不要修改代码，不要运行会写库的接口，不要执行迁移，不要创建测试数据。
```
