package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config with ARCADYAN",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "192.168.12.1",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
				Retries:  0,
			},
			wantErr: nil,
		},
		{
			name: "valid config with NOK5G21",
			config: Config{
				Model:    "NOK5G21",
				IP:       "192.168.12.1",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: nil,
		},
		{
			name: "IP as hostname",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "gateway.local",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: nil,
		},
		{
			name: "missing model",
			config: Config{
				IP:       "192.168.12.1",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "invalid model",
			config: Config{
				Model:    "INVALID",
				IP:       "192.168.12.1",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "missing IP",
			config: Config{
				Model:    "ARCADYAN",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "invalid IP format",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "invalid ip with spaces",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "valid hostname with hyphen",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "invalid-ip",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: nil,
		},
		{
			name: "missing username",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "192.168.12.1",
				Password: "password",
				Timeout:  5 * time.Second,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "invalid timeout",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "192.168.12.1",
				Username: "admin",
				Password: "password",
				Timeout:  0,
			},
			wantErr: ErrInvalidConfig,
		},
		{
			name: "negative retries",
			config: Config{
				Model:    "ARCADYAN",
				IP:       "192.168.12.1",
				Username: "admin",
				Password: "password",
				Timeout:  5 * time.Second,
				Retries:  -1,
			},
			wantErr: ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFlagNameFromField_KnownField(t *testing.T) {
	result := flagNameFromField("Model")
	assert.Equal(t, ConfigModel, result)
}

func TestFlagNameFromField_UnknownField(t *testing.T) {
	result := flagNameFromField("UnknownField")
	assert.Equal(t, "unknownfield", result)
}
