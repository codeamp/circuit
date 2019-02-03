package kubernetes_test

import (
	"fmt"

	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	uuid "github.com/satori/go.uuid"
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

func (l MockCoreService) Create(clientset kubernetes.Interface, namespace string, service *corev1.Service) (*corev1.Service, error) {
	return clientset.Core().Services(namespace).Create(service)
}

func (l MockCoreService) Update(clientset kubernetes.Interface, namespace string, service *corev1.Service) (*corev1.Service, error) {
	return clientset.Core().Services(namespace).Update(service)
}

/////////////////////////////////////////////////////////////////////////

type MockCoreSecret struct{}

func (l MockCoreSecret) Create(clientset kubernetes.Interface, namespace string, secretParams *corev1.Secret) (*corev1.Secret, error) {
	var secretsCopy corev1.Secret

	if secretParams != nil {
		secretsCopy = *secretParams

		genSuffix := uuid.NewV4()

		if secretsCopy.GenerateName != "" {
			secretsCopy.Name = fmt.Sprintf("%s-%s", secretsCopy.GenerateName, genSuffix)
		}
	}

	return clientset.Core().Secrets(namespace).Create(&secretsCopy)
}

/////////////////////////////////////////////////////////////////////////

type MockCoreDeployment struct {}

///////////////////////////////////////////////////

type MockExtDeployment struct {}

func (l MockExtDeployment) Get(clientset kubernetes.Interface, namespace string, deploymentName string, getOptions meta_v1.GetOptions) (*v1beta1.Deployment, error) {
	return clientset.Extensions().Deployments(namespace).Get(deploymentName, getOptions)
}

func (l MockExtDeployment) Delete(clientset kubernetes.Interface, namespace string, deploymentName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.Extensions().Deployments(namespace).Delete(deploymentName, deleteOptions)
}

func (l MockExtDeployment) Create(clientset kubernetes.Interface, namespace string, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return clientset.Extensions().Deployments(namespace).Create(deployment)
}

func (l MockExtDeployment) List(clientset kubernetes.Interface, namespace string, listOptions *meta_v1.ListOptions) (*v1beta1.DeploymentList, error) {
	return clientset.Extensions().Deployments(namespace).List(*listOptions)
}

func (l MockExtDeployment) Update(clientset kubernetes.Interface, namespace string, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return clientset.Extensions().Deployments(namespace).Update(deployment)
}

func (l MockExtDeployment) UpdateScale(clientset kubernetes.Interface, namespace string, deploymentName string, scale *v1beta1.Scale) (*v1beta1.Scale, error) {
	return clientset.Extensions().Deployments(namespace).UpdateScale(deploymentName, scale)
}

///////////////////////////////////////////////////


///////////////////////////////////////////////////

type MockCorePod struct {}

func (l MockCorePod) List(clientset kubernetes.Interface, namespace string, listOptions *meta_v1.ListOptions) (*corev1.PodList, error) {
	return clientset.Core().Pods(namespace).List(*listOptions)
}

func (l MockCorePod) Delete(clientset kubernetes.Interface, namespace string, podName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.Core().Pods(namespace).Delete(podName, deleteOptions)
}