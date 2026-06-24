package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elliotxin/ai-hms-backend/config"
	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"gorm.io/gorm"
)

type HisPriceSyncService struct {
	db     *gorm.DB
	oracle *his_oracle.Client
}

func NewHisPriceSyncService(oracleCfg config.HisOracleConfig) (*HisPriceSyncService, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	client, err := his_oracle.NewClient(his_oracle.Config{
		Host:     oracleCfg.Host,
		Port:     oracleCfg.Port,
		Service:  oracleCfg.Service,
		Username: oracleCfg.Username,
		Password: oracleCfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("HIS Oracle client init failed: %w", err)
	}
	return &HisPriceSyncService{db: db, oracle: client}, nil
}

func (s *HisPriceSyncService) SyncPriceList(runID string) (fetched, created, updated int, errMsg string) {
	if s.oracle == nil {
		return 0, 0, 0, "HIS Oracle client not initialized"
	}

	var cfg models.SyncJobConfig
	batchSize := 1000
	timeoutSec := 60
	if err := s.db.Where("job_code = ?", models.SyncJobCodeHisPriceList).First(&cfg).Error; err == nil {
		if cfg.BatchSize > 0 {
			batchSize = cfg.BatchSize
		}
		if cfg.TimeoutSeconds > 0 {
			timeoutSec = cfg.TimeoutSeconds
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	totalCount, err := s.oracle.PriceListCount(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Sprintf("count failed: %v", err)
	}

	now := time.Now()

	for offset := 0; offset < totalCount; offset += batchSize {
		rows, err := s.oracle.QueryPriceList(ctx, his_oracle.QueryPriceListParams{
			Offset: offset,
			Limit:  batchSize,
		})
		if err != nil {
			return fetched, created, updated, fmt.Sprintf("batch query offset=%d: %v", offset, err)
		}

		for _, row := range rows {
			fetched++
			item := s.mapRowToItem(row)
			item.SyncedAt = now
			item.SyncRunID = &runID
			item.UpdatedAt = now

			isActive := item.StopDate == nil || item.StopDate.After(now)
			item.IsActive = isActive

			var existing models.HisPriceItem
			err := s.db.Where("source_system = ? AND item_code = ?",
				"HIS_ORACLE", item.ItemCode).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				item.ID = utils.GenerateID()
				item.CreatedAt = now
				if ce := s.db.Create(item).Error; ce != nil {
					log.Printf("[his_price_sync] create item_code=%s failed: %v", item.ItemCode, ce)
					continue
				}
				created++
			} else if err == nil {
				item.ID = existing.ID
				item.CreatedAt = existing.CreatedAt
				if ue := s.db.Model(&existing).Select("*").Updates(item).Error; ue != nil {
					log.Printf("[his_price_sync] update item_code=%s failed: %v", item.ItemCode, ue)
					continue
				}
				updated++
			} else {
				log.Printf("[his_price_sync] query item_code=%s failed: %v", item.ItemCode, err)
				continue
			}
		}
	}

	if errMsg == "" {
		log.Printf("[his_price_sync] sync complete: fetched=%d created=%d updated=%d", fetched, created, updated)
	}
	return
}

func (s *HisPriceSyncService) mapRowToItem(row his_oracle.PriceListRow) *models.HisPriceItem {
	item := &models.HisPriceItem{
		SourceSystem: "HIS_ORACLE",
		ItemCode:     row.ItemCode,
	}

	if row.ItemClass != nil {
		item.ItemClass = row.ItemClass
	}
	if row.ItemName != nil {
		item.ItemName = row.ItemName
	}
	if row.ItemSpec != nil {
		item.ItemSpec = row.ItemSpec
	}
	if row.Units != nil {
		item.Units = row.Units
	}
	if row.Price != nil {
		item.Price = row.Price
	}
	if row.PreferPrice != nil {
		item.PreferPrice = row.PreferPrice
	}
	if row.ForeignerPrice != nil {
		item.ForeignerPrice = row.ForeignerPrice
	}
	if row.PerformedBy != nil {
		item.PerformedBy = row.PerformedBy
	}
	if row.FeeTypeMask != nil {
		item.FeeTypeMask = row.FeeTypeMask
	}
	if row.ClassOnInpRcpt != nil {
		item.ClassOnInpRcpt = row.ClassOnInpRcpt
	}
	if row.ClassOnOutpRcpt != nil {
		item.ClassOnOutpRcpt = row.ClassOnOutpRcpt
	}
	if row.ClassOnReckoning != nil {
		item.ClassOnReckoning = row.ClassOnReckoning
	}
	if row.SubjCode != nil {
		item.SubjCode = row.SubjCode
	}
	if row.ClassOnMr != nil {
		item.ClassOnMr = row.ClassOnMr
	}
	if row.Memo != nil {
		item.Memo = row.Memo
	}
	if row.StartDate != nil {
		item.StartDate = row.StartDate
	}
	if row.StopDate != nil {
		item.StopDate = row.StopDate
	}
	if row.OperatorCode != nil {
		item.OperatorCode = row.OperatorCode
	}
	if row.EnterDate != nil {
		item.EnterDate = row.EnterDate
	}
	if row.HighPrice != nil {
		item.HighPrice = row.HighPrice
	}
	if row.MaterialCode != nil {
		item.MaterialCode = row.MaterialCode
	}
	if row.Score1 != nil {
		item.Score1 = row.Score1
	}
	if row.Score2 != nil {
		item.Score2 = row.Score2
	}
	if row.PriceNameCode != nil {
		item.PriceNameCode = row.PriceNameCode
	}
	if row.ControlFlag != nil {
		item.ControlFlag = row.ControlFlag
	}
	if row.InputCode != nil {
		item.InputCode = row.InputCode
	}
	if row.InputCodeWb != nil {
		item.InputCodeWb = row.InputCodeWb
	}
	if row.StdCode1 != nil {
		item.StdCode1 = row.StdCode1
	}
	if row.ChangedMemo != nil {
		item.ChangedMemo = row.ChangedMemo
	}
	if row.ClassOnInsurMr != nil {
		item.ClassOnInsurMr = row.ClassOnInsurMr
	}
	if row.PackageSpec != nil {
		item.PackageSpec = row.PackageSpec
	}
	if row.FirmID != nil {
		item.FirmID = row.FirmID
	}
	if row.ChargeAccording != nil {
		item.ChargeAccording = row.ChargeAccording
	}
	if row.LicenseID != nil {
		item.LicenseID = row.LicenseID
	}
	if row.UpdateFlag != nil {
		item.UpdateFlag = row.UpdateFlag
	}
	if row.DeptName != nil {
		item.DeptName = row.DeptName
	}
	if row.UpdateFlagSyb != nil {
		item.UpdateFlagSyb = row.UpdateFlagSyb
	}
	if row.MrBillClass != nil {
		item.MrBillClass = row.MrBillClass
	}
	if row.ClassOnMrAdd != nil {
		item.ClassOnMrAdd = row.ClassOnMrAdd
	}
	if row.CwtjCode != nil {
		item.CwtjCode = row.CwtjCode
	}
	if row.HighValue != nil {
		item.HighValue = row.HighValue
	}
	if row.DrgCode != nil {
		item.DrgCode = row.DrgCode
	}
	if row.InsurUpdate != nil {
		item.InsurUpdate = row.InsurUpdate
	}
	if row.StopOperator != nil {
		item.StopOperator = row.StopOperator
	}
	if row.LimitQuantity != nil {
		item.LimitQuantity = row.LimitQuantity
	}

	return item
}

func (s *HisPriceSyncService) IsSyncRunning() bool {
	var run models.SyncJobRun
	err := s.db.Where("job_code = ? AND status = ?",
		models.SyncJobCodeHisPriceList, models.SyncJobStatusRunning).
		Order("started_at DESC").First(&run).Error
	return err == nil
}
