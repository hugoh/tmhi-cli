package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pterm/pterm"
)

// ErrInvalidConfig is returned when configuration validation fails.
var ErrInvalidConfig = errors.New("invalid configuration")

// Config holds all configuration values for the CLI application.
type Config struct {
	Model    string
	IP       string
	Username string
	Password string
	Timeout  time.Duration
	Retries  int
	Debug    bool
	DryRun   bool
}

//nolint:gochecknoglobals
var fieldToFlag = map[string]string{
	"Model":    ConfigModel,
	"IP":       ConfigIP,
	"Username": ConfigUsername,
	"Password": ConfigPassword,
	"Timeout":  ConfigTimeout,
	"Retries":  ConfigRetries,
	"Debug":    ConfigDebug,
	"DryRun":   ConfigDryRun,
}

// Validate validates the Config struct and returns formatted errors.
func (c *Config) Validate() error {
	err := validation.ValidateStruct(c,
		validation.Field(&c.Model, validation.Required, validation.In(ARCADYAN, NOK5G21)),
		validation.Field(&c.IP, validation.Required, is.Host),
		validation.Field(&c.Username, validation.Required),
		validation.Field(&c.Timeout, validation.Required, validation.Min(1*time.Second)),
		validation.Field(&c.Retries, validation.Min(0)),
	)
	if err != nil {
		if errs, ok := errors.AsType[validation.Errors](err); ok {
			for field, fieldErr := range errs {
				if fieldErr != nil {
					flagName := flagNameFromField(field)
					pterm.Error.Printf("error: --%s: %s\n", flagName, fieldErr.Error())
				}
			}
		}

		return displayed(fmt.Errorf("%w: %w", ErrInvalidConfig, err))
	}

	return nil
}

func flagNameFromField(field string) string {
	if name, ok := fieldToFlag[field]; ok {
		return name
	}

	return strings.ToLower(field)
}
