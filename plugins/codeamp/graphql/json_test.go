package graphql_resolver_test

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/suite"
)

type JSONTestSuite struct {
	suite.Suite
}

func (suite *JSONTestSuite) SetupTest() {
	// migrators := []interface{}{
	// 	&model.Extension{},
	// }

	// db, err := test.SetupResolverTest(migrators)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// _ = codeamp.CodeAmp{}
	// _ = &graphql_resolver.Resolver{DB: db, Events: nil, Redis: nil}

	// suite.Resolver = &graphql_resolver.Resolver{DB: db}
	// suite.ExtensionResolver = &graphql_resolver.ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: db}}
}

func (ts *JSONTestSuite) TearDownTest() {
}

func TestSuiteJSON(t *testing.T) {
	ts := new(JSONTestSuite)
	suite.Run(t, ts)

	ts.TearDownTest()
}
