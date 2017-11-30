package kubernetesutils

import (
	"fmt"
	"log"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func GetFormValuePrefix(e transistor.Event, fallbackPrefix string) string {
	formValues := e.Payload.(plugins.Extension).FormValues

	prefix := formValues["EXTENSION_PREFIX"].(string)
	if prefix == "" {
		return fallbackPrefix
	}
	return prefix
}

func CreateExtensionEvent(e transistor.Event, state plugins.State, action plugins.Action, msg string, err error) transistor.Event {
	payload := e.Payload.(plugins.Extension)
	payload.State = state
	payload.Action = action
	payload.StateMessage = msg

	return e.NewEvent(payload, err)
}

func GenNamespaceName(suggestedEnvironment string, projectSlug string) string {
	return fmt.Sprintf("%s-%s", suggestedEnvironment, projectSlug)
}

func GenDeploymentName(slugName string, serviceName string) string {
	return slugName + "-" + serviceName
}

func GenOneShotServiceName(slugName string, serviceName string) string {
	return "os-" + slugName + "-" + serviceName
}

func CreateNamespaceIfNotExists(namespace string, coreInterface corev1.CoreV1Interface) error {
	// Create namespace if it does not exist.
	_, nameGetErr := coreInterface.Namespaces().Get(namespace, meta_v1.GetOptions{})
	if nameGetErr != nil {
		if errors.IsNotFound(nameGetErr) {
			log.Printf("Namespace %s does not yet exist, creating.", namespace)
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
				log.Printf("Error '%s' creating namespace %s", createNamespaceErr, namespace)
				return createNamespaceErr
			}
			log.Printf("Namespace created: %s", namespace)
		} else {
			log.Printf("Unhandled error occured looking up namespace %s: '%s'", namespace, nameGetErr)
			return nameGetErr
		}
	}
	return nil
}
