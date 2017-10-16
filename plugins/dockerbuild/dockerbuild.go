package dockerbuild

import (
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type Dockerbuild struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("dockerbuild", func() transistor.Plugin {
		return &Dockerbuild{}
	})
}

func (x *Dockerbuild) Description() string {
	return "Build Docker images"
}

func (x *Dockerbuild) SampleConfig() string {
	return ` `
}

func (x *Dockerbuild) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Dockerbuild")
	return nil
}

func (x *Dockerbuild) Stop() {
	log.Info("Stopping Dockerbuild")
}

func (x *Dockerbuild) Subscribe() []string {
	return []string{
		"plugins.DockerBuild:create",
	}
}

func (x *Dockerbuild) Process(e transistor.Event) error {
	log.InfoWithFields("Processing dockerbuild event", log.Fields{
		"event": e,
	})
	return nil
}
