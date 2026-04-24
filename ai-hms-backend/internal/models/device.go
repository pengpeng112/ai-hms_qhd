package models

import "time"

const (
	DeviceStatusNormal      = "normal"
	DeviceStatusWarning     = "warning"
	DeviceStatusAlarm       = "alarm"
	DeviceStatusOffline     = "offline"
	DeviceStatusMaintenance = "maintenance"
)

// Device is a DTO assembled from legacy equipment archive + bed binding tables.
type Device struct {
	ID               string     `json:"id"`
	TenantId         int64      `json:"tenantId"`
	Name             string     `json:"name"`
	IDNo             string     `json:"idNo"`
	SerialNo         string     `json:"serialNo"`
	Brand            string     `json:"brand"`
	Model            string     `json:"model"`
	DialysisMethod   string     `json:"dialysisMethod"`
	DeviceType       string     `json:"deviceType"`
	Manufacturer     string     `json:"manufacturer"`
	BedNumber        string     `json:"bedNumber"`
	BedId            *int64     `json:"bedId"`
	WardId           *int64     `json:"wardId"`
	WardName         string     `json:"wardName"`
	Status           string     `json:"status"`
	PurchaseDate     *time.Time `json:"purchaseDate"`
	ManufactureDate  *time.Time `json:"manufactureDate"`
	InstallDate      *time.Time `json:"installDate"`
	LastMaintained   *time.Time `json:"lastMaintained"`
	Maintenance      *int64     `json:"maintenance"`
	MaintenanceCycle string     `json:"maintenanceCycle"`
	Flux             string     `json:"flux"`
	Notes            string     `json:"notes"`
	IsDisabled       bool       `json:"isDisabled"`
	CreatorId        int64      `json:"creatorId"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
