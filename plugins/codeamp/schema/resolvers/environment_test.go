package resolvers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestEnvironment struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
	user    models.User
	context context.Context
}

func (suite *TestEnvironment) SetupSuite() {

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

func (suite *TestEnvironment) SetupDBAndContext() {
	suite.db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.Environment{},
	)

	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	suite.db.Save(&user)

	suite.context = context.WithValue(suite.context, "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
}

func (suite *TestEnvironment) TearDownSuite() {
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from user_permissions;")
	suite.db.Exec("delete from environments;")
}

func (suite *TestEnvironment) TestSuccessfulCreateEnvironment() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulCreateEnvironment")

	var envResolver *resolvers.EnvironmentResolver
	envInput := struct {
		Environment *resolvers.EnvironmentInput
	}{
		Environment: &resolvers.EnvironmentInput{
			Name: fmt.Sprintf("testfoo%s", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	envResolver, _ = resolver.CreateEnvironment(suite.context, &envInput)
	assert.Equal(suite.T(), fmt.Sprintf("testfoo%s", stamp), envResolver.Name(suite.context))

	envInput2 := struct {
		Environment *resolvers.EnvironmentInput
	}{
		Environment: &resolvers.EnvironmentInput{
			Name: fmt.Sprintf("testfoo2%s", stamp),
		},
	}
	envResolver, _ = resolver.CreateEnvironment(suite.context, &envInput2)
	assert.Equal(suite.T(), fmt.Sprintf("testfoo2%s", stamp), envResolver.Name(suite.context))

	envInput3 := struct {
		Environment *resolvers.EnvironmentInput
	}{
		Environment: &resolvers.EnvironmentInput{
			Name: fmt.Sprintf("testfoo3%s", stamp),
		},
	}
	envResolver, _ = resolver.CreateEnvironment(suite.context, &envInput3)
	assert.Equal(suite.T(), fmt.Sprintf("testfoo3%s", stamp), envResolver.Name(suite.context))

	// because created_at desc in Environments() method
	envInputs := []struct{ Environment *resolvers.EnvironmentInput }{
		envInput3, envInput2, envInput,
	}

	// get from Environments
	envResolvers, _ := resolver.Environments(suite.context)
	assert.Equal(suite.T(), 3, len(envResolvers))
	for idx, envResolver := range envResolvers {
		assert.Equal(suite.T(), envInputs[idx].Environment.Name, envResolver.Name(suite.context))
	}

	suite.TearDownSuite()
}

func (suite *TestEnvironment) TestSuccessfulUpdateEnvironment() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulUpdateEnvironment")

	env := models.Environment{
		Name: fmt.Sprintf("testfoo%s", stamp),
	}
	suite.db.Save(&env)
	envId := env.Model.ID.String()

	var envResolver *resolvers.EnvironmentResolver
	envInput := struct {
		Environment *resolvers.EnvironmentInput
	}{
		Environment: &resolvers.EnvironmentInput{
			ID:   &envId,
			Name: fmt.Sprintf("testfoo2%s", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	envResolver, _ = resolver.UpdateEnvironment(suite.context, &envInput)

	assert.Equal(suite.T(), fmt.Sprintf("testfoo2%s", stamp), envResolver.Name(suite.context))

	envResolvers, _ := resolver.Environments(suite.context)

	assert.Equal(suite.T(), 1, len(envResolvers))
	assert.Equal(suite.T(), fmt.Sprintf("testfoo2%s", stamp), envResolvers[0].Name(suite.context))

	suite.TearDownSuite()
}

func (suite *TestEnvironment) TestSuccessfulDeleteEnvironment() {
	suite.SetupDBAndContext()
	stamp := strings.ToLower("TestSuccessfulDeleteEnvironment")

	env := models.Environment{
		Name: fmt.Sprintf("testfoo%s", stamp),
	}
	suite.db.Save(&env)

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	envResolvers, _ := resolver.Environments(suite.context)
	assert.Equal(suite.T(), 1, len(envResolvers))
	assert.Equal(suite.T(), env.Name, envResolvers[0].Name(suite.context))

	envId := env.Model.ID.String()
	var envResolver *resolvers.EnvironmentResolver

	envInput := struct {
		Environment *resolvers.EnvironmentInput
	}{
		Environment: &resolvers.EnvironmentInput{
			ID:   &envId,
			Name: fmt.Sprintf("testfoo2%s", stamp),
		},
	}

	envResolver, _ = resolver.DeleteEnvironment(suite.context, &envInput)
	assert.Equal(suite.T(), envId, string(envResolver.ID()))

	envResolvers, _ = resolver.Environments(suite.context)
	assert.Equal(suite.T(), 0, len(envResolvers))

	suite.TearDownSuite()
}

func TestEnvironmentResolvers(t *testing.T) {
	suite.Run(t, new(TestEnvironment))
}
