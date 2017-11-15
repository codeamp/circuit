package resolvers_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/actions"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/resolvers"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestProjects struct {
	suite.Suite
	db      *gorm.DB
	t       *transistor.Transistor
	actions *actions.Actions
}

func (suite *TestProjects) SetupSuite() {

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
		&models.Extension{},
	)

	transistor.RegisterPlugin("codeamp", func() transistor.Plugin { return codeamp.NewCodeAmp() })
	t, _ := transistor.NewTestTransistor(transistor.Config{
		Server:    "",
		Password:  "",
		Database:  "",
		Namespace: "",
		Pool:      "",
		Process:   "",
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

func (suite *TestProjects) TearDownSuite() {
	suite.db.Exec("delete from projects;")
	suite.db.Exec("delete from features;")
	suite.db.Exec("delete from releases;")
	suite.db.Exec("delete from services;")
	suite.db.Exec("delete from environment_variables;")
	suite.db.Exec("delete from extensions;")
}

func (suite *TestProjects) TestSuccessfulCreateProject() {
	stamp := strings.ToLower("TestSuccessfulCreateProject")
	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			GitProtocol: "public",
			GitUrl:      fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, &actions.Actions{})
	projectResolver, _ := resolver.CreateProject(&projectInput)

	assert.Equal(suite.T(), fmt.Sprintf("test/testrepo%s", stamp), projectResolver.Repository())
	assert.Equal(suite.T(), "HTTPS", projectResolver.GitProtocol())
	assert.Equal(suite.T(), fmt.Sprintf("test-testrepo%s", stamp), projectResolver.Slug())
	assert.Equal(suite.T(), fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp), projectResolver.GitUrl())
}

func (suite *TestProjects) TestFailedCreateProjectAlreadyExists() {
	stamp := strings.ToLower("TestFailedCreateProjectAlreadyExists")
	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			GitProtocol: "public",
			GitUrl:      fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	projectResolver, _ := resolver.CreateProject(&projectInput)

	assert.Equal(suite.T(), fmt.Sprintf("test/testrepo%s", stamp), projectResolver.Repository())
	assert.Equal(suite.T(), "HTTPS", projectResolver.GitProtocol())
	assert.Equal(suite.T(), fmt.Sprintf("test-testrepo%s", stamp), projectResolver.Slug())
	assert.Equal(suite.T(), fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp), projectResolver.GitUrl())

	projectResolver, err := resolver.CreateProject(&projectInput)
	assert.Equal(suite.T(), "This repository already exists. Try again with a different git url.", err.Error())
}

func (suite *TestProjects) TestSuccessUpdateProject() {
	stamp := strings.ToLower("TestSuccessUpdateProject")

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
	pId := p.Model.ID.String()

	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			ID:          &pId,
			GitProtocol: "private",
			GitUrl:      fmt.Sprintf("ssh://git@github.com:test/testrepo2%s.git", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	projectResolver, _ := resolver.UpdateProject(&projectInput)

	assert.Equal(suite.T(), "SSH", projectResolver.GitProtocol())
	assert.Equal(suite.T(), fmt.Sprintf("ssh://git@github.com:test/testrepo2%s.git", stamp), projectResolver.GitUrl())
}

func (suite *TestProjects) TestFailedUpdateProjectDoesntExist() {
	stamp := strings.ToLower("TestFailedUpdateProjectDoesntExist")
	fakeId := uuid.NewV1().String()
	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			ID:          &fakeId,
			GitProtocol: "private",
			GitUrl:      fmt.Sprintf("ssh://git@github.com:test/testrepo%s.git", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.UpdateProject(&projectInput)

	assert.Equal(suite.T(), "Project not found.", err.Error())
}

func (suite *TestProjects) TestFailedUpdateProjectMissingArgumentId() {
	stamp := strings.ToLower("TestFailedUpdateProjectMissingArgumentId")
	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			GitProtocol: "private",
			GitUrl:      fmt.Sprintf("ssh://git@github.com:test/testrepo%s.git", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.UpdateProject(&projectInput)

	assert.Equal(suite.T(), "Missing argument id", err.Error())
}

func (suite *TestProjects) TestFailedUpdateProjectInvalidArgumentId() {
	stamp := strings.ToLower("TestFailedUpdateProjectMissingArgumentId")
	fakeId := "invalidfakeid"

	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			ID:          &fakeId,
			GitProtocol: "private",
			GitUrl:      fmt.Sprintf("ssh://git@github.com:test/testrepo%s.git", stamp),
		},
	}
  
	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.UpdateProject(&projectInput)

	assert.Equal(suite.T(), "Invalid argument id", err.Error())
}

func (suite *TestProjects) TestFailedUpdateProjectWithExistingRepoName() {
	stamp := strings.ToLower("TestFailedUpdateProjectWithExistingRepoName")

	p := models.Project{
		Name:          fmt.Sprintf("test/testrepo%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p)

	p2 := models.Project{
		Name:          fmt.Sprintf("test/testrepo2%s", stamp),
		Slug:          fmt.Sprintf("test-testrepo2%s", stamp),
		Repository:    fmt.Sprintf("test/testrepo2%s", stamp),
		Secret:        "",
		GitUrl:        fmt.Sprintf("https://github.com/test/testrepo2%s.git", stamp),
		GitProtocol:   "HTTPS",
		RsaPrivateKey: "",
		RsaPublicKey:  "",
	}
	suite.db.Save(&p2)

	p2Id := p2.Model.ID.String()

	projectInput := struct {
		Project *resolvers.ProjectInput
	}{
		Project: &resolvers.ProjectInput{
			ID:          &p2Id,
			GitProtocol: "private",
			GitUrl:      fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp),
		},
	}

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, suite.actions)
	_, err := resolver.UpdateProject(&projectInput)

	assert.Equal(suite.T(), "Project with repository name already exists.", err.Error())
}

func TestProjectResolvers(t *testing.T) {
	suite.Run(t, new(TestProjects))
}
