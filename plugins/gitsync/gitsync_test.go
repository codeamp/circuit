package gitsync_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/gitsync"
	log "github.com/codeamp/logger"
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
	log.Info("fetching")
	assert.Equal(suite.T(), e.State, plugins.GetState("fetching"))

	created := time.Now()
	for i := 0; i < 5; i++ {
		e = suite.transistor.GetTestEvent(plugins.GetEventName("gitsync:commit"), plugins.GetAction("create"), 60)
		payload := e.Payload().(plugins.GitCommit)

		log.Info(payload)

		// assert.Equal(suite.T(), payload.Repository, payload.Project.Repository)
		// assert.Equal(suite.T(), payload.Ref, fmt.Sprintf("refs/heads/%s", payload.Git.Branch))
		assert.True(suite.T(), payload.Created.Before(created), "Commit created time is older than previous commit")
	}

	e = suite.transistor.GetTestEvent(plugins.GetEventName("gitsync"), plugins.GetAction("status"), 60)
	assert.Equal(suite.T(), e.State, plugins.GetState("complete"))
}

func TestGitSync(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
