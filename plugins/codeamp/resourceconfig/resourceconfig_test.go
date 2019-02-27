package resourceconfig_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/codeamp/circuit/plugins"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/plugins/codeamp/resourceconfig"
	"github.com/codeamp/circuit/test"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var migrators = []interface{}{
	&model.Project{},
	&model.ProjectEnvironment{},
	&model.ProjectExtension{},
	&model.ProjectSettings{},
	&model.Environment{},
	&model.ProjectExtension{},
	&model.Service{},
	&model.Secret{},
	&model.Extension{},
}

type ResourceConfigTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (suite *ResourceConfigTestSuite) SetupTest() {
	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.db = db
}

func (suite *ResourceConfigTestSuite) TestExportProject() {
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	svcNames := []string{"foo", "bar"}
	for _, name := range svcNames {
		service := model.Service{
			Name:          name,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
		}
		suite.db.Create(&service)
	}

	extensionNames := []string{"e1", "e2"}
	for _, name := range extensionNames {
		extension := model.Extension{
			Name:   name,
			Key:    name,
			Config: postgres.Jsonb{[]byte(`[]`)},
		}

		suite.db.Create(&extension)

		projectExtension := model.ProjectExtension{
			ExtensionID:   extension.Model.ID,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
			Config:        postgres.Jsonb{[]byte(`[]`)},
			CustomConfig: postgres.Jsonb{[]byte(`
			{
				"type": "foobar", 
				"service": "foo", 
				"upstream_domains": [{"apex": "checkrhq-dev.net", "subdomain": "deploy-test"}]
			}			
			`)},
		}

		suite.db.Create(&projectExtension)
	}

	projectSettings := model.ProjectSettings{
		ProjectID:        project.Model.ID,
		EnvironmentID:    env.Model.ID,
		GitBranch:        "master",
		ContinuousDeploy: true,
	}

	suite.db.Create(&projectSettings)

	projectConfig := resourceconfig.CreateProjectConfig(suite.db, &project, &env)

	exportedProject, err := projectConfig.Export()
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.NotNil(suite.T(), exportedProject)
	assert.Equal(suite.T(), 2, len(exportedProject.Services))
	assert.Equal(suite.T(), 2, len(exportedProject.ProjectExtensions))
	assert.NotNil(suite.T(), exportedProject.ProjectSettings)
	assert.Equal(suite.T(), projectSettings.ContinuousDeploy, exportedProject.ProjectSettings.ContinuousDeploy)
	assert.Equal(suite.T(), projectSettings.GitBranch, exportedProject.ProjectSettings.GitBranch)
}

func (suite *ResourceConfigTestSuite) TestImportProject() {
	fmt.Println("TestImportProject")
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectSettings := model.ProjectSettings{
		EnvironmentID:    env.Model.ID,
		ProjectID:        project.Model.ID,
		GitBranch:        "master",
		ContinuousDeploy: true,
	}
	suite.db.Create(&projectSettings)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&projectEnv)

	importableProject := resourceconfig.Project{
		ProjectSettings: resourceconfig.ProjectSettings{
			GitBranch:        "test-branch",
			ContinuousDeploy: true,
		},
		Services: []resourceconfig.Service{
			resourceconfig.Service{
				Service: &model.Service{
					Command: "foocommand",
					Name:    "fooname",
					Type:    plugins.GetType("general"),
					Count:   int32(1),
					Ports: []model.ServicePort{
						model.ServicePort{
							Port:     int32(80),
							Protocol: "TCP",
						},
					},
					DeploymentStrategy: model.ServiceDeploymentStrategy{
						Type: plugins.GetType("default"),
					},
					ReadinessProbe: model.ServiceHealthProbe{
						Method: "default",
						Type:   plugins.GetType("readinessProbe"),
					},
					LivenessProbe: model.ServiceHealthProbe{
						Method: "default",
						Type:   plugins.GetType("readinessProbe"),
					},
					PreStopHook: "fooprestophook",
				},
			},
		},
		ProjectExtensions: []resourceconfig.ProjectExtension{
			resourceconfig.ProjectExtension{
				CustomConfig: `{}`,
				Config:       `[{ "key": "KEY", "value: "VALUE" }]`,
				Key:          "fookey",
			},
		},
		Secrets: []resourceconfig.Secret{
			resourceconfig.Secret{
				Key:      "FOOSECRET",
				Value:    "foovalue",
				IsSecret: false,
			},
			resourceconfig.Secret{
				Key:      "FOOSECRET2",
				Value:    "foovalue",
				IsSecret: false,
			},
		},
	}

	projectConfig := resourceconfig.CreateProjectConfig(suite.db, &project, &env)
	err := projectConfig.Import(&importableProject)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// check ProjectSettings update
	dbProjectSettings := model.ProjectSettings{}
	suite.db.Where("project_id = ? and environment_id = ?", project.Model.ID, env.Model.ID).First(&dbProjectSettings)
	assert.Equal(suite.T(), "test-branch", dbProjectSettings.GitBranch)
	assert.Equal(suite.T(), true, dbProjectSettings.ContinuousDeploy)

	// check secrets

	// check project extension

	// check services
}

func (suite *ResourceConfigTestSuite) TearDownTest() {
	suite.db.Delete(&migrators)
	suite.db.Close()
}

func TestResourceConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceConfigTestSuite))
}
