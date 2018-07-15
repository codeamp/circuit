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
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExtensionTestSuite struct {
	suite.Suite
	Resolver *graphql_resolver.Resolver

	helper Helper
}

func (suite *ExtensionTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}

	suite.helper.SetResolver(suite.Resolver, "TestExtension")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ExtensionTestSuite) TestCreateExtension() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Extension
	ts.helper.CreateExtension(ts.T(), envResolver)
}

func (ts *ExtensionTestSuite) TestUpdateExtensionSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)

	// Update Extension
	extensionID := string(extensionResolver.ID())
	extensionInput := model.ExtensionInput{
		ID:            &extensionID,
		EnvironmentID: string(envResolver.ID()),
		Config:        model.JSON{[]byte("[]")},
	}

	_, err := ts.Resolver.UpdateExtension(&struct{ Extension *model.ExtensionInput }{&extensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ExtensionTestSuite) TestUpdateExtensionFailureNoEnv() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)

	// Update Extension
	extensionID := string(extensionResolver.ID())
	extensionInput := model.ExtensionInput{
		ID: &extensionID,
	}

	_, err := ts.Resolver.UpdateExtension(&struct{ Extension *model.ExtensionInput }{&extensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ExtensionTestSuite) TestUpdateExtensionFailureNotFound() {
	// Update Extension
	uuid, _ := uuid.FromString("TestUpdateExtensionFailureNotFound")
	extensionID := uuid.String()
	extensionInput := model.ExtensionInput{
		ID: &extensionID,
	}

	_, err := ts.Resolver.UpdateExtension(&struct{ Extension *model.ExtensionInput }{&extensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ExtensionTestSuite) TestDeleteExtensionSuccess() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)

	extensionID := string(extensionResolver.ID())
	extensionInput := model.ExtensionInput{
		ID: &extensionID,
	}
	_, err := ts.Resolver.DeleteExtension(&struct{ Extension *model.ExtensionInput }{&extensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ExtensionTestSuite) TestDeleteExtensionFailureNoID() {
	extensionInput := model.ExtensionInput{}
	_, err := ts.Resolver.DeleteExtension(&struct{ Extension *model.ExtensionInput }{&extensionInput})
	if err != nil {
		assert.NotNil(ts.T(), err)
	}
}

func (ts *ExtensionTestSuite) TestDeleteExtensionFailureNotFound() {
	// Update Extension
	uuid, _ := uuid.FromString("TestUpdateExtensionFailureNotFound")
	extensionID := uuid.String()
	extensionInput := model.ExtensionInput{
		ID: &extensionID,
	}

	_, err := ts.Resolver.DeleteExtension(&struct{ Extension *model.ExtensionInput }{&extensionInput})
	assert.NotNil(ts.T(), err)
}

func (ts *ExtensionTestSuite) TestExtensionInterface() {
	// Environment
	envResolver := ts.helper.CreateEnvironment(ts.T())

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), envResolver)

	// Test Extension Interface
	_ = extensionResolver.ID()
	assert.Equal(ts.T(), "TestExtension", extensionResolver.Name())
	assert.Equal(ts.T(), "test-component", extensionResolver.Component())

	assert.Equal(ts.T(), "workflow", extensionResolver.Type())
	assert.Equal(ts.T(), "test-project-interface", extensionResolver.Key())

	environmentResolver, err := extensionResolver.Environment()
	assert.Nil(ts.T(), err)
	assert.Equal(ts.T(), environmentResolver.ID(), envResolver.ID())

	config := extensionResolver.Config()
	assert.NotNil(ts.T(), config.RawMessage)

	created_at_diff := time.Now().Sub(extensionResolver.Created().Time)
	if created_at_diff.Minutes() > 1 {
		assert.FailNow(ts.T(), "Created at time is too old")
	}

	data, err := extensionResolver.MarshalJSON()
	assert.Nil(ts.T(), err)
	assert.NotNil(ts.T(), data)

	err = extensionResolver.UnmarshalJSON(data)
	assert.Nil(ts.T(), err)

	// Test Query
	envId := string(environmentResolver.ID())

	var ctx context.Context
	_, err = ts.Resolver.Extensions(ctx, &struct{ EnvironmentID *string }{&envId})
	assert.NotNil(ts.T(), err)

	extensionResolvers, err := ts.Resolver.Extensions(test.ResolverAuthContext(), &struct{ EnvironmentID *string }{&envId})
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), extensionResolvers)

	extensionResolvers, err = ts.Resolver.Extensions(test.ResolverAuthContext(), &struct{ EnvironmentID *string }{})
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), extensionResolvers)
}

func (ts *ExtensionTestSuite) TearDownTest() {
	ts.helper.TearDownTest(ts.T())
}

func TestSuiteExtensionResolver(t *testing.T) {
	ts := new(ExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
