package resourceconfig_test

import (
	"context"
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

	user := model.User{
		Email: "foo@gmail.com",
	}
	suite.db.Create(&user)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      user.Model.ID.String(),
		Email:       user.Email,
		Permissions: []string{"admin"},
	})

	projectConfig := resourceconfig.CreateProjectConfig(authContext, suite.db, &project, &env)

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

func (suite *ResourceConfigTestSuite) TestImportProjectSuccess_FullFlow() {
	project := model.Project{
		Slug: "hello-there",
	}
	suite.db.Create(&project)

	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}
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

	extension := model.Extension{
		Type:          plugins.GetType("once"),
		Key:           "fookey",
		Name:          "name",
		Component:     "",
		Cacheable:     false,
		EnvironmentID: env.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
	}
	suite.db.Create(&extension)

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
				Config:       `[{ "key": "KEY", "value": "VALUE" }]`,
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

	user := model.User{
		Email: "foo@gmail.com",
	}
	suite.db.Create(&user)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      user.Model.ID.String(),
		Email:       user.Email,
		Permissions: []string{"admin"},
	})

	projectConfig := resourceconfig.CreateProjectConfig(authContext, suite.db, &project, &env)
	err := projectConfig.Import(&importableProject)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// check ProjectSettings update
	dbProjectSettings := model.ProjectSettings{}
	suite.db.Where("project_id = ? and environment_id = ?", project.Model.ID, env.Model.ID).First(&dbProjectSettings)
	assert.Equal(suite.T(), "test-branch", dbProjectSettings.GitBranch)
	assert.Equal(suite.T(), true, dbProjectSettings.ContinuousDeploy)

	// assert all objects have been created - service, secret, projectextension
	for _, service := range importableProject.Services {
		if err = suite.db.Where("environment_id = ? and project_id = ? and name = ?", env.Model.ID, project.Model.ID, service.Name).Find(&model.Service{}).Error; err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}

	for _, secret := range importableProject.Secrets {
		if err = suite.db.Where("environment_id = ? and project_id = ? and key = ?", env.Model.ID, project.Model.ID, secret.Key).Find(&model.Secret{}).Error; err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}

	for _, pExtension := range importableProject.ProjectExtensions {
		extension := model.Extension{}
		suite.db.Where("key = ? and environment_id = ?", pExtension.Key, env.Model.ID).Find(&extension)
		if err = suite.db.Where("environment_id = ? and project_id = ? and extension_id = ?", env.Model.ID, project.Model.ID, extension.Model.ID).Find(&model.ProjectExtension{}).Error; err != nil {
			assert.FailNow(suite.T(), err.Error())
		}
	}
}

// ProjectSettings related
func (suite *ResourceConfigTestSuite) TestImportProjectSettings_Failure_NilDependency() {
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

	projectSettingsConfig := resourceconfig.CreateProjectSettingsConfig(suite.db, &project, nil)
	err := projectSettingsConfig.Import(&resourceconfig.ProjectSettings{
		GitBranch:        "master",
		ContinuousDeploy: false,
	})

	assert.NotNil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TestImportProjectSettings_Failure_NoExistingProjectSettings() {
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&projectEnv)

	projectSettingsConfig := resourceconfig.CreateProjectSettingsConfig(suite.db, &project, &env)
	err := projectSettingsConfig.Import(&resourceconfig.ProjectSettings{
		GitBranch:        "master",
		ContinuousDeploy: false,
	})

	assert.NotNil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TestExportProjectSettings_NilDependency() {
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

	projectSettingsConfig := resourceconfig.CreateProjectSettingsConfig(suite.db, &project, nil)
	exportedProjectConfig, err := projectSettingsConfig.Export()

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), exportedProjectConfig)
}

func (suite *ResourceConfigTestSuite) TestExportProjectSettings_Failure_NoExistingProjectSettings() {
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&projectEnv)

	projectSettingsConfig := resourceconfig.CreateProjectSettingsConfig(suite.db, &project, &env)
	exportedProjectConfig, err := projectSettingsConfig.Export()

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), exportedProjectConfig)
}

// ProjectExtension related tests
func (suite *ResourceConfigTestSuite) TestImportProjectExtension_Success() {
	// create relevant base objects - project, projectenvironment, projectsettings, extension
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	extension := model.Extension{
		Type:          plugins.GetType("once"),
		Key:           "key",
		Name:          "name",
		Component:     "",
		Cacheable:     false,
		EnvironmentID: env.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
	}

	suite.db.Create(&projectEnv)
	suite.db.Create(&extension)

	// attempt import project extension process now
	projectExtensionConfig := resourceconfig.CreateProjectExtensionConfig(suite.db, nil, &project, &env)
	err := projectExtensionConfig.Import(&resourceconfig.ProjectExtension{
		CustomConfig: `{}`,
		Config:       `[]`,
		Key:          extension.Key,
	})
	assert.Nil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TestImportProjectExtension_Failure_NilDependency() {
	// create relevant base objects - project, projectenvironment, projectsettings, extension
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	extension := model.Extension{
		Type:          plugins.GetType("once"),
		Key:           "key",
		Name:          "name",
		Component:     "",
		Cacheable:     false,
		EnvironmentID: env.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
	}

	suite.db.Create(&projectEnv)
	suite.db.Create(&extension)

	// attempt import project extension process now
	projectExtensionConfig := resourceconfig.CreateProjectExtensionConfig(suite.db, nil, nil, &env)
	err := projectExtensionConfig.Import(&resourceconfig.ProjectExtension{
		CustomConfig: `{}`,
		Config:       `[]`,
		Key:          extension.Key,
	})
	assert.NotNil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TestImportProjectExtension_Failure_NoExtensionExistsWithDeclaredKey() {
	// create relevant base objects - project, projectenvironment, projectsettings, extension
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}

	suite.db.Create(&projectEnv)

	// attempt import project extension process now
	projectExtensionConfig := resourceconfig.CreateProjectExtensionConfig(suite.db, nil, &project, &env)
	err := projectExtensionConfig.Import(&resourceconfig.ProjectExtension{
		CustomConfig: `{}`,
		Config:       `[]`,
		Key:          "invalid key",
	})
	assert.NotNil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TestExportProjectExtension_Success() {
	// create relevant base objects - project, projectenvironment, projectsettings, extension
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	extension := model.Extension{
		Type:          plugins.GetType("once"),
		Key:           "key",
		Name:          "name",
		Component:     "",
		Cacheable:     false,
		EnvironmentID: env.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
	}

	suite.db.Create(&projectEnv)
	suite.db.Create(&extension)

	projectExtension := model.ProjectExtension{
		EnvironmentID: env.Model.ID,
		ExtensionID:   extension.Model.ID,
		ProjectID:     project.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
		CustomConfig:  postgres.Jsonb{[]byte(`{}`)},
	}
	suite.db.Create(&projectExtension)

	projectExtensionConfig := resourceconfig.CreateProjectExtensionConfig(suite.db, &projectExtension, &project, &env)
	exportedProjectExtension, err := projectExtensionConfig.Export()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), exportedProjectExtension)
	assert.Equal(suite.T(), extension.Key, exportedProjectExtension.Key)
	assert.Equal(suite.T(), "[]\n", exportedProjectExtension.Config)
	assert.Equal(suite.T(), "{}", exportedProjectExtension.CustomConfig)
}

func (suite *ResourceConfigTestSuite) TestExportProjectExtension_Failure_NilDependency() {
	// create relevant base objects - project, projectenvironment, projectsettings, extension
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	extension := model.Extension{
		Type:          plugins.GetType("once"),
		Key:           "key",
		Name:          "name",
		Component:     "",
		Cacheable:     false,
		EnvironmentID: env.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
	}

	suite.db.Create(&projectEnv)
	suite.db.Create(&extension)

	projectExtension := model.ProjectExtension{
		EnvironmentID: env.Model.ID,
		ExtensionID:   extension.Model.ID,
		ProjectID:     project.Model.ID,
		Config:        postgres.Jsonb{[]byte(`[]`)},
		CustomConfig:  postgres.Jsonb{[]byte(`{}`)},
	}
	suite.db.Create(&projectExtension)

	projectExtensionConfig := resourceconfig.CreateProjectExtensionConfig(suite.db, nil, &project, &env)
	exportedProjectExtension, err := projectExtensionConfig.Export()

	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), exportedProjectExtension)
}

// Secret related tests
func (suite *ResourceConfigTestSuite) TestImportSecret_Success() {
	// base objects - project, environment
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&projectEnv)

	user := model.User{
		Email: "foo@gmail.com",
	}
	suite.db.Create(&user)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      user.Model.ID.String(),
		Email:       user.Email,
		Permissions: []string{"admin"},
	})

	secretConfig := resourceconfig.CreateProjectSecretConfig(authContext, suite.db, nil, &project, &env)
	err := secretConfig.Import(&resourceconfig.Secret{
		Key:      "KEY",
		Value:    "value",
		Type:     "file",
		IsSecret: false,
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// confirm secret creation
	createdSecret := model.Secret{}
	suite.db.Where("key = ? and project_id = ? and environment_id = ?", "KEY", project.Model.ID, env.Model.ID).Find(&createdSecret)
	assert.Equal(suite.T(), plugins.GetType("file"), createdSecret.Type)
	assert.Equal(suite.T(), false, createdSecret.IsSecret)

	createdSecretValue := model.SecretValue{}
	suite.db.Where("secret_id = ?", createdSecret.Model.ID).Find(&createdSecretValue)
	assert.Equal(suite.T(), "value", createdSecretValue.Value)
}

func (suite *ResourceConfigTestSuite) TestImportSecret_Failure_NilDependency() {
	// base objects - project, environment
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&projectEnv)

	user := model.User{
		Email: "foo@gmail.com",
	}
	suite.db.Create(&user)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      user.Model.ID.String(),
		Email:       user.Email,
		Permissions: []string{"admin"},
	})

	secretConfig := resourceconfig.CreateProjectSecretConfig(authContext, suite.db, nil, &project, nil)
	err := secretConfig.Import(&resourceconfig.Secret{
		Key:      "KEY",
		Value:    "value",
		Type:     "file",
		IsSecret: false,
	})

	assert.NotNil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TestImportSecret_Failure_SecretWithSameKeyAlreadyExists() {
	// base objects - project, environment
	project := model.Project{
		Slug: "hello-there",
	}
	env := model.Environment{
		Key:  "dev",
		Name: "Dev",
	}

	suite.db.Create(&project)
	suite.db.Create(&env)

	projectEnv := model.ProjectEnvironment{
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&projectEnv)

	// create secret with key KEY
	secret := model.Secret{
		Key:           "KEY",
		EnvironmentID: env.Model.ID,
		ProjectID:     project.Model.ID,
	}
	suite.db.Create(&secret)

	user := model.User{
		Email: "foo@gmail.com",
	}
	suite.db.Create(&user)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      user.Model.ID.String(),
		Email:       user.Email,
		Permissions: []string{"admin"},
	})

	secretConfig := resourceconfig.CreateProjectSecretConfig(authContext, suite.db, nil, &project, nil)
	err := secretConfig.Import(&resourceconfig.Secret{
		Key:      "KEY",
		Value:    "value",
		Type:     "file",
		IsSecret: false,
	})

	assert.NotNil(suite.T(), err)
}

// Project related tests
func (suite *ResourceConfigTestSuite) TestImportProject_Failure_NilEnvironment() {
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

	user := model.User{
		Email: "foo@gmail.com",
	}
	suite.db.Create(&user)

	authContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      user.Model.ID.String(),
		Email:       user.Email,
		Permissions: []string{"admin"},
	})

	projectConfig := resourceconfig.CreateProjectConfig(authContext, suite.db, &project, nil)
	err := projectConfig.Import(&resourceconfig.Project{
		ProjectSettings: resourceconfig.ProjectSettings{
			GitBranch:        "master",
			ContinuousDeploy: true,
		},
	})

	assert.NotNil(suite.T(), err)
}

func (suite *ResourceConfigTestSuite) TearDownTest() {
	suite.db.Delete(&migrators)
	suite.db.Close()
}

func TestResourceConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceConfigTestSuite))
}
