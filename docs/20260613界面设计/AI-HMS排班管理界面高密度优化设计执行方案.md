# AI-HMS 排班管理界面高密度优化设计执行方案

适用页面：`ai-hms-frontend/src/pages/SmartSchedulePage.tsx`  
目标：不修改接口、不修改字段、不改变排班生成、确认、拖拽移动、临时透析、CRRT、冲突差异处理等原有功能，只优化排班管理界面的布局、密度和矩阵展示方式，让一屏能显示更多排班内容，尤其解决“下方还有排班没显示出来”的问题。

---

## 1. 当前代码结构核查

当前 `SmartSchedulePage.tsx` 功能比较完整：

1. 页面状态包括：
   - `currentDate`
   - `weeks`
   - `board`
   - `loading`
   - `conflicts`
   - `quality`
   - `diffs`
   - `crrts`
   - `moveSrc`
   - `dragSrc`
   - `confirmDate`

2. 当前 API 能力包括：
   - 获取排班矩阵 `getBoard`
   - 生成排班 `generateSchedule`
   - 整盘确认 `confirmPlan`
   - 次日/当日确认 `confirmDay`
   - 取消、缺席、移动排班
   - 临时透析
   - CRRT
   - 机器停机
   - 设置假日
   - 方案变更
   - 补排
   - 冲突列表与处理
   - 差异列表
   - 质量评分
   - 上机、下机

3. 当前排班矩阵已经支持：
   - sticky 表头
   - sticky 左侧机器列
   - 横向滚动
   - 纵向滚动
   - 拖拽移动
   - 点击单元格操作
   - 今日列高亮
   - 草稿虚线边框
   - 透析中发光
   - 完成状态置灰

因此本次只做前端布局和样式优化，不改接口和业务逻辑。

---

## 2. 当前界面主要问题

从截图看，当前页面的主要问题不是功能不足，而是空间利用率不高：

### 2.1 顶部区域占高过多

当前顶部依次有：

```text
标题
日期 / 周数 / 刷新 / 生成排班 / 管理
确认 + 扰动工具栏
图例 / 消息
质量评分 6 个卡片
Tabs
排班矩阵
```

这些控件加起来占用了较多垂直空间，导致真正的排班矩阵高度不足。

---

### 2.2 质量评分卡片过高

当前质量评分是 6 个 `Card + Statistic`，占用一整行较高空间：

```text
达标率
利用率
稳定率
综合分
冲突
患者
```

建议改成一行细条 chip，不要使用大卡。

---

### 2.3 排班单元格偏高偏宽

当前单元格大约：

```text
min-w-[90px]
h-[50px]
```

单元格内患者卡也有上下 padding。对于排班管理这种矩阵页，应该支持高密度模式：

```text
单元格宽度 70–78px
行高 34–40px
患者卡高度 26–30px
```

这样一屏可以显示更多机器行和更多下方病区。

---

### 2.4 空格显示“空”影响密度

当前空格显示：

```text
空
```

虽然透明度低，但大量空格仍然增加视觉噪声。建议高密度模式下空格仅显示淡灰背景或很淡的小点，不显示明显文字。

---

### 2.5 Tabs 占用空间

当前 `Tabs` 包含：

```text
周排班矩阵
冲突
差异
```

在高密度排班页中，Tabs 可以改为矩阵卡片内的紧凑胶囊按钮，减少高度。

---

## 3. 推荐新版布局

```text
排班管理
├── 紧凑页头
│   ├── 标题 + 简短说明
│   ├── 日期
│   ├── 周数
│   ├── 刷新
│   ├── 生成排班
│   ├── 管理
│   ├── 紧凑格子
│   ├── 全屏矩阵
│   └── 隐藏统计
│
├── 确认与扰动工具条
│   ├── 整盘确认
│   ├── 确认日期
│   ├── 次日确认
│   ├── 当日确认
│   ├── 设为假日
│   ├── 方案变更
│   ├── 临时透析
│   └── CRRT
│
├── 质量与图例细条
│   ├── 达标率 / 利用率 / 稳定率 / 综合分 / 冲突 / 患者
│   ├── HD / HDF / CRRT 图例
│   └── 周排班矩阵 / 冲突 / 差异
│
└── 排班矩阵主区域
    ├── 矩阵工具条
    ├── sticky 日期表头
    ├── sticky 机器列
    ├── 高密度单元格
    └── 内部横纵向滚动
```

---

## 4. 具体优化内容

### 4.1 页头压缩为 96px 左右

将标题、日期、周数、生成排班、管理、显示模式放到一个白色卡片中。

建议：

```text
排班管理
高密度矩阵模式：压缩顶部信息，把屏幕高度优先留给排班表。

[日期] [2周] [刷新] [生成排班] [管理] [紧凑格子] [全屏矩阵] [隐藏统计]
```

---

### 4.2 确认与扰动工具条合并到页头第二行

当前确认和扰动工具条单独占一行。建议保留所有按钮，但放在页头第二行：

```text
确认：整盘确认 [周一 06-08] 次日确认 当日确认
扰动：设为假日 方案变更 +临时透析 +CRRT
```

好处：

1. 功能不变。
2. 减少上下 margin。
3. 用户仍然能快速找到操作。

---

### 4.3 质量评分改为细条

将 6 个卡片改为一行小 chip：

```text
达标率 0%
利用率 308%
稳定率 0%
综合分 0/100
冲突 7,797
患者 0/360
```

建议高度控制在 44–52px。

---

### 4.4 图例和 Tabs 合并

原来的图例：

```text
HD
HDF
CRRT
✓=确认级别
```

原来的 Tabs：

```text
周排班矩阵
冲突(100)
差异(360)
```

建议合并到质量细条右侧：

```text
HD HDF CRRT
● 确认级别 · 虚线=草稿 · 绿色光晕=透析中
[周排班矩阵] [冲突(100)] [差异(360)]
```

这样减少一行空间。

---

### 4.5 排班矩阵区域最大化

当前矩阵容器：

```tsx
style={{ maxHeight: '70vh' }}
```

建议改为：

```tsx
style={{ height: compactMode ? 'calc(100vh - 244px)' : '70vh' }}
```

或者用 Tailwind：

```tsx
className="h-[calc(100vh-244px)] overflow-auto"
```

这样可以让矩阵占据剩余屏幕高度。

---

### 4.6 增加紧凑模式

新增状态：

```ts
const [density, setDensity] = useState<'compact' | 'comfortable'>('compact')
const [hideQuality, setHideQuality] = useState(false)
const [matrixFullscreen, setMatrixFullscreen] = useState(false)
```

单元格尺寸根据 density 切换：

```ts
const cellClass =
  density === 'compact'
    ? 'min-w-[74px] h-[36px] p-[2px]'
    : 'min-w-[90px] h-[50px] p-0.5'
```

患者卡尺寸：

```ts
const patientCardClass =
  density === 'compact'
    ? 'rounded border bg-white px-1 py-[1px] leading-none'
    : 'rounded-md border bg-white px-1 py-0.5 leading-tight'
```

---

### 4.7 压缩单元格内容

当前单元格展示：

```text
患者姓名
确认点
透析模式
状态标记
```

紧凑模式建议仍保留这些信息，但减少高度：

```text
患者名 · 确认点
HD  临/缺/透/完
```

高密度建议：

1. 姓名单行 truncate。
2. 模式 chip 小到 9–10px。
3. 确认点保留。
4. 临时/缺席/透析中/完成状态保留一个字。
5. 不显示明显“空”字，空格只保留浅灰底。

---

### 4.8 病区标题行压缩

当前病区标题行：

```text
A-1区 (8台)
```

建议高度控制在 22–26px，背景浅灰即可。这样多个病区不会占用太多高度。

---

### 4.9 增加全屏矩阵模式

全屏矩阵不是浏览器全屏，而是页面内隐藏顶部统计、减少页头高度：

```ts
const matrixTopOffset = matrixFullscreen ? 120 : hideQuality ? 180 : 244
```

逻辑：

1. 保留顶部最小操作条。
2. 隐藏质量统计和图例。
3. 矩阵高度扩大。
4. 再点一次恢复。

这个不影响原有功能，只是显示模式。

---

### 4.10 下方排班可见性优化

为解决“下方还有排班没显示出来”，建议：

1. 默认启用 `compact`。
2. 默认矩阵高度使用 `calc(100vh - 244px)`。
3. 质量区可以隐藏。
4. 空格弱化。
5. 病区标题行压缩。
6. 单元格高度从 50px 降至 36px。
7. 患者卡高度约 28px。
8. 内部纵向滚动条始终显示或更明显。

这样同屏机器行数量可明显增加。

---

## 5. 建议代码调整范围

只改：

```text
ai-hms-frontend/src/pages/SmartSchedulePage.tsx
```

不改：

```text
smartScheduleApi.ts
后端接口
WeekBoard
CellDTO
MachineDTO
generateSchedule
confirmPlan
confirmDay
moveShift
insertTemporary
insertCrrt
machineOutage
setHoliday
planChange
makeup
startTreatment
completeTreatment
```

---

## 6. 建议实施步骤

### Commit 1：压缩顶部工具栏

```text
fix: compact smart schedule toolbar layout
```

内容：

1. 标题、日期、周数、刷新、生成排班、管理合并为紧凑页头。
2. 确认和扰动工具条合并到页头第二行。
3. 减少 `mb-2`、`py`、`gap`。
4. 保留所有按钮和原 onClick。

---

### Commit 2：质量统计与图例细条化

```text
fix: convert schedule quality cards to compact status strip
```

内容：

1. 将 6 个 `Card + Statistic` 改成一行 chip。
2. 图例和消息合并。
3. 周排班矩阵、冲突、差异入口改为紧凑 tab。
4. 保留冲突和差异内容。

---

### Commit 3：高密度矩阵样式

```text
fix: add compact density for smart schedule matrix
```

内容：

1. 新增 `density` 状态。
2. 单元格 compact 模式：
   - min-width 74px
   - height 36px
   - padding 2px
3. 患者卡片压缩。
4. 空格弱化。
5. 病区标题行压缩。
6. sticky 表头和 sticky 左列保持。

---

### Commit 4：矩阵高度与全屏模式优化

```text
fix: maximize smart schedule matrix viewport
```

内容：

1. 矩阵容器高度改为 `calc(100vh - xxxpx)`。
2. 新增 `hideQuality`。
3. 新增 `matrixFullscreen`。
4. 全屏矩阵模式隐藏统计区，保留核心操作。
5. 横纵向滚动正常。

---

## 7. 关键代码参考

### 7.1 新增状态

```ts
const [density, setDensity] = useState<'compact' | 'comfortable'>('compact')
const [hideQuality, setHideQuality] = useState(false)
const [matrixFullscreen, setMatrixFullscreen] = useState(false)
```

### 7.2 矩阵高度

```ts
const matrixHeight = matrixFullscreen
  ? 'calc(100vh - 150px)'
  : hideQuality
    ? 'calc(100vh - 200px)'
    : 'calc(100vh - 244px)'
```

使用：

```tsx
<div className="overflow-auto rounded-xl border bg-white" style={{ height: matrixHeight }}>
```

### 7.3 单元格尺寸

```ts
const tdClass = density === 'compact'
  ? 'border p-[2px] min-w-[74px] h-[36px] align-middle cursor-pointer'
  : 'border p-0.5 min-w-[90px] h-[50px] align-middle cursor-pointer'
```

### 7.4 患者卡片

```ts
const cellCardClass = density === 'compact'
  ? 'rounded border bg-white px-1 py-[1px] leading-none'
  : 'rounded-md border bg-white px-1 py-0.5 leading-tight'
```

### 7.5 空格显示

```tsx
{cell ? (
  ...
) : (
  <div className={density === 'compact'
    ? 'h-full rounded bg-slate-50/40'
    : 'flex h-full items-center justify-center text-xs text-slate-300 opacity-30'
  }>
    {density === 'compact' ? null : '空'}
  </div>
)}
```

---

## 8. 验收清单

1. 排班页面正常打开。
2. 日期切换正常。
3. 周数切换正常。
4. 刷新正常。
5. 生成排班正常。
6. 整盘确认正常。
7. 次日确认正常。
8. 当日确认正常。
9. 设置假日正常。
10. 方案变更正常。
11. 临时透析正常。
12. CRRT 正常。
13. 冲突 tab 正常。
14. 差异 tab 正常。
15. 拖拽移动正常。
16. 点击格子操作正常。
17. 上机/下机正常。
18. 机器停机登记正常。
19. sticky 表头正常。
20. sticky 左侧机器列正常。
21. 横向滚动正常。
22. 纵向滚动正常。
23. 紧凑模式下下方病区和机器行显示更多。
24. 全屏矩阵模式正常。
25. 隐藏统计后矩阵高度增加。
26. 接口和字段未修改。
27. 控制台无 JS 报错。

---

## 9. 给开发 AI 的完整执行提示词

```markdown
请优化 AI-HMS 智能透析系统“排班管理”页面，文件为 `ai-hms-frontend/src/pages/SmartSchedulePage.tsx`。请保持接口、字段和现有业务逻辑不变，只优化布局、样式和显示密度。目标是尽量显示更多排班内容，解决当前页面下方排班需要大量滚动才能看到的问题。

重要要求：
1. 不修改 `smartScheduleApi.ts`。
2. 不修改后端接口。
3. 不修改 `WeekBoard`、`CellDTO`、`MachineDTO` 数据结构。
4. 不改变生成排班、确认、临时透析、CRRT、方案变更、设为假日、机器停机、冲突处理、差异补排、拖拽移动、上机下机等功能。
5. 保留 sticky 表头和 sticky 左侧机器列。
6. 保留横向滚动和纵向滚动。
7. 保留点击格子弹出排班操作。
8. 保留拖拽移动到空格。

当前问题：
1. 顶部工具栏、确认扰动工具栏、图例、质量评分卡和 Tabs 占用高度较多。
2. 质量评分 6 个卡片太高。
3. 单元格 min-w 90px、h 50px，导致一屏显示行数少。
4. 空格显示“空”造成视觉噪声。
5. 下方病区和机器排班不容易看到。

优化目标：
1. 顶部压缩为紧凑页头，包含标题、日期、周数、刷新、生成排班、管理。
2. 确认和扰动工具条合并到页头第二行。
3. 质量评分从 6 个大卡改成一行细条 chip。
4. 图例、消息、周排班矩阵/冲突/差异入口合并到一行细条。
5. 新增显示模式：
   - density: compact / comfortable
   - hideQuality
   - matrixFullscreen
6. 默认使用 compact。
7. compact 模式下：
   - 单元格宽度约 74px
   - 单元格高度约 36px
   - 患者卡片高度约 28px
   - 空格不显示明显“空”字，只保留淡背景
8. 病区标题行压缩到 22–26px。
9. 排班矩阵容器高度改为 `calc(100vh - xxxpx)`，让矩阵占满剩余高度。
10. 全屏矩阵模式隐藏统计和次要信息，只保留核心操作，让矩阵显示更多行。
11. 不影响拖拽、点击、状态显示、今日列高亮和草稿/透析中/完成状态样式。

建议新增状态：
const [density, setDensity] = useState<'compact' | 'comfortable'>('compact')
const [hideQuality, setHideQuality] = useState(false)
const [matrixFullscreen, setMatrixFullscreen] = useState(false)

验收：
1. 页面正常加载。
2. 原所有按钮功能正常。
3. 拖拽移动正常。
4. 点击格子操作正常。
5. 冲突和差异 tab 正常。
6. 紧凑模式下一屏能看到更多机器行和下方病区。
7. 全屏矩阵模式正常。
8. 接口和字段未修改。
9. 控制台无 JS 错误。
```
