---
title: Dockerfile General Standards
description: General standards for Dockerfiles.
scope:
- '*'
parent: GENERAL.md
topics:
- docker
- dockerfile
- container
- deployment
---

# Dockerfile General Standards

## 1. General Guidelines

`[DOCKER-001]` **MUST**: Dockerfiles must be provided for all applications.

`[DOCKER-002]` **MUST**: Dockerfiles must be implemented as multi-stage builds. This keeps the final image small by ensuring build-time dependencies do not leak into the runtime image.

`[DOCKER-003]` **MUST**: Non-essential files must be excluded from the final image.

`[DOCKER-004]` **MUST**: Images must be built for AMD linux architecture. Use the `--platform linux/amd64` flag to specify the architecture when building the image. Additionally, the `--provenance=false` flag must be used to disable provenance.

`[DOCKER-005]` **SHOULD**: A dedicated stage should run the application's unit tests. This ensures tests are exercised as part of the image build.

`[DOCKER-006]` **SHOULD**: The runtime image should be as small as possible. Prefer the smallest viable base image that still supports the application's runtime dependencies.
