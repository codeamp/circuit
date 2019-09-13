package kubernetes

import (
	"fmt"

	apis_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	uuid "github.com/satori/go.uuid"

	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	payload := e.Payload.(plugins.ProjectExtension)

	deploymentName, projectName := genRedisDeploymentName(payload)
	redisDeployment, err := x.createRedisDeploymentSpec(deploymentName, projectName)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	_, err = x.CoreDeploymenter.Create(clientset, redisDeploysNamespace.String(), redisDeployment)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// Create service name based on project slug, environment and unique id
	redisService, err := x.createRedisServiceSpec(payload, deploymentName)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	_, err = x.CoreServicer.Create(clientset, redisDeploysNamespace.String(), redisService)
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
	// Get name from REDIS_ENDPOINT
	redisEndpoint, err := e.GetArtifact("REDIS_ENDPOINT")
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	splitRedisEndpoint := strings.Split(redisEndpoint.String(), ".")
	if len(splitRedisEndpoint) > 0 {
		deploymentName := splitRedisEndpoint[0]
		namespace := splitRedisEndpoint[1]

		clientset, err := x.SetupClientset(e)
		if err != nil {
			log.Error("Error getting cluster config.  Aborting!")
			x.sendErrorResponse(e, err.Error())
			return err
		}

		depInterface := clientset.AppsV1().Deployments(namespace)
		svcInterface := clientset.Core().Services(namespace)

		depInterface.Delete(deploymentName, &apis_meta_v1.DeleteOptions{})
		svcInterface.Delete(deploymentName, &apis_meta_v1.DeleteOptions{})

		x.sendSuccessResponse(e, transistor.GetState("complete"), nil)
	} else {
		return fmt.Errorf("Invalid redis endpoint %s", redisEndpoint.String())
	}

	return nil
}

func (x *Kubernetes) createRedisDeploymentSpec(deploymentName string, projectName string) (*v1.Deployment, error) {
	return &v1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: deploymentName,
			Labels: map[string]string{
				"app":         deploymentName,
				"projectName": projectName,
			},
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
						"app":         deploymentName,
						"projectName": projectName,
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

func (x *Kubernetes) createRedisServiceSpec(payload plugins.ProjectExtension, svcName string) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: svcName,
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
				"app": svcName,
			},
		},
	}, nil
}

func genRedisDeploymentName(payload plugins.ProjectExtension) (string, string) {
	uniqueID := uuid.NewV4()
	env := payload.Environment
	projectSlug := payload.Project.Slug

	if len(payload.Environment) > 10 {
		env = payload.Environment[:10]
	}

	projectSlugLength := len(projectSlug)
	if projectSlugLength > 10 {
		projectSlugLength = 10
	}

	return fmt.Sprintf("%s-%s-%s", env, projectSlug[:projectSlugLength], uniqueID.String()[:8]), projectSlug
}

func genRedisServiceName(payload plugins.ProjectExtension) string {
	uniqueID := uuid.NewV4()
	env := payload.Environment
	projectSlug := payload.Project.Slug

	if len(payload.Environment) > 10 {
		env = payload.Environment[:10]
	}

	if len(payload.Project.Slug) > 10 {
		projectSlug = payload.Project.Slug[:10]
	}

	return fmt.Sprintf("%s-%s-%s", env, projectSlug, uniqueID.String()[:8])
}
