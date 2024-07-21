package internal

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Configuration struct {
	Login struct {
		Username string `validate:"required" yaml:"username"`
		Password string `validate:"required" yaml:"password"`
	} `yaml:"login"`
	Gateway struct {
		Model string `validate:"required,oneof=NOK5G21" yaml:"model"`
		IP    string `validate:"ipv4"                   yaml:"ip"`
	} `yaml:"gateway"`
}

func ReadConf(cfgFile string) (*Configuration, error) {
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".tmhi-cli")
		viper.AddConfigPath("$HOME/")
		viper.AddConfigPath(".")
	}

	logrus.WithField("file", viper.ConfigFileUsed()).Debug("[Config] File")
	if err := viper.ReadInConfig(); err != nil {
		var notFoundError *viper.ConfigFileNotFoundError
		if errors.As(err, &notFoundError) {
			return nil, fmt.Errorf("fatal error config file not found: %w", err)
		}
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	var conf Configuration
	if err := viper.Unmarshal(&conf); err != nil {
		logrus.Fatalf("unable to unmarshall the config %v", err)
	}
	validate := validator.New()
	if err := validate.Struct(&conf); err != nil {
		logrus.Fatalf("Missing required attributes %v\n", err)
	}

	return &conf, nil
}

func LogSetup(debugFlag bool) {
	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
