---
Title: Acceptance Testing Standards
Description: Standards for acceptance testing using Gherkin and Cucumber.
Language: testing
Topics:
- testing
- acceptance-testing
- gherkin
- cucumber
---

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

## General

**MUST**: Acceptance tests must be implemented using Gherkin `Given, When, Then` syntax.

**MUST**: Acceptance tests must be implemented using the Cucumber framework.

**SHOULD**: Acceptance tests should be placed in the `acceptance` directory.

**SHOULD**: Acceptance tests should be ran against live DEV environments.

**Guidline**: `.feature` files should be placed in the `acceptance/features` directory.

**Guidline**: Step definitions should be placed in the root `acceptance` directory.

**Guidline**: Acceptenance tests should be implemented in the same language as the application. Golang projects should use the `github.com/cucumber/godog` package. Python projects should use the `behave` package.

**SHOULD**: A dockerfile should be provided to run the acceptance tests.


### Example 1

Acceptance tests implemented in Golang should use the `github.com/cucumber/godog` package. A `main_test.go` file should be provided to scaffold the acceptance tests. The following is an example of a `main_test.go` file:

```go
package main

import (
    "net/http"
    "os"
    "testing"

    "github.com/cucumber/godog"
)

const (
    API_BASE_URL = "https://api-dev.alpn-software.com"
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
