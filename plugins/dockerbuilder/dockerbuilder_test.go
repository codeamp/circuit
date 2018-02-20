package dockerbuilder_test

import (
	"bytes"
	"testing"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
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
	formValues["DOCKERBUILDER_USER"] = "test"
	formValues["DOCKERBUILDER_PASSWORD"] = "test"
	formValues["DOCKERBUILDER_EMAIL"] = "test@checkr.com"
	formValues["DOCKERBUILDER_HOST"] = "registry-testing.checkrhq-dev.net:5000"
	formValues["DOCKERBUILDER_ORG"] = "testorg"

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc8"

	dockerBuildEvent := plugins.ReleaseExtension{
		Slug:   "dockerbuilder",
		Action: plugins.GetAction("create"),
		State:  plugins.GetState("waiting"),
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
			Action: plugins.GetAction("create"),
			Slug:   "dockerbuilder",
			Config: formValues,
		},
		Artifacts: make(map[string]string),
	}

	suite.transistor.Events <- transistor.NewEvent(dockerBuildEvent, nil)

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:dockerbuilder", 60)
	payload := e.Payload.(plugins.ReleaseExtension)
	assert.Equal(suite.T(), string(plugins.GetAction("status")), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.GetState("fetching")), string(payload.State))

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:dockerbuilder", 600)
	payload = e.Payload.(plugins.ReleaseExtension)
	assert.Equal(suite.T(), string(plugins.GetAction("status")), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.GetState("complete")), string(payload.State))

	assert.NotEmpty(suite.T(), payload.Artifacts)
	assert.Contains(suite.T(), payload.Artifacts["IMAGE"], deploytestHash)
}

func TestDockerBuilder(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
