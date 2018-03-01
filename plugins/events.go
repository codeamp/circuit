package plugins

import (
	"fmt"
	"time"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

func init() {
	transistor.RegisterEvent(Project{})
	transistor.RegisterEvent(GitCommit{})
	transistor.RegisterEvent(GitSync{})
	transistor.RegisterEvent(WebsocketMsg{})
	transistor.RegisterEvent(HeartBeat{})
	transistor.RegisterEvent(Release{})
	transistor.RegisterEvent(ProjectExtension{})
	transistor.RegisterEvent(ReleaseExtension{})
}

type State string

func GetState(s string) State {
	states := []string{
		"waiting",
		"running",
		"fetching",
		"building",
		"pushing",
		"complete",
		"failed",
		"deleting",
		"deleted",
	}

	for _, state := range states {
		if s == state {
			return State(state)
		}
	}

	log.Info(fmt.Sprintf("State not found: %s", s))

	return State("unknown")
}

type Type string

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
	}

	for _, t := range types {
		if s == t {
			return Type(t)
		}
	}

	log.Info(fmt.Sprintf("Type not found: %s", s))

	return Type("unknown")
}

type Action string

func GetAction(s string) Action {
	actions := []string{
		"create",
		"run",
		"update",
		"destroy",
		"rollback",
		"status",
	}

	for _, action := range actions {
		if s == action {
			return Action(action)
		}
	}

	log.Info(fmt.Sprintf("Action not found: %s", s))

	return Action("unknown")
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

type GitSync struct {
	Action       Action  `json:"action"`
	State        State   `json:"state"`
	StateMessage string  `json:"stateMessage"`
	Project      Project `json:"project"`
	Git          Git     `json:"git"`
	From         string  `json:"from"`
}

type Feature struct {
	ID         string    `json:"id"`
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
	ID           string      `json:"id"`
	Action       Action      `json:"action"`
	Name         string      `json:"name"`
	Command      string      `json:"command"`
	Listeners    []Listener  `json:"listeners"`
	Replicas     int64       `json:"replicas"`
	State        State       `json:"state"`
	StateMessage string      `json:"stateMessage"`
	Spec         ServiceSpec `json:"spec"`
	Type         string      `json:"type"`
}

type ServiceSpec struct {
	ID                            string `json:"id"`
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

type Artifact struct {
	Key   string `json:"key"`
	Value string `json:"value" role:"secret"`
}

type HeartBeat struct {
	Tick string `json:"tick"`
}

type WebsocketMsg struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Payload interface{} `json:"data"`
}

type ReleaseExtension struct {
	ID           string                 `json:"id"`
	Action       Action                 `json:"action"`
	Slug         string                 `json:"slug"`
	State        State                  `json:"state"`
	StateMessage string                 `json:"stateMessage"`
	Project      Project                `json:"project"`
	Release      Release                `json:"release"`
	Secrets      map[string]string      `json:"secrets"`
	Config       map[string]interface{} `json:"config"`
	Artifacts    map[string]interface{} `json:"artifacts"`
	Environment  string                 `json:"environment"`
}

type ProjectExtension struct {
	ID           string                 `json:"id"`
	Action       Action                 `json:"action"`
	Slug         string                 `json:"slug"`
	State        State                  `json:"state"`
	StateMessage string                 `json:"stateMessage"`
	Project      Project                `json:"project"`
	Secrets      map[string]string      `json:"secrets"`
	Config       map[string]interface{} `json:"config"`
	Artifacts    map[string]interface{} `json:"artifacts"`
	Environment  string                 `json:"environment"`
}

type Release struct {
	ID           string                 `json:"id"`
	Action       Action                 `json:"action"`
	State        State                  `json:"state"`
	StateMessage string                 `json:"stateMessage"`
	Project      Project                `json:"project"`
	Git          Git                    `json:"git"`
	HeadFeature  Feature                `json:"headFeature"`
	User         string                 `json:"user"`
	TailFeature  Feature                `json:"tailFeature"`
	Services     []Service              `json:"services"`
	Secrets      []Secret               `json:"secrets"`
	Artifacts    map[string]interface{} `json:"artifacts"`
	Environment  string                 `json:"environment"`
}

type Project struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Repository string `json:"repository"`
}
