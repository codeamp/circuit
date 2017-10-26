package dockerbuild

import (
	"time"

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
		"plugins.Extension:update",
		"plugins.ReleaseWorkflow:create",
	}
}

func (x *Dockerbuild) Process(e transistor.Event) error {

	log.InfoWithFields("Processing dockerbuild event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.ReleaseWorkflow:create" {

		rw := e.Payload.(plugins.ReleaseWorkflow)
		logLine := "dockerbuild log line"

		spew.Dump("DOCKERBUILD RELEASEWORKFLOW", rw)
		for i := 0; i < 5; i++ {
			time.Sleep(3 * time.Second)
			if i == 4 {
				logLine = "process complete"
			}
			releaseWorkflowRes := plugins.ReleaseWorkflow{
				Action: plugins.Status,
				Slug:   rw.Slug,
				Artifacts: map[string]*string{
					"log": &logLine,
				},
				Release: rw.Release,
				Project: rw.Project,
			}

			x.events <- transistor.NewEvent(releaseWorkflowRes, nil)
		}

		releaseWorkflowRes := plugins.ReleaseWorkflow{
			Action: plugins.Complete,
			Slug:   rw.Slug,
			Artifacts: map[string]*string{
				"log": &logLine,
			},
			Release: rw.Release,
			Project: rw.Project,
		}

		x.events <- transistor.NewEvent(releaseWorkflowRes, nil)

	}

	extension := e.Payload.(plugins.Extension)

	if extension.Action == plugins.Update {
		// create docker
		// fill artifacts

		// return complete event
		time.Sleep(5 * time.Second)

		hostname := *extension.FormValues["hostname"]
		credentials := *extension.FormValues["credentials"]
		dockerdata := "dockerdata2"

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
	}

	if extension.Action == plugins.Create {
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
	}

	return nil
}
