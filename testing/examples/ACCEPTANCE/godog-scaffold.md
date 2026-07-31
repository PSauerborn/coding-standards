# [ACPT-002] godog Test Scaffold

Statements: `[ACPT-001]` `[ACPT-002]` `[ACPT-003]` `[ACPT-004]`

Acceptance tests implemented in Golang should use the `github.com/cucumber/godog` package. A `main_test.go` file should be provided to scaffold the acceptance tests. The following is an example of a `main_test.go` file:

```go
// File: main_test.go
package main

// GOOD: implement tests using Cucumber framework
// GOOD: use github.com/cucumber/godog for golang projects
import (
    "net/http"
    "os"
    "testing"

    "github.com/cucumber/godog"
)

// GOOD: run against live DEV systems
const (
    API_BASE_URL = "https://api-dev.s31-software.com"
)

// TestMain is the entry point for running the acceptance tests using Godog.
func TestMain(m *testing.M) {
    opts := godog.Options{
        Format:    "pretty",             // Output format
        Paths:     []string{"features"}, // Path to feature files
        Randomize: 0,                    // Execute scenarios in order
    }

    status := godog.TestSuite{
        Name:                 "acceptance",
        TestSuiteInitializer: InitializeTestSuite,
        ScenarioInitializer:  InitializeScenario,
        Options:              &opts,
    }.Run()

    if st := m.Run(); st > status {
        status = st
    }
    os.Exit(status)
}

// InitializeTestSuite sets up the test suite, including global before/after hooks.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
    // Global setup (e.g., start API server, docker containers)
    ctx.BeforeSuite(func() {
        // StartServer()
    })
    ctx.AfterSuite(func() {
        // StopServer()
    })
}

// GOOD: define feature struct to store state between steps
type ApiFeature struct {
    resp   *http.Response
    client *http.Client
}

func (f *ApiFeature) IAmAUser() error {
    return nil
}

func (f *ApiFeature) IDoSomething() error {
    return nil
}

func (f *ApiFeature) IExpectSomething() error {
    return nil
}

// InitializeScenario sets up each scenario by registering steps and creating a new ApiFeature instance.
func InitializeScenario(ctx *godog.ScenarioContext) {
    api := ApiFeature{
        client: &http.Client{},
    }

    ctx.Step(`^I am a user$`, api.IAmAUser)
    ctx.Step(`^I do something$`, api.IDoSomething)
    ctx.Step(`^I expect something$`, api.IExpectSomething)
}

```
