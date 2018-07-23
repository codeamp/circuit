package kubernetes

import (
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	"k8s.io/api/core/v1"
)

// Load Balancer
type ListenerPair struct {
	Name       string
	Protocol   string
	SourcePort int32
	TargetPort int32
}

// Deployments
type SimplePodSpec struct {
	Name           string
	DeployPorts    []v1.ContainerPort
	ReadinessProbe v1.Probe
	LivenessProbe  v1.Probe
	RestartPolicy  v1.RestartPolicy
	NodeSelector   map[string]string
	Args           []string
	Service        plugins.Service
	Image          string
	Env            []v1.EnvVar
	VolumeMounts   []v1.VolumeMount
	Volumes        []v1.Volume
}

// Kubernetes
type Kubernetes struct {
	events chan transistor.Event
}

// ProbeDefaults
type ProbeDefaults struct {
	InitialDelaySeconds int32
	PeriodSeconds       int32
	SuccessThreshold    int32
	FailureThreshold    int32
	TimeoutSeconds      int32
}
