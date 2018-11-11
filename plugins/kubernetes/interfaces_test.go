package kubernetes_test

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/////////////////////////////////////////////////////////////////////////
type MockKubernetesNamespacer struct{}

func (l MockKubernetesNamespacer) NewForConfig(config *rest.Config) (kubernetes.Interface, error) {
	return kubefake.NewSimpleClientset(), nil
}

/////////////////////////////////////////////////////////////////////////
type MockContourNamespacer struct{}

func (l MockContourNamespacer) NewForConfig(config *rest.Config) (contour_client.Interface, error) {
	clientset, err := contour_client.NewForConfig(config)
	return clientset, err
}

type MockBatchV1Job struct {
	StatusOverride v1.JobStatus
}

func (l MockBatchV1Job) Get(clientset kubernetes.Interface, namespace string, jobName string, getOptions meta_v1.GetOptions) (*v1.Job, error) {
	job, err := clientset.BatchV1().Jobs(namespace).Get(jobName, getOptions)
	job.Status = l.StatusOverride

	return job, err
}

/////////////////////////////////////////////////////////////////////////

type MockCoreService struct{}

func (l MockCoreService) Get(clientset kubernetes.Interface, namespace string, serviceName string, getOptions meta_v1.GetOptions) (*corev1.Service, error) {
	service, err := clientset.Core().Services(namespace).Get(serviceName, getOptions)

	if service != nil {
		fakeIngressList := []corev1.LoadBalancerIngress{
			{
				IP:       "127.0.0.1",
				Hostname: "localhost",
			},
		}

		service.Status.LoadBalancer.Ingress = fakeIngressList
	}

	return service, err
}

func (l MockCoreService) Delete(clientset kubernetes.Interface, namespace string, serviceName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.Core().Services(namespace).Delete(serviceName, deleteOptions)
}
