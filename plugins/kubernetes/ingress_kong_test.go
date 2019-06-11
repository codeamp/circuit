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
	"github.com/kevholditch/gokong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteKong struct {
	suite.Suite
	transistor *transistor.Transistor
	MockCoreService
}

func (suite *TestSuiteKong) SetupSuite() {
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

func TestKong(t *testing.T) {
	suite.Run(t, new(TestSuiteKong))
}

func (suite *TestSuiteKong) TestCreateKongPrivateIPSuccess() {
	suite.transistor.Events <- KongEvent("", transistor.GetAction("create"), plugins.GetType("clusterip"))

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:ingresskong"), transistor.GetAction("create"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:ingresskong"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteKong) TestCreateKongLBCreateAndDeleteSuccess() {
	// TEST CREATE
	suite.transistor.Events <- KongEvent("", transistor.GetAction("create"), plugins.GetType("loadbalancer"))

	kongConfig := gokong.NewDefaultConfig()
	kongConfig.HostAddress = "http://kong:8001"
	kongClient := gokong.NewClient(kongConfig)

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:ingresskong"), transistor.GetAction("create"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:ingresskong"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)

	var serviceResults []*gokong.Service
	var serviceQuery gokong.ServiceQueryString

	serviceQuery = gokong.ServiceQueryString{Offset: 0, Size: 1000}
	serviceResults, err = kongClient.Services().GetServices(&serviceQuery)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), len(serviceResults), 1)

	// TEST DELETE
	suite.transistor.Events <- KongEvent("", transistor.GetAction("delete"), plugins.GetType("loadbalancer"))

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:ingresskong"), transistor.GetAction("delete"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:ingresskong"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)

	serviceQuery = gokong.ServiceQueryString{Offset: 0, Size: 1000}
	serviceResults, err = kongClient.Services().GetServices(&serviceQuery)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), len(serviceResults), 0)
}

func KongData(action transistor.Action, t plugins.Type) plugins.ProjectExtension {
	project := plugins.Project{
		Repository: "checkr/nginx-test-success",
	}

	lbe := plugins.ProjectExtension{
		Environment: "testing",
		Project:     project,
		ID:          "nginx-test-lb-asdf1234",
	}
	return lbe
}

func KongEvent(name string, action transistor.Action, t plugins.Type) transistor.Event {
	payload := KongData(action, t)
	event := transistor.NewEvent(plugins.GetEventName("project:kubernetes:ingresskong"), action, payload)

	kubeConfigPath := path.Join("testdata", "kubeconfig")
	kubeConfig, _ := ioutil.ReadFile(kubeConfigPath)

	// For Kube connectivity
	event.AddArtifact("kubeconfig", string(kubeConfig), false)
	event.AddArtifact("client_certificate", "", false)
	event.AddArtifact("client_key", "", false)
	event.AddArtifact("certificate_authority", "", false)

	event.AddArtifact("protocol", "TCP", false)
	event.AddArtifact("service", "www:80", false)

	if t == plugins.GetType("clusterip") {
		event.AddArtifact("type", fmt.Sprintf("%v", t), false)
		return event
	}

	event.AddArtifact("type", fmt.Sprintf("%v", t), false)

	event.AddArtifact("upstream_apex_domains", "test.com,test.net", false)
	event.AddArtifact("ingress_controllers", `
[
	{
		"name": "Private",
		"id": "kong-private",
		"api": "http://kong:8001",
		"elb": "internal-elb-url"
	},
	{
		"name": "Public",
		"id": "kong-public",
		"api": "http://kong:8001",
		"elb": "elb-url"
	}
]`, false)

	event.AddArtifact("controlled_apex_domain", "test.net", false)
	event.AddArtifact("ingress", "kong-private", false)

	upstreamRoutes := []interface{}{
		map[string]interface{}{
			"domains": []interface{}{
				map[string]interface{}{
					"apex":      "test.net",
					"subdomain": "test",
				},
			},
			"paths":   "/test",
			"methods": "GET,POST",
		},
	}

	event.AddArtifact("upstream_routes", upstreamRoutes, false)

	return event
}
