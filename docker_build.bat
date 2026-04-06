@echo off
chcp 65001 >nul 2>&1
setlocal EnableDelayedExpansion
cd /d "%~dp0"

REM ============================================================
REM   AI-HMS - Docker Image Build + Export (Run on Windows dev)
REM   Two separate images: backend + frontend
REM
REM   Output:
REM     ai-hms-images.tar         combined image tar
REM     ai-hms-docker\            deploy directory (compose + scripts)
REM
REM   Usage: double-click or run docker_build.bat in cmd
REM ============================================================

set BACKEND_IMAGE=ai-hms-backend
set FRONTEND_IMAGE=ai-hms-frontend
set IMAGE_TAG=latest
set OUTPUT_DIR=%~dp0..
set DEPLOY_PKG=ai-hms-docker

REM ---- Optional: frontend API URL (injected at build time) ----
REM Uncomment and modify if needed; default is in Dockerfile ARG
REM set VITE_API_BASE_URL=http://10.10.8.84:8080

echo.
echo ============================================================
echo   AI-HMS - Docker Image Build ^& Export
echo   Backend : %BACKEND_IMAGE%:%IMAGE_TAG%
echo   Frontend: %FRONTEND_IMAGE%:%IMAGE_TAG%
echo ============================================================
echo.

REM ---- [0] Check Docker ----
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker not found. Please install Docker Desktop.
    pause & exit /b 1
)
echo [OK] Docker ready
echo.

REM ---- [1/5] Build backend image ----
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

REM ---- [2/5] Build frontend image ----
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

REM ---- [3/5] Show image sizes ----
echo [3/5] Image sizes:
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}" %BACKEND_IMAGE%:%IMAGE_TAG%
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}" %FRONTEND_IMAGE%:%IMAGE_TAG%
echo.

REM ---- [4/5] Export images to tar ----
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

REM ---- [5/5] Package deployment files ----
echo [5/5] Packaging deployment files ...
set DEPLOY_DIR=%OUTPUT_DIR%\%DEPLOY_PKG%
if exist "%DEPLOY_DIR%" rmdir /s /q "%DEPLOY_DIR%"
mkdir "%DEPLOY_DIR%"

REM Core deployment files
copy /y "docker-compose.yml"         "%DEPLOY_DIR%\" >nul
copy /y "docker_deploy.sh"           "%DEPLOY_DIR%\" >nul
copy /y "docker_upgrade.sh"          "%DEPLOY_DIR%\" >nul
copy /y ".env.production.template"   "%DEPLOY_DIR%\" >nul
echo   [+] docker-compose.yml
echo   [+] docker_deploy.sh
echo   [+] docker_upgrade.sh
echo   [+] .env.production.template

REM Database seed script
mkdir "%DEPLOY_DIR%\scripts" >nul 2>&1
copy /y "ai-hms-backend\scripts\seed_phase1.sql"  "%DEPLOY_DIR%\scripts\" >nul
echo   [+] scripts\seed_phase1.sql

REM Create empty dir for server-side volume mounts
mkdir "%DEPLOY_DIR%\logs" >nul 2>&1

REM Convert CRLF to LF (avoid $'\r' errors on Linux)
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
