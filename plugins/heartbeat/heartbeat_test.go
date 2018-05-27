package heartbeat_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/heartbeat"
	"github.com/codeamp/circuit/tests"
	"github.com/codeamp/transistor"
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
	suite.transistor, _ = tests.SetupPluginTest("heartbeat", viperConfig, func() transistor.Plugin {
		return &heartbeat.Heartbeat{}
	})
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestHeartbeat() {
	var e transistor.Event

	e = suite.transistor.GetTestEvent(plugins.GetEventName("heartbeat"), transistor.GetAction("status"), 61)
	payload := e.Payload.(plugins.HeartBeat)
	assert.Equal(suite.T(), "minute", payload.Tick)
}

func TestHeartbeat(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
