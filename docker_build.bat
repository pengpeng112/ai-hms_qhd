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
set TARFILE=%OUTPUT_DIR%\ai-hms-images.tar
set META_FILE=%OUTPUT_DIR%\ai-hms-images.meta.txt
set SHA_FILE=%OUTPUT_DIR%\ai-hms-images.tar.sha256
set PULL_RETRY_MAX=3
set BUILD_RETRY_MAX=3
set RETRY_WAIT_SEC=6

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

REM ---- [0.5] Pre-pull frontend base images with retry ----
echo [0.5/5] Pre-pulling frontend base images (with retry) ...
call :pull_with_retry node:22-alpine
if %errorlevel% neq 0 (
    echo [ERROR] Failed to pull node:22-alpine after %PULL_RETRY_MAX% attempts
    echo         Please check network / proxy / Docker Hub access, then retry.
    pause & exit /b 1
)
call :pull_with_retry nginx:1.27-alpine
if %errorlevel% neq 0 (
    echo [ERROR] Failed to pull nginx:1.27-alpine after %PULL_RETRY_MAX% attempts
    echo         Please check network / proxy / Docker Hub access, then retry.
    pause & exit /b 1
)
echo [OK] Frontend base images ready
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
    set "FRONTEND_BUILD_CMD=docker build --build-arg VITE_API_BASE_URL=%VITE_API_BASE_URL% -t %FRONTEND_IMAGE%:%IMAGE_TAG% ./ai-hms-frontend"
) else (
    echo       VITE_API_BASE_URL=default (from Dockerfile)
    set "FRONTEND_BUILD_CMD=docker build -t %FRONTEND_IMAGE%:%IMAGE_TAG% ./ai-hms-frontend"
)

set /a FRONTEND_BUILD_ATTEMPT=1
:frontend_build_retry
echo       build attempt !FRONTEND_BUILD_ATTEMPT!/%BUILD_RETRY_MAX%
cmd /c "!FRONTEND_BUILD_CMD!"
if !errorlevel! equ 0 goto frontend_build_ok

if !FRONTEND_BUILD_ATTEMPT! geq %BUILD_RETRY_MAX% goto frontend_build_failed
echo       [WARN] Frontend build failed, retrying in %RETRY_WAIT_SEC%s ...
timeout /t %RETRY_WAIT_SEC% /nobreak >nul
set /a FRONTEND_BUILD_ATTEMPT+=1
goto frontend_build_retry

:frontend_build_failed
    echo [ERROR] Frontend build failed
    pause & exit /b 1

:frontend_build_ok
echo [OK] Frontend image built
echo.

REM ---- [3/5] Show image sizes ----
echo [3/5] Image sizes:
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}" %BACKEND_IMAGE%:%IMAGE_TAG%
docker images --format "  {{.Repository}}:{{.Tag}}  {{.Size}}" %FRONTEND_IMAGE%:%IMAGE_TAG%

for /f "delims=" %%I in ('docker image inspect --format "{{.Id}}" %BACKEND_IMAGE%:%IMAGE_TAG%') do set BACKEND_ID=%%I
for /f "delims=" %%I in ('docker image inspect --format "{{.Created}}" %BACKEND_IMAGE%:%IMAGE_TAG%') do set BACKEND_CREATED=%%I
for /f "delims=" %%I in ('docker image inspect --format "{{.Id}}" %FRONTEND_IMAGE%:%IMAGE_TAG%') do set FRONTEND_ID=%%I
for /f "delims=" %%I in ('docker image inspect --format "{{.Created}}" %FRONTEND_IMAGE%:%IMAGE_TAG%') do set FRONTEND_CREATED=%%I

echo [3/5] Image identities:
echo   %BACKEND_IMAGE%:%IMAGE_TAG%  !BACKEND_ID!
echo     created: !BACKEND_CREATED!
echo   %FRONTEND_IMAGE%:%IMAGE_TAG%  !FRONTEND_ID!
echo     created: !FRONTEND_CREATED!
echo.

REM ---- [4/5] Export images to tar ----
echo [4/5] Exporting images to tar file ...
if exist "%TARFILE%" del /f "%TARFILE%"
if exist "%META_FILE%" del /f "%META_FILE%"
if exist "%SHA_FILE%" del /f "%SHA_FILE%"
docker save -o "%TARFILE%" %BACKEND_IMAGE%:%IMAGE_TAG% %FRONTEND_IMAGE%:%IMAGE_TAG%
if %errorlevel% neq 0 (
    echo [ERROR] Image export failed
    pause & exit /b 1
)
for %%A in ("%TARFILE%") do set TAR_SIZE=%%~zA
set /a TAR_MB=%TAR_SIZE% / 1048576

set TAR_SHA256=
for /f "tokens=* delims=" %%H in ('powershell -NoProfile -Command "(Get-FileHash -Algorithm SHA256 -Path ''%TARFILE%'').Hash.ToLower()"') do set TAR_SHA256=%%H
if not defined TAR_SHA256 (
    for /f "tokens=* delims=" %%H in ('certutil -hashfile "%TARFILE%" SHA256 ^| findstr /r /c:"^[0-9a-fA-F][0-9a-fA-F]"') do (
        for /f "tokens=* delims=" %%L in ('powershell -NoProfile -Command "''%%H''.ToLower().Trim()"') do set TAR_SHA256=%%L
    )
)

if "!TAR_SHA256!"=="" (
    echo [ERROR] Failed to calculate TAR SHA256
    pause & exit /b 1
)

(
    echo generated_at=%DATE% %TIME%
    echo backend_image=%BACKEND_IMAGE%:%IMAGE_TAG%
    echo backend_id=!BACKEND_ID!
    echo backend_created=!BACKEND_CREATED!
    echo frontend_image=%FRONTEND_IMAGE%:%IMAGE_TAG%
    echo frontend_id=!FRONTEND_ID!
    echo frontend_created=!FRONTEND_CREATED!
    echo tar_file=%TARFILE%
    echo tar_size_bytes=%TAR_SIZE%
    echo tar_size_mb=%TAR_MB%
    echo tar_sha256=!TAR_SHA256!
) > "%META_FILE%"

if defined TAR_SHA256 (
    > "%SHA_FILE%" (
        <nul set /p="!TAR_SHA256!  ai-hms-images.tar"
        echo.
    )
)

echo [OK] Exported: ai-hms-images.tar (%TAR_MB% MB)
echo [OK] Metadata: ai-hms-images.meta.txt
if defined TAR_SHA256 echo [OK] Checksum : ai-hms-images.tar.sha256
if defined TAR_SHA256 echo [OK] TAR SHA256: !TAR_SHA256!
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
copy /y "%META_FILE%"                "%DEPLOY_DIR%\" >nul
if exist "%SHA_FILE%" copy /y "%SHA_FILE%" "%DEPLOY_DIR%\" >nul
echo   [+] docker-compose.yml
echo   [+] docker_deploy.sh
echo   [+] docker_upgrade.sh
echo   [+] .env.production.template
echo   [+] ai-hms-images.meta.txt
if exist "%SHA_FILE%" echo   [+] ai-hms-images.tar.sha256

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
echo   scp %OUTPUT_DIR%\ai-hms-images.meta.txt root@10.10.8.84:/opt/
if exist "%SHA_FILE%" echo   scp %OUTPUT_DIR%\ai-hms-images.tar.sha256 root@10.10.8.84:/opt/
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
exit /b 0

:pull_with_retry
set "PULL_IMAGE=%~1"
set /a PULL_ATTEMPT=1

:pull_retry_loop
echo       pull !PULL_IMAGE! attempt !PULL_ATTEMPT!/%PULL_RETRY_MAX%
docker pull !PULL_IMAGE!
if !errorlevel! equ 0 exit /b 0

if !PULL_ATTEMPT! geq %PULL_RETRY_MAX% exit /b 1
echo       [WARN] Pull failed, retrying in %RETRY_WAIT_SEC%s ...
timeout /t %RETRY_WAIT_SEC% /nobreak >nul
set /a PULL_ATTEMPT+=1
goto pull_retry_loop
