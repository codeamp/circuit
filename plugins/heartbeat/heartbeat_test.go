package heartbeat_test

import (
	"bytes"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/heartbeat"
	"github.com/codeamp/transistor"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  heartbeat:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	transistor.RegisterPlugin("heartbeat", func() transistor.Plugin {
		return &heartbeat.Heartbeat{}
	})

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"heartbeat"},
	}
	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestHeartbeat() {
	var e transistor.Event

	e = suite.transistor.GetTestEvent(plugins.GetEventName("heartbeat"), plugins.GetAction("status"), 61)
	payload := e.Payload.(plugins.HeartBeat)
	assert.Equal(suite.T(), "minute", payload.Tick)
}

func TestHeartbeat(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
