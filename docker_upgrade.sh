#!/bin/bash
# ============================================================
# AI-HMS - Docker 升级脚本（在 10.10.8.84 服务器上运行）
#
# 前提:
#   1. 首次部署已完成（docker_deploy.sh 已执行过）
#   2. 新镜像已加载: docker load -i ai-hms-images.tar
#
# 用法:
#   cd /opt/ai-hms-docker
#   bash docker_upgrade.sh                # 升级全部
#   bash docker_upgrade.sh backend        # 仅升级后端
#   bash docker_upgrade.sh frontend       # 仅升级前端
#
# 升级策略:
#   - 备份当前镜像 ID 以便回滚
#   - 滚动重启：先 backend 再 frontend（保证 API 先就绪）
#   - 健康检查通过后才算成功
#   - 失败时自动提示回滚命令
# ============================================================
set -euo pipefail

cd "$(dirname "$0")"
DEPLOY_DIR="$(pwd)"
ENV_FILE="$DEPLOY_DIR/.env"
COMPOSE_FILE="$DEPLOY_DIR/docker-compose.yml"
UPGRADE_TARGET="${1:-all}"   # all | backend | frontend

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# Docker Compose 命令
if docker compose version >/dev/null 2>&1; then
    DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
    DC="docker-compose"
else
    error "Docker Compose not found"
fi

TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo ""
echo "============================================================"
echo "  AI-HMS - Upgrade ($UPGRADE_TARGET)"
echo "  Time: $TIMESTAMP"
echo "============================================================"
echo ""

# ================================================================
# [1] 检查前提
# ================================================================
info "[1/5] Pre-flight checks ..."

[ -f "$ENV_FILE" ] || error ".env not found. Was docker_deploy.sh run first?"
[ -f "$COMPOSE_FILE" ] || error "docker-compose.yml not found"

# ================================================================
# [2] 备份当前镜像 ID（用于回滚）
# ================================================================
info "[2/5] Recording current image IDs for rollback ..."

ROLLBACK_FILE="$DEPLOY_DIR/.rollback_${TIMESTAMP}"

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "backend" ]]; then
    OLD_BACKEND_ID=$(docker inspect --format='{{.Image}}' ai-hms-backend 2>/dev/null || echo "none")
    echo "backend=$OLD_BACKEND_ID" >> "$ROLLBACK_FILE"
    info "  Backend  current: ${OLD_BACKEND_ID:0:12}"
fi

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "frontend" ]]; then
    OLD_FRONTEND_ID=$(docker inspect --format='{{.Image}}' ai-hms-frontend 2>/dev/null || echo "none")
    echo "frontend=$OLD_FRONTEND_ID" >> "$ROLLBACK_FILE"
    info "  Frontend current: ${OLD_FRONTEND_ID:0:12}"
fi

info "  Rollback info: $ROLLBACK_FILE"

# ================================================================
# [3] 检查新镜像是否已加载
# ================================================================
info "[3/5] Checking new images ..."

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "backend" ]]; then
    NEW_BACKEND_ID=$(docker inspect --format='{{.Id}}' ai-hms-backend:latest 2>/dev/null || echo "")
    if [ -z "$NEW_BACKEND_ID" ]; then
        error "ai-hms-backend:latest not found. Run: docker load -i ai-hms-images.tar"
    fi
    if [ "$NEW_BACKEND_ID" == "$OLD_BACKEND_ID" ]; then
        warn "Backend image unchanged (same ID), will restart anyway"
    else
        info "  Backend  new: ${NEW_BACKEND_ID:0:12}"
    fi
fi

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "frontend" ]]; then
    NEW_FRONTEND_ID=$(docker inspect --format='{{.Id}}' ai-hms-frontend:latest 2>/dev/null || echo "")
    if [ -z "$NEW_FRONTEND_ID" ]; then
        error "ai-hms-frontend:latest not found. Run: docker load -i ai-hms-images.tar"
    fi
    if [ "$NEW_FRONTEND_ID" == "$OLD_FRONTEND_ID" ]; then
        warn "Frontend image unchanged (same ID), will restart anyway"
    else
        info "  Frontend new: ${NEW_FRONTEND_ID:0:12}"
    fi
fi

# ================================================================
# [4] 滚动升级
# ================================================================
info "[4/5] Upgrading containers ..."

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "backend" ]]; then
    info "Recreating backend ..."
    $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d --no-deps --force-recreate backend
    info "Backend container recreated"

    # 等待后端健康
    info "Waiting for backend health check ..."
    BACKEND_OK=false
    for i in $(seq 1 20); do
        if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
            BACKEND_OK=true
            break
        fi
        sleep 3
    done

    if $BACKEND_OK; then
        info "Backend: healthy ✓"
    else
        warn "Backend not responding after 60s"
        warn "Check: docker logs ai-hms-backend"
        warn "Rollback: docker stop ai-hms-backend && docker rm ai-hms-backend"
        warn "  Then tag old image and redeploy"
        exit 1
    fi
fi

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "frontend" ]]; then
    info "Recreating frontend ..."
    $DC -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d --no-deps --force-recreate frontend
    info "Frontend container recreated"

    # 等待前端健康
    info "Waiting for frontend health check ..."
    FRONTEND_OK=false
    for i in $(seq 1 15); do
        if curl -sf http://localhost:3000/nginx-health >/dev/null 2>&1; then
            FRONTEND_OK=true
            break
        fi
        sleep 2
    done

    if $FRONTEND_OK; then
        info "Frontend: healthy ✓"
    else
        warn "Frontend not responding after 30s"
        warn "Check: docker logs ai-hms-frontend"
        exit 1
    fi
fi

# ================================================================
# [5] 验证 & 清理
# ================================================================
info "[5/5] Post-upgrade verification ..."

echo ""
$DC -f "$COMPOSE_FILE" ps
echo ""

# 清理悬空镜像（旧版本）
DANGLING=$(docker images -f "dangling=true" -q 2>/dev/null || true)
if [ -n "$DANGLING" ]; then
    echo ""
    read -r -p "  Clean up old dangling images? [Y/n] " clean_answer
    if [[ ! "$clean_answer" =~ ^[Nn] ]]; then
        docker image prune -f
        info "Old images cleaned"
    fi
fi

SERVER_IP=$(hostname -I | awk '{print $1}')

echo ""
echo "============================================================"
echo "  [DONE] Upgrade complete!"
echo "============================================================"
echo ""
echo "  Web UI   : http://${SERVER_IP}:3000"
echo "  API      : http://${SERVER_IP}:8080"
echo "  Rollback : cat $ROLLBACK_FILE"
echo ""
echo "  Rollback steps (if needed):"
echo "    1. docker stop ai-hms-backend ai-hms-frontend"
echo "    2. docker rm ai-hms-backend ai-hms-frontend"
echo "    3. docker tag <old-image-id> ai-hms-backend:latest"
echo "    4. docker tag <old-image-id> ai-hms-frontend:latest"
echo "    5. $DC -f $COMPOSE_FILE --env-file $ENV_FILE up -d"
echo "============================================================"
