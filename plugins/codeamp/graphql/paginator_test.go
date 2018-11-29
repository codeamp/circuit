package graphql_resolver_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"

	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PaginatorTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (ts *PaginatorTestSuite) SetupTest() {
	migrators := []interface{}{}
	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	ts.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	ts.helper.SetResolver(ts.Resolver, "TestProject")
	ts.helper.SetContext(test.ResolverAuthContext())
}

func (ts *PaginatorTestSuite) TestReleaseListPaginator() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Pagination
	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.ReleaseListResolver{
		DBReleaseListResolver: &db_resolver.ReleaseListResolver{
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	count, err := paginator.Count()
	assert.Nil(ts.T(), err)
	assert.NotEqual(ts.T(), 0, count)
}

func (ts *PaginatorTestSuite) TestReleaseListPaginatorNoInput() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Pagination
	paginator := &graphql_resolver.ReleaseListResolver{
		DBReleaseListResolver: &db_resolver.ReleaseListResolver{
			PaginatorInput: nil,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)
}

func (ts *PaginatorTestSuite) TestSecretListPaginator() {
	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	count, err := paginator.Count()
	assert.Nil(ts.T(), err)
	assert.NotEqual(ts.T(), 0, count)
}

func (ts *PaginatorTestSuite) TestSecretListPaginatorNoInput() {
	paginator := &graphql_resolver.SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			PaginatorInput: nil,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)
}

func (ts *PaginatorTestSuite) TestServiceListPaginator() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy, nil, nil, nil)

	// Pagination
	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.ServiceListResolver{
		DBServiceListResolver: &db_resolver.ServiceListResolver{
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	count, err := paginator.Count()
	assert.Nil(ts.T(), err)
	assert.NotEqual(ts.T(), 0, count)
}

func (ts *PaginatorTestSuite) TestServiceListPaginatorNoInput() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), envResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Service Spec ID
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), projectResolver, &deploymentStrategy, nil, nil, nil)

	// Pagination
	paginator := &graphql_resolver.ServiceListResolver{
		DBServiceListResolver: &db_resolver.ServiceListResolver{
			PaginatorInput: nil,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)
}

func (ts *PaginatorTestSuite) TestFeatureListPaginator() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, _ := ts.helper.CreateProject(ts.T(), environmentResolver)

	// Feature
	ts.helper.CreateFeatureWithParent(ts.T(), projectResolver)

	// Pagination
	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.FeatureListResolver{
		DBFeatureListResolver: &db_resolver.FeatureListResolver{
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	count, err := paginator.Count()
	assert.Nil(ts.T(), err)
	assert.NotEqual(ts.T(), 0, count)
}

func (ts *PaginatorTestSuite) TestFeatureListPaginatorNoInput() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, _ := ts.helper.CreateProject(ts.T(), environmentResolver)

	// Feature
	ts.helper.CreateFeatureWithParent(ts.T(), projectResolver)

	// Pagination
	paginator := &graphql_resolver.FeatureListResolver{
		DBFeatureListResolver: &db_resolver.FeatureListResolver{
			PaginatorInput: nil,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)
}

func (ts *PaginatorTestSuite) TestProjectListPaginator() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	ts.helper.CreateProject(ts.T(), environmentResolver)

	// Pagination
	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.ProjectListResolver{
		DBProjectListResolver: &db_resolver.ProjectListResolver{
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	count, err := paginator.Count()
	assert.Nil(ts.T(), err)
	assert.NotEqual(ts.T(), 0, count)
}

func (ts *PaginatorTestSuite) TestProjectListPaginatorNoInput() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	ts.helper.CreateProject(ts.T(), environmentResolver)

	// Pagination
	paginator := &graphql_resolver.ProjectListResolver{
		DBProjectListResolver: &db_resolver.ProjectListResolver{
			PaginatorInput: nil,
			DB:             ts.Resolver.DB.Order("created_at desc"),
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)
}

func (ts *PaginatorTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteAuth(t *testing.T) {
	suite.Run(t, new(PaginatorTestSuite))
}
