package graphql_resolver_test

import (
	"strings"
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

func (suite *PaginatorTestSuite) SetupTest() {
	migrators := []interface{}{}
	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestProject")
	suite.helper.SetContext(test.ResolverAuthContext())
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
	query := ts.Resolver.DB.Order("created_at desc")

	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.ReleaseListResolver{
		DBReleaseListResolver: &db_resolver.ReleaseListResolver{
			Query:          query,
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	_, err = paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	// assert.NotEqual(ts.T(), strings.Compare(cursor, ""), 0)
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
	query := ts.Resolver.DB.Order("created_at desc")

	paginator := &graphql_resolver.ReleaseListResolver{
		DBReleaseListResolver: &db_resolver.ReleaseListResolver{
			Query:          query,
			PaginatorInput: nil,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	cursor, err := paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), strings.Compare(cursor, ""), 0)
}

func (ts *PaginatorTestSuite) TestSecretListPaginator() {
	query := ts.Resolver.DB.Order("created_at desc")

	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			Query:          query,
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	_, err = paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	// assert.NotEqual(ts.T(), strings.Compare(cursor, ""), 0)
}

func (ts *PaginatorTestSuite) TestSecretListPaginatorNoInput() {
	query := ts.Resolver.DB.Order("created_at desc")

	paginator := &graphql_resolver.SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			Query:          query,
			PaginatorInput: nil,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	cursor, err := paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), strings.Compare(cursor, ""), 0)
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
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)

	// Pagination
	query := ts.Resolver.DB.Order("created_at desc")

	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.ServiceListResolver{
		DBServiceListResolver: &db_resolver.ServiceListResolver{
			Query:          query,
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	_, err = paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	// assert.NotEqual(ts.T(), strings.Compare(cursor, ""), 0)
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
	serviceSpecResolver := ts.helper.CreateServiceSpec(ts.T())

	// Deployment Strategy Input
	deploymentStrategy := model.DeploymentStrategyInput{
		Type: plugins.GetType("recreate"),
	}

	// Services
	ts.helper.CreateService(ts.T(), serviceSpecResolver, projectResolver, &deploymentStrategy)

	// Pagination
	query := ts.Resolver.DB.Order("created_at desc")

	paginator := &graphql_resolver.ServiceListResolver{
		DBServiceListResolver: &db_resolver.ServiceListResolver{
			Query:          query,
			PaginatorInput: nil,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	cursor, err := paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), strings.Compare(cursor, ""), 0)
}

func (ts *PaginatorTestSuite) TestFeatureListPaginator() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, _ := ts.helper.CreateProject(ts.T(), environmentResolver)

	// Feature
	ts.helper.CreateFeatureWithParent(ts.T(), projectResolver)

	// Pagination
	query := ts.Resolver.DB.Order("created_at desc")

	limit := int32(100)
	paginatorInput := model.PaginatorInput{
		Limit: &limit,
	}

	paginator := &graphql_resolver.FeatureListResolver{
		DBFeatureListResolver: &db_resolver.FeatureListResolver{
			Query:          query,
			PaginatorInput: &paginatorInput,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	_, err = paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	// assert.NotEqual(ts.T(), strings.Compare(cursor, ""), 0)
}

func (ts *PaginatorTestSuite) TestFeatureListPaginatorNoInput() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, _ := ts.helper.CreateProject(ts.T(), environmentResolver)

	// Feature
	ts.helper.CreateFeatureWithParent(ts.T(), projectResolver)

	// Pagination
	query := ts.Resolver.DB.Order("created_at desc")

	paginator := &graphql_resolver.FeatureListResolver{
		DBFeatureListResolver: &db_resolver.FeatureListResolver{
			Query:          query,
			PaginatorInput: nil,
			DB:             ts.Resolver.DB,
		},
	}

	entries, err := paginator.Entries()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), entries)
	assert.NotEmpty(ts.T(), entries)

	page, err := paginator.Page()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), page, int32(1))

	cursor, err := paginator.NextCursor()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), page)
	assert.Equal(ts.T(), strings.Compare(cursor, ""), 0)
}

func (ts *PaginatorTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteAuth(t *testing.T) {
	suite.Run(t, new(PaginatorTestSuite))
}
