package kubernetes_test

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	v1 "k8s.io/api/batch/v1"
)

type TestSuiteDeployment struct {
	suite.Suite
	transistor *transistor.Transistor
	MockBatchV1Job
}

func (suite *TestSuiteDeployment) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  kubernetes:
    workers: 1
`)

	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin {
		return &kubernetes.Kubernetes{K8sContourNamespacer: &MockContourNamespacer{}, K8sNamespacer: &MockKubernetesNamespacer{}, BatchV1Jobber: &suite.MockBatchV1Job}
	}, plugins.ReleaseExtension{}, plugins.ProjectExtension{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

// Deploys Tests
func (suite *TestSuiteDeployment) TestBasicSuccessDeploy() {
	suite.transistor.Events <- BasicReleaseEvent()
	suite.MockBatchV1Job.StatusOverride = v1.JobStatus{Succeeded: 1}

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

func (suite *TestSuiteDeployment) TestBasicFailedDeploy() {
	suite.transistor.Events <- BasicFailedReleaseEvent()
	suite.MockBatchV1Job.StatusOverride = v1.JobStatus{Failed: 1}

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
	suite.Run(t, new(TestSuiteDeployment))
}

func (suite *TestSuiteDeployment) TearDownSuite() {
	suite.transistor.Stop()
}

func BasicFailedReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()
	extension.Release.Services[0].Command = "/bin/false"

	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func addBasicReleaseExtensionArtifacts(extension plugins.ReleaseExtension, event *transistor.Event) {
	kubeConfigPath := path.Join("testdata", "kubeconfig")
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
			Slug:       "checkr-deploy-test",
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
				Type:     "one-shot",
			},
		},
		HeadFeature: plugins.Feature{
			Hash:       deploytestHash,
			ParentHash: deploytestHash,
			User:       "",
			Message:    "Test",
		},
		Environment: "testing",
		Secrets: []plugins.Secret{
			{
				Key:   "secret-key",
				Value: "secret-value",
				Type:  plugins.GetType("internal"),
			},
		},
	}

	releaseExtension := plugins.ReleaseExtension{
		//		Slug:    "kubernetesdeployments",
		Release: release,
	}

	return releaseExtension
}
