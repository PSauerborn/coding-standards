---
title: Golang REST API Standards
description: Standards for writing REST APIs in Go.
parent: golang/GENERAL.md
scope:
- '*.go'
topics:
- golang
- api
- rest
- gin-gonic
examples:
- examples/API/controller-api.md
- examples/API/api-anti-pattern.md
---

# Golang REST API Standards

## 1. REST API Guidelines

`[GO-API-001]` **SHOULD**: REST APIs should be implemented using the `github.com/gin-gonic/gin` package.

`[GO-API-002]` **SHOULD**: REST APIs should be implemented using a controller pattern. The controller should contain singletons for database clients, service clients, etc. This is especially important when using connection pools to ensure that a single connection pool is shared across all endpoints. See `examples/API/controller-api.md` for an illustration and `examples/API/api-anti-pattern.md` for an anti-pattern to avoid.

`[GO-API-003]` **SHOULD**: Packages defining REST APIs should have a `NewRouter` constructor that returns a new instance of the router with all plugins and endpoints registered. See `examples/API/controller-api.md` for an illustration and `examples/API/api-anti-pattern.md` for an anti-pattern to avoid.

`[GO-API-004]` **SHOULD**: The controller should have a `EndpointNameHandler` function for each defined endpoint that takes the `*gin.Context` as its only argument. It should return a `JSONResponse` struct that contains the HTTP response code, and a body. See `examples/API/controller-api.md` for an illustration and `examples/API/api-anti-pattern.md` for an anti-pattern to avoid.

`[GO-API-005]` **SHOULD**: CORS middleware should use the `github.com/gin-contrib/cors` package. See `examples/API/controller-api.md` for an illustration and `examples/API/api-anti-pattern.md` for an anti-pattern to avoid.
