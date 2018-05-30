package codeamp

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func (x *CodeAmp) ProjectExtensionEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.ProjectExtension)
	var extension graphql_resolver.ProjectExtension
	var project graphql_resolver.Project

	if e.Matches("project:.*:status") {
		if x.DB.Where("id = ?", payload.ID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension not found", log.Fields{
				"id": payload.ID,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because ProjectExtension not found given payload id: %s.", payload.ID))
		}

		if x.DB.Where("id = ?", extension.ProjectID).Find(&project).RecordNotFound() {
			log.ErrorWithFields("project not found", log.Fields{
				"id": extension.ProjectID,
			})
			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because Project not found given payload id: %s.", extension.ProjectID))
		}

		extension.State = transistor.GetState(string(e.State))
		extension.StateMessage = e.StateMessage

		if e.State == transistor.GetState("complete") {
			if len(e.Artifacts) > 0 {
				marshalledReArtifacts, err := json.Marshal(e.Artifacts)
				if err != nil {
					log.ErrorWithFields(err.Error(), log.Fields{})
				}
				extension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
			}
		}

		x.DB.Save(&extension)

		event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), extension)
		event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/extensions", project.Slug, payload.Environment), false)
		x.Events <- event
	}

	return nil
}
