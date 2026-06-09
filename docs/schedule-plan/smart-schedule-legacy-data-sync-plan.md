# 智能排班 v1.3 与老数据库患者信息对接方案

> 状态：待复核  
> 目的：打通老数据库 → v2 智能排班系统的数据链路，实现基于真实患者数据自动生成排班  
> 前置条件：`docs/sql/v1.3_v2_tables.sql` 已执行（14 张 `Schedule_v2_*` 表已创建）

---

## 1. 背景与现状

### 1.1 当前架构

v1.3 智能排班系统（`/api/v2/*`）与老系统**完全隔离**：

- **老系统**：读/写 `Register_PatientInfomation`、`Plan_PatientPlan`、`Schedule_Ward`、`Schedule_Bed`、`Schedule_Shift` 等老表
- **v2 系统**：仅读/写 `Schedule_v2_*` 表（Ward / Machine / Shift / Patient / PatientProfile / PatientShift 等 14 张表）

**没有任何桥接代码**。当前只能通过 `POST /api/v2/admin/seed` 写入 7 个硬编码假病人演示排班。

### 1.2 目标

开发 5 个数据同步组件，将老数据库中的真实患者、病区、机位、班次数据导入 v2 系统，然后调用 v2 的生成引擎自动创建排班。

### 1.3 设计原则

1. **只读老表**：同步程序只从老表读取，**绝不写入/修改老表**
2. **可重复执行**：同步操作必须幂等（按 `TenantId + Id` 去重 UPSERT），可随时重新同步
3. **增量友好**：支持全量同步（首次）和增量同步（新增/变更的患者）
4. **人工确认边界**：无法自动推断的字段（HdfWeekday、FixedMachine、ZoneTag 等）设默认值 + 标记为待确认，后续通过管理界面人工补充
5. **一个 API + 一个 CLI**：提供 `POST /api/v2/sync/legacy` 一键同步端点，同时支持命令行 `go run ./cmd/sync` 独立调用

---

## 2. 老数据库相关表结构

### 2.1 患者主档 — `Register_PatientInfomation`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | 患者唯一标识 |
| `TenantId` | bigint | 租户 |
| `Name` | varchar(256) | 姓名 |
| `Gender` | varchar(64) | 性别 |
| `TreatmentStatus` | varchar(64) | 在院/出院/死亡 |
| `PatientType` | varchar(64) | 长期10/临时20 |
| `ExpenseType` | varchar(64) | 医保类型 |
| `DialysisNo` | varchar(64) | 透析号 |
| `FirstDialysisDate` | timestamp | 首次透析日期 |

### 2.2 患者方案 — `Plan_PatientPlan`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | 方案 ID |
| `PatientId` | bigint | 关联患者 |
| `OddWeekFrequency` | integer | 单周治疗频次 |
| `EvenWeekFrequency` | integer | 双周治疗频次 |
| `DialysisMethod` | varchar(256) | 透析方法（HD/HDF/HP/HD+HP 等） |
| `DialysisDuration` | numeric | 透析时长（小时） |
| `DryWeight` | numeric | 干体重 |
| `BF` | numeric | 标准血流量 |
| `IsDisabled` | boolean | 是否禁用 |
| `Frequency` | varchar(128) | 频次描述文本（自由输入） |

### 2.3 病区 — `Schedule_Ward`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | 病区 ID |
| `TenantId` | bigint | 租户 |
| `Name` | varchar(256) | 病区名称 |
| `PatientType` | varchar(64) | 适用患者类型（长期10/临时20） |
| `InfectionType` | varchar(64) | 适用传染病（普通/乙肝/丙肝） |
| `IsDisabled` | boolean | 是否禁用 |

### 2.4 床位 — `Schedule_Bed`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | 床位 ID |
| `TenantId` | bigint | 租户 |
| `Name` | varchar(256) | 床位名称 |
| `WardId` | bigint | 所属病区 |
| `IsDisabled` | boolean | 是否禁用 |
| `FEPId` | bigint | 前置机 ID |

### 2.5 设备 — `Auxiliary_EquipmentInfomation`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | 设备 ID |
| `Name` | varchar(256) | 设备名称 |
| `DialysisMethod` | varchar(512) | 治疗方式（HD/HDF，枚举多选） |
| `Type` | varchar | 类型（血透机/血滤机/水处理机等） |
| `Flux` | varchar(64) | 通量（高/低） |
| `Brand` / `ModelNo` / `SerialNo` | varchar | 品牌/型号/序列号 |

### 2.6 床位-设备关联 — `Schedule_BedEquipmentRel`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | |
| `BedId` | bigint | 床位 ID |
| `EquipmentId` | bigint | 设备 ID |
| `IsDefault` | boolean | 是否默认设备 |
| `IsDisabled` | boolean | 是否禁用 |

### 2.7 班次 — `Schedule_Shift`

| 字段 | 类型 | 用途 |
|------|------|------|
| `Id` | bigint PK | 班次 ID |
| `Name` | varchar(256) | 班次名称 |
| `StartTime` | timestamp | 开始时间 |
| `EndTime` | timestamp | 结束时间 |
| `Type` | integer | 类别（长期10/临时20） |
| `Sort` | integer | 排序 |

---

## 3. 目标 v2 表结构（资源层）

### 3.1 `Schedule_v2_Ward`

| 字段 | 说明 |
|------|------|
| `Id` | 沿用老 Ward.Id |
| `TenantId` | |
| `Name` | 病区名称 |
| `ZoneType` | **A/B/C — 老系统无此字段，需推断** |
| `Sort` | 排序 |
| `IsDisabled` | |

### 3.2 `Schedule_v2_Machine`

| 字段 | 说明 |
|------|------|
| `Id` | **自动生成（老系统无统一 Machine ID）** |
| `WardId` | 所属病区 |
| `Code` | 机位编号（来自设备编号或床位名称） |
| `Name` | 机位名称 |
| `MachineType` | **HD/HDF/CRRT — 从设备能力推断** |
| `PositionIndex` | 区内物理排位 |
| `LegacyBedId` | 回指老 Schedule_Bed.Id |

### 3.3 `Schedule_v2_Shift`

| 字段 | 说明 |
|------|------|
| `Id` | 沿用老 Shift.Id |
| `Name` | 班次名称 |
| `ShiftCode` | **MORNING/AFTERNOON/NIGHT — 从老 Type/Sort 推断** |
| `StartTime` / `EndTime` | |
| `Sort` | |

### 3.4 `Schedule_v2_Patient`

| 字段 | 说明 |
|------|------|
| `Id` | 沿用老 Register_PatientInfomation.Id |
| `TenantId` | |
| `Name` | |
| `Gender` | |
| `InfectionStatus` | **需从 Register_Infection 推断（见 4.2.4）** |

### 3.5 `Schedule_v2_PatientProfile`

| 字段 | 说明 | 映射难度 |
|------|------|----------|
| `PatientId` | 沿用老 Patient.Id | 直接 |
| `ZoneTag` | A/B/C 区域标签 | **需推断** |
| `HomeWardId` | 归属病区 | **需从历史排班推断** |
| `WeeklyCount` | 每周次数 | **需从 OddWeek/EvenWeekFrequency 推断** |
| `FreqPattern` | 星期组合（一三五/二四六/etc.） | **需从历史排班推断** |
| `ShiftId` | 全周同一班次 | **需从历史排班推断** |
| `DefaultMode` | HD/HDF/HFD/HF | 从 DialysisMethod 映射 |
| `HdfEnabled` | 是否启用 HDF 替换 | **需人工配置** |
| `HdfWeekday` | HDF 日（1=周一..6=周六） | **需人工配置** |
| `HdfWeekParity` | 0=偶周 1=奇周 | 由引擎自动分配 |
| `FixedHdMachineId` | 固定 HD 机位 | 从历史排班推算 |
| `FixedHdfMachineId` | 固定 HDF 机位 | 从历史排班推算 |

---

## 4. 字段映射规则（逐项详细说明）

### 4.1 组件一：SyncWards — 病区同步

```
老表：Schedule_Ward
目标：Schedule_v2_Ward
```

#### 映射逻辑

| 老字段 | v2 字段 | 映射规则 |
|--------|---------|----------|
| `Id` | `Id` | 直接复制 |
| `TenantId` | `TenantId` | 直接复制 |
| `Name` | `Name` | 直接复制 |
| `Sort` | `Sort` | 直接复制 |
| `IsDisabled` | `IsDisabled` | 直接复制 |

#### ZoneType 推断规则

老系统无 `ZoneType` 概念，需根据 `PatientType` + `InfectionType` 推断：

| `Schedule_Ward` 条件 | 推断 ZoneType | 理由 |
|-------|-------------|------|
| `PatientType='10'`（长期）+ `InfectionType` 含'普通' | **A** | 长期普通门诊区 |
| `PatientType='10'`（长期）+ `InfectionType` 含'乙肝'/'丙肝' | **C** | 传染病隔离区 |
| `PatientType='20'`（临时）| **B** | 临时/住院区 |
| 无法确定 | **A**（默认） | 安全回退 |

**⚠️ 需人工确认**：此规则为推测。实际映射请根据医院病区规划调整。ZoneType 错误会导致排班引擎对所有患者错误分配机位。

#### 同步策略

- 按 `TenantId + Id` UPSERT
- `IsDisabled=true` 的 ward 也会导入（标记为禁用）

---

### 4.2 组件二：SyncShifts — 班次同步

```
老表：Schedule_Shift
目标：Schedule_v2_Shift
```

#### 映射逻辑

| 老字段 | v2 字段 | 映射规则 |
|--------|---------|----------|
| `Id` | `Id` | 直接复制 |
| `TenantId` | `TenantId` | 直接复制 |
| `Name` | `Name` | 直接复制 |
| `Sort` | `Sort` | 直接复制 |
| `StartTime` | `StartTime` | 取 HH:MM 部分 |
| `EndTime` | `EndTime` | 取 HH:MM 部分 |
| `IsDisabled` | `IsDisabled` | 直接复制 |

#### ShiftCode 推断规则

| 条件 | ShiftCode |
|------|-----------|
| `Sort=1` 或 `Type=10`（长期第一个班） | **MORNING** |
| `Sort=2` | **AFTERNOON** |
| `Sort=3` | **NIGHT** |
| 无法确定 | **MORNING**（默认） |

**⚠️ 需人工确认**：不同医院的班次排序约定可能不同。建议同步后通过管理界面确认。

---

### 4.3 组件三：SyncMachines — 机位同步（最复杂）

```
老表：Schedule_Bed + Auxiliary_EquipmentInfomation + Schedule_BedEquipmentRel
目标：Schedule_v2_Machine
```

#### 数据组装逻辑

老系统是"床位 ↔ 设备"分离模型，v2 是统一 Machine 模型。组装步骤：

1. 从 `Schedule_Bed` 获取启用的床位列表（`IsDisabled=false`）
2. 对每个床位，通过 `Schedule_BedEquipmentRel`（`IsDisabled=false`, `IsDefault=true`）找到默认设备
3. 从 `Auxiliary_EquipmentInfomation` 获取设备的透析方式、通量、类型
4. 组装 v2 Machine

#### 映射规则

| v2 Machine 字段 | 来源 | 规则 |
|-----------------|------|------|
| `Id` | **新生成** | 按顺序自增（与 Bed.Id 区分命名空间） |
| `LegacyBedId` | `Schedule_Bed.Id` | 保留回指，方便追溯 |
| `WardId` | `Schedule_Bed.WardId` | 直接映射 |
| `Code` | `Auxiliary_EquipmentInfomation.IDNo` 或 `SerialNo` | 优先设备编号，其次床位编号 |
| `Name` | `Schedule_Bed.Name` 拼接设备型号 | 例如 "A区01床 费森尤斯4008S" |
| `MachineType` | **见推断规则** | |
| `PositionIndex` | `Schedule_Bed.Sort` 或 `Id` | 按 WardId 分组后递增 |
| `IsDisabled` | `Schedule_Bed.IsDisabled` | |

#### MachineType 推断规则

从 `Auxiliary_EquipmentInfomation.DialysisMethod` + `Type` + `Flux` 推断：

| 设备字段值 | 推断 MachineType |
|-----------|-----------------|
| `DialysisMethod` 含 "HDF" | **HDF** |
| `DialysisMethod` 含 "HD"（不含 HDF）且 `Flux` 为"高" | **HDF**（高通量 HD 机可做 HDF） |
| `DialysisMethod` 含 "HD"（不含 HDF）且 `Flux` 为"低" | **HD** |
| `Type` 含 "CRRT" 或 "血滤机" | **CRRT** |
| `DialysisMethod` 含 "HP" | **HDF**（HP 需要 HDF 机） |
| 无法确定（无设备关联或设备信息不全） | **HD**（默认） |

**⚠️ 需人工确认**：设备型号决定实际能力（如费森尤斯 4008S 标称 HD 但高通量实际可做 HDF；5008S 可做 HDF）。同步后需通过管理界面逐台确认 MachineType。

#### 无设备关联的床位处理

如果 `Schedule_Bed` 没有关联设备（`Schedule_BedEquipmentRel` 无匹配行），默认设为 HD 型机位，`Code` = "BED-{BedId}"。

---

### 4.4 组件四：SyncPatients — 患者与骨架同步（核心）

```
老表：Register_PatientInfomation + Plan_PatientPlan + Register_Infection
目标：Schedule_v2_Patient + Schedule_v2_PatientProfile
```

#### 4.4.1 Patient 映射

| 老字段（Register_PatientInfomation） | v2 字段（Schedule_v2_Patient） |
|--------------------------------------|-------------------------------|
| `Id` | `Id` |
| `TenantId` | `TenantId` |
| `Name` | `Name` |
| `Gender` | `Gender` |

#### 4.4.2 InfectionStatus 推断

从 `Register_Infection` 表查询（按 PatientId + TenantId）：

| `Register_Infection` 条件 | 推断值 |
|--------------------------|--------|
| `InfectionDesc` 含 "乙肝" 或 "HBsAg" | `positive` |
| `InfectionDesc` 含 "丙肝" 或 "HCV" | `positive` |
| `InfectionDesc` 含 "HIV" 或 "艾滋" | `positive` |
| `InfectionDesc` 含 "梅毒" | `positive` |
| 无匹配记录 或 `InfectionDesc` 为空 | `unknown` |
| 所有其他情况 | `unknown`（保守默认） |

#### 4.4.3 患者筛选条件

同步范围：`Register_PatientInfomation` 中满足以下全部条件的患者：

- `TreatmentStatus = '在院'` （在院治疗中）
- `PatientType = '10'` （长期患者，非临时）
- `TenantId = 指定租户`

排班引擎（决策27）会跳过 `PatientStatus=20`（已出组）的患者。出组由 v2 系统的 `DischargePatient()` 手动触发，不同步。

#### 4.4.4 PatientProfile 映射 — 可自动推断的字段

| v2 Profile 字段 | 来源 | 规则 |
|-----------------|------|------|
| `PatientId` | `Register_PatientInfomation.Id` | 直接 |
| `TenantId` | | 直接 |
| `DefaultMode` | `Plan_PatientPlan.DialysisMethod` | **见下表** |
| `WeeklyCount` | `Plan_PatientPlan` | **见 §4.4.5** |
| `PatientStatus` | — | 默认 `10`（在透） |
| `CreatorId` | — | `0`（系统同步） |

#### DialysisMethod → DefaultMode 映射

| 老 DialysisMethod 值 | v2 DefaultMode | 说明 |
|----------------------|----------------|------|
| `"HD"` | `HD` | 常规血液透析 |
| `"HDF"` | `HDF` | HDF 替换（需同时设 HdfEnabled=true） |
| `"HP"` 或 `"血液灌流"` | `HF` | 血液灌流 |
| `"HD+HP"` | `HD` | 复合治疗→取主模式 HD，HP 通过临时透析处理 |
| `"HD+HDF"` | `HD` | 复合治疗→取主模式 HD，HDF 部件通过 HdfEnabled 控制 |
| `"HFD"` | `HFD` | 高通量透析（决策24/25） |
| 空值或无法映射 | `HD` | 默认安全值 |

**⚠️ 重要**：如果 DialysisMethod 含 "HDF"，同步时将 `HdfEnabled=true` + `HdfWeekday=3`（默认周三）。之后需人工确认。

#### 4.4.5 频率映射 — 核心难点

老系统频率字段：

| 老字段 | 含义 |
|--------|------|
| `OddWeekFrequency` | 单周（奇数周）治疗次数 |
| `EvenWeekFrequency` | 双周（偶数周）治疗次数 |
| `Frequency` | 自由文本（如 "周一三五"、"隔天"） |

v2 需要：`WeeklyCount`（每周次数）+ `FreqPattern`（星期组合枚举）+ `HdfWeekday`（具体 HDF 日）

**映射策略**：

| 情况 | WeeklyCount | FreqPattern | 说明 |
|------|-------------|-------------|------|
| Odd=3 且 Even=3 | 3 | **10**（一三五） | 最常见，默认 |
| Odd=3 且 Even=2 | 3（取大值） | 10（一三五） | ⚠️ 偶数周少排1次，系统不支持双周变频率。按较大值同步 |
| Odd=2 且 Even=3 | 3（取大值） | 10（一三五） | ⚠️ 同上 |
| Odd=2 且 Even=2 | 2 | **30**（二四六） | |
| Odd=1 且 Even=1 | 1 | **50**（仅周四） | |
| Odd=4 且 Even=4 | 4 | **40**（一三五+六?） | ⚠️ v2 FreqPattern=40 代表"一三五+周六"四天，需确认 |
| Frequency 含 "一三五" | 按文本 | 10 | 利用文本字段校验 |
| Frequency 含 "二四六" | 按文本 | 30 | |
| 其他 | 2 | **40**（默认两/周） | 安全回退 |

**⚠️ 需人工确认**：
1. 奇数周/偶数周频率不一致的患者，当前策略取较大值。这意味着某些周会"多排"（生成排班后被取消或标记缺勤）。是否可接受？
2. `FreqPattern` 的星期分布是否正确符合患者实际透析日？
3. 建议同步后在"待确认"列表中列出所有频率不一致的患者

#### 4.4.5 ZoneTag 推断

ZoneTag 决定患者被分配到哪个区域（A/B/C）：

| 条件 | 推断 ZoneTag |
|------|-------------|
| 从 `Schedule_PatientShift` 历史记录中发现患者最常在的 Ward → 查该 Ward 的 ZoneType | 使用历史 ZoneType |
| 从 `Plan_PatientPlan` 中 `DialysisMethod` 含 "CRRT" | **C** |
| 从 `Register_Infection` 中 `InfectionDesc` 含 "乙肝/丙肝/HIV/梅毒" | **C**（决策26） |
| 其他 | **A**（默认） |

#### 4.4.6 HomeWardId 推断

| 策略 | 方法 |
|------|------|
| **优先** | 查 `Schedule_PatientShift` 最近 4 周内该患者排班记录，取频率最高的 `WardId` |
| **次选** | 查 `Plan_PatientPlan` 关联的 `VascularAccessId` → `Register_VascularAccess` → 关联的病区 |
| **默认** | 取该 ZoneType 下第一个非禁用的 Ward |

#### 4.4.7 固定机位推断（可选，建议先不自动推）

| 字段 | 建议 |
|------|------|
| `FixedHdMachineId` | **不同步** — 设为 NULL。通过管理界面人工指定 |
| `FixedHdfMachineId` | **不同步** — 设为 NULL |
| `HdfEnabled` | DialysisMethod 含 "HDF" → true，否则 false |
| `HdfWeekday` | **默认 3（周三）**，需人工确认 |
| `HdfWeekParity` | NULL（由引擎 AssignHdfWeekParity 自动分配） |
| `ShiftId` | 查历史排班中最常用的 ShiftId；无历史则 NULL（由引擎匹配同区同班次） |

---

### 4.5 组件五：RebuildTemplates — 从骨架重建模板

```
输入：Schedule_v2_PatientProfile（已有）
输出：Schedule_v2_ScheduleTemplate + Schedule_v2_ScheduleTemplateItem
```

#### 逻辑

v2 系统已有 `RebuildTemplateFromProfiles()` 方法（`admin_service.go`），在同步完患者骨架后直接调用：

1. 获取所有活跃的 `Schedule_v2_PatientProfile`（`PatientStatus=10`）
2. 生成模板项（每患者 1 条）：ZoneTag、WardId、ShiftId、FreqPattern、DefaultMode、Hdf* 字段全量复制
3. 覆盖保存到 `Schedule_v2_ScheduleTemplate`（scope=ALL, IsActive=true）

**幂等**：再次执行会重建整个模板，覆盖之前的内容。

---

## 5. 待开发清单

### 5.1 新增文件

| 文件 | 包 | 说明 |
|------|-----|------|
| `internal/smart_schedule/service/sync_service.go` | service | 核心同步逻辑（5 个组件） |
| `internal/smart_schedule/api/api_sync.go` | api | `POST /api/v2/sync/legacy` 端点 |
| `internal/smart_schedule/service/sync_service_test.go` | service | 同步逻辑单元测试 |
| `cmd/sync/main.go` | main | CLI 独立同步入口 |

### 5.2 核心函数（sync_service.go）

```go
// SyncWards 同步病区数据
func SyncWards(g *gorm.DB, tenant int64) (created, updated int, err error)

// SyncShifts 同步班次数据
func SyncShifts(g *gorm.DB, tenant int64) (created, updated int, err error)

// SyncMachines 同步机位数据（床位+设备→Machine）
func SyncMachines(g *gorm.DB, tenant int64) (created, updated int, err error)

// SyncPatients 同步患者+骨架数据
func SyncPatients(g *gorm.DB, tenant int64) (created, updated int, warnCount int, err error)

// SyncAll 一键全同步
func SyncAll(g *gorm.DB, tenant int64) (*SyncResult, error)

// SyncResult 同步结果汇总
type SyncResult struct {
    TenantId   int64           `json:"tenantId"`
    Wards      SyncCountResult `json:"wards"`
    Shifts     SyncCountResult `json:"shifts"`
    Machines   SyncCountResult `json:"machines"`
    Patients   SyncCountResult `json:"patients"`
    Warnings   []string        `json:"warnings"`   // 需人工确认的问题列表
    Templates  SyncCountResult `json:"templates"`
}

type SyncCountResult struct {
    Created int `json:"created"`
    Updated int `json:"updated"`
    Skipped int `json:"skipped"`
    Errors  int `json:"errors"`
}
```

### 5.3 API 端点

```
POST /api/v2/sync/legacy
  Headers: Authorization Bearer <JWT>
  Body: { "tenantId": 1, "components": ["wards","shifts","machines","patients","templates"] }
  Response: SyncResult

GET /api/v2/sync/legacy/status
  返回上次同步结果缓存

GET /api/v2/sync/legacy/warnings
  返回需人工确认的项（频率不一致的患者、未确定 ZoneType 的 ward 等）
```

### 5.4 CLI 入口（cmd/sync/main.go）

```go
// 用法：
//   go run ./cmd/sync --tenant=1 --dsn="host=... dbname=..."
//
// 流程：
//   1. 连接老数据库（读取 Register_*, Schedule_*, Plan_*）
//   2. 连接新数据库（写入 Schedule_v2_*）
//   3. 依次执行 wards → shifts → machines → patients → templates
//   4. 打印结果表格
```

### 5.5 main.go 路由注册

在 `cmd/server/main.go` 中注册同步路由组（需要 head_nurse 角色）：

```go
syncGroup := smartSchedule.Group("/sync")
syncGroup.Use(guard(RoleHeadNurse))
syncGroup.POST("/legacy", smartServer.SyncLegacy)
syncGroup.GET("/legacy/status", smartServer.SyncStatus)
syncGroup.GET("/legacy/warnings", smartServer.SyncWarnings)
```

---

## 6. 待决策问题（需人工确认）

以下问题影响数据映射准确性，建议在开发前确认：

| # | 问题 | 影响范围 | 建议默认值 |
|---|------|---------|-----------|
| 1 | **频率映射**：奇数周/偶数周频次不同时，取大值还是平均值？ | `PatientProfile.WeeklyCount` + `FreqPattern` | **取大值**（宁可多排，人工取消） |
| 2 | **FreqPattern 星期分布**：每周3次是否总是一三五？有无二四六患者？ | `PatientProfile.FreqPattern` | **检查 Frequency 文本字段**，无信息默认一三五 |
| 3 | **ZoneType 推断**：当前规则（长期+普通→A，传染病→C，临时→B）是否正确？ | `Ward.ZoneType` | **按规则，同步后人工校验** |
| 4 | **HDF 日默认周三**：老系统无 HdfWeekday，默认周三是否合理？ | `PatientProfile.HdfWeekday` | **默认3（周三），可改** |
| 5 | **复合治疗模式**："HD+HP"这类复合模式如何映射？ | `PatientProfile.DefaultMode` | **取主模式 HD**，HP 作为临时透析另行处理 |
| 6 | **固定机位**：是否从历史排班记录中自动推算 `FixedMachineId`？ | `PatientProfile.FixedHdMachineId` | **建议不同步**（避免错误固定），人工指定 |
| 7 | **历史排班记录处理**：老 `Schedule_PatientShift` 数据是否需要迁移到 v2？ | 不直接迁移。v2 从头开始生成 | **不同步历史**，从下一个自然周开始生成 |
| 8 | **租户范围**：同步所有租户还是指定租户？ | `TenantId` | **按请求中指定**，默认 `1` |

---

## 7. 同步执行顺序与依赖

```
1. SyncWards()       ← 无依赖
2. SyncShifts()      ← 无依赖
3. SyncMachines()    ← 依赖 Ward（WardId）
4. SyncPatients()    ← 依赖 Ward + Machine（ZoneTag, HomeWardId）
5. RebuildTemplates() ← 依赖 Patient + PatientProfile
6. GenerateSchedule() ← 用户手动触发 POST /api/v2/schedule/generate
```

**第 6 步（排班生成）不在同步范围内**。同步只负责数据准备，排班由用户通过 `POST /api/v2/schedule/generate` 手动触发（因为需要指定起始周和生成周数）。

---

## 8. 测试策略

### 8.1 单元测试（sync_service_test.go）

- `TestSyncWards_AllFields`: 模拟老 ward 数据，验证 ZoneType 推断
- `TestSyncMachines_Assemble`: 验证 Bed+Equipment → Machine 组装
- `TestSyncPatients_DialysisMethodMapping`: 所有 DialysisMethod 变体的 DefaultMode 映射
- `TestSyncPatients_FrequencyMapping`: 奇数周/偶数周不同组合的映射
- `TestSyncPatients_InfectionDetection`: Register_Infection → InfectionStatus
- `TestSyncIdempotent`: 多次同步不产生重复数据

### 8.2 集成测试

使用 `TEST_DATABASE_URL` 连接测试库，种子老表数据 → 执行同步 → 验证 v2 表数据完整性 → 调用 GenerateSchedule 验证可生成排班。

---

## 9. 风险与边界

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| ZoneType 推断错误 → 患者分配到错误区域 | 高 | 同步后输出 warnings；管理界面提供批量修改 ZoneType |
| 频率映射错误 → 患者多排或少排 | 中 | 差异检测 (`ComputeDiffs`) 会暴露；通过取消/补透修正 |
| MachineType 推断错误 → HDF 患者分配到 HD 机 | 中 | 引擎有 `MachineSupports()` 校验；不支持的组合进冲突队列 |
| 同步中断 → 部分数据残留 | 低 | UPSERT 幂等保证可重新执行；同步不在事务中（每条记录独立） |
| 大量患者（>500）→ 同步耗时 | 低 | 单租户内分页处理（每页 200 条） |

---

## 10. 开发预估

| 模块 | 文件 | 代码量（估） | 复杂度 |
|------|------|-------------|--------|
| sync_service.go | 核心逻辑 | ~400 行 | 中 |
| api_sync.go | API 端点 | ~100 行 | 低 |
| sync_service_test.go | 单元测试 | ~200 行 | 中 |
| cmd/sync/main.go | CLI 工具 | ~80 行 | 低 |
| main.go 路由注册 | 路由 | ~10 行 | 低 |
| **总计** | | **~790 行** | |

---

## 附录 A：老表 → v2 表完整对照表

| 老表 | v2 表 | 同步函数 | 关键推断 |
|------|-------|---------|---------|
| `Schedule_Ward` | `Schedule_v2_Ward` | `SyncWards()` | ZoneType |
| `Schedule_Shift` | `Schedule_v2_Shift` | `SyncShifts()` | ShiftCode |
| `Schedule_Bed` + `Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` | `Schedule_v2_Machine` | `SyncMachines()` | MachineType, Code |
| `Register_PatientInfomation` | `Schedule_v2_Patient` | `SyncPatients()` | InfectionStatus |
| `Register_PatientInfomation` + `Plan_PatientPlan` + `Register_Infection` | `Schedule_v2_PatientProfile` | `SyncPatients()` | ZoneTag, WeeklyCount, FreqPattern, DefaultMode, Hdf*, HomeWardId |
| `Schedule_v2_PatientProfile` | `Schedule_v2_ScheduleTemplate` + `Schedule_v2_ScheduleTemplateItem` | `RebuildTemplateFromProfiles()` | 直接复制 |

## 附录 B：老表查询涉及的已有代码

主系统已有以下 Go 结构体和查询，可直接复用：

- `internal/services/patient_service.go`: `legacyPatientPlan`（`Plan_PatientPlan` 全字段映射）
- `internal/services/patient_core_service.go`: `buildInfection()`（`Register_Infection` 查询）
- `internal/services/device_service.go`: `getLegacyDeviceByID()`（`Auxiliary_EquipmentInfomation` + `Schedule_BedEquipmentRel` + `Schedule_Bed` + `Schedule_Ward` 多表关联查询）
- `internal/models/patient.go`: `Patient`（`Register_PatientInfomation` 的 GORM 模型）
- `internal/models/schedule.go`: `Ward`、`Bed`、`Shift`、`PatientShift`（老 `Schedule_*` 表的 GORM 模型）

**注意**：同步服务应直接使用 GORM 查询老表，而非通过主系统 services 接口调用，以保持 smart_schedule 包的独立性。
