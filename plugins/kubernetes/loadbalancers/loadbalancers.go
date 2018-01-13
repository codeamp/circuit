package kubernetesloadbalancers

import (
	"strconv"
	"github.com/davecgh/go-spew/spew"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/codeamp/circuit/plugins"
	utils "github.com/codeamp/circuit/plugins/kubernetes"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type LoadBalancers struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("kubernetesloadbalancers", func() transistor.Plugin {
		return &LoadBalancers{}
	})
}

func (x *LoadBalancers) Description() string {
	return "Create an ELB load balancer associated to a service"
}

func (x *LoadBalancers) SampleConfig() string {
	return ``
}

func (x *LoadBalancers) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started kubernetesloadbalancer plugin")

	return nil
}

func (x *LoadBalancers) Stop() {
	log.Info("Deleting Load Balancer") // Stop the creation of the load balancer, or delete an existing one?
}

func (x *LoadBalancers) Subscribe() []string {
	return []string{
		"plugins.Extension:create:kubernetesloadbalancers",
		"plugins.Extension:update:kubernetesloadbalancers",
		"plugins.Extension:destroy:kubernetesloadbalancers",
	}
}

func (x *LoadBalancers) Process(e transistor.Event) error {
	log.InfoWithFields("Processing load balancer event", log.Fields{
		"event": e,
	})

	event := e.Payload.(plugins.Extension)

	var err error
	switch event.Action {
	case plugins.GetAction("destroy"):
		err = x.doDeleteLoadBalancer(e)
	case plugins.GetAction("create"):
		err = x.doLoadBalancer(e)
	case plugins.GetAction("update"):
		err = x.doLoadBalancer(e)
	}

	if err != nil {
		event.State = plugins.GetState("failed")
		event.StateMessage = fmt.Sprintf("%v (Action: %v, Step: LoadBalancer", err.Error(), event.State)
		log.Debug(event.StateMessage)
		event := e.NewEvent(event, err)
		x.events <- event
		return err
	}

	log.Info("Processed LoadBalancer Events")

	return nil
}

type ListenerPair struct {
	Name       string
	Protocol   string
	SourcePort int32
	TargetPort int32
}

func (x *LoadBalancers) doLoadBalancer(e transistor.Event) error {
	log.Println("doLoadBalancer")

	payload := e.Payload.(plugins.Extension)
	// configPrefix := utils.GetFormValuePrefix(e, "LOADBALANCERS_")
	configPrefix := "KUBERNETESLOADBALANCERS_"

	spew.Dump("configPrefix", configPrefix)

	spew.Dump("PAYLOAD", payload)
	svcName := payload.Config[configPrefix+"SERVICE"].(string)
	lbName := payload.Config[configPrefix+"NAME"].(string)
	sslARN := payload.Config[configPrefix+"SSL_CERT_ARN"].(string)
	s3AccessLogs := payload.Config[configPrefix+"ACCESS_LOG_S3_BUCKET"].(string)
	lbType := plugins.GetType(payload.Config[configPrefix+"TYPE"].(string))
	projectSlug := plugins.GetSlug(payload.Project.Repository)

	spew.Dump("INPUTS", svcName, lbName, sslARN, s3AccessLogs, lbType, projectSlug)

	kubeconfig := payload.Config[configPrefix+"KUBECONFIG"].(string)
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{Timeout: "60"}).ClientConfig()

	if err != nil {
		failMessage := fmt.Sprintf("ERROR: %s; you must set the environment variable CF_PLUGINS_KUBEDEPLOY_KUBECONFIG=/path/to/kubeconfig", err.Error())
		log.Println(failMessage)
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), failMessage, err)
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		failMessage := fmt.Sprintf("ERROR: %s; setting NewForConfig in doLoadBalancer", err.Error())
		log.Println(failMessage)
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), failMessage, err)
		return nil
	}

	coreInterface := clientset.Core()
	deploymentName := utils.GenDeploymentName(projectSlug, svcName)

	var serviceType v1.ServiceType
	var servicePorts []v1.ServicePort
	serviceAnnotations := make(map[string]string)

	spew.Dump("HELLLLLLLOI", projectSlug, payload.Environment)
	namespace := utils.GenNamespaceName(payload.Environment, projectSlug)
	createNamespaceErr := utils.CreateNamespaceIfNotExists(namespace, coreInterface)
	if createNamespaceErr != nil {
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), createNamespaceErr.Error(), err)
		return nil
	}

	spew.Dump("NAMESPACE", namespace)
	// Begin create
	switch lbType {
	case plugins.GetType("internal"):
		serviceType = v1.ServiceTypeClusterIP
	case plugins.GetType("external"):
		serviceType = v1.ServiceTypeLoadBalancer
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"] = "true"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-timeout"] = "300"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"] = "true"
		if s3AccessLogs != "" {
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-emit-interval"] = "5"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-enabled"] = "true"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-name"] = s3AccessLogs
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix"] = fmt.Sprintf("%s/%s", projectSlug, svcName)
		}
	case plugins.GetType("office"):
		serviceType = v1.ServiceTypeLoadBalancer
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-internal"] = "0.0.0.0/0"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-enabled"] = "true"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-connection-draining-timeout"] = "300"
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled"] = "true"
		if s3AccessLogs != "" {
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-emit-interval"] = "5"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-enabled"] = "true"
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-name"] = s3AccessLogs
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix"] = fmt.Sprintf("%s/%s", projectSlug, svcName)
		}
	}
	spew.Dump("GOING THROUGH LISTENER PAIRS", payload.Config[configPrefix+"LISTENER_PAIRS"])
	listenerPairs := payload.Config[configPrefix+"LISTENER_PAIRS"]
	var sslPorts []string
	spew.Dump("WENT THROUGH LISTENER PAIRS", listenerPairs)
	for _, p := range listenerPairs.([]interface{}) {
		spew.Dump("llopping through pairs", p)
		
		spew.Dump(p.(map[string]interface{})["serviceProtocol"].(string))
		spew.Dump(p.(map[string]interface{})["serviceProtocol"].(string))
		spew.Dump(p.(map[string]interface{})["port"].(string))

		var realProto string
		switch strings.ToUpper(p.(map[string]interface{})["serviceProtocol"].(string)) {
		case "HTTPS":
			serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-backend-protocol"] = "tcp"
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
			x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), err.Error(), err)
			return nil			
		}

		intContainerPort, err := strconv.Atoi(p.(map[string]interface{})["containerPort"].(string))
		if err != nil {
			x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), err.Error(), err)
			return nil			
		}		

		spew.Dump("intPort", intPort)
		convPort := intstr.IntOrString{
			IntVal: int32(intContainerPort),
		}
		// random 5 letter sequence
		// randomLetters := "abcdev"
		newPort := v1.ServicePort{
			// TODO: remove this toLower when we fix the data in mongo, kube only allows lowercase port names
			Name:       strings.ToLower(fmt.Sprintf("%s", p.(map[string]interface{})["serviceProtocol"])),
			Port:       int32(intPort),
			TargetPort: convPort,
			Protocol:   v1.Protocol(realProto),
		}
		spew.Dump("newPort", newPort)
		if strings.ToUpper(p.(map[string]interface{})["serviceProtocol"].(string)) == "HTTPS" || 
			strings.ToUpper(p.(map[string]interface{})["serviceProtocol"].(string)) == "SSL" {
			sslPorts = append(sslPorts, fmt.Sprintf("%d", intPort))
		}
		servicePorts = append(servicePorts, newPort)
	}

	spew.Dump(servicePorts, sslPorts)
	if len(sslPorts) > 0 {
		sslPortsCombined := strings.Join(sslPorts, ",")
		spew.Dump(sslPortsCombined, sslARN)
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-ssl-ports"] = sslPortsCombined
		serviceAnnotations["service.beta.kubernetes.io/aws-load-balancer-ssl-cert"] = sslARN
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
			Name:        lbName,
			Annotations: serviceAnnotations,
		},
		Spec: serviceSpec,
	}

	// Implement service update-or-create semantics.
	service := coreInterface.Services(namespace)
	svc, err := service.Get(lbName, meta_v1.GetOptions{})
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
			errMsg := fmt.Sprintf("Error: failed to update service: %s", err.Error())
			x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), errMsg, err)
			return nil
		}
		fmt.Printf("Service updated: %s", lbName)
	case errors.IsNotFound(err):
		_, err = service.Create(&serviceParams)
		if err != nil {
			errMsg := fmt.Sprintf("Error: failed to create service: %s", err.Error())
			x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), errMsg, err)
			return nil
		}
		fmt.Printf("Service created: %s", lbName)
	default:
		errMsg := fmt.Sprintf("Unexpected error: %s", err.Error())
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), errMsg, err)
		return nil
	}

	// If ELB grab the DNS name for the response
	ELBDNS := ""
	if lbType == plugins.GetType("external") || lbType == plugins.GetType("office") {
		fmt.Printf("Waiting for ELB address for %s", lbName)
		// Timeout waiting for ELB DNS name after 900 seconds
		timeout := 90
		for {
			elbResult, elbErr := coreInterface.Services(namespace).Get(lbName, meta_v1.GetOptions{})
			spew.Dump(elbResult.Status.LoadBalancer.Ingress)
			if elbErr != nil {
				fmt.Printf("Error '%s' describing service %s", elbErr, lbName)
			} else {
				ingressList := elbResult.Status.LoadBalancer.Ingress
				if len(ingressList) > 0 {
					ELBDNS = ingressList[0].Hostname
					break
				}
				if timeout <= 0 {
					failMessage := fmt.Sprintf("Error: timeout waiting for ELB DNS name for: %s", lbName)
					x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), failMessage, err)
					return nil
				}
			}
			time.Sleep(time.Second * 5)
			timeout -= 5
		}
	} else {
		ELBDNS = fmt.Sprintf("%s.%s", lbName, utils.GenNamespaceName(payload.Environment, projectSlug))
	}

	route53Config := e.Payload.(plugins.Extension).Config
	route53Config["KUBERNETESLOADBALANCERS_ELBDNS"] = ELBDNS
	
	route53Event := e
	route53Event.Payload = plugins.Extension{
		Action: plugins.GetAction("create"),
		Slug: "route53",
		State: plugins.GetState("waiting"),
		StateMessage: "",
		Config: route53Config,
		Artifacts: map[string]string{},
		Environment: e.Payload.(plugins.Extension).Environment,
		Project: e.Payload.(plugins.Extension).Project,
	}	
	
	// send event to codeamp to signal loadbalancers completion
	event := utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("waiting"), "waiting for route53", nil)	
	event.Payload.(plugins.Extension).Artifacts["ELBDNS"] = ELBDNS
	spew.Dump("KLB EVENT", event)
	x.events <- event

	// send route53 event
	route53Event = e.NewEvent(route53Event.Payload, err)
	spew.Dump("ROUTE 53 EVENT", route53Event)
	x.events <- route53Event

	return nil
}

func (x *LoadBalancers) doDeleteLoadBalancer(e transistor.Event) error {
	log.Println("doDeleteLoadBalancer")
	var err error
	payload := e.Payload.(plugins.Extension)
	configPrefix := utils.GetFormValuePrefix(e, "KUBERNETESLOADBALANCERS_")
	kubeconfig, err := utils.GetFormValue(payload.Config, configPrefix, "KUBECONFIG")
	if err != nil {
		log.Debug(err)
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), err.Error(), err)
		return nil
	}

	svcName, err := utils.GetFormValue(payload.Config, configPrefix, "SERVICE")
	if err != nil {
		log.Debug(err)
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), err.Error(), err)
		return nil
	}

	lbName, err := utils.GetFormValue(payload.Config, configPrefix, "NAME")
	if err != nil {
		log.Debug(err)
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), err.Error(), err)
		return nil
	}
	projectSlug := plugins.GetSlug(payload.Project.Repository)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig.(string))
	if err != nil {
		fmt.Printf("ERROR '%s' while building kubernetes api client config.  Aborting!", err)
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error getting cluster config.  Aborting!")
		return nil
	}

	coreInterface := clientset.Core()
	namespace := utils.GenNamespaceName(payload.Environment, projectSlug)

	log.Println(namespace, lbName.(string))
	_, svcGetErr := coreInterface.Services(namespace).Get(lbName.(string), meta_v1.GetOptions{})
	if svcGetErr == nil {
		// Service was found, ready to delete
		svcDeleteErr := coreInterface.Services(namespace).Delete(lbName.(string), &meta_v1.DeleteOptions{})
		if svcDeleteErr != nil {
			failMessage := fmt.Sprintf("Error '%s' deleting service %s", svcDeleteErr, lbName)
			fmt.Printf("ERROR managing loadbalancer %s: %s", svcName, failMessage)
			x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), failMessage, err)
			return nil
		}
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("deleted"), "", nil)
	} else {
		// Send failure message that we couldn't find the service to delete
		failMessage := fmt.Sprintf("Error finding %s service: '%s'", lbName, svcGetErr)
		fmt.Printf("ERROR managing loadbalancer %s: %s", svcName, failMessage)
		x.events <- utils.CreateExtensionEvent(e, plugins.GetAction("status"), plugins.GetState("failed"), failMessage, err)
	}
	return nil
}
