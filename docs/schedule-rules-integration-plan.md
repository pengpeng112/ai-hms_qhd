# 透析排班新规则融合详细开发计划

## 1. 总体结论

在“继续使用老血透库主表”的前提下,允许新增扩展表后,`透析排班子程序规则规范_v1(1).md` 和 `新规则决策记录(1).md` 中的功能可以完整实现。

推荐方案不是替换老库排班主链路,而是采用“老表主记录 + 新表补规则”的增量架构:

- 真实排班主记录继续写 `Schedule_PatientShift`。
- 基础资源继续复用 `Schedule_Ward`、`Schedule_Bed`、`Schedule_Shift`。
- 患者和方案继续复用 `Register_PatientInfomation`、`Plan_PatientPlan`。
- 老库没有表达能力的规则字段全部落到新增 `Schedule_*Ext` 或独立新表。
- 后端新增适配层,把老库 Bed/Ward/Shift/PatientShift 组装成算法需要的 Board 快照。
- 参考 `docs/排班功能说明/透析排班-backend` 的 `sched.Board`、`sched.Engine`、`repo.LoadBoard`、`service.GenerateSchedule`、扰动处理服务,但不要直接复制模型字段到老库主表。

最终目标是:老系统已有功能不破坏,新规则完整落地,并且后续可以逐步从扩展表演进到正式新模型。

## 2. 关键原则

- 老库主表优先:实际排班仍以 `Schedule_PatientShift.Id` 作为业务主键。
- 扩展表不反向破坏老表:新字段通过 `PatientShiftId`、`BedId`、`WardId`、`PatientId` 关联。
- 算法只读骨架,只分机位:患者分区、频率、固定班次、治疗模式、HDF 日都来自人工维护表或医嘱映射。
- 主方案 -> 备选方案 -> 冲突队列:排不开不硬排,生成建议或报警。
- 模板必须独立存储:禁止继续用 `Schedule_PatientShift.Status=60` 表示模板。
- CRRT 作为特殊形态:不走三班网格、不进模板,只占用 CRRT 机和起止时间。
- 所有 SQL 使用老库大小写字段规范,表列名需要双引号。
- 后端禁止 AutoMigrate;新增表必须以人工审核 SQL 脚本方式落库。
- 所有新增表必须显式设计索引、唯一约束和迁移脚本;不能只依赖 GORM tag。
- 所有批量生成/迁移操作必须先 preview 再 apply,并有并发保护。
- v1.1 原型中的 AutoMigrate 只适用于独立演示项目;本项目继续禁止 AutoMigrate,新增表只能通过人工审核 SQL 脚本执行。
- 写接口不能存在“未鉴权默认超级用户”行为;本地联调如需放行必须显式配置,生产必须关闭。
- API 错误响应必须统一格式并对 5xx 脱敏;真实错误写服务端日志,避免泄露表结构/SQL。
- 排班系统不得自动解决冲突或批量自动挪位;只允许“系统给建议 + 人工确认采纳”。

## 3. 当前老库可保留能力

### 3.1 继续复用的老表

| 老表 | 继续承担的职责 | 说明 |
| --- | --- | --- |
| `Schedule_Ward` | 病区基础信息 | 新规则 A/B/C、子区等放扩展表 |
| `Schedule_Bed` | 机器/机位基础实体 | 每个 Bed 适配为一台 Machine-like 资源 |
| `Schedule_Shift` | 班次 | 上午/下午/晚继续复用 |
| `Schedule_PatientShift` | 实际排班主记录 | 继续写 `TreatmentTime`、`BedId`、`WardId`、`ShiftId` |
| `Register_PatientInfomation` | 患者主档 | 患者姓名、门诊/住院等基础信息来源 |
| `Plan_PatientPlan` | 医嘱/方案来源 | 频率、治疗模式可做初始映射,不完整部分进人工确认 |
| `Treatment_Treatment` / `Treatment_Action` | 编辑保护 | 已开始治疗后不可改 |

### 3.2 老表原有功能保持

- 周视图:继续从 `Schedule_Ward`、`Schedule_Bed`、`Schedule_Shift`、`Schedule_PatientShift` 聚合。
- 手工排班:现有 `PatientShiftService.Create/Update/Delete/Swap/BatchSave` 保留。
- 基础冲突:同患者同日同班、同床同日同班继续有效。
- 历史保护和治疗中保护继续有效。
- 前端现有排班页面先不替换,逐步加预检、生成、冲突面板。

## 4. 新增表设计

以下表名沿用 `Schedule_` 前缀,但作为新增扩展表。字段命名使用老库风格 `PascalCase`。

### 4.1 `Schedule_WardExt` 病区扩展

用途: 给老库 `Schedule_Ward` 增加 A/B/C 分区和子区树能力。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `WardId` | bigint not null unique | 关联 `Schedule_Ward.Id` |
| `ZoneType` | varchar(8) not null | `A/B/C` |
| `ParentWardId` | bigint null | 父病区,用于 A1/A2 子区 |
| `IsSubZone` | boolean default false | 是否子区 |
| `Note` | varchar(512) | 备注 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

落地规则:

- A 区 = 门诊区,B 区 = 住院区,C 区 = 全警戒区。
- 初始映射不能猜,需提供后台配置页面或初始化导入表。
- `Schedule_Ward.PatientType/InfectionType` 只作为辅助展示,不要作为唯一规则源。

### 4.2 `Schedule_BedMachineExt` 床位/机器扩展

用途: 把老库 `Schedule_Bed` 适配为新规则中的 Machine。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `BedId` | bigint not null unique | 关联 `Schedule_Bed.Id` |
| `MachineCode` | varchar(64) | 机器编号,可默认使用 Bed.Name/Id |
| `MachineType` | varchar(8) not null | `HD/HDF/CRRT` |
| `SupportedModes` | varchar(64) not null | 支持模式,如 `HD`、`HD,HDF,HF`、`CRRT` |
| `PositionIndex` | int not null | 物理排位,默认可取 `Schedule_Bed.Sort` |
| `IsDisabled` | boolean default false | 扩展层停用 |
| `LegacyBedName` | varchar(256) | 冗余旧床名,便于审计 |
| `Note` | varchar(512) | 备注 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

落地规则:

- 算法中的 `MachineId` 在本项目内映射为 `BedId`。
- HDF/CRRT 能力必须由该表明确配置,不能从床名猜。
- `MachineType` 是主机型,`SupportedModes` 是能力矩阵;算法以 `SupportedModes` 判定是否支持某次治疗模式。
- `PositionIndex` 用于固定机位、就近、集中连片、组团挑机。

### 4.3 `Schedule_PatientProfile` 患者排班骨架

用途: 存新规则要求的人工权限属性。算法只读该表。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `PatientId` | bigint not null unique | 关联 `Register_PatientInfomation.Id` |
| `ZoneTag` | varchar(8) not null | `A/B/C` |
| `HomeWardId` | bigint null | 归属病区,关联 `Schedule_Ward.Id` |
| `FreqPattern` | smallint not null | 10 一三五,20 二四六,30 周二四,40 周四,90 临时 |
| `ShiftId` | bigint null | 固定班次 |
| `DefaultMode` | varchar(8) not null default `HD` | 默认治疗模式 |
| `HdfEnabled` | boolean default false | 是否每两周 HDF |
| `HdfWeekday` | smallint null | 1 周一 ... 6 周六 |
| `HdfWeekParity` | smallint null | 0 偶,1 奇 |
| `FixedHdBedId` | bigint null | 固定 HD 床/机器 |
| `FixedHdfBedId` | bigint null | 固定 HDF 床/机器 |
| `IsAdmissionRejected` | boolean default false | 是否拒收,拒收不生成排班 |
| `EffectiveFrom` | date null | 骨架生效日 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

落地规则:

- 分区标签变更必须受医生/护士长权限控制。
- 班次变更必须受护士长/主班护士权限控制。
- 频率/治疗模式/HDF 来源优先医嘱,但要落到 Profile 后供算法只读。
- `HdfWeekParity` 可由系统按 HDF 机负载初次分配,之后尽量固定。

### 4.4 `Schedule_PatientShiftExt` 排班记录扩展

用途: 给老库 `Schedule_PatientShift` 补齐新状态机正交字段和三级确认字段。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `PatientShiftId` | bigint not null unique | 关联 `Schedule_PatientShift.Id` |
| `DialysisMode` | varchar(8) not null default `HD` | 按次治疗模式 `HD/HDF/HF/CRRT` |
| `SourceType` | smallint not null default 10 | 10 常规,20 临时 |
| `RecordForm` | smallint not null default 10 | 10 规律,20 CRRT |
| `Confirm1At` | timestamp null | 第一次确认时间 |
| `Confirm2At` | timestamp null | 第二次确认时间 |
| `Confirm3At` | timestamp null | 第三次确认时间 |
| `Confirm1By` | bigint null | 第一次确认人 |
| `Confirm2By` | bigint null | 第二次确认人 |
| `Confirm3By` | bigint null | 第三次确认人 |
| `IsBorrowedSlot` | boolean default false | 是否借用请假/缺席空位 |
| `BorrowedFromShiftId` | bigint null | 借用来源排班,关联被取消/缺席记录 |
| `IsLocked` | boolean default false | 是否业务锁定 |
| `CancelReason` | varchar(256) | 取消/缺席原因 |
| `SourceTemplateItemId` | bigint null | 来源模板项 |
| `SourceTemplateVersion` | int null | 来源模板版本 |
| `RuleStatus` | smallint not null default 10 | 新规则状态,默认草稿 |
| `ApprovedBy` | bigint null | 临时透析/特殊排班审批人 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

落地规则:

- 老库 `Schedule_PatientShift.Status` 当前 10/20/30/40/50/60 含义与新规则 0/10/20/50/60/70/80 不一致,不能直接替换。
- 短期保留老状态映射,在扩展层用 `RuleStatus` 表达新语义。
- 推荐新增状态映射函数: `MapScheduleStatusLegacyToRule`、`MapScheduleStatusRuleToLegacy`。
- 三级确认不新增三个状态,只写 `Confirm1At/2At/3At`。
- 服务层必须统一走 `GetEffectiveScheduleStatus(shift, ext)` 获取状态,禁止业务代码直接混用老状态和新状态。

### 4.5 `Schedule_ScheduleTemplate` 模板头

用途: 替代 `Status=60` 伪模板。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `Name` | varchar(128) not null | 模板名 |
| `Scope` | varchar(8) | `ALL/A/B/C` 或 ward 范围 |
| `WardId` | bigint null | 单病区模板可填 |
| `IsActive` | boolean default true | 是否生效 |
| `Version` | int default 1 | 版本 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

### 4.6 `Schedule_ScheduleTemplateItem` 模板项

用途: 存稳定病人的时间骨架和固定机位快照。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `TemplateId` | bigint not null | 模板头 |
| `PatientId` | bigint not null | 患者 |
| `ZoneTag` | varchar(8) not null | A/B/C |
| `WardId` | bigint null | 病区 |
| `ShiftId` | bigint null | 固定班次 |
| `FreqPattern` | smallint not null | 频率模式 |
| `FixedHdBedId` | bigint null | 固定 HD 机器 |
| `FixedHdfBedId` | bigint null | 固定 HDF 机器 |
| `HdfEnabled` | boolean default false | 是否 HDF |
| `HdfWeekday` | smallint null | HDF 星期 |
| `HdfWeekParity` | smallint null | HDF 奇偶周 |
| `TemplateVersion` | int not null default 1 | 模板项版本快照 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

落地规则:

- 一个患者在一个模板中建议只有一条模板项。
- 模板生成时复制骨架,不是复制某天的状态 60 排班。
- 现有 `ListTemplates/SaveTemplate/ApplyTemplate` 需要迁移到新表,旧接口可保持路径不变但实现改写。
- 模板 `Scope` 只用于筛选模板适用范围;生成时仍以 `PatientProfile.ZoneTag/HomeWardId` 作为机位分配约束。
- `Scope` 与模板项患者 `ZoneTag` 不匹配时,生成预检必须报 `ZONE_MISMATCH`,默认跳过该模板项。

### 4.7 `Schedule_ConflictQueue` 冲突/待处理队列

用途: 承载“主方案 -> 备选 -> 报警人工裁决”。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `PatientId` | bigint null | 患者 |
| `ScheduleDate` | date null | 日期 |
| `ShiftId` | bigint null | 班次 |
| `WardId` | bigint null | 病区 |
| `ConflictType` | varchar(24) not null | 冲突类型 |
| `Severity` | smallint default 10 | 10 提示,20 报警 |
| `Detail` | text | 详情 |
| `SuggestedDate` | date null | 建议日期 |
| `SuggestedShiftId` | bigint null | 建议班次 |
| `SuggestedBedId` | bigint null | 建议床/机器 |
| `SuggestedPatientShiftId` | bigint null | 建议草稿记录 |
| `Status` | smallint default 0 | 0 待处理,10 已处理,20 已忽略 |
| `ResolvedBy` | bigint null | 处理人 |
| `ResolvedAt` | timestamp null | 处理时间 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

冲突类型建议:

- `NO_MACHINE`
- `HDF_NO_MACHINE`
- `WARD_FULL` 或 `NO_MACHINE`
- `NEW_PATIENT_UNPLACED`
- `MACHINE_OUTAGE`
- `HOLIDAY_REPLAN`
- `PLAN_CHANGE`
- `MAKEUP_SUGGEST`
- `SLOT_SPILLED`
- `MISSING_PROFILE`
- `MISSING_MACHINE_TYPE`
- `ZONE_MISMATCH`

说明:

- 决策 22 已取消“加床”,因此排满时优先使用 `NO_MACHINE` 或 `SLOT_SPILLED`。
- 如保留 `WARD_FULL`,其语义只能是“该区该班所有可用机器均已占用”,不能表示加床。

### 4.8 `Schedule_MachineOutage` 设备停机时段

用途: 冻结机器在某段时间的机位。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `BedId` | bigint not null | 对应老库 `Schedule_Bed.Id` |
| `StartAt` | timestamp not null | 开始 |
| `EndAt` | timestamp null | 结束,NULL 表示未定 |
| `ShiftId` | bigint null | 可选;为空表示整天/全时段停机,非空表示仅影响某班次 |
| `OutageType` | smallint not null | 10 临时,20 长期/报废 |
| `Reason` | varchar(512) | 原因 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

### 4.9 `Schedule_Calendar` 机构日历

用途: 非透析日、假日值班模式。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `CalDate` | date not null | 日期 |
| `IsDialysisDay` | boolean not null | 是否透析日 |
| `HolidayMode` | smallint default 0 | 0 正常,10 全院停,20 假日值班 |
| `Note` | varchar(256) | 备注 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

不要在 `Schedule_Calendar` 中用逗号分隔或 JSON 存开放资源。假日值班开放范围必须使用关联表:

#### `Schedule_CalendarOpenWard`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `CalendarId` | bigint not null | 关联 `Schedule_Calendar.Id` |
| `WardId` | bigint not null | 开放病区 |

#### `Schedule_CalendarOpenBed`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `CalendarId` | bigint not null | 关联 `Schedule_Calendar.Id` |
| `BedId` | bigint not null | 开放机器/床位 |

### 4.10 `Schedule_CrrtSession` CRRT 占用

用途: CRRT 不走三班网格,记录占用区间。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `PatientShiftId` | bigint not null unique | 关联主排班 |
| `BedId` | bigint not null | CRRT 机器/床 |
| `StartAt` | timestamp not null | 开始 |
| `EndAt` | timestamp null | 结束 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

落地规则:

- `EndAt IS NULL` 表示进行中,服务层重叠检测必须统一按远未来时间处理,例如 `COALESCE("EndAt", '2999-01-01')`。
- CRRT 主记录仍写 `Schedule_PatientShift`,但必须确认老库 `BedId/WardId` 是否 NOT NULL;如是,需为 C 区 CRRT 建立专用 Bed 主记录。
- 周视图必须额外查询与目标周有时间区间交叠的 `Schedule_CrrtSession`,不能只按 `TreatmentTime` 开始日期展示。

### 4.11 `Schedule_PlanChange` 方案变更记录

用途: 生效日后影响评估和审计。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `PatientId` | bigint not null | 患者 |
| `ChangeType` | varchar(16) not null | `FREQ/MODE/SHIFT/ZONE/HDF` |
| `OldValue` | varchar(64) | 旧值 |
| `NewValue` | varchar(64) | 新值 |
| `EffectiveDate` | date not null | 生效日 |
| `AffectedCount` | int default 0 | 影响记录数 |
| `ProcessedAt` | timestamp null | 处理时间 |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

### 4.12 `Schedule_TenantSetting` 排班配置

用途: 奇偶周锚点、预警阈值、生成周期默认值。

建议字段:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Id` | bigint PK | 主键 |
| `TenantId` | bigint not null | 租户 |
| `SettingKey` | varchar(64) not null | 配置键 |
| `SettingValue` | varchar(256) not null | 配置值 |
| `SettingType` | varchar(16) not null | `date/int/string/bool` |
| `CreatorId` | bigint | 创建人 |
| `CreateTime` | timestamp | 创建时间 |
| `LastModifyTime` | timestamp | 修改时间 |

配置键建议:

- `OddEvenWeekAnchorMonday`: 默认 `2025-01-06`
- `LowSlotWarnThreshold`: 余位预警阈值
- `DraftWeeks`: 默认 2
- `LastGenerationVersion`: 生成乐观锁版本号
- `SpillHorizonDays`: 排满顺延窗口天数,默认 14

服务层必须按 `SettingType` 校验值格式,例如锚点必须是 `YYYY-MM-DD`,预警阈值必须是整数。

## 5. 状态兼容策略

### 5.1 老状态现状

当前代码注释中的老库状态:

| 老库 Status | 当前含义 |
| --- | --- |
| 10 | 草稿 |
| 20 | 已确认 |
| 30 | 用户确认 |
| 40 | 用户取消 |
| 50 | 排班取消 |
| 60 | 转出人员/当前被模板复用 |

### 5.2 新规则状态

| 新规则状态 | 含义 |
| --- | --- |
| 0 | 待排 |
| 10 | 草稿 |
| 20 | 已确认 |
| 50 | 透析中 |
| 60 | 已完成 |
| 70 | 已取消 |
| 80 | 缺席 |

### 5.3 推荐兼容方式

- 不直接改变老库 `Status` 的既有业务含义。
- 主表继续使用老状态,保证旧页面/治疗流程不崩。
- 新规则状态在服务层统一 DTO 输出。
- 新增排班必须同步写 `Schedule_PatientShiftExt.RuleStatus`。
- 读取排班时以 `RuleStatus` 优先,没有扩展记录时回退老状态映射。
- 取消/缺席等新状态必须写 `RuleStatus` 和 `CancelReason`,不要只依赖老库 `Status`。
- `RuleStatus` 使用 `smallint not null default 10`,避免 Go 零值 `0` 被误解为“待排”。
- 所有排班查询 DTO 必须同时返回 `legacyStatus` 与 `ruleStatus`,便于迁移期排查。

### 5.4 Status=60 模板迁移策略

迁移前必须审计所有 `Schedule_PatientShift.Status=60` 记录,因为它既可能是“转出人员”,也可能是当前模板伪记录。

建议迁移步骤:

1. 查询 `Status=60` 记录,按 `TenantId/WardId/PatientId/PatientPlanId/ShiftTiming/TreatmentTime/CreateTime` 导出审计表。
2. 用人工规则标记 `IsTemplateCandidate`。不能仅凭 `Status=60` 自动迁移。
3. 将确认属于模板的记录写入 `Schedule_ScheduleTemplate` 和 `Schedule_ScheduleTemplateItem`。
4. 迁移后禁用旧 `SaveTemplate` 对 `Schedule_PatientShift` 的写入和删除。
5. 暂时保留旧 `ListTemplates` 只读迁移入口,并在前端提示“旧模板已迁移/待迁移”。
6. 确认无旧模板依赖后,删除或隐藏旧模板入口。

迁移期间禁止执行全租户 `DELETE WHERE "Status" = 60`。

## 6. 后端目标架构

参考原型 `透析排班-backend` 的分层,但在当前项目中落地为以下文件。

### 6.1 新增模型文件

- `internal/models/schedule_ext.go`
  - `WardExt`
  - `BedMachineExt`
  - `PatientProfile`
  - `PatientShiftExt`
  - `ScheduleTemplate`
  - `ScheduleTemplateItem`
  - `ConflictQueue`
  - `MachineOutage`
  - `Calendar`
  - `CrrtSession`
  - `PlanChange`
  - `TenantSetting`
  - `CalendarOpenWard`
  - `CalendarOpenBed`

### 6.2 新增算法包或服务文件

推荐先放在 `internal/services`,减少包迁移成本:

- `schedule_rule_constants.go`
  - 状态、频率、机型、模式、冲突类型、配置键常量。
  - 参考原型 `internal/sched/constants.go`。

- `schedule_board_service.go`
  - 从老库 + 扩展表加载 Board 快照。
  - 参考原型 `internal/repo/repo.go` 和 `internal/sched/board.go`。

- `schedule_engine_service.go`
  - HDF/HD 两轮分配。
  - 固定机位、就近、集中连片、组团打分。
  - 排满顺延。
  - 参考原型 `internal/sched/engine.go`。

- `schedule_template_service.go`
  - 独立模板 CRUD。
  - 模板生成 2/4 周草稿。
  - 替代 `PatientShiftService.ListTemplates/SaveTemplate/ApplyTemplate` 的 Status=60 实现。

- `schedule_conflict_service.go`
  - 冲突队列列表、处理、忽略、按建议应用。
  - 参考原型 `raiseConflictDB`。

- `schedule_ops_service.go`
  - 三级确认、取消、缺席、移动。
  - 参考原型 `internal/service/ops_service.go`。
  - 必须实现二次确认的非工作日跳转:周一排班的二次确认默认落在前一个可透析日,通常是周六。

- `schedule_perturb_service.go`
  - 临时透析、停机迁移、假日挪班、方案变更。
  - 参考原型 `perturb_service.go`。

- `schedule_crrt_service.go`
  - CRRT 插入和列表。
  - 参考原型 `crrt_service.go`。

- `schedule_diff_service.go`
  - 应排/已排差异检测。
  - 参考原型 `diff_service.go`。

### 6.3 新增 API Handler

- `internal/api/v1/schedule_rule_handler.go`
  - 生成、预览、差异、冲突、确认、临时透析、CRRT、停机、日历、方案变更。

已有 `schedule_handler.go` 保留,逐步把新能力挂到 `/api/v1/schedule/*` 下。

## 7. Board 快照设计

### 7.1 Board 输入

Board 加载时需要:

- Wards: `Schedule_Ward` + `Schedule_WardExt`
- Machines: `Schedule_Bed` + `Schedule_BedMachineExt`
- Shifts: `Schedule_Shift`
- Existing shifts: `Schedule_PatientShift` + `Schedule_PatientShiftExt`
- Profiles: `Schedule_PatientProfile`
- Outages: `Schedule_MachineOutage`
- Calendar: `Schedule_Calendar`

### 7.2 Machine 适配

内部 DTO:

```go
type ScheduleMachine struct {
    ID            int64 // 等于 BedId
    BedID         int64
    WardID        int64
    Code          string
    Name          string
    MachineType   string // HD/HDF/CRRT
    PositionIndex int
    IsDisabled    bool
}
```

注意:

- 内部算法叫 Machine,落库仍写 `BedId`。
- CRRT 也用 `BedId` 指向 CRRT 机器。

### 7.3 PatientShift 适配

内部 DTO:

```go
type ScheduleSlot struct {
    ID              int64 // Schedule_PatientShift.Id
    PatientID       int64
    ScheduleDate    time.Time // 来自 TreatmentTime 日期部分
    ShiftID         *int64
    WardID          int64
    BedID           *int64
    PatientPlanID   *int64
    LegacyStatus    int
    RuleStatus      int16
    DialysisMode    string
    SourceType      int16
    RecordForm      int16
    Confirm1At      *time.Time
    Confirm2At      *time.Time
    Confirm3At      *time.Time
    IsBorrowedSlot  bool
}
```

### 7.4 占用规则

- 有效占用:草稿、已确认、透析中、已完成。
- 不占用:已取消、缺席。
- 患者同日同班已有任意记录时,生成时默认跳过,避免复活已取消记录。
- 设备停机覆盖某日期时,该机器该日所有班次冻结;后续可细化到班次时间段。

## 8. 核心算法落地

### 8.1 频率模式

常量:

- 10: 一三五
- 20: 二四六
- 30: 每周二、四
- 40: 每周四
- 90: 临时

实现:

- `IsDue(freq, date)` 返回该日是否应排。
- 周日默认非透析日,除非 `Schedule_Calendar` 特别开放。
- 假日非透析日自动跳过。

### 8.2 HDF 奇偶周

实现:

- 从 `Schedule_TenantSetting.OddEvenWeekAnchorMonday` 读取锚点。
- 默认 `2025-01-06`。
- `WeekParity(date)` = `(MondayOf(date)-anchor)/7 % 2`。
- `AssignHdfWeekParity` 按 `HdfWeekday + parity` 统计负载,新患者放到较轻侧。

### 8.3 治疗模式决策

- 默认 `HD`。
- 患者 `HdfEnabled=true` 且日期等于 `HdfWeekday` 且奇偶周匹配时,本次为 `HDF`。
- CRRT 不走该逻辑。

### 8.4 机型能力

- HD 机只支持 HD。
- HDF 机支持 HD/HDF/HF。
- CRRT 机只支持 CRRT。

### 8.5 规律排班两轮分配

每个 `日期 x 病区 x 班次` 单元:

1. 先筛出该单元应排的 HDF 次。
2. HDF 次先排 HDF 机。
3. HDF 优先固定 HDF 机,不可用则就近 HDF 机,仍无则冲突 `HDF_NO_MACHINE`。
4. 再筛出 HD 次。
5. HD 优先固定 HD 机。
6. 固定不可用时找同区空闲 HD 机。
7. HD 机全满后才溢出到空闲 HDF 机。
8. 仍无空位则顺延到后续班次/后续透析日,生成提示冲突 `SLOT_SPILLED` 或报警 `NO_MACHINE`。

### 8.6 挑机评分

评分建议:

- 相邻机位已占用加分,用于集中连片和组团。
- 靠近固定机位加分或距离扣分。
- 同方案/同班次相邻加分可作为第二阶段增强。
- 不做负载均衡。

### 8.7 新病人分配

逻辑:

1. 根据 Profile 展开未来 2/4 周应排日期。
2. 首选找到所有应排日同班次都空闲的同一台机器。
3. 找不到则逐次分配。
4. 仍失败写 `NEW_PATIENT_UNPLACED`。

### 8.8 排满顺延

逻辑:

- 先同日后续班次。
- 再后续透析日,优先原班次,再其他班次。
- 窗口默认 14 天。
- 命中后只生成建议草稿或冲突建议,不自动强制确认。

## 9. 功能模块开发计划

### 阶段 0: 数据库脚本和模型准备

目标: 新表可人工建表,后端有 GORM 模型但不 AutoMigrate。

任务:

- 新增 `docs/sql/schedule_extension_tables.sql`。
- SQL 包含上述新增表、唯一索引和常用查询索引。
- SQL 必须显式包含以下核心索引:
  - `Schedule_WardExt("TenantId", "WardId") unique`
  - `Schedule_WardExt("TenantId", "ZoneType")`
  - `Schedule_BedMachineExt("TenantId", "BedId") unique`
  - `Schedule_BedMachineExt("TenantId", "MachineType")`
  - `Schedule_PatientProfile("TenantId", "PatientId") unique`
  - `Schedule_PatientProfile("TenantId", "ZoneTag", "ShiftId")`
  - `Schedule_PatientShiftExt("TenantId", "PatientShiftId") unique`
  - `Schedule_PatientShiftExt("TenantId", "RuleStatus")`
  - `Schedule_ScheduleTemplate("TenantId", "IsActive")`
  - `Schedule_ScheduleTemplateItem("TenantId", "TemplateId", "PatientId") unique`
  - `Schedule_ConflictQueue("TenantId", "Status", "ScheduleDate")`
  - `Schedule_MachineOutage("TenantId", "BedId", "StartAt", "EndAt")`
  - `Schedule_Calendar("TenantId", "CalDate") unique`
  - `Schedule_CalendarOpenWard("TenantId", "CalendarId", "WardId") unique`
  - `Schedule_CalendarOpenBed("TenantId", "CalendarId", "BedId") unique`
  - `Schedule_CrrtSession("TenantId", "BedId", "StartAt", "EndAt")`
  - `Schedule_PlanChange("TenantId", "PatientId", "EffectiveDate")`
- 新增 `internal/models/schedule_ext.go`。
- 增加注释说明禁止 AutoMigrate。
- 增加 `docs/schedule-extension-field-mapping.md`,记录老表到新规则字段映射。
- 增加 `docs/sql/schedule_status60_template_audit.sql`,专门审计并迁移旧 Status=60 模板。

验收:

- `go test ./internal/services -count=1 -short` 通过。
- 不修改 `internal/database/migrate.go` 的禁止 AutoMigrate 策略。
- SQL 脚本可重复执行或至少带存在性检查,避免重复建表失败。

### 阶段 0.5: 工程化与安全基线

目标: 吸收 v1.1 原型的工程化修正,先把可观测性、安全和运行稳定性补齐。

后端任务:

- 健康检查:
  - 新增或复用 `GET /health`。
  - 返回数据库连通性,正常返回 `{"status":"ok","db":"ok"}`,数据库不可达返回 503。
- 连接池:
  - 支持环境变量配置数据库连接池。
  - 建议变量: `DB_MAX_OPEN_CONNS`、`DB_MAX_IDLE_CONNS`、`DB_CONN_MAX_LIFE`、`DB_CONN_MAX_IDLE`。
  - 默认值可参考 v1.1: 50/10/30分钟/5分钟,但需结合当前生产环境确认。
- 统一错误响应:
  - 统一响应形态为 `{"code": number, "error": string}`。
  - 5xx 对外脱敏,只返回通用错误;真实错误写服务端日志。
  - 4xx 可返回业务可读错误,但不能包含 SQL、表名、堆栈。
- 请求审计:
  - 新增或复用审计中间件,记录方法、路径、状态码、耗时、租户、用户/角色、客户端 IP。
  - 只记录必要元数据,不要记录敏感请求体。
- 鉴权默认值:
  - 写接口必须显式鉴权。
  - 不允许“未带角色/未登录默认超级用户”。
  - 如需本地联调放行,必须通过显式环境变量控制,并在生产配置检查中禁止启用。
- 参数校验:
  - 所有日期、weeks、patientId、wardId、bedId、mode、role 入参必须做格式和值域校验。
  - 可后续引入 validator 标签库,但第一版至少手写校验关键写接口。

前端任务:

- 写操作失败时显示红色错误提示,成功显示绿色提示。
- 401/403 与业务冲突错误要有明确提示。
- 角色/权限状态在页面显著位置展示,避免测试时误判。

验收:

- 停掉数据库后 `/health` 返回 503。
- 未鉴权调用写接口返回 401 或当前系统统一未登录错误。
- 低权限角色调用写接口返回 403。
- 人为制造数据库错误时,客户端看不到 SQL/表名/堆栈。
- 审计日志能定位一次生成/确认/移动请求的租户、用户、角色和耗时。

### 阶段 1: 基础配置和维护页面 API

目标: 补齐算法前置数据。

后端任务:

- Ward 扩展 CRUD: 配置 A/B/C、子区。
- BedMachine 扩展 CRUD: 配置 HD/HDF/CRRT、物理排位。
- PatientProfile CRUD: 配置患者分区、频率、班次、HDF、固定机。
- TenantSetting CRUD: 配置奇偶周锚点、预警阈值。
- Calendar CRUD: 配置非透析日、假日值班。
- 配置服务必须校验 `SupportedModes` 与 `MachineType` 一致性,例如 `MachineType=HDF` 默认支持 `HD,HDF,HF`。

前端任务:

- 在排班配置中新增:
  - 病区规则配置
  - 机器能力配置
  - 患者排班骨架配置
  - 日历配置
  - 排班参数配置

验收:

- 可以完整配置一个 A 区、一个 B 区、一个 C 区。
- 可以给每个 Bed 配置机器类型。
- 可以给每个 Bed 配置支持模式,且 HDF 机能支持 HD/HDF/HF。
- 可以给患者配置固定班次和频率。
- 没有配置 MachineType 的 Bed 在预检中提示 `MISSING_MACHINE_TYPE`。

### 阶段 2: Board 快照和只读预检

目标: 在不写排班的情况下,验证数据是否足以执行新规则。

后端任务:

- 实现 `LoadScheduleBoard(tenantID,start,end)`。
- 实现占用计算、余位计算、停机冻结、日历跳过。
- Board 加载必须使用批量查询和内存 map,禁止按患者/床位逐条查库。
- Board 预检需要检查 CRRT 专用 Bed/Ward 是否存在。
- 实现预检接口:
  - `GET /api/v1/schedule/rules/board?date=YYYY-MM-DD`
  - `POST /api/v1/schedule/rules/precheck`
  - `GET /api/v1/schedule/conflicts`
- 预检返回:
  - 缺 Profile 的患者
  - 缺机器类型的 Bed
  - 频率无法映射的患者
  - 区域不匹配
  - 每区每班余位

前端任务:

- 排班页面新增“规则预检”按钮。
- 显示缺失配置、余位、冲突。

验收:

- 不写任何 `Schedule_PatientShift`。
- 可在页面看到每区每班容量和余位。
- 可识别 HDF 机不足、Bed 未配置机型。
- 4 周、3 区、100 床、200 患者的预检应在可接受时间内完成,建议记录耗时日志。

### 阶段 3: 独立模板替代 Status=60

目标: 模板不再污染真实排班。

后端任务:

- 实现 `ScheduleTemplateService`。
- 改写现有模板接口,路径可保持:
  - `GET /api/v1/schedule/template`
  - `POST /api/v1/schedule/template/save`
  - `POST /api/v1/schedule/template/apply`
- 新实现读写 `Schedule_ScheduleTemplate` 和 `Schedule_ScheduleTemplateItem`。
- 保留旧 Status=60 查询只作为一次性迁移工具,不要继续写。
- 增加迁移接口或脚本: 把当前 Status=60 模板转成新模板项,人工确认后禁用旧模板写入。
- 新模板保存必须使用版本号;模板修改时 `Version+1`,模板项同步写 `TemplateVersion`。

前端任务:

- `ScheduleTemplateList.tsx` 改为展示独立模板。
- 保存模板时提交模板项,不创建 Status=60 排班。
- 应用模板时生成草稿或预览,不直接确认。

验收:

- 保存模板不再删除 `Schedule_PatientShift.Status=60`。
- 应用模板生成 `Status=10` 草稿主记录 + `Schedule_PatientShiftExt` 扩展记录。
- 旧 `ApplyTemplate` 不能再直接生成 `Status=20` 已确认记录。

### 阶段 4: 草稿生成引擎

目标: 基于模板和 Profile 自动生成未来 2/4 周草稿。

后端任务:

- 实现 `GenerateSchedule(tenantID,start,weeks)`。
- 参考原型 `service.GenerateSchedule`。
- 读取 active template items。
- 加载 Board。
- 分配 HDF 奇偶周。
- 展开透析日。
- 运行两轮分配。
- 事务写入:
  - `Schedule_PatientShift`
  - `Schedule_PatientShiftExt`
  - `Schedule_ConflictQueue`
  - 必要时更新 `PatientProfile.HdfWeekParity` 和模板项奇偶周。
- 生成前读取 `Schedule_TenantSetting.LastGenerationVersion`,写入时做乐观锁检查,防止两人并发生成重复草稿。
- 写入前必须二次校验目标患者/床位/日期/班次是否已被其他事务占用。

接口:

- `POST /api/v1/schedule/generate-preview`
  - 只返回草稿建议和冲突,不写库。
- `POST /api/v1/schedule/generate`
  - 写入草稿。

验收:

- 2 周/4 周生成可选。
- 草稿状态为老库兼容草稿状态,扩展表写 `DialysisMode/SourceType/RecordForm`。
- HDF 病人每两周一次替换,总次数不增加。
- 无 HDF 机时不硬排,写冲突。
- 并发触发生成时,其中一个请求必须失败并提示重新预检。

### 阶段 5: 三级确认和编辑保护

目标: 落实确认工作流。

后端任务:

- `POST /api/v1/schedule/confirm-plan`
  - 第一次确认,整盘草稿生效。
- `POST /api/v1/schedule/confirm-day`
  - level=2 次日前确认。
  - level=3 当日确认。
- `POST /api/v1/schedule/shifts/:id/cancel`
  - 提前请假/计划取消。
- `POST /api/v1/schedule/shifts/:id/absent`
  - 当日缺席。
- `POST /api/v1/schedule/shifts/:id/move`
  - 移床/换班/改日期。

规则:

- 历史日期不可改。
- 已开始治疗不可改。
- Confirm2/Confirm3 后遇方案变化不自动改,写冲突。
- level=2 的“次日前确认”必须按 `Schedule_Calendar` 跳过非透析日/非工作日;周一排班默认在前一个周六确认。
- 移动时校验目标机器能力、目标床位占用、患者双排。

验收:

- 确认时间戳写入 `Schedule_PatientShiftExt`。
- 取消/缺席机位可被临时透析借用。
- 移床到不支持模式的机器返回冲突。
- 缺席记录应写 `RuleStatus=80` 和 `CancelReason`,并且临时透析借用时写 `BorrowedFromShiftId`。

### 阶段 6: 冲突队列闭环

目标: 把所有排不开的场景变成可处理工作项。

后端任务:

- `GET /api/v1/schedule/conflicts`
- `POST /api/v1/schedule/conflicts/:id/resolve`
- `POST /api/v1/schedule/conflicts/:id/ignore`
- `POST /api/v1/schedule/conflicts/:id/apply-suggestion`

规则红线:

- 禁止实现“一键自动解决全部冲突”或无人确认的批量挪位。
- 可以实现“对单条冲突建议一键采纳”,但必须由有权限用户点击确认后才改排。
- 批量采纳也必须展示待改排明细并二次确认,不能静默执行。
- 采纳建议前必须重新校验目标床位/班次/日期是否仍可用。

前端任务:

- 排班页面增加冲突队列面板。
- 支持按类型、严重度、日期、患者筛选。

验收:

- `HDF_NO_MACHINE`、`NO_MACHINE`、`SLOT_SPILLED` 可见。
- 人工处理后记录 `ResolvedBy/ResolvedAt/Status`。
- 采纳建议成功后,冲突记录状态变为已处理,并记录实际改动的排班 ID。
- 并发情况下,若建议目标已被占用,采纳失败并要求重新生成建议。

### 阶段 7: 扰动处理

目标: 实现规则文档 §9。

#### 7.1 临时透析

接口:

- `POST /api/v1/schedule/temporary`

规则:

- 不进模板。
- 从当前/指定班次开始找空位。
- 按模式匹配机型。
- 可借用请假/缺席空位,写 `IsBorrowedSlot=true`。
- 借用请假/缺席空位时必须写 `BorrowedFromShiftId`,用于后续恢复/审计。
- 临时透析需记录医嘱/审批来源;如暂无医嘱关联字段,至少写 `ApprovedBy`。
- 无空位写 `NO_MACHINE`。

#### 7.2 停机迁移

接口:

- `POST /api/v1/schedule/machines/:bedId/outage`

规则:

- 写 `Schedule_MachineOutage`。
- 临时停机: 找同区同班同模式替代机位,生成调整草稿或直接移动待确认。
- 停机可按整天或班次粒度冻结;`ShiftId` 为空表示全时段,非空表示指定班次。
- 长期停机: 写冲突,人工永久迁移固定机位。

#### 7.3 假日处理

接口:

- `POST /api/v1/schedule/calendar/holiday`

规则:

- 写 `Schedule_Calendar`。
- 假日值班开放范围写 `Schedule_CalendarOpenWard` / `Schedule_CalendarOpenBed`,不要写逗号字符串。
- 当日排班取消或标记需处理。
- 前后 7 天找建议空位,写 `HOLIDAY_REPLAN`。

#### 7.4 方案变更

接口:

- `POST /api/v1/schedule/patients/:id/plan-change`

规则:

- 更新 `Schedule_PatientProfile`。
- 写 `Schedule_PlanChange`。
- 生效日后未二/三次确认排班取消待重排。
- 已二/三次确认写 `PLAN_CHANGE` 冲突。

#### 7.5 补透建议

接口:

- `POST /api/v1/schedule/patients/:id/makeup`

规则:

- 系统只提示差异。
- 补不补由人工触发。
- 无空位写 `MAKEUP_SUGGEST`。

验收:

- 所有扰动都不硬排覆盖已有规律病人。
- 所有失败都有冲突队列记录。

### 阶段 8: CRRT

目标: 支持 C 区 CRRT 特殊占用。

接口:

- `POST /api/v1/schedule/crrt`
- `GET /api/v1/schedule/crrt?date=YYYY-MM-DD`

规则:

- 只能选择 MachineType=CRRT 的 Bed。
- CRRT 病人必须在 C 区。
- 不要求 ShiftId。
- 主表 `Schedule_PatientShift` 仍写一条记录,`TreatmentTime` 用开始日期。
- 扩展表 `RecordForm=20,DialysisMode=CRRT,SourceType=20`。
- `Schedule_CrrtSession` 写起止时间。
- 区间重叠时拒绝或写冲突。
- 若老库 `Schedule_PatientShift.BedId/WardId` 为 NOT NULL,CRRT 必须使用 C 区 CRRT 专用 Bed/Ward 写主记录。
- `EndAt IS NULL` 视为进行中,重叠检测必须使用 `COALESCE` 或服务层远未来时间。
- 周视图/日视图要按 `StartAt/EndAt` 区间交叠展示 CRRT,不能只看 `TreatmentTime`。

验收:

- CRRT 占用在周视图/日视图中显示。
- CRRT 机器不会被 HD/HDF 排班误用。
- 跨天 CRRT 在每个被占用日期都能显示。

### 阶段 9: 差异检测和余位预警

目标: 常驻提示应排/已排差异和容量风险。

接口:

- `GET /api/v1/schedule/diffs?weekStart=YYYY-MM-DD&weeks=2`
- `GET /api/v1/schedule/capacity?date=YYYY-MM-DD`

规则:

- 应排次数来自 `PatientProfile.FreqPattern` + 日历。
- 已排次数来自有效 `Schedule_PatientShift`。
- 缺席是否算已排需业务确认,建议算留痕但提示缺席。
- 若患者存在不对称奇偶周频率,差异检测必须按业务确认后的扩展频率字段计算,不能强行套五种固定频率。
- 余位按区、班次、日期统计。

验收:

- 能看到每位患者应排 X、已排 Y、差 Z。
- 余位低于阈值时提示。

## 10. API 汇总

建议新增或改造以下接口:

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| GET | `/health` | 健康检查,含数据库连通性 |
| GET | `/api/v1/schedule/rules/board` | 新规则 Board 快照 |
| POST | `/api/v1/schedule/rules/precheck` | 规则预检 |
| POST | `/api/v1/schedule/generate-preview` | 生成预览 |
| POST | `/api/v1/schedule/generate` | 生成草稿 |
| GET | `/api/v1/schedule/conflicts` | 冲突列表 |
| POST | `/api/v1/schedule/conflicts/:id/resolve` | 处理冲突 |
| POST | `/api/v1/schedule/conflicts/:id/ignore` | 忽略冲突 |
| POST | `/api/v1/schedule/confirm-plan` | 第一次确认 |
| POST | `/api/v1/schedule/confirm-day` | 第二/三次确认 |
| POST | `/api/v1/schedule/shifts/:id/cancel` | 取消/请假 |
| POST | `/api/v1/schedule/shifts/:id/absent` | 缺席 |
| POST | `/api/v1/schedule/shifts/:id/move` | 移床/换班 |
| POST | `/api/v1/schedule/temporary` | 临时透析 |
| POST | `/api/v1/schedule/crrt` | 新增 CRRT |
| GET | `/api/v1/schedule/crrt` | CRRT 列表 |
| POST | `/api/v1/schedule/machines/:bedId/outage` | 停机登记 |
| POST | `/api/v1/schedule/calendar/holiday` | 假日配置 |
| POST | `/api/v1/schedule/patients/:id/plan-change` | 方案变更 |
| POST | `/api/v1/schedule/patients/:id/makeup` | 补透 |
| GET | `/api/v1/schedule/diffs` | 差异检测 |
| GET | `/api/v1/schedule/capacity` | 余位预警 |

API 统一要求:

- 所有写接口必须走当前系统正式鉴权和权限中间件。
- 排班写接口按权限矩阵限制角色,不要使用演示版 `X-Role` 作为生产鉴权。
- 401/403/409/500 使用统一错误格式。
- 500 错误对外脱敏,服务端日志记录真实错误。
- 所有写接口写入审计日志。

配置类接口:

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| GET/PUT | `/api/v1/schedule/config/wards` | 病区扩展配置 |
| GET/PUT | `/api/v1/schedule/config/machines` | 机器能力配置 |
| GET/PUT | `/api/v1/schedule/config/patient-profiles` | 患者骨架配置 |
| GET/PUT | `/api/v1/schedule/config/settings` | 排班参数 |
| GET/PUT | `/api/v1/schedule/config/calendar` | 日历配置 |

## 11. 前端实施计划

### 11.1 服务模块

新增:

- `src/services/scheduleRuleApi.ts`
- `src/services/scheduleConfigApi.ts`
- `src/services/scheduleConflictApi.ts`

不要批量重写 `restClient.ts`,只新增模块并从 `src/services/index.ts` 导出。

### 11.2 页面改造

现有页面保留:

- `src/pages/Schedule.tsx`
- `src/pages/ScheduleTemplateList.tsx`
- `src/components/schedule/ApplyTemplateModal.tsx`

新增或扩展:

- 排班规则预检面板。
- 冲突队列抽屉/页面。
- 2/4 周生成弹窗。
- 三级确认按钮。
- 机器能力配置页面。
- 患者排班骨架配置入口。
- 日历配置页面。
- CRRT 占用展示。

### 11.3 UI 行为

- 未通过预检时禁止直接生成。
- 生成预览先展示草稿数量、冲突数量、HDF 分布。
- 用户确认后才写草稿。
- 冲突项可跳转到对应日期/班次/患者。
- 旧模板列表增加迁移提示,不再保存 Status=60。
- CRRT 记录应在周视图中以跨班/跨天占用条展示,避免误认为只占开始日。
- 日期选择器选择任意日期时,排班周视图应自动吸附到该日期所在周的周一。
- 顶部统计卡建议展示:患者数、已排班次数、待确认数、冲突数、差异数、本周范围。
- 今天列高亮,历史列置灰;历史排班不可编辑的原因要可见。
- 左侧病区/机器列应固定,横向滚动时不丢失行上下文。
- HD/HDF/CRRT、草稿/已确认/缺席/取消要用稳定颜色或标签区分。
- 差异面板应支持“补排建议/人工触发补排”,但不能自动补透。
- 工具栏按功能分组,建议分为“查看/数据”“确认/扰动”“配置/管理”。

## 12. 原型代码参考映射

| 原型文件 | 可借鉴内容 | 当前项目落地 |
| --- | --- | --- |
| `internal/model/models.go` | 新模型字段定义 | 拆成老表扩展模型,不要直接替换主表 |
| `internal/sched/constants.go` | 常量、频率、机型能力 | `schedule_rule_constants.go` |
| `internal/sched/board.go` | Board 快照、占用、日历、停机 | `schedule_board_service.go` |
| `internal/sched/engine.go` | 两轮分配、HDF、顺延 | `schedule_engine_service.go` |
| `internal/repo/repo.go` | LoadBoard、SaveDrafts、SaveConflicts | 改为老表 + 扩展表事务写入 |
| `internal/service/schedule_service.go` | GenerateSchedule 编排 | `schedule_template_service.go` 或 `schedule_generation_service.go` |
| `internal/service/ops_service.go` | 确认、取消、缺席、移动 | `schedule_ops_service.go` |
| `internal/service/perturb_service.go` | 临时透析、停机、假日、方案变更 | `schedule_perturb_service.go` |
| `internal/service/crrt_service.go` | CRRT 区间占用 | `schedule_crrt_service.go` |
| `internal/service/diff_service.go` | 应排/已排差异 | `schedule_diff_service.go` |
| `internal/api/api.go` | API 形态 | 融入当前 Gin v1 handler 和权限中间件 |
| `透析排班-backend-v1.1/internal/api/api.go` | 健康检查、错误脱敏、审计日志、写接口鉴权默认安全 | 融入当前正式鉴权/日志体系,不要照搬演示 `X-Role` |
| `透析排班-backend-v1.1/internal/db/db.go` | 数据库连接池和 Ping | 复用当前数据库连接配置方式,禁止 AutoMigrate |
| `透析排班-backend-v1.1/internal/config/config.go` | `SpillHorizonDays`、锚点、生成周数、余位阈值配置读取 | `Schedule_TenantSetting` + 服务层类型校验 |
| `透析排班-backend-v1.1/web/index.html` | 周一吸附、统计卡、角色反馈、历史置灰、冲突/差异入口 | 拆入当前 React 页面和组件,不要使用 CDN 单页演示代码 |
| `透析排班系统-修改答复与测试建议(1).md` | 测试建议与红线说明 | 作为验收 checklist 和禁止自动解决冲突的依据 |

## 13. 待业务确认项

这些项不能由开发自行猜:

- 老库哪个字段可靠表示门诊/住院,用于初始化 A/B。
- 哪些患者属于 C 区,以及“无明确院感指标”的数据来源。
- 患者传染病标识来源是哪张表/哪个字段,用于实现“门诊+传染病拒收”。
- `Plan_PatientPlan.OddWeekFrequency/EvenWeekFrequency` 到五种频率的精确映射。
- 是否存在奇周/偶周频率不一致的病人;如存在,是否扩展 `PatientProfile` 存奇偶周透析日集合。
- HDF 日由医嘱指定还是护士长配置。
- 缺席是否计入“已排次数”。
- 老库 `Status=30/40/50/60` 与新状态的最终映射。
- Bed 是否严格等价于一台透析机器。
- 是否存在一床多机、一机多床、备用机等特殊场景。
- CRRT 机器是否存在于 `Schedule_Bed`,如不存在需先建 Bed 主记录。
- 二次确认 11:00 截止是否需要定时任务自动提醒。
- 子区 A1/A2 是否视为独立排班区;排满时是否允许在同一顶层 A 区内跨子区顺延。
- Profile 变更的历史留存方式:是否新增 ProfileChangeLog,还是完全复用 `Schedule_PlanChange`。
- 模板修改后,已按旧模板生成的草稿是否跟随新模板自动重算。

所有不确定项记录到 `docs/legacy-migration-uncertain-field-checklist.md`。

## 14. 风险和防护

### 14.1 高风险点

- `Status=60` 当前被模板复用,同时老库注释为转出人员,必须迁移。
- 新旧状态机冲突,不能直接改老状态常量。
- HDF/CRRT 能力配置错误会导致错误排班。
- 方案变更、停机、假日处理会批量影响未来排班,必须先预览。
- 生成草稿可能与人工同时修改冲突,写入前必须二次校验。
- 日历开放范围如果用文本存 ID 会导致脏数据和无法索引,必须使用关联表。
- CRRT 跨天占用如果不改周视图查询,会漏显示后续日期占用。
- `RuleStatus` 如果允许 NULL 或默认 0,会与“待排”语义冲突。

### 14.2 防护措施

- 所有批量操作先 preview 后 apply。
- 所有生成写入必须事务化。
- 所有生成写入必须带乐观锁或排他锁,防止并发重复生成。
- 所有自动失败写冲突队列。
- 所有接口按租户过滤。
- 关键写操作加权限守卫。
- 生成时只处理指定 ward/scope,避免全租户误操作。
- 引入审计字段 `CreatorId/CreateTime/LastModifyTime`。
- 冲突队列需要归档/清理策略,例如已处理 30 天后归档。

## 15. 测试计划

### 15.1 后端单元测试

- 频率 `IsDue`。
- 奇偶周 `WeekParity`。
- HDF 替换不增加次数。
- HD 优先 HD 机,HDF 优先 HDF 机。
- HD 溢出 HDF 机。
- HDF 无机写冲突。
- 排满顺延。
- 停机冻结。
- CRRT 区间重叠。
- 二次确认非工作日跳转。
- 日历开放资源关联表查询。
- RuleStatus 回退老状态映射。
- SpillHorizonDays 租户配置默认值和非法值回退。
- AnchorMonday 非周一配置回退默认值。

### 15.2 服务测试

- 模板生成 2 周。
- 模板生成 4 周。
- 三级确认。
- 取消/缺席后机位释放。
- 临时透析借用请假机位。
- 方案变更影响未确认排班。
- 假日挪班建议。
- 差异检测。
- Status=60 模板迁移审计。
- 并发生成只允许一个成功。
- CRRT 跨天周视图展示。
- 健康检查数据库不可达返回 503。
- 未鉴权写接口返回 401,低权限写接口返回 403。
- 5xx 错误响应不包含 SQL、表名、堆栈。
- 审计日志包含方法、路径、状态码、耗时、租户、用户/角色、IP。

### 15.3 前端验收测试

- 选择任意日期后自动吸附到周一并刷新周视图。
- 顶部统计卡能正确显示患者数、已排班次、待确认、冲突、差异。
- 今天列高亮,历史列置灰且不可编辑。
- 横向滚动时机器/病区列固定。
- 操作成功/失败有明确绿色/红色反馈。
- 角色切换后权限错误可见,401/403 不被吞掉。
- 冲突按钮在有待处理冲突时高亮。
- 差异面板少排项只能人工触发补排,不能自动批量补透。

### 15.4 集成验证命令

后端:

```powershell
go test ./internal/services ./internal/api/v1 -count=1 -short
go build -o "$env:TEMP\ai-hms-backend-check.exe" ./cmd/server
```

前端:

```powershell
npm run lint
npm run build
```

## 16. 推荐开发顺序

最推荐顺序:

1. 建扩展表 SQL 和 GORM 模型。
2. 先补工程化基线:健康检查、连接池配置、统一错误、审计日志、写接口鉴权默认安全。
3. 做配置 CRUD: WardExt、BedMachineExt、PatientProfile、Calendar、TenantSetting。
4. 做 Board 快照和预检接口。
5. 迁移模板到独立表,停止 Status=60 写入。
6. 做生成预览。
7. 做生成草稿写入主表 + 扩展表。
8. 做冲突队列闭环,只允许人工采纳建议。
9. 做三级确认和移动/取消/缺席。
10. 做临时透析、停机、假日、方案变更、补透。
11. 做 CRRT。
12. 做差异检测和余位预警。
13. 最后统一前端体验和权限细化。

## 17. 第一批最小可交付版本

如果需要先快速可用,建议第一批只做:

- 新增扩展表 SQL。
- 健康检查、统一错误、审计日志、写接口鉴权默认安全。
- 机器能力配置。
- 患者排班骨架配置。
- Board 预检。
- 独立模板表。
- 2 周生成草稿。
- 冲突队列列表。
- HDF 两轮分配。

暂缓:

- 停机自动迁移。
- 假日挪班。
- 方案变更自动清理。
- CRRT。
- 复杂前端配置页。

这样能先验证新规则最核心的“时间骨架 + 分机位 + HDF 错峰 + 冲突队列”。
