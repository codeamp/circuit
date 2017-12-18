package testdata

import (
	"os"

	"github.com/codeamp/circuit/plugins"
)

func BasicFailedReleaseExtension() plugins.ReleaseExtension {
	d := BasicReleaseExtension()
	d.Release.Services[0].Command = "/bin/false"
	return d
}

func BasicReleaseExtension() plugins.ReleaseExtension {
	var formValues map[string]interface{}
	var kubeconfig string
	// If this is not set the test will use inClusterConfig
	kubeconfig = os.Getenv("KUBECONFIG")
	formValues["KUBECONFIG"] = kubeconfig
	formValues["DOCKERBUILDER_PASSWORD"] = "test"
	formValues["DOCKERBUILDER_USER"] = "test"
	formValues["DOCKERBUILDER_EMAIL"] = "test"
	formValues["DOCKERBUILDER_HOST"] = "test"

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"
	artifacts := make(map[string]string)
	artifacts["IMAGE"] = "dev-registry.checkrhq.net/checkr/checkr-deploy-test:latest"

	releaseEvent := plugins.Release{
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
				Name:    "www",
				Command: "nginx -g 'daemon off';",
				Listeners: []plugins.Listener{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
				State: plugins.Waiting,
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
		Artifacts:   artifacts,
	}

	releaseExtensionEvent := plugins.ReleaseExtension{
		Slug:    "kubernetesdeployments",
		Action:  plugins.Create,
		State:   plugins.Waiting,
		Release: releaseEvent,
		Extension: plugins.Extension{
			Action:     plugins.Create,
			Slug:       "kubernetesdeployments",
			FormValues: formValues,
		},
		Artifacts: artifacts,
	}

	return releaseExtensionEvent
}
