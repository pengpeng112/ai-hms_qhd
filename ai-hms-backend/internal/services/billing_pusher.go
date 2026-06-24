package services

import (
	"errors"

	"github.com/elliotxin/ai-hms-backend/internal/models"
)

type PushResult struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
	Ref      string `json:"ref"`
}

type HISPusher interface {
	Channel() string
	Push(rec *models.ChargeRecord) (PushResult, error)
}

type NoopPusher struct{}

func (NoopPusher) Channel() string { return "noop" }

func (NoopPusher) Push(rec *models.ChargeRecord) (PushResult, error) {
	if rec == nil || rec.Status != models.ChargeStatusChecked {
		return PushResult{}, errors.New("清单状态不允许推送")
	}
	return PushResult{
		Accepted: false,
		Message:  "HIS 推送接口暂未启用：请先导出 Excel 由护士录入 HIS。清单已标记为待推送。",
	}, nil
}
