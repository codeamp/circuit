package kubernetesdeployments_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes/deployments/testdata"
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
  kubernetesdeployments:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("YAML")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CODEAMP")
	viper.AutomaticEnv()
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"kubernetesdeployments"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	// TODO:
	// teardown docker-io secret?
	// teardown the deployment / namespaces
	suite.transistor.Stop()
}

func (suite *TestSuite) TestBasicSuccessDeploy() {
	var e transistor.Event
	suite.transistor.Events <- transistor.NewEvent(testdata.BasicReleaseExtension(), nil)
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 20)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.Running))
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 120)
	suite.T().Log(e.Payload.(plugins.ReleaseExtension).StateMessage)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.Complete))
}

func (suite *TestSuite) TestBasicFailedDeploy() {
	var e transistor.Event
	suite.transistor.Events <- transistor.NewEvent(testdata.BasicFailedReleaseExtension(), nil)
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 20)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.Running))
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 120)
	suite.T().Log(e.Payload.(plugins.ReleaseExtension).StateMessage)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.Failed))
}

func TestDeployments(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
