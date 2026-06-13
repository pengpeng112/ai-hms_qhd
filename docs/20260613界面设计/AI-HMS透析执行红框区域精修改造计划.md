# AI-HMS 透析执行页红框区域精修改造计划

参考预览文件：`AI-HMS透析执行红框区域优化预览.html`  
目标页面：AI-HMS 智能透析系统「透析执行」相关页面  
优化范围：截图红框标注的 4 个区域：

1. 左侧主菜单栏
2. 顶部二级业务菜单
3. 患者信息摘要区
4. 暂无治疗记录提示区

本次优化原则：**保持原有功能、接口、字段、路由、业务逻辑不变，只调整样式、字号、间距、布局层级和视觉密度。**

---

## 1. 改造目标

当前界面整体功能完整，但红框区域存在以下问题：

| 区域 | 当前问题 | 优化目标 |
|---|---|---|
| 左侧主菜单 | 字体偏小、层级感不够、菜单项识别度一般 | 放大字体，增强分组标题、菜单项、选中状态层级 |
| 顶部二级菜单 | 字体偏小，按钮高度和点击区域偏弱 | 增大字号和按钮高度，突出当前 Tab |
| 患者信息区 | 上下间距偏大，占用首屏高度较多 | 压缩为一行紧凑摘要卡，保留全部信息 |
| 暂无治疗记录提示 | 绿色提示区高度偏大，挤占表单空间 | 改为紧凑横向提示条，保留创建按钮 |

整体目标：

```text
不大改页面结构
不改变功能
不改变接口
不改变字段
不改变当前页面路由
提升可读性
减少首屏无效占高
让表单主体更早出现
```

---

## 2. 需要先定位的代码文件

请开发 AI 先在仓库中搜索以下关键词，定位真实组件文件：

```text
透析执行
透前评估
当日处方
双人核对
透析医嘱
透中监测
透后评估
健康宣教
透析小结
暂无治疗记录
创建治疗记录
当前登录身份
AI-HMS 智能透析
```

可能涉及文件包括但不限于：

```text
ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx
ai-hms-frontend/src/pages/dialysis-processing/execution/PreAssessment.tsx
ai-hms-frontend/src/pages/dialysis-processing/execution/*.tsx
ai-hms-frontend/src/components/layout/*.tsx
ai-hms-frontend/src/components/sidebar*.tsx
ai-hms-frontend/src/App.tsx
ai-hms-frontend/src/styles/*.css
ai-hms-frontend/src/index.css
```

如果当前项目仍有静态版本，也可能涉及：

```text
static/index.html
static/scripts/app.js
static/styles/components/sidebar.css
static/styles/*.css
```

实际改造时以仓库真实文件为准，不要凭空新增重复组件。

---

## 3. 不允许改动的内容

本次只做 UI 精修，以下内容不要改：

```text
后端接口
API 请求参数
字段名
治疗记录创建逻辑
患者切换逻辑
透析执行 Tab 切换逻辑
打印按钮逻辑
跳转按钮逻辑
当前登录身份逻辑
菜单权限逻辑
路由 path
状态枚举
保存/提交/暂存逻辑
```

如果发现某个信息字段为空或数据不准，本次只标记为“待确认”，不要借 UI 优化直接改业务逻辑。

---

## 4. 左侧主菜单优化方案

### 4.1 当前问题

截图中左侧主菜单整体字体偏小，尤其是：

```text
日常工作
工作台
病区概览
实时监控
透析执行
患者中心
排班
资源管理
系统配置
```

在 1920 宽屏上看起来偏弱，不像主导航。当前选中项虽然有蓝色背景，但文字和图标层级还可以加强。

---

### 4.2 目标样式

建议参数：

```text
侧边栏宽度：保持当前宽度，约 220px ~ 244px
品牌区高度：56px ~ 60px
分组标题字号：14px
分组标题字重：700 / 800
菜单项字号：15px
菜单项字重：650 / 700
菜单项高度：38px
菜单项圆角：10px
菜单项左右 padding：12px
图标字号：16px ~ 18px
图标宽度：18px ~ 20px
选中项左侧色条：3px
选中项背景：#1d63ff
选中项文字：#ffffff
```

---

### 4.3 具体改造点

#### 分组标题

将分组标题样式调整为：

```css
.menu-group-title {
  height: 30px;
  padding: 0 8px;
  font-size: 14px;
  font-weight: 800;
  color: #b8cae6;
  letter-spacing: 0.2px;
}
```

如果使用 Tailwind，可参考：

```tsx
className="flex h-[30px] items-center justify-between px-2 text-[14px] font-extrabold tracking-[0.2px] text-slate-300"
```

#### 菜单项

```css
.menu-item {
  height: 38px;
  margin: 4px 0;
  padding: 0 12px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 11px;
}
```

Tailwind 参考：

```tsx
className="relative my-1 flex h-[38px] items-center gap-3 rounded-[10px] px-3 text-[15px] font-bold"
```

#### 选中状态

```css
.menu-item.active {
  background: #1d63ff;
  color: #ffffff;
  box-shadow: 0 8px 20px rgba(29, 99, 255, 0.28);
}

.menu-item.active::before {
  content: "";
  position: absolute;
  left: 0;
  top: 9px;
  width: 3px;
  height: 20px;
  border-radius: 3px;
  background: #bfdbfe;
}
```

Tailwind 参考：

```tsx
className={clsx(
  'relative my-1 flex h-[38px] items-center gap-3 rounded-[10px] px-3 text-[15px] font-bold',
  active
    ? 'bg-blue-600 text-white shadow-[0_8px_20px_rgba(29,99,255,0.28)] before:absolute before:left-0 before:top-[9px] before:h-5 before:w-[3px] before:rounded before:bg-blue-200'
    : 'text-slate-200 hover:bg-white/5'
)}
```

---

### 4.4 验收标准

1. 左侧菜单文字明显比当前更清楚。
2. 分组标题和菜单项层级清楚。
3. 当前选中的“透析执行”一眼可识别。
4. 不影响菜单展开/折叠。
5. 不影响权限控制。
6. 不影响当前路由高亮。

---

## 5. 顶部二级业务菜单优化方案

### 5.1 当前问题

截图中顶部二级菜单包括：

```text
透前评估
当日处方
双人核对
透析医嘱
透中监测
透后评估
健康宣教
透析小结
```

当前按钮偏小，点击区域不够明显，活跃 Tab 虽然有蓝色背景，但整体看起来偏轻。

---

### 5.2 目标样式

建议参数：

```text
Tab 区域高度：48px
Tab 按钮高度：36px
Tab 字号：15px
Tab 字重：800 / 900
Tab 左右 padding：16px ~ 18px
Tab 间距：10px
Tab 圆角：11px
选中背景：#1d63ff
未选中背景：#eef2f7
未选中文字：#334155
选中阴影：0 8px 20px rgba(29,99,255,.22)
```

---

### 5.3 代码参考

```tsx
const tabClass = (active: boolean) =>
  clsx(
    'flex h-9 items-center rounded-[11px] px-4 text-[15px] font-black whitespace-nowrap transition',
    active
      ? 'bg-blue-600 text-white shadow-[0_8px_20px_rgba(29,99,255,0.22)]'
      : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
  )
```

如果二级菜单较多导致小屏溢出，外层使用横向滚动：

```tsx
<div className="flex h-12 items-center gap-2 overflow-x-auto whitespace-nowrap">
  ...
</div>
```

---

### 5.4 交互保留要求

1. 原 Tab 切换逻辑不变。
2. 原 active tab 识别逻辑不变。
3. 不改变 Tab 对应枚举值。
4. 不改变路由或 query 参数。
5. 小屏时允许横向滚动，不要换行挤压患者信息区。

---

### 5.5 验收标准

1. Tab 字体比当前更大。
2. 当前 Tab 更突出。
3. 所有 Tab 在 1920 宽屏下一行展示。
4. 小屏不换行挤乱布局。
5. 点击切换功能正常。

---

## 6. 患者信息摘要区优化方案

### 6.1 当前问题

红框中的患者信息区显示：

```text
支英俊
未排床
患者ID 300410
性别 / 年龄 女 / 72岁
费用类别 市职工普通
透龄 待补充
右侧：干体重 0kg、治疗方案 HD、打印/跳转按钮
```

当前上下高度偏大，占用首屏空间。建议压缩为一行紧凑摘要卡，同时保留所有信息。

---

### 6.2 目标布局

```text
┌──────────────────────────────────────────────────────────────┐
│ 支英俊  未排床     患者ID 300410   性别/年龄 女/72岁          │
│                  费用类别 市职工普通   透龄 待补充             │
│                                      干体重 0kg  治疗方案 HD  │
└──────────────────────────────────────────────────────────────┘
```

更推荐一行结构：

```text
左侧：姓名 + 状态
中间：4 个信息项
右侧：干体重 / 治疗方案 / 打印 / 跳转
```

---

### 6.3 样式参数

```text
摘要卡高度：72px ~ 78px
外边距：下方 12px
内边距：12px 16px
圆角：16px
背景：#ffffff
边框：#dbeafe
阴影：轻微
姓名字号：24px
姓名字重：900
状态标签字号：13px
信息 label 字号：12px
信息 value 字号：15px
右侧小卡高度：54px
右侧小卡宽度：80px ~ 88px
```

---

### 6.4 代码结构建议

```tsx
<section className="flex min-h-[76px] items-center gap-4 rounded-2xl border border-blue-100 bg-white px-4 py-3 shadow-sm">
  <div className="min-w-[180px]">
    <div className="flex items-center gap-2">
      <h2 className="m-0 text-2xl font-black text-slate-950">{patient.name}</h2>
      <span className="rounded-lg bg-blue-600 px-2.5 py-1 text-[13px] font-black text-white">
        {bedStatus || '未排床'}
      </span>
    </div>
  </div>

  <div className="grid flex-1 grid-cols-4 gap-4">
    <InfoItem label="患者ID" value={patient.patientId || patient.id || '--'} />
    <InfoItem label="性别 / 年龄" value={`${patient.gender || '--'} / ${patient.age || '--'}岁`} />
    <InfoItem label="费用类别" value={patient.feeType || '--'} />
    <InfoItem label="透龄" value={patient.dialysisAge || '待补充'} />
  </div>

  <div className="ml-auto flex items-center gap-2">
    <MiniMetric label="干体重" value={`${dryWeight || 0} kg`} tone="blue" />
    <MiniMetric label="治疗方案" value={dialysisMode || 'HD'} />
    <IconButton ... />
    <IconButton ... />
  </div>
</section>
```

---

### 6.5 响应式要求

在宽屏：

```text
姓名、患者信息、右侧指标同一行
```

在较窄屏：

```text
允许患者信息 grid 从 4 列变 2 列
右侧按钮可以换到下一行，但不要遮挡
```

Tailwind 可参考：

```tsx
<div className="grid flex-1 grid-cols-2 gap-3 xl:grid-cols-4">
```

---

### 6.6 验收标准

1. 患者信息区高度明显下降。
2. 患者姓名仍然突出。
3. 患者 ID、性别年龄、费用类别、透龄不丢失。
4. 干体重和治疗方案仍在右侧。
5. 打印和跳转按钮仍可点击。
6. 患者切换后信息正常刷新。
7. 小屏不发生错位。

---

## 7. 暂无治疗记录提示区优化方案

### 7.1 当前问题

绿色提示条占用上下空间较大，尤其在首屏中会推低表单主体区域。当前内容为：

```text
暂无治疗记录
可先创建今日治疗记录，再继续录入和查看。
创建治疗记录
```

这个提示很重要，但不需要占太高。

---

### 7.2 目标样式

改为紧凑横向提示条：

```text
[+] 暂无治疗记录  可先创建今日治疗记录，再继续录入和查看。        [创建治疗记录]
```

样式参数：

```text
高度：56px ~ 60px
圆角：15px
背景：#ecfdf5
边框：#86efac
图标尺寸：34px
标题字号：15px
说明字号：13px
按钮高度：34px
按钮圆角：10px
```

---

### 7.3 代码参考

```tsx
<section className="my-3 flex min-h-[58px] items-center gap-3 rounded-2xl border border-emerald-300 bg-emerald-50 px-4 py-2">
  <div className="flex h-8.5 w-8.5 items-center justify-center rounded-xl bg-emerald-600 text-white font-black">
    +
  </div>

  <div className="min-w-0">
    <div className="text-[15px] font-black text-emerald-800">暂无治疗记录</div>
    <div className="mt-0.5 text-[13px] font-semibold text-emerald-700">
      可先创建今日治疗记录，再继续录入和查看。
    </div>
  </div>

  <button
    type="button"
    onClick={handleCreateTreatment}
    className="ml-auto h-[34px] rounded-[10px] bg-emerald-600 px-4 text-[13px] font-black text-white hover:bg-emerald-700"
  >
    创建治疗记录
  </button>
</section>
```

如果当前没有治疗记录时还会展示多个提示，不要重复；保留一个主提示即可。

---

### 7.4 验收标准

1. 提示信息仍然清楚。
2. 创建治疗记录按钮仍可用。
3. 高度明显小于当前。
4. 不影响表单显示。
5. 治疗记录存在时该提示隐藏逻辑不变。

---

## 8. 顶部区域吸附策略

截图中顶部 Tab、患者摘要区位于表单上方。建议将 Tab 和患者摘要区作为同一个 sticky 区域，但要控制高度。

推荐：

```tsx
<div className="sticky top-0 z-20 border-b border-slate-200/70 bg-slate-50/95 pt-2 backdrop-blur">
  <TabBar />
  <PatientSummary />
</div>
```

注意：

1. sticky 区域不要包含“暂无治疗记录提示”，否则会占用滚动空间。
2. sticky 区域高度建议控制在 130px 内。
3. 如果页面已有全局 topbar，sticky 的 `top` 需要按实际布局调整。

---

## 9. 建议拆分组件

为降低改造风险，建议只做展示层组件拆分：

```text
SidebarMenu
ExecutionTabBar
PatientExecutionHeader
NoTreatmentNotice
InfoItem
MiniMetric
```

如果当前代码已经有这些组件，只改样式，不重复创建。

如果当前是单文件实现，可以先在同文件内抽小组件：

```tsx
function InfoItem({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div>
      <div className="text-xs font-extrabold text-slate-500">{label}</div>
      <div className="mt-1 text-[15px] font-black text-slate-950">{value || '--'}</div>
    </div>
  )
}
```

```tsx
function MiniMetric({ label, value, tone = 'default' }: { label: string; value: string; tone?: 'blue' | 'default' }) {
  return (
    <div className="flex h-[54px] w-20 flex-col items-center justify-center rounded-[13px] border border-slate-200 bg-slate-50">
      <span className="text-xs font-extrabold text-slate-500">{label}</span>
      <b className={clsx('mt-1 text-lg font-black', tone === 'blue' ? 'text-blue-600' : 'text-slate-950')}>
        {value}
      </b>
    </div>
  )
}
```

---

## 10. 推荐实施顺序

### Commit 1：左侧菜单字号和选中态优化

提交信息：

```text
fix: polish dialysis sidebar menu readability
```

内容：

1. 分组标题字号调到 14px。
2. 菜单项字号调到 15px。
3. 菜单项高度调到 38px。
4. 图标调到 16px ~ 18px。
5. 选中状态增加左侧色条和轻阴影。
6. 保留原菜单数据和权限逻辑。

---

### Commit 2：透析执行二级 Tab 优化

提交信息：

```text
fix: improve dialysis execution tab bar readability
```

内容：

1. Tab 高度调到 36px。
2. Tab 字号调到 15px。
3. Tab 字重加粗。
4. Active Tab 视觉增强。
5. 小屏加横向滚动。
6. 保留原切换逻辑。

---

### Commit 3：患者摘要区压缩

提交信息：

```text
fix: compact dialysis execution patient summary header
```

内容：

1. 患者摘要区改为一行紧凑卡片。
2. 保留姓名、状态、患者ID、性别年龄、费用类别、透龄。
3. 保留干体重、治疗方案、打印、跳转按钮。
4. 高度控制在 72px ~ 78px。
5. 小屏 grid 自适应。

---

### Commit 4：治疗记录提示条压缩

提交信息：

```text
fix: compact no treatment record notice
```

内容：

1. 暂无治疗记录提示改为紧凑横向提示条。
2. 保留创建治疗记录按钮。
3. 高度控制在 56px ~ 60px。
4. 治疗记录存在时隐藏逻辑不变。

---

### Commit 5：整体间距回归检查

提交信息：

```text
chore: verify dialysis execution header layout spacing
```

内容：

1. 检查首屏可见内容是否增加。
2. 检查不同 Tab 下是否都正常。
3. 检查 1366、1440、1920 宽度。
4. 检查侧边栏、患者列表、主内容滚动是否互不影响。

---

## 11. 详细验收清单

### 左侧菜单

- [ ] 菜单分组标题字号明显增大。
- [ ] 菜单项字号明显增大。
- [ ] 当前选中菜单更清楚。
- [ ] 图标和文字对齐。
- [ ] 菜单展开/折叠正常。
- [ ] 权限控制正常。
- [ ] 当前路由高亮正常。

### 二级 Tab

- [ ] Tab 字号增大。
- [ ] Tab 点击区域增大。
- [ ] 当前 Tab 高亮明显。
- [ ] Tab 切换正常。
- [ ] 小屏横向滚动正常。
- [ ] 不遮挡患者信息区。

### 患者摘要区

- [ ] 患者姓名显示正常。
- [ ] 未排床/床位状态显示正常。
- [ ] 患者 ID 显示正常。
- [ ] 性别/年龄显示正常。
- [ ] 费用类别显示正常。
- [ ] 透龄显示正常。
- [ ] 干体重显示正常。
- [ ] 治疗方案显示正常。
- [ ] 打印按钮正常。
- [ ] 跳转/打开按钮正常。
- [ ] 患者切换后数据刷新正常。
- [ ] 高度比原来更紧凑。

### 治疗记录提示

- [ ] 无治疗记录时显示提示。
- [ ] 有治疗记录时隐藏逻辑不变。
- [ ] 创建治疗记录按钮正常。
- [ ] 提示高度减少。
- [ ] 不影响下面表单。

### 整体

- [ ] 接口无变化。
- [ ] 字段无变化。
- [ ] 路由无变化。
- [ ] 保存/提交逻辑无变化。
- [ ] 控制台无 JS 错误。
- [ ] 1920 宽屏显示美观。
- [ ] 1366 宽度不错乱。
- [ ] 首屏表单主体更靠上。

---

## 12. 给开发 AI 的完整执行提示词

```markdown
请对 AI-HMS 智能透析系统「透析执行」页面红框区域进行低风险 UI 精修。参考用户确认的网页预览文件 `AI-HMS透析执行红框区域优化预览.html`，优化范围包括：左侧主菜单、顶部二级业务菜单、患者信息摘要区、暂无治疗记录提示区。

重要要求：
1. 不修改后端接口。
2. 不修改字段名。
3. 不修改治疗记录创建逻辑。
4. 不修改患者切换逻辑。
5. 不修改透析执行 Tab 切换逻辑。
6. 不修改打印和跳转按钮逻辑。
7. 不修改菜单权限逻辑。
8. 不修改路由 path。
9. 只调整样式、字号、间距、卡片高度、布局层级。
10. 保持原有功能完全可用。

需要先搜索定位组件：
- AI-HMS 智能透析
- 当前登录身份
- 透析执行
- 透前评估
- 当日处方
- 双人核对
- 透析医嘱
- 透中监测
- 透后评估
- 健康宣教
- 透析小结
- 暂无治疗记录
- 创建治疗记录

一、左侧主菜单优化：
- 分组标题字号调整为 14px，font-weight 800。
- 菜单项字号调整为 15px，font-weight 700。
- 菜单项高度调整为 38px。
- 图标字号调整为 16px~18px。
- 菜单项圆角 10px。
- 选中项保持蓝色背景，增加左侧 3px 高亮色条和轻微阴影。
- 不改变菜单数据来源、权限过滤和路由跳转。

二、顶部二级菜单优化：
- Tab 高度调整为 36px。
- Tab 字号调整为 15px。
- 字重调整为 800/900。
- 左右 padding 16px~18px。
- Tab 间距 10px。
- Active Tab 使用 #1d63ff 蓝色背景、白色文字和轻微阴影。
- 未选中 Tab 使用 #eef2f7 背景、#334155 文字。
- 小屏时允许横向滚动，不要换行挤压患者信息区。
- 不改变 Tab 枚举和值。

三、患者信息摘要区优化：
- 将患者信息区压缩为一行紧凑摘要卡。
- 高度控制在 72px~78px。
- 保留姓名、床位状态、患者ID、性别年龄、费用类别、透龄、干体重、治疗方案、打印按钮、跳转按钮。
- 姓名字号约 24px，font-weight 900。
- 信息 label 12px，value 15px。
- 右侧干体重和治疗方案用 54px 高的小卡片展示。
- 1366 宽度下患者信息 grid 可从 4 列变 2 列，避免错位。
- 不改变患者字段来源。

四、暂无治疗记录提示优化：
- 改为紧凑横向提示条。
- 高度控制在 56px~60px。
- 背景 #ecfdf5，边框 #86efac，圆角 15px。
- 左侧增加 34px 图标。
- 标题 15px，说明 13px。
- 创建治疗记录按钮高度 34px。
- 不改变按钮 onClick 和显示/隐藏逻辑。

五、整体布局：
- 可以将 Tab 和患者摘要区放到一个 sticky 区域，但不要把“暂无治疗记录”提示也 sticky。
- sticky 区域高度控制在 130px 内。
- 表单主体应比当前更靠上。
- 保持患者列表、左侧菜单和主内容滚动互不影响。

验收标准：
1. 左侧菜单字体明显更大。
2. 顶部二级菜单字体明显更大。
3. 患者信息区上下间距明显减少。
4. 暂无治疗记录提示条明显变矮。
5. 首屏能看到更多表单内容。
6. 所有原有功能正常。
7. 1366、1440、1920 宽度下不乱。
8. 控制台无 JS 错误。
```

---

## 13. 风险与注意事项

### 13.1 不要把患者摘要做得过小

患者姓名、床位状态、治疗方案是临床高频信息，压缩高度可以，但不能牺牲识别度。

### 13.2 不要让 Tab 换行

Tab 一旦换行，会比当前更占高度。小屏应横向滚动。

### 13.3 不要破坏 sticky 和滚动关系

如果当前页面已经有顶部固定栏，新增 sticky 时要确认 `top` 值，避免遮挡。

### 13.4 不要重复渲染“暂无治疗记录”

如果每个 Tab 都有自己的暂无治疗记录提示，应优先统一为一个公共提示，或保持当前逻辑但样式统一。

### 13.5 保留数据为空时的兜底

患者字段为空时统一显示：

```text
--
```

不要为了界面美观写死假数据。

---

## 14. 推荐最终效果

改造后页面应该呈现为：

```text
左侧菜单：更清楚、更像主导航
顶部 Tab：更大、更容易点
患者摘要：更紧凑，但信息完整
治疗记录提示：更轻、更省高度
表单主体：更靠上，首屏可见内容更多
```

本次属于低风险、高收益的界面精修，可以作为透析执行其他 Tab 的通用头部样式基础。
