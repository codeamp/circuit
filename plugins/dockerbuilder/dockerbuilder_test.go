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
	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"

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
	}

	ev := transistor.NewEvent(dockerBuildEvent, nil)
	ev.AddArtifact("DOCKERBUILDER_USER", "test", false)
	ev.AddArtifact("DOCKERBUILDER_PASSWORD", "test", false)
	ev.AddArtifact("DOCKERBUILDER_EMAIL", "test@checkr.com", false)
	ev.AddArtifact("DOCKERBUILDER_HOST", "registry-testing.checkrhq-dev.net:5000", false)
	ev.AddArtifact("DOCKERBUILDER_ORG", "testorg", false)
	suite.transistor.Events <- ev

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:dockerbuilder", 60)
	payload := e.Payload.(plugins.ReleaseExtension)
	assert.Equal(suite.T(), string(plugins.GetAction("status")), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.GetState("fetching")), string(payload.State))

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:dockerbuilder", 600)
	payload = e.Payload.(plugins.ReleaseExtension)
	assert.Equal(suite.T(), string(plugins.GetAction("status")), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.GetState("complete")), string(payload.State))

	image, err := e.GetArtifact("DOCKERBUILDER_IMAGE")
	if err != nil {
		log.Error(err)
	}

	assert.Equal(suite.T(), image.Value.(string), "registry-testing.checkrhq-dev.net:5000/testorg/checkr-deploy-test:4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9.testing")
}

func TestDockerBuilder(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
