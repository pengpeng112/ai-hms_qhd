# `/api/v1/patients/:id/basic-info` 三方字段对照说明

## 1. 目标

本文用于说明当前 `/api/v1/patients/:id/basic-info` 在以下三侧之间的差异：

1. **前端期望显示/保存的字段**
2. **后端当前实际返回的字段与取值来源**
3. **legacy 数据库真实可提供的字段来源**

目的不是直接给出改造实现，而是先明确：

- 哪些字段前端已经在等
- 哪些字段后端已经返回
- 哪些字段数据库明明有，但当前接口没取
- 哪些字段当前即使查了，也没有进入响应

---

## 2. 本次核对使用的关键文件

### 前端

- `ai-hms-frontend/src/pages/patient-detail/tabs/BasicInfoTab.tsx`
- `ai-hms-frontend/src/services/restClient.ts`

### 后端

- `ai-hms-backend/internal/api/v1/patient_basic_handler.go`
- `ai-hms-backend/internal/services/patient_basic_service.go`
- `ai-hms-backend/internal/services/patient_basic_types.go`
- `ai-hms-backend/internal/models/patient_basic_info.go`

### 数据库文档

- `数据库表结构.md`
- `数据库表设计.pdf`

### 本次重点函数

- `PatientBasicInfoHandler.GetBasicInfo`
- `PatientBasicService.GetBasicInfo`
- `PatientBasicService.buildPersonalInfo`
- `PatientBasicService.buildMedicalInfo`
- `PatientBasicService.buildVitalSocialInfo`
- `PatientBasicService.buildContactInfo`

---

## 3. 先说结论

当前 basic-info 接口的核心问题不是“前端没字段”，而是：

1. **前端已经定义并渲染了大量字段**。
2. **后端 GET 只从两类模型取值**：
   - `models.Patient`
   - `models.PatientBasicInfo`
3. `models.PatientBasicInfo` 对应的是**新表** `patient_basic_infos`，不是 legacy 表。
4. 在 legacy 库里，basic-info 需要的大量字段其实分散在：
   - `Register_PatientInfomation`
   - `Register_Hospitalization`
   - `Register_IDInfomation`
   - `Register_FamilyMember`
5. 因此只要 `patient_basic_infos` 没有同步数据，当前接口就会出现大面积空值。
6. 另外，后端 `GetBasicInfo` 当前还查询了：
   - `models.VascularAccess`
   - `models.MedicalHistory`
   但**查询结果完全没有进入返回 JSON**。

---

## 4. 前端当前期望的 basic-info 结构

前端 `restClient.ts` 中定义的响应类型为：

```ts
interface PatientBasicInfoResponse {
  personalInfo: PatientBasicPersonal
  medicalInfo: PatientBasicMedical
  vitalSocialInfo: PatientBasicVitalSocial
  contactInfo: PatientBasicContact
  // TODO: familyContacts, electronicDocuments
}
```

也就是说，前端当前正式依赖的还是 4 个 section：

- `personalInfo`
- `medicalInfo`
- `vitalSocialInfo`
- `contactInfo`

其中：

- `familyContacts`：目前只是 **前端本地状态**，不是接口正式返回字段
- `electronicDocuments`：目前仍是 **TODO**，未纳入 basic-info 响应

---

## 5. 前端页面实际展示/编辑的字段

来源文件：`ai-hms-frontend/src/pages/patient-detail/tabs/BasicInfoTab.tsx`

### 5.1 核心身份信息 `personalInfo`

页面期望字段：

- `name`
- `pinyin`
- `birthday`
- `age`（只读显示）
- `gender`
- `ethnicity`
- `idType`
- `idNumber`
- `patientType`

额外页面项：

- `internalName`：只读显示 `patient.name`
- `autoConfirm`：纯 UI，不属于后端字段

### 5.2 医疗登记信息 `medicalInfo`

页面期望字段：

- `visitCategory`
- `admissionNo`
- `visitNo`
- `medicalRecordNo`
- `insuranceNo`
- `hdisPatientId`
- `insuranceType`
- `dialysisNo`
- `doctorName`
- `nurseName`
- `firstDialysisDate`
- `firstHospitalDate`
- `firstDialysisHospital`
- `currentDialysisAge`（只读显示）

### 5.3 体征与社会状态 `vitalSocialInfo`

页面期望字段：

- `height`
- `dryWeight`
- `aboBloodType`
- `rhBloodType`
- `educationLevel`
- `occupation`
- `maritalStatus`
- `workplace`

### 5.4 联系方式 `contactInfo`

页面期望字段：

- `phone`
- `wechat`
- `landline`
- `address`
- `district`
- `contactName`
- `contactPhone`

### 5.5 家属与紧急联系人

前端当前不是直接读取 `familyContacts[]`。

当前行为是：

- 在 `fetchBasicInfo()` 里读取 `data.contactInfo.contactName` 和 `data.contactInfo.contactPhone`
- 如果二者同时存在，则前端临时生成一条本地联系人：
  - `id: 'emergency'`
  - `type: 'emergency'`
  - `relation: '-'`
- 然后放进本地状态 `familyContacts`

因此当前“家属/紧急联系人”区块本质上只是：

- 把 `contactInfo.contactName/contactPhone` 转成一条本地展示项
- 新增/删除联系人仅影响前端本地状态
- **不会随 `updatePatientBasicInfo` 持久化回后端**

### 5.6 电子文书

`electronicDocuments` 在 UI 上已有占位和弹窗入口，但不在 basic-info 正式响应中，也不在当前保存请求中。

---

## 6. 后端当前 GET `/basic-info` 的真实取值逻辑

来源文件：`ai-hms-backend/internal/services/patient_basic_service.go`

### 6.1 `GetBasicInfo` 当前做了什么

`PatientBasicService.GetBasicInfo(patientID string)` 当前流程：

1. 解析 `patientID`
2. 从 `models.Patient` 查询主患者记录
3. 从 `models.PatientBasicInfo` 查询扩展记录
4. 额外查询：
   - `models.VascularAccess`
   - `models.MedicalHistory`
5. 调用以下函数构建响应：
   - `buildPersonalInfo(patient, basicInfo)`
   - `buildMedicalInfo(patient, basicInfo)`
   - `buildVitalSocialInfo(patient, basicInfo)`
   - `buildContactInfo(basicInfo)`

### 6.2 但真正进入响应的只有这两类来源

- `models.Patient`
- `models.PatientBasicInfo`

`vascularAccess` 和 `medicalHistory` 虽然查了，但**没有参与任何 build* 返回值**。

---

## 7. 后端响应字段来源矩阵

> 说明：本节只描述“当前代码实际上怎么返回”，不是推荐方案。

### 7.1 `personalInfo`

| 返回字段 | 当前来源 | 说明 |
|---|---|---|
| `name` | `models.Patient.Name` | 直接返回 |
| `pinyin` | `models.PatientBasicInfo.Pinyin` | 新表扩展字段 |
| `birthday` | `models.PatientBasicInfo.Birthday` | 通过 `formatDatePtr` 转成 `YYYY-MM-DD` |
| `age` | `models.Patient.Age` | 直接返回 |
| `gender` | `models.Patient.Gender` | 直接返回 |
| `ethnicity` | `models.PatientBasicInfo.Ethnicity` | 新表扩展字段 |
| `idType` | `models.PatientBasicInfo.IDType` | 新表扩展字段 |
| `idNumber` | `models.PatientBasicInfo.IDNumber` | 通过 `stringPtrOrEmpty`；nil 时返回空字符串 |
| `patientType` | `models.Patient.PatientType` | 为空时通过 `stringOrDefault(..., "门诊")` 默认值兜底 |

### 7.2 `medicalInfo`

| 返回字段 | 当前来源 | 说明 |
|---|---|---|
| `visitCategory` | `models.PatientBasicInfo.VisitCategory` | 新表扩展字段 |
| `admissionNo` | `models.PatientBasicInfo.AdmissionNo` | 新表扩展字段 |
| `visitNo` | `models.PatientBasicInfo.VisitNo` | 新表扩展字段 |
| `medicalRecordNo` | `models.PatientBasicInfo.MedicalRecordNo` | 新表扩展字段 |
| `insuranceNo` | `models.PatientBasicInfo.InsuranceNo` | 新表扩展字段 |
| `hdisPatientId` | `models.PatientBasicInfo.HdisPatientID` | 新表扩展字段 |
| `insuranceType` | `models.Patient.InsuranceType` | 为空时 `stringOrDefault(..., "自费")` |
| `dialysisNo` | `models.PatientBasicInfo.DialysisNo` | 新表扩展字段 |
| `doctorName` | `models.Patient.DoctorName` | 当前来自主患者模型 |
| `nurseName` | `models.PatientBasicInfo.NurseName` | 新表扩展字段 |
| `firstDialysisDate` | `models.PatientBasicInfo.FirstDialysisDate` | 通过 `formatDatePtr` 格式化 |
| `firstHospitalDate` | `models.PatientBasicInfo.FirstHospitalDate` | 通过 `formatDatePtr` 格式化 |
| `firstDialysisHospital` | `models.PatientBasicInfo.FirstDialysisHospital` | 新表扩展字段 |
| `currentDialysisAge` | 由 `FirstDialysisDate` 计算 | 使用 `CalculateDialysisAge` 计算得出 |

### 7.3 `vitalSocialInfo`

| 返回字段 | 当前来源 | 说明 |
|---|---|---|
| `height` | `models.PatientBasicInfo.Height` | 新表扩展字段 |
| `dryWeight` | `models.Patient.DryWeight` | 主患者表字段 |
| `aboBloodType` | `models.PatientBasicInfo.ABOBloodType` | 新表扩展字段 |
| `rhBloodType` | `models.PatientBasicInfo.RhBloodType` | 新表扩展字段 |
| `educationLevel` | `models.PatientBasicInfo.EducationLevel` | 新表扩展字段 |
| `occupation` | `models.PatientBasicInfo.Occupation` | 新表扩展字段 |
| `maritalStatus` | `models.PatientBasicInfo.MaritalStatus` | 新表扩展字段 |
| `workplace` | `models.PatientBasicInfo.Workplace` | 新表扩展字段 |

### 7.4 `contactInfo`

| 返回字段 | 当前来源 | 说明 |
|---|---|---|
| `phone` | `models.PatientBasicInfo.Phone` | 新表扩展字段 |
| `wechat` | `models.PatientBasicInfo.Wechat` | 新表扩展字段 |
| `landline` | `models.PatientBasicInfo.Landline` | 新表扩展字段 |
| `address` | `models.PatientBasicInfo.Address` | 新表扩展字段 |
| `district` | `models.PatientBasicInfo.District` | 新表扩展字段 |
| `contactName` | `models.PatientBasicInfo.ContactName` | 新表扩展字段 |
| `contactPhone` | `models.PatientBasicInfo.ContactPhone` | 新表扩展字段 |

---

## 8. 当前后端已查询但没有返回的数据

来源函数：`PatientBasicService.GetBasicInfo`

当前代码还查询了：

- `models.VascularAccess`
- `models.MedicalHistory`

但查询结果没有传入：

- `buildPersonalInfo`
- `buildMedicalInfo`
- `buildVitalSocialInfo`
- `buildContactInfo`

所以现状是：

- **查了，但没用**
- 即使 legacy 表中有相关信息，当前 `/basic-info` 也不会返回它们

---

## 9. legacy 数据库里哪些表其实能提供这些字段

来源文档：`数据库表结构.md`、`数据库表设计.pdf`

### 9.1 主来源：`Register_PatientInfomation`

这是当前 basic-info 最重要的 legacy 来源表，已经覆盖了大量页面需要字段：

- 身份/人口学：
  - `Name`
  - `Spell`
  - `Gender`
  - `BirthDate`
  - `Nation`
  - `IDName`
  - `PatientType`
- 血型/体征/社会属性：
  - `ABOType`
  - `RHType`
  - `Height`
  - `Weight`
  - `Occupation`
  - `MaritalStatus`
  - `EducationLevel`
  - `Workunit`
- 地址/联系方式：
  - `Province`
  - `City`
  - `County`
  - `Address`
  - `PhoneNo`
  - `HomePhoneNo`
  - `WeChatNo`
- 医保/透析登记：
  - `ExpenseType`
  - `SSN`
  - `DialysisNo`
  - `ResponsibilityDrId`
  - `ResponsibilityNurseId`
  - `FirstDialysisDate`
  - `FirstDialysisHospital`
  - `OurHospitalFirstDialysisDate`

### 9.2 医疗登记补充来源：`Register_Hospitalization`

更适合补充住院/院内登记类字段：

- `CaseNo`
- `HospNo`
- `BarCode`
- `HospPatientType`
- `HospReceiveDept`
- `HospWard`
- `HospBed`
- `AttendDr`
- `ReceptionDr`
- `MedicalRecordNo`

### 9.3 证件权威来源：`Register_IDInfomation`

证件类字段更适合从这里取：

- `IDType`
- `IDNo`

### 9.4 家属/紧急联系人来源：`Register_FamilyMember`

家属与联系人更适合从这里取：

- `Name`
- `Kinship`
- `PhoneNo`
- `Type`

---

## 10. 三方字段对照（前端期望 vs 当前后端返回 vs legacy 可提供）

> 状态说明：
>
> - **已返回**：当前接口已有值来源
> - **依赖新表**：后端已返回，但来源是 `patient_basic_infos`
> - **legacy 可提供但未接**：数据库有，当前 GET 没取
> - **仅前端本地/未建模**：前端有 UI，但接口未正式承载

### 10.1 核心身份信息

| 前端字段 | 当前后端返回 | 当前来源 | legacy 可提供来源 | 判断 |
|---|---|---|---|---|
| `name` | 是 | `Patient.Name` | `Register_PatientInfomation.Name` | 已返回 |
| `pinyin` | 是 | `PatientBasicInfo.Pinyin` | `Register_PatientInfomation.Spell` | legacy 可提供但当前依赖新表 |
| `birthday` | 是 | `PatientBasicInfo.Birthday` | `Register_PatientInfomation.BirthDate` | legacy 可提供但当前依赖新表 |
| `age` | 是 | `Patient.Age` | 可由 `BirthDate` 推导 | 已返回 |
| `gender` | 是 | `Patient.Gender` | `Register_PatientInfomation.Gender` | 已返回 |
| `ethnicity` | 是 | `PatientBasicInfo.Ethnicity` | `Register_PatientInfomation.Nation` | legacy 可提供但当前依赖新表 |
| `idType` | 是 | `PatientBasicInfo.IDType` | `Register_IDInfomation.IDType` | legacy 可提供但当前依赖新表 |
| `idNumber` | 是 | `PatientBasicInfo.IDNumber` | `Register_IDInfomation.IDNo` / `Register_PatientInfomation.SSN` | legacy 可提供但当前依赖新表 |
| `patientType` | 是 | `Patient.PatientType` | `Register_PatientInfomation.PatientType` / `Type` | 已返回 |

### 10.2 医疗登记信息

| 前端字段 | 当前后端返回 | 当前来源 | legacy 可提供来源 | 判断 |
|---|---|---|---|---|
| `visitCategory` | 是 | `PatientBasicInfo.VisitCategory` | `Register_Hospitalization.HospPatientType` / `Register_PatientInfomation.Type` | legacy 可提供但当前依赖新表 |
| `admissionNo` | 是 | `PatientBasicInfo.AdmissionNo` | `Register_Hospitalization.HospNo` | legacy 可提供但当前依赖新表 |
| `visitNo` | 是 | `PatientBasicInfo.VisitNo` | `Register_Hospitalization.CaseNo` / `BarCode` | legacy 可提供但当前依赖新表 |
| `medicalRecordNo` | 是 | `PatientBasicInfo.MedicalRecordNo` | `Register_Hospitalization.MedicalRecordNo` | legacy 可提供但当前依赖新表 |
| `insuranceNo` | 是 | `PatientBasicInfo.InsuranceNo` | `Register_PatientInfomation.SSN` | legacy 可提供但当前依赖新表 |
| `hdisPatientId` | 是 | `PatientBasicInfo.HdisPatientID` | legacy 文档未见直接同名字段 | 当前依赖新表 |
| `insuranceType` | 是 | `Patient.InsuranceType` | `Register_PatientInfomation.ExpenseType` | 已返回，但 legacy 也可供 |
| `dialysisNo` | 是 | `PatientBasicInfo.DialysisNo` | `Register_PatientInfomation.DialysisNo` | legacy 可提供但当前依赖新表 |
| `doctorName` | 是 | `Patient.DoctorName` | `Register_Hospitalization.AttendDr` / `ReceptionDr` / `Register_PatientInfomation.ResponsibilityDrId` | 已返回，但 legacy 补充路径未显式接入 |
| `nurseName` | 是 | `PatientBasicInfo.NurseName` | `Register_PatientInfomation.ResponsibilityNurseId` | legacy 可提供但当前依赖新表 |
| `firstDialysisDate` | 是 | `PatientBasicInfo.FirstDialysisDate` | `Register_PatientInfomation.FirstDialysisDate` | legacy 可提供但当前依赖新表 |
| `firstHospitalDate` | 是 | `PatientBasicInfo.FirstHospitalDate` | `Register_PatientInfomation.OurHospitalFirstDialysisDate` | legacy 可提供但当前依赖新表 |
| `firstDialysisHospital` | 是 | `PatientBasicInfo.FirstDialysisHospital` | `Register_PatientInfomation.FirstDialysisHospital` | legacy 可提供但当前依赖新表 |
| `currentDialysisAge` | 是 | 由 `FirstDialysisDate` 计算 | 可由 legacy `FirstDialysisDate` 计算 | 已返回，但基底字段当前依赖新表 |

### 10.3 体征与社会状态

| 前端字段 | 当前后端返回 | 当前来源 | legacy 可提供来源 | 判断 |
|---|---|---|---|---|
| `height` | 是 | `PatientBasicInfo.Height` | `Register_PatientInfomation.Height` | legacy 可提供但当前依赖新表 |
| `dryWeight` | 是 | `Patient.DryWeight` | 可参考 `Weight` / 业务口径另定 | 已返回 |
| `aboBloodType` | 是 | `PatientBasicInfo.ABOBloodType` | `Register_PatientInfomation.ABOType` | legacy 可提供但当前依赖新表 |
| `rhBloodType` | 是 | `PatientBasicInfo.RhBloodType` | `Register_PatientInfomation.RHType` | legacy 可提供但当前依赖新表 |
| `educationLevel` | 是 | `PatientBasicInfo.EducationLevel` | `Register_PatientInfomation.EducationLevel` | legacy 可提供但当前依赖新表 |
| `occupation` | 是 | `PatientBasicInfo.Occupation` | `Register_PatientInfomation.Occupation` | legacy 可提供但当前依赖新表 |
| `maritalStatus` | 是 | `PatientBasicInfo.MaritalStatus` | `Register_PatientInfomation.MaritalStatus` | legacy 可提供但当前依赖新表 |
| `workplace` | 是 | `PatientBasicInfo.Workplace` | `Register_PatientInfomation.Workunit` | legacy 可提供但当前依赖新表 |

### 10.4 联系方式

| 前端字段 | 当前后端返回 | 当前来源 | legacy 可提供来源 | 判断 |
|---|---|---|---|---|
| `phone` | 是 | `PatientBasicInfo.Phone` | `Register_PatientInfomation.PhoneNo` | legacy 可提供但当前依赖新表 |
| `wechat` | 是 | `PatientBasicInfo.Wechat` | `Register_PatientInfomation.WeChatNo` | legacy 可提供但当前依赖新表 |
| `landline` | 是 | `PatientBasicInfo.Landline` | `Register_PatientInfomation.HomePhoneNo` | legacy 可提供但当前依赖新表 |
| `address` | 是 | `PatientBasicInfo.Address` | `Register_PatientInfomation.Address` | legacy 可提供但当前依赖新表 |
| `district` | 是 | `PatientBasicInfo.District` | `Register_PatientInfomation.Province/City/County` | legacy 可提供但当前依赖新表 |
| `contactName` | 是 | `PatientBasicInfo.ContactName` | `Register_FamilyMember.Name` | legacy 可提供但当前依赖新表 |
| `contactPhone` | 是 | `PatientBasicInfo.ContactPhone` | `Register_FamilyMember.PhoneNo` | legacy 可提供但当前依赖新表 |

### 10.5 家属联系人与电子文书

| 前端能力 | 当前后端返回 | legacy 可提供来源 | 判断 |
|---|---|---|---|
| `familyContacts` 本地列表 | 否 | `Register_FamilyMember` | 仅前端本地/未建模 |
| `electronicDocuments` | 否 | 暂未纳入本次 basic-info 范围 | TODO |

---

## 11. 为什么当前页面会出现大量空值

根因基本可以归纳为 4 点：

### 11.1 后端扩展字段主要依赖新表 `patient_basic_infos`

而不是直接从 legacy 表读取：

- `Register_PatientInfomation`
- `Register_Hospitalization`
- `Register_IDInfomation`
- `Register_FamilyMember`

只要新表没有对应数据，以下字段就大概率为空：

- `pinyin`
- `birthday`
- `ethnicity`
- `idType`
- `idNumber`
- `visitCategory`
- `admissionNo`
- `visitNo`
- `medicalRecordNo`
- `insuranceNo`
- `dialysisNo`
- `nurseName`
- `firstDialysisDate`
- `firstHospitalDate`
- `firstDialysisHospital`
- `height`
- `aboBloodType`
- `rhBloodType`
- `educationLevel`
- `occupation`
- `maritalStatus`
- `workplace`
- `phone`
- `wechat`
- `landline`
- `address`
- `district`
- `contactName`
- `contactPhone`

### 11.2 证件、住院登记、联系人本来就在别的 legacy 表

也就是说这些信息不是“数据库没有”，而是：

- **库里有**
- **当前 GET 没去读**

尤其典型的是：

- `Register_IDInfomation`：证件类型/证件号
- `Register_Hospitalization`：住院号、病历号、床位等
- `Register_FamilyMember`：家属联系人

### 11.3 家属联系人当前只是前端临时拼装

前端并没有真正消费 `familyContacts[]`，只是把：

- `contactInfo.contactName`
- `contactInfo.contactPhone`

转成一条本地 emergency 联系人展示出来。

### 11.4 后端存在“查了但没返回”的冗余查询

这说明当前实现也还没完成 legacy 信息整合：

- `vascularAccess`
- `medicalHistory`

目前只是查了，没有进入输出结构。

---

## 12. 对后续改造的直接启示

如果下一步要修 `/basic-info`，应优先把字段来源从“新表优先”切换为“legacy 可用字段优先回填”，至少覆盖以下几个方向：

1. **身份信息回填**
   - `Spell` → `pinyin`
   - `BirthDate` → `birthday`
   - `Nation` → `ethnicity`
   - `IDType/IDNo` → `idType/idNumber`

2. **医疗登记回填**
   - `HospNo` / `CaseNo` / `MedicalRecordNo`
   - `ExpenseType`
   - `DialysisNo`
   - `FirstDialysisDate`
   - `OurHospitalFirstDialysisDate`
   - `FirstDialysisHospital`

3. **联系方式与地址回填**
   - `PhoneNo`
   - `HomePhoneNo`
   - `WeChatNo`
   - `Province/City/County/Address`

4. **体征与社会属性回填**
   - `ABOType`
   - `RHType`
   - `Height`
   - `Occupation`
   - `MaritalStatus`
   - `EducationLevel`
   - `Workunit`

5. **家属联系人建模**
   - 将 `Register_FamilyMember` 正式映射到响应数组，而不是继续依赖 `contactName/contactPhone` 单值占位

---

## 13. 当前建议的判断口径

在后续逐字段修复时，建议统一按以下口径判断：

- **数据库无字段**：才算真正缺设计
- **数据库有字段但接口未取**：属于映射缺失
- **接口有字段但来源依赖新表**：属于 legacy 兼容不足
- **前端有 UI 但接口未建模**：属于接口能力缺口

按这份文档判断，当前 `/basic-info` 的大部分问题属于：

> **legacy 数据库可提供，但当前 GET 映射没有接上。**

---

## 14. 可直接作为下一步改造清单的高优先字段

建议优先处理以下高价值字段：

- `idType`
- `idNumber`
- `admissionNo`
- `visitNo`
- `medicalRecordNo`
- `insuranceNo`
- `dialysisNo`
- `nurseName`
- `firstDialysisDate`
- `firstHospitalDate`
- `firstDialysisHospital`
- `height`
- `aboBloodType`
- `rhBloodType`
- `educationLevel`
- `occupation`
- `maritalStatus`
- `workplace`
- `phone`
- `wechat`
- `landline`
- `address`
- `district`
- `contactName`
- `contactPhone`

这些字段的共同特点是：

- 前端已定义
- 用户已关注
- legacy 表存在来源
- 当前接口却仍高度依赖 `patient_basic_infos`
