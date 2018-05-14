package testdata

import (
	"fmt"
	"os"
	"path"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
)

func GetCreateProjectExtension() plugins.ProjectExtension {
	d := GetBasicProjectExtension()
	d.Action = plugins.GetAction("create")
	d.State = plugins.GetState("waiting")
	return d
}

func GetDestroyProjectExtension() plugins.ProjectExtension {
	d := GetBasicProjectExtension()
	d.Action = plugins.GetAction("destroy")
	d.State = plugins.GetState("waiting")
	return d
}

func GetBasicProjectExtension() plugins.ProjectExtension {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig == "" {
		kubeconfig = path.Join(os.Getenv("HOME"), ".kube", "config")
	}

	extensionEvent := plugins.ProjectExtension{
		Slug:        "kubernetesloadbalancers",
		Environment: "testing",
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		Project: plugins.Project{
			Repository: "checkr/deploy-test",
			// Services: []plugins.Service{
			// 	plugins.Service{},
			// },
		},
	}

	return extensionEvent
}

func LBDataForTCP(action plugins.Action, t plugins.Type) plugins.ProjectExtension {
	project := plugins.Project{
		Repository: "checkr/nginx-test-success",
	}

	lbe := plugins.ProjectExtension{
		Slug:        "kubernetesloadbalancers",
		Action:      action,
		Environment: "testing",
		Project:     project,
		ID:          "nginx-test-lb-asdf1234",
	}
	return lbe
}

func LBTCPEvent(action plugins.Action, t plugins.Type) transistor.Event {
	data := LBDataForTCP(action, t)
	event := transistor.NewEvent(data, nil)

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

	event := transistor.NewEvent(extension, nil)
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

	event := transistor.NewEvent(extension, nil)
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
		Action:  plugins.GetAction("create"),
		State:   plugins.GetState("waiting"),
		Release: release,
	}

	return releaseExtension
}
