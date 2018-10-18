package kubernetes

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

///////////////////////////////////////////////////
type K8sNamespacer interface {
	NewForConfig(*rest.Config) (K8sClienter, error)
}

type LegitimateKubernetesNamespacer struct {
}

func (l LegitimateKubernetesNamespacer) NewForConfig(config *rest.Config) (K8sClienter, error) {
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

type K8sContourNamespacer interface {
	NewForConfig(*rest.Config) (K8sClienter, error)
}

type LegitimateContourNamespacer struct {}

func (l LegitimateContourNamespacer) NewForConfig(config *rest.Config) (K8sClienter, error) {
	clientset, err := contour_client.NewForConfig(config)
	return clientset, err
}

///////////////////////////////////////////////////

type K8sContourer interface {}
type LegitimateContourClient struct {
	ContourClient *contour_client.Clientset
}



///////////////////////////////////////////////////

type K8sClientSet interface {}
type 