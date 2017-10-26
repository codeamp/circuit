package plugins

import (
	"time"

	"github.com/codeamp/transistor"
)

func init() {
	transistor.RegisterApi(Project{})
	transistor.RegisterApi(GitPing{})
	transistor.RegisterApi(GitCommit{})
	transistor.RegisterApi(GitStatus{})
	transistor.RegisterApi(GitSync{})
	transistor.RegisterApi(Release{})
	transistor.RegisterApi(DockerBuild{})
	transistor.RegisterApi(DockerDeploy{})
	transistor.RegisterApi(LoadBalancer{})
	transistor.RegisterApi(WebsocketMsg{})
	transistor.RegisterApi(HeartBeat{})
	transistor.RegisterApi(Route53{})
	transistor.RegisterApi(Extension{})
	transistor.RegisterApi(ReleaseWorkflow{})
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

type Action string

const (
	Create   Action = "create"
	Update          = "update"
	Destroy         = "destroy"
	Rollback        = "rollback"
	Status          = "status"
)

type Project struct {
	Action         Action   `json:"action,omitempty"`
	Slug           string   `json:"slug"`
	Repository     string   `json:"repository"`
	NotifyChannels []string `json:"notifyChannels,omitempty"`
}

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

type GitPing struct {
	Repository string `json:"repository"`
	User       string `json:"user"`
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
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parentHash"`
	User       string    `json:"user"`
	Message    string    `json:"message"`
	Created    time.Time `json:"created"`
}

type Release struct {
	Id          string  `json:"id"`
	HeadFeature Feature `json:"headFeature"`
	TailFeature Feature `json:"tailFeature"`
	User        string  `json:"user"`
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
	CpuRequest                    string `json:"cpuRequest"`
	CpuLimit                      string `json:"cpuLimit"`
	MemoryRequest                 string `json:"memoryRequest"`
	MemoryLimit                   string `json:"memoryLimit"`
	TerminationGracePeriodSeconds int64  `json:"terminationGracePeriodSeconds"`
}

type DockerRegistry struct {
	Host     string `json:"host"`
	Org      string `json:"org"`
	Username string `json:"username"`
	Password string `json:"password" role:"secret"`
	Email    string `json:"email"`
}

type Docker struct {
	Image    string         `json:"image"`
	Registry DockerRegistry `json:"registry"`
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

type DockerBuild struct {
	Action       Action         `json:"action"`
	State        State          `json:"state"`
	StateMessage string         `json:"stateMessage"`
	Project      Project        `json:"project"`
	Git          Git            `json:"git"`
	Feature      Feature        `json:"feature"`
	Registry     DockerRegistry `json:"registry"`
	BuildArgs    []Arg          `json:"buildArgs"`
	BuildLog     string         `json:"buildLog"`
	BuildError   string         `json:"buildError"`
	Image        string         `json:"image"`
}

type HeartBeat struct {
	Tick string `json:"tick"`
}

// Deploy
type DockerDeploy struct {
	Action             Action    `json:"action"`
	State              State     `json:"state"`
	StateMessage       string    `json:"stateMessage"`
	Project            Project   `json:"project"`
	Release            Release   `json:"release"`
	Docker             Docker    `json:"docker"`
	Services           []Service `json:"services"`
	Secrets            []Secret  `json:"secrets"`
	Timeout            int       `json:"timeout"`
	DeploymentStrategy string    `json:"deploymentStrategy"`
	Environment        string    `json:"environment"`
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
	Action       Action `json:"action"`
	State        State  `json:"state"`
	StateMessage string `json:"stateMessage"`

	Slug       string             `json:"slug"`
	FormValues map[string]*string `json:"formValues"`
	Artifacts  map[string]*string `json:"artifacts"`
}

type ReleaseWorkflow struct {
	Action Action `json:"action"`
	Slug   string `json:"slug"`

	Project          Project            `json:"project"`
	Git              Git                `json:"git"`
	Release          Release            `json:"release"`
	ReleaseExtension ReleaseExtension   `json:"releaseExtension"`
	Services         []Service          `json:"services"`
	Secrets          []Secret           `json:"secrets"` // secrets = build args + artifacts
	Artifacts        map[string]*string `json:"artifacts"`

	State        State  `json:"state"`
	StateMessage string `json:"stateMessage"`
}

type ReleaseExtension struct {
	Id           string `json:"id"`
	State        State  `json:"state"`
	StateMessage string `json:"stateMessage"`
}
