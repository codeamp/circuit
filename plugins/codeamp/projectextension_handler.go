package codeamp

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func (x *CodeAmp) ProjectExtensionEventHandler(e transistor.Event) error {
	receivedPayload := e.Payload.(plugins.ProjectExtension)
	var extension model.ProjectExtension
	var project model.Project

	if e.Matches("project:.*:status") {
		if x.DB.Unscoped().Where("id = ?", receivedPayload.ID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension not found", log.Fields{
				"id": receivedPayload.ID,
			})

			return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because ProjectExtension not found given payload id: %s.", receivedPayload.ID))
		} else {
			if extension.DeletedAt != nil {
				return fmt.Errorf(fmt.Sprintf("Could not handle ProjectExtension status event because ProjectExtension has been deleted id: %s.", receivedPayload.ID))
			}
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

		err := x.DB.Save(&extension).Error
		if err != nil {
			log.Error(err.Error())
		}

		sendPayload := plugins.WebsocketMsg{
			Event: fmt.Sprintf("projects/%s/%s/extensions", project.Slug, receivedPayload.Environment),
		}
		event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), sendPayload)
		event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/extensions", project.Slug, receivedPayload.Environment), false)
		x.Events <- event
	}

	return nil
}
