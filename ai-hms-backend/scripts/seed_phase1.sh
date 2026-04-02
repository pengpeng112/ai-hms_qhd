#!/usr/bin/env bash
# =============================================================================
# seed_phase1.sh — 首轮环境 seed 数据执行脚本
# 用途：执行 seed_phase1.sql，创建初始用户和患者样本数据
# 执行前提：数据库已创建，AutoMigrate 已执行（表结构已就绪）
# =============================================================================

set -euo pipefail

# ---- 颜色输出 -----------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# ---- 数据库配置（支持环境变量覆盖）---------------------------------------------
DB_HOST="${DB_HOST:-10.10.8.83}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-amdin}"
DB_NAME="${DB_NAME:-Postgre}"
# 密码通过 PGPASSWORD 环境变量传递（推荐），或通过 .pgpass 文件
# export PGPASSWORD="${DB_PASS:-admin123}"

# ---- 脚本目录 -----------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="${SCRIPT_DIR}/seed_phase1.sql"

# ---- 打印配置 -----------------------------------------------------------------
echo -e "${CYAN}============================================================${NC}"
echo -e "${CYAN}  AI-HMS Phase 1 Seed 数据初始化${NC}"
echo -e "${CYAN}============================================================${NC}"
echo -e "\n${YELLOW}数据库配置：${NC}"
echo -e "  Host:     ${DB_HOST}"
echo -e "  Port:     ${DB_PORT}"
echo -e "  User:     ${DB_USER}"
echo -e "  Database: ${DB_NAME}"
echo -e "  SQL File: ${SQL_FILE}"
echo ""

# ---- 检查 psql 是否存在 -------------------------------------------------------
if ! command -v psql &>/dev/null; then
    echo -e "${RED}❌ psql 未找到，请安装 PostgreSQL client 工具${NC}"
    echo -e "   Ubuntu/Debian: sudo apt-get install postgresql-client"
    echo -e "   macOS:         brew install postgresql"
    exit 1
fi

# ---- 检查 SQL 文件是否存在 ----------------------------------------------------
if [[ ! -f "${SQL_FILE}" ]]; then
    echo -e "${RED}❌ SQL 文件不存在: ${SQL_FILE}${NC}"
    exit 1
fi

# ---- 检查数据库连接 -----------------------------------------------------------
echo -e "${YELLOW}1. 检查数据库连接...${NC}"
if ! PGPASSWORD="${DB_PASS:-admin123}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -c "SELECT 1" \
    >/dev/null 2>&1; then
    echo -e "${RED}❌ 数据库连接失败${NC}"
    echo -e "   请检查配置或通过环境变量覆盖："
    echo -e "   DB_HOST=... DB_PORT=... DB_USER=... DB_NAME=... DB_PASS=... bash seed_phase1.sh"
    exit 1
fi
echo -e "${GREEN}✅ 数据库连接成功${NC}"

# ---- 执行 seed SQL -----------------------------------------------------------
echo -e "\n${YELLOW}2. 执行 seed_phase1.sql...${NC}"
PGPASSWORD="${DB_PASS:-admin123}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -f "${SQL_FILE}" \
    -v ON_ERROR_STOP=1

echo -e "${GREEN}✅ Seed 数据执行成功${NC}"

# ---- 验证插入结果 -------------------------------------------------------------
echo -e "\n${YELLOW}3. 验证插入结果...${NC}"

USER_COUNT=$(PGPASSWORD="${DB_PASS:-admin123}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -t -c "SELECT COUNT(*) FROM users WHERE id LIKE 'seed-%'")

PATIENT_COUNT=$(PGPASSWORD="${DB_PASS:-admin123}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -t -c "SELECT COUNT(*) FROM patients WHERE id LIKE 'seed-%'")

echo -e "  已创建用户:  ${USER_COUNT// /} 条"
echo -e "  已创建患者:  ${PATIENT_COUNT// /} 条"

echo -e "\n${CYAN}============================================================${NC}"
echo -e "${GREEN}  Seed 初始化完成！${NC}"
echo -e "${CYAN}============================================================${NC}"
echo -e "\n下一步操作："
echo -e "  1. 启动后端服务:    ${YELLOW}go run cmd/server/main.go${NC}"
echo -e "  2. 初始化字典数据:  ${YELLOW}curl -X POST http://localhost:8080/api/v1/dict/items/init${NC}"
echo -e "  3. 验证登录:        ${YELLOW}curl -X POST http://localhost:8080/api/v1/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"test_admin\",\"password\":\"Test@123456\"}'${NC}"
echo ""
