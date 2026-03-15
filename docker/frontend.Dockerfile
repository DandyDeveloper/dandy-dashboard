# syntax=docker/dockerfile:1

# Stage 1 — Build
FROM node:24.14.0-alpine AS builder
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# Stage 2 — Runtime (Nginx)
FROM nginx:1.29-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
RUN chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx \
 && chmod -R 755 /usr/share/nginx/html

USER nginx
EXPOSE 80
