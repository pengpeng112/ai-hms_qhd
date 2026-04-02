package services

import (
	"log"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"gorm.io/gorm"
)

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
	if err := markExpiredOrders(db, time.Now()); err != nil {
		log.Printf("Warning: failed to mark expired orders: %v", err)
	}
}

func markExpiredOrders(db *gorm.DB, now time.Time) error {
	result := db.Model(&models.Order{}).
		Where("end_time IS NOT NULL AND end_time < ? AND status IN ?", now, activeOrderStatuses).
		Updates(map[string]interface{}{
			"status": models.OrderStatusStopped,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		log.Printf("Marked %d expired orders as stopped", result.RowsAffected)
	}
	return nil
}
