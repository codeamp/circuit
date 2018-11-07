package codeamp

import (
	"github.com/spf13/viper"
	"fmt"
	"github.com/codeamp/transistor"
	"github.com/codeamp/circuit/plugins"
	"github.com/davecgh/go-spew/spew"
	"github.com/codeamp/circuit/plugins/codeamp/model"
)

// SmartProfiles
func (x *CodeAmp) SmartProfiles(project *model.Project) error {
	spew.Dump("SmartProfiles")

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
				Name: svc.Name,
			})
		}

		payload := plugins.Project{
			Slug: project.Slug,
			Environment: env.Name,
			Services: pluginServices,
		}
		
		fmt.Println(fmt.Sprintf("Sending event %s - %s", project.Slug, env.Name))
		ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("update"), payload)
		ev.AddArtifact("INFLUX_HOST", viper.GetString("plugins.smartprofiles.influxdb.host"), false)
		ev.AddArtifact("INFLUX_DB", viper.GetString("plugins.smartprofiles.influxdb.db"), false)

		x.Events <- ev
	}

	return nil
}

// SmartProfilesEventHandler
func (x *CodeAmp) ProjectEventHandler(e transistor.Event) error {
	// check how many upgrades have been made to the particular service. if more than 2, then create a service modification request to admins

	// check if downgrade or upgrade. if upgrade, create a service modification request for admins to accept
	
	return nil
}