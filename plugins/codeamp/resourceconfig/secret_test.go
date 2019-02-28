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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var migrators = []interface{}{
	&model.Secret{},
	&model.SecretValue{},
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

func (suite *ResourceConfigTestSuite) TearDownTest() {
	suite.db.Delete(&migrators)
	suite.db.Close()
}

func TestResourceConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceConfigTestSuite))
}
