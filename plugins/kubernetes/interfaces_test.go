package kubernetes_test


/////////////////////////////////////////////////////////////////////////
type MockKubernetesClient struct {}

func (l LegitimateKubernetes) NewForConfig(config *rest.Config) (K8sClienter, error) {
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}


/////////////////////////////////////////////////////////////////////////
type MockContourNamespace struct {}

func (l LegitimateKubernetes) NewForConfig(config *rest.Config) (K8sClienter, error) {
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

/////////////////////////////////////////////////////////////////////////
type MockContourClient struct {}