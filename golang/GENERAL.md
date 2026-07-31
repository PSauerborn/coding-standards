---
title: Golang General Standards
description: General standards for writing Go applications.
scope:
- '*.go'
parent: GENERAL.md
topics:
- golang
examples:
- examples/GENERAL/config.md
- examples/GENERAL/unittests.md
- examples/GENERAL/logging.md
- examples/GENERAL/persistence-interface.md
- examples/GENERAL/persistence-anti-pattern.md
- examples/GENERAL/persistence-postgres.md
---

# Golang General Standards

## 1. Versions and Tooling

`[GO-001]` **MUST**: Golang 1.25 or higher must be used for all applications.

`[GO-002]` **MUST**: All Golang applications must use `go` modules.

`[GO-003]` **MUST**: All code must be formatted using `gofmt`.

`[GO-004]` **MUST**: All code must be linted using `golangci-lint`.

## 2. Syntax, Naming & Style

`[GO-005]` **MUST**: All functions must have a doc comment that clearly describes the purpose of the function, its parameters, and its return values. The first word of every doc comment should be the name of the function. This ensures that the doc comment is easily searchable.

`[GO-006]` **MUST**: Filenames must all be snake_case.

`[GO-007]` **SHOULD**: Prefer functional programming patterns (pure functions, immutability) over object-oriented patterns (classes, inheritance).

`[GO-008]` **SHOULD**: Projects that build a single binary/application (i.e. a single API) should be structured in a flat directory structure. This prevents over-engineering and deep nesting for simple services, making file navigation and package imports cleaner.

`[GO-009]` **SHOULD**: Projects that build multiple binaries/applications that share code (i.e an API, CLI and workers) should be structured in a nested directory structure. At minimum this should include a `cmd` directory for the main application, a `bin` directory for binary files, and an `internal` directory for shared code.

## 3. Data Models and Validation

`[GO-010]` **MUST**: Data models must be defined in a dedicated `types.go` file.

`[GO-011]` **SHOULD**: Data models should be used throughout the application to group related data. This ensures that the code base is easier to navigate, and makes the code more readable.

`[GO-012]` **SHOULD**: Data models should have validation rules in place. Validation should be done via the `github.com/go-playground/validator/v10` package. At minimum, fields that are required should be tagged with `validate:"required"`.

`[GO-013]` **SHOULD**: DTOs and domain models should be defined separately. This ensures that API logic is decoupled from database/storage logic.

`[GO-014]` **SHOULD**: DTOs should have strict validation rules to ensure that they are valid before being used in business logic. This ensures that external data is validated before being used in business logic.

## 4. Error Handling

`[GO-015]` **MUST**: Business logic handlers must return errors rather than panicking. Top level handlers such as `main.go` can panic if an error is returned.

`[GO-016]` **SHOULD**: Custom errors should be used in favor of generic errors. This ensures that errors are more descriptive and can be handled more effectively.

`[GO-017]` **SHOULD**: Custom errors should be defined in a dedicated `errors.go` file.

## 5. Configuration

`[GO-018]` **MUST**: Configuration must be handled via a combination of `yaml` configuration files and environment variables. This ensures that configuration is decoupled from code and can be easily changed without modifying code.

`[GO-019]` **MUST**: Sensitive configuration values must be stored in environment variables and not in committed `yaml` configuration files.

`[GO-020]` **MUST**: Configuration must be validated at application startup to ensure that all required variables are set. This ensures that the application fails fast if a required variable is not set.

`[GO-021]` **SHOULD**: Default variables should be defined in the `yaml` configuration files. This minimizes the amount of environment variables that need to be set.

`[GO-022]` **SHOULD**: Packages should define a `Config` struct that contains all configuration settings required by the package. Each config field should be tagged with a `validate` tag to validate the values passed via environment variables. Validation should be done via the `github.com/go-playground/validator/v10` package. See `examples/GENERAL/config.md` for an illustration.

`[GO-023]` **SHOULD**: Configuration variables should be loaded via the `github.com/spf13/viper` package. The `Config` struct should then be populated with the values loaded from environment variables. See `examples/GENERAL/config.md` for an illustration.

`[GO-024]` **SHOULD**: Configuration variables should be loaded from environment variables first, and then from `yaml` configuration files. This ensures that environment variables take precedence over `yaml` configuration files.

## 6. Unittests

`[GO-025]` **MUST**: Unittests must be implemented for most business logic. 100% coverage is not required, but unittests should be comprehensive and cover at least 80% of the code as a guideline.

`[GO-026]` **MUST**: Unittests must be implemented using the `testing` package.

`[GO-027]` **MUST**: Each `.go` file should have a corresponding `_test.go` file that contains unittests for the logic defined in the `.go` file.

`[GO-028]` **SHOULD**: Each function should have a corresponding unittest. Unittests should be named in the format `TestFunctionName`.

`[GO-029]` **SHOULD**: All possible paths through a function should be tested. Each path should be tested within the same `TestFunctionName` function, but in a separate `t.Run` block. This ensures that unittests remain maintainable and easy to understand while still grouping related tests together.

`[GO-030]` **SHOULD**: Unittests should mock database and other service connections rather than using real connections. Connection to live databases should only be used in integration tests. This ensures that unittests are fast and do not depend on external services.

`[GO-031]` **SHOULD**: Additional test data used to test business logic (PDF, CSV files etc) should be stored in a separate `tests/data` directory. This ensures that test files do not become cluttered with test data.

`[GO-032]` **SHOULD**: Unittests should make use of the `github.com/stretchr/testify/assert` package to make assertions. Favor `github.com/stretchr/testify/assert` over comparisons using `==` or `!=` to make assertions.

See `examples/GENERAL/unittests.md` for a reference unittest implementation.

## 7. Logging

`[GO-033]` **SHOULD**: Structured logging should be implemented using the `github.com/sirupsen/logrus` package with the `logrus.JSONFormatter`. See `examples/GENERAL/logging.md` for an illustration.

## 8. Persistence Layers

`[GO-034]` **MUST**: Persistence layers must have their own dedicated file that contains all storage logic.

`[GO-035]` **MUST**: Persistence layers must be implemented via an interface. Each interface must have an associated `New` constructor that returns a new instance of the interface, which should accept connection parameters as arguments. Database clients must be initialized in the `New` constructor and set as a field of the interface implementation. See `examples/GENERAL/persistence-interface.md` for an illustration and `examples/GENERAL/persistence-anti-pattern.md` for an anti-pattern to avoid.

`[GO-036]` **MUST**: Any database operations that require multiple queries/steps must be executed within a transaction statement. This ensures that database operations are atomic and consistent, and minimizes the risk of incomplete or inconsistent data.

`[GO-037]` **MUST**: Database connections must be closed when the application is terminated.

`[GO-038]` **MUST**: Persistence layers must be implemented in a thread-safe manner.

`[GO-039]` **SHOULD**: Persistence layers should be implemented using the repository pattern.

`[GO-040]` **SHOULD**: Favor returning domain models from persistence layers unless only a single value is being returned. This ensures that related data items are grouped, and that code stays readable.

`[GO-041]` **SHOULD**: IDs and timestamps should be generated within the persistence layer function(s) rather than in the application layer. This minimizes the number of arguments required and ensures that IDs and timestamps are consistently generated across the application.

`[GO-042]` **SHOULD**: Persistence layer functions that create new records should return the ID of the generated resources and any associated errors.

`[GO-043]` **SHOULD**: PostgreSQL should be used by default if not otherwise specified.

### PostgreSQL

`[GO-044]` **MUST**: PostgreSQL persistence layers must be implemented using the `github.com/jackc/pgx/v5` package. This enforces consistency across all applications. See `examples/GENERAL/persistence-postgres.md` for an illustration.

`[GO-045]` **SHOULD**: Connection pools should be used in most cases rather than isolated connections. This ensures that connections are reused and that the application does not consume too many resources. See `examples/GENERAL/persistence-postgres.md` for an illustration.
