package dockerbuild

import (
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
)

type Dockerbuilder struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("dockerbuilder", func() transistor.Plugin {
		return &Dockerbuilder{}
	})
}

func (x *Dockerbuilder) Description() string {
	return "Build Docker images"
}

func (x *Dockerbuilder) SampleConfig() string {
	return ` `
}

func (x *Dockerbuilder) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Dockerbuild")
	return nil
}

func (x *Dockerbuilder) Stop() {
	log.Info("Stopping Dockerbuild")
}

func (x *Dockerbuilder) Subscribe() []string {
	return []string{
		"plugins.Extension:create",
		"plugins.Extension:update",
		"plugins.ReleaseExtension:create",
	}
}

func (x *Dockerbuilder) Process(e transistor.Event) error {

	log.InfoWithFields("Processing dockerbuilder event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.ReleaseExtension:create" {
		reEvent := e.Payload.(plugins.ReleaseExtension)

		if reEvent.Key == "dockerbuilder" {
			logLine := "dockerbuild log line"

			time.Sleep(5 * time.Second)

			reRes := reEvent
			reRes.Action = plugins.Complete
			reRes.Artifacts = map[string]*string{
				"log": &logLine,
			}

			spew.Dump("COMPLETED RE!", reRes)

			x.events <- transistor.NewEvent(reRes, nil)
		}
	}

	var extension plugins.Extension

	if e.Name == "plugins.Extension:update" {

		extension = e.Payload.(plugins.Extension)
		// create docker
		// fill artifacts

		// return complete event
		time.Sleep(5 * time.Second)

		dockerdata := "dockerdata2"

		completeExtensionEvent := plugins.Extension{
			Action:       plugins.Status,
			Slug:         extension.Slug,
			State:        plugins.Complete,
			StateMessage: "Complete",
			FormValues:   extension.FormValues,
			Artifacts: map[string]*string{
				"dockerdata": &dockerdata,
			},
		}

		x.events <- transistor.NewEvent(completeExtensionEvent, nil)
	}

	if e.Name == "plugins.Extension:create" {
		extension = e.Payload.(plugins.Extension)
		// check if extension is actually docker
		extensionSlugSlice := strings.Split(extension.Slug, "|")

		if extensionSlugSlice[0] == "dockerbuilder" {
			// create docker
			// fill artifacts

			// return complete event
			time.Sleep(5 * time.Second)

			dockerdata := "dockerdata"

			extensionRes := extension
			extensionRes.Action = plugins.Complete
			extensionRes.StateMessage = "Completed initialization"
			extensionRes.FormValues = extension.FormValues
			extensionRes.Artifacts = map[string]*string{
				"dockerdata": &dockerdata,
			}
			x.events <- transistor.NewEvent(extensionRes, nil)
		}
	}

	return nil
}
