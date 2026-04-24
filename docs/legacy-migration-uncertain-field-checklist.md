# 老血透迁移待确认字段核查表（人工复核）

更新时间：2026-04-23  
来源文档：`docs/legacy-migration-session-summary-2026-04-21.md`

> 说明：以下仅保留“当前仍无法完全确定”的字段与内容。请在“人工确认结果”列填写最终口径，我将按确认结果继续改造并从待确认中移除。

| 编号 | 菜单名称 | 接口 | 相关表 | 待确认字段/内容 | 当前处理 | 需你确认 | 人工确认结果 |
|---|---|---|---|---|---|---|---|
| A-1 | 患者管理 -> 方案调整记录 | `POST /api/v1/patients/:id/adjustment-records` | `Plan_PatientPlanPrescriptionAdjustment`, `Plan_PatientPrescription` | `PatientPlanPrescriptionId` 关联规则 | 当前按“患者最新处方”写入 | 是否必须按“被调整方案”精准落到指定处方；前端是否会传方案ID/处方ID | 待填写 |
| A-2 | 患者管理 -> 方案调整记录 | `GET /api/v1/patients/:id/adjustment-records` | `Plan_PatientPlanPrescriptionAdjustment`, `Organ_Employee`, `Identity_Users` | 操作人展示口径 | 优先姓名，缺失回退用户名/ID | 是否允许回退 ID；还是必须强制显示姓名 | 待填写 |
| D-1 | 患者管理 -> 检验检查 | `GET /api/v1/patients/:id/key-indicators` | `LIS_Examination`, `LIS_ExaminationItem`, `LIS_ExaminationItem_Ret` | `indexName` 取值口径 | 当前优先 `RetExaminationName`，再回退其他名称 | 是否固定只用 `RetExaminationName` | 待填写 |
| D-2 | 患者管理 -> 检验检查 | `GET /api/v1/patients/:id/key-indicators` | `LIS_ExaminationItem`, `LIS_ExaminationItem_Ret` | `evaluationResult` 判定口径 | 当前按 `ResultSign` 映射偏高/偏低/正常 | 是否仅以 `ResultSign` 判定，是否需结合参考值范围 | 待填写 |
| E-1 | 透析执行 -> 透中监测 | `POST/PUT /api/v1/treatments/:id/during-params`, `GET /api/v1/patients/:id/treatment?date=...` | `Treatment_DuringParam`, `Treatment_DuringSigns` | `dialysateFlow` 老库承接字段 | 当前仅接口透传，未落老库显式列 | 是否有指定老库字段可落（表名+字段名） | 待填写 |
| E-2 | 透析执行 -> 透前评估/透后评估 | （当前无后端提交接口，仅前端展示） | `Auxiliary_JsonData` | `Action/FirstCheck/AgainCheck` 的 `Code/Value` 结构 | 已接 `hp_during_other`/`hp_treatment_details`/`hp_treatment_feelcontent` | 这三类核对环节的提交接口路径和 `Value` JSON 结构 | 待填写 |
| F-1 | 诊疗配置 -> 方案模板 | `GET /api/v1/treatment-templates`, `GET /api/v1/treatment-templates/:id` | `CodeDictionary_CodeDictionarys`, `Plan_PlanTPL` | `DialysisMethod` 字典口径 | 当前标准值+透传兼容 | 是否统一按 `Type='DialysisMethod'` 字典显示/保存 | 待填写 |
| F-2 | 诊疗配置 -> 方案模板 | `POST/PUT /api/v1/treatment-templates` | `Plan_PlanTPL` | `Note` JSON 扩展字段白名单 | 当前部分字段写入 `Note` 兼容 | 哪些字段允许继续进 `Note`，哪些必须映射到显式列 | 待填写 |
| F-3 | 诊疗配置 -> 医嘱模板 | `GET/POST/PUT/DELETE /api/v1/order-templates` | `Order_OrderTPL` | `OrderGroup` 主键语义；`type/priority/isDefault/sort` 落库语义 | 当前按 `OrderGroup` 聚合，`type` 暷存 `Note`，顺序用 `UseNum` | 是否以 `OrderGroup` 作为唯一模板键；其余字段对应列/字典 | 待填写 |
| G-1 | 患者管理 -> 月度评估小结 | （建议新增）`GET/PUT /api/v1/patients/:id/monthly-summaries` | `Treatment_TreatmentMonthSummarySheet` | 接口契约与 `ContentJsonb` 字段映射 | 当前前端Tab静态，后端未落地该接口 | 确认最终接口路径、请求/响应结构、`ContentJsonb` 键名映射 | 待填写 |
| DEV-1 | 设备管理 | 设备与床位绑定相关接口 | `Schedule_BedEquipmentRelChange` | 是否必须记录绑定变更历史 | 当前未强制写入变更表 | 是否要求每次变更写 `RelChange` | 待填写 |
| DEV-2 | 设备管理 | `GET/PUT /api/v1/devices/...` | `Schedule_BedEquipmentRel` | `ParameterS` 状态值字典 | 当前未锁定统一枚举 | 是否固定为 `normal/warning/alarm/offline/maintenance` | 待填写 |
| DEV-3 | 设备管理 | `GET/PUT /api/v1/devices/...` | `Auxiliary_EquipmentInfomation` | `Maintenance` 字段业务语义 | 当前按“运维信息”兼容处理 | 是否为“运维厂家ID”；若是请给关联主数据表 | 待填写 |

