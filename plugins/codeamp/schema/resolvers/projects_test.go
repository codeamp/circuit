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
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestProjects struct {
	suite.Suite
	db *gorm.DB
	t  *transistor.Transistor
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
					"dbname":   "codeamp",
					"sslmode":  "disable",
					"password": "",
				},
			},
		},
		EnabledPlugins: []string{},
		Queueing:       false,
	})

	suite.db = db
	suite.t = t
}

func (suite *TestProjects) TearDownSuite() {
	spew.Dump("dropping test db")
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

	resolver := resolvers.NewResolver(suite.t.TestEvents, suite.db, &actions.Actions{})
	projectResolver, _ := resolver.CreateProject(&projectInput)

	assert.Equal(suite.T(), fmt.Sprintf("test/testrepo%s", stamp), projectResolver.Repository())
	assert.Equal(suite.T(), "HTTPS", projectResolver.GitProtocol())
	assert.Equal(suite.T(), fmt.Sprintf("test-testrepo%s", stamp), projectResolver.Slug())
	assert.Equal(suite.T(), fmt.Sprintf("https://github.com/test/testrepo%s.git", stamp), projectResolver.GitUrl())

	projectResolver, err := resolver.CreateProject(&projectInput)
	assert.Equal(suite.T(), "This repository already exists. Try again with a different git url.", err.Error())
}

func (suite *TestProjects) TestSuccessUpdateProject() {
}

func (suite *TestProjects) TestFailedUpdateProjectAlreadyExists() {
}

func TestProjectResolvers(t *testing.T) {
	suite.Run(t, new(TestProjects))
}
