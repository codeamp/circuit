package kubernetes

import (
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
			x.sendErrorResponse(e, err.Error())
		}
	}

	return
}

func (x *Kubernetes) deleteIngress(e transistor.Event) error {
	log.Info("deleteIngress")
	var err error
	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.SetupClientset(e)
	if err != nil {
		return err
	}

	ingType, err := e.GetArtifact("type")
	if err != nil {
		return err
	}

	service, err := ParseService(e)
	if err != nil {
		return err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	coreInterface := clientset.Core()
	namespace := x.GenNamespaceName(payload.Environment, projectSlug)

	// Delete Service
	_, svcGetErr := coreInterface.Services(namespace).Get(service.ID, metav1.GetOptions{})
	if svcGetErr == nil {
		// Service was found, ready to delete
		svcDeleteErr := coreInterface.Services(namespace).Delete(service.ID, &metav1.DeleteOptions{})
		if svcDeleteErr != nil {
			return fmt.Errorf("Error managing loadbalancer '%s' deleting service %s", service.ID, svcDeleteErr)
		}
	} else {
		// Send failure message that we couldn't find the service to delete
		return fmt.Errorf("Error managing loadbalancer finding %s service: '%s'", service.ID, svcGetErr)
	}

	if ingType.String() == "loadbalancer" {
		//Delete Ingress
		networkInterface := clientset.ExtensionsV1beta1()
		ingresses := networkInterface.Ingresses(namespace)
		_, err = ingresses.Get(service.ID, metav1.GetOptions{})
		if err == nil {
			// ingress found, ready to delete
			ingressDeleteErr := ingresses.Delete(service.ID, &metav1.DeleteOptions{})
			if ingressDeleteErr != nil {
				return fmt.Errorf("Error managing ingress '%s' deleting service %s", service.ID, ingressDeleteErr)
			}
		} else {
			// Send failure message that we couldn't find the service to delete
			return fmt.Errorf("Error managing ingress finding %s service: '%s'", service.ID, svcGetErr)
		}
	}

	return nil

}

func (x *Kubernetes) createIngress(e transistor.Event) error {
	var artifacts []transistor.Artifact

	inputs, err := getIngressInputs(e)
	if err != nil {
		return err
	}

	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.SetupClientset(e)
	if err != nil {
		return err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	coreInterface := clientset.Core()
	deploymentName := x.GenDeploymentName(projectSlug, inputs.Service.Name)

	// var servicePorts []v1.ServicePort
	namespace := x.GenNamespaceName(payload.Environment, projectSlug)
	createNamespaceErr := x.CreateNamespaceIfNotExists(namespace, coreInterface)
	if createNamespaceErr != nil {
		return createNamespaceErr
	}

	servicePort := v1.ServicePort{
		Name: inputs.Service.Port.Name,
		Port: inputs.Service.Port.SourcePort,
		TargetPort: intstr.IntOrString{
			IntVal: inputs.Service.Port.TargetPort,
		},
		Protocol: v1.Protocol(inputs.Service.Port.Protocol),
	}

	serviceSpec := v1.ServiceSpec{
		Selector: map[string]string{"app": deploymentName},
		Type:     v1.ServiceTypeClusterIP,
		Ports:    []v1.ServicePort{servicePort},
	}

	serviceParams := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: inputs.Service.ID,
		},
		Spec: serviceSpec,
	}

	service := coreInterface.Services(namespace)
	svc, err := service.Get(inputs.Service.ID, metav1.GetOptions{})
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
			return fmt.Errorf("Error: failed to update service: %s", err.Error())
		}
		log.Debug(fmt.Sprintf("Service updated: %s", inputs.Service.ID))
	case k8s_errors.IsNotFound(err):
		serviceObj, err = service.Create(&serviceParams)
		if err != nil {
			return fmt.Errorf("Error: failed to create service: %s", err.Error())
		}
		log.Debug(fmt.Errorf("Service created: %s", inputs.Service.ID))
	default:
		return fmt.Errorf("Unexpected error: %s", err.Error())
	}

	if inputs.Type == "loadbalancer" {

		// check duplicates
		isDuplicate, err := x.isDuplicateIngressHost(e)
		if err != nil || isDuplicate {
			return err
		}

		networkInterface := clientset.ExtensionsV1beta1()
		ingresses := networkInterface.Ingresses(namespace)

		ingressRuleValue := v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{
					{
						Backend: v1beta1.IngressBackend{
							ServiceName: inputs.Service.ID,
							ServicePort: intstr.IntOrString{
								IntVal: inputs.Service.Port.SourcePort,
							},
						},
					},
				},
			},
		}

		var rules []v1beta1.IngressRule

		// Build for Upstream Domains
		for _, domain := range inputs.UpstreamFQDNs {
			rule := v1beta1.IngressRule{
				Host:             domain.FQDN,
				IngressRuleValue: ingressRuleValue,
			}

			rules = append(rules, rule)
		}

		ingressSpec := v1beta1.IngressSpec{
			Rules: rules,
		}

		ingressConfig := v1beta1.Ingress{
			TypeMeta: metav1.TypeMeta{

				Kind:       "Ingress",
				APIVersion: "extensions/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: inputs.Service.ID,
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": inputs.Controller.ControllerID,
				},
			},
			Spec: ingressSpec,
		}

		_, err = ingresses.Get(inputs.Service.ID, metav1.GetOptions{})
		var nIng *v1beta1.Ingress
		switch {
		case err == nil:
			nIng, err = ingresses.Update(&ingressConfig)
			if err != nil {
				return fmt.Errorf("Error: failed to update ingress: %s", err.Error())
			}
		case k8s_errors.IsNotFound(err):
			nIng, err = ingresses.Create(&ingressConfig)
			if err != nil {
				return fmt.Errorf("Error: failed to create ingress: %s", err.Error())
			}

		default:
			return fmt.Errorf("Unexpected error: %s", err.Error())
		}

		artifacts = append(artifacts, transistor.Artifact{Key: "ingress_id", Value: nIng.CreationTimestamp, Secret: false})

	}

	artifacts = append(artifacts, transistor.Artifact{Key: "ingress_controller", Value: inputs.Controller.ControllerName, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "elb_dns", Value: inputs.Controller.ELB, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "name", Value: inputs.Service.Name, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "controlled_apex_domain", Value: inputs.ControlledApexDomain, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "cluster_ip", Value: serviceObj.Spec.ClusterIP, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "internal_dns", Value: fmt.Sprintf("%s.%s", inputs.Service.ID, namespace), Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "table_view", Value: getTableViewFromDomains(inputs.UpstreamFQDNs), Secret: false})

	x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)

	return nil
}

func (x *Kubernetes) isDuplicateIngressHost(e transistor.Event) (bool, error) {
	inputs, err := getIngressInputs(e)
	if err != nil {
		return false, err
	}

	clientset, err := x.SetupClientset(e)
	if err != nil {
		return false, err
	}

	networkInterface := clientset.ExtensionsV1beta1()
	ingresses := networkInterface.Ingresses("")

	existingIngresses, err := ingresses.List(metav1.ListOptions{})

	//TODO Rewrite lookups using hashtable

	// check for duplicate secondary upstream fqdns
	for _, ingress := range existingIngresses.Items {
		for _, rule := range ingress.Spec.Rules {
			for _, domain := range inputs.UpstreamFQDNs {
				if rule.Host == domain.FQDN && ingress.GetName() != inputs.Service.ID {
					return true, fmt.Errorf("Error: An ingress for Upstream Domain %s already configured. Namespace: %s", domain, ingress.GetNamespace())
				}
			}
		}
	}

	return false, nil

}

func getIngressInputs(e transistor.Event) (*IngressInput, error) {
	input := IngressInput{}
	var err error

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

	serviceType, err := e.GetArtifact("type")
	if err != nil {
		return nil, err
	}
	input.Type = serviceType.String()

	service, err := ParseService(e)
	if err != nil {
		return nil, err
	}

	input.Service = service

	if serviceType.String() == "loadbalancer" {

		apexDomain, err := e.GetArtifact("controlled_apex_domain")
		if err != nil {
			return nil, err
		}
		input.ControlledApexDomain = apexDomain.String()

		upstreamDomains, err := e.GetArtifact("upstream_domains")
		if err != nil {
			return nil, err
		}

		input.UpstreamFQDNs, err = parseUpstreamDomains(upstreamDomains)
		if err != nil {
			return nil, err
		}

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

	}

	return &input, nil

}
