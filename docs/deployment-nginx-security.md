# 安全部署参考 — Nginx 反向代理

> 本文档仅作参考，不替代运维团队的部署规范。HSTS 等 TLS 相关头部由 Nginx 在 HTTPS 终止层统一下发。

## 推荐 Nginx 安全配置

```nginx
# /etc/nginx/sites-available/hms

server {
    listen 80;
    server_name hms.example.com;
    # 强制 HTTPS 重定向
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name hms.example.com;

    # TLS 1.2 最低版本（兼容旧客户端），推荐 1.3
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305;
    ssl_prefer_server_ciphers off;

    ssl_certificate     /etc/ssl/certs/hms-fullchain.pem;
    ssl_certificate_key /etc/ssl/private/hms-privkey.pem;

    # === HSTS（应用层不设，在此统一下发）===
    # max-age=63072000 (2年) + includeSubDomains + preload 就绪
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;

    # === 安全响应头（与后端 middleware.SecurityHeaders 互补，双保险）===
    # X-Frame-Options（后端已设 DENY，此处不覆盖）
    # X-Content-Type-Options（后端已设 nosniff）
    # Referrer-Policy（后端已设 no-referrer）
    # 后端设了 CSP Report-Only，Nginx 不重复
    # 后端设了 Permissions-Policy，Nginx 不重复

    # === 请求体大小限制 ===
    client_max_body_size 20m;

    # 日志
    access_log /var/log/nginx/hms-access.log;
    error_log  /var/log/nginx/hms-error.log warn;

    # === 代理到后端应用 ===
    location / {
        # 前端静态资源（由 ai-hms-frontend dist/ 提供）
        root /opt/hms/frontend/dist;
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        # 后端 Go 服务
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 长连接：排班周视图 / 治疗监测等需 SSE
        proxy_read_timeout 120s;
        proxy_buffering off;
    }

    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
    }
}
```

## TLS 证书

```bash
# Let's Encrypt 自动续期（推荐）
certbot certonly --nginx -d hms.example.com
```

## 验证清单

部署后逐项验证：

| 检查项 | 方法 | 预期 |
|--------|------|------|
| HTTP→HTTPS 重定向 | `curl -I http://hms.example.com` | `301 Moved` → `Location: https://...` |
| HSTS | `curl -I https://hms.example.com` | `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload` |
| TLS 版本 | `openssl s_client -connect hms.example.com:443 -tls1_1` | 握手失败（拒绝 TLS 1.0/1.1） |
| 安全头 | 浏览器 DevTools → Network → Response Headers | 所有安全头存在 |
| CSP Report-Only | 浏览器 Console | 后端下发 CSP-Report-Only，不拦截资源 |
| 登录限流 | 连续 POST /api/v1/auth/login 6 次 | 第 6 次 429 |

## 注意事项

- HSTS `max-age` 初始部署时建议先用短值（如 300）验证，确认无误后再升至长期值
- `preload` 标记启用后可提交到 [hstspreload.org](https://hstspreload.org)
- 内网/本地开发环境**不要**开启 HSTS，或使用独立域名
- 后端已在 `middleware/security_headers.go` 下发 X-Frame-Options/X-Content-Type-Options/CSP 等，Nginx 层不重复
