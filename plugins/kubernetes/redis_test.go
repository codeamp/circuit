package kubernetes_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteRedis struct {
	suite.Suite
	transistor       *transistor.Transistor
	KubernetesPlugin kubernetes.Kubernetes
}

func (suite *TestSuiteRedis) SetupRedisSuite() {
	var viperConfig = []byte(`
plugins:
  kubernetes:
    workers: 1
`)

	suite.KubernetesPlugin = kubernetes.Kubernetes{
		K8sNamespacer:    &MockKubernetesNamespacer{},
		CoreServicer:     &MockCoreService{},
		CoreSecreter:     &MockCoreSecret{},
		CoreDeploymenter: &MockCoreDeployment{},
	}

	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin { return &suite.KubernetesPlugin }, plugins.ReleaseExtension{}, plugins.ProjectExtension{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

// Deploys Tests
func (suite *TestSuiteRedis) TestBasicSuccessRedisDeployment() {
	projectExtension := plugins.ProjectExtension{
		ID: "id",
		Project: plugins.Project{
			ID:         "id",
			Slug:       "foo-project",
			Repository: "project",
		},
		Environment: "fooenv",
	}

	payload := transistor.NewEvent(plugins.GetEventName("project:kubernetes:redis"), transistor.GetAction("create"), projectExtension)

	suite.transistor.Events <- payload

	var e transistor.Event
	e, err := suite.transistor.GetTestEvent("project:kubernetes:redis", transistor.GetAction("status"), 30)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(TestSuiteRedis))
}

func (suite *TestSuiteRedis) TearDownSuite() {
	suite.transistor.Stop()
}
