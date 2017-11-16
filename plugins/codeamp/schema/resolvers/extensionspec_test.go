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

type TestExtensionSpecs struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
	user    models.User
	context context.Context
	envVar  models.EnvironmentVariable
}

func (suite *TestExtensionSpecs) SetupSuite() {

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

	suite.db = db
	suite.t = t
	suite.actions = actions

}

func (suite *TestExtensionSpecs) SetupDBAndContext() {
	suite.db.AutoMigrate(
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

	env := models.Environment{
		Name: "Production",
	}
	suite.db.Save(&env)
	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	suite.db.Save(&user)

	envVar := models.EnvironmentVariable{
		Key:           "key",
		Value:         "value",
		ProjectId:     uuid.UUID{},
		Version:       int32(0),
		Type:          plugins.Env,
		Scope:         plugins.ExtensionScope,
		UserId:        user.Model.ID,
		Created:       time.Now(),
		EnvironmentId: env.Model.ID,
	}
	suite.db.Save(&envVar)

	suite.context = context.WithValue(context.TODO(), "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
	suite.envVar = envVar
}

func (suite *TestExtensionSpecs) TearDownSuite() {
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

func (suite *TestExtensionSpecs) TestSuccessfulCreateExtensionSpec() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulCreateExtensionSpec")

	extensionSpecInput := struct {
		ExtensionSpec *resolvers.ExtensionSpecInput
	}{
		ExtensionSpec: &resolvers.ExtensionSpecInput{
			Name:      fmt.Sprintf("esname%s", stamp),
			Component: fmt.Sprintf("escomponent%s", stamp),
			FormSpec: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key",
					Value: "required|string",
				},
			},
			EnvironmentVariables: []map[string]interface{}{
				map[string]interface{}{
					"envVar": suite.envVar.Model.ID.String(),
				},
			},
			Type: "workflow",
			Key:  fmt.Sprintf("key%s", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	esResolver, _ := resolver.CreateExtensionSpec(&extensionSpecInput)

	assert.Equal(suite.T(), fmt.Sprintf("esname%s", stamp), esResolver.Name())
	suite.TearDownSuite()
}

func (suite *TestExtensionSpecs) TestFailedCreateExtensionSpecInvalidType() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestFailedCreateExtensionSpecInvalidType")

	extensionSpecInput := struct {
		ExtensionSpec *resolvers.ExtensionSpecInput
	}{
		ExtensionSpec: &resolvers.ExtensionSpecInput{
			Name:      fmt.Sprintf("esname%s", stamp),
			Component: fmt.Sprintf("escomponent%s", stamp),
			FormSpec: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key",
					Value: "required|string",
				},
			},
			EnvironmentVariables: []map[string]interface{}{
				map[string]interface{}{
					"envVar": suite.envVar.Model.ID.String(),
				},
			},
			Type: "invalidtype",
			Key:  fmt.Sprintf("key%s", stamp),
		},
	}
	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtensionSpec(&extensionSpecInput)

	assert.Equal(suite.T(), "Invalid extension type: invalidtype", err.Error())
	suite.TearDownSuite()
}

func (suite *TestExtensionSpecs) TestFailedCreateExtensionSpecInvalidEnvVarsFormat() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestFailedCreateExtensionSpecInvalidEnvVarsFormat")

	extensionSpecInput := struct {
		ExtensionSpec *resolvers.ExtensionSpecInput
	}{
		ExtensionSpec: &resolvers.ExtensionSpecInput{
			Name:      fmt.Sprintf("esname%s", stamp),
			Component: fmt.Sprintf("escomponent%s", stamp),
			FormSpec: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key",
					Value: "required|string",
				},
			},
			EnvironmentVariables: []map[string]interface{}{
				map[string]interface{}{
					"invalidKey": suite.envVar.Model.ID.String(),
				},
			},
			Type: "workflow",
			Key:  fmt.Sprintf("key%s", stamp),
		},
	}
	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtensionSpec(&extensionSpecInput)

	assert.Equal(suite.T(), "Invalid env. vars format", err.Error())
	suite.TearDownSuite()
}

func (suite *TestExtensionSpecs) TestFailedCreateExtensionSpecEnvVarsDontExist() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestFailedCreateExtensionSpecEnvVarsDontExist")

	extensionSpecInput := struct {
		ExtensionSpec *resolvers.ExtensionSpecInput
	}{
		ExtensionSpec: &resolvers.ExtensionSpecInput{
			Name:      fmt.Sprintf("esname%s", stamp),
			Component: fmt.Sprintf("escomponent%s", stamp),
			FormSpec: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key",
					Value: "required|string",
				},
			},
			EnvironmentVariables: []map[string]interface{}{
				map[string]interface{}{
					"envVar": "notrealenvvar",
				},
			},
			Type: "workflow",
			Key:  fmt.Sprintf("key%s", stamp),
		},
	}
	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateExtensionSpec(&extensionSpecInput)

	assert.Equal(suite.T(), "Specified env vars don't exist.", err.Error())
	suite.TearDownSuite()
}

func (suite *TestExtensionSpecs) TestSuccessfulUpdateExtensionSpec() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulUpdateExtensionSpec")

	valString := "val"
	es := models.ExtensionSpec{
		Type:      plugins.Workflow,
		Key:       fmt.Sprintf("key%s", stamp),
		Name:      fmt.Sprintf("name%s", stamp),
		Component: fmt.Sprintf("component%s", stamp),
		FormSpec: map[string]*string{
			"key": &valString,
		},
		Created: time.Now(),
	}

	suite.db.Save(&es)

	esString := es.Model.ID.String()

	extensionSpecInput := struct {
		ExtensionSpec *resolvers.ExtensionSpecInput
	}{
		ExtensionSpec: &resolvers.ExtensionSpecInput{
			ID:        &esString,
			Name:      fmt.Sprintf("name2%s", stamp),
			Component: fmt.Sprintf("component2%s", stamp),
			FormSpec: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key",
					Value: "required|string",
				},
			},
			EnvironmentVariables: []map[string]interface{}{
				map[string]interface{}{
					"envVar": suite.envVar.Model.ID.String(),
				},
			},
			Type: "",
			Key:  "",
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	esResolver, _ := resolver.UpdateExtensionSpec(&extensionSpecInput)

	envVarResolvers, _ := resolver.EnvironmentVariables(suite.context)

	assert.Equal(suite.T(), fmt.Sprintf("name2%s", stamp), esResolver.Name())
	assert.Equal(suite.T(), fmt.Sprintf("component2%s", stamp), esResolver.Component())
	assert.Equal(suite.T(), 1, len(envVarResolvers))
	assert.Equal(suite.T(), graphql.ID(suite.envVar.Model.ID.String()), envVarResolvers[0].ID())

	suite.TearDownSuite()
}

func (suite *TestExtensionSpecs) TestSuccessfulDeleteExtensionSpec() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulDeleteExtensionSpec")

	valString := "val"
	es := models.ExtensionSpec{
		Type:      plugins.Workflow,
		Key:       fmt.Sprintf("key%s", stamp),
		Name:      fmt.Sprintf("name%s", stamp),
		Component: fmt.Sprintf("component%s", stamp),
		FormSpec: map[string]*string{
			"key": &valString,
		},
		Created: time.Now(),
	}

	suite.db.Save(&es)

	esString := es.Model.ID.String()

	extensionSpecInput := struct {
		ExtensionSpec *resolvers.ExtensionSpecInput
	}{
		ExtensionSpec: &resolvers.ExtensionSpecInput{
			ID:        &esString,
			Name:      fmt.Sprintf("name2%s", stamp),
			Component: fmt.Sprintf("component2%s", stamp),
			FormSpec: []plugins.KeyValue{
				plugins.KeyValue{
					Key:   "key",
					Value: "required|string",
				},
			},
			EnvironmentVariables: []map[string]interface{}{
				map[string]interface{}{
					"envVar": suite.envVar.Model.ID.String(),
				},
			},
			Type: "",
			Key:  "",
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	resolver.DeleteExtensionSpec(&extensionSpecInput)

	emptyEs := models.ExtensionSpec{}

	suite.db.Where("id = ?", esString).Find(&emptyEs)

	assert.Equal(suite.T(), models.ExtensionSpec{}, emptyEs)
	suite.TearDownSuite()
}

func TestExtensionSpecResolvers(t *testing.T) {
	suite.Run(t, new(TestExtensionSpecs))
}
