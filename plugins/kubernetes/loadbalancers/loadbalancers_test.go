package kubernetesloadbalancers_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes/loadbalancers/testdata"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
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
  kubernetesloadbalancers:
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
		EnabledPlugins: []string{"kubernetesloadbalancers"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()

	// Test teardown of any existing LBs
	// suite.transistor.Events <- testdata.TearDownLBTCP(plugins.Internal)
	// _ = suite.agent.GetTestEvent("plugins.Release:status", 60)
	// suite.agent.Events <- testdata.TearDownLBHTTPS(plugins.Internal)
	// _ = suite.agent.GetTestEvent("plugins.LoadBalancer:status", 60)
	// suite.agent.Events <- testdata.TearDownLBTCP(plugins.External)
	// _ = suite.agent.GetTestEvent("plugins.LoadBalancer:status", 60)
	// suite.agent.Events <- testdata.TearDownLBHTTPS(plugins.External)
	// _ = suite.agent.GetTestEvent("plugins.LoadBalancer:status", 60)
}

func (suite *TestSuite) TestLBCreate() {
	var e transistor.Event
	payload := testdata.LBDataForTCP(plugins.Update, plugins.Internal)
	suite.transistor.Events <- transistor.NewEvent(payload, nil)

	e = suite.transistor.GetTestEvent("plugins.Extension:status", 10)
	spew.Dump(e)

	assert.Equal(suite.T(), string(e.Payload.(plugins.Extension).State), string(plugins.Running))

}

func (suite *TestSuite) TestLBDestroy() {
	var e transistor.Event
	suite.transistor.Events <- transistor.NewEvent(testdata.GetDestroyExtension(), nil)
	e = suite.transistor.GetTestEvent("plugins.Extension:status", 10)
	spew.Dump(e)

	assert.Equal(suite.T(), string(e.Payload.(plugins.Extension).State), string(plugins.Running))
}

func TestLoadBalancers(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
