# AI-HMS 去 Mock 化整改开发计划

> **初版日期**：2026-04-06
> **二次核实日期**：2026-04-06（对照代码逐任务验证，全部前提假设确认准确）
> **审查人**：医院信息系统高级架构师 + 资深全栈工程师 + 代码审计专家
> **状态**：✅ 已核实，待 Codex 执行
> **执行完成后**：提交复审，对照第8节检查清单逐项核对；变更记录见 `DEMOCK_CHANGELOG.md`

---

# 1. 项目概览

**项目**：AI-HMS 血液透析管理系统
**仓库路径**：`F:\python\前后端代码\ai-hms_qhd`
**技术栈**：Go 1.24 · Gin · GORM · PostgreSQL 15 · React 19 · TypeScript · Vite · Ant Design 6

**总体结论**：主框架和核心 CRUD 接口对接已完成，但去 Mock 化工作仅完成约 55%。constants.ts 仍持有 15 个 MOCK 常量（均为死代码），4 个主要业务页面完全无 API 调用，护士姓名/日期在表单中大量硬编码，存在个人信息泄露风险。

---

# 2. 数据来源现状盘点

## 2.1 已完成真实 API 对接的模块

| 模块 | 页面文件 | 对接接口 | 状态 |
|------|---------|---------|------|
| 患者列表 | `PatientList.tsx` | `GET /patients` | ✅ 完成 |
| 患者详情 | `PatientDetail.tsx` | `GET /patients/:id` + 多子接口 | ✅ 完成 |
| 透析排班 | `Schedule.tsx` | `GET /users` + 排班 API | ✅ 完成 |
| 库存管理 | `Inventory.tsx` | 库存全套 CRUD | ✅ 完成 |
| 设备管理 | `DeviceManagement.tsx` | 设备 CRUD | ✅ 完成 |
| 看板统计 | `Dashboard.tsx` | `GET /dashboard/stats` | ✅ 完成 |
| 监控大屏设备列表 | `Monitoring.tsx` | `GET /devices` | ✅ 完成 |
| 主数据（药品/耗材）| `MasterData.tsx` | 药品/耗材目录 API | ✅ 完成 |
| 诊疗配置 | `TreatmentConfig/` | 配置 CRUD | ✅ 完成 |
| 字典配置 | `DictConfig.tsx` | 字典 CRUD | ✅ 完成 |
| HDIS 集成 | `Settings.tsx` | HDIS 设置 API | ✅ 完成 |
| 透析处理-患者加载 | `DialysisProcessing.tsx` | `GET /patients` | ✅ 完成 |
| 透析处理-监控记录 | `DialysisProcessing.tsx` | `GET /patients/:id/treatment` + duringParams | ✅ 完成 |
| 病区总览基础数据 | `WardOverview.tsx` | 设备/班次 API | ✅ 部分完成 |

## 2.2 部分完成但仍残留 Mock/硬编码的模块

| 模块 | 页面文件 | 已完成部分 | 残留问题 |
|------|---------|-----------|---------|
| 透析处理-表单 | `DialysisProcessing.tsx` | 患者/监控记录 API | 护士下拉选项硬编码（武琪迪/李俊雅），11 处日期 hardcode |
| 病区总览 | `WardOverview.tsx` | 设备/班次/患者总数 | processData（等待/透析中/消毒/完成统计）仍硬编码 |
| 透析历史记录 | `HistoryTab.tsx` | 无（注释已说明待对接）| 25 条动态生成假历史记录，useMemo 依赖数组为空 |
| 监控图表历史 | `Monitoring.tsx` | 设备列表来自 API | generateHistoryData 生成假监控历史，硬编码患者名"高敬兰" |

## 2.3 完全未完成真实数据对接的模块

| 模块 | 页面文件 | 问题描述 | 所需后端接口 |
|------|---------|---------|------------|
| 质控统计 | `Statistics.tsx` | 零 API 调用，全部硬编码 | 需新增统计汇总接口 |
| 任务/通知栏 | `MainLayout.tsx` | 5 条假任务，无 API | 需新增临床任务/预警接口 |

## 2.4 已失效但仍保留的 Mock 定义/辅助代码

| 文件 | 内容 | 状态 |
|------|------|------|
| `src/constants.ts` | 15 个 MOCK_ 常量（见第3节详表）| 全部死代码，无页面引用 |
| `src/utils/mockHelpers.ts` | `getPatientList()`、`getPatientById()` 等 4 个函数 | 无任何调用方，待整体删除 |

---

# 3. 残留 Mock 数据问题总表

## 3.1 constants.ts 中的死代码 MOCK 常量（全部无引用）

| # | 常量名 | 行号 | 内容摘要 | 处置建议 |
|---|--------|------|---------|---------|
| 1 | `MOCK_TREATMENT_PLAN` | 16-65 | 透析方案（物料/参数/调整历史）| **直接删除** |
| 2 | `MOCK_DEVICE_INVENTORY` | 68-77 | 15 条设备清单 | **直接删除** |
| 3 | `MOCK_DEVICE_MAINTENANCE_LOGS` | 80-84 | 3 条设备维护日志 | **直接删除** |
| 4 | `MOCK_SESSION_PROCESS` | 87-94 | 监控记录（已被替换）| **直接删除** |
| 5 | `MOCK_PATIENTS` | 98-189 | 2 个完整患者（张伟 P001/李娜 P002）| **连同 mockHelpers.ts 一起删除** |
| 6 | `MOCK_STATS_DATA` | 207-213 | 5 条图表统计数据 | **直接删除** |
| 7 | `MOCK_VITALS_HISTORY` | 215-222 | 6 条生命体征历史 | **直接删除** |
| 8 | `MOCK_MONITOR_DEVICES` | 224-252 | 25 台模拟设备（已被替换）| **直接删除** |
| 9 | `MOCK_STAFF` | 254-260 | 5 名护士（已被替换）| **直接删除** |
| 10 | `MOCK_SCHEDULE` | 262-280 | 动态生成 7 天排班（已被替换）| **直接删除** |
| 11 | `MOCK_PATIENT_SCHEDULE` | 282-313 | 25 床×7 天排班 | **直接删除** |
| 12 | `MOCK_TREATMENT_HISTORY` | 315-319 | 3 条治疗历史 | **直接删除** |
| 13 | `MOCK_DETAILED_ORDERS` | 324-331 | 6 条医嘱（含 2024/2025 年日期）| **直接删除** |
| 14 | `MOCK_ORDER_DATES` | 334 | 4 个医嘱日期 | **直接删除** |
| 15 | `MOCK_MED_GROUPS` | 337-383 | 10 个药品分组 | **直接删除** |

> **注意**：删除后 constants.ts 仅保留第 191 行起的 `DASHBOARD_CARDS` 导出，文件从 384 行缩减至约 30 行。

## 3.2 页面内联 Mock 数据

| # | 文件 | 位置 | 问题 | 风险级别 |
|---|------|------|------|---------|
| 1 | `DialysisProcessing.tsx` | 7 处 `<option>` | 护士名"武琪迪""李俊雅"硬编码 | **高**（个人信息泄露）|
| 2 | `DialysisProcessing.tsx` | 11 处 `defaultValue` | 日期全为 `2025-12-13 xx:xx` | **高**（逻辑错误）|
| 3 | `Statistics.tsx` | 38-74 行 | qualityData/infectionData/vascularData/workloadData 全部硬编码，零 API | **高**（数据失真）|
| 4 | `MainLayout.tsx` | 27-53 行 | MOCK_TASKS：5 条假通知（引用 P001/P002/P003/P006）| **中** |
| 5 | `HistoryTab.tsx` | 26-42 行 | 25 条循环生成假历史，注释标注"待 API 替换"，useMemo 依赖为空 `[]` | **高** |
| 6 | `Monitoring.tsx` | generateHistoryData | 图表历史数据模拟生成，硬编码患者名"高敬兰" | **中**（个人信息）|
| 7 | `Monitoring.tsx` | PrescriptionEditModal | 5 种医疗耗材、体重/超滤量/血压大量临床参数 hardcoded | **中** |
| 8 | `WardOverview.tsx` | 47-63 行 | totalCapacity:50、workCompletion:85%、processData 各状态计数全部 hardcoded | **中** |

## 3.3 Mock 辅助函数

| # | 文件 | 函数 | 调用方 | 处置 |
|---|------|------|-------|------|
| 1 | `utils/mockHelpers.ts` | `getPatientList()` | 无 | **直接删除整个文件** |
| 2 | `utils/mockHelpers.ts` | `getPatientById()` | 无 | 同上 |
| 3 | `utils/mockHelpers.ts` | `convertAPIPatientList()` | 无 | 同上 |
| 4 | `utils/mockHelpers.ts` | `convertAPIPatientToFullUI()` | 无 | 同上 |

---

# 4. 关键问题分级

## 4.1 P0 — 必须立即处理（数据安全/系统可信度风险）

| ID | 问题 | 文件 | 风险说明 |
|----|------|------|---------|
| P0-1 | 真实护士姓名硬编码在下拉选项中 | `DialysisProcessing.tsx` | 个人信息泄露；选项无法反映真实人员变动 |
| P0-2 | Statistics.tsx 零 API，所有统计数据假冒 | `Statistics.tsx` | 管理人员依此做决策会导致医疗安全风险 |
| P0-3 | HistoryTab 展示假治疗历史 | `HistoryTab.tsx` | 医生/护士查阅患者历史时看到的是假数据，临床安全风险 |
| P0-4 | MainLayout 通知栏全为假任务 | `MainLayout.tsx` | 医护人员无法识别真实临床预警 |

## 4.2 P1 — 应在下一个版本处理（功能完整性问题）

| ID | 问题 | 文件 |
|----|------|------|
| P1-1 | constants.ts 中 15 个 MOCK 常量死代码 | `constants.ts` |
| P1-2 | mockHelpers.ts 整体失效 | `utils/mockHelpers.ts` |
| P1-3 | DialysisProcessing.tsx 11 处硬编码日期 | `DialysisProcessing.tsx` |
| P1-4 | WardOverview.tsx processData 硬编码 | `WardOverview.tsx` |
| P1-5 | Monitoring.tsx 患者名"高敬兰" hardcode | `Monitoring.tsx` |

## 4.3 P2 — 计划内处理（体验与可维护性问题）

| ID | 问题 | 文件 |
|----|------|------|
| P2-1 | Monitoring.tsx 图表历史数据模拟生成 | `Monitoring.tsx` |
| P2-2 | MOCK_TASKS 患者 ID 不一致（P003/P006 无定义）| `MainLayout.tsx` |
| P2-3 | HistoryTab useMemo 依赖数组为 `[]` | `HistoryTab.tsx` |
| P2-4 | Statistics 后端接口设计与实现 | 后端新增 |

## 4.4 P3 — 技术债务（可延后处理）

| ID | 问题 |
|----|------|
| P3-1 | constants.ts 结构混乱（MOCK 与配置混存）|
| P3-2 | Monitoring.tsx PrescriptionEditModal hardcoded 耗材 |
| P3-3 | MOCK_DETAILED_ORDERS 中的 2024 年历史日期 |

---

# 5. 后端新增接口需求

| # | 接口 | 方法 | 路径 | 是否已有 | 优先级 |
|---|------|------|------|---------|-------|
| 1 | 护士列表（按角色过滤）| GET | `/api/v1/users?role=nurse&status=active` | 已有，确认参数即可 | P1 |
| 2 | 治疗历史列表（按患者）| GET | `/api/v1/treatments?patientId=:id&page=1&pageSize=25` | 已有，确认排序 | P1 |
| 3 | 临床任务列表 | GET | `/api/v1/clinical-tasks?status=pending` | **需新增** | P0 |
| 4 | 临床任务状态更新 | PUT | `/api/v1/clinical-tasks/:id/status` | **需新增** | P0 |
| 5 | 质控统计月报 | GET | `/api/v1/statistics/quality?year=2026` | **需新增** | P2 |
| 6 | 感染标志物统计 | GET | `/api/v1/statistics/infection?year=2026` | **需新增** | P2 |
| 7 | 血管通路统计 | GET | `/api/v1/statistics/vascular?year=2026` | **需新增** | P2 |
| 8 | 护士工作量统计 | GET | `/api/v1/statistics/workload?yearMonth=2026-04` | **需新增** | P2 |
| 9 | 病区实时状态 | GET | `/api/v1/wards/:id/status` | **需新增** | P1 |

---

# 6. 数据库变更需求

| # | 内容 | 类型 | 说明 |
|---|------|------|------|
| 1 | `clinical_tasks` 表 | 新增表 | 存储临床预警/任务（类型/患者/严重程度/状态/处理人）|
| 2 | 护士级别字段 | User 表扩展（可选）| 现有 User 无护士级别（N1-N4），确认是否通过字典表维护 |
| 3 | 病区容量字段 | Ward 表扩展 | 现有 Ward 无 capacity 字段，需添加 |
| 4 | 字典数据导入 | 现有 DictType/DictItem 表 | 确认护理级别字典（N1/N2/N3/N4）是否已录入 |

---

# 7. 给 Codex 的执行任务列表

---

## ✅ 阶段一：去 Mock 化与死代码清理（无需新增后端，立即可执行）

---

### T-A1：清理 constants.ts 死代码

**任务标题**：删除 constants.ts 中全部 MOCK_ 常量，仅保留 DASHBOARD_CARDS

**涉及文件**：`src/constants.ts`

**问题描述**：
文件共 384 行，含 15 个 MOCK_ 开头的常量，经确认均无任何页面引用，全部为死代码。唯一仍在使用的导出是第 191 行起的 `DASHBOARD_CARDS`（Dashboard.tsx 引用）。

**修改目标**：
- 删除第 16-189 行（MOCK_TREATMENT_PLAN 至 MOCK_PATIENTS）
- 删除第 207-383 行（MOCK_STATS_DATA 至 MOCK_MED_GROUPS）
- 保留 DASHBOARD_CARDS 定义及其类型引用
- 文件从 384 行缩减至约 30 行

**依赖后端接口**：否
**依赖数据库**：否
**允许先前端后后端**：是（纯删除）

**完成标准**：
- `npx tsc --noEmit` 零错误
- `grep -r "MOCK_" src/pages/` 零匹配
- `grep -r "MOCK_" src/layouts/` 零匹配
- constants.ts 仅导出 DASHBOARD_CARDS

**风险提示**：经验证 constants.ts **不导出任何类型**（所有类型来自 `types/original.ts`），删除 MOCK 常量不影响类型系统。但需确认删除后保留的 import 语句仅引入 DASHBOARD_CARDS 所需的 `DashboardCardConfig` 类型

---

### T-A2：删除 mockHelpers.ts

**任务标题**：删除整个 `utils/mockHelpers.ts` 文件

**涉及文件**：`src/utils/mockHelpers.ts`

**问题描述**：
文件含 4 个函数（getPatientList、getPatientById、convertAPIPatientList、convertAPIPatientToFullUI），全部返回假数据或 pass-through。经确认无任何文件 import 该模块，文件头注释也明确说明"临时 Mock 辅助函数，等待页面对接 API 后删除"。

**修改目标**：
- 直接删除文件
- **必须同步清理 re-export 链**（经验证确实存在）：
  - `src/utils/index.ts` 中删除 `export * from './mockHelpers'`
  - `src/services/index.ts` 中删除 `export { getPatientList, getPatientById } from '@/utils/mockHelpers'`

**依赖后端接口**：否

**完成标准**：
- `ls src/utils/mockHelpers.ts` 报 No such file
- `grep -r "mockHelpers" src/` 零匹配
- `grep -r "getPatientById" src/utils/` 零匹配（确认 utils 层不再导出该函数）

**风险提示**：**必须清理 re-export 链**，否则 TypeScript 编译会因找不到模块而报错

---

### T-A3：DialysisProcessing — 护士下拉动态化

**任务标题**：替换 DialysisProcessing.tsx 中所有硬编码护士姓名选项

**涉及文件**：`src/pages/DialysisProcessing.tsx`

**问题描述**：
文件中共 **12 处**硬编码护士姓名（经逐行验证：武琪迪 4 次 + 李俊雅 8 次）。分布在 `<option>` 下拉、`<span>` 签名、`<td>` 表格中（行 100/101/336/672/680/834/1031/1032/1033/1216/1444/1456），护士信息无法随人员变动更新，且存在个人信息泄露风险。

**修改目标**：
1. 在组件顶部添加：`const [nurses, setNurses] = useState<{id:string,name:string}[]>([])`
2. 在已有 useEffect 中追加加载：
   ```typescript
   restApi.getUserList({ status: 'active' })
     .then(res => {
       const nurseList = res.data.items
         .filter(u => u.role.includes('护') || u.role === 'nurse')
         .map(u => ({ id: u.id, name: u.realName || u.username }))
       setNurses(nurseList)
     })
   ```
3. 将所有硬编码 `<select>` 内的 `<option>` 替换为：
   ```tsx
   {nurses.map(n => <option key={n.id} value={n.id}>{n.name}</option>)}
   ```
4. 护士列表为空时显示 `<option value="">--请选择--</option>`

**依赖后端接口**：`GET /api/v1/users`（已有接口，用前端 filter 按 role 过滤护士）
**依赖数据库**：否（复用现有 User 表）
**允许先前端后后端**：是

**完成标准**：
- `grep "武琪迪\|李俊雅" src/pages/DialysisProcessing.tsx` 零匹配
- 护士下拉正常从 API 加载
- 无数据时显示"--请选择--"占位项

**风险提示**：现有 getUserList 接口需确认是否支持按 role 参数过滤；若不支持则在前端 `.filter()` 处理

---

### T-A4：DialysisProcessing — 日期 defaultValue 动态化

**任务标题**：替换 DialysisProcessing.tsx 中全部 `2025-12-13` 硬编码日期

**涉及文件**：`src/pages/DialysisProcessing.tsx`

**问题描述**：
共 11 处 `defaultValue` 含固定日期字符串（`2025-12-13 14:32`、`2025-12-13 08:18` 等），导致所有时间相关表单字段默认值错误。

**修改目标**：
- 普通文本输入（`type="text"`）：`defaultValue={new Date().toLocaleString('zh-CN',{hour12:false}).replace(/\//g,'-').slice(0,16)}`
- `datetime-local` 输入：`defaultValue={new Date().toISOString().slice(0,16)}`
- 若使用 antd DatePicker：`defaultValue={dayjs()}`

**依赖后端接口**：否

**完成标准**：
- `grep "2025-12-13" src/pages/DialysisProcessing.tsx` 零匹配
- 打开透析处理页面，时间字段默认显示当天日期

**风险提示**：部分 input 为受控/非受控混用，修改 defaultValue 不影响受控 input 的当前值，需注意区分

---

### T-A5：Monitoring — 删除硬编码患者名

**任务标题**：删除 Monitoring.tsx 中硬编码患者名"高敬兰"及其他假值

**涉及文件**：`src/pages/Monitoring.tsx`

**问题描述**：
`device.patientName || '高敬兰'` 导致设备空闲时显示假患者名；PrescriptionEditModal 中有硬编码耗材列表和临床参数默认值。

**修改目标**：
- `device.patientName || '高敬兰'` → `device.patientName || '--'`
- PrescriptionEditModal 中硬编码体重/超滤量/血压参数改为从当前选中设备的 patient 数据读取，无数据时显示空字符串 `''`
- 硬编码耗材列表保留但添加注释：`// TODO: 从 MaterialCatalog API 加载`

**依赖后端接口**：否（当前改动不需要）

**完成标准**：
- `grep "高敬兰" src/pages/Monitoring.tsx` 零匹配
- 空闲设备 patientName 显示"--"

**风险提示**：低

---

## ✅ 阶段二：已有后端接口的前端对接（复用现有 API）

---

### T-B1：HistoryTab — 接入真实治疗历史

**任务标题**：替换 patient-detail/HistoryTab.tsx 中的 mockHistoryData

**涉及文件**：`src/pages/patient-detail/tabs/HistoryTab.tsx`

**问题描述**：
`mockHistoryData` 用 useMemo 动态生成 25 条假治疗历史，依赖数组为空 `[]`，注释第 25 行明确标注"Placeholder treatment history — will be replaced when treatment-history API integration is completed"。

**修改目标**：
1. 删除 mockHistoryData 的整个 useMemo 块
2. 添加状态：
   ```typescript
   const [historyList, setHistoryList] = useState<TreatmentHistoryItem[]>([])
   const [loading, setLoading] = useState(false)
   ```
3. 添加 useEffect（注意依赖数组为 `[patientId]`）：
   ```typescript
   useEffect(() => {
     if (!patientId) return
     setLoading(true)
     restApi.getTreatments({ patientId: Number(patientId), pageSize: 25 })
       .then(res => {
         setHistoryList(res.data.items.map(t => ({
           id: String(t.id),
           date: t.treatmentDate?.slice(0, 10) ?? '',
           mode: t.treatmentType ?? '',
           duration: '',
           doctorSummary: t.notes ?? '',
           treatmentSummary: '',
         })))
       })
       .finally(() => setLoading(false))
   }, [patientId])
   ```
4. 无数据时显示"暂无治疗历史记录"空态提示

**依赖后端接口**：`GET /api/v1/treatments?patientId=:id&pageSize=25`（已有）
**依赖数据库**：否（Treatment 表已有）
**允许先前端后后端**：是

**完成标准**：
- `grep "mockHistoryData" src/` 零匹配
- 有数据时展示真实记录；无数据时显示空态
- useMemo 依赖数组不再为 `[]`

**风险提示**：`TreatmentHistoryItem.doctorSummary` 等字段在 RestTreatment 中无直接对应，映射时缺少字段用空字符串填充，**禁止填入假字符串**

---

### T-B2：WardOverview — processData 动态化

**任务标题**：WardOverview.tsx 的实时状态统计改为从 Treatment API 获取

**涉及文件**：`src/pages/WardOverview.tsx`

**问题描述**：
`processData = [{value:5,name:'等待'},{value:18,name:'透析中'},{value:3,name:'消毒'},{value:10,name:'完成'}]` 全部硬编码，无实际意义。

**修改目标**：
1. 在现有 loadData() 函数中追加：
   ```typescript
   const today = new Date().toISOString().slice(0, 10)
   const treatRes = await restApi.getTreatments({ treatmentDate: today, pageSize: 200 })
   const items = treatRes.data.items
   const statusCount = { 0: 0, 1: 0, 2: 0, 3: 0 }
   items.forEach(t => { statusCount[t.status as keyof typeof statusCount]++ })
   setProcessData([
     { value: statusCount[0], name: '等待' },
     { value: statusCount[1], name: '透析中' },
     { value: statusCount[2], name: '完成' },
     { value: statusCount[3], name: '已取消' },
   ])
   ```
2. processData 改为 useState，初始值为 `[]`
3. totalCapacity 改为 `// TODO: 从 Ward 表读取` 注释，暂时使用 patients.total 或保留固定值但明确标注

**依赖后端接口**：`GET /api/v1/treatments?treatmentDate=today`（已有）

**完成标准**：
- `grep "value:5\|value:18\|value:3\|value:10" src/pages/WardOverview.tsx` 零匹配
- 病区状态饼图数据来自真实统计

**风险提示**：Treatment API 按日期查询需确认时区处理正确，建议传 `treatmentDate` 参数而非在后端计算 today

---

### T-B3：DialysisProcessing — 治疗记录提交（流程走通关键）

**任务标题**：补充 DialysisProcessing.tsx 中治疗记录的创建/更新/状态变更 API 调用

**涉及文件**：`src/pages/DialysisProcessing.tsx`、`src/services/restClient.ts`

**问题描述**：
经代码审查确认，DialysisProcessing.tsx **目前只读不写**：能通过 `GET /patients/:id/treatment` 查询治疗数据，但没有调用任何 POST/PUT 接口来创建或更新治疗记录。而后端已有完整的治疗 CRUD 接口：
```
POST /api/v1/treatments              — 创建治疗记录（已有）
PUT  /api/v1/treatments/:id          — 更新治疗记录（已有）
PUT  /api/v1/treatments/:id/status   — 更新治疗状态（已有）
DELETE /api/v1/treatments/:id        — 删除治疗记录（已有）
```
这是"患者登记 → 排班 → **治疗创建** → 监控记录 → **治疗完成**"核心流程走通的关键缺口。

**修改目标**：
1. 在 restClient.ts 确认以下方法存在（若无则补充）：
   ```typescript
   async createTreatment(data: CreateTreatmentRequest): Promise<ApiSuccessResponse<RestTreatment>>
   async updateTreatment(id: number, data: UpdateTreatmentRequest): Promise<ApiSuccessResponse<RestTreatment>>
   async updateTreatmentStatus(id: number, status: number): Promise<ApiSuccessResponse<void>>
   ```
2. 在 DialysisProcessing.tsx 的"开始透析"操作中：
   - 若当日无治疗记录 → 调用 `POST /treatments` 创建
   - 若已有治疗记录 → 调用 `PUT /treatments/:id` 更新
3. 在"结束透析"操作中：
   - 调用 `PUT /treatments/:id/status` 将状态改为已完成（status=2）
4. 透前检查/透中参数/透后体征的提交：
   - 当前后端 Treatment 表通过关联表（TreatmentBeforeCheck、TreatmentDuringParam、TreatmentAfterSigns）存储
   - 若后端有对应子表的 CRUD 接口则直接调用
   - 若无，则在治疗主记录的 notes 字段临时存储，并添加 `// TODO: 补充子表 API` 注释

**依赖后端接口**：全部已有（`POST/PUT/DELETE /treatments`、`PUT /treatments/:id/status`）
**依赖数据库**：否（Treatment 表已存在）
**允许先前端后后端**：是（后端接口已存在，纯前端补齐）

**完成标准**：
- 患者可以通过界面发起一次透析治疗（点击"开始"后 DB 中 Treatment 表有新增记录）
- 可以结束透析（Treatment 的 status 变为 2-已完成）
- `npx tsc --noEmit` 零错误

**风险提示**：
- DialysisProcessing.tsx 文件较大（约1500行），修改时注意不破坏现有 UI 结构
- 创建治疗记录需要 patientId、treatmentDate、type 等必填字段，需从当前页面状态正确取值
- tenantId 和 creatorId 需从 AuthContext 获取当前用户信息

---

## ✅ 阶段三：新增后端接口与前端联调

---

### T-C1：新建 ClinicalTask 数据模型

**任务标题**：新建 `clinical_tasks` 表和 Go Model

**涉及文件**：
- 新增：`ai-hms-backend/internal/models/clinical_task.go`
- 修改：`ai-hms-backend/internal/database/migrate.go`

**Model 设计**：
```go
package models

import "time"

type ClinicalTask struct {
    ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    TenantId    int64      `gorm:"index;not null" json:"tenantId"`
    Type        string     `gorm:"size:32;not null" json:"type"`        // ALERT/PRESCRIPTION/ORDER/ASSESSMENT
    Title       string     `gorm:"size:100;not null" json:"title"`
    Description string     `gorm:"size:512" json:"description"`
    PatientId   *int64     `gorm:"index" json:"patientId"`
    PatientName string     `gorm:"size:50" json:"patientName"`
    BedNumber   string     `gorm:"size:20" json:"bedNumber"`
    Severity    string     `gorm:"size:20;default:medium" json:"severity"` // high/medium/low
    Status      string     `gorm:"size:20;default:pending" json:"status"`  // pending/handled/dismissed
    AssignedTo  *int64     `gorm:"index" json:"assignedTo"`
    HandledAt   *time.Time `json:"handledAt"`
    HandledBy   *int64     `json:"handledBy"`
    CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ClinicalTask) TableName() string {
    return "clinical_tasks"
}
```

在 migrate.go 的 AutoMigrate 列表中追加 `&models.ClinicalTask{}`

**依赖数据库**：是（新增 clinical_tasks 表）

**完成标准**：
- `go build ./...` 零错误
- AutoMigrate 在 debug 模式运行后表存在
- 表初始为空（正常，无需种子数据）

---

### T-C2：新建 ClinicalTask Service + Handler

**任务标题**：新建临床任务 Service 和 Handler

**涉及文件**：
- 新增：`ai-hms-backend/internal/services/clinical_task_service.go`
- 新增：`ai-hms-backend/internal/api/v1/clinical_task_handler.go`
- 修改：`ai-hms-backend/cmd/server/main.go`

**Service 方法**：
- `List(status string, tenantId int64) ([]ClinicalTask, error)` — 按状态查询
- `UpdateStatus(id int64, status string, handledBy int64) error` — 更新任务状态

**Handler 路由**：
```
GET  /api/v1/clinical-tasks          — 查询任务列表（支持 ?status=pending 过滤）
PUT  /api/v1/clinical-tasks/:id/status — 更新任务状态（handled/dismissed）
```

在 main.go 注册：`v1api.RegisterClinicalTaskRoutes(protected)`

**完成标准**：
- `GET /api/v1/clinical-tasks` 返回 `{"success":true,"data":{"items":[],"total":0}}`
- `go build ./...` 零错误

---

### T-C3：MainLayout — 替换 MOCK_TASKS

**任务标题**：MainLayout.tsx 通知栏接入真实 ClinicalTask 接口

**涉及文件**：
- `src/layouts/MainLayout.tsx`
- `src/services/restClient.ts`（添加 RestClinicalTask 接口和 getClinicalTasks 方法）

**修改目标**：
1. 在 restClient.ts 添加：
   ```typescript
   export interface RestClinicalTask {
     id: number
     type: string        // ALERT/PRESCRIPTION/ORDER/ASSESSMENT
     title: string
     description: string
     patientId?: number
     patientName: string
     bedNumber: string
     severity: string    // high/medium/low
     status: string      // pending/handled/dismissed
     createdAt: string
   }
   ```
   添加方法：`getClinicalTasks(params?: { status?: string })`

2. 在 MainLayout.tsx 中：
   - 删除 `MOCK_TASKS` 常量定义（第 27-53 行）
   - 添加 `const [tasks, setTasks] = useState<RestClinicalTask[]>([])`
   - useEffect 加载：`restApi.getClinicalTasks({ status: 'pending' }).then(res => setTasks(res.data.items))`
   - `visibleTasks` 改为 `tasks.filter(按角色权限过滤)`
   - 空态时显示"暂无待处理任务"

**依赖后端接口**：T-C2 完成后的接口（依赖 T-C1~T-C2）
**注意**：T-C2 未完成前，直接让 tasks 为 `[]`，显示空态，**禁止 fallback 到 MOCK_TASKS**

**完成标准**：
- `grep "MOCK_TASKS\|张伟.*A01\|李娜.*A02" src/layouts/MainLayout.tsx` 零匹配
- 通知栏加载中显示骨架；无任务显示"暂无待处理任务"

---

### T-C4 ~ T-C7：Statistics 后端接口

**任务标题**：新建质控统计相关后端接口

**涉及文件**：
- 新增：`ai-hms-backend/internal/services/statistics_service.go`
- 新增：`ai-hms-backend/internal/api/v1/statistics_handler.go`
- 修改：`ai-hms-backend/cmd/server/main.go`

**接口设计**：

```
GET /api/v1/statistics/quality?year=2026
  — 返回：{ items: [{month:1, ktv:0.0, hb:0.0, alb:0.0}, ...] }
  — 数据源：LabReport 表按月聚合

GET /api/v1/statistics/infection?year=2026
  — 返回：{ items: [{month:1, hbsAg:0, hcv:0, hiv:0, tp:0}, ...] }
  — 数据源：InfectionInfo 表按月聚合

GET /api/v1/statistics/vascular?year=2026
  — 返回：{ items: [{month:1, avf:0, avg:0, tcc:0}, ...] }
  — 数据源：VascularAccess 表按类型月统计

GET /api/v1/statistics/workload?yearMonth=2026-04
  — 返回：{ items: [{userId:1, name:'', treatments:0, punctures:0}, ...] }
  — 数据源：Treatment 表按 creatorId 聚合
```

**数据库依赖**：依赖 LabReport/InfectionInfo/VascularAccess/Treatment 表有真实数据
**允许先前端后后端**：是。后端未完成前，前端展示空态 `[]`，图表显示"暂无数据"

**完成标准**：
- `go build ./...` 零错误
- 无数据时接口返回 `{ items: [] }`，不返回假数据

---

### T-C8：Statistics.tsx — 接入真实统计接口

**任务标题**：Statistics.tsx 删除全部硬编码数组，接入真实统计接口

**涉及文件**：`src/pages/Statistics.tsx`

**修改目标**：
1. 删除第 38-74 行全部硬编码数组（qualityData/infectionData/vascularData/workloadData）
2. 删除第 84/91/98/105 行等硬编码统计卡片数值
3. 添加 4 个 useState，初始值均为 `[]`
4. useEffect 并行请求 4 个接口（参考 Dashboard.tsx 的 Promise.all 模式）
5. 加载中显示 Loading 组件；失败显示"数据加载失败，请刷新重试"
6. 统计卡片数值从接口数据中计算，无数据时显示"--"

**依赖后端接口**：T-C4~T-C7（依赖统计接口完成）
**注意**：T-C4~C7 未完成前，直接让数据为 `[]`，图表显示空态，**禁止保留任何硬编码数字**

**完成标准**：
- `grep "qualityData.*\[{\|infectionData.*\[{\|workloadData.*\[{" src/pages/Statistics.tsx` 零匹配
- 所有统计卡片数值不再是固定字面量
- `npx tsc --noEmit` 零错误

---

## ✅ 阶段四：上线前收尾

---

### T-D1：全量编译检查

```bash
cd ai-hms-frontend && npx tsc --noEmit     # 零错误
cd ai-hms-backend && go build ./...        # 零错误
cd ai-hms-backend && go vet ./...          # 零警告
```

### T-D2：Mock 残留扫描

```bash
grep -r "MOCK_" src/pages/
grep -r "MOCK_" src/layouts/
grep -r "武琪迪\|李俊雅\|高敬兰" src/
grep -r "2025-12-13" src/pages/
grep -r "mockHistoryData\|mockHelpers\|generateHistoryData" src/
```
以上全部应返回零匹配。

### T-D3：手工回归测试主流程

| 步骤 | 预期结果 |
|------|---------|
| 登录 → 患者列表 | 正常显示真实患者 |
| 患者详情 → 历史记录 tab | 显示真实治疗记录或"暂无历史记录" |
| 透析处理 → 护士下拉 | 显示数据库中真实护士，不含武琪迪/李俊雅 |
| 透析处理 → 时间字段 | 默认显示今天日期，不是 2025-12-13 |
| 统计报表页 | 显示真实数据或"暂无数据"，不是固定数字 |
| 主界面通知栏 | 显示真实任务或"暂无待处理任务" |
| 监控大屏 → 空闲设备 | 患者名显示"--"，不是"高敬兰" |
| 病区总览 → 饼图 | 显示今日实际透析状态分布 |

---

# 8. 整改完成后的复审检查清单

> Codex 完成整改后，将此清单交给架构师复审，逐项核对。

## 8.1 Mock 常量彻底清除检查

```bash
grep -r "MOCK_" src/pages/          # 应零匹配
grep -r "MOCK_" src/components/     # 应零匹配
grep -r "MOCK_" src/layouts/        # 应零匹配
grep -r "MOCK_" src/hooks/          # 应零匹配
grep -r "MOCK_" src/utils/          # 应零匹配
grep -c "^export const MOCK_" src/constants.ts   # 应为 0
wc -l src/constants.ts              # 应 < 50 行
```

## 8.2 硬编码人名检查

```bash
grep -rn "武琪迪\|李俊雅\|高敬兰" src/pages/
grep -rn "武琪迪\|李俊雅\|高敬兰" src/layouts/
grep -rn "张伟\|李娜\|王强\|孙行" src/pages/      # 应零匹配（人名不应出现在代码中）
grep -rn "刘护士长\|赵护士\|孙护士" src/pages/    # 应零匹配
```

## 8.3 硬编码日期检查

```bash
grep -rn "2025-12-13\|2024-05-04\|2024-05-10" src/pages/
grep -rn "defaultValue=.*202[0-9]-[0-9][0-9]-[0-9][0-9]" src/pages/DialysisProcessing.tsx
```
以上应全部零匹配。

## 8.4 模拟数据生成函数检查

```bash
grep -rn "generateHistoryData\|generateMiniGraphData" src/
grep -rn "mockHistoryData" src/
grep -rn "mockHelpers" src/
ls src/utils/mockHelpers.ts          # 应报 No such file
```

## 8.5 直接返回假数据的函数检查

- [ ] `utils/mockHelpers.ts` 文件已删除
- [ ] 无任何函数体内包含 `return MOCK_` 语句
- [ ] 无任何函数返回形如 `[{id:'P001', name:'张伟'...}]` 的硬编码数组

## 8.6 新增临时逻辑检查

- [ ] 所有新增的 useEffect 在错误处理中**不 fallback 到假数据**（只展示空态）
- [ ] Statistics.tsx：错误/空数据时显示"数据加载失败"或"暂无数据"，不展示任何具体数值
- [ ] MainLayout.tsx：接口失败时任务列表为空，不显示任何假任务
- [ ] 所有 `// TODO` 注释均有明确描述，无 `// 临时` 标记的假数据逻辑

## 8.7 伪整改检查（禁止以下模式）

```typescript
// ❌ 禁止出现的 fallback 模式：
const data = await fetchAPI() || MOCK_SOME_DATA
setStats(res.data ?? MOCK_STATS_DATA)
const list = response.data.length > 0 ? response.data : MOCK_DATA

// ✅ 正确的空态处理：
if (!data || data.length === 0) {
  setEmpty(true)  // 展示空态 UI
}
```

## 8.8 编译检查

```bash
cd ai-hms-frontend && npx tsc --noEmit    # 零错误
cd ai-hms-backend && go build ./...       # 零错误
cd ai-hms-backend && go vet ./...         # 零警告
```

## 8.9 页面功能验证（手工）

| 页面 | 检查项 | 预期结果 |
|------|-------|---------|
| DialysisProcessing | 护士下拉选项 | 真实护士列表，不含武琪迪/李俊雅 |
| DialysisProcessing | 时间字段默认值 | 显示今天日期，不是 2025-12-13 |
| patient-detail/历史记录 | 历史记录 tab | 真实治疗记录 或 "暂无历史记录" |
| Statistics | 统计图表 | 真实数据 或 "暂无数据"，不是固定数字 |
| MainLayout | 通知栏 | 真实待处理任务 或 "暂无待处理任务" |
| Monitoring | 空闲设备 | patientName 显示"--" |
| WardOverview | 病区状态饼图 | 今日实际透析状态分布 |

---

# 9. 结论

## 当前整体去 Mock 化完成度

| 维度 | 完成度 | 说明 |
|------|-------|------|
| 核心业务数据（患者/排班/设备/库存）| **95%** | 已全部接入真实 API |
| 统计报表数据 | **0%** | Statistics.tsx 完全 hardcoded |
| 临床通知/预警 | **0%** | 无后端支撑，全部假数据 |
| 患者治疗历史 | **10%** | API 已有，前端未对接 |
| 表单默认值/人员选项 | **30%** | 护士姓名/日期大量硬编码 |
| 死代码/废弃 MOCK 常量 | **5%** | 15 个 MOCK 常量仍保留在 constants.ts |

## 建议执行顺序

```
本周：T-A1 → T-A2 → T-A3 → T-A4 → T-A5
下周：T-B1 → T-B2 → T-B3（治疗提交，流程走通关键）→ T-C1 → T-C2
下下周：T-C3 → T-C4~C7 → T-C8
收尾：T-D1 → T-D2 → T-D3
```

## 关键风险提示

1. **临床安全风险**：HistoryTab 假历史、Statistics 假统计可能误导医护决策，**P0 优先解决**
2. **个人信息风险**：武琪迪、李俊雅、高敬兰 3 位真实姓名硬编码在源代码中，**下次部署前必须清除**
3. **技术债务风险**：constants.ts 约 350 行死代码若不清理，会持续误导新加入的开发者
4. **流程断点**：DialysisProcessing.tsx 当前只读不写，治疗记录无法提交到数据库，T-B3 是核心流程走通的必要前提

---

# 10. Codex 执行变更记录要求

> Codex 完成每个任务后，**必须**在以下文件中追加变更记录，便于 Claude Code 审查。

**变更记录文件**：`DEMOCK_CHANGELOG.md`（Codex 在项目根目录创建）

每个任务完成后追加一条记录，格式如下：

```markdown
## [T-A1] 清理 constants.ts 死代码
- **执行日期**：YYYY-MM-DD
- **修改文件**：
  - `src/constants.ts` — 删除 15 个 MOCK_ 常量，保留 DASHBOARD_CARDS
- **删除行数**：约 350 行
- **新增行数**：0
- **编译验证**：`npx tsc --noEmit` ✅ 零错误
- **残留扫描**：`grep -r "MOCK_" src/pages/` ✅ 零匹配
- **备注**：（如有特殊情况说明）
```

**记录要求**：
1. 每个 T-xx 任务独立一条记录
2. 列出所有实际修改/删除/新增的文件
3. 记录编译结果（前端 tsc + 后端 go build）
4. 记录残留扫描的 grep 命令和结果
5. 若任务执行中遇到计划外问题（如类型依赖、引用链未在计划中列出），需在备注中详细说明偏差原因和解决方式

---

# 11. 后续路线图（当前计划完成后）

> 本计划完成后的中长期开发路线。本节仅作为方向参考，具体任务拆解在对应阶段再做。

```
当前 → 阶段一~四：去 Mock 化 + 补齐缺失接口 + 核心流程走通（本计划）
  ↓
阶段五：角色与权限精细化
  - 后端：RequireRoles 中间件已有，需补充菜单级/操作级权限定义
  - 前端：根据角色动态隐藏/禁用菜单项和操作按钮
  - 数据库：新增 Permission 表 或 在 DictType 中管理权限配置
  ↓
阶段六：三方系统对接
  - HDIS 检验/检查报告同步（已完成 80%，需收尾测试）
  - 病历文书对接（需新增接口和数据模型）
  - 血透设备实时数据对接（需设备协议适配层）
  - 旧血透系统数据迁移（需编写迁移脚本和映射规则）
  ↓
阶段七：生产稳定性
  - 前端：auth.ts 与 restAuth.ts 双认证实现统一（当前存在冗余）
  - 后端：release 模式数据库迁移策略
  - 监控与告警：治疗中断、设备离线等场景的实时通知
  - 性能优化：大数据量患者列表分页、统计报表查询索引
```
