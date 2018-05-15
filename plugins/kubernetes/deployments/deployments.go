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
		"plugins.ProjectExtension:create:kubernetesdeployments",
		"plugins.ProjectExtension:update:kubernetesdeployments",
	}
}

func (x *Deployments) Process(e transistor.Event) error {

	log.InfoWithFields("Processing Deployments event", log.Fields{
		"event": e,
	})

	if e.Name == "plugins.ProjectExtension:create:kubernetesdeployments" {
		var extensionEvent plugins.ProjectExtension
		extensionEvent = e.Payload.(plugins.ProjectExtension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")
		x.events <- e.NewEvent(extensionEvent, nil)
		return nil
	}

	if e.Name == "plugins.ProjectExtension:update:kubernetesdeployments" {
		var extensionEvent plugins.ProjectExtension
		extensionEvent = e.Payload.(plugins.ProjectExtension)
		extensionEvent.Action = plugins.GetAction("status")
		extensionEvent.State = plugins.GetState("complete")
		x.events <- e.NewEvent(extensionEvent, nil)
		return nil
	}

	if e.Name == "plugins.ReleaseExtension:create:kubernetesdeployments" {
		event := e.Payload.(plugins.ReleaseExtension)

		err := x.doDeploy(e)
		if err != nil {
			event.Action = plugins.GetAction("status")
			event.State = plugins.GetState("failed")
			event.StateMessage = err.Error()
			x.events <- e.NewEvent(event, nil)
			return err
		}
	}

	return nil
}