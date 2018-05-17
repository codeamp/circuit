package kubernetes

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (x *Kubernetes) ProcessLoadBalancer(e transistor.Event) {
	log.InfoWithFields("Processing load balancer event", log.Fields{
		"event": e,
	})

	var err error
	switch e.Action {
	case plugins.GetAction("destroy"):
		err = x.doDeleteLoadBalancer(e)
	case plugins.GetAction("create"):
		err = x.doLoadBalancer(e)
	case plugins.GetAction("update"):
		err = x.doLoadBalancer(e)
	}

	if err != nil {
		log.Error(err)
		//x.sendErrorResponse(e, fmt.Sprintf("%v (Action: %v, Step: LoadBalancer", err.Error(), event.State))
		x.sendErrorResponse(e, err.Error())
	}
}

func (x *Kubernetes) doLoadBalancer(e transistor.Event) error {
	payload := e.Payload.(plugins.ProjectExtension)
	svcName, err := e.GetArtifact("service")
	if err != nil {
		log.Warn("missing service")
		return err
	}

	lbName, err := e.GetArtifact("name")
	if err != nil {
		name := fmt.Sprintf("%s-%s", svcName.String(), payload.ID[0:5])
		e.AddArtifact("name", name, false)

		lbName, err = e.GetArtifact("name")
		if err != nil {
			return err
		}
	}

	// Delete old LB if service was changed and update the name
	if !strings.HasPrefix(lbName.String(), fmt.Sprintf("%s-", svcName.String())) {
		err := deleteLoadBalancer(e, x)
		// The load balancer might fail to delete because it does not exist in the first place
		// The point is to make sure it's not there when we try to create it later.
		if err != nil {
			log.Warn(err)
		}

		name := fmt.Sprintf("%s-%s", svcName.String(), payload.ID[0:5])
		e.AddArtifact("name", name, false)

		lbName, err = e.GetArtifact("name")
		if err != nil {
			return err
		}
	}

	sslARN, err := e.GetArtifact("ssl_cert_arn")
	if err != nil {
		return err
	}

	s3AccessLogs, err := e.GetArtifact("access_log_s3_bucket")
	if err != nil {
		return err
	}

	_lbType, err := e.GetArtifact("type")
	if err != nil {
		return err
	}

	lbType := plugins.GetType(_lbType.String())

	projectSlug := plugins.GetSlug(payload.Project.Repository)
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
		failMessage := fmt.Sprintf("ERROR: %s; setting NewForConfig in doLoadBalancer", err.Error())
		log.Error(failMessage)
		return err
	}

	coreInterface := clientset.Core()
	deploymentName := x.GenDeploymentName(projectSlug, svcName.String())

	var serviceType v1.ServiceType
	var servicePorts []v1.ServicePort
	serviceAnnotations := make(map[string]string)
	namespace := x.GenNamespaceName(payload.Environment, projectSlug)
	createNamespaceErr := x.CreateNamespaceIfNotExists(namespace, coreInterface)
	if createNamespaceErr != nil {
		return createNamespaceErr
	}

	// Begin create
	switch lbType {
	case plugins.GetType("internal"):
		serviceType = v1.ServiceTypeClusterIP
	case plugins.GetType("external"):
		serviceType = v1.ServiceTypeLoadBalancer
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"] = "true"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-timeout"] = "300"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"] = "true"
		if s3AccessLogs.String() != "" {
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-emit-interval"] = "5"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-enabled"] = "true"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-name"] = s3AccessLogs.String()
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix"] = fmt.Sprintf("%s/%s", projectSlug, svcName.String())
		}
	case plugins.GetType("office"):
		serviceType = v1.ServiceTypeLoadBalancer
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-internal"] = "0.0.0.0/0"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"] = "true"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-timeout"] = "300"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"] = "true"
		if s3AccessLogs.String() != "" {
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-emit-interval"] = "5"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-enabled"] = "true"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-name"] = s3AccessLogs.String()
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix"] = fmt.Sprintf("%s/%s", projectSlug, svcName.String())
		}
	}
	listenerPairs, err := e.GetArtifact("listener_pairs")
	if err != nil {
		return err
	}

	var sslPorts []string
	for _, p := range listenerPairs.StringSlice() {
		var realProto string
		switch strings.ToUpper(p.(map[string]interface{})["serviceProtocol"].(string)) {
		case "HTTPS":
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-backend-protocol"] = "http"
			realProto = "TCP"
		case "SSL":
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-backend-protocol"] = "tcp"
			realProto = "TCP"
		case "HTTP":
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-backend-protocol"] = "http"
			realProto = "TCP"
		case "TCP":
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-backend-protocol"] = "tcp"
			realProto = "TCP"
		case "UDP":
			realProto = "UDP"
		}
		intPort, err := strconv.Atoi(p.(map[string]interface{})["port"].(string))
		if err != nil {
			return err
		}

		intContainerPort, err := strconv.Atoi(p.(map[string]interface{})["containerPort"].(string))
		if err != nil {
			return err
		}
		convPort := intstr.IntOrString{
			IntVal: int32(intContainerPort),
		}
		// random 5 letter sequence
		// randomLetters := "abcdev"
		newPort := v1.ServicePort{
			// TODO: remove this toLower when we fix the data in mongo, kube only allows lowercase port names
			Name: strings.ToLower(fmt.Sprintf("%s-%s-%s", p.(map[string]interface{})["serviceProtocol"], p.(map[string]interface{})["port"],
				p.(map[string]interface{})["containerPort"])),
			Port:       int32(intPort),
			TargetPort: convPort,
			Protocol:   v1.Protocol(realProto),
		}
		if strings.ToUpper(p.(map[string]interface{})["serviceProtocol"].(string)) == "HTTPS" ||
			strings.ToUpper(p.(map[string]interface{})["serviceProtocol"].(string)) == "SSL" {
			sslPorts = append(sslPorts, fmt.Sprintf("%d", intPort))
		}
		servicePorts = append(servicePorts, newPort)
	}
	if len(sslPorts) > 0 {
		sslPortsCombined := strings.Join(sslPorts, ",")
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-ssl-ports"] = sslPortsCombined
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-ssl-cert"] = sslARN.String()
	}
	serviceSpec := v1.ServiceSpec{
		Selector: map[string]string{"app": deploymentName},
		Type:     serviceType,
		Ports:    servicePorts,
	}
	serviceParams := v1.Service{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        lbName.String(),
			Annotations: serviceAnnotations,
		},
		Spec: serviceSpec,
	}

	// Implement service update-or-create semantics.
	log.Debug("Implement service update-or-create semantics.")
	service := coreInterface.Services(namespace)
	svc, err := service.Get(lbName.String(), meta_v1.GetOptions{})
	switch {
	case err == nil:
		// Preserve the NodePorts for PATCH service.
		if svc.Spec.Type == "LoadBalancer" {
			for i := range svc.Spec.Ports {
				for j := range serviceParams.Spec.Ports {
					// TODO: remove this toLower when we fix the data in mongo, kube only allows lowercase port names
					if strings.ToLower(svc.Spec.Ports[i].Name) == strings.ToLower(serviceParams.Spec.Ports[j].Name) {
						serviceParams.Spec.Ports[j].NodePort = svc.Spec.Ports[i].NodePort
					}
				}
			}
		}
		serviceParams.ObjectMeta.ResourceVersion = svc.ObjectMeta.ResourceVersion
		serviceParams.Spec.ClusterIP = svc.Spec.ClusterIP
		_, err = service.Update(&serviceParams)
		if err != nil {
			return errors.New(fmt.Sprintf("Error: failed to update service: %s", err.Error()))
		}
		log.Debug(fmt.Sprintf("Service updated: %s", lbName.String()))
	case k8s_errors.IsNotFound(err):
		_, err = service.Create(&serviceParams)
		if err != nil {
			return errors.New(fmt.Sprintf("Error: failed to create service: %s", err.Error()))
		}
		log.Debug(fmt.Sprintf("Service created: %s", lbName.String()))
	default:
		return errors.New(fmt.Sprintf("Unexpected error: %s", err.Error()))
	}

	// If ELB grab the DNS name for the response
	log.Debug("If ELB grab the DNS name for the response ", lbType)
	ELBDNS := ""
	if lbType == plugins.GetType("external") || lbType == plugins.GetType("office") {
		log.Debug(fmt.Sprintf("Waiting for ELB address for %s", lbName.String()))
		// Timeout waiting for ELB DNS name after 90 seconds
		timeout := 90
		for {
			elbResult, elbErr := coreInterface.Services(namespace).Get(lbName.String(), meta_v1.GetOptions{})
			if elbErr != nil {
				log.Error(fmt.Sprintf("Error '%s' describing service %s", elbErr, lbName.String()))
			} else {
				ingressList := elbResult.Status.LoadBalancer.Ingress
				if len(ingressList) > 0 {
					ELBDNS = ingressList[0].Hostname
					break
				}
				if timeout <= 0 {
					return errors.New(fmt.Sprintf("Error: timeout waiting for ELB DNS name for: %s", lbName.String()))
				}
			}
			time.Sleep(time.Second * 5)
			timeout -= 5
		}
	} else {
		ELBDNS = fmt.Sprintf("%s.%s", lbName.String(), x.GenNamespaceName(payload.Environment, projectSlug))
	}

	artifacts := make([]transistor.Artifact, 2, 2)
	artifacts[0] = transistor.Artifact{Key: "dns", Value: ELBDNS, Secret: false}
	artifacts[1] = transistor.Artifact{Key: "name", Value: lbName.String(), Secret: false}

	x.sendSuccessResponse(e, plugins.GetState("complete"), artifacts)
	return nil
}

func (x *Kubernetes) doDeleteLoadBalancer(e transistor.Event) error {
	err := deleteLoadBalancer(e, x)

	if err != nil {
		x.sendErrorResponse(e, err.Error())
	} else {
		log.Warn("sending success deleted")
		x.sendSuccessResponse(e, plugins.GetState("deleted"), nil)
	}

	return nil
}

func deleteLoadBalancer(e transistor.Event, x *Kubernetes) error {
	log.Info("deleteLoadBalancer")
	var err error
	payload := e.Payload.(plugins.ProjectExtension)

	kubeconfig, err := x.SetupKubeConfig(e)
	if err != nil {
		return err
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{Timeout: "60"}).ClientConfig()

	if err != nil {
		return errors.New(fmt.Sprintf("ERROR: %s; you must set the environment variable CF_PLUGINS_KUBEDEPLOY_KUBECONFIG=/path/to/kubeconfig", err.Error()))
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR: %s; setting NewForConfig in doLoadBalancer", err.Error()))
	}

	svcName, err := e.GetArtifact("service")
	if err != nil {
		return err
	}

	lbName, err := e.GetArtifact("name")
	if err != nil {
		return err
	}

	projectSlug := plugins.GetSlug(payload.Project.Repository)

	coreInterface := clientset.Core()
	namespace := x.GenNamespaceName(payload.Environment, projectSlug)

	_, svcGetErr := coreInterface.Services(namespace).Get(lbName.String(), meta_v1.GetOptions{})
	if svcGetErr == nil {
		// Service was found, ready to delete
		svcDeleteErr := coreInterface.Services(namespace).Delete(lbName.String(), &meta_v1.DeleteOptions{})
		if svcDeleteErr != nil {
			return errors.New(fmt.Sprintf("Error managing loadbalancer '%s' deleting service %s. %s.", svcDeleteErr, lbName.String(), svcName.String()))
		}
	} else {
		// Send failure message that we couldn't find the service to delete
		return errors.New(fmt.Sprintf("Error managing loadbalancer finding %s service: '%s'. '%s'", lbName.String(), svcGetErr, svcName.String()))
	}

	return nil
}
