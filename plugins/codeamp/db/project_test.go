package db_resolver_test

import (
	"context"
	"fmt"

	"testing"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver
}

func (suite *ProjectTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectEnvironment{},
		&model.ProjectBookmark{},
		&model.UserPermission{},
		&model.ProjectSettings{},
		&model.Environment{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
}

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestUpdateProjectEnvironments() {
	// setup
	project := model.Project{
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

	env := model.Environment{
		Name:      "dev",
		Color:     "purple",
		Key:       "dev",
		IsDefault: true,
	}
	suite.Resolver.DB.Create(&env)

	projectEnvironmentsInput := model.ProjectEnvironmentsInput{
		ProjectID: project.Model.ID.String(),
		Permissions: []model.ProjectEnvironmentInput{
			{
				EnvironmentID: env.Model.ID.String(),
				Grant:         true,
			},
		},
	}

	updateProjectEnvironmentsResp, err := suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	// check if env is found in response
	assert.Equal(suite.T(), 1, len(updateProjectEnvironmentsResp))
	assert.Equal(suite.T(), env.Model.ID, updateProjectEnvironmentsResp[0].DBEnvironmentResolver.Environment.Model.ID)

	projectEnvironments := []model.ProjectEnvironment{}
	suite.Resolver.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectEnvironments)

	assert.Equal(suite.T(), 1, len(projectEnvironments))
	assert.Equal(suite.T(), env.Model.ID.String(), projectEnvironments[0].EnvironmentID.String())

	// take away access
	projectEnvironmentsInput.Permissions[0].Grant = false
	updateProjectEnvironmentsResp, err = suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{ProjectEnvironments: &projectEnvironmentsInput})
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 0, len(updateProjectEnvironmentsResp))

	projectEnvironments = []model.ProjectEnvironment{}
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
		project := model.Project{
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

		projectBookmark := model.ProjectBookmark{
			UserID:    userId,
			ProjectID: project.Model.ID,
		}

		suite.Resolver.DB.Create(&projectBookmark)
		deleteIds = append(deleteIds, project.Model.ID.String(),
			projectBookmark.Model.ID.String())
	}

	adminContext := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      userId.String(),
		Email:       "codeamp",
		Permissions: []string{"admin"},
	})
	projectList, err := suite.Resolver.Projects(adminContext, &struct {
		ProjectSearch *model.ProjectSearchInput
		Params        *model.PaginatorInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: true,
		},
		Params: &model.PaginatorInput{},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	entries, err := projectList.Entries()
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 3, len(entries))

	// do a search for 'foo'
	searchQuery := "foo"
	projectList, err = suite.Resolver.Projects(adminContext, &struct {
		ProjectSearch *model.ProjectSearchInput
		Params        *model.PaginatorInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: false,
			Repository: &searchQuery,
		},
		Params: &model.PaginatorInput{},
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	entries, err = projectList.Entries()
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 2, len(entries))

	suite.TearDownTest(deleteIds)
}

func (suite *ProjectTestSuite) TearDownTest(ids []string) {
	suite.Resolver.DB.Delete(&model.Project{})
	suite.Resolver.DB.Delete(&model.ProjectEnvironment{})
	suite.Resolver.DB.Delete(&model.Environment{})
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
