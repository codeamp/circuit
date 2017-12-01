package dockerbuilder_test

import (
	"bytes"
	"testing"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
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
  dockerbuilder:
    workers: 1
    workdir: "/tmp/dockerbuilder"   
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"dockerbuilder"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestDockerBuilder() {
	var e transistor.Event

	log.SetLogLevel(logrus.DebugLevel)

	formValues := make(map[string]interface{})
	formValues["USER"] = "test"
	formValues["PASSWORD"] = "test"
	formValues["EMAIL"] = "test@checkr.com"
	formValues["HOST"] = "registry-testing.checkrhq-dev.net:5000"
	formValues["ORG"] = "testorg"

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"

	dockerBuildEvent := plugins.ReleaseExtension{
		Slug:   "dockerbuilder",
		Action: plugins.Create,
		State:  plugins.Waiting,
		Release: plugins.Release{
			Project: plugins.Project{
				Repository: "checkr/deploy-test",
			},

			Git: plugins.Git{
				Url:           "https://github.com/checkr/deploy-test.git",
				Protocol:      "HTTPS",
				Branch:        "master",
				RsaPrivateKey: "",
				RsaPublicKey:  "",
				Workdir:       "/tmp/something",
			},

			HeadFeature: plugins.Feature{
				Hash:       deploytestHash,
				ParentHash: deploytestHash,
				User:       "",
				Message:    "Test",
			},
			Environment: "testing",
		},
		Extension: plugins.Extension{
			Action:     plugins.Create,
			Slug:       "dockerbuilder",
			FormValues: formValues,
		},
		Artifacts: make(map[string]string),
	}

	suite.transistor.Events <- transistor.NewEvent(dockerBuildEvent, nil)

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:dockerbuilder", 60)
	payload := e.Payload.(plugins.ReleaseExtension)
	spew.Dump(payload.StateMessage)
	assert.Equal(suite.T(), string(plugins.Status), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.Fetching), string(payload.State))

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:dockerbuilder", 600)
	payload = e.Payload.(plugins.ReleaseExtension)
	spew.Dump(payload.StateMessage)
	assert.Equal(suite.T(), string(plugins.Status), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.Complete), string(payload.State))

	assert.NotEmpty(suite.T(), payload.Artifacts)
	assert.Contains(suite.T(), payload.Artifacts["IMAGE"], deploytestHash)
}

func TestDockerBuilder(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
