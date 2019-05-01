package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/intstr"

	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
)

func (x *Kubernetes) ProcessRedis(e transistor.Event) {
	if e.Matches("project:kubernetes:redis") {
		var err error
		switch e.Action {
		case transistor.GetAction("delete"):
			err = x.doDeleteRedis(e)
		case transistor.GetAction("create"):
			err = x.doRedis(e)
		case transistor.GetAction("update"):
			redisEndpoint, err := e.GetArtifact("REDIS_ENDPOINT")
			if err != nil {
				log.Error(err)
				x.sendErrorResponse(e, err.Error())
			}

			// Send back event with endpoint artifact (in this case, <service-name>.<redis-deploys-namespace>)
			artifacts := []transistor.Artifact{
				{Key: "REDIS_ENDPOINT", Value: redisEndpoint.String()},
			}

			x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)
		}

		if err != nil {
			log.Error(err)
			x.sendErrorResponse(e, err.Error())
		}
	}
}

func (x *Kubernetes) doRedis(e transistor.Event) error {
	log.Debug("Received Redis Event")
	clientset, err := x.SetupClientset(e)
	if err != nil {
		log.Error("Error getting cluster config.  Aborting!")
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// get namespace of where redis deploys live
	redisDeploysNamespace, err := e.GetArtifact("redis_deployments_namespace")
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// Create namespace for redis deployments if it doesn't exist
	createNamespaceErr := x.createNamespaceIfNotExists(redisDeploysNamespace.String(), clientset)
	if createNamespaceErr != nil {
		x.sendErrorResponse(e, createNamespaceErr.Error())
		return createNamespaceErr
	}

	depInterface := clientset.AppsV1().Deployments(redisDeploysNamespace.String())
	svcInterface := clientset.Core().Services(redisDeploysNamespace.String())

	payload := e.Payload.(plugins.ProjectExtension)

	deploymentName := genRedisDeploymentName(payload)
	redisDeployment, err := x.createRedisDeploymentSpec(deploymentName, depInterface)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	_, err = depInterface.Create(redisDeployment)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// Create service name based on project slug, environment and unique id
	redisService, err := x.createRedisServiceSpec(payload, deploymentName, svcInterface)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	_, err = svcInterface.Create(redisService)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// Send back event with endpoint artifact (in this case, <service-name>.<redis-deploys-namespace>)
	artifacts := []transistor.Artifact{
		{Key: "REDIS_ENDPOINT", Value: fmt.Sprintf("%s.%s", deploymentName, redisDeploysNamespace.String()), Secret: false},
	}

	x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)

	return nil
}

func (x *Kubernetes) doDeleteRedis(e transistor.Event) error {
	err := deleteRedis(e, x)
	return err
}

func deleteRedis(e transistor.Event, x *Kubernetes) error {
	return nil
}

func (x *Kubernetes) createRedisDeploymentSpec(deploymentName string, depInterface appsv1.DeploymentInterface) (*v1.Deployment, error) {
	return &v1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: v1.DeploymentSpec{
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentName,
							Image: REDIS_IMAGE_TAG,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(REDIS_CONTAINER_PORT),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (x *Kubernetes) createRedisServiceSpec(payload plugins.ProjectExtension, deploymentName string, svcInterface corev1typed.ServiceInterface) (*corev1.Service, error) {
	serviceName := genRedisServiceName(payload)

	return &corev1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: int32(REDIS_CONTAINER_PORT),
					TargetPort: intstr.IntOrString{
						IntVal: int32(REDIS_CONTAINER_PORT),
					},
				},
			},
			Selector: map[string]string{
				"app":     deploymentName,
				"project": fmt.Sprintf("%s-%s", payload.Environment, payload.Project.Slug),
			},
		},
	}, nil
}

func genRedisDeploymentName(payload plugins.ProjectExtension) string {
	uniqueID := uuid.NewV4()
	return fmt.Sprintf("%s-%s-%s", payload.Environment[:10], payload.Project.Slug[:10], uniqueID.String()[:8])
}

func genRedisServiceName(payload plugins.ProjectExtension) string {
	uniqueID := uuid.NewV4()
	return fmt.Sprintf("%s-%s-%s", payload.Environment[:10], payload.Project.Slug[:10], uniqueID.String()[:8])
}
