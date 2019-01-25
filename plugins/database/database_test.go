package database_test

import (
	"testing"

	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
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
	return
}

func TestDatabase(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
