package kubernetesutils

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

func CreateProjectExtensionEvent(e transistor.Event, action plugins.Action, state plugins.State, msg string, err error) transistor.Event {
	payload := e.Payload.(plugins.ProjectExtension)
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
			log.Info("Namespace %s does not yet exist, creating.", namespace)
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
				log.Info("Error '%s' creating namespace %s", createNamespaceErr, namespace)
				return createNamespaceErr
			}
			log.Info("Namespace created: %s", namespace)
		} else {
			log.Info("Unhandled error occured looking up namespace %s: '%s'", namespace, nameGetErr)
			return nameGetErr
		}
	}
	return nil
}

func GetTempDir() (string, error) {
	for {
		filePath := fmt.Sprintf("/tmp/%s", uuid.NewV1().String())
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Info("directory does not exist")
			// create the file
			err = os.MkdirAll(filePath, os.ModeDir)
			if err != nil {
				log.Info(err.Error())
				return "", err
			}
			return filePath, nil
		}
	}
}

func SetupKubeConfig(e transistor.Event) (string, error) {
	randomDirectory, err := GetTempDir()
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
		log.Info(err.Error())
		return "", err
	}

	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}
	log.Info("Using kubeconfig file: %s", fmt.Sprintf("%s/kubeconfig", randomDirectory))

	// generate client cert, client key
	// certificate authority
	err = ioutil.WriteFile(fmt.Sprintf("%s/admin.pem", randomDirectory),
		[]byte(clientCert.String()), 0644)
	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/admin-key.pem", randomDirectory),
		[]byte(clientKey.String()), 0644)
	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/ca.pem", randomDirectory),
		[]byte(certificateAuthority.String()), 0644)
	if err != nil {
		log.Info("ERROR: %s", err.Error())
		return "", err
	}

	return fmt.Sprintf("%s/kubeconfig", randomDirectory), nil
}
