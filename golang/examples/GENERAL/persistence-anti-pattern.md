# [GO-035] Persistence Layer Anti-Pattern

Statements: `[GO-013]` `[GO-014]` `[GO-035]` `[GO-040]`

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
