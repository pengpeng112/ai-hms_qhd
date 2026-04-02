# 阶段 1：Node.js 构建
FROM node:22-alpine AS builder

WORKDIR /app

# 复制 package 文件
COPY package.json package-lock.json ./

# 安装依赖（CI 模式，使用 package-lock.json 精确锁定）
RUN npm ci

# 复制源代码
COPY . .

# 构建生产产物
# VITE_API_BASE_URL 可在 docker build 时通过 --build-arg 注入
ARG VITE_API_BASE_URL=http://10.10.8.84:8080
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

# 阶段 2：Nginx 服务静态文件
FROM nginx:1.27-alpine

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/nginx-health || exit 1

CMD ["nginx", "-g", "daemon off;"]
