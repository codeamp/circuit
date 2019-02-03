package kubernetes

import (
	"github.com/go-errors/errors"
)

// Deployment errors
var ErrDeployNoSecrets = errors.Errorf("There were no secrets provided for deploying")
var ErrDeploySetupEnvironmentNoSecrets = errors.Errorf("There were no secrets provided for setting up environment")
var ErrDeploySecretCreate = errors.Errorf("There was an error creating the secret")
var ErrDeployServiceTypeNotSupported = errors.Errorf("This service type is not supported. Try again with either 'one-shot' or 'general' as the service type.")

var ErrDeployOneShotActive = errors.Errorf("Canceled deployment because one-shot service is still active. Redeploy your release once the currently running deployment process completes.")
var ErrDeployJobCreate = errors.Errorf("Job has failed to create")
var ErrDeployJobStarting = errors.Errorf("Job has failed to start")

var ErrDeployListingPods = errors.Errorf("Failed to get list of pods")
var ErrDeployListingJobs = errors.Errorf("Failed to get list of jobs")
var ErrDeployListingReplicaSets = errors.Errorf("Failed to get list of replica-sets")

var ErrDeployJobUpdate = errors.Errorf("Faild to create job")
var ErrDeployServiceCreate = errors.Errorf("Failed to create service")
var ErrDeployServiceUpdate = errors.Errorf("Failed to update service")

var ErrDeployPodWaitingForever = errors.Errorf("Pod is waiting forever")
var ErrDeployPodWaitingForeverUnwindingDeploy = errors.Errorf("Pod is waiting forever - Unwinding Deploy")
var ErrDeployTimeout = errors.Errorf("Error, timeout reached waiting for all deployments to succeed.")

// Services Errors
var ErrKubernetesClientSetup = errors.Errorf("You must set the environment variable CF_PLUGINS_KUBEDEPLOY_KUBECONFIG=/path/to/kubeconfig")
var ErrKubernetesNewForConfig = errors.Errorf("Could not create config for Kubernetes")
var ErrServiceUnexpectedError = errors.Errorf("There was an unexpected error involving services")
var ErrServiceUpdateFailed = errors.Errorf("Failed to update service")
var ErrServiceCreateFailed = errors.Errorf("Failed to create service")
var ErrServiceDeleteFailed = errors.Errorf("Failed to delete service")
var ErrServiceDeleteNotFound = errors.Errorf("Failed to find service when deleting")

var ErrServiceDNSTimeout = errors.Errorf("Timeout waiting for ELB DNS name")
