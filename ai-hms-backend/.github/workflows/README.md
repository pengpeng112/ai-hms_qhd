# CI/CD 配置指南

## 概述

项目使用 GitHub Actions 实现自动化部署，推送代码到 `main` 或 `develop` 分支时自动触发部署。

## 工作流

### 前端部署 (`.github/workflows/deploy-frontend.yml`)

**触发条件：**
- 推送到 `main` 或 `develop` 分支
- 修改了 `ai-hms-frontend/` 目录下的文件
- 手动触发 (workflow_dispatch)

**部署流程：**
1. 安装依赖
2. 运行 ESLint 检查
3. 构建生产版本
4. 备份现有部署
5. 上传到服务器
6. 重启 Nginx
7. 健康检查

### 后端部署 (`.github/workflows/deploy-backend.yml`)

**触发条件：**
- 推送到 `main` 或 `develop` 分支
- 修改了 `ai-hms-backend/` 目录下的文件
- 手动触发 (workflow_dispatch)

**部署流程：**
1. 运行测试 (test job)
2. 构建 Linux 二进制文件
3. 备份现有版本
4. 上传二进制和配置
5. 重启 systemd 服务
6. 健康检查

## 配置步骤

### 1. 生成 SSH 密钥

在本地机器上生成 SSH 密钥对（如果还没有）：

```bash
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_actions
```

### 2. 添加公钥到服务器

将公钥添加到服务器的 `authorized_keys`：

```bash
# 方式 1: 手动复制
cat ~/.ssh/github_actions.pub
# 然后登录服务器，添加到 ~/.ssh/authorized_keys

# 方式 2: 使用 ssh-copy-id
ssh-copy-id -i ~/.ssh/github_actions.pub root@47.239.21.188
```

### 3. 配置 GitHub Secrets

进入 GitHub 仓库页面：`Settings` → `Secrets and variables` → `Actions` → `New repository secret`

添加以下 Secret：

| Name | Value | 说明 |
|------|-------|------|
| `SSH_PRIVATE_KEY` | 私钥内容 | `~/.ssh/github_actions` 的内容 |

**获取私钥内容：**
```bash
cat ~/.ssh/github_actions
```

复制整个输出（包括 `-----BEGIN` 和 `-----END` 行）。

### 4. 测试 SSH 连接

在本地测试 SSH 密钥是否配置正确：

```bash
ssh -i ~/.ssh/github_actions root@47.239.21.188
```

如果无需密码即可登录，说明配置成功。

### 5. 触发部署

配置完成后，推送代码即可触发部署：

```bash
# 推送到 main 分支
git push origin main

# 或推送到 develop 分支
git push origin develop
```

### 6. 查看部署状态

在 GitHub 仓库页面：`Actions` 标签页查看部署进度和日志。

## 手动触发部署

在 GitHub 仓库页面：`Actions` → 选择工作流 → `Run workflow` 按钮。

## 监控和日志

### 前端部署日志
```bash
# Nginx 访问日志
ssh root@47.239.21.188 "sudo tail -f /var/log/nginx/access.log"

# Nginx 错误日志
ssh root@47.239.21.188 "sudo tail -f /var/log/nginx/error.log"
```

### 后端部署日志
```bash
# 应用日志
ssh root@47.239.21.188 "sudo tail -f /opt/ai-hms-backend/logs/app.log"

# 错误日志
ssh root@47.239.21.188 "sudo tail -f /opt/ai-hms-backend/logs/error.log"

# Systemd 日志
ssh root@47.239.21.188 "sudo journalctl -u ai-hms-backend -f"
```

## 服务管理

### 前端 (Nginx)
```bash
# 查看状态
ssh root@47.239.21.188 "sudo systemctl status nginx"

# 重启服务
ssh root@47.239.21.188 "sudo systemctl restart nginx"

# 测试配置
ssh root@47.239.21.188 "sudo nginx -t"
```

### 后端
```bash
# 查看状态
ssh root@47.239.21.188 "sudo systemctl status ai-hms-backend"

# 重启服务
ssh root@47.239.21.188 "sudo systemctl restart ai-hms-backend"

# 停止服务
ssh root@47.239.21.188 "sudo systemctl stop ai-hms-backend"
```

## 回滚操作

### 前端回滚
```bash
ssh root@47.239.21.188
# 列出备份
ls -la /var/www/ai-hms/backups/
# 恢复备份
sudo cp -r /var/www/ai-hms/backups/[备份目录]/* /var/www/ai-hms/current/
sudo systemctl reload nginx
```

### 后端回滚
```bash
ssh root@47.239.21.188
# 列出备份
ls -la /opt/ai-hms-backend/backups/
# 恢复备份
sudo cp /opt/ai-hms-backend/backups/[备份文件] /opt/ai-hms-backend/bin/ai-hms-server
sudo systemctl restart ai-hms-backend
```

## 访问地址

| 服务 | 地址 |
|------|------|
| 前端 | http://47.239.21.188 |
| 后端 API | http://47.239.21.188:8080 |
| API 文档 | http://47.239.21.188:8080/swagger/index.html |

## 故障排查

### 部署失败

1. 查看 Actions 日志，找到失败步骤
2. 检查服务器连接和权限
3. 检查服务状态和日志

### 健康检查失败

```bash
# 手动测试后端健康检查
curl http://47.239.21.188:8080/api/v1/health

# 手动测试前端
curl -I http://47.239.21.188
```

### 权限问题

确保 GitHub Actions 用户有足够权限：
```bash
# 服务器上检查目录权限
ls -la /var/www/ai-hms/
ls -la /opt/ai-hms-backend/
```
