package testdata

import (
	"os"
	"path"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/kubernetes/loadbalancers"
)

func GetCreateExtension() plugins.Extension {
	d := GetBasicExtension()
	d.Action = plugins.Create
	d.State = plugins.Waiting
	return d
}

func GetDestroyExtension() plugins.Extension {
	d := GetBasicExtension()
	d.Action = plugins.Destroy
	d.State = plugins.Waiting
	return d
}

func GetBasicExtension() plugins.Extension {
	var kubeconfig string
	if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig == "" {
		kubeconfig = path.Join(os.Getenv("HOME"), ".kube", "config")
	}

	formValues := map[string]interface{}{
		"SERVICE_NAME":         "test-service-name",
		"ACCESS_LOG_S3_BUCKET": "test-s3-logs-bucket",
		"SSL_CERT_ARN":         "arn:1234:arnid",
		"KUBECONFIG":           kubeconfig,
	}

	extensionEvent := plugins.Extension{
		Slug:        "kubernetesloadbalancers",
		Environment: "testing",
		Action:      plugins.Create,
		State:       plugins.Waiting,
		FormValues:  formValues,
		Artifacts:   map[string]string{},
		Project: plugins.Project{
			Repository: "checkr/deploy-test",
			Services: []plugins.Service{
				plugins.Service{},
			},
		},
	}

	return extensionEvent
}

func LBDataForTCP(action plugins.Action, t plugins.Type) plugins.Extension {
	project := plugins.Project{
		Repository: "checkr/nginx-test-success",
	}

	formPrefix := "LOADBALANCERS_"
	formValues := map[string]interface{}{
		"EXTENSION_PREFIX":                  formPrefix,
		formPrefix + "NAME":                 "nginx-test-lb-asdf1234",
		formPrefix + "SSL_CERT_ARN":         "arn",
		formPrefix + "ACCESS_LOG_S3_BUCKET": "bucket",
		formPrefix + "TYPE":                 t,
		formPrefix + "SERVICE":              "nginx-test-service-asdf",
		formPrefix + "LISTENER_PAIRS": []kubernetesloadbalancers.ListenerPair{
			{
				Name:       "port1",
				Protocol:   "TCP",
				SourcePort: 443,
				TargetPort: 3000,
			},
			{
				Name:       "port2",
				Protocol:   "TCP",
				SourcePort: 444,
				TargetPort: 3001,
			},
		},
	}

	lbe := plugins.Extension{
		Action:      action,
		Environment: "testing",
		Project:     project,
		FormValues:  formValues,
	}
	return lbe
}
