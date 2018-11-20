package kubernetes

import (
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	contour_client "github.com/heptio/contour/apis/generated/clientset/versioned"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
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
	K8sNamespacer
	K8sContourNamespacer
	BatchV1Jobber
	CoreServicer
	CoreSecreter

	ContourClient    contour_client.Interface
	KubernetesClient kubernetes.Interface
}

// ProbeDefaults
type ProbeDefaults struct {
	InitialDelaySeconds int32
	PeriodSeconds       int32
	SuccessThreshold    int32
	FailureThreshold    int32
	TimeoutSeconds      int32
}

type Service struct {
	ID   string
	Name string
	Port ListenerPair
}

type IngressController struct {
	ControllerName string
	ControllerID   string
	ELB            string
}

type IngressInput struct {
	Type                 string
	KubeConfig           string
	ClientCertificate    string
	ClientKey            string
	CertificateAuthority string
	Controller           IngressController
	Service              Service
	ControlledApexDomain string
	UpstreamFQDNs        []Domain
}

type IngressRouteInput struct {
	Type                 string
	KubeConfig           string
	ClientCertificate    string
	ClientKey            string
	CertificateAuthority string
	Controller           IngressController
	Service              Service
	EnableWebsockets     bool
	ControlledApexDomain string
	UpstreamDomains      []Domain
}

type Domain struct {
	Apex      string
	Subdomain string
	FQDN      string
}
