package kubernetes

import (
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
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
		"plugins.DockerDeploy:create",
		"plugins.DockerDeploy:delete",
		"plugins.DockerDeploy:update",
		"plugins.LoadBalancer:create",
		"plugins.LoadBalancer.update",
		"plugins.LoadBalancer.destroy",
	}
}

func (x *Kubernetes) Process(e transistor.Event) error {
	log.Info("Processing kubernetes event")
	spew.Dump(e.Name, e.ID)

	switch e.Name {
	case "plugins.DockerDeploy:create":
		x.doDeploy(e)
	case "plugins.DockerDeploy:update":
		x.doDeploy(e)
	case "plugins.DockerDeploy:destroy":
		x.doDeploy(e)
	case "plugins.LoadBalancer:create":
		x.doLoadBalancer(e)
	case "plugins.LoadBalancer:update":
		x.doLoadBalancer(e)
	case "plugins.LoadBalancer:destroy":
		x.doDeleteLoadBalancer(e)
	}

	log.Info("Processed kubernetes event")
	spew.Dump(e.Name, e.ID)
	return nil
}
