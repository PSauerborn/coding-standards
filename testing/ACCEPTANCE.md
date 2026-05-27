---
title: Acceptance Testing Standards
description: Standards for acceptance testing using Gherkin and Cucumber.
parent: GENERAL.md
scope: '*'
topics:
- testing
- acceptance-testing
- gherkin
- cucumber
---

## General

`[ACPT-001]` **MUST**: Acceptance tests must be implemented using Gherkin `Given, When, Then` syntax. The specific Gherkin-compatible BDD framework is chosen per language (see `[ACPT-002]`).

`[ACPT-002]` **SHOULD**: Acceptance tests should be implemented in the same language as the application. Golang projects should use the `github.com/cucumber/godog` package. Python projects should use the `behave` package.

`[ACPT-003]` **SHOULD**: Acceptance tests should be ran against live DEV environments to ensure that live systems are comprehensively tested.

`[ACPT-004]` **SHOULD**: Acceptance tests should be placed in the `acceptance` directory. `.feature` files should be placed in the `acceptance/features` directory. Step definitions should be placed in the root `acceptance` directory.

`[ACPT-005]` **SHOULD**: A dockerfile and Makefile should be provided to run the acceptance tests so that tests can be ran locally and as part of CI/CD processes interchangeably.


### Example 1

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
