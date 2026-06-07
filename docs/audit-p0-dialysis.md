# 透析执行全子页面 字段级核查报告

> 核查时间：2026-05-31
> 核查范围：透析执行主框架及全部 16 个子页面
> 核查依据：`老血透数据库表结构-合并版.md`、`LEGACY_TABLE_FIELD_MAPPING.md`、`legacy-migration-uncertain-field-checklist.md`
> 说明：✅=一致/已闭环 · ⚠️=待确认/部分闭环 · ❌=缺失/不一致 · 🔲=占位页未实现

---

## 一、总览

| 子页面 | 前端文件 | 实现状态 | 核查结论 |
|--------|----------|----------|----------|
| 患者列表侧栏 | `PatientListSidebar.tsx` | ✅ 已实现 | ✅ 读 `Register_PatientInfomation`，字段一致 |
| 患者摘要头部 | `PatientSummaryHeader.tsx` | ✅ 已实现 | ✅ 聚合展示，无独立落库 |
| 透前评估 | `PreAssessment.tsx` | ✅ 已实现 | ⚠️ 部分 symptomItem code 无老库对照 |
| 透中监测 | `MidMonitoring.tsx` | ✅ 已实现 | ⚠️ `dialysateFlow` 仅存 JSON，未落显式列 |
| 首次核对 | `FirstCheck.tsx` | 🔲 占位页 | ❌ 未实现表单，后端已有 `Treatment_BeforeCheck` |
| 双人核对 | `SecondCheck.tsx` | 🔲 占位页 | ❌ 未实现表单，后端已有 `Auxiliary_JsonData(again_check)` |
| 透后评估 | `PostAssessment.tsx` | ✅ 已实现 | ⚠️ 部分 symptomItem code 无老库对照 |
| 透析小结 | `DialysisSummary.tsx` | ✅ 已实现 | ✅ 写 `Treatment_Treatment.NurseSummary/TreatmentSummary` |
| 透析医嘱 | `MedicalOrders.tsx` | ✅ 已实现 | ✅ 读写 `Order_PatientOrder` |
| 健康宣教 | `HealthEducation.tsx` | ✅ 已实现 | ✅ 读写 `Auxiliary_HealthEducation` + `Auxiliary_PatientHealthEducation` |
| 消毒登记 | `Disinfection.tsx` | 🔲 占位页 | ⚠️ 后端已实现 `Auxiliary_EquipmentDisinfection`，前端未接 |
| 耗材记录 | `Consumables.tsx` | 🔲 占位页 | ❌ 前后端均未实现 |
| 今日处方 | `TodayPrescription.tsx` | ✅ 已实现 | ✅ 读写 `Plan_PatientPrescription` |
| 核查验证 | `Verification.tsx` | ✅ 已实现 | ⚠️ 首次核对写 `Treatment_BeforeCheck`；二次核对写 `Auxiliary_JsonData` |

---

## 二、字段级详细核查

### 2.1 透前评估 (PreAssessment)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| P-01 | P0 | PreAssessment.tsx:176 | 透前体重 | `weight` | `PUT /treatments/:id/before-signs` | treatment_service.go:1931 | `Treatment_BeforeSigns.Weight` | `Treatment_BeforeSigns.Weight numeric` | 类型一致 | ✅ | 后端 `values["Weight"] = req.Weight` | — | — |
| P-02 | P0 | PreAssessment.tsx:177 | 额外体重 | `extraWeight` | 同上 | treatment_service.go:1932 | `Treatment_BeforeSigns.ExtraWeight` | `Treatment_BeforeSigns.ExtraWeight numeric` | 类型一致 | ✅ | — | — | — |
| P-03 | P0 | PreAssessment.tsx:179 | 收缩压 | `sbp` | 同上 | treatment_service.go:1933 | `Treatment_BeforeSigns.SBP` | `Treatment_BeforeSigns.SBP numeric` | 类型一致 | ✅ | — | — | — |
| P-04 | P0 | PreAssessment.tsx:180 | 舒张压 | `dbp` | 同上 | treatment_service.go:1934 | `Treatment_BeforeSigns.DBP` | `Treatment_BeforeSigns.DBP numeric` | 类型一致 | ✅ | — | — | — |
| P-05 | P0 | PreAssessment.tsx:181 | 心率 | `heartRate` | 同上 | treatment_service.go:1935 | `Treatment_BeforeSigns.HeartRate` | `Treatment_BeforeSigns.HeartRate numeric` | 类型一致 | ✅ | — | — | — |
| P-06 | P0 | PreAssessment.tsx:182 | 呼吸 | `respiration` | 同上 | treatment_service.go:1936 | `Treatment_BeforeSigns.Respiration` | `Treatment_BeforeSigns.Respiration numeric` | 类型一致 | ✅ | — | — | — |
| P-07 | P0 | PreAssessment.tsx:183 | 体温 | `temperature` | 同上 | treatment_service.go:1937 | `Treatment_BeforeSigns.BodyTemp` | `Treatment_BeforeSigns.BodyTemp numeric` | 字段名映射 temperature→BodyTemp | ✅ | 后端 `values["BodyTemp"] = req.Temperature` | — | — |
| P-08 | P0 | PreAssessment.tsx:184 | 测压部位 | `pressurePoint` | 同上 | treatment_service.go:1938 | `Treatment_BeforeSigns.PressurePoint` | `Treatment_BeforeSigns.PressurePoint varchar(64)` | 类型一致 | ✅ | — | — | — |
| P-09 | P0 | PreAssessment.tsx:318 | 目标超滤量 | `symptomItems[code=uf_volume]` | 同上 (symptomItems) | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` (Code=uf_volume) | `Treatment_BeforeSymptom.Code/Value varchar` | 以 Code/Value 对存储 | ⚠️ | 老库 `Treatment_BeforeSymptom` 是通用 Code/Value 结构，新前端用自定义 code `uf_volume` | 确认老库是否曾用此 code | ✅ 需确认 |
| P-10 | P1 | PreAssessment.tsx:319 | A端位点 | `symptomItems[code=a_site]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` (Code=a_site) | 同上 | 自定义 code | ⚠️ | 老库无 `a_site` 历史 code | 确认老库原 code 名 | ✅ 需确认 |
| P-11 | P1 | PreAssessment.tsx:320 | V端位点 | `symptomItems[code=v_site]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` (Code=v_site) | 同上 | 自定义 code | ⚠️ | 同上 | 同上 | ✅ 需确认 |
| P-12 | P1 | PreAssessment.tsx:321 | 神志状态 | `symptomItems[code=consciousness]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 老库无此 code 记录 | 确认老库原 code 名 | ✅ 需确认 |
| P-13 | P1 | PreAssessment.tsx:322 | 护理分级 | `symptomItems[code=nurse_level]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 同上 | 同上 | ✅ 需确认 |
| P-14 | P1 | PreAssessment.tsx:323 | 跌倒风险 | `symptomItems[code=fall_risk]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 同上 | 同上 | ✅ 需确认 |
| P-15 | P1 | PreAssessment.tsx:324 | 疼痛评分 | `symptomItems[code=pain_score]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 同上 | 同上 | ✅ 需确认 |
| P-16 | P1 | PreAssessment.tsx:325 | 内瘘描述 | `symptomItems[code=fistula_tags]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 同上 | 同上 | ✅ 需确认 |
| P-17 | P1 | PreAssessment.tsx:326 | 症状历史 | `symptomItems[code=symptoms]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | code `symptoms` 与老库 `Treatment_BeforeSymptom.Code` 兼容 | ✅ | 老库设计 md 注释"属性名，枚举Code" | — | — |
| P-18 | P1 | PreAssessment.tsx:327 | 皮肤记录 | `symptomItems[code=skin_record]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 老库无此 code | 确认 | ✅ 需确认 |
| P-19 | P2 | PreAssessment.tsx:328 | 患者拒测 | `symptomItems[code=pre_weight_refused]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code，值为"是/否" | ⚠️ | 老库无布尔标志 code | 确认 | ✅ 需确认 |
| P-20 | P2 | PreAssessment.tsx:329 | 卧床 | `symptomItems[code=pre_bedridden]` | 同上 | treatment_service.go:1945 | `Treatment_BeforeSymptom.Value` | 同上 | 自定义 code | ⚠️ | 同上 | 确认 | ✅ 需确认 |
| P-21 | P0 | PreAssessment.tsx:191 | 透前备注 | `notes` | 同上 | treatment_service.go:1939 | `Treatment_BeforeSigns.Note` | `Treatment_BeforeSigns.Note varchar(1024)` | 类型一致 | ✅ | — | — | — |

### 2.2 透中监测 (MidMonitoring)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| M-01 | P0 | MidMonitoring.tsx:113 | 收缩压 | `sbp` | `POST /treatments/:id/during-params` | treatment_service.go:1753+1670 | `Treatment_DuringSigns.SBP` | `Treatment_DuringSigns.SBP numeric` | 类型一致 | ✅ | 通过 `upsertDuringSignsByTime` 写入 | — | — |
| M-02 | P0 | MidMonitoring.tsx:114 | 舒张压 | `dbp` | 同上 | treatment_service.go:1673 | `Treatment_DuringSigns.DBP` | `Treatment_DuringSigns.DBP numeric` | 类型一致 | ✅ | — | — | — |
| M-03 | P0 | MidMonitoring.tsx:115 | 心率 | `heartRate` | 同上 | treatment_service.go:1676 | `Treatment_DuringSigns.HeartRate` | `Treatment_DuringSigns.HeartRate numeric` | 类型一致 | ✅ | — | — | — |
| M-04 | P1 | MidMonitoring.tsx:116 | 呼吸 | `respiration` | 同上 | treatment_service.go:1679 | `Treatment_DuringSigns.Respiration` | `Treatment_DuringSigns.Respiration numeric` | 类型一致 | ✅ | — | — | — |
| M-05 | P1 | MidMonitoring.tsx:117 | 血氧 | `spO2` | 同上 | treatment_service.go:1682 | `Treatment_DuringSigns.SpO2` | `Treatment_DuringSigns.SpO2 numeric` | 类型一致 | ✅ | — | — | — |
| M-06 | P0 | MidMonitoring.tsx:119 | 超滤量 | `ufVolume` | 同上 | treatment_service.go:1742 | `Treatment_DuringParam.UFQuantity` | `Treatment_DuringParam.UFQuantity numeric` | 字段名映射 ufVolume→UFQuantity | ✅ | 后端 `values["UFQuantity"] = req.UFVolume` | — | — |
| M-07 | P0 | MidMonitoring.tsx:120 | 血流量 | `bloodFlow` | 同上 | treatment_service.go:1744 | `Treatment_DuringParam.BF` | `Treatment_DuringParam.BF numeric` | 字段名映射 bloodFlow→BF | ✅ | 后端 `values["BF"] = req.BloodFlow` | — | — |
| M-08 | P0 | MidMonitoring.tsx:121 | 透析液流量 | `dialysateFlow` | 同上 | treatment_service.go:1761 | **仅写入 `Auxiliary_JsonData` (Code=hp_during_other)** | `Treatment_DuringParam` 无 `DialysateFlow` 列 | **未落显式列，仅存 JSON** | ⚠️ | 后端 `upsertTreatmentSignsJSONSnapshot` 写入 JSON；`Treatment_DuringParam` 表无此列 | E-1 待确认：是否有指定老库字段可落 | ✅ 已标记 E-1 |
| M-09 | P0 | MidMonitoring.tsx:122 | 静脉压 | `venousPressure` | 同上 | treatment_service.go:1738 | `Treatment_DuringParam.VenousPressure` | `Treatment_DuringParam.VenousPressure numeric` | 类型一致 | ✅ | — | — | — |
| M-10 | P0 | MidMonitoring.tsx:123 | 动脉压 | `arterialPressure` | 同上 | treatment_service.go:1739 | `Treatment_DuringParam.ArterialPressure` | `Treatment_DuringParam.ArterialPressure numeric` | 类型一致 | ✅ | — | — | — |
| M-11 | P0 | MidMonitoring.tsx:124 | 跨膜压 | `tmp` | 同上 | treatment_service.go:1740 | `Treatment_DuringParam.TMP` | `Treatment_DuringParam.TMP numeric` | 类型一致 | ✅ | — | — | — |
| M-12 | P1 | MidMonitoring.tsx:125 | 机温 | `temperature` | 同上 | treatment_service.go:1743 | `Treatment_DuringParam.MachineTmp` | `Treatment_DuringParam.MachineTmp numeric` | 字段名映射 temperature→MachineTmp | ✅ | 后端 `values["MachineTmp"] = req.Temperature` | — | — |
| M-13 | P1 | MidMonitoring.tsx:126 | 电导度 | `conductivity` | 同上 | treatment_service.go:1741 | `Treatment_DuringParam.Conductivity` | `Treatment_DuringParam.Conductivity numeric` | 类型一致 | ✅ | — | — | — |
| M-14 | P2 | MidMonitoring.tsx:127 | 超滤率 | `ufRate` | 同上 | treatment_service.go:1786 | **仅写入响应，未落库** | 老库 `Treatment_DuringParam` 无 `UFRate` 列 | 前端传但后端不落库 | ⚠️ | 后端 DTO 有 `UFRate` 但 `CreateDuringParam` 不写入任何表 | 确认是否需要落库 | ✅ 需确认 |
| M-15 | P1 | MidMonitoring.tsx:128 | 备注 | `notes` | 同上 | treatment_service.go:1745 | `Treatment_DuringParam.Note` | `Treatment_DuringParam.Note` 列不存在（老库 `Treatment_DuringParam` 无 Note 列） | ⚠️ 后端写入 `values["Note"]` 但老库表无此列 | ⚠️ | 老库 `Treatment_DuringParam` 结构中无 `Note` 列；但后端代码确实写入了 | 确认老库是否有 Note 列或改用 JSON | ✅ 需确认 |

### 2.3 首次核对 (FirstCheck / Verification)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| F-01 | P0 | FirstCheck.tsx:1-10 | 首次核对页面 | **占位页，仅显示患者名** | — | — | — | — | 前端未实现表单 | ❌ | 文件仅 10 行，无表单 | 需实现：透析物品/参数/血管通路/管路 四项核查 | — |
| F-02 | P0 | Verification.tsx:77-86 | 首次核对表单(核查验证页) | `materials/param/vascular/pipeline` 各有 result+mistake | `PUT /treatments/:id/first-check` | treatment_service.go:2002-2107 | `Treatment_BeforeCheck` 全部字段 | `Treatment_BeforeCheck` 18 字段 | 字段完全匹配 | ✅ | 后端 `SaveFirstCheck` 写入 `MaterialsResult/MaterialsMistake/ParamResult/ParamMistake/VascularAccessResult/VascularAccessMistake/PipelineResult/PipelineMistake` | — | — |
| F-03 | P0 | Verification.tsx:35-41 | 首次核对操作人 | `operatorId` | 同上 | treatment_service.go:2018-2019 | `Treatment_BeforeCheck.OperatorId` + `Treatment_Action` (Code=70) | `Treatment_BeforeCheck.OperatorId bigint` + `Treatment_Action.Code` | 双写：BeforeCheck + Action | ✅ | 后端 `upsertLegacyAction` 写入 Action Code="70" | — | — |

### 2.4 双人核对 (SecondCheck / Verification)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| S-01 | P0 | SecondCheck.tsx:1-10 | 双人核对页面 | **占位页，仅显示患者名** | — | — | — | — | 前端未实现表单 | ❌ | 文件仅 10 行 | 需实现：透析模式/处方/抗凝/血管通路/管路连接 五项核查 | — |
| S-02 | P0 | Verification.tsx:89-101 | 双人核对表单(核查验证页) | `dialysisMode/prescription/anticoagulant/vascular/lineConnection` 各有 result+mistake | `PUT /treatments/:id/second-check` | treatment_service.go:2110-2232 | `Auxiliary_JsonData` (Code=hp_again_check) Value JSON | 老库无专用双人核对表，使用 `Auxiliary_JsonData` | JSON 结构存储 | ⚠️ | E-2 待确认：`Action/FirstCheck/AgainCheck` 的 Code/Value 结构 | 确认 JSON Value 的标准 schema | ✅ 已标记 E-2 |
| S-03 | P0 | Verification.tsx:44-51 | 双人核对-复核护士/质控护士 | `recheckNurseId/qcNurseId` | 同上 | treatment_service.go:2186-2187 | `Auxiliary_JsonData.Value->recheckNurseId/qcNurseId` | 老库无此字段 | 新增字段，存 JSON | ⚠️ | 老库 `Treatment_BeforeCheck` 无复核/质控护士字段 | 确认是否需扩展老库表 | ✅ 需确认 |
| S-04 | P0 | Verification.tsx:96 | 双人核对操作人 | `operatorId` | 同上 | treatment_service.go:2134 | `Treatment_Action` (Code=150) + JSON | `Treatment_Action.Code varchar(32)` | Action Code=150 对应二次核对 | ✅ | 后端 `upsertLegacyAction` Code="150" | — | — |

### 2.5 透后评估 (PostAssessment)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| Q-01 | P0 | PostAssessment.tsx:161 | 开始时间 | `startTime` | `PUT /treatments/:id/after-signs` 或 `/post-assessment-submit` | treatment_service.go:2369 | `Treatment_Treatment.StartTime` | `Treatment_Treatment.StartTime timestamp` | 类型一致 | ✅ | 后端 `updateTreatmentSummaryFields` 写入 | — | — |
| Q-02 | P0 | PostAssessment.tsx:162 | 结束时间 | `endTime` | 同上 | treatment_service.go:2372 | `Treatment_Treatment.EndTime` | `Treatment_Treatment.EndTime timestamp` | 类型一致 | ✅ | 使用 `COALESCE("EndTime", ?)` 保护 | — | — |
| Q-03 | P0 | PostAssessment.tsx:163 | 实际超滤量 | `realUfVolume` | 同上 | treatment_service.go:2376 | `Treatment_Treatment.RealUFQuantity` | `Treatment_Treatment.RealUFQuantity numeric` | 字段名映射 realUfVolume→RealUFQuantity | ✅ | — | — | — |
| Q-04 | P0 | PostAssessment.tsx:164 | 实际置换液量 | `realSubstituteVolume` | 同上 | treatment_service.go:2379 | `Treatment_Treatment.RealSubstituateVolume` | `Treatment_Treatment.RealSubstituateVolume numeric` | 字段名映射 | ✅ | — | — | — |
| Q-05 | P0 | PostAssessment.tsx:165 | 透后体重 | `weight` | 同上 | treatment_service.go:1959 | `Treatment_AfterSigns.Weight` | `Treatment_AfterSigns.Weight numeric` | 类型一致 | ✅ | — | — | — |
| Q-06 | P1 | PostAssessment.tsx:166 | 额外体重 | `extraWeight` | 同上 | treatment_service.go:1960 | `Treatment_AfterSigns.ExtraWeight` | `Treatment_AfterSigns.ExtraWeight numeric` | 类型一致 | ✅ | — | — | — |
| Q-07 | P1 | PostAssessment.tsx:167 | 体重丢失 | `lossWeight` | 同上 | treatment_service.go:1961 | `Treatment_AfterSigns.LossWeight` | `Treatment_AfterSigns.LossWeight numeric` | 类型一致 | ✅ | — | — | — |
| Q-08 | P0 | PostAssessment.tsx:168 | 收缩压 | `sbp` | 同上 | treatment_service.go:1962 | `Treatment_AfterSigns.SBP` | `Treatment_AfterSigns.SBP numeric` | 类型一致 | ✅ | — | — | — |
| Q-09 | P0 | PostAssessment.tsx:169 | 舒张压 | `dbp` | 同上 | treatment_service.go:1963 | `Treatment_AfterSigns.DBP` | `Treatment_AfterSigns.DBP numeric` | 类型一致 | ✅ | — | — | — |
| Q-10 | P1 | PostAssessment.tsx:170 | 心率 | `heartRate` | 同上 | treatment_service.go:1964 | `Treatment_AfterSigns.HeartRate` | `Treatment_AfterSigns.HeartRate numeric` | 类型一致 | ✅ | — | — | — |
| Q-11 | P1 | PostAssessment.tsx:171 | 呼吸 | `respiration` | 同上 | treatment_service.go:1965 | `Treatment_AfterSigns.Respiration` | `Treatment_AfterSigns.Respiration numeric` | 类型一致 | ✅ | — | — | — |
| Q-12 | P1 | PostAssessment.tsx:172 | 体温 | `temperature` | 同上 | treatment_service.go:1966 | `Treatment_AfterSigns.BodyTemp` | `Treatment_AfterSigns.BodyTemp numeric` | 字段名映射 temperature→BodyTemp | ✅ | — | — | — |
| Q-13 | P1 | PostAssessment.tsx:173 | 实际摄入 | `realIntake` | 同上 | treatment_service.go:1967 | `Treatment_AfterSigns.RealIntake` | `Treatment_AfterSigns.RealIntake numeric` | 类型一致 | ✅ | — | — | — |
| Q-14 | P1 | PostAssessment.tsx:174 | 测压部位 | `pressurePoint` | 同上 | treatment_service.go:1968 | `Treatment_AfterSigns.PressurePoint` | `Treatment_AfterSigns.PressurePoint varchar(64)` | 类型一致 | ✅ | — | — | — |
| Q-15 | P0 | PostAssessment.tsx:175 | 透析事件/并发症 | `complication` | 同上 | treatment_service.go:1969 | `Treatment_AfterSigns.Complication` | `Treatment_AfterSigns.Complication` 列不存在 | **老库 `Treatment_AfterSigns` 无 `Complication` 列** | ⚠️ | 后端写入 `values["Complication"]` 但老库表无此列；同时写入 `Auxiliary_JsonData` (Code=hp_after_symptom/hp_treatment_details) | 确认老库是否通过其他方式存储并发症 | ✅ 需确认 |
| Q-16 | P1 | PostAssessment.tsx:176 | 症状 | `symptoms` | 同上 | treatment_service.go:1970 | `Treatment_AfterSigns.Symptoms` | `Treatment_AfterSigns` 无 `Symptoms` 列 | **同上，老库无此列** | ⚠️ | 同上，通过 JSON 兼容 | 同上 | ✅ 需确认 |
| Q-17 | P1 | PostAssessment.tsx:150-157 | 凝血分级/内瘘护理等 | `symptomItems[code=dialyzer_coag/line_a_coag/line_v_coag/fistula_care]` | 同上 (symptomItems) | treatment_service.go:1980 | `Treatment_AfterSymptom.Value` (Code=dialyzer_coag 等) | `Treatment_AfterSymptom.Code/Value varchar` | Code/Value 对存储 | ⚠️ | 老库 `Treatment_AfterSymptom` 设计为通用 Code/Value，但 `hp_after_symptom` JSON 也存储 | 确认老库原 code 名 | ✅ 需确认 |
| Q-18 | P0 | PostAssessment.tsx:177 | 透后备注 | `notes` | 同上 | treatment_service.go:1971 | `Treatment_AfterSigns.Note` | `Treatment_AfterSigns.Note varchar(1024)` | 类型一致 | ✅ | — | — | — |
| Q-19 | P0 | PostAssessment.tsx:161-162 | 实际治疗时长 | 自动计算 `endTime - startTime` | 同上 | treatment_service.go:2381-2383 | `Treatment_Treatment.RealDuration` | `Treatment_Treatment.RealDuration numeric` | 后端自动计算写入 | ✅ | `updates["RealDuration"] = req.EndTime.Sub(*req.StartTime).Minutes()` | — | — |

### 2.6 透析小结 (DialysisSummary)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| T-01 | P0 | DialysisSummary.tsx:55-56 | 医生小结/治疗小结 | `doctorSummary/treatmentSummary` | `PUT /treatments/:id/summary` | treatment_service.go:2660-2688 | `Treatment_Treatment.TreatmentSummary/NurseSummary` | `Treatment_Treatment.TreatmentSummary varchar(1024)` / `NurseSummary varchar(1024)` | 类型一致 | ✅ | 后端分别写入两个字段 | — | — |
| T-02 | P1 | DialysisSummary.tsx:62-63 | 小结回显 | 读 `treatment.doctorSummary/treatmentSummary` | `GET /treatments/:id` | treatment_service.go:1040-1041 | 读 `Treatment_Treatment.TreatmentSummary/NurseSummary` | 同上 | 读写一致 | ✅ | `doctorSummary := row.TreatmentSummary; treatmentSummary := row.NurseSummary` | — | — |

### 2.7 透析医嘱 (MedicalOrders)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| O-01 | P0 | MedicalOrders.tsx:17-33 | 医嘱表单 | `type/category/name/content/dose/unit/route/timing/notes` 等 | `POST /patients/:id/orders` | order_service.go | `Order_PatientOrder` 全部字段 | `Order_PatientOrder` 27 字段 | 字段映射一致 | ✅ | 医嘱模块独立，已有完整映射 | — | — |

### 2.8 健康宣教 (HealthEducation)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| H-01 | P0 | HealthEducation.tsx:97-100 | 新增患者宣教记录 | `contentId/educationType/educationTime/educationResult/nurseSign/patientSign/note` | `POST /patients/:id/health-educations` | health_education_service.go:152-205 | `Auxiliary_PatientHealthEducation` 全部字段 | `Auxiliary_PatientHealthEducation` 15 字段 | 字段完全匹配 | ✅ | 后端读写 `Auxiliary_HealthEducation` + `Auxiliary_PatientHealthEducation`，JOIN 查询 `Organ_Employee` 获取操作人名 | — | — |
| H-02 | P0 | HealthEducation.tsx:67-70 | 读取宣教内容列表 | — | `GET /health-educations` | health_education_service.go:87-112 | `Auxiliary_HealthEducation` | `Auxiliary_HealthEducation` 13 字段 | 字段完全匹配 | ✅ | — | — | — |
| H-03 | P0 | HealthEducation.tsx:69 | 读取患者宣教记录 | — | `GET /patients/:id/health-educations` | health_education_service.go:114-150 | `Auxiliary_PatientHealthEducation` JOIN `Auxiliary_HealthEducation` + `Organ_Employee` | 同上 | 多表 JOIN 正确 | ✅ | — | — | — |

### 2.9 消毒登记 (Disinfection)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| D-01 | P0 | Disinfection.tsx:1-10 | 消毒登记页面 | **占位页，仅显示患者名** | — | — | — | — | 前端未实现 | 🔲 | 文件仅 10 行 | 后端 `PUT /treatments/:id/disinfection` 已实现，需对接前端 | — |
| D-02 | P0 | — | 后端消毒登记保存 | — | `PUT /treatments/:id/disinfection` | treatment_service.go:2615-2658 | `Auxiliary_EquipmentDisinfection` 全部字段 | `Auxiliary_EquipmentDisinfection` 16 字段 | 后端字段完全匹配 | ✅ | `EquipmentId/DisinfectUserId/DisinfectWay/Type/Disinfectant/StartTime/EndTime/Description/Note/TreatmentId/CreatorId` | 前端需实现表单并对接 | — |

### 2.10 耗材记录 (Consumables)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| C-01 | P0 | Consumables.tsx:1-10 | 耗材记录页面 | **占位页，仅显示患者名** | — | — | — | — | 前后端均未实现 | ❌ | 文件仅 10 行；后端无对应 API | 需设计：对接 `Plan_PatientPrescriptionMaterial` 或 `Treatment_MaterialTrace` | ✅ 需确认 |

### 2.11 今日处方 (TodayPrescription)

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| R-01 | P0 | TodayPrescription.tsx:21-46 | 处方表单 | `duration/dryWeight/dialysisMode/bloodFlow/substituteVolume/initialDrug/...` | `GET/PUT /patients/:id/prescriptions` | prescription_service.go | `Plan_PatientPrescription` 全部字段 | `Plan_PatientPrescription` 49 字段 | 字段映射一致 | ✅ | 处方模块独立，已有完整映射 | — | — |

### 2.12 核查验证 (Verification) — 首次核对 + 双人核对 + 消毒登记

| 编号 | 优先级 | 前端文件:行号 | 功能点 | 前端字段/操作 | API路径 | 后端文件:行号 | 当前表字段 | 老库标准表字段 | 类型/内容核查 | 结论 | 证据 | 建议改造方向 | 需人工确认 |
|------|--------|--------------|--------|--------------|---------|--------------|-----------|---------------|--------------|------|------|-------------|-----------|
| V-01 | P0 | Verification.tsx:200- | 核查验证页面 | 整合首次核对+双人核对+消毒登记 | 多个 API | 多个 | 多表 | 多表 | 集成页 | ✅ | Verification.tsx 是实际的核对表单实现页 | — | — |
| V-02 | P1 | Verification.tsx:300+ | 消毒登记子表单 | `equipmentId/disinfectUserId/disinfectWay/type/disinfectant/startTime/endTime/description/note` | `PUT /treatments/:id/disinfection` | treatment_service.go:2615-2658 | `Auxiliary_EquipmentDisinfection` | `Auxiliary_EquipmentDisinfection` 16 字段 | 字段完全匹配 | ✅ | 后端已有完整实现 | — | — |

---

## 三、Auxiliary_JsonData Code/Value Schema 核查

| 编号 | Code | 用途 | 写入位置 | Value JSON 结构 | 老库标准 | 结论 | 需人工确认 |
|------|------|------|----------|----------------|----------|------|-----------|
| J-01 | `hp_before_symptom` | 透前症状快照 | treatment_service.go:1948 | `{complication, symptoms, notes, symptomItems[{code,value}]}` | `Auxiliary_JsonData.Code varchar(64)` + `Value jsonb` | ⚠️ JSON schema 非固定，由业务自行定义 | ✅ 确认标准 schema |
| J-02 | `hp_after_symptom` | 透后症状快照 | treatment_service.go:1983 | 同上 | 同上 | ⚠️ 同上 | ✅ |
| J-03 | `hp_during_other` | 透中其他监测 | treatment_service.go:1756 | `{recordTime, code, notes, bloodFlow, dialysateFlow, ufVolume, sbp, dbp, heartRate, respiration, spO2}` | 同上 | ⚠️ 同上 | ✅ |
| J-04 | `hp_treatment_details` | 治疗详情快照 | treatment_service.go:1986 | `{complication, symptoms, symptomItems, notes}` | 同上 | ⚠️ 同上 | ✅ |
| J-05 | `hp_treatment_feelcontent` | 治疗感受快照 | treatment_service.go:1994 | `{notes}` | 同上 | ⚠️ 同上 | ✅ |
| J-06 | `hp_again_check` | 双人核对 | treatment_service.go:2204 | `{operatorId, recheckNurseId, qcNurseId, operateTime, paramResult, paramMistake, vascularAccessResult, vascularAccessMistake, pipelineResult, pipelineMistake, dialysisModeResult, dialysisModeMistake, prescriptionResult, prescriptionMistake, anticoagulantResult, anticoagulantMistake, lineConnectionResult, lineConnectionMistake}` | 同上 | ⚠️ 同上 | ✅ |

---

## 四、核心安全/一致性核查

### 4.1 startTime/endTime 不覆盖保护

| 编号 | 场景 | 保护机制 | 结论 | 代码位置 |
|------|------|----------|------|----------|
| S-01 | 状态改为"治疗中"时写 StartTime | `COALESCE("StartTime", ?)` — 仅当 StartTime 为 NULL 时才写入 | ✅ | treatment_service.go:1394, 1446 |
| S-02 | 状态改为"已完成"时写 EndTime | `COALESCE("EndTime", ?)` — 同上 | ✅ | treatment_service.go:1397 |
| S-03 | 透后评估显式传 startTime/endTime | 直接覆盖，不使用 COALESCE | ⚠️ 透后评估允许覆盖 StartTime/EndTime | treatment_service.go:2369-2373 | 

**结论**：UpdateStatus 接口有 COALESCE 保护；但透后评估 SubmitPostAssessment → updateTreatmentSummaryFields 允许直接覆盖 StartTime/EndTime。这是**设计意图**（透后评估需要修正时间），但需确认是否应增加"仅允许修正一次"或"仅允许修正未来时间"的限制。

### 4.2 不串患者保护

| 编号 | 场景 | 保护机制 | 结论 |
|------|------|----------|------|
| P-01 | Treatment 查询 | 所有查询均带 `"TenantId" = ?` 过滤 | ✅ |
| P-02 | DuringParam 写入 | 需先验证 treatmentId 存在：`Count` 检查 | ✅ |
| P-03 | BeforeSigns/AfterSigns 写入 | `upsertTreatmentSigns` 内先验证 treatment 存在 | ✅ |
| P-04 | FirstCheck/SecondCheck 写入 | 先验证 treatment 存在 | ✅ |
| P-05 | 但无 PatientId 交叉验证 | 写入时只验证 TreatmentId 存在，不验证 TreatmentId 是否属于当前操作的 PatientId | ⚠️ 如果前端传错 treatmentId，会串患者 |

**结论**：后端通过 TreatmentId 存在性检查提供基本保护，但**不验证 TreatmentId 与 PatientId 的归属关系**。前端应确保不串患者；后端可考虑增加归属校验。

---

## 五、待确认事项汇总（需人工确认）

| 编号 | 模块 | 待确认内容 | 当前处理 | 影响范围 |
|------|------|-----------|----------|----------|
| E-1 | 透中监测 | `dialysateFlow` 老库承接字段 | 仅写入 `Auxiliary_JsonData` (Code=hp_during_other)，未落 `Treatment_DuringParam` 显式列 | 透中监测 |
| E-2 | 核查验证 | `Action/FirstCheck/AgainCheck` 的 Code/Value JSON 结构标准 | 首次核对写 `Treatment_BeforeCheck` 表；二次核对写 `Auxiliary_JsonData` (Code=hp_again_check) | 核查验证 |
| E-3 | 透前评估 | symptomItem code 命名规范 (uf_volume/a_site/v_site/consciousness/nurse_level/fall_risk/pain_score/fistula_tags/skin_record/pre_weight_refused/pre_bedridden) | 新定义的 code，老库 `Treatment_BeforeSymptom` 无历史记录 | 透前评估 |
| E-4 | 透后评估 | symptomItem code 命名规范 (bp_site/real_intake/heart_rate/respiration/temperature/dialyzer_coag/line_a_coag/line_v_coag/fistula_care) | 新定义的 code，老库 `Treatment_AfterSymptom` 无历史记录 | 透后评估 |
| E-5 | 透后评估 | `Complication/Symptoms` 字段写入 `Treatment_AfterSigns` 但老库表可能无此列 | 后端同时写 `Treatment_AfterSigns` + `Auxiliary_JsonData` | 透后评估 |
| E-6 | 透中监测 | `Treatment_DuringParam.Note` 列是否存在 | 后端写入 `values["Note"]` 但老库结构文档未列出此列 | 透中监测 |
| E-7 | 透中监测 | `UFRate` 是否需要落库 | 前端传值，后端仅在响应中返回，不落库 | 透中监测 |
| E-8 | 耗材记录 | 耗材记录应使用哪张老库表 | 前后端均未实现；候选表：`Plan_PatientPrescriptionMaterial` 或 `Treatment_MaterialTrace` | 耗材记录 |
| E-9 | 核查验证 | 二次核对的 `recheckNurseId/qcNurseId` 是否需扩展到老库表 | 当前仅存 JSON | 核查验证 |
| E-10 | 透后评估 | 透后 StartTime/EndTime 是否允许直接覆盖 | 当前透后评估允许直接覆盖，UpdateStatus 用 COALESCE 保护 | 全局 |

---

## 六、占位页清单（需实现）

| 页面 | 前端文件 | 后端 API 状态 | 需实现内容 |
|------|----------|--------------|-----------|
| 首次核对 | `FirstCheck.tsx` | ✅ `PUT /treatments/:id/first-check` 已实现 | 表单：透析物品核查、参数核查、血管通路核查、管路核查（4项，各含 result+mistake） |
| 双人核对 | `SecondCheck.tsx` | ✅ `PUT /treatments/:id/second-check` 已实现 | 表单：透析模式核查、处方核查、抗凝核查、血管通路核查、管路连接核查（5项，各含 result+mistake）+ 复核护士/质控护士 |
| 消毒登记 | `Disinfection.tsx` | ✅ `PUT /treatments/:id/disinfection` 已实现 | 表单：设备ID、消毒人、消毒方式、消毒类型、消毒液、开始/结束时间、情况登记、备注 |
| 耗材记录 | `Consumables.tsx` | ❌ 后端未实现 | 需设计 API + 前端表单 |

> 注：首次核对和双人核对的表单已在 `Verification.tsx` 中实现，`FirstCheck.tsx` 和 `SecondCheck.tsx` 为独立占位页。
