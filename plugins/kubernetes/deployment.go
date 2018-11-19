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
	"k8s.io/client-go/kubernetes"

	apis_batch_v1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/google/shlex"
	"github.com/spf13/viper"
)

var deploySleepTime = 5 * time.Second
var timeout = 600

func (x *Kubernetes) ProcessDeployment(e transistor.Event) {
	if e.Matches("project:") {
		if e.Action == transistor.GetAction("create") {
			event := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), fmt.Sprintf("%s has completed successfully", e.Event()))
			x.events <- event
			return
		}

		if e.Action == transistor.GetAction("update") {
			event := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), fmt.Sprintf("%s has completed successfully", e.Event()))
			x.events <- event

			return
		}
	}

	if e.Matches("release:") {
		if e.Action == transistor.GetAction("create") {
			err := x.doDeploy(e)
			if err != nil {
				log.Error(err)
			}

			return
		}
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

func (x *Kubernetes) createDockerIOSecretIfNotExists(namespace string, clientset kubernetes.Interface, e transistor.Event) error {
	coreInterface := clientset.Core()

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
			_, createDockerIOSecretErr := x.CoreSecreter.Create(clientset, namespace, &v1.Secret{
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

func (x *Kubernetes) createNamespaceIfNotExists(namespace string, clientset kubernetes.Interface) error {
	coreInterface := clientset.Core()

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

func getDeploymentStrategy(service plugins.Service, rollback bool) v1beta1.DeploymentStrategy {
	var defaultDeploymentStrategy = v1beta1.DeploymentStrategy{
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

	if rollback {
		return v1beta1.DeploymentStrategy{
			Type: v1beta1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &v1beta1.RollingUpdateDeployment{
				MaxUnavailable: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "70%",
				},
				MaxSurge: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "100%",
				},
			},
		}
	}

	if service.DeploymentStrategy == (plugins.DeploymentStrategy{}) {
		return defaultDeploymentStrategy
	}

	switch service.DeploymentStrategy.Type {
	case plugins.GetType("default"):
		return defaultDeploymentStrategy
	case plugins.GetType("recreate"):
		return v1beta1.DeploymentStrategy{
			Type: v1beta1.RecreateDeploymentStrategyType,
		}
	case plugins.GetType("rollingUpdate"):
		customDeploymentStrategy := defaultDeploymentStrategy
		customDeploymentStrategy.RollingUpdate = &v1beta1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: fmt.Sprintf("%d%%", service.DeploymentStrategy.MaxUnavailable),
			},
			MaxSurge: &intstr.IntOrString{
				Type:   intstr.String,
				StrVal: fmt.Sprintf("%d%%", service.DeploymentStrategy.MaxSurge),
			},
		}

		return customDeploymentStrategy
	default:
		return defaultDeploymentStrategy
	}
}

func getReadinessProbe(service plugins.Service) v1.Probe {
	defaults := ProbeDefaults{
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
		TimeoutSeconds:      1,
	}

	if service.ReadinessProbe.Type != "" {
		return getHealthProbe(service.ReadinessProbe, defaults)
	}

	// no service listeners defined,
	var probe plugins.ServiceHealthProbe
	if len(service.Listeners) >= 1 && service.Listeners[0].Protocol == "TCP" {
		probe = plugins.ServiceHealthProbe{
			Method: "tcp",
			Port:   service.Listeners[0].Port,
		}
	} else {
		probe = plugins.ServiceHealthProbe{
			Method:  "exec",
			Command: "/bin/true",
		}
	}
	return getHealthProbe(probe, defaults)
}

func getLivenessProbe(service plugins.Service) v1.Probe {
	defaults := ProbeDefaults{
		InitialDelaySeconds: 15,
		PeriodSeconds:       20,
		SuccessThreshold:    1,
		FailureThreshold:    3,
		TimeoutSeconds:      1,
	}

	if service.LivenessProbe.Type != "" {
		return getHealthProbe(service.LivenessProbe, defaults)
	}

	var probe plugins.ServiceHealthProbe
	if len(service.Listeners) >= 1 && service.Listeners[0].Protocol == "TCP" {
		probe = plugins.ServiceHealthProbe{
			Method: "tcp",
			Port:   service.Listeners[0].Port,
		}
	} else {
		probe = plugins.ServiceHealthProbe{
			Method:  "exec",
			Command: "/bin/true",
		}
	}

	return getHealthProbe(probe, defaults)
}

func getHealthProbe(probe plugins.ServiceHealthProbe, defaults ProbeDefaults) v1.Probe {

	var v1Probe v1.Probe
	var handler v1.Handler

	// set handler
	switch method := probe.Method; method {
	case "http":
		var scheme v1.URIScheme
		if probe.Scheme == "https" {
			scheme = v1.URISchemeHTTPS
		} else {
			scheme = v1.URISchemeHTTP
		}
		var headers []v1.HTTPHeader
		for _, h := range probe.HttpHeaders {
			header := v1.HTTPHeader{
				Name:  h.Name,
				Value: h.Value,
			}
			headers = append(headers, header)
		}
		handler = v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path:        probe.Path,
				Port:        intstr.IntOrString{IntVal: probe.Port},
				Scheme:      scheme,
				HTTPHeaders: headers,
			},
		}
	case "exec":
		command := strings.Split(probe.Command, " ")
		handler = v1.Handler{
			Exec: &v1.ExecAction{
				Command: command,
			},
		}
	case "tcp":
		handler = v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.IntOrString{IntVal: probe.Port},
			},
		}
	default:
		handler = v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.IntOrString{IntVal: probe.Port},
			},
		}
	}
	v1Probe.Handler = handler

	// set default thresholds
	if probe.InitialDelaySeconds > 0 {
		v1Probe.InitialDelaySeconds = probe.InitialDelaySeconds
	} else {
		v1Probe.InitialDelaySeconds = defaults.InitialDelaySeconds
	}

	if probe.PeriodSeconds > 0 {
		v1Probe.PeriodSeconds = probe.PeriodSeconds
	} else {
		v1Probe.PeriodSeconds = defaults.PeriodSeconds
	}

	if probe.SuccessThreshold > 0 {
		v1Probe.SuccessThreshold = probe.SuccessThreshold
	} else {
		v1Probe.SuccessThreshold = defaults.SuccessThreshold
	}

	if probe.FailureThreshold > 0 {
		v1Probe.FailureThreshold = probe.FailureThreshold
	} else {
		v1Probe.FailureThreshold = defaults.FailureThreshold
	}

	if probe.TimeoutSeconds > 0 {
		v1Probe.TimeoutSeconds = probe.TimeoutSeconds
	} else {
		v1Probe.TimeoutSeconds = defaults.TimeoutSeconds
	}

	return v1Probe
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
		ImagePullPolicy: v1.PullIfNotPresent,
		Env:             podConfig.Env,
		VolumeMounts:    podConfig.VolumeMounts,
	}
	if kind == "Deployment" {
		container.ReadinessProbe = &podConfig.ReadinessProbe
		container.LivenessProbe = &podConfig.LivenessProbe
		if podConfig.PreStopHook != (v1.Handler{}) {
			container.Lifecycle = &v1.Lifecycle{
				PreStop: &podConfig.PreStopHook,
			}
		}
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

// Create the secrets for the deployment
func (x *Kubernetes) createSecretsForDeploy(clientset kubernetes.Interface, namespace string, projectSlug string, secrets *[]plugins.Secret) (string, error) {
	var secretMap map[string]string
	secretMap = make(map[string]string)

	// This map is used in to create the secrets themselves
	for _, secret := range secrets {
		secretMap[secret.Key] = secret.Value
	}

	secretParams := v1.Secret{
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

	secretResult, secErr := x.CoreSecreter.Create(clientset, namespace, &secretParams)
	if secErr != nil {
		failMessage := fmt.Sprintf("Error '%s' creating secret %s", secErr, projectSlug)
		return "", fmt.Errorf(failMessage)
	}

	return secretResult.Name, nil
}

// Build the configuration needed for the environment of the deploy
func (x *Kubernetes) setupEnvironmentForDeploy(secretName string, secrets *[]plugins.Secret) ([]v1.EnvVar, []v1.VolumeMount, []v1.Volume, []v1.KeyToPath, error) {
	log.Warn("setupEnvironmentForDeploy")
	// This is for building the configuration to use the secrets from inside the deployment
	// as ENVs
	var envVars []v1.EnvVar
	for _, secret := range secrets {
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
			envVars = append(envVars, newEnv)
		}
	}
	// expose pod details to running container via env variables
	envVars = x.exposePodInfoViaEnvVariable(envVars)

	/******************************************
	*	Place File-type Env Vars on FS
	*******************************************/
	var volumeMounts []v1.VolumeMount
	var deployVolumes []v1.Volume
	var volumeSecretItems []v1.KeyToPath
	volumeMounts = append(volumeMounts, v1.VolumeMount{
		Name:      secretName,
		MountPath: "/etc/secrets",
		ReadOnly:  true,
	})

	for _, secret := range secrets {
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

	return envVars, volumeMounts, deployVolumes, volumeSecretItems, nil
}

// Deploy a one shot array of services
func (x *Kubernetes) deployOneShotServices(clientset kubernetes.Interface,
	e transistor.Event, namespace string, projectSlug string,
	envVars []v1.EnvVar, volumeMounts []v1.VolumeMount, deployVolumes []v1.Volume, secretItems []v1.KeyToPath,
	oneShotServices []plugins.Service) error {
	batchv1DepInterface := clientset.BatchV1()
	coreInterface := clientset.Core()

	log.Warn("deployOneShotServices")

	// For all OneShot Services
	for index, service := range oneShotServices {
		oneShotServiceName := strings.ToLower(genOneShotServiceName(projectSlug, service.Name))

		// Check and delete any completed or failed jobs, and delete respective pods
		listOptions := meta_v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", "app", oneShotServiceName)}
		existingJobs, err := batchv1DepInterface.Jobs(namespace).List(listOptions)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to list existing jobs with label app=%s, with error: %s", oneShotServiceName, err)
			log.Error(errMsg)
			return fmt.Errorf(errMsg)
		}

		for _, job := range existingJobs.Items {
			if *job.Spec.Completions > 0 {
				if (job.Status.Active == 0 && job.Status.Failed == 0 && job.Status.Succeeded == 0) || job.Status.Active > 0 {
					errMsg := fmt.Sprintf("Cancelled deployment as a previous one-shot (%s) is still active. Redeploy your release once the currently running deployment process completes.", job.Name)
					log.Error(errMsg)
					return fmt.Errorf(errMsg)
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

		// expose codeamp service name via env variable
		podEnvVars := append(envVars, v1.EnvVar{
			Name:  "CODEAMP_SERVICE_NAME",
			Value: service.Name,
		})

		simplePod := SimplePodSpec{
			Name:          oneShotServiceName,
			RestartPolicy: v1.RestartPolicyNever,
			NodeSelector:  nodeSelector,
			Args:          commandArray,
			Service:       service,
			Image:         dockerImage.String(),
			Env:           podEnvVars,
			VolumeMounts:  volumeMounts,
			Volumes:       deployVolumes,
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
			},
		}

		// Create the job
		createdJob, err := batchv1DepInterface.Jobs(namespace).Create(jobParams)
		if err != nil {
			errMsg := fmt.Errorf("Failed to create job %s, with error: %s", createdJob.Name, err)
			log.Error(errMsg)

			return errMsg
		}

		// Loop and block any other jobs/ deployments from running until
		// the current job is terminated
		for {
			job, err := x.BatchV1Jobber.Get(clientset, namespace, createdJob.Name, meta_v1.GetOptions{})
			if err != nil {
				log.Error(fmt.Sprintf("Error '%s' fetching job status for %s", err, createdJob.Name))
				time.Sleep(deploySleepTime)
				continue
			}

			log.Info(fmt.Sprintf("Job Status: Active: %v ; Succeeded: %v, Failed: %v \n", job.Status.Active, job.Status.Succeeded, job.Status.Failed))

			// Container is still creating
			if int32(service.Replicas) != 0 && job.Status.Active == 0 && job.Status.Failed == 0 && job.Status.Succeeded == 0 {
				time.Sleep(deploySleepTime)
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

				return fmt.Errorf("Error job has failed %s", oneShotServiceName)
			}

			if job.Status.Active == int32(0) {
				// Check for success
				if job.Status.Succeeded == int32(service.Replicas) {
					oneShotServices[index].State = transistor.GetState("complete")
					break
				} else {
					// Job has failed!
					return fmt.Errorf("Error job has failed %s", oneShotServiceName)
				}
			}

			// Check Job's Pod status
			if pods, err := coreInterface.Pods(job.Namespace).List(meta_v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", "app", oneShotServiceName)}); err != nil {
				return fmt.Errorf("List Pods of service[%s] error: %v", job.Name, err)
			} else {
				for _, item := range pods.Items {
					if message, result := detectPodFailure(item); result {
						// Job has failed
						return fmt.Errorf(message)
					}
				}
			}

			time.Sleep(deploySleepTime)
		}
	}

	return nil
}

// Deploy a general array of services
func (x *Kubernetes) deployServices(clientset kubernetes.Interface,
	e transistor.Event, namespace string, projectSlug string, isRollback bool,
	envVars []v1.EnvVar, volumeMounts []v1.VolumeMount, deployVolumes []v1.Volume, secretItems []v1.KeyToPath, deploymentServices []plugins.Service) error {
	depInterface := clientset.Extensions()

	log.Warn("deployServices")

	// Now process all deployment services
	for _, service := range deploymentServices {
		deploymentName := genDeploymentName(projectSlug, service.Name)
		deployPorts := getContainerPorts(service)

		var deployStrategy v1beta1.DeploymentStrategy

		// Support ready and liveness probes
		readinessProbe := getReadinessProbe(service)
		livenessProbe := getLivenessProbe(service)

		deployStrategy = getDeploymentStrategy(service, isRollback)

		// Deployment
		replicas := int32(service.Replicas)
		if service.Action == transistor.GetAction("delete") {
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

		// expose codeamp service name via env variable
		podEnvVars := append(envVars, v1.EnvVar{
			Name:  "CODEAMP_SERVICE_NAME",
			Value: service.Name,
		})

		var preStopHook v1.Handler
		if service.PreStopHook != "" {
			preStopHook = v1.Handler{
				Exec: &v1.ExecAction{
					Command: strings.Split(service.PreStopHook, " "),
				},
			}
		}

		simplePod := SimplePodSpec{
			Name:           deploymentName,
			DeployPorts:    deployPorts,
			ReadinessProbe: readinessProbe,
			LivenessProbe:  livenessProbe,
			PreStopHook:    preStopHook,
			RestartPolicy:  v1.RestartPolicyAlways,
			NodeSelector:   nodeSelector,
			Args:           commandArray,
			Service:        service,
			Image:          dockerImage.String(),
			Env:            podEnvVars,
			VolumeMounts:   volumeMounts,
			Volumes:        deployVolumes,
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
				errMsg := fmt.Errorf("Failed to create service deployment %s, with error: %s", deploymentName, myError)
				log.Error(errMsg)
				return errMsg
			}
		} else {
			// Deployment exists, update deployment with new configuration
			_, myError = depInterface.Deployments(namespace).Update(deployParams)
			if myError != nil {
				errMsg := fmt.Errorf("Failed to update service deployment %s, with error: %s", deploymentName, myError)
				log.Error(errMsg)
				return errMsg
			}
		}

	} // All service deployments initiated.

	return nil
}

// Blocks the main thread while waiting for a succesful deployment
func (x *Kubernetes) waitForDeploymentSuccess(clientset kubernetes.Interface,
	namespace string, projectSlug string, deploymentServices []plugins.Service) error {

	log.Warn("waitForDeploymentSuccess")

	coreInterface := clientset.Core()
	depInterface := clientset.Extensions()

	for i := range deploymentServices {
		deploymentServices[i].State = transistor.GetState("waiting")
	}
	successfulDeploys := 0
	curTime := 0

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
					deploymentServices[index].State = transistor.GetState("complete")
					successfulDeploys = 0
					for _, d := range deploymentServices {
						if d.State == transistor.GetState("complete") {
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
				return fmt.Errorf(errMsg)
			}
			time.Sleep(deploySleepTime)
			curTime += int(deploySleepTime / time.Second)
		}
	}

	return nil
}

// Cleans up any resources leftover from a previous or inactive deployment
func (x *Kubernetes) cleanupOrphans(clientset kubernetes.Interface,
	namespace string, projectSlug string, oneShotServices []plugins.Service, services []plugins.Service) error {

	log.Warn("Cleanup orphans")

	batchv1DepInterface := clientset.BatchV1()
	depInterface := clientset.Extensions()
	coreInterface := clientset.Core()

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
		for _, service := range services {
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

func (x *Kubernetes) doDeploy(e transistor.Event) error {
	log.Warn("doDeploy")
	/******************************************
	*
	*	Build Kubernetes Configuration
	*
	*******************************************/
	reData := e.Payload.(plugins.ReleaseExtension)
	projectSlug := plugins.GetSlug(reData.Release.Project.Repository)

	kubeconfig, err := x.SetupKubeConfig(e)
	if err != nil {
		log.Error(err.Error())
		x.sendErrorResponse(e, "failed writing kubeconfig")
		return err
	}

	/******************************************
	*
	*	Create ClientSet
	*
	*******************************************/
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

	clientset, err := x.K8sNamespacer.NewForConfig(config)
	if err != nil {
		log.Error("Error getting cluster config.  Aborting!")
		x.sendErrorResponse(e, err.Error())
		return err
	}
	log.Warn("built config")

	/******************************************
	*
	*	Report: Deploy In Progress
	*
	*******************************************/
	x.sendInProgress(e, "Deploy in-progress")

	/******************************************
	*
	*	Build Prospective Namespace Name
	*
	*******************************************/
	log.Warn("building namespace name")
	namespace := x.GenNamespaceName(reData.Release.Environment, projectSlug)

	// TODO: get timeout from formValues
	//timeout := e.Payload.(plugins.ReleaseExtension).Release.Timeout

	/******************************************
	*
	*	Ensure Namespace Exists
	*
	*******************************************/
	log.Warn("creating namespace name")
	createNamespaceErr := x.createNamespaceIfNotExists(namespace, clientset)
	if createNamespaceErr != nil {
		x.sendErrorResponse(e, createNamespaceErr.Error())
		return createNamespaceErr
	}

	/******************************************
	*
	*	Create Docker IO Secret
	*
	*******************************************/
	log.Warn("creting docker io secret")
	createDockerIOSecretErr := x.createDockerIOSecretIfNotExists(namespace, clientset, e)
	if createDockerIOSecretErr != nil {
		x.sendErrorResponse(e, createDockerIOSecretErr.Error())
		return createDockerIOSecretErr
	}

	/******************************************
	*
	*	Create Secrets for Deploy
	*
	*******************************************/
	secrets := reData.Release.Secrets

	log.Warn("creating secrets for deploy")
	secretName, err := x.createSecretsForDeploy(clientset, namespace, projectSlug, secrets)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	x.sendInProgress(e, "Secrets created")

	/******************************************
	*
	*	Build Environment / EnvVars
	*
	*******************************************/
	log.Warn("setup environment for deploy")
	envVars, volumeMounts, volumes, volumeSecretItems, err := x.setupEnvironmentForDeploy(secretName, secrets)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	/******************************************
	*
	*	Update/Create Deployment & Services
	*
	*******************************************/
	// Validate we have some services to deploy
	log.Warn("validating services to deploy")
	if len(reData.Release.Services) == 0 {
		zeroServicesErr := fmt.Errorf("ERROR: Zero services were found in the deploy message.")
		x.sendErrorResponse(e, zeroServicesErr.Error())
		return zeroServicesErr
	}

	/******************************************
	*
	*	Enable Docker Socket for Specified Deployments
	*
	*******************************************/
	// Codeflow docker building container requires docker socket.
	if projectSlug == "codeamp-circuit" {
		volumes = append(volumes, v1.Volume{
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

	log.Warn("releasing loop services")
	for _, service := range reData.Release.Services {
		if service.Type == "one-shot" {
			if !reData.Release.IsRollback {
				oneShotServices = append(oneShotServices, service)
			}
		} else if service.Type == "general" {
			deploymentServices = append(deploymentServices, service)
		} else {
			return fmt.Errorf("This service type is not supported. Try again with either 'one-shot' or 'general' as the service type.")
		}
	}

	// One Shot Services
	log.Warn("one shot services")
	err = x.deployOneShotServices(clientset, e, namespace, projectSlug, envVars, volumeMounts, volumes, volumeSecretItems, oneShotServices)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// Deployment Services
	log.Warn("deployment services")
	err = x.deployServices(clientset, e, namespace, projectSlug, reData.Release.IsRollback, envVars, volumeMounts, volumes, volumeSecretItems, deploymentServices)
	if err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	/******************************************
	*
	*	Wait for deployment to succeed
	*
	*******************************************/
	log.Info(fmt.Sprintf("Waiting %d seconds for deployment to succeed.", timeout))
	if err := x.waitForDeploymentSuccess(clientset, namespace, projectSlug, deploymentServices); err != nil {
		x.sendErrorResponse(e, err.Error())
		return err
	}

	// all success!
	x.sendSuccessResponse(e, transistor.GetState("complete"), nil)
	log.Info(fmt.Sprintf("All deployments successful."))

	/******************************************
	*
	*	Cleanup orphans and environment
	*
	*******************************************/
	log.Warn("cleaning up orphans")
	if err := x.cleanupOrphans(clientset, namespace, projectSlug, oneShotServices, deploymentServices); err != nil {
		log.Error(err)
	}

	return nil
}

func (x *Kubernetes) exposePodInfoViaEnvVariable(myEnvVars []v1.EnvVar) []v1.EnvVar {
	log.Warn("exposePodInfoViaEnvVariable")

	// TODO rename to KUBE_POD_IP for consistency when all consumers get updated
	myEnvVars = append(myEnvVars, v1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	})

	myEnvVars = append(myEnvVars, v1.EnvVar{
		Name: "KUBE_NODE_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "spec.nodeName",
			},
		},
	})

	myEnvVars = append(myEnvVars, v1.EnvVar{
		Name: "KUBE_POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	})

	myEnvVars = append(myEnvVars, v1.EnvVar{
		Name: "KUBE_POD_NAMESPACE",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.namespace",
			},
		},
	})

	return myEnvVars
}
