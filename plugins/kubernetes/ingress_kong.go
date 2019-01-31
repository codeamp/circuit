package kubernetes

import (
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/spf13/viper"
	"k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (x *Kubernetes) ProcessKongIngress(e transistor.Event) {
	log.Println("processing kong ingress")

	kongAPI := viper.GetString("plugins.kubernetes.kong_api_url")
	log.Error(kongAPI)

	if e.Matches("project:kubernetes:ingresskong") {
		var err error
		switch e.Action {
		case transistor.GetAction("delete"):
			err = x.deleteKongIngress(e)
		case transistor.GetAction("create"):
			err = x.createKongIngress(e)
		case transistor.GetAction("update"):
			err = x.updateKongIngress(e)
		}

		if err != nil {
			log.Error(err)
			x.sendErrorResponse(e, err.Error())
		}
	}
}

func (x *Kubernetes) deleteKongIngress(e transistor.Event) error {
	return nil
}

func (x *Kubernetes) createKongIngress(e transistor.Event) error {

	service, err := x.createK8sService(e)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Error(service)

	input, err := getKongIngressInputs(e)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Error(input)

	return nil
}

func (x *Kubernetes) updateKongIngress(e transistor.Event) error {
	return nil
}

func (x *Kubernetes) createK8sService(e transistor.Event) (*v1.Service, error) {
	inputs, err := getKongIngressInputs(e)
	if err != nil {
		return nil, err
	}

	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.getKubernetesClient(e)
	if err != nil {
		return nil, err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	coreInterface := clientset.Core()
	deploymentName := x.GenDeploymentName(projectSlug, inputs.Service.Name)

	// var servicePorts []v1.ServicePort
	namespace := x.GenNamespaceName(payload.Environment, projectSlug)
	createNamespaceErr := x.CreateNamespaceIfNotExists(namespace, coreInterface)
	if createNamespaceErr != nil {
		return nil, createNamespaceErr
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
			return nil, fmt.Errorf("Error: failed to update service: %s", err.Error())
		}
		log.Debug(fmt.Sprintf("Service updated: %s", inputs.Service.ID))
	case k8s_errors.IsNotFound(err):
		serviceObj, err = service.Create(&serviceParams)
		if err != nil {
			return nil, fmt.Errorf("Error: failed to create service: %s", err.Error())
		}
		log.Debug(fmt.Errorf("Service created: %s", inputs.Service.ID))
	default:
		return nil, fmt.Errorf("Unexpected error: %s", err.Error())
	}

	return serviceObj, nil
}

func getKongIngressInputs(e transistor.Event) (*KongIngressInput, error) {
	input := KongIngressInput{}
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

		apexDomain, err := e.GetArtifact("controlled_apex_domain")
		if err != nil {
			return nil, err
		}
		input.ControlledApexDomain = apexDomain.String()

		upstreamRoutes, err := e.GetArtifact("upstream_routes")
		if err != nil {
			return nil, err
		}

		input.UpstreamRoutes = parseUpstreamRoutes(upstreamRoutes)

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

func parseUpstreamRoutes(a transistor.Artifact) []UpstreamRoute {
	var upstreamRoutes []UpstreamRoute
	for _, upstream := range a.Value.([]interface{}) {

		methods := strings.Split(strings.ToLower(upstream.(map[string]interface{})["methods"].(string)), ",")
		paths := strings.Split(strings.ToLower(upstream.(map[string]interface{})["paths"].(string)), ",")
		domains := upstream.(map[string]interface{})["domains"].(transistor.Artifact)

		fqdns := parseUpstreamDomains(domains)

		upstreamRoutes = append(upstreamRoutes, UpstreamRoute{
			FQDNs:   fqdns,
			Methods: methods,
			Paths:   paths,
		})
	}

	return upstreamRoutes
}
