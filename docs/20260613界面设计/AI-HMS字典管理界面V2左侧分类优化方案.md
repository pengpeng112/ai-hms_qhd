# AI-HMS 字典管理界面 V2 优化方案：左侧分类更直观版

适用页面：`ai-hms-frontend/src/pages/DictConfig.tsx`  
约束：不修改后端接口、不修改字段名、不改变新增/编辑/删除/启停逻辑。  
目标：重点解决“左侧字典分类不直观、其他字典列表太长、分类和类型混在一起”的问题。

---

## 1. 本次优化重点

上一版已经解决了部分密度问题，但左侧仍然容易让用户觉得：

1. 字典分类卡片和具体字典类型混在一起。
2. “其他字典”下面类型太多，用户不清楚该先看分类还是看类型。
3. `ABO血型 / ABOType` 这种具体字典藏在很长列表里，定位不直观。
4. 右侧如果按分类展示多个表格，也容易出现信息堆叠。

因此 V2 不改接口、不改字段，只改前端组织方式：

```text
原方式：
左侧 = 业务分类卡片 + 其他字典长列表

V2：
左侧 = 一级业务域导航 + 二级字典类型列表
右侧 = 当前选中字典类型的字典值表格
```

---

## 2. 推荐新版布局

```text
字典配置
├── 顶部说明区
│   ├── 字典配置
│   ├── 总类型 / 本地维护 / 老库只读
│   └── 刷新
│
├── 左侧一级：业务域导航
│   ├── 常用快捷入口
│   ├── 透析治疗
│   ├── 血管通路
│   ├── 医嘱处方
│   ├── 人员信息
│   ├── 临床诊疗
│   ├── 转归记录
│   └── 通用/未归类
│
├── 中间二级：当前业务域下的字典类型
│   ├── 搜索类型名称 / TypeCode
│   ├── 全部 / 本地 / 老库
│   └── 类型列表
│
└── 右侧：当前字典值维护
    ├── 当前维护字典摘要
    ├── 表内搜索
    ├── 全部 / 启用 / 停用
    └── 字典值表格
```

---

## 3. 左侧一级分类优化

### 3.1 分类名称调整

保持代码里的 `CategoryKey` 不变，但前端展示文案更符合用户理解。

| CategoryKey | 当前含义 | 建议展示 |
|---|---|---|
| dialysis | 透析治疗 | 透析治疗 |
| vascular | 血管通路 | 血管通路 |
| order | 医嘱处方 | 医嘱处方 |
| staff | 人员信息 | 人员信息 |
| clinical | 临床诊疗 | 临床诊疗 |
| outcome | 转归记录 | 转归记录 |
| other | 其他字典 | 通用/未归类 |

其中 `other` 建议前端显示为：

```text
通用/未归类
```

说明：

```text
其他来源字典统一归档
```

这样比“其他字典”更容易理解：不是业务上的“其他”，而是“未归类/通用”。

### 3.2 业务域导航卡片样式

每个业务域用一条清晰的导航项：

```text
图标   分类名称              数量
       简短说明
```

示例：

```text
诊  临床诊疗        18类
    诊断、医保、患者属性
```

选中状态：

```text
深色背景
左侧 3px 蓝色高亮条
图标蓝色
文字白色
```

---

## 4. 中间二级字典类型列表

### 4.1 从“其他字典长列表”改为“当前分类类型列表”

当前逻辑里：

```ts
visibleTypes = typesByCategory[selectedCategory]
```

V2 建议仍然使用这个数据，但展示方式改变：

1. 左侧第一级只负责选择 `selectedCategory`。
2. 中间第二级展示 `typesByCategory[selectedCategory]`。
3. 用户点击具体类型后，设置 `selectedTypeCode`。
4. 右侧只展示 `selectedTypeCode` 对应的 `TypeValueTable`。

这比当前普通分类直接 `visibleTypes.map(TypeValueTable)` 更直观。

### 4.2 建议新增状态

```ts
const [selectedTypeCode, setSelectedTypeCode] = useState<string>('')
const [typeKeyword, setTypeKeyword] = useState('')
const [typeSourceFilter, setTypeSourceFilter] = useState<'all' | 'local' | 'legacy'>('all')
```

### 4.3 自动选择默认类型

当 `selectedCategory` 改变时，自动选中该分类第一个类型：

```ts
useEffect(() => {
  const list = typesByCategory[selectedCategory] || []
  if (selectedCategory === 'other' && selectedOtherTypeCode) {
    setSelectedTypeCode(selectedOtherTypeCode)
    return
  }
  if (!list.some((type) => type.code === selectedTypeCode)) {
    setSelectedTypeCode(list[0]?.code || '')
  }
}, [selectedCategory, selectedOtherTypeCode, selectedTypeCode, typesByCategory])
```

也可以逐步替换掉 `selectedOtherTypeCode`，但为了降低风险，第一版可以保留 `selectedOtherTypeCode`，只是让 UI 层统一使用 `selectedTypeCode`。

---

## 5. 常用快捷入口

在一级业务域上方增加“常用快捷入口”，不改 API，只是前端快速定位：

```text
ABO血型
费用类别
透析方式
血管通路
```

点击后：

```ts
setSelectedCategory(resolveCategory(typeCode))
setSelectedTypeCode(typeCode)
if (resolveCategory(typeCode) === 'other') {
  setSelectedOtherTypeCode(typeCode)
}
```

建议第一批快捷入口：

```ts
const COMMON_DICT_SHORTCUTS = [
  { label: 'ABO血型', typeCode: 'ABOType' },
  { label: '费用类别', typeCode: 'ExpenseType' },
  { label: '透析方式', typeCode: 'DIALYSIS_MODE' },
  { label: '血管通路', typeCode: 'VASCULAR_ACCESS' },
]
```

快捷入口只是 UI，失败也不影响原功能。

---

## 6. 右侧只展示当前选中字典类型

当前代码在普通分类下会：

```tsx
visibleTypes.map((type) => (
  <TypeValueTable ... />
))
```

如果一个分类包含多个字典类型，右侧会连续出现多个表格，用户不容易聚焦。

V2 建议改为：

```ts
const selectedType = useMemo(() => {
  return sortedTypes.find((type) => type.code === selectedTypeCode) || null
}, [selectedTypeCode, sortedTypes])
```

右侧渲染：

```tsx
{selectedType ? (
  <TypeValueTable
    type={selectedType}
    items={itemsByType[selectedType.code] || []}
    loading={loadingTypeCodes.has(selectedType.code)}
    onAdd={openAddModal}
    onEdit={openEditModal}
    onToggle={handleToggleItem}
    onDelete={handleDeleteItem}
  />
) : (
  <EmptyState />
)}
```

这样用户始终只维护一个字典类型，界面更清楚。

---

## 7. 右侧当前字典摘要

在表格上方增加：

```text
当前维护字典
ABO血型
ABOType · 临床诊疗 / 血型信息 · 字段与接口不变

本地维护    5项
[新增字典值]
```

好处：

1. 用户知道当前维护的具体字典。
2. `新增字典值` 不再和表格混在一起。
3. 老库来源可在摘要区直接提示“只读”。

---

## 8. 表格仍保留原字段

字段不变：

```text
编码
名称
描述
上级编码
排序
来源
状态
操作
```

只优化：

1. 表格工具栏。
2. 表内搜索。
3. 启用/停用筛选。
4. 表头 sticky。
5. 表格内部滚动。
6. 空状态文案。

---

## 9. 建议代码改造范围

只改：

```text
ai-hms-frontend/src/pages/DictConfig.tsx
```

不改：

```text
ai-hms-frontend/src/services/dictApi.ts
后端接口
数据库字段
API 字段名
```

可选：如果页面 class 太长，后续再拆组件：

```text
DictCategoryRail.tsx
DictTypeList.tsx
DictValueTable.tsx
```

第一版不建议拆太多，避免引入风险。

---

## 10. 执行步骤

### Commit 1：左侧分类改为两级导航

```text
fix: make dictionary category navigation more intuitive
```

内容：

1. 新增 `selectedTypeCode`。
2. 一级业务域导航改为纵向业务域列表。
3. 中间增加当前分类下的字典类型列表。
4. `other` 展示文案改为“通用/未归类”。
5. 字典类型列表固定高度滚动。

### Commit 2：常用快捷入口和类型筛选

```text
fix: add dictionary shortcuts and type search
```

内容：

1. 增加常用字典快捷入口。
2. 增加类型搜索 `typeKeyword`。
3. 增加来源筛选 `typeSourceFilter`。
4. 点击快捷入口自动定位分类和类型。

### Commit 3：右侧只展示当前字典类型

```text
fix: show selected dictionary type values only
```

内容：

1. 不再在普通分类下 `visibleTypes.map` 渲染多个表格。
2. 右侧只显示 `selectedType` 的 `TypeValueTable`。
3. 增加当前字典摘要卡。
4. 保留新增、编辑、启停、删除功能。

### Commit 4：表格工具栏与密度优化

```text
fix: polish dictionary value table filters and density
```

内容：

1. 表内搜索。
2. 启用/停用筛选。
3. sticky 表头。
4. 表格内部滚动。
5. 空状态优化。

---

## 11. 验收清单

1. 页面进入后默认选中一个业务域和一个字典类型。
2. 一级业务域切换正常。
3. 二级字典类型列表切换正常。
4. `通用/未归类` 能正常显示原 `other` 类型。
5. 常用快捷入口能正确定位字典类型。
6. ABO血型能正常显示字典值。
7. 新增字典值正常。
8. 编辑字典值正常。
9. 启用/停用正常。
10. 删除正常。
11. 老库来源仍然禁用维护。
12. 接口和字段没有改动。
13. 页面不会被其他字典长列表撑高。
14. 1366x768 下左侧分类直观、右侧表格清晰。
15. 控制台无 JS 报错。

---

## 12. 给开发 AI 的完整提示词

```markdown
请继续优化 AI-HMS 智能透析系统“字典配置”页面，页面文件为 `ai-hms-frontend/src/pages/DictConfig.tsx`。接口和字段不要动，不要修改 `dictApi.ts` 的 API 请求方式，不要修改后端接口，不要改变新增、编辑、删除、启停逻辑。

用户反馈：当前界面布局仍不合适，特别是左侧字典分类不直观。请重点优化左侧分类展示。

当前问题：
1. 左侧把业务分类和具体字典类型混在一起。
2. “其他字典”展开后列表很长，不直观。
3. 用户不清楚应该先选分类还是先选具体字典。
4. 普通分类右侧会展示多个表格，用户焦点不明确。

优化目标：
1. 左侧改成两级导航：一级业务域 + 二级字典类型。
2. 一级业务域包括：透析治疗、血管通路、医嘱处方、人员信息、临床诊疗、转归记录、通用/未归类。
3. 注意：只是前端展示文案优化，CategoryKey 仍然保持 `other`，不要改接口字段。
4. 二级字典类型列表展示当前业务域下的具体类型。
5. 新增 `selectedTypeCode`，右侧只展示当前选中字典类型的字典值。
6. 不要再对普通分类使用 `visibleTypes.map` 连续渲染多个 TypeValueTable。
7. “其他字典”前端展示为“通用/未归类”，并加入搜索框，固定高度滚动。
8. 增加常用快捷入口：ABO血型、费用类别、透析方式、血管通路。
9. 常用快捷入口只做前端定位，不改 API。
10. 右侧增加当前字典摘要卡，展示名称、code、来源、数量和新增按钮。
11. TypeValueTable 字段保持不变：编码、名称、描述、上级编码、排序、来源、状态、操作。
12. 可增加表内搜索和启用/停用筛选，但仅前端过滤。
13. 老库来源仍然只读，编辑、启停、删除仍然禁用。
14. 页面整体适配 1366x768，不要让左侧长列表撑高整个页面。

建议新增状态：
- selectedTypeCode
- typeKeyword
- typeSourceFilter
- itemKeyword
- itemStatusFilter

验收：
1. 默认进入页面能看到清晰的业务域导航和字典类型列表。
2. 点击业务域后，中间类型列表切换。
3. 点击具体字典类型后，右侧表格切换。
4. 点击 ABO血型 快捷入口能定位到 ABOType。
5. 新增、编辑、启停、删除功能不受影响。
6. 老库来源仍然禁止维护。
7. 接口和字段未修改。
8. 控制台无 JS 错误。
```
