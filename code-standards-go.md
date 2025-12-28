# Golang Code Standards

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. Verisons and Tooling

**MUST**: Golang applications must use go modules.

**MUST**: Golang version 1.25 or higher must be used.

**MUST**: All code must be formatted using `gofmt`.

**MUST**: All code must be linted using `golangci-lint`.

# 3. Syntax, Naming & Style

**MUST**: Use spaces for indentation instead of tabs.

**MUST**: All functions must have a doc string that clearly describes the purpose of the function, its parameters, and its return values. The first word of every doc string should be the name of the function. This ensures that the doc string is easily searchable.

**MUST**: Filename must all be snake_case.

**SHOULD**: Prefer functional programming patterns (pure functions, immutability) over object-oriented patterns (classes, inheritance).

**SHOULD**: Projects that build a single binary/application (i.e. a single API) should be structured in a flat directory structure. This prevents over-engineering and deep nesting for simple services, making file navigation and package imports cleaner.

**SHOULD**: Projects that build multiple binaries/applications that share code (i.e an API, CLI and workers) should be structured in a nested directory structure. At minimum this should include a `cmd` directory for the main application, a `bin` directory for binary files, and an `internal` directory for shared code.

# 4. Data Models and Validation

**MUST**: Data models must be defined in a dedicated file.

**SHOULD**: Data models should be used throught the application to group related data. This ensures that the code base is more type-safe, and makes the code more readable.

**SHOULD**: Data models should have validation rules in place. Validation should be done via the `github.com/go-playground/validator/v10` package. At minimum, fields that are required should be tagged with `validate:"required"`.

**SHOULD**: DTOs and domain models should be defined separately. This ensures that API logic is decoupled from database/storage logic.

**SHOULD**: DTOs should have strict validation rules to ensure that they are valid before being used in business logic. This ensures that external data is validated before being used in business logic.

# 5. Error Handling

**MUST**: Business logic handlers must return errors rather than panicking. Top level handlers such as `main.go` can panic if an error is returned.

**SHOULD**: Custom errors should be used in favor of generic errors. This ensures that errors are more descriptive and can be handled more effectively.

**SHOULD**: Custom errors should be defined in a dedicated `errors.go` file.


# 6. Configuration

**MUST**: Configuration must be handled via environment variables. This ensures that configuration is decoupled from code and can be easily changed without modifying code.

**MUST**: Configuration must be validated at application startup to ensure that all required variables are set. This ensures that the application fails fast if a required variable is not set.

**SHOULD**: Packages should define a `Config` struct that contains all configuration settings required by the package. Each config field should be tagged with a `validate` tag to validate the values passed via environment variables. Validation should be done via the `github.com/go-playground/validator/v10` package. See `Example 1` for an illustration.

**SHOULD**: Environment variables should be loaded via the `github.com/spf13/viper` package. The `Config` struct should then be populated with the values loaded from environment variables. See `Example 1` for an illustration.

### Example 1

The following example illustrates how application configuration should be handled.

```go
// GOOD
// File: config.go
package main

import (
    "github.com/go-playground/validator/v10"
    "github.com/spf13/viper"

    log "github.com/sirupsen/logrus"
)

// GOOD: define config struct with validation tags
type Config struct {
    LogLevel string `validate:"required,oneof=debug info warn error"`
    DbUrl    string `validate:"required"`
}

// Validate validates the configuration using the validator package
func (c Config) Validate() error {
    // GOOD: use validator package to validate config struct
    validate := validator.New(validator.WithRequiredStructEnabled())
    return validate.Struct(c)
}

// LoadConfig loads the configuration from environment variables. If a required variable is not set, the application will panic.
func LoadConfig() *Config {
    // GOOD: use viper to load environment variables
    viper.AutomaticEnv()
    // GOOD: set default values for environment variables
    viper.SetDefault("LOG_LEVEL", "info")

    // GOOD: populate config struct with environment variables
    cfg := &Config{
        LogLevel: viper.GetString("LOG_LEVEL"),
        DbUrl: viper.GetString("DB_URL"),
    }

    // GOOD: validate configuration at load time
    if err := cfg.Validate(); err != nil {
        panic(err)
    }
    return cfg
}

```

# 7. Unittests

**MUST**: Unittests must be implemented for most business logic. 100% coverage is unrealistic, but unittests should be comprehensive and cover at least 80% of the code as a guideline.

**MUST**: Unittests must be implemented using the `testing` package.

**MUST**: Each `.go` file should have a corresponding `_test.go` file that contains unittests for the logic defined in the `.go` file.

**SHOULD**: Each function should have a corresponding unittest. Unittests should be named in the format `TestFunctionName`.

**SHOULD**: All possible paths through a function should be tested. Each path should be tested within the same `TestFunctionName` function, but in a separate `t.Run` block. This ensures that unittests remain maintainable and easy to understand while still grouping related tests together.

**SHOULD**: Unittests should mock database and other service connections rather than using real connections. Connection to live databases should only be used in integration tests. This ensures that unittests are fast and do not depend on external services.

**SHOULD**: Additional test data used to test business logic (PDF, CSV files etc) should be stored in a separate `tests/data` directory. This ensures that test files do not become cluttered with test data.

**SHOULD**: Unittests should make use of the `github.com/stretchr/testify/assert` package to make assertions.

### Example 2

The following illustrates unittests for a `SomeFunction` defined in `main.go`.

```go
// GOOD
// File: main_test.go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSomeFunction(t *testing.T) {
    // GOOD: unittests should be grouped by test case
    t.Run("test case 1", func(t *testing.T) {
        result, err := SomeFunction()

        // GOOD: unittests should make use of the assert package
        assert.NoError(t, err)
        assert.Equal(t, expected, result)
    })

    t.Run("test case 2", func(t *testing.T) {
        result, err := SomeFunction()
        assert.Error(t, err)
    })
}
```

The following testing structure should be avoided:

```go
// BAD
// File: main_test.go
import (
    "testing"
)

// BAD: unittests should be grouped by test case and not be spread across multiple functions
func TestSomeFunctionPath1(t *testing.T) {
    result, err := SomeFunction()
    if err != nil {
        t.Fatal(err)
    }
     
    expected := "expected result"
    // BAD: unittests should make use of the assert package
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}

// BAD: unittests should be grouped by test case and not be spread across multiple functions
func TestSomeFunctionPath2(t *testing.T) {
    result, err := SomeFunction()
    if err != nil {
        t.Fatal(err)
    }
    
    expected := "expected result"
    // BAD: unittests should make use of the assert package
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}

```

# 8. Logging

**MUST**: All applications must implement logging. Logging should be present at all levels of the application.

**MUST**: Logging must follow a structured logging format. All log messages should contain at minimum the timestamp, log level, and message.

**SHOULD**: The log level should be passed down as environment variables.

**SHOULD**: Logging should be handled via the `github.com/sirupsen/logrus` package, ideally using the `logrus.JSONFormatter`. 

### Example 3

The following example illustrates a logging implementation.

```go
// GOOD
// File: main.go
package main

import (
    "os"

    // GOOD: use github.com/sirupsen/logrus for logging
    log "github.com/sirupsen/logrus"
)

func main() {
    // GOOD: get log level from environment variables
    // and convert to logrus level
    logLevel := os.Getenv("LOG_LEVEL")
    if logLevel == "" {
        logLevel = "info"
    }
    log.SetLevel(log.Level(logLevel))

    // GOOD: use structlogging
    log.SetFormatter(&log.JSONFormatter{})

    // GOOD: implement logging at all levels of the application
    log.WithFields(log.Fields{
        "version": "1.0.0",
    }).Info("Application started")

    if err := DoSomething(); err != nil {
        log.WithError(err).Fatal("Failed to do something")
    }
}
```

# 9. Persistence Layers

**MUST**: Persistence layers must have their own dedicated file that contains all storage logic.

**MUST**: Persistence layers must be implemented via an interface. Each interface must have an associated `New` function that returns a new instance of the interface, which should accept connection parameters as arguments. Database clients must be initialized in the `New` function and set as a field of the interface implementation. See `Example 4` for an illustration.

**SHOULD**: Prefer transactional operations that execute multiple database operations as a single unit of work. This ensures that database operations are atomic and consistent, and minimizes the risk of incomplete or inconsistent data.

**SHOULD**: Persistence layers should be implemented using the repository pattern.

**SHOULD**: Persistence layers should return domain models where multiple input and return values are required. See `Example 4` for an illustration.

**SHOULD**: PostgreSQL should be used by default if not otherwise specified.

### Example 4

The following example illustrates a persistence layer for a generic SQLite database.

```go
// GOOD
// File: persistence.go

// GOOD: define domain model separately from DTOs
type User struct {
    Id    string `validate:"required"`
    Name  string `validate:"required"`
    Email string `validate:"required,email"`
}

// GOOD: define interface for persistence layer
type PersistenceLayer interface {
    CreateUser(user User) error
    GetUserById(id string) (User, error)
}

type ExampleRepository struct {
    db *sql.DB
}

// GOOD: implement New function for persistence layer
// NewExampleRepository creates a new ExampleRepository instance.
func NewExampleRepository(dsn string) (*ExampleRepository, error) {
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to open database")
        return nil, err
    }
    return &ExampleRepository{db: db}, nil
}

type UserExistsError struct {
    Id string
}

// GOOD: implement custom error types
func (e UserExistsError) Error() string {
    return fmt.Sprintf("user exists: %s", e.Id)
}

// CreateUser creates a new user in the database.
func (r ExampleRepository) CreateUser(user User) error {
    // GOOD: use custom error types to provide more context about the error
    if userExists {
        return UserExistsError{Id: user.Id}
    }
    return nil
}

// GOOD: implement custom error types
type UserNotFoundError struct {
    Id string
}

func (e UserNotFoundError) Error() string {
    return fmt.Sprintf("user not found: %s", e.Id)
}

// GetUserById retrieves a user by ID.
func (r ExampleRepository) GetUserById(id string) (User, error) {
    // GOOD: use custom error types to provide more context about the error
    if userNotFound {
        return User{}, UserNotFoundError{Id: id}
    }
    return User{}, nil
}

```

The following example illustrates how NOT to implement a persistence layer:

```go
// BAD
// File: persistence.go

// BAD: DTOs have no validation
type User struct {
    Id    string
    Name  string
    Email string 
}

// BAD: DTOs are not used as return values
// BAD: persistence layer is not implenented using interface
func CreateUser(name, email string) (string, string, error) {
    // BAD: dependencies are initialized within the function
    connection, err := sql.Open("sqlite", "")
    if err != nil {
        return "", "", err
    }

    response, err := connection.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
    if err != nil {
        return "", "", err
    }

    return response.LastInsertId(), nil
}
```

### PostgreSQL

**MUST**: PostgreSQL persistence layers must be implemented using the `github.com/jackc/pgx/v5` package. This enforces consistency across all applications.

**SHOULD**: Connection pools should be used in most cases rather than isolated connections. This ensures that connections are reused and that the application does not consume too many resources.

# 10. REST APIs

**MUST**: REST APIs must accept and return JSON data. Exceptions can be made for file uploads and responses where binary data is required. 

**MUST**: REST APIs must structure responses in a consistent manner. 

**MUST**: REST APIs must have a version prefix in the URL. This ensures that APIs can be versioned and that old APIs can be deprecated.

**MUST**: Error responses must contain an `error` field, and an optional `details` field. The `error` field must contain a generic error message i.e. "Internal Server Error", "Bad Request" etc. The `details` field should contain additional details about the error where applicable. 

**MUST**: Success responses must return data in a `data` field.

**MUST**: All REST APIs must have an associated `openapi.yaml` file that defines the API contract. This ensures that the API is well-documented.

**SHOULD**: REST API endpoints should follow a dependency injection pattern. Prefer initialization of dependencies within the endpoint handler rather than within business logic. See `Example 5` for an illustration.

**SHOULD**: REST APIs should be implemented using the `github.com/gin-gonic/gin` package.

**SHOULD**: Packages defining REST APIs should have a `NewRouter` constructor that returns a new instance of the router with all plugins and endpoints registered. 

**SHOULD**: Registration of endpoints should be kept minimal and only contain the basic logic for routing, creation of depedencies such as database clients, and error handling. All business logic should be implemented outside of the endpoint definition.

**SHOULD**: Database clients and other dependencies should be initialized within the endpoint handler, when the endpoint is invoked. Prefer a new database/client/service connection for each request as this avoids long-lived connections and reduces the risk of connection leaks. See `Example 5` for an illustration.

**SHOULD**: DTO structs should be define separately from domain models, and should include validation for all fields. This should be done using the `binding` tag.

**SHOULD**: Each endpoint should have a `EndpointNameHandler` function that takes the `*gin.Context` as the first argument, followed by any additional dependencies such as database clients. It should return a `JSONResponse` struct that contains the HTTP response code, and body. See `Example 5` for an illustration.

**SHOULD**: CORS should be enabled for REST APIs by default via the `github.com/gin-contrib/cors` package.

**SHOULD**: Each endpoint should have its own unittest.

### Example 5

The following example illustrates how a REST API should be structured.

```go
// GOOD
// main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"

    log "github.com/sirupsen/logrus"
)

// GOOD: use structs to define a consistent response format
type JSONResponse struct {
    Code int
    Body interface{}
}

func(response JSONResponse) Send(c *gin.Context) {
    c.JSON(response.Code, response.Body)
}

// NewRouter returns a new router with all plugins and endpoints registered
func NewRouter() *gin.Engine {
    router := gin.Default()
    router.Use(cors.Default())

    // GOOD: endpoint registration is kept minimal. Business logic should be in handler.
    router.GET("/v1/health", func (c *gin.Context) {
        // GOOD: Database connections are initialized within the endpoint handler
        // for each request
        db, err := sql.Open("sqlite", "./test.db")
        if err != nil {
            // ensure logging is used throughout application
            log.WithError(err).Error("failed to connect to database")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
            return
        }

        // GOOD: dependency injection is implemented
        response := HealthCheckHandler(c, db)
        response.Send(c)
    })

    // GOOD: endpoint registration is kept minimal. Business logic should be in handler.
    router.POST("/v1/resource", func (c *gin.Context) {
        // GOOD: Database connections are initialized within the endpoint handler
        // for each request
        db, err := sql.Open("sqlite", "./test.db")
        if err != nil {
            // ensure logging is used throughout application
            log.WithError(err).Error("failed to connect to database")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
            return
        }

        // GOOD: dependency injection is implemented
        response := CreateResourceHandler(c, db)
        response.Send(c)
    })

    return router
}

// HealthCheckHandler handles health checks by pinging the database.
// If connection to the database fails, an error response is returned.
func HealthCheckHandler(c *gin.Context, db *sql.DB)  JSONResponse {
    if err := db.Ping(); err != nil {
        // GOOD: logging is used throughout application
        log.WithError(err).Error("failed to ping database")

        // GOOD: response is structured as per contract
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

type NewResourceRequest struct {
    Name        string `json:"name" binding:"required"` // GOOD: use binding tags to validate request data
    Description string `json:"description" binding:"required"`
}

// CreateResourceHandler handles the creation of a new resource.
func CreateResourceHandler(c *gin.Context, db *sql.DB) JSONResponse {
    var body NewResourceRequest
    if err := c.ShouldBindJSON(&body); err != nil {
        // GOOD: structured error response
        log.WithError(err).Error("failed to parse request data")
        return JSONResponse{
            Code: http.StatusBadRequest,
            Body: gin.H{
                "error": "Bad Request",
                "details": "Invalid request data.",
            },
        }
    }

    // business logic to create resource
    id, err := Todo(body)
    if err != nil {
        log.WithError(err).Error("failed to create resource")
        return JSONResponse{
            Code: http.StatusInternalServerError,
            Body: gin.H{
                "error": "Internal Server Error",
                "details": "Failed to create resource.",
            },
        }
    }

    // GOOD: structured logging is used throughout application
    log.WithFields(log.Fields{
        "name": body.Name,
        "description": body.Description,
    }).Info("resource created")

    return JSONResponse{
        Code: http.StatusCreated,
        Body: gin.H{
            "data": id,
        },
    }
}

func main() {
    config := LoadConfig()
    log.SetLevel(log.Level(config.LogLevel))
    // GOOD: Structured logging is used throughout application
    log.SetFormatter(&log.JSONFormatter{})

    router := NewRouter()

    log.WithFields(log.Fields{
        "port": ":8080",
    }).Info("starting application")

    if err := router.Run(":8080"); err != nil {
        // GOOD: panic is used for critical errors only in top level functions
        log.WithError(err).Fatal("failed to run server")
    }
}

```

The following example illustrates how REST APIs should NOT be implemented:

```go
// BAD
// File: main.go

// BAD: no docstring present on function
func CreateResourceHandler(c *gin.Context) {
    // BAD: dependency injection is not implemented
    db, err := sql.Open("sqlite", "./test.db")
    if err != nil {
        // BAD: JSONResponse contract is not implemented
        c.JSON(http.StatusInternalServerError, gin.H{"msg": "Internal Server Error"})
        return
    }

    // BAD: DTOs should be defined separately from domain models
    var body Resource
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"msg": "Bad Request"})
        return
    }

    id, err := Todo(body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"msg": "Internal Server Error"})
        return
    }
    // BAD: response does not conform to contract
    c.JSON(http.StatusCreated, id)
}

func main() {
    // BAD: routers should have a dedicated constructor
    // BAD: CORS should be configured via middleware
    router := gin.Default()

    // BAD: database connections should be initialized within the endpoint handler
    // to avoid long-lived, globally shared connections. They should NOT be shared
    // across multiple handlers.
    db, err := sql.Open("sqlite", "./test.db")
    if err != nil {
        panic(err)
    }

    // BAD: endpoint does not implement dependency injection
    // BAD: endpoint does not implement JSONResponse contract
    router.GET("/v1/health", func (c *gin.Context) {
        if err := db.Ping(); err != nil {
            // BAD: logging is not used throughout application
            // BAD: response does not conform to contract
            c.JSON(http.StatusInternalServerError, gin.H{"msg": "Internal Server Error"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"data": "OK"})
    })

    router.POST("/v1/resources", func (c *gin.Context) {
        CreateResourceHandler(c)
    })

    // BAD: logging is not implemented
    if err := router.Run(":8080"); err != nil {
        panic(err)
    }
}

```

# 11. Dockerfiles

**MUST**: Dockerfiles must be provided for all applications. 

**MUST**: Dockerfiles must be implemented as multi-stage builds. 

**MUST**: Images must be built for AMD linux architecture. Use the `--platform linux/amd64` flag to specify the architecture when building the image. Additionally, the `--provenance=false` flag must be used to disable provenance.

**MUST**: Non-essential files should be excluded from the final image.

**SHOULD**: Dockerfiles should consist of three stages. The first stage should run unittests, the second stage should build the application, and the third stage should run the application. 

**SHOULD**: Any stages that do not run the application should be based on the full golang image. Stages that run the application should be based on the `gcr.io/distroless/static:nonroot` image.

### Example

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
