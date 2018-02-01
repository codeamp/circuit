package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
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
		x.db.Where("project_id = ?", project.ID).Order("created_at DESC").First(&feature)
		hash = feature.Hash
	} else {
		hash = release.HeadFeature.Hash
	}

	// get branches of entire environments
	envProjectBranches := []models.EnvironmentBasedProjectBranch{}
	if x.db.Where("project_id = ?", project.Model.ID.String()).Find(&envProjectBranches).RecordNotFound() {
		log.Info("no env project branches found")
		gitSync := plugins.GitSync{
			Action: plugins.GetAction("update"),
			State:  plugins.GetState("waiting"),
			Project: plugins.Project{
				Id:         project.Model.ID.String(),
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
	} else {
		for _, envProjectBranch := range envProjectBranches {
			gitSync := plugins.GitSync{
				Action: plugins.GetAction("update"),
				State:  plugins.GetState("waiting"),
				Project: plugins.Project{
					Id:         project.Model.ID.String(),
					Repository: project.Repository,
				},
				Git: plugins.Git{
					Url:           project.GitUrl,
					Protocol:      project.GitProtocol,
					Branch:        envProjectBranch.GitBranch,
					RsaPrivateKey: project.RsaPrivateKey,
					RsaPublicKey:  project.RsaPublicKey,
				},
				From: hash,
			}

			x.events <- transistor.NewEvent(gitSync, nil)
		}
	}
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
	} else {
		log.InfoWithFields("feature already exists", log.Fields{
			"repository": commit.Repository,
			"hash":       commit.Hash,
		})
	}
}

func (x *Actions) GitBranch(branch plugins.GitBranch) {
	project := models.Project{}
	gitBranch := models.GitBranch{}

	if x.db.Where("repository = ?", branch.Repository).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"repository": branch.Repository,
		})
		return
	}

	if x.db.Where("project_id = ? and name = ?", project.ID, branch.Name).First(&gitBranch).RecordNotFound() {
		gitBranch = models.GitBranch{
			ProjectId: project.ID,
			Name:      branch.Name,
		}
		x.db.Save(&gitBranch)
	} else {
		log.InfoWithFields("branch already exists", log.Fields{
			"repository": branch.Repository,
			"name":       branch.Name,
		})
	}
}

func (x *Actions) ProjectCreated(project *models.Project) {

}

func (x *Actions) ServiceCreated(service *models.Service) {
	project := models.Project{}
	if x.db.Where("id = ?", service.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": service,
		})
	}
}

func (x *Actions) ServiceUpdated(service *models.Service) {
	project := models.Project{}
	if x.db.Where("id = ?", service.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": service,
		})
	}
}

func (x *Actions) ServiceDeleted(service *models.Service) {
	project := models.Project{}
	if x.db.Where("id = ?", service.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"service": service,
		})
	}
}

func (x *Actions) EnvironmentBasedProjectBranchCreated(service *models.EnvironmentBasedProjectBranch) {
}

func (x *Actions) EnvironmentBasedProjectBranchUpdated(service *models.EnvironmentBasedProjectBranch) {
}

func (x *Actions) EnvironmentBasedProjectBranchDeleted(service *models.EnvironmentBasedProjectBranch) {
}

func (x *Actions) ServiceSpecCreated(service *models.ServiceSpec) {
}

func (x *Actions) ServiceSpecDeleted(service *models.ServiceSpec) {
}

func (x *Actions) ServiceSpecUpdated(service *models.ServiceSpec) {
}

func (x *Actions) ExtensionSpecCreated(extensionSpec *models.ExtensionSpec) {
}

func (x *Actions) ExtensionSpecDeleted(extensionSpec *models.ExtensionSpec) {
}

func (x *Actions) ExtensionSpecUpdated(extensionSpec *models.ExtensionSpec) {
}

func (x *Actions) EnvironmentCreated(env *models.Environment) {
}

func (x *Actions) EnvironmentUpdated(env *models.Environment) {
}

func (x *Actions) EnvironmentDeleted(env *models.Environment) {
}

func (x *Actions) EnvironmentVariableCreated(envVar *models.EnvironmentVariable) {
}

func (x *Actions) EnvironmentVariableDeleted(envVar *models.EnvironmentVariable) {
}

func (x *Actions) EnvironmentVariableUpdated(envVar *models.EnvironmentVariable) {
}

func (x *Actions) GetSecrets(project models.Project) ([]plugins.Secret, error) {
	secrets := []plugins.Secret{}
	adminEnvVars := []models.EnvironmentVariable{}
	if x.db.Where("scope = ?", "global").Find(&adminEnvVars).RecordNotFound() {
		log.InfoWithFields("no global admin env vars", log.Fields{})
	}
	for _, val := range adminEnvVars {
		evValue := models.EnvironmentVariableValue{}
		if x.db.Where("environment_variable_id = ?", val.Model.ID.String()).Order("created_at desc").First(&evValue).RecordNotFound() {
			log.InfoWithFields("envvar value not found", log.Fields{
				"id": val.Model.ID.String(),
			})
		} else {
			secrets = append(secrets, plugins.Secret{
				Key:   val.Key,
				Value: evValue.Value,
				Type:  val.Type,
			})
		}
	}

	projectEnvVars := []models.EnvironmentVariable{}
	if x.db.Where("scope = ? and project_id = ?", "project", project.Model.ID.String()).Find(&projectEnvVars).RecordNotFound() {
		log.InfoWithFields("no project env vars found", log.Fields{})
	}
	for _, val := range projectEnvVars {
		evValue := models.EnvironmentVariableValue{}
		if x.db.Where("environment_variable_id = ?", val.Model.ID.String()).Order("created_at desc").First(&evValue).RecordNotFound() {
			log.InfoWithFields("envvar value not found", log.Fields{
				"id": val.Model.ID.String(),
			})
		} else {
			secrets = append(secrets, plugins.Secret{
				Key:   val.Key,
				Value: evValue.Value,
				Type:  val.Type,
			})
		}
	}
	return secrets, nil
}

func (x *Actions) GetSecretsAndServicesFromSnapshot(release *models.Release) ([]plugins.Secret, []plugins.Service, error) {
	secrets := []plugins.Secret{}
	unmarshalledSnapshot := map[string]interface{}{}
	err := json.Unmarshal(release.Snapshot.RawMessage, &unmarshalledSnapshot)
	if err != nil {
		log.Info(err.Error())
		return nil, nil , err
	}

	for _, envvar := range unmarshalledSnapshot["environmentVariables"].([]interface{}) {
		key := envvar.(map[string]interface{})["key"].(string)
		val := envvar.(map[string]interface{})["value"].(string)
		evType := plugins.GetType(envvar.(map[string]interface{})["type"].(string))

		secrets = append(secrets, plugins.Secret{
			Key: key,
			Value: val,
			Type: evType,
		})
	}

	pluginServices := []plugins.Service{}
	for _, service := range unmarshalledSnapshot["services"].([]interface{}) {
		pluginListeners := []plugins.Listener{}
		for _, listener := range service.(map[string]interface{})["container_ports"].([]interface{}) {
			intPort, _ := strconv.Atoi(listener.(map[string]interface{})["port"].(string))
			pluginListeners = append(pluginListeners, plugins.Listener{
				Port:     int32(intPort),
				Protocol: listener.(map[string]interface{})["protocol"].(string),
			})
		}

		intTerminationGracePeriod, _ := strconv.Atoi(service.(map[string]interface{})["service_spec"].(map[string]interface{})["termination_grace_period"].(string))
		intReplicas, _ := strconv.Atoi(service.(map[string]interface{})["count"].(string))
		pluginServices = append(pluginServices, plugins.Service{
			Id:        service.(map[string]interface{})["id"].(string),
			Command:   service.(map[string]interface{})["command"].(string),
			Name:      service.(map[string]interface{})["name"].(string),
			Listeners: pluginListeners,
			State:     plugins.GetState("waiting"),
			Spec: plugins.ServiceSpec{
				Id:                            service.(map[string]interface{})["service_spec"].(map[string]interface{})["id"].(string),
				CpuRequest:                    fmt.Sprintf("%sm", service.(map[string]interface{})["service_spec"].(map[string]interface{})["cpu_request"].(string)),
				CpuLimit:                      fmt.Sprintf("%sm", service.(map[string]interface{})["service_spec"].(map[string]interface{})["cpu_limit"].(string)),
				MemoryRequest:                 fmt.Sprintf("%sMi", service.(map[string]interface{})["service_spec"].(map[string]interface{})["memory_request"].(string)),
				MemoryLimit:                   fmt.Sprintf("%sMi", service.(map[string]interface{})["service_spec"].(map[string]interface{})["memory_limit"].(string)),
				TerminationGracePeriodSeconds: int64(intTerminationGracePeriod),
			},
			Type:     string(service.(map[string]interface{})["type"].(string)),
			Replicas: int64(intReplicas),
		})
	}	
	return secrets, pluginServices, nil
}

func (x *Actions) ExtensionCreated(extension *models.Extension) {
	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}
	environment := models.Environment{}

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

	if x.db.Where("id = ?", extension.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("env not found", log.Fields{
			"id": extension.EnvironmentId,
		})
	}

	unmarshalledConfig := make(map[string]interface{})
	err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
	if err != nil {
		log.Info(err.Error())
	}

	formValues, err := utils.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.db)
	if err != nil {
		log.Info(err.Error())
	}

	services := []models.Service{}
	if x.db.Where("project_id = ?", extension.ProjectId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"extension": extension.ProjectId,
		})
	}

	// get env vars in project and admin and insert into secrets
	secrets, err := x.GetSecrets(project)
	if err != nil {
		log.Info(err.Error())
	}

	// get all branches relevant for the projec
	branch := "master"
	envProjectBranch := models.EnvironmentBasedProjectBranch{}
	if x.db.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&envProjectBranch).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = envProjectBranch.GitBranch
	}

	pluginServices := []plugins.Service{}
	for _, service := range services {
		spec := models.ServiceSpec{}
		if x.db.Where("id = ?", service.ServiceSpecId).First(&spec).RecordNotFound() {
			log.InfoWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecId,
			})
			return
		}

		listeners := []models.ContainerPort{}
		if x.db.Where("service_id = ?", service.Model.ID).Find(&listeners).RecordNotFound() {
			log.InfoWithFields("container ports not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return
		}

		pluginListeners := []plugins.Listener{}
		for _, listener := range listeners {
			intPort, _ := strconv.Atoi(listener.Port)
			pluginListeners = append(pluginListeners, plugins.Listener{
				Port:     int32(intPort),
				Protocol: listener.Protocol,
			})
		}

		intTerminationGracePeriod, _ := strconv.Atoi(spec.TerminationGracePeriod)
		intReplicas, _ := strconv.Atoi(service.Count)
		pluginServices = append(pluginServices, plugins.Service{
			Id:        service.Model.ID.String(),
			Command:   service.Command,
			Name:      service.Name,
			Listeners: pluginListeners,
			State:     plugins.GetState("waiting"),
			Spec: plugins.ServiceSpec{
				Id:                            spec.Model.ID.String(),
				CpuRequest:                    fmt.Sprintf("%sm", spec.CpuRequest),
				CpuLimit:                      fmt.Sprintf("%sm", spec.CpuLimit),
				MemoryRequest:                 fmt.Sprintf("%sMi", spec.MemoryRequest),
				MemoryLimit:                   fmt.Sprintf("%sMi", spec.MemoryLimit),
				TerminationGracePeriodSeconds: int64(intTerminationGracePeriod),
			},
			Type:     string(service.Type),
			Replicas: int64(intReplicas),
		})
	}

	eventExtension := plugins.Extension{
		Id:           extension.Model.ID.String(),
		Action:       plugins.GetAction("create"),
		Slug:         extensionSpec.Key,
		State:        plugins.GetState("waiting"),
		StateMessage: "onCreate",
		Config:       formValues,
		Artifacts:    map[string]string{},
		Environment:  environment.Name,
		Project: plugins.Project{
			Id: project.Model.ID.String(),
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        branch,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			Services:   pluginServices,
			Secrets:    secrets,
			Repository: project.Repository,
		},
	}

	x.events <- transistor.NewEvent(eventExtension, nil)
}

func (x *Actions) ExtensionUpdated(extension *models.Extension) {
	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}
	environment := models.Environment{}

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

	if x.db.Where("id = ?", extension.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("env not found", log.Fields{
			"id": extension.EnvironmentId,
		})
	}

	unmarshalledConfig := make(map[string]interface{})
	err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
	if err != nil {
		log.Info(err.Error())
	}

	formValues, err := utils.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.db)
	if err != nil {
		log.Info(err.Error())
	}

	services := []models.Service{}
	if x.db.Where("project_id = ?", extension.ProjectId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"extension": extension.ProjectId,
		})
	}

	secrets, err := x.GetSecrets(project)
	if err != nil {
		log.Info(err.Error())
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	envProjectBranch := models.EnvironmentBasedProjectBranch{}
	if x.db.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(), project.Model.ID.String()).First(&envProjectBranch).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = envProjectBranch.GitBranch
	}

	pluginServices := []plugins.Service{}
	for _, service := range services {
		spec := models.ServiceSpec{}
		if x.db.Where("id = ?", service.ServiceSpecId).First(&spec).RecordNotFound() {
			log.InfoWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecId,
			})
			return
		}

		listeners := []models.ContainerPort{}
		if x.db.Where("service_id = ?", service.Model.ID).Find(&listeners).RecordNotFound() {
			log.InfoWithFields("container ports not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return
		}

		pluginListeners := []plugins.Listener{}
		for _, listener := range listeners {
			intPort, _ := strconv.Atoi(listener.Port)
			pluginListeners = append(pluginListeners, plugins.Listener{
				Port:     int32(intPort),
				Protocol: listener.Protocol,
			})
		}

		intTerminationGracePeriod, _ := strconv.Atoi(spec.TerminationGracePeriod)
		intReplicas, _ := strconv.Atoi(service.Count)
		pluginServices = append(pluginServices, plugins.Service{
			Id:        service.Model.ID.String(),
			Command:   service.Command,
			Name:      service.Name,
			Listeners: pluginListeners,
			State:     plugins.GetState("waiting"),
			Spec: plugins.ServiceSpec{
				Id:                            spec.Model.ID.String(),
				CpuRequest:                    fmt.Sprintf("%sm", spec.CpuRequest),
				CpuLimit:                      fmt.Sprintf("%sm", spec.CpuLimit),
				MemoryRequest:                 fmt.Sprintf("%sMi", spec.MemoryRequest),
				MemoryLimit:                   fmt.Sprintf("%sMi", spec.MemoryLimit),
				TerminationGracePeriodSeconds: int64(intTerminationGracePeriod),
			},
			Type:     string(service.Type),
			Replicas: int64(intReplicas),
		})
	}

	eventExtension := plugins.Extension{
		Id:           extension.Model.ID.String(),
		Action:       plugins.GetAction("update"),
		Slug:         extensionSpec.Key,
		State:        plugins.GetState("waiting"),
		StateMessage: "onUpdate",
		Config:       formValues,
		Artifacts:    map[string]string{},
		Environment:  environment.Name,
		Project: plugins.Project{
			Id: project.Model.ID.String(),
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        branch,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			Services:   pluginServices,
			Secrets:    secrets,
			Repository: project.Repository,
		},
	}

	x.events <- transistor.NewEvent(eventExtension, nil)
}

func (x *Actions) ExtensionDeleted(extension *models.Extension) {
	project := models.Project{}
	extensionSpec := models.ExtensionSpec{}
	environment := models.Environment{}

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

	if x.db.Where("id = ?", extension.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("env not found", log.Fields{
			"id": extension.EnvironmentId,
		})
	}

	unmarshalledConfig := make(map[string]interface{})
	err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
	if err != nil {
		log.Info(err.Error())
	}

	formValues, err := utils.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.db)
	if err != nil {
		log.Info(err.Error())
	}

	services := []models.Service{}
	if x.db.Where("project_id = ?", extension.ProjectId).Find(&services).RecordNotFound() {
		log.InfoWithFields("no services found for this project", log.Fields{
			"extension": extension.ProjectId,
		})
	}

	// get env vars in project and admin and insert into secrets
	secrets := []plugins.Secret{}
	adminEnvVars := []models.EnvironmentVariable{}
	if x.db.Where("scope = ?", "global").Find(&adminEnvVars).RecordNotFound() {
		log.InfoWithFields("no global admin env vars", log.Fields{})
	}
	for _, val := range adminEnvVars {
		evValue := models.EnvironmentVariableValue{}
		if x.db.Where("environment_variable_id = ?", val.Model.ID.String()).Order("created_at desc").First(&evValue).RecordNotFound() {
			log.InfoWithFields("envvar value not found", log.Fields{
				"id": val.Model.ID.String(),
			})
		} else {
			secrets = append(secrets, plugins.Secret{
				Key:   val.Key,
				Value: evValue.Value,
				Type:  val.Type,
			})
		}
	}

	projectEnvVars := []models.EnvironmentVariable{}
	if x.db.Where("scope = ? and project_id = ?", "project", project.Model.ID.String()).Find(&projectEnvVars).RecordNotFound() {
		log.InfoWithFields("no project env vars found", log.Fields{})
	}
	for _, val := range projectEnvVars {
		evValue := models.EnvironmentVariableValue{}
		if x.db.Where("environment_variable_id = ?", val.Model.ID.String()).Order("created_at desc").First(&evValue).RecordNotFound() {
			log.InfoWithFields("envvar value not found", log.Fields{
				"id": val.Model.ID.String(),
			})
		} else {
			secrets = append(secrets, plugins.Secret{
				Key:   val.Key,
				Value: evValue.Value,
				Type:  val.Type,
			})
		}
	}

	// get all branches relevant for the projec
	branch := "master"
	envProjectBranch := models.EnvironmentBasedProjectBranch{}
	if x.db.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&envProjectBranch).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = envProjectBranch.GitBranch
	}

	pluginServices := []plugins.Service{}
	for _, service := range services {
		spec := models.ServiceSpec{}
		if x.db.Where("id = ?", service.ServiceSpecId).First(&spec).RecordNotFound() {
			log.InfoWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecId,
			})
			return
		}

		listeners := []models.ContainerPort{}
		if x.db.Where("service_id = ?", service.Model.ID).Find(&listeners).RecordNotFound() {
			log.InfoWithFields("container ports not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return
		}

		pluginListeners := []plugins.Listener{}
		for _, listener := range listeners {
			intPort, _ := strconv.Atoi(listener.Port)
			pluginListeners = append(pluginListeners, plugins.Listener{
				Port:     int32(intPort),
				Protocol: listener.Protocol,
			})
		}

		intTerminationGracePeriod, _ := strconv.Atoi(spec.TerminationGracePeriod)
		intReplicas, _ := strconv.Atoi(service.Count)
		pluginServices = append(pluginServices, plugins.Service{
			Id:        service.Model.ID.String(),
			Command:   service.Command,
			Name:      service.Name,
			Listeners: pluginListeners,
			State:     plugins.GetState("waiting"),
			Spec: plugins.ServiceSpec{
				Id:                            spec.Model.ID.String(),
				CpuRequest:                    fmt.Sprintf("%sm", spec.CpuRequest),
				CpuLimit:                      fmt.Sprintf("%sm", spec.CpuLimit),
				MemoryRequest:                 fmt.Sprintf("%sMi", spec.MemoryRequest),
				MemoryLimit:                   fmt.Sprintf("%sMi", spec.MemoryLimit),
				TerminationGracePeriodSeconds: int64(intTerminationGracePeriod),
			},
			Type:     string(service.Type),
			Replicas: int64(intReplicas),
		})
	}

	eventExtension := plugins.Extension{
		Id:           extension.Model.ID.String(),
		Action:       plugins.GetAction("destroy"),
		Slug:         extensionSpec.Key,
		State:        plugins.GetState("waiting"),
		StateMessage: "onDestroy",
		Config:       formValues,
		Artifacts:    map[string]string{},
		Environment:  environment.Name,
		Project: plugins.Project{
			Id: project.Model.ID.String(),
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        branch,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			Services:   pluginServices,
			Secrets:    secrets,
			Repository: project.Repository,
		},
	}

	x.events <- transistor.NewEvent(eventExtension, nil)
}

func (x *Actions) ExtensionInitCompleted(extension *models.Extension) {
}

func (x *Actions) ReleaseExtensionCompleted(re *models.ReleaseExtension) {
	x.db.Save(&re)

	project := models.Project{}
	release := models.Release{}
	fellowReleaseExtensions := []models.ReleaseExtension{}

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

	// loop through and check if all release extensions are completed
	done := true
	for _, fre := range fellowReleaseExtensions {
		if fre.Type == re.Type && fre.State != plugins.GetState("complete") {
			done = false
		}
	}


	if done {
		switch re.Type {
		case plugins.GetType("workflow"):
			x.WorkflowExtensionsCompleted(&release)
		case plugins.GetType("deployment"):
			x.DeploymentExtensionsCompleted(&release)
		}
	}
}

func (x *Actions) ReleaseExtensionsCompleted(release *models.Release) {
	project := models.Project{}

	release.StateMessage = "Finished"
	release.State = plugins.GetState("complete")

	x.db.Save(&release)

	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}
}

func (x *Actions) WorkflowExtensionsCompleted(release *models.Release) {
	// find all related deployment extensions
	depExtensions := []models.Extension{}
	aggregateReleaseExtensionArtifacts := make(map[string]interface{})
	found := false

	if x.db.Where("project_id = ? and environment_id = ?", release.ProjectId, release.EnvironmentId).Find(&depExtensions).RecordNotFound() {
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
		if plugins.Type(extensionSpec.Type) == plugins.GetType("workflow") {
			releaseExtension := models.ReleaseExtension{}

			if x.db.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, de.Model.ID, string(plugins.GetState("complete"))).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": de.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}

			// put all releaseextension artifacts inside release artifacts
			unmarshalledArtifacts := make(map[string]interface{})
			err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &unmarshalledArtifacts)
			if err != nil {
				log.InfoWithFields(err.Error(), log.Fields{})
				return
			}

			for k, v := range unmarshalledArtifacts {
				key := fmt.Sprintf("%s_%s", strings.ToUpper(extensionSpec.Key), strings.ToUpper(k))
				aggregateReleaseExtensionArtifacts[key] = v
			}
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("deployment") {
			found = true
		}
	}

	// persist workflow artifacts
	// release.Artifacts = plugins.MapStringStringToHstore(releaseExtensionArtifacts)
	x.db.Save(release)

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

	// get secrets from release snapshot env vars
	secrets, pluginServices, err := x.GetSecretsAndServicesFromSnapshot(release)
	if err != nil {
		log.Info(err.Error())
		return
	}

	headFeature := models.Feature{}
	if x.db.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := models.Feature{}
	if x.db.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := models.Environment{}
	if x.db.Where("id = ?", release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentId,
		})
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	envProjectBranch := models.EnvironmentBasedProjectBranch{}
	if x.db.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&envProjectBranch).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = envProjectBranch.GitBranch
	}

	releaseEvent := plugins.Release{
		Action:       plugins.GetAction("create"),
		State:        plugins.GetState("waiting"),
		Environment:  environment.Name,
		StateMessage: "create release event",
		Id:           release.Model.ID.String(),
		HeadFeature: plugins.Feature{
			Id:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			Id:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: "",
		Project: plugins.Project{
			Id:             project.Model.ID.String(),
			Action:         plugins.GetAction("update"),
			Repository:     project.Repository,
			NotifyChannels: []string{}, // not sure what channels can be notified with this
			Services:       pluginServices,
		},
		Git: plugins.Git{
			Url:    project.GitUrl,
			Branch: branch,
		},
		Secrets: secrets,
	}
	releaseExtensionEvents := []plugins.ReleaseExtension{}

	for _, extension := range depExtensions {
		extensionSpec := models.ExtensionSpec{}
		if x.db.Where("id= ?", extension.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"extension": extension,
			})
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("workflow") {
			releaseExtension := models.ReleaseExtension{}

			if x.db.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, extension.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": extension.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("deployment") {

			// create ReleaseExtension
			releaseExtension := models.ReleaseExtension{
				ReleaseId:         release.Model.ID,
				FeatureHash:       "",
				ServicesSignature: "",
				SecretsSignature:  "",
				ExtensionId:       extension.Model.ID,
				State:             plugins.GetState("waiting"),
				Type:              plugins.GetType("deployment"),
				StateMessage:      "initialized",
			}

			x.db.Save(&releaseExtension)
			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			formValues, err := utils.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.db)
			if err != nil {
				log.Info(err.Error())
			}

			extensionEvent := plugins.Extension{
				Id:     extension.Model.ID.String(),
				Config: formValues,
				// Artifacts: plugins.HstoreToMapStringString(extension.Artifacts),
			}

			releaseExtensionEvents = append(releaseExtensionEvents, plugins.ReleaseExtension{
				Id:           releaseExtension.Model.ID.String(),
				Action:       plugins.GetAction("create"),
				Slug:         extensionSpec.Key,
				State:        releaseExtension.State,
				Artifacts:    map[string]string{},
				Release:      releaseEvent,
				Extension:    extensionEvent,
				StateMessage: releaseExtension.StateMessage,
			})

		}
	}

	// send out release extension event for each re
	for _, re := range releaseExtensionEvents {
		re.Release.Artifacts = aggregateReleaseExtensionArtifacts
		x.events <- transistor.NewEvent(re, nil)
	}
}

func (x *Actions) DeploymentExtensionsCompleted(release *models.Release) {
	// find all related deployment extensions
	depExtensions := []models.Extension{}
	// releaseExtensionArtifacts := map[string]string{}

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
				"id": de.ExtensionSpecId,
			})
		}

		if plugins.Type(extensionSpec.Type) == plugins.GetType("deployment") {
			releaseExtension := models.ReleaseExtension{}

			if x.db.Where("release_id = ? AND extension_id = ? AND state = ?", release.Model.ID, de.Model.ID, plugins.GetState("complete")).Find(&releaseExtension).RecordNotFound() {
				log.InfoWithFields("release extension not found", log.Fields{
					"release_id":   release.Model.ID,
					"extension_id": de.Model.ID,
					"state":        plugins.GetState("complete"),
				})
			}
		}
	}

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
	release.State = plugins.GetState("complete")
	release.StateMessage = "Release completed"

	x.db.Save(release)
}

func (x *Actions) ReleaseCreated(release *models.Release) {
	project := models.Project{}

	if x.db.Where("id = ?", release.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"release": release,
		})
		return
	}

	// loop through extensions and send ReleaseWorkflow events
	projectExtensions := []models.Extension{}
	if x.db.Where("project_id = ? and environment_id = ?", release.ProjectId, release.EnvironmentId).Find(&projectExtensions).RecordNotFound() {
		log.InfoWithFields("project has no extensions", log.Fields{
			"project_id":     release.ProjectId,
			"environment_id": release.EnvironmentId,
		})
	}

	services := []models.Service{}
	if x.db.Where("project_id = ? and environment_id = ?", release.ProjectId, release.EnvironmentId).Find(&services).RecordNotFound() {
		log.InfoWithFields("project has no services", log.Fields{
			"project_id":     release.ProjectId,
			"environment_id": release.EnvironmentId,
		})
	}

	headFeature := models.Feature{}
	if x.db.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := models.Feature{}
	if x.db.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := models.Environment{}
	if x.db.Where("id = ?", release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentId,
		})
		return
	}

	// get all branches relevant for the projec
	branch := "master"
	envProjectBranch := models.EnvironmentBasedProjectBranch{}
	if x.db.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&envProjectBranch).RecordNotFound() {
		log.InfoWithFields("no env project branch found", log.Fields{})
	} else {
		branch = envProjectBranch.GitBranch
	}

	secrets, pluginServices, err := x.GetSecretsAndServicesFromSnapshot(release)
	if err != nil {
		log.Info(err.Error())
		return
	}

	releaseEvent := plugins.Release{
		Id:          release.Model.ID.String(),
		Action:      plugins.GetAction("create"),
		State:       plugins.GetState("waiting"),
		Environment: environment.Name,
		HeadFeature: plugins.Feature{
			Id:         headFeature.Model.ID.String(),
			Hash:       headFeature.Hash,
			ParentHash: headFeature.ParentHash,
			User:       headFeature.User,
			Message:    headFeature.Message,
			Created:    headFeature.Created,
		},
		TailFeature: plugins.Feature{
			Id:         tailFeature.Model.ID.String(),
			Hash:       tailFeature.Hash,
			ParentHash: tailFeature.ParentHash,
			User:       tailFeature.User,
			Message:    tailFeature.Message,
			Created:    tailFeature.Created,
		},
		User: release.User.Email,
		Project: plugins.Project{
			Id:         project.Model.ID.String(),
			Repository: project.Repository,
			Services:   pluginServices,
		},
		Git: plugins.Git{
			Url:    project.GitUrl,
			Branch: branch,
		},
		Secrets: secrets,
	}
	for _, extension := range projectExtensions {
		extensionSpec := models.ExtensionSpec{}
		if x.db.Where("id= ?", extension.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("extension spec not found", log.Fields{
				"id": extension.ExtensionSpecId,
			})
		}

		// ONLY SEND WORKFLOW TYPE, EVENTs
		if plugins.Type(extensionSpec.Type) == plugins.GetType("workflow") {
			// create ReleaseExtension
			releaseExtension := models.ReleaseExtension{
				ReleaseId:         release.Model.ID,
				FeatureHash:       "",
				ServicesSignature: "",
				SecretsSignature:  "",
				ExtensionId:       extension.Model.ID,
				State:             plugins.GetState("waiting"),
				Type:              plugins.GetType("workflow"),
			}

			x.db.Save(&releaseExtension)

			unmarshalledConfig := make(map[string]interface{})

			err := json.Unmarshal(extension.Config.RawMessage, &unmarshalledConfig)
			if err != nil {
				log.Info(err.Error())
			}

			formValues, err := utils.GetFilledFormValues(unmarshalledConfig, extensionSpec.Key, x.db)
			if err != nil {
				log.Info(err.Error())
			}

			extensionEvent := plugins.Extension{
				Id:     extension.Model.ID.String(),
				Config: formValues,
				// Artifacts:  plugins.HstoreToMapStringString(extension.Artifacts),
			}
			x.events <- transistor.NewEvent(plugins.ReleaseExtension{
				Id:        releaseExtension.Model.ID.String(),
				Action:    plugins.GetAction("create"),
				Slug:      extensionSpec.Key,
				State:     releaseExtension.State,
				Release:   releaseEvent,
				Extension: extensionEvent,
				Artifacts: map[string]string{},
			}, nil)
		}
	}
}
