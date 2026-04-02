-- 迁移脚本: 为 medical_histories 表添加专科记录扩展字段
-- 适用于: 生产环境（AutoMigrate 在 release 模式下跳过）
-- 数据库引擎: PostgreSQL 9.6+（本项目使用 gorm.io/driver/postgres）
-- 执行前请备份数据库

-- 原发病扩展字段
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS primary_disease_type VARCHAR(255) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS primary_disease_check_time VARCHAR(32) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS primary_disease_check_doc VARCHAR(100) DEFAULT '';

-- 病理诊断扩展字段
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS pathology_type VARCHAR(255) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS pathology_check_time VARCHAR(32) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS pathology_check_doc VARCHAR(100) DEFAULT '';

-- 过敏信息扩展字段
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS allergen_type VARCHAR(255) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS allergen_check_time VARCHAR(32) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS allergen_check_doc VARCHAR(100) DEFAULT '';

-- 肿瘤病史扩展字段
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS tumor_history_type VARCHAR(255) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS tumor_history_check_time VARCHAR(32) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS tumor_history_check_doc VARCHAR(100) DEFAULT '';

-- 并发症扩展字段
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS complication_type VARCHAR(255) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS complication_check_time VARCHAR(32) DEFAULT '';
ALTER TABLE medical_histories ADD COLUMN IF NOT EXISTS complication_check_doc VARCHAR(100) DEFAULT '';
