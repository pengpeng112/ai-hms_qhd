package services

import (
	"errors"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/database"
	"github.com/elliotxin/ai-hms-backend/internal/models"
	"github.com/elliotxin/ai-hms-backend/internal/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// SignService 统一电子签名留痕服务（处方/方案/小结三类共用，契约02 待签线 / 契约05 #5）。
// 依赖独立新表 sign_record；本项目 AutoMigrate 永久禁用，应由部署阶段执行 docs/sql/deploy_new_tables.sql。
type SignService struct {
	db *gorm.DB
}

func NewSignService() *SignService {
	return &SignService{db: database.GetDB()}
}

var validSignTargets = map[string]struct{}{
	models.SignTargetPrescription:          {},
	models.SignTargetPlan:                  {},
	models.SignTargetSummary:               {},
	models.SignTargetInfectiousDisposition: {},
}

func normalizeSignRecordError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
		return errors.New("sign_record 表不存在，请先在部署阶段执行 docs/sql/deploy_new_tables.sql")
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "sign_record") && (strings.Contains(lower, "does not exist") || strings.Contains(lower, "undefined_table")) {
		return errors.New("sign_record 表不存在，请先在部署阶段执行 docs/sql/deploy_new_tables.sql")
	}
	return err
}

// Sign 写入一条电子签名留痕（谁 / 何时 / 签了什么）。v1 不写法律级 signature_blob。
func (s *SignService) Sign(tenantID int64, targetType, targetID, signerID, signerName string) (*models.SignRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	targetType = strings.TrimSpace(targetType)
	targetID = strings.TrimSpace(targetID)
	if _, ok := validSignTargets[targetType]; !ok {
		return nil, errors.New("invalid sign target type")
	}
	if targetID == "" {
		return nil, errors.New("target id is required")
	}
	if strings.TrimSpace(signerID) == "" {
		return nil, errors.New("signer id is required")
	}

	rec := models.SignRecord{
		ID:         utils.GenerateID(),
		TenantID:   tenantID,
		TargetType: targetType,
		TargetID:   targetID,
		SignerID:   strings.TrimSpace(signerID),
		SignerName: strings.TrimSpace(signerName),
		SignTime:   time.Now(),
	}
	if err := s.db.Create(&rec).Error; err != nil {
		return nil, normalizeSignRecordError(err)
	}
	return &rec, nil
}

// ListSigns 返回某对象的签名留痕（审计/展示），按签名时间倒序。
func (s *SignService) ListSigns(tenantID int64, targetType, targetID string) ([]models.SignRecord, error) {
	if s.db == nil {
		return nil, errors.New("database not available")
	}
	var rows []models.SignRecord
	if err := s.db.
		Where("tenant_id = ? AND target_type = ? AND target_id = ?", tenantID, strings.TrimSpace(targetType), strings.TrimSpace(targetID)).
		Order("sign_time DESC").
		Find(&rows).Error; err != nil {
		return nil, normalizeSignRecordError(err)
	}
	return rows, nil
}
