package codeamp_resolvers_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	Resolver *resolvers.Resolver
}

func (suite *ProjectTestSuite) SetupTest() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("../../../configs/circuit.test.yml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s",
		viper.GetString("plugins.codeamp.postgres.host"),
		viper.GetString("plugins.codeamp.postgres.port"),
		viper.GetString("plugins.codeamp.postgres.user"),
		viper.GetString("plugins.codeamp.postgres.dbname"),
		viper.GetString("plugins.codeamp.postgres.sslmode"),
		viper.GetString("plugins.codeamp.postgres.password"),
	))
	if err != nil {
		log.Fatal(err.Error())
	}
	db.AutoMigrate(
		&resolvers.Project{},
		&resolvers.ProjectEnvironment{},
		&resolvers.UserPermission{},
		&resolvers.ProjectSettings{},
		&resolvers.Environment{},
	)
	suite.Resolver = &resolvers.Resolver{DB: db}
}

func (suite *ProjectTestSuite) TestCreateProject() {
	// setup
	env := resolvers.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	projectInput := resolvers.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: env.Model.ID.String(),
	}
	authContext := context.WithValue(context.Background(), "jwt", resolvers.Claims{
		UserID:      "foo",
		Email:       "foo@gmail.com",
		Permissions: []string{"admin"},
	})

	createProjectResolver, err := suite.Resolver.CreateProject(authContext, &struct {
		Project *resolvers.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// assert permissions exist for dev env
	assert.Equal(suite.T(), createProjectResolver.Permissions(), []string{env.Model.ID.String()})
	suite.TearDownTest([]string{string(createProjectResolver.ID())})
}

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestUpdateProjectEnvironments() {
	// setup
	project := resolvers.Project{
		Name:          "foo",
		Slug:          "foo",
		Repository:    "foo/foo",
		Secret:        "foo",
		GitUrl:        "foo",
		GitProtocol:   "foo",
		RsaPrivateKey: "foo",
		RsaPublicKey:  "foo",
	}
	suite.Resolver.DB.Create(&project)

	env := resolvers.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	projectEnvironmentsInput := resolvers.ProjectEnvironmentsInput{
		ProjectID: project.Model.ID.String(),
		Permissions: []resolvers.ProjectEnvironmentInput{
			resolvers.ProjectEnvironmentInput{
				EnvironmentID: env.Model.ID.String(),
				Grant:         true,
			},
		},
	}

	updateProjectEnvironmentsResp, err := suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *resolvers.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// check if env is found in response
	assert.Equal(suite.T(), 1, len(updateProjectEnvironmentsResp))
	assert.Equal(suite.T(), env.Model.ID.String(), updateProjectEnvironmentsResp[0])

	projectEnvironments := []resolvers.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 1, len(projectEnvironments))
	assert.Equal(suite.T(), env.Model.ID.String(), projectEnvironments[0].EnvironmentID.String())

	// take away access
	projectEnvironmentsInput.Permissions[0].Grant = false
	updateProjectEnvironmentsResp, err = suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *resolvers.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 0, len(updateProjectEnvironmentsResp))

	projectEnvironments = []resolvers.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 0, len(projectEnvironments))

	deleteIds := []string{project.Model.ID.String()}
	for _, projectEnvironment := range projectEnvironments {
		deleteIds = append(deleteIds, projectEnvironment.Model.ID.String())
	}

	suite.TearDownTest(deleteIds)
}

func (suite *ProjectTestSuite) TearDownTest(ids []string) {
	suite.Resolver.DB.Delete(&resolvers.Project{})
	suite.Resolver.DB.Delete(&resolvers.ProjectEnvironment{})
	suite.Resolver.DB.Delete(&resolvers.Environment{})
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
