package dockerbuild

import (
	"time"

	"github.com/codeamp/circuit/plugins"
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
		"plugins.Extension:create",
	}
}

func (x *Dockerbuild) Process(e transistor.Event) error {
	log.InfoWithFields("Processing dockerbuild event", log.Fields{
		"event": e,
	})

	extension := e.Payload.(plugins.Extension)

	// create docker
	// fill artifacts

	// return complete event
	time.Sleep(5 * time.Second)

	hostname := *extension.FormValues["hostname"]
	credentials := *extension.FormValues["credentials"]
	dockerdata := "dockerdata"

	completeExtensionEvent := plugins.Extension{
		Action:       plugins.Status,
		Slug:         extension.Slug,
		State:        plugins.Complete,
		StateMessage: "Complete",
		FormValues:   extension.FormValues,
		Artifacts: map[string]*string{
			"hostname":    &hostname,
			"credentials": &credentials,
			"dockerdata":  &dockerdata,
		},
	}

	x.events <- transistor.NewEvent(completeExtensionEvent, nil)

	return nil
}
