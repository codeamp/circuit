package kubernetes

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

///////////////////////////////////////////////////
type Kuberneteser interface {
	NewForConfig(*rest.Config) (K8sClienter, error)
}
type LegitimateKubernetes struct {
}

func (l LegitimateKubernetes) NewForConfig(config *rest.Config) (K8sClienter, error) {
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

///////////////////////////////////////////////////

type ClientCommander interface{}
type LegitimateClientCmd struct{}

///////////////////////////////////////////////////

type K8sClienter interface{}
type LegitimateKubernetesClient struct {
	KubernetesClient *kubernetes.Clientset
}

///////////////////////////////////////////////////

type K8sCountourer interface {
	NewForConfig(*rest.Config) (K8sClienter, error)
}

type LegitimateK8sCountourClient struct {
	ContourClient *contour_client.Clientset
}

func (l LegitimateKubernetes) NewForConfig(config *rest.Config) (K8sClienter, error) {
	clientset, err := contour_client.NewForConfig(config)
	return clientset, err
}

///////////////////////////////////////////////////
