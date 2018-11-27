package kubernetes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	contour_v1beta1 "github.com/heptio/contour/apis/contour/v1beta1"
	"k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//ProcessIngress Processes Kubernetes Ingress Events
func (x *Kubernetes) ProcessIngressRoute(e transistor.Event) {
	log.Println("processing IngressRoute")

	if e.Matches("project:kubernetes:ingressroute") {
		var err error
		switch e.Action {
		case transistor.GetAction("delete"):
			err = x.deleteIngressRoute(e)
		case transistor.GetAction("create"):
			err = x.createIngressRoute(e)
		case transistor.GetAction("update"):
			err = x.createIngressRoute(e)
		}

		if err != nil {
			log.Error(err)
			x.sendErrorResponse(e, err.Error())
		}
	}

	return
}

func (x *Kubernetes) isDuplicateIngressRoute(e transistor.Event) (bool, error) {
	inputs, err := getIngressRouteInputs(e)
	if err != nil {
		return false, err
	}

	contourClient, err := x.getContourClient(e)
	if err != nil {
		return false, err
	}

	existingRoutes, err := contourClient.ContourV1beta1().IngressRoutes("").List(metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	// TODO Rewrite lookups using hashtable
	// check for duplicate secondary upstream fqdns
	for _, route := range existingRoutes.Items {
		for _, domain := range inputs.UpstreamDomains {
			if route.Spec.VirtualHost.Fqdn == domain.FQDN && route.GetName() != x.getIngressRouteID(domain, *inputs) {
				return true, fmt.Errorf("Error: An ingressroute for Upstream Domain %s already configured. Namespace: %s", domain.FQDN, route.GetNamespace())
			}
		}
	}

	return false, nil

}

func getIngressRouteInputs(e transistor.Event) (*IngressRouteInput, error) {
	input := IngressRouteInput{}
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

	service, err := parseService(e)
	if err != nil {
		return nil, err
	}

	input.Service = service

	if serviceType.String() == "loadbalancer" {
		isWebsocket := false
		enableWebsockets, err := e.GetArtifact("enable_websockets")
		if err != nil {
			log.DebugWithFields("enable_websockets property not found, defaulting to false", log.Fields{
				"id":      e.ID,
				"message": err.Error(),
			})
		} else {
			isWebsocket, err = strconv.ParseBool(enableWebsockets.String())
			if err != nil {
				return nil, fmt.Errorf("enable_websockets property is malformed: %s", err.Error())
			}
		}
		input.EnableWebsockets = isWebsocket

		apexDomain, err := e.GetArtifact("controlled_apex_domain")
		if err != nil {
			return nil, err
		}
		input.ControlledApexDomain = apexDomain.String()

		upstreamDomains, err := e.GetArtifact("upstream_domains")
		if err != nil {
			return nil, err
		}

		input.UpstreamDomains = parseUpstreamDomains(upstreamDomains)

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

func (x *Kubernetes) getIngressRouteID(domain Domain, extensionInput IngressRouteInput) string {
	return fmt.Sprintf("%s-%s-%s", extensionInput.Service.ID, domain.Subdomain, domain.Apex)
}

func (x *Kubernetes) deleteIngressRoute(e transistor.Event) error {
	log.Info("deleteIngress")
	var err error
	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.getKubernetesClient(e)
	if err != nil {
		return err
	}

	ingType, err := e.GetArtifact("type")
	if err != nil {
		return err
	}

	service, err := parseService(e)
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
		inputs, err := getIngressRouteInputs(e)
		if err != nil {
			return err
		}

		contourClient, err := x.getContourClient(e)
		if err != nil {
			return err
		}
		ingressClient := contourClient.ContourV1beta1().IngressRoutes(namespace)

		for _, domain := range inputs.UpstreamDomains {
			ingID := x.getIngressRouteID(domain, *inputs)
			_, err = ingressClient.Get(ingID, metav1.GetOptions{})
			if err == nil {
				ingressDeleteErr := ingressClient.Delete(ingID, &metav1.DeleteOptions{})
				if ingressDeleteErr != nil {
					return fmt.Errorf("Error managing ingressroute '%s' deleting ingressroute %s", ingID, ingressDeleteErr)
				}
			} else {
				return fmt.Errorf("Error managing ingressroute finding %s ingressroute: '%s'", ingID, err)
			}
		}
	}

	return nil

}

func (x *Kubernetes) createIngressRoute(e transistor.Event) error {
	var artifacts []transistor.Artifact

	inputs, err := getIngressRouteInputs(e)
	if err != nil {
		return err
	}

	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.getKubernetesClient(e)
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
		isDuplicate, err := x.isDuplicateIngressRoute(e)
		if err != nil || isDuplicate {
			return err
		}

		contourClient, err := x.getContourClient(e)
		if err != nil {
			return err
		}

		ingressClient := contourClient.ContourV1beta1().IngressRoutes(namespace)

		// cleanup removed ingresses
		currentIngresses, err := ingressClient.List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		var cleanupList []string
		for _, ingress := range currentIngresses.Items {
			found := false
			needle := ingress.GetName()
			if !strings.HasPrefix(needle, inputs.Service.ID) {
				continue
			}
			for _, domain := range inputs.UpstreamDomains {
				if x.getIngressRouteID(domain, *inputs) == needle {
					found = true

				}
			}
			if !found {
				cleanupList = append(cleanupList, needle)
			}
		}

		for _, ingress := range cleanupList {
			err := ingressClient.Delete(ingress, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}

		for _, domain := range inputs.UpstreamDomains {
			spec := contour_v1beta1.IngressRouteSpec{
				VirtualHost: &contour_v1beta1.VirtualHost{
					Fqdn: domain.FQDN,
				},
				Routes: []contour_v1beta1.Route{
					{
						Match:            "/",
						EnableWebsockets: inputs.EnableWebsockets,
						Services: []contour_v1beta1.Service{
							{
								Name: inputs.Service.ID,
								Port: int(inputs.Service.Port.SourcePort),
							},
						},
					},
				},
			}

			route := contour_v1beta1.IngressRoute{
				TypeMeta: metav1.TypeMeta{

					Kind:       "IngressRoute",
					APIVersion: "contour.heptio.com/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: x.getIngressRouteID(domain, *inputs),
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": inputs.Controller.ControllerID,
					},
				},
				Spec: spec,
			}

			existingIngressRoute, err := ingressClient.Get(x.getIngressRouteID(domain, *inputs), metav1.GetOptions{})
			// var nIng *contour_v1beta1.IngressRoute
			switch {
			case err == nil:
				route.SetResourceVersion(existingIngressRoute.ObjectMeta.ResourceVersion)
				_, err = ingressClient.Update(&route)
				if err != nil {
					return fmt.Errorf("Error: failed to update ingress: %s", err.Error())
				}
			case k8s_errors.IsNotFound(err):
				_, err = ingressClient.Create(&route)
				if err != nil {
					return fmt.Errorf("Error: failed to create ingress: %s", err.Error())
				}

			default:
				return fmt.Errorf("Unexpected error: %s", err.Error())
			}

		}

	}

	var upstreamFQDNs []string

	for _, domain := range inputs.UpstreamDomains {
		upstreamFQDNs = append(upstreamFQDNs, domain.FQDN)
	}

	artifacts = append(artifacts, transistor.Artifact{Key: "ingress_controller", Value: inputs.Controller.ControllerName, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "elb_dns", Value: inputs.Controller.ELB, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "name", Value: inputs.Service.Name, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "controlled_apex_domain", Value: inputs.ControlledApexDomain, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "cluster_ip", Value: serviceObj.Spec.ClusterIP, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "internal_dns", Value: fmt.Sprintf("%s.%s", inputs.Service.ID, namespace), Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "table_view", Value: getTableViewFromDomains(inputs.UpstreamDomains), Secret: false})

	x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)

	return nil
}
