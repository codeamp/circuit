package kubernetes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
)

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

func parseUpstreamDomains(a transistor.Artifact) ([]Domain, error) {
	var upstreamFQDNs []Domain

	domains, ok := a.Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("Expected type []interface{} but got %T", domains))
	}
	for _, domain := range a.Value.([]interface{}) {
		apex := strings.ToLower(domain.(map[string]interface{})["apex"].(string))
		subdomain := strings.ToLower(domain.(map[string]interface{})["subdomain"].(string))

		fqdn := fmt.Sprintf("%s.%s", subdomain, apex)

		upstreamFQDNs = append(upstreamFQDNs, Domain{
			Apex:      apex,
			Subdomain: subdomain,
			FQDN:      fqdn,
		})
	}

	return upstreamFQDNs, nil
}

func getTableViewFromDomains(domains []Domain) string {
	var upstreamFQDNs []string
	for _, domain := range domains {
		upstreamFQDNs = append(upstreamFQDNs, domain.FQDN)
	}

	return strings.Join(upstreamFQDNs[:], ", ")
}

// Service should be in the format servicename:port
func parseService(e transistor.Event) (Service, error) {
	payload := e.Payload.(plugins.ProjectExtension)

	protocol, err := e.GetArtifact("protocol")
	if err != nil {
		return Service{}, err
	}

	serviceRaw, err := e.GetArtifact("service")
	if err != nil {
		return Service{}, err
	}

	serviceParts := strings.Split(serviceRaw.String(), ":")
	if len(serviceParts) != 2 {
		return Service{}, fmt.Errorf("Malformed service reference: %s", serviceRaw.String())
	}

	serviceName := serviceParts[0]
	servicePort := serviceParts[1]

	portInt, err := strconv.Atoi(servicePort)
	if err != nil {
		return Service{}, fmt.Errorf("%s: Invalid Port type", serviceParts[1])
	}

	port := ListenerPair{
		Name:       fmt.Sprintf("http-%s-%.0f", serviceName, float64(portInt)),
		Protocol:   protocol.String(),
		SourcePort: int32(portInt),
		TargetPort: int32(portInt),
	}

	service := Service{
		ID:   fmt.Sprintf("%s-%s", serviceParts[0], payload.ID[0:5]),
		Name: serviceName,
		Port: port,
	}

	return service, nil

}
