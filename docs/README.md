# AI-HMS 文档索引

本目录作为当前仓库的主文档入口。后续优先维护这里列出的主文档，避免同主题文档在多个位置重复演化。

## 1. 部署与环境

- `docs/docker-migration-deploy-upgrade-guide.md`
  - Docker 首次部署、升级、回滚，以及前端同源 `/api` 代理实机经验
- `docs/environment-contract.md`
  - 环境变量与运行时约定
- `本地开发-内网迁移-开发对接指南.md`
  - 本地开发、准生产模拟、内网 Linux 迁移总览

## 2. 老库迁移主线

- `docs/migration-plan-legacy.md`
  - 新系统切换到老血透库的迁移主计划
- `docs/migration-field-map.md`
  - 新表到老表的字段级映射总表
- `docs/legacy-db-schema.md`
  - 结构化整理后的老血透库表结构摘要
- `docs/basic-info-legacy-gap-analysis.md`
  - 患者基础信息迁移差异分析
- `docs/dictionary-type-mapping-dev.md`
  - `CodeDictionary_CodeDictionarys` 的 `Type -> 统一字典编码` 映射与“其他”归类规则

## 3. 会话记录与持续跟进

- `docs/legacy-migration-session-summary-2026-04-21.md`
  - 迁移会话总结与待确认项
- `docs/treatment-execution-legacy-dev-record-2026-04-21.md`
  - 治疗执行链路迁移开发记录
- `ai-hms-backend/LEGACY_TABLE_FIELD_MAPPING.md`
  - 后端侧 legacy 替换与字段更新实施记录

## 4. 原始参考

- `老血透数据库表结构-合并版.md`
  - 老血透库原始结构主参考
- `DATABASE_DESIGN.md`
  - 新系统数据库设计主参考

## 5. 已整合说明

- 原 `docs/2026-04-20-frontend-only-access-proxy-plan.md` 已并入 `docs/docker-migration-deploy-upgrade-guide.md` 的同源 `/api` 升级与排障章节，不再单独保留。
- 原 `新血透DATABASE_DESIGN.md` 与 `DATABASE_DESIGN.md` 内容重复，保留后者。
- 原 `数据库表结构.md`、`数据库表设计.md` 的老库内容已由 `docs/legacy-db-schema.md`、`docs/migration-field-map.md` 接管。
- 原 `CODEX_FIXPLAN*.md`、`DEMOCK_PLAN.md` 为一次性执行计划，不再作为持续维护文档保留。
