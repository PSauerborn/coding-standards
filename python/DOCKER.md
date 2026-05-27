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
---

# Python Dockerfile Standards

## 1. Dockerfile Guidelines

`[PY-DOCKER-001]` **SHOULD**: Dockerfiles should consist of two stages. The first stage should run unittests, the second stage should build the application and run the application. See `Example 1` for an illustration.

`[PY-DOCKER-002]` **SHOULD**: The runtime image should be based on a slim image. Prefer debian based images over alpine based images. This avoids issues with missing C libraries and bindings when installing dependencies. See `Example 1` for an illustration.

### Example 1

The following example shows a minimal dockerfile for a python application that implements a `tests` and a `runtime` stage. Note that a smaller runtime image is used.

```dockerfile
# GOOD: Use bookworm as base image for tests
FROM python:3.13-bookworm AS tests

COPY requirements.txt ./
COPY tests/requirements.txt ./requirements-tests.txt

RUN pip install -U pip && \
    pip install -r requirements.txt && \
    pip install -r requirements-tests.txt

COPY src ./
COPY tests ./tests

CMD ["pytest", "-vv"]

# GOOD: Use slim as base image for runtime
FROM python:3.13-slim AS runtime

COPY requirements.txt ./

RUN pip install -U pip && \
    pip install -r requirements.txt

COPY src ./

CMD ["python", "src/main.py"]
```
