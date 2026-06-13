# AI-HMS 透析小结界面优化设计执行方案

适用页面：`ai-hms-frontend/src/pages/dialysis-processing/execution/DialysisSummary.tsx`  
上层页面：`ai-hms-frontend/src/pages/dialysis-processing/DialysisExecution.tsx`  
目标：不修改接口、不修改字段、不改变小结保存、自动汇总、医嘱列表加载逻辑，只优化“透析小结”界面的布局、层级、书写体验和结项核查效率。

---

## 1. 当前代码结构核查

当前 `DialysisSummary.tsx` 已经具备完整的小结展示和保存能力：

1. 页面接收：
   - `patient`
   - `treatment`
   - `treatmentLoading`
   - `onTreatmentUpdated`

2. 页面状态包括：
   - `orders`
   - `doctorSummary`
   - `treatmentSummary`
   - `saving`
   - `lastTreatmentId`

3. 当前会加载：
   - `orderApi.list(patient.id, { includeExpired: false })`

4. 当前保存逻辑：
   - `restApi.updateTreatmentSummary(treatment.id, { doctorSummary, treatmentSummary })`
   - 保存后调用 `onTreatmentUpdated(updated.data)`

5. 当前已有展示内容：
   - 容量负荷评估
   - 关键生命体征趋势
   - 质控与安全汇总
   - 医生小结
   - 治疗/护理小结
   - 自动汇总参考
   - 透前/透后摘要
   - 双人核对摘要
   - 凝血与通路摘要
   - 本次执行医嘱明细汇总
   - 最新治疗时间提示

因此本次只做前端 UI 和信息层级优化，不改 API。

---

## 2. 当前界面主要问题

### 2.1 顶部三张卡片信息分散

当前顶部是：

```text
容量负荷评估
关键生命体征趋势
质控与安全汇总
```

这三张卡片视觉比较重，但用户进入“透析小结”时，真正需要的是先判断：

```text
容量是否达标
生命体征有没有趋势/异常
安全质控是否完整
医嘱执行是否完成
```

建议把这部分改为更紧凑的“结项总览”。

---

### 2.2 小结书写区缺少辅助动作

当前医生小结、治疗/护理小结两个文本框可用，但“自动汇总参考”在下方，用户需要先看参考再手动复制。建议把自动汇总放到右侧，形成：

```text
左：小结书写
右：自动汇总助手
```

这样更符合书写路径。

---

### 2.3 自动汇总参考应该更像“助手”

当前自动汇总是普通段落。建议改为：

```text
自动汇总参考
复制容量和超滤要点
补充生命体征、通路、并发症和医嘱执行
```

第一版可以只做展示，不一定增加复制逻辑；如果实现简单，可增加“复制自动汇总”按钮。

---

### 2.4 下方三块摘要重复占高

当前下方还有：

```text
透前/透后摘要
双人核对摘要
凝血与通路摘要
```

这些信息有价值，但不需要用大卡片占太多高度。建议改为紧凑核查卡，行高缩小。

---

### 2.5 医嘱明细表偏低、行高偏大

当前医嘱表格是完整表格，字段合理，但行高较高。建议压缩为：

```text
类型
项目名称
执行编码
执行时间
执行人
状态
```

保留字段，降低行高，让它作为“小结依据明细”而不是页面主体。

---

## 3. 推荐新版布局

```text
透析小结
├── 顶部患者摘要
│
├── 结项总览
│   ├── 容量负荷
│   ├── 生命体征
│   ├── 安全质控
│   └── 医嘱执行
│
├── 小结书写区 + 自动汇总助手
│   ├── 医生小结
│   ├── 治疗/护理小结
│   └── 自动汇总参考
│
├── 结项核查摘要
│   ├── 透前/透后摘要
│   ├── 双人核对摘要
│   └── 凝血与通路摘要
│
├── 本次执行医嘱明细
│
└── 底部固定操作条
    ├── 最新治疗时间
    ├── 透中监测点
    ├── 最新备注
    ├── 缺项提示
    ├── 复制自动汇总
    └── 保存小结
```

---

## 4. 具体优化内容

### 4.1 结项总览

把顶部三张大卡改成一张“结项总览”卡，里面 4 个小卡片：

```text
容量负荷
--kg / --L
透前/透后净重、实际超滤量

生命体征
暂无趋势
透中监测点与最新血压心率

安全质控
凝血未记录
凝血、通路、并发症

医嘱执行
1 项已执行
本次医嘱执行明细
```

字段都来自现有 `treatment`、`monitoringRows`、`orders`、`afterSymptomItems`。

---

### 4.2 小结书写区

小结书写区左侧放两个文本框：

```text
医生小结
治疗/护理小结
```

仍然对应：

```ts
doctorSummary
treatmentSummary
```

保存逻辑不变：

```ts
restApi.updateTreatmentSummary(treatment.id, {
  doctorSummary: doctorSummary.trim(),
  treatmentSummary: treatmentSummary.trim(),
})
```

---

### 4.3 自动汇总助手

右侧展示 `autoSummary`，并增加说明：

```text
自动汇总参考
患者于 -- 开始 HD 治疗...
```

建议增加两个辅助提示：

```text
1 复制容量和超滤要点到小结
2 补充生命体征、通路、并发症和医嘱执行
```

可选增加按钮：

```text
复制自动汇总
```

如果开发觉得会改变交互，可以先只做静态按钮或保留到底部。

---

### 4.4 结项核查摘要

将原来的三张大卡压缩为三张小卡：

```text
透前/透后摘要
透前血压
透后血压
透后情况

双人核对摘要
首次核对
二次核对

凝血与通路摘要
透析器
A端 / V端
血管通路
并发症
```

注意：当前代码里血管通路还是 `未记录`，不建议编造数据。保持原逻辑，后续有字段再补。

---

### 4.5 医嘱执行明细

保留当前表格字段：

```text
类型
项目名称
执行编码
执行时间
执行人
状态
```

优化：

1. 行高压缩。
2. 表头 sticky 可选。
3. 状态用绿色 badge 或图标。
4. 无数据时显示：
   `暂无本次执行医嘱`
5. 仍然使用 `orders.length`。

---

### 4.6 底部固定操作条

建议新增底部操作条，保留保存按钮：

```text
最新治疗时间：--
透中监测点：0 个
最新备注：--
缺项：小结未保存
[复制自动汇总] [保存小结]
```

新增派生变量：

```ts
const missingSummaryFields = [
  !doctorSummary.trim() ? '医生小结' : '',
  !treatmentSummary.trim() ? '护理小结' : '',
].filter(Boolean)
```

第一版只提示，不强制阻断保存。

---

## 5. 建议代码调整范围

只改：

```text
ai-hms-frontend/src/pages/dialysis-processing/execution/DialysisSummary.tsx
```

不改：

```text
restApi.updateTreatmentSummary
orderApi.list
Order
RestTreatment
onTreatmentUpdated
autoSummary 的字段来源
doctorSummary / treatmentSummary 字段名
```

---

## 6. 可复用组件建议

### 6.1 SummaryMetric

```tsx
function SummaryMetric({
  title,
  value,
  description,
  tone,
}: {
  title: string
  value: string
  description: string
  tone: 'blue' | 'emerald' | 'rose' | 'indigo'
}) {
  return (...)
}
```

### 6.2 ReviewCard

```tsx
function ReviewCard({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (...)
}
```

### 6.3 MissingSummaryFields

```ts
const missingSummaryFields = useMemo(() => [
  !doctorSummary.trim() ? '医生小结' : '',
  !treatmentSummary.trim() ? '护理小结' : '',
].filter(Boolean), [doctorSummary, treatmentSummary])
```

### 6.4 复制自动汇总，可选

```ts
const handleCopyAutoSummary = async () => {
  await navigator.clipboard.writeText(autoSummary)
  message.success('自动汇总已复制')
}
```

如果考虑兼容性，可以 try/catch。

---

## 7. 实施步骤

### Commit 1：顶部结项总览优化

```text
fix: refine dialysis summary overview dashboard
```

内容：

1. 将顶部三张大卡合并为“结项总览”。
2. 增加容量负荷、生命体征、安全质控、医嘱执行 4 个小卡。
3. 字段来源保持不变。

---

### Commit 2：小结书写和自动汇总助手优化

```text
fix: reorganize dialysis summary writing workspace
```

内容：

1. 左侧医生小结和护理小结。
2. 右侧自动汇总参考。
3. 增加书写提示。
4. 保存逻辑不变。

---

### Commit 3：核查摘要与医嘱表格压缩

```text
fix: compact dialysis summary review cards and order table
```

内容：

1. 三个摘要卡压缩。
2. 医嘱明细表行高降低。
3. 保留现有字段。
4. 空状态优化。

---

### Commit 4：底部保存操作条

```text
fix: add dialysis summary action bar and missing hints
```

内容：

1. 新增 `missingSummaryFields`。
2. 底部显示最新治疗时间、监测点、最新备注、缺项。
3. 保存小结按钮固定在底部右侧。
4. 可选增加复制自动汇总。

---

## 8. 验收清单

1. 透析小结页面正常打开。
2. 患者切换后小结内容重置正常。
3. treatment 数据回填正常。
4. 医生小结输入正常。
5. 治疗/护理小结输入正常。
6. 保存小结正常。
7. 保存后 `onTreatmentUpdated` 正常。
8. 自动汇总参考正常显示。
9. 透中监测点数量正常。
10. 凝血分级显示正常。
11. 医嘱列表正常加载。
12. 医嘱为空时空状态正常。
13. 缺项提示显示正常。
14. 接口和字段未修改。
15. 控制台无 JS 报错。
16. 1366x768 下页面层级更清晰，底部明细不抢占主体空间。

---

## 9. 给开发 AI 的执行提示词

```markdown
请优化 AI-HMS 智能透析系统“透析执行 / 透析小结”页面。页面文件为 `ai-hms-frontend/src/pages/dialysis-processing/execution/DialysisSummary.tsx`。请保持接口、字段和保存逻辑不变，只优化布局、样式和展现方式。

重要要求：
1. 不修改后端接口。
2. 不修改字段名。
3. 不修改 `restApi.updateTreatmentSummary` 调用方式。
4. 不修改 `orderApi.list(patient.id, { includeExpired: false })` 调用方式。
5. 不修改 `doctorSummary` 和 `treatmentSummary` 字段。
6. 不修改 `onTreatmentUpdated(updated.data)` 逻辑。
7. 不修改 `autoSummary` 的字段来源。
8. 保留 `treatmentLoading` 提示。

当前页面问题：
1. 顶部三张摘要卡占空间，结项重点不够集中。
2. 小结书写区和自动汇总参考上下分离，填写效率一般。
3. 下方透前/透后摘要、双人核对摘要、凝血通路摘要占高较多。
4. 医嘱明细表行高偏大。
5. 底部缺少小结保存状态和缺项提示。

优化目标：
1. 顶部改为“结项总览”，包含：
   - 容量负荷
   - 生命体征
   - 安全质控
   - 医嘱执行
2. 小结书写区改为左侧医生小结/护理小结，右侧自动汇总助手。
3. 自动汇总助手显示 `autoSummary`，并增加书写引导。
4. 将透前/透后摘要、双人核对摘要、凝血与通路摘要压缩为三张紧凑卡。
5. 医嘱执行明细表保留字段，但压缩行高。
6. 新增 `missingSummaryFields` 派生变量，只做提示，不阻断保存。
7. 底部固定操作条显示：
   - 最新治疗时间
   - 透中监测点
   - 最新备注
   - 缺项
   - 保存小结
8. 可选增加“复制自动汇总”按钮。

验收：
1. 页面加载正常。
2. 患者切换正常。
3. 医生小结和护理小结输入正常。
4. 保存小结正常。
5. 自动汇总显示正常。
6. 医嘱明细加载正常。
7. 缺项提示正常。
8. 接口和字段未修改。
9. 控制台无 JS 错误。
```
