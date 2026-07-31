# [GO-035] Persistence Layer Interface

Statements: `[GO-013]` `[GO-016]` `[GO-035]` `[GO-039]` `[GO-040]` `[GO-041]` `[GO-042]`

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
        return "", UserExistsError{Id: user.Id}
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
