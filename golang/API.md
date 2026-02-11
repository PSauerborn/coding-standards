---
title: Golang REST API Standards
description: Standards for writing REST APIs in Go.
parent: golang/GENERAL.md
scope: '*.go'
topics:
- golang
- api
- rest
- gin-gonic
---

## 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.

1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

## 2. REST API Guidelines

**MUST**: REST APIs must accept and return JSON data. Exceptions can be made for file uploads and responses where binary data is required.

**MUST**: REST APIs must structure responses in a consistent manner.

**MUST**: REST APIs must have a version prefix in the URL. This ensures that APIs can be versioned and that old APIs can be deprecated.

**MUST**: Error responses must contain an `error` field, and an optional `details` field. The `error` field must contain a generic error message i.e. "Internal Server Error", "Bad Request" etc. The `details` field should contain additional details about the error where applicable.

**MUST**: REST APIs must have a `/health` endpoint that checks the health of the API. The health endpoint should check the health of all dependencies and return a `200 OK` response if all dependencies are healthy.

**MUST**: REST APIs must have a `/version` endpoint that returns the version of the API.

**SHOULD**: REST API endpoints should follow a dependency injection pattern. Prefer initialization of dependencies within the endpoint handler rather than within business logic. See `Example 1` for an illustration.

**SHOULD**: REST APIs should be implemented using the `github.com/gin-gonic/gin` package.

**SHOULD**: REST APIs should be implemented using a controller pattern. The controller should contain singletons for database clients, service clients, etc. See `Example 1` for an illustration.

**SHOULD**: Packages defining REST APIs should have a `NewRouter` constructor that returns a new instance of the router with all plugins and endpoints registered. See `Example 1` for an illustration.

**SHOULD**: Registration of endpoints should be kept minimal and only contain the basic logic for routing, creation of dependencies such as database clients, and error handling. All business logic should be implemented in the controller. See `Example 1` for an illustration.

**SHOULD**: The controller should have a `EndpointNameHandler` function for each defined endpoint that takes the `*gin.Context` as its only argument. It should return a `JSONResponse` struct that contains the HTTP response code, and a body. See `Example 1` for an illustration.

**SHOULD**: CORS should be enabled for REST APIs by default via the `github.com/gin-contrib/cors` package. See `Example 1` for an illustration.

**SHOULD**: Each endpoint should have its own unittest.

### Example 1

The following example illustrates how a REST API should be structured.

```go
// GOOD
// filename: main.go
package main

import (
	"fmt"
    "net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)


// GOOD: define JSONResponse struct to ensure that all responses are consistent
type JSONResponse struct {
    Code int
    Body interface{}
}

// Send sends the response to the client.
func (r JSONResponse) Send(c *gin.Context) {
    c.JSON(r.Code, r.Body)
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(code int, err, details string) JSONResponse {
    return JSONResponse{
        Code: code,
        Body: map[string]string{
            "error": err,
            "details": details,
        },
    }
}

// GOOD: define controller struct with dependencies
// GOOD: use singleton pattern for database and service clients
type Controller struct {
    db      Persistence
    config *Config
}

// NewController creates a new controller with the given dependencies.
func NewController(db Persistence, config *Config) *Controller {
    return &Controller{
        db:      db,
        config: config,
    }
}

// HealthCheck checks the health of the application. All dependencies are checked.
func (ct *Controller) HealthCheck(c *gin.Context) JSONResponse {
    // GOOD: use structured logging throughout application
    log.WithFields(log.Fields{
        "host": ct.config.DatabaseHost,
        "port": ct.config.DatabasePort,
    }).Info("checking health")

    if err := ct.db.HealthCheck(); err != nil {
        log.WithError(err).Error("failed to connect to database")
        // GOOD: return consistent error response
        return NewErrorResponse(http.StatusInternalServerError, "Internal Server Error", "Failed to connect to database")
    }

    return JSONResponse{
        Code: http.StatusOK,
        Body: map[string]interface{}{
            "status": "ok",
        },
    }
}

// VersionCheck returns the version of the application as defined in config.
func (ct *Controller) VersionCheck(c *gin.Context) JSONResponse {
    return JSONResponse{
        Code: http.StatusOK,
        Body: map[string]interface{}{
            "version": ct.config.AppVersion,
        },
    }
}

// GOOD: define request struct to ensure that all requests are consistent
// GOOD: ensure that all requests are validated
type NewResourceRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required"`
}

// GOOD: enforce consistent JSON response structure
// CreateResource creates a new resource.
func (ct *Controller) CreateResource(c *gin.Context) JSONResponse {
    var req NewResourceRequest
    // GOOD: validate request body
    if err := c.ShouldBindJSON(&req); err != nil {
        log.WithError(err).Error("failed to bind request body")
        return NewErrorResponse(http.StatusBadRequest, "Bad Request", "Invalid request body")
    }

    // GOOD: use structured logging throughout application
    log.WithFields(log.Fields{
        "name": req.Name,
        "email": req.Email,
    }).Info("creating resource")

    id, err := ct.db.CreateResource(req.Name, req.Email)
    if err != nil {
        log.WithError(err).Error("failed to create resource")
        return NewErrorResponse(http.StatusInternalServerError, "Internal Server Error", "Failed to create resource")
    }

    log.WithFields(log.Fields{
        "id": id,
    }).Info("created resource")

    return JSONResponse{
        Code: http.StatusCreated,
        Body: map[string]interface{}{
            "id": id,
        },
    }
}

// GOOD: use github.com/gin-gonic/gin for REST APIs
// GOOD: define constructor for router
// NewRouter creates a new router for the application and adds
// all endpoints.
func NewRouter(controller *Controller) *gin.Engine {
    router := gin.Default()
    // GOOD: enable CORS by default
    router.Use(cors.DefaultConfig().Handler)

    // GOOD: keep endpoint definitions minimal
    // GOOD: all REST APIs should have a /health endpoint
    router.POST("/v1/health", func(c *gin.Context) {
        response := controller.HealthCheck(c)
        response.Send(c)
    })

    // GOOD: all REST APIs should have a /version endpoint
    router.POST("/v1/version", func(c *gin.Context) {
        response := controller.VersionCheck(c)
        response.Send(c)
    })

    // GOOD: endpoints should be versioned
    router.POST("/v1/resources", func(c *gin.Context) {
        response := controller.CreateResource(c)
        response.Send(c)
    })
    return router
}

func main() {
    // GOOD: load and validate configuration at
    // application start time
    config := LoadConfig()

    // GOOD: use structured logging throughout application
    log.SetFormatter(&log.JSONFormatter{})
    log.SetLevel(config.LogLevel)

    log.WithFields(log.Fields{
        "host": config.DatabaseHost,
        "port": config.DatabasePort,
    }).Info("connecting to database")

    // GOOD: create database connection using singleton pattern
    db := NewDatabase(
        config.DatabaseHost,
        config.DatabasePort,
        config.DatabaseName,
        config.DatabaseUser,
        config.DatabasePassword,
    )
    defer db.Close()

    // GOOD: create controller with dependencies
    controller := NewController(db, config)
    router := NewRouter(controller)

    if err := router.Run(fmt.Sprintf(":%s", config.Port)); err != nil {
        log.WithError(err).Error("failed to start server")
        os.Exit(1)
    }
}

```

### Example 2

The following example illustrates how NOT to implement a REST API, and should be avoided:

```go
// BAD
// filename: main.go
package main

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)


func main() {
    // BAD: CORS should be enabled by default
    // BAD: APIs should be versioned
    // BAD: APIs should have a /health endpoint
    // BAD: APIs should have a /version endpoint
    router := gin.Default()


    // BAD: endpoint definitions should be minimal. avoid embedding business logic in endpoint definitions
    router.GET("/resource", func(c *gin.Context) {
        // BAD: dependencies should be injected via controller
        // BAD: database connection should be created using singleton pattern
        db, err := sql.Open("postgres", "user:password@localhost:5432/dbname")
        if err != nil {
            // BAD: no structured logging
            fmt.Println("failed to connect to database: %+v", err)
            // BAD: inconsistent response structures must be avoided
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to database"})
            return
        }
        defer db.Close()

        resource := db.GetResource()
        c.JSON(http.StatusOK, gin.H{"resource": resource})
    })

    router.PUT("/resource/:id", func(c *gin.Context) {
        id := c.Param("id")
        // BAD: structured logging should be used throughout application
        resource := UpdateResource(c, id)
        c.JSON(http.StatusOK, gin.H{"resource": resource})
    })

    if err := router.Run(":8080"); err != nil {
        // BAD: no structured logging
        fmt.Println("failed to start server")
        os.Exit(1)
    }
}

// BAD: no docstring present on function
func UpdateResource(c *gin.Context, id string) error {
    // BAD: dependency injection is not implemented
    db, err := sql.Open("postgres", "user:password@localhost:5432/dbname")
    if err != nil {
        // BAD: inconsistent response structures must be avoided
        c.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to connect to database"})
        return
    }
    defer db.Close()

    // BAD: DTOs should be defined separately from domain models
    var body Resource
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"msg": "Bad Request"})
        return
    }

    err := db.UpdateResource(id, body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"msg": "Internal Server Error"})
        return
    }
    // BAD: response does not conform to contract
    c.JSON(http.StatusCreated, id)
}

```

## 3. Documentation

**MUST**: All REST APIs must have an associated `openapi.yaml` file that defines the API contract. All endpoints should be documented in the `openapi.yaml` file. This ensures that the API is well-documented.

## 4. Dockerfiles

**MUST**: Dockerfiles must be provided for all applications.

**MUST**: Dockerfiles must be implemented as multi-stage builds.

**MUST**: Images must be built for AMD linux architecture. Use the `--platform linux/amd64` flag to specify the architecture when building the image. Additionally, the `--provenance=false` flag must be used to disable provenance.

**MUST**: Non-essential files should be excluded from the final image.

**SHOULD**: Dockerfiles should consist of three stages. The first stage should run unittests, the second stage should build the application, and the third stage should run the application.

**SHOULD**: Any stages that do not run the application should be based on the full golang image. Stages that run the application should be based on the `gcr.io/distroless/static:nonroot` image.

### Example 3

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
FROM gcr.io/distroless/static:nonroot AS runtime

WORKDIR /app

COPY --from=build /app/api .

COPY etc ./etc

CMD ["./api"]
```
