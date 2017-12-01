package plugins

import (
	"time"

	"github.com/codeamp/transistor"
)

func init() {
	transistor.RegisterApi(Project{})
	transistor.RegisterApi(GitCommit{})
	transistor.RegisterApi(GitStatus{})
	transistor.RegisterApi(GitSync{})
	transistor.RegisterApi(WebsocketMsg{})
	transistor.RegisterApi(HeartBeat{})
	transistor.RegisterApi(Project{})
	transistor.RegisterApi(Extension{})
	transistor.RegisterApi(Release{})
	transistor.RegisterApi(ReleaseExtension{})
}

type State string

const (
	Waiting  State = "waiting"
	Running        = "running"
	Fetching       = "fetching"
	Building       = "building"
	Pushing        = "pushing"
	Complete       = "complete"
	Failed         = "failed"
	Deleting       = "deleting"
	Deleted        = "deleted"
)

type Type string

const (
	File         Type = "file"
	Env               = "env"
	ProtectedEnv      = "protected-env"
	Build             = "build"
	Internal          = "internal"
	External          = "external"
	Office            = "office"
)

type ExtensionType string

const (
	Deployment   ExtensionType = "deployment"
	Workflow                   = "workflow"
	Notification               = "notification"
	Once                       = "once"
)

type EnvVarScope string

const (
	ProjectScope   EnvVarScope = "project"
	ExtensionScope             = "extension"
	GlobalScope                = "global"
)

type Action string

const (
	Create   Action = "create"
	Update          = "update"
	Destroy         = "destroy"
	Rollback        = "rollback"
	Status          = "status"
)

type Git struct {
	Url           string `json:"gitUrl"`
	Protocol      string `json:"protocol"`
	Branch        string `json:"branch"`
	Workdir       string `json:"workdir"`
	HeadHash      string `json:"headHash,omitempty"`
	RsaPrivateKey string `json:"rsaPrivateKey" role:"secret"`
	RsaPublicKey  string `json:"rsaPublicKey" role:"secret"`
}

type GitCommit struct {
	Repository string    `json:"repository"`
	User       string    `json:"user"`
	Message    string    `json:"message"`
	Ref        string    `json:"ref"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parentHash"`
	Created    time.Time `json:"created"`
}

type GitStatus struct {
	Repository string `json:"repository"`
	User       string `json:"user"`
	Hash       string `json:"hash"`
	State      string `json:"state"`
	Context    string `json:"context"`
}

type GitSync struct {
	Action       Action  `json:"action"`
	State        State   `json:"state"`
	StateMessage string  `json:"stateMessage"`
	Project      Project `json:"project"`
	Git          Git     `json:"git"`
	From         string  `json:"from"`
}

type Feature struct {
	Id         string    `json:"id"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parentHash"`
	User       string    `json:"user"`
	Message    string    `json:"message"`
	Created    time.Time `json:"created"`
}

type Listener struct {
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

type ListenerPair struct {
	Source      Listener `json:"source"`
	Destination Listener `json:"destination"`
}

type Service struct {
	Id           string      `json:"id"`
	Action       Action      `json:"action"`
	Name         string      `json:"name"`
	Command      string      `json:"command"`
	Listeners    []Listener  `json:"listeners"`
	Replicas     int64       `json:"replicas"`
	State        State       `json:"state"`
	StateMessage string      `json:"stateMessage"`
	Spec         ServiceSpec `json:"spec"`
	OneShot      bool        `json:"oneShot"`
}

type ServiceSpec struct {
	Id                            string `json:"id"`
	CpuRequest                    string `json:"cpuRequest"`
	CpuLimit                      string `json:"cpuLimit"`
	MemoryRequest                 string `json:"memoryRequest"`
	MemoryLimit                   string `json:"memoryLimit"`
	TerminationGracePeriodSeconds int64  `json:"terminationGracePeriodSeconds"`
}

type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value" role:"secret"`
	Type  Type   `json:"type"`
}

type Arg struct {
	Key   string `json:"key"`
	Value string `json:"value" role:"secret"`
}

type HeartBeat struct {
	Tick string `json:"tick"`
}

// LoadBalancer
type LoadBalancer struct {
	Action        Action         `json:"action"`
	State         State          `json:"state"`
	StateMessage  string         `json:"stateMessage"`
	Name          string         `json:"name"`
	Type          Type           `json:"type"`
	Project       Project        `json:"project"`
	Service       Service        `json:"service"`
	ListenerPairs []ListenerPair `json:"portPairs"`
	DNS           string         `json:"dns"`
	Environment   string         `json:"environment"`
	Subdomain     string         `json:"subdomain"`
}

// Route53
type Route53 struct {
	State        State   `json:"state"`
	StateMessage string  `json:"stateMessage"`
	Project      Project `json:"project"`
	Service      Service `json:"service"`
	DNS          string  `json:"dns"`
	FQDN         string  `json:"fqdn"`
	Environment  string  `json:"environment"`
	Subdomain    string  `json:"subdomain"`
}

type WebsocketMsg struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Payload interface{} `json:"data"`
}

type Extension struct {
	Id           string                 `json:"id"`
	Action       Action                 `json:"action"`
	Slug         string                 `json:"slug"`
	State        State                  `json:"state"`
	StateMessage string                 `json:"stateMessage"`
	FormValues   map[string]interface{} `json:"formValues"`
	Artifacts    map[string]string      `json:"artifacts"`
	Environment  string                 `json:"environment"`
	Project      Project                `json:"project"`
}

type Release struct {
	Id           string            `json:"id"`
	Action       Action            `json:"action"`
	State        State             `json:"state"`
	StateMessage string            `json:"stateMessage"`
	Project      Project           `json:"project"`
	Git          Git               `json:"git"`
	HeadFeature  Feature           `json:"headFeature"`
	User         string            `json:"user"`
	TailFeature  Feature           `json:"tailFeature"`
	Services     []Service         `json:"services"`
	Secrets      []Secret          `json:"secrets"` // secrets = build args + artifacts
	Artifacts    map[string]string `json:"artifacts"`
	Environment  string            `json:"environment"`
}

type ReleaseExtension struct {
	Id           string            `json:"id"`
	Action       Action            `json:"action"`
	Slug         string            `json:"slug"`
	Release      Release           `json:"release"`
	Extension    Extension         `json:"extension"`
	Artifacts    map[string]string `json:"artifacts"`
	State        State             `json:"state"`
	StateMessage string            `json:"stateMessage"`
}

type Project struct {
	Id             string            `json:"id"`
	Action         Action            `json:"action"`
	State          State             `json:"state"`
	StateMessage   string            `json:"stateMessage"`
	Git            Git               `json:"git"`
	Services       []Service         `json:"services"`
	Secrets        []Secret          `json:"secrets"` // secrets = build args + artifacts
	Artifacts      map[string]string `json:"artifacts"`
	Repository     string            `json:"repository"`
	NotifyChannels []string          `json:"notifyChannels,omitempty"`
}
