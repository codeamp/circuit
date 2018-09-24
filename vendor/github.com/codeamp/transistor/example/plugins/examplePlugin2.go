package plugins

import (
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

type ExamplePlugin2 struct {
	events chan transistor.Event
	Hello  string `mapstructure:"hello"`
}

func init() {
	transistor.RegisterPlugin("examplePlugin2", func() transistor.Plugin {
		return &ExamplePlugin2{}
	})
}

func (x *ExamplePlugin2) Start(e chan transistor.Event) error {
	log.Info("starting ExamplePlugin2")

	payload := Hello{
		Action:  "examplePlugin2",
		Message: "Hello World from ExamplePlugin2",
	}

	e <- transistor.CreateEvent(transistor.EventName("examplePlugin2"), payload)

	return nil
}

func (x *ExamplePlugin2) Stop() {
	log.Info("stopping ExamplePlugin")
}

func (x *ExamplePlugin2) Subscribe() []string {
	return []string{
		"examplePlugin1:create",
	}
}

func (x *ExamplePlugin2) Process(e transistor.Event) error {
	if e.Event() == "examplePlugin2:create" {
		hello := e.Payload.(Hello)
		log.Info("ExamplePlugin2 received a message:", hello)
	}
	return nil
}
