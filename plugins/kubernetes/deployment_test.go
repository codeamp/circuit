package kubernetes_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
	"github.com/codeamp/circuit/tests"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteDeployments struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfigDeployments = []byte(`
plugins:
  kubernetes:
    workers: 1
`)

func (suite *TestSuiteDeployments) SetupSuite() {
	suite.transistor, _ = tests.SetupPluginTest("kubernetes", viperConfigDeployments, func() transistor.Plugin {
		return &kubernetes.Kubernetes{Simulated: true}
	})

	go suite.transistor.Run()
}

// Load Balancers Tests
func (suite *TestSuiteDeployments) TestCleanupLBOffice() {
	suite.transistor.Events <- LBTCPEvent(transistor.GetAction("delete"), plugins.GetType("office"))

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("release:kubernetes:loadbalancer"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), transistor.GetState("deleted"), e.State, e.StateMessage)
}

func (suite *TestSuiteDeployments) TestLBTCPOffice() {
	suite.transistor.Events <- LBTCPEvent(transistor.GetAction("update"), plugins.GetType("office"))

	var e transistor.Event
	var err error
	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:kubernetes:loadbalancer"), transistor.GetAction("status"), 20)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}

	assert.Equal(suite.T(), transistor.GetState("complete"), e.State, e.StateMessage)
	if e.State != transistor.GetState("complete") {
		return
	}

	for {
		e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:kubernetes:loadbalancer"), transistor.GetAction("status"), 20)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.transistor.Events <- LBTCPEvent(transistor.GetAction("delete"), plugins.GetType("office"))

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:kubernetes:loadbalancer"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error())
		return
	}
	assert.Equal(suite.T(), transistor.GetState("deleted"), e.State)
}

// Deploys Tests
func (suite *TestSuiteDeployments) TestBasicSuccessDeploy() {
	suite.transistor.Events <- BasicReleaseEvent()

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("release:kubernetes:deployment", transistor.GetAction("status"), 30)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteDeployments) TestBasicFailedDeploy() {
	suite.transistor.Events <- BasicFailedReleaseEvent()

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("status"), 30)

		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("failed"), e.State)
}

func TestDeployments(t *testing.T) {
	proceed := true

	if err := verifyDeploymentArtifacts(); err != nil {
		proceed = false
		assert.Nil(t, err, err.Error())
	}

	if proceed {
		suite.Run(t, new(TestSuiteDeployments))
	}
}

func (suite *TestSuiteDeployments) TearDownSuite() {
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

func BasicFailedReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()
	extension.Release.Services[0].Command = "/bin/false"

	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func addBasicReleaseExtensionArtifacts(extension plugins.ReleaseExtension, event *transistor.Event) {
	kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube", "config")
	kubeConfig, _ := ioutil.ReadFile(kubeConfigPath)

	event.AddArtifact("user", "test", false)
	event.AddArtifact("password", "test", false)
	event.AddArtifact("host", "test", false)
	event.AddArtifact("email", "test", false)
	event.AddArtifact("image", "nginx", false)

	for idx := range event.Artifacts {
		event.Artifacts[idx].Source = "dockerbuilder"
	}

	event.AddArtifact("kubeconfig", string(kubeConfig), false)
	event.AddArtifact("client_certificate", "", false)
	event.AddArtifact("client_key", "", false)
	event.AddArtifact("certificate_authority", "", false)
}

func BasicReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()

	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), extension)
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
			{
				Name: "www",
				Listeners: []plugins.Listener{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
				State: transistor.GetState("waiting"),
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
		//		Slug:    "kubernetesdeployments",
		Release: release,
	}

	return releaseExtension
}
