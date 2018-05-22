package gitsync_test

import (
	"bytes"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/gitsync"
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
  gitsync:
    workers: 1
    workdir: /tmp/gitsync
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"gitsync"},
	}

	transistor.RegisterPlugin("gitsync", func() transistor.Plugin {
		return &gitsync.GitSync{}
	})

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestGitSync() {
	var e transistor.Event

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

	event := transistor.NewEvent(plugins.GetEventName("gitsync"), plugins.GetAction("create"), gitSync)
	suite.transistor.Events <- event

	e = suite.transistor.GetTestEvent(plugins.GetEventName("gitsync"), plugins.GetAction("status"), 30)
	assert.Equal(suite.T(), e.State, plugins.GetState("fetching"))

	e = suite.transistor.GetTestEvent(plugins.GetEventName("gitsync"), plugins.GetAction("status"), 30)
	assert.Equal(suite.T(), e.State, plugins.GetState("complete"))
	assert.NotNil(suite.T(), e.Payload().(plugins.GitSync).Commits)
	assert.NotEqual(suite.T(), 0, len(e.Payload().(plugins.GitSync).Commits), "commits should not be empty")
}

func TestGitSync(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
