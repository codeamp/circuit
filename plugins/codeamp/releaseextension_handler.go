package codeamp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func (x *CodeAmp) ReleaseExtensionEventHandler(e transistor.Event) error {
	log.Warn("ReleaseExtensionEventHandler")
	payload := e.Payload.(plugins.ReleaseExtension)

	var releaseExtension model.ReleaseExtension
	var release model.Release

	if e.Matches("release:.*:status") {
		if x.DB.Where("id = ?", payload.Release.ID).Find(&release).RecordNotFound() {
			log.InfoWithFields("release", log.Fields{
				"id": payload.Release.ID,
			})
			return fmt.Errorf("Release %s not found", payload.Release.ID)
		}

		if x.DB.Where("id = ?", payload.ID).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.ID,
			})
			return fmt.Errorf("Release extension %s not found", payload.ID)
		}

		releaseExtension.State = e.State
		releaseExtension.StateMessage = e.StateMessage
		marshalledReArtifacts, err := json.Marshal(e.Artifacts)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		releaseExtension.Finished = time.Now()
		if err := x.DB.Save(&releaseExtension).Error; err != nil {
			log.Error(err)
			return err
		}

		switch e.State {
		case transistor.GetState("complete"):
			x.ReleaseExtensionCompleted(&payload, &releaseExtension)
		case transistor.GetState("failed"):
			x.ReleaseFailed(&release, e.StateMessage)
		case transistor.GetState("running"):
			break
		default:
			log.Error("RELEASE IN ERROR STATE: ", e.State)
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionCompleted(payload *plugins.ReleaseExtension, re *model.ReleaseExtension) {
	log.Warn("ReleaseExtensionCompleted")

	releaseExtensions := []model.ReleaseExtension{}
	if x.DB.Where("release_id = ?", re.ReleaseID).Find(&releaseExtensions).RecordNotFound() {
		log.ErrorWithFields("release extensions not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	// Notify clientside a release extension has completed
	{
		wssPayload := plugins.WebsocketMsg{
			Event: fmt.Sprintf("projects/%s/%s/releases/reCompleted", payload.Release.Project.Slug, payload.Release.Environment),
		}
		event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), wssPayload)
		event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases/reCompleted", payload.Release.Project.Slug, payload.Release.Environment), false)

		x.Events <- event
	}

	// loop through and check if all same-type release extensions are completed
	for _, releaseExtension := range releaseExtensions {
		if releaseExtension.Type == re.Type && releaseExtension.State != transistor.GetState("complete") {
			return
		}
	}

	var release model.Release
	if x.DB.Where("id = ?", payload.Release.ID).First(&release).RecordNotFound() {
		log.WarnWithFields("release not found", log.Fields{
			"release": release,
		})
		return
	}

	switch re.Type {
	case plugins.GetType("workflow"):
		x.WorkflowReleaseExtensionsCompleted(payload, &release)
	case plugins.GetType("deployment"):
		log.Warn("Releases completed")
		x.ReleaseCompleted(&release)
	}
}

func (x *CodeAmp) WorkflowReleaseExtensionsCompleted(payload *plugins.ReleaseExtension, release *model.Release) {
	log.Debug("WorkflowReleaseExtensionsCompleted")

	releaseExtensionDeploymentsCount := 0
	/**************************************
	*
	* COLLECT WORKFLOW ARTIFACTS
	*
	/*************************************/
	releaseExtensions := []model.ReleaseExtension{}
	artifacts := []transistor.Artifact{}

	x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions)
	for _, releaseExtension := range releaseExtensions {
		if releaseExtension.Type == plugins.GetType("workflow") {
			projectExtension := model.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.WarnWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := model.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.WarnWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			var unmarshalledArtifacts []transistor.Artifact
			err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &unmarshalledArtifacts)
			if err != nil {
				log.ErrorWithFields(err.Error(), log.Fields{})
				return
			}

			for _, artifact := range unmarshalledArtifacts {
				artifact.Source = extension.Key
				artifacts = append(artifacts, artifact)
			}
		}
	}

	/**************************************
	*
	* SEND 'DEPLOYMENT' TYPE RE EVENTS
	*
	**************************************/
	for _, releaseExtension := range releaseExtensions {
		releaseExtensionAction := transistor.GetAction("create")
		if releaseExtension.Type == plugins.GetType("deployment") {
			_artifacts := artifacts

			// Fail deployment if the release is in a failed state
			if release.State == transistor.GetState("failed") {
				releaseExtensionAction = transistor.GetAction("status")
				releaseExtension.State = transistor.GetState("failed")
				releaseExtension.StateMessage = release.StateMessage

				x.DB.Save(&releaseExtension)
				log.Error("RELEASE IN FAILED STATE")
				continue
			}

			projectExtension := model.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.WarnWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := model.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.WarnWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			projectExtensionArtifacts, err := graphql_resolver.ExtractArtifacts(projectExtension, extension, x.DB)
			if err != nil {
				log.Error(err.Error())
			}

			for _, artifact := range projectExtensionArtifacts {
				_artifacts = append(_artifacts, artifact)
			}

			releaseExtension.Started = time.Now()
			x.DB.Save(&releaseExtension)

			releaseExtensionEvent := plugins.ReleaseExtension{
				ID:      releaseExtension.Model.ID.String(),
				Release: payload.Release,
			}

			ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), releaseExtensionAction, releaseExtensionEvent)
			ev.Artifacts = _artifacts
			x.Events <- ev

			releaseExtensionDeploymentsCount++
		}
	}

	log.Warn("releaseExtensionDeploymentsCount ", releaseExtensionDeploymentsCount)
	if releaseExtensionDeploymentsCount == 0 {
		x.ReleaseCompleted(release)
	}
}
