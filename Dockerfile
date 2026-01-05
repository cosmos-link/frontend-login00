# 阶段1：构建阶段 - 复制配置文件读取工具
FROM golang:1.24-alpine AS config-builder

WORKDIR /app

# 复制配置读取脚本（用于构建时读取config.ini）
COPY scripts/get_config.go ./
RUN go build -o get_config ./get_config.go

# 阶段2：运行阶段 - Nginx静态文件服务
FROM nginx:alpine

# 构建参数（从config.ini读取，默认值做兜底）
ARG APP_NAME=DID-Client
ARG CONTAINER_PORT=50100

# 安装基础依赖（时区、wget）
RUN apk --no-cache add tzdata ca-certificates wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata

# 删除默认Nginx静态文件和配置
RUN rm -rf /usr/share/nginx/html/* && \
    rm -f /etc/nginx/conf.d/default.conf

# 创建自定义Nginx配置
RUN echo 'server {' > /etc/nginx/conf.d/default.conf && \
    echo '    listen 50107;' >> /etc/nginx/conf.d/default.conf && \
    echo '    server_name _;' >> /etc/nginx/conf.d/default.conf && \
    echo '    root /usr/share/nginx/html;' >> /etc/nginx/conf.d/default.conf && \
    echo '    index index.html index.htm;' >> /etc/nginx/conf.d/default.conf && \
    echo '' >> /etc/nginx/conf.d/default.conf && \
    echo '    # Gzip 压缩' >> /etc/nginx/conf.d/default.conf && \
    echo '    gzip on;' >> /etc/nginx/conf.d/default.conf && \
    echo '    gzip_vary on;' >> /etc/nginx/conf.d/default.conf && \
    echo '    gzip_min_length 1024;' >> /etc/nginx/conf.d/default.conf && \
    echo '    gzip_types text/plain text/css text/xml text/javascript ' >> /etc/nginx/conf.d/default.conf && \
    echo '               application/x-javascript application/xml+rss ' >> /etc/nginx/conf.d/default.conf && \
    echo '               application/json application/javascript;' >> /etc/nginx/conf.d/default.conf && \
    echo '' >> /etc/nginx/conf.d/default.conf && \
    echo '    # 缓存静态资源' >> /etc/nginx/conf.d/default.conf && \
    echo '    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {' >> /etc/nginx/conf.d/default.conf && \
    echo '        expires 30d;' >> /etc/nginx/conf.d/default.conf && \
    echo '        add_header Cache-Control "public, immutable";' >> /etc/nginx/conf.d/default.conf && \
    echo '    }' >> /etc/nginx/conf.d/default.conf && \
    echo '' >> /etc/nginx/conf.d/default.conf && \
    echo '    # 所有请求都路由到index.html（SPA模式）' >> /etc/nginx/conf.d/default.conf && \
    echo '    location / {' >> /etc/nginx/conf.d/default.conf && \
    echo '        try_files $uri $uri/ /index.html;' >> /etc/nginx/conf.d/default.conf && \
    echo '    }' >> /etc/nginx/conf.d/default.conf && \
    echo '}' >> /etc/nginx/conf.d/default.conf

# 从构建阶段复制配置读取工具（可选，用于调试）
COPY --from=config-builder /app/get_config /usr/local/bin/

# 复制前端静态文件到Nginx目录
COPY src/ /usr/share/nginx/html/

# 赋予Nginx用户对静态文件的访问权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

# 声明容器内部端口（Nginx内部端口）
EXPOSE 50107

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

# 使用Nginx前台运行
CMD ["nginx", "-g", "daemon off;"]