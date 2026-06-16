-- 统一电子签名留痕表（契约05 #5 / 契约02 待签线）
-- 处方 / 方案 / 小结 三类待签共用，避免改老库加字段。
-- ⚠️ 本项目 AutoMigrate 永久禁用：此表须由 DBA 手工建表后，待签端点方可运行。
-- 对应 GORM 模型：internal/models/sign_record.go（snake_case，契约03/05）。

CREATE TABLE IF NOT EXISTS sign_record (
    id             VARCHAR(36)  NOT NULL,
    tenant_id      BIGINT       NOT NULL,
    target_type    VARCHAR(16)  NOT NULL,   -- prescription / plan / summary
    target_id      VARCHAR(64)  NOT NULL,   -- 被签对象 Id
    signer_id      VARCHAR(64)  NOT NULL,   -- 签名人 Id
    signer_name    VARCHAR(64),
    sign_time      TIMESTAMP    NOT NULL,
    signature_blob TEXT,                     -- 法律级签名(CA/图像)，v1 留空
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_sign_record PRIMARY KEY (id)
);

-- 审计/展示查询：按 (租户, 对象类型, 对象Id) 取留痕
CREATE INDEX IF NOT EXISTS idx_sign_record_target
    ON sign_record (tenant_id, target_type, target_id);
