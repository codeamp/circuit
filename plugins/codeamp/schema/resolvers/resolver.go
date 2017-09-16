package codeamp_schema_resolvers

import (
	codeamp_actions "github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
)

type Resolver struct {
	db      *gorm.DB
	events  chan transistor.Event
	actions *codeamp_actions.Actions
}

func NewResolver(events chan transistor.Event, db *gorm.DB, actions *codeamp_actions.Actions) *Resolver {
	return &Resolver{
		events:  events,
		db:      db,
		actions: actions,
	}
}
