---
title: Golang General Standards
description: General standards for writing Go applications.
scope: '*.go'
topics:
- golang
---

## 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.

1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

## 2. Versions and Tooling

**MUST**: Golang 1.25 or higher must be used for all applications.

**MUST**: All Golang applications must use `go` modules.

**MUST**: All code must be formatted using `gofmt`.

**MUST**: All code must be linted using `golangci-lint`.

## 3. Syntax, Naming & Style

**MUST**: Use spaces for indentation instead of tabs.

**MUST**: All functions must have a doc string that clearly describes the purpose of the function, its parameters, and its return values. The first word of every doc string should be the name of the function. This ensures that the doc string is easily searchable.

**MUST**: Filename must all be snake_case.

**SHOULD**: Prefer functional programming patterns (pure functions, immutability) over object-oriented patterns (classes, inheritance).

**SHOULD**: Projects that build a single binary/application (i.e. a single API) should be structured in a flat directory structure. This prevents over-engineering and deep nesting for simple services, making file navigation and package imports cleaner.

**SHOULD**: Projects that build multiple binaries/applications that share code (i.e an API, CLI and workers) should be structured in a nested directory structure. At minimum this should include a `cmd` directory for the main application, a `bin` directory for binary files, and an `internal` directory for shared code.

## 4. Data Models and Validation

**MUST**: Data models must be defined in a dedicated `types.go` file.

**SHOULD**: Data models should be used throughout the application to group related data. This ensures that the code base is easier to navigate, and makes the code more readable.

**SHOULD**: Data models should have validation rules in place. Validation should be done via the `github.com/go-playground/validator/v10` package. At minimum, fields that are required should be tagged with `validate:"required"`.

**SHOULD**: DTOs and domain models should be defined separately. This ensures that API logic is decoupled from database/storage logic.

**SHOULD**: DTOs should have strict validation rules to ensure that they are valid before being used in business logic. This ensures that external data is validated before being used in business logic.

## 5. Error Handling

**MUST**: Business logic handlers must return errors rather than panicking. Top level handlers such as `main.go` can panic if an error is returned.

**SHOULD**: Custom errors should be used in favor of generic errors. This ensures that errors are more descriptive and can be handled more effectively.

**SHOULD**: Custom errors should be defined in a dedicated `errors.go` file.

## 6. Configuration

**MUST**: Configuration must be handled via a combination of `yaml` configuration files and environment variables. This ensures that configuration is decoupled from code and can be easily changed without modifying code.

**MUST**: Sensitive configuration values must be stored in environment variables and not in committed `yaml` configuration files.

**MUST**: Configuration must be validated at application startup to ensure that all required variables are set. This ensures that the application fails fast if a required variable is not set.

**SHOULD**: Default variables should be defined in the `yaml` configuration files. This minimizes the amount of environment variables that need to be set.

**SHOULD**: Packages should define a `Config` struct that contains all configuration settings required by the package. Each config field should be tagged with a `validate` tag to validate the values passed via environment variables. Validation should be done via the `github.com/go-playground/validator/v10` package. See `Example 1` for an illustration.

**SHOULD**: Configuration variables should be loaded via the `github.com/spf13/viper` package. The `Config` struct should then be populated with the values loaded from environment variables. See `Example 1` for an illustration.

**SHOULD**: Configuration variables should be loaded from environment variables first, and then from `yaml` configuration files. This ensures that environment variables take precedence over `yaml` configuration files.

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
    LogLevel string    `validate:"required,oneof=debug info warn error"`
    DbHost      string `validate:"required"`
	DbUser      string `validate:"required"`
	DbPassword  string `validate:"required"`
	DbDatabase  string `validate:"required"`
	DbPort      int    `validate:"required"`
}

// Validate validates the configuration using the validator package
func (c Config) Validate() error {
    // GOOD: use validator package to validate config struct
    validate := validator.New(validator.WithRequiredStructEnabled())
    return validate.Struct(c)
}

// LoadConfig loads the configuration from environment variables. If a required variable is not set, the application will panic.
func LoadConfig() *Config {
    replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

    // GOOD: use viper to load environment variables.
    // GOOD: load environment variables BEFORE loading yaml config files
	viper.AutomaticEnv()

    // GOOD: load yaml config files.
    // GOOD: load yaml config files AFTER loading environment variables
    viper.AddConfigPath("etc")
	viper.SetConfigType("yaml")

	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			panic(err)
		}
	}

    // GOOD: populate config struct with environment variables
    cfg := &Config{
        LogLevel: viper.GetString("log_level"),
        DbHost: viper.GetString("db.host"),
        DbUser: viper.GetString("db.user"),
        DbPassword: viper.GetString("db.password"),
        DbDatabase: viper.GetString("db.database"),
        DbPort: viper.GetInt("db.port"),
    }

    // GOOD: validate configuration at load time
    if err := cfg.Validate(); err != nil {
        panic(err)
    }
    return cfg
}

```

## 7. Unittests

**MUST**: Unittests must be implemented for most business logic. 100% coverage is not required, but unittests should be comprehensive and cover at least 80% of the code as a guideline.

**MUST**: Unittests must be implemented using the `testing` package.

**MUST**: Each `.go` file should have a corresponding `_test.go` file that contains unittests for the logic defined in the `.go` file.

**SHOULD**: Each function should have a corresponding unittest. Unittests should be named in the format `TestFunctionName`.

**SHOULD**: All possible paths through a function should be tested. Each path should be tested within the same `TestFunctionName` function, but in a separate `t.Run` block. This ensures that unittests remain maintainable and easy to understand while still grouping related tests together.

**SHOULD**: Unittests should mock database and other service connections rather than using real connections. Connection to live databases should only be used in integration tests. This ensures that unittests are fast and do not depend on external services.

**SHOULD**: Additional test data used to test business logic (PDF, CSV files etc) should be stored in a separate `tests/data` directory. This ensures that test files do not become cluttered with test data.

**SHOULD**: Unittests should make use of the `github.com/stretchr/testify/assert` package to make assertions. Favor `github.com/stretchr/testify/assert` over comparisons using `==` or `!=` to make assertions.

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
    // GOOD: unittests should be grouped by test case and ran in their own t.Run block
    t.Run("test case 1", func(t *testing.T) {
        result, err := SomeFunction()
        expected := "expected result"

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

## 8. Logging

**MUST**: All applications must implement logging. Logging should be present at all levels of the application.

**MUST**: Logging must follow a structured logging format. All log messages should contain at minimum the timestamp, log level, and message.

**SHOULD**: The log level should be configurable.

**SHOULD**: Kept logged messages minimal. Favor providing additional data and context via structured logging fields.

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
    // GOOD: use log.WithFields to provide additional context
    log.WithFields(log.Fields{
        "version": "1.0.0",
    }).Info("Application started")

    if err := DoSomething(); err != nil {
        log.WithError(err).Fatal("Failed to do something")
    }
}
```

## 9. Persistence Layers

**MUST**: Persistence layers must have their own dedicated file that contains all storage logic.

**MUST**: Persistence layers must be implemented via an interface. Each interface must have an associated `New` constructor that returns a new instance of the interface, which should accept connection parameters as arguments. Database clients must be initialized in the `New` constructor and set as a field of the interface implementation. See `Example 4` for an illustration.

**MUST**: Any database operations that require multiple queries/steps must be executed within a transaction statement. This ensures that database operations are atomic and consistent, and minimizes the risk of incomplete or inconsistent data.

**MUST**: Database connections must be closed when the application is terminated.

**MUST**: Persistence layers must be implemented in a thread-safe manner.

**SHOULD**: Persistence layers should be implemented using the repository pattern.

**SHOULD**: Favor returning domain models from persistence layers unless only a single value is being returned. This ensures that related data items are grouped, and that code stays readable.

**SHOULD**: IDs and timestamps should be generated within the persistence layer function(s) rather than in the application layer. This minimizes the number of arguments required and ensures that IDs and timestamps are consistently generated across the application.

**SHOULD**: Persistence layer functions that create new records should return the ID of the generated resources and any associated errors.

**SHOULD**: PostgreSQL should be used by default if not otherwise specified.

### Example 4

The following example illustrates a persistence layer for a generic SQLite database.

```go
// GOOD
// File: persistence.go

import (
    "database/sql"
    "time"
    "github.com/google/uuid"
)

// GOOD: define domain model separately from DTOs
type User struct {
    Id        string    `validate:"required"`
    Name      string    `validate:"required"`
    Email     string    `validate:"required,email"`
    CreatedAt time.Time `validate:"required"`
    UpdatedAt time.Time `validate:"required"`
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
// GOOD: return ID of the created resource and any associated errors.
func (r ExampleRepository) CreateUser(name, email string) (string, error) {

    // GOOD: generate IDs and timestamps within persistence layers
    ts := time.Now().UTC()
    id := uuid.New().String()

    // GOOD: use custom error types to provide more context about the error
    if userExists {
        return UserExistsError{Id: user.Id}
    }
    return id, nil
}

// GOOD: implement custom error types
type UserNotFoundError struct {
    Id string
}

func (e UserNotFoundError) Error() string {
    return fmt.Sprintf("user not found: %s", e.Id)
}

// GetUserById retrieves a user by ID.
// GOOD: return domain model and any associated errors.
func (r ExampleRepository) GetUserById(id string) (User, error) {
    // GOOD: use custom error types to provide more context about the error
    if userNotFound {
        return User{}, UserNotFoundError{Id: id}
    }
    return User{}, nil
}

```

### Example 5

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
// BAD: persistence layer is not implemented using interface
func GetUser(email string) (string, string, error) {
    // BAD: dependencies are initialized within the function
    connection, err := sql.Open("sqlite", "")
    if err != nil {
        return "", "", err
    }

    response, err := connection.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
    if err != nil {
        return "", "", err
    }

    return response.LastInsertId(), "", nil
}
```

### PostgreSQL

**MUST**: PostgreSQL persistence layers must be implemented using the `github.com/jackc/pgx/v5` package. This enforces consistency across all applications.

**SHOULD**: Connection pools should be used in most cases rather than isolated connections. This ensures that connections are reused and that the application does not consume too many resources.

### Example 6

The following example illustrates how to configure PostgreSQL persistence layers using the `github.com/jackc/pgx/v5` package, connection pools and transactions.

```go
// GOOD
// File: persistence.go
import (
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// GOOD: define interface for persistence layer
type PersistenceLayer interface {
    CreateUser(user User) error
    GetUserById(id string) (User, error)
}

type PostgresPersistenceLayer struct {
    connection *pgxpool.Pool
}

// GOOD: implement New function for persistence layer
// NewPostgresPersistenceLayer creates a new PostgresPersistenceLayer instance.
func NewPostgresPersistenceLayer(dsn string) (*PostgresPersistenceLayer, error) {
    connection, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to open database")
        return nil, err
    }
    return &PostgresPersistenceLayer{connection: connection}, nil
}

// CreateUser creates a new user in the database.
// GOOD: return ID of the created resource and any associated errors.
func (db *PostgresPersistenceLayer) CreateUser(name, email, role string) (string, error) {

    // GOOD: use transactions to ensure data consistency
    tx, err := db.connection.Begin(context.Background())
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to begin transaction")
        return "", err
    }
    defer tx.Rollback(context.Background())

    // GOOD: generate IDs and timestamps within persistence layers
    ts := time.Now().UTC()
    id := uuid.New().String()

    query := "INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)"
    _, err = tx.Exec(context.Background(), query, id, name, email, ts, ts)
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to execute query")
        return "", err
    }

    query = "INSERT INTO user_roles (user_id, role) VALUES ($1, $2)"
    _, err = tx.Exec(context.Background(), query, id, role)
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to execute query")
        return "", err
    }

    err = tx.Commit(context.Background())
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to commit transaction")
        return "", err
    }

    // GOOD: return ID of the created resource and any associated errors.
    return id, nil
}

// GetUserById retrieves a user by ID.
// GOOD: return domain model and any associated errors.
func (db *PostgresPersistenceLayer) GetUserById(id string) (User, error) {
    query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
    row := db.connection.QueryRow(context.Background(), query, id)
    var user User
    err := row.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
    if err != nil {
        // GOOD: log error with context
        log.WithError(err).Error("failed to scan row")
        return User{}, err
    }
    return user, nil
}
```
