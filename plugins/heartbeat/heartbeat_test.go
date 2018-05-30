package heartbeat_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/heartbeat"
	"github.com/codeamp/circuit/test"
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
	creatorsMap := map[string]transistor.Creator{
		"heartbeat": func() transistor.Plugin {
			return &heartbeat.Heartbeat{}
		},
	}

	suite.transistor, _ = test.SetupPluginTest(viperConfig, creatorsMap)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestHeartbeat() {
	var e transistor.Event
	var err error

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("heartbeat"), transistor.GetAction("status"), 61)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	payload := e.Payload.(plugins.HeartBeat)
	assert.Equal(suite.T(), "minute", payload.Tick)
}

func TestHeartbeat(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
