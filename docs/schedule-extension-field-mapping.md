# 老表 → 新规则扩展表 字段映射

## 1. 病区映射

| 老表.字段 | 扩展表.字段 | 说明 |
|-----------|------------|------|
| `Schedule_Ward.Id` | `Schedule_WardExt.WardId` | 老病区主键 |
| `Schedule_Ward.PatientType` | `Schedule_WardExt.ZoneType` | 辅助映射，不能作为唯一规则源 |
| (无) | `Schedule_WardExt.ParentWardId` | 子区树，新字段 |
| (无) | `Schedule_WardExt.IsSubZone` | 子区标记，新字段 |

## 2. 机器/床位映射

| 老表.字段 | 扩展表.字段 | 说明 |
|-----------|------------|------|
| `Schedule_Bed.Id` | `Schedule_BedMachineExt.BedId` | 老床位主键 |
| `Schedule_Bed.Name` | `Schedule_BedMachineExt.MachineCode` | 机器编号 |
| `Schedule_Bed.Sort` | `Schedule_BedMachineExt.PositionIndex` | 物理排位 |
| (无) | `Schedule_BedMachineExt.MachineType` | HD/HDF/CRRT，需人工配置 |
| (无) | `Schedule_BedMachineExt.SupportedModes` | 能力矩阵，需人工配置 |

## 3. 患者排班骨架映射

| 老表.字段 | 扩展表.字段 | 说明 |
|-----------|------------|------|
| `Register_PatientInfomation.Id` | `Schedule_PatientProfile.PatientId` | 患者主键 |
| (无) | `Schedule_PatientProfile.ZoneTag` | A/B/C，需业务确认映射 |
| (无) | `Schedule_PatientProfile.HomeWardId` | 归属病区 |
| `Plan_PatientPlan.OddWeekFrequency` | 推断 `FreqPattern` | 初始映射，不完整项需人工 |
| `Plan_PatientPlan.EvenWeekFrequency` | 推断 `FreqPattern` | 同上 |
| (无) | `Schedule_PatientProfile.FreqPattern` | 五种频率枚举 |
| `Plan_PatientPlan.DialysisMethod` | `Schedule_PatientProfile.DefaultMode` | 治疗模式 |
| (无) | `Schedule_PatientProfile.ShiftId` | 固定班次 |
| (无) | `Schedule_PatientProfile.HdfEnabled` | 是否每两周 HDF |
| (无) | `Schedule_PatientProfile.HdfWeekday` | HDF 固定星期 |
| (无) | `Schedule_PatientProfile.FixedHdBedId` | 固定 HD 机器 |
| (无) | `Schedule_PatientProfile.FixedHdfBedId` | 固定 HDF 机器 |

## 4. 排班记录映射

| 老表.字段 | 扩展表.字段 | 说明 |
|-----------|------------|------|
| `Schedule_PatientShift.Id` | `Schedule_PatientShiftExt.PatientShiftId` | 排班主键 |
| `Schedule_PatientShift.Status` | `Schedule_PatientShiftExt.RuleStatus` | 老状态↔新状态双轨兼容 |
| (无) | `Schedule_PatientShiftExt.DialysisMode` | 按次 HD/HDF/HF/CRRT |
| (无) | `Schedule_PatientShiftExt.SourceType` | 10=常规，20=临时 |
| (无) | `Schedule_PatientShiftExt.RecordForm` | 10=规律，20=CRRT |
| (无) | `Schedule_PatientShiftExt.Confirm1/2/3At` | 三级确认时间戳 |
| (无) | `Schedule_PatientShiftExt.CancelReason` | 取消/缺席原因 |

## 5. 模板映射 [阶段 3 已完成]

| 老用法 | 新表 | 说明 |
|--------|------|------|
| `Schedule_PatientShift.Status=60` 伪模板 | `Schedule_ScheduleTemplate` + `Schedule_ScheduleTemplateItem` | 独立存储，不再复用状态值 |
| `ListTemplates` 查 Status=60 | 查新表 `GET /api/v1/patient-shifts/templates` | 旧方法已标记 Deprecated |
| `SaveTemplate` 删 Status=60 | 写新表 `POST /api/v1/schedule/template/save` | 禁止再 DELETE Status=60 |
| `ApplyTemplate` 生成 Status=20 | 生成草稿 `Status=10` + `PatientShiftExt(RuleStatus=10)` | 不直接确认，事务写入 |

### 阶段 3 落地状态

- **后端**: `ScheduleTemplateService` 已完整替代旧 `PatientShiftService` 模板方法。
  - `ListTemplates` / `SaveTemplate` / `ApplyTemplate` 路由均已切到新服务。
  - 旧 `Status=60` 模板方法已标记 `Deprecated`，无路由调用。
  - 冲突检测在事务内通过 `checkConflictTx` / `checkBedConflictTx` 完成。
  - 业务校验错误返回 400 (TemplateBusinessError)，内部错误返回 500。

- **前端**: 已适配新接口结构。
  - `ScheduleTemplateList`: 展示 name/scope/wardId/version/itemCount，支持编辑。
  - `ScheduleTemplateEditor`: 编辑模板头信息并重新保存。
  - `ApplyTemplateModal`: 选择模板+日期+病区后应用生成草稿。
  - 暂不支持模板项增删改 UI（需阶段 4 补齐）。

- **限制**:
  - ApplyTemplate 不做自动排班算法，仅按显式模板项逐条生成。
  - 模板项无增删改 UI，新建模板入口已隐藏。
  - ShiftTiming 硬编码为 20（长期），未从模板项/PatientProfile 派生。
  - PatientPlanId 未写入。`

## 6. 状态双轨映射

| 老库 Status | 含义 | 新 RuleStatus | 含义 |
|-------------|------|--------------|------|
| 10 | 草稿 | 10 | 草稿 |
| 20 | 已确认 | 20 | 已确认 |
| 30 | 用户确认 | 50 | 透析中 |
| 40 | 用户取消 | 70 | 已取消 |
| 50 | 排班取消 | 70 | 已取消 |
| 60 | 转出/模板 | 60 | 已完成 |
| (无) | (无) | 80 | 缺席（新） |
| (无) | (无) | 0 | 待排（新） |

注意：上表中老状态与新 RuleStatus 的映射是建议方案，需业务最终确认后再执行。

## 7. 未确认项（待业务澄清）

- A/B/C 分区与 `PatientType`/`InfectionType` 的映射
- 患者传染病标识来源字段
- `Plan_PatientPlan.OddWeekFrequency/EvenWeekFrequency` 到五种频率的精确映射
- HDF 日的指定方式（医嘱 vs 护士长配置）
- 奇周/偶周频率不一致的患者是否存在
- Schdule_Bed 是否严格等价于机器
- CRRT 机器是否存在于 `Schedule_Bed`
- 老库 `Status=30/40/50/60` 与新状态最终映射
- `Schedule_PatientShift.BedId/WardId/ShiftId` 真实 NOT NULL 约束
