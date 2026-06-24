package models

import "time"

type HisPriceItem struct {
	ID              string     `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	SourceSystem    string     `gorm:"column:source_system;type:varchar(32);not null;default:HIS_ORACLE" json:"sourceSystem"`
	ItemClass       *string    `gorm:"column:item_class;type:varchar(1)" json:"itemClass"`
	ItemCode        string     `gorm:"column:item_code;type:varchar(20);not null" json:"itemCode"`
	ItemName        *string    `gorm:"column:item_name;type:varchar(120)" json:"itemName"`
	ItemSpec        *string    `gorm:"column:item_spec;type:varchar(50)" json:"itemSpec"`
	Units           *string    `gorm:"column:units;type:varchar(30)" json:"units"`
	Price           *float64   `gorm:"column:price;type:decimal(9,3)" json:"price"`
	PreferPrice     *float64   `gorm:"column:prefer_price;type:decimal(9,3)" json:"preferPrice"`
	ForeignerPrice  *float64   `gorm:"column:foreigner_price;type:decimal(9,3)" json:"foreignerPrice"`
	PerformedBy     *string    `gorm:"column:performed_by;type:varchar(8)" json:"performedBy"`
	FeeTypeMask     *int       `gorm:"column:fee_type_mask" json:"feeTypeMask"`
	ClassOnInpRcpt  *string    `gorm:"column:class_on_inp_rcpt;type:varchar(1)" json:"classOnInpRcpt"`
	ClassOnOutpRcpt *string    `gorm:"column:class_on_outp_rcpt;type:varchar(1)" json:"classOnOutpRcpt"`
	ClassOnReckoning *string   `gorm:"column:class_on_reckoning;type:varchar(10)" json:"classOnReckoning"`
	SubjCode        *string    `gorm:"column:subj_code;type:varchar(10)" json:"subjCode"`
	ClassOnMr       *string    `gorm:"column:class_on_mr;type:varchar(4)" json:"classOnMr"`
	Memo            *string    `gorm:"column:memo;type:varchar(100)" json:"memo"`
	StartDate       *time.Time `gorm:"column:start_date" json:"startDate"`
	StopDate        *time.Time `gorm:"column:stop_date" json:"stopDate"`
	OperatorCode    *string    `gorm:"column:operator_code;type:varchar(8)" json:"operatorCode"`
	EnterDate       *time.Time `gorm:"column:enter_date" json:"enterDate"`
	HighPrice       *float64   `gorm:"column:high_price;type:decimal(10,4)" json:"highPrice"`
	MaterialCode    *string    `gorm:"column:material_code;type:varchar(20)" json:"materialCode"`
	Score1          *float64   `gorm:"column:score_1;type:decimal(10,2)" json:"score1"`
	Score2          *float64   `gorm:"column:score_2;type:decimal(10,2)" json:"score2"`
	PriceNameCode   *string    `gorm:"column:price_name_code;type:varchar(20)" json:"priceNameCode"`
	ControlFlag     *string    `gorm:"column:control_flag;type:varchar(1)" json:"controlFlag"`
	InputCode       *string    `gorm:"column:input_code;type:varchar(100)" json:"inputCode"`
	InputCodeWb     *string    `gorm:"column:input_code_wb;type:varchar(100)" json:"inputCodeWb"`
	StdCode1        *string    `gorm:"column:std_code_1;type:varchar(20)" json:"stdCode1"`
	ChangedMemo     *string    `gorm:"column:changed_memo;type:varchar(40)" json:"changedMemo"`
	ClassOnInsurMr  *string    `gorm:"column:class_on_insur_mr;type:varchar(24)" json:"classOnInsurMr"`
	PackageSpec     *string    `gorm:"column:package_spec;type:varchar(20)" json:"packageSpec"`
	FirmID          *string    `gorm:"column:firm_id;type:varchar(10)" json:"firmId"`
	ChargeAccording *string    `gorm:"column:charge_according;type:varchar(23)" json:"chargeAccording"`
	LicenseID       *string    `gorm:"column:license_id;type:varchar(20)" json:"licenseId"`
	UpdateFlag      *float64   `gorm:"column:update_flag;type:decimal" json:"updateFlag"`
	DeptName        *string    `gorm:"column:dept_name;type:varchar(100)" json:"deptName"`
	UpdateFlagSyb   *float64   `gorm:"column:update_flag_syb;type:decimal" json:"updateFlagSyb"`
	MrBillClass     *string    `gorm:"column:mr_bill_class;type:varchar(4)" json:"mrBillClass"`
	ClassOnMrAdd    *string    `gorm:"column:class_on_mr_add;type:varchar(4)" json:"classOnMrAdd"`
	CwtjCode        *string    `gorm:"column:cwtj_code;type:varchar(20)" json:"cwtjCode"`
	HighValue       *float64   `gorm:"column:high_value;type:decimal(9,3)" json:"highValue"`
	DrgCode         *string    `gorm:"column:drg_code;type:varchar(8)" json:"drgCode"`
	InsurUpdate     *int       `gorm:"column:insur_update" json:"insurUpdate"`
	StopOperator    *string    `gorm:"column:stop_operator;type:varchar(8)" json:"stopOperator"`
	LimitQuantity   *float64   `gorm:"column:limit_quantity;type:decimal(10,0)" json:"limitQuantity"`
	IsActive        bool       `gorm:"column:is_active;not null;default:true" json:"isActive"`
	SyncedAt        time.Time  `gorm:"column:synced_at;not null;default:now()" json:"syncedAt"`
	SyncRunID       *string    `gorm:"column:sync_run_id;type:varchar(36)" json:"syncRunId"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (HisPriceItem) TableName() string {
	return "his_price_item"
}
