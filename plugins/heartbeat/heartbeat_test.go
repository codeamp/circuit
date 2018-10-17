package heartbeat_test

import (
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/heartbeat"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

type CronMock struct{}

func (c CronMock) NewCronJob(month, day, weekday, hour, minute, second int8, task func(time.Time)) {
	log.Debug("Mocked Cron Response - Firing Immediately")
	task(time.Now())
}

var viperConfig = []byte(`
plugins:
  heartbeat:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	log.Warn("SETUP SUITE REGISTERING HeartBeat")
	transistor.RegisterPlugin("heartbeat", func() transistor.Plugin {
		return &heartbeat.Heartbeat{Cron: CronMock{}}
	}, plugins.HeartBeat{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
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
	assert.Contains(suite.T(), []string{"minute", "hour"}, payload.Tick)
}

func TestHeartbeat(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
