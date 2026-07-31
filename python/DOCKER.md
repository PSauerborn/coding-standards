---
title: Python Dockerfile Standards
description: Standards for writing Dockerfiles for Python applications.
parent: general/DOCKER.md
scope:
- '*'
topics:
- python
- docker
- dockerfile
- container
examples:
- examples/DOCKER/two-stage-build.md
---

# Python Dockerfile Standards

## 1. Dockerfile Guidelines

`[PY-DOCKER-001]` **SHOULD**: Dockerfiles should consist of two stages. The first stage should run unittests, the second stage should build the application and run the application. See `examples/DOCKER/two-stage-build.md` for an illustration.

`[PY-DOCKER-002]` **SHOULD**: The runtime image should be based on a slim image. Prefer debian based images over alpine based images. This avoids issues with missing C libraries and bindings when installing dependencies. See `examples/DOCKER/two-stage-build.md` for an illustration.
