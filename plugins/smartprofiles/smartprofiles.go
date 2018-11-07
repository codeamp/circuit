package smartprofiles

import (
	// "github.com/davecgh/go-spew/spew"
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
)

//SmartProfiles is a local struct for smartprofiles plugin
type SmartProfiles struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("smartprofiles", func() transistor.Plugin {
		return &SmartProfiles{}
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
	
	// new event with project service
	influxClient, err := InitInfluxClient()
	if err != nil {
		panic(err)
	}
	ch := make(chan *Service)

	go influxClient.GetService("web", "production-checkr-checkr", "72h", ch)

	svc := <- ch
	
	ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("status"), plugins.Project{
		Slug: "checkr-checkr",
		Environment: "production",
		Services: []plugins.Service{
			plugins.Service{
				Name: "web",
				Spec: plugins.ServiceSpec{
					CpuRequest: svc.RecommendedState.CPU.Request,
					CpuLimit: svc.RecommendedState.CPU.Limit,
					MemoryRequest: svc.RecommendedState.Memory.Request,
					MemoryLimit: svc.RecommendedState.Memory.Limit,
				},
			},
		},
	})
	x.events <- ev

	return nil
}
