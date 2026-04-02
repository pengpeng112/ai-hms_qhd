-- 迁移脚本: 为 outcome_records 表添加 is_door_rule 字段
-- 适用于: 生产环境（AutoMigrate 在 release 模式下跳过）
-- 数据库引擎: PostgreSQL 9.6+（本项目使用 gorm.io/driver/postgres）
-- 执行前请备份数据库

-- 添加门规字段到转归记录表
ALTER TABLE outcome_records ADD COLUMN IF NOT EXISTS is_door_rule BOOLEAN DEFAULT FALSE;
