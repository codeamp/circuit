package kubernetes

import (
	"github.com/go-errors/errors"
)

var ErrorNoSecretsDeploy = errors.Errorf("There were no secrets provided for deploying")
var ErrorNoSecretsSetupEnvironment = errors.Errorf("There were no secrets provided for setting up environment")
var ErrorSecretCreate = errors.Errorf("There was an error creating the secret")
var ErrorServiceNotSupported = errors.Errorf("This service type is not supported. Try again with either 'one-shot' or 'general' as the service type.")

var ErrorOneShotActive = errors.Errorf("Canceled deployment because one-shot service is still active")
var ErrorJobFailedCreate = errors.Errorf("Job has failed to create")
var ErrorJobFailedStarting = errors.Errorf("Job has failed to start")

var ErrorFailedListingPods = errors.Errorf("Failed to get list of pods")
var ErrorFailedListingJobs = errors.Errorf("Failed to get list of jobs")

var ErrorFailedJobUpdate = errors.Errorf("Faild to create job")
var ErrorFailedServiceCreate = errors.Errorf("Failed to create service")
var ErrorFailedServiceUpdate = errors.Errorf("Failed to update service")

var ErrorPodWaitingForever = errors.Errorf("Pod is waiting forever")
var ErrorDeploymentTimeout = errors.Errorf("Error, timeout reached waiting for all deployments to succeed.")
