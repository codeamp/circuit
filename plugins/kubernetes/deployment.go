package kubernetes

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	apis_batch_v1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/google/shlex"
	"github.com/spf13/viper"
)

func (x *Kubernetes) ProcessDeployment(e transistor.Event) {
	log.InfoWithFields("Processing Kubernetes Deployments event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.ProjectExtension:create:kubernetesdeployments" {
		var extensionEvent plugins.ProjectExtension
		extensionEvent = e.Payload.(plugins.ProjectExtension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")

		x.events <- e.NewEvent(extensionEvent, nil)
		return
	}

	if e.Name == "plugins.ProjectExtension:update:kubernetesdeployments" {
		var extensionEvent plugins.ProjectExtension
		extensionEvent = e.Payload.(plugins.ProjectExtension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")

		x.events <- e.NewEvent(extensionEvent, nil)
		return
	}

	if e.Name == "plugins.ReleaseExtension:create:kubernetesdeployments" {
		err := x.doDeploy(e)
		if err != nil {
			log.Error(err)
		}

		return
	}
}

func int32Ptr(i int32) *int32 { return &i }

func genDeploymentName(slugName string, serviceName string) string {
	return slugName + "-" + serviceName
}

func genOneShotServiceName(slugName string, serviceName string) string {
	return "os-" + slugName + "-" + serviceName
}

func secretifyDockerCred(e transistor.Event) (string, error) {
	user, err := e.GetArtifactFromSource("user", "dockerbuilder")
	log.Warn(user, " ", err)
	if err != nil {
		return "", err
	}

	pass, err := e.GetArtifactFromSource("password", "dockerbuilder")
	if err != nil {
		return "", err
	}

	email, err := e.GetArtifactFromSource("email", "dockerbuilder")
	if err != nil {
		return "", err
	}

	host, err := e.GetArtifactFromSource("host", "dockerbuilder")
	if err != nil {
		return "", err
	}

	encodeMe := fmt.Sprintf("%s:%s", user.String(), pass.String())
	encodeResult := []byte(encodeMe)
	authField := base64.StdEncoding.EncodeToString(encodeResult)
	jsonFilled := fmt.Sprintf("{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"email\":\"%s\",\"auth\":\"%s\"}}",
		host.String(),
		user.String(),
		pass.String(),
		email.String(),
		authField,
	)
	return jsonFilled, nil
}

func (x *Kubernetes) createDockerIOSecretIfNotExists(namespace string, coreInterface corev1.CoreV1Interface, e transistor.Event) error {
	// Load up the docker-io secrets for image pull if not exists
	_, dockerIOSecretErr := coreInterface.Secrets(namespace).Get("docker-io", meta_v1.GetOptions{})
	if dockerIOSecretErr != nil {
		if errors.IsNotFound(dockerIOSecretErr) {
			dockerCred, err := secretifyDockerCred(e)
			if err != nil {
				log.Error(fmt.Sprintf("Error '%s' creating docker-io secret for %s.", err, namespace))
				return err
			}
			secretMap := map[string]string{
				".dockercfg": dockerCred,
			}
			_, createDockerIOSecretErr := coreInterface.Secrets(namespace).Create(&v1.Secret{
				TypeMeta: meta_v1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      "docker-io",
					Namespace: namespace,
				},
				StringData: secretMap,
				Type:       v1.SecretTypeDockercfg,
			})
			if createDockerIOSecretErr != nil {
				log.Error(fmt.Sprintf("Error '%s' creating docker-io secret for %s.", createDockerIOSecretErr, namespace))
				return createDockerIOSecretErr
			}
		} else {
			log.Error(fmt.Sprintf("Error unhandled '%s' while attempting to lookup docker-io secret.", dockerIOSecretErr))
			return dockerIOSecretErr
		}
	}

	return nil
}

func (x *Kubernetes) createNamespaceIfNotExists(namespace string, coreInterface corev1.CoreV1Interface) error {
	// Create namespace if it does not exist.
	_, nameGetErr := coreInterface.Namespaces().Get(namespace, meta_v1.GetOptions{})
	if nameGetErr != nil {
		if errors.IsNotFound(nameGetErr) {
			log.Debug(fmt.Sprintf("Namespace %s does not yet exist, creating.", namespace))
			namespaceParams := &v1.Namespace{
				TypeMeta: meta_v1.TypeMeta{
					Kind:       "Namespace",
					APIVersion: "v1",
				},
				ObjectMeta: meta_v1.ObjectMeta{
					Name: namespace,
				},
			}
			_, createNamespaceErr := coreInterface.Namespaces().Create(namespaceParams)
			if createNamespaceErr != nil {
				log.Error(fmt.Sprintf("Error '%s' creating namespace %s", createNamespaceErr, namespace))
				return createNamespaceErr
			}
			log.Debug(fmt.Sprintf("Namespace created: %s", namespace))
		} else {
			log.Error(fmt.Sprintf("Unhandled error occured looking up namespace %s: '%s'", namespace, nameGetErr))
			return nameGetErr
		}
	}
	return nil
}

// Returns false if there is no failures detected and true if there is an error waiting
func detectPodFailure(pod v1.Pod) (string, bool) {
	if len(pod.Status.ContainerStatuses) > 0 {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				switch waitingReason := containerStatus.State.Waiting.Reason; waitingReason {
				case "CrashLoopBackOff", "ImageInspectError", "ErrImageNeverPull", "RegistryUnavilable", "InvalidImageName":
					failmessage := fmt.Sprintf("Detected Pod '%s' is waiting forever because of '%s'", pod.Name, waitingReason)
					// Pod is waiting forever
					return failmessage, true
				default:
					return fmt.Sprintf("Pod '%s' is waiting because '%s'", pod.Name, waitingReason), false
				}
			}
		}
	}
	return "", false
}

// Returns an array of services with all Waiting statuses re-set to Failed
func setFailServices(deploymentServices []plugins.Service) []plugins.Service {
	var deploymentServicesFailed []plugins.Service
	for index := range deploymentServices {
		if deploymentServices[index].State == plugins.GetState("waiting") {
			deploymentServices[index].State = plugins.GetState("failed")
			deploymentServices[index].StateMessage = "Failed from waiting too long"
		}
	}
	return deploymentServicesFailed
}

func getContainerPorts(service plugins.Service) []v1.ContainerPort {
	var deployPorts []v1.ContainerPort

	// ContainerPorts for the deployment
	for _, cPort := range service.Listeners {
		// Build the deployments containerports array
		newContainerPort := v1.ContainerPort{
			//Name:          //fmt.Sprintf("%d-%s", cPort.Port, strings.ToLower(cPort.Protocol)),
			ContainerPort: cPort.Port,
			Protocol:      v1.Protocol(cPort.Protocol),
		}
		deployPorts = append(deployPorts, newContainerPort)
	}

	return deployPorts
}

func genPodTemplateSpec(e transistor.Event, podConfig SimplePodSpec, kind string) v1.PodTemplateSpec {
	releaseExtension := e.Payload.(plugins.ReleaseExtension)
	container := v1.Container{
		Name:  strings.ToLower(podConfig.Service.Name),
		Image: podConfig.Image,
		Ports: podConfig.DeployPorts,
		Args:  podConfig.Args,
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(podConfig.Service.Spec.CpuLimit),
				v1.ResourceMemory: resource.MustParse(podConfig.Service.Spec.MemoryLimit),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(podConfig.Service.Spec.CpuRequest),
				v1.ResourceMemory: resource.MustParse(podConfig.Service.Spec.MemoryRequest),
			},
		},
		ImagePullPolicy: v1.PullAlways,
		Env:             podConfig.Env,
		VolumeMounts:    podConfig.VolumeMounts,
	}
	if kind == "Deployment" {
		container.ReadinessProbe = &podConfig.ReadyProbe
		container.LivenessProbe = &podConfig.LiveProbe
	}
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: podConfig.Name,
			Labels: map[string]string{
				"app":                podConfig.Name,
				"environment":        releaseExtension.Release.Environment,
				"releaseFeatureHash": releaseExtension.Release.HeadFeature.Hash,
				"releaseID":          releaseExtension.Release.ID,
			},
		},
		Spec: v1.PodSpec{
			NodeSelector:                  podConfig.NodeSelector,
			TerminationGracePeriodSeconds: &podConfig.Service.Spec.TerminationGracePeriodSeconds,
			ImagePullSecrets: []v1.LocalObjectReference{
				{
					Name: "docker-io",
				},
			},
			Containers: []v1.Container{
				container,
			},
			Volumes:       podConfig.Volumes,
			RestartPolicy: podConfig.RestartPolicy,
			DNSPolicy:     v1.DNSClusterFirst,
		},
	}
	return podTemplateSpec
}

func (x *Kubernetes) doDeploy(e transistor.Event) error {
	// write kubeconfig
	reData := e.Payload.(plugins.ReleaseExtension)
	projectSlug := plugins.GetSlug(reData.Release.Project.Repository)

	kubeconfig, err := x.SetupKubeConfig(e)
	if err != nil {
		log.Error(err.Error())
		x.sendErrorResponse(e, "failed writing kubeconfig")
		return err
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{Timeout: "60"}).ClientConfig()
	if err != nil {
		log.Error(fmt.Sprintf("ERROR '%s' while building kubernetes api client config.  Falling back to inClusterConfig.", err))
		config, err = clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			log.Error(fmt.Sprintf("ERROR '%s' while attempting inClusterConfig fallback. Aborting!", err))
			x.sendErrorResponse(e, "failed writing kubeconfig")
			return err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Error getting cluster config.  Aborting!")
		x.sendErrorResponse(e, err.Error())
		return err
	}

	x.sendInProgress(e, "Deploy in-progress")
	namespace := x.GenNamespaceName(reData.Release.Environment, projectSlug)
	coreInterface := clientset.Core()

	successfulDeploys := 0
	// TODO: get timeout from formValues
	//timeout := e.Payload.(plugins.ReleaseExtension).Release.Timeout
	// Set default timeout to 600 seconds if not specified.
	//if timeout == 0 {
	timeout := 600
	//}
	curTime := 0

	createNamespaceErr := x.createNamespaceIfNotExists(namespace, coreInterface)
	if createNamespaceErr != nil {
		x.sendErrorResponse(e, createNamespaceErr.Error())
		return createNamespaceErr
	}

	createDockerIOSecretErr := x.createDockerIOSecretIfNotExists(namespace, coreInterface, e)
	if createDockerIOSecretErr != nil {
		x.sendErrorResponse(e, createDockerIOSecretErr.Error())
		return createDockerIOSecretErr
	}
	// Create secrets for this deploy
	var secretMap map[string]string
	secretMap = make(map[string]string)
	var myEnvVars []v1.EnvVar

	// This map is used in to create the secrets themselves
	for _, secret := range reData.Release.Secrets {
		secretMap[secret.Key] = secret.Value
	}

	secretParams := &v1.Secret{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			GenerateName: fmt.Sprintf("%v-", projectSlug),
			Namespace:    namespace,
		},
		StringData: secretMap,
		Type:       v1.SecretTypeOpaque,
	}

	secretResult, secErr := coreInterface.Secrets(namespace).Create(secretParams)
	if secErr != nil {
		failMessage := fmt.Sprintf("Error '%s' creating secret %s", secErr, projectSlug)
		x.sendErrorResponse(e, failMessage)
		return fmt.Errorf(failMessage)
	}
	secretName := secretResult.Name
	x.sendInProgress(e, "Secrets created")

	// This is for building the configuration to use the secrets from inside the deployment
	// as ENVs
	for _, secret := range reData.Release.Secrets {
		if secret.Type == plugins.GetType("env") || secret.Type == plugins.GetType("protected-env") {
			newEnv := v1.EnvVar{
				Name: secret.Key,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Key: secret.Key,
					},
				},
			}
			myEnvVars = append(myEnvVars, newEnv)
		}
	}

	// as Files
	var volumeMounts []v1.VolumeMount
	var deployVolumes []v1.Volume
	var volumeSecretItems []v1.KeyToPath
	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      secretName,
		MountPath: "/etc/secrets",
		ReadOnly:  true,
	})

	for _, secret := range reData.Release.Secrets {
		if secret.Type == plugins.GetType("file") {
			volumeSecretItems = append(volumeSecretItems, v1.KeyToPath{
				Path: secret.Key,
				Key:  secret.Key,
				Mode: int32Ptr(256),
			})
		}
	}
	secretVolume := v1.SecretVolumeSource{
		SecretName:  secretName,
		Items:       volumeSecretItems,
		DefaultMode: int32Ptr(256),
	}

	// Add the secrets
	deployVolumes = append(deployVolumes, v1.Volume{
		Name: secretName,
		VolumeSource: v1.VolumeSource{
			Secret: &secretVolume,
		},
	})

	x.sendInProgress(e, "Secrets added to deployVolumes")

	// Do update/create of deployments and services
	depInterface := clientset.Extensions()
	batchv1DepInterface := clientset.BatchV1()

	// Validate we have some services to deploy
	if len(reData.Release.Services) == 0 {
		zeroServicesErr := fmt.Errorf("ERROR: Zero services were found in the deploy message.")
		x.sendErrorResponse(e, zeroServicesErr.Error())
		return zeroServicesErr
	}

	// Codeflow docker building container requires docker socket.
	if projectSlug == "codeamp-circuit" {
		deployVolumes = append(deployVolumes, v1.Volume{
			Name: "dockersocket",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/var/run/docker.sock",
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      "dockersocket",
			ReadOnly:  false,
			MountPath: "/var/run/docker.sock",
		})
	}

	// prioritize one-shot services over deployments
	// because migrations (which are one-shot jobs) should be
	// run before app code deployments

	var deploymentServices []plugins.Service
	var oneShotServices []plugins.Service

	for _, service := range reData.Release.Services {
		if service.Type == "one-shot" {
			oneShotServices = append(oneShotServices, service)
		} else {
			deploymentServices = append(deploymentServices, service)
		}
	}
	for index, service := range oneShotServices {
		oneShotServiceName := strings.ToLower(genOneShotServiceName(projectSlug, service.Name))

		// Check and delete any completed or failed jobs, and delete respective pods
		existingJobs, err := batchv1DepInterface.Jobs(namespace).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", "app", oneShotServiceName)})
		if err != nil {
			errMsg := fmt.Sprintf("Failed to list existing jobs with label app=%s, with error: %s", oneShotServiceName, err)
			oneShotServices[index].State = plugins.GetState("failed")
			oneShotServices[index].StateMessage = errMsg
			x.sendErrorResponse(e, oneShotServices[index].StateMessage)
			return nil
		}

		for index, job := range existingJobs.Items {
			if *job.Spec.Completions > 0 {
				if (job.Status.Active == 0 && job.Status.Failed == 0 && job.Status.Succeeded == 0) || job.Status.Active > 0 {
					oneShotServices[index].State = plugins.GetState("failed")
					oneShotServices[index].StateMessage = fmt.Sprintf("Cancelled deployment as a previous one-shot (%s) is still active. Redeploy your release once the currently running deployment process completes.", job.Name)
					x.sendErrorResponse(e, oneShotServices[index].StateMessage)
					return fmt.Errorf(oneShotServices[index].StateMessage)
				}
			}

			// delete old job
			gracePeriod := int64(0)
			deleteOptions := meta_v1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
			}

			err = batchv1DepInterface.Jobs(namespace).Delete(job.Name, &deleteOptions)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to delete job %s with err %s", job.Name, err))
			}

			correspondingPods, err := coreInterface.Pods(namespace).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", "app", oneShotServiceName)})
			if err != nil {
				log.Error(fmt.Sprintf("Failed to find corresponding pods with job-name %s with err %s", job.Name, err))
			}

			// delete associated pods
			for _, cp := range correspondingPods.Items {
				err := coreInterface.Pods(namespace).Delete(cp.Name, &meta_v1.DeleteOptions{})
				if err != nil {
					log.Error(fmt.Sprintf("Failed to delete pod %s with err %s", cp.Name, err))
				}
			}

			if err != nil {
				log.Error(fmt.Sprintf("Failed to delete job %s with err %s", job.Name, err))
			}
		}

		// Command parsing into entrypoint vs. args
		commandArray, _ := shlex.Split(service.Command)

		// Node selector
		var nodeSelector map[string]string
		if viper.IsSet("plugins.deployments.node_selector") {
			arrayKeyValue := strings.SplitN(viper.GetString("plugins.deployments.node_selector"), "=", 2)
			nodeSelector = map[string]string{arrayKeyValue[0]: arrayKeyValue[1]}
		}

		dockerImage, err := e.GetArtifactFromSource("image", "dockerbuilder")
		if err != nil {
			return err
		}

		simplePod := SimplePodSpec{
			Name:          oneShotServiceName,
			RestartPolicy: v1.RestartPolicyNever,
			NodeSelector:  nodeSelector,
			Args:          commandArray,
			Service:       service,
			Image:         dockerImage.String(),
			Env:           myEnvVars,
			VolumeMounts:  volumeMounts,
			Volumes:       deployVolumes,
		}

		jobSelector := meta_v1.LabelSelector{
			MatchLabels: map[string]string{
				"app": oneShotServiceName,
			},
		}

		podTemplateSpec := genPodTemplateSpec(e, simplePod, "Job")

		numParallelPods := int32(1)
		numCompletionsToTerminate := int32(service.Replicas)

		var jobParams *apis_batch_v1.Job
		jobParams = &apis_batch_v1.Job{
			TypeMeta: meta_v1.TypeMeta{
				Kind:       "Job",
				APIVersion: "batch/v1",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				GenerateName: fmt.Sprintf("%v-", oneShotServiceName),
				Labels:       map[string]string{"app": oneShotServiceName},
			},
			Spec: apis_batch_v1.JobSpec{
				Parallelism: &numParallelPods,
				Completions: &numCompletionsToTerminate,
				Template:    podTemplateSpec,
				Selector:    &jobSelector,
			},
		}

		createdJob, err := batchv1DepInterface.Jobs(namespace).Create(jobParams)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to create service job %s, with error: %s", createdJob.Name, err))
			oneShotServices[index].State = plugins.GetState("failed")
			oneShotServices[index].StateMessage = fmt.Sprintf("Failed to create job %s, with error: %s", createdJob.Name, err)
			x.sendErrorResponse(e, oneShotServices[index].StateMessage)
			return nil
		}

		// Loop and block any other jobs/ deployments from running until
		// the current job is terminated
		for {
			job, err := batchv1DepInterface.Jobs(namespace).Get(createdJob.Name, meta_v1.GetOptions{})
			if err != nil {
				log.Error(fmt.Sprintf("Error '%s' fetching job status for %s", err, createdJob.Name))
				time.Sleep(5 * time.Second)
				continue
			}

			log.Info(fmt.Sprintf("Job Status: Active: %v ; Succeeded: %v, Failed: %v \n", job.Status.Active, job.Status.Succeeded, job.Status.Failed))

			// Container is still creating
			if int32(service.Replicas) != 0 && job.Status.Active == 0 && job.Status.Failed == 0 && job.Status.Succeeded == 0 {
				time.Sleep(5 * time.Second)
				continue
			}

			if job.Status.Failed > 0 {
				// Job has failed. Delete job and report
				activeDeadlineSeconds := int64(1)

				job.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds
				job, err = batchv1DepInterface.Jobs(namespace).Update(job)
				if err != nil {
					log.Error(fmt.Sprintf("Error %s updating job %s before deletion", job.Name, err))
				}

				oneShotServices[index].State = plugins.GetState("failed")
				oneShotServices[index].StateMessage = fmt.Sprintf("Error job has failed %s", oneShotServiceName)
				x.sendErrorResponse(e, oneShotServices[index].StateMessage)
				return fmt.Errorf(oneShotServices[index].StateMessage)
			}

			if job.Status.Active == int32(0) {
				// Check for success
				if job.Status.Succeeded == int32(service.Replicas) {
					oneShotServices[index].State = plugins.GetState("complete")
					break
				} else {
					// Job has failed!
					oneShotServices[index].State = plugins.GetState("failed")
					oneShotServices[index].StateMessage = fmt.Sprintf("Error job has failed %s", oneShotServiceName)
					x.sendErrorResponse(e, oneShotServices[index].StateMessage)
					return fmt.Errorf(oneShotServices[index].StateMessage)
				}
			}

			// Check Job's Pod status
			if pods, err := clientset.Core().Pods(job.Namespace).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", "app", oneShotServiceName)}); err != nil {
				log.Error(fmt.Sprintf("List Pods of service[%s] error: %v", job.Name, err))
				oneShotServices[index].State = plugins.GetState("failed")
				oneShotServices[index].StateMessage = fmt.Sprintf("List Pods of service[%s] error: %v", job.Name, err)
				x.sendErrorResponse(e, oneShotServices[index].StateMessage)
			} else {
				for _, item := range pods.Items {
					if message, result := detectPodFailure(item); result {
						// Job has failed
						oneShotServices[index].State = plugins.GetState("failed")
						oneShotServices[index].StateMessage = fmt.Sprintf(message)
						x.sendErrorResponse(e, message)
						return fmt.Errorf(message)
					}
				}
			}
			time.Sleep(5 * time.Second)
		}
	}

	for index, service := range deploymentServices {
		deploymentName := genDeploymentName(projectSlug, service.Name)
		deployPorts := getContainerPorts(service)

		// Support ready and liveness probes
		var readyProbe v1.Probe
		var liveProbe v1.Probe
		var deployStrategy v1beta1.DeploymentStrategy
		if len(service.Listeners) >= 1 && service.Listeners[0].Protocol == "TCP" {
			// If the service is TCP, use a TCP Probe
			myPort := service.Listeners[0].Port
			readyProbe = v1.Probe{
				InitialDelaySeconds: 5,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				FailureThreshold:    3,
				TimeoutSeconds:      1,
				Handler: v1.Handler{
					TCPSocket: &v1.TCPSocketAction{
						Port: intstr.IntOrString{IntVal: myPort},
					},
				},
			}
			liveProbe = v1.Probe{
				InitialDelaySeconds: 15,
				PeriodSeconds:       20,
				SuccessThreshold:    1,
				FailureThreshold:    3,
				TimeoutSeconds:      1,
				Handler: v1.Handler{
					TCPSocket: &v1.TCPSocketAction{
						Port: intstr.IntOrString{IntVal: myPort},
					},
				},
			}
			deployStrategy = v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "30%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "60%",
					},
				},
			}
		} else {
			// If the service is non-TCP or has no ports use a simple exec probe
			runThis := []string{"/bin/true"}
			readyProbe = v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: runThis,
					},
				},
			}
			liveProbe = v1.Probe{
				InitialDelaySeconds: 15,
				PeriodSeconds:       20,
				SuccessThreshold:    1,
				FailureThreshold:    3,
				TimeoutSeconds:      1,
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: runThis,
					},
				},
			}
			deployStrategy = v1beta1.DeploymentStrategy{
				Type: v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "30%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "60%",
					},
				},
			}
		}

		// Deployment
		replicas := int32(service.Replicas)
		if service.Action == plugins.GetAction("destroy") {
			replicas = 0
		}

		// Command parsing into entrypoint vs. args
		commandArray, _ := shlex.Split(service.Command)

		// Node selector
		var nodeSelector map[string]string
		if viper.IsSet("plugins.deployments.node_selector") {
			arrayKeyValue := strings.SplitN(viper.GetString("plugins.deployments.node_selector"), "=", 2)
			nodeSelector = map[string]string{arrayKeyValue[0]: arrayKeyValue[1]}
		}

		var revisionHistoryLimit int32 = 10

		dockerImage, err := e.GetArtifactFromSource("image", "dockerbuilder")
		if err != nil {
			return err
		}

		simplePod := SimplePodSpec{
			Name:          deploymentName,
			DeployPorts:   deployPorts,
			ReadyProbe:    readyProbe,
			LiveProbe:     liveProbe,
			RestartPolicy: v1.RestartPolicyAlways,
			NodeSelector:  nodeSelector,
			Args:          commandArray,
			Service:       service,
			Image:         dockerImage.String(),
			Env:           myEnvVars,
			VolumeMounts:  volumeMounts,
			Volumes:       deployVolumes,
		}
		podTemplateSpec := genPodTemplateSpec(e, simplePod, "Deployment")

		var deployParams *v1beta1.Deployment

		deploySelector := meta_v1.LabelSelector{
			MatchLabels: map[string]string{
				"app": deploymentName,
			},
		}

		deployParams = &v1beta1.Deployment{
			TypeMeta: meta_v1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "extensions/v1beta1",
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name: deploymentName,
			},
			Spec: v1beta1.DeploymentSpec{
				ProgressDeadlineSeconds: int32Ptr(300),
				Replicas:                &replicas,
				Strategy:                deployStrategy,
				RevisionHistoryLimit:    &revisionHistoryLimit,
				Template:                podTemplateSpec,
				Selector:                &deploySelector,
			},
		}

		x.sendInProgress(e, "Deploy setup is complete. Created Replica-Set. Now Creating Deployment.")

		log.Info("Deploy setup is complete. Created Replica-Set. Now Creating Deployment.")
		log.Info(fmt.Sprintf("Getting list of deployments/ jobs matching %s", deploymentName))

		deployments := depInterface.Deployments(namespace)
		_, err = deployments.List(meta_v1.ListOptions{})
		if err != nil {
			log.Panic(err)
		}

		_, err = depInterface.Deployments(namespace).Get(deploymentName, meta_v1.GetOptions{})

		var myError error
		if err != nil {
			// Create deployment if it does not exist
			log.Warn(fmt.Sprintf("Existing deployment not found for %s. requested action: %s.", deploymentName, service.Action))
			// Sanity check that we were told to create this service or error out.

			x.sendInProgress(e, "Successfully creating Deployment.")
			_, myError = depInterface.Deployments(namespace).Create(deployParams)
			if myError != nil {
				// send failed status
				log.Error(fmt.Sprintf("Failed to create service deployment %s, with error: %s", deploymentName, myError))
				deploymentServices[index].State = plugins.GetState("failed")
				deploymentServices[index].StateMessage = fmt.Sprintf("Error creating deployment: %s", myError)
				// shorten the timeout in this case so that we can fail without waiting
				curTime = timeout
				x.sendErrorResponse(e, fmt.Sprintf("Service deployment failed: %s.", myError.Error()))
				return myError
			}
		} else {
			// Deployment exists, update deployment with new configuration
			_, myError = depInterface.Deployments(namespace).Update(deployParams)
			if myError != nil {
				log.Error(fmt.Sprintf("Failed to update service deployment %s, with error: %s", deploymentName, myError))
				deploymentServices[index].State = plugins.GetState("failed")
				deploymentServices[index].StateMessage = fmt.Sprintf("Failed to update deployment %s, with error: %s", deploymentName, myError)
				// shorten the timeout in this case so that we can fail without waiting
				curTime = timeout
				x.sendErrorResponse(e, fmt.Sprintf("Service deployment failed: %s.", myError.Error()))
				return myError
			}
		}

	} // All service deployments initiated.

	log.Info(fmt.Sprintf("Waiting %d seconds for deployment to succeed.", timeout))
	for i := range deploymentServices {
		deploymentServices[i].State = plugins.GetState("waiting")
	}

	if len(deploymentServices) > 0 {
		// Check status of all deployments till the succeed or timeout.
		replicaFailures := 0
		for {
			for index, service := range deploymentServices {
				deploymentName := strings.ToLower(genDeploymentName(projectSlug, service.Name))
				deployment, err := depInterface.Deployments(namespace).Get(deploymentName, meta_v1.GetOptions{})
				if err != nil {
					log.Error(fmt.Sprintf("Error '%s' fetching deployment status for %s", err, deploymentName))
					continue
				}
				log.Info(fmt.Sprintf("Waiting for %s; ObservedGeneration: %d, Generation: %d, UpdatedReplicas: %d, Replicas: %d, AvailableReplicas: %d, UnavailableReplicas: %d", deploymentName, deployment.Status.ObservedGeneration, deployment.ObjectMeta.Generation, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas, deployment.Status.AvailableReplicas, deployment.Status.UnavailableReplicas))
				if deployment.Status.ObservedGeneration >= deployment.ObjectMeta.Generation && deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas && deployment.Status.AvailableReplicas >= deployment.Status.UpdatedReplicas && deployment.Status.UnavailableReplicas == 0 {
					// deployment success
					deploymentServices[index].State = plugins.GetState("complete")
					successfulDeploys = 0
					for _, d := range deploymentServices {
						if d.State == plugins.GetState("complete") {
							successfulDeploys++
						}
					}
					log.Info(fmt.Sprintf("%s deploy: %d of %d deployments successful.", deploymentName, successfulDeploys, len(deploymentServices)))

					if successfulDeploys == len(deploymentServices) {
						break
					}
					// If this deployment has succeeded then we can skip the failure checks below.
					continue
				}

				latestRevision := deployment.Annotations["deployment.kubernetes.io/revision"]

				// Check for indications of pod failures on the latest replicaSet so we can fail faster than waiting for a timeout.
				matchLabel := make(map[string]string)
				matchLabel["app"] = deploymentName
				replicaSetList, err := depInterface.ReplicaSets(namespace).List(meta_v1.ListOptions{
					LabelSelector: "app=" + deploymentName,
				})

				var currentReplica v1beta1.ReplicaSet

				for _, r := range replicaSetList.Items {
					if r.Annotations["deployment.kubernetes.io/revision"] == latestRevision {
						currentReplica = r
						break
					}
				}

				allPods, podErr := coreInterface.Pods(namespace).List(meta_v1.ListOptions{})
				if podErr != nil {
					log.Error(fmt.Sprintf("Error retrieving list of pods for %s", namespace))
					continue
				}

				for _, pod := range allPods.Items {
					for _, ref := range pod.ObjectMeta.OwnerReferences {
						if ref.Kind == "ReplicaSet" {
							if ref.Name == currentReplica.Name {
								// This is a pod we want to check status for
								if message, result := detectPodFailure(pod); result {
									// Pod is waiting forever, fail the deployment.
									x.sendErrorResponse(e, message)
									return fmt.Errorf(message)
								}
							}
						}
					}
				}
			}

			if successfulDeploys == len(deploymentServices) {
				break
			}

			if curTime >= timeout || replicaFailures > 1 {
				errMsg := fmt.Sprintf("Error, timeout reached waiting for all deployments to succeed.")
				log.Error(fmt.Sprintf(errMsg))
				x.sendErrorResponse(e, errMsg)
				return fmt.Errorf(errMsg)
			}
			time.Sleep(5 * time.Second)
			curTime += 5
		}

	}

	x.sendSuccessResponse(e, plugins.GetState("complete"), nil)

	// all success!
	log.Info(fmt.Sprintf("All deployments successful."))
	// Cleanup orphan jobs
	existingJobs, err := batchv1DepInterface.Jobs(namespace).List(meta_v1.ListOptions{})
	if err != nil {
		log.Error(fmt.Sprintf("Failed to list existing jobs in namespace %s, with error: %s", namespace, err))
	}

	for _, job := range existingJobs.Items {
		var foundIt bool
		for _, service := range oneShotServices {
			oneShotServiceName := strings.ToLower(genOneShotServiceName(projectSlug, service.Name))
			if oneShotServiceName == job.Labels["app"] {
				foundIt = true
			}
		}

		if foundIt == false {
			log.Debug(fmt.Sprintf("Deleting orphan job %s", job.Name))
			gracePeriod := int64(0)
			isOrphan := true
			deleteOptions := meta_v1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
				OrphanDependents:   &isOrphan,
			}

			err = batchv1DepInterface.Jobs(namespace).Delete(job.Name, &deleteOptions)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to delete orphan job %s with err %s", job.Name, err))
			}
		}
	}

	// cleanup Orphans! (these are deployments leftover from rename or etc.)
	allDeploymentsList, listErr := depInterface.Deployments(namespace).List(meta_v1.ListOptions{})
	if listErr != nil {
		// If we can't list the deployments just return.  We have already sent the success message.
		log.Error(fmt.Sprintf("Fatal Error listing deployments during cleanup.  %s", listErr))
		return nil
	}
	var foundIt bool
	var orphans []v1beta1.Deployment
	for _, deployment := range allDeploymentsList.Items {
		foundIt = false
		for _, service := range reData.Release.Services {
			if deployment.Name == genDeploymentName(projectSlug, service.Name) {
				foundIt = true
			}
		}
		if foundIt == false {
			orphans = append(orphans, deployment)
		}
	}

	// Preload list of all replica sets
	repSets, repErr := depInterface.ReplicaSets(namespace).List(meta_v1.ListOptions{})
	if repErr != nil {
		log.Error(fmt.Sprintf("Error retrieving list of replicasets for %s", namespace))
		return repErr
	}

	// Preload list of all pods
	allPods, podErr := coreInterface.Pods(namespace).List(meta_v1.ListOptions{})
	if podErr != nil {
		log.Error(fmt.Sprintf("Error retrieving list of pods for %s", namespace))
		return podErr
	}

	// Delete the deployments
	for _, deleteThis := range orphans {
		matched, _ := regexp.MatchString("^keep", deleteThis.Name)
		if matched {
			continue
		}

		log.Debug(fmt.Sprintf("Deleting deployment orphan: %s", deleteThis.Name))
		err := depInterface.Deployments(namespace).Delete(deleteThis.Name, &meta_v1.DeleteOptions{})
		if err != nil {
			log.Error(fmt.Sprintf("Error when deleting: %s", err))
		}

		// Delete the replicasets (cascade)
		for _, repSet := range repSets.Items {
			if repSet.ObjectMeta.Labels["app"] == deleteThis.Name {
				log.Debug(fmt.Sprintf("Deleting replicaset orphan: %s", repSet.Name))
				err := depInterface.ReplicaSets(namespace).Delete(repSet.Name, &meta_v1.DeleteOptions{})
				if err != nil {
					log.Error(fmt.Sprintf("Error '%s' while deleting replica set %s", err, repSet.Name))
				}
			}
		}

		// Delete the pods (cascade) or scale down the repset
		for _, pod := range allPods.Items {
			if pod.ObjectMeta.Labels["app"] == deleteThis.Name {
				log.Debug(fmt.Sprintf("Deleting pod orphan: %s", pod.Name))
				err := coreInterface.Pods(namespace).Delete(pod.Name, &meta_v1.DeleteOptions{})
				if err != nil {
					log.Error(fmt.Sprintf("Error '%s' while deleting pod %s", err, pod.Name))
				}
			}
		}
	}

	return nil
}
