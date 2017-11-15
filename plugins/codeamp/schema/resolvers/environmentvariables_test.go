package resolvers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestEnvironmentVariables struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
	env     models.Environment
	user    models.User
	context context.Context
}

func (suite *TestEnvironmentVariables) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"15432",
		"postgres",
		"codeamp_test",
		"disable",
		"",
	))

	db.Exec(fmt.Sprintf("CREATE DATABASE %s", "codeamp_test"))
	db.Exec("CREATE EXTENSION \"uuid-ossp\"")
	db.Exec("CREATE EXTENSION IF NOT EXISTS hstore")

	db.AutoMigrate(
		&models.Project{},
		&models.Feature{},
		&models.Release{},
		&models.Service{},
		&models.EnvironmentVariable{},
		&models.ExtensionSpecEnvironmentVariable{},
		&models.Extension{},
		&models.ExtensionSpec{},
		&models.Environment{},
		&models.User{},
	)

	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return codeamp.NewCodeAmp() })
	t, _ := transistor.NewTestTransistor(transistor.Config{
		Server:    "http://127.0.0.1:16379",
		Password:  "",
		Database:  "0",
		Namespace: "",
		Pool:      "30",
		Process:   "1",
		Plugins: map[string]interface{}{
			"codeamp": map[string]interface{}{
				"workers": 1,
				"postgres": map[string]interface{}{
					"host":     "0.0.0.0",
					"port":     "15432",
					"user":     "postgres",
					"dbname":   "codeamp_test",
					"sslmode":  "disable",
					"password": "",
				},
			},
		},
		EnabledPlugins: []string{},
		Queueing:       false,
	})

	actions := actions.NewActions(t.TestEvents, db)

	env := models.Environment{
		Name: "Production",
	}
	db.Save(&env)
	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	db.Save(&user)

	suite.db = db
	suite.t = t
	suite.actions = actions
	suite.context = context.WithValue(context.TODO(), "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
	suite.env = env
}

func (suite *TestEnvironmentVariables) TearDownSuite() {
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from features;")
	suite.db.Exec("delete from releases;")
	suite.db.Exec("delete from services;")
	suite.db.Exec("delete from environment_variables;")
	suite.db.Exec("delete from extension_spec_environment_variables;")
	suite.db.Exec("delete from extensions;")
	suite.db.Exec("delete from extension_specs;")
	suite.db.Exec("delete from environments;")
	suite.db.Exec("delete from users;")
}

func (suite *TestEnvironmentVariables) TestSuccessfulCreateEnvironmentVariable() {
	stamp := strings.ToLower("TestSuccessfulCreateEnvironmentVariable")

	envVarInput := struct {
		EnvironmentVariable *resolvers.EnvironmentVariableInput
	}{
		EnvironmentVariable: &resolvers.EnvironmentVariableInput{
			Key:           fmt.Sprintf("key%s", stamp),
			Value:         fmt.Sprintf("value%s", stamp),
			Type:          plugins.Env,
			Scope:         plugins.ExtensionScope,
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	envVarResolver, _ := resolver.CreateEnvironmentVariable(suite.context, &envVarInput)

	envResolver, _ := envVarResolver.Environment(context.TODO())

	assert.Equal(suite.T(), fmt.Sprintf("key%s", stamp), envVarResolver.Key())
	assert.Equal(suite.T(), fmt.Sprintf("value%s", stamp), envVarResolver.Value())
	assert.Equal(suite.T(), string(plugins.Env), envVarResolver.Type())
	assert.Equal(suite.T(), string(plugins.ExtensionScope), envVarResolver.Scope())
	assert.Equal(suite.T(), "Production", envResolver.Name(context.TODO()))
}

func (suite *TestEnvironmentVariables) TestFailedCreateEnvironmentVariableAlreadyExistsInSameEnvironment() {
	stamp := strings.ToLower("TestFailedCreateEnvironmentVariableAlreadyExists")

	envVar := models.EnvironmentVariable{
		Key:           fmt.Sprintf("key%s", stamp),
		Value:         fmt.Sprintf("value%s", stamp),
		ProjectId:     uuid.UUID{},
		Version:       int32(0),
		Type:          plugins.Env,
		Scope:         plugins.ExtensionScope,
		UserId:        suite.user.Model.ID,
		Created:       time.Now(),
		EnvironmentId: suite.env.Model.ID,
	}
	suite.db.Save(&envVar)

	envVarInput := struct {
		EnvironmentVariable *resolvers.EnvironmentVariableInput
	}{
		EnvironmentVariable: &resolvers.EnvironmentVariableInput{
			Key:           fmt.Sprintf("key%s", stamp),
			Value:         fmt.Sprintf("value%s", stamp),
			Type:          plugins.Env,
			Scope:         plugins.ExtensionScope,
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateEnvironmentVariable(suite.context, &envVarInput)

	assert.Equal(suite.T(), "CreateEnvironmentVariable: key already exists", err.Error())
}

func (suite *TestEnvironmentVariables) TestSuccessfulUpdateEnvironmentVariable() {
	stamp := strings.ToLower("TestSuccessfulUpdateEnvironmentVariable")

	e2 := models.Environment{
		Name: fmt.Sprintf("Production2%s", stamp),
	}
	suite.db.Save(&e2)

	envVar := models.EnvironmentVariable{
		Key:           fmt.Sprintf("key%s", stamp),
		Value:         fmt.Sprintf("value%s", stamp),
		ProjectId:     uuid.UUID{},
		Version:       int32(0),
		Type:          plugins.Env,
		Scope:         plugins.ExtensionScope,
		UserId:        suite.user.Model.ID,
		Created:       time.Now(),
		EnvironmentId: suite.env.Model.ID,
	}
	suite.db.Save(&envVar)
	envVarId := envVar.Model.ID.String()

	envVarInput := struct {
		EnvironmentVariable *resolvers.EnvironmentVariableInput
	}{
		EnvironmentVariable: &resolvers.EnvironmentVariableInput{
			ID:            &envVarId,
			Key:           fmt.Sprintf("key%s", stamp),
			Value:         fmt.Sprintf("value2%s", stamp),
			Type:          plugins.Env,
			Scope:         plugins.GlobalScope,
			EnvironmentId: e2.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	envVarResolver, _ := resolver.UpdateEnvironmentVariable(suite.context, &envVarInput)

	envResolver, _ := envVarResolver.Environment(suite.context)

	assert.Equal(suite.T(), fmt.Sprintf("value2%s", stamp), envVarResolver.Value())
	assert.Equal(suite.T(), int32(1), envVarResolver.Version())
	assert.Equal(suite.T(), fmt.Sprintf("Production2%s", stamp), envResolver.Name(suite.context))
	assert.Equal(suite.T(), plugins.GlobalScope, envVarResolver.Scope())
}

func (suite *TestEnvironmentVariables) TestFailedUpdateEnvironmentVariableDoesntExist() {
	stamp := strings.ToLower("TestFailedUpdateEnvironmentVariableDoesntExist")

	fakeEnvVarId := uuid.NewV1().String()

	envVarInput := struct {
		EnvironmentVariable *resolvers.EnvironmentVariableInput
	}{
		EnvironmentVariable: &resolvers.EnvironmentVariableInput{
			ID:            &fakeEnvVarId,
			Key:           fmt.Sprintf("key%s", stamp),
			Value:         fmt.Sprintf("value2%s", stamp),
			Type:          plugins.Env,
			Scope:         plugins.GlobalScope,
			EnvironmentId: suite.env.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.UpdateEnvironmentVariable(suite.context, &envVarInput)

	assert.Equal(suite.T(), "UpdateEnvironmentVariable: env var doesn't exist.", err.Error())
}

func (suite *TestEnvironmentVariables) TestSuccessfulDeleteEnvironmentVariable() {
	stamp := strings.ToLower("TestSuccessfulDeleteEnvironmentVariable")

	envVar := models.EnvironmentVariable{
		Key:           fmt.Sprintf("key%s", stamp),
		Value:         fmt.Sprintf("value%s", stamp),
		ProjectId:     uuid.UUID{},
		Version:       int32(0),
		Type:          plugins.Env,
		Scope:         plugins.ExtensionScope,
		UserId:        suite.user.Model.ID,
		Created:       time.Now(),
		EnvironmentId: suite.env.Model.ID,
	}
	suite.db.Save(&envVar)
	envVarId := envVar.Model.ID.String()

	envVarInput := struct {
		EnvironmentVariable *resolvers.EnvironmentVariableInput
	}{
		EnvironmentVariable: &resolvers.EnvironmentVariableInput{
			ID:            &envVarId,
			Key:           fmt.Sprintf("key%s", stamp),
			Value:         fmt.Sprintf("value%s", stamp),
			Type:          plugins.Env,
			Scope:         plugins.ExtensionScope,
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	envId := struct {
		ID graphql.ID
	}{
		ID: graphql.ID(envVar.Model.ID.String()),
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	resolver.DeleteEnvironmentVariable(suite.context, &envVarInput)
	_, err := resolver.EnvironmentVariable(suite.context, &envId)

	assert.Equal(suite.T(), "record not found", err.Error())
}

func TestEnvironmentVariableResolvers(t *testing.T) {
	suite.Run(t, new(TestEnvironmentVariables))
}
