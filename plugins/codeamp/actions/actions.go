package actions

import (
	"fmt"
	"strings"

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

	if x.db.Where("repository = ?", commit.Repository).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"repository": commit.Repository,
		})
		return
	}

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

func (x *Actions) EnvironmentVariableCreated(envVar *models.EnvironmentVariable) {
	project := models.Project{}
	if x.db.Where("id = ?", envVar.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": envVar,
		})
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/environmentVariables/created", project.Slug),
		Payload: envVar,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentVariableDeleted(envVar *models.EnvironmentVariable) {
	project := models.Project{}
	if x.db.Where("id = ?", envVar.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("envvar not found", log.Fields{
			"service": envVar,
		})
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/environmentVariables/deleted", project.Slug),
		Payload: envVar,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}

func (x *Actions) EnvironmentVariableUpdated(envVar *models.EnvironmentVariable) {
	project := models.Project{}
	if x.db.Where("id = ?", envVar.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("envvar not found", log.Fields{
			"envVar": envVar,
		})
	}

	wsMsg := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/environmentVariables/updated", project.Slug),
		Payload: envVar,
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

	if x.db.Where("release_id = ?", re.ReleaseId).Find(&fellowReleaseExtensions).RecordNotFound() {
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
	for _, re := range fellowReleaseExtensions {
		if re.State != plugins.Complete {
			done = false
		}
	}

	if done {
		x.ReleaseExtensionsCompleted(&release)
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

	extensionEvents := []plugins.Extension{}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range projectExtensions {

		extensionSplitSlug := strings.Split(extension.Slug, "|")

		// create ReleaseExtension
		releaseExtension := models.ReleaseExtension{
			ReleaseId:         release.Model.ID,
			FeatureHash:       "",
			ServicesSignature: "",
			SecretsSignature:  "",
			ExtensionId:       extension.Model.ID,
			State:             plugins.Waiting,
			StateMessage:      "initialized",
		}

		x.db.Save(&releaseExtension)

		releaseExtension.Slug = fmt.Sprintf("%s|%s", extensionSplitSlug[0], releaseExtension.Model.ID.String())

		extensionEvents = append(extensionEvents, plugins.Extension{
			Slug:       extension.Slug,
			FormValues: extension.FormSpecValues,
			Artifacts:  extension.Artifacts,
		})
		releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
			Id:           releaseExtension.Model.ID.String(),
			Slug:         releaseExtension.Slug,
			State:        releaseExtension.State,
			StateMessage: releaseExtension.StateMessage,
		})
	}

	releaseEvent := plugins.Release{
		Action:            plugins.Create,
		State:             plugins.Waiting,
		StateMessage:      "create release event",
		Id:                release.Model.ID.String(),
		HeadFeature:       plugins.Feature{},
		TailFeature:       plugins.Feature{},
		User:              "",
		ReleaseExtensions: releaseExtensionEvents,
		Extensions:        extensionEvents,
		Project: plugins.Project{
			Action:         plugins.Update,
			Slug:           project.Slug,
			Repository:     project.GitUrl,
			NotifyChannels: []string{}, // not sure what channels can be notified with this
		},
	}

	x.events <- transistor.NewEvent(releaseEvent, nil)

}
