package kubernetes

import (
	"fmt"
	"strings"

	"github.com/codeamp/transistor"
)

// func parseUpstreamDomains(a transistor.Artifact) []string {

// 	var upstreamFQDNs []string
// 	for _, domain := range a.Value.([]interface{}) {
// 		apex := strings.ToLower(domain.(map[string]interface{})["apex"].(string))
// 		subdomain := strings.ToLower(domain.(map[string]interface{})["subdomain"].(string))

// 		fqdn := fmt.Sprintf("%s.%s", subdomain, apex)

// 		upstreamFQDNs = append(upstreamFQDNs, fqdn)
// 	}

// 	return upstreamFQDNs
// }

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

func parseUpstreamDomains(a transistor.Artifact) []Domain {

	var upstreamFQDNs []Domain
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

	return upstreamFQDNs
}

func getTableViewFromDomains(domains []Domain) string {
	var upstreamFQDNs []string
	for _, domain := range domains {
		upstreamFQDNs = append(upstreamFQDNs, domain.FQDN)
	}

	return strings.Join(upstreamFQDNs[:], ", ")
}
