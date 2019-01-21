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

	"github.com/davecgh/go-spew/spew"
)

func (x *CodeAmp) ReleaseEventHandler(e transistor.Event) error {
	releaseEvent := e.Payload.(plugins.Release)

	if e.Matches("release:create") {
		release := model.Release{}

		if x.DB.Where("id = ?", releaseEvent.ID).First(&release).RecordNotFound() {
			log.InfoWithFields("release not found", log.Fields{
				"id": releaseEvent.ID,
			})
			return fmt.Errorf("release %s not found", releaseEvent.ID)
		}

		if x.IsReleaseStartable(&release) == true {
			err := x.StartRelease(&release, &releaseEvent)
			if err != nil {
				log.Error(err)
			}
		}
	}

	return nil
}

func (x *CodeAmp) IsReleaseStartable(release *model.Release) bool {
	var releases []model.Release
	if err := x.DB.
		Where("project_id = ? and environment_id = ? and state IN ('running', 'waiting')", release.ProjectID, release.EnvironmentID).
		Order("created_at asc").
		Limit(1).
		Find(&releases).
		Error; err != nil {
		log.Error(err)
	}
	return (releases[0].ID == release.ID && release.State == transistor.GetState("waiting"))
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

	err = x.DB.Save(&release).Error
	if err != nil {
		return err
	}

	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", releaseEvent.Project.Slug, releaseEvent.Environment),
		Payload: release,
	}

	spew.Dump(payload)

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
		err := x.DB.Save(&re).Error
		if err != nil {
			return err
		}

		var artifacts []transistor.Artifact
		artifacts, err = graphql_resolver.ExtractArtifacts(projectExtension, extension, x.DB)
		if err != nil {
			log.Error(err.Error())
			return nil
		}	

		spew.Dump(artifacts)

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

	x.DB.Save(release)

	// Mark all release extensions as failed
	releaseExtensions := []model.ReleaseExtension{}
	x.DB.Where("release_id = ? AND state <> ?", release.Model.ID, transistor.GetState("complete")).Find(&releaseExtensions)
	for _, re := range releaseExtensions {
		re.State = transistor.GetState("failed")
		x.DB.Save(&re)
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
	// x.RunQueuedReleases(release)
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

		x.DB.Save(release)

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
	
	var nextQueuedRelease model.Release

	/******************************************
	*
	*	Check for waiting/queued releases
	*
	*******************************************/
	// If there are no queued/waiting releases, no work to be done here.
	x.DB.LogMode(true)
	defer x.DB.LogMode(false)
	if x.DB.Where("id != ? and state = ? and project_id = ? and environment_id = ? and created_at > ?", release.Model.ID, "waiting", release.ProjectID, release.EnvironmentID, release.CreatedAt).Order("created_at asc").First(&nextQueuedRelease).RecordNotFound() {
		log.WarnWithFields("No queued releases found.", log.Fields{
			"id":             release.Model.ID,
			"state":          "waiting",
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
			"created_at":     release.CreatedAt,
		})
		return nil
	}
	x.DB.LogMode(false)

	var project model.Project
	var services []model.Service
	var secrets []model.Secret

	projectSecrets := []model.Secret{}
	// get all the env vars/secrets related to this release and store
	x.DB.Where("environment_id = ? AND project_id = ? AND scope = ?", nextQueuedRelease.EnvironmentID, nextQueuedRelease.ProjectID, "project").Find(&projectSecrets)
	for _, secret := range projectSecrets {
		var secretValue model.SecretValue
		x.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	globalSecrets := []model.Secret{}
	x.DB.Where("environment_id = ? AND scope = ?", nextQueuedRelease.EnvironmentID, "global").Find(&globalSecrets)
	for _, secret := range globalSecrets {
		var secretValue model.SecretValue
		x.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	x.DB.Where("project_id = ? and environment_id = ?", nextQueuedRelease.ProjectID, nextQueuedRelease.EnvironmentID).Find(&services)
	if len(services) == 0 {
		log.WarnWithFields("no services found", log.Fields{
			"project_id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	if x.DB.Where("id = ?", nextQueuedRelease.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	for i, service := range services {
		ports := []model.ServicePort{}
		x.DB.Where("service_id = ?", service.Model.ID).Find(&ports)
		services[i].Ports = ports
	}

	if x.DB.Where("id = ?", nextQueuedRelease.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	/******************************************
	*
	*	Gather ProjectSettings, Environment,
	*	Head/Tail Features, and all Services
	*
	*******************************************/
	// get all branches relevant for the project
	var branch string
	var projectSettings model.ProjectSettings

	if x.DB.Where("environment_id = ? and project_id = ?", nextQueuedRelease.EnvironmentID, nextQueuedRelease.ProjectID).First(&projectSettings).RecordNotFound() {
		log.WarnWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var environment model.Environment
	if x.DB.Where("id = ?", nextQueuedRelease.EnvironmentID).Find(&environment).RecordNotFound() {
		log.WarnWithFields("no env found", log.Fields{
			"id": nextQueuedRelease.EnvironmentID,
		})
		return nil
	}

	var headFeature model.Feature
	if x.DB.Where("id = ?", nextQueuedRelease.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.WarnWithFields("head feature not found", log.Fields{
			"id": nextQueuedRelease.HeadFeatureID,
		})
		return nil
	}

	var tailFeature model.Feature
	if x.DB.Where("id = ?", nextQueuedRelease.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.WarnWithFields("tail feature not found", log.Fields{
			"id": nextQueuedRelease.TailFeatureID,
		})
		return nil
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec model.ServiceSpec
		if x.DB.Where("service_id = ?", service.Model.ID).First(&spec).RecordNotFound() {
			log.WarnWithFields("servicespec not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return nil
		}

		pluginServices = graphql_resolver.AppendPluginService(pluginServices, service, spec)
	}

	/******************************************
	*
	*	Insert Environment Variables
	*
	*******************************************/
	var pluginSecrets []plugins.Secret
	for _, secret := range secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
	}

	// pluginSecrets = x.injectReleaseEnvVars(pluginSecrets, &project, headFeature, tailFeature)

	/******************************************
	*
	*	Build Release Payload
	*
	*******************************************/
	releasePayload := graphql_resolver.BuildReleasePayload(nextQueuedRelease, project, environment, branch, headFeature, tailFeature, pluginServices, pluginSecrets)

	nextQueuedRelease.Started = time.Now()
	x.DB.Save(&nextQueuedRelease)

	var err error
	releaseExtensions := []model.ReleaseExtension{}

	x.DB.Where("release_id = ?", nextQueuedRelease.Model.ID).Find(&releaseExtensions)
	for _, releaseExtension := range releaseExtensions {
		// Find associated project extensions
		projectExtension := model.ProjectExtension{}
		if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
			log.InfoWithFields("project extensions not found", log.Fields{
				"id": releaseExtension.ProjectExtensionID,
				"release_extension_id": releaseExtension.Model.ID,
			})
			return fmt.Errorf("project extension %s not found", releaseExtension.ProjectExtensionID)
		}

		// Find Extensions to match the searched for Project Extensions
		extension := model.Extension{}
		if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
			log.InfoWithFields("extension not found", log.Fields{
				"id": projectExtension.Model.ID,
				"release_extension_id": releaseExtension.Model.ID,
			})
			return fmt.Errorf("extension %s not found", projectExtension.ExtensionID)
		}

		// Mark the workflow extensions as having 'started' now
		if plugins.Type(extension.Type) == plugins.GetType("workflow") {
			// check if the last release extension has the same
			// ServicesSignature and SecretsSignature. If so,
			// mark the action as completed before sending the event
			lastReleaseExtension := model.ReleaseExtension{}
			artifacts := []transistor.Artifact{}

			eventAction := transistor.GetAction("create")
			eventState := transistor.GetState("waiting")
			eventStateMessage := ""
			needsExtract := true

			// If this extension is cacheable and there hasn't been an explicit rebuild,
			// try to find a previous release and use its extension configuration
			// e.g. dockerbuilder use the configuration the last release's dockerbuilder used.
			if !release.ForceRebuild && extension.Cacheable {
				// query for the most recent complete release extension that has the same services, secrets and feature hash as this one
				err = x.DB.Where("project_extension_id = ? and services_signature = ? and secrets_signature = ? and feature_hash = ? and state in (?)",
					projectExtension.Model.ID, releaseExtension.ServicesSignature,
					releaseExtension.SecretsSignature, releaseExtension.FeatureHash,
					[]string{"complete"}).Order("created_at desc").First(&lastReleaseExtension).Error
				if err != nil {
					eventAction = transistor.GetAction("status")
					eventState = lastReleaseExtension.State
					eventStateMessage = lastReleaseExtension.StateMessage

					err := json.Unmarshal(lastReleaseExtension.Artifacts.RawMessage, &artifacts)
					if err != nil {
						log.Error(err.Error())
						return nil
					}
					eventAction = transistor.GetAction("status")
					eventState = lastReleaseExtension.State
					eventStateMessage = lastReleaseExtension.StateMessage

					err = json.Unmarshal(lastReleaseExtension.Artifacts.RawMessage, &artifacts)
					if err != nil {
						log.Error(err.Error())
						return nil
					}

					needsExtract = false
				}
			}

			if needsExtract {
				artifacts, err = graphql_resolver.ExtractArtifacts(projectExtension, extension, x.DB)
				if err != nil {
					log.Error(err.Error())
					return nil
				}
			}

			payload := plugins.ReleaseExtension{
				ID:      releaseExtension.Model.ID.String(),
				Release: releasePayload,
			}

			ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), eventAction, payload)
			ev.State = eventState
			ev.StateMessage = eventStateMessage
			ev.Artifacts = artifacts

			// Set release extension as 'started' and notify the plugin it has started
			releaseExtension.Started = time.Now()
			x.DB.Save(&releaseExtension)

			x.Events <- ev
		}
	}

	return nil
}
