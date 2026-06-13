# AI-HMS 排班管理界面 V3 优化方案：字体放大与右侧空白自适应

适用页面：`ai-hms-frontend/src/pages/SmartSchedulePage.tsx`

目标：在不修改接口、不修改字段、不影响排班生成、确认、拖拽、临时透析、CRRT、冲突/差异处理等原有功能的前提下，继续优化当前“排班管理”页面，重点解决两个问题：

1. 排班格子字体偏小，患者姓名、治疗模式、状态标识不够清楚。
2. 矩阵右侧出现明显空白，排班表没有充分利用页面宽度。

---

## 1. 当前代码问题定位

### 1.1 字体偏小的原因

当前矩阵单元格里患者姓名使用：

```tsx
<span className="font-semibold text-xs text-gray-800 truncate">
```

治疗模式和状态标识使用：

```tsx
text-[9px]
```

这会导致在 1920 宽屏或医院大屏上看起来偏小，尤其排班护士需要快速识别患者姓名时不够直观。

---

### 1.2 单元格高度和字体不匹配

当前单元格为：

```tsx
className="border p-0.5 min-w-[90px] h-[50px] ..."
```

患者卡片内部字体很小，但是单元格高度仍是 50px，所以会出现：

```text
格子不算特别小，但字很小；
一屏内容不少，但阅读费劲。
```

---

### 1.3 右侧空白的原因

当前 table 使用：

```tsx
<table className="border-collapse text-xs">
```

单元格使用：

```tsx
min-w-[90px]
```

但是没有让 table 最小宽度撑满容器，也没有根据容器宽度动态分配列宽。

当实际列数较少，或者当前只返回一周 6 天数据时，表格总宽度可能小于右侧容器，于是右侧出现大块空白。

---

## 2. 优化原则

本次不要继续盲目压缩，而是改成：

```text
高可读矩阵
```

具体原则：

1. 患者姓名必须清晰。
2. 治疗模式必须清楚。
3. 今日列仍然突出。
4. 草稿、透析中、完成、缺席、临时等状态仍然保留。
5. 表格宽度要优先撑满容器。
6. 只有当列数过多超过屏幕时，再启用横向滚动。
7. 不影响拖拽移动。
8. 不影响点击格子操作。

---

## 3. 推荐最终效果

### 3.1 字体大小建议

| 元素 | 当前 | 建议 |
|---|---:|---:|
| 表格整体 | `text-xs` | `text-[13px]` 或 `text-sm` |
| 患者姓名 | `text-xs` | `text-[13px]` |
| 治疗模式 | `text-[9px]` | `text-[10px]` 或 `text-[11px]` |
| 状态字：临/透/缺/完 | `text-[9px]` | `text-[11px]` |
| 机器列 | 默认小字 | `text-[13px] font-bold` |
| 日期表头 | 偏小 | `text-[13px]` |
| 班次表头 | 偏小 | `text-[12px]` |

---

### 3.2 单元格尺寸建议

不要用过小的 36px 高度。建议改成：

```text
舒适默认：
单元格宽度：自适应，最小 86px
单元格高度：44px
患者卡片高度：34px

紧凑模式：
单元格宽度：自适应，最小 78px
单元格高度：38px
患者卡片高度：30px
```

默认建议使用“舒适默认”，因为当前主要问题是字小，而不是内容太少。

---

### 3.3 右侧空白处理

核心方案：

```text
列宽 = max(最小列宽, 可用宽度 / 总班次数)
```

例如：

```ts
const leftMachineWidth = 96
const shiftCount = (board.dates?.length || 0) * (board.shifts?.length || 1)

const availableMatrixWidth = matrixWidth - leftMachineWidth
const autoShiftWidth = Math.floor(availableMatrixWidth / Math.max(shiftCount, 1))

const shiftColWidth = Math.max(density === 'compact' ? 78 : 86, autoShiftWidth)
```

这样：

1. 列少时自动变宽，填满右侧。
2. 列多时保持最小宽度，横向滚动。
3. 不会出现右侧大空白。

---

## 4. 具体代码改造建议

### 4.1 新增显示密度状态

在组件内新增：

```tsx
const [density, setDensity] = useState<'comfortable' | 'compact'>('comfortable')
```

可在顶部增加按钮：

```tsx
<Button size="small" type={density === 'comfortable' ? 'primary' : 'default'} onClick={() => setDensity('comfortable')}>
  舒适
</Button>
<Button size="small" type={density === 'compact' ? 'primary' : 'default'} onClick={() => setDensity('compact')}>
  紧凑
</Button>
```

---

### 4.2 新增矩阵宽度测量

在组件中引入：

```tsx
import { useRef } from 'react'
```

如果当前已经只引入 `useState, useEffect, useCallback`，改为：

```tsx
import { useState, useEffect, useCallback, useRef } from 'react'
```

新增：

```tsx
const matrixWrapRef = useRef<HTMLDivElement | null>(null)
const [matrixWidth, setMatrixWidth] = useState(0)

useEffect(() => {
  const el = matrixWrapRef.current
  if (!el) return

  const update = () => setMatrixWidth(el.clientWidth)
  update()

  const ro = new ResizeObserver(update)
  ro.observe(el)

  return () => ro.disconnect()
}, [])
```

---

### 4.3 计算自适应列宽

在渲染前计算：

```tsx
const machineColWidth = density === 'compact' ? 82 : 96
const minShiftColWidth = density === 'compact' ? 78 : 86
const cellHeight = density === 'compact' ? 38 : 44
const patientCardHeight = density === 'compact' ? 30 : 34

const shiftCount = Math.max(
  1,
  (board?.dates?.length || 0) * (board?.shifts?.length || 1)
)

const autoShiftColWidth = matrixWidth > 0
  ? Math.floor((matrixWidth - machineColWidth) / shiftCount)
  : minShiftColWidth

const shiftColWidth = Math.max(minShiftColWidth, autoShiftColWidth)

const tableMinWidth = machineColWidth + shiftColWidth * shiftCount
```

---

### 4.4 修改矩阵容器

当前：

```tsx
<div className="overflow-auto border rounded bg-white" style={{ maxHeight: '70vh' }}>
```

建议改为：

```tsx
<div
  ref={matrixWrapRef}
  className="overflow-auto border rounded bg-white"
  style={{
    height: 'calc(100vh - 215px)',
  }}
>
```

如果担心高度太高，可以先用：

```tsx
style={{ maxHeight: 'calc(100vh - 215px)' }}
```

但建议使用 `height`，这样右侧和下方滚动更稳定。

---

### 4.5 修改 table

当前：

```tsx
<table className="border-collapse text-xs">
```

建议：

```tsx
<table
  className={density === 'compact'
    ? 'border-collapse text-[12px]'
    : 'border-collapse text-[13px]'
  }
  style={{
    minWidth: tableMinWidth,
    width: tableMinWidth,
  }}
>
```

如果想让表格在列少时严格撑满容器，也可以：

```tsx
style={{
  minWidth: Math.max(tableMinWidth, matrixWidth),
  width: Math.max(tableMinWidth, matrixWidth),
}}
```

推荐后者。

---

### 4.6 修改表头宽度

机器列表头：

```tsx
<th
  className="sticky left-0 z-30 bg-gray-100 border px-2"
  rowSpan={2}
  style={{ width: machineColWidth, minWidth: machineColWidth }}
>
  病区 / 机器
</th>
```

日期表头：

```tsx
<th
  key={d}
  colSpan={board.shifts?.length || 1}
  className={`border px-1 text-[13px] ${isToday ? 'bg-amber-100' : ''}`}
  style={{
    width: shiftColWidth * (board.shifts?.length || 1),
    minWidth: shiftColWidth * (board.shifts?.length || 1),
  }}
>
```

班次表头：

```tsx
<th
  key={d + s.id}
  className={`border px-1 font-normal text-gray-500 ${isToday ? 'bg-amber-50' : ''}`}
  style={{ width: shiftColWidth, minWidth: shiftColWidth }}
>
  {s.name.replace('班', '')}
</th>
```

---

### 4.7 修改机器列

当前：

```tsx
<td className="sticky left-0 z-10 bg-white border px-2 font-medium cursor-pointer ...">
```

建议：

```tsx
<td
  className="sticky left-0 z-10 bg-white border px-2 font-bold cursor-pointer hover:bg-rose-50 whitespace-nowrap text-[13px]"
  style={{ width: machineColWidth, minWidth: machineColWidth }}
  onClick={...}
>
  <span className="text-slate-700">{m.code}</span>
  <span className="ml-1 text-[11px] font-semibold text-slate-400">{m.machineType}</span>
</td>
```

这样机器编号更清楚。

---

### 4.8 修改普通单元格

当前：

```tsx
className={`border p-0.5 min-w-[90px] h-[50px] align-middle cursor-pointer ...`}
```

建议改成：

```tsx
className={`border align-middle cursor-pointer ${colTint} ${dropHL ? 'bg-green-50' : ''}`}
style={{
  width: shiftColWidth,
  minWidth: shiftColWidth,
  height: cellHeight,
  padding: density === 'compact' ? 2 : 3,
}}
```

---

### 4.9 修改患者卡片

当前患者姓名：

```tsx
<span className="font-semibold text-xs text-gray-800 truncate">
```

建议：

```tsx
<span
  className={density === 'compact'
    ? 'truncate text-[12px] font-bold text-slate-900'
    : 'truncate text-[13px] font-bold text-slate-900'
  }
>
```

患者卡片：

```tsx
<div
  draggable
  ...
  className={`rounded-md border bg-white leading-tight ${
    density === 'compact' ? 'px-1 py-[1px]' : 'px-1.5 py-1'
  } ${cell.status === 10 ? 'border-dashed' : ''} ...`}
  style={{
    borderColor: modeColor(cell.dialysisMode),
    minHeight: patientCardHeight,
  }}
>
```

---

### 4.10 修改治疗模式和状态字体

当前：

```tsx
<span className="text-[9px] font-bold px-1 rounded">
```

建议：

```tsx
<span
  className={density === 'compact'
    ? 'rounded px-1 text-[10px] font-black'
    : 'rounded px-1.5 text-[11px] font-black'
  }
  style={{
    color: modeColor(cell.dialysisMode),
    background: modeColor(cell.dialysisMode) + '1a',
  }}
>
  {cell.dialysisMode}
</span>
```

状态：

```tsx
{cell.sourceType === 20 ? (
  <span className="text-[11px] font-black text-amber-600">临</span>
) : cell.status === 80 ? (
  <span className="text-[11px] font-black text-rose-500">缺</span>
) : cell.status === 50 ? (
  <span className="text-[11px] font-black text-emerald-600">透</span>
) : cell.status === 60 ? (
  <span className="text-[11px] font-black text-slate-400">完</span>
) : null}
```

---

### 4.11 空格不再显示明显“空”

当前：

```tsx
<div className="flex items-center justify-center h-full opacity-10 text-gray-400 text-xs">空</div>
```

建议：

```tsx
<div
  className={dropHL
    ? 'h-full rounded border border-dashed border-emerald-300 bg-emerald-50'
    : 'h-full rounded bg-slate-50/60'
  }
/>
```

这样不会干扰视线。

---

## 5. 右侧空白的两个可选方案

### 方案 A：自适应列宽，推荐

优点：

1. 最符合当前问题。
2. 列少时自动撑满右侧。
3. 列多时保持滚动。
4. 不改数据结构。

核心：

```tsx
const shiftColWidth = Math.max(minShiftColWidth, autoShiftColWidth)
```

---

### 方案 B：右侧空白展示辅助面板，不推荐作为第一版

可以在右侧空白显示：

```text
今日冲突
未排患者
操作提示
```

但这样需要重新布局，也可能影响矩阵操作区域。第一版不建议。

---

## 6. 当前可能还有一个业务展示问题

当前页面选择了 `2周`，但截图矩阵只显示周一到周六一周数据。代码里 `getBoard` 只传了：

```tsx
getBoard(dateStr)
```

而不是：

```tsx
getBoard(dateStr, weeks)
```

如果后端本来支持多周展示，前端可能没有把 `weeks` 传给 `getBoard`。如果后端不支持，则 `2周` 目前只影响生成和质量评分，不影响矩阵展示。

这点需要开发 AI 结合 `smartScheduleApi.ts` 和后端接口确认：

1. `getBoard` 是否支持 `weeks` 参数？
2. 如果支持，应改为 `getBoard(dateStr, weeks)` 或 `getBoard({ startDate: dateStr, weeks })`。
3. 如果不支持，则界面上应把 `2周` 文案说明为“生成范围”，不要让用户误以为矩阵展示两周。

这个是待确认项，不要直接改接口。

---

## 7. 建议实施步骤

### Commit 1：放大矩阵字体

```text
fix: improve smart schedule matrix readability
```

内容：

1. 患者姓名从 `text-xs` 提升到 `text-[13px]`。
2. 治疗模式从 `text-[9px]` 提升到 `text-[10px]/text-[11px]`。
3. 状态字“临/透/缺/完”提升到 `text-[11px]`。
4. 机器列字体加粗。
5. 日期和班次表头字体提升。

---

### Commit 2：增加舒适/紧凑密度

```text
feat: add smart schedule density toggle
```

内容：

1. 新增 `density` 状态。
2. 增加舒适/紧凑切换按钮。
3. 舒适模式默认。
4. 根据 density 切换：
   - cellHeight
   - minShiftColWidth
   - patientCardHeight
   - 字体大小
   - padding

---

### Commit 3：矩阵列宽自适应撑满右侧

```text
fix: adapt smart schedule columns to fill available width
```

内容：

1. 新增 `matrixWrapRef`。
2. 用 `ResizeObserver` 获取矩阵容器宽度。
3. 根据容器宽度、日期数、班次数计算 `shiftColWidth`。
4. table 设置 `width/minWidth`。
5. th/td 统一设置 width/minWidth。
6. 右侧不再出现大片空白。

---

### Commit 4：空格弱化和拖拽目标增强

```text
fix: reduce empty cell noise in schedule matrix
```

内容：

1. 空格不再显示明显“空”。
2. 普通空格显示浅灰背景。
3. 移动/拖拽目标空格显示绿色虚线。
4. 保留 onClick/onDrop 逻辑。

---

### Commit 5：确认 2周展示逻辑

```text
chore: verify schedule board week range display
```

内容：

1. 查看 `smartScheduleApi.ts` 的 `getBoard` 是否支持 weeks。
2. 如果支持，补传 weeks。
3. 如果不支持，不改接口，只在 UI 文案说明 `2周` 是生成范围。
4. 避免误导用户。

---

## 8. 验收清单

1. 排班页面正常打开。
2. 患者姓名明显变大。
3. 治疗模式 HD/HDF/CRRT 清楚。
4. 状态字“临/透/缺/完”清楚。
5. 机器列清楚。
6. 今日列高亮仍正常。
7. 草稿虚线仍正常。
8. 透析中光效仍正常。
9. 完成状态置灰仍正常。
10. 空格不再干扰视线。
11. 拖拽到空格仍正常。
12. 点击格子仍弹出操作菜单。
13. 右侧大空白消失。
14. 窗口缩放后列宽能重新计算。
15. 舒适/紧凑模式切换正常。
16. 横向滚动正常。
17. 纵向滚动正常。
18. sticky 表头正常。
19. sticky 机器列正常。
20. 原按钮功能全部正常。
21. 接口和字段未修改。
22. 控制台无 JS 报错。

---

## 9. 给开发 AI 的完整执行提示词

```markdown
请继续优化 AI-HMS 智能透析系统“排班管理”页面，文件为 `ai-hms-frontend/src/pages/SmartSchedulePage.tsx`。

当前页面已经是高密度矩阵，但存在两个问题：
1. 排班格子字体偏小，患者姓名、HD/HDF/CRRT 和状态字不够清楚。
2. 右侧出现明显空白，排班矩阵没有填满可用宽度。

重要要求：
1. 不修改后端接口。
2. 不修改字段名。
3. 不修改 `WeekBoard`、`CellDTO`、`MachineDTO`。
4. 不修改生成排班、确认、设为假日、方案变更、临时透析、CRRT、冲突处理、差异补排、拖拽移动、上机下机等业务逻辑。
5. 保留 sticky 表头和 sticky 左侧机器列。
6. 保留点击格子弹出操作菜单。
7. 保留拖拽移动到空格。

请执行以下优化：

一、字体放大：
- 患者姓名从 `text-xs` 调整为舒适模式 `text-[13px]`，紧凑模式 `text-[12px]`。
- 治疗模式从 `text-[9px]` 调整为 `text-[10px]` 或 `text-[11px]`。
- 状态字“临/透/缺/完”调整为 `text-[11px] font-black`。
- 机器列改为 `text-[13px] font-bold`。
- 日期表头使用 `text-[13px]`，班次表头使用 `text-[12px]`。

二、增加密度状态：
新增：
const [density, setDensity] = useState<'comfortable' | 'compact'>('comfortable')

根据 density 设置：
- machineColWidth: comfortable 96, compact 82
- minShiftColWidth: comfortable 86, compact 78
- cellHeight: comfortable 44, compact 38
- patientCardHeight: comfortable 34, compact 30

三、右侧空白自适应：
- 新增 matrixWrapRef 和 matrixWidth。
- 使用 ResizeObserver 获取矩阵容器宽度。
- 根据 `(matrixWidth - machineColWidth) / shiftCount` 计算 autoShiftColWidth。
- shiftColWidth = Math.max(minShiftColWidth, autoShiftColWidth)。
- table 设置 `width` 和 `minWidth` 为 `Math.max(tableMinWidth, matrixWidth)`。
- 所有日期 th、班次 th、td 都设置 width/minWidth。

四、空格弱化：
- 空格不再显示明显“空”字。
- 普通空格显示浅灰背景。
- 拖拽/移动目标空格显示绿色虚线背景。
- 保留 onClick/onDrop。

五、确认 2周展示：
- 检查 `smartScheduleApi.ts` 中 `getBoard` 是否支持 weeks。
- 如果支持，将 weeks 传入。
- 如果不支持，不要擅自改接口，只在 UI 说明 `2周` 是生成范围，不是矩阵展示范围。

验收：
1. 字体明显比当前大。
2. 右侧空白消失或显著减少。
3. 排班矩阵仍可横向/纵向滚动。
4. 拖拽、点击、确认、临时透析、CRRT、冲突和差异功能正常。
5. 控制台无 JS 错误。
```
