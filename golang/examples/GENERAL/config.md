# [GO-022] Configuration Loading

Statements: `[GO-018]` `[GO-020]` `[GO-021]` `[GO-022]` `[GO-023]` `[GO-024]`

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
