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
    registry_host: "read-registry.checkrhq.net"
    registry_org: "checkr"
    registry_username: ""
    registry_password: ""
    registry_user_email: ""
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

	dockerBuildEvent := plugins.DockerBuild{
		Action: plugins.Create,
		State:  plugins.Waiting,
		Project: plugins.Project{
			Slug:       "codeamp-circuit",
			Repository: "codeamp/circuit",
		},
		Git: plugins.Git{
			Url:           "https://github.com/codeamp/circuit.git",
			Protocol:      "HTTPS",
			Branch:        "master",
			RsaPrivateKey: "",
			RsaPublicKey:  "",
			Workdir:       viper.GetString("plugins.dockerbuilder.workdir"),
		},
		Feature: plugins.Feature{
			Hash:       "b82f00530a7186d5b03ead5bd3d3600053b71ee7",
			ParentHash: "b5021f702069ac6160fe5f0e9395351a36462c59",
			User:       "Saso Matejina",
			Message:    "Test",
		},
		Registry: plugins.DockerRegistry{
			Host:     viper.GetString("plugins.dockerbuilder.registry_host"),
			Org:      viper.GetString("plugins.dockerbuilder.registry_org"),
			Username: viper.GetString("plugins.dockerbuilder.registry_username"),
			Password: viper.GetString("plugins.dockerbuilder.registry_password"),
			Email:    viper.GetString("plugins.dockerbuilder.registry_user_email"),
		},
		BuildArgs: []plugins.Arg{},
	}

	suite.transistor.Events <- transistor.NewEvent(dockerBuildEvent, nil)

	e = suite.transistor.GetTestEvent("plugins.DockerBuild:status", 60)
	payload := e.Payload.(plugins.DockerBuild)
	assert.Equal(suite.T(), string(plugins.Status), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.Fetching), string(payload.State))

	e = suite.transistor.GetTestEvent("plugins.DockerBuild:status", 600)
	payload = e.Payload.(plugins.DockerBuild)
	assert.Equal(suite.T(), string(plugins.Status), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.Complete), string(payload.State))

}

func TestDockerBuilder(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
