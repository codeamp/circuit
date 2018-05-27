package tests

import (
	"bytes"
	"strings"

	"github.com/codeamp/transistor"
	"github.com/spf13/viper"
)

func SetupPluginTest(pluginName string, viperConfig []byte, creator transistor.Creator) (*transistor.Transistor, error) {
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CODEAMP")
	viper.AutomaticEnv()
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	transistor.RegisterPlugin(pluginName, creator)

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{pluginName},
	}

	return transistor.NewTestTransistor(config)
}
