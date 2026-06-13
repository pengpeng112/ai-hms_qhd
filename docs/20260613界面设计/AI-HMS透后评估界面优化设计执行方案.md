# AI-HMS 透后评估界面优化设计执行方案

适用页面：`ai-hms-frontend/src/pages/dialysis-processing/execution/PostAssessment.tsx`  
上层页面：`ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`  
目标：不修改接口、不修改字段、不改变透后评估暂存和提交逻辑，只优化当前“透后评估”界面的样式、分组、流程和临床结项体验。

---

## 1. 当前代码结构核查

当前 `PostAssessment.tsx` 已经具备完整的透后评估能力：

1. 页面接收：
   - `patient`
   - `treatment`
   - `treatmentLoading`
   - `onSave`
   - `onSubmit`

2. 表单字段包括：
   - startTime
   - endTime
   - realUfVolume
   - realSubstituteVolume
   - weight
   - extraWeight
   - lossWeight
   - postNetWeight
   - sbp
   - dbp
   - heartRate
   - respiration
   - temperature
   - realIntake
   - pressurePoint
   - complication
   - hasDialysisEvent
   - dialyzerCoag
   - lineACoag
   - lineVCoag
   - symptoms
   - notes
   - fistulaCareGuidance

3. 现有逻辑包括：
   - 从 `treatment` 回填透后体征和症状项
   - 从 `treatment.endBp` 拆分血压
   - 自动计算透后净重
   - 自动计算体重丢失
   - 暂存报告
   - 提交结项

4. 上层 `DialysisExecution.tsx` 中 `POST_ASSESSMENT` 会渲染 `PostAssessment`，并传入 `onSave={handleSavePostAssessment}`、`onSubmit={handleSubmitPostAssessment}`。

因此本次只做前端 UI 和展示层级优化，不改 API。

---

## 2. 当前界面主要问题

从截图和代码看，当前页面已经能用，但还可以继续优化：

### 2.1 页面像表单堆叠，不像“下机结项流程”

当前顺序是：

```text
治疗时间 / 实际超滤量
体重与生命体征
临床观察与记录
底部按钮
```

功能完整，但视觉上还是一张长表单。透后评估其实是“下机后结项核查”，需要更突出流程：

```text
下机时间和超滤量 → 体重生命体征 → 凝血/事件/内瘘 → 护理交接 → 提交结项
```

---

### 2.2 治疗时间和超滤量信息没有形成“结项摘要”

截图中治疗时间、实际超滤量、实际置换液量放在一个普通卡片内，但这三个信息对结项非常关键，建议作为“下机结项摘要条”突出。

---

### 2.3 体重和生命体征字段较多，分组还可以更明确

当前“体重与生命体征”里字段较多：

```text
透后体重
透后净重
体重丢失
额外体重
透后血压
测压部位
透后心率
透后呼吸
透后体温
实际摄入
血压实际测量时间
```

建议拆成：

```text
体重与容量
透后血压与体征
```

这样阅读更快。

---

### 2.4 临床观察与记录右侧区域略散

当前右侧包含：

```text
内瘘情况
是否进行内瘘/导管护理健康指导
其他意外情况提示
```

这些和护理交接强相关，建议与透析事件、透后备注一起形成“事件、内瘘与交接记录”。

---

### 2.5 提交前缺项提示不够明确

当前底部只显示：

```text
评估时间
下机护士
患者姓名
暂存报告
提交结项
```

建议增加必填缺项提示，例如：

```text
必填缺项：透后心率、透后呼吸、透后体温
```

---

## 3. 推荐新版布局

建议改成：

```text
透后评估
├── 顶部患者摘要
│   ├── 患者信息
│   ├── 干体重
│   └── 治疗方案
│
├── 暂无治疗记录提示
│
├── 下机结项摘要条
│   ├── 治疗时间
│   ├── 实际超滤量
│   ├── 实际置换液量
│   └── 结项状态
│
├── 主体区
│   ├── 左侧：体重与生命体征
│   │   ├── 体重与容量
│   │   └── 透后血压与体征
│   └── 右侧：临床观察与风险
│       ├── 透析器凝血分级
│       ├── 血路管A端凝血分级
│       └── 血路管V端凝血分级
│
├── 事件、内瘘与交接记录
│   ├── 是否发生透析事件
│   ├── 透后备注
│   ├── 内瘘情况
│   ├── 内瘘/导管护理健康指导
│   └── 其他意外情况提示
│
└── 底部固定操作条
    ├── 评估时间
    ├── 下机护士
    ├── 当前患者
    ├── 必填缺项
    ├── 暂存报告
    └── 提交结项
```

---

## 4. 具体优化内容

### 4.1 顶部患者摘要保留但压缩

当前上层已经有患者摘要，所以 `PostAssessment` 内部不建议再重复太多信息。保留：

```text
支英俊  未排床
ID: 300410   女 / 72岁   费用类别: 市职工普通   透龄: 待补充
干体重 0kg
治疗方案 HD
```

建议高度控制在 100–110px。

---

### 4.2 下机结项摘要条

将治疗时间、实际超滤量、实际置换液量统一成一个紧凑摘要条：

```text
治疗时间：-- ~ 2026-06-13T21:05
结束时间建议以超滤量稳定后自动判断

实际超滤量：取超滤量最大值
实际置换液量：取置换液量最大值
结项状态：待暂存 / 待提交
```

这样用户一眼能看到结项关键指标。

---

### 4.3 体重与生命体征卡

拆成两个子组：

#### 体重与容量

```text
透后体重
额外体重
透后净重
体重丢失
实际摄入
```

其中：

```text
透后净重 = 透后体重 - 额外体重
体重丢失 = 透前体重 - 透后净重
```

自动计算逻辑继续保留，不改字段。

#### 透后血压与体征

```text
透后血压
测压部位
透后心率
透后呼吸
透后体温
血压测量时间
```

透后心率、呼吸、体温继续保留必填标记。

---

### 4.4 临床观察与风险卡

将凝血分级放到右侧独立风险卡中：

```text
临床观察与风险
无负面事件 / 已记录事件

透析器凝血分级：0级 1级 2级 3级
血路管A端凝血分级：0级 1级 2级 3级
血路管V端凝血分级：0级 1级 2级 3级
```

优势：

1. 凝血风险独立显示，更像核查项。
2. 和体重体征区分开。
3. 页面更清晰。

---

### 4.5 事件、内瘘与交接记录

将原来的事件、备注、内瘘情况、健康指导集中到一个卡片中：

```text
事件、内瘘与交接记录
发生透析事件：否 / 是
透后备注
内瘘情况
是否进行内瘘/导管护理健康指导
其他意外情况提示
```

如果 `hasDialysisEvent = true`，展开事件说明输入框。

---

### 4.6 底部固定操作条

保留现有按钮：

```text
暂存报告
提交结项
```

增强左侧状态：

```text
评估时间：21:05
下机护士：test_admin
当前患者：支英俊
必填缺项：透后心率、透后呼吸、透后体温
```

建议新增派生变量：

```ts
const missingRequiredFields = [
  !form.heartRate.trim() ? '透后心率' : '',
  !form.respiration.trim() ? '透后呼吸' : '',
  !form.temperature.trim() ? '透后体温' : '',
].filter(Boolean)
```

第一版只提示，不阻断提交；后续可考虑提交前校验。

---

## 5. 建议代码调整范围

只改：

```text
ai-hms-frontend/src/pages/dialysis-processing/execution/PostAssessment.tsx
```

不改：

```text
TreatmentAfterSignsRequest
onSave
onSubmit
buildPayload
mapTreatmentToForm
parseOptionalNumber
toIsoOrUndefined
restApi
后端接口
字段名
```

---

## 6. 可复用组件优化建议

当前已有：

```text
Field
CoagSelect
```

建议保留并增强。

### 6.1 Field 增加 compact 样式

当前 Field 高度和边距较稳定，可以保留。只建议在布局层减少卡片高度，不强制改 Field 组件。

### 6.2 CoagSelect 视觉增强

当前 `CoagSelect` 已经支持 0级/1级/2级/3级按钮选择，建议保持逻辑不变，只让它放在“临床观察与风险”右侧卡片里。

### 6.3 新增 MissingRequired 提示

```tsx
const missingRequiredFields = useMemo(() => [
  !form.heartRate.trim() ? '透后心率' : '',
  !form.respiration.trim() ? '透后呼吸' : '',
  !form.temperature.trim() ? '透后体温' : '',
].filter(Boolean), [form.heartRate, form.respiration, form.temperature])
```

展示：

```tsx
{missingRequiredFields.length > 0 ? (
  <span className="font-bold text-rose-500">
    必填缺项：{missingRequiredFields.join('、')}
  </span>
) : (
  <span className="font-bold text-emerald-600">必填项已完成</span>
)}
```

---

## 7. 实施步骤

### Commit 1：透后评估顶部摘要优化

```text
fix: polish post assessment header and finish summary
```

内容：

1. 顶部患者摘要压缩。
2. 增加“下机结项摘要条”。
3. 治疗时间、实际超滤量、实际置换液量集中展示。
4. 增加结项状态提示。

---

### Commit 2：体重与生命体征分组优化

```text
fix: regroup post assessment weight and vital signs
```

内容：

1. 体重与生命体征卡拆为：
   - 体重与容量
   - 透后血压与体征
2. 字段不变。
3. 自动计算逻辑不变。
4. 必填项标识保留。

---

### Commit 3：临床观察与风险优化

```text
fix: improve post assessment clinical observation layout
```

内容：

1. 凝血分级移入右侧“临床观察与风险”卡。
2. 事件、内瘘、备注、健康指导集中到一个交接记录卡。
3. `hasDialysisEvent` 展开事件说明逻辑不变。

---

### Commit 4：底部操作与缺项提示优化

```text
fix: add post assessment required field hints and action bar polish
```

内容：

1. 新增 `missingRequiredFields` 派生变量。
2. 底部显示必填缺项。
3. 暂存报告和提交结项按钮样式优化。
4. 保存和提交逻辑不变。

---

## 8. 验收清单

1. 透后评估页面正常打开。
2. 患者切换后表单正常重置。
3. treatment 回填逻辑正常。
4. 透后净重自动计算正常。
5. 体重丢失自动计算正常。
6. 暂存报告正常。
7. 提交结项正常。
8. 透后心率、呼吸、体温必填标识正常。
9. 凝血分级选择正常。
10. 发生透析事件切换正常。
11. 事件说明输入正常。
12. 内瘘情况输入正常。
13. 内瘘/导管护理健康指导勾选正常。
14. buildPayload 输出字段不变。
15. 接口和字段未修改。
16. 控制台无 JS 报错。
17. 1366x768 下首屏能看到核心结项内容。

---

## 9. 给开发 AI 的执行提示词

```markdown
请优化 AI-HMS 智能透析系统“透析执行 / 透后评估”页面。页面文件为 `ai-hms-frontend/src/pages/dialysis-processing/execution/PostAssessment.tsx`。请保持接口、字段、暂存和提交逻辑不变，只优化布局、样式和展现方式。

重要要求：
1. 不修改后端接口。
2. 不修改字段名。
3. 不修改 `TreatmentAfterSignsRequest`。
4. 不修改 `buildPayload` 字段结构。
5. 不修改 `mapTreatmentToForm` 回填逻辑。
6. 不修改 `onSave`、`onSubmit` 调用逻辑。
7. 不改变透后净重和体重丢失自动计算逻辑。
8. 不改变 `hasDialysisEvent` 和事件说明展示逻辑。
9. 不改变凝血分级选择逻辑。

当前页面问题：
1. 页面像普通纵向表单，不像下机结项流程。
2. 治疗时间、实际超滤量、实际置换液量没有形成明显结项摘要。
3. 体重和生命体征字段较多，分组不够清楚。
4. 凝血分级、事件、内瘘情况、备注和健康指导之间层级不够明确。
5. 底部提交前没有必填缺项提示。

优化目标：
1. 顶部患者摘要压缩展示。
2. 增加“下机结项摘要条”，集中展示治疗时间、实际超滤量、实际置换液量和结项状态。
3. 体重与生命体征拆成：
   - 体重与容量
   - 透后血压与体征
4. 凝血分级放入右侧“临床观察与风险”卡。
5. 事件、内瘘与交接记录集中到一个卡片。
6. 底部固定操作条显示：
   - 评估时间
   - 下机护士
   - 当前患者
   - 必填缺项
   - 暂存报告
   - 提交结项
7. 新增 `missingRequiredFields` 派生变量，只做提示，不强制阻断提交。
8. 保持所有输入字段、保存、提交、自动计算逻辑不变。

验收：
1. 页面加载正常。
2. 患者切换后表单重置正常。
3. 字段回填正常。
4. 暂存报告正常。
5. 提交结项正常。
6. 自动计算正常。
7. 凝血分级选择正常。
8. 发生透析事件切换和备注输入正常。
9. 接口和字段未修改。
10. 控制台无 JS 错误。
```
