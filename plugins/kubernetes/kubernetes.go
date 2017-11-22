package kubernetes

import (
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
	case "plugins.ReleaseExtension:create:kubernetes":
		reEvent := e.Payload.(plugins.ReleaseExtension)
		time.Sleep(5 * time.Second)

		reRes := reEvent
		reRes.Artifacts["HELLO1"] = "world"
		reRes.Artifacts["hello2"] = "WORLD"
		reRes.Action = plugins.Complete
		reRes.StateMessage = "Finished deployment"
		x.events <- transistor.NewEvent(reRes, nil)

	case "plugins.Extension:create:kubernetes":
		extensionEvent := e.Payload.(plugins.Extension)
		// create deployment
		// fill artifacts
		sampleKubeString := "checkrhq-dev.net"
		sampleKubeConfig := "/etc/secrets"

		extensionRes := extensionEvent
		extensionRes.Action = plugins.Complete
		extensionRes.StateMessage = "Finished initialization"
		extensionRes.Artifacts = map[string]string{
			"cluster":     sampleKubeString,
			"config path": sampleKubeConfig,
		}

		x.events <- transistor.NewEvent(extensionRes, nil)
	}

	log.Info("Processed kubernetes event")
	return nil
}
