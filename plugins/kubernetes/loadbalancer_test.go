package kubernetes_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
	"github.com/codeamp/circuit/tests"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteLB struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfigLB = []byte(`
plugins:
  kubernetes:
    workers: 1
`)

func (suite *TestSuiteLB) SetupSuite() {
	suite.transistor, _ = tests.SetupPluginTest("kubernetes", viperConfigLB, func() transistor.Plugin {
		return &kubernetes.Kubernetes{Simulated: true}
	})

	go suite.transistor.Run()
}

// Load Balancers Tests
func (suite *TestSuiteLB) TestCleanupLBOffice() {
	suite.transistor.Events <- LBTCPEvent(transistor.GetAction("delete"), plugins.GetType("office"))

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State, e.StateMessage)
}

func (suite *TestSuiteLB) TestLBTCPOffice() {
	suite.transistor.Events <- LBTCPEvent(transistor.GetAction("update"), plugins.GetType("office"))
	eventName := plugins.GetEventName("project:kubernetes:loadbalancer")

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(eventName, transistor.GetAction("status"), 100)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	assert.Equal(suite.T(), transistor.GetState("complete"), e.State, e.StateMessage)
	if e.State != transistor.GetState("complete") {
		return
	}

	for {
		e, err = suite.transistor.GetTestEvent(eventName, transistor.GetAction("status"), 20)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.transistor.Events <- LBTCPEvent(transistor.GetAction("delete"), plugins.GetType("office"))

	e, err = suite.transistor.GetTestEvent(eventName, transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), transistor.GetAction("delete"), e.State)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func strMapKeys(strMap map[string]string) string {
	keys := make([]string, len(strMap))

	i := 0
	for k := range strMap {
		keys[i] = k
		i++
	}

	return strings.Join(keys, "\n")
}

func TestLoadBalancers(t *testing.T) {
	proceed := true

	if err := verifyLoadBalancerArtifacts(); err != nil {
		proceed = false
		assert.Nil(t, err, err.Error())
	}

	if proceed {
		suite.Run(t, new(TestSuiteLB))
	}
}

func (suite *TestSuiteLB) TearDownSuite() {
	// TODO:
	// teardown docker-io secret?
	// teardown the deployment / namespaces
	suite.transistor.Stop()
}

func verifyLoadBalancerArtifacts() error {
	e := LBTCPEvent(transistor.GetAction("update"), plugins.GetType("office"))

	lbTCPArtifacts := map[string]string{
		"service":               "",
		"name":                  "",
		"ssl_cert_arn":          "",
		"access_log_s3_bucket":  "",
		"type":                  "",
		"kubeconfig":            "",
		"client_certificate":    "",
		"client_key":            "",
		"certificate_authority": "",
		"listener_pairs":        "",
	}

	for _, artifact := range e.Artifacts {
		delete(lbTCPArtifacts, artifact.Key)
	}

	if len(lbTCPArtifacts) != 0 {
		return errors.New("LoadBalancer\nMissing Artifacts:\n" + strMapKeys(lbTCPArtifacts))
	}

	return nil
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

func LBTCPEvent(action transistor.Action, t plugins.Type) transistor.Event {
	payload := LBDataForTCP(action, t)
	event := transistor.NewEvent(plugins.GetEventName("project:kubernetes:loadbalancer"), action, payload)

	kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube", "config")
	kubeConfig, _ := ioutil.ReadFile(kubeConfigPath)

	event.AddArtifact("service", "nginx-test-service-asdf", false)
	event.AddArtifact("name", "nginx-test-lb-asdf1234", false)
	event.AddArtifact("ssl_cert_arn", "arn:1234:arnid", false)
	event.AddArtifact("access_log_s3_bucket", "test-s3-logs-bucket", false)
	event.AddArtifact("type", fmt.Sprintf("%v", t), false)

	// For Kube connectivity
	event.AddArtifact("kubeconfig", string(kubeConfig), false)
	event.AddArtifact("client_certificate", "", false)
	event.AddArtifact("client_key", "", false)
	event.AddArtifact("certificate_authority", "", false)

	var listener_pairs []interface{} = make([]interface{}, 2, 2)
	listener_pairs[0] = map[string]interface{}{
		"serviceProtocol": "TCP",
		"port":            "443",
		"containerPort":   "3000",
	}
	listener_pairs[1] = map[string]interface{}{
		"serviceProtocol": "TCP",
		"port":            "444",
		"containerPort":   "3001",
	}
	event.AddArtifact("listener_pairs", listener_pairs, false)

	return event
}
