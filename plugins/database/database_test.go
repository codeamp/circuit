package database_test

import (
	"log"
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

func (suite *DatabaseTestSuite) TestDatabase_Success() {
	log.Println("TestDatabase_Success")

	// inputs
	dbInstanceHost := "0.0.0.0"
	dbAdminUsername := "postgres"
	dbAdminPassword := ""
	dbInstancePort := "5432"
	dbType := "postgresql"

	payload := plugins.ProjectExtension{
		Project: plugins.Project{
			Slug: "foobar",
		},
		Environment: "foobarenv",
	}

	dbProjectExtensionEvent := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("create"), payload)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_HOST", dbInstanceHost, false)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_ADMIN_USERNAME", dbAdminUsername, false)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_ADMIN_PASSWORD", dbAdminPassword, false)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_PORT", dbInstancePort, false)
	dbProjectExtensionEvent.AddArtifact("DB_TYPE", dbType, false)

	suite.transistor.Events <- dbProjectExtensionEvent

	respEvent, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), 60)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	// assert state is complete
	assert.Equal(suite.T(), transistor.GetState("complete"), respEvent.State)

	// assert non-nil artifacts
	dbName, err := respEvent.GetArtifact("DB_NAME")
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	dbUser, err := respEvent.GetArtifact("DB_USER")
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	dbPassword, err := respEvent.GetArtifact("DB_PASSWORD")
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.NotNil(suite.T(), dbName)
	assert.NotNil(suite.T(), dbUser)
	assert.NotNil(suite.T(), dbPassword)

	deleteDBEvent := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("delete"), payload)
	deleteDBEvent.AddArtifact("SHARED_DATABASE_HOST", dbInstanceHost, false)
	deleteDBEvent.AddArtifact("SHARED_DATABASE_ADMIN_USERNAME", dbAdminUsername, false)
	deleteDBEvent.AddArtifact("SHARED_DATABASE_ADMIN_PASSWORD", dbAdminPassword, false)
	deleteDBEvent.AddArtifact("SHARED_DATABASE_PORT", dbInstancePort, false)
	deleteDBEvent.AddArtifact("DB_TYPE", dbType, false)
	deleteDBEvent.AddArtifact("DB_NAME", dbName.String(), false)
	deleteDBEvent.AddArtifact("DB_USER", dbUser.String(), false)

	suite.transistor.Events <- deleteDBEvent
	respEvent, err = suite.transistor.GetTestEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), 60)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.Equal(suite.T(), transistor.GetState("complete"), respEvent.State)
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
