# syntax=docker/dockerfile:1.7
FROM node:25.9.0-alpine AS builder

WORKDIR /app

# Cache node_modules separately — audit fails on high/critical CVEs
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run typecheck && npm run build

# Minimal nginx image
FROM nginx:1.27-alpine

# Remove default configs
RUN rm -f /etc/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf.template

COPY --from=builder /app/dist /usr/share/nginx/html
# Main nginx.conf (http block + include conf.d)
COPY nginx.conf /etc/nginx/nginx.conf
# Server block template — envsubst'd at start via nginx entrypoint
COPY nginx-server.conf.template /etc/nginx/templates/default.conf.template

# Default backend URL (overridden per environment via env var BACKEND_URL)
ENV BACKEND_URL=http://backend:8080

# Make dirs writable by nginx so the entrypoint envsubst can write conf
# and nginx can write tmp/cache at runtime.
RUN chown -R nginx:nginx /usr/share/nginx/html \
    && chmod -R 755 /usr/share/nginx/html \
    && chown -R nginx:nginx /etc/nginx/conf.d \
    && chown -R nginx:nginx /etc/nginx/templates

USER nginx

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
