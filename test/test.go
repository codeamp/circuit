package test

import (
	"bytes"
	"strings"
	"time"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func SetupPluginTest(viperConfig []byte) (*transistor.Transistor, error) {
	setupViperConfig(viperConfig)

	pluginConfig := viper.GetStringMap("plugins")
	enabledPluginNames := make([]string, 0, len(pluginConfig))

	for pluginName, _ := range pluginConfig {
		enabledPluginNames = append(enabledPluginNames, pluginName)
	}

	config := transistor.Config{
		Plugins:        pluginConfig,
		EnabledPlugins: enabledPluginNames,
	}

	configLogLevel()
	configLogFormat()

	return transistor.NewTestTransistor(config)
}

func setupViperConfig(viperConfig []byte) {
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CODEAMP")
	viper.AutomaticEnv()

	viper.ReadConfig(bytes.NewBuffer(viperConfig))
}

func configLogLevel() {
	if _logLevel := viper.GetString("log_level"); _logLevel != "" {
		logLevel, err := log.ParseLevel(_logLevel)

		if err != nil {
			log.Fatal(err)
		}

		log.SetLogLevel(logLevel)
	}
}

func configLogFormat() {
	if logFormat := viper.GetString("log_format"); logFormat != "" {
		switch strings.ToLower(logFormat) {
		case "standard":
			break
		case "json":
			fallthrough
		default:
			log.SetLogFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
		}
	}
}
