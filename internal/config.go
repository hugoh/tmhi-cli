package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pterm/pterm"
)

// ErrInvalidConfig is returned when configuration validation fails.
var ErrInvalidConfig = errors.New("invalid configuration")

// Config holds all configuration values for the CLI application.
type Config struct {
	Model    string `validate:"required,oneof=ARCADYAN NOK5G21"`
	IP       string `validate:"required,hostname|ip"`
	Username string `validate:"required"`
	Password string
	Timeout  time.Duration `validate:"required,min=1s"`
	Retries  int           `validate:"min=0"`
	Debug    bool
	DryRun   bool
}

//nolint:gochecknoglobals
var validate = validator.New(validator.WithRequiredStructEnabled())

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
	if err := validate.Struct(c); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, fe := range ve {
				flagName := flagNameFromField(fe.Field())
				pterm.Error.Printf("error: --%s: %s\n", flagName, fe.Tag())
			}

			return fmt.Errorf("%w: %w", ErrInvalidConfig, err)
		}

		return fmt.Errorf("%w: %w", ErrInvalidConfig, err)
	}

	return nil
}

func flagNameFromField(field string) string {
	if name, ok := fieldToFlag[field]; ok {
		return name
	}

	return strings.ToLower(field)
}
