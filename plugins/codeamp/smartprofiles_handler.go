package codeamp

import (
	"fmt"
	"github.com/codeamp/transistor"
	"github.com/codeamp/circuit/plugins"
	"github.com/davecgh/go-spew/spew"
	"github.com/codeamp/circuit/plugins/codeamp/model"
)

// SmartProfiles
func (x *CodeAmp) SmartProfiles(project *model.Project) error {
	spew.Dump("SmartProfiles")

	projects := []model.Project{}

	if err := x.DB.Find(&projects).Error; err != nil {
		return err
	}

	for _, project := range projects {
		project_envs := []model.ProjectEnvironment{}
		if err := x.DB.Where("project_id = ?", project.Model.ID).Find(&project_envs).Error; err != nil {
			return err
		}

		for _, project_env := range project_envs {
			env := model.Environment{}
			if err := x.DB.Where("id = ?", project_env.EnvironmentID).Find(&env).Error; err != nil {
				return err
			}

			payload := plugins.Project{
				Slug: project.Slug,
				Environment: env.Name,
			}
			
			fmt.Println(fmt.Sprintf("Sending event %s - %s", project.Slug, env.Name))
			ev := transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("create"), payload)
			ev.AddArtifact("INFLUX_HOST", "", false)
			ev.AddArtifact("INFLUX_DB", "", false)

			x.Events <- ev
		}
	}

	return nil
}

// SmartProfilesEventHandler
func (x *CodeAmp) ProjectEventHandler(e transistor.Event) error {
	spew.Dump(e)

	// check how many upgrades have been made to the particular service. if more than 2, then create a service modification request to admins

	// check if downgrade or upgrade. if upgrade, create a service modification request for admins to accept
	
	return nil
}