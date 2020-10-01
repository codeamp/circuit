package kubernetes_test

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	appsv1 "k8s.io/api/apps/v1"
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

type MockBatchV1Job struct {
	StatusOverride v1.JobStatus
}

func (l MockBatchV1Job) Get(clientset kubernetes.Interface, namespace string, jobName string, getOptions meta_v1.GetOptions) (*v1.Job, error) {
	job, err := clientset.BatchV1().Jobs(namespace).Get(jobName, getOptions)
	job.Status = l.StatusOverride

	return job, err
}

func (l MockBatchV1Job) Create(clientset kubernetes.Interface, namespace string, job *v1.Job) (*v1.Job, error) {
	job.ObjectMeta.Name = job.ObjectMeta.GenerateName
	return clientset.BatchV1().Jobs(namespace).Create(job)
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

type MockCoreDeployment struct{}

func (l MockCoreDeployment) Delete(clientset kubernetes.Interface, namespace string, deploymentName string, deleteOptions *meta_v1.DeleteOptions) error {
	return clientset.AppsV1().Deployments(namespace).Delete(deploymentName, deleteOptions)
}

func (l MockCoreDeployment) Create(clientset kubernetes.Interface, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return clientset.AppsV1().Deployments(namespace).Create(deployment)
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
