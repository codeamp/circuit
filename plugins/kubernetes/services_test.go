package kubernetes_test

import (
	"fmt"
	"io/ioutil"
	"path"
	_ "strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteServices struct {
	suite.Suite
	transistor *transistor.Transistor
	MockCoreService
}

func (suite *TestSuiteServices) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  kubernetes:
    workers: 1
`)

	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin {
		return &kubernetes.Kubernetes{
			K8sContourNamespacer: &MockContourNamespacer{},
			K8sNamespacer:        &MockKubernetesNamespacer{},
			BatchV1Jobber:        &MockBatchV1Job{},
			CoreServicer:         &suite.MockCoreService,
			CoreSecreter:         &MockCoreSecret{},
		}
	}, plugins.ReleaseExtension{}, plugins.ProjectExtension{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func TestServices(t *testing.T) {
	suite.Run(t, new(TestSuiteServices))
}

func (suite *TestSuiteServices) TestCreateServiceSuccess() {
	suite.transistor.Events <- LBTCPEvent("", transistor.GetAction("create"), plugins.GetType("office"))

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("create"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteServices) TestUpdateServiceSuccess() {
	suite.transistor.Events <- LBTCPEvent("", transistor.GetAction("update"), plugins.GetType("office"))

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("update"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteServices) TestDeleteServiceSuccess() {
	// Create service to delete later
	suite.transistor.Events <- LBTCPEvent("", transistor.GetAction("create"), plugins.GetType("office"))

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("create"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)

	// Now send delete message
	suite.transistor.Events <- LBTCPEvent("", transistor.GetAction("delete"), plugins.GetType("office"))
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("status"), 5)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.State("complete"), e.State)
}

func (suite *TestSuiteServices) TestDeleteServiceNotExistFailed() {
	suite.transistor.Events <- LBTCPEvent("nginx-test-lb-snef-1", transistor.GetAction("delete"), plugins.GetType("office"))

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("status"), 5)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("failed"), e.State)
	assert.Equal(suite.T(), kubernetes.ErrServiceDeleteNotFound.Error(), e.StateMessage)
}

func (suite *TestSuiteServices) TearDownSuite() {
	suite.transistor.Stop()
}

func LBDataForTCP(action transistor.Action, t plugins.Type) plugins.ProjectExtension {
	project := plugins.Project{
		Repository: "checkr/nginx-test-success",
	}

	lbe := plugins.ProjectExtension{
		//		Slug:        "kubernetesloadbalancers",
		Environment: "testing",
		Project:     project,
		ID:          "nginx-test-lb-asdf1234",
	}
	return lbe
}

func LBTCPEvent(name string, action transistor.Action, t plugins.Type) transistor.Event {
	payload := LBDataForTCP(action, t)
	event := transistor.NewEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), action, payload)

	kubeConfigPath := path.Join("testdata", "kubeconfig")
	kubeConfig, _ := ioutil.ReadFile(kubeConfigPath)

	event.AddArtifact("service", "test-12345", false)

	if name == "" {
		event.AddArtifact("name", "nginx-test-lb-asdf1234", false)
	} else {
		event.AddArtifact("name", name, false)
	}
	event.AddArtifact("ssl_cert_arn", "arn:1234:arnid", false)
	event.AddArtifact("access_log_s3_bucket", "test-s3-logs-bucket", false)
	event.AddArtifact("type", fmt.Sprintf("%v", t), false)

	// For Kube connectivity
	event.AddArtifact("kubeconfig", string(kubeConfig), false)
	event.AddArtifact("client_certificate", "", false)
	event.AddArtifact("client_key", "", false)
	event.AddArtifact("certificate_authority", "", false)
	event.AddArtifact("private_source_ranges", "10.0.0.0/16", false)

	var listener_pairs []interface{} = make([]interface{}, 2, 2)
	listener_pairs[0] = map[string]interface{}{
		"serviceProtocol": "TCP",
		"port":            "443",
		"containerPort":   float64(3000),
	}
	listener_pairs[1] = map[string]interface{}{
		"serviceProtocol": "TCP",
		"port":            "444",
		"containerPort":   float64(3001),
	}
	event.AddArtifact("listener_pairs", listener_pairs, false)

	return event
}
