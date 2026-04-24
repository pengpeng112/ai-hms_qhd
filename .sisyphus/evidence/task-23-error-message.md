## Task 23 证据记录

### 已统一的模块
- 透析执行：DialysisExecution / PreAssessment / PostAssessment / Verification / MidMonitoring / MedicalOrders / TodayPrescription
- 字典配置：DictConfig
- 患者管理：PatientList / PatientDetail / BasicInfoTab / VascularTab / LabsExamsTab / TreatmentPlanTab / SchemeOrderTab / MedicalRecordTab

### 结果
- 统一使用 `getErrorMessage` 展示 API 错误。
- 401/403/404/409/500 和网络错误已有独立提示。
- 400/409 优先保留后端返回的校验信息。
- DialysisExecution 的 401/403 加载错误已补充提示，不再静默吞错。

### 备注
- 保留了部分非 API 校验提示（如必填项、无选择项）。
