#!/bin/bash

# 测试数据导入脚本
# 用途：快速创建 2-3 个测试患者，用于 API 验证

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  测试数据导入脚本${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查数据库连接
echo -e "\n${YELLOW}1. 检查数据库连接...${NC}"
if ! psql -h localhost -U elliotxin -d ai_hms_db -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${RED}❌ 数据库连接失败${NC}"
    echo "请检查数据库配置："
    echo "  - 主机: localhost"
    echo "  - 用户: elliotxin"
    echo "  - 数据库: ai_hms_db"
    exit 1
fi
echo -e "${GREEN}✅ 数据库连接成功${NC}"

# 导入测试数据
echo -e "\n${YELLOW}2. 导入测试数据...${NC}"
psql -h localhost -U elliotxin -d ai_hms_db -f scripts/test_data.sql

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ 测试数据导入成功${NC}"
else
    echo -e "${RED}❌ 测试数据导入失败${NC}"
    exit 1
fi

# 验证导入结果
echo -e "\n${YELLOW}3. 验证导入结果...${NC}"
PATIENT_COUNT=$(psql -h localhost -U elliotxin -d ai_hms_db -t -c "SELECT COUNT(*) FROM patients WHERE id LIKE 'test-%';")

echo -e "${GREEN}✅ 已创建 ${PATIENT_COUNT} 个测试患者${NC}"

# 显示测试患者列表
echo -e "\n${YELLOW}测试患者列表：${NC}"
psql -h localhost -U elliotxin -d ai_hms_db -c "
SELECT
  id,
  name,
  status,
  bed_number,
  risk_level
FROM patients
WHERE id LIKE 'test-%'
ORDER BY id;
"

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  导入完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\n下一步："
echo -e "  1. 启动后端服务: ${YELLOW}go run cmd/server/main.go${NC}"
echo -e "  2. 测试 API: ${YELLOW}curl http://localhost:8080/api/v1/patients${NC}"
echo -e "  3. 查看详情: ${YELLOW}curl http://localhost:8080/api/v1/patients/test-patient-001${NC}"
echo ""
