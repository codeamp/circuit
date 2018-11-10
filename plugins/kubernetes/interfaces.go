package kubernetes

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/batch/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

///////////////////////////////////////////////////
type K8sNamespacer interface {
	NewForConfig(*rest.Config) (kubernetes.Interface, error)
}

type KubernetesNamespace struct {
}

func (l KubernetesNamespace) NewForConfig(config *rest.Config) (kubernetes.Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

///////////////////////////////////////////////////

type K8sContourNamespacer interface {
	NewForConfig(*rest.Config) (contour_client.Interface, error)
}

type ContourNamespace struct{}

func (l ContourNamespace) NewForConfig(config *rest.Config) (contour_client.Interface, error) {
	clientset, err := contour_client.NewForConfig(config)
	return clientset, err
}

///////////////////////////////////////////////////

type BatchV1Jobber interface {
	Get(kubernetes.Interface, string, string, meta_v1.GetOptions) (*v1.Job, error)
}

type BatchV1Job struct {}

func (l BatchV1Job) Get(clientset kubernetes.Interface, namespace string, jobName string, getOptions meta_v1.GetOptions) (*v1.Job, error){
	return clientset.BatchV1().Jobs(namespace).Get(jobName, getOptions)
}