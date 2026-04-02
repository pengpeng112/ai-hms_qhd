# 最小环境变量契约

## 后端必填

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `JWT_SECRET`
- `APP_SECRET`
- `CORS_ALLOWED_ORIGINS`

## 后端可选

- `SERVER_HOST`（默认 `0.0.0.0`）
- `SERVER_PORT`（默认 `8080`）
- `GIN_MODE`（默认 `debug`）
- `DB_SSL_MODE`（默认 `disable`）
- `JWT_EXPIRATION_HOURS`（默认 `24`）

## 前端必填

- `VITE_API_BASE_URL`

## 失败策略

- 必填项缺失时，服务直接失败。
- 禁止回退到 `localhost`、空密钥或弱默认密钥。

---

## 系统级依赖（宿主机）

### chromedp / Chrome Headless（后续阶段需要）

> **首轮部署不需要**。HDIS Token 自动获取功能集成时（后续阶段），宿主机 10.10.8.84 必须安装 Chromium。

| 部署方式 | 安装命令 |
|---------|---------|
| openEuler 22.03（宿主机） | `yum install chromium-browser` |
| Docker 容器内（alpine）    | `apk add --no-cache chromium chromium-chromedriver` |

- go.mod 已包含 `github.com/chromedp/chromedp v0.14.2`
- 首轮部署时 chromedp 相关代码路径不会被调用，无需额外配置
- Docker 方式需将 runtime stage 的 `alpine:latest` 替换为包含 Chromium 的镜像，并设置 `CHROMIUM_PATH=/usr/bin/chromium-browser` 环境变量
