package kubernetes

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	uuid "github.com/satori/go.uuid"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func init() {
	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin {
		return &Kubernetes{}
	})
}

func (x *Kubernetes) Description() string {
	return "Kubernetes"

}

func (x *Kubernetes) SampleConfig() string {
	return ` `
}

func (x *Kubernetes) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Kubernetes (k8s)")

	return nil
}

func (x *Kubernetes) Stop() {
	log.Info("Stopping Kubernetes (k8s)")
}

func (x *Kubernetes) Subscribe() []string {
	return []string{
		"plugins.ReleaseExtension:create:kubernetesdeployments",

		"plugins.ProjectExtension:create:kubernetesdeployments",
		"plugins.ProjectExtension:update:kubernetesdeployments",

		"plugins.ProjectExtension:create:kubernetesloadbalancers",
		"plugins.ProjectExtension:update:kubernetesloadbalancers",
		"plugins.ProjectExtension:destroy:kubernetesloadbalancers",
	}
}

func (x *Kubernetes) Process(e transistor.Event) error {
	log.InfoWithFields("Processing kubernetes event", log.Fields{
		"event": e,
	})

	if e.Matches("kubernetesdeployments") == true {
		x.ProcessDeployment(e)
	} else if e.Matches("kubernetesloadbalancers") == true {
		x.ProcessLoadBalancer(e)
	}

	return nil
}

func buildEventPayload(e transistor.Event, state plugins.State, msg string) interface{} {
	switch e.PayloadModel {
	case "plugins.ReleaseExtension":
		payload := e.Payload.(plugins.ReleaseExtension)

		payload.Action = plugins.GetAction("status")
		payload.State = state
		payload.StateMessage = msg

		return payload
	case "plugins.ProjectExtension":
		payload := e.Payload.(plugins.ProjectExtension)

		payload.Action = plugins.GetAction("status")
		payload.State = state
		payload.StateMessage = msg

		return payload
	default:
		log.Fatal("unexpected type '%s'", e.PayloadModel)
	}

	return nil
}

func (x *Kubernetes) sendSuccessResponse(e transistor.Event, state plugins.State, artifacts []transistor.Artifact) {
	payload := buildEventPayload(e, state, "")

	event := e.NewEvent(payload, nil)
	event.Artifacts = artifacts

	x.events <- event
}

func (x *Kubernetes) sendErrorResponse(e transistor.Event, msg string) {
	payload := buildEventPayload(e, plugins.GetState("failed"), msg)
	x.events <- e.NewEvent(payload, nil)
}

func (x *Kubernetes) sendInProgress(e transistor.Event, msg string) {
	payload := buildEventPayload(e, plugins.GetState("running"), msg)
	x.events <- e.NewEvent(payload, nil)
}

func (x *Kubernetes) CreateProjectExtensionEvent(e transistor.Event, action plugins.Action, state plugins.State, msg string, err error) transistor.Event {
	payload := e.Payload.(plugins.ProjectExtension)
	payload.State = state
	payload.Action = action
	payload.StateMessage = msg

	return e.NewEvent(payload, err)
}

func (x *Kubernetes) GenNamespaceName(suggestedEnvironment string, projectSlug string) string {
	return fmt.Sprintf("%s-%s", suggestedEnvironment, projectSlug)
}

func (x *Kubernetes) GenDeploymentName(slugName string, serviceName string) string {
	return slugName + "-" + serviceName
}

func (x *Kubernetes) GenOneShotServiceName(slugName string, serviceName string) string {
	return "os-" + slugName + "-" + serviceName
}

func (x *Kubernetes) CreateNamespaceIfNotExists(namespace string, coreInterface corev1.CoreV1Interface) error {
	// Create namespace if it does not exist.
	_, nameGetErr := coreInterface.Namespaces().Get(namespace, meta_v1.GetOptions{})
	if nameGetErr != nil {
		if errors.IsNotFound(nameGetErr) {
			log.Warn("Namespace %s does not yet exist, creating.", namespace)
			namespaceParams := &v1.Namespace{
				TypeMeta: meta_v1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: meta_v1.ObjectMeta{
					Name: namespace,
				},
			}
			_, createNamespaceErr := coreInterface.Namespaces().Create(namespaceParams)
			if createNamespaceErr != nil {
				log.Error("Error '%s' creating namespace %s", createNamespaceErr, namespace)
				return createNamespaceErr
			}
			log.Debug("Namespace created: %s", namespace)
		} else {
			log.Error("Unhandled error occured looking up namespace %s: '%s'", namespace, nameGetErr)
			return nameGetErr
		}
	}
	return nil
}

func (x *Kubernetes) GetTempDir() (string, error) {
	for {
		filePath := fmt.Sprintf("/tmp/%s", uuid.NewV1().String())
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Warn("directory does not exist, creating.")
			// create the file
			err = os.MkdirAll(filePath, os.ModeDir|0700)
			if err != nil {
				log.Error(err.Error())
				return "", err
			}
			return filePath, nil
		}
	}
}

func (x *Kubernetes) SetupKubeConfig(e transistor.Event) (string, error) {
	randomDirectory, err := x.GetTempDir()
	if err != nil {
		log.Info(err.Error())
		return "", err
	}

	kubeconfig, err := e.GetArtifact("kubeconfig")
	if err != nil {
		return "", err
	}

	clientCert, err := e.GetArtifact("client_certificate")
	if err != nil {
		return "", err
	}

	clientKey, err := e.GetArtifact("client_key")
	if err != nil {
		return "", err
	}

	certificateAuthority, err := e.GetArtifact("certificate_authority")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/kubeconfig", randomDirectory), []byte(kubeconfig.String()), 0644)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	if err != nil {
		log.Error("ERROR: %s", err.Error())
		return "", err
	}
	log.Info("Using kubeconfig file: ", fmt.Sprintf("%s/kubeconfig", randomDirectory))

	// generate client cert, client key
	// certificate authority
	err = ioutil.WriteFile(fmt.Sprintf("%s/admin.pem", randomDirectory),
		[]byte(clientCert.String()), 0644)
	if err != nil {
		log.Error("ERROR: %s", err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/admin-key.pem", randomDirectory),
		[]byte(clientKey.String()), 0644)
	if err != nil {
		log.Error("ERROR: %s", err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/ca.pem", randomDirectory),
		[]byte(certificateAuthority.String()), 0644)
	if err != nil {
		log.Error("ERROR: %s", err.Error())
		return "", err
	}

	return fmt.Sprintf("%s/kubeconfig", randomDirectory), nil
}
