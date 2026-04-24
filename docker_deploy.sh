#!/bin/bash
# ============================================================
# AI-HMS - Docker 首次部署脚本（在 10.10.8.84 openEuler 服务器上运行）
#
# 前提:
#   1. docker load -i ai-hms-images.tar 已执行
#   2. 数据库 10.10.8.83 PostgreSQL 已可达
#   3. 本脚本所在目录包含 docker-compose.yml, .env.production.template
#
# 用法:
#   cd /opt/ai-hms-docker
#   bash docker_deploy.sh
#
# 说明:
#   - 前后端各一个 Docker 容器，通过 bridge 网络互通
#   - 前端 nginx 反向代理 /api/ 到后端 :8080
#   - 后端连接外部 PostgreSQL (10.10.8.83)
#   - 对外只暴露前端 :3000 端口，后端 :8080 可选暴露
# ============================================================
set -euo pipefail

cd "$(dirname "$0")"
DEPLOY_DIR="$(pwd)"
ENV_FILE="$DEPLOY_DIR/.env"
COMPOSE_FILE="$DEPLOY_DIR/docker-compose.yml"
ALLOW_METADATA_MISMATCH="${ALLOW_METADATA_MISMATCH:-0}"
TARFILE_DEFAULT="/opt/ai-hms-images.tar"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

resolve_first_existing() {
    for p in "$@"; do
        if [ -f "$p" ]; then
            echo "$p"
            return 0
        fi
    done
    return 1
}

get_meta_value() {
    local key="$1"
    local file="$2"
    awk -F'=' -v k="$key" '$1==k {print substr($0, index($0, "=")+1); exit}' "$file"
}

echo ""
echo "============================================================"
echo "  AI-HMS - Docker First Deploy"
echo "  Target: $(hostname) ($(hostname -I | awk '{print $1}'))"
echo "============================================================"
echo ""

# ================================================================
# [1/6] 检查 Docker
# ================================================================
info "[1/6] Checking Docker ..."
command -v docker >/dev/null 2>&1 || error "Docker not installed.\n  openEuler: dnf install -y docker-ce docker-ce-cli containerd.io\n  Then: systemctl enable --now docker"
docker info >/dev/null 2>&1 || error "Docker daemon not running. Run: systemctl start docker"
info "Docker OK: $(docker --version)"

# 检查 docker compose (v2 plugin 或 v1 独立)
if docker compose version >/dev/null 2>&1; then
    DC="docker compose"
    info "Docker Compose (plugin): $(docker compose version --short 2>/dev/null || echo 'v2')"
elif command -v docker-compose >/dev/null 2>&1; then
    DC="docker-compose"
    info "Docker Compose (standalone): $(docker-compose version --short 2>/dev/null || echo 'v1')"
else
    error "Docker Compose not found.\n  Install: dnf install -y docker-compose-plugin"
fi

TARFILE="$(resolve_first_existing "$TARFILE_DEFAULT" "$DEPLOY_DIR/ai-hms-images.tar" "" || true)"
SHA_FILE="$(resolve_first_existing "$TARFILE_DEFAULT.sha256" "$DEPLOY_DIR/ai-hms-images.tar.sha256" "" || true)"
META_FILE="$(resolve_first_existing "$DEPLOY_DIR/ai-hms-images.meta.txt" "${TARFILE_DEFAULT%.tar}.meta.txt" "" || true)"

if [ -n "$TARFILE" ] && [ -n "$SHA_FILE" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
        info "Checksum verifying image tar ..."
        EXPECTED_SHA=$(awk '{print $1}' "$SHA_FILE" | head -1 | tr -d '[:space:]')
        ACTUAL_SHA=$(sha256sum "$TARFILE" | awk '{print $1}')
        if [ -z "$EXPECTED_SHA" ]; then
            warn "sha256 file is empty or malformed, skip checksum verification"
        elif [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
            error "Image tar checksum mismatch!\n  expected: $EXPECTED_SHA\n  actual  : $ACTUAL_SHA"
        else
            info "Tar checksum OK ($ACTUAL_SHA)"
        fi
    else
        warn "sha256sum not found, skip checksum verification"
    fi
else
    warn "Tar or checksum file not found; skip checksum verification"
fi

if [ -n "$META_FILE" ]; then
    info "Using metadata file: $META_FILE"
else
    warn "ai-hms-images.meta.txt not found; metadata cross-check disabled"
fi

# ================================================================
# [2/6] 检查镜像
# ================================================================
info "[2/6] Checking Docker images ..."
MISSING=""
docker image inspect ai-hms-backend:latest  >/dev/null 2>&1 || MISSING="ai-hms-backend:latest "
docker image inspect ai-hms-frontend:latest >/dev/null 2>&1 || MISSING="${MISSING}ai-hms-frontend:latest"
if [ -n "$MISSING" ]; then
    error "Image(s) not found: $MISSING\n  Run first: docker load -i ai-hms-images.tar"
fi
info "Images ready:"
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}  ({{.CreatedSince}})" ai-hms-backend:latest
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}  ({{.CreatedSince}})" ai-hms-frontend:latest

BACKEND_ID=$(docker inspect --format='{{.Id}}' ai-hms-backend:latest 2>/dev/null || echo "")
FRONTEND_ID=$(docker inspect --format='{{.Id}}' ai-hms-frontend:latest 2>/dev/null || echo "")

if [ -n "$META_FILE" ]; then
    EXPECTED_BACKEND_ID=$(get_meta_value "backend_id" "$META_FILE")
    EXPECTED_FRONTEND_ID=$(get_meta_value "frontend_id" "$META_FILE")
    if [ -n "$EXPECTED_BACKEND_ID" ] && [ "$EXPECTED_BACKEND_ID" != "$BACKEND_ID" ]; then
        if [ "$ALLOW_METADATA_MISMATCH" != "1" ]; then
            error "Loaded backend image does not match metadata.\n  expected=${EXPECTED_BACKEND_ID:0:19}\n  actual=${BACKEND_ID:0:19}\n  If intentional, rerun with: ALLOW_METADATA_MISMATCH=1 bash docker_deploy.sh"
        fi
        warn "Backend image metadata mismatch, continue because ALLOW_METADATA_MISMATCH=1"
    fi
    if [ -n "$EXPECTED_FRONTEND_ID" ] && [ "$EXPECTED_FRONTEND_ID" != "$FRONTEND_ID" ]; then
        if [ "$ALLOW_METADATA_MISMATCH" != "1" ]; then
            error "Loaded frontend image does not match metadata.\n  expected=${EXPECTED_FRONTEND_ID:0:19}\n  actual=${FRONTEND_ID:0:19}\n  If intentional, rerun with: ALLOW_METADATA_MISMATCH=1 bash docker_deploy.sh"
        fi
        warn "Frontend image metadata mismatch, continue because ALLOW_METADATA_MISMATCH=1"
    fi
fi

# ================================================================
# [3/6] 配置 .env
# ================================================================
info "[3/6] Setting up environment configuration ..."

if [ ! -f "$ENV_FILE" ]; then
    if [ ! -f "$DEPLOY_DIR/.env.production.template" ]; then
        error ".env.production.template not found in $DEPLOY_DIR"
    fi
    cp "$DEPLOY_DIR/.env.production.template" "$ENV_FILE"

    # 自动生成强随机密钥
    JWT_KEY=$(openssl rand -hex 32 2>/dev/null || python3 -c "import secrets;print(secrets.token_hex(32))" 2>/dev/null || cat /proc/sys/kernel/random/uuid | tr -d '-')
    APP_KEY=$(openssl rand -hex 32 2>/dev/null || python3 -c "import secrets;print(secrets.token_hex(32))" 2>/dev/null || cat /proc/sys/kernel/random/uuid | tr -d '-')

    # 替换占位符
    sed -i "s|<generate-a-strong-secret-min-32-chars>|$JWT_KEY|" "$ENV_FILE"  # 第一个匹配 JWT_SECRET
    # APP_SECRET 需要单独处理（第二个同名占位符已在第一次被替换，这里用固定行匹配）
    sed -i "s|APP_SECRET=.*|APP_SECRET=$APP_KEY|" "$ENV_FILE"
    sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_KEY|" "$ENV_FILE"
    sed -i "s|<your-db-password>|admin123|" "$ENV_FILE"

    SERVER_IP=$(hostname -I | awk '{print $1}')
    sed -i "s|CORS_ALLOWED_ORIGINS=.*|CORS_ALLOWED_ORIGINS=http://${SERVER_IP}:3000|" "$ENV_FILE"
    sed -i "s|VITE_API_BASE_URL=.*|VITE_API_BASE_URL=http://${SERVER_IP}:8080|" "$ENV_FILE"

    info ".env generated from template (random keys auto-filled)"
    warn "Review and adjust: $ENV_FILE"
    warn "Especially: DB_PASSWORD, DB_HOST, DB_USER, DB_NAME"
    echo ""
    echo "  Current DB config:"
    grep -E "^DB_" "$ENV_FILE" | sed 's/^/    /'
    echo ""
    read -r -p "  Continue with these settings? [Y/n] " answer
    if [[ "$answer" =~ ^[Nn] ]]; then
        info "Edit $ENV_FILE then re-run: bash docker_deploy.sh"
        exit 0
    fi
else
    info ".env already exists, keeping it"
fi

# ================================================================
# [4/6] 检查数据库连通性
# ================================================================
info "[4/6] Checking database connectivity ..."
DB_HOST=$(grep -E "^DB_HOST=" "$ENV_FILE" | cut -d'=' -f2)
DB_PORT=$(grep -E "^DB_PORT=" "$ENV_FILE" | cut -d'=' -f2)
DB_HOST=${DB_HOST:-10.10.8.83}
DB_PORT=${DB_PORT:-5432}

# 简单 TCP 检测
if command -v nc >/dev/null 2>&1; then
    if nc -z -w3 "$DB_HOST" "$DB_PORT" 2>/dev/null; then
        info "Database $DB_HOST:$DB_PORT is reachable"
    else
        warn "Cannot reach $DB_HOST:$DB_PORT — database may not be running or firewall blocked"
        warn "Continuing anyway (backend will retry on startup)"
    fi
elif command -v timeout >/dev/null 2>&1; then
    if timeout 3 bash -c "echo >/dev/tcp/$DB_HOST/$DB_PORT" 2>/dev/null; then
        info "Database $DB_HOST:$DB_PORT is reachable"
    else
        warn "Cannot reach $DB_HOST:$DB_PORT — check network/firewall"
    fi
else
    warn "nc/timeout not available, skipping DB connectivity check"
fi

# ================================================================
# [5/6] 验证老血透数据库可访问性（seed_phase1.sql 已废弃）
# ================================================================
# NOTE: seed_phase1.sql 是为旧新建库（snake_case 表）设计的，现已废弃。
# 本系统直接连接老血透 PostgreSQL（dialysis@10.10.8.83），数据已在老库中，无需种子。
info "[5/6] Verifying legacy DB accessibility ..."
DB_USER=$(grep -E "^DB_USER=" "$ENV_FILE" | cut -d'=' -f2)
DB_PASS=$(grep -E "^DB_PASSWORD=" "$ENV_FILE" | cut -d'=' -f2)
DB_NAME=$(grep -E "^DB_NAME=" "$ENV_FILE" | cut -d'=' -f2)

if command -v psql >/dev/null 2>&1; then
    ROW_COUNT=$(PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -tAc "SELECT COUNT(*) FROM \"Identity_Users\" WHERE \"TenantId\"=3;" 2>/dev/null || echo "")
    if [ -n "$ROW_COUNT" ] && [ "$ROW_COUNT" -gt 0 ] 2>/dev/null; then
        info "Legacy DB OK — Identity_Users(TenantId=3): $ROW_COUNT user(s) found"
    else
        warn "Could not query Identity_Users from legacy DB — check DB connectivity or credentials"
    fi
else
    info "psql not available, skipping legacy DB verification (backend will check on startup)"
fi

# ================================================================
# [6/6] 启动容器
# ================================================================
info "[6/6] Starting AI-HMS containers ..."

# 停止旧容器（如果存在）
$DC -f "$COMPOSE_FILE" down 2>/dev/null || true

# 创建日志目录
mkdir -p "$DEPLOY_DIR/logs"

# 启动
$DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d

info "Waiting for services to start ..."
sleep 5

# 健康检查
BACKEND_OK=false
FRONTEND_OK=false

for i in $(seq 1 20); do
    if ! $BACKEND_OK && curl -sf http://localhost:8080/health >/dev/null 2>&1; then
        BACKEND_OK=true
    fi
    if ! $FRONTEND_OK && curl -sf http://localhost:3000/nginx-health >/dev/null 2>&1; then
        FRONTEND_OK=true
    fi
    if $BACKEND_OK && $FRONTEND_OK; then
        break
    fi
    sleep 3
done

echo ""
echo "============================================================"
echo "  AI-HMS Deploy Result"
echo "============================================================"
echo ""

SERVER_IP=$(hostname -I | awk '{print $1}')

if $BACKEND_OK; then
    echo -e "  Backend  : ${GREEN}✓ running${NC}  http://${SERVER_IP}:8080/health"
else
    echo -e "  Backend  : ${RED}✗ not responding${NC}  (check: docker logs ai-hms-backend)"
fi

if $FRONTEND_OK; then
    echo -e "  Frontend : ${GREEN}✓ running${NC}  http://${SERVER_IP}:3000"
else
    echo -e "  Frontend : ${RED}✗ not responding${NC}  (check: docker logs ai-hms-frontend)"
fi

echo ""
echo "  ---- Access ----"
echo "  Web UI   : http://${SERVER_IP}:3000"
echo "  API      : http://${SERVER_IP}:8080"
echo "  Login    : 使用老血透系统账号登录 (Identity_Users, TenantId=3)"
echo "             备用密码 admin@123qwe 仅在 GIN_MODE!=release 时有效"
echo ""
echo "  ---- Commands ----"
echo "  View logs     : docker logs -f ai-hms-backend"
echo "                  docker logs -f ai-hms-frontend"
echo "  Stop all      : $DC -f $COMPOSE_FILE down"
echo "  Restart all   : $DC -f $COMPOSE_FILE restart"
echo "  Status        : $DC -f $COMPOSE_FILE ps"
echo ""
echo "  ---- Upgrade ----"
echo "  1. Transfer new ai-hms-images.tar to server"
echo "  2. docker load -i ai-hms-images.tar"
echo "  3. bash docker_upgrade.sh"
echo "============================================================"

if ! $BACKEND_OK || ! $FRONTEND_OK; then
    echo ""
    warn "Some services are not healthy. Check logs above."
    exit 1
fi
