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
	}
}

func (x *Kubernetes) Process(e transistor.Event) error {
	log.InfoWithFields("Processing kubernetes event", log.Fields{
		"event": e,
	})

	switch e.Name {
	case "plugins.Release:complete":
		releaseEvent := e.Payload.(plugins.Release)
		valid := false
		kubeRe := plugins.ReleaseExtension{}

		// confirm this extension should be processed
		// by checking array of extensions and finding 'kubernetes' slug

		for _, re := range releaseEvent.ReleaseExtensions {
			slug := strings.Split(re.Slug, "|")
			if slug[0] == "kubernetes" {
				valid = true
				kubeRe = re
				break
			}
		}

		if valid {
			extensionSlugSlice := strings.Split(releaseEvent.Slug, "|")
			switch extensionSlugSlice[0] {
			case "dockerbuilder":
				// check if Release.Extension.ExtensionSpec.Name == "DockerBuilder"
				// doDeploy if it is

				time.Sleep(5 * time.Second)

				releaseRes := releaseEvent
				releaseRes.Action = plugins.Complete
				releaseRes.StateMessage = "Finished deployment"
				releaseRes.Slug = kubeRe.Slug

				x.events <- transistor.NewEvent(releaseRes, nil)
			default:
			}
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
