package codeamp

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

func (x *CodeAmp) ReleaseEventHandler(e transistor.Event) error {
	log.Error("RELEASE EVENT HANDLER")

	if e.Matches("release:create") {
		releaseEvent := e.Payload.(plugins.Release)
		release := model.Release{}

		if x.DB.Where("id = ?", releaseEvent.ID).First(&release).RecordNotFound() {
			log.InfoWithFields("release not found", log.Fields{
				"id": releaseEvent.ID,
			})
			return fmt.Errorf("release %s not found", releaseEvent.ID)
		}

		startableRelease, startableReleaseEvent, err := x.GetStartableRelease(&release, &releaseEvent)
		if err == nil {
			err := x.StartRelease(startableRelease, startableReleaseEvent)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			log.Error(err)
			return err
		}
	}

	return nil
}

func (x *CodeAmp) GetStartableRelease(release *model.Release, releaseEvent *plugins.Release) (*model.Release, *plugins.Release, error) {
	var queuedRelease model.Release
	if err := x.DB.
		Where("project_id = ? and environment_id = ? and state IN ('running', 'waiting')", release.ProjectID, release.EnvironmentID).
		Order("created_at asc").
		First(&queuedRelease).
		Error; err != nil {
		log.Error(err)
		return nil, nil, err
	}
	if queuedRelease.State == transistor.GetState("waiting") {
		if release == nil || queuedRelease.ID != release.ID {
			log.Warn("Forced to rebuild release event!")

			releaseMutation := graphql_resolver.ReleaseResolverMutation{DB: x.DB}

			releaseID := queuedRelease.Model.ID.String()
			releaseComponents, err := releaseMutation.PrepRelease(queuedRelease.ProjectID.String(), queuedRelease.EnvironmentID.String(),
				queuedRelease.HeadFeatureID.String(), &releaseID)			
			if err != nil {
				return nil, nil, err
			}

			_releaseEvent, err := releaseMutation.BuildReleaseEvent(&queuedRelease, releaseComponents)
			if err != nil {
				return nil, nil, err
			}
			pluginReleaseEvent := _releaseEvent.Payload.(plugins.Release)
			releaseEvent = &pluginReleaseEvent
		}

		return &queuedRelease, releaseEvent, nil
	}
	return nil, nil, errors.New("No releases startable")
}

func (x *CodeAmp) StartRelease(release *model.Release, releaseEvent *plugins.Release) error {
	err := x.CreateReleaseExtensionsForRelease(release, releaseEvent)
	if err != nil {
		return err
	}

	err = x.StartWorkflowExtensions(release, releaseEvent)
	if err != nil {
		return err
	}

	release.Started = time.Now()
	release.State = transistor.GetState("running")
	release.StateMessage = "Running Release"

	if err := x.DB.Save(&release).Error; err != nil {
		log.Error(err)
		return err
	}

	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", releaseEvent.Project.Slug, releaseEvent.Environment),
		Payload: release,
	}

	// Send event to websocket plugin to forward message to client
	// Tell the client to refetch its data re: this release
	event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
	event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases", releaseEvent.Project.Slug, releaseEvent.Environment), false)
	x.Events <- event

	return nil
}

func (x *CodeAmp) StartWorkflowExtensions(release *model.Release, releaseEvent *plugins.Release) error {
	log.Debug("StartWorkflowExtensions")
	var releaseExtensions []model.ReleaseExtension
	if err := x.DB.
		Where("release_id = ? and type = 'workflow'", release.Model.ID).
		Find(&releaseExtensions).Error; err != nil {
		log.Error(err)
		return err
	}

	for _, re := range releaseExtensions {
		projectExtension := model.ProjectExtension{}
		if x.DB.Where("id = ?", re.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
			log.InfoWithFields("project extensions not found", log.Fields{
				"id": re.ProjectExtensionID,
				"release_extension_id": re.Model.ID,
			})
			return fmt.Errorf("project extension %s not found", re.ProjectExtensionID)
		}

		extension := model.Extension{}
		if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": projectExtension.Model.ID,
				"release_extension_id": re.Model.ID,
			})
			return fmt.Errorf("extension %s not found", projectExtension.ExtensionID)
		}

		re.Started = time.Now()
		re.State = "running"
		re.StateMessage = "Release Extension Started"
		if err := x.DB.Save(&re).Error; err != nil {
			return err
		}

		artifacts, err := graphql_resolver.ExtractArtifacts(projectExtension, extension, x.DB)
		if err != nil {
			log.Error(err.Error())
			return nil
		}

		payload := plugins.ReleaseExtension{
			ID:      re.Model.ID.String(),
			Release: *releaseEvent,
		}

		eventAction := transistor.GetAction("create")
		eventState := transistor.GetState("waiting")

		log.Warn("Sending events for release extension: ", re.Model.ID)
		ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), eventAction, payload)
		ev.State = eventState
		ev.StateMessage = ""
		ev.Artifacts = artifacts

		x.Events <- ev
	}

	return nil
}

func (x *CodeAmp) CreateReleaseExtensionsForRelease(release *model.Release, releaseEvent *plugins.Release) error {
	log.Debug("CreateReleaseExtensionsForRelease")

	var err error
	var projectExtensions []model.ProjectExtension
	err = json.Unmarshal(release.ProjectExtensions.RawMessage, &projectExtensions)
	if err != nil {
		return err
	}

	for _, projectExtension := range projectExtensions {
		extension := model.Extension{}
		if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.ErrorWithFields("extension spec not found", log.Fields{
				"id": projectExtension.ExtensionID,
			})
			return errors.New("extension spec not found")
		}

		if plugins.Type(extension.Type) == plugins.GetType("workflow") || plugins.Type(extension.Type) == plugins.GetType("deployment") {
			var headFeature model.Feature
			if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
				log.ErrorWithFields("head feature not found", log.Fields{
					"id": release.HeadFeatureID,
				})
				return errors.New("head feature not found")
			}

			// create ReleaseExtension
			releaseExtension := model.ReleaseExtension{
				State:              transistor.GetState("waiting"),
				StateMessage:       "",
				ReleaseID:          release.Model.ID,
				FeatureHash:        headFeature.Hash,
				ServicesSignature:  fmt.Sprintf("%x", "servicesSig"),
				SecretsSignature:   fmt.Sprintf("%x", "secretsSig"),
				ProjectExtensionID: projectExtension.Model.ID,
				Type:               extension.Type,
			}

			if err := x.DB.Create(&releaseExtension).Error; err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseFailed(release *model.Release, stateMessage string) {
	log.Warn("ReleaseFailed")
	// Update DB to reflect release has failed
	release.State = transistor.GetState("failed")
	release.StateMessage = stateMessage
	release.Finished = time.Now()

	project := model.Project{}
	environment := model.Environment{}

	if err := x.DB.Save(release).Error; err != nil {
		log.Error(err)
	}

	// Mark all release extensions as failed
	releaseExtensions := []model.ReleaseExtension{}
	x.DB.Where("release_id = ? AND state <> ?", release.Model.ID, transistor.GetState("complete")).Find(&releaseExtensions)
	for _, re := range releaseExtensions {
		re.State = transistor.GetState("failed")
		if err := x.DB.Save(&re).Error; err != nil {
			log.Error(err)
		}
	}

	// Gather environment and project ID to send websocket message
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.WarnWithFields("Environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
	}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	// Notify all listeners release has failed
	x.SendNotifications("FAILED", release, &project)

	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key),
		Payload: release,
	}

	// Send event to websocket plugin to forward message to client
	// Tell the client to refetch its data re: this release
	event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
	event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key), false)
	x.Events <- event

	// Continue on and run other releases if they are queued and waiting
	x.RunQueuedReleases(release)
}

func (x *CodeAmp) ReleaseCompleted(release *model.Release) {
	log.Warn("Release Completed")

	project := model.Project{}
	environment := model.Environment{}

	// Find project and environment names to build websocket message
	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.WarnWithFields("Environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
	}

	// mark release as complete, unless it was canceled
	if release.State != transistor.GetState("canceled") {
		release.State = transistor.GetState("complete")
		release.StateMessage = "Completed"
		release.Finished = time.Now()

		if err := x.DB.Save(release).Error; err != nil {
			log.Error(err)
		}

		x.SendNotifications("SUCCESS", release, &project)
	} else {
		x.SendNotifications("CANCELED", release, &project)
	}

	// Notify the front end
	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key),
		Payload: release,
	}
	event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
	event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key), false)
	x.Events <- event

	// Try to run any queued releases once this one is handled
	x.RunQueuedReleases(release)
}

func (x *CodeAmp) RunQueuedReleases(release *model.Release) error {
	log.Warn("RunQueuedReleases")

	startableRelease, startableReleaseEvent, err := x.GetStartableRelease(release, nil)
	if err == nil {
		err := x.StartRelease(startableRelease, startableReleaseEvent)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		log.Error(err)
		return err
	}	

	return nil
}
