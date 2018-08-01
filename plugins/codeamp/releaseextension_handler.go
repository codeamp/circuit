package codeamp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func (x *CodeAmp) ReleaseExtensionEventHandler(e transistor.Event) error {
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

		spew.Dump("e.State", e.State)

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

			x.ReleaseExtensionCompleted(&releaseExtension)
		}

		if e.State == transistor.GetState("failed") {
			releaseExtension.Finished = time.Now()
			x.DB.Save(&releaseExtension)

			x.ReleaseFailed(&release, e.StateMessage)
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionCompleted(re *model.ReleaseExtension) {
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

	// loop through and check if all same-type elease extensions are completed
	done := true
	for _, releaseExtension := range releaseExtensions {
		spew.Dump("re", re.Type)
		spew.Dump("releaseExtension", releaseExtension.State, releaseExtension.Type)
		if releaseExtension.Type == re.Type && releaseExtension.State != transistor.GetState("complete") {
			done = false
		}
	}

	spew.Dump("done", done)

	if done {
		switch re.Type {
		case plugins.GetType("workflow"):
			x.WorkflowReleaseExtensionsCompleted(&release)
		case plugins.GetType("deployment"):
			x.ReleaseCompleted(&release)
		}
	}
}
