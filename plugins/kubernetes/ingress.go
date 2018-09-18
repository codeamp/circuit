package kubernetes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
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
			err = x.createIngress(e)
		case transistor.GetAction("update"):
			err = x.createIngress(e)
		}

		if err != nil {
			log.Error(err)
			spew.Dump(err)
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

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		failMessage := fmt.Sprintf("ERROR: %s; setting NewForConfig in createIngress", err.Error())
		log.Error(failMessage)
		return err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	coreInterface := clientset.Core()
	deploymentName := x.GenDeploymentName(projectSlug, inputs.Service)

	// var servicePorts []v1.ServicePort
	namespace := x.GenNamespaceName(payload.Environment, projectSlug)
	createNamespaceErr := x.CreateNamespaceIfNotExists(namespace, coreInterface)
	if createNamespaceErr != nil {
		return createNamespaceErr
	}

	// service, err := createService(*inputs, coreInterface)
	// if err != nil {
	// 	return err
	// }

	// spew.Dump(service)

	servicePort := v1.ServicePort{
		Name: inputs.Port.Name,
		Port: inputs.Port.SourcePort,
		TargetPort: intstr.IntOrString{
			IntVal: inputs.Port.TargetPort,
		},
		Protocol: v1.Protocol(inputs.Port.Protocol),
	}

	serviceSpec := v1.ServiceSpec{
		Selector: map[string]string{"app": deploymentName},
		Type:     v1.ServiceTypeLoadBalancer,
		Ports:    []v1.ServicePort{servicePort},
	}

	serviceParams := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: inputs.Service,
		},
		Spec: serviceSpec,
	}

	service := coreInterface.Services(namespace)
	svc, err := service.Get(inputs.Service, metav1.GetOptions{})
	var serviceObj *v1.Service
	switch {
	case err == nil:
		// Preserve the NodePorts for PATCH service.
		if svc.Spec.Type == "LoadBalancer" {
			for i := range svc.Spec.Ports {
				for j := range serviceParams.Spec.Ports {
					if strings.ToLower(svc.Spec.Ports[i].Name) == strings.ToLower(serviceParams.Spec.Ports[j].Name) {
						serviceParams.Spec.Ports[j].NodePort = svc.Spec.Ports[i].NodePort
					}
				}
			}
		}
		serviceParams.ObjectMeta.ResourceVersion = svc.ObjectMeta.ResourceVersion
		serviceParams.Spec.ClusterIP = svc.Spec.ClusterIP
		serviceObj, err = service.Update(&serviceParams)
		if err != nil {
			return errors.New(fmt.Sprintf("Error: failed to update service: %s", err.Error()))
		}
		log.Debug(fmt.Sprintf("Service updated: %s", inputs.Service))
	case k8s_errors.IsNotFound(err):
		serviceObj, err = service.Create(&serviceParams)
		if err != nil {
			return errors.New(fmt.Sprintf("Error: failed to create service: %s", err.Error()))
		}
		log.Debug(fmt.Sprintf("Service created: %s", inputs.Service))
	default:
		return errors.New(fmt.Sprintf("Unexpected error: %s", err.Error()))
	}

	//TODO: hook up ingress
	networkInterface := clientset.ExtensionsV1beta1()
	ingresses := networkInterface.Ingresses(namespace)

	ingressSpec := v1beta1.IngressSpec{
		Backend: &v1beta1.IngressBackend{
			ServiceName: inputs.Service,
			ServicePort: intstr.IntOrString{
				IntVal: inputs.Port.SourcePort,
			},
		},
		Rules: []v1beta1.IngressRule{
			v1beta1.IngressRule{
				Host: fmt.Sprintf("%s.%s", inputs.Subdomain, inputs.FQDN),
				// IngressRuleValue: v1beta1.HTTPIngressRuleValue,
			},
		},
	}

	ingressConfig := v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{

			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: inputs.Service,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": inputs.Controller.ControllerID,
			},
		},
		Spec: ingressSpec,
	}

	ing, err := ingresses.Create(&ingressConfig)
	if err != nil {
		return err
	}
	spew.Dump(ing.Namespace)

	artifacts := make([]transistor.Artifact, 6, 6)
	artifacts[0] = transistor.Artifact{Key: "elb", Value: inputs.Controller.ELB, Secret: false}
	artifacts[1] = transistor.Artifact{Key: "name", Value: inputs.Service, Secret: false}
	artifacts[2] = transistor.Artifact{Key: "fqdn", Value: fmt.Sprintf("%s.%s", inputs.Subdomain, inputs.FQDN), Secret: false}
	artifacts[3] = transistor.Artifact{Key: "cluster_ip", Value: serviceObj.Spec.ClusterIP, Secret: false}
	artifacts[4] = transistor.Artifact{Key: "internal_dns", Value: fmt.Sprintf("%s.%s", inputs.Port.Name, namespace), Secret: false}
	artifacts[5] = transistor.Artifact{Key: "ingress_id", Value: ing.CreationTimestamp, Secret: false}

	x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)

	// log.Println(clientset.Ingresses(x.GenNamespaceName(payload.Environment, projectSlug)))

	return nil
}

// func createService(input IngressInput, coreInterface corev1.CoreV1Interface) (*v1.ServicePort, error) {

// 	return nil, nil

// }

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

	subdomain, err := e.GetArtifact("subdomain")
	if err != nil {
		return nil, err
	}
	input.Subdomain = subdomain.String()

	service, err := e.GetArtifact("service")
	if err != nil {
		return nil, err
	}
	serviceParts := strings.Split(service.String(), ":")
	input.Service = serviceParts[0]
	portInt, err := strconv.Atoi(serviceParts[1])
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("%s: Invalid Port type", serviceParts[1]))
	}

	port := ListenerPair{
		Name:       fmt.Sprintf("http-%s-%.0f", input.Service, float64(portInt)),
		Protocol:   "TCP",
		SourcePort: int32(portInt),
		TargetPort: int32(portInt),
	}
	input.Port = port

	selectedIngress, err := e.GetArtifact("ingress")
	if err != nil {
		return nil, err
	}

	ingressControllers, err := e.GetArtifact("ingress_controllers")
	if err != nil {
		return nil, err
	}

	// Guarantee persisted ingress controller is configured on the extension side.
	found := false
	for _, controller := range strings.Split(ingressControllers.String(), ",") {
		if controller == selectedIngress.String() {
			found = true
		}
		continue
	}
	if found == false {
		return nil, fmt.Errorf("Selected Ingress Controller is Not Configured")
	}

	parsedController, err := parseController(selectedIngress.String())
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
		return nil, fmt.Errorf("%s is an invalid IngressController string. Must be in format: <ingress_name:ingress_controller_id:elb_dns>", ingressController)
	}

	controller := IngressController{
		ControllerName: parts[0],
		ControllerID:   parts[1],
		ELB:            parts[2],
	}

	return &controller, nil
}
