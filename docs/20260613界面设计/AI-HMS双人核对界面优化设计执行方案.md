# AI-HMS 双人核对界面优化设计执行方案

适用页面：`ai-hms-frontend/src/pages/dialysis-processing/execution/Verification.tsx`  
上层页面：`ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`  
目标：不修改接口、不修改字段、不改变首次核对、二次核对、消毒登记和提交逻辑，只优化当前“双人核对”界面的样式、层级、核对流程和操作体验。

---

## 1. 当前代码结构核查

当前 `Verification.tsx` 已经包含完整的业务能力：

1. 页面接收：
   - `patient`
   - `treatment`
   - `treatmentLoading`
   - `onSaveFirstCheck`
   - `onSaveSecondCheck`
   - `onSaveDisinfection`

2. 首次核对表单包括：
   - operatorId
   - operateTime
   - materials
   - param
   - vascular
   - pipeline

3. 二次核对表单包括：
   - operatorId
   - recheckNurseId
   - qcNurseId
   - operateTime
   - dialysisMode
   - prescription
   - anticoagulant
   - vascular
   - lineConnection

4. 机器消毒登记固定提交：
   - type: 机器
   - disinfectant: 500mg/L含氯消毒液
   - startTime
   - disinfectUserId

5. 当前已有：
   - 保存首次核对
   - 保存二次核对
   - 保存全部核对
   - 保存消毒登记
   - 人员配置与角色登记

因此本次不需要改 API，只调整前端布局和样式。

---

## 2. 当前界面主要问题

从截图看，当前页面已经可用，但还有几个体验问题：

### 2.1 三张大卡横排，视觉重量接近

当前首次核对、二次核对、机器消毒登记三个模块并排，每个都很重。用户不容易快速判断：

```text
先做什么？
哪个已经完成？
哪个还缺？
最终在哪里提交？
```

### 2.2 核对流程不够直观

双人核对实际是一个流程：

```text
首次核对 → 二次核对 → 机器消毒登记 → 人员确认 → 提交生效
```

当前页面没有流程进度提示，导致三块内容看起来是并列关系，而不是流程关系。

### 2.3 右侧机器消毒登记空白偏多

截图中机器消毒登记卡片高度很高，但实际字段较少，导致右侧有明显空白。

### 2.4 人员配置在底部弱化

“人员配置与角色登记”其实很重要，但目前放在底部小卡片中，和核对流程关联不强。

### 2.5 提交前缺项不够明确

底部有“暂存修改 / 提交并生效”，但没有明显提示哪些角色未选、哪些核对项异常、消毒是否登记。

---

## 3. 推荐新版布局

建议改为：

```text
双人核对
├── 顶部患者与流程摘要
│   ├── 患者简要信息
│   ├── 核对流程：首次核对 → 二次核对 → 消毒登记
│   └── 干体重 / 治疗方案
│
├── 暂无治疗记录提示
│
├── 主核对区
│   ├── 首次核对
│   ├── 二次核对
│   └── 右侧：机器消毒登记
│
├── 人员配置与角色登记
│
├── 提交前检查提示
└── 底部操作
```

---

## 4. 具体优化内容

### 4.1 顶部增加“核对流程条”

在患者摘要区域增加流程提示：

```text
核对流程
1 首次核对 → 2 二次核对 → 3 消毒登记
```

目的：

1. 用户知道这是一个顺序流程。
2. 当前页不是三个独立卡片，而是完整核对闭环。
3. 后续可以根据完成状态显示绿色/蓝色/灰色。

第一版只做静态展示，不改业务逻辑。

---

### 4.2 患者信息保留但压缩

保留：

```text
支英俊  未排床
ID: 300410   女 / 72岁   治疗方案: HD
干体重 0kg
治疗方案 HD
```

不建议在当前页面重复过多基础信息，因为上层已有患者摘要。

---

### 4.3 首次核对卡优化

首次核对卡展示：

```text
首次核对
默认取病区责任护士，重点核对身份、耗材、参数和管路安全。

核对人
核对时间

耗材规格（透析器/血路管）  正常 / 异常
处方参数（透析方式/处方内容） 正常 / 异常
血管通路与患者身份          正常 / 异常
管路连接与预冲              正常 / 异常

[确认并完成首次核对]
```

优化点：

1. 说明文案更短。
2. 核对项行高统一。
3. “正常/异常”按钮靠右。
4. 异常时展开原因输入框，保持原逻辑。
5. 按钮蓝色，表示首次核对主操作。

---

### 4.4 二次核对卡优化

二次核对卡展示：

```text
二次核对
需独立复核，不可与首次核对人相同，重点核对处方调整与批号一致性。

核对人
核对时间

透析模式      正常 / 异常
处方内容      正常 / 异常
抗凝剂        正常 / 异常
血管通路      正常 / 异常
管路连接      正常 / 异常

[确认并完成二次核对]
```

优化点：

1. 二次核对用橙色强调。
2. 和首次核对形成视觉区分。
3. 保留过滤“不可与首次核对人相同”的逻辑。

---

### 4.5 机器消毒登记改为右侧安全卡

消毒登记字段少，可以放在右侧安全卡内：

```text
机器消毒登记
固定类型与消毒液，仅登记时间和登记人。

消毒类型：机器
消毒液：500mg/L含氯消毒液
消毒时间：年/月/日 --:--
登记人：test_admin

[确认并保存消毒登记]
```

优化点：

1. 用浅绿色安全底色。
2. 卡片高度与核对卡一致。
3. 减少大面积空白。
4. 更突出“消毒登记已完成/待完成”。

---

### 4.6 人员配置与角色登记优化

当前 5 个角色卡保留，但建议作为一个清晰横向区域：

```text
人员配置与角色登记
系统根据排班与核对记录自动建议人员，提交前做最后确认。

预冲护士     test_admin
穿刺/注射    test_admin
上机护士     test_admin
质控护士     请选择
质检医生     请选择
```

建议：

1. 角色卡高度统一。
2. 未选择显示灰色“请选择”。
3. 已选择人员加粗。
4. 如果未选，可在提交前检查中提示。

---

### 4.7 新增提交前检查提示

在底部提交前增加一个提示条：

```text
提交前检查：
首次核对 4/4 正常，二次核对 5/5 正常，机器消毒已登记；
质控护士和质检医生尚未选择。

若出现“异常”，需填写异常原因后再提交生效。
```

这个提示条不一定第一版就做真实计算，可以先基于现有表单状态前端计算。

建议新增派生变量：

```ts
const firstNormalCount = [
  firstForm.materials,
  firstForm.param,
  firstForm.vascular,
  firstForm.pipeline,
].filter((item) => item.result).length

const secondNormalCount = [
  secondForm.dialysisMode,
  secondForm.prescription,
  secondForm.anticoagulant,
  secondForm.vascular,
  secondForm.lineConnection,
].filter((item) => item.result).length

const hasFirstMistakeWithoutReason = [
  firstForm.materials,
  firstForm.param,
  firstForm.vascular,
  firstForm.pipeline,
].some((item) => !item.result && !item.mistake.trim())

const hasSecondMistakeWithoutReason = [
  secondForm.dialysisMode,
  secondForm.prescription,
  secondForm.anticoagulant,
  secondForm.vascular,
  secondForm.lineConnection,
].some((item) => !item.result && !item.mistake.trim())
```

第一版可以只展示，不强制阻断提交；第二版再考虑提交校验。

---

## 5. 建议代码调整范围

只改：

```text
ai-hms-frontend/src/pages/dialysis-processing/execution/Verification.tsx
```

不改：

```text
restApi
TreatmentFirstCheckRequest
TreatmentSecondCheckRequest
TreatmentDisinfectionRequest
onSaveFirstCheck
onSaveSecondCheck
onSaveDisinfection
后端接口
字段名
```

---

## 6. 可复用组件建议

当前已有：

```text
StaffSelect
DateTimeInput
CheckResultRow
RoleCard
```

建议保留并轻量增强。

### 6.1 CheckResultRow 增加 compact / tone

```tsx
function CheckResultRow({
  label,
  value,
  onChange,
  tone = 'blue',
}: {
  label: string
  value: CheckItemState
  onChange: (value: CheckItemState) => void
  tone?: 'blue' | 'orange'
})
```

首次核对用 blue，二次核对用 orange。

### 6.2 新增 WorkflowStep

```tsx
function WorkflowStep({ index, title, tone }: { index: string; title: string; tone: string }) {
  return (
    <div className="flex items-center gap-2">
      <span className="flex h-7 w-7 items-center justify-center rounded-full bg-blue-600 text-xs font-black text-white">
        {index}
      </span>
      <span className="text-sm font-bold text-slate-800">{title}</span>
    </div>
  )
}
```

### 6.3 RoleCard 强化未选择状态

当前 `RoleCard` 直接显示“请选择”，建议区分颜色：

```tsx
<div className={value ? 'text-slate-900' : 'text-slate-400'}>
  {value || '请选择'}
</div>
```

---

## 7. 实施步骤

### Commit 1：顶部流程摘要优化

```text
fix: improve verification patient header and workflow status
```

内容：

1. 顶部患者信息压缩。
2. 增加核对流程条。
3. 干体重、治疗方案小卡保留。
4. 暂无治疗记录提示缩小。

---

### Commit 2：首次与二次核对卡优化

```text
fix: polish first and second verification panels
```

内容：

1. 首次核对卡更紧凑。
2. 二次核对卡使用橙色强调。
3. 核对项行高统一。
4. 保留异常原因输入逻辑。
5. 保存按钮保留原逻辑。

---

### Commit 3：消毒登记与人员角色优化

```text
fix: refine disinfection and role registration layout
```

内容：

1. 机器消毒登记改为右侧安全卡。
2. 消毒信息减少空白。
3. 人员配置横向区域优化。
4. 未选择角色显示灰色。

---

### Commit 4：提交前检查与底部操作优化

```text
fix: add verification pre-submit summary
```

内容：

1. 增加提交前检查提示条。
2. 统计首次核对正常数量。
3. 统计二次核对正常数量。
4. 提示异常项未填写原因。
5. 暂存修改 / 提交并生效按钮样式优化。

---

## 8. 验收清单

1. 双人核对页面正常打开。
2. 患者切换后表单正常重置。
3. 工作人员列表正常加载。
4. 首次核对保存正常。
5. 二次核对保存正常。
6. 批量提交保存正常。
7. 机器消毒登记保存正常。
8. 首次核对人默认取当前用户逻辑不变。
9. 二次核对人仍过滤首次核对人。
10. 异常原因输入框逻辑不变。
11. 消毒类型和消毒液仍保持固定值。
12. 人员配置显示正常。
13. 接口和字段未修改。
14. 控制台无 JS 报错。
15. 1366x768 下视觉更清晰，首屏能看到主要核对内容。

---

## 9. 给开发 AI 的执行提示词

```markdown
请优化 AI-HMS 智能透析系统“透析执行 / 双人核对”页面。页面文件为 `ai-hms-frontend/src/pages/dialysis-processing/execution/Verification.tsx`。请保持接口、字段和保存逻辑不变，只优化布局、样式和展现方式。

重要要求：
1. 不修改后端接口。
2. 不修改字段名。
3. 不修改 `TreatmentFirstCheckRequest`、`TreatmentSecondCheckRequest`、`TreatmentDisinfectionRequest`。
4. 不修改 `onSaveFirstCheck`、`onSaveSecondCheck`、`onSaveDisinfection` 的调用逻辑。
5. 不改变首次核对、二次核对、保存全部、机器消毒登记的业务流程。
6. 保留二次核对人不能与首次核对人相同的过滤逻辑。
7. 保留异常时填写异常原因的逻辑。

当前问题：
1. 首次核对、二次核对、消毒登记三张大卡并列，流程不直观。
2. 机器消毒登记字段少但卡片很高，右侧空白较多。
3. 人员配置与角色登记较弱，和提交关系不明显。
4. 底部提交前缺少核对完成度和缺项提示。
5. 页面看起来能用，但还不像一个“核对流程工作台”。

优化目标：
1. 顶部增加核对流程条：首次核对 → 二次核对 → 消毒登记。
2. 患者信息压缩展示，突出 ID、性别年龄、治疗方案。
3. 首次核对卡保留蓝色主色，内容更紧凑。
4. 二次核对卡使用橙色强调，说明“不可与首次核对人相同”。
5. 机器消毒登记改为右侧绿色安全卡，字段紧凑展示。
6. 人员配置与角色登记横向展示，未选显示灰色“请选择”。
7. 增加提交前检查提示条，展示首次核对正常数量、二次核对正常数量、消毒登记状态、未选人员提示。
8. 暂存修改 / 提交并生效按钮样式强化。
9. 1366x768 下首屏更清晰，减少杂乱和空白。

建议新增派生变量：
- firstNormalCount
- secondNormalCount
- hasFirstMistakeWithoutReason
- hasSecondMistakeWithoutReason

验收：
1. 首次核对保存正常。
2. 二次核对保存正常。
3. 保存全部正常。
4. 机器消毒登记保存正常。
5. 人员列表加载正常。
6. 患者切换后数据正常重置。
7. 异常原因输入逻辑正常。
8. 接口和字段未修改。
9. 控制台无 JS 错误。
```
