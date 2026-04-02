#!/bin/bash
# deploy/install-systemd.sh
# 在 openEuler 22.03 (10.10.8.84) 上安装 systemd 服务的操作步骤说明
# 此脚本仅供参考，手动执行各步骤时请确认路径和权限

set -e

APP_DIR=/opt/ai-hms
SERVICE_FILE=ai-hms-backend.service
BINARY_NAME=ai-hms-backend

echo "=== AI-HMS 后端 systemd 安装指引 ==="

# 1. 创建专用用户（如未存在）
# useradd -r -s /sbin/nologin -d $APP_DIR ai-hms

# 2. 创建应用目录
# mkdir -p $APP_DIR/logs
# chown -R ai-hms:ai-hms $APP_DIR

# 3. 将编译好的二进制复制到目标路径
# cp $BINARY_NAME $APP_DIR/
# chmod +x $APP_DIR/$BINARY_NAME

# 4. 将 .env.production.template 复制并填写真实值
# cp .env.production.template $APP_DIR/.env
# chmod 600 $APP_DIR/.env
# chown ai-hms:ai-hms $APP_DIR/.env
# vi $APP_DIR/.env

# 5. 安装 systemd unit 文件
# cp deploy/$SERVICE_FILE /etc/systemd/system/
# systemctl daemon-reload

# 6. 启用并启动服务
# systemctl enable $SERVICE_FILE
# systemctl start ai-hms-backend

# 7. 验证运行状态
# systemctl status ai-hms-backend
# journalctl -u ai-hms-backend -f

echo ""
echo "=== chromedp/Chrome 依赖（HDIS 集成阶段需要）==="
echo "首轮部署暂不需要 Chrome。待 HDIS Token 集成时，在 10.10.8.84 执行："
echo "  yum install chromium-browser"
echo "  # 或 Docker 内安装时在 Dockerfile 中添加："
echo "  # RUN apk add --no-cache chromium chromium-chromedriver"
echo ""
echo "chromedp 包已在 go.mod 中引入（github.com/chromedp/chromedp v0.14.2）"
echo "首轮部署时相关代码路径不会被调用，无需额外配置。"
