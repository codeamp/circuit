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
	PreStopHook    v1.Handler
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

type IngressInput struct {
	FQDN                 string
	KubeConfig           string
	ClientCertificate    string
	ClientKey            string
	CertificateAuthority string
	Port                 ListenerPair
	Controller           IngressController
}

type IngressController struct {
	Subdomain      string
	ControllerName string
	ControllerID   string
}
