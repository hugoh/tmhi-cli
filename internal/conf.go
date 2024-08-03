package internal

import (
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

func configFatal(msg string, err error) {
	logrus.WithFields(logrus.Fields{
		"file": viper.ConfigFileUsed(),
		"err":  err,
	}).Fatal(msg)
}

func ReadConf(cfgFile string) *Configuration {
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
		configFatal("Could not read config", err)
	}
	var conf Configuration
	if err := viper.Unmarshal(&conf); err != nil {
		configFatal("Unable to parse the config", err)
	}
	validate := validator.New()
	if err := validate.Struct(&conf); err != nil {
		configFatal("Missing required attributes", err)
	}

	return &conf
}

func LogSetup(debugFlag bool) {
	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
