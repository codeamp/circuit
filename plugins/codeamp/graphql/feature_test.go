package graphql_resolver_test

import (
	"context"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/transistor"
	graphql "github.com/graph-gophers/graphql-go"
	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	_ "github.com/codeamp/circuit/plugins/gitsync"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	Resolver   *graphql_resolver.Resolver
	transistor *transistor.Transistor

	helper Helper
}

func (suite *FeatureTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Feature{},
		&model.Project{},
		&model.ProjectSettings{},
		&model.Environment{},
		&model.UserPermission{},
		&model.ProjectBookmark{},
		&model.ProjectEnvironment{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestFeature")
	suite.helper.SetContext(test.ResolverAuthContext())

	httpmock.Activate()
	httpmock.RegisterResponder("GET", "https://github.com/golang/example.git", httpmock.NewStringResponder(200, "{}"))
}

func (suite *FeatureTestSuite) TestCreateFeature() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver, _ := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Feature
	suite.helper.CreateFeatureWithParent(suite.T(), projectResolver)
}

func (suite *FeatureTestSuite) TestFeatureResolverInterface() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver, _ := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Feature
	featureResolver := suite.helper.CreateFeatureWithParent(suite.T(), projectResolver)

	// Test FeatureResolver Interface
	_ = featureResolver.ID()
	_ = featureResolver.Project()
	message := featureResolver.Message()
	assert.Equal(suite.T(), "A test feature message", message)

	user := featureResolver.User()
	assert.Equal(suite.T(), "TestFeature", user)

	hash := featureResolver.Hash()
	assert.Equal(suite.T(), "42941a0900e952f7f78994d53b699aea23926804", hash)

	parentHash := featureResolver.ParentHash()
	assert.Equal(suite.T(), "7f78994d53b699aea239268950441a090952f0e9", parentHash)

	ref := featureResolver.Ref()
	assert.Equal(suite.T(), "refs/heads/master", ref)

	created_at_diff := time.Now().Sub(featureResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(suite.T(), "Created at time is too old")
	}

	data, err := featureResolver.MarshalJSON()
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), data)

	err = featureResolver.UnmarshalJSON(data)
	assert.Nil(suite.T(), err)
}

func (suite *FeatureTestSuite) TestFeatureQueryInterface() {
	emptyPaginatorInput := &struct {
		Params *model.PaginatorInput
	}{nil}

	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver, _ := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Feature
	suite.helper.CreateFeatureWithParent(suite.T(), projectResolver)

	// Test Features Query Interface
	var ctx context.Context
	_, err := suite.Resolver.Features(ctx, emptyPaginatorInput)
	assert.NotNil(suite.T(), err)

	featureListResolver, err := suite.Resolver.Features(test.ResolverAuthContext(), emptyPaginatorInput)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	featureResolvers, err := featureListResolver.Entries()
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), featureResolvers)

	featureResolver := featureResolvers[0]
	assert.NotNil(suite.T(), featureResolver)
}

func (suite *FeatureTestSuite) TestGetGitCommits() {
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver, _ := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Feature
	suite.helper.CreateFeatureWithParent(suite.T(), projectResolver)

	isNew := true

	// Test Features Query Interface
	var ctx context.Context
	eventSent, err := suite.Resolver.GetGitCommits(ctx, &struct {
		ProjectID     graphql.ID
		EnvironmentID graphql.ID
		New           *bool
	}{
		ProjectID:     projectResolver.ID(),
		EnvironmentID: environmentResolver.ID(),
		New:           &isNew,
	})

	for len(suite.Resolver.Events) > 0 {
		<-suite.Resolver.Events
	}

	assert.Nil(suite.T(), err)
	assert.True(suite.T(), eventSent)
}

func (suite *FeatureTestSuite) TearDownTest() {
	httpmock.DeactivateAndReset()
	suite.helper.TearDownTest(suite.T())
	suite.Resolver.DB.Close()
}

func TestSuiteFeatureResolver(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}
