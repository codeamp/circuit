package kubernetes

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/kevholditch/gokong"
	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (x *Kubernetes) ProcessKongIngress(e transistor.Event) {
	log.Println("processing kong ingress")

	if e.Matches("project:kubernetes:ingresskong") {
		var err error
		switch e.Action {
		case transistor.GetAction("delete"):
			err = x.deleteKongIngress(e)
		case transistor.GetAction("create"):
			err = x.createKongIngress(e)
		case transistor.GetAction("update"):
			err = x.createKongIngress(e)
		}

		if err != nil {
			log.Error(err)
			x.sendErrorResponse(e, err.Error())
		}
	}
}

func (x *Kubernetes) deleteKongIngress(e transistor.Event) error {
	// delete routes
	err := x.deleteK8sService(e)
	if err != nil {
		x.sendErrorResponse(e, "failed to delete service")
		return fmt.Errorf(fmt.Sprintf("Failed to delete K8s Service: %s", err.Error()))
	}

	inputs, err := getKongIngressInputs(e)
	if err != nil {
		x.sendErrorResponse(e, "failed to delete service")
		return err
	}

	kongConfig := gokong.NewDefaultConfig()
	kongConfig.HostAddress = inputs.Controller.API
	kongClient := gokong.NewClient(kongConfig)

	status, err := kongClient.Status().Get()
	if err != nil {
		log.Error(err)
		log.Error(status)
		x.sendErrorResponse(e, "failed to delete service")
		return err
	}

	ingType, err := e.GetArtifact("type")
	if err != nil {
		return fmt.Errorf("Failed to retrieve service type")
		x.sendErrorResponse(e, "failed to delete service")
	}

	if ingType.String() == "loadbalancer" {
		inputs, err := getKongIngressInputs(e)
		if err != nil {
			x.sendErrorResponse(e, "failed to delete service")
			return err
		}

		for _, route := range inputs.UpstreamRoutes {
			routeName := generateRouteName(route, inputs.Service)

			err := kongClient.Routes().DeleteByName(routeName)
			if err != nil {
				log.Error(err)
				x.sendErrorResponse(e, "failed to delete service")
				return err
			}

		}
		// Find kong service
		existingKongService, err := kongClient.Services().GetServiceByName(inputs.Service.ID)
		if err != nil {
			log.Error(err)
			x.sendErrorResponse(e, "failed to delete service")
		}
		// delete service if it exists
		if existingKongService != nil {
			err = kongClient.Services().DeleteServiceById(*existingKongService.Id)
			if err != nil {
				log.Error(fmt.Sprintf("failed to delete service: %s", err.Error()))
				x.sendErrorResponse(e, "failed to delete service")
				return err
			}
		}
	}

	x.sendSuccessResponse(e, transistor.GetState("complete"), nil)

	return nil
}

func (x *Kubernetes) createKongIngress(e transistor.Event) error {

	service, err := x.createK8sService(e)
	if err != nil {
		log.Error(err)
		return err
	}

	inputs, err := getKongIngressInputs(e)
	if err != nil {
		log.Error(err)
		return err
	}
	var artifacts []transistor.Artifact

	if inputs.Type == "clusterip" {
		clusterDNS := fmt.Sprintf("%s.%s", service.Name, service.Namespace)
		newArtifacts := []transistor.Artifact{
			transistor.Artifact{Key: "table_view", Value: clusterDNS, Secret: false},
			transistor.Artifact{Key: "cluster_dns", Value: clusterDNS, Secret: false},
			transistor.Artifact{Key: "cluster_ip", Value: service.Spec.ClusterIP, Secret: false},
			transistor.Artifact{Key: "name", Value: inputs.Service.Name, Secret: false},
		}
		artifacts = append(artifacts, newArtifacts...)
		x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)
		return nil
	}

	routesArtifact := ""
	var tableView string
	for idx, route := range inputs.UpstreamRoutes {
		methods := strings.Join(route.Methods, "\n\t")
		paths := strings.Join(route.Paths, "\n\t")

		routesArtifact = routesArtifact + fmt.Sprintf("\nUPSTREAM_%d:\n\n", idx+1)
		tableView = strings.Join([]string{tableView, route.Domain.FQDN}, ",")

		routesArtifact = routesArtifact + "Domain:\n\t" + route.Domain.FQDN + "\n"
		routesArtifact = routesArtifact + "Methods:\n\t" + methods + "\n"
		routesArtifact = routesArtifact + "Paths:\n\t" + paths + "\n"
		routesArtifact = routesArtifact + "Target: \n\t" + fmt.Sprintf("%s:%d\n", inputs.Service.Name, inputs.Service.Port.TargetPort)

	}

	artifacts = append(artifacts, transistor.Artifact{Key: "ingress_controller", Value: inputs.Controller.ControllerName, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "elb_dns", Value: inputs.Controller.ELB, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "name", Value: inputs.Service.Name, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "controlled_apex_domain", Value: inputs.ControlledApexDomain, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "cluster_ip", Value: service.Spec.ClusterIP, Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "internal_dns", Value: fmt.Sprintf("%s.%s", inputs.Service.ID, service.Namespace), Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "table_view", Value: strings.Trim(tableView, ","), Secret: false})
	artifacts = append(artifacts, transistor.Artifact{Key: "routes", Value: routesArtifact, Secret: false})

	kongConfig := gokong.NewDefaultConfig()
	kongConfig.HostAddress = inputs.Controller.API
	kongClient := gokong.NewClient(kongConfig)

	status, err := kongClient.Status().Get()
	if err != nil {
		log.Error(err)
		log.Error(status)
		return err
	}

	serviceRequest := &gokong.ServiceRequest{
		Name:     gokong.String(inputs.Service.ID),
		Protocol: gokong.String("http"),
		Host:     gokong.String(fmt.Sprintf("%s.%s", inputs.Service.ID, service.GetNamespace())),
		Port:     gokong.Int(int(inputs.Service.Port.SourcePort)),
	}

	// Find service if already exists
	existingKongService, err := kongClient.Services().GetServiceByName(inputs.Service.ID)
	if err != nil {
		log.Error(err)
	}

	kongService := &gokong.Service{}
	if existingKongService != nil {
		kongService, err = kongClient.Services().UpdateServiceById(*existingKongService.Id, serviceRequest)
		if err != nil {
			log.Error(fmt.Sprintf("failed to update service: %s", err.Error()))
			return err
		}
	} else {
		kongService, err = kongClient.Services().Create(serviceRequest)
		if err != nil {
			log.Error(err)
			log.Error(kongService)
			return err
		}

	}

	// get existing upstream routes to remove after new routes are provisioned
	routesToRemove, err := kongClient.Routes().GetRoutesFromServiceName(inputs.Service.ID)
	if err != nil {
		log.Error(err)
	}

	// check for duplicate routes
	for _, route := range inputs.UpstreamRoutes {
		isDuplicateRoute, err := checkDuplicateRoute(*inputs, route, kongClient)

		if isDuplicateRoute || err != nil {
			log.Error(err)
			return fmt.Errorf(fmt.Sprintf("ERR: %s", err.Error()))
		}
	}

	for _, route := range inputs.UpstreamRoutes {

		keepRoute, err := kongClient.Routes().GetByName(generateRouteName(route, inputs.Service))
		if err != nil {
			log.Error(err)
		}
		if keepRoute != nil {
			skipCreate := false
			for idx, removeRoute := range routesToRemove {
				if *keepRoute.Name == *removeRoute.Name {
					routesToRemove = append(routesToRemove[:idx], routesToRemove[idx+1:]...)
					skipCreate = true
				}
			}
			if skipCreate == true {
				continue
			}
		}

		routeRequest := &gokong.RouteRequest{
			Name:         gokong.String(generateRouteName(route, inputs.Service)),
			Protocols:    gokong.StringSlice([]string{"http", "https"}),
			Methods:      gokong.StringSlice(route.Methods),
			Paths:        gokong.StringSlice(route.Paths),
			Hosts:        gokong.StringSlice([]string{route.Domain.FQDN}),
			StripPath:    gokong.Bool(false),
			PreserveHost: gokong.Bool(true),
			Service:      gokong.ToId(*kongService.Id),
		}

		if len(route.Methods) > 0 {
			methods := strings.Split(strings.ToUpper(strings.Join(route.Methods, ",")), ",")
			for i := range methods {
				methods[i] = strings.TrimSpace(methods[i])
			}
			routeRequest.Methods = gokong.StringSlice(methods)
		}

		kongRoute, err := kongClient.Routes().Create(routeRequest)
		if err != nil {
			log.Error(err)
			log.Error(kongRoute)
			return err
		}

	}

	// delete oldroutes
	for _, route := range routesToRemove {
		err := kongClient.Routes().DeleteById(*route.Id)
		if err != nil {
			log.Error(err)
		}
	}

	x.sendSuccessResponse(e, transistor.GetState("complete"), artifacts)

	return nil
}

func generateRouteName(route UpstreamRoute, service Service) string {
	return fmt.Sprintf("%s_%s", generateRouteKey(route), service.ID)
}

func generateRouteKey(route UpstreamRoute) string {
	sort.Strings(route.Methods)
	sort.Strings(route.Paths)

	key := fmt.Sprintf("%s_%s_%s", route.Domain.FQDN, strings.Join(route.Methods, "_"), strings.Join(route.Paths, "_"))

	// return cleaned key
	return strings.Replace(key, "/", "_", -1)
}

func checkDuplicateRoute(input KongIngressInput, upstream UpstreamRoute, client *gokong.KongAdminClient) (bool, error) {

	// list all routes
	// iterate through routes, checking for duplicates
	query := gokong.RouteQueryString{Offset: 0, Size: 1000}
	routesResult, err := client.Routes().List(&query)
	if err != nil {
		log.Error("Failed to retrieve routes")
	}

	routes := routesResult

	for len(routesResult) == query.Size {
		query.Offset += query.Size
		routesResult, err = client.Routes().List(&query)
		if err != nil {
			log.Error("faild to retrieve routes")
		}
		routes = append(routes, routesResult...)
	}

	searchMap := map[string]string{}
	for _, route := range routes {
		var methods []string
		var paths []string
		var hosts []string

		for _, method := range route.Methods {
			methods = append(methods, strings.ToUpper(*method))
		}
		for _, path := range route.Paths {
			paths = append(paths, strings.ToUpper(*path))
		}
		upstreamRoute := UpstreamRoute{
			Domain: Domain{
				FQDN: *route.Hosts[0],
			},
			Methods: methods,
			Paths:   paths,
		}
		searchMap[generateRouteKey(upstreamRoute)] = *route.Id

		log.InfoWithFields("route", log.Fields{
			"Methods": strings.Join(methods, ","),
			"Paths":   strings.Join(paths, ","),
			"Hosts":   strings.Join(hosts, ","),
		})

	}
	searchKey := generateRouteKey(upstream)

	if searchMap[searchKey] != "" {
		service, err := client.Services().GetServiceFromRouteId(searchMap[searchKey])
		if err != nil {
			return false, fmt.Errorf("Failed to get route's associated service")
		}

		if *service.Name != input.Service.ID {
			return true, fmt.Errorf("Ingress contains routes controlled by another Ingress")
		}
	}

	return false, nil
}

func (x *Kubernetes) updateKongIngress(e transistor.Event) error {
	return nil
}

func (x *Kubernetes) deleteK8sService(e transistor.Event) error {
	log.Info("deleteKubernetesService")
	var err error
	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.SetupClientset(e)
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

	return nil
}

func (x *Kubernetes) createK8sService(e transistor.Event) (*v1.Service, error) {
	inputs, err := getKongIngressInputs(e)
	if err != nil {
		return nil, err
	}

	payload := e.Payload.(plugins.ProjectExtension)

	clientset, err := x.SetupClientset(e)
	if err != nil {
		return nil, err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	coreInterface := clientset.Core()
	deploymentName := x.GenDeploymentName(projectSlug, inputs.Service.Name)
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

		upstreamRoutes, err := e.GetArtifact("upstream_routes")
		if err != nil {
			return nil, err
		}

		input.UpstreamRoutes, err = parseUpstreamRoutes(upstreamRoutes)
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

		controllers, err := parseKongControllers(ingressControllers.String())
		if err != nil {
			return nil, err
		}

		var selectedController KongIngressController
		found := false
		for _, controller := range controllers {
			if controller.ControllerID == selectedIngress.String() {
				found = true
				selectedController = controller
				break
			}
		}

		if found == false {
			return nil, fmt.Errorf("Selected Ingress Controller is Not Configured")
		}
		input.Controller = selectedController
	}

	return &input, nil
}

func parseKongControllers(ingressControllers string) ([]KongIngressController, error) {
	controllers := []KongIngressController{}

	err := json.Unmarshal([]byte(ingressControllers), &controllers)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to parse kong controllers: %s", err.Error()))
	}

	return controllers, nil
}

func parseUpstreamRoutes(a transistor.Artifact) ([]UpstreamRoute, error) {
	var upstreamRoutes []UpstreamRoute
	upstreams, ok := a.Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("Expected type []interface{} but got %T", upstreams))
	}

	for _, upstream := range a.Value.([]interface{}) {
		var methods []string
		var paths []string

		methodsString := upstream.(map[string]interface{})["methods"].(string)
		if methodsString == "" {
			methods = []string{}
		} else {
			methods = strings.Split(strings.ToLower(methodsString), ",")
		}

		pathsString := upstream.(map[string]interface{})["paths"].(string)
		if pathsString == "" {
			paths = []string{}
		} else {
			paths = strings.Split(pathsString, ",")
		}

		domains := upstream.(map[string]interface{})["domains"]

		domainArtifact := transistor.Artifact{
			Key:    "domains",
			Value:  domains,
			Secret: false,
		}

		fqdns, err := parseUpstreamDomains(domainArtifact)
		if err != nil {
			return []UpstreamRoute{}, err
		}

		// normalize domains to hosts
		for _, host := range fqdns {
			upstreamRoutes = append(upstreamRoutes, UpstreamRoute{
				Domain:  host,
				Methods: methods,
				Paths:   paths,
			})
		}
	}

	return upstreamRoutes, nil
}
