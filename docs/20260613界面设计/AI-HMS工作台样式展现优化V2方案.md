# AI-HMS 工作台样式与展现方式 V2 优化方案

适用页面：`ai-hms-frontend/src/pages/Dashboard.tsx`  
目标：在不修改接口、不修改字段、不改变现有路由跳转的基础上，进一步优化工作台样式和展现方式，使其更适合日常值班使用，而不是偏展示型大屏。

---

## 1. 优化原则

本次不改后端、不改字段、不改服务层，只改 `Dashboard.tsx` 的前端展示。

必须保持：

1. `restApi.getPatientList`
2. `getAllEquipments`
3. `getTodayTreatments`
4. `restApi.getDashboardStats`
5. 现有路由：
   - `/dialysis-processing`
   - `/monitoring`
   - `/schedule`
   - `/patients`
   - `/statistics`
   - `/device-binding`
6. 现有字段和状态判断逻辑。

---

## 2. 当前工作台可继续优化的问题

### 2.1 Hero 区偏“大屏展示”，日常使用略重

当前深色大卡很好看，但在日常使用中，用户更需要快速看到：

- 今天是否有排班
- 是否已经开始治疗
- 是否有异常
- 下一步点哪里

因此建议弱化“展示型大标题”，改成更实用的“今日工作台”。

---

### 2.2 业务提示不够突出

截图中：

```text
今日排班 51
今日透析 0
透析中 0
完成率 0%
```

这应该提示：

```text
今日已有排班安排，但暂无治疗记录。建议优先进入透析执行开始接诊或补录。
```

这个提示比单纯展示 0% 更有价值。

---

### 2.3 空状态应该带下一步动作

当前“暂无今日治疗记录”只是空状态。建议改成：

```text
暂无今日治疗记录
已有排班时，建议创建今日治疗记录。
[进入透析执行]
```

---

### 2.4 今日治疗节奏不要展示假峰值

当前代码在没有治疗数据时，fallback hourBars 可能使用 `Math.max(1, ...)`，容易在 0 治疗时仍显示条形。建议 0 治疗时全部显示 0，避免误导。

---

## 3. 推荐新版结构

```text
工作台
├── 顶部清爽页头
│   ├── 今日工作台
│   ├── 今日排班卡
│   └── 系统状态卡
│
├── 关键业务提示条
│   └── 有排班但无治疗记录时提示
│
├── 核心指标卡
│   ├── 今日透析
│   ├── 透析中
│   ├── 待关注
│   ├── 在档患者
│   └── 完成率
│
├── 今日待办
├── 常用入口
├── 治疗节奏
├── 今日透析患者
├── 透析执行动态
└── 设备与床位关注
```

---

## 4. 样式优化方向

### 4.1 顶部从“深色大屏 Hero”改成“白色清爽页头 + 状态卡”

建议将当前大面积深色 Hero 改成白色卡片：

```text
今日工作台
聚焦今日排班、透析执行、设备风险与待办处理。

今日排班：51 项
运行状态：正常 · 无设备关注
```

如果仍想保留深色视觉，可以只保留很小的深色条，不要占太多首屏高度。

---

### 4.2 增加关键业务提示条

当满足：

```ts
todayScheduleCount > 0 && todayTreatmentCount === 0
```

显示：

```tsx
<div className="rounded-2xl border border-blue-200 bg-blue-50 px-5 py-4 text-sm font-semibold text-blue-800">
  今日已有 {todayScheduleCount} 个排班安排，但暂无治疗记录。建议优先进入「透析执行」开始接诊或补录。
  <button onClick={() => navigate('/dialysis-processing')}>进入透析执行</button>
</div>
```

这条提示是整个工作台最实用的部分之一。

---

### 4.3 核心指标卡变得更轻

保留 5 张指标卡：

```text
今日透析
透析中
待关注
在档患者
完成率
```

每张卡顶部用 4px 色条区分状态，不需要大块渐变背景。

样式建议：

```tsx
className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"
```

卡片内部：

```text
标题
大数字
单位
一句解释
```

例如：

```text
今日透析
0 人次
当前治疗记录数
```

---

### 4.4 今日待办面板

新增一个“今日待办”面板，放在第二行左侧。

建议待办项：

```text
1 创建/补录治疗记录
  51 项排班待转入执行        透析执行

2 确认未开始患者
  关注接诊和候诊状态          患者列表

3 设备与床位巡检
  当前无异常设备              实时监测

4 透后评估补齐
  完成后及时评估记录          透后评估
```

这些数据可以先用现有字段前端计算，不新增接口。

---

### 4.5 常用入口面板

原来的 4 个入口保留，但从横向大卡改成一个面板里的 2x2 小卡：

```text
进入透析执行    最常用
查看实时监测    异常优先
维护今日排班    排班
患者资料中心    资料
```

好处：

1. 更紧凑。
2. 保留所有路由。
3. 视觉上能区分最常用和异常优先。

---

### 4.6 治疗节奏面板

如果今日治疗为 0，显示：

```text
08:00 0
09:00 0
10:00 0
...
```

不要用 `Math.max(1, ...)` 生成假条形。

提示：

```text
真实无记录时显示 0，避免误导。
```

---

### 4.7 底部三栏优化

保留当前三个面板，但样式更统一：

```text
今日透析患者
透析执行动态
设备与床位关注
```

优化：

1. 面板高度统一。
2. 空状态增加操作按钮。
3. 设备卡片更紧凑。
4. 患者列表行高统一。
5. 卡片圆角统一为 18–20px。

---

## 5. 建议代码调整点

### 5.1 修改 fallback hourBars

当前逻辑建议改为：

```ts
const fallbackHourBars = todayTreatmentCount > 0
  ? [
      { name: '08:00', value: Math.round(todayTreatmentCount * 0.25) },
      { name: '10:00', value: Math.round(todayTreatmentCount * 0.35) },
      { name: '12:00', value: Math.round(todayTreatmentCount * 0.25) },
      { name: '14:00', value: Math.round(todayTreatmentCount * 0.15) },
    ]
  : [
      { name: '08:00', value: 0 },
      { name: '09:00', value: 0 },
      { name: '10:00', value: 0 },
      { name: '11:00', value: 0 },
      { name: '12:00', value: 0 },
      { name: '13:00', value: 0 },
    ]

const hourBars = (dashboardStats?.treatmentsByHour?.length
  ? dashboardStats.treatmentsByHour
  : fallbackHourBars
).slice(0, 6)
```

---

### 5.2 新增业务提示条件

```ts
const hasScheduleWithoutTreatment = todayScheduleCount > 0 && todayTreatmentCount === 0
```

---

### 5.3 新增 todoItems

```ts
const completedTreatments = treatments.filter(item =>
  ['completed', '2', '30', '已完成'].includes(item.Status ?? '')
).length

const pendingScheduleCount = Math.max(todayScheduleCount - todayTreatmentCount, 0)

const todoItems = [
  {
    index: '1',
    title: '创建/补录治疗记录',
    desc: `${pendingScheduleCount} 项排班待转入执行`,
    action: '透析执行',
    route: '/dialysis-processing',
    tone: 'blue',
  },
  {
    index: '2',
    title: '确认未开始患者',
    desc: '关注接诊和候诊状态',
    action: '患者列表',
    route: '/patients',
    tone: 'teal',
  },
  {
    index: '3',
    title: '设备与床位巡检',
    desc: attentionDevices > 0 ? `${attentionDevices} 台设备需关注` : '当前无异常设备',
    action: '实时监测',
    route: '/monitoring',
    tone: attentionDevices > 0 ? 'orange' : 'green',
  },
  {
    index: '4',
    title: '透后评估补齐',
    desc: `${Math.max(todayTreatmentCount - completedTreatments, 0)} 人待评估`,
    action: '透后评估',
    route: '/dialysis-processing',
    tone: 'indigo',
  },
]
```

---

### 5.4 扩展 EmptyState

```tsx
function EmptyState({
  text,
  hint,
  action,
  onAction,
}: {
  text: string
  hint?: string
  action?: string
  onAction?: () => void
}) {
  return (
    <div className="flex min-h-32 flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-4 text-center text-sm text-slate-500">
      <div className="font-semibold text-slate-600">{text}</div>
      {hint && <div className="mt-2 text-xs text-slate-400">{hint}</div>}
      {action && onAction && (
        <button
          type="button"
          onClick={onAction}
          className="mt-4 rounded-xl bg-blue-600 px-4 py-2 text-xs font-bold text-white transition hover:bg-blue-700"
        >
          {action}
        </button>
      )}
    </div>
  )
}
```

---

## 6. 建议实施步骤

### Commit 1：工作台页头和业务提示优化

```text
fix: refine dashboard header and schedule treatment signal
```

内容：

1. 顶部 Hero 改为清爽页头。
2. 增加今日排班和系统状态卡。
3. 增加排班有值但无治疗记录提示。
4. 保留原统计数据来源。

---

### Commit 2：核心指标和今日待办优化

```text
feat: add practical dashboard todo section
```

内容：

1. 核心指标卡轻量化。
2. 新增 `todoItems`。
3. 新增“今日待办”面板。
4. 待办项支持跳转。

---

### Commit 3：常用入口和治疗节奏优化

```text
fix: improve dashboard quick actions and rhythm accuracy
```

内容：

1. 常用入口改为 2x2 紧凑卡片。
2. `hourBars` 在 0 治疗时显示 0。
3. 治疗节奏增加提示。
4. 不再展示虚假压力条。

---

### Commit 4：底部面板和空状态优化

```text
fix: polish dashboard panels and guided empty states
```

内容：

1. `EmptyState` 支持 hint/action。
2. 透析执行动态空状态增加“进入透析执行”。
3. 患者列表和设备卡片视觉统一。
4. 三个底部面板高度统一。

---

## 7. 验收清单

1. 页面正常打开。
2. 原接口正常调用。
3. 原字段没有修改。
4. 今日透析、透析中、待关注、在档患者、完成率显示正确。
5. 今日排班显示正确。
6. `todayScheduleCount > 0 && todayTreatmentCount === 0` 时出现提示。
7. 点击提示中的“进入透析执行”正常跳转。
8. 今日待办项点击正常跳转。
9. 常用入口跳转正常。
10. 0 治疗时治疗节奏不显示假数据。
11. 空状态显示下一步指引。
12. 设备与床位关注显示正常。
13. 1366x768 下首屏信息更集中。
14. 控制台无 JS 报错。

---

## 8. 给开发 AI 的执行提示词

```markdown
请继续优化 AI-HMS 智能透析系统工作台页面，文件是 `ai-hms-frontend/src/pages/Dashboard.tsx`。本次只优化样式和展现方式，不修改接口、不修改字段、不修改路由。

当前页面已经能加载患者、设备、今日治疗和 dashboardStats。请保留这些数据来源。

优化目标：
1. 工作台从展示型大屏改为日常实用清爽版。
2. 顶部 Hero 改为白色清爽页头，显示“今日工作台”、今日排班和系统状态。
3. 当 todayScheduleCount > 0 且 todayTreatmentCount === 0 时，显示醒目的业务提示：今日已有排班安排，但暂无治疗记录，建议进入透析执行开始接诊或补录。
4. 核心指标卡轻量化，保留：今日透析、透析中、待关注、在档患者、完成率。
5. 新增“今日待办”面板，包括创建/补录治疗记录、确认未开始患者、设备与床位巡检、透后评估补齐。
6. 常用入口改成 2x2 紧凑卡片，保留原路由。
7. 治疗节奏在今日治疗为 0 时全部显示 0，不要用 Math.max(1, ...) 生成假条形。
8. EmptyState 增加 hint 和 action，空治疗记录时提供“进入透析执行”按钮。
9. 底部三个面板：今日透析患者、透析执行动态、设备与床位关注，保持原功能但统一样式。
10. 不改 services，不改 API，不改字段名。

验收：
1. 页面加载正常。
2. 所有跳转正常。
3. 统计值正确。
4. 有排班无治疗记录时能看到提示。
5. 0 治疗时节奏图不显示假数据。
6. 控制台无 JS 报错。
```
