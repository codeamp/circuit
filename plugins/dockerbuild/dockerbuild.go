package dockerbuild

import (
	"strings"
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
		"plugins.Release:create",
	}
}

func (x *Dockerbuild) Process(e transistor.Event) error {

	log.InfoWithFields("Processing dockerbuild event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.Release:create" {

		releaseEvent := e.Payload.(plugins.Release)

		valid := false
		dbRe := plugins.ReleaseExtension{}
		// confirm this extension should be processed
		// by checking array of extensions and finding 'dockerbuilder' slug

		for _, re := range releaseEvent.ReleaseExtensions {
			slug := strings.Split(re.Slug, "|")
			if slug[0] == "dockerbuilder" {
				valid = true
				dbRe = re
				break
			}
		}

		if valid {
			logLine := "dockerbuild log line"

			time.Sleep(5 * time.Second)

			releaseRes := releaseEvent
			releaseRes.Action = plugins.Complete
			releaseRes.Artifacts = map[string]*string{
				"log": &logLine,
			}
			releaseRes.Slug = dbRe.Slug

			x.events <- transistor.NewEvent(releaseRes, nil)
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

		spew.Dump("Dockerbuild Release done -> sending out event", completeExtensionEvent)
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
