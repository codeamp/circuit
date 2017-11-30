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
	}
}

func (x *Deployments) Process(e transistor.Event) error {

	log.InfoWithFields("Processing Deployments event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.Extension:create:kubernetesdeployments" {
		var extensionEvent plugins.Extension
		extensionEvent = e.Payload.(plugins.Extension)
		extensionEvent.Action = plugins.Complete
		x.events <- e.NewEvent(extensionEvent, nil)
		return nil
	}

	x.doDeploy(e)

	log.Info("Processed Deployments event")
	return nil
}
