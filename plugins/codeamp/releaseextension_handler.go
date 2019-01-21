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
			log.Error(err.Error(), log.Fields{})
			return err
		}

		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		x.DB.Save(&releaseExtension)

		if e.State == transistor.GetState("complete") {
			releaseExtension.Finished = time.Now()
			x.DB.Save(&releaseExtension)

			x.ReleaseExtensionCompleted(&e, &releaseExtension)
		}

		if e.State == transistor.GetState("failed") {
			releaseExtension.Finished = time.Now()
			x.DB.Save(&releaseExtension)

			x.ReleaseFailed(&release, e.StateMessage)
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionCompleted(e *transistor.Event, re *model.ReleaseExtension) {
	log.Warn("ReleaseExtensionCompleted")

	project := model.Project{}
	release := model.Release{}
	environment := model.Environment{}
	releaseExtensions := []model.ReleaseExtension{}

	if x.DB.Where("id = ?", re.ReleaseID).First(&release).RecordNotFound() {
		log.ErrorWithFields("release not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.ErrorWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	if x.DB.Where("release_id = ?", re.ReleaseID).Find(&releaseExtensions).RecordNotFound() {
		log.ErrorWithFields("release extensions not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.ErrorWithFields("Environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases/reCompleted", project.Slug, environment.Key),
		Payload: release,
	}
	event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
	event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases/reCompleted", project.Slug, environment.Key), false)

	x.Events <- event

	// loop through and check if all same-type release extensions are completed
	done := true
	for _, releaseExtension := range releaseExtensions {
		if releaseExtension.Type == re.Type && releaseExtension.State != transistor.GetState("complete") {
			done = false
		}
	}

	if done {
		switch re.Type {
		case plugins.GetType("workflow"):
			x.WorkflowReleaseExtensionsCompleted(e, &release)
		case plugins.GetType("deployment"):
			x.ReleaseCompleted(&release)
		}
	}
}

func (x *CodeAmp) WorkflowReleaseExtensionsCompleted(e *transistor.Event, release *model.Release) {
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

			payload := e.Payload.(plugins.ReleaseExtension)
			releaseExtensionEvent := plugins.ReleaseExtension{
				ID: releaseExtension.Model.ID.String(),
				Release: payload.Release,
			}

			log.Warn("Sending ", fmt.Sprintf("release:%s", extension.Key), " message")
			ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), releaseExtensionAction, releaseExtensionEvent)
			ev.Artifacts = _artifacts

			releaseExtension.Started = time.Now()
			x.DB.Save(&releaseExtension)

			x.Events <- ev

			releaseExtensionDeploymentsCount++
		}
	}

	log.Warn("releaseExtensionDeploymentsCount ", releaseExtensionDeploymentsCount)
	if releaseExtensionDeploymentsCount == 0 {
		x.ReleaseCompleted(release)
	}
}
