# 患者管理字典改造待确认（2026-04-23）

## 本次已完成
- 菜单：患者管理 -> 基本信息档案
- 页面：[BasicInfoTab.tsx](/F:/python/前后端代码/ai-hms_qhd/ai-hms-frontend/src/pages/patient-detail/tabs/BasicInfoTab.tsx)
- 调整：
  - `idType/patientType/visitCategory/aboBloodType/rhBloodType/educationLevel/maritalStatus` 编辑态下拉改为优先读字典数据。
  - 去除上述字段硬编码枚举兜底；字典不可用时仅回退当前值，避免写死业务值。
  - `insuranceType` 级联选择去除“自费”硬编码默认项，改为仅使用字典树。

## 待确认项
- 菜单：患者管理 -> 基本信息档案 -> 家属与紧急联系人
- 页面字段：`与患者关系`
- 页面位置：[BasicInfoTab.tsx](/F:/python/前后端代码/ai-hms_qhd/ai-hms-frontend/src/pages/patient-detail/tabs/BasicInfoTab.tsx:1041)
- 现状：固定选项 `配偶/父亲/母亲/子女/兄弟/姐妹/其他`
- 不确定点：
  - legacy 字典表 `CodeDictionary_CodeDictionarys` 对应 `Type` 未确认。
  - 是否要求该字段完全改为字典驱动（含排序、启停、增删）。
- 建议确认：
  - 请提供该字段字典 `Type`（若有）与期望菜单行为；确认后可立即改造为统一字典加载。
