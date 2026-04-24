package services

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"gorm.io/gorm"
)

type legacyPatientOrderCron struct{}

func (legacyPatientOrderCron) TableName() string {
	return "Order_PatientOrder"
}

// StartOrderCron 启动医嘱自动停用任务
// TODO: 当前假设单实例部署。多实例时需加分布式锁或 advisory lock
func StartOrderCron() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		runOrderCronOnce()
		for range ticker.C {
			runOrderCronOnce()
		}
	}()
	log.Println("Order cron job started: marking expired orders every 5 minutes")
}

func runOrderCronOnce() {
	db := database.GetDB()
	if db == nil {
		return
	}
	if err := disableExpiredLegacyOrders(db, time.Now()); err != nil {
		log.Printf("Warning: failed to disable expired legacy orders: %v", err)
	}
}

func disableExpiredLegacyOrders(db *gorm.DB, now time.Time) error {
	result := db.Model(&legacyPatientOrderCron{}).
		Where(`"EndTime" IS NOT NULL AND "EndTime" < ? AND COALESCE("IsDisabled", false) = false`, now).
		Updates(map[string]interface{}{
			"IsDisabled":     true,
			"LastModifyTime": now,
		})

	if isIgnorableOrderCronError(result.Error) {
		return nil
	}
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("Disabled %d expired legacy patient orders", result.RowsAffected)
	}
	return nil
}

func isIgnorableOrderCronError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "undefined_table") || strings.Contains(lower, "does not exist")
}
