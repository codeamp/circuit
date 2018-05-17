package kubernetes_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
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
  kubernetes:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("YAML")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("CODEAMP")
	viper.AutomaticEnv()
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin {
		return &kubernetes.Kubernetes{}
	})

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"kubernetes"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

// func (suite *TestSuite) TestCleanupLBOffice() {
// 	suite.transistor.Events <- LBTCPEvent(plugins.GetAction("destroy"), plugins.GetType("office"))

// 	e := suite.transistor.GetTestEvent(plugins.GetEventName("kubernetes:loadbalancer"), plugins.GetAction("status"), 60)
// 	assert.Equal(suite.T(), plugins.GetState("deleted"), e.State, e.StateMessage)
// }

// // Load Balancers Tests
// func (suite *TestSuite) TestLBTCPOffice() {
// 	timer := time.NewTimer(time.Second * 100)
// 	defer timer.Stop()

// 	go func() {
// 		<-timer.C
// 		log.Fatal("TestLBTCPOffice: Test timeout")
// 	}()

// 	suite.transistor.Events <- LBTCPEvent(plugins.GetAction("update"), plugins.GetType("office"))

// 	var e transistor.Event
// 	e = suite.transistor.GetTestEvent(plugins.GetEventName("kubernetes:loadbalancer"), plugins.GetAction("status"), 120)
// 	assert.Equal(suite.T(), plugins.GetState("complete"), e.State, e.StateMessage)
// 	if e.State != plugins.GetState("complete") {
// 		return
// 	}

// 	for {
// 		e = suite.transistor.GetTestEvent(plugins.GetEventName("kubernetes:loadbalancer"), plugins.GetAction("status"), 120)
// 		if e.State != "running" {
// 			break
// 		}
// 	}

// 	suite.transistor.Events <- LBTCPEvent(plugins.GetAction("destroy"), plugins.GetType("office"))

// 	e = suite.transistor.GetTestEvent(plugins.GetEventName("kubernetes:loadbalancer"), plugins.GetAction("status"), 10)
// 	assert.Equal(suite.T(), plugins.GetState("deleted"), e.State)
// }

func strMapKeys(strMap map[string]string) string {
	keys := make([]string, len(strMap))

	i := 0
	for k := range strMap {
		keys[i] = k
		i++
	}

	return strings.Join(keys, "\n")
}

// Deploys Tests
func (suite *TestSuite) TestBasicSuccessDeploy() {
	timer := time.NewTimer(time.Second * 60)
	defer timer.Stop()

	suite.transistor.Events <- BasicReleaseEvent()

	go func() {
		<-timer.C
		log.Fatal("TestBasicSuccessDeploy: Test timeout")
	}()

	var e transistor.Event
	for {
		e = suite.transistor.GetTestEvent("kubernetes:deployment", plugins.GetAction("status"), 30)
		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), plugins.GetState("complete"), e.State)
}

func (suite *TestSuite) TestBasicFailedDeploy() {
	timer := time.NewTimer(time.Second * 60)
	defer timer.Stop()

	go func() {
		<-timer.C
		log.Fatal("TestBasicFailedDeploy: Test timeout")
	}()

	suite.transistor.Events <- BasicFailedReleaseEvent()

	var e transistor.Event
	for {
		e = suite.transistor.GetTestEvent(plugins.GetEventName("kubernetes:deployment"), plugins.GetAction("status"), 30)
		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), plugins.GetState("failed"), e.State)
}

func TestDeployments(t *testing.T) {
	proceed := true

	if err := verifyDeploymentArtifacts(); err != nil {
		proceed = false
		assert.Nil(t, err, err.Error())
	}

	if err := verifyLoadBalancerArtifacts(); err != nil {
		proceed = false
		assert.Nil(t, err, err.Error())
	}

	if proceed {
		suite.Run(t, new(TestSuite))
	}
}

func (suite *TestSuite) TearDownSuite() {
	// TODO:
	// teardown docker-io secret?
	// teardown the deployment / namespaces
	suite.transistor.Stop()
}

func verifyDeploymentArtifacts() error {
	e := BasicReleaseEvent()

	basicReleaseEventArtifacts := map[string]string{
		"user":                  "",
		"password":              "",
		"host":                  "",
		"email":                 "",
		"image":                 "",
		"kubeconfig":            "",
		"client_certificate":    "",
		"client_key":            "",
		"certificate_authority": "",
	}

	for _, artifact := range e.Artifacts {
		delete(basicReleaseEventArtifacts, artifact.Key)
	}

	if len(basicReleaseEventArtifacts) != 0 {
		return errors.New("BasicReleaseEvent\nMissing Artifacts:\n" + strMapKeys(basicReleaseEventArtifacts))
	}

	return nil
}

func verifyLoadBalancerArtifacts() error {
	e := LBTCPEvent(plugins.GetAction("update"), plugins.GetType("office"))

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
		Slug:        "kubernetesloadbalancers",
		Environment: "testing",
		Project:     project,
		ID:          "nginx-test-lb-asdf1234",
	}
	return lbe
}

func LBTCPEvent(action transistor.Action, t plugins.Type) transistor.Event {
	payload := LBDataForTCP(action, t)
	event := transistor.NewEvent(plugins.GetEventName("kubernetes:loadbalancer"), action, payload)

	var kubeConfigPath string
	if kubeConfigPath = os.Getenv("KUBECONFIG_PATH"); kubeConfigPath == "" {
		kubeConfigPath = path.Join(os.Getenv("HOME"), ".kube", "config")
	}

	event.AddArtifact("service", "nginx-test-service-asdf", false)
	event.AddArtifact("name", "nginx-test-lb-asdf1234", false)
	event.AddArtifact("ssl_cert_arn", "arn:1234:arnid", false)
	event.AddArtifact("access_log_s3_bucket", "test-s3-logs-bucket", false)
	event.AddArtifact("type", fmt.Sprintf("%v", t), false)

	// For Kube connectivity
	event.AddArtifact("kubeconfig", kubeConfigPath, false)
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

func BasicFailedReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()
	extension.Release.Services[0].Command = "/bin/false"

	event := transistor.NewEvent(plugins.GetEventName("kubernetes:deployment"), plugins.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func addBasicReleaseExtensionArtifacts(extension plugins.ReleaseExtension, event *transistor.Event) {
	var kubeConfigPath string
	if kubeConfigPath = os.Getenv("KUBECONFIG_PATH"); kubeConfigPath == "" {
		kubeConfigPath = path.Join(os.Getenv("HOME"), ".kube", "config")
	}

	event.AddArtifact("user", "test", false)
	event.AddArtifact("password", "test", false)
	event.AddArtifact("host", "test", false)
	event.AddArtifact("email", "test", false)
	event.AddArtifact("image", "nginx", false)

	for idx := range event.Artifacts {
		event.Artifacts[idx].Source = "dockerbuilder"
	}

	event.AddArtifact("kubeconfig", kubeConfigPath, false)
	event.AddArtifact("client_certificate", "", false)
	event.AddArtifact("client_key", "", false)
	event.AddArtifact("certificate_authority", "", false)
}

func BasicReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()

	event := transistor.NewEvent(plugins.GetEventName("kubernetes:deployment"), plugins.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func BasicReleaseExtension() plugins.ReleaseExtension {

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"

	release := plugins.Release{
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
		Services: []plugins.Service{
			plugins.Service{
				Name: "www",
				Listeners: []plugins.Listener{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
				State: plugins.GetState("waiting"),
				Spec: plugins.ServiceSpec{

					CpuRequest:                    "10m",
					CpuLimit:                      "500m",
					MemoryRequest:                 "1Mi",
					MemoryLimit:                   "500Mi",
					TerminationGracePeriodSeconds: int64(1),
				},
				Replicas: 1,
			},
		},
		HeadFeature: plugins.Feature{
			Hash:       deploytestHash,
			ParentHash: deploytestHash,
			User:       "",
			Message:    "Test",
		},
		Environment: "testing",
	}

	releaseExtension := plugins.ReleaseExtension{
		Slug:    "kubernetesdeployments",
		Release: release,
	}

	return releaseExtension
}
