package resourceconfig_test

import (
	"log"
	"testing"

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
			CustomConfig:  postgres.Jsonb{[]byte(`{}`)},
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

	projectConfig := resourceconfig.CreateProjectConfig(nil, suite.db, &project, &env)

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

func (suite *ResourceConfigTestSuite) TearDownTest() {
	suite.db.Delete(&migrators)
	suite.db.Close()
}

func TestResourceConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceConfigTestSuite))
}
