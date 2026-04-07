# AI-HMS Docker 迁移 / 首次部署 / 升级操作手册（已结合真实部署问题修订）

本文档基于**当前仓库脚本**与**本次 10.10.8.84 实际部署过程**修订，适用于以下场景：

- Windows 开发机构建镜像
- openEuler 应用服务器通过 Docker 部署前后端
- PostgreSQL 独立部署在另一台服务器
- 首次空库部署、后续升级、回滚排查

> 本文档不是理想流程文档，而是**已结合真实踩坑后的可执行文档**。

---

## 1. 当前部署架构

### 1.1 服务器角色

#### 数据库服务器
- IP：`10.10.8.83`
- 数据库：`PostgreSQL`
- 数据库名：`ai_hms_db`
- 用户名：`admin`
- 密码：`admin123`

#### 应用服务器
- IP：`10.10.8.84`
- 系统：`openEuler 22.03 LTS-SP2`
- 运行内容：
  - 前端容器：`ai-hms-frontend`
  - 后端容器：`ai-hms-backend`

### 1.2 容器架构

```text
浏览器
  -> http://10.10.8.84:3000
  -> ai-hms-frontend (nginx:80)
  -> /api/ 反向代理
  -> ai-hms-backend (:8080)
  -> PostgreSQL 10.10.8.83:5432
```

### 1.3 当前镜像与端口

- 后端镜像：`ai-hms-backend:latest`
- 前端镜像：`ai-hms-frontend:latest`
- 后端容器：`ai-hms-backend`
- 前端容器：`ai-hms-frontend`
- 前端端口：`3000`
- 后端端口：`8080`

---

## 2. 这次真实部署暴露出的关键问题

本次部署中，实际遇到了以下问题，后续部署时必须提前规避：

### 2.1 `.env.production.template` 默认数据库参数与真实环境不一致

模板文件里的默认值曾写成：

- `DB_USER=amdin`
- `DB_NAME=Postgre`

而当前真实环境实际可用的是：

- `DB_USER=admin`
- `DB_NAME=ai_hms_db`

所以：

> **不要直接信任模板默认数据库参数，首次部署必须人工核对 `.env`。**

### 2.2 `docker-compose.yml` 默认强制后端使用 `GIN_MODE=release`

当前仓库里的 `docker-compose.yml` 明确写了：

```yaml
environment:
  - GIN_MODE=release
```

而后端代码在 `release` 模式下会明确跳过 AutoMigrate：

```text
Production environment: skipping AutoMigrate
```

这意味着：

> **首次部署到空库时，后端不会自动建表。**

### 2.3 `docker_deploy.sh` 的 seed 顺序对空库不成立

当前 `docker_deploy.sh` 的顺序是：

1. 检查镜像
2. 生成 `.env`
3. 检查数据库连通性
4. **先执行 seed**
5. **再启动容器**

但空库第一次部署时，表还不存在，seed 会报：

```text
relation "users" does not exist
relation "patients" does not exist
...
```

所以：

> **空库首次部署时，不应该先执行 seed。**

### 2.4 `docker_deploy.sh` 的健康检查只代表“服务活着”，不代表“数据已就绪”

后端 `/health` 只是活性检查，不会验证：

- 表是否存在
- seed 是否成功
- 字典是否初始化成功

所以出现了这种情况：

- 容器健康
- 页面可打开
- 但库里没有表/没有种子数据

### 2.5 Docker 网络可能因宿主机环境冲突而自动分配失败

真实部署中遇到过：

```text
could not find an available, non-overlapping IPv4 address pool
```

因此建议：

> **部署时给 compose 网络明确指定固定 subnet。**

---

## 3. 当前推荐的标准流程（已经过本次环境验证）

### 总体原则

当前仓库脚本可用，但对于**空库首次部署**，必须采用以下顺序：

1. Windows 构建镜像并打包
2. 上传到服务器
3. 配置 `.env`
4. 修正 `docker-compose.yml` 网络配置
5. **临时以 debug 模式启动后端一次，用于 AutoMigrate 建表**
6. 手工执行 `seed_phase1.sql`
7. 再切回 release 正式运行

> 这是当前仓库在“生产空库首次部署”条件下最稳妥的办法。

---

## 4. Windows 开发机：构建与打包

假设项目目录：

```text
F:\python\前后端代码\ai-hms_qhd
```

### 4.1 进入项目目录

```bat
cd /d F:\python\前后端代码\ai-hms_qhd
```

### 4.2 执行构建脚本

```bat
docker_build.bat
```

### 4.3 构建产物位置

`docker_build.bat` 当前将产物输出到项目上一级目录。

例如项目目录是：

```text
F:\python\前后端代码\ai-hms_qhd
```

则产物一般在：

```text
F:\python\前后端代码\ai-hms-images.tar
F:\python\前后端代码\ai-hms-docker\
```

### 4.4 检查镜像

```bat
docker images | findstr ai-hms
```

---

## 5. 将文件迁移到 10.10.8.84

推荐目标目录：`/opt`

### 5.1 scp 传输

```bat
scp F:\python\前后端代码\ai-hms-images.tar root@10.10.8.84:/opt/
scp -r F:\python\前后端代码\ai-hms-docker root@10.10.8.84:/opt/
```

### 5.2 目标目录结构

```text
/opt/ai-hms-images.tar
/opt/ai-hms-docker/
├── docker-compose.yml
├── docker_deploy.sh
├── docker_upgrade.sh
├── .env.production.template
├── logs/
└── scripts/
    └── seed_phase1.sql
```

---

## 6. openEuler 服务器准备

以下命令在 `10.10.8.84` 执行。

### 6.1 安装基础环境

```bash
dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin curl postgresql nmap-ncat
```

如果你的环境使用 `docker-compose` 独立二进制，也可以继续使用现有版本。

### 6.2 启动 Docker

```bash
systemctl enable --now docker
systemctl status docker --no-pager
```

### 6.3 验证版本

```bash
docker --version
docker compose version || docker-compose version
```

---

## 7. 首次部署前必须人工修改的配置

### 7.1 导入镜像

```bash
cd /opt
docker load -i ai-hms-images.tar
docker images | grep ai-hms
```

### 7.2 检查并修正 `.env`

进入部署目录：

```bash
cd /opt/ai-hms-docker
```

如果 `.env` 不存在，先运行一次：

```bash
bash docker_deploy.sh
```

当脚本第一次生成 `.env` 后，先不要急着继续，先编辑：

```bash
vi /opt/ai-hms-docker/.env
```

确认至少以下内容：

```env
DB_HOST=10.10.8.83
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=admin123
DB_NAME=ai_hms_db
DB_SSL_MODE=disable

SERVER_HOST=0.0.0.0
SERVER_PORT=8080
GIN_MODE=release

CORS_ALLOWED_ORIGINS=http://10.10.8.84:3000
VITE_API_BASE_URL=http://10.10.8.84:8080
```

### 7.3 强烈建议给 Docker 网络固定子网

编辑：

```bash
vi /opt/ai-hms-docker/docker-compose.yml
```

把最后的：

```yaml
networks:
  ai-hms-net:
    driver: bridge
```

改成：

```yaml
networks:
  ai-hms-net:
    driver: bridge
    ipam:
      config:
        - subnet: 172.30.88.0/24
```

> 如果宿主机路由表已占用 `172.30.x.x`，请换一个不冲突的私网段。

可用以下命令检查：

```bash
ip route show
docker network ls
```

---

## 8. 首次空库部署：正确步骤（非常重要）

### 8.1 先验证数据库连接是否正确

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c '\conninfo'
```

如果能正常返回连接信息，再继续。

### 8.2 先不要执行 seed

原因：当前仓库里的 seed 脚本依赖表已经存在，而当前生产部署默认不会自动建表。

### 8.3 备份 compose 文件

```bash
cd /opt/ai-hms-docker
cp docker-compose.yml docker-compose.yml.bak.$(date +%Y%m%d_%H%M%S)
```

### 8.4 临时把后端模式改成 debug，用于建表

```bash
sed -i 's/GIN_MODE=release/GIN_MODE=debug/' /opt/ai-hms-docker/docker-compose.yml
grep -n 'GIN_MODE' /opt/ai-hms-docker/docker-compose.yml
```

### 8.5 只启动/重建后端容器，让它执行 AutoMigrate

```bash
cd /opt/ai-hms-docker
docker compose --env-file .env up -d --no-deps --force-recreate backend || docker-compose -f docker-compose.yml --env-file .env up -d --no-deps --force-recreate backend
```

查看日志：

```bash
docker logs --tail 200 ai-hms-backend
```

确认不要再看到：

```text
Production environment: skipping AutoMigrate
```

### 8.6 确认表已经创建成功

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c '\dt'
```

如果能看到：

- `users`
- `patients`
- `patient_basic_infos`
- `vascular_accesses`
- `infection_infos`
- `treatment_plans`
- `medical_histories`

以及其他业务表，说明建表完成。

### 8.7 手工执行 seed（必须开启 `ON_ERROR_STOP`）

```bash
cd /opt/ai-hms-docker
PGPASSWORD='admin123' psql -v ON_ERROR_STOP=1 -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -f scripts/seed_phase1.sql
```

### 8.8 验证 seed 是否成功

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c "select username, real_name, role, status from users;"
```

当前种子里至少应看到：

- `test_admin`（角色 `ADMIN`）
- `test_doctor`（角色 `DOCTOR_CHIEF`）

### 8.9 改回正式模式 release

```bash
sed -i 's/GIN_MODE=debug/GIN_MODE=release/' /opt/ai-hms-docker/docker-compose.yml
grep -n 'GIN_MODE' /opt/ai-hms-docker/docker-compose.yml
```

### 8.10 重建后端并正式运行整套服务

```bash
cd /opt/ai-hms-docker
docker compose --env-file .env up -d || docker-compose -f docker-compose.yml --env-file .env up -d
```

---

## 9. 如果数据库不是空库

如果数据库里已经存在完整表结构，则可以直接走：

```bash
cd /opt/ai-hms-docker
bash docker_deploy.sh
```

但仍然建议你先确认：

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c '\dt'
```

有表再 seed，才是安全顺序。

---

## 10. 当前 `docker_deploy.sh` 的已知缺陷（必须知道）

### 10.1 空库首次部署时，seed 顺序不对

脚本会先 seed，再启动容器。对于空库，这是不成立的。

### 10.2 seed 没有使用 `-v ON_ERROR_STOP=1`

这会导致：

- 中途大量 SQL 错误继续执行
- 最后脚本仍可能显示成功/警告

### 10.3 健康检查不能代表业务初始化成功

即使看到：

```text
Backend : ✓ running
Frontend: ✓ running
```

也不能说明：

- 表已创建
- seed 已成功
- 字典已初始化完成

所以首次部署后必须手工验证数据库。

---

## 11. 当前 `docker_upgrade.sh` 的使用方式

后续升级时，如果数据库表结构已经存在，则流程会简单很多。

### 11.1 Windows 重新构建

```bat
cd /d F:\python\前后端代码\ai-hms_qhd
docker_build.bat
```

### 11.2 重新上传镜像包与部署目录

```bat
scp F:\python\前后端代码\ai-hms-images.tar root@10.10.8.84:/opt/
scp -r F:\python\前后端代码\ai-hms-docker root@10.10.8.84:/opt/
```

### 11.3 服务器导入新镜像

```bash
cd /opt
docker load -i ai-hms-images.tar
```

### 11.4 执行升级脚本

升级全部：

```bash
cd /opt/ai-hms-docker
 bash /opt/ai-hms-docker/docker_upgrade.sh
```

仅升级后端：

```bash
bash docker_upgrade.sh backend
```

仅升级前端：

```bash
bash docker_upgrade.sh frontend
```

---

## 12. 当前环境可用的测试账号

种子成功导入后，可用以下账号测试：

### 管理员账号

```text
用户名: test_admin
密码:   Test@123456
角色:   ADMIN
```

### 医生账号

```text
用户名: test_doctor
密码:   Test@123456
角色:   DOCTOR_CHIEF
```

其中当前系统里权限最大的测试管理员账号是：

```text
test_admin / Test@123456
```

---

## 13. 部署完成后的验证命令

### 13.1 容器状态

```bash
cd /opt/ai-hms-docker
docker compose ps || docker-compose -f docker-compose.yml ps
```

### 13.2 后端日志

```bash
docker logs --tail 200 ai-hms-backend
```

### 13.3 前端日志

```bash
docker logs --tail 200 ai-hms-frontend
```

### 13.4 健康检查

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:3000/nginx-health
```

### 13.5 检查管理员账号是否存在

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c "select username, real_name, role, status from users;"
```

### 13.6 浏览器验证

访问：

```text
http://10.10.8.84:3000
```

登录：

```text
test_admin / Test@123456
```

验证：

- 登录成功
- 患者列表可打开
- 患者详情可打开

---

## 14. 常见问题排查

### 14.1 `password authentication failed`

优先检查 `.env`：

```bash
grep -E '^(DB_HOST|DB_PORT|DB_USER|DB_PASSWORD|DB_NAME)=' /opt/ai-hms-docker/.env
```

再手工验证：

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c '\conninfo'
```

### 14.2 `database does not exist`

说明 `DB_NAME` 配置错了。先查库清单：

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d postgres -c '\l'
```

### 14.3 `could not find an available, non-overlapping IPv4 address pool`

说明 Docker 自动分配网络失败。优先处理方式不是删全局 Docker，而是给当前 compose 指定固定 subnet。

### 14.4 `relation "xxx" does not exist`

说明表还没建出来。对空库首次部署，要先走：

1. 临时 debug 启动后端
2. AutoMigrate 建表
3. 再执行 seed

### 14.5 后端健康但数据库为空

这是当前仓库脚本的真实行为之一。必须手工执行：

```bash
PGPASSWORD='admin123' psql -h 10.10.8.83 -p 5432 -U admin -d ai_hms_db -c '\dt'
```

---

## 15. 当前这套部署方案的总结

当前已经验证可用的方案特征如下：

- 前后端分两个容器
- 数据库独立部署在 10.10.8.83
- 适合后续独立升级 frontend / backend
- 首次空库部署时，需要一次性 debug 建表
- 建表完成后即可长期使用 release 正式运行

这套方式符合当前项目状态，也符合后续迭代升级要求。
