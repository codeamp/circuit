package graphql_resolver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	graphql "github.com/graph-gophers/graphql-go"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *ProjectTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Project{},
		&model.ProjectBookmark{},
		&model.ProjectEnvironment{},
		&model.ProjectExtension{},
		&model.ProjectSettings{},
		&model.UserPermission{},
		&model.Environment{},
		&model.Extension{},
		&model.Service{},
		&model.ServiceSpec{},
		&model.Secret{},
		&model.Feature{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestProject")
}

func (suite *ProjectTestSuite) TestProjectInterface() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Secret
	_ = suite.helper.CreateSecret(suite.T(), projectResolver)

	// Extension
	extensionResolver := suite.helper.CreateExtension(suite.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := suite.helper.CreateProjectExtension(suite.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	suite.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Test ProjectExtension Interface while its here.
	// ADB This should probably be moved out into a separate test.
	// Leaving this here now until all this is organized a bit better
	// Project Extension Interface
	_ = projectExtensionResolver.ID()

	assert.Equal(suite.T(), projectResolver.ID(), projectExtensionResolver.Project().ID())
	assert.Equal(suite.T(), extensionResolver.ID(), projectExtensionResolver.Extension().ID())

	_ = projectExtensionResolver.Artifacts()

	assert.Equal(suite.T(), model.JSON{[]byte("[]")}, projectExtensionResolver.Config())
	assert.Equal(suite.T(), model.JSON{[]byte("{}")}, projectExtensionResolver.CustomConfig())

	_ = projectExtensionResolver.State()
	_ = projectExtensionResolver.StateMessage()

	environment, err := projectExtensionResolver.Environment()
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), environment)
	assert.Equal(suite.T(), environmentResolver.ID(), environment.ID())

	created_at_diff := time.Now().Sub(projectExtensionResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(suite.T(), "Created at time is too old")
	}

	data, err := projectExtensionResolver.MarshalJSON()
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), data)

	err = projectExtensionResolver.UnmarshalJSON(data)
	assert.Nil(suite.T(), err)

	// Test ProjectExtension Query Interface
	var ctx context.Context
	_, err = suite.Resolver.ProjectExtensions(ctx)
	assert.NotNil(suite.T(), err)

	projectExtensionResolvers, err := suite.Resolver.ProjectExtensions(test.ResolverAuthContext())
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), projectExtensionResolvers)
	assert.NotEmpty(suite.T(), projectExtensionResolvers)

	// Features
	featureResolver := suite.helper.CreateFeature(suite.T(), projectResolver)

	// Releases
	_ = suite.helper.CreateRelease(suite.T(), featureResolver, projectResolver)

	// Test Releases Query Interface
	_, err = suite.Resolver.Releases(ctx, nil)
	assert.NotNil(suite.T(), err)

	releasesList, err := suite.Resolver.Releases(test.ResolverAuthContext(), &struct {
		Params *model.PaginatorInput
	}{nil})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), releasesList)

	// Service Spec ID
	serviceSpecResolver := suite.helper.CreateServiceSpec(suite.T())

	// Services
	_ = suite.helper.CreateService(suite.T(), serviceSpecResolver, projectResolver)

	// Test
	_ = projectResolver.ID()
	_ = projectResolver.Name()
	_ = projectResolver.Repository()
	_ = projectResolver.Secret()

	assert.Equal(suite.T(), "https://github.com/foo/goo.git", projectResolver.GitUrl())
	assert.Equal(suite.T(), "HTTPS", projectResolver.GitProtocol())

	_ = projectResolver.RsaPrivateKey()
	_ = projectResolver.RsaPublicKey()

	showDeployed := false
	featuresList := projectResolver.Features(&struct {
		ShowDeployed *bool
		Params       *model.PaginatorInput
	}{&showDeployed, nil})
	assert.NotEmpty(suite.T(), featuresList, "Features List Empty")

	_, _ = projectResolver.CurrentRelease()

	emptyPaginatorInput := &struct {
		Params *model.PaginatorInput
	}{nil}

	releasesList = projectResolver.Releases(emptyPaginatorInput)
	assert.NotEmpty(suite.T(), releasesList, "Releases List Empty")

	servicesList := projectResolver.Services(emptyPaginatorInput)
	assert.NotEmpty(suite.T(), servicesList, "Services List Empty")

	_, err = projectResolver.Secrets(ctx, emptyPaginatorInput)
	assert.NotNil(suite.T(), err)

	secretsList, err := projectResolver.Secrets(test.ResolverAuthContext(), emptyPaginatorInput)
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), secretsList, "Secrets List Empty")

	extensionsList, err := projectResolver.Extensions()
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), extensionsList, "Extensions List Empty")

	assert.Equal(suite.T(), "master", projectResolver.GitBranch())
	_ = projectResolver.ContinuousDeploy()
	projectEnvironments := projectResolver.Environments()
	assert.NotEmpty(suite.T(), projectEnvironments, "Project Environments Empty")

	_ = projectResolver.Bookmarked(ctx)
	_ = projectResolver.Bookmarked(test.ResolverAuthContext())

	created_at_diff = time.Now().Sub(projectResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(suite.T(), "Created at time is too old")
	}

	data, err = projectResolver.MarshalJSON()
	assert.Nil(suite.T(), err)

	err = projectResolver.UnmarshalJSON(data)
	assert.Nil(suite.T(), err)
}

func (suite *ProjectTestSuite) TestCreateProject() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	_ = suite.helper.CreateProject(suite.T(), environmentResolver)

	// assert permissions exist for dev env
	//assert.Equal(suite.T(), createProjectResolver.Permissions(), []string{env.Model.ID.String()})
}

func (suite *ProjectTestSuite) TestQueryProject() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	initialProjectResolver := suite.helper.CreateProject(suite.T(), environmentResolver)

	var ctx context.Context
	_, err := suite.Resolver.Projects(ctx, &struct {
		ProjectSearch *model.ProjectSearchInput
		Params        *model.PaginatorInput
	}{})
	assert.NotNil(suite.T(), err)

	// do a search for 'foo'
	searchQuery := "foo"
	projects, err := suite.Resolver.Projects(test.ResolverAuthContext(), &struct {
		ProjectSearch *model.ProjectSearchInput
		Params        *model.PaginatorInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: false,
			Repository: &searchQuery,
		},
	})
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), projects)

	envID := string(environmentResolver.ID())
	projectID := initialProjectResolver.ID()

	// By ID - Should Fail
	projectResolver, err := suite.Resolver.Project(ctx, &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectID,
		EnvironmentID: &envID,
	})
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), projectResolver)

	// By ID - Should Succeed
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectID,
		EnvironmentID: &envID,
	})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), projectResolver)

	// By Slug
	projectSlug := projectResolver.Slug()
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		Slug:          &projectSlug,
		EnvironmentID: &envID,
	})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), projectResolver)

	// By Name
	projectName := projectResolver.Name()
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		Name:          &projectName,
		EnvironmentID: &envID,
	})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), projectResolver)

	// Environment Errors
	// No ID
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectID,
		EnvironmentID: nil,
	})
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), projectResolver)

	// Should Fail
	// Not a UUID
	invalidEnvironmentID := "not-a-valid-id"
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectID,
		EnvironmentID: &invalidEnvironmentID,
	})
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), projectResolver)

	// Expected Failure, no ID provided
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            nil,
		EnvironmentID: nil,
	})
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), projectResolver)

	// Permission to access environment
	// Delete the project_environments entry for this
	suite.Resolver.DB.Unscoped().Where("project_id = ?", initialProjectResolver.ID()).Delete(&model.ProjectEnvironment{})

	// Should Fail
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectID,
		EnvironmentID: &envID,
	})
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), projectResolver)

	// Environment does not exist
	// Delete the environment
	suite.Resolver.DB.Unscoped().Where("id = ?", envID).Delete(&model.Environment{})

	// Should fail, no environment exists now
	projectResolver, err = suite.Resolver.Project(test.ResolverAuthContext(), &struct {
		ID            *graphql.ID
		Slug          *string
		Name          *string
		EnvironmentID *string
	}{
		ID:            &projectID,
		EnvironmentID: &envID,
	})
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), projectResolver)
}

/* Test successful project permissions update */
func (suite *ProjectTestSuite) TestUpdateProjectEnvironments() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Update Project Environments
	projectEnvironmentsInput := model.ProjectEnvironmentsInput{
		ProjectID: string(projectResolver.ID()),
		Permissions: []model.ProjectEnvironmentInput{
			{
				EnvironmentID: string(environmentResolver.ID()),
				Grant:         true,
			},
		},
	}

	updateProjectEnvironmentsResp, err := suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{&projectEnvironmentsInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// check if env is found in response
	assert.Equal(suite.T(), 1, len(updateProjectEnvironmentsResp))
	assert.Equal(suite.T(), environmentResolver.ID(), updateProjectEnvironmentsResp[0].ID())
	projectEnvironmentResolvers := projectResolver.Environments()

	assert.Equal(suite.T(), 1, len(projectEnvironmentResolvers), string(projectResolver.ID()))
	assert.Equal(suite.T(), environmentResolver.ID(), projectEnvironmentResolvers[0].ID())

	// take away access
	projectEnvironmentsInput.Permissions[0].Grant = false
	updateProjectEnvironmentsResp, err = suite.Resolver.UpdateProjectEnvironments(nil, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{&projectEnvironmentsInput})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.Empty(suite.T(), updateProjectEnvironmentsResp)
	assert.Empty(suite.T(), projectResolver.Environments(), string(projectResolver.ID()))
}

func (suite *ProjectTestSuite) TestGetBookmarkedAndQueryProjects() {
	// init 3 projects into db
	projectNames := []string{"foo", "foobar", "boo"}

	environmentResolver := suite.helper.CreateEnvironment(suite.T())
	for _, name := range projectNames {
		projectResolver := suite.helper.CreateProjectWithRepo(suite.T(), environmentResolver, fmt.Sprintf("https://github.com/test/%s", name))
		suite.Resolver.BookmarkProject(test.ResolverAuthContext(), &struct{ ID graphql.ID }{projectResolver.ID()})
	}

	projectList, err := suite.Resolver.Projects(test.ResolverAuthContext(), &struct {
		ProjectSearch *model.ProjectSearchInput
		Params        *model.PaginatorInput
	}{
		ProjectSearch: &model.ProjectSearchInput{
			Bookmarked: true,
		},
		Params: &model.PaginatorInput{},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	entries, err := projectList.Entries()
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 3, len(entries))

	// do a search for 'foo'
	searchQuery := "foo"
	projectList, err = suite.Resolver.Projects(test.ResolverAuthContext(), &struct {
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
		assert.FailNow(suite.T(), err.Error())
	}

	entries, err = projectList.Entries()
	if err != nil {
		log.Fatal(err.Error())
	}

	assert.Equal(suite.T(), 2, len(entries))
}

func (suite *ProjectTestSuite) TearDownTest() {
	suite.helper.TearDownTest(suite.T())
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}
