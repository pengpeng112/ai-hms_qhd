# 老血透迁移会话核查（2026-04-21，按人工标记更新）

## 2026-04-22 再核查结论

- 本轮按“人工标记规则”复核后，文档中每条待确认项已包含：`菜单名称 + 接口 + 关键字段`。
- 本轮无新增可“直接删除的待确认项”（未引入主观假设前提下）。
- 已确认并完成改造的事项仍保留在“已完成项”章节，待确认事项继续保留在“待确认项”章节，便于你逐条回执后继续收敛。
- 待确认项已同步整理为表格版清单：`docs/legacy-migration-uncertain-field-checklist.md`（便于逐条人工核查）。

本文档已根据你在原“待确认”中的人工标记做了二次核查。  
规则如下：

- 已能明确且已完成代码改造的：从“待确认”移除，记录到“已完成项”
- 仍有不确定性或依赖字典/业务口径确认的：保留在“待确认项”
- 每条都补充：菜单名称、接口、字段，便于你再次核实

---
## 2026-04-22 实库连通核查（10.20.1.153:5432 / dialysis）

### 连接与本地复用
- 已验证可连通：`select now()` 成功。
- 已写入本地连接文件（git 已忽略）：`/.env.legacy.local`。

### 患者 300410 关键表实际数据量（用于映射核查）
- `Register_PatientInfomation`: `1`
- `Plan_PatientPlan`: `2`
- `Plan_PatientPrescription`: `3`
- `Order_PatientOrder`: `0`
- `Order_PatientDayOrder`（关联长期医嘱）: `0`
- `Register_MedicalHistory`: `0`
- `Register_OutCome`: `1`
- `Register_VascularAccess`: `1`
- `Treatment_Treatment`: `3`
- `LIS_Examination`: `23`

### 月度评估相关实表
- 已确认存在：`Treatment_TreatmentMonthSummarySheet`
- 字段含：`PatientId/Year/Month/ContentJsonb/ImageBase64String/...`

---

## 一、已完成项（已从待确认移除）

### 1. 治疗方式来源改造
- 菜单：`患者管理 -> 治疗详情历史 / 治疗记录单`
- 接口：
  - `GET /api/v1/treatments`
  - `GET /api/v1/treatments/:id`
  - `GET /api/v1/patients/:id/treatment?date=YYYY-MM-DD`
- 改造：
  - `treatmentType` 优先从 `Plan_PatientPrescription.DialysisMethod` 读取
  - 通过 `Plan_PatientPrescription.TreatmentId -> Treatment_Treatment.Id` 关联
  - 仅在处方中无值时才回退原推断逻辑
- 字段：
  - `Plan_PatientPrescription.TreatmentId`
  - `Plan_PatientPrescription.DialysisMethod`
  - `Treatment_Treatment.Id`
- 代码：
  - `ai-hms-backend/internal/services/treatment_service.go`

### 2. 血管通路图片主键字段处理
- 菜单：`患者管理 -> 血管通路评估`
- 接口：
  - `POST /api/v1/patients/:id/vascular-accesses`
  - `PUT /api/v1/patients/:id/vascular-accesses/:accessId`
- 改造：
  - 停止将图片首图 ID 回写到 `Register_VascularAccess.PictureIds`
  - 图片读取/展示完全以 `Register_VascularAccessImage` 为准（按 `Sort` 升序）
- 字段：
  - `Register_VascularAccessImage.VascularAccessId`
  - `Register_VascularAccessImage.Sort`
  - `Register_VascularAccess.PictureIds`（不再写入）
- 代码：
  - `ai-hms-backend/internal/services/vascular_access_service.go`

### 3. 检查报告来源确认
- 菜单：`患者管理 -> 检验检查`
- 接口：
  - `GET /api/v1/patients/:id/exam-reports`
- 处理结果：
  - 按人工标记“无对应老表”，保持空列表策略，不再按“待确认 PACS 表”推进
- 字段：无（老血透库无对应本地表）

### 4. 血管通路数组字段存储格式确认
- 菜单：`患者管理 -> 血管通路评估`
- 接口：
  - `GET /api/v1/patients/:id/vascular-accesses`
  - `POST /api/v1/patients/:id/vascular-accesses`
  - `PUT /api/v1/patients/:id/vascular-accesses/:accessId`
- 处理结果：
  - 以逗号分隔字符串为主口径（如 `1,2,3`）
  - 读取保持兼容 JSON 数组，避免历史数据回归问题
- 字段：
  - `Register_VascularAccess.Artery`
  - `Register_VascularAccess.Venous`
  - `Register_VascularAccess.ASidePointCount`
  - `Register_VascularAccess.VSidePointCount`

### 5. 处方状态码与创建状态口径
- 菜单：`患者管理 -> 处方`
- 接口：
  - `GET /api/v1/patients/:id/prescriptions`
  - `POST /api/v1/patients/:id/prescriptions`
  - `PUT /api/v1/patients/:id/prescriptions/:prescriptionId`
- 改造：
  - `Plan_PatientPrescription.Status` 按已确认口径调整为：`1/0 -> 待执行`、`2 -> 已执行`、`3 -> 已取消`
  - 新建处方默认写入状态改为 `1`（草稿/待执行）
- 字段：
  - `Plan_PatientPrescription.Status`
- 代码：
  - `ai-hms-backend/internal/services/prescription_service.go`

### 6. 处方日期按治疗表时间优先
- 菜单：`患者管理 -> 处方`
- 接口：
  - `GET /api/v1/patients/:id/prescriptions`
  - `GET /api/v1/patients/:id/prescriptions/:prescriptionId`
- 改造：
  - 当 `Plan_PatientPrescription.TreatmentId > 0` 时，优先使用 `Treatment_Treatment.StartTime/ReceptionTime/CreateTime` 作为处方日期
  - 仅在未关联治疗记录时，才回退 `Note` 扩展日期与 `CreateTime`
- 字段：
  - `Plan_PatientPrescription.TreatmentId`
  - `Treatment_Treatment.StartTime`
  - `Treatment_Treatment.ReceptionTime`
  - `Treatment_Treatment.CreateTime`
- 代码：
  - `ai-hms-backend/internal/services/prescription_service.go`

### 7. 医嘱状态来源改为当日医嘱优先
- 菜单：`患者管理 -> 医嘱`
- 接口：
  - `GET /api/v1/patients/:id/orders`
- 改造：
  - 优先读取 `Order_PatientDayOrder.Status`（按同 `PatientOrderId` 最新记录）
  - 优先读取字典 `CodeDictionary_CodeDictionarys.Type='PatientDayOrderStatus'` 进行状态语义映射
  - 字典或表缺失时自动回退原 `Order_PatientOrder` 推断逻辑，避免中断
- 字段：
  - `Order_PatientOrder.Id`
  - `Order_PatientDayOrder.PatientOrderId`
  - `Order_PatientDayOrder.Status`
  - `CodeDictionary_CodeDictionarys.Type/Code/Name`
- 代码：
  - `ai-hms-backend/internal/services/order_service.go`

### 8. 临床病史输血史切换到 Auxiliary_JsonData
- 菜单：`患者管理 -> 临床病史`
- 接口：
  - `GET /api/v1/patients/:id/medical-history`
  - `PUT /api/v1/patients/:id/medical-history`
- 改造：
  - 输血史读写已接入 `Auxiliary_JsonData`，固定 `Code='hp_bloodtransfusion_history'`
  - `GET` 优先读取 `Auxiliary_JsonData.Value`，为空时回退 `Register_MedicalHistory.PersonalHistory`
  - `PUT` 写入 `Auxiliary_JsonData` 同时保留旧字段兼容写入，避免历史页面回归
- 字段：
  - `Auxiliary_JsonData.PatientId/Code/Value`
  - `Register_MedicalHistory.PersonalHistory`（兼容回退）
- 代码：
  - `ai-hms-backend/internal/services/medical_history_service.go`

### 9. 透中生命体征落库到 Treatment_DuringSigns
- 菜单：`透析执行 -> 透中监测`
- 接口：
  - `GET /api/v1/patients/:id/treatment?date=...`
  - `POST /api/v1/treatments/:id/during-params`
  - `PUT /api/v1/treatments/:id/during-params/:paramId`
- 改造：
  - `during-params` 新增接收并持久化：`sbp/dbp/heartRate/respiration/spO2` 到 `Treatment_DuringSigns`
  - 查询 `duringParams` 时按 `OperateTime` 将 `Treatment_DuringSigns` 生命体征回填到 `Treatment_DuringParam` 结果
  - 机参仍以 `Treatment_DuringParam` 为主（`dialysateFlow` 当前仍无老库显式列承接）
- 字段：
  - `Treatment_DuringSigns.SBP/DBP/HeartRate/Respiration/SpO2/BodyTemp`
  - `Treatment_DuringParam.OperateTime/*`
- 代码：
  - `ai-hms-backend/internal/services/treatment_service.go`

### 10. 透前/透后扩展字段接入 Auxiliary_JsonData（双写兼容）
- 菜单：`透析执行 -> 透前评估 / 透后评估`
- 接口：
  - `PUT /api/v1/treatments/:id/before-signs`
  - `PUT /api/v1/treatments/:id/after-signs`
  - `GET /api/v1/patients/:id/treatment?date=...`
- 改造：
  - 透前保存时写入 `Auxiliary_JsonData.Code='hp_before_symptom'`（保存请求快照 JSON）
  - 透后保存时写入 `Auxiliary_JsonData.Code='hp_after_symptom'`（保存请求快照 JSON）
  - 保留原 `Treatment_BeforeSymptom / Treatment_AfterSymptom` 写入，避免历史流程回归
  - 治疗详情读取时，当 `Treatment_AfterSigns` 字段为空，会从 `hp_after_symptom` 解析 `complication/symptoms/notes` 回填
- 字段：
  - `Auxiliary_JsonData.PatientId/TreatmentId/Code/Value`
  - `Treatment_AfterSigns.Complication/Symptoms/Note`（读取回退）
  - `Treatment_BeforeSymptom`、`Treatment_AfterSymptom`（兼容保留）
- 代码：
  - `ai-hms-backend/internal/services/treatment_service.go`

### 11. 透前/透后扩展字段回显补齐（治疗详情接口）
- 菜单：`透析执行 -> 透前评估 / 透后评估`
- 接口：
  - `GET /api/v1/patients/:id/treatment?date=...`
  - `GET /api/v1/treatments/:id`
- 改造：
  - 接口新增回显字段：`beforeSymptomItems`、`afterSymptomItems`
  - 读取顺序：优先 `Treatment_BeforeSymptom/Treatment_AfterSymptom`，若为空则回退解析 `Auxiliary_JsonData` (`hp_before_symptom` / `hp_after_symptom`)
  - 透后 `complications` 也增加 JsonData 回退解析（`complication/symptoms/notes`）
- 字段：
  - `Treatment_BeforeSymptom.Code/Value`
  - `Treatment_AfterSymptom.Code/Value`
  - `Auxiliary_JsonData.Code/Value`
- 代码：
  - `ai-hms-backend/internal/services/treatment_service.go`

### 12. 治疗状态接入字典映射（兼容回退）
- 菜单：`透析执行 -> 治疗详情历史`
- 接口：
  - `GET /api/v1/treatments`
  - `GET /api/v1/treatments/:id`
  - `GET /api/v1/patients/:id/treatment?date=...`
- 改造：
  - 已接入 `CodeDictionary_CodeDictionarys.Type='Treatment_TreatmentStatus'` 进行状态语义映射
  - 字典缺失或无法识别时，自动回退原 `0/10/60/90/100` 兼容映射
- 字段：
  - `Treatment_Treatment.Status`
  - `CodeDictionary_CodeDictionarys.Type/Code/Name`
- 代码：
  - `ai-hms-backend/internal/services/treatment_service.go`

### 13. 新增 JsonData code 接入（透中其他项/治疗详情/感受内容）
- 菜单：`透析执行 -> 透中监测 / 透后评估`
- 接口：
  - `POST/PUT /api/v1/treatments/:id/during-params`
  - `PUT /api/v1/treatments/:id/after-signs`
  - `GET /api/v1/patients/:id/treatment?date=...`
- 改造：
  - 透中监测写入 `Auxiliary_JsonData.Code='hp_during_other'`（保存当前透中项快照）
  - 透后保存写入：
    - `hp_treatment_details`（并发症/症状/症状项/备注）
    - `hp_treatment_feelcontent`（透后感受内容，当前映射 `notes`）
  - 读取治疗详情时，上述 code 会参与透后并发症/备注回退合并
- 字段：
  - `Auxiliary_JsonData.Code/Value`
- 代码：
  - `ai-hms-backend/internal/services/treatment_service.go`

### 14. 检验检查关键指标切换到老库 LIS 表（实库验证后）
- 菜单：`患者管理 -> 检验检查`
- 接口：
  - `GET /api/v1/patients/:id/key-indicators`
- 改造：
  - 查询来源由新库 `patient_key_indicators` 改为老库：
    - `LIS_Examination`
    - `LIS_ExaminationItem`
    - `LIS_ExaminationItem_Ret`（重要指标配置）
  - 按 `ItemCode/ItemName` 与配置表匹配后返回，保持前端原响应结构不变。
- 字段：
  - `LIS_Examination.PatientId/ResultTime/TestNO`
  - `LIS_ExaminationItem.ExaminationId/ItemCode/ItemName/Result/Unit/Reference/ResultSign`
  - `LIS_ExaminationItem_Ret.ItemCode/ItemName/RetItemName/RetExaminationName/ExaminationName`
- 代码：
  - `ai-hms-backend/internal/services/key_indicator_service.go`

### 15. 方案调整记录操作人字段兼容修复（实库列差异）
- 菜单：`患者管理 -> 方案调整记录`
- 接口：
  - `GET /api/v1/patients/:id/adjustment-records`
  - `POST /api/v1/patients/:id/adjustment-records`
- 改造：
  - 针对 `Organ_Employee` 在不同库可能缺少 `UserId/TenantId` 的情况，增加多分支回退查询：
    - `UserId + TenantId`
    - `UserId`
    - `Id + TenantId`
    - `Id`
  - 消除 `column "UserId" does not exist` 类型报错。
- 字段（实库确认）：
  - `Plan_PatientPlanPrescriptionAdjustment` 含：`AdjustUserId`（不含 `UserId`）
  - `Organ_Employee` 实库仅有：`Id`, `Name`
- 代码：
  - `ai-hms-backend/internal/services/prescription_service.go`

### 16. 全息视图档案查询列名修正（老库字段大小写）
- 菜单：`患者管理 -> 全息视图档案`
- 接口：
  - `GET /api/v1/patients/:id/core`
- 改造：
  - `buildInfection`、`buildCurrentPlan` 查询条件从新库风格 `patient_id` 改为老库列：
    - `"PatientId"` + `"TenantId"`
  - 避免因列名不匹配导致的“无数据/降级空块”。
- 字段：
  - `Register_Infection.PatientId/TenantId`
  - `Plan_PatientPlan.PatientId/TenantId`
- 代码：
  - `ai-hms-backend/internal/services/patient_core_service.go`

### 17. 方案调整记录按人工标注收敛
- 菜单：`患者管理 -> 方案调整记录`
- 接口：
  - `POST /api/v1/patients/:id/adjustment-records`
  - `GET /api/v1/patients/:id/adjustment-records`
- 改造：
  - 新增可选入参 `patientPlanPrescriptionId`，传入时按该处方精确绑定到 `Plan_PatientPlanPrescriptionAdjustment.PatientPlanPrescriptionId`
  - 未传时保留“患者最新处方”回退
  - 操作人名称缺失时显示 `--`（按人工标注）
- 代码：
  - `ai-hms-backend/internal/services/patient_service.go`

### 18. 关键指标定义来源切换（重要指标）
- 菜单：`患者管理 -> 检验检查`
- 接口：
  - `GET /api/v1/patients/:id/key-indicators`
- 改造：
  - 关键指标定义优先从 `LIS_ExaminationItem_Config` 获取（过滤 `RetExaminationName='重要指标'`，使用 `RetItemName`）
  - 当配置表为空时，回退 `LIS_ExaminationItem_Ret`
  - `evaluationResult` 继续基于 `LIS_ExaminationItem.ResultSign(H/L)` 判定
- 代码：
  - `ai-hms-backend/internal/services/key_indicator_service.go`

---

## 二、待确认项（保留）


### A. 患者管理 -> 方案调整记录
1. 前端是否传 `patientPlanPrescriptionId`  
- 接口：`POST /api/v1/patients/:id/adjustment-records`  
- 当前：后端已支持可选入参 `patientPlanPrescriptionId`，传入时按该处方精确关联；未传时回退患者最新处方。  
- 已按人工标注改造：
  - 调整记录支持精确绑定 `Plan_PatientPrescription.Id`
  - 操作人缺失展示统一为 `--`
- 仍需确认：
  - 前端是否在“修改处方”动作中统一传 `patientPlanPrescriptionId`（用于彻底关闭“回退最新处方”分支）
- 关键字段：
  - `Plan_PatientPlanPrescriptionAdjustment.PatientPlanPrescriptionId`
  - `Plan_PatientPrescription.Id`
  - `Plan_PatientPlanPrescriptionAdjustment.AdjustUserId`
  - `Organ_Employee.Name`

### B. 患者管理 -> 处方与医嘱
（已在“已完成项”第 5、6、7 条落地）

### C. 患者管理 -> 临床病史
（已在“已完成项”第 8 条落地）

### D. 患者管理 -> 检验检查
（已按人工标注完成并移至“已完成项补充”）

### E. 治疗执行页
（已在“已完成项”第 12 条落地）

9. 透中监测 `dialysateFlow` 老库字段承接  
- 菜单：`透析执行 -> 透中监测`
- 接口：
  - `POST/PUT /api/v1/treatments/:id/during-params`
  - `GET /api/v1/patients/:id/treatment?date=...`
- 当前：生命体征已落 `Treatment_DuringSigns`，`dialysateFlow` 仍仅 API 层透传
- 需确认：
  - 是否有明确老字段可承接 `dialysateFlow`（当前未在 `Treatment_DuringParam`/`Treatment_DuringSigns` 找到直连列）
- 关键字段：
  - `Treatment_DuringParam.*`
  - `Treatment_DuringSigns.*`

10. 核对环节 code（Action / FirstCheck / AgainCheck）接入  
- 菜单：`透析执行 -> 透前评估 / 透后评估`
- 接口：
  - 当前页面 `check1/check2` 无后端提交接口（仅前端展示）
- 当前：`hp_during_other` / `hp_treatment_details` / `hp_treatment_feelcontent` 已接入
- 需确认：
  - `Action / FirstCheck / AgainCheck` 的提交接口与 Value 结构
  
  人工标注：Action：{"20": 200000, "40": 200000, "60": 200000, "80": 200000, "100": 200000, "110": 200000, "120": 200000, "130": 200000, "140": 300162} key为治疗状态码，值为操作人Id
  FirstCheck：[{"name": "透析方式", "Result": true, "checked": true, "content": "HFD", "OperatorId": 200007, "OperateTime": "2025-10-13T02:20:09.971Z"}, {"name": "透析器", "Result": true, "checked": true, "content": "FB-17U", "OperatorId": 200007, "OperateTime": "2025-10-13T02:20:09.971Z"}, {"name": "血路管", "Result": true, "checked": true, "OperatorId": 200007, "OperateTime": "2025-10-13T02:20:09.971Z"}, {"name": "预冲", "Result": true, "checked": true, "OperatorId": 200007, "OperateTime": "2025-10-13T02:20:09.971Z"}, {"name": "处方内容核对", "Result": true, "checked": true, "OperatorId": 200007, "OperateTime": "2025-10-13T02:20:09.971Z"}, {"name": "患者身份识别", "Result": true, "checked": true, "OperatorId": 200007, "OperateTime": "2025-10-13T02:20:09.971Z"}]

  AgainCheck：[{"name": "透析方式", "Result": true, "checked": true, "content": "HFD", "OperatorId": 200000, "OperateTime": "2025-07-31T00:40:02.356Z"}, {"name": "处方内容核对", "Result": true, "checked": true, "OperatorId": 200000, "OperateTime": "2025-07-31T00:40:02.356Z"}, {"name": "抗凝剂使用", "Result": true, "checked": true, "OperatorId": 200000, "OperateTime": "2025-07-31T00:40:02.356Z"}, {"name": "血管通路核查", "Result": true, "checked": true, "OperatorId": 200000, "OperateTime": "2025-07-31T00:40:02.356Z"}, {"name": "管路连接核查", "Result": true, "checked": true, "OperatorId": 200000, "OperateTime": "2025-07-31T00:40:02.356Z"}]


- 关键字段：
  - `Auxiliary_JsonData.Code`
  - `Auxiliary_JsonData.Value`
  - `Treatment_Action.Code/Name/OperatorId/OperateTime`

当前补充说明：
- 已完成：
  - `PUT /api/v1/treatments/:id/first-check` 保存时同步写 `Treatment_Action(Code=70, Name=首次核对)`
  - `PUT /api/v1/treatments/:id/second-check` 保存时同步写 `Treatment_Action(Code=150, Name=二次核对)`
- 仍待确认：
  - `FirstCheck/AgainCheck` 是否必须按“逐行数组结构”固化（当前仍为可回显的聚合结构）

（已在“已完成项”第 11 条落地）

### F. 诊疗配置
12. `DialysisMethod` 字典来源（你标注字典表）  
- 菜单：`诊疗配置 -> 方案模板`
- 接口：
  - `GET /api/v1/treatment-templates`
  - `GET /api/v1/treatment-templates/:id`
- 当前：标准值+保底透传
- 需确认：
  - 是否统一按 `CodeDictionary_CodeDictionarys.Type = 'DialysisMethod'` 显示和保存
- 关键字段：
  - `CodeDictionary_CodeDictionarys.Type/Code/Name`
  - `Plan_PlanTPL.DialysisMethod`

13. `Plan_PlanTPL.Note` 扩展字段清单  
- 菜单：`诊疗配置 -> 方案模板`
- 接口：
  - `POST/PUT /api/v1/treatment-templates`
- 当前：部分前端字段写入 `Note` JSON 兼容
- 需确认：
  - 允许写入 `Note` 的字段白名单
  - 哪些字段必须改成老库显式列
  
  人工标注：不允许写入 `Note` JSON 兼容，哪些字段在老库中没有对应，请列出清单。
- 关键字段：
  - `Plan_PlanTPL.Note`

14. 医嘱模板字段语义  
- 菜单：`诊疗配置 -> 医嘱模板`
- 接口：
  - `GET/POST/PUT/DELETE /api/v1/order-templates`
- 当前：按 `OrderGroup` 聚合，`type` 暂存 `Note`，顺序暂用 `UseNum`
- 需确认：
  - `OrderGroup` 是否唯一模板键
  - `type/priority/isDefault/sort` 的真实落库列或字典

  人工标注：
  - OrderGroup为组合医嘱的组合Id

  - 顺序不能用UseNum,UseNum为使用数量

  - type/priority/isDefault/sort` 的真实落库列或字典 医嘱模板有类型、优先级、默认的需求吗？ Sort可以考虑增加

- 关键字段：
  - `Order_OrderTPL.OrderGroup`
  - `Order_OrderTPL.Note`
  - `Order_OrderTPL.UseNum`

### G. 患者管理 -> 月度评估小结
15. 月度小结接口与字段映射尚未落地
- 菜单：`患者管理 -> 月度评估小结`
- 接口：
  - （当前前端 Tab 仅静态展示，未发现已接入后端 API）
- 实库现状：
  - 已发现目标表：`Treatment_TreatmentMonthSummarySheet`
  - 关键列：`PatientId/Year/Month/ContentJsonb/ImageBase64String`
- 需确认：
  - 前端期望接口路径（建议新增 `GET/PUT /api/v1/patients/:id/monthly-summaries`）
  - `ContentJsonb` 字段内 JSON 结构与页面字段的一一映射

人工标注：ContentJsonb结构如下：

{"": {"sort": 20, "血压": {"sort": 30}, "透中": {"sort": 50, "widget": "radio", "default": "正常", "options": [{"label": "高", "value": "高"}, {"label": "正常", "value": "正常"}, {"label": "低", "value": "低"}]}, "透后": {"sort": 60, "widget": "radio", "default": "正常", "options": [{"label": "高", "value": "高"}, {"label": "正常", "value": "正常"}, {"label": "低", "value": "低"}]}, "干体重": {"sort": 0}, "水肿部位": {"mode": "default", "sort": 90, "widget": "select", "default": "", "options": [{"label": "颜面", "value": "颜面"}, {"label": "下肢", "value": "下肢"}, {"label": "心包", "value": "心包"}, {"label": "胸腔", "value": "胸腔"}, {"label": "腹腔", "value": "腹腔"}]}, "透析间期": {"sort": 40, "widget": "radio", "default": "正常", "options": [{"label": "高", "value": "高"}, {"label": "正常", "value": "正常"}, {"label": "低", "value": "低"}]}, "治疗依从性": {"sort": 20, "widget": "radio", "default": "一般", "options": [{"label": "一般", "value": "一般"}, {"label": "差", "value": "差"}]}, "透析中并发症": {"mode": "tags", "sort": 70, "widget": "select", "default": ["无"], "options": [{"label": "低血压", "value": "低血压"}, {"label": "肌肉痉挛", "value": "肌肉痉挛"}, {"label": "低血糖", "value": "低血糖"}, {"label": "心律失常", "value": "心律失常"}, {"label": "心绞痛", "value": "心绞痛"}, {"label": "心肌梗塞", "value": "心肌梗塞"}, {"label": "肺栓塞", "value": "肺栓塞"}, {"label": "透析器反应", "value": "透析器反应"}, {"label": "致热源反应", "value": "致热源反应"}, {"label": "失衡综合征", "value": "失衡综合征"}, {"label": "无", "value": "无"}, {"label": "其他", "value": "其他"}]}, "透析间期水肿": {"sort": 80, "widget": "radio", "default": "无", "options": [{"label": "有", "value": "有"}, {"label": "无", "value": "无"}]}, "透析间平均体重增加": {"sort": 10, "widget": "radio", "default": "<5kg", "options": [{"label": ">5kg", "value": ">5kg"}, {"label": "<5kg", "value": "<5kg"}]}}, "CTR": {"": {"sort": 0, "unit": "%", "widget": "number", "default": ""}, "sort": 30, "备注": {"sort": 10, "widget": "input", "default": ""}}, "其他": {"sort": 60, "住院": {"sort": 0, "widget": "radio", "default": "否", "options": [{"label": "是", "value": "是"}, {"label": "否", "value": "否"}]}, "转归": {"mode": "default", "sort": 30, "widget": "select", "default": "", "options": [{"label": "好转", "value": "好转"}, {"label": "恶化", "value": "恶化"}, {"label": "转院", "value": "转院"}, {"label": "死亡", "value": "死亡"}]}, "住院日期": {"sort": 10, "widget": "input", "default": ""}, "出院日期": {"sort": 20, "widget": "input", "default": ""}, "主要就诊原因": {"sort": 40, "widget": "input", "default": ""}}, "贫血": {"Hb": {"sort": 0, "unit": "g/L", "widget": "number", "default": ""}, "sort": 50}, "骨病": {"P": {"sort": 20, "unit": "mmol/L", "widget": "number", "default": ""}, "Ca": {"sort": 10, "unit": "mmol/L", "widget": "number", "default": ""}, "iPTH": {"sort": 0, "unit": "pg/mL", "widget": "number", "default": ""}, "sort": 50}, "一般情况": {"sort": 0, "尿量": {"sort": 50, "unit": "ml/日", "widget": "radio", "default": "少尿", "options": [{"label": "无尿", "value": "无尿"}, {"label": "少尿", "value": "少尿"}]}, "服药": {"sort": 60, "widget": "radio", "default": "遵医嘱", "options": [{"label": "遵医嘱", "value": "遵医嘱"}, {"label": "不遵医嘱", "value": "不遵医嘱"}]}, "睡眠": {"sort": 10, "widget": "radio", "default": "一般", "options": [{"label": "好", "value": "好"}, {"label": "一般", "value": "一般"}, {"label": "差", "value": "差"}]}, "营养": {"sort": 40, "widget": "radio", "default": "一般", "options": [{"label": "好", "value": "好"}, {"label": "一般", "value": "一般"}, {"label": "差", "value": "差"}]}, "血压": {"sort": 70, "widget": "radio", "default": "自测", "options": [{"label": "自测", "value": "自测"}, {"label": "规律", "value": "规律"}, {"label": "偶", "value": "偶"}, {"label": "未自测", "value": "未自测"}]}, "血糖": {"sort": 80, "widget": "radio", "default": "未自测", "options": [{"label": "自测", "value": "自测"}, {"label": "规律", "value": "规律"}, {"label": "偶", "value": "偶"}, {"label": "未自测", "value": "未自测"}]}, "饮食": {"sort": 30, "widget": "radio", "default": "一般", "options": [{"label": "好", "value": "好"}, {"label": "一般", "value": "一般"}, {"label": "差", "value": "差"}]}, "生活自理": {"sort": 0, "widget": "radio", "default": "正常", "options": [{"label": "正常", "value": "正常"}, {"label": "依赖", "value": "依赖"}, {"label": "部分", "value": "部分"}, {"label": "完全", "value": "完全"}]}}, "急诊透析": {"": {"sort": 0, "widget": "radio", "default": "无", "options": [{"label": "有", "value": "有"}, {"label": "无", "value": "无"}]}, "sort": 70, "原因": {"sort": 10, "widget": "radio", "default": "其他", "options": [{"label": "高钾血症", "value": "高钾血症"}, {"label": "心力衰竭", "value": "心力衰竭"}, {"label": "其他", "value": "其他"}]}}, "血透情况": {"sort": 10, "平均血流速": {"sort": 0, "unit": "ml/min", "widget": "number", "default": 200}}, "透析充分性": {"": {"sort": 20, "widget": "radio", "default": "", "options": [{"label": "充分", "value": "充分"}, {"label": "不充分", "value": "不充分"}]}, "URR": {"sort": 0, "unit": "%", "widget": "number"}, "Kt/V": {"sort": 10, "unit": "%", "widget": "number", "default": ""}, "sort": 40}, "本阶段透析总评价以及治疗建议": {"": {"sort": 0, "widget": "textarea", "default": "继续规律血液透析治疗，及时调整治疗方案。"}, "sort": 80}}


  
- 关键字段：
  - `Treatment_TreatmentMonthSummarySheet.ContentJsonb`
  - `Treatment_TreatmentMonthSummarySheet.ImageBase64String`

---

## 三、设备管理（本轮新增）

### 已完成
- 菜单：`设备管理`
- 接口：
  - `GET /api/v1/devices`
  - `GET /api/v1/devices/:id`
  - `POST /api/v1/devices`
  - `PUT /api/v1/devices/:id`
  - `DELETE /api/v1/devices/:id`
  - `PUT /api/v1/devices/:id/status`
  - `GET /api/v1/devices/:id/usage-logs`
  - `GET /api/v1/devices/:id/maintenance-records`
  - `GET /api/v1/devices/:id/disinfections`
- 表：
  - `Auxiliary_EquipmentInfomation`
  - `Schedule_BedEquipmentRel`
  - `Schedule_Bed`
  - `Schedule_Ward`
  - `Auxiliary_EquipmentUsageLog`
  - `Auxiliary_EquipmentMaintenance`
  - `Auxiliary_EquipmentDisinfection`

### 待确认
- `Schedule_BedEquipmentRelChange` 是否要求在设备绑定变更时同步写入变更历史
- `Schedule_BedEquipmentRel.ParameterS` 状态值字典是否固定（normal/warning/alarm/offline/maintenance）
- `Auxiliary_EquipmentInfomation.Maintenance` 语义是否是运维厂家 ID（及其对应主数据表）

人工标注：
- `Schedule_BedEquipmentRelChange`设备绑定变更时同步写入变更历史

- `Schedule_BedEquipmentRel.ParameterS` 老系统暂时没有使用，jsonb类型

- `Auxiliary_EquipmentInfomation.Maintenance` 从CodeDictionary_CodeDictionarys表获得，Type字段等于EquipmentInfomationMaintenance，Code为Id，Name为显示名称

---



## 2026-04-23 本轮补充：透析执行-首次核对（check1/check2）

### 已完成改造
- 菜单：`透析执行 -> 首次核对 / 二次核对`
- 接口：
  - 新增 `PUT /api/v1/treatments/:id/first-check`
  - 既有 `GET /api/v1/treatments/:id`、`GET /api/v1/patients/:id/treatment?date=...` 回传新增 `firstCheck` 快照
- 表映射：
  - `新逻辑 first-check` -> `Treatment_BeforeCheck`
  - 前端保持原 `check1/check2` 页面与路径不变，后端做兼容
- 本轮字段落库：
  - `Treatment_BeforeCheck.BeforeSignsId/BeforeSymptomId/OperatorId/OperateTime`
  - `Treatment_BeforeCheck.MaterialsResult/MaterialsMistake`
  - `Treatment_BeforeCheck.ParamResult/ParamMistake`
  - `Treatment_BeforeCheck.VascularAccessResult/VascularAccessMistake`
  - `Treatment_BeforeCheck.PipelineResult/PipelineMistake`

### 待确认（保留）
1. 页面/模块名称：`透析执行 -> 首次核对 / 二次核对`
- API：`PUT /api/v1/treatments/:id/first-check`
- 相关表：`Treatment_BeforeCheck`
- 相关字段：
  - `MaterialsResult/MaterialsMistake`
  - `ParamResult/ParamMistake`
  - `VascularAccessResult/VascularAccessMistake`
  - `PipelineResult/PipelineMistake`
  - `OperatorId/OperateTime`
- 当前判断：
  - 已按“最小侵入”将前端 check1/check2 的核对结果聚合映射到以上 4 组字段并可持久化。
- 不明确点：
  - check1/check2 中每一行（透析方式、透析器、血路管、预冲、患者身份识别、抗凝剂等）与上述 4 组字段是否需要固定一一映射，还是允许聚合映射。
  - 二次核对区域中的 `再核对护士`、`质控护士` 在老库无直接列，是否需要落其他表（例如 `Treatment_Action` 或 `Auxiliary_JsonData`）。
- 待确认项：
  - 请确认“首次/二次核对每个 UI 字段 -> 老库具体列”的最终口径。
  - 请确认 `OperatorId` 应取 `上机护士` 还是其它角色。
- 临时处理：
  - 当前实现为可运行方案：按 check1/check2 分组聚合后写入 `Treatment_BeforeCheck`；
  - 未明确的一对多字段映射保持在本清单，待你人工确认后再精确收敛。

---

## 四、本次核查后建议回归接口

- 患者方案与处方：
  - `/api/v1/patients/:id/treatment-plans`
  - `/api/v1/patients/:id/prescriptions`
  - `/api/v1/patients/:id/adjustment-records`
- 医嘱：
  - `/api/v1/patients/:id/orders`
- 治疗执行：
  - `/api/v1/treatments?patientId=...`
  - `/api/v1/treatments/:id`
  - `/api/v1/patients/:id/treatment?date=YYYY-MM-DD`
- 诊疗配置：
  - `/api/v1/treatment-templates?pageSize=9999`
  - `/api/v1/order-templates?pageSize=9999`
- 设备管理：
  - `/api/v1/devices?pageSize=20`
  - `/api/v1/devices/:id/maintenance-records?pageSize=20`
  - `/api/v1/devices/:id/usage-logs?pageSize=20`

---

## 2026-04-23 透析执行 -> 透前评估（本轮补充）

### 已完成
- 菜单：透析执行 -> 透前评估
- 接口：
  - `GET /api/v1/patients/:id/treatment?date=YYYY-MM-DD`
  - `PUT /api/v1/treatments/:id/before-signs`
- 老库映射（保持前端原接口契约不变）：
  - `Treatment_BeforeSigns` -> 返回新增 `beforeSigns`（weight/extraWeight/sbp/dbp/heartRate/respiration/temperature/pressurePoint/notes/operateTime）
  - `Treatment_BeforeSymptom` + `Auxiliary_JsonData(Code='hp_before_symptom')` -> 返回 `beforeSymptomItems`，并参与 `beforeSigns.pressurePoint/symptoms` 回填
- 前端页面改造：
  - 透前评估页从 `currentTreatment.beforeSigns + beforeSymptomItems + startBp` 动态回填，不再使用固定演示值。
  - 保持提交仍走 `PUT /before-signs` 原结构，不改前端对外行为。

### 待确认（不确定项）
- 菜单：透析执行 -> 透前评估
- 字段：测压部位 `PressurePoint`
  - 当前页面选项仅有“左上臂/右上臂”，而老库设计备注包含上下肢更多枚举。
  - 待确认是否需要扩充到完整老库枚举。
- 字段：上机护士/评估人（`on_machine_nurse` / `assessor`）
  - `beforeSymptomItems` 中历史值可能是姓名或ID；当前下拉按用户ID取值。
  - 待确认标准存储口径（统一ID或姓名）后再做严格归一化。
- 字段：神志/内瘘情况可视项
  - 页面存在展示项，但老库 `Treatment_BeforeSigns` 无对应显式列，当前仍通过 symptomItems 或前端固定展示。
  - 待确认这些字段是否必须统一落到 `Treatment_BeforeSymptom` 的固定 code 清单。

---

## 2026-04-23 透析执行 -> 当日处方（本轮补充）

### 已完成
- 菜单：透析执行 -> 当日处方
- 接口：
  - `GET /api/v1/patients/:id/prescriptions`
  - `GET /api/v1/patients/:id/treatment?date=YYYY-MM-DD`
- 老库落表：
  - 处方主表：`Plan_PatientPrescription`
  - 处方耗材：`Plan_PatientPrescriptionMaterial`
- 改造点：
  - 页面不再仅用静态演示值，已接入当日处方数据回填。
  - 当日处方优先按 `TreatmentId` 关联当天治疗记录，次选按 `prescriptionDate=当天`，再回退最新一条。
  - 处方材料表改为优先展示当日处方的 `materials`。
  - 后端 DTO 新增 `treatmentId`（仅响应字段，不改变前端既有入参契约）。

### 待确认（不确定项）
- 菜单：透析执行 -> 当日处方
- 字段：`血管通路` 展示值
  - 当前 `Plan_PatientPrescription` 可直接用的是 `VascularAccessId`（ID），页面需显示名称。
  - 待确认是否统一通过哪张字典/主数据映射（如 AccessType 或其他通路字典）。
- 字段：`上次透后体重`
  - 当前页面缺少稳定且单一来源的上一疗程透后体重字段（需跨治疗记录回溯）。
  - 待确认口径：是否取上一条 `Treatment_AfterSigns.Weight`。
- 字段：`肝素类型`
  - 老库显式字段为首剂/维持药及剂量，页面“普通/相对/绝对”属于业务枚举派生。
  - 待确认是否有固定映射规则，或仅按药物是否为空进行展示。

---

## 2026-04-23 本轮补充：二次核对 + 透析医嘱（透析执行页）

### 已完成
- 菜单：`透析执行 -> 二次核对(check2)`
- 接口：
  - 新增 `PUT /api/v1/treatments/:id/second-check`
  - 既有 `GET /api/v1/treatments/:id`、`GET /api/v1/patients/:id/treatment?date=...` 返回新增 `secondCheck`
- 表映射：
  - 二次核对动作：`Treatment_Action`（`Code='again_check'`, `Name='二次核对'`）
  - 二次核对明细快照：`Auxiliary_JsonData`（`Code='hp_again_check'`）
- 字段：
  - `operatorId / operateTime / recheckNurseId / qcNurseId`
  - `paramResult / paramMistake`
  - `vascularAccessResult / vascularAccessMistake`
  - `pipelineResult / pipelineMistake`
  - 行级细分：`dialysisMode* / prescription* / anticoagulant* / lineConnection*`
- 前端改造：
  - `check2` 从 `first-check` 改为调用 `second-check`
  - `check2` 回显改用 `currentTreatment.secondCheck`

- 菜单：`透析执行 -> 透析医嘱(orders_process)`
- 接口：
  - `GET /api/v1/patients/:id/orders`
  - `POST /api/v1/patients/:id/orders`
- 表映射：
  - 创建医嘱主表：`Order_PatientOrder`
  - 同步创建当日医嘱：`Order_PatientDayOrder`
- 前端改造：
  - 透析医嘱表格从静态假数据改为真实接口加载
  - 新增弹窗改为真实提交（类型/内容/备注）

### 待确认（保留）
1. 菜单：`透析执行 -> 二次核对(check2)`
- 接口：`PUT /api/v1/treatments/:id/second-check`
- 相关表：`Treatment_Action`, `Auxiliary_JsonData`
- 相关字段：
  - `Treatment_Action.Code/Name/OperatorId/OperateTime`
  - `Auxiliary_JsonData.Code='hp_again_check'`, `Value`
- 当前判断：
  - 老库无独立 `AgainCheck` 明细表，采用 `Action + JsonData` 组合承载。
- 不明确点：
  - `recheckNurse/qcNurse` 是否应落到其他专用表而非 JsonData。
- 临时处理：
  - 已按实库口径改为 `Treatment_Action.Code=150`，并可稳定读写回显；后续按你确认口径再做字段归位。

2. 菜单：`透析执行 -> 透析医嘱(orders_process)`
- 接口：`POST /api/v1/patients/:id/orders`
- 相关表：`Order_PatientOrder`, `Order_PatientDayOrder`, `Plan_PatientPlan`
- 相关字段：
  - `Order_PatientOrder.PatientPlanId`
  - `Order_PatientDayOrder.Status`
  - `Order_PatientDayOrder.DealOpportunity`
- 当前判断：
  - `PatientPlanId` 采用“患者最新方案ID，找不到则 0”。
  - 新建当日医嘱默认 `Status=20`（已确定）。
  - `DealOpportunity` 按字典 `DealOpportunityType` 写编码（`10=立即执行`,`20=普通`）。
- 不明确点：
  - 新建后是否还需立即生成 `Order_PatientDayOrderDeal` 执行流水（当前未自动生成）。
## 2026-04-23 设备管理模块补充（老库字段增强）

### 已完成
- 菜单：`设备管理`
- 接口：`GET /api/v1/devices`、`GET /api/v1/devices/:id`、`GET /api/v1/devices/:id/usage-logs`、`GET /api/v1/devices/:id/maintenance-records`、`GET /api/v1/devices/:id/disinfections`
- 表映射：
  - `Auxiliary_EquipmentInfomation`（设备档案）
  - `Auxiliary_EquipmentUsageLog`（设备使用记录）
  - `Auxiliary_EquipmentMaintenance`（设备保养维修记录）
  - `Schedule_BedEquipmentRel` / `Schedule_Bed` / `Schedule_Ward`（床位与病区绑定）
- 前端设备详情已补充显示：
  - 设备类型、生产厂家、床位、病区、生产日期、安装日期、维护周期、通量
  - 最近保养维修记录（Type/Mode/OperatorId/OperateTime/Description/Note）
  - 最近使用记录（UseUserId/UseStartTime/UseDuration/Note）
- 设备状态展示改为优先使用后端返回状态（`normal/warning/alarm/offline/maintenance`），不再只依赖前端随机模拟。

### 待确认（不确定项）
1. 菜单：`设备管理` -> 设备档案
- 接口：`GET /api/v1/devices`
- 字段：`Auxiliary_EquipmentInfomation.Maintenance`
- 当前判断：字段语义更像“运维厂家ID”。
- 不确定点：是否始终引用字典 `CodeDictionary_CodeDictionarys(Type='EquipmentInfomationMaintenance')`，以及是否存在其他主数据表作为唯一来源。
- 临时处理：当前先按数值透传，不强制转换成名称。

2. 菜单：`设备管理` -> 设备档案
- 接口：`GET /api/v1/devices`
- 字段：`Schedule_BedEquipmentRel.ParameterS`
- 当前判断：被当作设备运行状态承载（normal/warning/alarm/offline/maintenance）。
- 不确定点：你之前标注“老系统暂未使用，jsonb类型”；是否后续应改为仅状态字段或保留完整JSON结构。
- 临时处理：继续兼容字符串状态读取，未知值原样返回。
## 2026-04-24 前端透析执行改造补记

### 待确认
1. 菜单：透析执行 -> 透析医嘱
- 接口：
  - `GET /api/v1/patients/:id/orders`
  - `POST /api/v1/patients/:id/orders`
  - `PUT /api/v1/patients/:id/orders/:orderId`
  - `POST /api/v1/patients/:id/orders/:orderId/stop`
- 相关字段：
  - `Order.type`
  - `Order.status`
  - `Order.startTime`
  - `Order.endTime`
  - `Order.execTiming`
  - `Order.frequency`
  - `Order.notes`
- 当前判断：
  - 新版前端“透析医嘱”已改成真实接口加载，支持新增、编辑、停嘱。
  - 右侧终止按钮已按“停嘱”语义接到 `stop` 接口，不再做前端假删除。
- 不确定点：
  - 当前页面按钮文案/交互是否应保持为“停嘱”而不是“删除/作废”。
  - `stop` 接口返回值当前按“返回最新医嘱列表”处理；若后端后续改为只返回单条医嘱，前端需再调整。
- 临时处理：
  - 先以“停嘱”作为唯一终止动作，保留编辑能力，不做物理删除入口。

2. 菜单：透析执行 -> 透析医嘱
- 接口：
  - `POST /api/v1/patients/:id/orders`
  - `PUT /api/v1/patients/:id/orders/:orderId`
- 相关字段：
  - `Order.category`
  - `Order.name`
  - `Order.content`
  - `Order.dose`
  - `Order.unit`
  - `Order.route`
  - `Order.timing`
  - `Order.execTiming`
  - `Order.spec`
  - `Order.priority`
- 当前判断：
  - 新版前端已按现有接口能力接入通用字段编辑。
  - 页面暂未接模板选单、药品目录选取、分组医嘱、复制医嘱等增强动作。
- 不确定点：
  - 透析执行页中的透析医嘱是否需要继续完全复刻旧页的“模板开立/成组医嘱/复制上一条”操作流。
  - `drugId/groupId` 在此页是否需要显式暴露给前端，而不是仅由后端隐式处理。
- 临时处理：
  - 当前先确保真实医嘱数据可查、可增、可改、可停；复杂开立流待你确认后再补。
3. 菜单：透析执行 -> 透中监测
- 接口：
  - `GET /api/v1/patients/:id/treatment?date=YYYY-MM-DD`
  - `POST /api/v1/treatments/:id/during-params`
  - `PUT /api/v1/treatments/:id/during-params/:paramId`
  - `DELETE /api/v1/treatments/:id/during-params/:paramId`
- 相关字段：
  - `Treatment_DuringParam.OperateTime`
  - `Treatment_DuringParam.BloodFlow`
  - `Treatment_DuringParam.UFVolume`
  - `Treatment_DuringParam.TMP`
  - `Treatment_DuringSigns.SBP`
  - `Treatment_DuringSigns.DBP`
  - `Treatment_DuringSigns.HeartRate`
  - `Treatment_DuringSigns.Respiration`
  - `Treatment_DuringSigns.SpO2`
- 当前判断：
  - 新版前端已接真实透中监测列表，并支持新增、编辑、删除。
  - 前端 `restClient` 已补齐 `sbp/dbp/heartRate/respiration/spO2` 字段，避免丢失后端已支持的生命体征。
- 不确定点：
  - 当前在“无当日治疗记录”时，前端会先创建治疗并将状态置为 `1` 后再写透中监测；这是否符合你的业务口径。
  - 页面中“记录人”当前按 `creatorId -> 用户姓名` 展示；若老页口径应显示其他操作人字段，需要再调整。
- 临时处理：
  - 先保证透中监测数据可落库、可回显、可维护；治疗状态自动置 `1` 与记录人显示口径待你确认后再收口。

4. 菜单：透析执行 -> 透后评估
- 接口：
  - `GET /api/v1/patients/:id/treatment?date=YYYY-MM-DD`
  - `PUT /api/v1/treatments/:id/after-signs`
  - `PUT /api/v1/treatments/:id/status`
- 相关字段：
  - `Treatment_AfterSigns.Weight`
  - `Treatment_AfterSigns.ExtraWeight`
  - `Treatment_AfterSigns.LossWeight`
  - `Treatment_AfterSigns.SBP`
  - `Treatment_AfterSigns.DBP`
  - `Treatment_AfterSigns.HeartRate`
  - `Treatment_AfterSigns.Respiration`
  - `Treatment_AfterSigns.Temperature`
  - `Treatment_AfterSigns.RealIntake`
  - `Treatment_AfterSigns.PressurePoint`
  - `Treatment_AfterSigns.Complication`
  - `Treatment_AfterSigns.Symptoms`
  - `Treatment_AfterSymptom.Code/Value`
  - `Treatment_Treatment.StartTime`
  - `Treatment_Treatment.EndTime`
  - `Treatment_Treatment.RealUFQuantity`
  - `Treatment_Treatment.RealSubstituateVolume`
- 当前判断：
  - 新版前端已接真实透后评估保存，支持“暂存”和“提交透后评估”。
  - “提交”当前实现为：先保存 `after-signs`，再调用 `updateTreatmentStatus(..., 2)` 完成治疗。
  - 页面回填优先使用现有接口已返回的顶层字段 `endTime/endBp/weightLossKg/complications/treatmentSummary/afterSymptomItems`。
- 不确定点：
  - 当前治疗详情接口没有独立 `afterSigns` 对象，部分透后字段只能从 `afterSymptomItems` 或顶层摘要字段回填；若你要求完整回显，可能需要后端再补一个明确的 `afterSigns` 快照对象。
  - `heartRate/respiration/temperature/realIntake/pressurePoint` 当前按 `afterSymptomItems` 中的约定 code 回填：`heart_rate/respiration/temperature/real_intake/bp_site`；如果你的实际 code 口径不同，需要再对齐。
  - “提交即完成治疗”是否严格符合现场流程口径，仍需你确认。
- 临时处理：
  - 先保证透后评估可保存、可提交、可结束治疗；详细回显口径和 code 命名待你确认后再精修。

5. 菜单：透析执行 -> 健康宣教
- 接口：
  - 当前未接独立后端接口
- 相关字段：
  - 待确认老库表/接口
- 当前判断：
  - 仓库内未发现现成的宣教记录接口。
  - 新版前端已去掉静态假记录，改成“真实患者/治疗上下文 + 无数据源提示”，避免联调误判。
- 不确定点：
  - 健康宣教最终应落在哪个菜单接口或老库表上，目前仓库内没有明确数据源。
- 临时处理：
  - 页面先作为非误导占位页保留，待你确认宣教数据源后再补录入和历史记录。

6. 菜单：透析执行 -> 透析小结
- 接口：
  - `GET /api/v1/patients/:id/treatment?date=YYYY-MM-DD`
- 相关字段：
  - `Treatment_Treatment.TreatmentSummary`
  - `Treatment_Treatment.NurseSummary`
  - `Treatment_Treatment.StartTime`
  - `Treatment_Treatment.EndTime`
  - `Treatment_Treatment.RealUFQuantity`
  - `Treatment_DuringParam.*`
- 当前判断：
  - 新版前端已去掉静态摘要与静态监测表，改为基于当前治疗详情和透中监测的真实只读汇总页。
- 不确定点：
  - 当前小结页仍为只读展示，尚未新增独立“保存小结”动作；如果你要求继续支持单独编辑 doctor/nurse summary，需要再补保存接口和字段口径。
- 临时处理：
  - 先按真实治疗数据生成只读汇总，避免展示假数据。
