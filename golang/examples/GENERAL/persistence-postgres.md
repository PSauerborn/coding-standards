# [GO-044] PostgreSQL Persistence with pgx

Statements: `[GO-036]` `[GO-040]` `[GO-041]` `[GO-042]` `[GO-044]` `[GO-045]`

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
