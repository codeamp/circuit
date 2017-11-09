package kubernetes

import (
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type Kubernetes struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("kubernetes", func() transistor.Plugin {
		return &Kubernetes{}
	})
}

func (x *Kubernetes) Description() string {
	return "Deploy projects to Kubernetes"
}

func (x *Kubernetes) SampleConfig() string {
	return ` `
}

func (x *Kubernetes) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Kubernetes")

	return nil
}

func (x *Kubernetes) Stop() {
	log.Info("Stopping Kubernetes")
}

func (x *Kubernetes) Subscribe() []string {
	return []string{
		"plugins.Release:complete",
		"plugins.Extension:create",
		"plugins.ReleaseExtension:create",
	}
}

func (x *Kubernetes) Process(e transistor.Event) error {
	log.InfoWithFields("Processing kubernetes event", log.Fields{
		"event": e,
	})

	switch e.Name {
	case "plugins.ReleaseExtension:create":
		reEvent := e.Payload.(plugins.ReleaseExtension)

		if reEvent.Key == "kubernetes" {
			// doDeploy if it is

			time.Sleep(5 * time.Second)

			reRes := reEvent
			reRes.Action = plugins.Complete
			reRes.StateMessage = "Finished deployment"
			reRes.Slug = reRes.Slug

			x.events <- transistor.NewEvent(reRes, nil)
		}

	case "plugins.Extension:create":
		// check if extension slug is kubernetes
		extensionEvent := e.Payload.(plugins.Extension)
		extensionSlugSlice := strings.Split(extensionEvent.Slug, "|")
		switch extensionSlugSlice[0] {
		case "kubernetes":
			// create deployment
			// fill artifacts
			sampleKubeString := "checkrhq-dev.net"
			sampleKubeConfig := "/etc/secrets"

			extensionRes := extensionEvent
			extensionRes.Action = plugins.Complete
			extensionRes.StateMessage = "Finished initialization"
			extensionRes.Artifacts = map[string]*string{
				"cluster":     &sampleKubeString,
				"config path": &sampleKubeConfig,
			}

			x.events <- transistor.NewEvent(extensionRes, nil)
		default:
		}
	}

	log.Info("Processed kubernetes event")
	return nil
}
