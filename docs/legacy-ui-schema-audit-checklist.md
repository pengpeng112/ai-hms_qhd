# 前端界面与老血透库表结构核查清单（复核修订版）

复核日期：2026-06-01

核查范围：当前前端路由页面、子 Tab、弹窗、组件，以及对应前端 service、后端 handler/service/model 与老库表结构。

核查依据：`老血透数据库表结构-合并版.md`、`老数据库表设计.md`、`字典类型对照表(1).md`、`ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`、`docs/legacy-migration-uncertain-field-checklist.md`。

分组报告：`docs/audit-p0-patient.md`、`docs/audit-p0-schedule.md`、`docs/audit-p0-dialysis.md`、`docs/audit-p0-monitoring.md`、`docs/audit-p1-config.md`、`docs/audit-p2-auxiliary.md`。

## 复核修订说明

原汇总清单已经完成主体核查，但存在几处需要修订：

| 修订项 | 原表述 | 修订后口径 |
|---|---|---|
| 统计数量 | 直接汇总各 agent 统计数 | 各分组统计口径不完全一致，不再作为执行依据；以本清单的“确认问题”和“待确认项”为准 |
| 透中备注 | `Treatment_DuringParam.Note` 是否存在待确认 | 老库结构已确认 `Treatment_DuringParam` 无 `Note` 列，属于确认问题 |
| 透后并发症/症状 | `Treatment_AfterSigns.Complication/Symptoms` 是否存在待确认 | 老库结构已确认 `Treatment_AfterSigns` 无这两列，属于确认问题 |
| 病区感染类型 | 前端多了 `梅毒/HIV`，老库设计仅 `普通/乙肝/丙肝` | 不能只按设计文档判断；应以 `CodeDictionary_CodeDictionarys.Type='InfectionType'` 为准。当前问题是前端硬编码，未走老库字典，且缺少部分实际字典值 |
| 排班模板接口 | 后端无模板列表接口 | 更准确：前端调用 `/api/v1/schedule/template` 不存在；后端已有 `/api/v1/patient-shifts/templates`，但页面未对接 |
| 统计页面 | 4 个统计接口均使用新表模型 | 更准确：质量统计使用新表 `lab_report_items`；感染/通路/治疗模型部分 TableName 指向老库，但服务仍使用新 schema 字段/列名检查，结果可能为空或不准确 |
| 一致项 | 个别一致项与待确认项重复 | 改为“基本一致/核心字段一致”，不再把待确认字段列为完全一致 |

## 执行原则

- “确认问题”可以交给其他 AI 按计划改造。
- “待确认项”不要直接改造，先查库或人工确认最终口径。
- 禁止新增 DDL、AutoMigrate、seed、DropTables。
- 禁止把不确定字段猜测写入老库。
- 涉及字典字段时优先核对 `CodeDictionary_CodeDictionarys`。

## 一、确认问题清单（可进入改造）

| 编号 | 优先级 | 页面/模块 | 问题 | 当前依据 | 建议改造方向 |
|---|---|---|---|---|---|
| AUD-001 | P0 | 患者列表/详情 | `DryWeight` 映射到 `Register_PatientInfomation.Weight`，但老库 `Weight` 是体重，不是干体重 | `internal/models/patient.go:62`；老库 `Register_PatientInfomation.Weight`；老库 `Plan_PatientPlan.DryWeight` | 列表和详情干体重从最新有效 `Plan_PatientPlan.DryWeight` 读取；基本信息保存不要覆盖 `Register_PatientInfomation.Weight` |
| AUD-002 | P0 | 患者列表/详情 | `Diagnosis` 映射到 `Register_PatientInfomation.Note`，但老库 `Note` 是备注，不是诊断 | `internal/models/patient.go:61`；老库 `Register_Diagnosis.DiagnosisDesc` | 患者诊断从 `Register_Diagnosis.DiagnosisDesc` 读取，必要时 `Note` 只作为备注或 fallback 展示 |
| AUD-003 | P0 | 患者详情-基本信息 | 干体重保存写入 `Register_PatientInfomation.Weight`，会覆盖体重语义 | `patient_basic_service.go` 写入 Weight；老库 `Plan_PatientPlan.DryWeight` | 保存干体重时写 `Plan_PatientPlan.DryWeight`，如仍需记录当前体重则另设 `weight` 字段口径 |
| AUD-004 | P1 | 患者详情-概览 | 感染信息用 `InfectionDesc/OtherDesc/Note` 自由文本关键词解析，存在误判/漏判风险 | `patient_core_service.go` buildInfection；老库 `Register_Infection` 为文本描述 | 保持兼容读取，但需明确“非结构化解析”；如要结构化阳性/阴性，先确认老系统实际文本样本和关键词规则 |
| AUD-005 | P0 | 排班模板列表 | 前端调用 `/api/v1/schedule/template`，后端未注册该 GET 路由 | `ScheduleTemplateList.tsx`；`schedule_handler.go` 只有 `/patient-shifts/templates` 与 `/schedule/template/save/apply` | 前端模板列表改接 `GET /api/v1/patient-shifts/templates`，或后端补兼容 GET 路由，二选一 |
| AUD-006 | P0 | 排班模板 | 前端 `ScheduleTemplate` 类型与后端模板记录结构不一致 | `scheduleTemplate.ts`；后端模板来自 `Schedule_PatientShift.Status=60` | 统一模板 DTO；如果继续用 `Schedule_PatientShift.Status=60`，前端不要使用独立模板主表字段如 `cycleWeeks/isDefault/isEnabled` |
| AUD-007 | P0 | 应用排班模板 | 前端应用模板发送空 body `{}`，后端需要 `targetDate` | `ApplyTemplateModal.tsx`；`schedule_handler.go` ApplyTemplateHandler | 前端传当前周起始日期或用户选择日期作为 `targetDate` |
| AUD-008 | P0 | 保存排班模板 | 前端传 `weekday`，后端 `PatientShiftCreateRequest` 无该字段 | `restClient.ts`；`patient_shift_service.go` | 删除无效 `weekday` 或后端明确支持并转换到 `TreatmentTime` |
| AUD-009 | P0 | 排班模板编辑器 | `ScheduleTemplateEditor.tsx` 未实现 | `ScheduleTemplateEditor.tsx` | 实现模板编辑器，或暂时隐藏入口避免误用 |
| AUD-010 | P0 | 班次配置 | `Schedule_Shift.StartTime/EndTime` 老库为 `timestamp`，模型和前端按 `HH:mm` 字符串处理 | 老库 `Schedule_Shift.StartTime/EndTime`；`models/schedule.go:87-88` | 查询老库实际样本后统一转换策略；写库时避免直接把 `HH:mm` 当 timestamp 写入失败 |
| AUD-011 | P0 | 班次配置 | `Schedule_Shift.Type` 老库为 integer，前端不传，模型为 string | 老库 `Schedule_Shift.Type`；`models/schedule.go:89` | 如 Type 仍有业务意义，前端增加类型选择并按 10/20 保存；如废弃，后端设安全默认值 |
| AUD-012 | P0 | 排班创建 | `BedId/WardId/PatientPlanId/ShiftTiming` 老库非空，模型仍允许空 | 老库 `Schedule_PatientShift`；`models/schedule.go:121-124` | 创建/更新时强制校验非空，无法确定时阻止保存而不是写 NULL |
| AUD-013 | P0 | 排班历史弹窗 | 后端返回新状态值，前端用老状态值映射显示，状态显示可能错误 | `Schedule.tsx:799-802`；`patient_shift_service.go` 状态转换 | 前端历史弹窗改用新状态值映射，或后端同时返回 `statusName` |
| AUD-014 | P0 | 监控页面 | 生命体征、透中参数、超滤、趋势图全部是硬编码或空数组 | `monitoring/types.ts`；`StatusGrid.tsx`；`Monitoring.tsx` | 新增监控数据 API，从当前治疗记录关联 `Treatment_DuringSigns` + `Treatment_DuringParam` 取最新和趋势数据 |
| AUD-015 | P0 | 监控页面 | 页面只加载一次，无轮询/WebSocket/SSE，不符合监控页面实时性 | `useMonitoringData.ts` | 增加轮询或实时推送；优先短轮询，避免大改架构 |
| AUD-016 | P0 | 病区管理 | `patientType` 前端硬编码中文值，老库实际保存代码值（已知测试库存在 `10`/空） | `WardManagement.tsx:151-155`；老库 `Schedule_Ward.PatientType` | 改用字典/枚举 code 保存，显示层再翻译名称 |
| AUD-017 | P1 | 病区管理 | `infectionType` 前端硬编码，未走老库字典；现有选项缺少已知字典值如 `结核`，是否需要 `普通/空` 也未统一 | `WardManagement.tsx:159-165`；`CodeDictionary_CodeDictionarys.Type='InfectionType'` | 下拉改从老库字典 `InfectionType` 加载，保存字典值；不要在前端硬编码 |
| AUD-018 | P0 | 病区管理 | `bedCount` DTO 有字段但列表未统计，前端可能显示 0 或空 | `ward_service.go:75-106` 未 join `Schedule_Bed` | List 查询聚合 `Schedule_Bed` 统计，或单独查询后回填 |
| AUD-019 | P0 | 病区/床位/宣教管理 | 删除为物理删除，老库表已有 `IsDisabled` | `ward_service.go:196-200`；`bed_service.go:279-290`；`health_education_service.go:306-310` | 改为软删除 `IsDisabled=true`，保留历史引用关系 |
| AUD-020 | P0 | 透中监测 | 后端写 `Treatment_DuringParam.Note`，但老库 `Treatment_DuringParam` 已确认无 `Note` 列 | 老库 `Treatment_DuringParam` 字段 22 列；`treatment_service.go:1745` | 不再写 `Treatment_DuringParam.Note`；备注写入 `Auxiliary_JsonData` 或确认其他承接字段 |
| AUD-021 | P0 | 透后评估 | 后端写 `Treatment_AfterSigns.Complication/Symptoms`，但老库 `Treatment_AfterSigns` 已确认无这两列 | 老库 `Treatment_AfterSigns` 19 列；`treatment_service.go:1969-1970` | 不再写不存在列；并发症/症状写 `Treatment_AfterSymptom` 或 `Auxiliary_JsonData`，按确认 schema 固定 |
| AUD-022 | P0 | 透析执行-独立子页 | `FirstCheck.tsx`、`SecondCheck.tsx`、`Disinfection.tsx` 是占位页，实际表单在 `Verification.tsx`，独立入口不闭环 | 对应文件仅 10 行；后端 API 已存在 | 二选一：实现独立子页，或移除/重定向独立入口到 `Verification.tsx` |
| AUD-023 | P0 | 透析执行-耗材记录 | `Consumables.tsx` 前后端均未实现 | `Consumables.tsx` 占位；无对应 API | 先确认老库承接表，再设计 API；不可猜表写入 |
| AUD-024 | P0 | 患者详情-月份小结 | `MonthlySummaryTab` 纯前端静态 UI，无读写 API；老库有 `Treatment_TreatmentMonthSummarySheet` | `MonthlySummaryTab.tsx`；老库表存在 | 实现月度小结读写 API，对接 `Treatment_TreatmentMonthSummarySheet` |
| AUD-025 | P1 | 患者详情-待办任务 | `TodoPopover` 使用硬编码 mock 数据 | `TodoPopover.tsx` | 接入真实待办/任务 API，或隐藏模拟入口 |
| AUD-026 | P0 | 库存管理 | 使用新表 `inventory_items/stock_logs`，未对接老库 `Stock_*` 表 | `inventory_service.go`；老库 `Stock_Stock/Stock_InOutStorage/...` | 切换到老库库存表，标签任务如老库无对应可保留新功能但需标记 |
| AUD-027 | P0 | 统计页面 | 质量统计使用新表 `lab_report_items`；其他统计部分虽模型 TableName 指向老库，但仍使用新 schema 字段/列名检查，可能返回空或不准 | `statistics_service.go`；`models/lab_report.go`；`models/patient.go`；`models/treatment.go` | 改为显式 SQL 查询老库 `LIS_*`、`Register_*`、`Treatment_*` 真实字段，不用新 schema 字段名判断 |
| AUD-028 | P0 | 用户管理 | 重置密码直接写入明文到 `Identity_Users.PasswordHash` | `user_service.go:232-252` | 使用 ASP.NET Identity V3 PasswordHash 生成逻辑写入 |
| AUD-029 | P0 | 角色管理 | 角色管理使用 `Authorization_Roles`，登录/用户使用 `Identity_Roles`，角色体系不统一 | `permission_service.go`；`auth_service.go`；`user_service.go` | 确认目标角色体系；如沿用老库，角色管理应对接 `Identity_Roles/Identity_UserRoles` |
| AUD-030 | P1 | Dashboard | 在档患者数未过滤 `TreatmentStatus`，可能包含出院/死亡患者 | `dashboard_service.go:75-80` | 按业务确认过滤条件，一般应过滤在科/在院且未禁用患者 |
| AUD-031 | P1 | WardOverview | 平均透析时长硬编码 3.8h，床位矩阵非真实床位状态 | `WardOverview.tsx` | 平均时长查 `Treatment_Treatment.RealDuration`；床位矩阵查 `Schedule_Bed`/排班/治疗状态 |
| AUD-032 | P1 | 设备管理 | `Notes` 映射为病区名而不是设备备注 | `device_service.go` 分组报告证据；老库 `Auxiliary_EquipmentInfomation.Note` | 备注字段映射到设备 `Note`，病区名单独字段展示 |

## 二、待确认项（确认前不要改造）

| 编号 | 优先级 | 模块 | 待确认内容 | 建议确认方式 |
|---|---|---|---|---|
| TODO-001 | P0 | 患者基本信息 | `Register_PatientInfomation.Gender` 实际存储 `M/F` 还是 `男/女` | 查 `SELECT DISTINCT "Gender" FROM "Register_PatientInfomation"` |
| TODO-002 | P0 | 患者列表/详情 | `TreatmentStatus` 实际枚举值及“在档患者”的过滤口径 | 查 `SELECT DISTINCT "TreatmentStatus" FROM "Register_PatientInfomation"` 并人工确认在档口径 |
| TODO-003 | P0 | 患者医嘱 | `Order_PatientOrder.Type/Classification` 实际枚举值及前端长期/临时映射 | 查 distinct 值 |
| TODO-004 | P1 | 患者证件 | 前端 `ID_TYPE` 字典 code 与 `Register_IDInfomation.IDType` 是否一致 | 对比字典接口与老库 distinct |
| TODO-005 | P1 | 血管通路 | `Artery/Venous` 是单值、逗号分隔还是多值字典 | 抽样 `Register_VascularAccess` |
| TODO-006 | P0 | 班次配置 | `Schedule_Shift.StartTime/EndTime` 实际样本如何保存，是否只有日期占位 + 时间 | 抽样查询班次表 |
| TODO-007 | P0 | 班次配置 | `Schedule_Shift.Type` 是否仍有业务意义，默认值应是什么 | 查 distinct 值并与老系统界面确认 |
| TODO-008 | P1 | 排班 | 单双周算法是否按 ISO 周号奇偶判断 | 对照老系统排班源码/文档/样本 |
| TODO-009 | P0 | 排班模板 | `Schedule_PatientShift.Status=60` 用作模板是否会与“转出”语义冲突 | 人工确认最终模板承接口径 |
| TODO-010 | P0 | 透中监测 | `dialysateFlow` 是否有其他老库承接字段；已确认 `Treatment_DuringParam` 无显式列 | 查老系统写入逻辑或实际 `Auxiliary_JsonData` 样本 |
| TODO-011 | P0 | 透析执行 JSON | `hp_before_symptom/hp_after_symptom/hp_during_other/hp_again_check` 的标准 JSON schema | 查历史数据样本并固定 schema |
| TODO-012 | P1 | 透前/透后症状 | 前端新定义 symptomItem code 是否与老系统 code 一致 | 查 `Treatment_BeforeSymptom/Treatment_AfterSymptom` 历史 code |
| TODO-013 | P1 | 透析执行 | 透后 StartTime/EndTime 是否允许直接覆盖已有值 | 人工确认流程权限和修正规则 |
| TODO-014 | P0 | 设备/监控 | `Schedule_BedEquipmentRel.ParameterS` jsonb 原始用途，是否可用作设备状态 | 查历史样本和老系统代码 |
| TODO-015 | P1 | 诊疗配置 | 方案模板透析方式是否必须支持 `HD+HP` | 查字典 `DialysisMethod` 和老系统模板样本 |
| TODO-016 | P1 | 诊疗配置 | `Order_OrderTPL.Type` 应如何保存，是否可从文本推断 | 查老库 distinct 和老系统模板保存逻辑 |
| TODO-017 | P1 | 材料目录 | `standardType` 是否应双写 `StdCat` 和 `Type` | 确认两个字段业务语义 |
| TODO-018 | P1 | 药品目录 | `genericName/concentration/specUnit` 分别落到哪一列 | 查老库字段和老系统界面 |
| TODO-019 | P1 | 字典配置 | 新表 `parent_code` 树形结构是否需要同步到老库 | 人工确认字典层级需求 |
| TODO-020 | P1 | 用户管理 | 删除用户是否允许物理删除，用户类型字段存储到哪里 | 人工确认用户体系 |
| TODO-021 | P1 | 设置 | HDIS 配置应使用新表还是老库 `Applications_AppSetting` | 人工确认配置存储口径 |
| TODO-022 | P2 | 日志 | 系统日志使用文件日志还是老库 `Log_Logs` | 人工确认审计日志需求 |

## 三、基本一致项（执行时只需回归验证）

| 模块 | 已基本一致的内容 | 注意事项 |
|---|---|---|
| 患者基本信息 | 姓名、拼音、生日、民族、ABO/RH、身高、文化程度、职业、婚姻、单位、电话、微信、座机、地址、医保号、透析号、首次透析时间/医院等 | `Gender`、`IDType`、`dryWeight` 仍按上方问题处理 |
| 住院/证件/联系人 | `Register_Hospitalization`、`Register_IDInfomation`、`Register_FamilyMember` 聚合读写基本正确 | 字典值仍需对齐老库 |
| 血管通路 | 主要字段映射到 `Register_VascularAccess` | `Artery/Venous` 多值格式需确认 |
| 检验报告 | 检验报告详情读取 `LIS_Examination/LIS_ExaminationItem` 的核心字段 | 统计页面另有问题，不等同于检验报告 Tab |
| 转归记录 | `Register_OutCome.Type/Reason/OutComeTime/Note` 映射基本正确 | 无 |
| 透前体征 | `Treatment_BeforeSigns` 核心体征字段一致 | symptomItem code 需确认 |
| 透中核心参数 | `BF/TMP/UFQuantity/VenousPressure/ArterialPressure/Conductivity/MachineTmp` 映射一致 | `dialysateFlow`、`notes` 不在该表显式列 |
| 透后核心体征 | `Treatment_AfterSigns` 核心体征字段一致 | `Complication/Symptoms` 不在该表显式列 |
| 透析小结 | `Treatment_Treatment.NurseSummary/TreatmentSummary` 读写基本一致 | 无 |
| 健康宣教 | `Auxiliary_HealthEducation` 与 `Auxiliary_PatientHealthEducation` 主流程基本闭环 | 删除应软删除，附件字段未完善 |
| 透析医嘱 | `Order_PatientOrder` 主流程基本闭环 | 类型/分类枚举需确认 |
| 今日处方 | `Plan_PatientPrescription` 主流程基本闭环 | 无 |
| 排班核心 | 周视图、创建、更新、取消、换床、拖拽、互换核心逻辑基本正确 | 模板、班次类型/时间、非空约束仍需修复 |
| 方案/医嘱模板 | `Plan_PlanTPL`、`Plan_PlanTPLMaterial`、`Order_OrderTPL` 主流程基本对接 | `HD+HP`、Order Type、AllDosage 等需确认 |
| 材料/药品目录 | 主字段基本对接 `Auxiliary_MaterialInfomation/Auxiliary_DrugInfomation` | 部分扩展字段待确认 |
| 登录认证 | 登录校验 ASP.NET Identity V3 Hash 基本正确 | 重置密码仍错误 |
| 用户列表 | 列表 JOIN `Identity_Users/Organ_Employee/Identity_UserRoles/Identity_Roles` 基本正确 | 创建/重置/删除仍需修复 |

## 四、建议执行顺序

1. 第一批：修复明确的数据破坏风险。包括 AUD-001、AUD-002、AUD-003、AUD-020、AUD-021、AUD-028。
2. 第二批：修复临床闭环阻塞。包括 AUD-005 到 AUD-015、AUD-022 到 AUD-024。
3. 第三批：修复主数据一致性。包括 AUD-016 到 AUD-019、AUD-026、AUD-027、AUD-029。
4. 第四批：完善 P1/P2 体验和统计展示。包括 AUD-025、AUD-030 到 AUD-032。
5. 待确认项 TODO-001 到 TODO-022 必须确认后再纳入改造，不得猜测落库。

## 五、给执行 AI 的提示词

```text
你要按照 docs/legacy-ui-schema-audit-checklist.md 执行改造。

只改“确认问题清单”里的 AUD 项。不要改 TODO 项，除非用户已经明确给出确认结果。

硬性限制：
- 不允许 AutoMigrate、DDL、seed、DropTables。
- 不允许新增新表替代老库表。
- 老库表字段大小写敏感，SQL 使用双引号。
- 字段含义不确定时停止并记录，不要猜。
- 每次只改一个模块，改完运行对应后端/前端验证。

执行顺序：
1. 先修数据破坏风险：AUD-001、AUD-002、AUD-003、AUD-020、AUD-021、AUD-028。
2. 再修临床闭环：AUD-005 到 AUD-015、AUD-022 到 AUD-024。
3. 再修主数据一致性：AUD-016 到 AUD-019、AUD-026、AUD-027、AUD-029。
4. 最后处理 P1/P2：AUD-025、AUD-030 到 AUD-032。

每个 AUD 项完成后，在提交说明中写清楚：前端字段、API、后端 service、老库表字段、验证命令结果。
```
