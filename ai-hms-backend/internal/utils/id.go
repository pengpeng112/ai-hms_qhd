package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateID 生成唯一 ID（通用）
func GenerateID() string {
	return uuid.New().String()
}

// GeneratePatientID 生成患者编号
// 格式: P + 年月日 + 3位序号，例如: P20250128001
// 注意: 此函数需要传入当天已有的患者数量来生成序号
func GeneratePatientID(count int) string {
	now := time.Now()
	dateStr := now.Format("20060102")
	seq := count + 1
	return fmt.Sprintf("P%s%03d", dateStr, seq)
}
