package plugins

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

func GetEventName(s string) transistor.EventName {
	eventNames := []string{
		"project:kubernetes:deployment",
		"release:kubernetes:deployment",
		"project:kubernetes:loadbalancer",
		"release:kubernetes:loadbalancer",
		"project:githubstatus",
		"release:githubstatus",
		"gitsync",
		"gitsync:commit",
		"heartbeat",
		"release",
		"project:dockerbuilder",
		"release:dockerbuilder",
		"route53",
		"websocket",
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

	errMsg := fmt.Sprintf("Type not found: '%s' ", s)
	_, file, line, ok := runtime.Caller(1)
	if ok {
		errMsg += fmt.Sprintf("%s : ln %d", file, line)
	}

	log.Panic(errMsg)
	return Type("unknown")
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
	Head       bool      `json:"head"`
	Created    time.Time `json:"created"`
}

type GitSync struct {
	Project Project     `json:"project"`
	Git     Git         `json:"git"`
	From    string      `json:"from"`
	Commits []GitCommit `json:"commits"`
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
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Command   string      `json:"command"`
	Listeners []Listener  `json:"listeners"`
	Replicas  int64       `json:"replicas"`
	Spec      ServiceSpec `json:"spec"`
	Type      string      `json:"type"`

	State        transistor.State  `json:"state"`
	StateMessage string            `json:"stateMessage"`
	Action       transistor.Action `json:"action"`
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

type HeartBeat struct {
	Tick string `json:"tick"`
}

type WebsocketMsg struct {
	Channel string      `json:"channel"`
	Event   string      `json:"event"`
	Payload interface{} `json:"data" role:"secret"`
}

type ReleaseExtension struct {
	ID          string  `json:"id"`
	Project     Project `json:"project"`
	Release     Release `json:"release"`
	Environment string  `json:"environment"`
}

type ProjectExtension struct {
	ID          string  `json:"id"`
	Project     Project `json:"project"`
	Environment string  `json:"environment"`
}

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
}

type Project struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Repository string `json:"repository"`
}
