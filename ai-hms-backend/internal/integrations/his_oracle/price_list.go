package his_oracle

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type PriceListRow struct {
	ItemClass        *string
	ItemCode         string
	ItemName         *string
	ItemSpec         *string
	Units            *string
	Price            *float64
	PreferPrice      *float64
	ForeignerPrice   *float64
	PerformedBy      *string
	FeeTypeMask      *int
	ClassOnInpRcpt   *string
	ClassOnOutpRcpt  *string
	ClassOnReckoning *string
	SubjCode         *string
	ClassOnMr        *string
	Memo             *string
	StartDate        *time.Time
	StopDate         *time.Time
	OperatorCode     *string
	EnterDate        *time.Time
	HighPrice        *float64
	MaterialCode     *string
	Score1           *float64
	Score2           *float64
	PriceNameCode    *string
	ControlFlag      *string
	InputCode        *string
	InputCodeWb      *string
	StdCode1         *string
	ChangedMemo      *string
	ClassOnInsurMr   *string
	PackageSpec      *string
	FirmID           *string
	ChargeAccording  *string
	LicenseID        *string
	UpdateFlag       *float64
	DeptName         *string
	UpdateFlagSyb    *float64
	MrBillClass      *string
	ClassOnMrAdd     *string
	CwtjCode         *string
	HighValue        *float64
	DrgCode          *string
	InsurUpdate      *int
	StopOperator     *string
	LimitQuantity    *float64
}

const priceListQuerySQL = `
SELECT
    item_class,
    item_code,
    item_name,
    item_spec,
    units,
    price,
    prefer_price,
    foreigner_price,
    performed_by,
    fee_type_mask,
    class_on_inp_rcpt,
    class_on_outp_rcpt,
    class_on_reckoning,
    subj_code,
    class_on_mr,
    memo,
    start_date,
    stop_date,
    operator,
    enter_date,
    high_price,
    material_code,
    score_1,
    score_2,
    price_name_code,
    control_flag,
    input_code,
    input_code_wb,
    std_code_1,
    changed_memo,
    class_on_insur_mr,
    package_spec,
    firm_id,
    charge_according,
    license_id,
    update_flag,
    dept_name,
    update_flag_syb,
    mr_bill_class,
    class_on_mr_add,
    cwtj_code,
    high_value,
    drg_code,
    insur_update,
    stop_operator,
    limit_quantity
FROM his.price_list
ORDER BY item_code
`

const priceListCountSQL = `SELECT COUNT(*) FROM his.price_list`

type QueryPriceListParams struct {
	Offset int
	Limit  int
}

func scanPriceRow(scanner interface {
	Scan(dest ...interface{}) error
}, r *PriceListRow) error {
	return scanner.Scan(
		&r.ItemClass, &r.ItemCode, &r.ItemName, &r.ItemSpec, &r.Units,
		&r.Price, &r.PreferPrice, &r.ForeignerPrice,
		&r.PerformedBy, &r.FeeTypeMask, &r.ClassOnInpRcpt, &r.ClassOnOutpRcpt, &r.ClassOnReckoning,
		&r.SubjCode, &r.ClassOnMr, &r.Memo,
		&r.StartDate, &r.StopDate, &r.OperatorCode, &r.EnterDate,
		&r.HighPrice, &r.MaterialCode, &r.Score1, &r.Score2,
		&r.PriceNameCode, &r.ControlFlag, &r.InputCode, &r.InputCodeWb, &r.StdCode1,
		&r.ChangedMemo, &r.ClassOnInsurMr, &r.PackageSpec, &r.FirmID, &r.ChargeAccording, &r.LicenseID,
		&r.UpdateFlag, &r.DeptName, &r.UpdateFlagSyb, &r.MrBillClass, &r.ClassOnMrAdd,
		&r.CwtjCode, &r.HighValue, &r.DrgCode, &r.InsurUpdate, &r.StopOperator, &r.LimitQuantity,
	)
}

func (c *Client) QueryPriceList(ctx context.Context, params QueryPriceListParams) ([]PriceListRow, error) {
	if c.db == nil {
		return nil, fmt.Errorf("HIS Oracle client not initialized")
	}
	if params.Limit <= 0 {
		params.Limit = 1000
	}

	sql := fmt.Sprintf(`SELECT * FROM (
		SELECT inner_q.*, ROWNUM rn FROM (
			%s
		) inner_q WHERE ROWNUM <= :2
	) WHERE rn > :1`, priceListQuerySQL)

	rows, err := c.db.QueryContext(ctx, sql, params.Offset, params.Offset+params.Limit)
	if err != nil {
		return nil, fmt.Errorf("HIS Oracle price_list query failed: %w", err)
	}
	defer rows.Close()

	var results []PriceListRow
	for rows.Next() {
		var r PriceListRow
		var rnDummy int
		if err := rows.Scan(
			&r.ItemClass, &r.ItemCode, &r.ItemName, &r.ItemSpec, &r.Units,
			&r.Price, &r.PreferPrice, &r.ForeignerPrice,
			&r.PerformedBy, &r.FeeTypeMask, &r.ClassOnInpRcpt, &r.ClassOnOutpRcpt, &r.ClassOnReckoning,
			&r.SubjCode, &r.ClassOnMr, &r.Memo,
			&r.StartDate, &r.StopDate, &r.OperatorCode, &r.EnterDate,
			&r.HighPrice, &r.MaterialCode, &r.Score1, &r.Score2,
			&r.PriceNameCode, &r.ControlFlag, &r.InputCode, &r.InputCodeWb, &r.StdCode1,
			&r.ChangedMemo, &r.ClassOnInsurMr, &r.PackageSpec, &r.FirmID, &r.ChargeAccording, &r.LicenseID,
			&r.UpdateFlag, &r.DeptName, &r.UpdateFlagSyb, &r.MrBillClass, &r.ClassOnMrAdd,
			&r.CwtjCode, &r.HighValue, &r.DrgCode, &r.InsurUpdate, &r.StopOperator, &r.LimitQuantity,
			&rnDummy,
		); err != nil {
			return results, fmt.Errorf("HIS Oracle price_list row scan failed: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return results, fmt.Errorf("HIS Oracle price_list rows error: %w", err)
	}
	return results, nil
}

func (c *Client) PriceListCount(ctx context.Context) (int, error) {
	if c.db == nil {
		return 0, fmt.Errorf("HIS Oracle client not initialized")
	}
	var cnt int
	if err := c.db.QueryRowContext(ctx, priceListCountSQL).Scan(&cnt); err != nil {
		return 0, fmt.Errorf("HIS Oracle price_list count failed: %w", err)
	}
	return cnt, nil
}

func (c *Client) QueryPriceListByClass(ctx context.Context, itemClass string) ([]PriceListRow, error) {
	if c.db == nil {
		return nil, fmt.Errorf("HIS Oracle client not initialized")
	}
	sql := strings.Replace(priceListQuerySQL, "FROM his.price_list", "FROM his.price_list WHERE item_class = :1", 1)
	rows, err := c.db.QueryContext(ctx, sql, itemClass)
	if err != nil {
		return nil, fmt.Errorf("HIS Oracle price_list query failed: %w", err)
	}
	defer rows.Close()

	var results []PriceListRow
	for rows.Next() {
		var r PriceListRow
		if err := scanPriceRow(rows, &r); err != nil {
			return results, fmt.Errorf("HIS Oracle price_list row scan failed: %w", err)
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return results, fmt.Errorf("HIS Oracle price_list rows error: %w", err)
	}
	return results, nil
}
