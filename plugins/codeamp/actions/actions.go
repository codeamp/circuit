package actions

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
)

type Actions struct {
	events chan transistor.Event
	db     *gorm.DB
}

func NewActions(events chan transistor.Event, db *gorm.DB) *Actions {
	return &Actions{
		events: events,
		db:     db,
	}
}

func (x *Actions) HeartBeat(tick string) {
	var projects []models.Project

	x.db.Find(&projects)
	for _, project := range projects {
		if tick == "minute" {
			x.GitSync(&project)
		}
	}
}

func (x *Actions) GitSync(project *models.Project) {
	var feature models.Feature
	var release models.Release
	hash := ""

	// Get latest release and deployed feature hash
	if x.db.Where("project_id = ?", project.ID).Order("created_at DESC").First(&release).RecordNotFound() {
		// get latest feature if there is no releases
		x.db.Where("project_id = ?", project.ID).Order("created DESC").First(&feature)
		hash = feature.Hash
	} else {
		hash = release.HeadFeature.Hash
	}

	gitSync := plugins.GitSync{
		Action: plugins.Update,
		State:  plugins.Waiting,
		Project: plugins.Project{
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Git: plugins.Git{
			Url:           project.GitUrl,
			Protocol:      project.GitProtocol,
			Branch:        "master",
			RsaPrivateKey: project.RsaPrivateKey,
			RsaPublicKey:  project.RsaPublicKey,
		},
		From: hash,
	}

	x.events <- transistor.NewEvent(gitSync, nil)
}

func (x *Actions) GitCommit(commit plugins.GitCommit) {
	project := models.Project{}
	feature := models.Feature{}

	if x.db.Where("project_id = ? AND hash = ?", project.ID, commit.Hash).First(&feature).RecordNotFound() {
		feature = models.Feature{
			ProjectId:  project.ID,
			Message:    commit.Message,
			User:       commit.User,
			Hash:       commit.Hash,
			ParentHash: commit.ParentHash,
			Ref:        commit.Ref,
			Created:    commit.Created,
		}
		x.db.Save(&feature)

		wsMsg := plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/features", project.Slug),
			Payload: feature,
		}
		x.events <- transistor.NewEvent(wsMsg, nil)
	} else {
		log.InfoWithFields("feature already exists", log.Fields{
			"repository": commit.Repository,
			"hash":       commit.Hash,
		})
	}
}

func (x *Actions) ProjectCreated(project *models.Project) {
	wsMsg := plugins.WebsocketMsg{
		Event:   "projects",
		Payload: project,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ServiceCreated(service *models.Service) {
	project := models.Project{}
	if x.db.Where("id = ?", service.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": service,
		})
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/services/new", project.Slug),
		Payload: service,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ServiceUpdated(service *models.Service) {
	project := models.Project{}
	if x.db.Where("id = ?", service.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": service,
		})
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/services/updated", project.Slug),
		Payload: service,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ServiceDeleted(service *models.Service) {
	project := models.Project{}
	if x.db.Where("id = ?", service.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": service,
		})
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/services/deleted", project.Slug),
		Payload: service,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ServiceSpecCreated(service *models.ServiceSpec) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("serviceSpecs/new"),
		Payload: service,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ServiceSpecDeleted(service *models.ServiceSpec) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("serviceSpecs/deleted"),
		Payload: service,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ServiceSpecUpdated(service *models.ServiceSpec) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("serviceSpecs/updated"),
		Payload: service,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ExtensionSpecCreated(extensionSpec *models.ExtensionSpec) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("extensionSpecs/new"),
		Payload: extensionSpec,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ExtensionSpecDeleted(extensionSpec *models.ExtensionSpec) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("extensionSpecs/deleted"),
		Payload: extensionSpec,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ExtensionSpecUpdated(extensionSpec *models.ExtensionSpec) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("extensionSpecs/updated"),
		Payload: extensionSpec,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentCreated(env *models.Environment) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("environments/new"),
		Payload: env,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentUpdated(env *models.Environment) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("environments/updated"),
		Payload: env,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentDeleted(env *models.Environment) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("environments/deleted"),
		Payload: env,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentVariableCreated(envVar *models.EnvironmentVariable) {
	/*
		sends a websocket message to notify the env var has been created
		input: env var
		emits: plugins.WebSocketMsg
	*/

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("environmentVariables/created"),
		Payload: envVar,
	}
	if envVar.Scope == plugins.ProjectScope {
		project := models.Project{}
		if x.db.Where("id = ?", envVar.ProjectId).First(&project).RecordNotFound() {
			log.InfoWithFields("project not found", log.Fields{
				"service": envVar,
			})
		}

		wsMsg = plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/environmentVariables/created", project.Slug),
			Payload: envVar,
		}
	}

	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentVariableDeleted(envVar *models.EnvironmentVariable) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("environmentVariables/deleted"),
		Payload: envVar,
	}
	if envVar.Scope == plugins.ProjectScope {

		project := models.Project{}
		if x.db.Where("id = ?", envVar.ProjectId).First(&project).RecordNotFound() {
			log.InfoWithFields("envvar not found", log.Fields{
				"service": envVar,
			})
		}

		wsMsg = plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/environmentVariables/deleted", project.Slug),
			Payload: envVar,
		}
	}

	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentVariableUpdated(envVar *models.EnvironmentVariable) {
	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("environmentVariables/updated"),
		Payload: envVar,
	}
	if envVar.Scope == plugins.ProjectScope {
		project := models.Project{}
		if x.db.Where("id = ?", envVar.ProjectId).First(&project).RecordNotFound() {
			log.InfoWithFields("envvar not found", log.Fields{
				"envVar": envVar,
			})
		}

		wsMsg = plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/environmentVariables/updated", project.Slug),
			Payload: envVar,
		}
	}

	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ExtensionCreated(extension *models.Extension) {

	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}

	if x.db.Where("id = ?", extension.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"extension": extension,
		})
	}

	if x.db.Where("id = ?", extension.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extensionSpec not found", log.Fields{
			"extension": extension,
		})
	}

	payload := map[string]interface{}{
		"extension":     extension,
		"extensionSpec": extensionSpec,
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/extensions/created", project.Slug),
		Payload: payload,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)

	eventExtension := plugins.Extension{
		Action:       plugins.Create,
		Slug:         extension.Slug,
		State:        plugins.Waiting,
		StateMessage: "onCreate",
		FormValues:   extension.FormSpecValues,
		Artifacts:    map[string]*string{},
	}

	x.events <- transistor.NewEvent(eventExtension, nil)
}

func (x *Actions) ExtensionUpdated(extension *models.Extension) {
	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}

	if x.db.Where("id = ?", extension.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"extension": extension,
		})
	}

	if x.db.Where("id = ?", extension.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extensionSpec not found", log.Fields{
			"extension": extension,
		})
	}

	payload := map[string]interface{}{
		"extension":     extension,
		"extensionSpec": extensionSpec,
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/extensions/updated", project.Slug),
		Payload: payload,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)

	eventExtension := plugins.Extension{
		Action:       plugins.Update,
		Slug:         extension.Slug,
		State:        plugins.Waiting,
		StateMessage: "onUpdate",
		FormValues:   extension.FormSpecValues,
		Artifacts:    map[string]*string{},
	}
	x.events <- transistor.NewEvent(eventExtension, nil)
}

func (x *Actions) ExtensionDeleted(extension *models.Extension) {
	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}

	if x.db.Where("id = ?", extension.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"extension": extension,
		})
	}

	if x.db.Where("id = ?", extension.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extensionSpec not found", log.Fields{
			"extension": extension,
		})
	}

	payload := map[string]interface{}{
		"extension":     extension,
		"extensionSpec": extensionSpec,
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/extensions/deleted", project.Slug),
		Payload: payload,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ExtensionInitCompleted(extension *models.Extension) {
	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}

	if x.db.Where("id = ?", extension.ProjectId).Find(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"extension": extension,
		})
		return
	}

	if x.db.Where("id = ?", extension.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extensionSpec not found", log.Fields{
			"extension": extension,
		})
	}

	payload := map[string]interface{}{
		"extension":     extension,
		"extensionSpec": extensionSpec,
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/extensions/initCompleted", project.Slug),
		Payload: payload,
	}

	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ReleaseExtensionCompleted(re *models.ReleaseExtension) {
	project := models.Project{}
	release := models.Release{}

	fellowReleaseExtensions := []models.ReleaseExtension{}

	re.State = plugins.Complete
	re.StateMessage = "Finished"
	x.db.Save(&re)

	if x.db.Where("id = ?", re.ReleaseId).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.db.Where("release_id = ? and type = ?", re.ReleaseId, re.Type).Find(&fellowReleaseExtensions).RecordNotFound() {
		log.InfoWithFields("fellow release extensions not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	wsMsg := plugins.WebsocketMsg{
		Event: fmt.Sprintf("projects/%s/releases/releaseExtensionComplete", project.Slug),
		Payload: map[string]interface{}{
			"releaseExtension": re,
		},
	}

	x.events <- transistor.NewEvent(wsMsg, nil)

	// loop through and check if all release extensions are completed
	done := true
	for _, fre := range fellowReleaseExtensions {
		if fre.Type == re.Type && fre.State != plugins.Complete {
			done = false
		}
	}

	if done {
		switch re.Type {
		case plugins.Workflow:
			x.WorkflowExtensionsCompleted(&release)
		case plugins.Deployment:
			x.DeploymentExtensionsCompleted(&release)
		}
	}
}

func (x *Actions) ReleaseExtensionsCompleted(release *models.Release) {

	project := models.Project{}

	release.StateMessage = "Finished"
	release.State = plugins.Complete

	x.db.Save(&release)

	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	// send notif to client
	wsMsg := plugins.WebsocketMsg{
		Event: fmt.Sprintf("projects/%s/releases/releaseExtensionsCompleted", project.Slug),
		Payload: map[string]interface{}{
			"release": release,
		},
	}

	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) WorkflowExtensionsCompleted(release *models.Release) {
	// find all related deployment extensions
	depExtensions := []models.Extension{}
	found := false

	if x.db.Where("project_id = ?", release.ProjectId).Find(&depExtensions).RecordNotFound() {
		log.InfoWithFields("deployment extensions not found", log.Fields{
			"release": release,
		})
		return
	}

	for _, de := range depExtensions {
		var extensionSpec models.ExtensionSpec
		if x.db.Where("id = ?", de.ExtensionSpecId).First(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension spec": de,
			})
		}
		if plugins.ExtensionType(extensionSpec.Type) == plugins.Deployment {
			found = true
		}
	}

	// if there are no deployment workflows, then release is complete
	if !found {
		x.ReleaseCompleted(release)
	}

	project := models.Project{}

	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	services := []models.Service{}
	if x.db.Where("project_id = ?", release.ProjectId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"release": release,
		})
	}

	releaseEvent := plugins.Release{
		Action:       plugins.Create,
		State:        plugins.Waiting,
		StateMessage: "create release event",
		Id:           release.Model.ID.String(),
		HeadFeature:  plugins.Feature{},
		TailFeature:  plugins.Feature{},
		User:         "",
		Project: plugins.Project{
			Action:         plugins.Update,
			Slug:           project.Slug,
			Repository:     project.GitUrl,
			NotifyChannels: []string{}, // not sure what channels can be notified with this
		},
	}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range depExtensions {
		extensionSpec := models.ExtensionSpec{}
		if x.db.Where("id= ?", extension.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension": extension,
			})
		}

		if plugins.ExtensionType(extensionSpec.Type) == plugins.Deployment {

			// create ReleaseExtension
			releaseExtension := models.ReleaseExtension{
				ReleaseId:         release.Model.ID,
				FeatureHash:       "",
				ServicesSignature: "",
				SecretsSignature:  "",
				ExtensionId:       extension.Model.ID,
				State:             plugins.Waiting,
				Type:              plugins.Deployment,
				StateMessage:      "initialized",
			}

			x.db.Save(&releaseExtension)

			releaseExtension.Slug = fmt.Sprintf("%s:%s", extensionSpec.Key, releaseExtension.Model.ID.String())

			extensionEvent := plugins.Extension{
				Slug:       extension.Slug,
				FormValues: extension.FormSpecValues,
				Artifacts:  extension.Artifacts,
			}

			releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
				Id:           releaseExtension.Model.ID.String(),
				Slug:         releaseExtension.Slug,
				Action:       plugins.Create,
				State:        releaseExtension.State,
				Artifacts:    map[string]*string{},
				Key:          extensionSpec.Key,
				Release:      releaseEvent,
				Extension:    extensionEvent,
				StateMessage: releaseExtension.StateMessage,
			})

		}
	}

	// send out release extension event for each re
	for _, re := range releaseExtensionEvents {
		x.events <- transistor.NewEvent(re, nil)
	}
}

func (x *Actions) DeploymentExtensionsCompleted(release *models.Release) {
	x.ReleaseCompleted(release)
}

func (x *Actions) ReleaseCompleted(release *models.Release) {
	project := models.Project{}
	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	// mark release as complete
	(*release).State = plugins.Complete
	(*release).StateMessage = "Release completed"

	x.db.Save(release)

	payload := map[string]interface{}{
		"release": release,
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/releases/completed", project.Slug),
		Payload: payload,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) ReleaseCreated(release *models.Release) {
	project := models.Project{}

	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	payload := map[string]interface{}{
		"release": release,
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/releases/created", project.Slug),
		Payload: payload,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)

	// loop through extensions and send ReleaseWorkflow events
	projectExtensions := []models.Extension{}
	if x.db.Where("project_id = ?", release.ProjectId).Find(&projectExtensions).RecordNotFound() {
		log.InfoWithFields("no extensions to be found", log.Fields{
			"release": release,
			"project": project,
		})
	}

	services := []models.Service{}
	if x.db.Where("project_id = ?", release.ProjectId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"release": release,
		})
	}

	releaseEvent := plugins.Release{
		Action:       plugins.Create,
		State:        plugins.Waiting,
		StateMessage: "create release event",
		Id:           release.Model.ID.String(),
		HeadFeature:  plugins.Feature{},
		TailFeature:  plugins.Feature{},
		User:         "",
		Project: plugins.Project{
			Action:         plugins.Update,
			Slug:           project.Slug,
			Repository:     project.GitUrl,
			NotifyChannels: []string{}, // not sure what channels can be notified with this
		},
	}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range projectExtensions {
		extensionSpec := models.ExtensionSpec{}
		if x.db.Where("id= ?", extension.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension": extension,
			})
		}

		if plugins.ExtensionType(extensionSpec.Type) == plugins.Workflow {
			// create ReleaseExtension
			releaseExtension := models.ReleaseExtension{
				ReleaseId:         release.Model.ID,
				FeatureHash:       "",
				ServicesSignature: "",
				SecretsSignature:  "",
				ExtensionId:       extension.Model.ID,
				State:             plugins.Waiting,
				Type:              plugins.Workflow,
				StateMessage:      "initialized",
			}

			x.db.Save(&releaseExtension)

			releaseExtension.Slug = fmt.Sprintf("%s|%s", extensionSpec.Key, releaseExtension.Model.ID.String())

			extensionEvent := plugins.Extension{
				Slug:       extension.Slug,
				FormValues: extension.FormSpecValues,
				Artifacts:  extension.Artifacts,
			}

			releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
				Id:           releaseExtension.Model.ID.String(),
				Slug:         releaseExtension.Slug,
				Action:       plugins.Create,
				State:        releaseExtension.State,
				Artifacts:    map[string]*string{},
				Key:          extensionSpec.Key,
				Release:      releaseEvent,
				Extension:    extensionEvent,
				StateMessage: releaseExtension.StateMessage,
			})
		}
	}

	// send out release extension event for each re
	for _, re := range releaseExtensionEvents {
		x.events <- transistor.NewEvent(re, nil)
	}
}
