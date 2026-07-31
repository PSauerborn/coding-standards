# [GO-API-002] REST API Anti-Pattern

Statements: `[GO-API-002]` `[GO-API-003]` `[GO-API-004]` `[GO-API-005]`

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
