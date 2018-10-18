package kubernetes_test

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

/////////////////////////////////////////////////////////////////////////
type MockKubernetesNamespacer struct{}

func (l MockKubernetesNamespacer) NewForConfig(config *rest.Config) (kubernetes.Interface, error) {
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

/////////////////////////////////////////////////////////////////////////
type MockContourNamespacer struct{}

func (l MockContourNamespacer) NewForConfig(config *rest.Config) (contour_client.Interface, error) {
	clientset, err := contour_client.NewForConfig(config)
	return clientset, err
}
