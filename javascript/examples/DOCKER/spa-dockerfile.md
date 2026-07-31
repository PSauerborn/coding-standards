# [JS-DOCKER-002] SPA Dockerfile

Statements: `[JS-DOCKER-001]` `[JS-DOCKER-002]`

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
