package graphql

import (
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
)

type GraphQL struct {
	Events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("graphql", func() transistor.Plugin {
		return &GraphQL{}
	}, plugins.Project{})
}

func (x *GraphQL) Start(events chan transistor.Event) error {
	x.Events = events

	log.Info("Starting GraphQL service")
	return nil
}

func (x *GraphQL) Stop() {
	log.Info("Stopping GraphQL service")
}

func (x *GraphQL) Subscribe() []string {
	return []string{
		"gitsync:status",
		"heartbeat",
		"websocket",
		"project",
		"release",
	}
}

func (x *GraphQL) Process(e transistor.Event) error {
	log.DebugWithFields("Processing GraphQL event", log.Fields{"event": e})

	return nil
}

func (x *Resolver) DB() *gorm.DB {
	return nil
}
