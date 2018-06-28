package graphql_resolver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/test"

	log "github.com/codeamp/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExtensionTestSuite struct {
	suite.Suite
	Resolver          *graphql_resolver.Resolver
	ExtensionResolver *graphql_resolver.ExtensionResolver

	cleanupEnvironmentIDs []uuid.UUID
	cleanupExtensionIDs   []uuid.UUID
}

func (suite *ExtensionTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	_ = codeamp.CodeAmp{}

	suite.Resolver = &graphql_resolver.Resolver{DB: db}
	suite.ExtensionResolver = &graphql_resolver.ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: db}}
}

func (ts *ExtensionTestSuite) TestExtensionInterface() {
	// Environment
	envInput := model.EnvironmentInput{
		Name:      "TestExtensionInterface",
		Key:       "foo",
		IsDefault: true,
		Color:     "color",
	}

	envResolver, err := ts.Resolver.CreateEnvironment(nil, &struct {
		Environment *model.EnvironmentInput
	}{Environment: &envInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupEnvironmentIDs = append(ts.cleanupEnvironmentIDs, envResolver.DBEnvironmentResolver.Environment.Model.ID)
	envId := fmt.Sprintf("%v", envResolver.DBEnvironmentResolver.Environment.Model.ID)

	// Extension
	extensionInput := model.ExtensionInput{
		Name:          "TestExtensionInterface",
		Key:           "test-extension-interface",
		Component:     "",
		EnvironmentID: envId,
		Config:        model.JSON{[]byte("[]")},
		Type:          "workflow",
	}
	extensionResolver, err := ts.Resolver.CreateExtension(&struct {
		Extension *model.ExtensionInput
	}{Extension: &extensionInput})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
	ts.cleanupExtensionIDs = append(ts.cleanupExtensionIDs, extensionResolver.DBExtensionResolver.Extension.Model.ID)

	// Test Extension Interface
	_ = extensionResolver.ID()
	assert.Equal(ts.T(), extensionInput.Name, extensionResolver.Name())
	assert.Equal(ts.T(), extensionInput.Component, extensionResolver.Component())

	assert.Equal(ts.T(), extensionInput.Type, extensionResolver.Type())
	assert.Equal(ts.T(), extensionInput.Key, extensionResolver.Key())

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
	var ctx context.Context
	_, err = ts.Resolver.Extensions(ctx, &struct{ EnvironmentID *string }{&envId})
	assert.NotNil(ts.T(), err)

	extensionResolvers, err := ts.Resolver.Extensions(test.ResolverAuthContext(), &struct{ EnvironmentID *string }{&envId})
	assert.Nil(ts.T(), err)
	assert.NotEmpty(ts.T(), extensionResolvers)
}

func (ts *ExtensionTestSuite) TearDownTest() {
	for _, id := range ts.cleanupExtensionIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Extension{Model: model.Model{ID: id}}).Error
		if err != nil {
			log.Error(err)
		}
	}
	ts.cleanupExtensionIDs = make([]uuid.UUID, 0)

	for _, id := range ts.cleanupEnvironmentIDs {
		err := ts.Resolver.DB.Unscoped().Delete(&model.Environment{Model: model.Model{ID: id}}).Error
		if err != nil {
			assert.FailNow(ts.T(), err.Error())
		}
	}
	ts.cleanupEnvironmentIDs = make([]uuid.UUID, 0)
}

func TestSuiteExtensionResolver(t *testing.T) {
	ts := new(ExtensionTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
