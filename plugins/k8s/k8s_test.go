package k8s_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/k8s"
	"github.com/codeamp/circuit/plugins/k8s/testdata"
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
  k8s:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("YAML")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CODEAMP")
	viper.AutomaticEnv()
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	transistor.RegisterPlugin("k8s", func() transistor.Plugin {
		return &k8s.K8s{}
	})

	log.Debug("issue here")
	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"k8s"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()

	suite.transistor.Events <- transistor.NewEvent(testdata.LBDataForTCP(plugins.GetAction("destroy"), plugins.GetType("office")), nil)
	_ = suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 60)
}

// Deploys Tests
func (suite *TestSuite) TestBasicSuccessDeploy() {
	var e transistor.Event
	suite.transistor.Events <- transistor.NewEvent(testdata.BasicReleaseExtension(), nil)
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 20)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.GetState("running")))
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 120)
	suite.T().Log(e.Payload.(plugins.ReleaseExtension).StateMessage)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.GetState("complete")))
}

func (suite *TestSuite) TestBasicFailedDeploy() {
	var e transistor.Event
	suite.transistor.Events <- transistor.NewEvent(testdata.BasicFailedReleaseExtension(), nil)
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 20)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.GetState("running")))
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status", 120)
	suite.T().Log(e.Payload.(plugins.ReleaseExtension).StateMessage)
	assert.Equal(suite.T(), string(e.Payload.(plugins.ReleaseExtension).State), string(plugins.GetState("failed")))
}

// Load Balancers Tests
func (suite *TestSuite) TestLBTCPOffice() {
	var e transistor.Event
	payload := testdata.LBDataForTCP(plugins.GetAction("update"), plugins.GetType("office"))
	suite.transistor.Events <- transistor.NewEvent(payload, nil)

	e = suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 120)
	assert.Equal(suite.T(), string(plugins.GetAction("completed")), string(e.Payload.(plugins.ProjectExtension).State))

	payload = testdata.LBDataForTCP(plugins.GetAction("destroy"), plugins.GetType("office"))
	suite.transistor.Events <- transistor.NewEvent(payload, nil)
	e = suite.transistor.GetTestEvent("plugins.ProjectExtension:status", 120)
	assert.Equal(suite.T(), string(plugins.GetAction("deleted")), string(e.Payload.(plugins.ProjectExtension).State))

}

func TestDeployments(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) TearDownSuite() {
	// TODO:
	// teardown docker-io secret?
	// teardown the deployment / namespaces
	suite.transistor.Stop()
}
