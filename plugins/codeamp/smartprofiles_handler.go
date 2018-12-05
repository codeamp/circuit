package codeamp

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"

	"github.com/codeamp/circuit/plugins/codeamp/model"
)

// SmartProfiles
func (x *CodeAmp) SmartProfiles(project *model.Project) error {
	project_envs := []model.ProjectEnvironment{}
	if err := x.DB.Where("project_id = ?", project.Model.ID).Find(&project_envs).Error; err != nil {
		return err
	}

	for _, project_env := range project_envs {
		env := model.Environment{}
		if err := x.DB.Where("id = ?", project_env.EnvironmentID).Find(&env).Error; err != nil {
			return err
		}

		services := []model.Service{}
		if err := x.DB.Where("project_id = ? and environment_id = ?", project.Model.ID, project_env.EnvironmentID).Find(&services).Error; err != nil {
			return err
		}

		pluginServices := []plugins.Service{}
		for _, svc := range services {
			pluginServices = append(pluginServices, plugins.Service{
				ID:   svc.Model.ID.String(),
				Name: svc.Name,
			})
		}

		payload := plugins.Project{
			Slug:        project.Slug,
			Environment: env.Name,
			Services:    pluginServices,
		}

		fmt.Println(fmt.Sprintf("Sending event %s - %s", project.Slug, env.Name))
		ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), payload)
		ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
		ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

		x.Events <- ev
	}

	return nil
}

// ProjectEventHandler
func (x *CodeAmp) ProjectEventHandler(e transistor.Event) error {
	tx := x.DB.Begin()

	// For each service's service spec, find + update or create the corresponding suggested service spec
	projectPayload := e.Payload.(plugins.Project)
	for _, service := range projectPayload.Services {
		dbService := model.Service{}
		if err := tx.Where("id = ?", service.ID).First(&dbService).Error; err != nil {
			tx.Rollback()
			return err
		} else {
			// get the non-suggested service spec for the service in order to copy over any
			// non-resource related attributes e.g. TerminationGracePeriod
			nonSuggestedServiceSpec := model.ServiceSpec{}
			if err := tx.Where("service_id = ? and type != ?", dbService.Model.ID.String(), "suggested").First(&nonSuggestedServiceSpec).Error; err != nil {
				tx.Rollback()
				return err
			}

			spew.Dump(service.Spec)

			// check if previous suggested service spec already exists with exact same spec
			previousSuggestedServiceSpec := model.ServiceSpec{}

			suggestedServiceSpec := model.ServiceSpec{
				ServiceID:              dbService.Model.ID,
				CpuRequest:             service.Spec.CpuRequest,
				CpuLimit:               service.Spec.CpuLimit,
				MemoryRequest:          service.Spec.MemoryRequest,
				MemoryLimit:            service.Spec.MemoryLimit,
				Type:                   "suggested",
				TerminationGracePeriod: nonSuggestedServiceSpec.TerminationGracePeriod,
				Name:                   nonSuggestedServiceSpec.Name,
			}

			if err := tx.Where(suggestedServiceSpec).Order("created_at desc").First(&previousSuggestedServiceSpec).Error; err == nil {
				tx.Rollback()
				return nil
			} else {
				if err := tx.Create(&suggestedServiceSpec).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil
	}	

	return nil
}
