# AI-HMS 健康宣教界面优化设计执行方案

适用页面：`ai-hms-frontend/src/pages/dialysis-processing/execution/HealthEducation.tsx`  
目标：不修改接口、不修改字段、不改变宣教内容加载、患者宣教历史加载、重置和保存逻辑，只优化“健康宣教”界面的布局、表单层级、记录体验和历史展示。

---

## 1. 当前代码结构核查

当前 `HealthEducation.tsx` 已经具备完整功能：

1. 页面接收：
   - `patient`
   - `treatment`
   - `treatmentLoading`

2. 表单字段包括：
   - `contentId`
   - `educationType`
   - `educationTime`
   - `finishTime`
   - `educationResult`
   - `nurseSign`
   - `patientSign`
   - `note`

3. 页面加载数据：
   - `restApi.getHealthEducationContents()`
   - `restApi.getPatientHealthEducations(patient.id)`

4. 保存逻辑：
   - `healthEducationId: form.contentId`
   - `educationTime`
   - `educationType`
   - `educationResult`
   - `nurseSign`
   - `patientSign`
   - `finishTime`
   - `note`
   - 调用 `restApi.createPatientHealthEducation(patient.id, payload)`

5. 当前界面结构：
   - 来源提示条
   - 左侧核心宣教参数
   - 右侧宣教详情记录描述
   - 下方患者宣教历史
   - 底部重置/保存操作

因此本次只改前端 UI，不改接口和字段。

---

## 2. 当前界面主要问题

从截图和代码看，当前页面功能完整，但展现上有几个问题：

### 2.1 左侧表单过长

当前左侧从宣教内容题目、内容描述、宣教方式、宣教日期、宣教人、效果评价、签字、完成日期一路纵向排列，用户需要从上到下填写，密度偏低。

### 2.2 右侧大文本框过空

右侧 `textarea` 高度约 520px，如果用户只是记录一小段宣教内容，会显得空旷。它也缺少结构化提示，例如“宣教重点、患者反馈、后续提醒”。

### 2.3 历史记录被放在较下方

患者宣教历史对避免重复宣教很重要，但当前位置在下方，首屏内存在感偏弱。

### 2.4 缺少宣教完成前核对提示

健康宣教记录保存前，至少应提醒：

```text
患者是否理解本次宣教重点
是否需要家属共同知情
是否需要继续宣教
签字和日期是否完整
```

### 2.5 底部缺项提示不明显

当前底部只显示“保存状态、所属患者、治疗方式”，建议增加“缺项提示”，例如宣教人未填、宣教内容未选。

---

## 3. 推荐新版布局

建议改成“三栏宣教记录工作台”：

```text
健康宣教
├── 顶部患者摘要
│   ├── 患者信息
│   ├── 干体重
│   └── 治疗方案
│
├── 来源提示条
│
├── 主体三栏
│   ├── 左栏：宣教内容选择与基础参数
│   │   ├── 常用宣教内容卡片
│   │   ├── 宣教内容描述
│   │   ├── 宣教方式
│   │   ├── 宣教日期
│   │   └── 完成日期
│   │
│   ├── 中栏：本次宣教记录
│   │   ├── 详情记录描述
│   │   ├── 快捷模板
│   │   ├── 效果评价
│   │   ├── 病患/家属签字
│   │   └── 宣教人
│   │
│   └── 右栏：历史与核对提示
│       ├── 患者宣教历史
│       ├── 宣教完成前核对
│       └── 字段映射提示
│
├── 历史时间线
└── 底部固定操作条
```

---

## 4. 具体优化内容

### 4.1 顶部患者摘要保留但压缩

保留：

```text
支英俊  未排床
ID: 300410   女 / 72岁   费用类别: 市职工普通   透龄: 待补充
干体重 0kg
治疗方案 HD
```

不要放太多其他信息，避免和上层患者摘要重复。

---

### 4.2 来源提示条变成更轻的状态条

当前提示：

```text
健康宣教内容来源于 Auxiliary_HealthEducation，保存记录写入 Auxiliary_PatientHealthEducation。
```

建议保留，但改为 48px 左右的细条，并右侧显示：

```text
可保存
```

这样不占太多高度。

---

### 4.3 左栏：宣教内容选择与基础参数

当前是下拉选择。建议不取消下拉，但在 UI 上可以增强为“常用内容卡片 + 选择框逻辑”。

展示：

```text
宣教内容选择
饮食与水分控制
血管通路保护
用药与复诊提醒
```

点击卡片仍然是设置 `contentId`，不改字段。

保留：

```text
宣教内容描述
宣教方式
宣教日期
完成日期
```

如果内容很多，卡片区可滚动。

---

### 4.4 中栏：本次宣教记录正文

右侧大文本框建议改为中间主区域，并减少高度：

```text
本次宣教记录
请输入本次宣教的具体内容、患者反馈或重点强调的注意事项...
```

在文本框下方增加快捷模板：

```text
已掌握
需继续宣教
家属已知晓
```

第一版可以只是按钮样式，点击后追加到 `note`；如果担心改逻辑，也可以先仅展示为提示，不增加点击逻辑。

下方放：

```text
效果评价
病患/家属签字
宣教人
```

这样填写路径更自然：

```text
写记录 → 评价 → 签字 → 宣教人 → 保存
```

---

### 4.5 右栏：历史与宣教核对提示

右侧显示：

```text
患者宣教历史
0 条
暂无患者宣教历史
保存后会在这里形成时间线记录。
```

下方增加核对提示：

```text
宣教完成前建议核对
1. 患者是否理解本次宣教重点
2. 是否需要家属共同知情
3. 是否需要后续继续宣教
4. 签字和日期是否完整
```

这个不需要新接口，只是静态提示。

---

### 4.6 历史时间线

当前历史记录在下方以卡片列表展示。建议压缩为历史时间线区域：

```text
历史时间线
暂无患者宣教历史
```

有记录时按 `educationTime` 倒序展示：

```text
宣教内容名称
宣教人 / 方式 / 评价
日期
```

保持现有 `records.map`，只是样式调整。

---

### 4.7 底部固定操作条

保留：

```text
重置当前表单
保存宣教记录
```

增加缺项提示：

```text
保存状态：可保存
所属患者：支英俊
治疗方式：HD
缺项：宣教人
```

建议新增派生变量：

```ts
const missingRequiredFields = [
  !form.contentId ? '宣教内容' : '',
  !form.educationTime ? '宣教日期' : '',
  !form.nurseSign.trim() ? '宣教人' : '',
].filter(Boolean)
```

第一版只提示，不阻断保存。当前保存逻辑只强制 contentId，这一点不改变。

---

## 5. 建议代码调整范围

只改：

```text
ai-hms-frontend/src/pages/dialysis-processing/execution/HealthEducation.tsx
```

不改：

```text
restApi.getHealthEducationContents
restApi.getPatientHealthEducations
restApi.createPatientHealthEducation
CreatePatientHealthEducationRequest
字段名
handleSave
resetForm
formatDate
```

---

## 6. 可复用组件建议

当前已有 `Field` 组件，建议保留。

可新增轻量组件：

### 6.1 ContentCard

```tsx
function ContentCard({
  item,
  active,
  onClick,
}: {
  item: HealthEducationContentApi
  active: boolean
  onClick: () => void
}) {
  return (
    <button type="button" onClick={onClick}>
      <div>{item.name}</div>
      <div>{item.description || '暂无描述'}</div>
    </button>
  )
}
```

### 6.2 MissingRequiredFields

```ts
const missingRequiredFields = useMemo(() => [
  !form.contentId ? '宣教内容' : '',
  !form.educationTime ? '宣教日期' : '',
  !form.nurseSign.trim() ? '宣教人' : '',
].filter(Boolean), [form.contentId, form.educationTime, form.nurseSign])
```

### 6.3 appendNoteTemplate 可选

```ts
const appendNoteTemplate = (text: string) => {
  setForm((current) => ({
    ...current,
    note: current.note ? `${current.note}\n${text}` : text,
  }))
}
```

如果担心影响逻辑，第一版可以不做点击，仅展示快捷标签。

---

## 7. 实施步骤

### Commit 1：顶部摘要和来源提示优化

```text
fix: polish health education header and source hint
```

内容：

1. 顶部患者摘要压缩。
2. 来源提示条变轻。
3. 保存状态显示更清楚。

---

### Commit 2：宣教内容选择区优化

```text
fix: improve health education content selection layout
```

内容：

1. 左侧增加宣教内容卡片展示。
2. 点击卡片仍然更新 `contentId`。
3. 保留原 selectedContent 描述。
4. 宣教方式、宣教日期、完成日期保持原字段。

---

### Commit 3：记录正文和评价签字优化

```text
fix: reorganize health education narrative form
```

内容：

1. 中间展示宣教详情记录。
2. 文本框高度适当降低。
3. 增加快捷模板提示。
4. 效果评价、病患/家属签字、宣教人放到正文下方。
5. 字段和保存 payload 不变。

---

### Commit 4：历史记录与底部操作优化

```text
fix: refine health education history and action bar
```

内容：

1. 历史记录改成右侧摘要 + 下方时间线。
2. 增加宣教完成前核对提示。
3. 底部增加缺项提示。
4. 重置和保存按钮样式优化。

---

## 8. 验收清单

1. 健康宣教页面正常打开。
2. 患者切换后重新加载宣教内容和历史记录。
3. 宣教内容列表正常加载。
4. 默认选中第一个宣教内容逻辑不变。
5. 点击宣教内容卡片能正确更新 `contentId`。
6. 宣教内容描述显示正常。
7. 宣教方式选择正常。
8. 宣教日期、完成日期输入正常。
9. 宣教人输入正常。
10. 效果评价输入正常。
11. 病患/家属签字输入正常。
12. 详情记录 note 输入正常。
13. 重置当前表单正常。
14. 保存宣教记录正常。
15. 保存后历史记录追加到顶部。
16. treatmentLoading 时仍显示加载提示。
17. API 和字段未修改。
18. 控制台无 JS 报错。
19. 1366x768 下页面空白减少，记录区更实用。

---

## 9. 给开发 AI 的执行提示词

```markdown
请优化 AI-HMS 智能透析系统“透析执行 / 健康宣教”页面。页面文件为 `ai-hms-frontend/src/pages/dialysis-processing/execution/HealthEducation.tsx`。请保持接口、字段和保存逻辑不变，只优化布局、样式和展现方式。

重要要求：
1. 不修改后端接口。
2. 不修改字段名。
3. 不修改 `CreatePatientHealthEducationRequest`。
4. 不修改 `restApi.getHealthEducationContents`、`restApi.getPatientHealthEducations`、`restApi.createPatientHealthEducation` 的调用方式。
5. 不改变 `handleSave` 的 payload 字段结构。
6. 不改变 `resetForm` 逻辑。
7. 保留 treatmentLoading 提示。
8. 保留保存后 records 追加逻辑。

当前页面问题：
1. 左侧核心宣教参数表单较长。
2. 右侧大文本框过空，缺少结构化填写引导。
3. 患者宣教历史位置偏下，存在感较弱。
4. 缺少宣教完成前核对提示。
5. 底部缺少必填缺项提示。

优化目标：
1. 页面改成三栏宣教记录工作台：
   - 左栏：宣教内容选择与基础参数
   - 中栏：本次宣教记录正文、效果评价、签字和宣教人
   - 右栏：患者宣教历史和核对提示
2. 宣教内容仍使用 `contentId`，可用卡片方式展示内容列表，点击卡片更新 `contentId`。
3. 保留 `selectedContent.description`，但展示为更紧凑的描述卡。
4. 右侧大文本框改为中间主记录区，减少高度，并增加快捷模板提示：
   - 已掌握
   - 需继续宣教
   - 家属已知晓
5. 历史记录增加右侧摘要和下方时间线。
6. 增加宣教完成前核对提示：
   - 患者是否理解本次宣教重点
   - 是否需要家属共同知情
   - 是否需要后续继续宣教
   - 签字和日期是否完整
7. 底部固定操作条增加缺项提示：
   - 宣教内容
   - 宣教日期
   - 宣教人
8. 新增 `missingRequiredFields` 派生变量，只做提示，不强制阻断保存。
9. 重置当前表单和保存宣教记录按钮保留。
10. 页面在 1366x768 下减少空白，提高填写效率。

验收：
1. 页面加载正常。
2. 宣教内容加载正常。
3. 历史记录加载正常。
4. 点击内容卡片更新 `contentId` 正常。
5. 所有输入项正常。
6. 重置正常。
7. 保存正常。
8. 保存后历史记录追加正常。
9. 接口和字段未修改。
10. 控制台无 JS 错误。
```
