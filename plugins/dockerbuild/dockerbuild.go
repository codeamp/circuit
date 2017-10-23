package dockerbuild

import (
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
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
		"plugins.Extension:create",
	}
}

func (x *Dockerbuild) Process(e transistor.Event) error {
	spew.Dump("DOCKER BUILD")
	spew.Dump(e)
	log.InfoWithFields("Processing dockerbuild event", log.Fields{
		"event": e,
	})

	extension := e.Payload.(plugins.Extension)

	// create docker
	// fill artifacts

	// return complete event
	completeExtensionEvent := plugins.Extension{
		Action:       plugins.Status,
		Slug:         extension.Slug,
		State:        plugins.Complete,
		StateMessage: "Complete",
		FormValues:   extension.FormValues,
		Artifacts:    map[string]string{},
	}

	spew.Dump("COMPLETE EXTENSION!")
	spew.Dump(completeExtensionEvent)
	x.events <- transistor.NewEvent(completeExtensionEvent, nil)

	return nil
}
