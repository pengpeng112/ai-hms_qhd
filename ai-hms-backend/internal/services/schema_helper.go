package services

import "gorm.io/gorm"

// ensureTables 历史兼容占位函数。
// 老血透生产库禁止执行任何 DDL，此函数永远不执行建表/改表操作。
func ensureTables(_ *gorm.DB, _ ...interface{}) error { return nil }

// ensureSchema 历史兼容占位函数。
// 老血透生产库禁止执行任何 DDL，此函数永远不执行建表/改表操作。
func ensureSchema(_ *gorm.DB, _ ...interface{}) error { return nil }
