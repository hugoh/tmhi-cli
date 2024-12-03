package internal

import (
	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/sirupsen/logrus"
)

const (
	DefaultConfig string = ".tmhi-cli.yaml"
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

func configFatal(msg string, cfgFile string, err error) {
	logrus.WithField("file", cfgFile).WithError(err).Fatal(msg)
}

func ReadConf(cfgFile string) *Configuration {
	k := koanf.New(".")

	if cfgFile == "" {
		cfgFile = DefaultConfig
	}
	if err := k.Load(file.Provider(cfgFile), yaml.Parser()); err != nil {
		configFatal("Could not read config", cfgFile, err)
	}
	logrus.WithField("file", cfgFile).Debug("[Config] config file used")
	var conf Configuration
	if err := k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{}); err != nil {
		configFatal("Unable to parse the config", cfgFile, err)
	}

	validate := validator.New()
	if err := validate.Struct(&conf); err != nil {
		configFatal("Missing required attributes", cfgFile, err)
	}

	return &conf
}

func LogSetup(debugFlag bool) {
	if debugFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
