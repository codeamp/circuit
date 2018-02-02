package route53

import (
	"fmt"
	"strings"
	"time"

	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
)

type Route53 struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("route53", func() transistor.Plugin {
		return &Route53{}
	})
}

func (x *Route53) Description() string {
	return "Set Route53 DNS for Kubernetes services"
}

func (x *Route53) SampleConfig() string {
	return ` `
}

func (x *Route53) Start(e chan transistor.Event) error {
	x.events = e
	log.Println("Started Route53")
	return nil
}

func (x *Route53) Stop() {
	log.Println("Stopping Route53")
}

func (x *Route53) Subscribe() []string {
	return []string{
		"plugins.Extension:create:route53",
	}
}

func (x *Route53) Process(e transistor.Event) error {
	log.InfoWithFields("Processing route53 event", log.Fields{
		"event": e,
	})
	event := e.Payload.(plugins.Extension)
	var err error
	switch event.Action {
	case plugins.GetAction("create"):
		log.InfoWithFields(fmt.Sprintf("Process Route53 event: %s", e.Name), log.Fields{})
		err = x.updateRoute53(e)
	}

	if err != nil {
		failMessage := fmt.Sprintf("%v (Action: %v, Step: Route53", err.Error(), event.State)
		failedEvent := e.Payload.(plugins.Extension)
		failedEvent.Action = plugins.GetAction("status")
		failedEvent.State = plugins.GetState("failed")
		failedEvent.StateMessage = failMessage
		x.events <- e.NewEvent(failedEvent, fmt.Errorf("%s", failMessage))
		return nil
	}

	return nil
}

func (x *Route53) sendRoute53Response(e transistor.Event, state plugins.State, failureMessage string, lbPayload plugins.Extension) {
	event := e.NewEvent(plugins.Extension{
		Action:       plugins.GetAction("create"),
		State:        state,
		Slug:         "route53",
		StateMessage: failureMessage,
		Artifacts: map[string]string{
			"DNS":       lbPayload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string),
			"SUBDOMAIN": lbPayload.Config["KUBERNETESLOADBALANCERS_SUBDOMAIN"].(string),
			"FQDN":      lbPayload.Config["KUBERNETESLOADBALANCERS_HOSTED_ZONE_NAME"].(string),
		},
		Environment: lbPayload.Environment,
		Project:     lbPayload.Project,
	}, nil)

	x.events <- event
}

func (x *Route53) updateRoute53(e transistor.Event) error {
	payload := e.Payload.(plugins.Extension)
	// Sanity checks
	if payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string) == "" {
		failMessage := fmt.Sprintf("DNS was blank for %s, skipping Route53.", payload.Project.Repository)
		x.sendRoute53Response(e, plugins.GetState("failed"), failMessage, payload)
		return nil
	}
	if payload.Config["KUBERNETESLOADBALANCERS_NAME"] == "" {
		failMessage := fmt.Sprintf("Subdomain was blank for %s, skipping Route53.", payload.Project.Repository)
		x.sendRoute53Response(e, plugins.GetState("failed"), failMessage, payload)
	}
	if plugins.GetType(payload.Config["KUBERNETESLOADBALANCERS_TYPE"].(string)) == plugins.GetType("internal") {
		fmt.Printf("Internal service type ignored for %s", payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string))
		return nil
	}
	route53Name := fmt.Sprintf("%s.%s", payload.Config["KUBERNETESLOADBALANCERS_NAME"], payload.Config["KUBERNETESLOADBALANCERS_HOSTED_ZONE_NAME"].(string))
	if payload.Action == plugins.GetAction("create") {
		log.Info("Route53 plugin received LoadBalancer success message for %s, %s, %s.  Processing.\n", payload.Project.Repository, payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string), payload.Action)

		// Wait for DNS from the ELB to settle, abort if it does not resolve in initial_wait
		// Trying to be conservative with these since we don't want to update Route53 before the new ELB dns record is available

		// time.Sleep(time.Second * 300) // TODO: replace initial_wait_seconds

		// Query the DNS until it resolves or timeouts
		dnsTimeout := 600 // TODO: replace with dns_resolve_timeout_seconds
		dnsValid := false
		var failMessage string
		var dnsLookup []string
		var dnsLookupErr error
		for dnsValid == false {
			dnsLookup, dnsLookupErr = net.LookupHost(payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string))
			dnsTimeout -= 10
			if dnsLookupErr != nil {
				failMessage = fmt.Sprintf("Error '%s' resolving DNS for: %s", dnsLookupErr, payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string))
			} else if len(dnsLookup) == 0 {
				failMessage = fmt.Sprintf("Error 'found no names associated with ELB record' while resolving DNS for: %s", payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string))
			} else {
				dnsValid = true
			}
			if dnsTimeout <= 0 || dnsValid {
				break
			}
			time.Sleep(time.Second * 10)
			fmt.Println(failMessage + ".. Retrying in 10s")
		}
		if dnsValid == false {
			failedEvent := e.Payload.(plugins.Extension)
			failedEvent.Action = plugins.GetAction("status")
			failedEvent.State = plugins.GetState("failed")
			failedEvent.StateMessage = failMessage

			x.events <- e.NewEvent(failedEvent, fmt.Errorf("%s", failMessage))
			return nil
		}
		fmt.Printf("DNS for %s resolved to: %s\n", payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string), strings.Join(dnsLookup, ","))
		// Create the client
		sess := awssession.Must(awssession.NewSessionWithOptions(
			awssession.Options{
				Config: aws.Config{
					Credentials: credentials.NewStaticCredentials(payload.Config["KUBERNETESLOADBALANCERS_AWS_ACCESS_KEY_ID"].(string), payload.Config["KUBERNETESLOADBALANCERS_AWS_SECRET_KEY"].(string), ""),
				},
			},
		))

		client := route53.New(sess)
		// Look for this dns name
		params := &route53.ListResourceRecordSetsInput{
			HostedZoneId: aws.String(payload.Config["KUBERNETESLOADBALANCERS_HOSTED_ZONE_ID"].(string)), // Required
		}
		foundRecord := false
		pageNum := 0
		// Route53 has a . on the end of the name
		lookFor := fmt.Sprintf("%s.", route53Name)
		errList := client.ListResourceRecordSetsPages(params,
			func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
				pageNum++
				for _, p := range page.ResourceRecordSets {
					if *p.Name == lookFor {
						foundRecord = true
						// break out of pagination
						return true
					}
				}
				return false
			})
		if errList != nil {
			log.Info(fmt.Sprintf("Error listing ResourceRecordSets for Route53: %s", errList))
			return errList
		}
		if foundRecord {
			log.Info(fmt.Sprintf("Route53 found existing record for: %s\n", route53Name))
		} else {
			log.Info(fmt.Sprintf("Route53 record not found, creating %s\n", route53Name))
		}
		updateParams := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(payload.Config["KUBERNETESLOADBALANCERS_HOSTED_ZONE_ID"].(string)),
			ChangeBatch: &route53.ChangeBatch{
				Changes: []*route53.Change{
					{
						Action: aws.String("UPSERT"),
						ResourceRecordSet: &route53.ResourceRecordSet{

							Name: aws.String(route53Name),
							Type: aws.String("CNAME"),
							ResourceRecords: []*route53.ResourceRecord{
								{
									Value: aws.String(payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string)),
								},
							},
							TTL: aws.Int64(60),
						},
					},
				},
			},
		}

		_, err := client.ChangeResourceRecordSets(updateParams)
		if err != nil {
			failMessage := fmt.Sprintf("ERROR '%s' setting Route53 DNS for %s", err, route53Name)
			failedEvent := e.Payload.(plugins.Extension)
			failedEvent.State = plugins.GetState("failed")
			failedEvent.Action = plugins.GetAction("status")
			failedEvent.StateMessage = failMessage

			x.events <- e.NewEvent(failedEvent, fmt.Errorf("%s", failMessage))
			return nil
		}
		log.Info(fmt.Sprintf("Route53 record UPSERTed for %s: %s", route53Name, payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string)))

		// TODO: create aws manager that managers route53 and klb
		// on behalf of klb, we send a klb event complete from route53
		lbEventPayload := plugins.Extension{
			Id:           e.Payload.(plugins.Extension).Id,
			Action:       plugins.GetAction("status"),
			Slug:         "kubernetesloadbalancers",
			State:        plugins.GetState("complete"),
			StateMessage: "route53 cname created!",
			Config:       e.Payload.(plugins.Extension).Config,
			Artifacts: map[string]string{
				"ELBDNS":    payload.Config["KUBERNETESLOADBALANCERS_ELBDNS"].(string),
				"SUBDOMAIN": payload.Config["KUBERNETESLOADBALANCERS_NAME"].(string),
				"FQDN":      fmt.Sprintf("%s.%s", payload.Config["KUBERNETESLOADBALANCERS_NAME"].(string), payload.Config["KUBERNETESLOADBALANCERS_HOSTED_ZONE_NAME"].(string)),
			},
			Environment: payload.Environment,
			Project:     payload.Project,
		}

		x.events <- e.NewEvent(lbEventPayload, err)
	}

	return nil
}
