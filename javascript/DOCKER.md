---
title: Javascript Dockerfile Standards
description: Standards for writing Dockerfiles for Javascript applications.
parent: general/DOCKER.md
scope:
- '*'
topics:
- javascript
- docker
- dockerfile
- container
---

# Javascript Dockerfile Standards

## 1. Dockerfile Guidelines

`[JS-DOCKER-001]` **SHOULD**: Dockerfiles should consist of two stages. The first stage should build the application, and the second stage should run the application.

`[JS-DOCKER-002]` **SHOULD**: SPA applications should be served using the `nginx` base image. See `Example 1` for an example NGINX config and `Example 2` for a full Dockerfile example.

`[JS-DOCKER-002]` **SHOULD**: SSR applications should be served using the `node` base image. See `Example 3` for an example dockerfile.


### Example 1

```nginx
server {
    listen 8080;
    server_name localhost;

    root /usr/share/nginx/html;

    index index.html;

    location / {
        # This is critical for SPAs (History Mode).
        # If the requested file isn't found, fall back to index.html
        try_files $uri $uri/ /index.html;
    }

    # Basic error handling
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

### Example 2

```dockerfile
FROM node:22 AS build

ARG BUILD_ENV=prod

WORKDIR /app

RUN npm install -g @quasar/cli
RUN rm -f package-lock.json

COPY package.json yarn.lock ./

RUN yarn install --ignore-scripts

COPY . .

RUN BUILD_ENV=${BUILD_ENV} quasar build

FROM nginx

COPY --from=build /app/dist/spa /usr/share/nginx/html

COPY etc/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

```

### Example 3

```dockerfile
FROM node:22 AS build

ENV NODE_ENV=development

WORKDIR /app

RUN npm install -g @quasar/cli
RUN rm -f package-lock.json

COPY package.json yarn.lock ./

RUN yarn install --ignore-scripts

COPY . .

RUN quasar build -m ssr

FROM node:22-alpine AS runtime

ENV NODE_ENV=production
ENV PORT=8080

WORKDIR /app

# Quasar's SSR build outputs its own minimal Node project in dist/ssr
# (including a package.json and entry index.js)
COPY --from=build /app/dist/ssr ./

# Install only production dependencies for the SSR server
RUN yarn install --production --ignore-scripts

EXPOSE 8080

# Start the SSR server
CMD ["node", "index.js"]

```