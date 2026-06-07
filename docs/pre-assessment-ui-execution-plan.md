# 透析执行 - 透前评估界面优化执行计划

> 状态：可交给其他 AI 执行  
> 页面：`ai-hms-frontend/src/pages/dialysis-processing/execution/PreAssessment.tsx`  
> 容器：`ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`  
> 目标：只优化透前评估布局、间距、对齐和视觉层级；不改接口、不改字段语义、不改保存逻辑。

## 1. 当前问题定位

### 1.1 主布局不均衡
- 当前 `PreAssessment.tsx` 第 187 行使用 `grid grid-cols-1 xl:grid-cols-4 gap-6`。
- 里面只有 3 个 `Card`，在 `xl:grid-cols-4` 下会占 3/4 宽度，右侧留空，整体不对称。
- 三张卡片内容量差异很大：`体重与容量评估` 4 项、`生命体征` 6 项、`通路与症状` 含症状和备注，导致高度明显不一致。

### 1.2 卡片样式过重且不符合现有 UI 约束趋势
- `Card` 使用 `rounded-3xl p-6`，`Input` 使用 `rounded-2xl h-12`，底部栏使用 `rounded-[32px]`。
- 当前 `eslint.config.js` 对 `src/pages/dialysis-processing/**` 暂时豁免了 `rounded-xl/2xl/3xl`，但全局计划正在收敛到 `rounded-md/rounded-lg`。
- 继续新增大圆角会扩大后续清理成本。

### 1.3 字段标签和输入框密度不统一
- `Input` 标签使用 `text-[11px] uppercase tracking-wide`，中文标签不适合 uppercase 风格。
- 多数字段是中文临床字段，建议使用普通中文标签、统一行高和 label 宽度，不要用英文式小写标签样式。
- 输入框内 `font-bold` 较重，临床录入页更适合中等字重，减少视觉噪音。

### 1.4 生命体征字段顺序导致读写不顺
- 当前顺序为：收缩压、舒张压、测压部位、心率、呼吸、体温。
- `测压部位` 插入在血压和心率之间，导致一行视觉信息不连续。
- 建议改为：收缩压、舒张压、心率、呼吸、体温、测压部位。

### 1.5 症状区和备注区高度占用过大
- 症状列表、添加输入、备注全部塞在第三张卡片中，第三卡明显高于其他卡。
- 备注 textarea 固定 `rows={4}`，在大屏和中屏上都会拉大纵向空间。

### 1.6 底部操作栏错行
- 当前第 277-286 行按钮区域存在缩进不齐。
- 操作栏使用深色整块 `bg-slate-900 rounded-[32px] px-8 py-6`，与上方浅色表单割裂。
- `暂存` 按钮没有事件，容易误导用户。

## 2. 改造原则

1. 保持全部业务字段不变：`weight`、`targetUf`、`sbp`、`dbp`、`pressurePoint`、`aSite`、`vSite`、`symptoms`、`notes` 等字段不可删改。
2. 保持保存 payload 不变：`handleSave` 中 `TreatmentBeforeSignsRequest` 的字段和 `symptomItems` code 不变。
3. 不改父组件接口：`patient`、`treatment`、`saving`、`treatmentLoading`、`onSave` props 不变。
4. 不引入新依赖，不使用图片资源，不使用外部 URL。
5. 优先使用 `rounded-md` / `rounded-lg`，避免新增 `rounded-xl/2xl/3xl`。
6. 中文标签不要 `uppercase`。
7. 以医护录入效率为主：字段紧凑、分组清楚、按钮固定在操作区。

## 3. 推荐目标布局

### 3.1 页面整体结构

将当前三卡 + 深色底栏改为：

```tsx
<div className="space-y-4 pb-6">
  <LoadingBanner />
  <section className="grid grid-cols-1 gap-4 2xl:grid-cols-[minmax(0,1fr)_360px]">
    <div className="space-y-4">
      <AssessmentSection title="体重与容量评估">...</AssessmentSection>
      <AssessmentSection title="生命体征">...</AssessmentSection>
      <AssessmentSection title="通路与观察">...</AssessmentSection>
    </div>
    <aside className="space-y-4">
      <PatientContextCard />
      <SymptomsCard />
      <ActionCard />
    </aside>
  </section>
</div>
```

桌面端：左侧是主要表单，右侧是患者状态、症状、保存操作。  
中小屏：全部单列，顺序为表单字段 -> 症状 -> 操作。

### 3.2 左侧主表单分组

#### 体重与容量评估

使用 4 列响应式网格：

```tsx
<div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
  透前体重
  干体重
  体重增长
  目标超滤量
</div>
```

要求：
- 每个字段同高，建议 `h-10` 或 `h-11`。
- `干体重`、`体重增长` disabled，但视觉不能过灰到不可读。
- 体重增长可根据正负值加轻量提示色：正值普通，负值 amber 或 slate；非必须。

#### 生命体征

使用 6 项均分：

```tsx
<div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
  收缩压
  舒张压
  心率
  呼吸
  体温
  测压部位
</div>
```

要求：
- 先放血压，再放心率/呼吸/体温，最后放测压部位。
- 避免 `测压部位` 插入血压组中间。
- 单位 suffix 宽度固定，例如 `min-w-10 text-right`，避免 `mmHg` 和 `℃` 造成输入框内文本错位。

#### 通路与观察

使用 4 项均分：

```tsx
<div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4">
  A 端位点
  V 端位点
  神志状态
  护理分级
</div>
```

要求：
- 只放短字段，不把症状和备注塞进同一张主卡。
- 保持与体重卡片视觉宽度一致。

### 3.3 右侧辅助区

#### 患者上下文卡

展示只读信息：
- 患者状态：`patient.status`
- 透析方案：`patient.treatmentPlan`
- 干体重：`patient.dryWeight` kg
- 当前床位：`patient.bedId`

用途：医护录入时不用回头看页面顶部。

#### 症状记录卡

从主卡拆出：
- 症状 chip 列表
- 添加症状输入框
- 添加按钮

要求：
- 空状态显示“暂无透前症状”。
- chip 使用 `rounded-md`，高度统一。
- 删除按钮需要 `aria-label={`删除症状 ${item}`}`。
- 添加按钮文字建议改为“添加症状”，不要“添加真实症状”，减少文案压力。

#### 备注与操作卡

建议把备注与保存按钮放在同一卡内：
- 备注 textarea `min-h-24`，不要用过大的固定行数。
- 按钮右对齐或全宽：中小屏全宽，大屏右对齐。
- 删除无事件的“暂存”按钮，除非同步实现暂存逻辑。若保留，必须禁用并标注“暂存（待接入）”。
- 主按钮文案保持：`提交透前评估` / `保存中...` / `治疗加载中...`。

## 4. 组件级重构建议

### 4.1 替换 `Card` 为 `AssessmentSection`

目标：减少圆角和内边距，统一标题栏。

```tsx
function AssessmentSection({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <div className="mb-4 flex items-center gap-2 border-b border-slate-100 pb-3">
        {icon}
        <h3 className="text-base font-bold text-slate-800">{title}</h3>
      </div>
      {children}
    </section>
  )
}
```

### 4.2 替换 `Input` 为 `FieldInput`

目标：统一中文 label、输入框高度、suffix 对齐。

```tsx
function FieldInput({ label, value, suffix, onChange, disabled }: Props) {
  return (
    <label className="block min-w-0">
      <span className="mb-1.5 block text-sm font-medium text-slate-600">{label}</span>
      <span className="flex h-10 items-center rounded-md border border-slate-200 bg-white px-3 focus-within:border-blue-400 focus-within:ring-2 focus-within:ring-blue-100">
        <input className="min-w-0 flex-1 bg-transparent text-sm font-medium text-slate-800 outline-none disabled:text-slate-500" />
        {suffix ? <span className="ml-2 min-w-10 text-right text-xs font-medium text-slate-400">{suffix}</span> : null}
      </span>
    </label>
  )
}
```

注意：
- 不要新增表单库。
- 不要使用 `text-[11px]`、`rounded-2xl`。
- `input` 仍使用受控 `value` 和 `onChange`。

## 5. 分步执行清单

### Step 1：只改结构，不改保存逻辑

文件：`PreAssessment.tsx`

- [ ] 保留 `EMPTY_FORM`、`mapTreatmentToForm`、`parseOptionalNumber`、`handleSave` 原逻辑。
- [ ] 将 `Card` 改名或替换为 `AssessmentSection`。
- [ ] 将 `Input` 改名或替换为 `FieldInput`。
- [ ] 删除 `rounded-3xl`、`rounded-2xl`、`rounded-[32px]`、`text-[11px]` 的新增/现有使用。

### Step 2：重排主表单

- [ ] 外层从 `xl:grid-cols-4` 改为 `2xl:grid-cols-[minmax(0,1fr)_360px]`。
- [ ] 体重卡改为 4 项横向均分。
- [ ] 生命体征卡改为 3 列 / 2 行，顺序为：收缩压、舒张压、心率、呼吸、体温、测压部位。
- [ ] 通路与观察卡只包含 A端、V端、神志、护理分级。

### Step 3：拆出右侧辅助区

- [ ] 新增患者上下文卡，显示患者状态、透析方案、干体重、床位。
- [ ] 症状记录从原第三张卡拆到右侧。
- [ ] 备注和提交按钮放入右侧底部操作卡。
- [ ] 删除无业务事件的“暂存”按钮，或禁用显示“暂存（待接入）”。推荐删除。

### Step 4：处理加载与禁用状态

- [ ] `treatmentLoading` 提示从大圆角蓝块改为 `rounded-lg border border-blue-100 bg-blue-50 px-4 py-3`。
- [ ] 提交按钮 disabled 条件保持 `saving || treatmentLoading`。
- [ ] 加载中不能误显示旧患者数据，这一点继续依赖现有 `useEffect` 清空逻辑，不要删除。

### Step 5：移动端检查

- [ ] 宽度 1366：左主表单 + 右辅助栏不挤压。
- [ ] 宽度 1024：仍可保持左右布局或自然变单列，不能横向滚动。
- [ ] 宽度 768：单列，字段间距一致。
- [ ] 宽度 375：输入框 suffix 不挤压输入内容，按钮全宽或可点击区域足够。

## 6. 验收标准

### 6.1 视觉验收

- [ ] 页面没有明显右侧空白或三卡错位。
- [ ] 体重、生命体征、通路观察三组宽度一致、间距一致。
- [ ] 字段 label、输入框、suffix 横向对齐。
- [ ] 症状和备注不再把主表单撑得过高。
- [ ] 底部操作区不再深色割裂，按钮不缩进错行。
- [ ] 全页面中文文案自然、无英文 uppercase 标签风格。

### 6.2 功能验收

- [ ] 切换患者后，透前评估不会显示上一患者旧数据。
- [ ] 编辑所有字段后点击提交，请求 payload 与改造前一致。
- [ ] 症状添加、删除可用。
- [ ] `saving` 和 `treatmentLoading` 时提交按钮禁用。
- [ ] 无今日治疗时仍可通过保存流程触发 `ensureTodayTreatment` 创建治疗记录。

### 6.3 工程验收

必须执行：

```bash
cd ai-hms-frontend && npm run lint
cd ai-hms-frontend && npm run build
```

补充检查：

```bash
rg "rounded-(2xl|3xl)|rounded-\[32px\]|text-\[11px\]" ai-hms-frontend/src/pages/dialysis-processing/execution/PreAssessment.tsx
```

期望：不再命中上述旧样式。

## 7. 不要做的事

- 不要改 `TreatmentBeforeSignsRequest` 类型。
- 不要改 `restApi.saveTreatmentBeforeSigns`。
- 不要把字段名改成新接口字段。
- 不要把症状改成固定字典，当前仍是自由输入 chip。
- 不要引入 AntD Form、Formik、React Hook Form 等新表单方案。
- 不要同时改 `PostAssessment`、`TodayPrescription` 等其他 Tab，除非用户另行要求。
- 不要提交 `dist/`。

## 8. 建议交付说明模板

执行完成后回复用户时使用：

```md
已按计划优化 `PreAssessment.tsx`：
- 主表单改为左侧三组均衡布局，右侧为患者上下文、症状、备注与提交操作。
- 统一字段高度、label 样式、suffix 对齐和间距。
- 删除无事件暂存按钮，保留提交保存逻辑。
- 验证：`npm run lint`、`npm run build` 通过。
```
