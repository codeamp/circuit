package kubernetes

import (
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	beta "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

//ProcessIngress Processes Kubernetes Ingress Events
func (x *Kubernetes) ProcessIngress(e transistor.Event) {
	log.Println("processing ingress")

	if e.Matches("project:kubernetes:ingress") {
		var err error
		switch e.Action {
		case transistor.GetAction("delete"):
			err = x.deleteIngress(e)
		case transistor.GetAction("create"):
			spew.Dump(e.Payload.(plugins.ProjectExtension))
			err = x.createIngress(e)
		case transistor.GetAction("update"):
			err = x.updateIngress(e)
		}

		if err != nil {
			log.Error(err)
			x.sendErrorResponse(e, err.Error())
		}
	}

	return
}

func (x *Kubernetes) deleteIngress(e transistor.Event) error {
	return nil
}

func (x *Kubernetes) createIngress(e transistor.Event) error {
	inputs, err := getInputs(e)
	if err != nil {
		return err
	}
	spew.Dump(inputs)

	payload := e.Payload.(plugins.ProjectExtension)

	kubeconfig, err := x.SetupKubeConfig(e)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{Timeout: "60"}).ClientConfig()

	if err != nil {
		failMessage := fmt.Sprintf("ERROR: %s; you must set the environment variable CF_PLUGINS_KUBEDEPLOY_KUBECONFIG=/path/to/kubeconfig", err.Error())
		log.Error(failMessage)
		return err
	}

	clientset, err := beta.NewForConfig(config)
	if err != nil {
		failMessage := fmt.Sprintf("ERROR: %s; setting NewForConfig in createIngress", err.Error())
		log.Error(failMessage)
		return err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	log.Println(clientset.Ingresses(x.GenNamespaceName(payload.Environment, projectSlug)))

	return nil
}

func (x *Kubernetes) updateIngress(e transistor.Event) error {
	return nil
}

func getInputs(e transistor.Event) (*IngressInput, error) {
	input := IngressInput{}
	var err error

	fqdn, err := e.GetArtifact("fqdn")
	if err != nil {
		return nil, err
	}
	input.FQDN = fqdn.String()

	kubeconfig, err := e.GetArtifact("kubeconfig")
	if err != nil {
		return nil, err
	}
	input.KubeConfig = kubeconfig.String()

	clientCertificate, err := e.GetArtifact("client_certificate")
	if err != nil {
		return nil, err
	}
	input.ClientCertificate = clientCertificate.String()

	clientKey, err := e.GetArtifact("client_key")
	if err != nil {
		return nil, err
	}
	input.ClientKey = clientKey.String()

	certificateAuthority, err := e.GetArtifact("certificate_authority")
	if err != nil {
		return nil, err
	}
	input.CertificateAuthority = certificateAuthority.String()

	ingressController, err := e.GetArtifact("ingress_controller")
	if err != nil {
		return nil, err
	}
	parsedController, err := parseController(ingressController.String())
	if err != nil {
		return nil, err
	}
	input.Controller = *parsedController

	return &input, nil

}

/*
parseController accepts a string delimited by the `:` character.
Each controller string must be in this format:
	<subdomain:ingress_controler_id:elb_dns>
*/
func parseController(ingressController string) (*IngressController, error) {

	parts := strings.Split(ingressController, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%s is an invalid IngressController string. Must be in format: <subdomain:ingress_controler_id:elb_dns>", ingressController)
	}

	controller := IngressController{
		Subdomain:      parts[0],
		ControllerName: parts[1],
		ControllerID:   parts[2],
	}

	return &controller, nil
}
