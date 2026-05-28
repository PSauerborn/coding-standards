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
---

# Golang Dockerfile Standards

## 1. Dockerfile Guidelines

`[GO-DOCKER-001]` **SHOULD**: Dockerfiles should consist of three stages. The first stage should run unittests, the second stage should build the application, and the third stage should run the application.

`[GO-DOCKER-002]` **SHOULD**: Stages that run the application should use the `gcr.io/distroless/static:nonroot` image where possible. Where the application binary cannot run on distroless (e.g. it requires `libc` or other shared dependencies), fall back to a `debian:*-slim` image. Any stages that do not run the application should be based on the full golang image.

### Example 1

The following dockerfile illustrates a three-stage build for a Go application using the distroless runtime image.

```dockerfile
# GOOD: Use golang:1.25 as base image for tests
# GOOD: implement unittests as first stage
FROM golang:1.25 AS tests

WORKDIR /app/tests

COPY go.mod go.sum ./
RUN go mod download

RUN go install gotest.tools/gotestsum@latest

COPY etc ./etc
COPY *.go ./

CMD ["gotestsum", "--format", "testname"]

# GOOD: Use golang:1.25 as base image for build
# GOOD: implement build as second stage
FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o api .

# GOOD: Use gcr.io/distroless/static:nonroot as base image for runtime
# GOOD: fall back to debian:*-slim only if the binary cannot run on distroless
FROM gcr.io/distroless/static:nonroot AS runtime

WORKDIR /app

COPY --from=build /app/api .

COPY etc ./etc

CMD ["./api"]
```
