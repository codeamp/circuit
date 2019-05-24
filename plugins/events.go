package plugins

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"
)

// GetEventName returns registered events.
func GetEventName(s string) transistor.EventName {
	eventNames := []string{
		"release:kubernetes:loadbalancer",
		"release:kubernetes:deployment",
		"release:dockerbuilder",
		"release:githubstatus",

		"project:githubstatus",
		"project:dockerbuilder",
		"project:database",
		"project:s3",
		"project:mongo",
		"project:scheduledbranchreleaser",
		"project:kubernetes:loadbalancer",
		"project:kubernetes:deployment",
		"project:kubernetes:redis",
		"project:heartbeat",

		"scheduledbranchreleaser:pulse",
		"scheduledbranchreleaser:scheduled",

		"gitsync",
		"gitsync:commit",
		"heartbeat",
		"release",

		"route53",
		"websocket",
		"slack",
		"slack:notify",
	}

	for _, t := range eventNames {
		if s == t {
			return transistor.EventName(t)
		}
	}

	errMsg := fmt.Sprintf("EventName not found: '%s' ", s)
	_, file, line, ok := runtime.Caller(1)
	if ok {
		errMsg += fmt.Sprintf("%s : ln %d", file, line)
	}

	log.Panic(errMsg)
	return transistor.EventName("unknown")
}

// Type is the plugins Type data structure
type Type string

// GetType returns registered plugin values of type Type
func GetType(s string) Type {
	types := []string{
		"file",
		"env",
		"protected-env",
		"build",
		"internal",
		"external",
		"office",
		"workflow",
		"notification",
		"once",
		"deployment",
		"general",
		"one-shot",
		"default",
		"recreate",
		"rollingUpdate",
		"livenessProbe",
		"readinessProbe",
	}

	for _, t := range types {
		if s == t {
			return Type(t)
		}
	}

	errMsg := fmt.Sprintf("Type not found: '%s' ", s)
	_, file, line, ok := runtime.Caller(1)
	if ok {
		errMsg += fmt.Sprintf("%s : ln %d", file, line)
	}

	log.Error(errMsg)
	return Type("unknown")
}

// Git event data struct
type Git struct {
	Url           string `json:"gitUrl"`
	Protocol      string `json:"protocol"`
	Branch        string `json:"branch"`
	Workdir       string `json:"workdir"`
	HeadHash      string `json:"headHash,omitempty"`
	RsaPrivateKey string `json:"rsaPrivateKey" role:"secret"`
	RsaPublicKey  string `json:"rsaPublicKey" role:"secret"`
}

// GitCommit event data struct
type GitCommit struct {
	Repository string    `json:"repository"`
	User       string    `json:"user"`
	Message    string    `json:"message"`
	Ref        string    `json:"ref"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parentHash"`
	Head       bool      `json:"head"`
	Created    time.Time `json:"created"`
}

// GitSync event data struct
type GitSync struct {
	Project Project     `json:"project"`
	Git     Git         `json:"git"`
	From    string      `json:"from"`
	Commits []GitCommit `json:"commits"`
}

// Feature event data struct
type Feature struct {
	ID         string    `json:"id"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parentHash"`
	User       string    `json:"user"`
	Message    string    `json:"message"`
	Created    time.Time `json:"created"`
}

// Listener event data struct
type Listener struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// ListenerPair event data struct
type ListenerPair struct {
	Source      Listener `json:"source"`
	Destination Listener `json:"destination"`
}

// Service event data struct
type Service struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Command   string      `json:"command"`
	Listeners []Listener  `json:"listeners"`
	Replicas  int64       `json:"replicas"`
	Spec      ServiceSpec `json:"spec"`
	Type      string      `json:"type"`

	State              transistor.State   `json:"state"`
	StateMessage       string             `json:"stateMessage"`
	Action             transistor.Action  `json:"action"`
	DeploymentStrategy DeploymentStrategy `json:"deploymentStrategy"`
	ReadinessProbe     ServiceHealthProbe `json:"readinessProbe"`
	LivenessProbe      ServiceHealthProbe `json:"livenessProbe"`
	PreStopHook        string             `json:"preStopHook"`
}

type DeploymentStrategy struct {
	Type           Type  `json:"type"`
	MaxUnavailable int32 `json:"maxUnavailable,string`
	MaxSurge       int32 `json:"MaxSurge,string`
}

type ServiceHealthProbe struct {
	// ServiceID
	ServiceID uuid.UUID `bson:"serviceID" json:"-" gorm:"type:uuid"`
	// Type: required; accepts `readinessProbe` and `livenessProbe`
	Type Type `json:"type"`
	// Method: required; accepts `exec`, `http`, and `tcp`
	Method string `json:"method"`
	// Command: Required with Method `exec`
	Command string `json:"command"`
	// Port: Required with Method `http` or `tcp`
	Port int32 `json:"port"`
	// Scheme: required with method `http`; accepts `http` or `https`
	Scheme string `json:"scheme"`
	// Path: required with Method `http`
	Path string `json:"path"`
	// InitialDelaySeconds is the delay before the probe begins to evaluate service health
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	// PeriodSeconds is how frequently the probe is executed
	PeriodSeconds int32 `json:"periodSeconds"`
	// TimeoutSeconds is the number of seconds before the probe times out
	TimeoutSeconds int32 `json:"timeoutSeconds"`
	// SuccessThreshold minimum consecutive success before the probe is considered successfull
	SuccessThreshold int32 `json:"successThreshold"`
	// FailureThreshold is the number of attempts before a probe is considered failed
	FailureThreshold int32 `json:"failureThreshold"`
	// HttpHeaders
	HttpHeaders []HealthProbeHttpHeader `json:"httpHeaders"`
}

type HealthProbeHttpHeader struct {
	// Name
	Name string `json:"name"`
	// Value
	Value string `json:"value"`
}

// ServiceSpec event data struct
type ServiceSpec struct {
	ID                            string `json:"id"`
	CpuRequest                    string `json:"cpuRequest"`
	CpuLimit                      string `json:"cpuLimit"`
	MemoryRequest                 string `json:"memoryRequest"`
	MemoryLimit                   string `json:"memoryLimit"`
	TerminationGracePeriodSeconds int64  `json:"terminationGracePeriodSeconds"`
}

// Secret event data struct
type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value" role:"secret"`
	Type  Type   `json:"type"`
}

// HeartBeat event data struct
type HeartBeat struct {
	Tick string `json:"tick"`
}

// ScheduledBranchReleaser data struct
type ScheduledBranchReleaser struct {
	ProjectExtension  `json:"projectextension"`
	Git               `json:"git"`
	ProjectSettingsID string `json:"projectSettingsID"`
}

// WebsocketMsg event data struct
type WebsocketMsg struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Payload interface{} `json:"data" role:"secret"`
}

// ReleaseExtension event data struct
type ReleaseExtension struct {
	ID          string  `json:"id"`
	Project     Project `json:"project"`
	Release     Release `json:"release"`
	Environment string  `json:"environment"`
}

// ProjectExtension event data struct
type ProjectExtension struct {
	ID          string  `json:"id"`
	Project     Project `json:"project"`
	Environment string  `json:"environment"`
}

// NotificationExtension event data struct
type NotificationExtension struct {
	ID          string  `json:"id"`
	Project     Project `json:"project"`
	Release     Release `json:"release"`
	Environment string  `json:"environment"`
}

// Release event data struct
type Release struct {
	ID          string    `json:"id"`
	Project     Project   `json:"project"`
	Git         Git       `json:"git"`
	HeadFeature Feature   `json:"headFeature"`
	User        string    `json:"user"`
	TailFeature Feature   `json:"tailFeature"`
	Services    []Service `json:"services"`
	Secrets     []Secret  `json:"secrets" role:"secret"`
	Environment string    `json:"environment"`
	IsRollback  bool      `json:"isRollback"`
}

// Project event data struct
type Project struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Repository string `json:"repository"`
}
