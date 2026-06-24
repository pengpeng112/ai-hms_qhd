# C4 收费归集模块 实现方案

> **执行说明**：用 subagent-driven-development 或 executing-plans 逐任务实现。步骤用 `- [ ]` 跟踪。
> **依据**：`ai-hms_规则C4_收费归集.md` v1.0（2026-06-23 已确认）
> **基线**：origin/master `5d66b8d`（A1-A4/B1-B4/C1-C3 全部已合入）

**目标**：在 ai-hms 内建"治疗费用项归集引擎"——每次治疗自动归集 A治疗费/B耗材费/C护理费/D注射费/E药品 五类费用项，护士可增删改、确认、双人核对，导出护士友好的 Excel 给 HIS 人工录入；HIS 推送走预留接口（本期不实现）。

**架构**：
- 后端 = 数据真值源 + 预留推送接口。归集引擎读 `Treatment_MaterialTrace`（耗材实际用量）+ `Plan_PatientPrescription`（治疗模式/时长）+ `medication_admin`（给药→注射费/药品项）+ 患者通路类型，套 `billing_catalog.json`（用户①表最新参考价）产出清单。状态机 `draft→confirmed→checked→pushed→settled/cancelled`。
- **参考价单一来源** = `billing_catalog.json`（go:embed，院方可改价重构建即生效）。老库 `Stock_ChargeItem.Price` 全是 0，**不可用**。
- **Excel 导出 = 前端生成**（`xlsx` SheetJS 已装，3处在用），house 风格 `json_to_sheet`+`!cols`，护士读取/复制友好。
- **HIS 推送 = 预留**：后端定义 `HISPusher` 接口 + `NoopPusher` 桩（照搬 CNRDS `Exporter` 模式），本期只把清单标 `pushed` 并交前端导出 Excel。

**技术栈**：Go + GORM + gin（后端）；React + TypeScript + antd + lucide-react + xlsx（前端）。

---

## 文件结构

**后端（在 ai-hms-backend/ 各 clone 内，最终交付基线 5d66b8d）**
- `internal/config/billing_catalog.json` — 新建，参考价目（治疗费/护理费/注射费/耗材按 ChargeItemId/不可计费集）
- `internal/config/billing_catalog.go` — 新建，go:embed 加载器 + 查询方法
- `internal/models/charge_record.go` — 新建，清单头（一次结算一条）
- `internal/models/charge_line.go` — 新建，清单明细行
- `internal/services/billing_service.go` — 新建，归集引擎 + 编辑 + 状态流 + 报表
- `internal/services/billing_pusher.go` — 新建，HISPusher 接口 + NoopPusher 桩（预留）
- `internal/services/billing_service_test.go` — 新建，引擎/编辑/状态/报表测试（SQLite 内存）
- `internal/api/v1/billing_handler.go` — 新建，路由 + handler
- `internal/database/health.go` — 改，RequiredNewTables 加 charge_record/charge_line
- `cmd/server/main.go` — 改，加 `RegisterBillingRoutes(protected)`
- `docs/sql/deploy_new_tables.sql` — 改，加表 22/23 DDL

**前端（在 ai-hms-frontend/）**
- `src/services/billingApi.ts` — 新建，API 客户端
- `src/lib/billingExcel.ts` — 新建，护士友好 Excel 生成
- `src/pages/patient-detail/tabs/BillingTab.tsx` — 新建，患者「收费归集」Tab
- `src/pages/patient-detail/tabs/index.ts` — 改，导出 BillingTab
- `src/pages/PatientDetail.tsx` — 改，菜单项 + switch 分支 + SectionID
- `src/pages/BillingPage.tsx` — 新建，独立「收费管理」页（报表/批量，Phase 6）

---

## 数据模型

**charge_record（清单头）**
| 列 | 类型 | 说明 |
|----|------|------|
| id | varchar(36) PK | utils.GenerateID() |
| tenant_id | bigint NOT NULL | LegacyTenantID=3 |
| patient_id | bigint | 患者 |
| treatment_id | bigint | 本次治疗 Treatment.Id |
| prescription_id | bigint | 当日处方 Plan_PatientPrescription.Id |
| charge_date | timestamptz | 治疗日期 |
| shift | varchar(16) | 班次码 early/late |
| dialysis_mode | varchar(16) | HD/HDF/HF/HP/HD+HP/PE/DFPP/CRRT |
| access_type | varchar(16) | AVF/AVG/TCC/NCC（护理费用） |
| crrt_hours | decimal(5,2) | CRRT 时长，按时长计价用 |
| total_amount | decimal(10,2) | 可计费行参考金额合计 |
| status | varchar(16) | draft/confirmed/checked/pushed/settled/cancelled |
| recorded_by/_name | varchar(64) | 记账护士 |
| checked_by/_name | varchar(64) | 核对护士 |
| checked_at | timestamptz | |
| pushed_at | timestamptz | 推送/导出时间 |
| note | varchar(256) | |
| created_at/updated_at | timestamptz | |

**charge_line（清单明细）**
| 列 | 类型 | 说明 |
|----|------|------|
| id | varchar(36) PK | |
| tenant_id | bigint NOT NULL | |
| charge_record_id | varchar(36) | FK→charge_record，index |
| category | varchar(16) | treatment/material/nursing/injection/drug |
| item_code | varchar(64) | 收费/医保编码 |
| item_name | varchar(128) | |
| spec | varchar(64) | 规格（药品用） |
| unit | varchar(16) | |
| quantity | decimal(10,2) | |
| unit_price | decimal(10,2) NULL | 参考单价；药品为 NULL |
| amount | decimal(10,2) NULL | quantity×unit_price；NULL=待核价 |
| billable | boolean | false=不可计费（仍展示，不计入合计） |
| source | varchar(8) | auto/manual |
| charge_item_id | bigint | 关联老库耗材 Id（material 行用） |
| note | varchar(256) | |
| created_at | timestamptz | |

---

## Task 1：参考价目 billing_catalog.json + 加载器

**Files:**
- Create: `internal/config/billing_catalog.json`
- Create: `internal/config/billing_catalog.go`
- Create: `internal/config/billing_catalog_test.go`
- 数据源（本会话已读，在桌面）：用户①表 `01ai-hms_C4_收费医保_调研表医疗部分.xlsx` + 团队 `系统导出数据及说明.rar/charge_item.csv`

- [ ] **Step 1：生成 catalog JSON 的 join 脚本**（一次性，产出 JSON，避免手抄 130 行出错）

把以下脚本存为 `C:\tmp\build_billing_catalog.js`，`cd C:\tmp && node build_billing_catalog.js`（xlsx 已装在 C:\tmp）。脚本：读①表②表+charge_item.csv，按"型号 token"把①表参考价贴到 charge_item 的 Id 上，输出 `billing_catalog.json`。

```js
const XLSX = require('xlsx'); const fs = require('fs'); const path = require('path')
const SRC = 'C:\\Users\\docto\\Desktop\\HMS开发\\01ai-hms_C4_收费医保_调研表医疗部分.xlsx'
const CHARGE_CSV = 'C:\\tmp\\rar-extract\\系统导出数据及说明\\charge_item.csv'

// 1) 读①表材料价目（name, price, unit, insuranceCode, category）
const wb = XLSX.readFile(SRC)
const mat = XLSX.utils.sheet_to_json(wb.Sheets['①材料价目库_可编辑'], { header: 1, defval: '' })
// R3 是表头；R4+ 数据。列：[2]名称 [3]类别 [4]单价 [5]单位 [6]医保编码
const priceRows = []
for (let i = 4; i < mat.length; i++) {
  const r = mat[i]; const name = String(r[2]).trim(); if (!name) continue
  const price = parseFloat(String(r[4]).replace(/[^\d.]/g, '')) || null
  priceRows.push({ name, category: String(r[3]).trim(), price, unit: String(r[5]).trim(), insuranceCode: String(r[6]).trim().replace(/\s/g,'') })
}

// 2) 读 charge_item.csv（Id, Name, Unit）— 运行时 material_trace.ChargeItemId 精确命中
const csv = fs.readFileSync(CHARGE_CSV, 'utf8').split(/\r?\n/).filter(Boolean)
const cols = csv[0].split(',').map(s => s.replace(/"/g, ''))
const idI = cols.indexOf('Id'), nameI = cols.indexOf('Name'), unitI = cols.indexOf('Unit')
const items = []
for (let i = 1; i < csv.length; i++) {
  // 简单 CSV 解析（字段无内嵌逗号的按逗号切；带引号的去引号）
  const f = csv[i].match(/("([^"]*)")|([^,]*)/g).filter((_,j)=>j%2===0).map(s=>s.replace(/^"|"$/g,''))
  if (!f[idI]) continue
  items.push({ id: parseInt(f[idI],10), name: f[nameI], unit: f[unitI] || '' })
}

// 3) 不可计费 ChargeItem 名（精确/包含匹配 material_trace.material_name）
const NON_BILLABLE_KEYWORDS = ['内瘘包','上机包','下机包','护理包','注射器','敷贴','敷料','A液','B液','浓缩透析液']
const isNonBillable = (nm) => NON_BILLABLE_KEYWORDS.some(k => nm.includes(k))

// 4) 把①表参考价贴到 charge_item（按型号 token：取①表名称里大写字母+数字的型号串，与 charge_item.name 比对）
const tokenOf = (s) => (s.toUpperCase().match(/[A-Z]{1,}[A-Z0-9\-]*[0-9]/g) || []).sort((a,b)=>b.length-a.length)[0] || s
const priceByToken = new Map()
for (const p of priceRows) { if (p.price != null) { const t = tokenOf(p.name); if (!priceByToken.has(t)) priceByToken.set(t, p) } }

const materials = items.map(it => {
  if (isNonBillable(it.name)) return { chargeItemId: it.id, name: it.name, unit: it.unit, billable: false }
  const hit = priceByToken.get(tokenOf(it.name))
  return { chargeItemId: it.id, name: it.name, unit: it.unit, billable: true,
           unitPrice: hit ? hit.price : null, insuranceCode: hit ? hit.insuranceCode : '' }
})

// 5) 治疗费/护理费/注射费（直接来自规则，权威）
const catalog = {
  version: '1.0', effectiveDate: '2026-06-23',
  treatmentFees: [
    { mode:'HD',   name:'血液透析',     price:399, unit:'次',  insuranceCode:'013110000010000', billType:'perSession' },
    { mode:'HDF',  name:'血液透析滤过', price:599, unit:'次',  insuranceCode:'013110000030000', billType:'perSession' },
    { mode:'HF',   name:'血液滤过',     price:399, unit:'次',  insuranceCode:'013110000020000', billType:'perSession' },
    { mode:'HP',   name:'血液灌流',     price:550, unit:'次',  insuranceCode:'013110000040000', billType:'perSession' },
    { mode:'HD+HP',name:'血液透析+灌流',price:850, unit:'次',  insuranceCode:'013110000050000', billType:'perSession' },
    { mode:'PE',   name:'血浆置换',     price:1680,unit:'次',  insuranceCode:'013110000060000', billType:'perSession' },
    { mode:'DFPP', name:'双重血浆置换', price:336, unit:'次',  insuranceCode:'013110000060001', billType:'perSession' },
    { mode:'CRRT', name:'连续性肾脏替代治疗', price:115, unit:'小时', insuranceCode:'013110000080000', billType:'perHour', surcharge:{ name:'连续性血浆吸附滤过(加收)', price:35, insuranceCode:'013110000080001' } }
  ],
  nursingFees: {
    AVF: [{ name:'造口/造瘘护理', price:17, unit:'次', insuranceCode:'011303000090000', qty:1 }],
    AVG: [{ name:'造口/造瘘护理', price:17, unit:'次', insuranceCode:'011303000090000', qty:1 }],
    TCC: [{ name:'置管护理', price:12, unit:'次', insuranceCode:'011303000040000', qty:1 }, { name:'中换药', price:21, unit:'次', insuranceCode:'GL120600003', qty:2 }],
    NCC: [{ name:'置管护理', price:12, unit:'次', insuranceCode:'011303000040000', qty:1 }, { name:'中换药', price:21, unit:'次', insuranceCode:'GL120600003', qty:2 }]
  },
  injectionFee: { name:'静脉注射', price:3, unit:'次', insuranceCode:'' },
  materials
}
fs.writeFileSync('C:\\Users\\docto\\AppData\\Local\\Temp\\aihms-b2\\ai-hms-backend\\internal\\config\\billing_catalog.json', JSON.stringify(catalog, null, 2))
console.log('materials:', materials.length, 'priced:', materials.filter(m=>m.billable&&m.unitPrice!=null).length, 'unpriced:', materials.filter(m=>m.billable&&m.unitPrice==null).length, 'nonBillable:', materials.filter(m=>!m.billable).length)
```

- [ ] **Step 2：跑脚本生成 JSON，人工核一遍 unpriced 项**

Run: `cd C:\tmp && node build_billing_catalog.js`
Expected: 打印 `materials: ~130 priced: N unpriced: M nonBillable: K`。打开生成的 `billing_catalog.json`，把 `unitPrice:null` 的可计费耗材（型号没自动贴上价的）对照①表手工补价；型号确实不在①表的留 null（运行时显示"待核价"，护士补）。

- [ ] **Step 3：写加载器 billing_catalog.go**

```go
package config

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed billing_catalog.json
var billingCatalogJSON []byte

type TreatmentFee struct {
	Mode          string  `json:"mode"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Unit          string  `json:"unit"`
	InsuranceCode string  `json:"insuranceCode"`
	BillType      string  `json:"billType"` // perSession / perHour
	Surcharge     *struct {
		Name          string  `json:"name"`
		Price         float64 `json:"price"`
		InsuranceCode string  `json:"insuranceCode"`
	} `json:"surcharge,omitempty"`
}

type NursingFee struct {
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Unit          string  `json:"unit"`
	InsuranceCode string  `json:"insuranceCode"`
	Qty           float64 `json:"qty"`
}

type InjectionFee struct {
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Unit          string  `json:"unit"`
	InsuranceCode string  `json:"insuranceCode"`
}

type MaterialPrice struct {
	ChargeItemID  int64    `json:"chargeItemId"`
	Name          string   `json:"name"`
	Unit          string   `json:"unit"`
	Billable      bool     `json:"billable"`
	UnitPrice     *float64 `json:"unitPrice"`
	InsuranceCode string   `json:"insuranceCode"`
}

type BillingCatalog struct {
	Version       string                  `json:"version"`
	EffectiveDate string                  `json:"effectiveDate"`
	TreatmentFees []TreatmentFee          `json:"treatmentFees"`
	NursingFees   map[string][]NursingFee `json:"nursingFees"`
	InjectionFee  InjectionFee            `json:"injectionFee"`
	Materials     []MaterialPrice         `json:"materials"`

	treatmentByMode map[string]TreatmentFee
	materialByID    map[int64]MaterialPrice
}

var loadedBillingCatalog *BillingCatalog

func LoadBillingCatalog() (*BillingCatalog, error) {
	if loadedBillingCatalog != nil {
		return loadedBillingCatalog, nil
	}
	var c BillingCatalog
	if err := json.Unmarshal(billingCatalogJSON, &c); err != nil {
		return nil, err
	}
	c.treatmentByMode = make(map[string]TreatmentFee, len(c.TreatmentFees))
	for _, t := range c.TreatmentFees {
		c.treatmentByMode[strings.ToUpper(t.Mode)] = t
	}
	c.materialByID = make(map[int64]MaterialPrice, len(c.Materials))
	for _, m := range c.Materials {
		c.materialByID[m.ChargeItemID] = m
	}
	loadedBillingCatalog = &c
	return loadedBillingCatalog, nil
}

// TreatmentFeeFor 按模式查治疗费（大小写/别名归一）。
func (c *BillingCatalog) TreatmentFeeFor(mode string) (TreatmentFee, bool) {
	t, ok := c.treatmentByMode[strings.ToUpper(strings.TrimSpace(mode))]
	return t, ok
}

// MaterialFor 按老库 ChargeItemId 查耗材参考价。
func (c *BillingCatalog) MaterialFor(chargeItemID int64) (MaterialPrice, bool) {
	m, ok := c.materialByID[chargeItemID]
	return m, ok
}

// NursingFeeFor 按通路类型查护理费项（AVF/AVG/TCC/NCC）。
func (c *BillingCatalog) NursingFeeFor(accessType string) []NursingFee {
	return c.NursingFees[strings.ToUpper(strings.TrimSpace(accessType))]
}
```

- [ ] **Step 4：写测试**

```go
package config

import "testing"

func TestLoadBillingCatalog(t *testing.T) {
	c, err := LoadBillingCatalog()
	if err != nil { t.Fatalf("load: %v", err) }
	if hd, ok := c.TreatmentFeeFor("hd"); !ok || hd.Price != 399 {
		t.Fatalf("HD want 399 got %+v ok=%v", hd, ok)
	}
	if crrt, ok := c.TreatmentFeeFor("CRRT"); !ok || crrt.BillType != "perHour" || crrt.Surcharge == nil || crrt.Surcharge.Price != 35 {
		t.Fatalf("CRRT perHour+surcharge35 got %+v", crrt)
	}
	if n := c.NursingFeeFor("TCC"); len(n) != 2 {
		t.Fatalf("TCC want 2 nursing lines got %d", len(n))
	}
	if len(c.Materials) < 50 {
		t.Fatalf("materials too few: %d", len(c.Materials))
	}
}
```

- [ ] **Step 5：Run** `go test ./internal/config/ -run TestLoadBillingCatalog -v` → PASS
- [ ] **Step 6：Commit** `git add internal/config/billing_catalog.* && git commit -m "feat(C4): 收费归集参考价目 catalog（go:embed）"`

---

## Task 2：charge_record + charge_line 模型 + DDL + 自检

**Files:** Create `internal/models/charge_record.go`、`internal/models/charge_line.go`；Modify `docs/sql/deploy_new_tables.sql`、`internal/database/health.go`

- [ ] **Step 1：charge_record.go**

```go
package models

import "time"

type ChargeRecord struct {
	ID             string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID       int64      `gorm:"column:tenant_id;index:idx_cr_tenant_patient;not null" json:"tenantId"`
	PatientID      int64      `gorm:"column:patient_id;index:idx_cr_tenant_patient" json:"patientId"`
	TreatmentID    int64      `gorm:"column:treatment_id;index:idx_cr_treatment" json:"treatmentId"`
	PrescriptionID int64      `gorm:"column:prescription_id" json:"prescriptionId"`
	ChargeDate     *time.Time `gorm:"column:charge_date;index:idx_cr_date" json:"chargeDate"`
	Shift          string     `gorm:"column:shift;type:varchar(16)" json:"shift"`
	DialysisMode   string     `gorm:"column:dialysis_mode;type:varchar(16)" json:"dialysisMode"`
	AccessType     string     `gorm:"column:access_type;type:varchar(16)" json:"accessType"`
	CrrtHours      float64    `gorm:"column:crrt_hours;type:decimal(5,2)" json:"crrtHours"`
	TotalAmount    float64    `gorm:"column:total_amount;type:decimal(10,2)" json:"totalAmount"`
	Status         string     `gorm:"column:status;type:varchar(16);index:idx_cr_status;not null;default:'draft'" json:"status"`
	RecordedBy     string     `gorm:"column:recorded_by;type:varchar(64)" json:"recordedBy"`
	RecordedName   string     `gorm:"column:recorded_name;type:varchar(64)" json:"recordedName"`
	CheckedBy      string     `gorm:"column:checked_by;type:varchar(64)" json:"checkedBy"`
	CheckedName    string     `gorm:"column:checked_name;type:varchar(64)" json:"checkedName"`
	CheckedAt      *time.Time `gorm:"column:checked_at" json:"checkedAt"`
	PushedAt       *time.Time `gorm:"column:pushed_at" json:"pushedAt"`
	Note           string     `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	Lines []ChargeLine `gorm:"-" json:"lines,omitempty"`
}

func (ChargeRecord) TableName() string { return "charge_record" }

const (
	ChargeStatusDraft     = "draft"
	ChargeStatusConfirmed = "confirmed"
	ChargeStatusChecked   = "checked"
	ChargeStatusPushed    = "pushed"
	ChargeStatusSettled   = "settled"
	ChargeStatusCancelled = "cancelled"
)
```

- [ ] **Step 2：charge_line.go**

```go
package models

import "time"

type ChargeLine struct {
	ID             string    `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	TenantID       int64     `gorm:"column:tenant_id;not null" json:"tenantId"`
	ChargeRecordID string    `gorm:"column:charge_record_id;type:varchar(36);index:idx_cl_record;not null" json:"chargeRecordId"`
	Category       string    `gorm:"column:category;type:varchar(16);not null" json:"category"` // treatment/material/nursing/injection/drug
	ItemCode       string    `gorm:"column:item_code;type:varchar(64)" json:"itemCode"`
	ItemName       string    `gorm:"column:item_name;type:varchar(128);not null" json:"itemName"`
	Spec           string    `gorm:"column:spec;type:varchar(64)" json:"spec"`
	Unit           string    `gorm:"column:unit;type:varchar(16)" json:"unit"`
	Quantity       float64   `gorm:"column:quantity;type:decimal(10,2)" json:"quantity"`
	UnitPrice      *float64  `gorm:"column:unit_price;type:decimal(10,2)" json:"unitPrice"`
	Amount         *float64  `gorm:"column:amount;type:decimal(10,2)" json:"amount"`
	Billable       bool      `gorm:"column:billable;not null;default:true" json:"billable"`
	Source         string    `gorm:"column:source;type:varchar(8);not null;default:'auto'" json:"source"`
	ChargeItemID   int64     `gorm:"column:charge_item_id" json:"chargeItemId"`
	Note           string    `gorm:"column:note;type:varchar(256)" json:"note"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (ChargeLine) TableName() string { return "charge_line" }

const (
	ChargeCatTreatment = "treatment"
	ChargeCatMaterial  = "material"
	ChargeCatNursing   = "nursing"
	ChargeCatInjection = "injection"
	ChargeCatDrug      = "drug"
)
```

- [ ] **Step 3：DDL（deploy_new_tables.sql 末尾追加表 22/23）**

```sql
-- 22. charge_record 收费归集清单头（规则C4）
CREATE TABLE IF NOT EXISTS charge_record (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    patient_id bigint,
    treatment_id bigint,
    prescription_id bigint,
    charge_date timestamptz,
    shift varchar(16),
    dialysis_mode varchar(16),
    access_type varchar(16),
    crrt_hours decimal(5,2),
    total_amount decimal(10,2),
    status varchar(16) NOT NULL DEFAULT 'draft',
    recorded_by varchar(64), recorded_name varchar(64),
    checked_by varchar(64), checked_name varchar(64), checked_at timestamptz,
    pushed_at timestamptz, note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cr_tenant_patient ON charge_record (tenant_id, patient_id);
CREATE INDEX IF NOT EXISTS idx_cr_treatment ON charge_record (treatment_id);
CREATE INDEX IF NOT EXISTS idx_cr_date ON charge_record (tenant_id, charge_date);
CREATE INDEX IF NOT EXISTS idx_cr_status ON charge_record (tenant_id, status);

-- 23. charge_line 收费归集清单明细（规则C4）
CREATE TABLE IF NOT EXISTS charge_line (
    id varchar(36) PRIMARY KEY,
    tenant_id bigint NOT NULL,
    charge_record_id varchar(36) NOT NULL,
    category varchar(16) NOT NULL,
    item_code varchar(64),
    item_name varchar(128) NOT NULL,
    spec varchar(64), unit varchar(16),
    quantity decimal(10,2),
    unit_price decimal(10,2),
    amount decimal(10,2),
    billable boolean NOT NULL DEFAULT true,
    source varchar(8) NOT NULL DEFAULT 'auto',
    charge_item_id bigint, note varchar(256),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cl_record ON charge_line (charge_record_id);
COMMENT ON TABLE charge_record IS '收费归集清单头（规则C4）';
COMMENT ON TABLE charge_line IS '收费归集清单明细（规则C4）';
```

- [ ] **Step 4：health.go RequiredNewTables 加两表**（在 consent_record 行后追加）

```go
	{Table: "charge_record", Feature: "收费归集清单", DDL: "docs/sql/deploy_new_tables.sql"},
	{Table: "charge_line", Feature: "收费归集明细", DDL: "docs/sql/deploy_new_tables.sql"},
```

- [ ] **Step 5：Run** `go build ./...` → exit 0
- [ ] **Step 6：Commit** `git add -A && git commit -m "feat(C4): charge_record/charge_line 模型+DDL+自检"`

---

## Task 3：归集引擎 — BuildDraft 骨架（治疗费 + 注射费）

**Files:** Create `internal/services/billing_service.go`、`internal/services/billing_service_test.go`

- [ ] **Step 1：先写失败测试**（SQLite 内存；只验证治疗费+注射费两行）

```go
package services

import (
	"testing"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newBillingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil { t.Fatal(err) }
	if err := db.AutoMigrate(&models.ChargeRecord{}, &models.ChargeLine{}); err != nil { t.Fatal(err) }
	// 老库表（最小列）
	db.Exec(`CREATE TABLE IF NOT EXISTS "Plan_PatientPrescription" ("Id" INTEGER PRIMARY KEY,"TenantId" INTEGER,"PatientId" INTEGER,"TreatmentId" INTEGER,"DialysisMethod" TEXT,"DialysisDuration" REAL)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS "Treatment_MaterialTrace" ("Id" INTEGER PRIMARY KEY,"TenantId" INTEGER,"TreatmentId" INTEGER,"ChargeItemId" INTEGER,"Num" REAL,"CreateTime" DATETIME)`)
	return db
}

func TestBilling_BuildDraft_TreatmentAndInjection(t *testing.T) {
	db := newBillingTestDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","TreatmentId","DialysisMethod") VALUES (5001,3,1001,9001,'HD')`)
	svc := &BillingService{db: db, tenantID: 3}
	rec, err := svc.BuildDraft(BuildDraftInput{
		PatientID: 1001, TreatmentID: 9001, PrescriptionID: 5001,
		AccessType: "AVF", HasInjection: true,
	})
	if err != nil { t.Fatalf("BuildDraft: %v", err) }
	var treat, inj *models.ChargeLine
	for i := range rec.Lines {
		switch rec.Lines[i].Category {
		case models.ChargeCatTreatment: treat = &rec.Lines[i]
		case models.ChargeCatInjection: inj = &rec.Lines[i]
		}
	}
	if treat == nil || treat.UnitPrice == nil || *treat.UnitPrice != 399 {
		t.Fatalf("treatment HD want 399 got %+v", treat)
	}
	if inj == nil || inj.UnitPrice == nil || *inj.UnitPrice != 3 {
		t.Fatalf("injection want 3 got %+v", inj)
	}
}
```

- [ ] **Step 2：Run** `go test ./internal/services/ -run TestBilling_BuildDraft_TreatmentAndInjection -v` → FAIL（BillingService 未定义）

- [ ] **Step 3：写 billing_service.go 骨架 + BuildDraft（先只产治疗费+注射费）**

```go
package services

import (
	"errors"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type BillingService struct {
	db       *gorm.DB
	tenantID int64
}

func NewBillingService() *BillingService {
	return &BillingService{db: database.GetDB(), tenantID: LegacyTenantID}
}

type BuildDraftInput struct {
	PatientID      int64
	TreatmentID    int64
	PrescriptionID int64
	AccessType     string  // AVF/AVG/TCC/NCC
	CrrtHours      float64 // CRRT 时长，可由 handler 算好传入
	HasInjection   bool    // 本次是否有静脉给药
}

func ptr(f float64) *float64 { return &f }

func (s *BillingService) newLine(recID, cat, code, name, unit string, qty float64, price *float64, billable bool, source string, chargeItemID int64) models.ChargeLine {
	var amount *float64
	if price != nil {
		amount = ptr(qty * (*price))
	}
	return models.ChargeLine{
		ID: utils.GenerateID(), TenantID: s.tenantID, ChargeRecordID: recID,
		Category: cat, ItemCode: code, ItemName: name, Unit: unit,
		Quantity: qty, UnitPrice: price, Amount: amount,
		Billable: billable, Source: source, ChargeItemID: chargeItemID,
	}
}

func (s *BillingService) BuildDraft(in BuildDraftInput) (*models.ChargeRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	cat, err := config.LoadBillingCatalog()
	if err != nil {
		return nil, err
	}
	// 读处方拿治疗模式
	var presc struct {
		DialysisMethod   string
		DialysisDuration float64
	}
	if err := s.db.Table(`"Plan_PatientPrescription"`).
		Select(`"DialysisMethod","DialysisDuration"`).
		Where(`"Id" = ? AND "TenantId" = ?`, in.PrescriptionID, s.tenantID).
		Scan(&presc).Error; err != nil {
		return nil, err
	}
	mode := presc.DialysisMethod
	now := time.Now()
	rec := &models.ChargeRecord{
		ID: utils.GenerateID(), TenantID: s.tenantID, PatientID: in.PatientID,
		TreatmentID: in.TreatmentID, PrescriptionID: in.PrescriptionID,
		ChargeDate: &now, DialysisMode: mode, AccessType: in.AccessType,
		CrrtHours: in.CrrtHours, Status: models.ChargeStatusDraft,
	}
	var lines []models.ChargeLine

	// A 治疗费
	if tf, ok := cat.TreatmentFeeFor(mode); ok {
		if tf.BillType == "perHour" {
			hrs := in.CrrtHours
			if hrs <= 0 { hrs = 1 }
			lines = append(lines, s.newLine(rec.ID, models.ChargeCatTreatment, tf.InsuranceCode, tf.Name, tf.Unit, hrs, ptr(tf.Price), true, "auto", 0))
			// CRRT 加收：默认不勾，护士手动开启 → billable=false 占位
			if tf.Surcharge != nil {
				l := s.newLine(rec.ID, models.ChargeCatTreatment, tf.Surcharge.InsuranceCode, tf.Surcharge.Name, "次", 1, ptr(tf.Surcharge.Price), false, "auto", 0)
				l.Note = "护士手动勾选启用"
				lines = append(lines, l)
			}
		} else {
			lines = append(lines, s.newLine(rec.ID, models.ChargeCatTreatment, tf.InsuranceCode, tf.Name, tf.Unit, 1, ptr(tf.Price), true, "auto", 0))
		}
	}

	// D 注射费（最多 1 项）
	if in.HasInjection {
		inj := cat.InjectionFee
		lines = append(lines, s.newLine(rec.ID, models.ChargeCatInjection, inj.InsuranceCode, inj.Name, inj.Unit, 1, ptr(inj.Price), true, "auto", 0))
	}

	rec.Lines = lines
	return rec, nil
}
```

- [ ] **Step 4：Run** 同 Step 2 命令 → PASS
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 归集引擎 BuildDraft（治疗费+注射费）"`

---

## Task 4：耗材行（material_trace → 可计费/不可计费/待核价）

**Files:** Modify `internal/services/billing_service.go`、`billing_service_test.go`

- [ ] **Step 1：失败测试**

```go
func TestBilling_BuildDraft_Materials(t *testing.T) {
	db := newBillingTestDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","TreatmentId","DialysisMethod") VALUES (5002,3,1002,9002,'HD')`)
	// 27=SUREFLUX-13G(可计费), 19=内瘘包(不可计费), 999999=未知(待核价)
	db.Exec(`INSERT INTO "Treatment_MaterialTrace" ("Id","TenantId","TreatmentId","ChargeItemId","Num") VALUES (1,3,9002,27,1),(2,3,9002,19,1),(3,3,9002,999999,2)`)
	svc := &BillingService{db: db, tenantID: 3}
	rec, err := svc.BuildDraft(BuildDraftInput{PatientID: 1002, TreatmentID: 9002, PrescriptionID: 5002, AccessType: "AVF"})
	if err != nil { t.Fatal(err) }
	var mats []models.ChargeLine
	for _, l := range rec.Lines { if l.Category == models.ChargeCatMaterial { mats = append(mats, l) } }
	if len(mats) != 3 { t.Fatalf("want 3 material lines got %d", len(mats)) }
	for _, l := range mats {
		switch l.ChargeItemID {
		case 19:
			if l.Billable { t.Fatalf("内瘘包 should be non-billable") }
		case 999999:
			if l.UnitPrice != nil { t.Fatalf("unknown item should have nil price (待核价)") }
		case 27:
			if !l.Billable || l.UnitPrice == nil { t.Fatalf("SUREFLUX-13G should be billable+priced") }
		}
	}
}
```

- [ ] **Step 2：Run** → FAIL（只有 0 条 material）
- [ ] **Step 3：在 BuildDraft 的注射费之前插入 B 耗材段**

```go
	// B 耗材费：读 material_trace
	type matRow struct {
		ChargeItemID int64   `gorm:"column:ChargeItemId"`
		Num          float64 `gorm:"column:Num"`
	}
	var mrows []matRow
	if err := s.db.Table(`"Treatment_MaterialTrace"`).
		Select(`"ChargeItemId","Num"`).
		Where(`"TreatmentId" = ? AND "TenantId" = ?`, in.TreatmentID, s.tenantID).
		Scan(&mrows).Error; err != nil {
		return nil, err
	}
	for _, m := range mrows {
		mp, ok := cat.MaterialFor(m.ChargeItemID)
		if !ok {
			// 未知耗材：列入但待核价
			l := s.newLine(rec.ID, models.ChargeCatMaterial, "", "未知耗材", "", m.Num, nil, true, "auto", m.ChargeItemID)
			l.Note = "待核价：目录无此项"
			lines = append(lines, l)
			continue
		}
		if !mp.Billable {
			lines = append(lines, s.newLine(rec.ID, models.ChargeCatMaterial, "", mp.Name, mp.Unit, m.Num, nil, false, "auto", m.ChargeItemID))
			continue
		}
		lines = append(lines, s.newLine(rec.ID, models.ChargeCatMaterial, mp.InsuranceCode, mp.Name, mp.Unit, m.Num, mp.UnitPrice, true, "auto", m.ChargeItemID))
	}
```

> 注意：此段放在 `var lines []models.ChargeLine` 之后、A 治疗费段之前或之后均可（顺序只影响展示）。建议放 A 之后、D 之前，符合 A→B→C→D→E 清单顺序。

- [ ] **Step 4：Run** → PASS（同时 Task3 测试仍 PASS）
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 归集引擎 B耗材费（可计费/不可计费/待核价三档）"`

---

## Task 5：护理费（按通路类型）

**Files:** Modify `billing_service.go`、`billing_service_test.go`

- [ ] **Step 1：失败测试**

```go
func TestBilling_BuildDraft_Nursing(t *testing.T) {
	db := newBillingTestDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","TreatmentId","DialysisMethod") VALUES (5003,3,1003,9003,'HD')`)
	svc := &BillingService{db: db, tenantID: 3}
	// AVF → 1 项造瘘护理 17
	avf, _ := svc.BuildDraft(BuildDraftInput{PatientID: 1003, TreatmentID: 9003, PrescriptionID: 5003, AccessType: "AVF"})
	if n := countCat(avf.Lines, models.ChargeCatNursing); n != 1 { t.Fatalf("AVF want 1 nursing got %d", n) }
	// TCC → 置管护理 + 中换药×2 = 2 行
	tcc, _ := svc.BuildDraft(BuildDraftInput{PatientID: 1003, TreatmentID: 9003, PrescriptionID: 5003, AccessType: "TCC"})
	var changeDressing *models.ChargeLine
	cnt := 0
	for i := range tcc.Lines {
		if tcc.Lines[i].Category == models.ChargeCatNursing {
			cnt++
			if tcc.Lines[i].ItemName == "中换药" { changeDressing = &tcc.Lines[i] }
		}
	}
	if cnt != 2 { t.Fatalf("TCC want 2 nursing got %d", cnt) }
	if changeDressing == nil || changeDressing.Quantity != 2 { t.Fatalf("中换药 qty want 2 got %+v", changeDressing) }
}

func countCat(lines []models.ChargeLine, cat string) int {
	n := 0
	for _, l := range lines { if l.Category == cat { n++ } }
	return n
}
```

- [ ] **Step 2：Run** → FAIL
- [ ] **Step 3：在 B 耗材段之后插入 C 护理费段**

```go
	// C 护理费：按通路类型
	for _, nf := range cat.NursingFeeFor(in.AccessType) {
		qty := nf.Qty
		if qty <= 0 { qty = 1 }
		lines = append(lines, s.newLine(rec.ID, models.ChargeCatNursing, nf.InsuranceCode, nf.Name, nf.Unit, qty, ptr(nf.Price), true, "auto", 0))
	}
```

- [ ] **Step 4：Run** → PASS
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 归集引擎 C护理费（AVF/AVG/TCC/NCC）"`

---

## Task 6：药品项（medication_admin → 列名不带价）+ 落库 + 合计

**Files:** Modify `billing_service.go`、`billing_service_test.go`

- [ ] **Step 1：失败测试**（含药品行 + Save 落库 + total 合计只算 billable）

```go
func TestBilling_SaveDraft_TotalAndDrugs(t *testing.T) {
	db := newBillingTestDB(t)
	db.Exec(`CREATE TABLE IF NOT EXISTS medication_admin ("id" TEXT,"tenant_id" INTEGER,"patient_id" INTEGER,"treatment_id" INTEGER,"drug_name" TEXT,"dose" TEXT,"route" TEXT)`)
	db.Exec(`INSERT INTO medication_admin VALUES ('m1',3,1004,9004,'蔗糖铁注射液','100mg','静脉滴注')`)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","TreatmentId","DialysisMethod") VALUES (5004,3,1004,9004,'HD')`)
	db.Exec(`INSERT INTO "Treatment_MaterialTrace" ("Id","TenantId","TreatmentId","ChargeItemId","Num") VALUES (1,3,9004,27,1)`)
	svc := &BillingService{db: db, tenantID: 3}
	rec, err := svc.BuildDraft(BuildDraftInput{PatientID: 1004, TreatmentID: 9004, PrescriptionID: 5004, AccessType: "AVF", HasInjection: true})
	if err != nil { t.Fatal(err) }
	if countCat(rec.Lines, models.ChargeCatDrug) != 1 { t.Fatalf("want 1 drug line") }
	if err := svc.SaveDraft(rec); err != nil { t.Fatalf("SaveDraft: %v", err) }
	// total = HD 399 + 注射 3 + 造瘘 17 + SUREFLUX-13G 价 ；至少 > 419
	if rec.TotalAmount < 419 { t.Fatalf("total too low: %v", rec.TotalAmount) }
	var cnt int64
	db.Model(&models.ChargeLine{}).Where("charge_record_id = ?", rec.ID).Count(&cnt)
	if cnt == 0 { t.Fatalf("lines not persisted") }
}
```

- [ ] **Step 2：Run** → FAIL
- [ ] **Step 3：BuildDraft 末尾（rec.Lines=lines 之前）加 E 药品段；新增 computeTotal + SaveDraft**

E 药品段（注射费之后）：
```go
	// E 药品项：列名不带价（供 HIS 查价）
	type drugRow struct {
		DrugName string `gorm:"column:drug_name"`
		Dose     string `gorm:"column:dose"`
		Route    string `gorm:"column:route"`
	}
	var drows []drugRow
	if err := s.db.Table("medication_admin").
		Select("drug_name, dose, route").
		Where("treatment_id = ? AND tenant_id = ?", in.TreatmentID, s.tenantID).
		Scan(&drows).Error; err == nil {
		for _, d := range drows {
			l := s.newLine(rec.ID, models.ChargeCatDrug, "", d.DrugName, "", "", 1, nil, true, "auto", 0)
			l.Spec = d.Dose
			l.Note = d.Route + "（HIS 查价）"
			lines = append(lines, l)
		}
	}
```
> ⚠️ `newLine` 第 6 个参数是 unit(string)。上面误写了空串占位，请改为正确签名：`s.newLine(rec.ID, models.ChargeCatDrug, "", d.DrugName, "", 1, nil, true, "auto", 0)`，再单独 `l.Spec = d.Dose; l.Note = d.Route + "（HIS 查价）"`。

computeTotal + SaveDraft：
```go
func computeTotal(lines []models.ChargeLine) float64 {
	var t float64
	for _, l := range lines {
		if l.Billable && l.Amount != nil {
			t += *l.Amount
		}
	}
	return t
}

// SaveDraft 落库清单头+明细（事务）。
func (s *BillingService) SaveDraft(rec *models.ChargeRecord) error {
	rec.TotalAmount = computeTotal(rec.Lines)
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rec).Error; err != nil {
			return err
		}
		if len(rec.Lines) > 0 {
			if err := tx.Create(&rec.Lines).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
```
在 BuildDraft 的 `rec.Lines = lines` 后加 `rec.TotalAmount = computeTotal(lines)`（让未落库的草稿也带合计供预览）。

- [ ] **Step 4：Run** → PASS
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 归集引擎 E药品项+合计+SaveDraft落库"`

---

## Task 7：清单读取 + 护士编辑（增删改行，重算合计）

**Files:** Modify `billing_service.go`、`billing_service_test.go`

- [ ] **Step 1：失败测试**

```go
func TestBilling_EditLines(t *testing.T) {
	db := newBillingTestDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","TreatmentId","DialysisMethod") VALUES (5005,3,1005,9005,'HD')`)
	svc := &BillingService{db: db, tenantID: 3}
	rec, _ := svc.BuildDraft(BuildDraftInput{PatientID: 1005, TreatmentID: 9005, PrescriptionID: 5005, AccessType: "AVF"})
	svc.SaveDraft(rec)
	// 加一行手工项
	added, err := svc.AddLine(rec.ID, models.ChargeLine{Category: models.ChargeCatMaterial, ItemName: "手工耗材", Unit: "个", Quantity: 2, UnitPrice: ptr(10)})
	if err != nil { t.Fatal(err) }
	if added.Source != "manual" || added.Amount == nil || *added.Amount != 20 { t.Fatalf("added line bad: %+v", added) }
	// 改数量
	if _, err := svc.UpdateLine(added.ID, ChargeLinePatch{Quantity: ptr(3)}); err != nil { t.Fatal(err) }
	// 删一行
	if err := svc.DeleteLine(added.ID); err != nil { t.Fatal(err) }
	got, _ := svc.GetRecord(rec.ID)
	for _, l := range got.Lines { if l.ID == added.ID { t.Fatalf("line not deleted") } }
}
```

- [ ] **Step 2：Run** → FAIL
- [ ] **Step 3：实现 GetRecord / AddLine / UpdateLine / DeleteLine / recomputeRecordTotal**

```go
type ChargeLinePatch struct {
	ItemName  *string
	Quantity  *float64
	UnitPrice *float64
	Billable  *bool
	Note      *string
}

func (s *BillingService) GetRecord(id string) (*models.ChargeRecord, error) {
	var rec models.ChargeRecord
	if err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&rec).Error; err != nil {
		return nil, errors.New("清单不存在")
	}
	if err := s.db.Where("charge_record_id = ?", id).Order("created_at").Find(&rec.Lines).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *BillingService) mutableOrErr(recID string) error {
	var st string
	if err := s.db.Model(&models.ChargeRecord{}).Select("status").Where("id = ? AND tenant_id = ?", recID, s.tenantID).Scan(&st).Error; err != nil {
		return err
	}
	if st != models.ChargeStatusDraft && st != models.ChargeStatusConfirmed {
		return errors.New("清单已核对/推送，不可编辑")
	}
	return nil
}

func (s *BillingService) AddLine(recID string, in models.ChargeLine) (*models.ChargeLine, error) {
	if err := s.mutableOrErr(recID); err != nil { return nil, err }
	in.ID = utils.GenerateID()
	in.TenantID = s.tenantID
	in.ChargeRecordID = recID
	in.Source = "manual"
	if in.UnitPrice != nil { in.Amount = ptr(in.Quantity * (*in.UnitPrice)) }
	if !in.Billable && in.Category != "" { /* 保留传入 */ } else if in.Category == "" { in.Billable = true }
	if err := s.db.Create(&in).Error; err != nil { return nil, err }
	return &in, s.recomputeRecordTotal(recID)
}

func (s *BillingService) UpdateLine(lineID string, p ChargeLinePatch) (*models.ChargeLine, error) {
	var l models.ChargeLine
	if err := s.db.Where("id = ? AND tenant_id = ?", lineID, s.tenantID).First(&l).Error; err != nil {
		return nil, errors.New("明细行不存在")
	}
	if err := s.mutableOrErr(l.ChargeRecordID); err != nil { return nil, err }
	if p.ItemName != nil { l.ItemName = *p.ItemName }
	if p.Quantity != nil { l.Quantity = *p.Quantity }
	if p.UnitPrice != nil { l.UnitPrice = p.UnitPrice }
	if p.Billable != nil { l.Billable = *p.Billable }
	if p.Note != nil { l.Note = *p.Note }
	if l.UnitPrice != nil { l.Amount = ptr(l.Quantity * (*l.UnitPrice)) } else { l.Amount = nil }
	if err := s.db.Save(&l).Error; err != nil { return nil, err }
	return &l, s.recomputeRecordTotal(l.ChargeRecordID)
}

func (s *BillingService) DeleteLine(lineID string) error {
	var l models.ChargeLine
	if err := s.db.Where("id = ? AND tenant_id = ?", lineID, s.tenantID).First(&l).Error; err != nil {
		return errors.New("明细行不存在")
	}
	if err := s.mutableOrErr(l.ChargeRecordID); err != nil { return err }
	if err := s.db.Delete(&l).Error; err != nil { return err }
	return s.recomputeRecordTotal(l.ChargeRecordID)
}

func (s *BillingService) recomputeRecordTotal(recID string) error {
	var lines []models.ChargeLine
	if err := s.db.Where("charge_record_id = ?", recID).Find(&lines).Error; err != nil {
		return err
	}
	return s.db.Model(&models.ChargeRecord{}).Where("id = ?", recID).
		Updates(map[string]any{"total_amount": computeTotal(lines), "updated_at": time.Now()}).Error
}
```
> `recomputeRecordTotal` 返回值在 AddLine 里被当作第二返回值——注意 AddLine 末尾应为 `if err := s.recomputeRecordTotal(recID); err != nil { return &in, err }; return &in, nil`。请按此修正三处调用以免把 nil error 误传。

- [ ] **Step 4：Run** → PASS
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 清单读取+护士增删改行（重算合计，状态门禁）"`

---

## Task 8：状态流 confirm/check/cancel + 列表查询

**Files:** Modify `billing_service.go`、`billing_service_test.go`

- [ ] **Step 1：失败测试**

```go
func TestBilling_StatusFlow(t *testing.T) {
	db := newBillingTestDB(t)
	db.Exec(`INSERT INTO "Plan_PatientPrescription" ("Id","TenantId","PatientId","TreatmentId","DialysisMethod") VALUES (5006,3,1006,9006,'HD')`)
	svc := &BillingService{db: db, tenantID: 3}
	rec, _ := svc.BuildDraft(BuildDraftInput{PatientID: 1006, TreatmentID: 9006, PrescriptionID: 5006, AccessType: "AVF"})
	svc.SaveDraft(rec)
	if _, err := svc.Confirm(rec.ID, "u1", "护士甲"); err != nil { t.Fatal(err) }
	if _, err := svc.Check(rec.ID, "u2", "护士乙"); err != nil { t.Fatal(err) }
	// checked 后不能再 confirm
	if _, err := svc.Confirm(rec.ID, "u1", "护士甲"); err == nil { t.Fatalf("should reject confirm after checked") }
	got, _ := svc.GetRecord(rec.ID)
	if got.Status != models.ChargeStatusChecked || got.CheckedName != "护士乙" { t.Fatalf("bad status %+v", got.Status) }
}
```

- [ ] **Step 2：Run** → FAIL
- [ ] **Step 3：实现 Confirm / Check / MarkPushed / Cancel / List**

```go
var billingTransitions = map[string][]string{
	models.ChargeStatusDraft:     {models.ChargeStatusConfirmed, models.ChargeStatusCancelled},
	models.ChargeStatusConfirmed: {models.ChargeStatusChecked, models.ChargeStatusDraft, models.ChargeStatusCancelled},
	models.ChargeStatusChecked:   {models.ChargeStatusPushed, models.ChargeStatusCancelled},
	models.ChargeStatusPushed:    {models.ChargeStatusSettled, models.ChargeStatusCancelled},
}

func canTransit(from, to string) bool {
	for _, t := range billingTransitions[from] {
		if t == to { return true }
	}
	return false
}

func (s *BillingService) setStatus(id, to string, fields map[string]any) (*models.ChargeRecord, error) {
	rec, err := s.GetRecord(id)
	if err != nil { return nil, err }
	if !canTransit(rec.Status, to) {
		return nil, errors.New("状态流转不合法：" + rec.Status + "→" + to)
	}
	if fields == nil { fields = map[string]any{} }
	fields["status"] = to
	fields["updated_at"] = time.Now()
	if err := s.db.Model(&models.ChargeRecord{}).Where("id = ?", id).Updates(fields).Error; err != nil {
		return nil, err
	}
	return s.GetRecord(id)
}

func (s *BillingService) Confirm(id, userID, name string) (*models.ChargeRecord, error) {
	return s.setStatus(id, models.ChargeStatusConfirmed, map[string]any{"recorded_by": userID, "recorded_name": name})
}

func (s *BillingService) Check(id, userID, name string) (*models.ChargeRecord, error) {
	now := time.Now()
	return s.setStatus(id, models.ChargeStatusChecked, map[string]any{"checked_by": userID, "checked_name": name, "checked_at": now})
}

func (s *BillingService) MarkPushed(id string) (*models.ChargeRecord, error) {
	now := time.Now()
	return s.setStatus(id, models.ChargeStatusPushed, map[string]any{"pushed_at": now})
}

func (s *BillingService) Cancel(id, reason string) (*models.ChargeRecord, error) {
	return s.setStatus(id, models.ChargeStatusCancelled, map[string]any{"note": reason})
}

// List 按患者/日期/状态过滤清单头。
func (s *BillingService) List(patientID *int64, date, status string) ([]models.ChargeRecord, error) {
	q := s.db.Where("tenant_id = ?", s.tenantID)
	if patientID != nil { q = q.Where("patient_id = ?", *patientID) }
	if status != "" { q = q.Where("status = ?", status) }
	if date != "" { q = q.Where("charge_date >= ? AND charge_date < datetime(?, '+1 day')", date, date) }
	var rows []models.ChargeRecord
	if err := q.Order("charge_date DESC").Find(&rows).Error; err != nil { return nil, err }
	return rows, nil
}
```
> ⚠️ `List` 里 `datetime(?, '+1 day')` 是 SQLite 写法；生产 PostgreSQL 需用 `charge_date < (?::date + interval '1 day')`。改为方言无关：传 `date` 时用 `charge_date >= ? AND charge_date < ?`，由 handler 传 `date` 和 `date+1天` 两个边界字符串。请按此调整签名为 `List(patientID *int64, dateFrom, dateTo, status string)`，handler 计算边界。

- [ ] **Step 4：Run** → PASS（注意按上面注释把 List 改成方言无关再跑）
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 状态流 confirm/check/push/cancel + 列表查询"`

---

## Task 9：报表（班次/日/月小结）

**Files:** Modify `billing_service.go`、`billing_service_test.go`

- [ ] **Step 1：失败测试**（班次小结：治疗数 vs 已结算数）

```go
func TestBilling_ShiftSummary(t *testing.T) {
	db := newBillingTestDB(t)
	svc := &BillingService{db: db, tenantID: 3}
	// 造 2 条清单：1 checked 1 draft，同一天
	for i, st := range []string{models.ChargeStatusChecked, models.ChargeStatusDraft} {
		db.Create(&models.ChargeRecord{ID: utils.GenerateID(), TenantID: 3, PatientID: int64(2000 + i), Status: st, TotalAmount: 419, ChargeDate: ptrTime("2026-06-23")})
	}
	sum, err := svc.DaySummary("2026-06-23", "2026-06-24")
	if err != nil { t.Fatal(err) }
	if sum.Total != 2 || sum.Checked != 1 || sum.Draft != 1 {
		t.Fatalf("summary bad: %+v", sum)
	}
}
func ptrTime(s string) *time.Time { tm, _ := time.Parse("2006-01-02", s); return &tm }
```

- [ ] **Step 2：Run** → FAIL
- [ ] **Step 3：实现 DaySummary（按状态分桶 + 参考金额合计）**

```go
type ChargeSummary struct {
	DateFrom    string  `json:"dateFrom"`
	Total       int     `json:"total"`
	Draft       int     `json:"draft"`
	Confirmed   int     `json:"confirmed"`
	Checked     int     `json:"checked"`
	Pushed      int     `json:"pushed"`
	Settled     int     `json:"settled"`
	Cancelled   int     `json:"cancelled"`
	TotalAmount float64 `json:"totalAmount"`
}

func (s *BillingService) DaySummary(dateFrom, dateTo string) (*ChargeSummary, error) {
	var rows []models.ChargeRecord
	if err := s.db.Where("tenant_id = ? AND charge_date >= ? AND charge_date < ?", s.tenantID, dateFrom, dateTo).Find(&rows).Error; err != nil {
		return nil, err
	}
	sum := &ChargeSummary{DateFrom: dateFrom}
	for _, r := range rows {
		sum.Total++
		sum.TotalAmount += r.TotalAmount
		switch r.Status {
		case models.ChargeStatusDraft: sum.Draft++
		case models.ChargeStatusConfirmed: sum.Confirmed++
		case models.ChargeStatusChecked: sum.Checked++
		case models.ChargeStatusPushed: sum.Pushed++
		case models.ChargeStatusSettled: sum.Settled++
		case models.ChargeStatusCancelled: sum.Cancelled++
		}
	}
	return sum, nil
}
```
> 月报/按项目排行/按医保类型分布属增量，v1 先给 DaySummary（班次=按天传时间边界即可复用）。月报 = 传月初/月末边界调 DaySummary + 额外按 patient_id GROUP（可后续加 MonthByPatient）。本期 v1 到 DaySummary 即可，标注"月报/按项目/按医保后续迭代"。

- [ ] **Step 4：Run** → PASS
- [ ] **Step 5：Commit** `git commit -am "feat(C4): 报表 DaySummary（班次/日状态分桶+参考金额）"`

---

## Task 10：HISPusher 预留接口 + NoopPusher 桩

**Files:** Create `internal/services/billing_pusher.go`、加测试到 `billing_service_test.go`

- [ ] **Step 1：失败测试**

```go
func TestBilling_NoopPusher(t *testing.T) {
	p := NoopPusher{}
	if p.Channel() != "noop" { t.Fatalf("channel") }
	res, err := p.Push(&models.ChargeRecord{ID: "x", Status: models.ChargeStatusChecked})
	if err != nil { t.Fatal(err) }
	if res.Accepted { t.Fatalf("noop should not claim accepted") }
	if res.Message == "" { t.Fatalf("want guidance message") }
}
```

- [ ] **Step 2：Run** → FAIL
- [ ] **Step 3：写 billing_pusher.go**（照搬 CNRDS Exporter 抽象，留真实推送给以后）

```go
package services

import "github.com/elliotxin/ai-hms-backend/internal/models"

// PushResult 推送结果。
type PushResult struct {
	Accepted bool   `json:"accepted"`
	Ref      string `json:"ref"`     // HIS 回执号（实现后填）
	Message  string `json:"message"` // 人读说明
}

// HISPusher 费用清单推送 HIS 的抽象。本期只有 Noop 桩；
// 接 HIS 时新增实现（HTTP/HL7/文件）即可，调用方不变。
type HISPusher interface {
	Channel() string
	Push(rec *models.ChargeRecord) (PushResult, error)
}

// NoopPusher 占位实现：不真正推送，提示走 Excel 导出人工录入。
type NoopPusher struct{}

func (NoopPusher) Channel() string { return "noop" }

func (NoopPusher) Push(rec *models.ChargeRecord) (PushResult, error) {
	return PushResult{
		Accepted: false,
		Message:  "HIS 推送接口待对接：请导出 Excel 由护士录入 HIS。清单已标记为已推送。",
	}, nil
}

// 默认推送器（接 HIS 后改这里或注入配置）。
var DefaultPusher HISPusher = NoopPusher{}
```

- [ ] **Step 4：Run** → PASS
- [ ] **Step 5：Commit** `git commit -am "feat(C4): HISPusher 预留接口+NoopPusher 桩"`

---

## Task 11：billing_handler 路由 + 装配

**Files:** Create `internal/api/v1/billing_handler.go`；Modify `cmd/server/main.go`

- [ ] **Step 1：写 handler**（端点：归集/读/编辑/状态/推送/报表）

```go
package v1

import (
	"strconv"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type BillingHandler struct{ svc *services.BillingService }

func RegisterBillingRoutes(rg *gin.RouterGroup) {
	h := &BillingHandler{svc: services.NewBillingService()}
	rg.POST("/charges/build", h.Build)            // 归集生成草稿并落库
	rg.GET("/charges", h.List)                    // 列表（患者/日期/状态）
	rg.GET("/charges/:id", h.Get)                 // 清单详情（头+明细）
	rg.POST("/charges/:id/lines", h.AddLine)      // 加行
	rg.PUT("/charges/lines/:lineId", h.UpdateLine)// 改行
	rg.DELETE("/charges/lines/:lineId", h.DeleteLine)
	rg.POST("/charges/:id/confirm", h.Confirm)    // 确认
	rg.POST("/charges/:id/check", h.Check)        // 核对
	rg.POST("/charges/:id/push", h.Push)          // 推送（走 NoopPusher）
	rg.POST("/charges/:id/cancel", h.Cancel)      // 取消
	rg.GET("/charges/summary", h.Summary)         // 日/班次小结
}

func (h *BillingHandler) Build(c *gin.Context) {
	var raw struct {
		PatientID      int64   `json:"patientId"`
		TreatmentID    int64   `json:"treatmentId"`
		PrescriptionID int64   `json:"prescriptionId"`
		AccessType     string  `json:"accessType"`
		CrrtHours      float64 `json:"crrtHours"`
		HasInjection   bool    `json:"hasInjection"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil { response.BadRequest(c, "请求体无效"); return }
	rec, err := h.svc.BuildDraft(services.BuildDraftInput{
		PatientID: raw.PatientID, TreatmentID: raw.TreatmentID, PrescriptionID: raw.PrescriptionID,
		AccessType: raw.AccessType, CrrtHours: raw.CrrtHours, HasInjection: raw.HasInjection,
	})
	if err != nil { response.BadRequest(c, err.Error()); return }
	if err := h.svc.SaveDraft(rec); err != nil { response.InternalErrorSafe(c); return }
	response.SuccessCreated(c, rec)
}

func (h *BillingHandler) List(c *gin.Context) {
	var pid *int64
	if v := c.Query("patientId"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil { response.BadRequest(c, "无效患者ID"); return }
		pid = &n
	}
	from, to := "", ""
	if d := c.Query("date"); d != "" {
		from = d
		if t, err := time.Parse("2006-01-02", d); err == nil { to = t.AddDate(0, 0, 1).Format("2006-01-02") }
	}
	rows, err := h.svc.List(pid, from, to, c.Query("status"))
	if err != nil { response.InternalErrorSafe(c); return }
	response.Success(c, rows)
}

func (h *BillingHandler) Get(c *gin.Context) {
	rec, err := h.svc.GetRecord(c.Param("id"))
	if err != nil { response.NotFound(c, err.Error()); return }
	response.Success(c, rec)
}

func (h *BillingHandler) AddLine(c *gin.Context) {
	var l models.ChargeLine
	if err := c.ShouldBindJSON(&l); err != nil { response.BadRequest(c, "请求体无效"); return }
	out, err := h.svc.AddLine(c.Param("id"), l)
	if err != nil { response.BadRequest(c, err.Error()); return }
	response.SuccessCreated(c, out)
}

func (h *BillingHandler) UpdateLine(c *gin.Context) {
	var p services.ChargeLinePatch
	if err := c.ShouldBindJSON(&p); err != nil { response.BadRequest(c, "请求体无效"); return }
	out, err := h.svc.UpdateLine(c.Param("lineId"), p)
	if err != nil { response.BadRequest(c, err.Error()); return }
	response.Success(c, out)
}

func (h *BillingHandler) DeleteLine(c *gin.Context) {
	if err := h.svc.DeleteLine(c.Param("lineId")); err != nil { response.BadRequest(c, err.Error()); return }
	response.Success(c, gin.H{"ok": true})
}

func (h *BillingHandler) actor(c *gin.Context) (string, string) {
	// 与既有 handler 一致地从上下文取登录用户；缺省取请求体
	uid, _ := c.Get("userID")
	uname, _ := c.Get("userName")
	s1, _ := uid.(string); s2, _ := uname.(string)
	return s1, s2
}

func (h *BillingHandler) Confirm(c *gin.Context) {
	uid, uname := h.actor(c)
	rec, err := h.svc.Confirm(c.Param("id"), uid, uname)
	if err != nil { response.BadRequest(c, err.Error()); return }
	response.Success(c, rec)
}

func (h *BillingHandler) Check(c *gin.Context) {
	uid, uname := h.actor(c)
	rec, err := h.svc.Check(c.Param("id"), uid, uname)
	if err != nil { response.BadRequest(c, err.Error()); return }
	response.Success(c, rec)
}

func (h *BillingHandler) Push(c *gin.Context) {
	rec, err := h.svc.GetRecord(c.Param("id"))
	if err != nil { response.NotFound(c, err.Error()); return }
	res, _ := services.DefaultPusher.Push(rec)
	updated, err := h.svc.MarkPushed(rec.ID)
	if err != nil { response.BadRequest(c, err.Error()); return }
	response.Success(c, gin.H{"record": updated, "push": res})
}

func (h *BillingHandler) Cancel(c *gin.Context) {
	var raw struct{ Reason string `json:"reason"` }
	_ = c.ShouldBindJSON(&raw)
	rec, err := h.svc.Cancel(c.Param("id"), raw.Reason)
	if err != nil { response.BadRequest(c, err.Error()); return }
	response.Success(c, rec)
}

func (h *BillingHandler) Summary(c *gin.Context) {
	d := c.Query("date")
	if d == "" { response.BadRequest(c, "缺少 date"); return }
	to := d
	if t, err := time.Parse("2006-01-02", d); err == nil { to = t.AddDate(0, 0, 1).Format("2006-01-02") }
	sum, err := h.svc.DaySummary(d, to)
	if err != nil { response.InternalErrorSafe(c); return }
	response.Success(c, sum)
}
```
> `actor()` 取上下文 key 名需与项目鉴权中间件一致（查 medication/consent handler 怎么取当前用户，照抄 key）。若项目从 JWT claims 取，照同款。

- [ ] **Step 2：main.go 注册**（在 RegisterConsentRoutes 行后）

```go
	// 收费归集路由（C4）
	v1api.RegisterBillingRoutes(protected)
```

- [ ] **Step 3：Run** `go build ./... && go vet ./... && go test ./internal/... -run 'TestBilling|TestLoadBillingCatalog'` → 全 PASS
- [ ] **Step 4：Commit** `git commit -am "feat(C4): billing handler 11端点 + main 装配"`

---

## Task 12：前端 billingApi.ts

**Files:** Create `src/services/billingApi.ts`

- [ ] **Step 1：写 API 客户端**（仿 consentApi.ts 信封 `res.data?.data`）

```ts
// 收费归集 API（C4）。/api/v1/charges，走主系统 v1 信封。
import { apiClient } from './restClient'

export type ChargeStatus = 'draft' | 'confirmed' | 'checked' | 'pushed' | 'settled' | 'cancelled'
export type ChargeCategory = 'treatment' | 'material' | 'nursing' | 'injection' | 'drug'

export interface ChargeLine {
  id: string
  chargeRecordId: string
  category: ChargeCategory
  itemCode?: string
  itemName: string
  spec?: string
  unit?: string
  quantity: number
  unitPrice?: number | null
  amount?: number | null
  billable: boolean
  source: 'auto' | 'manual'
  chargeItemId?: number
  note?: string
}

export interface ChargeRecord {
  id: string
  patientId: number
  treatmentId: number
  prescriptionId: number
  chargeDate?: string
  shift?: string
  dialysisMode?: string
  accessType?: string
  crrtHours?: number
  totalAmount: number
  status: ChargeStatus
  recordedName?: string
  checkedName?: string
  checkedAt?: string
  pushedAt?: string
  note?: string
  lines?: ChargeLine[]
}

export interface ChargeSummary {
  dateFrom: string; total: number; draft: number; confirmed: number
  checked: number; pushed: number; settled: number; cancelled: number; totalAmount: number
}

const base = '/api/v1'

export async function buildCharge(body: {
  patientId: number; treatmentId: number; prescriptionId: number
  accessType?: string; crrtHours?: number; hasInjection?: boolean
}): Promise<ChargeRecord> {
  const res = await apiClient.post(`${base}/charges/build`, body)
  return res.data?.data
}
export async function listCharges(params: { patientId?: number; date?: string; status?: string } = {}): Promise<ChargeRecord[]> {
  const res = await apiClient.get(`${base}/charges`, { params })
  return res.data?.data ?? []
}
export async function getCharge(id: string): Promise<ChargeRecord> {
  const res = await apiClient.get(`${base}/charges/${id}`)
  return res.data?.data
}
export async function addChargeLine(id: string, line: Partial<ChargeLine>): Promise<ChargeLine> {
  const res = await apiClient.post(`${base}/charges/${id}/lines`, line)
  return res.data?.data
}
export async function updateChargeLine(lineId: string, patch: Partial<ChargeLine>): Promise<ChargeLine> {
  const res = await apiClient.put(`${base}/charges/lines/${lineId}`, patch)
  return res.data?.data
}
export async function deleteChargeLine(lineId: string): Promise<void> {
  await apiClient.delete(`${base}/charges/lines/${lineId}`)
}
export async function confirmCharge(id: string): Promise<ChargeRecord> {
  const res = await apiClient.post(`${base}/charges/${id}/confirm`, {})
  return res.data?.data
}
export async function checkCharge(id: string): Promise<ChargeRecord> {
  const res = await apiClient.post(`${base}/charges/${id}/check`, {})
  return res.data?.data
}
export async function pushCharge(id: string): Promise<{ record: ChargeRecord; push: { accepted: boolean; message: string } }> {
  const res = await apiClient.post(`${base}/charges/${id}/push`, {})
  return res.data?.data
}
export async function cancelCharge(id: string, reason: string): Promise<ChargeRecord> {
  const res = await apiClient.post(`${base}/charges/${id}/cancel`, { reason })
  return res.data?.data
}
export async function getChargeSummary(date: string): Promise<ChargeSummary> {
  const res = await apiClient.get(`${base}/charges/summary`, { params: { date } })
  return res.data?.data
}
```

- [ ] **Step 2：Run** `cd ai-hms-frontend && npx tsc --noEmit` → exit 0
- [ ] **Step 3：Commit** `git commit -am "feat(C4): 前端 billingApi 客户端"`

---

## Task 13：护士友好 Excel 导出 billingExcel.ts

**Files:** Create `src/lib/billingExcel.ts`

设计：一张清单导出一个 sheet，**护士能直接读、整列复制到 HIS**。列顺序固定、有合计行、不可计费项灰显标注。

- [ ] **Step 1：写导出器**（house 风格 `json_to_sheet` + `!cols`）

```ts
import * as XLSX from 'xlsx'
import type { ChargeRecord, ChargeLine } from '@/services/billingApi'

const CAT_LABEL: Record<string, string> = {
  treatment: 'A 治疗费', material: 'B 耗材费', nursing: 'C 护理费',
  injection: 'D 注射费', drug: 'E 药品（HIS查价）',
}

function lineRow(l: ChargeLine) {
  return {
    '类别': CAT_LABEL[l.category] || l.category,
    '项目名称': l.itemName,
    '规格': l.spec || '',
    '医保编码': l.itemCode || '',
    '数量': l.quantity,
    '单位': l.unit || '',
    '参考单价': l.unitPrice ?? '',
    '参考金额': l.amount ?? '',
    '是否计费': l.billable ? '是' : '否（不计费）',
    '来源': l.source === 'manual' ? '手工' : '自动',
    '备注': l.note || '',
  }
}

/** 导出单张清单为护士友好 xlsx。 */
export function exportChargeSheet(rec: ChargeRecord, patientName: string) {
  const lines = rec.lines || []
  const data = lines.map(lineRow)
  // 合计行（只计 billable+amount）
  const total = lines.reduce((s, l) => s + (l.billable && typeof l.amount === 'number' ? l.amount : 0), 0)
  data.push({
    '类别': '', '项目名称': '参考总价（仅供核对，实际以HIS为准）', '规格': '', '医保编码': '',
    '数量': '', '单位': '', '参考单价': '', '参考金额': total,
    '是否计费': '', '来源': '', '备注': '',
  } as any)

  const ws = XLSX.utils.json_to_sheet(data)
  ws['!cols'] = [
    { wch: 14 }, // 类别
    { wch: 34 }, // 项目名称
    { wch: 14 }, // 规格
    { wch: 18 }, // 医保编码
    { wch: 6 },  // 数量
    { wch: 6 },  // 单位
    { wch: 10 }, // 参考单价
    { wch: 10 }, // 参考金额
    { wch: 14 }, // 是否计费
    { wch: 8 },  // 来源
    { wch: 24 }, // 备注
  ]
  const wb = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(wb, ws, '收费清单')
  const date = (rec.chargeDate || '').slice(0, 10) || new Date().toLocaleDateString('zh-CN')
  XLSX.writeFile(wb, `收费清单_${patientName}_${rec.dialysisMode || ''}_${date}.xlsx`)
}

/** 批量导出多张清单（每张一个 sheet），用于班次/日导出。 */
export function exportChargeBatch(records: { rec: ChargeRecord; patientName: string }[], fileLabel: string) {
  const wb = XLSX.utils.book_new()
  records.forEach(({ rec, patientName }, idx) => {
    const data = (rec.lines || []).map(lineRow)
    const ws = XLSX.utils.json_to_sheet(data)
    ws['!cols'] = [{ wch: 14 }, { wch: 34 }, { wch: 14 }, { wch: 18 }, { wch: 6 }, { wch: 6 }, { wch: 10 }, { wch: 10 }, { wch: 14 }, { wch: 8 }, { wch: 24 }]
    const name = `${idx + 1}_${patientName}`.slice(0, 28).replace(/[\\/?*\[\]:]/g, '_')
    XLSX.utils.book_append_sheet(wb, ws, name)
  })
  XLSX.writeFile(wb, `收费批量_${fileLabel}.xlsx`)
}
```

- [ ] **Step 2：Run** `npx tsc --noEmit` → exit 0
- [ ] **Step 3：Commit** `git commit -am "feat(C4): 护士友好 Excel 导出（单张+批量，xlsx house风格）"`

---

## Task 14：患者「收费归集」Tab — BillingTab.tsx

**Files:** Create `src/pages/patient-detail/tabs/BillingTab.tsx`

UI：顶部状态条 + 「归集生成清单」按钮 → 清单卡（5 类分组、参考价、不可计费灰显、行内增删改）→ 底部参考总价 + 确认/核对/导出Excel/推送 按钮。

- [ ] **Step 1：写组件**（仿 ConsentTab：SectionHeader/DetailCard + antd + lucide）

```tsx
import { useState, useEffect, useCallback } from 'react'
import { message, Modal, Button, InputNumber, Input, Tag, Space, Popconfirm } from 'antd'
import { Receipt, Plus, Loader2, FileSpreadsheet, CheckCircle2, UserCheck, Send, Ban, Trash2 } from 'lucide-react'
import { SectionHeader, DetailCard } from '@/components/ui'
import {
  buildCharge, listCharges, getCharge, addChargeLine, updateChargeLine, deleteChargeLine,
  confirmCharge, checkCharge, pushCharge, cancelCharge,
  type ChargeRecord, type ChargeLine, type ChargeStatus,
} from '@/services/billingApi'
import { exportChargeSheet } from '@/lib/billingExcel'
import type { Patient } from '@/types/original'

const STATUS_META: Record<ChargeStatus, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' }, confirmed: { label: '已确认', color: 'blue' },
  checked: { label: '已核对', color: 'cyan' }, pushed: { label: '已推送', color: 'processing' },
  settled: { label: '已结算', color: 'success' }, cancelled: { label: '已取消', color: 'error' },
}
const CAT_LABEL: Record<string, string> = {
  treatment: 'A 治疗费', material: 'B 耗材费', nursing: 'C 护理费', injection: 'D 注射费', drug: 'E 药品',
}

interface Props { patient: Patient }

export default function BillingTab({ patient }: Props) {
  const [records, setRecords] = useState<ChargeRecord[]>([])
  const [active, setActive] = useState<ChargeRecord | null>(null)
  const [loading, setLoading] = useState(false)
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    if (!patient.id) return
    setLoading(true)
    try {
      const rows = await listCharges({ patientId: Number(patient.id) })
      setRecords(rows)
      if (rows[0]) setActive(await getCharge(rows[0].id))
      else setActive(null)
    } catch { message.error('加载收费清单失败') } finally { setLoading(false) }
  }, [patient.id])

  useEffect(() => { load() }, [load])

  const reloadActive = useCallback(async (id: string) => { setActive(await getCharge(id)) }, [])

  async function handleBuild() {
    // 需当前治疗/处方上下文；此处取患者最近一次治疗（实际接 PatientDetail 上下文或让用户选）
    const treatmentId = Number((patient as any).currentTreatmentId || 0)
    const prescriptionId = Number((patient as any).currentPrescriptionId || 0)
    if (!treatmentId || !prescriptionId) { message.warning('未找到本次治疗/处方，请在治疗执行页发起归集'); return }
    setBusy(true)
    try {
      const accessType = (patient.vascularAccess?.type || '').toUpperCase()
      const rec = await buildCharge({ patientId: Number(patient.id), treatmentId, prescriptionId, accessType, hasInjection: false })
      message.success('已归集生成清单草稿')
      await load(); setActive(await getCharge(rec.id))
    } catch (e: any) { message.error(e?.response?.data?.error?.message || '归集失败') } finally { setBusy(false) }
  }

  const editable = active && (active.status === 'draft' || active.status === 'confirmed')

  async function patchQty(line: ChargeLine, qty: number) {
    await updateChargeLine(line.id, { quantity: qty }); await reloadActive(active!.id)
  }
  async function delLine(line: ChargeLine) {
    await deleteChargeLine(line.id); await reloadActive(active!.id)
  }
  async function addManual() {
    let name = '', price = 0, qty = 1
    Modal.confirm({
      title: '新增收费项',
      content: (
        <div className="space-y-2 pt-2">
          <Input placeholder="项目名称" onChange={e => (name = e.target.value)} />
          <InputNumber placeholder="参考单价" style={{ width: '100%' }} onChange={v => (price = Number(v) || 0)} />
          <InputNumber placeholder="数量" defaultValue={1} style={{ width: '100%' }} onChange={v => (qty = Number(v) || 1)} />
        </div>
      ),
      onOk: async () => {
        if (!name) { message.error('请输入名称'); throw new Error() }
        await addChargeLine(active!.id, { category: 'material', itemName: name, unitPrice: price, quantity: qty, billable: true })
        await reloadActive(active!.id)
      },
    })
  }

  async function doStatus(fn: () => Promise<ChargeRecord>, ok: string) {
    setBusy(true)
    try { await fn(); message.success(ok); await reloadActive(active!.id); await load() }
    catch (e: any) { message.error(e?.response?.data?.error?.message || '操作失败') } finally { setBusy(false) }
  }

  return (
    <div className="space-y-5 animate-fade-in pb-10">
      <div className="flex items-center justify-between">
        <SectionHeader icon={Receipt} title="收费归集" />
        <Button type="primary" icon={<Plus size={16} />} loading={busy} onClick={handleBuild}>归集生成清单</Button>
      </div>

      {loading ? (
        <div className="py-12 text-center text-slate-400"><Loader2 size={20} className="inline animate-spin" /> 加载中…</div>
      ) : !active ? (
        <div className="py-10 text-center text-slate-400 text-[13px]">暂无收费清单。点击右上「归集生成清单」从本次治疗自动归集费用项。</div>
      ) : (
        <DetailCard>
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <span className="font-bold text-slate-800">{active.dialysisMode} 清单</span>
              <Tag color={STATUS_META[active.status].color}>{STATUS_META[active.status].label}</Tag>
              <span className="text-[12px] text-slate-400">{(active.chargeDate || '').slice(0, 16).replace('T', ' ')}</span>
            </div>
            <div className="text-right">
              <span className="text-[12px] text-slate-400">参考总价</span>
              <span className="ml-2 text-lg font-black text-rose-600">¥{active.totalAmount?.toFixed(2)}</span>
            </div>
          </div>

          <table className="w-full text-[13px]">
            <thead>
              <tr className="text-slate-400 border-b">
                <th className="text-left py-1.5 font-medium">类别</th>
                <th className="text-left font-medium">项目</th>
                <th className="text-right font-medium">数量</th>
                <th className="text-right font-medium">参考单价</th>
                <th className="text-right font-medium">参考金额</th>
                <th className="text-right font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {(active.lines || []).map(l => (
                <tr key={l.id} className={`border-b border-slate-50 ${l.billable ? '' : 'text-slate-300'}`}>
                  <td className="py-1.5">{CAT_LABEL[l.category]}</td>
                  <td>
                    {l.itemName}{l.spec ? ` · ${l.spec}` : ''}
                    {!l.billable && <Tag className="ml-1 text-[10px]">不计费</Tag>}
                    {l.unitPrice == null && l.billable && l.category !== 'drug' && <Tag color="orange" className="ml-1 text-[10px]">待核价</Tag>}
                    {l.source === 'manual' && <Tag color="purple" className="ml-1 text-[10px]">手工</Tag>}
                  </td>
                  <td className="text-right">
                    {editable
                      ? <InputNumber size="small" min={0} value={l.quantity} onChange={v => patchQty(l, Number(v) || 0)} style={{ width: 64 }} />
                      : l.quantity}
                  </td>
                  <td className="text-right">{l.unitPrice == null ? '—' : `¥${l.unitPrice}`}</td>
                  <td className="text-right">{l.amount == null ? '—' : `¥${l.amount.toFixed(2)}`}</td>
                  <td className="text-right">
                    {editable && (
                      <Popconfirm title="删除该项？" onConfirm={() => delLine(l)}>
                        <Button type="text" size="small" danger icon={<Trash2 size={14} />} />
                      </Popconfirm>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="flex items-center gap-2 mt-4 flex-wrap">
            {editable && <Button size="small" icon={<Plus size={14} />} onClick={addManual}>加项</Button>}
            <div className="flex-1" />
            <Button size="small" icon={<FileSpreadsheet size={14} />} onClick={() => exportChargeSheet(active, patient.name)}>导出 Excel</Button>
            {active.status === 'draft' && <Button size="small" type="primary" icon={<CheckCircle2 size={14} />} loading={busy} onClick={() => doStatus(() => confirmCharge(active.id), '已确认')}>确认</Button>}
            {active.status === 'confirmed' && <Button size="small" type="primary" icon={<UserCheck size={14} />} loading={busy} onClick={() => doStatus(() => checkCharge(active.id), '已核对')}>核对</Button>}
            {active.status === 'checked' && <Button size="small" type="primary" icon={<Send size={14} />} loading={busy} onClick={() => doStatus(async () => { const r = await pushCharge(active.id); message.info(r.push.message); return r.record }, '已标记推送')}>推送 HIS</Button>}
            {active.status !== 'cancelled' && active.status !== 'settled' && (
              <Popconfirm title="取消该清单？" onConfirm={() => doStatus(() => cancelCharge(active.id, '护士取消'), '已取消')}>
                <Button size="small" danger icon={<Ban size={14} />}>取消</Button>
              </Popconfirm>
            )}
          </div>
        </DetailCard>
      )}
    </div>
  )
}
```
> 归集需要"本次治疗/处方"上下文。`patient.currentTreatmentId/currentPrescriptionId` 若 Patient 类型没有，则二选一：①在 PatientDetail 载入时补这两个字段；②本 Tab 仅展示已生成清单，把"归集生成"动作放到透析执行流页（更贴合规则"治疗启动后归集"）。v1 建议：Tab 先做展示+编辑+导出+状态，"归集生成"按钮在拿不到上下文时给出引导提示（已实现）。

- [ ] **Step 2：Run** `npx tsc --noEmit` → exit 0
- [ ] **Step 3：Commit** `git commit -am "feat(C4): 患者收费归集 Tab（清单/增删改/状态/导出）"`

---

## Task 15：PatientDetail 装配（菜单 + switch + barrel）

**Files:** Modify `src/pages/patient-detail/tabs/index.ts`、`src/pages/PatientDetail.tsx`

- [ ] **Step 1：barrel 导出**（index.ts 末尾）

```ts
export { default as BillingTab } from './BillingTab'
```

- [ ] **Step 2：PatientDetail.tsx — import 加 BillingTab**（与其它 tabs 同一 import 块）
```tsx
  NursingTab, ConsentTab, BillingTab, HistoryTab, MonthlySummaryTab,
```
- [ ] **Step 3：lucide import 加 Receipt**（第 5 行图标 import 追加 `Receipt`）
- [ ] **Step 4：SectionID 加 'billing'**
```tsx
  | 'consent'         // 知情同意
  | 'billing'         // 收费归集
```
- [ ] **Step 5：MENU_ITEMS 加项**（consent 行后）
```tsx
  { key: 'billing',         label: '收费归集',     icon: <Receipt size={18} /> },
```
- [ ] **Step 6：renderContent switch 加分支**（consent case 后）
```tsx
      case 'billing':
        return <BillingTab patient={patient} />
```
- [ ] **Step 7：Run** `npx tsc --noEmit` → exit 0
- [ ] **Step 8：Commit** `git commit -am "feat(C4): 收费归集 Tab 挂载患者详情侧栏"`

---

## Task 16（Phase 2）：独立「收费管理」页 + 报表 + 批量导出

**Files:** Create `src/pages/BillingPage.tsx`；改路由表 + 一级菜单（按项目既有 admin 页注册方式照做）

- [ ] **Step 1**：页面 = 顶部日期选择 + `getChargeSummary(date)` 状态分桶卡（治疗总数/已核对/未完成/已推送）+ `listCharges({date})` 表格（患者/模式/状态/参考金额/操作）+ 「批量导出 Excel」按钮（对每条 `getCharge` 后 `exportChargeBatch`）。
- [ ] **Step 2**：路由 + 一级菜单项「收费管理」按项目现有管理页（如 CNRDS 报表页 `/cnrds-report`）的注册方式照搬：找到 router 配置与侧栏菜单数组，各加一行 `/billing` → BillingPage。
- [ ] **Step 3：Run** `npx tsc --noEmit` → exit 0；**Commit** `git commit -am "feat(C4): 独立收费管理页（报表+批量导出）"`

> 班次小结 = 同一 DaySummary 接口按"今天"调用即可；月报/按项目排行/按医保分布列为 C4 v2 增量（后端补 MonthByPatient/ByItem 聚合，前端加 Tab）。

---

## 回归验证（每阶段末 + 总收尾）

```
cd ai-hms-backend  && go build ./... && go vet ./... && go test ./internal/...
cd ai-hms-frontend && npx tsc --noEmit
```
全绿后按既有交付模式（git bundle + format-patch + README_移交说明 + 本方案 + 规则 md）打包 `feat/billing-c4`，基线注明 5d66b8d。**无 push 权限，补丁移交团队**（见 [[deliver-to-github-cloud]]）。

## 自检对照规则（覆盖确认）
- 五类费用项 A/B/C/D/E → Task 3-6 ✅
- CRRT 按时长 + 加收手动勾 → Task 3（perHour + surcharge billable=false 占位）✅
- 不可计费 10 类排除 → Task 1 nonBillable + Task 4 billable=false ✅
- 护理费按通路 AVF17/TCC54 → Task 5 ✅
- 注射费 3 元最多 1 项 → Task 3 ✅
- 药品列名不带价 → Task 6 ✅
- 护士增删改 → Task 7 ✅
- 状态机 draft→confirmed→checked→pushed→settled/cancelled + 双人核对（不强制≠）→ Task 8 ✅
- 报表班次/日 → Task 9（月报标 v2）✅
- HIS 推送预留 → Task 10 NoopPusher ✅
- 两入口（患者Tab+独立页）→ Task 14/15 + Task 16 ✅
- Excel 护士友好导出 → Task 13 ✅
- 参考价单一来源 catalog（老库 Price=0 不用）→ Task 1 ✅
```
