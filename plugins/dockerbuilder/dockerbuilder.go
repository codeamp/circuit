package dockerbuilder

import (
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type DockerBuilder struct {
	events chan transistor.Event
	Socket string
}

func init() {
	transistor.RegisterPlugin("dockerbuilder", func() transistor.Plugin {
		return &DockerBuilder{Socket: "unix:///var/run/docker.sock"}
	})
}

func (x *DockerBuilder) Description() string {
	return "Clone git repository and build a docker image"
}

func (x *DockerBuilder) SampleConfig() string {
	return ` `
}

func (x *DockerBuilder) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started DockerBuilder")

	return nil
}

func (x *DockerBuilder) Stop() {
	log.Println("Stopping DockerBuilder")
}

func (x *DockerBuilder) Subscribe() []string {
	return []string{
		"plugins.ReleaseExtension:create:dockerbuilder",
		"plugins.Extension:create",
	}
}

func (x *DockerBuilder) Process(e transistor.Event) error {
	log.InfoWithFields("Process DockerBuilder event", log.Fields{
		"event": e.Name,
	})

	// var err error
	if e.Name == "plugins.ReleaseExtension:create:dockerbuilder" {
		event := e.Payload.(plugins.ReleaseExtension)
		event.Action = plugins.Complete
		event.State = plugins.Complete
		event.Artifacts["IMAGE"] = "dockerhub.io/it-werks"
		event.Artifacts["image1"] = "dockerhub.io/it-werks-1"
		x.events <- e.NewEvent(event, nil)
	}

	if e.Name == "plugins.Extension:create:dockerbuilder" {
		event := e.Payload.(plugins.Extension)
		event.State = plugins.Complete
		event.Action = plugins.Complete
		x.events <- e.NewEvent(event, nil)
	}

	return nil
}
