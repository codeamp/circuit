package smartprofiles

import (
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	// "github.com/davecgh/go-spew/spew"

	log "github.com/codeamp/logger"
)

//SmartProfiles is a local struct for smartprofiles plugin
type SmartProfiles struct {
	events chan transistor.Event
	InfluxClienter
}

func init() {
	transistor.RegisterPlugin("smartprofiles", func() transistor.Plugin {
		return &SmartProfiles{
			InfluxClienter: &InfluxClient{},
		}
	}, plugins.Project{})
}

// Description: Plugin description
func (x *SmartProfiles) Description() string {
	return "Get service resource recommendations"
}

// SampleConfig return plugin sample config
func (x *SmartProfiles) SampleConfig() string {
	return ` `
}

// Start plugin
func (x *SmartProfiles) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Smart Profiles")
	return nil
}

// Stop spins slack down
func (x *SmartProfiles) Stop() {
	log.Info("Stopping Smart Profiles")
}

// Subscribe to events
func (x *SmartProfiles) Subscribe() []string {
	return []string{
		"smartprofiles:update",
	}
}

// Process slack webhook events
func (x *SmartProfiles) Process(e transistor.Event) error {	
	log.DebugWithFields("Processing SmartProfiles event", log.Fields{
		"event": e.Event(),
	})

	project := e.Payload.(plugins.Project)
	projectNamespace := fmt.Sprintf("%s-%s", strings.ToLower(project.Environment), strings.ToLower(project.Slug))

	// new event with project service
	influxHost, err := e.GetArtifact("INFLUX_HOST")
	if err != nil {
		return err
	}

	influxDBName, err := e.GetArtifact("INFLUX_DB")
	if err != nil {
		return err
	}

	err = x.InfluxClienter.InitInfluxClient(influxHost.String(), influxDBName.String())
	if err != nil {
		return err
	}

	ch := make(chan *Service)

	for _, service := range project.Services {
		go x.InfluxClienter.GetService(service.ID, service.Name, projectNamespace, "72h", ch)
	}

	respProject := project
	respProject.Services = []plugins.Service{}

	for range project.Services {
		var svc *Service
		svc = <-ch

		projectService := plugins.Service{
			ID:   svc.ID,
			Name: svc.Name,
			Spec: plugins.ServiceSpec{
				CpuRequest:    svc.RecommendedState.CPU.Request,
				CpuLimit:      svc.RecommendedState.CPU.Limit,
				MemoryRequest: svc.RecommendedState.Memory.Request,
				MemoryLimit:   svc.RecommendedState.Memory.Limit,
			},
		}

		respProject.Services = append(respProject.Services, projectService)
	}

	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), respProject)
	x.events <- ev

	return nil
}
