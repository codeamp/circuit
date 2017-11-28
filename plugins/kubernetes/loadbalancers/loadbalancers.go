package kubernetesloadbalancers

import (
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
)

func (x *Kubernetes) doLoadBalancer(e transistor.Event) error {
	log.Println("doLoadBalancer")

	spew.Dump(e)

	return nil
}

func (x *Kubernetes) doDeleteLoadBalancer(e transistor.Event) error {
	log.Println("doDeleteLoadBalancer")

	spew.Dump(e)

	return nil
}
