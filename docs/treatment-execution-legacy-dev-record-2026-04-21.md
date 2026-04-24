# 治疗执行页老库迁移开发记录（2026-04-21）

本文记录本轮围绕“患者管理 / 治疗详情历史 / 透析执行页”的老血透库迁移改造，供后续持续开发和回归验证使用。

## 1. 目标范围

本轮主要处理以下能力：

- 患者治疗历史列表改读老库
- 单条治疗详情改读老库
- 治疗记录单弹窗展示老库字段
- 当天治疗记录 `GET /patients/:id/treatment`
- 透中监测记录新增 / 编辑 / 删除
- 透前评估保存
- 透后评估保存
- 透后保存时同步更新治疗主表摘要字段

## 2. 主要涉及文件

- `ai-hms-backend/internal/services/treatment_service.go`
- `ai-hms-backend/internal/api/v1/treatment_handler.go`
- `ai-hms-frontend/src/services/restClient.ts`
- `ai-hms-frontend/src/pages/DialysisProcessing.tsx`
- `ai-hms-frontend/src/pages/patient-detail/tabs/HistoryTab.tsx`
- `ai-hms-frontend/src/pages/patient-detail/types.ts`

## 3. 已完成的后端迁移

### 3.1 治疗历史 / 单条详情 / 当天治疗记录

已切换到老库表：

- `Treatment_Treatment`
- `Treatment_Action`
- `Treatment_BeforeSigns`
- `Treatment_AfterSigns`
- `Treatment_DuringParam`

已补充的返回字段：

- `treatmentDate`
- `treatmentType`
- `status`
- `startTime`
- `endTime`
- `timeRange`
- `doctorSummary`
- `treatmentSummary`
- `notes`
- `durationMinutes`
- `weightLossKg`
- `doctorName`
- `startBp`
- `endBp`
- `complications`
- `actions`
- `duringParams`
- `tmrPath`
- `tmrTime`
- `tmrPages`

### 3.2 治疗状态与主表写回

已兼容老表状态值：

- `0 -> "0"`
- `1 -> "10"`
- `2 -> "60"`
- `3 -> "90"`

已支持：

- `POST /api/v1/treatments`
- `PUT /api/v1/treatments/:id`
- `PUT /api/v1/treatments/:id/status`

透后保存时，额外会回写主表：

- `StartTime`
- `EndTime`
- `RealDuration`
- `RealUFQuantity`
- `RealSubstituateVolume`

### 3.3 透中监测记录

新增接口：

- `POST /api/v1/treatments/:id/during-params`
- `PUT /api/v1/treatments/:id/during-params/:paramId`
- `DELETE /api/v1/treatments/:id/during-params/:paramId`

当前稳定落库字段：

- `OperateTime`
- `BF`
- `UFQuantity`
- `VenousPressure`
- `ArterialPressure`
- `TMP`
- `MachineTmp`
- `Conductivity`
- `Note`

### 3.4 透前 / 透后评估主表

新增接口：

- `PUT /api/v1/treatments/:id/before-signs`
- `PUT /api/v1/treatments/:id/after-signs`

已写入 `Treatment_BeforeSigns`：

- `Weight`
- `ExtraWeight`
- `BodyTemp`
- `SBP`
- `DBP`
- `HeartRate`
- `Respiration`
- `PressurePoint`
- `Note`

已写入 `Treatment_AfterSigns`：

- `Weight`
- `ExtraWeight`
- `LossWeight`
- `RealIntake`
- `BodyTemp`
- `SBP`
- `DBP`
- `HeartRate`
- `Respiration`
- `PressurePoint`
- `Note`

### 3.5 透前 / 透后扩展项子表

已新增使用：

- `Treatment_BeforeSymptom`
- `Treatment_AfterSymptom`

当前采用 `Code + Value` 通用存储模型，已接入的页面扩展项包括：

- 透前：
  - `bp_site`
  - `symptoms`
  - `fistula_status`
  - `a_point`
  - `v_point`
  - `fall_risk`
  - `pain_score`
  - `nursing_level`
  - `check_in_time`
  - `admission_time`
  - `assess_time`
  - `doctor`
  - `on_machine_nurse`
  - `assessor`
  - `start_time`
- 透后：
  - `bp_site`
  - `actual_replacement`
  - `dialyzer_coag`
  - `line_a_coag`
  - `line_v_coag`
  - `fistula_care`
  - `accident`
  - `dialysis_event`
  - `fistula_status`
  - `assess_time`
  - `on_machine_nurse`
  - `assessor`

## 4. 已完成的前端接线

### 4.1 治疗历史页

- 历史列表已优先展示老库字段
- “治疗记录单”弹窗已能展示老库历史记录
- 支持按年 / 月 / 日筛选历史记录

### 4.2 透析执行页

已从“展示假界面”改为真实调用接口：

- 透前页：
  - “暂存” -> 保存 `before-signs`
  - “提交下一步” -> 保存 `before-signs` 后切到处方页
- 透中监测页：
  - 新增监测
  - 编辑监测
  - 删除监测
  - 保存后刷新当天治疗记录
- 透后页：
  - “提交下一步” -> 保存 `after-signs`，再更新治疗状态为完成

若当天尚无治疗记录：

- 透中监测保存时会自动创建当天治疗
- 透前 / 透后保存时也会自动确保存在当天治疗记录

## 5. 验证结果

本地已通过：

- `cd ai-hms-backend && go build ./...`
- `cd ai-hms-backend && go test ./...`
- `cd ai-hms-frontend && npm run build`

## 6. 后续建议

下一步优先级建议如下：

1. 将 `BeforeSymptom / AfterSymptom` 已写入的扩展字段在页面初始化时反查并回显。
2. 若需要与旧系统更一致，补充 `Code` 的旧系统标准编码映射，而不是继续使用当前业务代码字符串。
3. 继续处理 `Treatment_BeforeCheck`、`Treatment_BeforeCheck` 相关核对结果、透析执行页其他保存链。

## 7. 当前实现口径

本轮策略是：

- 先保证“页面提交能真实落老库”
- 对有明确旧字段的内容，写主表
- 对旧库中是通用扩展结构的内容，写 `Code + Value` 子表
- 对还没有明确落点的复杂字段，暂不强写，避免误映射

## 8. 诊疗配置老库迁移补充（2026-04-22）

本次继续完成了“诊疗配置”模块的老库切换，覆盖范围从只读扩展到完整读写。

涉及文件：
- `ai-hms-backend/internal/services/treatment_config_service.go`
- `ai-hms-backend/internal/api/v1/treatment_config_handler.go`

已切换到老库的模块与表：
- 方案模板：
  - `Plan_PlanTPL`
  - `Plan_PlanTPLMaterial`
- 材料目录：
  - `Auxiliary_MaterialInfomation`
- 药品目录：
  - `Auxiliary_DrugInfomation`
- 医嘱模板：
  - `Order_OrderTPL`

已打通接口：
- 方案模板：
  - `GET /api/v1/treatment-templates`
  - `GET /api/v1/treatment-templates/:id`
  - `POST /api/v1/treatment-templates`
  - `PUT /api/v1/treatment-templates/:id`
  - `DELETE /api/v1/treatment-templates/:id`
  - `POST /api/v1/treatment-templates/:id/toggle`
- 材料目录：
  - `GET /api/v1/materials/catalog`
  - `GET /api/v1/materials/catalog/:id`
  - `GET /api/v1/materials/categories`
  - `POST /api/v1/materials/catalog`
  - `PUT /api/v1/materials/catalog/:id`
  - `DELETE /api/v1/materials/catalog/:id`
  - `POST /api/v1/materials/catalog/:id/toggle`
- 药品目录：
  - `GET /api/v1/drugs/catalog`
  - `GET /api/v1/drugs/catalog/:id`
  - `GET /api/v1/drugs/categories`
  - `POST /api/v1/drugs/catalog`
  - `PUT /api/v1/drugs/catalog/:id`
  - `DELETE /api/v1/drugs/catalog/:id`
  - `POST /api/v1/drugs/catalog/:id/toggle`
- 医嘱模板：
  - `GET /api/v1/order-templates`
  - `GET /api/v1/order-templates/:id`
  - `POST /api/v1/order-templates`
  - `PUT /api/v1/order-templates/:id`
  - `DELETE /api/v1/order-templates/:id`
  - `POST /api/v1/order-templates/:id/toggle`

当前实现口径：
- 方案模板材料从 `Plan_PlanTPLMaterial` 读写，材料名通过 `Auxiliary_MaterialInfomation` 反查。
- 药品目录的启停用通过 `Auxiliary_DrugInfomation.IsDisabled` 维护。
- 医嘱模板按 `OrderGroup` 聚合视为一个模板；模板条目落为多条 `Order_OrderTPL`。
- 医嘱模板的 `type` 由于老表无独立字段，当前兼容保存在 `Note`，列表/详情再回读推断。
- 删除操作当前统一按老库“禁用”口径处理，不做物理删除。

验证结果：
- `cd ai-hms-backend && go build ./...`
- `cd ai-hms-backend && go test ./...`
