package kubernetes

import (
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
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

type CoreSecreter interface {
	Create(kubernetes.Interface, string, *corev1.Secret) (*corev1.Secret, error)
}

type CoreSecret struct{}

func (l CoreSecret) Create(clientset kubernetes.Interface, namespace string, secretParams *corev1.Secret) (*corev1.Secret, error) {
	return clientset.Core().Secrets(namespace).Create(secretParams)
}

///////////////////////////////////////////////////

type ExtDeploymenter interface {
	Create(kubernetes.Interface, string, *v1beta1.Deployment) (*v1beta1.Deployment, error)
	Update(kubernetes.Interface, string, *v1beta1.Deployment) (*v1beta1.Deployment, error)
	List(kubernetes.Interface, string, *meta_v1.ListOptions) (*v1beta1.DeploymentList, error)
	Get(kubernetes.Interface, string, string, *meta_v1.GetOptions) (*v1beta1.Deployment, error)
	Delete(kubernetes.Interface, string, string, *meta_v1.DeleteOptions) error
	UpdateScale(kubernetes.Interface, string, string, *v1beta1.Scale) (*v1beta1.Scale, error)
}

type ExtDeployment struct{}

func (l ExtDeployment) Get(clientset kubernetes.Interface, namespace string, deploymentName string, getOptions *meta_v1.GetOptions) (*v1beta1.Deployment, error) {
	return clientset.Extensions().Deployments(namespace).Get(deploymentName, *getOptions)
}

func (l ExtDeployment) Delete(clientset kubernetes.Interface, namespace string, deploymentName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.Extensions().Deployments(namespace).Delete(deploymentName, deleteOptions)
}

func (l ExtDeployment) Create(clientset kubernetes.Interface, namespace string, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return clientset.Extensions().Deployments(namespace).Create(deployment)
}

func (l ExtDeployment) List(clientset kubernetes.Interface, namespace string, listOptions *meta_v1.ListOptions) (*v1beta1.DeploymentList, error) {
	return clientset.Extensions().Deployments(namespace).List(*listOptions)
}

func (l ExtDeployment) Update(clientset kubernetes.Interface, namespace string, deployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return clientset.Extensions().Deployments(namespace).Update(deployment)
}

func (l ExtDeployment) UpdateScale(clientset kubernetes.Interface, namespace string, deploymentName string, scale *v1beta1.Scale) (*v1beta1.Scale, error) {
	return clientset.Extensions().Deployments(namespace).UpdateScale(deploymentName, scale)
}

///////////////////////////////////////////////////

type ExtReplicaSetter interface {
	List(kubernetes.Interface, string, *meta_v1.ListOptions) (*v1beta1.ReplicaSetList, error)
	Delete(kubernetes.Interface, string, string, *meta_v1.DeleteOptions) error
	UpdateScale(kubernetes.Interface, string, string, *v1beta1.Scale) (*v1beta1.Scale, error)
}

type ExtReplicaSet struct{}

func (l ExtReplicaSet) List(clientset kubernetes.Interface, namespace string, listOptions *meta_v1.ListOptions) (*v1beta1.ReplicaSetList, error) {
	return clientset.Extensions().ReplicaSets(namespace).List(*listOptions)
}

func (l ExtReplicaSet) Delete(clientset kubernetes.Interface, namespace string, replicaSetName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.Extensions().ReplicaSets(namespace).Delete(replicaSetName, deleteOptions)
}

func (l ExtReplicaSet) UpdateScale(clientset kubernetes.Interface, namespace string, replicaSetName string, scale *v1beta1.Scale) (*v1beta1.Scale, error) {
	return clientset.Extensions().ReplicaSets(namespace).UpdateScale(replicaSetName, scale)
}
