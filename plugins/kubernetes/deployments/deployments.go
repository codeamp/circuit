package kubernetesdeployments

import (
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type Deployments struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("kubernetesdeployments", func() transistor.Plugin {
		return &Deployments{}
	})
}

func (x *Deployments) Description() string {
	return "Deploy projects to Deployments"
}

func (x *Deployments) SampleConfig() string {
	return ` `
}

func (x *Deployments) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Deployments")

	return nil
}

func (x *Deployments) Stop() {
	log.Info("Stopping Deployments")
}

func (x *Deployments) Subscribe() []string {
	return []string{
		"plugins.ReleaseExtension:create:kubernetesdeployments",
		"plugins.Extension:create:kubernetesdeployments",
		"plugins.Extension:update:kubernetesdeployments",
	}
}

func (x *Deployments) Process(e transistor.Event) error {

	log.InfoWithFields("Processing Deployments event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.Extension:create:kubernetesdeployments" || e.Name == "plugins.Extension:update:kubernetesdeployments" {
		var extensionEvent plugins.Extension
		extensionEvent = e.Payload.(plugins.Extension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")
		x.events <- e.NewEvent(extensionEvent, nil)
		return nil
	}

	event := e.Payload.(plugins.ReleaseExtension)

	event.Action = plugins.GetAction("status")
	event.State = plugins.GetState("complete")
	event.StateMessage = "Completed"

	x.doDeploy(e)
	log.Info("Processed Deployments event")
	x.events <- e.NewEvent(event, nil)
	return nil
}
