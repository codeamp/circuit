package resolvers_test

import (
	"context"
	"fmt"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestRelease struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
	user    models.User
	project models.Project
	env     models.Environment
	feature models.Feature
	context context.Context
}

func (suite *TestRelease) SetupSuite() {

	db, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		"0.0.0.0",
		"4500",
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

func (suite *TestRelease) SetupDBAndContext() {
	stamp := "SetupDBAndContext"

	suite.db.AutoMigrate(
		&models.User{},
		&models.UserPermission{},
		&models.Release{},
		&models.Feature{},
		&models.Environment{},
		&models.Project{},
	)

	user := models.User{
		Email:       "foo@boo.com",
		Password:    "secret",
		Permissions: []models.UserPermission{},
	}
	suite.db.Save(&user)

	p := models.Project{
		Name:          fmt.Sprintf("test%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p)

	env := models.Environment{
		Name: fmt.Sprintf("env%s", stamp),
	}
	suite.db.Save(&env)

	feature := models.Feature{
		ProjectId:  p.Model.ID,
		Message:    "message",
		User:       "drshrey",
		Hash:       "hash",
		ParentHash: "parentHash",
		Ref:        "ref",
		Created:    time.Now(),
	}
	suite.db.Save(&feature)

	suite.context = context.WithValue(suite.context, "jwt", utils.Claims{UserId: user.Model.ID.String()})
	suite.user = user
	suite.project = p
	suite.env = env
	suite.feature = feature
}

func (suite *TestRelease) TearDownSuite() {
	suite.db.Exec("delete from users;")
	suite.db.Exec("delete from user_permissions;")
	suite.db.Exec("delete from Releases;")
}

func (suite *TestRelease) TestSuccessfulCreateRelease() {
	suite.SetupDBAndContext()
	// stamp := strings.ToLower("TestSuccessfulCreateRelease")

	releaseInput := struct {
		Release *resolvers.ReleaseInput
	}{
		Release: &resolvers.ReleaseInput{
			ProjectId:     suite.project.Model.ID.String(),
			HeadFeatureId: suite.feature.Model.ID.String(),
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	releaseResolver, _ := resolver.CreateRelease(suite.context, &releaseInput)

	assert.Equal(suite.T(), "Release created", releaseResolver.StateMessage())
	assert.Equal(suite.T(), string(plugins.GetState("waiting")), releaseResolver.State())

	suite.TearDownSuite()
}

func (suite *TestRelease) TestFailedCreateReleaseInvalidHeadFeatureId() {
	suite.SetupDBAndContext()
	// stamp := strings.ToLower("TestSuccessfulCreateRelease")

	releaseInput := struct {
		Release *resolvers.ReleaseInput
	}{
		Release: &resolvers.ReleaseInput{
			ProjectId:     suite.project.Model.ID.String(),
			HeadFeatureId: "invalidfeatureid",
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateRelease(suite.context, &releaseInput)

	assert.Equal(suite.T(), "Couldn't parse headFeatureId", err.Error())

	suite.TearDownSuite()
}

func (suite *TestRelease) TestFailedCreateReleaseInvalidEnvironmentId() {
	suite.SetupDBAndContext()
	// stamp := strings.ToLower("TestSuccessfulCreateRelease")

	releaseInput := struct {
		Release *resolvers.ReleaseInput
	}{
		Release: &resolvers.ReleaseInput{
			ProjectId:     suite.project.Model.ID.String(),
			HeadFeatureId: suite.feature.Model.ID.String(),
			EnvironmentId: "invalidenvid",
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateRelease(suite.context, &releaseInput)

	assert.Equal(suite.T(), "Couldn't parse environmentId", err.Error())

	suite.TearDownSuite()
}

func (suite *TestRelease) TestFailedCreateReleaseInvalidProjectId() {
	suite.SetupDBAndContext()
	// stamp := strings.ToLower("TestSuccessfulCreateRelease")

	releaseInput := struct {
		Release *resolvers.ReleaseInput
	}{
		Release: &resolvers.ReleaseInput{
			ProjectId:     "invalidprojectid",
			HeadFeatureId: suite.feature.Model.ID.String(),
			EnvironmentId: suite.env.Model.ID.String(),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.CreateRelease(suite.context, &releaseInput)

	assert.Equal(suite.T(), "Couldn't parse projectId", err.Error())

	suite.TearDownSuite()
}

func TestReleaseResolvers(t *testing.T) {
	suite.Run(t, new(TestRelease))
}
