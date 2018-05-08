package testdata

import (
	"os"
	"path"

	"github.com/codeamp/circuit/plugins"
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
		// FormValues:  formValues,
		// Artifacts:   map[string]string{},
	}
	return lbe
}

func BasicFailedReleaseExtension() plugins.ReleaseExtension {
	d := BasicReleaseExtension()
	d.Release.Services[0].Command = "/bin/false"
	return d
}

func BasicReleaseExtension() plugins.ReleaseExtension {
	formValues := make(map[string]interface{})
	var kubeconfig string
	// If this is not set the test will use inClusterConfig
	kubeconfig = os.Getenv("KUBECONFIG")
	formValues["KUBECONFIG"] = kubeconfig
	formValues["DOCKERBUILDER_PASSWORD"] = "test"
	formValues["DOCKERBUILDER_USER"] = "test"
	formValues["DOCKERBUILDER_EMAIL"] = "test"
	formValues["DOCKERBUILDER_HOST"] = "test"

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"
	artifacts := make(map[string]interface{})
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
		// Artifacts:   artifacts,
	}

	releaseExtensionEvent := plugins.ReleaseExtension{
		Slug:    "kubernetesdeployments",
		Action:  plugins.GetAction("create"),
		State:   plugins.GetState("waiting"),
		Release: releaseEvent,
		// Artifacts: artifacts,
	}

	return releaseExtensionEvent
}
