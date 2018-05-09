package testdata

import (
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

	// formPrefix := "LOADBALANCERS_"
	// formValues := map[string]interface{}{
	// 	"EXTENSION_PREFIX":                  formPrefix,
	// 	formPrefix + "NAME":                 "nginx-test-lb-asdf1234",
	// 	formPrefix + "SERVICE":              "nginx-test-service-asdf",
	// 	formPrefix + "ACCESS_LOG_S3_BUCKET": "test-s3-logs-bucket",
	// 	formPrefix + "SSL_CERT_ARN":         "arn:1234:arnid",
	// 	formPrefix + "KUBECONFIG":           kubeconfig,
	// }

	extensionEvent := plugins.ProjectExtension{
		Slug:        "kubernetesloadbalancers",
		Environment: "testing",
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		// FormValues:  formValues,
		// Artifacts:   map[string]string{},
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
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig == "" {
		kubeconfig = path.Join(os.Getenv("HOME"), ".kube", "config")
	}
	project := plugins.Project{
		Repository: "checkr/nginx-test-success",
	}

	// formPrefix := "LOADBALANCERS_"
	// formValues := map[string]interface{}{
	// 	"EXTENSION_PREFIX":                  formPrefix,
	// 	formPrefix + "NAME":                 "nginx-test-lb-asdf1234",
	// 	formPrefix + "SSL_CERT_ARN":         "",
	// 	formPrefix + "ACCESS_LOG_S3_BUCKET": "",
	// 	formPrefix + "TYPE":                 t,
	// 	formPrefix + "SERVICE":              "nginx-test-service-asdf1234",
	// 	formPrefix + "KUBECONFIG":           kubeconfig,
	// 	formPrefix + "LISTENER_PAIRS": []k8s.ListenerPair{
	// 		{
	// 			Name:       "port1",
	// 			Protocol:   "TCP",
	// 			SourcePort: 443,
	// 			TargetPort: 3000,
	// 		},
	// 		{
	// 			Name:       "port2",
	// 			Protocol:   "TCP",
	// 			SourcePort: 444,
	// 			TargetPort: 3001,
	// 		},
	// 	},
	// }

	lbe := plugins.ProjectExtension{
		Slug:        "kubernetesloadbalancers",
		Action:      action,
		Environment: "testing",
		Project:     project,
	}
	return lbe
}

func BasicFailedReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()
	extension.Release.Services[0].Command = "/bin/false"

	event := transistor.NewEvent(extension, nil)
	AddBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func AddBasicReleaseExtensionArtifacts(extension plugins.ReleaseExtension, event *transistor.Event) {
	kubeConfigPath := os.Getenv("KUBECONFIG_PATH")

	event.AddArtifact("kubeconfig", kubeConfigPath, false)
	event.AddArtifact("client_certificate", "", false)
	event.AddArtifact("client_key", "", false)
	event.AddArtifact("certificate_authority", "", false)
	event.AddArtifact("client_key", "", false)

	event.AddArtifact("dockerbuilder_user", "test", false)
	event.AddArtifact("dockerbuilder_password", "test", false)
	event.AddArtifact("dockerbuilder_email", "test", false)
	event.AddArtifact("dockerbuilder_host", "test", false)
}

func BasicReleaseEvent() transistor.Event {
	extension := BasicReleaseExtension()

	event := transistor.NewEvent(extension, nil)
	AddBasicReleaseExtensionArtifacts(extension, &event)

	return event
}

func BasicReleaseExtension() plugins.ReleaseExtension {

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"
	artifacts := make(map[string]interface{})
	artifacts["IMAGE"] = "dev-registry.checkrhq.net/checkr/checkr-deploy-test:latest"

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
				Name:    "www",
				Command: "nginx -g 'daemon off';",
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
