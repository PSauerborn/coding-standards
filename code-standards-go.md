# Golang Code Standards

This document contains coding standards for Golang projects. The document outlines a series of **rules** and **guidelines**. **rules** are mandatory and must be followed. **guidelines** are best practices and should be implemented where reasonable. Examples should be treated as guidelines.

## General

**Rule**: Golang applications must use go modules.

**Rule**: All functions must have a doc string that clearly describes the purpose of the function, its parameters, and its return values. The first word of every doc string should be the name of the function.

**Rule**: Filename must all be snake_case.

**Guideline**: Object-oriented programming should be avoided in favor of functional programming. 

**Guideline**: Projects that build a single binary/application should be structured in a flat directory structure. 

**Guideline**: Projects that build multiple binaries/applications that share code (i.e an API, CLI and workers) should be structured in a nested directory structure. At minimum this should include a `cmd` directory for the main application, a `bin` directory for binary files, and an `internal` directory for shared code.


## Linting & Formatting

**Rule**: Linting must be performed using the `golangci-lint` tool.

**Rule**: Code must be formatted using the `gofmt` tool.

## Data Models and Validation

**Rule**: Data models must be defined in a dedicated file.

**Rule**: Data models must have validation rules defined. Validation should be done via the `github.com/go-playground/validator/v10` package.

**Guideline**: Structs than have multiple custom struct types should contain a `Validate` function that validates nested structs.

**Guideline**: DTOs and domain models should be defined separately. DTOs must be have strict validation rules to ensure that they are valid before being used in business logic. Validation for domain models can be less strict.

**Guideline**: Domain models that belong to entities stored in a database should all have a `createdAt`, `updatedAt`, `createdBy` and `lastUpdatedBy` field. 

## Error Handling

**Rule**: Business logic handlers must return errors rather than panicking. Top level handlers such as `main.go` can panic if an error is returned.

**Guideline**: Custom erraors should be used where appropriate. Custom errors should be defined in a dedicated file.

## Configuration

**Rule**: Configuration must be handled via environment variables.

**Rule**: Configuration must be validated at application startup to ensure that all required variables are set. Validation should be done via the `github.com/go-playground/validator/v10` package.

**Guidelines**: Packages should define a `Config` struct that contains all configuration variables. Each field should be tagged with a `validate` tag validate the values passed via environment variables. For instance, the `required` tag can be used to ensure that a variable is set.

**Guidelines**: Environment variables should be loaded via the `github.com/spf13/viper` package. The `Config` should then be populated with the values loaded from environment variables.


### Example

The following example illustrates how application configuration can be handled.

```go
package main

import (
    "github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
    LogLevel string `validate:"required,oneof=debug info warn error"`
    DbUrl string `validate:"required"`
}

func (c Config) Validate() error {
    validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(c)
}

func LoadConfig() *Config {
    viper.AutomaticEnv()
	// Set default values for optional variables
	viper.SetDefault("LOG_LEVEL", "info")

	cfg := &Config{
		LogLevel: viper.GetString("LOG_LEVEL"),
        DbUrl: viper.GetString("DB_URL"),
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

```

## Unittests

**Rule**: Unittests must be implemented for all business logic.

**Rule**: Unittests must be implemented using the `testing` package.

**Guidline**: Each function should have a corresponding unittest. Unittests should be named in the format `TestFunctionName`.

**Guideline**: All possible paths through a function should be tested. Each path should be tested within the same `TestFunctionName` function, but in a separate `t.Run` block.

**Guideline**: Each `.go` file should have a corresponding `_test.go` file that contains unittests for the logic defined in the `.go` file.

**Guideline**: Unittests should be run via the `gotestsum` package.

**Guideline**: Unittests should mock database and other service connections rather than using real connections to ensure that unittests are fast and do not depend on external services.

**Guideline**: Additional test data used to test business logic (PDF, CSV files etc) should be stored in a separate `tests/data` directory.

## Logging

**Rule**: All applications must implement logging.

**Rule**: Logging must follow a structured logging format.All log messages should contain at minimum the timestamp, log level, and message.

**Guideline**: Log levels should be passed down as environment variables.

**Guideline**: Logging should be handled via the `github.com/sirupsen/logrus` package, ideally using the `logrus.JSONFormatter`. 

## Persistence Layers

**Rule**: persistence layers must have their own dedicated file that contains all database queries.

**Rule**: persistence layers must be implemented via an interface. Each interface should have an associated `New` function that returns a new instance of the interface, which should accept connection parameters as arguments. Database clients should be initialized in the `New` function and set as a field of the interface implementation.

**Guideline**: persistence layers should be implemented using the repository pattern.

**Guideline**: persistence layers should should accept and return domain models where multiple input and return values are required.

### Example

The following example illustrates a persistence layer for a generic SQLite database.

```go
type User struct {
    Id    string
    Name  string
    Email string
}

type PersistenceLayer interface {
    CreateUser(user User) error
    GetUserById(id string) (User, error)
}

type ExampleRepository struct {
    db *sql.DB
}

func NewExampleRepository(dsn string) (*ExampleRepository, error) {
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return ExampleRepository{}, err
    }
    return &ExampleRepository{db: db}, nil
}

func (r ExampleRepository) CreateUser(user User) error {
    return nil
}

func (r ExampleRepository) GetUserById(id string) (User, error) {
    return User{}, nil
}

```

### PostgreSQL

**Rule**: PostgreSQL persistence layers must be implemented using the `github.com/jackc/pgx/v5` package.

**Guidline**: Connection pools should be used in most cases rather than isolated connections.

## REST APIs

**Rule**: REST APIs must accept and return JSON data. Exceptions can be made for file uploads and responses where binary data is required. 

**Rule**: REST APIs must structure responses in a consistent manner. Error responses must contain an `error` field, and an optional `details` field. The `error` field must contain a generic error message i.e. "Internal Server Error", "Bad Request" etc. The `details` field should contain additional details about the error where applicable. Success responses must return data in a `data` field.

**Rule**: All REST APIs must have an associated `openapi.yaml` file that defines the API contract.

**Rule**: DTO structs must define struct validation rules using the `binding` tag.

**Guideline**: REST APIs should be implemented using the `github.com/gin-gonic/gin` package.

**Guidelines**: Packages defining REST APIs should have a `NewRouter` function that returns a new instance of the router with all plugins and endpoints registered. Registration of endpoints should be kept minimal and only contain the basic logic for routing, creation of depedencies such as database clients, and error handling. All business logic should be implemented outside of the endpoint definition.

**Guidelines**: REST API endpoints should follow a dependency injection pattern. Database and other service client should not be initialized in business logic functions.

**Guidelines**: Database clients and other dependencies should be initialized within the endpoint handler, when the endpoint is invoked. Database and other service connections should NOT be shared across endpoints.

**Guidelines**: Each endpoint should have a `EndpointNameHandler` function that takes the `*gin.Context` as the first argument, followed by any additional dependencies such as database clients. It should return a `JSONResponse` struct that contains the HTTP response code, and body.

**Guidelines**: CORS should be enabled for REST APIs by default via the `github.com/gin-contrib/cors` package.

**Guidelines**: Each endpoint should have its own unittest.

### Example

The following example illustrates a REST API endpoint for a generic SQLite database.

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"

    log "github.com/sirupsen/logrus"
)

type JSONResponse struct {
    Code int
    Body interface{}
}

func(response JSONResponse) Send(c *gin.Context) {
    c.JSON(response.Code, response.Body)
}

func NewRouter() *gin.Engine {
    router := gin.Default()
    router.Use(cors.Default())

    router.GET("/health", func (c *gin.Context) {
        db, err := sql.Open("sqlite", "./test.db")
        if err != nil {
            log.Error(fmt.Sprintf("failed to open database: %v", err))
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
            return
        }

        response := HealthCheckHandler(c)
        response.Send(c)
    })

    return router
}

func HealthCheckHandler(c *gin.Context, db *sql.DB)  JSONResponse {
    if err := db.Ping(); err != nil {
        return JSONResponse{
            Code: http.StatusInternalServerError,
            Body: gin.H{
                "error": "Internal Server Error", 
                "details": "Failed to connect to database server.",
            },
        }
    }

    return JSONResponse{
        Code: http.StatusOK,
        Body: gin.H{"data": "OK"},
    }
}

func main() {
    router := NewRouter()
    if err := router.Run(":8080"); err != nil {
        log.Fatal(fmt.Sprintf("failed to run router: %v", err))
    }
}

```

## Dockerfiles

**Rule**: Dockerfiles must be provided for all applications. 

**Rule**: Dockerfiles must be implemented as multi-stage builds. 

**Rule**: Images must be built for AMD linux architecture. Use the --platform linux/amd64 flag to specify the architecture when building the image. Additionally, the --provenance=false flag must be used to disable provenance.

**Rule**: Non-essential files should be excluded from the final image.

**Guideline**: Dockerfiles should consist of three stages. The first stage should run unittests, the second stage should build the application, and the third stage should run the application. 

**Guideline**: Any stages that do not run the application should be based on the full golang image. Stages that run the application should be based on the `gcr.io/distroless/static:nonroot` image where possible.


### Example

```dockerfile
FROM golang:1.25 AS tests

WORKDIR /app/tests

COPY go.mod go.sum ./
RUN go mod download

RUN go install gotest.tools/gotestsum@latest

COPY etc ./etc
COPY *.go ./

CMD ["gotestsum", "--format", "testname"]

FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o api .

FROM gcr.io/distroless/static:nonroot AS runtime

WORKDIR /app

COPY --from=build /app/api .

COPY etc ./etc

CMD ["./api"]
```
