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
examples:
- examples/DOCKER/nginx-spa-config.md
- examples/DOCKER/spa-dockerfile.md
- examples/DOCKER/ssr-dockerfile.md
---

# Javascript Dockerfile Standards

## 1. Dockerfile Guidelines

`[JS-DOCKER-001]` **SHOULD**: Dockerfiles should consist of two stages. The first stage should build the application, and the second stage should run the application.

`[JS-DOCKER-002]` **SHOULD**: SPA applications should be served using the `nginx` base image. See `examples/DOCKER/nginx-spa-config.md` for an example NGINX config and `examples/DOCKER/spa-dockerfile.md` for a full Dockerfile example.

`[JS-DOCKER-003]` **SHOULD**: SSR applications should be served using the `node` base image. See `examples/DOCKER/ssr-dockerfile.md` for an example dockerfile.
