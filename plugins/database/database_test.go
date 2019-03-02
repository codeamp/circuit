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
		Project:     plugins.Project{},
		Environment: "",
	}

	dbProjectExtensionEvent := transistor.NewEvent(plugins.GetEventName("project:database"), transistor.GetAction("create"), payload)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_HOST", dbInstanceHost, false)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_ADMIN_USERNAME", dbAdminUsername, false)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_ADMIN_PASSWORD", dbAdminPassword, false)
	dbProjectExtensionEvent.AddArtifact("SHARED_DATABASE_PORT", dbInstancePort, false)
	dbProjectExtensionEvent.AddArtifact("DB_TYPE", dbType, false)

	suite.transistor.Events <- dbProjectExtensionEvent

	e, err := suite.transistor.GetTestEvent(plugins.GetEventName("project:database"), transistor.GetAction("status"), 60)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	dbProjectExtensionEvent.Action = transistor.GetAction("delete")
	suite.transistor.Events <- dbProjectExtensionEvent
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
