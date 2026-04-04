@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion
cd /d "%~dp0"

REM ============================================================
REM   AI-HMS - Docker 镜像构建 + 打包脚本（Windows 开发机上运行）
REM   前后端分为两个独立镜像，便于独立升级
REM
REM   产物：
REM     ai-hms-images.tar        两个镜像合并的 tar 包
REM     ai-hms-docker\            部署目录（含 compose + 脚本）
REM
REM   用法：双击运行或在 cmd 中 docker_build.bat
REM ============================================================

set BACKEND_IMAGE=ai-hms-backend
set FRONTEND_IMAGE=ai-hms-frontend
set IMAGE_TAG=latest
set OUTPUT_DIR=%~dp0..
set DEPLOY_PKG=ai-hms-docker

REM ---- 可选: 前端 API 地址（构建时注入）----
REM 如需自定义，取消注释并修改；默认值已写在 Dockerfile ARG 中
REM set VITE_API_BASE_URL=http://10.10.8.84:8080

echo.
echo ============================================================
echo   AI-HMS - Docker Image Build ^& Export
echo   Backend : %BACKEND_IMAGE%:%IMAGE_TAG%
echo   Frontend: %FRONTEND_IMAGE%:%IMAGE_TAG%
echo ============================================================
echo.

REM ---- [0] 检查 Docker ----
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker not found. Please install Docker Desktop.
    pause & exit /b 1
)
echo [OK] Docker ready
echo.

REM ---- [1/5] 构建后端镜像 ----
echo [1/5] Building backend image %BACKEND_IMAGE%:%IMAGE_TAG% ...
echo       (multi-stage: golang:1.24-alpine -^> alpine:latest)
echo.
docker build -t %BACKEND_IMAGE%:%IMAGE_TAG% ./ai-hms-backend
if %errorlevel% neq 0 (
    echo [ERROR] Backend build failed
    pause & exit /b 1
)
echo [OK] Backend image built
echo.

REM ---- [2/5] 构建前端镜像 ----
echo [2/5] Building frontend image %FRONTEND_IMAGE%:%IMAGE_TAG% ...
echo       (multi-stage: node:22-alpine -^> nginx:1.27-alpine)
echo.

if defined VITE_API_BASE_URL (
    echo       VITE_API_BASE_URL=%VITE_API_BASE_URL%
    docker build --build-arg VITE_API_BASE_URL=%VITE_API_BASE_URL% -t %FRONTEND_IMAGE%:%IMAGE_TAG% ./ai-hms-frontend
) else (
    echo       VITE_API_BASE_URL=default (from Dockerfile)
    docker build -t %FRONTEND_IMAGE%:%IMAGE_TAG% ./ai-hms-frontend
)
if %errorlevel% neq 0 (
    echo [ERROR] Frontend build failed
    pause & exit /b 1
)
echo [OK] Frontend image built
echo.

REM ---- [3/5] 显示镜像大小 ----
echo [3/5] Image sizes:
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}" %BACKEND_IMAGE%:%IMAGE_TAG%
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}" %FRONTEND_IMAGE%:%IMAGE_TAG%
echo.

REM ---- [4/5] 导出镜像为 tar ----
echo [4/5] Exporting images to tar file ...
set TARFILE=%OUTPUT_DIR%\ai-hms-images.tar
if exist "%TARFILE%" del /f "%TARFILE%"
docker save -o "%TARFILE%" %BACKEND_IMAGE%:%IMAGE_TAG% %FRONTEND_IMAGE%:%IMAGE_TAG%
if %errorlevel% neq 0 (
    echo [ERROR] Image export failed
    pause & exit /b 1
)
for %%A in ("%TARFILE%") do set TAR_SIZE=%%~zA
set /a TAR_MB=%TAR_SIZE% / 1048576
echo [OK] Exported: ai-hms-images.tar (%TAR_MB% MB)
echo.

REM ---- [5/5] 打包部署文件 ----
echo [5/5] Packaging deployment files ...
set DEPLOY_DIR=%OUTPUT_DIR%\%DEPLOY_PKG%
if exist "%DEPLOY_DIR%" rmdir /s /q "%DEPLOY_DIR%"
mkdir "%DEPLOY_DIR%"

REM 核心部署文件
copy /y "docker-compose.yml"         "%DEPLOY_DIR%\" >nul
copy /y "docker_deploy.sh"           "%DEPLOY_DIR%\" >nul
copy /y "docker_upgrade.sh"          "%DEPLOY_DIR%\" >nul
copy /y ".env.production.template"   "%DEPLOY_DIR%\" >nul
echo   [+] docker-compose.yml
echo   [+] docker_deploy.sh
echo   [+] docker_upgrade.sh
echo   [+] .env.production.template

REM 数据库种子脚本
mkdir "%DEPLOY_DIR%\scripts" >nul 2>&1
copy /y "ai-hms-backend\scripts\seed_phase1.sql"  "%DEPLOY_DIR%\scripts\" >nul
echo   [+] scripts\seed_phase1.sql

REM 创建空目录（服务器上挂载用）
mkdir "%DEPLOY_DIR%\logs" >nul 2>&1

REM LF 转换（避免 Windows CRLF 在 Linux 上出 $'\r' 错误）
python -c "import os,sys; d=sys.argv[1]; [open(os.path.join(d,f),'wb').write(open(os.path.join(d,f),'rb').read().replace(b'\r\n',b'\n')) for f in os.listdir(d) if f.endswith(('.sh','.yml','.yaml','.env','.template'))]" "%DEPLOY_DIR%" 2>nul
if exist "%DEPLOY_DIR%\scripts" (
    python -c "import os,sys; d=sys.argv[1]; [open(os.path.join(d,f),'wb').write(open(os.path.join(d,f),'rb').read().replace(b'\r\n',b'\n')) for f in os.listdir(d) if f.endswith(('.sh','.sql'))]" "%DEPLOY_DIR%\scripts" 2>nul
)

echo.
echo ============================================================
echo   [DONE] Build complete!
echo ============================================================
echo.
echo   Image tar  : %TARFILE% (%TAR_MB% MB)
echo   Deploy dir : %DEPLOY_DIR%\
echo.
echo   ---- Transfer to server (10.10.8.84) ----
echo   scp %OUTPUT_DIR%\ai-hms-images.tar root@10.10.8.84:/opt/
echo   scp -r %OUTPUT_DIR%\%DEPLOY_PKG% root@10.10.8.84:/opt/
echo.
echo   ---- On server: first deploy ----
echo   cd /opt
echo   docker load -i ai-hms-images.tar
echo   cd ai-hms-docker
echo   bash docker_deploy.sh
echo.
echo   ---- On server: upgrade ----
echo   docker load -i ai-hms-images.tar
echo   cd /opt/ai-hms-docker
echo   bash docker_upgrade.sh
echo ============================================================
echo.
pause
