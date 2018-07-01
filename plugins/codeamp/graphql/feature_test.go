package graphql_resolver_test

import (
	"context"
	"testing"
	"time"

	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FeatureTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *FeatureTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Feature{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.helper.SetResolver(suite.Resolver, "TestFeature")
	suite.helper.SetContext(test.ResolverAuthContext())
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
	// Environment
	environmentResolver := suite.helper.CreateEnvironment(suite.T())

	// Project
	projectResolver, _ := suite.helper.CreateProject(suite.T(), environmentResolver)

	// Feature
	suite.helper.CreateFeatureWithParent(suite.T(), projectResolver)

	// Test Features Query Interface
	var ctx context.Context
	_, err := suite.Resolver.Features(ctx)
	assert.NotNil(suite.T(), err)

	featureResolvers, err := suite.Resolver.Features(test.ResolverAuthContext())
	assert.Nil(suite.T(), err)
	assert.NotEmpty(suite.T(), featureResolvers)

	featureResolver := featureResolvers[0]
	assert.NotNil(suite.T(), featureResolver)
}

func (suite *FeatureTestSuite) TearDownTest() {
	suite.helper.TearDownTest(suite.T())
}

func TestSuiteFeatureResolver(t *testing.T) {
	suite.Run(t, new(FeatureTestSuite))
}
