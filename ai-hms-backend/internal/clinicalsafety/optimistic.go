// Package clinicalsafety 提供临床写入安全的可复用原语：乐观锁、版本令牌(ETag)。
//
// 这是《临床写入安全加固设计.md》的参考实现，独立于业务 service，
// 不依赖也不改动任何迁移中的文件，供各写路径按需接入。
//
// 列名约定：本系统对接老血透库，列名为 PascalCase（如 "Id" "Version" "LastModifyTime"），
// 故 SQL 中显式加双引号。调用方传入的列名必须是**库内真实列名**（先核对 schema 文档）。
package clinicalsafety

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// VersionConflictError 乐观锁版本冲突：目标行已被他人修改或删除。
// handler 层应将其映射为 HTTP 409（或 412），提示前端重新加载后重试。
type VersionConflictError struct {
	Entity string
	ID     int64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("乐观锁冲突: %s#%d 已被他人修改或删除，请重新加载后重试", e.Entity, e.ID)
}

// IsConflict 判断 err 链中是否为乐观锁冲突，便于 handler 统一映射 409。
func IsConflict(err error) bool {
	var c *VersionConflictError
	return errors.As(err, &c)
}

// UpdateWithVersion 基于整型版本列的乐观更新（A 档，最稳）。
//
// 等价 SQL：UPDATE <table> SET <fields...>, "Version"="Version"+1 WHERE "Id"=? AND "Version"=?
// 命中 0 行（版本不符或行已删）返回 *VersionConflictError。
//
// fields 的 key 必须是真实列名，且**不应包含 Version**（由本函数自增）。
// 传入的 db 可以是事务句柄(tx)，以便与审计写入同事务。
func UpdateWithVersion(db *gorm.DB, table, entity string, id int64, expectedVersion int, fields map[string]any) error {
	if len(fields) == 0 {
		return errors.New("clinicalsafety: 空更新字段")
	}
	set := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		set[k] = v
	}
	set["Version"] = gorm.Expr(`"Version" + 1`)

	res := db.Table(table).
		Where(`"Id" = ? AND "Version" = ?`, id, expectedVersion).
		Updates(set)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return &VersionConflictError{Entity: entity, ID: id}
	}
	return nil
}

// UpdateWithTimestamp 基于时间戳列(如 LastModifyTime)的乐观更新（B 档，零改表过渡）。
//
// 等价 SQL：UPDATE <table> SET <fields...> WHERE "Id"=? AND "<tsColumn>"=?
// 命中 0 行（时间戳不符或行已删）返回 *VersionConflictError。
//
// ⚠️ 时间戳相等比较存在精度风险：expected 必须是"更新前从同一列读出的原值"，
// 且读出/回填须保持同一精度与时区（详见设计文档 §2.1）。
// 若列带 autoUpdateTime，ORM 会在本次更新时自动改写该列，不影响 WHERE 用旧值匹配。
func UpdateWithTimestamp(db *gorm.DB, table, entity, tsColumn string, id int64, expected time.Time, fields map[string]any) error {
	if len(fields) == 0 {
		return errors.New("clinicalsafety: 空更新字段")
	}
	res := db.Table(table).
		Where(`"Id" = ? AND "`+tsColumn+`" = ?`, id, expected).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return &VersionConflictError{Entity: entity, ID: id}
	}
	return nil
}
