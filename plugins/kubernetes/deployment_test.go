package kubernetes_test

import (
	"io/ioutil"
	"path"
	"testing"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	v1 "k8s.io/api/batch/v1"
	"k8s.io/api/extensions/v1beta1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestSuiteDeployment struct {
	suite.Suite
	transistor *transistor.Transistor
	MockBatchV1Job
	MockExtDeployment
	MockCorePod
}

func (suite *TestSuiteDeployment) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  kubernetes:
    workers: 1
`)

	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin {
		return &kubernetes.Kubernetes{
			K8sContourNamespacer: &MockContourNamespacer{},
			K8sNamespacer:        &MockKubernetesNamespacer{},
			BatchV1Jobber:        &suite.MockBatchV1Job,
			CoreSecreter:         &MockCoreSecret{},
			CorePodder:           &suite.MockCorePod,
			ExtReplicaSetter:     &MockExtReplicaSet{},
			ExtDeploymenter:	  &suite.MockExtDeployment,
		}
	}, plugins.ReleaseExtension{}, plugins.ProjectExtension{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

// Deploys Tests
func (suite *TestSuiteDeployment) TestBasicSuccessOneShotDeploy() {
	suite.transistor.Events <- ReleaseEvent(BasicReleaseExtensionJob())
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

func (suite *TestSuiteDeployment) TestBasicFailedOneShotDeploy() {
	suite.transistor.Events <- ReleaseEvent(BasicReleaseExtensionJob())
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
	assert.Equal(suite.T(), kubernetes.ErrDeployJobStarting.Error(), e.StateMessage)
}

func (suite *TestSuiteDeployment) TestFailedDeployUnwindFirstDeploy() {
	suite.transistor.Events <- ReleaseEvent(FailingReleaseExtension())
	suite.MockBatchV1Job.StatusOverride = v1.JobStatus{Succeeded: 1}
	suite.MockExtDeployment.DeploymentStatusOverride = &v1beta1.DeploymentStatus{UnavailableReplicas: 1}

	suite.MockCorePod.Pods = []corev1.Pod{
		corev1.Pod{
			ObjectMeta: meta_v1.ObjectMeta{
				OwnerReferences: []meta_v1.OwnerReference{
					meta_v1.OwnerReference{
						Kind: "ReplicaSet",
						Name: "",
					},
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{corev1.ContainerStatus{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}}},
			},
		},
	}

	var e transistor.Event
	var err error
	for {
		log.Warn("waiting for event")
		e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("status"), 30)
		log.Warn(e.Event())
		log.Warn(e.StateMessage)

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
	assert.Equal(suite.T(), kubernetes.ErrDeployPodWaitingForeverUnwindingDeploy.Error(), e.StateMessage)
}

func (suite *TestSuiteDeployment) TestFailedDeployUnwind() {
	suite.transistor.Events <- ReleaseEvent(FailingReleaseExtension())
	suite.MockBatchV1Job.StatusOverride = v1.JobStatus{Succeeded: 1}
	suite.MockExtDeployment.DeploymentStatusOverride = &v1beta1.DeploymentStatus{UnavailableReplicas: 1}
	suite.MockCorePod.Pods = []corev1.Pod{
		corev1.Pod{
			ObjectMeta: meta_v1.ObjectMeta{
				OwnerReferences: []meta_v1.OwnerReference{
					meta_v1.OwnerReference{
						Kind: "ReplicaSet",
						Name: "",
					},
				},
			},
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{corev1.ContainerStatus{State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}}},
			},
		},
	}


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
	assert.Equal(suite.T(), kubernetes.ErrDeployPodWaitingForeverUnwindingDeploy.Error(), e.StateMessage)
}


func (suite *TestSuiteDeployment) TestDeployFailureNoSecrets() {
	suite.transistor.Events <- BuildReleaseEvent(BasicReleaseExtensionNoSecrets())

	// Setting this to succeeded. If we pass the secrets test for some (failure) reason
	// we want to catch it with the assert at the end of this (failed/completed)
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
	assert.Equal(suite.T(), transistor.GetState("failed"), e.State)
	assert.Equal(suite.T(), kubernetes.ErrDeployNoSecrets.Error(), e.StateMessage)
}

func TestDeployments(t *testing.T) {
	suite.Run(t, new(TestSuiteDeployment))
}

func (suite *TestSuiteDeployment) TearDownSuite() {
	suite.transistor.Stop()
}

func FailingReleaseExtension() plugins.ReleaseExtension{
	extension := BasicReleaseExtensionJobAndService()
	extension.Release.Services[0].Command = "/bin/false"

	return extension
}

func BasicFailedReleaseEvent() transistor.Event {
	extension := BasicReleaseExtensionJob()
	extension.Release.Services[0].Command = "/bin/false"

	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(&extension, &event)

	return event
}

func addBasicReleaseExtensionArtifacts(extension *plugins.ReleaseExtension, event *transistor.Event) {
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
	extension := BasicReleaseExtensionJob()

	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(&extension, &event)

	return event
}

func ReleaseEvent(extension plugins.ReleaseExtension) transistor.Event {
	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), extension)
	addBasicReleaseExtensionArtifacts(&extension, &event)

	return event
}

func BuildReleaseEvent(extension *plugins.ReleaseExtension) transistor.Event {
	event := transistor.NewEvent(plugins.GetEventName("release:kubernetes:deployment"), transistor.GetAction("create"), *extension)
	addBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func BasicReleaseExtensionNoSecrets() *plugins.ReleaseExtension {
	extension := BasicReleaseExtensionJob()

	extension.Release.Secrets = nil
	return &extension
}

func BasicReleaseExtensionJob() plugins.ReleaseExtension {

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"
	uuid := fmt.Sprintf("%s", uuid.NewV4())[:4]	

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
				Name: fmt.Sprintf("www%s", uuid),
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


func BasicReleaseExtensionJobAndService() plugins.ReleaseExtension {

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"
	uuid := fmt.Sprintf("%s", uuid.NewV4())[:4]	

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
				Name: fmt.Sprintf("ws%s", uuid),
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
			{
				Name: fmt.Sprintf("www%s", uuid),
				Listeners: []plugins.Listener{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
				Action: transistor.GetAction("create"),
				State:  transistor.GetState("waiting"),
				Spec: plugins.ServiceSpec{

					CpuRequest:                    "10m",
					CpuLimit:                      "500m",
					MemoryRequest:                 "1Mi",
					MemoryLimit:                   "500Mi",
					TerminationGracePeriodSeconds: int64(1),
				},
				Replicas: 1,
				Type:     "general",
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