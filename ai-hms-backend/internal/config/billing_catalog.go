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
	BillType      string  `json:"billType"`
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

func (c *BillingCatalog) TreatmentFeeFor(mode string) (TreatmentFee, bool) {
	t, ok := c.treatmentByMode[strings.ToUpper(strings.TrimSpace(mode))]
	return t, ok
}

func (c *BillingCatalog) MaterialFor(chargeItemID int64) (MaterialPrice, bool) {
	m, ok := c.materialByID[chargeItemID]
	return m, ok
}

func (c *BillingCatalog) NursingFeeFor(accessType string) []NursingFee {
	return c.NursingFees[strings.ToUpper(strings.TrimSpace(accessType))]
}
