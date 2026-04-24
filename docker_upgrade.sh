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
ALLOW_SAME_IMAGE="${ALLOW_SAME_IMAGE:-1}"
ALLOW_METADATA_MISMATCH="${ALLOW_METADATA_MISMATCH:-1}"
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

short_id() {
    local id="$1"
    if [ -z "$id" ] || [ "$id" = "none" ]; then
        echo "none"
    else
        echo "${id:0:19}"
    fi
}

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

TARFILE="$(resolve_first_existing "$TARFILE_DEFAULT" "$DEPLOY_DIR/ai-hms-images.tar" "" || true)"
SHA_FILE="$(resolve_first_existing "$TARFILE_DEFAULT.sha256" "$DEPLOY_DIR/ai-hms-images.tar.sha256" "" || true)"
META_FILE="$(resolve_first_existing "$DEPLOY_DIR/ai-hms-images.meta.txt" "${TARFILE_DEFAULT%.tar}.meta.txt" "" || true)"

if [ -n "$TARFILE" ] && [ -n "$SHA_FILE" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
        info "  Verifying tar checksum: $TARFILE"
        EXPECTED_SHA=$(awk '{print $1}' "$SHA_FILE" | head -1 | tr -d '[:space:]')
        ACTUAL_SHA=$(sha256sum "$TARFILE" | awk '{print $1}')
        if [ -z "$EXPECTED_SHA" ]; then
            warn "sha256 file is empty or malformed, skip checksum verification"
        elif [ "$EXPECTED_SHA" != "$ACTUAL_SHA" ]; then
            error "Image tar checksum mismatch!\n  expected: $EXPECTED_SHA\n  actual  : $ACTUAL_SHA\n  File: $SHA_FILE"
        else
            info "  Tar checksum OK ($ACTUAL_SHA)"
        fi
    else
        warn "sha256sum not found, skip checksum verification"
    fi
else
    warn "Tar or checksum file not found; skip checksum verification"
fi

if [ -n "$META_FILE" ]; then
    info "  Using metadata file: $META_FILE"
else
    warn "ai-hms-images.meta.txt not found; metadata cross-check disabled"
fi

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
    NEW_BACKEND_CREATED=$(docker inspect --format='{{.Created}}' ai-hms-backend:latest 2>/dev/null || echo "unknown")
    info "  Backend  loaded : $(short_id "$NEW_BACKEND_ID")  created=$NEW_BACKEND_CREATED"
    if [ -n "$META_FILE" ]; then
        EXPECTED_BACKEND_ID=$(get_meta_value "backend_id" "$META_FILE")
        if [ -n "$EXPECTED_BACKEND_ID" ] && [ "$EXPECTED_BACKEND_ID" != "$NEW_BACKEND_ID" ]; then
            if [ "$ALLOW_METADATA_MISMATCH" != "1" ]; then
                error "Backend image ID does not match metadata.\n  expected=$(short_id "$EXPECTED_BACKEND_ID")\n  actual=$(short_id "$NEW_BACKEND_ID")\n  If this is intentional, rerun with: ALLOW_METADATA_MISMATCH=1 bash docker_upgrade.sh ${UPGRADE_TARGET}"
            fi
            warn "Backend image ID mismatches metadata, continue because ALLOW_METADATA_MISMATCH=1"
        fi
    fi
    if [ "$NEW_BACKEND_ID" == "$OLD_BACKEND_ID" ]; then
        if [ "$ALLOW_SAME_IMAGE" != "1" ]; then
            error "Backend image unchanged (same ID: $(short_id "$NEW_BACKEND_ID")).\n  This usually means your ai-hms-images.tar is old.\n  Rebuild/export on build machine, re-transfer tar, docker load again.\n  If you really want restart-only, run: ALLOW_SAME_IMAGE=1 bash docker_upgrade.sh ${UPGRADE_TARGET}"
        fi
        warn "Backend image unchanged (same ID), continue because ALLOW_SAME_IMAGE=1"
    else
        info "  Backend  change: $(short_id "$OLD_BACKEND_ID") -> $(short_id "$NEW_BACKEND_ID")"
    fi
fi

if [[ "$UPGRADE_TARGET" == "all" || "$UPGRADE_TARGET" == "frontend" ]]; then
    NEW_FRONTEND_ID=$(docker inspect --format='{{.Id}}' ai-hms-frontend:latest 2>/dev/null || echo "")
    if [ -z "$NEW_FRONTEND_ID" ]; then
        error "ai-hms-frontend:latest not found. Run: docker load -i ai-hms-images.tar"
    fi
    NEW_FRONTEND_CREATED=$(docker inspect --format='{{.Created}}' ai-hms-frontend:latest 2>/dev/null || echo "unknown")
    info "  Frontend loaded : $(short_id "$NEW_FRONTEND_ID")  created=$NEW_FRONTEND_CREATED"
    if [ -n "$META_FILE" ]; then
        EXPECTED_FRONTEND_ID=$(get_meta_value "frontend_id" "$META_FILE")
        if [ -n "$EXPECTED_FRONTEND_ID" ] && [ "$EXPECTED_FRONTEND_ID" != "$NEW_FRONTEND_ID" ]; then
            if [ "$ALLOW_METADATA_MISMATCH" != "1" ]; then
                error "Frontend image ID does not match metadata.\n  expected=$(short_id "$EXPECTED_FRONTEND_ID")\n  actual=$(short_id "$NEW_FRONTEND_ID")\n  If this is intentional, rerun with: ALLOW_METADATA_MISMATCH=1 bash docker_upgrade.sh ${UPGRADE_TARGET}"
            fi
            warn "Frontend image ID mismatches metadata, continue because ALLOW_METADATA_MISMATCH=1"
        fi
    fi
    if [ "$NEW_FRONTEND_ID" == "$OLD_FRONTEND_ID" ]; then
        if [ "$ALLOW_SAME_IMAGE" != "1" ]; then
            error "Frontend image unchanged (same ID: $(short_id "$NEW_FRONTEND_ID")).\n  This usually means your ai-hms-images.tar is old.\n  Rebuild/export on build machine, re-transfer tar, docker load again.\n  If you really want restart-only, run: ALLOW_SAME_IMAGE=1 bash docker_upgrade.sh ${UPGRADE_TARGET}"
        fi
        warn "Frontend image unchanged (same ID), continue because ALLOW_SAME_IMAGE=1"
    else
        info "  Frontend change: $(short_id "$OLD_FRONTEND_ID") -> $(short_id "$NEW_FRONTEND_ID")"
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
