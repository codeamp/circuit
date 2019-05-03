package kubernetes

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	Create(kubernetes.Interface, string, *v1.Job) (*v1.Job, error)
	Get(kubernetes.Interface, string, string, meta_v1.GetOptions) (*v1.Job, error)
}

type BatchV1Job struct{}

func (l BatchV1Job) Get(clientset kubernetes.Interface, namespace string, jobName string, getOptions meta_v1.GetOptions) (*v1.Job, error) {
	return clientset.BatchV1().Jobs(namespace).Get(jobName, getOptions)
}

func (l BatchV1Job) Create(clientset kubernetes.Interface, namespace string, job *v1.Job) (*v1.Job, error) {
	return clientset.BatchV1().Jobs(namespace).Create(job)
}

///////////////////////////////////////////////////

type CoreServicer interface {
	Get(kubernetes.Interface, string, string, meta_v1.GetOptions) (*corev1.Service, error)
	Delete(kubernetes.Interface, string, string, *meta_v1.DeleteOptions) error

	Create(kubernetes.Interface, string, *corev1.Service) (*corev1.Service, error)
	Update(kubernetes.Interface, string, *corev1.Service) (*corev1.Service, error)
}

type CoreService struct{}

func (l CoreService) Get(clientset kubernetes.Interface, namespace string, serviceName string, getOptions meta_v1.GetOptions) (*corev1.Service, error) {
	return clientset.Core().Services(namespace).Get(serviceName, getOptions)
}

func (l CoreService) Delete(clientset kubernetes.Interface, namespace string, serviceName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.Core().Services(namespace).Delete(serviceName, deleteOptions)
}

func (l CoreService) Create(clientset kubernetes.Interface, namespace string, service *corev1.Service) (*corev1.Service, error) {
	return clientset.Core().Services(namespace).Create(service)
}

func (l CoreService) Update(clientset kubernetes.Interface, namespace string, service *corev1.Service) (*corev1.Service, error) {
	return clientset.Core().Services(namespace).Update(service)
}

///////////////////////////////////////////////////

type CoreDeploymenter interface {
	Create(kubernetes.Interface, string, *appsv1.Deployment) (*appsv1.Deployment, error)
}

type CoreDeployment struct{}

func (l CoreDeployment) Create(clientset kubernetes.Interface, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return clientset.AppsV1().Deployments(namespace).Create(deployment)
}

func (l CoreDeployment) Delete(clientset kubernetes.Interface, namespace string, deploymentName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.AppsV1().Deployments(namespace).Delete(deploymentName, deleteOptions)
}

///////////////////////////////////////////////////

type CoreSecreter interface {
	Create(kubernetes.Interface, string, *corev1.Secret) (*corev1.Secret, error)
}

type CoreSecret struct{}

func (l CoreSecret) Create(clientset kubernetes.Interface, namespace string, secretParams *corev1.Secret) (*corev1.Secret, error) {
	return clientset.Core().Secrets(namespace).Create(secretParams)
}
