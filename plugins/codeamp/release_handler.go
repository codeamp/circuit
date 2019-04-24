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
)

func (x *CodeAmp) releasedInDependentEnvironments(e transistor.Event) (bool, error) {
	payload := e.Payload.(plugins.Release)
	primaryEnvironmentName := "production"
	secondaryEnvironmentName := "development"

	log.Println("CHECKING IF DEPENDENT ENVIRONMENT IS RELEASED")

	primaryEnvironment := model.Environment{}
	secondaryEnvironment := model.Environment{}

	if x.DB.Where("name = ?", primaryEnvironmentName).First(&primaryEnvironment).RecordNotFound() {
		log.Error("could not find primary environment")
	}
	if x.DB.Where("name = ?", secondaryEnvironmentName).First(&secondaryEnvironment).RecordNotFound() {
		log.Error("could not find secondary environment")
	}

	if payload.Environment != primaryEnvironment.Name {
		log.Info("No required environments set, nothing to check")
		return true, nil
	}
	log.Println("release is for primary environment. Checking release in qualifying environments")

	// check project exists in both primary and secondary environment
	primaryProjectEnvironment := model.ProjectEnvironment{}
	secondaryProjectEnvironment := model.ProjectEnvironment{}

	if x.DB.Where("environment_id = ? and project_id = ?", primaryEnvironment.ID, payload.Project.ID).First(&primaryProjectEnvironment).RecordNotFound() {
		log.Error("project is not configured in the primary environment")
	}
	if x.DB.Where("environment_id = ? and project_id = ?", secondaryEnvironment.ID, payload.Project.ID).First(&secondaryProjectEnvironment).RecordNotFound() {
		log.Error("project is not configured in the secondary environment")
	}

	releaseQuery := fmt.Sprintf("environment_id = '%s' and project_id = '%s' and head_feature_id = '%s' and state = '%s'",
		secondaryProjectEnvironment.EnvironmentID, secondaryProjectEnvironment.ProjectID,
		payload.HeadFeature.ID,
		string(transistor.GetState("complete")))

	log.Println(releaseQuery)

	//TODO: Handle project not defined in each environment
	secondaryRelease := model.Release{}
	if x.DB.Where(releaseQuery).First(&secondaryRelease).RecordNotFound() {

		log.Println(secondaryRelease.State)

		log.Error(fmt.Sprintf("RELEASE CREATED IN %s WITHOUT BEING DEPLOYED TO %s", primaryEnvironmentName, secondaryEnvironmentName))

		return false, nil
	}
	return true, nil
}

func (x *CodeAmp) alertReleaseNotReady(e transistor.Event) error {
	//TODO: Dispatch alert to user/team
	return nil
}

func (x *CodeAmp) checkDependentEnvironmentsDeployed(e transistor.Event) (bool, []error) {
	payload := e.Payload.(plugins.Release)

	projectSettings := model.ProjectSettings{}
	environment := model.Environment{}

	if x.DB.Where("name = ?", payload.Environment).First(&environment).RecordNotFound(){
		log.ErrorWithFields("release environment not found", log.Fields{
			"id": payload.Environment
		})
	}

	if x.DB.Where("environment_id = ? and project_id = ?", environment.ID, payload.Project.ID).First(&projectSettings).RecordNotFound()}{
		log.Error("project settings do not exist for this release")
	}

	undeployedDependencies := []model.Environment{}
	for _, id := range projectSettings.DependentEnvironments {

		releaseQuery := fmt.Sprintf("environment_id = '%s' and project_id = '%s' and head_feature_id = '%s' and state = '%s'",
			secondaryProjectEnvironment.EnvironmentID, secondaryProjectEnvironment.ProjectID,
			payload.HeadFeature.ID,
			string(transistor.GetState("complete")))
		
		release := model.Release{}
		x.DB.Where(releaseQuery).First(&release).RecordNotFound(){
			log.error("Release created without first deploying to dependencies")
		}
		
	}
}

func (x *CodeAmp) PassesPreReleaseChecks(e transistor.Event) (bool, []error) {
	// dependent environments have been deployed

}

func (x *CodeAmp) ReleaseEventHandler(e transistor.Event) error {
	// var err error
	payload := e.Payload.(plugins.Release)
	release := model.Release{}
	releaseExtensions := []model.ReleaseExtension{}

	if x.DB.Where("id = ?", payload.ID).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"id": payload.ID,
		})
		return fmt.Errorf("release %s not found", payload.ID)
	}

	if e.Matches("release:create") {
		x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions)

		readyToRelease, err := x.releasedInDependentEnvironments(e)
		if err != nil {
			log.Error(fmt.Errorf("failed to validate release %s status in dependent environments: %s", payload.ID, err.Error()))
		}
		if readyToRelease != true {
			err = x.alertReleaseNotReady(e)
			if err != nil {
				log.Error(fmt.Sprintf("failed to alert on release not ready: %s", err.Error()))
			}
		}

		for _, releaseExtension := range releaseExtensions {
			projectExtension := model.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.InfoWithFields("project extensions not found", log.Fields{
					"id":                   releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return fmt.Errorf("project extension %s not found", releaseExtension.ProjectExtensionID)
			}

			extension := model.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.InfoWithFields("extension not found", log.Fields{
					"id":                   projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return fmt.Errorf("extension %s not found", projectExtension.ExtensionID)
			}

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

				if !release.ForceRebuild && extension.Cacheable {
					// query for the most recent complete release extension that has the same services, secrets and feature hash as this one
					err = x.DB.Where("project_extension_id = ? and services_signature = ? and secrets_signature = ? and feature_hash = ? and state in (?)",
						projectExtension.Model.ID, releaseExtension.ServicesSignature,
						releaseExtension.SecretsSignature, releaseExtension.FeatureHash,
						[]string{"complete"}).Order("created_at desc").First(&lastReleaseExtension).Error
					if err != nil {
						log.Error(err.Error())
					} else {
						eventAction = transistor.GetAction("status")
						eventState = lastReleaseExtension.State
						eventStateMessage = fmt.Sprintf("Using cache from previous release: %s", lastReleaseExtension.StateMessage)

						err := json.Unmarshal(lastReleaseExtension.Artifacts.RawMessage, &artifacts)
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
					Release: payload,
				}

				ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), eventAction, payload)
				ev.State = eventState
				ev.StateMessage = eventStateMessage
				ev.Artifacts = artifacts

				releaseExtension.Started = time.Now()
				x.DB.Save(&releaseExtension)

				x.Events <- ev
			}
		}
	}
	return nil
}

func (x *CodeAmp) ReleaseFailed(release *model.Release, stateMessage string) {
	release.State = transistor.GetState("failed")
	release.StateMessage = stateMessage
	release.Finished = time.Now()

	project := model.Project{}
	environment := model.Environment{}

	x.DB.Save(release)

	releaseExtensions := []model.ReleaseExtension{}
	x.DB.Where("release_id = ? AND state <> ?", release.Model.ID, transistor.GetState("complete")).Find(&releaseExtensions)
	for _, re := range releaseExtensions {
		re.State = transistor.GetState("failed")
		x.DB.Save(&re)
	}

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

	x.SendNotifications("FAILED", release, &project)

	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key),
		Payload: release,
	}

	event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
	event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key), false)
	x.Events <- event

	x.RunQueuedReleases(release)
}

func (x *CodeAmp) ReleaseCompleted(release *model.Release) {
	project := model.Project{}
	environment := model.Environment{}

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

	x.RunQueuedReleases(release)
}

func (x *CodeAmp) RunQueuedReleases(release *model.Release) error {
	var nextQueuedRelease model.Release

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

	var project model.Project
	var services []model.Service
	var secrets []model.Secret

	projectSecrets := []model.Secret{}
	// get all the env vars related to this release and store
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

	var pluginSecrets []plugins.Secret
	for _, secret := range secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
	}

	// insert CodeAmp envs
	slugSecret := plugins.Secret{
		Key:   "CODEAMP_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, slugSecret)

	hashSecret := plugins.Secret{
		Key:   "CODEAMP_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, hashSecret)

	timeSecret := plugins.Secret{
		Key:   "CODEAMP_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, timeSecret)

	// insert Codeflow envs - remove later
	_slugSecret := plugins.Secret{
		Key:   "CODEFLOW_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _slugSecret)

	_hashSecret := plugins.Secret{
		Key:   "CODEFLOW_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _hashSecret)

	_timeSecret := plugins.Secret{
		Key:   "CODEFLOW_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _timeSecret)

	releasePayload := graphql_resolver.BuildReleasePayload(nextQueuedRelease, project, environment, branch, headFeature, tailFeature, pluginServices, pluginSecrets)

	nextQueuedRelease.Started = time.Now()
	nextQueuedRelease.State = transistor.GetState("running")
	nextQueuedRelease.StateMessage = "Running Release"
	x.DB.Save(&nextQueuedRelease)

	x.Events <- transistor.NewEvent(plugins.GetEventName("release"), transistor.GetAction("create"), releasePayload)
	return nil
}
