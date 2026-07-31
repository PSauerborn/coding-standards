# [GO-033] Structured Logging

Statements: `[GO-033]`

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

    f, err := os.OpenFile("/var/log/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	mw := io.MultiWriter(os.Stdout, f)
    // GOOD: Write both to file and stdout
	log.SetOutput(mw)
	log.SetFormatter(&logrus.JSONFormatter{})

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
