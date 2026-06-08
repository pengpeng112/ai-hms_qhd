# 前端流畅度整改报告

> 整改日期: 2026-06-08

## 已实施

### 1. 保存后局部更新（避免刷新整周）

**旧行为**: `handleSave()` → `closeModal(); await loadWeek()` → 重拉整周 412KB

**新行为**: `handleSave()` → 直接修改本地 `data.patientShifts` 数组（追加或替换）

`Schedule.tsx:252` — 新建排班追加到数组末尾，编辑排班替换对应项。

### 2. 删除后局部移除

**旧行为**: `handleDelete()` → `await loadWeek()` → 重拉整周

**新行为**: `handleDelete()` → `setData(p => filter(s => s.id !== item.id))` 直接移除

### 3. 待排患者列表去重

`schedule_week_service.go:329` — 同一患者多 Plan 只保留首条，消除前端重复 key

### 4. 待排患者 key 防重复

`Schedule.tsx:628` — key 从 `p.id` 改为 `` `${p.id}-${p.patientPlanId}` ``

### 5. loadMap 性能修复

循环从 `beds × days × shifts` (2700次) 降为 `days × shifts` (21次)

### 6. 患者拖拽支持

右侧待排患者卡片可拖拽到空格子，弹窗自动填入所有字段

## 未实施（待后续）

| 优化项 | 原因 | 建议 |
|--------|------|------|
| 虚拟滚动 | 25床位×7天×3班次=525单元格，量不大，虚拟化性价比低 | 床位>100时再考虑 |
| React.memo | 需拆分子组件，改动大 | 后续阶段 |
| 按病区懒加载 | 需改造前端选中逻辑 | 当前默认全部病区 |
| 搜索 debounce | 待排患者已限制200条 | 数据量<1000可暂缓 |
