package gitsync_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/gitsync"
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
  gitsync:
    workers: 1
    workdir: /tmp/gitsync
`)

func (suite *TestSuite) SetupSuite() {
	suite.transistor, _ = tests.SetupPluginTest("gitsync", viperConfig, func() transistor.Plugin {
		return &gitsync.GitSync{}
	})
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestGitSync() {
	var e transistor.Event
	var err error

	gitSync := plugins.GitSync{
		Project: plugins.Project{
			Repository: "codeamp/circuit",
		},
		Git: plugins.Git{
			Url:           "https://github.com/codeamp/circuit.git",
			Protocol:      "HTTPS",
			Branch:        "master",
			RsaPrivateKey: "",
			RsaPublicKey:  "",
		},
		From: "",
	}

	event := transistor.NewEvent(plugins.GetEventName("gitsync"), transistor.GetAction("create"), gitSync)
	suite.transistor.Events <- event

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("gitsync"), transistor.GetAction("status"), 30)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), e.State, transistor.GetState("running"))

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("gitsync"), transistor.GetAction("status"), 30)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), e.State, transistor.GetState("complete"))
	assert.NotNil(suite.T(), e.Payload.(plugins.GitSync).Commits)
	assert.NotEqual(suite.T(), 0, len(e.Payload.(plugins.GitSync).Commits), "commits should not be empty")
}

func TestGitSync(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
