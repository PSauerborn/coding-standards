---
title: Golang Dockerfile Standards
description: Standards for writing Dockerfiles for Go applications.
parent: general/DOCKER.md
scope:
- '*'
topics:
- golang
- docker
- dockerfile
- container
examples:
- examples/DOCKER/three-stage-build.md
---

# Golang Dockerfile Standards

## 1. Dockerfile Guidelines

`[GO-DOCKER-001]` **SHOULD**: Dockerfiles should consist of three stages. The first stage should run unittests, the second stage should build the application, and the third stage should run the application. See `examples/DOCKER/three-stage-build.md` for an illustration.

`[GO-DOCKER-002]` **SHOULD**: Stages that run the application should use the `gcr.io/distroless/static:nonroot` image where possible. Where the application binary cannot run on distroless (e.g. it requires `libc` or other shared dependencies), fall back to a `debian:*-slim` image. Any stages that do not run the application should be based on the full golang image. See `examples/DOCKER/three-stage-build.md` for an illustration.
