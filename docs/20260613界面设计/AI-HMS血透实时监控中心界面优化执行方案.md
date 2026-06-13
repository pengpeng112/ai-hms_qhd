# AI-HMS 智能透析实时监控中心界面优化执行方案

适用页面：血透系统“实时监控中心”  
目标：在不修改后端接口、不改变原有搜索、区域筛选、床位点击、患者详情等功能的基础上，重构页面信息层级和床位卡片样式，使界面更清晰、更像医疗监控看板。

---

## 1. 当前界面问题

根据现有截图，当前界面主要问题不是功能错误，而是视觉层级和卡片密度问题：

1. 空床卡片过大：大部分床位为空床，但仍显示完整监控字段，导致一屏大量重复空值。
2. 黑色粗边框过重：卡片边框过强，像调试面板，不像医疗业务系统。
3. 信息层级不清：空床、透析中、预警、异常没有明显区分。
4. 顶部统计弱：只有“当前在治 0 位患者”，缺少总床位、空床、预警、异常、离线等核心指标。
5. 页面密度不合理：一行 5 张大卡，空值很多，实际有效信息少。
6. 左侧菜单与主内容过近：主内容与侧栏之间建议增加视觉留白和浅色分区背景。

---

## 2. 改造原则

必须遵守：

1. 不修改后端 API。
2. 不修改字段名。
3. 不改变床位点击、搜索、区域筛选、患者详情等原有功能。
4. 不删除原始监控指标。
5. 只调整页面布局、卡片结构、样式和状态展示。
6. 空床减少展示内容；透析中、预警、异常才展示完整监控信息。
7. 以“床位监控看板”为目标，不做复杂大屏特效。

---

## 3. 新页面结构

建议页面结构：

```text
实时监控中心
├── 顶部标题区
│   ├── 标题：实时监控中心
│   ├── 副标题：当前在治 / 总床位 / 空床 / 预警 / 异常
│   └── 右侧：搜索、区域筛选、状态筛选、刷新
│
├── 统计概览卡
│   ├── 总床位
│   ├── 透析中
│   ├── 空床
│   ├── 预警
│   ├── 异常
│   └── 离线
│
├── 状态说明条
│   └── 提示透析中、空床、预警、异常的展示规则
│
└── 床位卡片区
    ├── 透析中卡片
    ├── 预警卡片
    ├── 异常卡片
    └── 空床紧凑卡片
```

---

## 4. 顶部区域设计

### 4.1 标题区

文案建议：

```text
实时监控中心
当前在治 8 位患者 · 总床位 32 · 空床 20 · 预警 3 · 异常 1
```

### 4.2 工具区

保留现有功能：

```text
搜索患者、床号
全部 / A区 / B区 / C区
状态筛选：全部 / 透析中 / 空床 / 预警 / 异常 / 离线
刷新按钮
```

### 4.3 顶部统计卡

统计卡建议由前端根据床位列表计算，不需要新增接口。

```javascript
monitorSummary() {
  const beds = this.beds || [];
  return {
    total: beds.length,
    active: beds.filter(b => b.status === 'dialysis').length,
    empty: beds.filter(b => b.status === 'empty').length,
    warning: beds.filter(b => b.status === 'warning').length,
    danger: beds.filter(b => b.status === 'danger').length,
    offline: beds.filter(b => b.status === 'offline').length,
  };
}
```

---

## 5. 床位卡片设计

### 5.1 空床卡片

空床卡片只显示必要信息：

```text
A01
空床
当前无透析患者
设备状态：在线
区域：A区
```

高度建议：110–130px。不要展示干体重、增量、血管通路、血压、心率、超滤进度等空值。

### 5.2 透析中卡片

透析中卡片展示完整监控数据：

```text
床号：A01
患者：张明 男 56岁
状态：透析中
血压 BP：132/78 mmHg
心率 HR：78 bpm
干体重：60.5 kg
增量：2.1 kg
血管通路：左前臂内瘘
超滤进度：62%
```

高度建议：220–240px。

### 5.3 预警/异常卡片

预警卡片使用橙色强调，异常卡片使用红色强调：

```text
顶部状态条
有色边框
异常指标高亮
状态标签
```

---

## 6. 布局尺寸建议

```css
.monitor-page {
  padding: 20px 24px;
  background: #f5f7fb;
  min-height: 100%;
}

.bed-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}
```

卡片高度：

```text
空床：120px 左右
透析中：220–240px
预警/异常：230–250px
```

---

## 7. CSS 样式建议

```css
.monitor-page {
  padding: 20px 24px;
  background: #f5f7fb;
  min-height: 100%;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.monitor-title {
  font-size: 24px;
  font-weight: 800;
  color: #0f172a;
}

.monitor-subtitle {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
}

.monitor-tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.monitor-summary-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.monitor-summary-card {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 14px 16px;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.04);
}

.bed-card {
  border-radius: 16px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
  padding: 14px;
  transition: all 0.2s ease;
}

.bed-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.10);
}

.bed-card-empty {
  min-height: 120px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.bed-card-active {
  min-height: 220px;
  border-color: #bfdbfe;
}

.bed-card-warning {
  min-height: 230px;
  border-color: #f59e0b;
  box-shadow: 0 8px 24px rgba(245, 158, 11, 0.14);
}

.bed-card-danger {
  min-height: 230px;
  border-color: #ef4444;
  box-shadow: 0 8px 24px rgba(239, 68, 68, 0.14);
}

.bed-card-offline {
  opacity: 0.72;
  filter: grayscale(0.2);
}

.bed-card-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f1f5f9;
}

.bed-no {
  min-width: 42px;
  height: 32px;
  border-radius: 12px;
  background: #0f1f3a;
  color: #ffffff;
  font-weight: 800;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bed-vitals-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 12px;
}

.bed-vitals-grid div {
  background: #f8fafc;
  border-radius: 10px;
  padding: 10px;
}

.bed-vitals-grid span {
  display: block;
  font-size: 12px;
  color: #64748b;
}

.bed-vitals-grid b {
  display: block;
  margin-top: 4px;
  color: #0f172a;
}

.bed-card-footer {
  margin-top: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #64748b;
  font-size: 12px;
}
```

响应式：

```css
@media (max-width: 1280px) {
  .monitor-summary-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .monitor-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .monitor-tools {
    width: 100%;
    flex-wrap: wrap;
  }

  .monitor-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bed-grid {
    grid-template-columns: 1fr;
  }
}
```

---

## 8. 前端逻辑建议

### 8.1 床位状态判断

```javascript
bedStatus(bed) {
  if (bed.deviceStatus === 'offline') return 'offline';
  if (!bed.patientId && !bed.patientName) return 'empty';
  if (bed.hasDanger || bed.alarmLevel === 'danger') return 'danger';
  if (bed.hasWarning || bed.alarmLevel === 'warning') return 'warning';
  return 'active';
}
```

### 8.2 卡片 class

```javascript
bedCardClass(bed) {
  const status = this.bedStatus(bed);
  return {
    'bed-card-empty': status === 'empty',
    'bed-card-active': status === 'active',
    'bed-card-warning': status === 'warning',
    'bed-card-danger': status === 'danger',
    'bed-card-offline': status === 'offline',
  };
}
```

### 8.3 筛选逻辑

```javascript
filteredBeds() {
  return (this.bedList || []).filter((bed) => {
    const areaOk = this.areaFilter === '全部' || bed.areaName === this.areaFilter;
    const keyword = String(this.monitorKeyword || '').trim();
    const keywordOk = !keyword
      || String(bed.bedNo || '').includes(keyword)
      || String(bed.patientName || '').includes(keyword)
      || String(bed.patientId || '').includes(keyword);

    const status = this.bedStatus(bed);
    const statusOk = this.statusFilter === '全部' || this.statusFilter === status;

    return areaOk && keywordOk && statusOk;
  });
}
```

---

## 9. 执行步骤

### Commit 1：监控页结构和统计卡

```text
feat: redesign dialysis monitor header and summary cards
```

内容：

1. 顶部标题和副标题。
2. 统计卡。
3. 搜索、区域筛选、状态筛选。
4. 主内容 padding 调整。

### Commit 2：床位卡片分状态展示

```text
feat: compact empty beds and highlight active dialysis cards
```

内容：

1. 空床紧凑卡。
2. 透析中完整卡。
3. 预警橙色卡。
4. 异常红色卡。
5. 离线灰态卡。
6. 去掉黑色粗边框。

### Commit 3：细节和响应式优化

```text
fix: polish dialysis monitor responsive layout and visual hierarchy
```

内容：

1. 1366x768 适配。
2. 移动端单列。
3. hover 效果。
4. 空状态提示。
5. 保留床位点击事件。

---

## 10. 验收清单

1. 页面打开后顶部能看到总床位、透析中、空床、预警、异常。
2. 空床卡片明显变小，不再展示大量 `--`。
3. 透析中卡片展示完整监控指标。
4. 预警卡片橙色高亮。
5. 异常卡片红色高亮。
6. 搜索患者、床号正常。
7. A区 / B区 / C区筛选正常。
8. 状态筛选正常。
9. 点击床位原有功能不受影响。
10. 页面滚动正常。
11. 1366x768 下不拥挤。
12. 移动端不横向溢出。
13. 控制台无 JS 报错。
