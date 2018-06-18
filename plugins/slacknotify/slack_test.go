package slacknotify_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
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
  slack:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestSlack() {

	var ev, re transistor.Event
	var err error
	// var re transistor.Event

	ev = transistor.NewEvent(plugins.GetEventName("slack"), transistor.GetAction("create"), nil)
	ev.AddArtifact("webhook_url", "https://hooks.slack.com/services/token/token/valid_token", false)
	ev.AddArtifact("channel", "devops-test", false)

	suite.transistor.Events <- ev

	re, err = suite.transistor.GetTestEvent(plugins.GetEventName("slack"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error)
		return
	}
	assert.Equal(suite.T(), re.State, transistor.GetState("complete"))

	ev = transistor.NewEvent(plugins.GetEventName("slack"), transistor.GetAction("create"), nil)
	ev.AddArtifact("webhook_url", "https://hooks.slack.com/services/token/token/invalid_token", false)
	ev.AddArtifact("channel", "devops-test", false)

	suite.transistor.Events <- ev

	re, err = suite.transistor.GetTestEvent(plugins.GetEventName("slack"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error)
		return
	}
	assert.Equal(suite.T(), re.State, transistor.GetState("failed"))

	ev = transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("status"), nil)
	ev.AddArtifact("webhook_url", "https://hooks.slack.com/services/token/token/valid_token", false)
	ev.AddArtifact("channel", "devops-test", false)

	ev.State = transistor.GetState("complete")
	suite.transistor.Events <- ev
	_, err = suite.transistor.GetTestEvent(plugins.GetEventName("slack"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error)
		return
	}

	ev.State = transistor.GetState("failed")
	suite.transistor.Events <- ev
	_, err = suite.transistor.GetTestEvent(plugins.GetEventName("slack"), transistor.GetAction("status"), 100)
	if err != nil {
		assert.NotNil(suite.T(), err, err.Error)
		return
	}
}

func TestSlack(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
