# [JS-DOCKER-003] SSR Dockerfile

Statements: `[JS-DOCKER-001]` `[JS-DOCKER-003]`

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
