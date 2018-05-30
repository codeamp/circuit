package graphql_resolver_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	resolver "github.com/codeamp/circuit/plugins/graphql/resolver"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	Resolver *resolver.Resolver
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
		&resolver.Project{},
		&resolver.ProjectEnvironment{},
		&resolver.ProjectBookmark{},
		&resolver.UserPermission{},
		&resolver.ProjectSettings{},
		&resolver.Environment{},
	)
	suite.Resolver = &resolver.Resolver{DB: db}
}

/*
func (suite *ProjectTestSuite) TestCreateProject() {
	// setup
	env := resolver.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	projectInput := resolver.ProjectInput{
		GitProtocol:   "HTTPS",
		GitUrl:        "https://github.com/foo/goo.git",
		EnvironmentID: env.Model.ID.String(),
	}
	authContext := context.WithValue(context.Background(), "jwt", resolver.Claims{
		UserID:      "foo",
		Email:       "foo@gmail.com",
		Permissions: []string{"admin"},
	})

	createProjectResolver, err := suite.Resolver.CreateProject(authContext, &struct {
		Project *resolver.ProjectInput
	}{Project: &projectInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// assert permissions exist for dev env
	assert.Equal(suite.T(), createProjectResolver.Permissions(), []string{env.Model.ID.String()})
	suite.TearDownTest([]string{string(createProjectResolver.ID())})
}
*/

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestUpdateProjectEnvironments() {
	// setup
	project := resolver.Project{
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

	env := resolver.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	projectEnvironmentsInput := resolver.ProjectEnvironmentsInput{
		ProjectID: project.Model.ID.String(),
		Permissions: []resolver.ProjectEnvironmentInput{
			{
				EnvironmentID: env.Model.ID.String(),
				Grant:         true,
			},
		},
	}

	updateProjectEnvironmentsResp, err := suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *resolver.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// check if env is found in response
	assert.Equal(suite.T(), 1, len(updateProjectEnvironmentsResp))
	assert.Equal(suite.T(), env.Model.ID, updateProjectEnvironmentsResp[0].Environment.Model.ID)

	projectEnvironments := []resolver.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 1, len(projectEnvironments))
	assert.Equal(suite.T(), env.Model.ID.String(), projectEnvironments[0].EnvironmentID.String())

	// take away access
	projectEnvironmentsInput.Permissions[0].Grant = false
	updateProjectEnvironmentsResp, err = suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *resolver.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 0, len(updateProjectEnvironmentsResp))

	projectEnvironments = []resolver.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 0, len(projectEnvironments))

	deleteIds := []string{project.Model.ID.String()}
	for _, projectEnvironment := range projectEnvironments {
		deleteIds = append(deleteIds, projectEnvironment.Model.ID.String())
	}

	suite.TearDownTest(deleteIds)
}

func (suite *ProjectTestSuite) TestGetBookmarkedAndQueryProjects() {
	// init 3 projects into db
	projectNames := []string{"foo", "foobar", "boo"}
	userId := uuid.NewV1()
	deleteIds := []string{}

	for _, name := range projectNames {
		project := resolver.Project{
			Name:          name,
			Slug:          name,
			Repository:    fmt.Sprintf("test/%s", name),
			Secret:        "foo",
			GitUrl:        "foo",
			GitProtocol:   "foo",
			RsaPrivateKey: "foo",
			RsaPublicKey:  "foo",
		}

		suite.Resolver.DB.Create(&project)

		projectBookmark := resolver.ProjectBookmark{
			UserID:    userId,
			ProjectID: project.Model.ID,
		}

		suite.Resolver.DB.Create(&projectBookmark)
		deleteIds = append(deleteIds, project.Model.ID.String(),
			projectBookmark.Model.ID.String())
	}

	adminContext := context.WithValue(context.Background(), "jwt", resolver.Claims{
		UserID:      userId.String(),
		Email:       "codeamp",
		Permissions: []string{"admin"},
	})
	projects, err := suite.Resolver.Projects(adminContext, &struct{ ProjectSearch *resolver.ProjectSearchInput }{
		ProjectSearch: &resolver.ProjectSearchInput{
			Bookmarked: true,
		},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 3, len(projects))

	// do a search for 'foo'
	searchQuery := "foo"
	projects, err = suite.Resolver.Projects(adminContext, &struct{ ProjectSearch *resolver.ProjectSearchInput }{
		ProjectSearch: &resolver.ProjectSearchInput{
			Bookmarked: false,
			Repository: &searchQuery,
		},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 2, len(projects))

	suite.TearDownTest(deleteIds)
}

func (suite *ProjectTestSuite) TearDownTest(ids []string) {
	suite.Resolver.DB.Delete(&resolver.Project{})
	suite.Resolver.DB.Delete(&resolver.ProjectEnvironment{})
	suite.Resolver.DB.Delete(&resolver.Environment{})
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
