package database_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

/*
	test cases:
	- missing pre-req inputs
	- missing inputs
	- api error when creating db on shared Amazon RDS instance
	- success test case
*/

type DatabaseTestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  database:
    workers: 1
`)

func (suite *DatabaseTestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *DatabaseTestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *DatabaseTestSuite) TestDatabase_FailMissingPreRequisiteInputs() {
	// attempt install of database extension with option 'postgres'
	// and omit required pre-req input SHARED_DATABASE_ENDPOINT
	databasePayload := plugins.ProjectExtension{
		Project: plugins.Project{
			Slug:       "checkr-deploy-test",
			Repository: "TestDatabase_FailMissingPreRequisiteInputs",
		},
		Environment: "foo-env",
	}
	ev := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("create"), databasePayload)

	ev.AddArtifact("SHARED_DATABASE_ADMIN_PASSWORD", "password", true)
	ev.AddArtifact("SHARED_DATABASE_ADMIN_USERNAME", "username", true)
	ev.AddArtifact("SHARED_DATABASE_PORT", "5432", true)

	suite.transistor.Events <- ev

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err)
	}

	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("failed"), e.State)
	assert.NotNil(suite.T(), e.StateMessage)
}

func (suite *DatabaseTestSuite) TestDatabase_FailMissingInputs() {
	// attempt install of database extension with option 'postgres'
	// and omit required input environment
	databasePayload := plugins.ProjectExtension{
		Project: plugins.Project{
			Slug:       "checkr-deploy-test",
			Repository: "TestDatabase_FailMissingInputs",
		},
	}
	ev := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("create"), databasePayload)

	ev.AddArtifact("SHARED_DATABASE_ENDPOINT", "https://database-endpoint.com/create", true)
	ev.AddArtifact("SHARED_DATABASE_ADMIN_PASSWORD", "password", true)
	ev.AddArtifact("SHARED_DATABASE_ADMIN_USERNAME", "username", true)
	ev.AddArtifact("SHARED_DATABASE_PORT", "5432", true)

	suite.transistor.Events <- ev

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err)
	}

	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("failed"), e.State)
	assert.NotNil(suite.T(), e.StateMessage)
}

func (suite *DatabaseTestSuite) TestDatabase_FailExternalAPIError() {
	// attempt install of database extension with option 'postgres'
	// and mock failed API error
	databasePayload := plugins.ProjectExtension{
		Project: plugins.Project{
			Slug:       "checkr-deploy-test",
			Repository: "TestDatabase_FailExternalAPIError",
		},
		Environment: "foo-env",
	}
	ev := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("create"), databasePayload)

	ev.AddArtifact("SHARED_DATABASE_ENDPOINT", "https://database-endpoint.com/create", true)
	ev.AddArtifact("SHARED_DATABASE_ADMIN_PASSWORD", "password", true)
	ev.AddArtifact("SHARED_DATABASE_ADMIN_USERNAME", "username", true)
	ev.AddArtifact("SHARED_DATABASE_PORT", "5432", true)

	suite.transistor.Events <- ev

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err)
	}

	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *DatabaseTestSuite) TestDatabase_Success() {
	return
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
